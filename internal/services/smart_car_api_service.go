package services

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type SmartCarService struct {
	BaseURL string
}

// possibly rename this as we break it up
func getSmartCarVehicleData() error {
	const url = "https://smartcar.com/page-data/product/compatible-vehicles/page-data.json"
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("received a non 200 response from smart car page. status code: %d", res.StatusCode)
	}

	compatibleVehicles := SmartCarCompatibilityData{}
	err = json.NewDecoder(res.Body).Decode(&compatibleVehicles)
	if err != nil {
		return errors.Wrap(err, "failed to marshal json from smart car")
	}

	// pull above json into memory and put into a struct, or use json parsing

	// transpose data into something to insert
	// do we start tracking country of origin?
	// todo: migration to make squishvin nullable

	// store data in db. integration capabilities - where to store? in intermediary table as jsonb. Only issue is would we use fixes
	// struct or dynamic map, i feel like struct - and we can change it in the future. Log if there are unmapped properties?
	// lookup if make, model and year exists, if not create new record with no vin
	// create if not exists integration for smart car.
	// create if not exists link btw smart car integration and device, if exists check if capabilities are any different

	return nil
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
	Location bool `json:"location"`
	Odometer bool `json:"odometer"`
	LockUnlock bool `json:"lock_unlock"`

}