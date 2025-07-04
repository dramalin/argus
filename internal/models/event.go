// File: internal/models/event.go
// Brief: Event-related data models for Argus
// Detailed: Contains type definition for AlertEvent.
// Author: drama.lin@aver.com
// Date: 2024-07-03

package models

import (
	"time"
)

// AlertEvent represents an alert state change event
type AlertEvent struct {
	AlertID      string       // ID of the alert that changed state
	OldState     AlertState   // Previous state
	NewState     AlertState   // New state
	CurrentValue float64      // Current metric value
	Threshold    float64      // Alert threshold value
	Timestamp    time.Time    // When the state change occurred
	Message      string       // Human-readable message
	Alert        *AlertConfig // The full alert configuration
	Status       *AlertStatus // The current alert status
}
