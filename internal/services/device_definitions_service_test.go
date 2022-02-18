package services

import (
	"testing"

	"github.com/DIMO-INC/devices-api/models"
	"github.com/stretchr/testify/assert"
)

func TestSubModelsFromStylesDB(t *testing.T) {
	styles := models.DeviceStyleSlice{
		&models.DeviceStyle{
			ID:       "123",
			SubModel: "XLT",
		},
		&models.DeviceStyle{
			ID:       "124",
			SubModel: "XLT",
		},
		&models.DeviceStyle{
			ID:       "125",
			SubModel: "XLT",
		},
		&models.DeviceStyle{
			ID:       "126",
			SubModel: "Lariat",
		},
		&models.DeviceStyle{
			ID:       "127",
			SubModel: "Lariat",
		},
		&models.DeviceStyle{
			ID:       "127",
			SubModel: "King Cab",
		},
	}

	subModels := SubModelsFromStylesDB(styles)

	assert.Len(t, subModels, 3)
	assert.Equal(t, []string{"King Cab", "Lariat", "XLT"}, subModels)
}
