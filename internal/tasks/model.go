package tasks

import (
	"context"

	"argus/internal/models"
)

// TaskRepository defines operations for storing and retrieving tasks
type TaskRepository interface {
	CreateTask(ctx context.Context, task *models.TaskConfig) error
	GetTask(ctx context.Context, id string) (*models.TaskConfig, error)
	ListTasks(ctx context.Context) ([]models.TaskConfig, error)
	UpdateTask(ctx context.Context, id string, task *models.TaskConfig) error
	DeleteTask(ctx context.Context, id string) error

	// Execution-related methods
	RecordExecution(ctx context.Context, execution *models.TaskExecution) error
	GetExecution(ctx context.Context, id string) (*models.TaskExecution, error)
	ListExecutions(ctx context.Context, taskID string, limit int) ([]models.TaskExecution, error)
}

// Removed duplicate TaskRunner interface. Use the canonical definition from services/models.
