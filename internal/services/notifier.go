// File: internal/sync/notifier.go
// Brief: Unified notification system for alerts (migrated from internal/alerts/notifier/)
// Detailed: Contains Notifier, NotificationChannel, EmailChannel, InAppChannel, and all related logic for alert notifications.
// Author: drama.lin@aver.com
// Date: 2024-07-03

// Package services provides notification logic for the Argus system.
package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"argus/internal/models"
	"argus/internal/utils"
)

// ... (migrated and refactored code from notifier.go, email.go, inapp.go goes here) ...

// NotificationTemplate represents a template for notification messages
type NotificationTemplate struct {
	Subject string // Template for notification subject
	Body    string // Template for notification body
}

// CompiledTemplate holds pre-compiled templates for performance
type CompiledTemplate struct {
	Subject *template.Template
	Body    *template.Template
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
	// Email worker pool configuration
	EmailWorkerCount int
	EmailQueueSize   int
	// SMTP connection pool configuration
	SMTPPoolSize    int
	SMTPIdleTimeout time.Duration
}

func DefaultConfig() *NotifierConfig {
	return &NotifierConfig{
		RateLimit:        5,
		RateLimitWindow:  1 * time.Hour,
		Templates:        DefaultTemplates,
		EmailWorkerCount: 3,
		EmailQueueSize:   100,
		SMTPPoolSize:     5,
		SMTPIdleTimeout:  5 * time.Minute,
	}
}

type NotificationChannel interface {
	Send(event models.AlertEvent, subject, body string) error
	Type() models.NotificationType
	Name() string
}

// Efficient rate limiter using time-based expiry
type rateLimiter struct {
	entries sync.Map // map[string]*rateLimitEntry
	config  *NotifierConfig
}

type rateLimitEntry struct {
	count     int64
	expiresAt int64
}

func newRateLimiter(config *NotifierConfig) *rateLimiter {
	rl := &rateLimiter{config: config}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) isAllowed(key string) bool {
	now := time.Now().Unix()
	
	// Load or create entry
	entryInterface, _ := rl.entries.LoadOrStore(key, &rateLimitEntry{
		count:     0,
		expiresAt: now + int64(rl.config.RateLimitWindow.Seconds()),
	})

	entry := entryInterface.(*rateLimitEntry)

	// Check if entry is expired
	if entry.expiresAt < now {
		// Reset expired entry
		atomic.StoreInt64(&entry.count, 0)
		atomic.StoreInt64(&entry.expiresAt, now+int64(rl.config.RateLimitWindow.Seconds()))
	}

	// Check rate limit
	currentCount := atomic.LoadInt64(&entry.count)
	if currentCount >= int64(rl.config.RateLimit) {
		return false
	}

	// Increment counter
	atomic.AddInt64(&entry.count, 1)
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.RateLimitWindow)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Unix()
		rl.entries.Range(func(key, value interface{}) bool {
			entry := value.(*rateLimitEntry)
			if atomic.LoadInt64(&entry.expiresAt) < now {
				rl.entries.Delete(key)
			}
			return true
		})
	}
}

type Notifier struct {
	config            *NotifierConfig
	channels          map[models.NotificationType]NotificationChannel
	rateLimiter       *rateLimiter
	compiledTemplates map[models.AlertSeverity]map[models.AlertState]*CompiledTemplate
	mu                sync.RWMutex
}

func NewNotifier(config *NotifierConfig) *Notifier {
	if config == nil {
		config = DefaultConfig()
	}

	notifier := &Notifier{
		config:      config,
		channels:    make(map[models.NotificationType]NotificationChannel),
		rateLimiter: newRateLimiter(config),
	}

	// Pre-compile templates for performance
	if err := notifier.compileTemplates(); err != nil {
		slog.Error("Failed to compile templates", "error", err)
		// Use runtime compilation as fallback
		notifier.compiledTemplates = nil
	}

	return notifier
}

