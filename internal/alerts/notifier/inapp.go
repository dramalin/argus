package notifier

import (
	"sync"
	"time"

	"argus/internal/alerts"
	"argus/internal/alerts/evaluator"
)

// InAppNotification represents an in-app notification
type InAppNotification struct {
	ID        string               // Unique identifier
	AlertID   string               // ID of the alert that triggered this notification
	AlertName string               // Name of the alert
	Severity  alerts.AlertSeverity // Alert severity
	State     alerts.AlertState    // Alert state
	Message   string               // Notification message
	Subject   string               // Notification subject
	Timestamp time.Time            // When the notification was created
	Read      bool                 // Whether the notification has been read
}

// InAppChannel is a notification channel that stores notifications in memory
type InAppChannel struct {
	notifications []InAppNotification
	maxSize       int
	mu            sync.RWMutex
}

// NewInAppChannel creates a new in-app notification channel
func NewInAppChannel(maxSize int) *InAppChannel {
	if maxSize <= 0 {
		maxSize = 100 // Default to storing 100 notifications
	}

	return &InAppChannel{
		notifications: make([]InAppNotification, 0, maxSize),
		maxSize:       maxSize,
	}
}

// Send sends a notification through this channel
func (c *InAppChannel) Send(event evaluator.AlertEvent, subject, body string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	notification := InAppNotification{
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

	// Add to the front of the list (most recent first)
	c.notifications = append([]InAppNotification{notification}, c.notifications...)

	// Trim if we exceed max size
	if len(c.notifications) > c.maxSize {
		c.notifications = c.notifications[:c.maxSize]
	}

	return nil
}

// Type returns the type of this notification channel
func (c *InAppChannel) Type() alerts.NotificationType {
	return alerts.NotificationInApp
}

// Name returns a human-readable name for this channel
func (c *InAppChannel) Name() string {
	return "In-App Notifications"
}

// GetNotifications returns all stored notifications
func (c *InAppChannel) GetNotifications() []InAppNotification {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	result := make([]InAppNotification, len(c.notifications))
	copy(result, c.notifications)

	return result
}

// GetUnreadNotifications returns all unread notifications
func (c *InAppChannel) GetUnreadNotifications() []InAppNotification {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []InAppNotification

	for _, notification := range c.notifications {
		if !notification.Read {
			result = append(result, notification)
		}
	}

	return result
}

// MarkAsRead marks a notification as read
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

// MarkAllAsRead marks all notifications as read
func (c *InAppChannel) MarkAllAsRead() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.notifications {
		c.notifications[i].Read = true
	}
}

// ClearNotifications removes all notifications
func (c *InAppChannel) ClearNotifications() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.notifications = make([]InAppNotification, 0, c.maxSize)
}

// generateID generates a unique ID for a notification
func generateID() string {
	return time.Now().Format("20060102150405.000") + "-" + randomString(8)
}

// randomString generates a random string of the given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond) // Ensure uniqueness
	}

	return string(result)
}
