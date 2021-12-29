package services

import (
	"context"
	"strings"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

//go:generate mockgen -source device_definitions_service.go -destination mocks/device_definitions_service_mock.go

type IDeviceDefinitionService interface {
	FindDeviceDefinitionByMMY(ctx context.Context, db boil.ContextExecutor, mk, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error)
	CheckAndSetImage(dd *models.DeviceDefinition) error
}

type DeviceDefinitionService struct {
	DBS        func() *database.DBReaderWriter
	EdmundsSvc IEdmundsService
	log        *zerolog.Logger
}

func NewDeviceDefinitionService(settings *config.Settings, DBS func() *database.DBReaderWriter, log *zerolog.Logger) *DeviceDefinitionService {
	return &DeviceDefinitionService{DBS: DBS, log: log, EdmundsSvc: NewEdmundsService(settings.TorProxyURL)}
}

// FindDeviceDefinitionByMMY builds and execs query to find device definition for MMY, returns db object and db error if occurs. if db is nil, just uses one from service, useful for tx
func (d *DeviceDefinitionService) FindDeviceDefinitionByMMY(ctx context.Context, tx boil.ContextExecutor, mk, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.Where("make = ?", strings.ToUpper(mk)),
		qm.And("model = ?", strings.ToUpper(model)),
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
func (d *DeviceDefinitionService) CheckAndSetImage(dd *models.DeviceDefinition) error {
	if dd.ImageURL.Valid {
		return nil
	}
	img, err := d.EdmundsSvc.GetDefaultImageForMMY(dd.Make, dd.Model, int(dd.Year))
	if err != nil {
		return err
	}
	dd.ImageURL = null.StringFromPtr(img)
	return nil
}

// todo: refactor lookups that use above logic

// todo: update image if not set on device def, new method for use cases with setting DD default image.
