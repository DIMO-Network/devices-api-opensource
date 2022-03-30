package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	autoPiBaseAPIURL  = "https://api.dimo.autopi.io"
	AutoPiVendor      = "AutoPi"
	AutoPiWebhookPath = "/webhooks/autopi-command"
)

//go:generate mockgen -source autopi_api_service.go -destination mocks/autopi_api_service_mock.go
type AutoPiAPIService interface {
	GetDeviceByUnitID(unitID string) (*AutoPiDongleDevice, error)
	GetDeviceByID(deviceID string) (*AutoPiDongleDevice, error)
	PatchVehicleProfile(vehicleID string, profile PatchVehicleProfile) error
	UnassociateDeviceTemplate(deviceID string, templateID int) error
	AssociateDeviceToTemplate(deviceID string, templateID int) error
	ApplyTemplate(deviceID string, templateID int) error
	CommandSyncDevice(deviceID string) (*AutoPiCommandResponse, error)
	CommandRaw(deviceID string, command string) (*AutoPiCommandResponse, error)
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

func GetOrCreateAutoPiIntegration(ctx context.Context, exec boil.ContextExecutor) (*models.Integration, error) {
	const (
		autoPiType        = "API"
		autoPiStyle       = models.IntegrationStyleAddon
		defaultTemplateID = 10
	)
	integration, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(AutoPiVendor),
		models.IntegrationWhere.Style.EQ(autoPiStyle), models.IntegrationWhere.Type.EQ(autoPiType)).
		One(ctx, exec)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// create
			im := IntegrationsMetadata{AutoPiDefaultTemplateID: defaultTemplateID}
			integration = &models.Integration{
				ID:     ksuid.New().String(),
				Vendor: AutoPiVendor,
				Type:   autoPiType,
				Style:  autoPiStyle,
			}
			_ = integration.Metadata.Marshal(im)
			err = integration.Insert(ctx, exec, boil.Infer())
			if err != nil {
				return nil, errors.Wrap(err, "error inserting autoPi integration")
			}
		} else {
			return nil, errors.Wrap(err, "error fetching autoPi integration from database")
		}
	}
	return integration, nil
}

// FindUserDeviceAutoPiIntegration gets the user_device_api_integration record and unmarshalled metadata, returns fiber error where makes sense
func FindUserDeviceAutoPiIntegration(ctx context.Context, exec boil.ContextExecutor, userDeviceID, userID string) (*models.UserDeviceAPIIntegration, *UserDeviceAPIIntegrationsMetadata, error) {
	autoPiInteg, err := GetOrCreateAutoPiIntegration(ctx, exec)
	if err != nil {
		return nil, nil, err
	}
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("could not find device with id %s for user %s", userDeviceID, userID))
		}
		return nil, nil, errors.Wrap(err, "Unexpected database error searching for user device")
	}
	udai := new(models.UserDeviceAPIIntegration)
	for _, apiInteg := range ud.R.UserDeviceAPIIntegrations {
		if apiInteg.IntegrationID == autoPiInteg.ID {
			udai = apiInteg
		}
	}
	if !(udai != nil && udai.ExternalID.Valid) {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "user does not have an autopi integration registered for userDeviceId: "+userDeviceID)
	}
	// get metadata for a little later
	md := new(UserDeviceAPIIntegrationsMetadata)
	err = udai.Metadata.Unmarshal(md)
	if err != nil {
		return nil, nil, errors.Wrap(err, "metadata for user device api integrations in wrong format for unmarshal")
	}
	return udai, md, nil
}

// AppendAutoPiCompatibility adds autopi compatibility for AmericasRegion and EuropeRegion regions
func AppendAutoPiCompatibility(ctx context.Context, dcs []DeviceCompatibility, writer *database.DB) ([]DeviceCompatibility, error) {
	integration, err := GetOrCreateAutoPiIntegration(ctx, writer)
	if err != nil {
		return nil, err
	}
	dcs = append(dcs, DeviceCompatibility{
		ID:           integration.ID,
		Type:         integration.Type,
		Style:        integration.Style,
		Vendor:       integration.Vendor,
		Region:       AmericasRegion.String(),
		Capabilities: nil,
	})
	dcs = append(dcs, DeviceCompatibility{
		ID:           integration.ID,
		Type:         integration.Type,
		Style:        integration.Style,
		Vendor:       integration.Vendor,
		Region:       EuropeRegion.String(),
		Capabilities: nil,
	})
	return dcs, nil
}

