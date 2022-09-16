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

type DeviceDefinitionService interface {
	FindDeviceDefinitionByMMY(ctx context.Context, mk, model string, year int) (*ddgrpc.GetDeviceDefinitionItemResponse, error)
	CheckAndSetImage(ctx context.Context, dd *ddgrpc.GetDeviceDefinitionItemResponse, overwrite bool) error
	UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error
	PullDrivlyData(ctx context.Context, userDeviceID, deviceDefinitionID string, vin string) error
	PullBlackbookData(ctx context.Context, userDeviceID, deviceDefinitionID string, vin string) error
	GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*models.DeviceMake, error)
	GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) ([]*ddgrpc.GetDeviceDefinitionItemResponse, error)
}

type deviceDefinitionService struct {
	dbs                 func() *database.DBReaderWriter
	edmundsSvc          EdmundsService
	drivlySvc           DrivlyAPIService
	blackbookSvc        BlackbookAPIService
	log                 *zerolog.Logger
	nhtsaSvc            INHTSAService
	definitionsGRPCAddr string
}

func NewDeviceDefinitionService(DBS func() *database.DBReaderWriter, log *zerolog.Logger, nhtsaService INHTSAService, settings *config.Settings) DeviceDefinitionService {
	return &deviceDefinitionService{
		dbs:                 DBS,
		log:                 log,
		edmundsSvc:          NewEdmundsService(settings.TorProxyURL, log),
		nhtsaSvc:            nhtsaService,
		drivlySvc:           NewDrivlyAPIService(settings, DBS),
		blackbookSvc:        NewBlackbookAPIService(settings, DBS),
		definitionsGRPCAddr: settings.DefinitionsGRPCAddr,
	}
}

// GetDeviceDefinitionsByIDs calls device definitions api via GRPC to get the definition. idea for testing: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
func (d *deviceDefinitionService) GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) ([]*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	definitions, err := definitionsClient.GetDeviceDefinitionByID(ctx, &ddgrpc.GetDeviceDefinitionRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}

	return definitions.GetDeviceDefinitions(), nil
}

// FindDeviceDefinitionByMMY builds and execs query to find device definition for MMY, calling out via gRPC. Includes compatible integrations.
func (d *deviceDefinitionService) FindDeviceDefinitionByMMY(ctx context.Context, mk, model string, year int) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// question: does this load the integrations? it should
	dd, err := definitionsClient.GetDeviceDefinitionByMMY(ctx, &ddgrpc.GetDeviceDefinitionByMMYRequest{
		Make:  mk,
		Model: model,
		Year:  int32(year),
	})
	if err != nil {
		return nil, err
	}

	return dd, nil
}

