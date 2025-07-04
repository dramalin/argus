// File: internal/models/task.go
// Brief: Task-related data models for Argus
// Detailed: Contains type definitions for TaskType, TaskStatus, Schedule, TaskConfig, TaskExecution, and related constants/methods.
// Author: drama.lin@aver.com
// Date: 2024-07-03

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TaskType represents the type of system maintenance task
type TaskType string

// Available task types for system maintenance
const (
	TaskLogRotation        TaskType = "log_rotation"        // Log file rotation task
	TaskMetricsAggregation TaskType = "metrics_aggregation" // System metrics aggregation task
	TaskHealthCheck        TaskType = "health_check"        // System health verification task
	TaskSystemCleanup      TaskType = "system_cleanup"      // Temporary file cleanup task
)

// TaskStatus represents the current execution status of a task
type TaskStatus string

// Available task status values
const (
	StatusPending   TaskStatus = "pending"   // Task is scheduled but not yet executed
	StatusRunning   TaskStatus = "running"   // Task is currently running
	StatusCompleted TaskStatus = "completed" // Task has completed successfully
	StatusFailed    TaskStatus = "failed"    // Task has failed during execution
)

// Schedule defines when and how often a task should run
type Schedule struct {
	CronExpression string    `json:"cron_expression"` // Cron expression for recurring tasks
	OneTime        bool      `json:"one_time"`        // Whether this is a one-time task
	NextRunTime    time.Time `json:"next_run_time"`   // Next scheduled execution time
}

// Validate checks if the schedule configuration is valid
func (s *Schedule) Validate() error {
	if s.CronExpression == "" && !s.OneTime {
		return errors.New("either cron_expression or one_time must be set")
	}
	return nil
}

// TaskConfig defines a complete task configuration
type TaskConfig struct {
	ID          string            `json:"id"`                    // Unique identifier for the task
	Name        string            `json:"name"`                  // Human-readable name
	Description string            `json:"description,omitempty"` // Optional description
	Type        TaskType          `json:"type"`                  // Type of task
	Enabled     bool              `json:"enabled"`               // Whether this task is active
	Schedule    Schedule          `json:"schedule"`              // When to run the task
	Parameters  map[string]string `json:"parameters,omitempty"`  // Task-specific parameters
	CreatedAt   time.Time         `json:"created_at"`            // Creation timestamp
	UpdatedAt   time.Time         `json:"updated_at"`            // Last update timestamp
}

// Validate checks if the task configuration is valid
func (t *TaskConfig) Validate() error {
	// Check for required fields
	if t.ID == "" {
		return errors.New("task ID is required")
	}
	if t.Name == "" {
		return errors.New("task name is required")
	}
	if t.Type == "" {
		return errors.New("task type is required")
	}
	// Validate task type
	validTaskTypes := map[TaskType]bool{
		TaskLogRotation:        true,
		TaskMetricsAggregation: true,
		TaskHealthCheck:        true,
		TaskSystemCleanup:      true,
	}
	if !validTaskTypes[t.Type] {
		return fmt.Errorf("invalid task type: %s", t.Type)
	}
	// Validate schedule
	if err := t.Schedule.Validate(); err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}
	return nil
}

// GenerateID creates a new unique ID for a task
func GenerateID() string {
	return uuid.New().String()
}

// TaskRunner defines the interface for executing system maintenance tasks
type TaskRunner interface {
	GetType() TaskType                                                 // Returns the type of task this runner handles
	Execute(ctx context.Context, task TaskConfig) (*TaskResult, error) // Executes a task and returns the result
}

// TaskResult represents the outcome of a task execution
type TaskResult struct {
	ExecutionID string            // Unique identifier for this execution
	StartTime   time.Time         // When the execution started
	EndTime     time.Time         // When the execution completed
	Status      TaskStatus        // Final execution status
	Output      string            // Task output or error message
	Metadata    map[string]string // Additional execution metadata
}

// TaskExecution stores the details of a single task execution
type TaskExecution struct {
	ExecutionID string            // Unique identifier for this execution
	TaskID      string            // ID of the task that was executed
	TaskName    string            // Name of the task that was executed
	TaskType    TaskType          // Type of the task that was executed
	StartTime   time.Time         // When the execution started
	EndTime     time.Time         // When the execution completed
	Status      TaskStatus        // Final execution status
	Output      string            // Task output or error message
	Error       string            // Error message if task failed
	Metadata    map[string]string // Additional execution metadata
}

// NewTaskExecution creates a new execution record for a task
func NewTaskExecution(taskID string) *TaskExecution {
	return &TaskExecution{
		ExecutionID: GenerateID(),
		TaskID:      taskID,
		Status:      StatusPending,
		StartTime:   time.Now(),
	}
}

// Complete marks an execution as completed successfully
func (e *TaskExecution) Complete(output string) {
	e.Status = StatusCompleted
	e.EndTime = time.Now()
	e.Output = output
}

// Fail marks an execution as failed
func (e *TaskExecution) Fail(errMsg string) {
	e.Status = StatusFailed
	e.EndTime = time.Now()
	e.Error = errMsg
}

// Start marks an execution as running
func (e *TaskExecution) Start() {
	e.Status = StatusRunning
	e.StartTime = time.Now()
}

// TaskRepository defines the interface for task storage operations
// Moved from internal/tasks/repository/repository.go as part of architecture migration
// @migration 2024-07-03
// @author Argus
// @see internal/tasks/repository/repository.go

type TaskRepository interface {
	// CreateTask saves a new task configuration
	CreateTask(ctx context.Context, task *TaskConfig) error

	// GetTask retrieves a task configuration by ID
	GetTask(ctx context.Context, id string) (*TaskConfig, error)

	// UpdateTask updates an existing task configuration
	UpdateTask(ctx context.Context, task *TaskConfig) error

	// DeleteTask removes a task configuration
	DeleteTask(ctx context.Context, id string) error

	// ListTasks retrieves all task configurations
	ListTasks(ctx context.Context) ([]*TaskConfig, error)

	// GetTasksByType retrieves task configurations of a specific type
	GetTasksByType(ctx context.Context, taskType TaskType) ([]*TaskConfig, error)

	// RecordExecution saves a task execution record
	RecordExecution(ctx context.Context, execution *TaskExecution) error

	// GetTaskExecutions retrieves execution records for a specific task
	GetTaskExecutions(ctx context.Context, taskID string, limit int) ([]*TaskExecution, error)

	// GetExecution retrieves a specific execution record by ID
	GetExecution(ctx context.Context, id string) (*TaskExecution, error)

	// GetExecutions retrieves all executions for a task
	GetExecutions(ctx context.Context, taskID string) ([]*TaskExecution, error)
}

// FileTaskRepository and all related logic have been moved to internal/database/task_repository.go as part of the Web Application File Storage Architecture Framework migration.
// Please update your imports and references accordingly.
