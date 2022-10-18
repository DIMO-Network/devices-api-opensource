package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
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
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -source device_definitions_service.go -destination mocks/device_definitions_service_mock.go

type DeviceDefinitionService interface {
	FindDeviceDefinitionByMMY(ctx context.Context, mk, model string, year int) (*ddgrpc.GetDeviceDefinitionItemResponse, error)
	CheckAndSetImage(ctx context.Context, dd *ddgrpc.GetDeviceDefinitionItemResponse, overwrite bool) error
	UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error
	PullDrivlyData(ctx context.Context, userDeviceID, deviceDefinitionID string, vin string) error
	PullBlackbookData(ctx context.Context, userDeviceID, deviceDefinitionID string, vin string) error
	GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*ddgrpc.DeviceMake, error)
	GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) ([]*ddgrpc.GetDeviceDefinitionItemResponse, error)
	GetDeviceDefinitionByID(ctx context.Context, id string) (*ddgrpc.GetDeviceDefinitionItemResponse, error)
	GetIntegrations(ctx context.Context) ([]*ddgrpc.Integration, error)
	GetIntegrationByID(ctx context.Context, id string) (*ddgrpc.Integration, error)
	GetIntegrationByVendor(ctx context.Context, vendor string) (*ddgrpc.Integration, error)
	GetIntegrationByFilter(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error)
	CreateIntegration(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error)
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

func (d *deviceDefinitionService) CreateIntegration(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error) {

	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	integration, err := definitionsClient.CreateIntegration(ctx, &ddgrpc.CreateIntegrationRequest{
		Vendor: vendor,
		Type:   integrationType,
		Style:  style,
	})

	if err != nil {
		return nil, err
	}

	return &ddgrpc.Integration{Id: integration.Id, Vendor: vendor, Type: integrationType, Style: style}, nil
}

// GetDeviceDefinitionsByIDs calls device definitions api via GRPC to get the definition. idea for testing: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
// if not found or other error from server, the error contains the grpc status code that can be interpreted for different conditions. example in api.GrpcErrorToFiber
func (d *deviceDefinitionService) GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) ([]*ddgrpc.GetDeviceDefinitionItemResponse, error) {

	if len(ids) == 0 {
		return nil, errors.New("Device Definition Ids is required")
	}

	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	definitions, err2 := definitionsClient.GetDeviceDefinitionByID(ctx, &ddgrpc.GetDeviceDefinitionRequest{
		Ids: ids,
	})

	if err2 != nil {
		return nil, err2
	}

	return definitions.GetDeviceDefinitions(), nil
}

// GetDeviceDefinitionByID is a helper for calling GetDeviceDefinitionsByIDs with one id.
func (d *deviceDefinitionService) GetDeviceDefinitionByID(ctx context.Context, id string) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	resp, err := d.GetDeviceDefinitionsByIDs(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, errors.New("no definitions returned")
	}

	return resp[0], nil
}

// GetIntegrations calls device definitions integrations api via GRPC to get the definition. idea for testing: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
func (d *deviceDefinitionService) GetIntegrations(ctx context.Context) ([]*ddgrpc.Integration, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	definitions, err := definitionsClient.GetIntegrations(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc endpoint GetIntegrations")
	}

	return definitions.GetIntegrations(), nil
}

// GetIntegrationByID get integration from grpc by id
func (d *deviceDefinitionService) GetIntegrationByID(ctx context.Context, id string) (*ddgrpc.Integration, error) {
	allIntegrations, err := d.GetIntegrations(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc to get integrations")
	}
	var integration *ddgrpc.Integration
	for _, in := range allIntegrations {
		if in.Id == id {
			integration = in
		}
	}
	if integration == nil {
		return nil, fmt.Errorf("no integration with id %s found in the %d existing", id, len(allIntegrations))
	}

	return integration, nil
}

func (d *deviceDefinitionService) GetIntegrationByFilter(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error) {
	allIntegrations, err := d.GetIntegrations(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc to get integrations")
	}
	var integration *ddgrpc.Integration
	for _, in := range allIntegrations {
		if in.Type == integrationType && in.Vendor == vendor && in.Style == style {
			integration = in
		}
	}
	if integration == nil {
		return nil, nil
	}

	return integration, nil
}

func (d *deviceDefinitionService) GetIntegrationByVendor(ctx context.Context, vendor string) (*ddgrpc.Integration, error) {
	allIntegrations, err := d.GetIntegrations(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc to get integrations")
	}
	var integration *ddgrpc.Integration
	for _, in := range allIntegrations {
		if in.Vendor == vendor {
			integration = in
		}
	}
	if integration == nil {
		return nil, fmt.Errorf("no integration with vendor %s found in the %d existing", vendor, len(allIntegrations))
	}

	return integration, nil
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
		return nil, errors.Wrap(err, "failed to call grpc endpoint GetDeviceDefinitionByMMY")
	}

	return dd, nil
}

