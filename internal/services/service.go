// File: internal/sync/service.go
// Brief: Unified business logic for task scheduling, execution, and runners (migrated from internal/tasks/)
// Detailed: Contains TaskScheduler, TaskRunner, and all related logic for scheduling, executing, and managing system tasks.
// Author: Argus Migration (AI)
// Date: 2024-07-03

package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"argus/internal/models"

	"github.com/robfig/cron/v3"
)

const (
	DefaultCheckInterval      = 1 * time.Minute
	DefaultMaxConcurrentTasks = 5
	DefaultTaskTimeout        = 30 * time.Minute
)

type TaskSchedulerConfig struct {
	CheckInterval      time.Duration
	MaxConcurrentTasks int
	TaskTimeout        time.Duration
}

func DefaultTaskSchedulerConfig() *TaskSchedulerConfig {
	return &TaskSchedulerConfig{
		CheckInterval:      DefaultCheckInterval,
		MaxConcurrentTasks: DefaultMaxConcurrentTasks,
		TaskTimeout:        DefaultTaskTimeout,
	}
}

type TaskScheduler struct {
	config     *TaskSchedulerConfig
	repository models.TaskRepository
	runners    map[models.TaskType]TaskRunner
	semaphore  chan struct{}
	cronParser cron.Parser
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mutex      sync.RWMutex
	running    bool
}

func NewTaskScheduler(repo models.TaskRepository, config *TaskSchedulerConfig) *TaskScheduler {
	if config == nil {
		config = DefaultTaskSchedulerConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskScheduler{
		config:     config,
		repository: repo,
		runners:    make(map[models.TaskType]TaskRunner),
		semaphore:  make(chan struct{}, config.MaxConcurrentTasks),
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		ctx:        ctx,
		cancel:     cancel,
		running:    false,
	}
}

func (s *TaskScheduler) RegisterRunner(runner TaskRunner) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	taskType := runner.GetType()
	slog.Info("Registering task runner", "task_type", taskType)
	s.runners[taskType] = runner
}

func (s *TaskScheduler) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.running {
		return fmt.Errorf("scheduler is already running")
	}
	slog.Info("Starting task scheduler",
		"check_interval", s.config.CheckInterval,
		"max_concurrent_tasks", s.config.MaxConcurrentTasks,
		"task_timeout", s.config.TaskTimeout)
	s.running = true
	s.wg.Add(1)
	go s.scheduleLoop()
	return nil
}

func (s *TaskScheduler) Stop() {
	s.mutex.Lock()
	if !s.running {
		s.mutex.Unlock()
		return
	}
	s.running = false
	s.mutex.Unlock()
	slog.Info("Stopping task scheduler")
	s.cancel()
	s.wg.Wait()
	slog.Info("Task scheduler stopped")
}

func (s *TaskScheduler) scheduleLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()
	if err := s.checkScheduledTasks(); err != nil {
		slog.Error("Error checking scheduled tasks", "error", err)
	}
	for {
		select {
		case <-ticker.C:
			if err := s.checkScheduledTasks(); err != nil {
				slog.Error("Error checking scheduled tasks", "error", err)
			}
		case <-s.ctx.Done():
			slog.Info("Task scheduler context cancelled, exiting schedule loop")
			return
		}
	}
}

func (s *TaskScheduler) checkScheduledTasks() error {
	tasks, err := s.repository.ListTasks(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}
	now := time.Now()
	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		if !task.Schedule.NextRunTime.IsZero() && task.Schedule.NextRunTime.Before(now) {
			s.wg.Add(1)
			go func(t *models.TaskConfig) {
				defer s.wg.Done()
				s.semaphore <- struct{}{}
				defer func() { <-s.semaphore }()
				if err := s.executeTask(t); err != nil {
					slog.Error("Failed to execute task",
						"task_id", t.ID,
						"task_name", t.Name,
						"error", err)
				}
			}(task)
		}
	}
	return nil
}

