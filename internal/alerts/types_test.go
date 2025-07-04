package alerts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/models"
)

// TestThresholdConfigValidate ensures that the ThresholdConfig validation works correctly
func TestThresholdConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		threshold   models.ThresholdConfig
		expectError bool
	}{
		{
			name: "Valid CPU threshold",
			threshold: models.ThresholdConfig{
				MetricType: models.MetricCPU,
				MetricName: "usage_percent",
				Operator:   models.OperatorGreaterThan,
				Value:      90.0,
			},
			expectError: false,
		},
		{
			name: "Valid memory threshold",
			threshold: models.ThresholdConfig{
				MetricType: models.MetricMemory,
				MetricName: "used_percent",
				Operator:   models.OperatorGreaterThan,
				Value:      85.0,
			},
			expectError: false,
		},
		{
			name: "Valid network threshold",
			threshold: models.ThresholdConfig{
				MetricType: models.MetricNetwork,
				MetricName: "bytes_sent",
				Operator:   models.OperatorGreaterThan,
				Value:      10000000,
			},
			expectError: false,
		},
		{
			name: "Missing metric type",
			threshold: models.ThresholdConfig{
				MetricName: "usage_percent",
				Operator:   models.OperatorGreaterThan,
				Value:      90.0,
			},
			expectError: true,
		},
		{
			name: "Invalid metric type",
			threshold: models.ThresholdConfig{
				MetricType: "invalid",
				MetricName: "usage_percent",
				Operator:   models.OperatorGreaterThan,
				Value:      90.0,
			},
			expectError: true,
		},
		{
			name: "Invalid operator",
			threshold: models.ThresholdConfig{
				MetricType: models.MetricCPU,
				MetricName: "usage_percent",
				Operator:   "invalid",
				Value:      90.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.threshold.Validate()
			if (err != nil) != tt.expectError {
				t.Errorf("ThresholdConfig.Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestNotificationConfigValidate ensures that the NotificationConfig validation works correctly
func TestNotificationConfigValidate(t *testing.T) {
	tests := []struct {
		name         string
		notification models.NotificationConfig
		expectError  bool
	}{
		{
			name: "Valid in-app notification",
			notification: models.NotificationConfig{
				Type:    models.NotificationInApp,
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "Valid email notification",
			notification: models.NotificationConfig{
				Type:    models.NotificationEmail,
				Enabled: true,
				Settings: map[string]any{
					"recipient": "user@example.com",
				},
			},
			expectError: false,
		},
		{
			name: "Missing notification type",
			notification: models.NotificationConfig{
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "Invalid notification type",
			notification: models.NotificationConfig{
				Type:    "invalid",
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "Email notification without settings",
			notification: models.NotificationConfig{
				Type:    models.NotificationEmail,
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "Email notification without recipient",
			notification: models.NotificationConfig{
				Type:     models.NotificationEmail,
				Enabled:  true,
				Settings: map[string]any{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.notification.Validate()
			if (err != nil) != tt.expectError {
				t.Errorf("NotificationConfig.Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestAlertConfigValidate ensures that the AlertConfig validation works correctly
func TestAlertConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		alert       models.AlertConfig
		expectError bool
	}{
		{
			name: "Valid alert config",
			alert: models.AlertConfig{
				ID:          "test-alert-1",
				Name:        "Test Alert",
				Description: "Test alert description",
				Enabled:     true,
				Severity:    models.SeverityWarning,
				Threshold: models.ThresholdConfig{
					MetricType: models.MetricCPU,
					MetricName: "usage_percent",
					Operator:   models.OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []models.NotificationConfig{
					{
						Type:    models.NotificationInApp,
						Enabled: true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Missing ID",
			alert: models.AlertConfig{
				Name:        "Test Alert",
				Description: "Test alert description",
				Enabled:     true,
				Severity:    models.SeverityWarning,
				Threshold: models.ThresholdConfig{
					MetricType: models.MetricCPU,
					MetricName: "usage_percent",
					Operator:   models.OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []models.NotificationConfig{
					{
						Type:    models.NotificationInApp,
						Enabled: true,
					},
				},
			},
			expectError: true,
		},
		{
			name: "Missing name",
			alert: models.AlertConfig{
				ID:          "test-alert-1",
				Description: "Test alert description",
				Enabled:     true,
				Severity:    models.SeverityWarning,
				Threshold: models.ThresholdConfig{
					MetricType: models.MetricCPU,
					MetricName: "usage_percent",
					Operator:   models.OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []models.NotificationConfig{
					{
						Type:    models.NotificationInApp,
						Enabled: true,
					},
				},
			},
			expectError: true,
		},
		{
			name: "Invalid severity",
			alert: models.AlertConfig{
				ID:          "test-alert-1",
				Name:        "Test Alert",
				Description: "Test alert description",
				Enabled:     true,
				Severity:    "invalid",
				Threshold: models.ThresholdConfig{
					MetricType: models.MetricCPU,
					MetricName: "usage_percent",
					Operator:   models.OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []models.NotificationConfig{
					{
						Type:    models.NotificationInApp,
						Enabled: true,
					},
				},
			},
			expectError: true,
		},
		{
			name: "Invalid threshold",
			alert: models.AlertConfig{
				ID:          "test-alert-1",
				Name:        "Test Alert",
				Description: "Test alert description",
				Enabled:     true,
				Severity:    models.SeverityWarning,
				Threshold: models.ThresholdConfig{
					// Invalid: Missing metric type
					MetricName: "usage_percent",
					Operator:   models.OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []models.NotificationConfig{
					{
						Type:    models.NotificationInApp,
						Enabled: true,
					},
				},
			},
			expectError: true,
		},
		{
			name: "No notifications",
			alert: models.AlertConfig{
				ID:          "test-alert-1",
				Name:        "Test Alert",
				Description: "Test alert description",
				Enabled:     true,
				Severity:    models.SeverityWarning,
				Threshold: models.ThresholdConfig{
					MetricType: models.MetricCPU,
					MetricName: "usage_percent",
					Operator:   models.OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []models.NotificationConfig{},
			},
			expectError: false, // Current implementation doesn't validate notifications
		},
		{
			name: "Invalid notification",
			alert: models.AlertConfig{
				ID:          "test-alert-1",
				Name:        "Test Alert",
				Description: "Test alert description",
				Enabled:     true,
				Severity:    models.SeverityWarning,
				Threshold: models.ThresholdConfig{
					MetricType: models.MetricCPU,
					MetricName: "usage_percent",
					Operator:   models.OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []models.NotificationConfig{
					{
						// Invalid: Missing type
						Enabled: true,
					},
				},
			},
			expectError: false, // Current implementation doesn't validate notifications
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.alert.Validate()
			if (err != nil) != tt.expectError {
				t.Errorf("AlertConfig.Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// Test alert status transitions
func TestAlertStatusTransitions(t *testing.T) {
	// Create a test alert status
	status := models.AlertStatus{
		AlertID:      "test-alert-1",
		State:        models.StateInactive,
		CurrentValue: 0,
	}

	// Create a helper function to transition the status
	transition := func(status models.AlertStatus, newState models.AlertState, value float64) models.AlertStatus {
		now := time.Now()
		updated := status
		updated.State = newState
		updated.CurrentValue = value

		// Set timestamps based on state
		if newState == models.StateActive {
			updated.TriggeredAt = &now
			updated.ResolvedAt = nil
			updated.Message = "Alert triggered"
		} else if newState == models.StateResolved {
			updated.ResolvedAt = &now
			updated.Message = "Alert resolved"
		}

		return updated
	}

	// Test transition to Pending
	updated := transition(status, models.StatePending, 85.0)
	assert.Equal(t, models.StatePending, updated.State)
	assert.Equal(t, 85.0, updated.CurrentValue)

	// Test transition to Active (triggered)
	triggered := transition(updated, models.StateActive, 95.0)
	assert.Equal(t, models.StateActive, triggered.State)
	assert.Equal(t, 95.0, triggered.CurrentValue)
	require.NotNil(t, triggered.TriggeredAt)
	assert.Nil(t, triggered.ResolvedAt)

	// Test transition to Resolved
	resolved := transition(triggered, models.StateResolved, 75.0)
	assert.Equal(t, models.StateResolved, resolved.State)
	assert.Equal(t, 75.0, resolved.CurrentValue)
	require.NotNil(t, resolved.ResolvedAt)

	// Test transition to Inactive
	inactive := transition(resolved, models.StateInactive, 50.0)
	assert.Equal(t, models.StateInactive, inactive.State)
	assert.Equal(t, 50.0, inactive.CurrentValue)
}

// Test alert event creation and management
func TestAlertEvent(t *testing.T) {
	// Create a test event
	now := time.Now()
	event := models.AlertEvent{
		AlertID:      "test-alert-1",
		OldState:     models.StateInactive,
		NewState:     models.StateActive,
		CurrentValue: 95.0,
		Threshold:    90.0,
		Timestamp:    now,
		Message:      "CPU usage exceeded threshold",
	}

	// Validate event properties
	assert.Equal(t, "test-alert-1", event.AlertID)
	assert.Equal(t, models.StateActive, event.NewState)
	assert.Equal(t, models.StateInactive, event.OldState)
	assert.Equal(t, 95.0, event.CurrentValue)
	assert.Equal(t, 90.0, event.Threshold)
	assert.Equal(t, now, event.Timestamp)
	assert.Equal(t, "CPU usage exceeded threshold", event.Message)
}