// GetOrCreateMake gets the make from the db or creates it if not found. optional tx - if not passed in uses db writer
func (d *deviceDefinitionService) GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*ddgrpc.DeviceMake, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// question: does this load the integrations? it should
	dm, err := definitionsClient.CreateDeviceMake(ctx, &ddgrpc.CreateDeviceMakeRequest{
		Name: makeName,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc endpoint CreateDeviceMake")
	}

	return &ddgrpc.DeviceMake{Id: dm.Id, Name: makeName}, nil
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

	deviceDefinitionResponse, err := d.GetDeviceDefinitionsByIDs(ctx, []string{deviceDefinitionID})
	if err != nil {
		return err
	}

	if len(deviceDefinitionResponse) == 0 {
		return errors.New("Device definition empty")
	}

	dbDeviceDef := deviceDefinitionResponse[0]

	nhtsaDecode, err := d.nhtsaSvc.DecodeVIN(vin)
	if err != nil {
		return err
	}
	dd := NewDeviceDefinitionFromNHTSA(nhtsaDecode)
	if dd.Type.Make == dbDeviceDef.Make.Name && dd.Type.Model == dbDeviceDef.Type.Model && int16(dd.Type.Year) == int16(dbDeviceDef.Type.Year) {
		if !(dbDeviceDef.Verified && dbDeviceDef.Source == "NHTSA") {
			definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
			if err != nil {
				return err
			}
			defer conn.Close()

			_, err = definitionsClient.UpdateDeviceDefinition(ctx, &ddgrpc.UpdateDeviceDefinitionRequest{
				DeviceDefinitionId: dbDeviceDef.DeviceDefinitionId,
				Verified:           true,
				Source:             "NHTSA",
				Year:               dbDeviceDef.Type.Year,
				Model:              dbDeviceDef.Type.Model,
				ImageUrl:           dbDeviceDef.ImageUrl,
				VehicleData:        dbDeviceDef.VehicleData,
			})

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

const MilesToKmFactor = 1.609344 // there is 1.609 kilometers in a mile. const should probably be KmToMilesFactor
const EstMilesPerYear = 12000.0

type ValuationRequestData struct {
	Mileage *float64 `json:"mileage,omitempty"`
	ZipCode *string  `json:"zipCode,omitempty"`
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
		return errors.New("Device definition empty")
	}

	deviceDef := deviceDefinitionResponse[0]

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
	externalVinData := &models.ExternalVinDatum{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(deviceDef.DeviceDefinitionId),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}
	if neverPulled {
		vinInfo, err := d.drivlySvc.GetVINInfo(vin)
		if err != nil {
			return errors.Wrapf(err, "error getting VIN %s. skipping", vin)
		}
		err = externalVinData.VinMetadata.Marshal(vinInfo)
		if err != nil {
			return err
		}
		// extra optional data that only needs to be pulled once.
		edmunds, err := d.drivlySvc.GetEdmundsByVIN(vin)
		if err == nil {
			_ = externalVinData.EdmundsMetadata.Marshal(edmunds)
		}
		build, err := d.drivlySvc.GetBuildByVIN(vin)
		if err == nil {
			_ = externalVinData.BuildMetadata.Marshal(build)
		}

		// pull out vehicleInfo useful data to update our device definition over gRPC. Do NOT update if property already has value.
		vehicleInfo := &ddgrpc.VehicleInfo{}
		if vinInfo["mpgCity"] != nil && deviceDef.VehicleData.MPGCity == 0 {
			v := fmt.Sprintf("%f", vinInfo["mpgCity"])
			if s, err := strconv.ParseFloat(v, 32); err == nil {
				vehicleInfo.MPGCity = float32(s)
			}
		}
		if vinInfo["mpgHighway"] != nil && deviceDef.VehicleData.MPGHighway == 0 {
			v := fmt.Sprintf("%f", vinInfo["mpgHighway"])
			if s, err := strconv.ParseFloat(v, 32); err == nil {
				vehicleInfo.MPGHighway = float32(s)
			}
		}
		if vinInfo["mpg"] != nil && deviceDef.VehicleData.MPG == 0 {
			v := fmt.Sprintf("%f", vinInfo["mpg"])
			if s, err := strconv.ParseFloat(v, 32); err == nil {
				vehicleInfo.MPG = float32(s)
			}
		}
		if vinInfo["msrpBase"] != nil && deviceDef.VehicleData.Base_MSRP == 0 {
			v := fmt.Sprintf("%s", vinInfo["msrpBase"])
			if s, err := strconv.Atoi(v); err == nil {
				vehicleInfo.Base_MSRP = int32(s)
			}
		}
		if vinInfo["fuelTankCapacityGal"] != nil && deviceDef.VehicleData.FuelTankCapacityGal == 0 {
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

		_, err = definitionsClient.UpdateDeviceDefinition(ctx, &ddgrpc.UpdateDeviceDefinitionRequest{
			DeviceDefinitionId: deviceDef.DeviceDefinitionId,
			VehicleData:        vehicleInfo,
		})
		if err != nil {
			// just log if can't update device-definition
			d.log.Err(err).Str("vin", vin).Str("deviceDefinitionID", deviceDefinitionID).
				Msg("failed to update device definition over gRPC")
		}

		// fill in edmunds style_id in our user_device if it exists and not already set. None of these seen as bad errors so just logs
		if edmunds != nil && ud.DeviceStyleID.IsZero() {
			d.setUserDeviceStyleFromEdmunds(ctx, edmunds, ud)
		}

		// future: we could pull some specific data from this and persist in the user_device.metadata
		// future: did MMY from vininfo match the device definition? if not fixup year, or model? but need external_id etc
	}

	// get mileage for our requests
	deviceMileage, err := d.getDeviceMileage(userDeviceID, int(deviceDef.Type.Year))
	if err != nil {
		return err
	}
	// TODO(zavaboy): get zipcode of this vehicle for better valuations

	reqData := ValuationRequestData{
		Mileage: deviceMileage,
	}
	_ = externalVinData.RequestMetadata.Marshal(reqData)

	// only pull offers and pricing on every pull.
	offer, err := d.drivlySvc.GetOffersByVIN(vin, &reqData)
	if err == nil {
		_ = externalVinData.OfferMetadata.Marshal(offer)
	}
	pricing, err := d.drivlySvc.GetVINPricing(vin, &reqData)
	if err == nil {
		_ = externalVinData.PricingMetadata.Marshal(pricing)
	}

	err = externalVinData.Insert(ctx, d.dbs().Writer, boil.Infer())
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

		definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
		if err != nil {
			return
		}
		defer conn.Close()

		deviceStyle, err := definitionsClient.GetDeviceStyleByExternalID(ctx, &ddgrpc.GetDeviceStyleByIDRequest{
			Id: styleID,
		})

		if err != nil {
			d.log.Err(err).Msgf("unable to find device_style for edmunds style_id %s", styleID)
			return
		}
		ud.DeviceStyleID = null.StringFrom(deviceStyle.Id) // set foreign key
		_, err = ud.Update(ctx, d.dbs().Writer, boil.Infer())
		if err != nil {
			d.log.Err(err).Msgf("unable to update user_device_id %s with styleID %s", ud.ID, deviceStyle.Id)
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

	deviceDefinitionResponse, err := d.GetDeviceDefinitionsByIDs(ctx, []string{deviceDefinitionID})
	if err != nil {
		return err
	}

	if len(deviceDefinitionResponse) == 0 {
		return errors.New("Device definition empty")
	}

	dbDeviceDef := deviceDefinitionResponse[0]

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
		DeviceDefinitionID: null.StringFrom(dbDeviceDef.DeviceDefinitionId),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}

	// get mileage for our requests
	deviceMileage, err := d.getDeviceMileage(userDeviceID, int(dbDeviceDef.Type.Year))
	if err != nil {
		return err
	}
	// TODO(zavaboy): get zipcode of this vehicle for better valuations

	reqData := ValuationRequestData{
		Mileage: deviceMileage,
	}
	_ = blackbookData.RequestMetadata.Marshal(reqData)

	vinInfo, err := d.blackbookSvc.GetVINInfo(vin, &reqData)
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

// getDeviceDefsGrpcClient instanties new connection with client to dd service. You must defer conn.close from returned connection
func (d *deviceDefinitionService) getDeviceDefsGrpcClient() (ddgrpc.DeviceDefinitionServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(d.definitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	return definitionsClient, conn, nil
}

func (d *deviceDefinitionService) getDeviceMileage(udID string, modelYear int) (mileage *float64, err error) {
	var deviceMileage *float64

	// Get user device odometer
	deviceData, err := models.UserDeviceData(
		models.UserDeviceDatumWhere.UserDeviceID.EQ(udID),
		models.UserDeviceDatumWhere.Data.IsNotNull(),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(context.Background(), d.dbs().Writer)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		deviceOdometer := gjson.GetBytes(deviceData.Data.JSON, "odometer")
		if deviceOdometer.Exists() {
			deviceMileage = new(float64)
			*deviceMileage = deviceOdometer.Float() / MilesToKmFactor
		}
	}

	// Estimate mileage based on model year
	if deviceMileage == nil {
		deviceMileage = new(float64)
		yearDiff := time.Now().Year() - modelYear
		switch {
		case yearDiff > 0:
			// Past model year
			*deviceMileage = float64(yearDiff) * EstMilesPerYear
		case yearDiff == 0:
			// Current model year
			*deviceMileage = EstMilesPerYear / 2
		default:
			// Next model year
			*deviceMileage = 0
		}
	}

	return deviceMileage, nil
}
