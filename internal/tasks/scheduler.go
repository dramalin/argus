// Package tasks provides functionality for scheduling and managing system maintenance tasks
package tasks

// All scheduling and business logic has been migrated to internal/sync/service.go as part of the architecture migration.
// This file remains as a stub for compatibility and to avoid breaking imports during transition.

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	// DefaultCheckInterval is the default interval between scheduler checks
	DefaultCheckInterval = 1 * time.Minute

	// DefaultMaxConcurrentTasks is the default maximum number of concurrent tasks
	DefaultMaxConcurrentTasks = 5

	// DefaultTaskTimeout is the default timeout for task execution
	DefaultTaskTimeout = 30 * time.Minute
)

// TaskSchedulerConfig holds configuration for the task scheduler
type TaskSchedulerConfig struct {
	CheckInterval      time.Duration // Interval between checking for tasks to run
	MaxConcurrentTasks int           // Maximum number of tasks to run concurrently
	TaskTimeout        time.Duration // Maximum time a task can run before being cancelled
}

// DefaultTaskSchedulerConfig returns a default configuration for the task scheduler
func DefaultTaskSchedulerConfig() *TaskSchedulerConfig {
	return &TaskSchedulerConfig{
		CheckInterval:      DefaultCheckInterval,
		MaxConcurrentTasks: DefaultMaxConcurrentTasks,
		TaskTimeout:        DefaultTaskTimeout,
	}
}

// TaskScheduler manages the execution of scheduled tasks
type TaskScheduler struct {
	config     *TaskSchedulerConfig
	repository TaskRepository
	runners    map[TaskType]TaskRunner
	semaphore  chan struct{}
	cronParser cron.Parser
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mutex      sync.RWMutex
	running    bool
}

// NewTaskScheduler creates a new task scheduler
func NewTaskScheduler(repo TaskRepository, config *TaskSchedulerConfig) *TaskScheduler {
	if config == nil {
		config = DefaultTaskSchedulerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TaskScheduler{
		config:     config,
		repository: repo,
		runners:    make(map[TaskType]TaskRunner),
		semaphore:  make(chan struct{}, config.MaxConcurrentTasks),
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		ctx:        ctx,
		cancel:     cancel,
		running:    false,
	}
}

// RegisterRunner registers a runner for a specific task type
func (s *TaskScheduler) RegisterRunner(runner TaskRunner) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	taskType := runner.GetType()
	slog.Info("Registering task runner", "task_type", taskType)
	s.runners[taskType] = runner
}

// Start begins the scheduling loop in a goroutine
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

// Stop stops the scheduler and waits for tasks to complete
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

// scheduleLoop runs the main scheduling loop
func (s *TaskScheduler) scheduleLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	// Initial check when starting
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

// checkScheduledTasks checks for tasks that are due and executes them
func (s *TaskScheduler) checkScheduledTasks() error {
	// Get all enabled tasks
	tasks, err := s.repository.ListTasks(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	now := time.Now()
	for _, task := range tasks {
		// Skip tasks that are not enabled
		if !task.Enabled {
			continue
		}

		// Check if task is due to run
		if !task.Schedule.NextRunTime.IsZero() && task.Schedule.NextRunTime.Before(now) {
			// Execute the task in a goroutine
			s.wg.Add(1)
			go func(t *TaskConfig) {
				defer s.wg.Done()

				// Acquire semaphore to limit concurrent tasks
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

// executeTask executes a specific task
func (s *TaskScheduler) executeTask(task *TaskConfig) error {
	slog.Info("Executing scheduled task", "task_id", task.ID, "task_name", task.Name)

	// Get the runner for this task type
	s.mutex.RLock()
	runner, exists := s.runners[task.Type]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no runner registered for task type: %s", task.Type)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(s.ctx, s.config.TaskTimeout)
	defer cancel()

	// Execute the task
	execution, err := runner.Run(ctx, task)
	if err != nil {
		return fmt.Errorf("task execution failed: %w", err)
	}

	// Record the execution
	if err := s.repository.RecordExecution(s.ctx, execution); err != nil {
		return fmt.Errorf("failed to record task execution: %w", err)
	}

	// Update next run time
	if !task.Schedule.OneTime {
		if err := s.updateNextRunTime(task); err != nil {
			return fmt.Errorf("failed to update next run time: %w", err)
		}
	} else {
		// Disable one-time tasks after execution
		task.Enabled = false
		if err := s.repository.UpdateTask(s.ctx, task); err != nil {
			return fmt.Errorf("failed to disable one-time task: %w", err)
		}
	}

	return nil
}

// updateNextRunTime calculates and updates the next run time based on the cron expression
func (s *TaskScheduler) updateNextRunTime(task *TaskConfig) error {
	if task.Schedule.CronExpression == "" {
		return fmt.Errorf("task has no cron expression")
	}

	// Parse the cron expression
	schedule, err := s.cronParser.Parse(task.Schedule.CronExpression)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Calculate the next run time
	now := time.Now()
	nextRun := schedule.Next(now)
	task.Schedule.NextRunTime = nextRun

	// Update the task in the repository
	if err := s.repository.UpdateTask(s.ctx, task); err != nil {
		return fmt.Errorf("failed to update task with next run time: %w", err)
	}

	slog.Info("Updated next run time for task",
		"task_id", task.ID,
		"task_name", task.Name,
		"next_run", nextRun.Format(time.RFC3339))

	return nil
}

// RunTaskNow executes a task immediately, regardless of its schedule
func (s *TaskScheduler) RunTaskNow(taskID string) (*TaskExecution, error) {
	// Get the task
	task, err := s.repository.GetTask(s.ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Get the runner for this task type
	s.mutex.RLock()
	runner, exists := s.runners[task.Type]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no runner registered for task type: %s", task.Type)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(s.ctx, s.config.TaskTimeout)
	defer cancel()

	slog.Info("Running task immediately", "task_id", task.ID, "task_name", task.Name)

	// Execute the task
	execution, err := runner.Run(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("task execution failed: %w", err)
	}

	// Record the execution
	if err := s.repository.RecordExecution(s.ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to record task execution: %w", err)
	}

	return execution, nil
}
