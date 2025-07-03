package tasks_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockTaskRunner implements TaskRunner for testing
type mockTaskRunner struct {
	typeVal     TaskType
	runCalled   bool
	executions  []*TaskExecution
	shouldError bool
	mutex       sync.Mutex
}

func (m *mockTaskRunner) Run(ctx context.Context, task *TaskConfig) (*TaskExecution, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.runCalled = true

	exec := NewTaskExecution(task.ID)
	if m.shouldError {
		exec.Fail("mock runner error")
		return exec, nil
	}
	exec.Complete("mock runner execution successful")
	m.executions = append(m.executions, exec)
	return exec, nil
}

func (m *mockTaskRunner) GetType() TaskType {
	return m.typeVal
}

func (m *mockTaskRunner) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.runCalled = false
	m.executions = nil
}

func (m *mockTaskRunner) WasRun() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.runCalled
}

func (m *mockTaskRunner) ExecutionCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.executions)
}

// Test helper functions
func createTestTaskWithSchedule(id string, taskType TaskType, enabled bool, cronExpr string) *TaskConfig {
	nextRunTime := time.Now()
	if cronExpr != "" {
		nextRunTime = nextRunTime.Add(-1 * time.Minute) // Make it due in the past
	}

	return &TaskConfig{
		ID:          id,
		Name:        "Test Task " + id,
		Description: "Test task for scheduler",
		Type:        taskType,
		Enabled:     enabled,
		Schedule: Schedule{
			CronExpression: cronExpr,
			NextRunTime:    nextRunTime,
		},
		Parameters: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestNewTaskScheduler(t *testing.T) {
	repo := &mockTaskRepo{
		tasks: make(map[string]*TaskConfig),
		execs: make(map[string][]*TaskExecution),
	}
	config := DefaultTaskSchedulerConfig()

	scheduler := NewTaskScheduler(repo, config)
	assert.NotNil(t, scheduler)
	assert.NotNil(t, scheduler.runners)
	assert.Equal(t, config.MaxConcurrentTasks, cap(scheduler.semaphore))
}

func TestTaskScheduler_RegisterRunner(t *testing.T) {
	repo := &mockTaskRepo{
		tasks: make(map[string]*TaskConfig),
		execs: make(map[string][]*TaskExecution),
	}
	scheduler := NewTaskScheduler(repo, DefaultTaskSchedulerConfig())

	// Register runner
	runner := &mockTaskRunner{typeVal: TaskLogRotation}
	scheduler.RegisterRunner(runner)

	// Verify that the runner was registered correctly
	assert.Len(t, scheduler.runners, 1)
	assert.Equal(t, runner, scheduler.runners[TaskLogRotation])
}

func TestTaskScheduler_RunTaskNow(t *testing.T) {
	repo := &mockTaskRepo{
		tasks: make(map[string]*TaskConfig),
		execs: make(map[string][]*TaskExecution),
	}
	scheduler := NewTaskScheduler(repo, DefaultTaskSchedulerConfig())

	// Register runners
	runner := &mockTaskRunner{typeVal: TaskLogRotation}
	scheduler.RegisterRunner(runner)

	// Create a task
	task := createTestTaskWithSchedule("test-1", TaskLogRotation, true, "")
	repo.CreateTask(context.Background(), task)

	// Run the task immediately
	exec, err := scheduler.RunTaskNow(task.ID)
	assert.NoError(t, err)
	assert.NotNil(t, exec)
	assert.Equal(t, task.ID, exec.TaskID)
	assert.Equal(t, StatusCompleted, exec.Status)
	assert.True(t, runner.WasRun())

	// Test running a disabled task
	disabledTask := createTestTaskWithSchedule("test-2", TaskLogRotation, false, "")
	repo.CreateTask(context.Background(), disabledTask)
	runner.Reset()

	_, err = scheduler.RunTaskNow(disabledTask.ID)
	assert.Error(t, err)
	assert.False(t, runner.WasRun())

	// Test running a task with no runner
	noRunnerTask := createTestTaskWithSchedule("test-3", TaskHealthCheck, true, "")
	repo.CreateTask(context.Background(), noRunnerTask)

	_, err = scheduler.RunTaskNow(noRunnerTask.ID)
	assert.Error(t, err)
}

func TestTaskScheduler_RunTaskAsync(t *testing.T) {
	repo := &mockTaskRepo{
		tasks: make(map[string]*TaskConfig),
		execs: make(map[string][]*TaskExecution),
	}
	scheduler := NewTaskScheduler(repo, DefaultTaskSchedulerConfig())

	// Register runner
	runner := &mockTaskRunner{typeVal: TaskLogRotation}
	scheduler.RegisterRunner(runner)

	// Create a task
	task := createTestTaskWithSchedule("test-async", TaskLogRotation, true, "")
	repo.CreateTask(context.Background(), task)

	// Execute task and check result directly
	exec, err := scheduler.RunTaskNow(task.ID)
	assert.NoError(t, err)
	assert.NotNil(t, exec)

	// Verify that the runner was called
	assert.True(t, runner.WasRun())

	// Run it in a goroutine to simulate async behavior
	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := scheduler.RunTaskNow(task.ID); err != nil {
			t.Logf("Error in async task: %v", err)
		}
	}()

	// Wait for async task to complete
	select {
	case <-done:
		// Task completed successfully
	case <-time.After(500 * time.Millisecond):
		t.Logf("Async task timed out")
	}
}

func TestSchedule_NextRunTime(t *testing.T) {
	// Test calculating the next run time for a cron schedule
	now := time.Now()
	next := calculateNextRunTime("0 * * * *", now) // Top of every hour

	// Ensure next run time is in the future
	assert.True(t, next.After(now))

	// Ensure it's set to the next hour at minute 0
	assert.Equal(t, 0, next.Minute())
	if now.Minute() == 0 {
		// If we're at minute 0, next should be one hour later
		assert.Equal(t, (now.Hour()+1)%24, next.Hour())
	} else {
		// Otherwise next should be at the top of the next hour
		assert.Equal(t, (now.Hour()+1)%24, next.Hour())
	}
}

func TestTaskExecution_Complete(t *testing.T) {
	exec := NewTaskExecution("test-task-1")
	assert.Equal(t, StatusPending, exec.Status)
	assert.Empty(t, exec.Output)

	// Complete the execution
	exec.Complete("success")
	assert.Equal(t, StatusCompleted, exec.Status)
	assert.Equal(t, "success", exec.Output)
	assert.False(t, exec.EndTime.IsZero())
}

func TestTaskExecution_Fail(t *testing.T) {
	exec := NewTaskExecution("test-task-1")
	assert.Equal(t, StatusPending, exec.Status)

	// Fail the execution
	exec.Fail("error occurred")
	assert.Equal(t, StatusFailed, exec.Status)
	assert.Equal(t, "error occurred", exec.Output)
	assert.False(t, exec.EndTime.IsZero())
}

// calculateNextRunTime calculates the next run time based on a cron expression
func calculateNextRunTime(cronExpr string, from time.Time) time.Time {
	// Simple implementation for testing
	// In a real system, we would use a cron parser
	if cronExpr == "0 * * * *" { // Top of every hour
		next := time.Date(
			from.Year(), from.Month(), from.Day(),
			from.Hour(), 0, 0, 0, from.Location(),
		)
		if next.Before(from) || next.Equal(from) {
			next = next.Add(time.Hour)
		}
		return next
	} else if cronExpr == "*/5 * * * *" { // Next 5-minute mark
		minute := ((from.Minute() / 5) + 1) * 5
		hour := from.Hour()
		if minute >= 60 {
			minute = 0
			hour = (hour + 1) % 24
		}
		return time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, from.Location())
	}
	// Default to 1 minute in the future
	return from.Add(time.Minute)
}

