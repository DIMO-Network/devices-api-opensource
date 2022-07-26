package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
)

// load user devices.
func loadUserDeviceDrively(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	// get all devices from DB.
	all, err := models.UserDevices(models.UserDeviceWhere.VinConfirmed.EQ(true)).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	logger.Info().Msgf("found %d user device verified", len(all))

	client := &http.Client{}

	for _, device := range all {
		url := fmt.Sprintf("%s/api/%s", baseURL, device.VinIdentifier.String)
		logger.Info().Msgf("URL: %s", url)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("x-api-key", "")
		res, err := client.Do(req)

		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			logger.Info().Msgf("Unexpected response %#v", res)
			return err
		}

		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Info().Msgf("Could not read response body: %s\n", err)
			return err
		}

		vinModel := DrivlyVINMetaData{}

		err = json.Unmarshal(resBody, &vinModel)
		if err != nil {
			logger.Info().Msgf("Could not parse response body %", err)
			return err
		}

		metadata := map[string]string{}
		metadata["mpg"] = vinModel.mpg

		fmt.Println("Id :", metadata)

	}

	return nil
}

type DrivlyVINMetaData struct {
	mpg string
}
