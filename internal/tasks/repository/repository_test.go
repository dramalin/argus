package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/tasks"
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
func createTestTask(id string, taskType tasks.TaskType) *tasks.TaskConfig {
	return &tasks.TaskConfig{
		ID:          id,
		Name:        "Test Task",
		Description: "Test task for repository",
		Type:        taskType,
		Enabled:     true,
		Schedule: tasks.Schedule{
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
func createTestExecution(taskID string, status tasks.TaskStatus) *tasks.TaskExecution {
	return &tasks.TaskExecution{
		ID:        tasks.GenerateID(),
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
	task := createTestTask("test-task-1", tasks.TaskLogRotation)

	// Test creation
	err := repo.CreateTask(context.Background(), task)
	assert.NoError(t, err)

	// Verify the file was created
	taskPath := filepath.Join(tempDir, TasksDir, task.ID+".json")
	_, err = os.Stat(taskPath)
	assert.NoError(t, err)

	// Test duplicate creation
	err = repo.CreateTask(context.Background(), task)
	assert.Error(t, err)
}

func TestFileTaskRepository_GetTask(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-2", tasks.TaskHealthCheck)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Test getting the task
	retrieved, err := repo.GetTask(context.Background(), task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Name, retrieved.Name)
	assert.Equal(t, task.Type, retrieved.Type)

	// Test getting a non-existent task
	_, err = repo.GetTask(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestFileTaskRepository_UpdateTask(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-3", tasks.TaskMetricsAggregation)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Update the task
	task.Name = "Updated Task"
	task.Description = "Updated description"
	task.Parameters["new_param"] = "new_value"
	err = repo.UpdateTask(context.Background(), task)
	assert.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetTask(context.Background(), task.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Task", retrieved.Name)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, "new_value", retrieved.Parameters["new_param"])

	// Test updating a non-existent task
	nonExistentTask := createTestTask("non-existent", tasks.TaskSystemCleanup)
	err = repo.UpdateTask(context.Background(), nonExistentTask)
	assert.Error(t, err)
}

func TestFileTaskRepository_DeleteTask(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-4", tasks.TaskSystemCleanup)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Test deletion
	err = repo.DeleteTask(context.Background(), task.ID)
	assert.NoError(t, err)

	// Verify the file was deleted
	taskPath := filepath.Join(tempDir, TasksDir, task.ID+".json")
	_, err = os.Stat(taskPath)
	assert.True(t, os.IsNotExist(err))

	// Test deleting a non-existent task
	err = repo.DeleteTask(context.Background(), "non-existent")
	assert.Error(t, err)
}

func TestFileTaskRepository_ListTasks(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create multiple test tasks
	task1 := createTestTask("test-task-5", tasks.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task1)
	require.NoError(t, err)

	task2 := createTestTask("test-task-6", tasks.TaskHealthCheck)
	err = repo.CreateTask(context.Background(), task2)
	require.NoError(t, err)

	// Test listing tasks
	tasks, err := repo.ListTasks(context.Background())
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)

	// Verify task IDs are in the list
	ids := []string{tasks[0].ID, tasks[1].ID}
	assert.Contains(t, ids, task1.ID)
	assert.Contains(t, ids, task2.ID)
}

func TestFileTaskRepository_GetTasksByType(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create tasks of different types
	logTask1 := createTestTask("log-task-1", tasks.TaskLogRotation)
	err := repo.CreateTask(context.Background(), logTask1)
	require.NoError(t, err)

	logTask2 := createTestTask("log-task-2", tasks.TaskLogRotation)
	err = repo.CreateTask(context.Background(), logTask2)
	require.NoError(t, err)

	healthTask := createTestTask("health-task", tasks.TaskHealthCheck)
	err = repo.CreateTask(context.Background(), healthTask)
	require.NoError(t, err)

	// Test getting tasks by type
	logTasks, err := repo.GetTasksByType(context.Background(), tasks.TaskLogRotation)
	assert.NoError(t, err)
	assert.Len(t, logTasks, 2)

	healthTasks, err := repo.GetTasksByType(context.Background(), tasks.TaskHealthCheck)
	assert.NoError(t, err)
	assert.Len(t, healthTasks, 1)
	assert.Equal(t, healthTask.ID, healthTasks[0].ID)

	// Test getting tasks of non-existent type
	emptyTasks, err := repo.GetTasksByType(context.Background(), "non-existent")
	assert.NoError(t, err) // No error, just empty list
	assert.Len(t, emptyTasks, 0)
}

func TestFileTaskRepository_RecordExecution(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-7", tasks.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create a test execution
	execution := createTestExecution(task.ID, tasks.StatusCompleted)

	// Test recording execution
	err = repo.RecordExecution(context.Background(), execution)
	assert.NoError(t, err)

	// Verify the execution file was created
	execPath := filepath.Join(tempDir, ExecutionsDir, execution.ID+".json")
	_, err = os.Stat(execPath)
	assert.NoError(t, err)
}

func TestFileTaskRepository_GetExecutions(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-8", tasks.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create test executions
	exec1 := createTestExecution(task.ID, tasks.StatusCompleted)
	exec1.ID = "exec-1"
	exec1.StartTime = time.Now().Add(-2 * time.Minute)

	exec2 := createTestExecution(task.ID, tasks.StatusCompleted)
	exec2.ID = "exec-2"
	exec2.StartTime = time.Now().Add(-1 * time.Minute)

	// Record executions
	err = repo.RecordExecution(context.Background(), exec1)
	require.NoError(t, err)
	err = repo.RecordExecution(context.Background(), exec2)
	require.NoError(t, err)

	// Test getting executions
	executions, err := repo.GetExecutions(context.Background(), task.ID)
	assert.NoError(t, err)
	assert.Len(t, executions, 2)

	// Verify the executions are sorted by start time (newest first)
	assert.Equal(t, exec2.ID, executions[0].ID)
	assert.Equal(t, exec1.ID, executions[1].ID)

	// Test getting executions for non-existent task
	emptyExecs, err := repo.GetExecutions(context.Background(), "non-existent")
	assert.NoError(t, err) // No error, just empty list
	assert.Len(t, emptyExecs, 0)
}

func TestFileTaskRepository_GetExecution(t *testing.T) {
	// Setup
	repo, tempDir := setupTestTaskRepo(t)
	defer os.RemoveAll(tempDir)

	// Create a test task
	task := createTestTask("test-task-9", tasks.TaskLogRotation)
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create a test execution
	execution := createTestExecution(task.ID, tasks.StatusCompleted)
	execution.ID = "exec-3"

	// Record execution
	err = repo.RecordExecution(context.Background(), execution)
	require.NoError(t, err)

	// Test getting execution
	retrieved, err := repo.GetExecution(context.Background(), execution.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ID, retrieved.ID)
	assert.Equal(t, execution.TaskID, retrieved.TaskID)
	assert.Equal(t, execution.Status, retrieved.Status)

	// Test getting non-existent execution
	_, err = repo.GetExecution(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrExecutionNotFound, err)
}
