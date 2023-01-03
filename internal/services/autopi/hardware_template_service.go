package autopi

import (
	"fmt"
	"strconv"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

type HardwareTemplateService interface {
	GetTemplateID(ud *models.UserDevice, dd *ddgrpc.GetDeviceDefinitionItemResponse, integ *ddgrpc.Integration) (string, error)
}

type hardwareTemplateService struct {
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
