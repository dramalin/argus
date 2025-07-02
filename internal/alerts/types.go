// Package alerts provides functionality for configuring and managing system alerts
package alerts

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// MetricType represents the type of system metric to monitor
type MetricType string

// Available metric types for alerting
const (
	MetricCPU     MetricType = "cpu"     // CPU usage percentage
	MetricMemory  MetricType = "memory"  // Memory usage percentage
	MetricLoad    MetricType = "load"    // System load average
	MetricNetwork MetricType = "network" // Network traffic
	MetricDisk    MetricType = "disk"    // Disk usage/IO (for future implementation)
	MetricProcess MetricType = "process" // Process specific metrics (for future implementation)
)

// ComparisonOperator defines how a threshold is compared to the actual value
type ComparisonOperator string

// Available comparison operators for thresholds
const (
	OperatorGreaterThan        ComparisonOperator = ">"  // Greater than
	OperatorGreaterThanOrEqual ComparisonOperator = ">=" // Greater than or equal to
	OperatorLessThan           ComparisonOperator = "<"  // Less than
	OperatorLessThanOrEqual    ComparisonOperator = "<=" // Less than or equal to
	OperatorEqual              ComparisonOperator = "==" // Equal to
	OperatorNotEqual           ComparisonOperator = "!=" // Not equal to
)

// AlertSeverity represents the importance/urgency of an alert
type AlertSeverity string

// Available alert severity levels
const (
	SeverityInfo     AlertSeverity = "info"     // Informational, lowest severity
	SeverityWarning  AlertSeverity = "warning"  // Warning, medium severity
	SeverityCritical AlertSeverity = "critical" // Critical, highest severity
)

// NotificationType represents the channel through which an alert is delivered
type NotificationType string

// Available notification channels
const (
	NotificationInApp NotificationType = "in-app" // In-application notification
	NotificationEmail NotificationType = "email"  // Email notification
)

// ThresholdConfig defines a threshold condition that triggers an alert
type ThresholdConfig struct {
	MetricType   MetricType         `json:"metric_type"`             // Type of metric to monitor
	MetricName   string             `json:"metric_name"`             // Specific metric name (e.g., "usage_percent" for CPU)
	Operator     ComparisonOperator `json:"operator"`                // Comparison operator
	Value        float64            `json:"value"`                   // Threshold value
	Duration     time.Duration      `json:"duration,omitempty"`      // Duration the condition must persist (optional)
	SustainedFor int                `json:"sustained_for,omitempty"` // Number of consecutive checks (optional alternative to Duration)
}

// Validate checks if the threshold configuration is valid
func (t *ThresholdConfig) Validate() error {
	// Check for required fields
	if t.MetricType == "" {
		return errors.New("metric type is required")
	}

	// Validate metric type
	validMetricTypes := map[MetricType]bool{
		MetricCPU:     true,
		MetricMemory:  true,
		MetricLoad:    true,
		MetricNetwork: true,
		MetricDisk:    true,
		MetricProcess: true,
	}

	if !validMetricTypes[t.MetricType] {
		return fmt.Errorf("invalid metric type: %s", t.MetricType)
	}

	// Validate operator
	validOperators := map[ComparisonOperator]bool{
		OperatorGreaterThan:        true,
		OperatorGreaterThanOrEqual: true,
		OperatorLessThan:           true,
		OperatorLessThanOrEqual:    true,
		OperatorEqual:              true,
		OperatorNotEqual:           true,
	}

	if !validOperators[t.Operator] {
		return fmt.Errorf("invalid operator: %s", t.Operator)
	}

	// Validate metric name based on metric type
	switch t.MetricType {
	case MetricCPU:
		if t.MetricName != "usage_percent" && t.MetricName != "load1" &&
			t.MetricName != "load5" && t.MetricName != "load15" {
			return fmt.Errorf("invalid CPU metric name: %s", t.MetricName)
		}
	case MetricMemory:
		if t.MetricName != "used_percent" && t.MetricName != "used" &&
			t.MetricName != "free" {
			return fmt.Errorf("invalid memory metric name: %s", t.MetricName)
		}
	case MetricNetwork:
		if t.MetricName != "bytes_sent" && t.MetricName != "bytes_recv" &&
			t.MetricName != "packets_sent" && t.MetricName != "packets_recv" {
			return fmt.Errorf("invalid network metric name: %s", t.MetricName)
		}
	}

	return nil
}

// NotificationConfig defines how an alert is delivered
type NotificationConfig struct {
	Type     NotificationType `json:"type"`               // Type of notification
	Enabled  bool             `json:"enabled"`            // Whether this notification is enabled
	Settings map[string]any   `json:"settings,omitempty"` // Channel-specific settings
}

// Validate checks if the notification configuration is valid
func (n *NotificationConfig) Validate() error {
	// Check for required fields
	if n.Type == "" {
		return errors.New("notification type is required")
	}

	// Validate notification type
	validTypes := map[NotificationType]bool{
		NotificationInApp: true,
		NotificationEmail: true,
	}

	if !validTypes[n.Type] {
		return fmt.Errorf("invalid notification type: %s", n.Type)
	}

	// Validate settings based on notification type
	switch n.Type {
	case NotificationEmail:
		if n.Settings == nil {
			return errors.New("email notification requires settings")
		}

		recipient, ok := n.Settings["recipient"]
		if !ok || recipient == "" {
			return errors.New("email notification requires recipient")
		}
	}

	return nil
}

