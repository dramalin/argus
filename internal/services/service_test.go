// File: internal/sync/service_test.go
// Brief: Tests for unified business logic for task scheduling, execution, and runners (migrated from internal/tasks/)
// Detailed: Contains tests for TaskScheduler, TaskRunner, and all related logic for scheduling, executing, and managing system tasks.
// Author: drama.lin@aver.com
// Date: 2024-07-03

package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/database"
	"argus/internal/models"
)

// testingT is an interface that matches both *testing.T and *testing.B
type testingT interface {
	Helper()
	TempDir() string
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func createTestTaskConfig(t testingT) models.TaskConfig {
	t.Helper()
	return models.TaskConfig{
		ID:          "test-task-1",
		Name:        "Test Task",
		Description: "Test maintenance task",
		Type:        models.TaskSystemCleanup,
		Enabled:     true,
		Schedule: models.Schedule{
			CronExpression: "0 * * * *", // Every hour
			OneTime:        false,
			NextRunTime:    time.Now().Add(time.Hour),
		},
		Parameters: map[string]string{
			"retention_days": "7",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestTaskStore(t testingT) models.TaskRepository {
	t.Helper()
	tempDir := t.TempDir()
	store, err := database.NewFileTaskRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create task store: %v", err)
	}
	return store
}

func TestNewTaskScheduler(t *testing.T) {
	taskStore := createTestTaskStore(t)
	config := &TaskSchedulerConfig{
		CheckInterval:      30 * time.Second,
		MaxConcurrentTasks: 3,
		TaskTimeout:        15 * time.Minute,
	}

	scheduler := NewTaskScheduler(taskStore, config)
	assert.NotNil(t, scheduler)
	assert.Equal(t, config, scheduler.config)
	assert.NotNil(t, scheduler.runners)
	assert.Equal(t, config.MaxConcurrentTasks, cap(scheduler.semaphore))

	// Test with default config
	scheduler = NewTaskScheduler(taskStore, nil)
	assert.NotNil(t, scheduler)
	assert.Equal(t, DefaultCheckInterval, scheduler.config.CheckInterval)
	assert.Equal(t, DefaultMaxConcurrentTasks, scheduler.config.MaxConcurrentTasks)
	assert.Equal(t, DefaultTaskTimeout, scheduler.config.TaskTimeout)
}

func TestTaskSchedulerStartStop(t *testing.T) {
	taskStore := createTestTaskStore(t)
	scheduler := NewTaskScheduler(taskStore, nil)

	// Register a test runner
	testRunner := &mockTaskRunner{
		taskType: models.TaskSystemCleanup,
	}
	scheduler.RegisterRunner(testRunner)

	// Verify runner registration
	assert.Len(t, scheduler.runners, 1)
	assert.Equal(t, testRunner, scheduler.runners[models.TaskSystemCleanup])

	// Start scheduler
	err := scheduler.Start()
	require.NoError(t, err)

	// Create test task
	ctx := context.Background()
	testTask := createTestTaskConfig(t)
	err = taskStore.CreateTask(ctx, &testTask)
	require.NoError(t, err)

	// Wait for a bit to let the scheduler run
	time.Sleep(50 * time.Millisecond)

	// Verify running state
	scheduler.mutex.RLock()
	assert.True(t, scheduler.running)
	scheduler.mutex.RUnlock()

	// Clean shutdown
	scheduler.Stop()
	scheduler.mutex.RLock()
	assert.False(t, scheduler.running)
	scheduler.mutex.RUnlock()
}

// Test helpers and mocks

// mockTaskRunner is a mock implementation for testing
type mockTaskRunner struct {
	taskType      models.TaskType
	runFunc       func(context.Context, *models.TaskConfig) (*models.TaskExecution, error)
	delay         time.Duration           // simulates task execution time
	shouldTimeout bool                    // forces task to timeout
	errorOnRun    error                   // simulates run error
	executions    []*models.TaskExecution // tracks all executions
}

func newMockTaskRunner(taskType models.TaskType) *mockTaskRunner {
	return &mockTaskRunner{
		taskType:   taskType,
		executions: make([]*models.TaskExecution, 0),
	}
}

func (r *mockTaskRunner) GetType() models.TaskType {
	return r.taskType
}

func (r *mockTaskRunner) Run(ctx context.Context, task *models.TaskConfig) (*models.TaskExecution, error) {
	if r.errorOnRun != nil {
		return nil, r.errorOnRun
	}

	if r.delay > 0 {
		select {
		case <-ctx.Done():
			if r.shouldTimeout {
				return nil, ctx.Err()
			}
		case <-time.After(r.delay):
			// Continue with execution
		}
	}

	if r.runFunc != nil {
		return r.runFunc(ctx, task)
	}

	exec := &models.TaskExecution{
		ID:        uuid.New().String(),
		TaskID:    task.ID,
		Status:    models.StatusCompleted,
		StartTime: time.Now().Add(-r.delay), // Account for simulated execution time
		EndTime:   time.Now(),
		Output:    "Task completed successfully",
	}

	r.executions = append(r.executions, exec)
	return exec, nil
}

// verifyTaskExecution checks if a task was executed correctly
func verifyTaskExecution(t *testing.T, runner *mockTaskRunner, taskID string, expectedStatus models.TaskStatus) {
	t.Helper()
	var found bool
	for _, exec := range runner.executions {
		if exec.TaskID == taskID {
			found = true
			assert.Equal(t, expectedStatus, exec.Status, "Task execution status mismatch")
			assert.NotZero(t, exec.StartTime, "Task start time should be set")
			assert.NotZero(t, exec.EndTime, "Task end time should be set")
			assert.False(t, exec.EndTime.Before(exec.StartTime), "Task end time should not be before start time")
			break
		}
	}
	assert.True(t, found, "Task execution not found")
}

// waitForNExecutions waits for a specific number of task executions
func waitForNExecutions(t *testing.T, runner *mockTaskRunner, n int, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(runner.executions) >= n {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// createConcurrentTasks creates multiple tasks for concurrent execution testing
func createConcurrentTasks(t *testing.T, store models.TaskRepository, n int) []models.TaskConfig {
	t.Helper()
	tasks := make([]models.TaskConfig, n)
	for i := 0; i < n; i++ {
		tasks[i] = models.TaskConfig{
			ID:          fmt.Sprintf("test-task-%d", i+1),
			Name:        fmt.Sprintf("Test Task %d", i+1),
			Description: "Concurrent test task",
			Type:        models.TaskSystemCleanup,
			Enabled:     true,
			Schedule: models.Schedule{
				CronExpression: "* * * * *",
				OneTime:        false,
				NextRunTime:    time.Now(),
			},
			Parameters: map[string]string{
				"test_param": fmt.Sprintf("value-%d", i+1),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, store.CreateTask(context.Background(), &tasks[i]))
	}
	return tasks
}

func TestTaskSchedulerExecution(t *testing.T) {
	taskStore := createTestTaskStore(t)
	scheduler := NewTaskScheduler(taskStore, &TaskSchedulerConfig{
		CheckInterval:      50 * time.Millisecond,
		MaxConcurrentTasks: 1,
		TaskTimeout:        1 * time.Second,
	})

	executionCount := 0
	testRunner := &mockTaskRunner{
		taskType: models.TaskSystemCleanup,
		runFunc: func(ctx context.Context, task *models.TaskConfig) (*models.TaskExecution, error) {
			executionCount++
			return &models.TaskExecution{
				ID:        task.ID,
				TaskID:    task.ID,
				Status:    models.StatusCompleted,
				StartTime: time.Now(),
				EndTime:   time.Now(),
			}, nil
		},
	}
	scheduler.RegisterRunner(testRunner)

	// Start scheduler
	err := scheduler.Start()
	require.NoError(t, err)

	// Create test task
	testTask := createTestTaskConfig(t)
	testTask.Schedule.NextRunTime = time.Now() // Set to run immediately
	err = taskStore.CreateTask(context.Background(), &testTask)
	require.NoError(t, err)

	// Wait for execution
	time.Sleep(200 * time.Millisecond)

	// Verify task was executed
	assert.Greater(t, executionCount, 0)

	// Clean up
	scheduler.Stop()
}

func TestTaskSchedulerConcurrency(t *testing.T) {
	taskStore := createTestTaskStore(t)
	maxConcurrent := 3
	scheduler := NewTaskScheduler(taskStore, &TaskSchedulerConfig{
		CheckInterval:      50 * time.Millisecond,
		MaxConcurrentTasks: maxConcurrent,
		TaskTimeout:        1 * time.Second,
	})

	// Create a runner that simulates work with a delay
	runner := newMockTaskRunner(models.TaskSystemCleanup)
	runner.delay = 200 * time.Millisecond // Each task takes 200ms
	scheduler.RegisterRunner(runner)

	// Start scheduler
	err := scheduler.Start()
	require.NoError(t, err)
	defer scheduler.Stop()

	// Create multiple tasks
	tasks := createConcurrentTasks(t, taskStore, maxConcurrent+2) // Create more tasks than max concurrent

	// Wait for executions
	ok := waitForNExecutions(t, runner, len(tasks), 2*time.Second)
	assert.True(t, ok, "Not all tasks were executed")

	// Verify all tasks were executed
	for _, task := range tasks {
		verifyTaskExecution(t, runner, task.ID, models.StatusCompleted)
	}

	// Verify execution timing indicates concurrent execution
	if len(runner.executions) >= 2 {
		firstEnd := runner.executions[0].EndTime
		secondStart := runner.executions[1].StartTime
		assert.True(t, secondStart.Before(firstEnd),
			"Second task should start before first task ends, indicating concurrent execution")
	}
}

func TestTaskSchedulerTimeout(t *testing.T) {
	taskStore := createTestTaskStore(t)
	scheduler := NewTaskScheduler(taskStore, &TaskSchedulerConfig{
		CheckInterval:      50 * time.Millisecond,
		MaxConcurrentTasks: 1,
		TaskTimeout:        100 * time.Millisecond,
	})

	// Create a runner that will timeout
	runner := newMockTaskRunner(models.TaskSystemCleanup)
	runner.delay = 200 * time.Millisecond // Longer than timeout
	runner.shouldTimeout = true
	scheduler.RegisterRunner(runner)

	// Start scheduler
	err := scheduler.Start()
	require.NoError(t, err)
	defer scheduler.Stop()

	// Create test task
	task := createTestTaskConfig(t)
	task.Schedule.NextRunTime = time.Now()
	err = taskStore.CreateTask(context.Background(), &task)
	require.NoError(t, err)

	// Wait for execution
	time.Sleep(300 * time.Millisecond)

	// Verify task execution was attempted
	assert.NotEmpty(t, runner.executions, "Task should have been attempted")
}

func TestTaskSchedulerErrorHandling(t *testing.T) {
	taskStore := createTestTaskStore(t)
	scheduler := NewTaskScheduler(taskStore, &TaskSchedulerConfig{
		CheckInterval:      50 * time.Millisecond,
		MaxConcurrentTasks: 1,
		TaskTimeout:        1 * time.Second,
	})

	// Create a runner that returns an error
	testError := errors.New("test error")
	runner := newMockTaskRunner(models.TaskSystemCleanup)
	runner.errorOnRun = testError
	scheduler.RegisterRunner(runner)

	// Start scheduler
	err := scheduler.Start()
	require.NoError(t, err)
	defer scheduler.Stop()

	// Create test task
	task := createTestTaskConfig(t)
	task.Schedule.NextRunTime = time.Now()
	err = taskStore.CreateTask(context.Background(), &task)
	require.NoError(t, err)

	// Wait for execution attempt
	time.Sleep(100 * time.Millisecond)

	// Verify task state reflects error
	executions, err := taskStore.GetExecutions(context.Background(), task.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, executions, "Should have at least one execution")
	lastExec := executions[len(executions)-1]
	assert.Equal(t, models.StatusFailed, lastExec.Status)
}

func TestTaskSchedulerRescheduling(t *testing.T) {
	taskStore := createTestTaskStore(t)
	scheduler := NewTaskScheduler(taskStore, &TaskSchedulerConfig{
		CheckInterval:      50 * time.Millisecond,
		MaxConcurrentTasks: 1,
		TaskTimeout:        1 * time.Second,
	})

	runner := newMockTaskRunner(models.TaskSystemCleanup)
	scheduler.RegisterRunner(runner)

	// Start scheduler
	err := scheduler.Start()
	require.NoError(t, err)
	defer scheduler.Stop()

	// Create recurring task
	task := createTestTaskConfig(t)
	task.Schedule.CronExpression = "* * * * * *" // Run every second
	task.Schedule.NextRunTime = time.Now()
	err = taskStore.CreateTask(context.Background(), &task)
	require.NoError(t, err)

	// Wait for multiple executions
	time.Sleep(200 * time.Millisecond)
	firstCount := len(runner.executions)
	assert.Greater(t, firstCount, 0, "Task should have executed at least once")

	time.Sleep(1 * time.Second)
	secondCount := len(runner.executions)
	assert.Greater(t, secondCount, firstCount, "Task should have been rescheduled and executed again")
}

func TestTaskSchedulerEdgeCases(t *testing.T) {
	taskStore := createTestTaskStore(t)
	scheduler := NewTaskScheduler(taskStore, nil) // Use default config

	t.Run("UnknownTaskType", func(t *testing.T) {
		task := createTestTaskConfig(t)
		task.Type = "unknown_type"
		task.Schedule.NextRunTime = time.Now()
		err := taskStore.CreateTask(context.Background(), &task)
		require.NoError(t, err)

		// Start scheduler without any runners
		err = scheduler.Start()
		require.NoError(t, err)
		defer scheduler.Stop()

		time.Sleep(100 * time.Millisecond)

		// Verify task was marked as failed due to unknown type
		executions, err := taskStore.GetExecutions(context.Background(), task.ID)
		require.NoError(t, err)
		if assert.NotEmpty(t, executions) {
			assert.Equal(t, models.StatusFailed, executions[0].Status)
		}
	})

	t.Run("DisabledTask", func(t *testing.T) {
		runner := newMockTaskRunner(models.TaskSystemCleanup)
		scheduler.RegisterRunner(runner)

		task := createTestTaskConfig(t)
		task.Enabled = false
		task.Schedule.NextRunTime = time.Now()
		err := taskStore.CreateTask(context.Background(), &task)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		assert.Empty(t, runner.executions, "Disabled task should not be executed")
	})

	t.Run("ConcurrentTaskLimit", func(t *testing.T) {
		scheduler := NewTaskScheduler(taskStore, &TaskSchedulerConfig{
			CheckInterval:      50 * time.Millisecond,
			MaxConcurrentTasks: 1,
			TaskTimeout:        1 * time.Second,
		})

		runner := newMockTaskRunner(models.TaskSystemCleanup)
		runner.delay = 200 * time.Millisecond
		scheduler.RegisterRunner(runner)

		err := scheduler.Start()
		require.NoError(t, err)
		defer scheduler.Stop()

		// Create multiple tasks
		_ = createConcurrentTasks(t, taskStore, 3)
		time.Sleep(400 * time.Millisecond)

		// Verify that tasks were executed sequentially
		assert.LessOrEqual(t, len(runner.executions), 2,
			"With 200ms delay and 400ms wait, no more than 2 tasks should complete")
	})
}

// BenchmarkTaskScheduler provides performance metrics for task scheduling and execution
func BenchmarkTaskScheduler(b *testing.B) {
	taskStore := createTestTaskStore(b)
	scheduler := NewTaskScheduler(taskStore, &TaskSchedulerConfig{
		CheckInterval:      10 * time.Millisecond,
		MaxConcurrentTasks: 10,
		TaskTimeout:        1 * time.Second,
	})

	runner := newMockTaskRunner(models.TaskSystemCleanup)
	scheduler.RegisterRunner(runner)

	err := scheduler.Start()
	require.NoError(b, err)
	defer scheduler.Stop()

	ctx := context.Background()

	b.Run("TaskCreationAndScheduling", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			task := models.TaskConfig{
				ID:      fmt.Sprintf("bench-task-%d", i),
				Name:    fmt.Sprintf("Bench Task %d", i),
				Type:    models.TaskSystemCleanup,
				Enabled: true,
				Schedule: models.Schedule{
					CronExpression: "* * * * *",
					NextRunTime:    time.Now(),
				},
			}
			if err := taskStore.CreateTask(ctx, &task); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ConcurrentTaskExecution", func(b *testing.B) {
		// Pre-create tasks
		tasks := make([]models.TaskConfig, b.N)
		for i := 0; i < b.N; i++ {
			tasks[i] = models.TaskConfig{
				ID:      fmt.Sprintf("bench-concurrent-%d", i),
				Name:    fmt.Sprintf("Bench Concurrent %d", i),
				Type:    models.TaskSystemCleanup,
				Enabled: true,
				Schedule: models.Schedule{
					CronExpression: "* * * * *",
					NextRunTime:    time.Now(),
				},
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := taskStore.CreateTask(ctx, &tasks[i]); err != nil {
				b.Fatal(err)
			}
		}
		// Wait for execution completion
		time.Sleep(100 * time.Millisecond)
	})
}
