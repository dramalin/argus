package alerts

import (
	"encoding/json"
	"testing"
	"time"
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
			name: "Invalid metric type",
			threshold: ThresholdConfig{
				MetricType: "invalid",
				MetricName: "usage_percent",
				Operator:   OperatorGreaterThan,
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
			name: "Invalid CPU metric name",
			threshold: ThresholdConfig{
				MetricType: MetricCPU,
				MetricName: "invalid",
				Operator:   OperatorGreaterThan,
				Value:      90.0,
			},
			expectError: true,
		},
		{
			name: "Invalid memory metric name",
			threshold: ThresholdConfig{
				MetricType: MetricMemory,
				MetricName: "invalid",
				Operator:   OperatorGreaterThan,
				Value:      85.0,
			},
			expectError: true,
		},
		{
			name: "Invalid network metric name",
			threshold: ThresholdConfig{
				MetricType: MetricNetwork,
				MetricName: "invalid",
				Operator:   OperatorGreaterThan,
				Value:      10000000,
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

func TestNotificationConfigValidate(t *testing.T) {
	tests := []struct {
		name         string
		notification NotificationConfig
		expectError  bool
	}{
		{
			name: "Valid in-app notification",
			notification: NotificationConfig{
				Type:    NotificationInApp,
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "Valid email notification",
			notification: NotificationConfig{
				Type:    NotificationEmail,
				Enabled: true,
				Settings: map[string]any{
					"recipient": "user@example.com",
				},
			},
			expectError: false,
		},
		{
			name: "Missing notification type",
			notification: NotificationConfig{
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "Invalid notification type",
			notification: NotificationConfig{
				Type:    "invalid",
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "Email notification without settings",
			notification: NotificationConfig{
				Type:    NotificationEmail,
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "Email notification without recipient",
			notification: NotificationConfig{
				Type:     NotificationEmail,
				Enabled:  true,
				Settings: map[string]any{},
			},
			expectError: true,
		},
		{
			name: "Email notification with empty recipient",
			notification: NotificationConfig{
				Type:    NotificationEmail,
				Enabled: true,
				Settings: map[string]any{
					"recipient": "",
				},
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

func TestAlertConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		alert       AlertConfig
		expectError bool
	}{
		{
			name: "Valid alert config",
			alert: AlertConfig{
				ID:       "test-alert",
				Name:     "Test Alert",
				Enabled:  true,
				Severity: SeverityWarning,
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
			name: "Missing ID",
			alert: AlertConfig{
				Name:     "Test Alert",
				Enabled:  true,
				Severity: SeverityWarning,
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
			name: "Missing name",
			alert: AlertConfig{
				ID:       "test-alert",
				Enabled:  true,
				Severity: SeverityWarning,
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
			name: "Invalid severity",
			alert: AlertConfig{
				ID:       "test-alert",
				Name:     "Test Alert",
				Enabled:  true,
				Severity: "invalid",
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
			alert: AlertConfig{
				ID:       "test-alert",
				Name:     "Test Alert",
				Enabled:  true,
				Severity: SeverityWarning,
				Threshold: ThresholdConfig{
					MetricType: "invalid",
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
			alert: AlertConfig{
				ID:       "test-alert",
				Name:     "Test Alert",
				Enabled:  true,
				Severity: SeverityWarning,
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
			name: "Invalid notification",
			alert: AlertConfig{
				ID:       "test-alert",
				Name:     "Test Alert",
				Enabled:  true,
				Severity: SeverityWarning,
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
			err := tt.alert.Validate()
			if (err != nil) != tt.expectError {
				t.Errorf("AlertConfig.Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestAlertConfigJSON(t *testing.T) {
	alert := AlertConfig{
		ID:          "test-alert",
		Name:        "Test Alert",
		Description: "This is a test alert",
		Enabled:     true,
		Severity:    SeverityWarning,
		Threshold: ThresholdConfig{
			MetricType: MetricCPU,
			MetricName: "usage_percent",
			Operator:   OperatorGreaterThan,
			Value:      90.0,
			Duration:   5 * time.Minute,
		},
		Notifications: []NotificationConfig{
			{
				Type:    NotificationInApp,
				Enabled: true,
			},
			{
				Type:    NotificationEmail,
				Enabled: true,
				Settings: map[string]any{
					"recipient": "user@example.com",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test marshaling
	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("Failed to marshal alert: %v", err)
	}

	// Test unmarshaling
	var unmarshaledAlert AlertConfig
	err = json.Unmarshal(data, &unmarshaledAlert)
	if err != nil {
		t.Fatalf("Failed to unmarshal alert: %v", err)
	}

	// Check key fields
	if unmarshaledAlert.ID != alert.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaledAlert.ID, alert.ID)
	}
	if unmarshaledAlert.Name != alert.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaledAlert.Name, alert.Name)
	}
	if unmarshaledAlert.Description != alert.Description {
		t.Errorf("Description mismatch: got %s, want %s", unmarshaledAlert.Description, alert.Description)
	}
	if unmarshaledAlert.Enabled != alert.Enabled {
		t.Errorf("Enabled mismatch: got %v, want %v", unmarshaledAlert.Enabled, alert.Enabled)
	}
	if unmarshaledAlert.Severity != alert.Severity {
		t.Errorf("Severity mismatch: got %s, want %s", unmarshaledAlert.Severity, alert.Severity)
	}
	if unmarshaledAlert.Threshold.MetricType != alert.Threshold.MetricType {
		t.Errorf("Threshold.MetricType mismatch: got %s, want %s", unmarshaledAlert.Threshold.MetricType, alert.Threshold.MetricType)
	}
	if len(unmarshaledAlert.Notifications) != len(alert.Notifications) {
		t.Errorf("Notifications length mismatch: got %d, want %d", len(unmarshaledAlert.Notifications), len(alert.Notifications))
	}
}

func TestCreateDefaultAlertConfig(t *testing.T) {
	tests := []struct {
		name       string
		metricType MetricType
		severity   AlertSeverity
	}{
		{
			name:       "CPU default config",
			metricType: MetricCPU,
			severity:   SeverityWarning,
		},
		{
			name:       "Memory default config",
			metricType: MetricMemory,
			severity:   SeverityCritical,
		},
		{
			name:       "Load default config",
			metricType: MetricLoad,
			severity:   SeverityInfo,
		},
		{
			name:       "Network default config",
			metricType: MetricNetwork,
			severity:   SeverityWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CreateDefaultAlertConfig(tt.metricType, tt.severity)

			// Validate the config
			err := config.Validate()
			if err != nil {
				t.Errorf("CreateDefaultAlertConfig() produced invalid config: %v", err)
			}

			// Check that the metric type and severity match
			if config.Threshold.MetricType != tt.metricType {
				t.Errorf("CreateDefaultAlertConfig() metric type = %s, want %s", config.Threshold.MetricType, tt.metricType)
			}
			if config.Severity != tt.severity {
				t.Errorf("CreateDefaultAlertConfig() severity = %s, want %s", config.Severity, tt.severity)
			}

			// Check that the ID is properly formatted
			if config.ID == "" {
				t.Error("CreateDefaultAlertConfig() ID is empty")
			}
		})
	}
}