func (n *Notifier) compileTemplates() error {
	n.compiledTemplates = make(map[models.AlertSeverity]map[models.AlertState]*CompiledTemplate)

	templates := n.config.Templates
	if templates == nil {
		templates = DefaultTemplates
	}

	for severity, stateTemplates := range templates {
		n.compiledTemplates[severity] = make(map[models.AlertState]*CompiledTemplate)

		for state, tmpl := range stateTemplates {
			subjTmpl, err := template.New("subject").Parse(tmpl.Subject)
			if err != nil {
				return fmt.Errorf("failed to compile subject template for %s/%s: %w", severity, state, err)
			}

			bodyTmpl, err := template.New("body").Parse(tmpl.Body)
			if err != nil {
				return fmt.Errorf("failed to compile body template for %s/%s: %w", severity, state, err)
			}

			n.compiledTemplates[severity][state] = &CompiledTemplate{
				Subject: subjTmpl,
				Body:    bodyTmpl,
			}
		}
	}

	slog.Info("Pre-compiled notification templates", "count", len(templates))
	return nil
}

func (n *Notifier) RegisterChannel(channel NotificationChannel) {
	n.mu.Lock()
	defer n.mu.Unlock()
	channelType := channel.Type()
	n.channels[channelType] = channel
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
		// Check rate limit using efficient time-based expiry
		rateLimitKey := fmt.Sprintf("%s:%s", string(typ), event.AlertID)
		if !n.rateLimiter.isAllowed(rateLimitKey) {
			slog.Warn("Notification rate limited", "type", typ, "alert_id", event.AlertID)
			continue
		}

		// Render templates (using pre-compiled templates if available)
		subject, body, err := n.renderTemplates(event)
		if err != nil {
			slog.Error("Failed to render notification template", "error", err)
			continue
		}

		// Send notification (non-blocking for email)
		if err := channel.Send(event, subject, body); err != nil {
			slog.Error("Failed to send notification", "type", typ, "error", err)
			continue
		}
	}
}

func (n *Notifier) renderTemplates(event models.AlertEvent) (string, string, error) {
	sev := event.Alert.Severity
	state := event.NewState

	// Use pre-compiled templates if available
	if n.compiledTemplates != nil {
		if sevTemplates, ok := n.compiledTemplates[sev]; ok {
			if compiled, ok := sevTemplates[state]; ok {
				return n.executeCompiledTemplate(compiled, event)
			}
		}

		// Fallback to default template
		if compiled, ok := n.compiledTemplates[models.SeverityInfo][models.StateActive]; ok {
			return n.executeCompiledTemplate(compiled, event)
		}
	}

	// Fallback to runtime compilation (backward compatibility)
	return n.renderTemplatesRuntime(event)
}

func (n *Notifier) executeCompiledTemplate(compiled *CompiledTemplate, event models.AlertEvent) (string, string, error) {
	// Use pooled buffers for template rendering
	subjBuf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(subjBuf)
	
	bodyBuf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(bodyBuf)

	if err := compiled.Subject.Execute(subjBuf, event); err != nil {
		return "", "", fmt.Errorf("failed to execute subject template: %w", err)
	}

	if err := compiled.Body.Execute(bodyBuf, event); err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %w", err)
	}

	return subjBuf.String(), bodyBuf.String(), nil
}

func (n *Notifier) renderTemplatesRuntime(event models.AlertEvent) (string, string, error) {
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
	
	// Use pooled buffers for template rendering
	subjBuf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(subjBuf)
	
	bodyBuf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(bodyBuf)
	
	err = subjTmpl.Execute(subjBuf, event)
	if err != nil {
		return "", "", err
	}
	err = bodyTmpl.Execute(bodyBuf, event)
	if err != nil {
		return "", "", err
	}
	return subjBuf.String(), bodyBuf.String(), nil
}

// Email notification channel with worker pool and connection pooling

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

type EmailJob struct {
	Event   models.AlertEvent
	Subject string
	Body    string
}

type SMTPConnection struct {
	client   *smtp.Client
	lastUsed time.Time
	inUse    bool
}

