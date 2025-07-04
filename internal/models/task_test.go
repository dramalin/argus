package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		schedule    Schedule
		expectError bool
	}{
		{
			name: "Valid cron schedule",
			schedule: Schedule{
				CronExpression: "0 * * * *", // Every hour
				OneTime:        false,
				NextRunTime:    now,
			},
			expectError: false,
		},
		{
			name: "Valid one-time schedule",
			schedule: Schedule{
				CronExpression: "",
				OneTime:        true,
				NextRunTime:    now,
			},
			expectError: false,
		},
		{
			name: "Invalid: no cron and not one-time",
			schedule: Schedule{
				CronExpression: "",
				OneTime:        false,
				NextRunTime:    now,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schedule.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTaskType(t *testing.T) {
	tests := []struct {
		name     string
		taskType TaskType
		valid    bool
	}{
		{
			name:     "Valid log rotation task",
			taskType: TaskLogRotation,
			valid:    true,
		},
		{
			name:     "Valid metrics aggregation task",
			taskType: TaskMetricsAggregation,
			valid:    true,
		},
		{
			name:     "Valid health check task",
			taskType: TaskHealthCheck,
			valid:    true,
		},
		{
			name:     "Valid system cleanup task",
			taskType: TaskSystemCleanup,
			valid:    true,
		},
		{
			name:     "Invalid task type",
			taskType: "invalid_type",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if task type is one of the defined constants
			var found bool
			switch tt.taskType {
			case TaskLogRotation, TaskMetricsAggregation, TaskHealthCheck, TaskSystemCleanup:
				found = true
			}
			assert.Equal(t, tt.valid, found, "Task type validity check failed")
		})
	}
}

func TestTaskStatus(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		valid  bool
	}{
		{
			name:   "Valid pending status",
			status: StatusPending,
			valid:  true,
		},
		{
			name:   "Valid running status",
			status: StatusRunning,
			valid:  true,
		},
		{
			name:   "Valid completed status",
			status: StatusCompleted,
			valid:  true,
		},
		{
			name:   "Valid failed status",
			status: StatusFailed,
			valid:  true,
		},
		{
			name:   "Invalid status",
			status: "invalid_status",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if status is one of the defined constants
			var found bool
			switch tt.status {
			case StatusPending, StatusRunning, StatusCompleted, StatusFailed:
				found = true
			}
			assert.Equal(t, tt.valid, found, "Status validity check failed")
		})
	}
}

func TestTaskConfig(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		config      TaskConfig
		expectError bool
	}{
		{
			name: "Valid log rotation task config",
			config: TaskConfig{
				ID:          "test-task-1",
				Name:        "Log Rotation Task",
				Type:        TaskLogRotation,
				Description: "Daily log rotation task",
				Enabled:     true,
				Schedule: Schedule{
					CronExpression: "0 0 * * *", // Daily at midnight
					OneTime:        false,
					NextRunTime:    now,
				},
				Parameters: map[string]string{
					"retention_days": "30",
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectError: false,
		},
		{
			name: "Missing name",
			config: TaskConfig{
				Type: TaskLogRotation,
				Schedule: Schedule{
					CronExpression: "0 0 * * *",
					OneTime:        false,
					NextRunTime:    now,
				},
			},
			expectError: true,
		},
		{
			name: "Invalid schedule",
			config: TaskConfig{
				Name: "Invalid Schedule Task",
				Type: TaskLogRotation,
				Schedule: Schedule{
					CronExpression: "",
					OneTime:        false,
					NextRunTime:    now,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTaskExecution(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		wantErr    bool
		wantFailed bool
	}{
		{
			name:       "Valid execution with success",
			taskID:     "test-task-1",
			wantErr:    false,
			wantFailed: false,
		},
		{
			name:       "Valid execution with failure",
			taskID:     "test-task-2",
			wantErr:    false,
			wantFailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new execution
			execution := NewTaskExecution(tt.taskID)
			require.NotNil(t, execution)
			assert.Equal(t, tt.taskID, execution.TaskID)
			assert.Equal(t, StatusPending, execution.Status)

			// Test Start
			execution.Start()
			assert.Equal(t, StatusRunning, execution.Status)
			assert.False(t, execution.StartTime.IsZero())

			if tt.wantFailed {
				// Test Fail
				errMsg := "Task failed: timeout"
				execution.Fail(errMsg)
				assert.Equal(t, StatusFailed, execution.Status)
				assert.Equal(t, errMsg, execution.Error)
				assert.False(t, execution.EndTime.IsZero())
			} else {
				// Test Complete
				output := "Task completed successfully"
				execution.Complete(output)
				assert.Equal(t, StatusCompleted, execution.Status)
				assert.Equal(t, output, execution.Output)
				assert.False(t, execution.EndTime.IsZero())
			}
		})
	}
}