// GetDeviceByUnitID calls /dongle/devices/by_unit_id/{unit_id}/ to get the device for the unitID.
// Errors if it finds none or more than one device, as there should only be one device attached to a unit.
func (a *autoPiAPIService) GetDeviceByUnitID(unitID string) (*AutoPiDongleDevice, error) {
	res, err := a.executeRequest(fmt.Sprintf("/dongle/devices/by_unit_id/%s/", unitID), "GET", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling autopi api to get unit with ID %s", unitID)
	}
	defer res.Body.Close() // nolint

	u := new(AutoPiDongleDevice)
	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding json from autopi api to get device by unitID %s", unitID)
	}

	return u, nil
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
func (a *autoPiAPIService) UnassociateDeviceTemplate(deviceID string, templateID int) error {
	p := postDeviceIDs{
		Devices:         []string{deviceID},
		UnassociateOnly: false,
	}
	j, _ := json.Marshal(p)
	res, err := a.executeRequest(fmt.Sprintf("/dongle/templates/%d/unassociate_devices/", templateID), "POST", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to unassociate_devices. template %d", templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// AssociateDeviceToTemplate set a new templateID on the device by doing a Patch request
func (a *autoPiAPIService) AssociateDeviceToTemplate(deviceID string, templateID int) error {
	p := postDeviceIDs{
		Devices: []string{deviceID},
	}
	j, _ := json.Marshal(p)
	res, err := a.executeRequest(fmt.Sprintf("/dongle/templates/%d/", templateID), "PATCH", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to associate device %s with new template %d", deviceID, templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// ApplyTemplate When device awakes, it checks if it has templates to be applied. If device is awake, this won't do anything until next cycle.
func (a *autoPiAPIService) ApplyTemplate(deviceID string, templateID int) error {
	p := postDeviceIDs{
		Devices: []string{deviceID},
	}
	j, _ := json.Marshal(p)
	res, err := a.executeRequest(fmt.Sprintf("/dongle/templates/%d/apply_explicit", templateID), "POST", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to apply template for device %s with new template %d", deviceID, templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// CommandSyncDevice sends raw command to autopi only if it is online. Invokes syncing the pending changes (eg. template change) on the device.
func (a *autoPiAPIService) CommandSyncDevice(deviceID string) (*AutoPiCommandResponse, error) {
	webhookURL := fmt.Sprintf("%s/api/v1%s", a.Settings.DeploymentBaseURL, AutoPiWebhookPath)
	syncCommand := autoPiCommandRequest{
		Command:     "state.sls pending",
		CallbackURL: &webhookURL,
	}
	j, err := json.Marshal(syncCommand)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshall json for autoPiCommandRequest")
	}

	res, err := a.executeRequest(fmt.Sprintf("/dongle/%s/execute_raw", deviceID), "POST", j)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling autopi api to command state.sls.pending for device %s", deviceID)
	}
	defer res.Body.Close() // nolint

	d := new(AutoPiCommandResponse)
	err = json.NewDecoder(res.Body).Decode(d)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to decode responde from autopi command")
	}

	return d, nil
}

// CommandRaw sends raw command to autopi only if it is online, pending whitelisted
func (a *autoPiAPIService) CommandRaw(deviceID string, command string) (*AutoPiCommandResponse, error) {
	// todo: whitelist command
	webhookURL := fmt.Sprintf("%s/api/v1%s", a.Settings.DeploymentBaseURL, AutoPiWebhookPath)
	syncCommand := autoPiCommandRequest{
		Command:     command,
		CallbackURL: &webhookURL,
	}
	j, err := json.Marshal(syncCommand)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshall json for autoPiCommandRequest")
	}

	res, err := a.executeRequest(fmt.Sprintf("/dongle/%s/execute_raw", deviceID), "POST", j)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling autopi api to command %s for device %s", command, deviceID)
	}
	defer res.Body.Close() // nolint

	d := new(AutoPiCommandResponse)
	err = json.NewDecoder(res.Body).Decode(d)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to decode responde from autopi command")
	}

	return d, nil
}

// executeRequest calls an api endpoint with autopi creds, optional body and error handling.
// If request results in non 2xx response, will always return error with payload body in err message
// respone should have defer response.Body.Close() after the error check as it could be nil when err is != nil
func (a *autoPiAPIService) executeRequest(path, method string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, autoPiBaseAPIURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "APIToken "+a.Settings.AutoPiAPIToken)
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
	ID                string              `json:"id"`
	UnitID            string              `json:"unit_id"`
	Token             string              `json:"token"`
	CallName          string              `json:"callName"`
	Owner             int                 `json:"owner"`
	Vehicle           AutoPiDongleVehicle `json:"vehicle"`
	Display           string              `json:"display"`
	LastCommunication time.Time           `json:"last_communication"`
	IsUpdated         bool                `json:"is_updated"`
	Release           struct {
		Version string `json:"version"`
	} `json:"release"`
	OpenAlerts struct {
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Critical int `json:"critical"`
		Low      int `json:"low"`
	} `json:"open_alerts"`
	IMEI     string `json:"imei"`
	Template int    `json:"template"`
	Warnings []struct {
		DeviceHasNoMakeModel struct {
			Header  string `json:"header"`
			Message string `json:"message"`
		} `json:"device_has_no_make_model"`
	} `json:"warnings"`
	KeyState           string   `json:"key_state"`
	Access             string   `json:"access"`
	DockerReleases     []string `json:"docker_releases"`
	DataUsage          int      `json:"data_usage"`
	PhoneNumber        string   `json:"phone_number"`
	Icc                string   `json:"icc"`
	MaxDataUsage       int      `json:"max_data_usage"`
	IsBlockedByRelease bool     `json:"is_blocked_by_release"`
	// only exists when get by unitID
	HwRevision string   `json:"hw_revision"`
	Tags       []string `json:"tags"`
}

type AutoPiDongleVehicle struct {
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

type autoPiCommandRequest struct {
	Command     string  `json:"command"`
	CallbackURL *string `json:"callback_url,omitempty"`
	// CallbackTimeout default is 120 seconds
	CallbackTimeout *int `json:"callback_timeout,omitempty"`
}

type AutoPiCommandResponse struct {
	Jid     string   `json:"jid"`
	Minions []string `json:"minions"`
}