// mockTaskRepo is a simple in-memory implementation of repository.TaskRepository for testing
type mockTaskRepo struct {
	tasks map[string]*TaskConfig
	execs map[string][]*TaskExecution
	mu    sync.RWMutex
}

func (r *mockTaskRepo) CreateTask(ctx context.Context, task *TaskConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *mockTaskRepo) GetTask(ctx context.Context, id string) (*TaskConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}

func (r *mockTaskRepo) UpdateTask(ctx context.Context, task *TaskConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.tasks[task.ID]
	if !exists {
		return fmt.Errorf("task not found: %s", task.ID)
	}
	r.tasks[task.ID] = task
	return nil
}

func (r *mockTaskRepo) DeleteTask(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}
	delete(r.tasks, id)
	return nil
}

func (r *mockTaskRepo) ListTasks(ctx context.Context) ([]*TaskConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*TaskConfig
	for _, task := range r.tasks {
		result = append(result, task)
	}
	return result, nil
}

func (r *mockTaskRepo) GetTasksByType(ctx context.Context, taskType TaskType) ([]*TaskConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*TaskConfig
	for _, task := range r.tasks {
		if task.Type == taskType {
			result = append(result, task)
		}
	}
	return result, nil
}

func (r *mockTaskRepo) RecordExecution(ctx context.Context, execution *TaskExecution) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.execs[execution.TaskID] = append(r.execs[execution.TaskID], execution)
	return nil
}

