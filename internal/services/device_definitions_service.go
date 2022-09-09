package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -source device_definitions_service.go -destination mocks/device_definitions_service_mock.go
const vehicleInfoJSONNode = "vehicle_info"

type IDeviceDefinitionService interface {
	FindDeviceDefinitionByMMY(ctx context.Context, db boil.ContextExecutor, mk, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error)
	CheckAndSetImage(dd *models.DeviceDefinition, overwrite bool) error
	UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error
	PullDrivlyData(ctx context.Context, userDeviceID, deviceDefinitionID string, vin string) error
	PullBlackbookData(ctx context.Context, userDeviceID, deviceDefinitionID string, vin string) error
	GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*models.DeviceMake, error)
	GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) (*ddgrpc.GetDeviceDefinitionResponse, error)
}

type DeviceDefinitionService struct {
	dbs                 func() *database.DBReaderWriter
	edmundsSvc          EdmundsService
	drivlySvc           DrivlyAPIService
	blackbookSvc        BlackbookAPIService
	log                 *zerolog.Logger
	nhtsaSvc            INHTSAService
	definitionsGRPCAddr string
}

func NewDeviceDefinitionService(DBS func() *database.DBReaderWriter, log *zerolog.Logger, nhtsaService INHTSAService, settings *config.Settings) *DeviceDefinitionService {
	return &DeviceDefinitionService{dbs: DBS, log: log, edmundsSvc: NewEdmundsService(settings.TorProxyURL, log),
		nhtsaSvc: nhtsaService, drivlySvc: NewDrivlyAPIService(settings, DBS), blackbookSvc: NewBlackbookAPIService(settings, DBS), definitionsGRPCAddr: settings.DefinitionsGRPCAddr}
}

// GetDeviceDefinitionsByIDs calls device definitions api via GRPC to get the definition
func (d *DeviceDefinitionService) GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) (*ddgrpc.GetDeviceDefinitionResponse, error) {
	// to test this we could use http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
	conn, err := grpc.Dial(d.definitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)

	definitions, err := definitionsClient.GetDeviceDefinitionByID(ctx, &ddgrpc.GetDeviceDefinitionRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}

	return definitions, nil
}

// FindDeviceDefinitionByMMY builds and execs query to find device definition for MMY, returns db object and db error if occurs. if db tx is nil, just uses one from service, useful for tx
func (d *DeviceDefinitionService) FindDeviceDefinitionByMMY(ctx context.Context, tx boil.ContextExecutor, mk, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.InnerJoin("device_makes dm on dm.id = device_definitions.device_make_id"),
		qm.Where("dm.name ilike ?", mk),
		qm.And("model ilike ?", model),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
	}
	if loadIntegrations {
		qms = append(qms,
			qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
			qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration)))
	}

	query := models.DeviceDefinitions(qms...)
	if tx == nil {
		tx = d.dbs().Reader
	}
	dd, err := query.One(ctx, tx)
	if err != nil {
		return nil, err
	}
	return dd, nil
}

// GetOrCreateMake gets the make from the db or creates it if not found. optional tx - if not passed in uses db writer
func (d *DeviceDefinitionService) GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*models.DeviceMake, error) {
	if tx == nil {
		tx = d.dbs().Writer
	}
	m, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(strings.TrimSpace(makeName))).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// create
			m = &models.DeviceMake{
				ID:   ksuid.New().String(),
				Name: makeName,
			}
			err = m.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return nil, errors.Wrapf(err, "error inserting make: %s", makeName)
			}
			return m, nil
		}
		return nil, errors.Wrapf(err, "error querying for make: %s", makeName)
	}
	return m, nil
}

// CheckAndSetImage just checks if the device definitions has an image set, and if not gets it from edmunds and sets it. does not update DB. This process could take a few seconds.
func (d *DeviceDefinitionService) CheckAndSetImage(dd *models.DeviceDefinition, overwrite bool) error {
	if !overwrite && dd.ImageURL.Valid {
		return nil
	}
	if dd.R.DeviceMake == nil {
		return errors.New("device make relation is required in dd.R.DeviceMake")
	}
	img, err := d.edmundsSvc.GetDefaultImageForMMY(dd.R.DeviceMake.Name, dd.Model, int(dd.Year))
	if err != nil {
		return err
	}
	if img != nil {
		dd.ImageURL = null.StringFromPtr(img)
	}
	return nil
}

