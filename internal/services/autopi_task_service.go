package services

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source autopi_task_service.go -destination mocks/autopi_task_service_mock.go
type AutoPiTaskService interface {
	StartAutoPiUpdate(ctx context.Context, deviceID, userID, unitID string) (taskID string, err error)
	GetTaskStatus(ctx context.Context, taskID string) (task AutoPiTask, err error)
}

func NewAutoPiTaskService(settings *config.Settings) AutoPiTaskService {
	return &autoPiTaskService{
		Settings: settings,
	}
}

type autoPiTaskService struct {
	Settings *config.Settings
}

func (ats *autoPiTaskService) StartAutoPiUpdate(ctx context.Context, deviceID, userID, unitID string) (taskID string, err error) {
	return ksuid.New().String(), nil
}

func (ats *autoPiTaskService) GetTaskStatus(ctx context.Context, taskID string) (task AutoPiTask, err error) {
	return AutoPiTask{
		TaskID:      taskID,
		Status:      "Pending",
		Description: "testing",
		Code:        100, // todo make enum of status, has code as value etc
	}, nil
}

// AutoPiTask describes a task that is being worked on asynchronously for autopi
type AutoPiTask struct {
	TaskID      string `json:"taskId"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Code        int    `json:"code"`
}
