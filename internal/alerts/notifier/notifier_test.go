package notifier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/alerts"
	"argus/internal/alerts/evaluator"
)

func TestNewNotifier(t *testing.T) {
	// Test with default config
	n := NewNotifier(nil)
	assert.NotNil(t, n)
	assert.NotNil(t, n.config)
	assert.Equal(t, 5, n.config.RateLimit)
	assert.Equal(t, 1*time.Hour, n.config.RateLimitWindow)
	assert.NotNil(t, n.channels)
	assert.NotNil(t, n.rateLimits)

	// Test with custom config
	customConfig := &NotifierConfig{
		RateLimit:       10,
		RateLimitWindow: 30 * time.Minute,
	}
	n = NewNotifier(customConfig)
	assert.NotNil(t, n)
	assert.Equal(t, customConfig, n.config)
	assert.Equal(t, 10, n.config.RateLimit)
	assert.Equal(t, 30*time.Minute, n.config.RateLimitWindow)
}

func TestRegisterChannel(t *testing.T) {
	n := NewNotifier(nil)
	inAppChannel := NewInAppChannel(100)

	// Register the channel
	n.RegisterChannel(inAppChannel)

	// Verify it was registered
	channel, ok := n.GetChannel(alerts.NotificationInApp)
	assert.True(t, ok)
	assert.Equal(t, inAppChannel, channel)
	assert.Equal(t, "In-App Notifications", channel.Name())
}

func TestProcessEvent(t *testing.T) {
	n := NewNotifier(nil)

	// Create a mock channel for testing
	mockChannel := &mockNotificationChannel{
		sendFunc: func(event evaluator.AlertEvent, subject, body string) error {
			return nil
		},
	}
	n.RegisterChannel(mockChannel)

	// Create a test alert event
	event := createTestAlertEvent(t)

	// Process the event
	n.ProcessEvent(event)

	// Verify the mock channel was called
	assert.Equal(t, 1, mockChannel.sendCount)
	assert.Equal(t, event.AlertID, mockChannel.lastEvent.AlertID)
}

func TestRateLimiting(t *testing.T) {
	// Create a notifier with a low rate limit
	config := &NotifierConfig{
		RateLimit:       2,
		RateLimitWindow: 1 * time.Hour,
	}
	n := NewNotifier(config)

	// Create a mock channel for testing
	mockChannel := &mockNotificationChannel{
		sendFunc: func(event evaluator.AlertEvent, subject, body string) error {
			return nil
		},
	}
	n.RegisterChannel(mockChannel)

	// Create a test alert event
	event := createTestAlertEvent(t)

	// Process the event multiple times
	for i := 0; i < 5; i++ {
		n.ProcessEvent(event)
	}

	// Verify the mock channel was only called twice (due to rate limiting)
	assert.Equal(t, 2, mockChannel.sendCount)
}

func TestRenderTemplates(t *testing.T) {
	n := NewNotifier(nil)
	event := createTestAlertEvent(t)

	// Test rendering templates
	subject, body, err := n.renderTemplates(event)
	require.NoError(t, err)
	assert.Contains(t, subject, "[CRITICAL] Argus Alert: Test Alert")
	assert.Contains(t, body, "Alert: Test Alert")
	assert.Contains(t, body, "Status: ACTIVE")
	assert.Contains(t, body, "Severity: CRITICAL")
}

// mockNotificationChannel is a mock implementation of NotificationChannel for testing
type mockNotificationChannel struct {
	sendFunc    func(event evaluator.AlertEvent, subject, body string) error
	sendCount   int
	lastEvent   evaluator.AlertEvent
	lastSubject string
	lastBody    string
}

func (m *mockNotificationChannel) Send(event evaluator.AlertEvent, subject, body string) error {
	m.sendCount++
	m.lastEvent = event
	m.lastSubject = subject
	m.lastBody = body
	return m.sendFunc(event, subject, body)
}

func (m *mockNotificationChannel) Type() alerts.NotificationType {
	return alerts.NotificationInApp
}

func (m *mockNotificationChannel) Name() string {
	return "Mock Notification Channel"
}

// createTestAlertEvent creates a test alert event for testing
func createTestAlertEvent(t *testing.T) evaluator.AlertEvent {
	alert := &alerts.AlertConfig{
		ID:          "test-alert",
		Name:        "Test Alert",
		Description: "Test alert description",
		Enabled:     true,
		Severity:    alerts.SeverityCritical,
		Threshold: alerts.ThresholdConfig{
			MetricType: alerts.MetricCPU,
			MetricName: "usage_percent",
			Operator:   alerts.OperatorGreaterThan,
			Value:      90.0,
		},
		Notifications: []alerts.NotificationConfig{
			{
				Type:    alerts.NotificationInApp,
				Enabled: true,
			},
		},
	}

	status := &alerts.AlertStatus{
		AlertID:      alert.ID,
		State:        alerts.StateActive,
		CurrentValue: 95.0,
	}

	return evaluator.AlertEvent{
		AlertID:      alert.ID,
		OldState:     alerts.StateInactive,
		NewState:     alerts.StateActive,
		CurrentValue: 95.0,
		Threshold:    90.0,
		Timestamp:    time.Now(),
		Message:      "CPU usage exceeded threshold",
		Alert:        alert,
		Status:       status,
	}
}
