package services

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/redisq"
)

//go:generate mockgen -source autopi_task_service.go -destination mocks/autopi_task_service_mock.go
type AutoPiTaskService interface {
	StartAutoPiUpdate(ctx context.Context, deviceID, userID, unitID string) (taskID string, err error)
	GetTaskStatus(ctx context.Context, taskID string) (task *AutoPiTask, err error)
	StartConsumer(ctx context.Context)
}

// task names
const (
	updateAutoPiTask = "updateTask"
)

func NewAutoPiTaskService(settings *config.Settings, autoPiSvc AutoPiAPIService, logger zerolog.Logger) AutoPiTaskService {
	// setup redis connection
	var tlsConfig *tls.Config
	if settings.RedisTLS {
		tlsConfig = new(tls.Config)
	}
	var r StandardRedis
	// handle redis cluster in prod
	if settings.Environment == "prod" {
		r = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:     []string{settings.RedisURL},
			Password:  settings.RedisPassword,
			TLSConfig: tlsConfig,
		})
	} else {
		r = redis.NewClient(&redis.Options{
			Addr:      settings.RedisURL,
			Password:  settings.RedisPassword,
			TLSConfig: tlsConfig,
		})
	}

	var QueueFactory = redisq.NewFactory()
	const workerQueueName = "autopi-worker"
	// Create a queue.
	mainQueue := QueueFactory.RegisterQueue(&taskq.QueueOptions{
		Name:  workerQueueName,
		Redis: r, // go-redis client
	})
	// register task, handler would be below as
	ats := &autoPiTaskService{
		Settings:  settings,
		MainQueue: mainQueue,
		autoPiSvc: autoPiSvc,
		log:       logger.With().Str("worker queue", workerQueueName).Logger(),
	}
	updateTask := taskq.RegisterTask(&taskq.TaskOptions{
		Name: updateAutoPiTask,
		Handler: func(ctx context.Context, taskID, deviceID, userID, unitID string) error {
			return ats.ProcessUpdate(ctx, taskID, deviceID, userID, unitID)
		},
	})

	ats.UpdateAutoPiTask = updateTask
	ats.Redis = r
	return ats
}

type autoPiTaskService struct {
	Settings         *config.Settings
	MainQueue        taskq.Queue
	UpdateAutoPiTask *taskq.Task
	Redis            StandardRedis
	autoPiSvc        AutoPiAPIService
	log              zerolog.Logger
}

// StartAutoPiUpdate produces a task
func (ats *autoPiTaskService) StartAutoPiUpdate(ctx context.Context, deviceID, userID, unitID string) (taskID string, err error) {
	taskID = ksuid.New().String()
	msg := ats.UpdateAutoPiTask.WithArgs(ctx, taskID, deviceID, userID, unitID)
	msg.Name = taskID
	err = ats.MainQueue.Add(msg)
	if err != nil {
		return "", err
	}
	err = ats.updateTaskState(taskID, "", Pending, 100, nil)
	if err != nil {
		return taskID, err
	}

	return taskID, nil
}

func (ats *autoPiTaskService) StartConsumer(ctx context.Context) {
	if err := ats.MainQueue.Consumer().Start(ctx); err != nil {
		ats.log.Err(err).Msg("consumer failed")
	}
	ats.log.Info().Msg("started autopi tasks consumer")
}

// GetTaskStatus gets the status from the redis backend - is there a way to do this? multistep
func (ats *autoPiTaskService) GetTaskStatus(ctx context.Context, taskID string) (task *AutoPiTask, err error) {
	// problem is taskq does not have a way to retrieve a task, and we want to persist state as we move along the task
	taskRaw := ats.Redis.Get(ctx, buildAutoPiTaskRedisKey(taskID))
	if taskRaw == nil {
		return nil, errors.New("task not found")
	}
	taskBytes, err := taskRaw.Bytes()
	if err != nil {
		return nil, err
	}
	apTask := new(AutoPiTask)
	err = json.Unmarshal(taskBytes, apTask)
	if err != nil {
		return nil, err
	}
	return apTask, nil
}

