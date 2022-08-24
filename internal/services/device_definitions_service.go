package services

import (
	"context"
	"database/sql"
	"sort"
	"strings"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
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
	GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*models.DeviceMake, error)
	GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) (*ddgrpc.GetDeviceDefinitionResponse, error)
}

type DeviceDefinitionService struct {
	DBS                 func() *database.DBReaderWriter
	EdmundsSvc          EdmundsService
	DrivlySvc           DrivlyAPIService
	log                 *zerolog.Logger
	nhtsaSvc            INHTSAService
	definitionsGRPCAddr string
}

func NewDeviceDefinitionService(DBS func() *database.DBReaderWriter, log *zerolog.Logger, nhtsaService INHTSAService, settings *config.Settings) *DeviceDefinitionService {
	return &DeviceDefinitionService{DBS: DBS, log: log, EdmundsSvc: NewEdmundsService(settings.TorProxyURL, log),
		nhtsaSvc: nhtsaService, DrivlySvc: NewDrivlyAPIService(settings, DBS), definitionsGRPCAddr: settings.DefinitionsGRPCAddr}
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
		tx = d.DBS().Reader
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
		tx = d.DBS().Writer
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
	img, err := d.EdmundsSvc.GetDefaultImageForMMY(dd.R.DeviceMake.Name, dd.Model, int(dd.Year))
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
	).One(ctx, d.DBS().Reader)
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
			_, err = dbDeviceDef.Update(ctx, d.DBS().Writer, boil.Infer())
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

// PullDrivlyData pulls vin info from drivly, and updates the device definition metadata
func (d *DeviceDefinitionService) PullDrivlyData(ctx context.Context, userDeviceID string, deviceDefinitionID string, vin string) error {
	dbDeviceDef, err := models.DeviceDefinitions(
		models.DeviceDefinitionWhere.ID.EQ(deviceDefinitionID),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
	).One(ctx, d.DBS().Reader)
	if err != nil {
		return err
	}

	// insert drivly raw json data
	drivlyData := &models.DrivlyDatum{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(dbDeviceDef.ID),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}

	summary, err := d.DrivlySvc.GetSummaryByVIN(vin)
	if err != nil {
		return err
	}

	_ = drivlyData.VinMetadata.Marshal(summary.VIN)
	_ = drivlyData.BuildMetadata.Marshal(summary.Build)
	_ = drivlyData.AutocheckMetadata.Marshal(summary.AutoCheck)
	_ = drivlyData.CargurusMetadata.Marshal(summary.Cargurus)
	_ = drivlyData.CarmaxMetadata.Marshal(summary.Carmax)
	_ = drivlyData.KBBMetadata.Marshal(summary.KBB)
	_ = drivlyData.CarstoryMetadata.Marshal(summary.Carstory)
	_ = drivlyData.CarvanaMetadata.Marshal(summary.Carvana)
	_ = drivlyData.EdmundsMetadata.Marshal(summary.Edmunds)
	_ = drivlyData.OfferMetadata.Marshal(summary.Offers)
	_ = drivlyData.TMVMetadata.Marshal(summary.TMV)
	_ = drivlyData.VroomMetadata.Marshal(summary.VRoom)

	err = drivlyData.Insert(ctx, d.DBS().Writer, boil.Infer())

	if err != nil {
		return err
	}

	defer appmetrics.DrivlyIngestTotalOps.Inc()

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
