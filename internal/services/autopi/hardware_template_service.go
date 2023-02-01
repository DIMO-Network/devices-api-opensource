package autopi

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/DIMO-Network/shared/api/devices"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
)

type HardwareTemplateService interface {
	GetTemplateID(ud *models.UserDevice, dd *ddgrpc.GetDeviceDefinitionItemResponse, integ *ddgrpc.Integration) (string, error)
	ApplyHardwareTemplate(ctx context.Context, req *pb.ApplyHardwareTemplateRequest) (*pb.ApplyHardwareTemplateResponse, error)
}

type hardwareTemplateService struct {
	dbs func() *db.ReaderWriter
	ap  services.AutoPiAPIService
}

func NewHardwareTemplateService() HardwareTemplateService {
	return &hardwareTemplateService{}
}

func (a *hardwareTemplateService) GetTemplateID(ud *models.UserDevice, dd *ddgrpc.GetDeviceDefinitionItemResponse, integ *ddgrpc.Integration) (string, error) {

	if ud.DeviceStyleID.Valid {
		if len(dd.DeviceStyles) > 0 {
			for _, item := range dd.DeviceStyles {
				if item.Id == ud.DeviceStyleID.String {
					return item.HardwareTemplateId, nil
				}
			}
		}
	}

	if len(dd.HardwareTemplateId) > 0 {
		return dd.HardwareTemplateId, nil
	}

	if len(dd.DeviceStyles) > 0 {
		for _, item := range dd.DeviceStyles {
			if len(item.HardwareTemplateId) > 0 {
				return item.HardwareTemplateId, nil
			}
		}
	}

	if len(dd.Make.HardwareTemplateId) > 0 {
		return dd.Make.HardwareTemplateId, nil
	}

	if integ.AutoPiDefaultTemplateId > 0 {
		if integ.AutoPiPowertrainTemplate != nil {
			udMd := services.UserDeviceMetadata{}
			err := ud.Metadata.Unmarshal(&udMd)
			if err != nil {
				return "", err
			}

			powertrainToTemplateID := powertrainToTemplate(udMd.PowertrainType, integ)

			return strconv.Itoa(int(powertrainToTemplateID)), nil
		}
		return strconv.Itoa(int(integ.AutoPiDefaultTemplateId)), nil
	}

	return "", fmt.Errorf("integration lacks a default template")
}

func (a *hardwareTemplateService) ApplyHardwareTemplate(ctx context.Context, req *pb.ApplyHardwareTemplateRequest) (*pb.ApplyHardwareTemplateResponse, error) {
	tx, err := a.dbs().Writer.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	udapi, err := models.UserDeviceAPIIntegrations(
		qm.Where("user_device_id = ?", req.UserDeviceId),
		qm.And("auto_pi_unit_id = ?", req.AutoApiUnitId),
	).One(ctx, tx)
	if err != nil {
		return nil, err
	}

	autoPiModel, err := models.AutopiUnits(
		models.AutopiUnitWhere.AutopiUnitID.EQ(req.AutoApiUnitId),
	).One(ctx, tx)
	if err != nil {
		return nil, err
	}

	autoPi, err := a.ap.GetDeviceByUnitID(autoPiModel.AutopiUnitID)
	if err != nil {
		return nil, err
	}

	err = a.ap.UnassociateDeviceTemplate(autoPi.ID, autoPi.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to unassociate template %d", autoPi.Template)
	}

	hardwareTemplateID, err := strconv.Atoi(req.HardwareTemplateId)
	if err != nil {
		return nil, err
	}

	// set our template on the autoPiDevice
	err = a.ap.AssociateDeviceToTemplate(autoPi.ID, hardwareTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to associate autoPiDevice %s to template %d", autoPi.ID, hardwareTemplateID)
	}

	// apply for next reboot
	err = a.ap.ApplyTemplate(autoPi.ID, hardwareTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to apply autoPiDevice %s with template %d", autoPi.ID, hardwareTemplateID)
	}

	// send sync command in case autoPiDevice is on at this moment (should be during initial setup)
	_, err = a.ap.CommandSyncDevice(ctx, autoPi.UnitID, autoPi.ID, req.UserDeviceId)
	if err != nil {
		return nil, fmt.Errorf("failed to sync changes to autoPiDevice %s", autoPi.ID)
	}

	udMetadata := services.UserDeviceAPIIntegrationsMetadata{
		AutoPiUnitID:          &autoPi.UnitID,
		AutoPiIMEI:            &autoPi.IMEI,
		AutoPiTemplateApplied: &hardwareTemplateID,
	}

	err = udapi.Metadata.Marshal(udMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall user device integration metadata")
	}

	_, err = udapi.Update(ctx, tx, boil.Whitelist(models.UserDeviceColumns.Metadata, models.UserDeviceColumns.UpdatedAt))
	if err != nil {
		return nil, fmt.Errorf("failed to update user device status to Pending")
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit new hardware template to user device")
	}

	return &pb.ApplyHardwareTemplateResponse{Applied: true}, nil
}
