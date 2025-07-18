// File: internal/services/notifier_test.go
// Brief: Tests for unified notification system for alerts (migrated from internal/alerts/notifier/)
// Detailed: Contains tests for Notifier, NotificationChannel, EmailChannel, InAppChannel, and all related logic for alert notifications.
// Author: drama.lin@aver.com
// Date: 2024-07-03

package services

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"

	"argus/internal/models"
	"argus/internal/server"
)

func TestNewNotifier(t *testing.T) {
	n := NewNotifier(nil)
	assert.NotNil(t, n)
	assert.NotNil(t, n.config)
	assert.Equal(t, 5, n.config.RateLimit)
	assert.Equal(t, 1*time.Hour, n.config.RateLimitWindow)
	assert.NotNil(t, n.channels)
	assert.NotNil(t, n.rateLimiter)

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
	n.RegisterChannel(inAppChannel)
	channel, ok := n.GetChannel(models.NotificationInApp)
	assert.True(t, ok)
	assert.Equal(t, inAppChannel, channel)
	assert.Equal(t, "In-App Notifications", channel.Name())
}

func TestProcessEvent(t *testing.T) {
	n := NewNotifier(nil)
	mockChannel := &mockNotificationChannel{
		sendFunc: func(event models.AlertEvent, subject, body string) error {
			return nil
		},
	}
	n.RegisterChannel(mockChannel)
	event := createTestAlertEvent(t)
	n.ProcessEvent(event)
	assert.Equal(t, 1, mockChannel.sendCount)
	assert.Equal(t, event.AlertID, mockChannel.lastEvent.AlertID)
}

func TestRateLimiting(t *testing.T) {
	config := &NotifierConfig{
		RateLimit:       2,
		RateLimitWindow: 1 * time.Hour,
	}
	n := NewNotifier(config)
	mockChannel := &mockNotificationChannel{
		sendFunc: func(event models.AlertEvent, subject, body string) error {
			return nil
		},
	}
	n.RegisterChannel(mockChannel)
	event := createTestAlertEvent(t)
	for i := 0; i < 5; i++ {
		n.ProcessEvent(event)
	}
	assert.Equal(t, 2, mockChannel.sendCount)
}

func TestRenderTemplates(t *testing.T) {
	n := NewNotifier(nil)
	event := createTestAlertEvent(t)
	subject, body, err := n.renderTemplates(event)
	require.NoError(t, err)
	assert.Contains(t, subject, "[CRITICAL] Argus Alert: Test Alert")
	assert.Contains(t, body, "Alert: Test Alert")
	assert.Contains(t, body, "Status: ACTIVE")
	assert.Contains(t, body, "Severity: CRITICAL")
}

type mockNotificationChannel struct {
	sendFunc    func(event models.AlertEvent, subject, body string) error
	sendCount   int
	lastEvent   models.AlertEvent
	lastSubject string
	lastBody    string
}

func (m *mockNotificationChannel) Send(event models.AlertEvent, subject, body string) error {
	m.sendCount++
	m.lastEvent = event
	m.lastSubject = subject
	m.lastBody = body
	return m.sendFunc(event, subject, body)
}

func (m *mockNotificationChannel) Type() models.NotificationType {
	return models.NotificationInApp
}

func (m *mockNotificationChannel) Name() string {
	return "Mock Notification Channel"
}

func createTestAlertEvent(t *testing.T) models.AlertEvent {
	alert := &models.AlertConfig{
		ID:          "test-alert",
		Name:        "Test Alert",
		Description: "Test alert description",
		Enabled:     true,
		Severity:    models.SeverityCritical,
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
	}
	status := &models.AlertStatus{
		AlertID:      alert.ID,
		State:        models.StateActive,
		CurrentValue: 95.0,
	}
	return models.AlertEvent{
		AlertID:      alert.ID,
		OldState:     models.StateInactive,
		NewState:     models.StateActive,
		CurrentValue: 95.0,
		Threshold:    90.0,
		Timestamp:    time.Now(),
		Message:      "CPU usage exceeded threshold",
		Alert:        alert,
		Status:       status,
	}
}

// EmailChannel tests
func TestNewEmailChannel(t *testing.T) {
	channel := NewEmailChannel(nil, nil)
	assert.NotNil(t, channel)
	assert.NotNil(t, channel.config)
	assert.Equal(t, "smtp.example.com", channel.config.Host)
	assert.Equal(t, 587, channel.config.Port)
	assert.Equal(t, "alerts@example.com", channel.config.Username)
	assert.Equal(t, "alerts@example.com", channel.config.From)
	assert.True(t, channel.config.UseSSL)

	customConfig := &EmailConfig{
		Host:     "smtp.custom.com",
		Port:     465,
		Username: "custom@example.com",
		Password: "password",
		From:     "custom@example.com",
		UseSSL:   false,
	}
	channel = NewEmailChannel(customConfig, nil)
	assert.NotNil(t, channel)
	assert.Equal(t, customConfig, channel.config)
	assert.Equal(t, "smtp.custom.com", channel.config.Host)
	assert.Equal(t, 465, channel.config.Port)
	assert.Equal(t, "custom@example.com", channel.config.Username)
	assert.Equal(t, "password", channel.config.Password)
	assert.Equal(t, "custom@example.com", channel.config.From)
	assert.False(t, channel.config.UseSSL)
}

func TestEmailChannelType(t *testing.T) {
	channel := NewEmailChannel(nil, nil)
	assert.Equal(t, models.NotificationEmail, channel.Type())
}

func TestEmailChannelName(t *testing.T) {
	channel := NewEmailChannel(nil, nil)
	assert.Equal(t, "Email Notifications", channel.Name())
}

