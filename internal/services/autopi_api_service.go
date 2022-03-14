package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/pkg/errors"
)

const baseAPIURL = "https://api.dimo.autopi.io"

//go:generate mockgen -source autopi_api_service.go -destination mocks/autopi_api_service_mock.go
type AutoPiAPIService interface {
	GetDeviceByUnitID(unitID string) (*AutoPiDongleDevice, error)
	GetDeviceByID(deviceID string) (*AutoPiDongleDevice, error)
	PatchVehicleProfile(vehicleID string, profile PatchVehicleProfile) error
	UnassociateDeviceTemplate(deviceID, templateID string) error
	AssociateDeviceToTemplate(deviceID, templateID string) error
	ApplyTemplate(deviceID, templateID string) error
	CommandSyncDevice(deviceID string) error
}

type autoPiAPIService struct {
	Settings   *config.Settings
	HTTPClient *http.Client
}

func NewAutoPiAPIService(settings *config.Settings) AutoPiAPIService {
	return &autoPiAPIService{
		Settings: settings,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDeviceByUnitID calls /dongle/devices/by_unit_id/{unit_id}/ to get the device for the unitID.
// Errors if it finds none or more than one device, as there should only be one device attached to a unit.
func (a *autoPiAPIService) GetDeviceByUnitID(unitID string) (*AutoPiDongleDevice, error) {
	res, err := a.executeRequest(fmt.Sprintf("/dongle/devices/by_unit_id/%s/", unitID), "GET", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling autopi api to get unit with ID %s", unitID)
	}
	defer res.Body.Close() // nolint

	u := new(autoPiUnits)
	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding json from autopi api to get device by unitID %s", unitID)
	}
	if u.Count != 1 {
		return nil, fmt.Errorf("expected to find only one device with autopi unitID %s", unitID)
	}
	return &u.Results[0], nil
}

// GetDeviceByID calls https://api.dimo.autopi.io/dongle/devices/{DEVICE_ID}/ Note that the deviceID is the autoPi one. This brings us the templateID
func (a *autoPiAPIService) GetDeviceByID(deviceID string) (*AutoPiDongleDevice, error) {
	res, err := a.executeRequest(fmt.Sprintf("/dongle/devices/%s/", deviceID), "GET", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling autopi api to get device %s", deviceID)
	}
	defer res.Body.Close() // nolint

	d := new(AutoPiDongleDevice)
	err = json.NewDecoder(res.Body).Decode(d)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding json from autopi api to get device %s", deviceID)
	}
	return d, nil
}