type EmailChannel struct {
	config      *EmailConfig
	notifierCfg *NotifierConfig
	emailQueue  chan EmailJob
	smtpPool    sync.Pool
	workers     sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewEmailChannel(config *EmailConfig, notifierConfig *NotifierConfig) *EmailChannel {
	if config == nil {
		config = DefaultEmailConfig()
	}
	if notifierConfig == nil {
		notifierConfig = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	channel := &EmailChannel{
		config:      config,
		notifierCfg: notifierConfig,
		emailQueue:  make(chan EmailJob, notifierConfig.EmailQueueSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize SMTP connection pool
	channel.smtpPool = sync.Pool{
		New: func() interface{} {
			return &SMTPConnection{
				client:   nil,
				lastUsed: time.Now(),
				inUse:    false,
			}
		},
	}

	// Start email worker pool
	channel.startWorkers()

	// Start connection pool cleanup
	go channel.cleanupConnections()

	return channel
}

func (c *EmailChannel) startWorkers() {
	for i := 0; i < c.notifierCfg.EmailWorkerCount; i++ {
		c.workers.Add(1)
		go c.emailWorker()
	}
}

func (c *EmailChannel) emailWorker() {
	defer c.workers.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case job := <-c.emailQueue:
			c.processEmailJob(job)
		}
	}
}

func (c *EmailChannel) processEmailJob(job EmailJob) {
	if job.Event.Alert == nil || len(job.Event.Alert.Notifications) == 0 {
		slog.Error("Alert has no notification settings", "alert_id", job.Event.AlertID)
		return
	}

	var recipient string
	for _, notif := range job.Event.Alert.Notifications {
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
		slog.Error("No valid email recipient found", "alert_id", job.Event.AlertID)
		return
	}

	// Get SMTP connection from pool
	conn := c.getSMTPConnection()
	if conn == nil {
		slog.Error("Failed to get SMTP connection", "alert_id", job.Event.AlertID)
		return
	}

	defer c.returnSMTPConnection(conn)

	// Send email using pooled connection
	if err := c.sendEmailWithConnection(conn, recipient, job.Subject, job.Body); err != nil {
		slog.Error("Failed to send email", "recipient", recipient, "error", err)
		// Mark connection as bad
		conn.client = nil
		return
	}

	slog.Info("Email sent successfully",
		"recipient", recipient,
		"subject", job.Subject,
		"alert_id", job.Event.AlertID)
}

func (c *EmailChannel) getSMTPConnection() *SMTPConnection {
	conn := c.smtpPool.Get().(*SMTPConnection)

	// Check if connection is still valid
	if conn.client == nil || time.Since(conn.lastUsed) > c.notifierCfg.SMTPIdleTimeout {
		// Create new connection
		client, err := c.createSMTPClient()
		if err != nil {
			slog.Error("Failed to create SMTP client", "error", err)
			return nil
		}
		conn.client = client
	}

	conn.lastUsed = time.Now()
	conn.inUse = true
	return conn
}

func (c *EmailChannel) returnSMTPConnection(conn *SMTPConnection) {
	conn.inUse = false
	c.smtpPool.Put(conn)
}

func (c *EmailChannel) createSMTPClient() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SMTP server: %w", err)
	}

	// Start TLS if required
	if c.config.UseSSL {
		tlsConfig := &tls.Config{
			ServerName: c.config.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate
	if c.config.Username != "" && c.config.Password != "" {
		auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
		if err := client.Auth(auth); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	return client, nil
}

func (c *EmailChannel) sendEmailWithConnection(conn *SMTPConnection, recipient, subject, body string) error {
	// Set sender
	if err := conn.client.Mail(c.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := conn.client.Rcpt(recipient); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Get data writer
	w, err := conn.client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer w.Close()

	// Write message
	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		recipient, c.config.From, subject, body)

	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (c *EmailChannel) cleanupConnections() {
	ticker := time.NewTicker(c.notifierCfg.SMTPIdleTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// This is a simple cleanup - in a real implementation,
			// we'd track connections more carefully
		}
	}
}

func (c *EmailChannel) Send(event models.AlertEvent, subject, body string) error {
	job := EmailJob{
		Event:   event,
		Subject: subject,
		Body:    body,
	}

	// Non-blocking send to queue
	select {
	case c.emailQueue <- job:
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

func (c *EmailChannel) Type() models.NotificationType {
	return models.NotificationEmail
}

func (c *EmailChannel) Name() string {
	return "Email Notifications"
}

func (c *EmailChannel) Stop() {
	c.cancel()
	close(c.emailQueue)
	c.workers.Wait()
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

// Stop gracefully shuts down the notifier
func (n *Notifier) Stop() {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, channel := range n.channels {
		if emailChannel, ok := channel.(*EmailChannel); ok {
			emailChannel.Stop()
		}
	}
}