// AlertConfig defines a complete alert configuration
type AlertConfig struct {
	ID            string               `json:"id"`                    // Unique identifier for the alert
	Name          string               `json:"name"`                  // Human-readable name
	Description   string               `json:"description,omitempty"` // Optional description
	Enabled       bool                 `json:"enabled"`               // Whether this alert is active
	Severity      AlertSeverity        `json:"severity"`              // Alert severity level
	Threshold     ThresholdConfig      `json:"threshold"`             // Threshold condition
	Notifications []NotificationConfig `json:"notifications"`         // Notification channels
	CreatedAt     time.Time            `json:"created_at"`            // Creation timestamp
	UpdatedAt     time.Time            `json:"updated_at"`            // Last update timestamp
}

// Validate checks if the alert configuration is valid
func (a *AlertConfig) Validate() error {
	// Check for required fields
	if a.ID == "" {
		return errors.New("alert ID is required")
	}

	if a.Name == "" {
		return errors.New("alert name is required")
	}

	// Validate severity
	validSeverities := map[AlertSeverity]bool{
		SeverityInfo:     true,
		SeverityWarning:  true,
		SeverityCritical: true,
	}

	if !validSeverities[a.Severity] {
		return fmt.Errorf("invalid severity: %s", a.Severity)
	}

	// Validate threshold
	if err := a.Threshold.Validate(); err != nil {
		return fmt.Errorf("invalid threshold: %w", err)
	}

	// Validate notifications
	if len(a.Notifications) == 0 {
		return errors.New("at least one notification channel is required")
	}

	for i, notification := range a.Notifications {
		if err := notification.Validate(); err != nil {
			return fmt.Errorf("invalid notification at index %d: %w", i, err)
		}
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling with validation
func (a *AlertConfig) MarshalJSON() ([]byte, error) {
	// Validate before marshaling
	if err := a.Validate(); err != nil {
		return nil, err
	}

	// Create a type alias to avoid infinite recursion
	type AlertConfigAlias AlertConfig
	return json.Marshal((*AlertConfigAlias)(a))
}

// UnmarshalJSON implements custom JSON unmarshaling with validation
func (a *AlertConfig) UnmarshalJSON(data []byte) error {
	// Create a type alias to avoid infinite recursion
	type AlertConfigAlias AlertConfig
	aux := (*AlertConfigAlias)(a)

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Validate after unmarshaling
	if err := a.Validate(); err != nil {
		return err
	}

	return nil
}

// AlertState represents the current state of an alert
type AlertState string

// Available alert states
const (
	StateInactive AlertState = "inactive" // Alert condition not met
	StateActive   AlertState = "active"   // Alert condition currently met
	StatePending  AlertState = "pending"  // Alert condition met but not for the required duration
)

// AlertStatus represents the current status of an alert
type AlertStatus struct {
	AlertID      string     `json:"alert_id"`               // Reference to the alert configuration
	State        AlertState `json:"state"`                  // Current state
	CurrentValue float64    `json:"current_value"`          // Current metric value
	TriggeredAt  *time.Time `json:"triggered_at,omitempty"` // When the alert was triggered (if active)
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`  // When the alert was resolved (if inactive)
	Message      string     `json:"message,omitempty"`      // Human-readable status message
}

// CreateDefaultAlertConfig creates a default alert configuration for the given metric type
func CreateDefaultAlertConfig(metricType MetricType, severity AlertSeverity) *AlertConfig {
	now := time.Now()

	var threshold ThresholdConfig
	var name, description string

	switch metricType {
	case MetricCPU:
		threshold = ThresholdConfig{
			MetricType: MetricCPU,
			MetricName: "usage_percent",
			Operator:   OperatorGreaterThan,
			Value:      90.0,
			Duration:   5 * time.Minute,
		}
		name = "High CPU Usage"
		description = "Alert when CPU usage exceeds 90% for 5 minutes"

	case MetricMemory:
		threshold = ThresholdConfig{
			MetricType: MetricMemory,
			MetricName: "used_percent",
			Operator:   OperatorGreaterThan,
			Value:      85.0,
			Duration:   5 * time.Minute,
		}
		name = "High Memory Usage"
		description = "Alert when memory usage exceeds 85% for 5 minutes"

	case MetricLoad:
		threshold = ThresholdConfig{
			MetricType: MetricLoad,
			MetricName: "load5",
			Operator:   OperatorGreaterThan,
			Value:      4.0,
			Duration:   10 * time.Minute,
		}
		name = "High System Load"
		description = "Alert when 5-minute load average exceeds 4.0 for 10 minutes"

	case MetricNetwork:
		threshold = ThresholdConfig{
			MetricType: MetricNetwork,
			MetricName: "bytes_sent",
			Operator:   OperatorGreaterThan,
			Value:      10000000, // 10MB/s
			Duration:   5 * time.Minute,
		}
		name = "High Network Usage"
		description = "Alert when network traffic exceeds 10MB/s for 5 minutes"
	}

	return &AlertConfig{
		ID:          fmt.Sprintf("default-%s-%d", metricType, now.Unix()),
		Name:        name,
		Description: description,
		Enabled:     true,
		Severity:    severity,
		Threshold:   threshold,
		Notifications: []NotificationConfig{
			{
				Type:    NotificationInApp,
				Enabled: true,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
