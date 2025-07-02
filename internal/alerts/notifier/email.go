package notifier

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"argus/internal/alerts"
	"argus/internal/alerts/evaluator"
)

// EmailConfig holds configuration for the email notification channel
type EmailConfig struct {
	Host     string // SMTP server host
	Port     int    // SMTP server port
	Username string // SMTP username
	Password string // SMTP password
	From     string // Sender email address
	UseSSL   bool   // Whether to use SSL/TLS
}

// DefaultEmailConfig returns a default email configuration
func DefaultEmailConfig() *EmailConfig {
	return &EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "alerts@example.com",
		Password: "",
		From:     "alerts@example.com",
		UseSSL:   true,
	}
}

// EmailChannel is a notification channel that sends emails
type EmailChannel struct {
	config *EmailConfig
}

// NewEmailChannel creates a new email notification channel
func NewEmailChannel(config *EmailConfig) *EmailChannel {
	if config == nil {
		config = DefaultEmailConfig()
	}

	return &EmailChannel{
		config: config,
	}
}

// Send sends a notification through this channel
func (c *EmailChannel) Send(event evaluator.AlertEvent, subject, body string) error {
	// Check if the alert has notification settings
	if event.Alert == nil || len(event.Alert.Notifications) == 0 {
		return fmt.Errorf("alert has no notification settings")
	}

	// Find the email notification config
	var recipient string
	for _, notif := range event.Alert.Notifications {
		if notif.Type == alerts.NotificationEmail && notif.Enabled {
			if notif.Settings != nil {
				if r, ok := notif.Settings["recipient"].(string); ok && r != "" {
					recipient = r
					break
				}
			}
		}
	}

	if recipient == "" {
		return fmt.Errorf("no valid email recipient found in alert notification settings")
	}

	// Prepare email message
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", recipient, c.config.From, subject, body))

	// Log the email attempt
	slog.Info("Sending email notification",
		"recipient", recipient,
		"subject", subject,
		"alert_id", event.AlertID,
		"alert_name", event.Alert.Name)

	// Send the email
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		auth,
		c.config.From,
		[]string{recipient},
		msg,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Type returns the type of this notification channel
func (c *EmailChannel) Type() alerts.NotificationType {
	return alerts.NotificationEmail
}

// Name returns a human-readable name for this channel
func (c *EmailChannel) Name() string {
	return "Email Notifications"
}

// ValidateRecipient checks if an email address is valid
func ValidateRecipient(email string) bool {
	if email == "" {
		return false
	}

	// Basic validation: must contain @ and at least one dot after @
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}

	// Domain part should have at least one dot
	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}