// ProcessUpdate handles the task when consumed
func (ats *autoPiTaskService) ProcessUpdate(ctx context.Context, taskID, deviceID, userID, unitID string) error {
	// todo check if receive ctx cancel - means we should start task again on next restart

	log := ats.log.With().Str("handler", updateAutoPiTask+"_ProcessUpdate").Str("taskID", taskID).Str("deviceID", deviceID).Logger()
	// store the userID?
	log.Info().Msg("Started processing autopi update")

	err := ats.updateTaskState(taskID, "Started", InProcess, 110, nil)
	if err != nil {
		log.Err(err).Msg("failed to persist state to redis")
		return err
	}
	//send command to update device (2M timeout)
	cmd, err := ats.autoPiSvc.CommandRaw(ctx, deviceID, "minionutil.update_release", "")
	if err != nil {
		time.Sleep(time.Minute)
		cmd, err = ats.autoPiSvc.CommandRaw(ctx, deviceID, "minionutil.update_release", "")
		if err != nil {
			log.Err(err).Msg("failed to call autopi api svc with update command")
			_ = ats.updateTaskState(taskID, "autopi api call failed after retries", Failure, 500, err)
			return err
		}
	}
	//check that command did not timeout
	backoffSchedule := []time.Duration{
		10 * time.Second,
		30 * time.Second,
		30 * time.Second,
		60 * time.Second,
	}
	for _, backoff := range backoffSchedule {
		time.Sleep(backoff)
		job, _, err := ats.autoPiSvc.GetCommandStatus(ctx, cmd.Jid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_ = ats.updateTaskState(taskID, "autopi job was not found in db", Failure, 500, err)
				log.Err(err).Msg("autopi job not found in db")
				return err
			}
			continue // try again if error
		}
		if job.CommandState == "COMMAND_EXECUTED" {
			_ = ats.updateTaskState(taskID, "autopi update job received by device", InProcess, 120, nil)
			break
		}
		if job.CommandState == "TIMEOUT" {
			_ = ats.updateTaskState(taskID, "autopi update job timed out, check device online", Failure, 400, nil)
			log.Warn().Msg("autopi update job timed out and did not reach device")
			return errors.New("job timeout")
		}
	}
	// note we could fire off a delayed sub-task for the next part, but right now this feels easier to follow and debug
	//query device endpoint to check that update was applied. (5M timeout)
	backoffSchedule = []time.Duration{
		35 * time.Second,
		30 * time.Second,
		60 * time.Second,
		60 * time.Second,
		60 * time.Second,
		60 * time.Second,
	}
	for _, backoff := range backoffSchedule {
		time.Sleep(backoff)
		d, err := ats.autoPiSvc.GetDeviceByID(deviceID)
		if err != nil {
			continue // keep trying if error
		}
		if d.IsUpdated {
			_ = ats.updateTaskState(taskID, "autopi update confirmed", Success, 200, nil)
			log.Info().Msg("Succesfully applied autopi update")
			return nil
		}
	}
	// if we got here we could not validate update was applied
	_ = ats.updateTaskState(taskID, "could not confirm update applied within 5 minutes", Failure, 420, nil)

	return nil
}

func (ats *autoPiTaskService) updateTaskState(taskID, description string, status TaskStatusEnum, code int, err error) error {
	t := AutoPiTask{
		TaskID:      taskID,
		Status:      string(status),
		Description: description,
		Code:        code, // need enum with codes
	}
	if err != nil {
		t.Error = err.Error()
	}
	set := ats.Redis.Set(context.Background(), buildAutoPiTaskRedisKey(taskID), t, time.Hour*72)
	return set.Err()
}

func buildAutoPiTaskRedisKey(taskID string) string {
	return updateAutoPiTask + "_" + taskID
}

// AutoPiTask describes a task that is being worked on asynchronously for autopi
type AutoPiTask struct {
	TaskID      string `json:"taskId"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Code        int    `json:"code"`
	Error       string `json:"error"`
}

type TaskStatusEnum string

const (
	Pending   TaskStatusEnum = "Pending"
	InProcess TaskStatusEnum = "InProcess"
	Success   TaskStatusEnum = "Success"
	Failure   TaskStatusEnum = "Failure"
)

type StandardRedis interface {
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Pipelined(ctx context.Context, fn func(pipe redis.Pipeliner) error) ([]redis.Cmder, error)

	// Eval Required by redislock
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, scripts ...string) *redis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}
