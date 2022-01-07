package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDeviceDefinitionFromNHTSA(t *testing.T) {
	vinResp := NHTSADecodeVINResponse{}
	_ = json.Unmarshal([]byte(testNhtsaDecodedVin), &vinResp)

	deviceDefinition := NewDeviceDefinitionFromNHTSA(&vinResp)
	var nilString *string

	assert.Equal(t, "", deviceDefinition.DeviceDefinitionID)
	assert.Equal(t, "2020 TESLA MODEL Y", deviceDefinition.Name)
	assert.Equal(t, "Vehicle", deviceDefinition.Type.Type)
	assert.Equal(t, 2020, deviceDefinition.Type.Year)
	assert.Equal(t, "TESLA", deviceDefinition.Type.Make)
	assert.Equal(t, "MODEL Y", deviceDefinition.Type.Model)
	assert.Equal(t, nilString, deviceDefinition.Type.SubModel)
	assert.Equal(t, "PASSENGER CAR", deviceDefinition.VehicleInfo.VehicleType)
	assert.Equal(t, 48000, deviceDefinition.VehicleInfo.BaseMSRP)
	assert.Equal(t, "5", deviceDefinition.VehicleInfo.NumberOfDoors)
	assert.Equal(t, "ELECTRIC", deviceDefinition.VehicleInfo.FuelType)
}
