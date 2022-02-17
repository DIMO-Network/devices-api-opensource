package services

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

//go:embed test_smart_car_vehicles.json
var testSmartCarVehicles string

func Test_parseSmartCarYears(t *testing.T) {
	singleYear := "2019"
	rangeYear := "2012-2015"
	plusYears := "2019+" // figure out way to stub out time so we always get same. https://stackoverflow.com/questions/18970265/is-there-an-easy-way-to-stub-out-time-now-globally-during-test
	garbage := "bobby"
	empty := ""

	tests := []struct {
		name     string
		yearsPtr *string
		want     []int
		wantErr  bool
	}{
		{
			name:     "nil",
			yearsPtr: nil,
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "empty",
			yearsPtr: &empty,
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "parse single year",
			yearsPtr: &singleYear,
			want:     []int{2019},
			wantErr:  false,
		},
		{
			name:     "parse range year",
			yearsPtr: &rangeYear,
			want:     []int{2012, 2013, 2014, 2015},
			wantErr:  false,
		},
		{
			name:     "parse year plus",
			yearsPtr: &plusYears,
			want:     []int{2019, 2020, 2021, 2022},
			wantErr:  false,
		},
		{
			name:     "error on garbage",
			yearsPtr: &garbage,
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSmartCarYears(tt.yearsPtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSmartCarYears() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSmartCarYears() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSmartCarVehicleData(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "https://smartcar.com/page-data/product/compatible-vehicles/page-data.json"
	httpmock.RegisterResponder(http.MethodGet, url, httpmock.NewStringResponder(200, testSmartCarVehicles))

	data, err := GetSmartCarVehicleData()
	assert.NoError(t, err)

	assert.Equal(t, "Audi", data.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US[0].Name)
	assert.Equal(t, "Location", data.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US[0].Headers[1].Text)
	assert.Equal(t, "A3", *data.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US[0].Rows[0][0].Text)
}

func Test_getColIdxForCapability(t *testing.T) {
	compatibleVehicles := SmartCarCompatibilityData{}
	err := json.Unmarshal([]byte(testSmartCarVehicles), &compatibleVehicles)
	assert.NoError(t, err)

	hdrs := compatibleVehicles.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US[0].Headers

	locationIdx := getHdrIdxForCapability("Location", hdrs)
	assert.Equal(t, 1, locationIdx)

	lockIdx := getHdrIdxForCapability("Lock & unlock", hdrs)
	assert.Equal(t, 3, lockIdx)

	vinIdx := getHdrIdxForCapability("VIN", hdrs)
	assert.Equal(t, 11, vinIdx)
}

func Test_getCapability(t *testing.T) {
	compatibleVehicles := SmartCarCompatibilityData{}
	err := json.Unmarshal([]byte(testSmartCarVehicles), &compatibleVehicles)
	assert.NoError(t, err)

	hdrs := compatibleVehicles.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US[0].Headers
	row := compatibleVehicles.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US[0].Rows[0]
	locCap := GetCapability("Location", hdrs, row)
	assert.True(t, locCap, "expected location to be true")

	odoCap := GetCapability("Odometer", hdrs, row)
	assert.True(t, odoCap, "expected odometer to be true")

	lockUnLockCap := GetCapability("Lock & unlock", hdrs, row)
	assert.True(t, lockUnLockCap, "expected lock and unlock to be true")

	evBattery := GetCapability("EV Battery", hdrs, row)
	assert.False(t, evBattery, "expected ev battery to be false")

	evChargeStat := GetCapability("EV Charging Status", hdrs, row)
	assert.False(t, evChargeStat, "expected EV charge status to be false")

	fuelTank := GetCapability("Fuel Tank", hdrs, row)
	assert.True(t, fuelTank, "expected fuel to be true")

	vinCap := GetCapability("VIN", hdrs, row)
	assert.True(t, vinCap, "expected vin capability to be true")
}
