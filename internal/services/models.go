package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// DeviceDefinition represents a device for to clients in generic form, ie. not specific to a user
type DeviceDefinition struct {
	DeviceDefinitionID string  `json:"deviceDefinitionId"`
	Name               string  `json:"name"`
	ImageURL           *string `json:"imageUrl"`
	// CompatibleIntegrations has systems this vehicle can integrate with
	CompatibleIntegrations []DeviceCompatibility `json:"compatibleIntegrations"`
	Type                   DeviceType            `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo DeviceVehicleInfo `json:"vehicleData,omitempty"`
	Metadata    interface{}       `json:"metadata"`
	Verified    bool              `json:"verified"`
}

// DeviceCompatibility represents what systems we know this is compatible with
type DeviceCompatibility struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Style        string          `json:"style"`
	Vendor       string          `json:"vendor"`
	Region       string          `json:"region"`
	Country      string          `json:"country,omitempty"`
	Capabilities json.RawMessage `json:"capabilities"`
}

// DeviceType whether it is a vehicle or other type and basic information
type DeviceType struct {
	// Type is eg. Vehicle, E-bike, roomba
	Type      string   `json:"type"`
	Make      string   `json:"make"`
	Model     string   `json:"model"`
	Year      int      `json:"year"`
	SubModels []string `json:"subModels"`
}

// DeviceVehicleInfo represents some standard vehicle specific properties stored in the metadata json field in DB
type DeviceVehicleInfo struct {
	FuelType      string `json:"fuel_type,omitempty"`
	DrivenWheels  string `json:"driven_wheels,omitempty"`
	NumberOfDoors string `json:"number_of_doors,omitempty"`
	BaseMSRP      int    `json:"base_msrp,omitempty"`
	EPAClass      string `json:"epa_class,omitempty"`
	VehicleType   string `json:"vehicle_type,omitempty"` // VehicleType PASSENGER CAR, from NHTSA
	MPGHighway    string `json:"mpg_highway,omitempty"`
	MPGCity       string `json:"mpg_city,omitempty"`
}

// Converters

// NewDeviceDefinitionFromNHTSA converts nhtsa response into our standard device definition struct
func NewDeviceDefinitionFromNHTSA(decodedVin *NHTSADecodeVINResponse) DeviceDefinition {
	dd := DeviceDefinition{}
	yr, _ := strconv.Atoi(decodedVin.LookupValue("Model Year"))
	msrp, _ := strconv.Atoi(decodedVin.LookupValue("Base Price ($)"))
	dd.Type = DeviceType{
		Type:  "Vehicle",
		Make:  decodedVin.LookupValue("Make"),
		Model: decodedVin.LookupValue("Model"),
		Year:  yr,
	}
	dd.Name = fmt.Sprintf("%d %s %s", dd.Type.Year, dd.Type.Make, dd.Type.Model)
	dd.VehicleInfo = DeviceVehicleInfo{
		FuelType:      decodedVin.LookupValue("Fuel Type - Primary"),
		NumberOfDoors: decodedVin.LookupValue("Doors"),
		BaseMSRP:      msrp,
		VehicleType:   decodedVin.LookupValue("Vehicle Type"),
	}

	return dd
}

type VehicleDriveType string

const (
	ICE  VehicleDriveType = "ICE"
	HEV  VehicleDriveType = "HEV"
	PHEV VehicleDriveType = "PHEV"
	BEV  VehicleDriveType = "BEV"
)

func (v VehicleDriveType) String() string {
	return string(v)
}

// IntegrationsMetadata represents json stored in integrations table metadata jsonb column
type IntegrationsMetadata struct {
	AutoPiDefaultTemplateID int `json:"auto_pi_default_template_id"`
}

// UserDeviceAPIIntegrationsMetadata represents json stored in user_device_api_integrations table metadata jsonb column
type UserDeviceAPIIntegrationsMetadata struct {
	AutoPiUnitID      *string                       `json:"auto_pi_unit_id,omitempty"`
	AutoPiIMEI        *string                       `json:"imei,omitempty"`
	AutoPiCommandJobs []UserDeviceAPIIntegrationJob `json:"auto_pi_command_jobs"`
}

// UserDeviceAPIIntegrationJob holds the autopi webhook jobs. We only expect a handful of this per device so not breaking out to own table for now
type UserDeviceAPIIntegrationJob struct {
	CommandJobID string    `json:"command_job_id"`
	CommandState string    `json:"command_state"`
	CommandRaw   string    `json:"command_raw"`
	LastUpdated  time.Time `json:"last_updated"`
}