// PatchVehicleProfile https://api.dimo.autopi.io/vehicle/profile/{device.vehicle.id}/ driveType: {"ICE", "BEV", "PHEV", "HEV"}
func (a *autoPiAPIService) PatchVehicleProfile(vehicleID string, profile PatchVehicleProfile) error {
	j, _ := json.Marshal(profile)
	res, err := a.executeRequest(fmt.Sprintf("/vehicle/profile/%s/", vehicleID), "PATCH", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to patch device %s", vehicleID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// UnassociateDeviceTemplate Unassociate the device from the existing templateID.
func (a *autoPiAPIService) UnassociateDeviceTemplate(deviceID, templateID string) error {
	p := postDeviceIDs{
		Devices:         []string{deviceID},
		UnassociateOnly: false,
	}
	j, _ := json.Marshal(p)
	res, err := a.executeRequest(fmt.Sprintf("/dongle/templates/%s/unassociate_devices/", templateID), "POST", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to unassociate_devices. template %s", templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// AssociateDeviceToTemplate set a new templateID on the device by doing a Patch request
func (a *autoPiAPIService) AssociateDeviceToTemplate(deviceID, templateID string) error {
	p := postDeviceIDs{
		Devices: []string{deviceID},
	}
	j, _ := json.Marshal(p)
	res, err := a.executeRequest(fmt.Sprintf("/dongle/templates/%s/", templateID), "PATCH", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to associate device %s with new template %s", deviceID, templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// ApplyTemplate When device awakes, it checks if it has templates to be applied. If device is awake, this won't do anything until next cycle.
func (a *autoPiAPIService) ApplyTemplate(deviceID, templateID string) error {
	p := postDeviceIDs{
		Devices: []string{deviceID},
	}
	j, _ := json.Marshal(p)
	res, err := a.executeRequest(fmt.Sprintf("/dongle/templates/%s/apply_explicit", templateID), "POST", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to apply template for device %s with new template %s", deviceID, templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// CommandSyncDevice sends raw command to autopi only if it is online. Invokes syncing the pending changes (eg. template change) on the device.
func (a *autoPiAPIService) CommandSyncDevice(deviceID string) error {
	p := `{"command":"state.sls pending"}`
	res, err := a.executeRequest(fmt.Sprintf("/dongle/%s/execute_raw", deviceID), "POST", []byte(p))
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to command state.sls.pending for device %s", deviceID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// executeRequest calls an api endpoint with autopi creds, optional body and error handling.
// If request results in non 2xx response, will always return error with payload body in err message
// respone should have defer response.Body.Close() after the error check as it could be nil when err is != nil
func (a *autoPiAPIService) executeRequest(path, method string, body []byte) (*http.Response, error) {
	reader := new(bytes.Reader)
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, baseAPIURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", a.Settings.AutoPiAPIToken)
	res, err := a.HTTPClient.Do(req)
	// handle error status codes
	if err == nil && res != nil && res.StatusCode > 299 {
		defer res.Body.Close() //nolint
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "error reading failed request body")
		}
		return nil, errors.Errorf("received non success status code %d with body: %s", res.StatusCode, string(body))
	}
	return res, err
}

// AutoPiDongleDevice https://api.dimo.autopi.io/#/dongle/dongle_devices_read
type AutoPiDongleDevice struct {
	ID       string `json:"id"`
	UnitID   string `json:"unit_id"`
	Token    string `json:"token"`
	CallName string `json:"callName"`
	Owner    int    `json:"owner"`
	Vehicle  struct {
		ID                    string `json:"id"`
		Vin                   string `json:"vin"`
		Display               string `json:"display"`
		CallName              string `json:"callName"`
		LicensePlate          string `json:"licensePlate"`
		Model                 int    `json:"model"`
		Make                  string `json:"make"`
		Year                  int    `json:"year"`
		Type                  string `json:"type"`
		BatteryNominalVoltage int    `json:"battery_nominal_voltage"`
	} `json:"vehicle"`
	Display           string    `json:"display"`
	LastCommunication time.Time `json:"last_communication"`
	IsUpdated         string    `json:"is_updated"`
	Release           struct {
		Version string `json:"version"`
	} `json:"release"`
	OpenAlerts         string `json:"open_alerts"`
	IMEI               string `json:"imei"`
	Template           int    `json:"template"`
	Warnings           string `json:"warnings"`
	KeyState           string `json:"key_state"`
	Access             string `json:"access"`
	DockerReleases     string `json:"docker_releases"`
	DataUsage          string `json:"data_usage"`
	PhoneNumber        string `json:"phone_number"`
	Icc                string `json:"icc"`
	MaxDataUsage       string `json:"max_data_usage"`
	IsBlockedByRelease string `json:"is_blocked_by_release"`
	// only exists when get by unitID
	HwRevision string   `json:"hw_revision"`
	Tags       []string `json:"tags"`
}

// PatchVehicleProfile used to update vehicle profile https://api.dimo.autopi.io/#/vehicle/vehicle_profile_partial_update
type PatchVehicleProfile struct {
	Vin      string `json:"vin,omitempty"`
	CallName string `json:"callName,omitempty"`
	Year     int    `json:"year,omitempty"`
	Type     string `json:"type,omitempty"`
}

// used to post an array of device ID's, for template and command operations
type postDeviceIDs struct {
	Devices         []string `json:"devices"`
	UnassociateOnly bool     `json:"unassociate_only,omitempty"`
}

// autoPiUnits used when get devices by unitID, basically just a result wrapper
type autoPiUnits struct {
	Count    int                  `json:"count"`
	Next     string               `json:"next"`
	Previous string               `json:"previous"`
	Results  []AutoPiDongleDevice `json:"results"`
}
