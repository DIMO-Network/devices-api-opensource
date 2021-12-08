package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	scIntegrationID, err := s.getOrCreateSmartCarIntegration(ctx)
	if err != nil {
		return err
	}
	// future: loop for each other country .EU .CA - difference is in integration capability but MMY may be the same.
	for _, usData := range data.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US {
		vehicleMake := usData.Name
		if strings.Contains(vehicleMake, "Nissan") || strings.Contains(vehicleMake, "Hyundai") {
			continue // skip if nissan or hyundai b/c not really supported
		}

		for _, row := range usData.Rows {
			vehicleModel := null.StringFromPtr(row[0].Text).String
			years := row[0].Subtext                                      // eg. 2017+ or 2012-2017
			vehicleType := null.StringFromPtr(row[1].VehicleType).String // ICE, PHEV, BEV

			ic := IntegrationCapabilities{
				Location:          getCapability("Location", usData.Headers, row),
				Odometer:          getCapability("Odometer", usData.Headers, row),
				LockUnlock:        getCapability("Lock & unlock", usData.Headers, row),
				EVBattery:         getCapability("EV battery", usData.Headers, row),
				EVChargingStatus:  getCapability("EV charging status", usData.Headers, row),
				EVStartStopCharge: getCapability("EV start & stop charge", usData.Headers, row),
				FuelTank:          getCapability("Fuel tank", usData.Headers, row),
				TirePressure:      getCapability("Tire pressure", usData.Headers, row),
				EngineOilLife:     getCapability("Engine oil life", usData.Headers, row),
				VehicleAttributes: getCapability("Vehicle attributes", usData.Headers, row),
				VIN:               getCapability("VIN", usData.Headers, row),
			}
			icJSON, err := json.Marshal(&ic)
			if err != nil {
				return err
			}
			dvi := DeviceVehicleInfo{VehicleType: "PASSENGER CAR", FuelType: smartCarVehicleTypeToNhtsaFuelType(vehicleType)}

			yearRange, err := parseSmartCarYears(years)
			if err != nil {
				return errors.Wrapf(err, "could not parse years: %s", *years)
			}

			tx, err := s.DBS().Writer.DB.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			// future: put below code in own function so we can defer tx.rollback here https://manse.cloud/posts/go-footuns-go-defer-rust-drop
			// loop over each year and insert into device definition same stuff just changing year
			for _, yr := range yearRange {
				err := s.saveDeviceDefinition(ctx, tx, vehicleMake, vehicleModel, yr, dvi, icJSON, scIntegrationID, "us")
				if err != nil {
					_ = tx.Rollback()
					return errors.Wrapf(err, "could not save device definition to db for mmy: %s %s %d", vehicleMake, vehicleModel, yr)
				}
			}
			err = tx.Commit()
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	return nil
}

// saveDeviceDefinition does not commit or rollback the transaction, just operates the insert
func (s *SmartCarService) saveDeviceDefinition(ctx context.Context, tx *sql.Tx, make, model string, year int, dvi DeviceVehicleInfo, icJSON []byte, integrationID string, integrationCountry string) error {
	// todo: idempotency - read all info from DB singleton and then compare MMY, but integration capabilities vary by country

	// db operation, note we are not setting vin
	dbDeviceDef := models.DeviceDefinition{
		ID:    ksuid.New().String(),
		Make:  make,
		Model: model,
		Year:  int16(year),
	}
	err := dbDeviceDef.Metadata.Marshal(map[string]interface{}{vehicleInfoJSONNode: dvi})
	if err != nil {
		s.log.Warn().Err(err).Msg("could not marshal DeviceVehicleInfo for DeviceDefinition metadata")
	}

	err = dbDeviceDef.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return err
	}
	// attach smart car integration in intermediary table
	deviceIntegration := &models.DeviceIntegration{
		IntegrationID:      integrationID,
		DeviceDefinitionID: dbDeviceDef.ID,
		Capabilities:       null.JSONFrom(icJSON),
		Country:            integrationCountry,
	}
	return deviceIntegration.Insert(ctx, tx, boil.Infer())
}

// getHdrIdxForCapability gets the column index based on matching header name, so you can get a row value
func getHdrIdxForCapability(capabilityName string, headers []struct {
	Text    string  `json:"text"`
	Tooltip *string `json:"tooltip"`
}) int {
	for i, header := range headers {
		if strings.EqualFold(header.Text, capabilityName) {
			return i
		}
	}
	return -1
}

func getCapability(capabilityName string, headers []struct {
	Text    string  `json:"text"`
	Tooltip *string `json:"tooltip"`
}, row []struct {
	Color       *string `json:"color"`
	Subtext     *string `json:"subtext"`
	Text        *string `json:"text"`
	Type        *string `json:"type"`
	VehicleType *string `json:"vehicleType"`
}) bool {
	hdrIdx := getHdrIdxForCapability(capabilityName, headers)
	// note that in some there are 12 header cols vs 13 row cols.
	rowIdx := hdrIdx + len(row) - len(headers)
	if rowIdx > len(row) {
		return false
	}
	return null.StringFromPtr(row[rowIdx].Type).String == "check"
}

// parseSmartCarYears parses out the years format in the smartcar document and returns an array of years
func parseSmartCarYears(yearsPtr *string) ([]int, error) {
	if yearsPtr == nil || len(*yearsPtr) == 0 {
		return nil, errors.New("years string was nil")
	}
	years := *yearsPtr
	if len(years) > 4 {
		var rangeYears []int
		startYear := years[:4]
		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			return nil, errors.Errorf("could not parse start year from: %s", years)
		}
		endYear := time.Now().Year()
		if strings.Contains(years, "-") {
			eyStr := years[5:]
			endYear, err = strconv.Atoi(eyStr)
			if err != nil {
				return nil, errors.Errorf("could not parse end year from: %s", years)
			}
		}
		for y := startYearInt; y <= endYear; y++ {
			rangeYears = append(rangeYears, y)
		}
		return rangeYears, nil
	}
	y, err := strconv.Atoi(years)
	if err != nil {
		return nil, errors.Errorf("could not parse single year from: %s", years)
	}
	return []int{y}, nil
}

func (s *SmartCarService) getOrCreateSmartCarIntegration(ctx context.Context) (string, error) {
	const (
		smartCarType   = "API"
		smartCarVendor = "SmartCar"
		smartCarStyle  = models.IntegrationStyleWebhook
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
				return "", errors.Wrap(err, "error inserting smart car integration")
			}
		} else {
			return "", errors.Wrap(err, "error fetching smart car integration from database")
		}
	}
	return integration.ID, nil
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
