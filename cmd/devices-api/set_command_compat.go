package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var teslaEnabledCommands = []string{"doors/lock", "doors/unlock", "trunk/open", "frunk/open", "charge/limit"}

func setCommandCompatibility(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {

	if err := setCommandCompatTesla(ctx, logger, settings, pdb); err != nil {
		return err
	}
	if err := setCommandCompatSmartcar(ctx, logger, settings, pdb); err != nil {
		return err
	}

	return nil
}

func setCommandCompatTesla(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	teslaInt, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(services.TeslaVendor)).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	teslaUDAIs, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(teslaInt.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	for _, tu := range teslaUDAIs {
		md := new(services.UserDeviceAPIIntegrationsMetadata)
		if err := tu.Metadata.Unmarshal(md); err != nil {
			return err
		}

		md.Commands = &services.UserDeviceAPIIntegrationsMetadataCommands{Enabled: teslaEnabledCommands}

		if err := tu.Metadata.Marshal(md); err != nil {
			return err
		}

		if _, err := tu.Update(ctx, pdb.DBS().Writer, boil.Whitelist("metadata")); err != nil {
			return err
		}
	}

	return nil
}

func setCommandCompatSmartcar(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	scInt, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(services.SmartCarVendor)).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	scUDAIs, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scInt.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice), // Need VIN and country.
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	for _, su := range scUDAIs {
		country := services.FindCountry(su.R.UserDevice.CountryCode.String)
		doors, err := checkSmartcarDoorCompatibility(settings, su.R.UserDevice.VinIdentifier.String, country.Alpha2)
		if err != nil {
			return err
		}
		if !doors {
			continue
		}
		md := new(services.UserDeviceAPIIntegrationsMetadata)
		if err := su.Metadata.Unmarshal(md); err != nil {
			return err
		}

		if md.Commands == nil {
			md.Commands = new(services.UserDeviceAPIIntegrationsMetadataCommands)
		}

		if len(md.Commands.Enabled) != 0 {
			continue
		}

		md.Commands.Capable = []string{"doors/lock", "doors/unlock"}

		if err := su.Metadata.Marshal(md); err != nil {
			return err
		}

		if _, err := su.Update(ctx, pdb.DBS().Writer, boil.Whitelist("metadata")); err != nil {
			return err
		}
	}

	return nil
}

type capResp struct {
	Capabilities []struct {
		Permission string `json:"permission"`
		Capable    bool   `json:"capable"`
	} `json:"capabilities"`
}

func checkSmartcarDoorCompatibility(settings *config.Settings, vin, countryAlpha2 string) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.smartcar.com/v2.0/compatibility?vin=%s&scope=control_security&country=%s", vin, countryAlpha2), nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(settings.SmartcarClientID+":"+settings.SmartcarClientSecret)))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status code %d", resp.StatusCode)
	}

	rb := new(capResp)
	if err := json.NewDecoder(resp.Body).Decode(rb); err != nil {
		return false, err
	}

	if len(rb.Capabilities) == 0 {
		return false, errors.New("no capabilities in response")
	}

	return rb.Capabilities[0].Capable, nil
}
