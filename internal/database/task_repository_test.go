package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/models"
)

// setupTestTaskRepo creates a temporary directory and returns a new FileTaskRepository for testing
func setupTestTaskRepo(t *testing.T) (*FileTaskRepository, string) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "task_repo_test")
	require.NoError(t, err)

	// Create a new file task repository
	repo, err := NewFileTaskRepository(tempDir)
	require.NoError(t, err)

	return repo, tempDir
}

// createTestTask creates a test task configuration
func createTestTask(id string, taskType models.TaskType) *models.TaskConfig {
	return &models.TaskConfig{
		ID:          id,
		Name:        "Test Task",
		Description: "Test task for repository",
		Type:        taskType,
		Enabled:     true,
		Schedule: models.Schedule{
			CronExpression: "*/5 * * * *",
			NextRunTime:    time.Now().Add(5 * time.Minute),
		},
		Parameters: map[string]string{
			"log_dir":     "/var/log",
			"max_size_mb": "10",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// createTestExecution creates a test task execution record
func createTestExecution(taskID string, status models.TaskStatus) *models.TaskExecution {
	return &models.TaskExecution{
		ID:        models.GenerateID(),
		TaskID:    taskID,
		Status:    status,
		StartTime: time.Now().Add(-1 * time.Minute),
		EndTime:   time.Now(),
		Output:    "Test execution output",
	}
}

func TestNewFileTaskRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "repo_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test creating a repository
	repo, err := NewFileTaskRepository(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	// Verify directories were created
	tasksDir := filepath.Join(tempDir, TasksDir)
	executionsDir := filepath.Join(tempDir, ExecutionsDir)

	_, err = os.Stat(tasksDir)
	assert.NoError(t, err)

	_, err = os.Stat(executionsDir)
	assert.NoError(t, err)

	// Test with invalid path
	_, err = NewFileTaskRepository("/nonexistent/path")
	assert.Error(t, err)
}

func TestFileTaskRepository_CreateTask(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-1", models.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task)
	assert.NoError(t, err)

	// Verify the task was created and can be retrieved
	retrieved, err := repo.GetTask(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Name, retrieved.Name)
	assert.Equal(t, task.Description, retrieved.Description)
	assert.Equal(t, task.Type, retrieved.Type)

	// Try to create the same task again (should fail)
	err = repo.CreateTask(context.Background(), task)
	assert.Error(t, err)

	// Create a task with an invalid configuration
	invalidTask := createTestTask("", models.TaskHealthCheck)
	err = repo.CreateTask(context.Background(), invalidTask)
	assert.Error(t, err)
}

func TestFileTaskRepository_GetTask(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-2", models.TaskHealthCheck)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Get the task
	retrieved, err := repo.GetTask(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Name, retrieved.Name)
	assert.Equal(t, task.Description, retrieved.Description)
	assert.Equal(t, task.Type, retrieved.Type)

	// Try to get a non-existent task
	_, err = repo.GetTask(context.Background(), "non-existent-id")
	assert.ErrorIs(t, err, ErrTaskNotFound)

	// Try to get with an empty ID
	_, err = repo.GetTask(context.Background(), "")
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestFileTaskRepository_UpdateTask(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-3", models.TaskMetricsAggregation)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Update the task
	task.Name = "Updated Task Name"
	task.Description = "Updated description"
	task.Parameters["new_param"] = "new_value"
	err = repo.UpdateTask(context.Background(), task)
	require.NoError(t, err)

	// Get the updated task
	retrieved, err := repo.GetTask(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Task Name", retrieved.Name)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, "new_value", retrieved.Parameters["new_param"])

	// Try to update a non-existent task
	nonExistentTask := createTestTask("non-existent", models.TaskSystemCleanup)
	err = repo.UpdateTask(context.Background(), nonExistentTask)
	assert.ErrorIs(t, err, ErrTaskNotFound)

	// Try to update with an empty ID
	invalidTask := createTestTask("", models.TaskSystemCleanup)
	err = repo.UpdateTask(context.Background(), invalidTask)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestFileTaskRepository_DeleteTask(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-4", models.TaskSystemCleanup)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Delete the task
	err = repo.DeleteTask(context.Background(), task.ID)
	require.NoError(t, err)

	// Verify the task was deleted
	_, err = repo.GetTask(context.Background(), task.ID)
	assert.ErrorIs(t, err, ErrTaskNotFound)

	// Try to delete a non-existent task
	err = repo.DeleteTask(context.Background(), "non-existent-id")
	assert.ErrorIs(t, err, ErrTaskNotFound)

	// Try to delete with an empty ID
	err = repo.DeleteTask(context.Background(), "")
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestFileTaskRepository_ListTasks(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create multiple test tasks
	task1 := createTestTask("test-task-5", models.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task1)
	require.NoError(t, err)

	task2 := createTestTask("test-task-6", models.TaskHealthCheck)
	err = repo.CreateTask(context.Background(), task2)
	require.NoError(t, err)

	// List all tasks
	tasks, err := repo.ListTasks(context.Background())
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	// Verify task IDs are in the list
	taskIDs := make(map[string]bool)
	for _, t := range tasks {
		taskIDs[t.ID] = true
	}
	assert.True(t, taskIDs[task1.ID])
	assert.True(t, taskIDs[task2.ID])
}

func TestFileTaskRepository_GetTasksByType(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create multiple test tasks with different types
	logTask1 := createTestTask("log-task-1", models.TaskLogRotation)
	err := repo.CreateTask(context.Background(), logTask1)
	require.NoError(t, err)

	logTask2 := createTestTask("log-task-2", models.TaskLogRotation)
	err = repo.CreateTask(context.Background(), logTask2)
	require.NoError(t, err)

	healthTask := createTestTask("health-task", models.TaskHealthCheck)
	err = repo.CreateTask(context.Background(), healthTask)
	require.NoError(t, err)

	// Get tasks by type
	logTasks, err := repo.GetTasksByType(context.Background(), models.TaskLogRotation)
	require.NoError(t, err)
	assert.Len(t, logTasks, 2)

	healthTasks, err := repo.GetTasksByType(context.Background(), models.TaskHealthCheck)
	require.NoError(t, err)
	assert.Len(t, healthTasks, 1)
}

func TestFileTaskRepository_RecordExecution(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-7", models.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Record a task execution
	execution := createTestExecution(task.ID, models.StatusCompleted)
	err = repo.RecordExecution(context.Background(), execution)
	require.NoError(t, err)

	// Get the recorded execution
	retrieved, err := repo.GetExecution(context.Background(), execution.ID)
	require.NoError(t, err)
	assert.Equal(t, execution.ID, retrieved.ID)
	assert.Equal(t, task.ID, retrieved.TaskID)
	assert.Equal(t, models.StatusCompleted, retrieved.Status)

	// Try to record with empty execution ID
	invalidExec := createTestExecution(task.ID, models.StatusCompleted)
	invalidExec.ID = ""
	err = repo.RecordExecution(context.Background(), invalidExec)
	assert.Error(t, err)
}

func TestFileTaskRepository_GetTaskExecutions(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-8", models.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create test executions
	exec1 := createTestExecution(task.ID, models.StatusCompleted)
	exec1.ID = "exec-1"
	exec1.StartTime = time.Now().Add(-2 * time.Minute)
	err = repo.RecordExecution(context.Background(), exec1)
	require.NoError(t, err)

	exec2 := createTestExecution(task.ID, models.StatusCompleted)
	exec2.ID = "exec-2"
	exec2.StartTime = time.Now().Add(-1 * time.Minute)
	err = repo.RecordExecution(context.Background(), exec2)
	require.NoError(t, err)

	// Get executions with limit
	executions, err := repo.GetTaskExecutions(context.Background(), task.ID, 2)
	assert.NoError(t, err)
	assert.Len(t, executions, 2)

	// Verify the executions are sorted by start time (newest first)
	assert.Equal(t, exec2.ID, executions[0].ID)
	assert.Equal(t, exec1.ID, executions[1].ID)

	// Test limit
	limitedExecs, err := repo.GetTaskExecutions(context.Background(), task.ID, 1)
	assert.NoError(t, err)
	assert.Len(t, limitedExecs, 1)
	assert.Equal(t, exec2.ID, limitedExecs[0].ID) // Should get the newest one

	// Test getting executions for non-existent task
	emptyExecs, err := repo.GetTaskExecutions(context.Background(), "non-existent", 10)
	assert.NoError(t, err) // No error, just empty list
	assert.Empty(t, emptyExecs)
}

func TestFileTaskRepository_GetExecution(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-9", models.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create and record an execution
	execution := createTestExecution(task.ID, models.StatusCompleted)
	err = repo.RecordExecution(context.Background(), execution)
	require.NoError(t, err)

	// Get the execution
	retrieved, err := repo.GetExecution(context.Background(), execution.ID)
	require.NoError(t, err)
	assert.Equal(t, execution.ID, retrieved.ID)
	assert.Equal(t, task.ID, retrieved.TaskID)
	assert.Equal(t, models.StatusCompleted, retrieved.Status)

	// Try to get a non-existent execution
	_, err = repo.GetExecution(context.Background(), "non-existent-id")
	assert.ErrorIs(t, err, ErrExecutionNotFound)
}
