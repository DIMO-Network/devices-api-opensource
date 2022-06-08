package controllers

import (
	"database/sql"
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// GetUserDeviceStatus godoc
// @Description  Returns the latest status update for the device. May return 404 if the
// @Description  user does not have a device with the ID, or if no status updates have come
// @Tags         user-devices
// @Produce      json
// @Param        user_device_id  path  string  true  "user device ID"
// @Success      200 {object} controllers.DeviceSnapshot
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/status [get]
func (udc *UserDevicesController) GetUserDeviceStatus(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
	).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}
	deviceData, err := models.UserDeviceData(models.UserDeviceDatumWhere.UserDeviceID.EQ(userDevice.ID),
		qm.OrderBy("updated_at asc")).All(c.Context(), udc.DBS().Reader)
	if errors.Is(err, sql.ErrNoRows) || len(deviceData) == 0 || !deviceData[0].Data.Valid {
		return fiber.NewError(fiber.StatusNotFound, "no status updates yet")
	}
	if err != nil {
		return err
	}
	// how should we handle the errorData, if at all?
	ds := DeviceSnapshot{}
	// merging data: foreach order by updatedAt desc, only set property if it exists in json data
	for _, datum := range deviceData {
		if datum.Data.Valid {
			// note this means the date updated we sent may be inaccurate if have both smartcar and autopi
			if ds.RecordUpdatedAt == nil {
				ds.RecordCreatedAt = &datum.CreatedAt
				ds.RecordUpdatedAt = &datum.UpdatedAt
			}
			// note we are assuming json property names are same accross smartcar, tesla, autopi, AND same types eg. int / float / string
			// we could use reflection and just have single line assuming json name in struct matches what is in data
			charging := gjson.GetBytes(datum.Data.JSON, "charging")
			if charging.Exists() {
				c := charging.Bool()
				ds.Charging = &c
			}
			fuelPercentRemaining := gjson.GetBytes(datum.Data.JSON, "fuelPercentRemaining")
			if fuelPercentRemaining.Exists() {
				f := fuelPercentRemaining.Float()
				ds.FuelPercentRemaining = &f
			}
			batteryCapacity := gjson.GetBytes(datum.Data.JSON, "batteryCapacity")
			if batteryCapacity.Exists() {
				b := batteryCapacity.Int()
				ds.BatteryCapacity = &b
			}
			oilLevel := gjson.GetBytes(datum.Data.JSON, "oil")
			if oilLevel.Exists() {
				o := oilLevel.Float()
				ds.OilLevel = &o
			}
			stateOfCharge := gjson.GetBytes(datum.Data.JSON, "soc")
			if stateOfCharge.Exists() {
				o := stateOfCharge.Float()
				ds.StateOfCharge = &o
			}
			odometer := gjson.GetBytes(datum.Data.JSON, "odometer")
			if odometer.Exists() {
				o := odometer.Float()
				ds.Odometer = &o
			}
			latitude := gjson.GetBytes(datum.Data.JSON, "latitude")
			if latitude.Exists() {
				l := latitude.Float()
				ds.Latitude = &l
			}
			longitude := gjson.GetBytes(datum.Data.JSON, "longitude")
			if longitude.Exists() {
				l := longitude.Float()
				ds.Longitude = &l
			}
			rangeG := gjson.GetBytes(datum.Data.JSON, "range")
			if rangeG.Exists() {
				r := rangeG.Float()
				ds.Range = &r
			}
			// TirePressure
			tires := gjson.GetBytes(datum.Data.JSON, "tires")
			if tires.Exists() {
				// weird thing here is in example payloads these are all ints, but the smartcar lib has as floats
				ds.TirePressure = &smartcar.TirePressure{
					FrontLeft:  tires.Get("frontLeft").Float(),
					FrontRight: tires.Get("frontRight").Float(),
					BackLeft:   tires.Get("backLeft").Float(),
					BackRight:  tires.Get("backRight").Float(),
				}
			}
		}
	}
	return c.JSON(ds)
}

// RefreshUserDeviceStatus godoc
// @Description  Starts the process of refreshing device status from Smartcar
// @Tags         user-devices
// @Param        user_device_id  path  string  true  "user device ID"
// @Success      204
// @Failure      429  "rate limit hit for integration"
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/commands/refresh [post]
func (udc *UserDevicesController) RefreshUserDeviceStatus(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	// We could probably do a smarter join here, but it's unclear to me how to handle that
	// in SQLBoiler.
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceData),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceData, models.UserDeviceDatumRels.Integration)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}

	for _, deviceDatum := range ud.R.UserDeviceData {
		if deviceDatum.R.Integration.Type == models.IntegrationTypeAPI && deviceDatum.R.Integration.Vendor == services.SmartCarVendor {

			nextAvailableTime := deviceDatum.UpdatedAt.Add(time.Second * time.Duration(deviceDatum.R.Integration.RefreshLimitSecs))
			if time.Now().Before(nextAvailableTime) {
				return fiber.NewError(fiber.StatusTooManyRequests, "rate limit for integration refresh hit")
			}

			udai, err := models.FindUserDeviceAPIIntegration(c.Context(), udc.DBS().Reader, deviceDatum.UserDeviceID, deviceDatum.IntegrationID)
			if err != nil {
				return err
			}
			if udai.TaskID.Valid {
				err = udc.smartcarTaskSvc.Refresh(udai)
				if err != nil {
					return err
				}
			} else {
				err = udc.taskSvc.StartSmartcarRefresh(udi, deviceDatum.IntegrationID)
				if err != nil {
					return err
				}
			}
			return c.SendStatus(204)
		}
	}
	return fiber.NewError(fiber.StatusBadRequest, "no active Smartcar integration found for this device")
}

// DeviceSnapshot is the response object for device status endpoint
// https://docs.google.com/document/d/1DYzzTOR9WA6WJNoBnwpKOoxfmrVwPWNLv0x0MkjIAqY/edit#heading=h.dnp7xngl47bw
type DeviceSnapshot struct {
	Charging             *bool                  `json:"charging,omitempty"`
	FuelPercentRemaining *float64               `json:"fuelPercentRemaining,omitempty"`
	BatteryCapacity      *int64                 `json:"batteryCapacity,omitempty"`
	OilLevel             *float64               `json:"oil,omitempty"`
	Odometer             *float64               `json:"odometer,omitempty"`
	Latitude             *float64               `json:"latitude,omitempty"`
	Longitude            *float64               `json:"longitude,omitempty"`
	Range                *float64               `json:"range,omitempty"`
	StateOfCharge        *float64               `json:"soc,omitempty"` // todo: change json to match after update frontend
	RecordUpdatedAt      *time.Time             `json:"recordUpdatedAt,omitempty"`
	RecordCreatedAt      *time.Time             `json:"recordCreatedAt,omitempty"`
	TirePressure         *smartcar.TirePressure `json:"tirePressure,omitempty"`
}
