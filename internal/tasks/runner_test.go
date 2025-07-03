package tasks

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskRunner(t *testing.T) {
	// Test creating valid runners
	runner, err := NewTaskRunner(TaskLogRotation)
	assert.NoError(t, err)
	assert.Equal(t, TaskLogRotation, runner.GetType())

	runner, err = NewTaskRunner(TaskMetricsAggregation)
	assert.NoError(t, err)
	assert.Equal(t, TaskMetricsAggregation, runner.GetType())

	runner, err = NewTaskRunner(TaskHealthCheck)
	assert.NoError(t, err)
	assert.Equal(t, TaskHealthCheck, runner.GetType())

	runner, err = NewTaskRunner(TaskSystemCleanup)
	assert.NoError(t, err)
	assert.Equal(t, TaskSystemCleanup, runner.GetType())

	// Test creating invalid runner
	_, err = NewTaskRunner("invalid_type")
	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedTaskType, err)
}

func createTestTask(taskType TaskType, enabled bool) *TaskConfig {
	return &TaskConfig{
		ID:          GenerateID(),
		Name:        "Test Task",
		Description: "Test task for runner",
		Type:        taskType,
		Enabled:     enabled,
		Schedule: Schedule{
			CronExpression: "* * * * *",
			NextRunTime:    time.Now(),
		},
		Parameters: map[string]string{
			"log_dir":     "/tmp/logs",
			"max_size_mb": "10",
			"keep_count":  "5",
			"url":         "http://localhost:8080",
			"timeout":     "30",
			"pattern":     "*.log",
			"retention":   "7",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTempDir(t *testing.T, prefix string) string {
	dir, err := os.MkdirTemp("", prefix)
	require.NoError(t, err)
	return dir
}

func TestLogRotationRunner_Run(t *testing.T) {
	// Create a test task
	task := createTestTask(TaskLogRotation, true)

	// Create a temporary log directory
	logDir := createTempDir(t, "log_rotation_test")
	defer os.RemoveAll(logDir)

	// Create some test log files
	testFiles := []string{"app.log", "error.log", "access.log"}
	for _, file := range testFiles {
		filePath := filepath.Join(logDir, file)
		err := os.WriteFile(filePath, []byte("test log content"), 0644)
		require.NoError(t, err)
	}

	// Set the task parameters
	task.Parameters["log_dir"] = logDir
	task.Parameters["pattern"] = "*.log"
	task.Parameters["max_size_mb"] = "1"
	task.Parameters["keep_count"] = "3"

	// Create the runner
	runner, err := NewTaskRunner(TaskLogRotation)
	require.NoError(t, err)

	// Run the task
	ctx := context.Background()
	execution, err := runner.Run(ctx, task)
	require.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, task.ID, execution.TaskID)
	assert.Equal(t, StatusCompleted, execution.Status)
	assert.NotEmpty(t, execution.Output)

	// Test with invalid parameters
	invalidTask := createTestTask(TaskLogRotation, true)
	invalidTask.Parameters = map[string]string{}

	execution, err = runner.Run(ctx, invalidTask)
	assert.NoError(t, err) // Should still complete but with warning
	assert.Contains(t, execution.Output, "warning")
}

func TestMetricsAggregationRunner_Run(t *testing.T) {
	// Create a test task
	task := createTestTask(TaskMetricsAggregation, true)

	// Create a temporary directory for metrics
	metricsDir := createTempDir(t, "metrics_test")
	defer os.RemoveAll(metricsDir)

	// Set the task parameters
	task.Parameters["metrics_dir"] = metricsDir
	task.Parameters["retention_days"] = "7"

	// Create the runner
	runner, err := NewTaskRunner(TaskMetricsAggregation)
	require.NoError(t, err)

	// Run the task
	ctx := context.Background()
	execution, err := runner.Run(ctx, task)
	require.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, task.ID, execution.TaskID)
	assert.Equal(t, StatusCompleted, execution.Status)
	assert.NotEmpty(t, execution.Output)

	// Test with context cancellation
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	execution, err = runner.Run(cancelCtx, task)
	assert.Error(t, err)
	assert.Equal(t, StatusFailed, execution.Status)
	assert.Contains(t, err.Error(), "context")
}

func TestHealthCheckRunner_Run(t *testing.T) {
	// Create a test task
	task := createTestTask(TaskHealthCheck, true)

	// Set a non-existent URL to simulate a failure
	task.Parameters["url"] = "http://localhost:12345/health" // Unlikely to exist
	task.Parameters["timeout"] = "1"                         // 1 second timeout

	// Create the runner
	runner, err := NewTaskRunner(TaskHealthCheck)
	require.NoError(t, err)

	// Run the task
	ctx := context.Background()
	execution, err := runner.Run(ctx, task)
	require.NoError(t, err) // Error in health check doesn't fail the task
	assert.NotNil(t, execution)
	assert.Equal(t, task.ID, execution.TaskID)
	assert.Equal(t, StatusCompleted, execution.Status)
	assert.Contains(t, execution.Output, "failed")

	// Test with invalid URL
	invalidTask := createTestTask(TaskHealthCheck, true)
	invalidTask.Parameters["url"] = "invalid-url"

	execution, err = runner.Run(ctx, invalidTask)
	assert.NoError(t, err)
	assert.Equal(t, StatusCompleted, execution.Status)
	assert.Contains(t, execution.Output, "invalid")
}

func TestSystemCleanupRunner_Run(t *testing.T) {
	// Create a test task
	task := createTestTask(TaskSystemCleanup, true)

	// Create a temporary directory for cleanup
	cleanupDir := createTempDir(t, "cleanup_test")
	defer os.RemoveAll(cleanupDir)

	// Create some test files with different modification times
	oldFile := filepath.Join(cleanupDir, "old.tmp")
	err := os.WriteFile(oldFile, []byte("old file"), 0644)
	require.NoError(t, err)

	// Set the file's modification time to 10 days ago
	oldTime := time.Now().Add(-240 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	require.NoError(t, err)

	// Create a new file
	newFile := filepath.Join(cleanupDir, "new.tmp")
	err = os.WriteFile(newFile, []byte("new file"), 0644)
	require.NoError(t, err)

	// Set the task parameters
	task.Parameters["cleanup_dir"] = cleanupDir
	task.Parameters["pattern"] = "*.tmp"
	task.Parameters["retention_days"] = "7" // Keep files newer than 7 days

	// Create the runner
	runner, err := NewTaskRunner(TaskSystemCleanup)
	require.NoError(t, err)

	// Run the task
	ctx := context.Background()
	execution, err := runner.Run(ctx, task)
	require.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, task.ID, execution.TaskID)
	assert.Equal(t, StatusCompleted, execution.Status)
	assert.NotEmpty(t, execution.Output)

	// Verify the old file was deleted and the new file remains
	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(newFile)
	assert.NoError(t, err)
}