// UpdateDeviceDefinitionFromNHTSA (deprecated) pulls vin info from nhtsa, and updates the device definition metadata if the MMY from nhtsa matches ours, and the Source is not NHTSA verified
func (d *DeviceDefinitionService) UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error {
	dbDeviceDef, err := models.DeviceDefinitions(
		models.DeviceDefinitionWhere.ID.EQ(deviceDefinitionID),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
	).One(ctx, d.dbs().Reader)
	if err != nil {
		return err
	}
	nhtsaDecode, err := d.nhtsaSvc.DecodeVIN(vin)
	if err != nil {
		return err
	}
	dd := NewDeviceDefinitionFromNHTSA(nhtsaDecode)
	if dd.Type.Make == dbDeviceDef.R.DeviceMake.Name && dd.Type.Model == dbDeviceDef.Model && int16(dd.Type.Year) == dbDeviceDef.Year {
		if !(dbDeviceDef.Verified && dbDeviceDef.Source.String == "NHTSA") {
			// update our device definition metadata `vehicle_info` with latest from nhtsa
			err = dbDeviceDef.Metadata.Marshal(map[string]interface{}{vehicleInfoJSONNode: dd.VehicleInfo})
			if err != nil {
				return err
			}
			dbDeviceDef.Verified = true
			dbDeviceDef.Source = null.StringFrom("NHTSA")
			_, err = dbDeviceDef.Update(ctx, d.dbs().Writer, boil.Infer())
			if err != nil {
				return err
			}
		}
	} else {
		// just log for now if no MMY match.
		d.log.Warn().Msgf("No MMY match between deviceDefinitionID: %s and NHTSA for VIN: %s, %s", deviceDefinitionID, vin, dd.Name)
	}

	return nil
}

// PullDrivlyData pulls vin info from drivly, and inserts a record with the data.
// Will only pull if haven't in last 2 weeks. Does not re-pull VIN info, updates DD metadata, sets the device_style_id using the edmunds data pulled.
func (d *DeviceDefinitionService) PullDrivlyData(ctx context.Context, userDeviceID string, deviceDefinitionID string, vin string) error {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return errors.Errorf("invalid VIN %s", vin)
	}

	dbDeviceDef, err := models.FindDeviceDefinition(ctx, d.dbs().Reader, deviceDefinitionID)
	if err != nil {
		return err
	}
	neverPulled := false
	existingData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.Vin.EQ(vin),
		models.ExternalVinDatumWhere.PricingMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)
	if errors.Is(err, sql.ErrNoRows) {
		neverPulled = true
	} else if err != nil {
		return err
	}
	// just return if already pulled recently for this VIN
	if existingData != nil && existingData.UpdatedAt.Add(repullWindow).After(time.Now()) {
		return nil
	}
	ud, err := models.FindUserDevice(ctx, d.dbs().Reader, userDeviceID)
	if err != nil {
		return err
	}

	// by this point we know we need to insert drivly raw json data
	drivlyData := &models.ExternalVinDatum{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(dbDeviceDef.ID),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}
	if neverPulled {
		vinInfo, err := d.drivlySvc.GetVINInfo(vin)
		if err != nil {
			return errors.Wrapf(err, "error getting VIN %s. skipping", vin)
		}
		err = drivlyData.VinMetadata.Marshal(vinInfo)
		if err != nil {
			return err
		}
		// todo grpc - update the device definition over grpc with this metadata
		metaData := new(DeviceVehicleInfo) // make as pointer
		if err := dbDeviceDef.Metadata.Unmarshal(metaData); err == nil {
			if vinInfo["mpgCity"] != nil && metaData.MPGCity == "" {
				metaData.MPGCity = fmt.Sprintf("%f", vinInfo["mpgCity"])
			}
			if vinInfo["mpgHighway"] != nil && metaData.MPGHighway == "" {
				metaData.MPGHighway = fmt.Sprintf("%f", vinInfo["mpgHighway"])
			}
			if vinInfo["mpg"] != nil && metaData.MPG == "" {
				metaData.MPG = fmt.Sprintf("%f", vinInfo["mpg"])
			}
			if vinInfo["msrpBase"] != nil && metaData.BaseMSRP == 0 {
				metaData.BaseMSRP, _ = strconv.Atoi(fmt.Sprintf("%s", vinInfo["msrpBase"]))
			}
			if vinInfo["fuelTankCapacityGal"] != nil && metaData.FuelTankCapacityGal == "" {
				metaData.FuelTankCapacityGal = fmt.Sprintf("%f", vinInfo["fuelTankCapacityGal"])
			}
		}
		err = dbDeviceDef.Metadata.Marshal(metaData)
		if err != nil {
			return err
		}

		_, err = dbDeviceDef.Update(ctx, d.dbs().Writer, boil.Infer())
		if err != nil {
			return err
		}
		// future: we could pull some specific data from this and persist in the user_device.metadata
		// future: did MMY from vininfo match the device definition? if not fixup year, or model? but need external_id etc
	}
	// As we understand what data changes and what doesn't, as well as what raw sources here are unnecessary to pull, we can clean this up.
	summary, err := d.drivlySvc.GetExtendedOffersByVIN(vin)
	if err != nil {
		return err
	}

	_ = drivlyData.PricingMetadata.Marshal(summary.Pricing)
	_ = drivlyData.BuildMetadata.Marshal(summary.Build)
	_ = drivlyData.OfferMetadata.Marshal(summary.Offers)
	_ = drivlyData.AutocheckMetadata.Marshal(summary.AutoCheck)
	_ = drivlyData.CargurusMetadata.Marshal(summary.Cargurus)
	_ = drivlyData.CarmaxMetadata.Marshal(summary.Carmax)
	_ = drivlyData.KBBMetadata.Marshal(summary.KBB)
	_ = drivlyData.CarstoryMetadata.Marshal(summary.Carstory)
	_ = drivlyData.CarvanaMetadata.Marshal(summary.Carvana)
	_ = drivlyData.EdmundsMetadata.Marshal(summary.Edmunds)
	_ = drivlyData.TMVMetadata.Marshal(summary.TMV)
	_ = drivlyData.VroomMetadata.Marshal(summary.VRoom)

	err = drivlyData.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return err
	}

	// fill in edmunds style_id in our user_device if it exists and not already set. None of these seen as bad errors so just logs & returns nil
	if summary.Edmunds != nil && ud.DeviceStyleID.IsZero() {
		edmundsJSON, err := json.Marshal(summary.Edmunds)
		if err != nil {
			d.log.Err(err).Msg("could not marshal edmunds response to json")
			return nil
		}
		styleIDResult := gjson.GetBytes(edmundsJSON, "edmundsStyle.data.style.id")
		styleID := styleIDResult.String()
		if styleIDResult.Exists() && len(styleID) > 0 {
			deviceStyle, err := models.DeviceStyles(models.DeviceStyleWhere.ExternalStyleID.EQ(styleID)).One(ctx, d.dbs().Reader)
			if err != nil {
				d.log.Err(err).Msgf("unable to find device_style for edmunds style_id %s", styleID)
				return nil
			}
			ud.DeviceStyleID = null.StringFrom(deviceStyle.ID) // set foreign key
			_, err = ud.Update(ctx, d.dbs().Writer, boil.Infer())
			if err != nil {
				d.log.Err(err).Msgf("unable to update user_device_id %s with styleID %s", ud.ID, deviceStyle.ID)
				return nil
			}
		}
	}

	defer appmetrics.DrivlyIngestTotalOps.Inc()

	return nil
}

