// Package tasks provides functionality for scheduling and managing system maintenance tasks
package tasks

import (
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

// TaskExecution represents a single execution of a task
type TaskExecution struct {
	ID        string     `json:"id"`                 // Unique identifier for this execution
	TaskID    string     `json:"task_id"`            // ID of the task being executed
	Status    TaskStatus `json:"status"`             // Current status of execution
	StartTime time.Time  `json:"start_time"`         // When execution started
	EndTime   time.Time  `json:"end_time,omitempty"` // When execution completed (if applicable)
	Error     string     `json:"error,omitempty"`    // Error message if execution failed
	Output    string     `json:"output,omitempty"`   // Output from the task execution
}

// NewTaskExecution creates a new execution record for a task
func NewTaskExecution(taskID string) *TaskExecution {
	return &TaskExecution{
		ID:        GenerateID(),
		TaskID:    taskID,
		Status:    StatusPending,
		StartTime: time.Now(),
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
