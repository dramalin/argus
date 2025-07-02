package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"argus/internal/alerts"
)

func TestNewEmailChannel(t *testing.T) {
	// Test with default config
	channel := NewEmailChannel(nil)
	assert.NotNil(t, channel)
	assert.NotNil(t, channel.config)
	assert.Equal(t, "smtp.example.com", channel.config.Host)
	assert.Equal(t, 587, channel.config.Port)
	assert.Equal(t, "alerts@example.com", channel.config.Username)
	assert.Equal(t, "alerts@example.com", channel.config.From)
	assert.True(t, channel.config.UseSSL)

	// Test with custom config
	customConfig := &EmailConfig{
		Host:     "smtp.custom.com",
		Port:     465,
		Username: "custom@example.com",
		Password: "password",
		From:     "custom@example.com",
		UseSSL:   false,
	}
	channel = NewEmailChannel(customConfig)
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
	channel := NewEmailChannel(nil)
	assert.Equal(t, alerts.NotificationEmail, channel.Type())
}

func TestEmailChannelName(t *testing.T) {
	channel := NewEmailChannel(nil)
	assert.Equal(t, "Email Notifications", channel.Name())
}

func TestValidateRecipient(t *testing.T) {
	// Valid email addresses
	assert.True(t, ValidateRecipient("user@example.com"))
	assert.True(t, ValidateRecipient("user.name@example.com"))
	assert.True(t, ValidateRecipient("user+tag@example.com"))
	assert.True(t, ValidateRecipient("user@subdomain.example.com"))

	// Invalid email addresses
	assert.False(t, ValidateRecipient(""))
	assert.False(t, ValidateRecipient("user"))
	assert.False(t, ValidateRecipient("user@"))
	assert.False(t, ValidateRecipient("@example.com"))
	assert.False(t, ValidateRecipient("user@example"))
}

func TestEmailChannelSend(t *testing.T) {
	// This is a basic test that doesn't actually send emails
	// In a real test, you might use a mock SMTP server or dependency injection

	channel := NewEmailChannel(nil)
	event := createTestAlertEvent(t)

	// Add email notification to the event
	event.Alert.Notifications = append(event.Alert.Notifications, alerts.NotificationConfig{
		Type:    alerts.NotificationEmail,
		Enabled: true,
		Settings: map[string]any{
			"recipient": "test@example.com",
		},
	})

	// We expect this to fail since we're not connecting to a real SMTP server
	// but we can verify that it attempts to use the correct recipient
	err := channel.Send(event, "Test Subject", "Test Body")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
}