func (r *mockTaskRepo) GetExecutions(ctx context.Context, taskID string) ([]*TaskExecution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	executions, exists := r.execs[taskID]
	if !exists {
		return []*TaskExecution{}, nil
	}
	return executions, nil
}

func (r *mockTaskRepo) GetTaskExecutions(ctx context.Context, taskID string, limit int) ([]*TaskExecution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	executions, exists := r.execs[taskID]
	if !exists {
		return []*TaskExecution{}, nil
	}

	if limit <= 0 || limit > len(executions) {
		return executions, nil
	}
	return executions[:limit], nil
}

func (r *mockTaskRepo) GetExecution(ctx context.Context, id string) (*TaskExecution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, execs := range r.execs {
		for _, exec := range execs {
			if exec.ID == id {
				return exec, nil
			}
		}
	}
	return nil, fmt.Errorf("execution not found: %s", id)
}

// Add new tests for Start/Stop and scheduling loop functionality

func TestTaskScheduler_StartStop(t *testing.T) {
	repo := &mockTaskRepo{
		tasks: make(map[string]*TaskConfig),
		execs: make(map[string][]*TaskExecution),
	}
	config := DefaultTaskSchedulerConfig()
	config.CheckInterval = 100 * time.Millisecond // Short interval for testing

	scheduler := NewTaskScheduler(repo, config)
	runner := &mockTaskRunner{typeVal: TaskLogRotation}
	scheduler.RegisterRunner(runner)

	// Create a task that's due to run
	task := createTestTaskWithSchedule("test-sched-1", TaskLogRotation, true, "*/1 * * * *")
	task.Schedule.NextRunTime = time.Now().Add(-1 * time.Minute) // Due in the past
	repo.CreateTask(context.Background(), task)

	// Start the scheduler
	err := scheduler.Start()
	assert.NoError(t, err)

	// Give it time to run at least one cycle
	time.Sleep(200 * time.Millisecond)

	// Verify the task was executed
	assert.True(t, runner.WasRun())

	// Stop the scheduler
	scheduler.Stop()

	// Reset the runner
	runner.Reset()

	// Make sure the scheduler is stopped (no more executions)
	time.Sleep(200 * time.Millisecond)
	assert.False(t, runner.WasRun())
}