func TestValidateRecipient(t *testing.T) {
	assert.True(t, ValidateRecipient("user@example.com"))
	assert.True(t, ValidateRecipient("user.name@example.com"))
	assert.True(t, ValidateRecipient("user+tag@example.com"))
	assert.True(t, ValidateRecipient("user@subdomain.example.com"))
	assert.False(t, ValidateRecipient(""))
	assert.False(t, ValidateRecipient("user"))
	assert.False(t, ValidateRecipient("user@"))
	assert.False(t, ValidateRecipient("@example.com"))
	assert.False(t, ValidateRecipient("user@example"))
}

func TestEmailChannelSend(t *testing.T) {
	channel := NewEmailChannel(nil, nil)
	event := createTestAlertEvent(t)
	event.Alert.Notifications = append(event.Alert.Notifications, models.NotificationConfig{
		Type:    models.NotificationEmail,
		Enabled: true,
		Settings: map[string]any{
			"recipient": "test@example.com",
		},
	})
	err := channel.Send(event, "Test Subject", "Test Body")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to send email") || strings.Contains(err.Error(), "no valid email recipient"))
}

// InAppChannel tests
func TestNewInAppChannel(t *testing.T) {
	channel := NewInAppChannel(0)
	assert.NotNil(t, channel)
	assert.Equal(t, 100, channel.maxSize)
	channel = NewInAppChannel(50)
	assert.NotNil(t, channel)
	assert.Equal(t, 50, channel.maxSize)
}

// MockHub is a mock implementation of the Hub
type MockHub struct {
	mock.Mock
}

func (m *MockHub) Broadcast(message []byte) {
	m.Called(message)
}

func (m *MockHub) Run() {
	// No-op for testing
}

func TestInAppChannel_Send(t *testing.T) {
	// Setup
	mockHub := new(MockHub)
	channel := NewInAppChannel(5, mockHub)

	event := models.AlertEvent{
		Alert: &models.AlertConfig{ID: "alert-1"},
	}
	subject := "Test Subject"
	body := "Test Body"

	// Expected notification
	var capturedNotification models.InAppNotification
	mockHub.On("Broadcast", mock.Anything).Run(func(args mock.Arguments) {
		msgBytes := args.Get(0).([]byte)
		json.Unmarshal(msgBytes, &capturedNotification)
	}).Return()

	// Action
	err := channel.Send(event, subject, body)

	// Assertions
	assert.NoError(t, err)
	mockHub.AssertCalled(t, "Broadcast", mock.Anything)

	assert.Equal(t, "alert-1", capturedNotification.AlertID)
	assert.Equal(t, subject, capturedNotification.Subject)
	assert.Equal(t, body, capturedNotification.Body)
	assert.False(t, capturedNotification.IsRead)

	// Verify in-memory store
	notifications := channel.GetNotifications()
	assert.Len(t, notifications, 1)
	assert.Equal(t, capturedNotification.ID, notifications[0].ID)
}

func TestInAppChannel_MaxSize(t *testing.T) {
	mockHub := new(MockHub)
	channel := NewInAppChannel(2, mockHub)
	mockHub.On("Broadcast", mock.Anything).Return()

	// Send 3 notifications
	channel.Send(models.AlertEvent{Alert: &models.AlertConfig{ID: "1"}}, "s1", "b1")
	time.Sleep(1 * time.Millisecond) // Ensure unique timestamps
	channel.Send(models.AlertEvent{Alert: &models.AlertConfig{ID: "2"}}, "s2", "b2")
	time.Sleep(1 * time.Millisecond)
	channel.Send(models.AlertEvent{Alert: &models.AlertConfig{ID: "3"}}, "s3", "b3")

	notifications := channel.GetNotifications()
	assert.Len(t, notifications, 2)
	assert.Equal(t, "2", notifications[0].AlertID) // Oldest (1) should be gone
	assert.Equal(t, "3", notifications[1].AlertID)
}

func TestInAppChannelGetUnreadNotifications(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)
	for i := 0; i < 3; i++ {
		err := channel.Send(event, "Test Subject", "Test Body")
		require.NoError(t, err)
	}
	notifications := channel.GetNotifications()
	channel.MarkAsRead(notifications[1].ID)
	unread := channel.GetUnreadNotifications()
	assert.Len(t, unread, 2)
}

func TestInAppChannelMarkAsRead(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)
	err := channel.Send(event, "Test Subject", "Test Body")
	require.NoError(t, err)
	notifications := channel.GetNotifications()
	require.Len(t, notifications, 1)
	id := notifications[0].ID
	result := channel.MarkAsRead(id)
	assert.True(t, result)
	notifications = channel.GetNotifications()
	assert.True(t, notifications[0].Read)
	result = channel.MarkAsRead("non-existent-id")
	assert.False(t, result)
}

func TestInAppChannelMarkAllAsRead(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)
	for i := 0; i < 3; i++ {
		err := channel.Send(event, "Test Subject", "Test Body")
		require.NoError(t, err)
	}
	channel.MarkAllAsRead()
	notifications := channel.GetNotifications()
	for _, notification := range notifications {
		assert.True(t, notification.Read)
	}
	unread := channel.GetUnreadNotifications()
	assert.Len(t, unread, 0)
}

func TestInAppChannelClearNotifications(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)
	for i := 0; i < 3; i++ {
		err := channel.Send(event, "Test Subject", "Test Body")
		require.NoError(t, err)
	}
	channel.ClearNotifications()
	notifications := channel.GetNotifications()
	assert.Len(t, notifications, 0)
}