// GetOrCreateMake gets the make from the db or creates it if not found. optional tx - if not passed in uses db writer
func (d *deviceDefinitionService) GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*models.DeviceMake, error) {
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
func (d *deviceDefinitionService) CheckAndSetImage(ctx context.Context, dd *ddgrpc.GetDeviceDefinitionItemResponse, overwrite bool) error {
	if !overwrite {
		return nil
	}
	img, err := d.edmundsSvc.GetDefaultImageForMMY(dd.Type.Make, dd.Type.Model, int(dd.Type.Year))
	if err != nil {
		return err
	}
	if img != nil {
		definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		_, err = definitionsClient.SetDeviceDefinitionImage(ctx, &ddgrpc.UpdateDeviceDefinitionImageRequest{
			DeviceDefinitionId: dd.DeviceDefinitionId,
			ImageUrl:           *img,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateDeviceDefinitionFromNHTSA (deprecated) pulls vin info from nhtsa, and updates the device definition metadata if the MMY from nhtsa matches ours, and the Source is not NHTSA verified
func (d *deviceDefinitionService) UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error {
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

const KmToMilesFactor = 1.609344

type ValuationRequestData struct {
	Mileage float64 `json:"mileage,omitempty"`
	ZipCode string  `json:"zipCode,omitempty"`
}

// PullDrivlyData pulls vin info from drivly, and inserts a record with the data.
// Will only pull if haven't in last 2 weeks. Does not re-pull VIN info, updates DD metadata, sets the device_style_id using the edmunds data pulled.
func (d *deviceDefinitionService) PullDrivlyData(ctx context.Context, userDeviceID string, deviceDefinitionID string, vin string) error {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return errors.Errorf("invalid VIN %s", vin)
	}

	deviceDefinitionResponse, err := d.GetDeviceDefinitionsByIDs(ctx, []string{deviceDefinitionID})
	if err != nil {
		return err
	}

	if len(deviceDefinitionResponse) == 0 {
		return err
	}

	dbDeviceDef := deviceDefinitionResponse[0]

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
		DeviceDefinitionID: null.StringFrom(dbDeviceDef.DeviceDefinitionId),
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
		// extra optional data that only needs to be pulled once.
		edmunds, err := d.drivlySvc.GetEdmundsByVIN(vin)
		if err == nil {
			_ = drivlyData.EdmundsMetadata.Marshal(edmunds)
		}
		build, err := d.drivlySvc.GetBuildByVIN(vin)
		if err == nil {
			_ = drivlyData.BuildMetadata.Marshal(build)
		}

		// todo grpc - update the device definition over grpc with this metadata
		vehicleInfo := &ddgrpc.VehicleInfo{}
		if vinInfo["mpgCity"] != nil && dbDeviceDef.VehicleData.MPGCity == 0 {
			v := fmt.Sprintf("%f", vinInfo["mpgCity"])
			if s, err := strconv.ParseFloat(v, 32); err == nil {
				vehicleInfo.MPGCity = float32(s)
			}
		}
		if vinInfo["mpgHighway"] != nil && dbDeviceDef.VehicleData.MPGHighway == 0 {
			v := fmt.Sprintf("%f", vinInfo["mpgHighway"])
			if s, err := strconv.ParseFloat(v, 32); err == nil {
				vehicleInfo.MPGHighway = float32(s)
			}
		}
		if vinInfo["mpg"] != nil && dbDeviceDef.VehicleData.MPG == 0 {
			v := fmt.Sprintf("%f", vinInfo["mpg"])
			if s, err := strconv.ParseFloat(v, 32); err == nil {
				vehicleInfo.MPG = float32(s)
			}
		}
		if vinInfo["msrpBase"] != nil && dbDeviceDef.VehicleData.Base_MSRP == 0 {
			v := fmt.Sprintf("%s", vinInfo["msrpBase"])
			if s, err := strconv.Atoi(v); err == nil {
				vehicleInfo.Base_MSRP = int32(s)
			}
		}
		if vinInfo["fuelTankCapacityGal"] != nil && dbDeviceDef.VehicleData.FuelTankCapacityGal == 0 {
			v := fmt.Sprintf("%f", vinInfo["fuelTankCapacityGal"])
			if s, err := strconv.ParseFloat(v, 32); err == nil {
				vehicleInfo.FuelTankCapacityGal = float32(s)
			}
		}

		definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		updateResponse, err := definitionsClient.UpdateDeviceDefinition(ctx, &ddgrpc.UpdateDeviceDefinitionRequest{
			DeviceDefinitionId: dbDeviceDef.DeviceDefinitionId,
			VehicleData:        vehicleInfo,
		})

		if err != nil {
			return err
		}

		if updateResponse == nil {
			return err
		}

		// fill in edmunds style_id in our user_device if it exists and not already set. None of these seen as bad errors so just logs
		if edmunds != nil && ud.DeviceStyleID.IsZero() {
			d.setUserDeviceStyleFromEdmunds(ctx, edmunds, ud)
		}

		// future: we could pull some specific data from this and persist in the user_device.metadata
		// future: did MMY from vininfo match the device definition? if not fixup year, or model? but need external_id etc
	}

	// get mileage and zip code for our requests
	var deviceMileage float64
	deviceData, err := models.UserDeviceData(
		models.UserDeviceDatumWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceDatumWhere.Data.IsNotNull(),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(context.Background(), d.dbs().Writer)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	} else {
		deviceOdometer := gjson.GetBytes(deviceData.Data.JSON, "odometer")
		if deviceOdometer.Exists() {
			deviceMileage = deviceOdometer.Float() / KmToMilesFactor
		}
	}
	reqData := ValuationRequestData{
		Mileage: deviceMileage,
		ZipCode: "", // TODO(zavaboy): add vehicle location to zipcode magic to set this zipcode
	}
	_ = drivlyData.RequestMetadata.Marshal(reqData)

	// only pull offers and pricing on every pull.
	offer, err := d.drivlySvc.GetOffersByVIN(vin, reqData.Mileage, reqData.ZipCode)
	if err == nil {
		_ = drivlyData.OfferMetadata.Marshal(offer)
	}
	pricing, err := d.drivlySvc.GetVINPricing(vin, reqData.Mileage, reqData.ZipCode)
	if err == nil {
		_ = drivlyData.PricingMetadata.Marshal(pricing)
	}

	err = drivlyData.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return err
	}

	defer appmetrics.DrivlyIngestTotalOps.Inc()

	return nil
}

// setUserDeviceStyleFromEdmunds given edmunds json, sets the device style_id in the user_device per what edmunds says.
// If errors just logs and continues, since non critical
func (d *deviceDefinitionService) setUserDeviceStyleFromEdmunds(ctx context.Context, edmunds map[string]interface{}, ud *models.UserDevice) {
	edmundsJSON, err := json.Marshal(edmunds)
	if err != nil {
		d.log.Err(err).Msg("could not marshal edmunds response to json")
		return
	}
	styleIDResult := gjson.GetBytes(edmundsJSON, "edmundsStyle.data.style.id")
	styleID := styleIDResult.String()
	if styleIDResult.Exists() && len(styleID) > 0 {
		deviceStyle, err := models.DeviceStyles(models.DeviceStyleWhere.ExternalStyleID.EQ(styleID)).One(ctx, d.dbs().Reader)
		if err != nil {
			d.log.Err(err).Msgf("unable to find device_style for edmunds style_id %s", styleID)
			return
		}
		ud.DeviceStyleID = null.StringFrom(deviceStyle.ID) // set foreign key
		_, err = ud.Update(ctx, d.dbs().Writer, boil.Infer())
		if err != nil {
			d.log.Err(err).Msgf("unable to update user_device_id %s with styleID %s", ud.ID, deviceStyle.ID)
			return
		}
	}
}

// PullBlackbookData pulls vin info from Blackbook, and inserts a record with the data.
// Will only pull if haven't in last 2 weeks.
func (d *deviceDefinitionService) PullBlackbookData(ctx context.Context, userDeviceID string, deviceDefinitionID string, vin string) error {
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

// getDeviceDefsGrpcClient instanties new connection with client to dd service. You must defer conn.close from returned connection
func (d *deviceDefinitionService) getDeviceDefsGrpcClient() (ddgrpc.DeviceDefinitionServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(d.definitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	return definitionsClient, conn, nil
}