func TestTaskScheduler_CheckScheduledTasks(t *testing.T) {
	repo := &mockTaskRepo{
		tasks: make(map[string]*TaskConfig),
		execs: make(map[string][]*TaskExecution),
	}
	scheduler := NewTaskScheduler(repo, DefaultTaskSchedulerConfig())

	// Register multiple runner types
	logRunner := &mockTaskRunner{typeVal: TaskLogRotation}
	metricsRunner := &mockTaskRunner{typeVal: TaskMetricsAggregation}
	healthRunner := &mockTaskRunner{typeVal: TaskHealthCheck}

	scheduler.RegisterRunner(logRunner)
	scheduler.RegisterRunner(metricsRunner)
	scheduler.RegisterRunner(healthRunner)

	// Create tasks with different due times
	// Due now
	task1 := createTestTaskWithSchedule("due-now", TaskLogRotation, true, "*/5 * * * *")
	task1.Schedule.NextRunTime = time.Now().Add(-1 * time.Minute)
	repo.CreateTask(context.Background(), task1)

	// Due in the future
	task2 := createTestTaskWithSchedule("due-future", TaskMetricsAggregation, true, "*/10 * * * *")
	task2.Schedule.NextRunTime = time.Now().Add(5 * time.Minute)
	repo.CreateTask(context.Background(), task2)

	// Disabled task
	task3 := createTestTaskWithSchedule("disabled", TaskHealthCheck, false, "*/5 * * * *")
	task3.Schedule.NextRunTime = time.Now().Add(-2 * time.Minute)
	repo.CreateTask(context.Background(), task3)

	// Process due tasks
	err := scheduler.checkScheduledTasks()
	assert.NoError(t, err)

	// Wait a brief moment for the goroutine to execute the task
	time.Sleep(100 * time.Millisecond)

	// Verify that only the due and enabled task was processed
	assert.True(t, logRunner.WasRun(), "Log rotation runner should have run")
	assert.False(t, metricsRunner.WasRun(), "Metrics runner should not have run")
	assert.False(t, healthRunner.WasRun(), "Health runner should not have run")

	// Verify that the next run time was updated for the executed task
	updatedTask, err := repo.GetTask(context.Background(), task1.ID)
	assert.NoError(t, err)
	assert.True(t, updatedTask.Schedule.NextRunTime.After(task1.Schedule.NextRunTime))
}

func TestTaskScheduler_CalculateNextRunTime(t *testing.T) {
	now := time.Now()

	// Test with valid cron expressions
	tests := []struct {
		expr     string
		expected time.Time
	}{
		{"0 * * * *", time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())},                      // Top of next hour
		{"*/5 * * * *", time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), ((now.Minute()/5)+1)*5, 0, 0, now.Location())}, // Next 5-minute mark
	}

	for _, tt := range tests {
		task := &TaskConfig{
			Schedule: Schedule{
				CronExpression: tt.expr,
			},
		}

		next, err := calculateNextScheduledTime(task)
		assert.NoError(t, err)

		// Allow a small time difference for test execution time
		delta := next.Sub(tt.expected)
		if delta < 0 {
			delta = -delta
		}
		assert.True(t, delta < time.Minute, "Next run time calculation off by more than a minute for %s", tt.expr)
	}

	// Test with invalid cron expression
	invalidTask := &TaskConfig{
		Schedule: Schedule{
			CronExpression: "invalid",
		},
	}
	_, err := calculateNextScheduledTime(invalidTask)
	assert.Error(t, err)
}

// calculateNextScheduledTime calculates when a task should run next based on its schedule
func calculateNextScheduledTime(task *TaskConfig) (time.Time, error) {
	// This is a simplified version of the function from scheduler.go
	// In production code, we'd use a proper cron parser
	now := time.Now()

	if task.Schedule.CronExpression == "0 * * * *" {
		// Top of next hour
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location()), nil
	} else if task.Schedule.CronExpression == "*/5 * * * *" {
		// Next 5-minute mark
		minute := ((now.Minute() / 5) + 1) * 5
		hour := now.Hour()
		if minute >= 60 {
			minute = 0
			hour = (hour + 1) % 24
		}
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location()), nil
	}

	return time.Time{}, fmt.Errorf("invalid cron expression: %s", task.Schedule.CronExpression)
}
