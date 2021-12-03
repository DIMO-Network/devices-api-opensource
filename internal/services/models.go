package services

// DeviceVehicleInfo represents some standard vehicle specific properties stored in the metadata json field in DB
type DeviceVehicleInfo struct {
	FuelType      string `json:"fuel_type,omitempty"`
	DrivenWheels  string `json:"driven_wheels,omitempty"`
	NumberOfDoors string `json:"number_of_doors,omitempty"`
	BaseMSRP      int    `json:"base_msrp,omitempty"`
	EPAClass      string `json:"epa_class,omitempty"`
	// PASSENGER CAR, from NHTSA
	VehicleType   string `json:"vehicle_type,omitempty"`
	MPGHighway    string `json:"mpg_highway,omitempty"`
	MPGCity       string `json:"mpg_city,omitempty"`
}
