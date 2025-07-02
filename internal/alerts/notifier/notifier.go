// Package notifier provides functionality for sending alert notifications through various channels
package notifier

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"sync"
	"time"

	"argus/internal/alerts"
	"argus/internal/alerts/evaluator"
)

// NotificationTemplate represents a template for notification messages
type NotificationTemplate struct {
	Subject string // Template for notification subject
	Body    string // Template for notification body
}

// DefaultTemplates provides default templates for different alert severities and states
var DefaultTemplates = map[alerts.AlertSeverity]map[alerts.AlertState]NotificationTemplate{
	alerts.SeverityInfo: {
		alerts.StateActive: {
			Subject: "[INFO] Argus Alert: {{ .Alert.Name }}",
			Body: `
Alert: {{ .Alert.Name }}
Status: ACTIVE
Severity: INFO
Time: {{ .Timestamp.Format "2006-01-02 15:04:05" }}
Value: {{ printf "%.2f" .CurrentValue }}
Threshold: {{ .Alert.Threshold.Operator }} {{ printf "%.2f" .Alert.Threshold.Value }}

{{ .Message }}

Description: {{ .Alert.Description }}
`,
		},
		alerts.StateInactive: {
			Subject: "[RESOLVED] Argus Alert: {{ .Alert.Name }}",
			Body: `
Alert: {{ .Alert.Name }}
Status: RESOLVED
Severity: INFO
Time: {{ .Timestamp.Format "2006-01-02 15:04:05" }}
Value: {{ printf "%.2f" .CurrentValue }}
Threshold: {{ .Alert.Threshold.Operator }} {{ printf "%.2f" .Alert.Threshold.Value }}

{{ .Message }}

Description: {{ .Alert.Description }}
`,
		},
	},
	alerts.SeverityWarning: {
		alerts.StateActive: {
			Subject: "[WARNING] Argus Alert: {{ .Alert.Name }}",
			Body: `
Alert: {{ .Alert.Name }}
Status: ACTIVE
Severity: WARNING
Time: {{ .Timestamp.Format "2006-01-02 15:04:05" }}
Value: {{ printf "%.2f" .CurrentValue }}
Threshold: {{ .Alert.Threshold.Operator }} {{ printf "%.2f" .Alert.Threshold.Value }}

{{ .Message }}

Description: {{ .Alert.Description }}
`,
		},
		alerts.StateInactive: {
			Subject: "[RESOLVED] Argus Alert: {{ .Alert.Name }}",
			Body: `
Alert: {{ .Alert.Name }}
Status: RESOLVED
Severity: WARNING
Time: {{ .Timestamp.Format "2006-01-02 15:04:05" }}
Value: {{ printf "%.2f" .CurrentValue }}
Threshold: {{ .Alert.Threshold.Operator }} {{ printf "%.2f" .Alert.Threshold.Value }}

{{ .Message }}

Description: {{ .Alert.Description }}
`,
		},
	},
	alerts.SeverityCritical: {
		alerts.StateActive: {
			Subject: "[CRITICAL] Argus Alert: {{ .Alert.Name }}",
			Body: `
Alert: {{ .Alert.Name }}
Status: ACTIVE
Severity: CRITICAL
Time: {{ .Timestamp.Format "2006-01-02 15:04:05" }}
Value: {{ printf "%.2f" .CurrentValue }}
Threshold: {{ .Alert.Threshold.Operator }} {{ printf "%.2f" .Alert.Threshold.Value }}

{{ .Message }}

Description: {{ .Alert.Description }}
`,
		},
		alerts.StateInactive: {
			Subject: "[RESOLVED] Argus Alert: {{ .Alert.Name }}",
			Body: `
Alert: {{ .Alert.Name }}
Status: RESOLVED
Severity: CRITICAL
Time: {{ .Timestamp.Format "2006-01-02 15:04:05" }}
Value: {{ printf "%.2f" .CurrentValue }}
Threshold: {{ .Alert.Threshold.Operator }} {{ printf "%.2f" .Alert.Threshold.Value }}

{{ .Message }}

Description: {{ .Alert.Description }}
`,
		},
	},
}

// NotifierConfig holds configuration for the notifier
type NotifierConfig struct {
	// RateLimit specifies the maximum number of notifications per alert ID per time window
	RateLimit int
	// RateLimitWindow specifies the time window for rate limiting
	RateLimitWindow time.Duration
	// Templates holds custom notification templates
	Templates map[alerts.AlertSeverity]map[alerts.AlertState]NotificationTemplate
}

// DefaultConfig returns a default notifier configuration
func DefaultConfig() *NotifierConfig {
	return &NotifierConfig{
		RateLimit:       5,                // Maximum 5 notifications per alert ID
		RateLimitWindow: 1 * time.Hour,    // Per hour
		Templates:       DefaultTemplates, // Use default templates
	}
}

// NotificationChannel is the interface that must be implemented by all notification channels
type NotificationChannel interface {
	// Send sends a notification through this channel
	Send(event evaluator.AlertEvent, subject, body string) error

	// Type returns the type of this notification channel
	Type() alerts.NotificationType

	// Name returns a human-readable name for this channel
	Name() string
}