// PullBlackbookData pulls vin info from Blackbook, and inserts a record with the data.
// Will only pull if haven't in last 2 weeks.
func (d *DeviceDefinitionService) PullBlackbookData(ctx context.Context, userDeviceID string, deviceDefinitionID string, vin string) error {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return errors.Errorf("invalid VIN %s", vin)
	}

	dbDeviceDef, err := models.FindDeviceDefinition(ctx, d.dbs().Reader, deviceDefinitionID)
	if err != nil {
		return err
	}
	existingData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.Vin.EQ(vin),
		models.ExternalVinDatumWhere.BlackbookMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)
	if err != nil {
		return err
	}
	// just return if already pulled recently for this VIN
	if existingData != nil && existingData.UpdatedAt.Add(repullWindow).After(time.Now()) {
		return nil
	}

	// by this point we know we need to insert blackbook raw json data
	blackbookData := &models.ExternalVinDatum{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(dbDeviceDef.ID),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}

	vinInfo, err := d.blackbookSvc.GetVINInfo(vin, "")
	if err != nil {
		return errors.Wrapf(err, "error getting VIN %s. skipping", vin)
	}
	err = blackbookData.BlackbookMetadata.UnmarshalJSON(vinInfo)
	if err != nil {
		return err
	}

	err = blackbookData.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return err
	}

	defer appmetrics.BlackbookRequestTotalOps.Inc()

	return nil
}

// SubModelsFromStylesDB gets the unique style.SubModel from the styles slice, deduping sub_model
func SubModelsFromStylesDB(styles models.DeviceStyleSlice) []string {
	items := map[string]string{}
	for _, style := range styles {
		if _, ok := items[style.SubModel]; !ok {
			items[style.SubModel] = style.Name
		}
	}

	sm := make([]string, len(items))
	i := 0
	for key := range items {
		sm[i] = key
		i++
	}
	sort.Strings(sm)
	return sm
}
