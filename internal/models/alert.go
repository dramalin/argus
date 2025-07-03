// File: internal/models/alert.go
// Brief: Alert-related data models for Argus
// Detailed: Contains type definitions for MetricType, ComparisonOperator, AlertSeverity, NotificationType, ThresholdConfig, NotificationConfig, AlertConfig, AlertState, AlertStatus, and related constants/methods.
// Author: Argus Migration (AI)
// Date: 2024-07-03

package models

import (
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
	MetricType   MetricType         `json:"metric_type"`
	MetricName   string             `json:"metric_name"`
	Operator     ComparisonOperator `json:"operator"`
	Value        float64            `json:"value"`
	Duration     time.Duration      `json:"duration,omitempty"`
	SustainedFor int                `json:"sustained_for,omitempty"`
}

// Validate checks if the threshold configuration is valid
func (t *ThresholdConfig) Validate() error {
	if t.MetricType == "" {
		return errors.New("metric type is required")
	}
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
	// Validate metric name based on metric type (partial, see original for full logic)
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
	Type     NotificationType `json:"type"`
	Enabled  bool             `json:"enabled"`
	Settings map[string]any   `json:"settings,omitempty"`
}

// Validate checks if the notification configuration is valid
func (n *NotificationConfig) Validate() error {
	if n.Type == "" {
		return errors.New("notification type is required")
	}
	validTypes := map[NotificationType]bool{
		NotificationInApp: true,
		NotificationEmail: true,
	}
	if !validTypes[n.Type] {
		return fmt.Errorf("invalid notification type: %s", n.Type)
	}
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
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Enabled       bool                 `json:"enabled"`
	Severity      AlertSeverity        `json:"severity"`
	Threshold     ThresholdConfig      `json:"threshold"`
	Notifications []NotificationConfig `json:"notifications"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// Validate checks if the alert configuration is valid
func (a *AlertConfig) Validate() error {
	if a.ID == "" {
		return errors.New("alert ID is required")
	}
	if a.Name == "" {
		return errors.New("alert name is required")
	}
	validSeverities := map[AlertSeverity]bool{
		SeverityInfo:     true,
		SeverityWarning:  true,
		SeverityCritical: true,
	}
	if !validSeverities[a.Severity] {
		return fmt.Errorf("invalid severity: %s", a.Severity)
	}
	if err := a.Threshold.Validate(); err != nil {
		return fmt.Errorf("invalid threshold: %w", err)
	}
	return nil
}

// AlertState represents the state of an alert
type AlertState string

// Available alert states
const (
	StateActive   AlertState = "active"
	StateInactive AlertState = "inactive"
)

// AlertStatus represents the current status of an alert
type AlertStatus struct {
	AlertID      string     `json:"alert_id"`
	State        AlertState `json:"state"`
	CurrentValue float64    `json:"current_value"`
	TriggeredAt  *time.Time `json:"triggered_at,omitempty"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
	Message      string     `json:"message,omitempty"`
}