// Notifier is responsible for dispatching notifications through registered channels
type Notifier struct {
	config     *NotifierConfig
	channels   map[alerts.NotificationType]NotificationChannel
	rateLimits map[string]*rateLimiter
	mu         sync.RWMutex
}

// rateLimiter tracks notification counts for rate limiting
type rateLimiter struct {
	counts     map[string]int       // Map of alert ID to count
	timestamps map[string]time.Time // Map of alert ID to last notification time
	mu         sync.Mutex
}

// NewNotifier creates a new notifier with the given configuration
func NewNotifier(config *NotifierConfig) *Notifier {
	if config == nil {
		config = DefaultConfig()
	}

	return &Notifier{
		config:     config,
		channels:   make(map[alerts.NotificationType]NotificationChannel),
		rateLimits: make(map[string]*rateLimiter),
	}
}

// RegisterChannel registers a notification channel
func (n *Notifier) RegisterChannel(channel NotificationChannel) {
	n.mu.Lock()
	defer n.mu.Unlock()

	channelType := channel.Type()
	n.channels[channelType] = channel
	n.rateLimits[string(channelType)] = &rateLimiter{
		counts:     make(map[string]int),
		timestamps: make(map[string]time.Time),
	}

	slog.Info("Registered notification channel", "type", channelType, "name", channel.Name())
}

// GetChannel returns a registered notification channel by type
func (n *Notifier) GetChannel(channelType alerts.NotificationType) (NotificationChannel, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	channel, ok := n.channels[channelType]
	return channel, ok
}

// ProcessEvent processes an alert event and sends notifications as needed
func (n *Notifier) ProcessEvent(event evaluator.AlertEvent) {
	// Only send notifications for active and inactive (resolved) states
	if event.NewState != alerts.StateActive &&
		!(event.NewState == alerts.StateInactive && event.OldState == alerts.StateActive) {
		return
	}

	// Check if notifications are configured for this alert
	if event.Alert == nil || len(event.Alert.Notifications) == 0 {
		slog.Debug("No notifications configured for alert", "alert_id", event.AlertID)
		return
	}

	// Process each notification configuration
	for _, notifConfig := range event.Alert.Notifications {
		if !notifConfig.Enabled {
			continue
		}

		// Get the notification channel
		channel, ok := n.GetChannel(notifConfig.Type)
		if !ok {
			slog.Warn("Notification channel not registered",
				"alert_id", event.AlertID,
				"channel_type", notifConfig.Type)
			continue
		}

		// Check rate limit
		if n.isRateLimited(string(notifConfig.Type), event.AlertID) {
			slog.Info("Notification rate limited",
				"alert_id", event.AlertID,
				"channel_type", notifConfig.Type)
			continue
		}

		// Render notification templates
		subject, body, err := n.renderTemplates(event)
		if err != nil {
			slog.Error("Failed to render notification templates",
				"alert_id", event.AlertID,
				"error", err)
			continue
		}

		// Send the notification
		if err := channel.Send(event, subject, body); err != nil {
			slog.Error("Failed to send notification",
				"alert_id", event.AlertID,
				"channel_type", notifConfig.Type,
				"error", err)
		} else {
			slog.Info("Notification sent successfully",
				"alert_id", event.AlertID,
				"alert_name", event.Alert.Name,
				"channel_type", notifConfig.Type,
				"state", event.NewState)
		}
	}
}

// isRateLimited checks if sending a notification would exceed the rate limit
func (n *Notifier) isRateLimited(channelType, alertID string) bool {
	n.mu.RLock()
	limiter, ok := n.rateLimits[channelType]
	n.mu.RUnlock()

	if !ok {
		return false
	}

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	// Check if we need to reset the counter (rate limit window has passed)
	lastTime, ok := limiter.timestamps[alertID]
	if !ok || time.Since(lastTime) > n.config.RateLimitWindow {
		limiter.counts[alertID] = 1
		limiter.timestamps[alertID] = time.Now()
		return false
	}

	// Increment counter and check if rate limit is exceeded
	count := limiter.counts[alertID] + 1
	if count > n.config.RateLimit {
		return true
	}

	limiter.counts[alertID] = count
	return false
}

// renderTemplates renders the notification templates for the given event
func (n *Notifier) renderTemplates(event evaluator.AlertEvent) (string, string, error) {
	// Get templates for the alert severity and state
	templates, ok := n.config.Templates[event.Alert.Severity]
	if !ok {
		return "", "", fmt.Errorf("no templates found for severity: %s", event.Alert.Severity)
	}

	tmpl, ok := templates[event.NewState]
	if !ok {
		return "", "", fmt.Errorf("no templates found for state: %s", event.NewState)
	}

	// Render subject template
	subjectBuf := new(bytes.Buffer)
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse subject template: %w", err)
	}

	if err := subjectTmpl.Execute(subjectBuf, event); err != nil {
		return "", "", fmt.Errorf("failed to execute subject template: %w", err)
	}

	// Render body template
	bodyBuf := new(bytes.Buffer)
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse body template: %w", err)
	}

	if err := bodyTmpl.Execute(bodyBuf, event); err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %w", err)
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}
