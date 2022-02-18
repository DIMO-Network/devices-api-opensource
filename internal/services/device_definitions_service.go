package services

import (
	"context"
	"sort"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

//go:generate mockgen -source device_definitions_service.go -destination mocks/device_definitions_service_mock.go

const vehicleInfoJSONNode = "vehicle_info"

type IDeviceDefinitionService interface {
	FindDeviceDefinitionByMMY(ctx context.Context, db boil.ContextExecutor, mk, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error)
	CheckAndSetImage(dd *models.DeviceDefinition, overwrite bool) error
	UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error
}

type DeviceDefinitionService struct {
	DBS        func() *database.DBReaderWriter
	EdmundsSvc IEdmundsService
	log        *zerolog.Logger
	nhtsaSvc   INHTSAService
}

func NewDeviceDefinitionService(settings *config.Settings, DBS func() *database.DBReaderWriter, log *zerolog.Logger, nhtsaService INHTSAService) *DeviceDefinitionService {
	return &DeviceDefinitionService{DBS: DBS, log: log, EdmundsSvc: NewEdmundsService(settings.TorProxyURL, log), nhtsaSvc: nhtsaService}
}

// FindDeviceDefinitionByMMY builds and execs query to find device definition for MMY, returns db object and db error if occurs. if db tx is nil, just uses one from service, useful for tx
func (d *DeviceDefinitionService) FindDeviceDefinitionByMMY(ctx context.Context, tx boil.ContextExecutor, mk, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.Where("make ilike ?", mk),
		qm.And("model ilike ?", model),
		qm.And("year = ?", year),
	}
	if loadIntegrations {
		qms = append(qms,
			qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
			qm.Load("DeviceIntegrations.Integration"))
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

// CheckAndSetImage just checks if the device definitions has an image set, and if not gets it from edmunds and sets it. does not update DB. This process could take a few seconds.
func (d *DeviceDefinitionService) CheckAndSetImage(dd *models.DeviceDefinition, overwrite bool) error {
	if !overwrite && dd.ImageURL.Valid {
		return nil
	}
	img, err := d.EdmundsSvc.GetDefaultImageForMMY(dd.Make, dd.Model, int(dd.Year))
	if err != nil {
		return err
	}
	if img != nil {
		dd.ImageURL = null.StringFromPtr(img)
	}
	return nil
}

// UpdateDeviceDefinitionFromNHTSA pulls vin info from nhtsa, and updates the device definition metadata if the MMY from nhtsa matches ours, and the Source is not NHTSA verified
func (d *DeviceDefinitionService) UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error {
	dbDeviceDef, err := models.FindDeviceDefinition(ctx, d.DBS().Reader, deviceDefinitionID)
	if err != nil {
		return err
	}
	nhtsaDecode, err := d.nhtsaSvc.DecodeVIN(vin)
	if err != nil {
		return err
	}
	dd := NewDeviceDefinitionFromNHTSA(nhtsaDecode)
	if dd.Type.Make == dbDeviceDef.Make && dd.Type.Model == dbDeviceDef.Model && int16(dd.Type.Year) == dbDeviceDef.Year {
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