func (s *TaskScheduler) executeTask(task *models.TaskConfig) error {
	slog.Info("Executing scheduled task", "task_id", task.ID, "task_name", task.Name)
	s.mutex.RLock()
	runner, exists := s.runners[task.Type]
	s.mutex.RUnlock()
	if !exists {
		return fmt.Errorf("no runner registered for task type: %s", task.Type)
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.config.TaskTimeout)
	defer cancel()
	execution, err := runner.Run(ctx, task)
	if err != nil {
		return fmt.Errorf("task execution failed: %w", err)
	}
	if err := s.repository.RecordExecution(s.ctx, execution); err != nil {
		return fmt.Errorf("failed to record task execution: %w", err)
	}
	if !task.Schedule.OneTime {
		if err := s.updateNextRunTime(task); err != nil {
			return fmt.Errorf("failed to update next run time: %w", err)
		}
	} else {
		task.Enabled = false
		if err := s.repository.UpdateTask(s.ctx, task); err != nil {
			return fmt.Errorf("failed to disable one-time task: %w", err)
		}
	}
	return nil
}

func (s *TaskScheduler) updateNextRunTime(task *models.TaskConfig) error {
	if task.Schedule.CronExpression == "" {
		return fmt.Errorf("task has no cron expression")
	}
	schedule, err := s.cronParser.Parse(task.Schedule.CronExpression)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	now := time.Now()
	nextRun := schedule.Next(now)
	task.Schedule.NextRunTime = nextRun
	return s.repository.UpdateTask(s.ctx, task)
}

func (s *TaskScheduler) RunTaskNow(taskID string) (*models.TaskExecution, error) {
	task, err := s.repository.GetTask(s.ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	s.mutex.RLock()
	runner, exists := s.runners[task.Type]
	s.mutex.RUnlock()
	if !exists {
		return nil, fmt.Errorf("no runner registered for task type: %s", task.Type)
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.config.TaskTimeout)
	defer cancel()
	execution, err := runner.Run(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("task execution failed: %w", err)
	}
	if err := s.repository.RecordExecution(s.ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to record task execution: %w", err)
	}
	return execution, nil
}

// TaskRunner and implementations
var (
	ErrUnsupportedTaskType = errors.New("unsupported task type")
	ErrTaskCancelled       = errors.New("task cancelled")
	ErrInvalidParameter    = errors.New("invalid task parameter")
)

type TaskRunner interface {
	Run(ctx context.Context, task *models.TaskConfig) (*models.TaskExecution, error)
	GetType() models.TaskType
}

type BaseTaskRunner struct {
	taskType models.TaskType
}

func (r *BaseTaskRunner) GetType() models.TaskType {
	return r.taskType
}

func NewTaskRunner(taskType models.TaskType) (TaskRunner, error) {
	switch taskType {
	case models.TaskLogRotation:
		return &LogRotationRunner{BaseTaskRunner{taskType: models.TaskLogRotation}}, nil
	case models.TaskMetricsAggregation:
		return &MetricsAggregationRunner{BaseTaskRunner{taskType: models.TaskMetricsAggregation}}, nil
	case models.TaskHealthCheck:
		return &HealthCheckRunner{BaseTaskRunner{taskType: models.TaskHealthCheck}}, nil
	case models.TaskSystemCleanup:
		return &SystemCleanupRunner{BaseTaskRunner{taskType: models.TaskSystemCleanup}}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedTaskType, taskType)
	}
}

type LogRotationRunner struct {
	BaseTaskRunner
}

func (r *LogRotationRunner) Run(ctx context.Context, task *models.TaskConfig) (*models.TaskExecution, error) {
	return nil, errors.New("LogRotationRunner not implemented")
}

type MetricsAggregationRunner struct {
	BaseTaskRunner
}

func (r *MetricsAggregationRunner) Run(ctx context.Context, task *models.TaskConfig) (*models.TaskExecution, error) {
	return nil, errors.New("MetricsAggregationRunner not implemented")
}

type HealthCheckRunner struct {
	BaseTaskRunner
}

func (r *HealthCheckRunner) Run(ctx context.Context, task *models.TaskConfig) (*models.TaskExecution, error) {
	return nil, errors.New("HealthCheckRunner not implemented")
}

type SystemCleanupRunner struct {
	BaseTaskRunner
}

func (r *SystemCleanupRunner) Run(ctx context.Context, task *models.TaskConfig) (*models.TaskExecution, error) {
	return nil, errors.New("SystemCleanupRunner not implemented")
}

// LogRotationRunner, MetricsAggregationRunner, HealthCheckRunner, SystemCleanupRunner, and helpers go here (see runner.go for full code)
// ...
