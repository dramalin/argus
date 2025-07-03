// File: internal/models/notification.go
// Brief: Notification-related data models for Argus
// Detailed: Contains type definition for InAppNotification.
// Author: Argus Migration (AI)
// Date: 2024-07-03

package models

import (
	"time"
)

// InAppNotification represents an in-app notification
type InAppNotification struct {
	ID        string        // Unique identifier
	AlertID   string        // ID of the alert that triggered this notification
	AlertName string        // Name of the alert
	Severity  AlertSeverity // Alert severity
	State     AlertState    // Alert state
	Message   string        // Notification message
	Subject   string        // Notification subject
	Timestamp time.Time     // When the notification was created
	Read      bool          // Whether the notification has been read
}
