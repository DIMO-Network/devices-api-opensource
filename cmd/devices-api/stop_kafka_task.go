package main

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

func stopKafkaTask(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, scTaskSvc services.SmartcarTaskService, taskID string) error {
	integ := &models.UserDeviceAPIIntegration{
		TaskID:        null.StringFrom(taskID),
		UserDeviceID:  "FAKE",
		IntegrationID: "FAKE",
	}

	if err := scTaskSvc.StopPoll(integ); err != nil {
		return err
	}

	return nil
}
