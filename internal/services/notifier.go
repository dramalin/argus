// File: internal/sync/notifier.go
// Brief: Unified notification system for alerts (migrated from internal/alerts/notifier/)
// Detailed: Contains Notifier, NotificationChannel, EmailChannel, InAppChannel, and all related logic for alert notifications.
// Author: drama.lin@aver.com
// Date: 2024-07-03

// Package services provides notification logic for the Argus system.
package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"argus/internal/models"
)

// ... (migrated and refactored code from notifier.go, email.go, inapp.go goes here) ...

// NotificationTemplate represents a template for notification messages
type NotificationTemplate struct {
	Subject string // Template for notification subject
	Body    string // Template for notification body
}

// DefaultTemplates provides default templates for different alert severities and states
var DefaultTemplates = map[models.AlertSeverity]map[models.AlertState]NotificationTemplate{
	models.SeverityInfo: {
		models.StateActive: {
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
		models.StateInactive: {
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
	models.SeverityWarning: {
		models.StateActive: {
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
		models.StateInactive: {
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
	models.SeverityCritical: {
		models.StateActive: {
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
		models.StateInactive: {
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

// ... (NotifierConfig, Notifier, NotificationChannel, EmailChannel, InAppChannel, and all related methods and helpers migrated here) ...

// NotifierConfig holds configuration for the notifier
type NotifierConfig struct {
	RateLimit       int
	RateLimitWindow time.Duration
	Templates       map[models.AlertSeverity]map[models.AlertState]NotificationTemplate
}

func DefaultConfig() *NotifierConfig {
	return &NotifierConfig{
		RateLimit:       5,
		RateLimitWindow: 1 * time.Hour,
		Templates:       DefaultTemplates,
	}
}

type NotificationChannel interface {
	Send(event models.AlertEvent, subject, body string) error
	Type() models.NotificationType
	Name() string
}

type Notifier struct {
	config     *NotifierConfig
	channels   map[models.NotificationType]NotificationChannel
	rateLimits map[string]*rateLimiter
	mu         sync.RWMutex
}

type rateLimiter struct {
	counts     map[string]int
	timestamps map[string]time.Time
	mu         sync.Mutex
}

func NewNotifier(config *NotifierConfig) *Notifier {
	if config == nil {
		config = DefaultConfig()
	}
	return &Notifier{
		config:     config,
		channels:   make(map[models.NotificationType]NotificationChannel),
		rateLimits: make(map[string]*rateLimiter),
	}
}

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

func (n *Notifier) GetChannel(channelType models.NotificationType) (NotificationChannel, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	ch, ok := n.channels[channelType]
	return ch, ok
}

func (n *Notifier) ProcessEvent(event models.AlertEvent) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for typ, channel := range n.channels {
		if n.isRateLimited(string(typ), event.AlertID) {
			slog.Warn("Notification rate limited", "type", typ, "alert_id", event.AlertID)
			continue
		}
		subject, body, err := n.renderTemplates(event)
		if err != nil {
			slog.Error("Failed to render notification template", "error", err)
			continue
		}
		if err := channel.Send(event, subject, body); err != nil {
			slog.Error("Failed to send notification", "type", typ, "error", err)
			continue
		}
		n.updateRateLimit(string(typ), event.AlertID)
	}
}

func (n *Notifier) isRateLimited(channelType, alertID string) bool {
	rl, ok := n.rateLimits[channelType]
	if !ok {
		return false
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	count := rl.counts[alertID]
	last := rl.timestamps[alertID]
	if time.Since(last) > n.config.RateLimitWindow {
		rl.counts[alertID] = 0
		rl.timestamps[alertID] = time.Now()
		return false
	}
	return count >= n.config.RateLimit
}

func (n *Notifier) updateRateLimit(channelType, alertID string) {
	rl, ok := n.rateLimits[channelType]
	if !ok {
		return
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.counts[alertID]++
	rl.timestamps[alertID] = time.Now()
}

func (n *Notifier) renderTemplates(event models.AlertEvent) (string, string, error) {
	sev := event.Alert.Severity
	state := event.NewState
	tmpls := n.config.Templates
	if tmpls == nil {
		tmpls = DefaultTemplates
	}
	tmpl, ok := tmpls[sev][state]
	if !ok {
		tmpl = DefaultTemplates[models.SeverityInfo][models.StateActive]
	}
	subjTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return "", "", err
	}
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return "", "", err
	}
	var subjBuf, bodyBuf bytes.Buffer
	err = subjTmpl.Execute(&subjBuf, event)
	if err != nil {
		return "", "", err
	}
	err = bodyTmpl.Execute(&bodyBuf, event)
	if err != nil {
		return "", "", err
	}
	return subjBuf.String(), bodyBuf.String(), nil
}

// Email notification channel

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseSSL   bool
}

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

type EmailChannel struct {
	config *EmailConfig
}

func NewEmailChannel(config *EmailConfig) *EmailChannel {
	if config == nil {
		config = DefaultEmailConfig()
	}
	return &EmailChannel{config: config}
}

func (c *EmailChannel) Send(event models.AlertEvent, subject, body string) error {
	if event.Alert == nil || len(event.Alert.Notifications) == 0 {
		return fmt.Errorf("alert has no notification settings")
	}
	var recipient string
	for _, notif := range event.Alert.Notifications {
		if notif.Type == models.NotificationEmail && notif.Enabled {
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
	msg := []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", recipient, c.config.From, subject, body))
	slog.Info("Sending email notification", "recipient", recipient, "subject", subject, "alert_id", event.AlertID, "alert_name", event.Alert.Name)
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", c.config.Host, c.config.Port), auth, c.config.From, []string{recipient}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (c *EmailChannel) Type() models.NotificationType {
	return models.NotificationEmail
}

func (c *EmailChannel) Name() string {
	return "Email Notifications"
}

func ValidateRecipient(email string) bool {
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// In-app notification channel

type InAppChannel struct {
	notifications []models.InAppNotification
	maxSize       int
	mu            sync.RWMutex
}

func NewInAppChannel(maxSize int) *InAppChannel {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &InAppChannel{
		notifications: make([]models.InAppNotification, 0, maxSize),
		maxSize:       maxSize,
	}
}

func (c *InAppChannel) Send(event models.AlertEvent, subject, body string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	notification := models.InAppNotification{
		ID:        generateID(),
		AlertID:   event.AlertID,
		AlertName: event.Alert.Name,
		Severity:  event.Alert.Severity,
		State:     event.NewState,
		Message:   body,
		Subject:   subject,
		Timestamp: time.Now(),
		Read:      false,
	}
	c.notifications = append([]models.InAppNotification{notification}, c.notifications...)
	if len(c.notifications) > c.maxSize {
		c.notifications = c.notifications[:c.maxSize]
	}
	return nil
}

func (c *InAppChannel) Type() models.NotificationType {
	return models.NotificationInApp
}

func (c *InAppChannel) Name() string {
	return "In-App Notifications"
}

func (c *InAppChannel) GetNotifications() []models.InAppNotification {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]models.InAppNotification, len(c.notifications))
	copy(result, c.notifications)
	return result
}

func (c *InAppChannel) GetUnreadNotifications() []models.InAppNotification {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []models.InAppNotification
	for _, notification := range c.notifications {
		if !notification.Read {
			result = append(result, notification)
		}
	}
	return result
}

func (c *InAppChannel) MarkAsRead(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, notification := range c.notifications {
		if notification.ID == id {
			c.notifications[i].Read = true
			return true
		}
	}
	return false
}

func (c *InAppChannel) MarkAllAsRead() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.notifications {
		c.notifications[i].Read = true
	}
}

func (c *InAppChannel) ClearNotifications() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.notifications = make([]models.InAppNotification, 0, c.maxSize)
}

func generateID() string {
	return time.Now().Format("20060102150405.000") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(result)
}

// GetNotifications returns all in-app notifications if the channel is registered.
func (n *Notifier) GetNotifications() []models.InAppNotification {
	ch, ok := n.channels[models.NotificationInApp]
	if !ok {
		return nil
	}
	inApp, ok := ch.(*InAppChannel)
	if !ok {
		return nil
	}
	return inApp.GetNotifications()
}

// MarkNotificationRead marks a notification as read by ID in the in-app channel.
func (n *Notifier) MarkNotificationRead(id string) bool {
	ch, ok := n.channels[models.NotificationInApp]
	if !ok {
		return false
	}
	inApp, ok := ch.(*InAppChannel)
	if !ok {
		return false
	}
	return inApp.MarkAsRead(id)
}

// MarkAllNotificationsRead marks all in-app notifications as read.
func (n *Notifier) MarkAllNotificationsRead() {
	ch, ok := n.channels[models.NotificationInApp]
	if !ok {
		return
	}
	inApp, ok := ch.(*InAppChannel)
	if !ok {
		return
	}
	inApp.MarkAllAsRead()
}

// ClearNotifications removes all in-app notifications.
func (n *Notifier) ClearNotifications() {
	ch, ok := n.channels[models.NotificationInApp]
	if !ok {
		return
	}
	inApp, ok := ch.(*InAppChannel)
	if !ok {
		return
	}
	inApp.ClearNotifications()
}
