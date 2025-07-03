// Package tasks provides functionality for scheduling and managing system maintenance tasks
package tasks

import (
	"context"
)

// TaskRepository defines the interface for persisting and retrieving tasks
type TaskRepository interface {
	// CreateTask creates a new task configuration
	CreateTask(ctx context.Context, task *TaskConfig) error

	// GetTask retrieves a task configuration by ID
	GetTask(ctx context.Context, id string) (*TaskConfig, error)

	// UpdateTask updates an existing task configuration
	UpdateTask(ctx context.Context, task *TaskConfig) error

	// DeleteTask deletes a task configuration by ID
	DeleteTask(ctx context.Context, id string) error

	// ListTasks retrieves all task configurations
	ListTasks(ctx context.Context) ([]*TaskConfig, error)

	// GetTasksByType retrieves all task configurations of a specific type
	GetTasksByType(ctx context.Context, taskType TaskType) ([]*TaskConfig, error)

	// RecordExecution records a task execution
	RecordExecution(ctx context.Context, execution *TaskExecution) error

	// GetExecutions retrieves all executions for a task
	GetExecutions(ctx context.Context, taskID string) ([]*TaskExecution, error)

	// GetTaskExecutions retrieves recent executions for a task with a limit
	GetTaskExecutions(ctx context.Context, taskID string, limit int) ([]*TaskExecution, error)

	// GetExecution retrieves a specific execution by ID
	GetExecution(ctx context.Context, id string) (*TaskExecution, error)
}
