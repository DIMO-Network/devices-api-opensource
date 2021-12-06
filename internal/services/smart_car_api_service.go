package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SmartCarService struct {
	baseURL string
	DBS     func() *database.DBReaderWriter
	log     zerolog.Logger // can't remember if best practice with this logger is to use *
}

func NewSmartCarService(apiBaseURL string, dbs func() *database.DBReaderWriter, logger zerolog.Logger) SmartCarService {
	return SmartCarService{
		baseURL: apiBaseURL,
		DBS:     dbs,
		log:     logger,
	}
}

const vehicleInfoJSONNode = "vehicle_info"

func (s *SmartCarService) SeedDeviceDefinitionsFromSmartCar(ctx context.Context) error {
	smartCarVehicleData, err := getSmartCarVehicleData()
	if err != nil {
		return err
	}

	err = s.saveSmartCarDataToDeviceDefs(ctx, smartCarVehicleData)
	return err
}

func (s *SmartCarService) saveSmartCarDataToDeviceDefs(ctx context.Context, data *SmartCarCompatibilityData) error {
	smartCarIntegration, err := s.getOrCreateSmartCarIntegration(ctx)
	if err != nil {
		return err
	}
	// todo: need to loop for each .US .EU .CA
	for _, usData := range data.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US {
		vehicleMake := usData.Name
		if strings.Contains(vehicleMake, "Nissan") || strings.Contains(vehicleMake, "Hyundai") {
			continue // skip if nissan or hyundai b/c not really supported
		}
		/// for now can hard code the position in the row, but later should look up the position
		for _, row := range usData.Rows {
			vehicleModel := null.StringFromPtr(row[0].Text).String
			years := row[0].Subtext                                      // eg. 2017+
			vehicleType := null.StringFromPtr(row[1].VehicleType).String // ICE, PHEV, BEV
			// these indexes may be out of whack
			ic := IntegrationCapabilities{
				Location:          null.StringFromPtr(row[2].Type).String == "check",
				Odometer:          null.StringFromPtr(row[3].Type).String == "check",
				LockUnlock:        null.StringFromPtr(row[4].Type).String == "check",
				EVBattery:         null.StringFromPtr(row[5].Type).String == "check",
				EVChargingStatus:  null.StringFromPtr(row[6].Type).String == "check",
				EVStartStopCharge: null.StringFromPtr(row[7].Type).String == "check",
				FuelTank:          null.StringFromPtr(row[8].Type).String == "check",
				TirePressure:      null.StringFromPtr(row[9].Type).String == "check",
				EngineOilLife:     null.StringFromPtr(row[10].Type).String == "check",
				VehicleAttributes: null.StringFromPtr(row[11].Type).String == "check",
				VIN:               null.StringFromPtr(row[12].Type).String == "check",
			}
			icJSON, err := json.Marshal(&ic)
			if err != nil {
				return err
			}
			// parse out year. todo: will need to create DB record for each year
			startYear := strings.Trim(null.StringFromPtr(years).String, "+")
			startYearInt, err := strconv.Atoi(startYear)
			if err != nil {
				s.log.Warn().Err(err).Msg("could not parse year so can't save smartcar device def to db")
				continue
			}
			dvi := DeviceVehicleInfo{VehicleType: "PASSENGER CAR", FuelType: smartCarVehicleTypeToNhtsaFuelType(vehicleType)}
			dviJSON, err := json.Marshal(map[string]interface{}{vehicleInfoJSONNode: dvi})
			// db operation, note we are not setting vin
			dbDeviceDef := models.DeviceDefinition{
				ID:    ksuid.New().String(),
				Make:  vehicleMake,
				Model: vehicleModel,
				Year:  int16(startYearInt),
			}
			if err != nil {
				dbDeviceDef.Metadata = null.JSONFrom(dviJSON)
			}
			// attach smart car integration in intermediary table
			dbDeviceDef.R = dbDeviceDef.R.NewStruct()
			dbDeviceDef.R.DeviceIntegrations = append(dbDeviceDef.R.DeviceIntegrations, &models.DeviceIntegration{
				IntegrationID: smartCarIntegration.ID,
				Capabilities:  null.JSONFrom(icJSON),
			})
			err = dbDeviceDef.Insert(ctx, s.DBS().Writer, boil.Infer())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SmartCarService) getOrCreateSmartCarIntegration(ctx context.Context) (*models.Integration, error) {
	const (
		smartCarType   = "API"
		smartCarVendor = "SmartCar"
		smartCarStyle  = "Webhook"
	)
	integration, err := models.Integrations(qm.Where("type = ?", smartCarType),
		qm.And("vendors = ?", smartCarVendor),
		qm.And("style = ?", smartCarStyle)).One(ctx, s.DBS().Writer)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// create
			integration = &models.Integration{}
			integration.ID = ksuid.New().String()
			integration.Vendors = smartCarVendor
			integration.Type = smartCarType
			integration.Style = smartCarStyle
			err = integration.Insert(ctx, s.DBS().Writer, boil.Infer())
			if err != nil {
				return nil, errors.Wrap(err, "error inserting smart car integration")
			}
		}
		return nil, errors.Wrap(err, "error fetching smart car integration from database")
	}
	return integration, nil
}

func smartCarVehicleTypeToNhtsaFuelType(vehicleType string) string {
	if vehicleType == "BEV" {
		return "electric"
	}
	return "gasoline"
}

// getSmartCarVehicleData gets all smartcar data on compatibility from their website
func getSmartCarVehicleData() (*SmartCarCompatibilityData, error) {
	const url = "https://smartcar.com/page-data/product/compatible-vehicles/page-data.json"
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received a non 200 response from smart car page. status code: %d", res.StatusCode)
	}

	compatibleVehicles := SmartCarCompatibilityData{}
	err = json.NewDecoder(res.Body).Decode(&compatibleVehicles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal json from smart car")
	}
	return &compatibleVehicles, nil
}

type SmartCarCompatibilityData struct {
	ComponentChunkName string `json:"componentChunkName"`
	Path               string `json:"path"`
	Result             struct {
		Data struct {
			AllMakesTable struct {
				Edges []struct {
					Node struct {
						CompatibilityData struct {
							US []struct {
								Name    string `json:"name"`
								Headers []struct {
									Text    string  `json:"text"`
									Tooltip *string `json:"tooltip"`
								} `json:"headers"`
								Rows [][]struct {
									Color       *string `json:"color"`
									Subtext     *string `json:"subtext"`
									Text        *string `json:"text"`
									Type        *string `json:"type"`
									VehicleType *string `json:"vehicleType"`
								} `json:"rows"`
							} `json:"US"`
						} `json:"compatibilityData"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"allMakesTable"`
		} `json:"data"`
	} `json:"result"`
}

// IntegrationCapabilities gets stored on the association table btw a device_definition and the integrations, device_integrations
type IntegrationCapabilities struct {
	Location          bool `json:"location"`
	Odometer          bool `json:"odometer"`
	LockUnlock        bool `json:"lock_unlock"`
	EVBattery         bool `json:"ev_battery"`
	EVChargingStatus  bool `json:"ev_charging_status"`
	EVStartStopCharge bool `json:"ev_start_stop_charge"`
	FuelTank          bool `json:"fuel_tank"`
	TirePressure      bool `json:"tire_pressure"`
	EngineOilLife     bool `json:"engine_oil_life"`
	VehicleAttributes bool `json:"vehicle_attributes"`
	VIN               bool `json:"vin"`
}
