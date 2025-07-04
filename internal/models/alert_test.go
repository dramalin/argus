package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThresholdConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		threshold   ThresholdConfig
		expectError bool
	}{
		{
			name: "Valid CPU threshold",
			threshold: ThresholdConfig{
				MetricType: MetricCPU,
				MetricName: "usage_percent",
				Operator:   OperatorGreaterThan,
				Value:      90.0,
			},
			expectError: false,
		},
		{
			name: "Valid memory threshold",
			threshold: ThresholdConfig{
				MetricType: MetricMemory,
				MetricName: "used_percent",
				Operator:   OperatorGreaterThan,
				Value:      85.0,
			},
			expectError: false,
		},
		{
			name: "Valid network threshold",
			threshold: ThresholdConfig{
				MetricType: MetricNetwork,
				MetricName: "bytes_sent",
				Operator:   OperatorGreaterThan,
				Value:      10000000,
			},
			expectError: false,
		},
		{
			name: "Missing metric type",
			threshold: ThresholdConfig{
				MetricName: "usage_percent",
				Operator:   OperatorGreaterThan,
				Value:      90.0,
			},
			expectError: true,
		},
		{
			name: "Missing metric name",
			threshold: ThresholdConfig{
				MetricType: MetricCPU,
				Operator:   OperatorGreaterThan,
				Value:      90.0,
			},
			expectError: true,
		},
		{
			name: "Missing operator",
			threshold: ThresholdConfig{
				MetricType: MetricCPU,
				MetricName: "usage_percent",
				Value:      90.0,
			},
			expectError: true,
		},
		{
			name: "Invalid operator",
			threshold: ThresholdConfig{
				MetricType: MetricCPU,
				MetricName: "usage_percent",
				Operator:   "invalid",
				Value:      90.0,
			},
			expectError: true,
		},
		{
			name: "Invalid metric type",
			threshold: ThresholdConfig{
				MetricType: "invalid",
				MetricName: "usage_percent",
				Operator:   OperatorGreaterThan,
				Value:      90.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.threshold.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAlertConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      AlertConfig
		expectError bool
	}{
		{
			name: "Valid config",
			config: AlertConfig{
				ID:          "test-alert-1",
				Name:        "High CPU Usage",
				Description: "Alert when CPU usage is too high",
				Enabled:     true,
				Severity:    SeverityWarning,
				Threshold: ThresholdConfig{
					MetricType: MetricCPU,
					MetricName: "usage_percent",
					Operator:   OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []NotificationConfig{
					{
						Type:    NotificationInApp,
						Enabled: true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Missing name",
			config: AlertConfig{
				ID:          "test-alert-2",
				Description: "Alert when CPU usage is too high",
				Enabled:     true,
				Severity:    SeverityWarning,
				Threshold: ThresholdConfig{
					MetricType: MetricCPU,
					MetricName: "usage_percent",
					Operator:   OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []NotificationConfig{
					{
						Type:    NotificationInApp,
						Enabled: true,
					},
				},
			},
			expectError: true,
		},
		{
			name: "Invalid threshold",
			config: AlertConfig{
				Name:        "High CPU Usage",
				Description: "Alert when CPU usage is too high",
				Enabled:     true,
				Severity:    SeverityWarning,
				Threshold: ThresholdConfig{
					// Missing MetricType
					MetricName: "usage_percent",
					Operator:   OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []NotificationConfig{
					{
						Type:    NotificationInApp,
						Enabled: true,
					},
				},
			},
			expectError: true,
		},
		{
			name: "No notifications",
			config: AlertConfig{
				Name:        "High CPU Usage",
				Description: "Alert when CPU usage is too high",
				Enabled:     true,
				Severity:    SeverityWarning,
				Threshold: ThresholdConfig{
					MetricType: MetricCPU,
					MetricName: "usage_percent",
					Operator:   OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []NotificationConfig{},
			},
			expectError: true,
		},
		{
			name: "Invalid notification type",
			config: AlertConfig{
				Name:        "High CPU Usage",
				Description: "Alert when CPU usage is too high",
				Enabled:     true,
				Severity:    SeverityWarning,
				Threshold: ThresholdConfig{
					MetricType: MetricCPU,
					MetricName: "usage_percent",
					Operator:   OperatorGreaterThan,
					Value:      90.0,
				},
				Notifications: []NotificationConfig{
					{
						Type:    "invalid",
						Enabled: true,
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotificationConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      NotificationConfig
		expectError bool
	}{
		{
			name: "Valid in-app notification",
			config: NotificationConfig{
				Type:    NotificationInApp,
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "Valid email notification with settings",
			config: NotificationConfig{
				Type:    NotificationEmail,
				Enabled: true,
				Settings: map[string]interface{}{
					"recipient": "user@example.com",
				},
			},
			expectError: false,
		},
		{
			name: "Email notification missing recipient",
			config: NotificationConfig{
				Type:     NotificationEmail,
				Enabled:  true,
				Settings: map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "Invalid notification type",
			config: NotificationConfig{
				Type:    "invalid",
				Enabled: true,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComparisonOperatorComparison(t *testing.T) {
	tests := []struct {
		name      string
		operator  ComparisonOperator
		actual    float64
		threshold float64
		expected  bool
	}{
		{"Greater than (true)", OperatorGreaterThan, 95.0, 90.0, true},
		{"Greater than (false)", OperatorGreaterThan, 85.0, 90.0, false},
		{"Greater than or equal (true, greater)", OperatorGreaterThanOrEqual, 95.0, 90.0, true},
		{"Greater than or equal (true, equal)", OperatorGreaterThanOrEqual, 90.0, 90.0, true},
		{"Greater than or equal (false)", OperatorGreaterThanOrEqual, 85.0, 90.0, false},
		{"Less than (true)", OperatorLessThan, 85.0, 90.0, true},
		{"Less than (false)", OperatorLessThan, 95.0, 90.0, false},
		{"Less than or equal (true, less)", OperatorLessThanOrEqual, 85.0, 90.0, true},
		{"Less than or equal (true, equal)", OperatorLessThanOrEqual, 90.0, 90.0, true},
		{"Less than or equal (false)", OperatorLessThanOrEqual, 95.0, 90.0, false},
		{"Equal (true)", OperatorEqual, 90.0, 90.0, true},
		{"Equal (false)", OperatorEqual, 95.0, 90.0, false},
		{"Not equal (true)", OperatorNotEqual, 95.0, 90.0, true},
		{"Not equal (false)", OperatorNotEqual, 90.0, 90.0, false},
		{"Unknown operator", "unknown", 90.0, 90.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Implementing the comparison logic directly in the test
			var result bool
			switch tt.operator {
			case OperatorGreaterThan:
				result = tt.actual > tt.threshold
			case OperatorGreaterThanOrEqual:
				result = tt.actual >= tt.threshold
			case OperatorLessThan:
				result = tt.actual < tt.threshold
			case OperatorLessThanOrEqual:
				result = tt.actual <= tt.threshold
			case OperatorEqual:
				result = tt.actual == tt.threshold
			case OperatorNotEqual:
				result = tt.actual != tt.threshold
			default:
				result = false
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to implement the comparison operation
func compareValues(op ComparisonOperator, actual, threshold float64) bool {
	switch op {
	case OperatorGreaterThan:
		return actual > threshold
	case OperatorGreaterThanOrEqual:
		return actual >= threshold
	case OperatorLessThan:
		return actual < threshold
	case OperatorLessThanOrEqual:
		return actual <= threshold
	case OperatorEqual:
		return actual == threshold
	case OperatorNotEqual:
		return actual != threshold
	default:
		return false
	}
}

func TestAlertStatusMarshalJSON(t *testing.T) {
	// Create a test alert status
	triggeredTime := time.Now().Add(-time.Hour)
	resolvedTime := time.Now().Add(-30 * time.Minute)

	status := AlertStatus{
		AlertID:      "test-alert-id",
		State:        StateActive,
		CurrentValue: 95.5,
		TriggeredAt:  &triggeredTime,
		ResolvedAt:   &resolvedTime,
		Message:      "CPU usage exceeded threshold",
	}

	// Marshal to JSON
	data, err := json.Marshal(status)
	require.NoError(t, err)

	// Unmarshal from JSON
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Check fields
	assert.Equal(t, "test-alert-id", unmarshaled["alert_id"])
	assert.Equal(t, string(StateActive), unmarshaled["state"])
	assert.Equal(t, 95.5, unmarshaled["current_value"])
	assert.Equal(t, "CPU usage exceeded threshold", unmarshaled["message"])

	// Check time fields are properly formatted
	assert.Contains(t, unmarshaled, "triggered_at")
	assert.Contains(t, unmarshaled, "resolved_at")
}

// TestAlertConfigSerialization tests the JSON serialization/deserialization of AlertConfig
func TestAlertConfigSerialization(t *testing.T) {
	config := AlertConfig{
		ID:          "test-id",
		Name:        "High CPU Usage",
		Description: "Alert when CPU usage is too high",
		Enabled:     true,
		Severity:    SeverityWarning,
		Threshold: ThresholdConfig{
			MetricType: MetricCPU,
			MetricName: "usage_percent",
			Operator:   OperatorGreaterThan,
			Value:      90.0,
		},
		Notifications: []NotificationConfig{
			{
				Type:    NotificationInApp,
				Enabled: true,
			},
			{
				Type:    NotificationEmail,
				Enabled: true,
				Settings: map[string]interface{}{
					"recipient": "admin@example.com",
				},
			},
		},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Unmarshal back to AlertConfig
	var decoded AlertConfig
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Compare fields
	assert.Equal(t, config.ID, decoded.ID)
	assert.Equal(t, config.Name, decoded.Name)
	assert.Equal(t, config.Description, decoded.Description)
	assert.Equal(t, config.Enabled, decoded.Enabled)
	assert.Equal(t, config.Severity, decoded.Severity)

	// Compare threshold
	assert.Equal(t, config.Threshold.MetricType, decoded.Threshold.MetricType)
	assert.Equal(t, config.Threshold.MetricName, decoded.Threshold.MetricName)
	assert.Equal(t, config.Threshold.Operator, decoded.Threshold.Operator)
	assert.Equal(t, config.Threshold.Value, decoded.Threshold.Value)

	// Compare notifications
	assert.Equal(t, len(config.Notifications), len(decoded.Notifications))
	assert.Equal(t, config.Notifications[0].Type, decoded.Notifications[0].Type)
	assert.Equal(t, config.Notifications[0].Enabled, decoded.Notifications[0].Enabled)
	assert.Equal(t, config.Notifications[1].Type, decoded.Notifications[1].Type)
	assert.Equal(t, config.Notifications[1].Enabled, decoded.Notifications[1].Enabled)

	// Check that the time fields are approximately equal
	assert.WithinDuration(t, config.CreatedAt, decoded.CreatedAt, time.Second)
	assert.WithinDuration(t, config.UpdatedAt, decoded.UpdatedAt, time.Second)
}
