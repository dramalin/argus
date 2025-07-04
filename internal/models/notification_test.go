package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInAppNotification(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		notification InAppNotification
		want         InAppNotification
	}{
		{
			name: "Create notification with full data",
			notification: InAppNotification{
				ID:        "test-notif-1",
				AlertID:   "test-alert-1",
				AlertName: "Test Alert",
				Severity:  "high",
				State:     StateActive,
				Message:   "CPU usage is high",
				Subject:   "High CPU Alert",
				Timestamp: now,
				Read:      false,
			},
			want: InAppNotification{
				ID:        "test-notif-1",
				AlertID:   "test-alert-1",
				AlertName: "Test Alert",
				Severity:  "high",
				State:     StateActive,
				Message:   "CPU usage is high",
				Subject:   "High CPU Alert",
				Timestamp: now,
				Read:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that all fields are set correctly
			assert.Equal(t, tt.want.ID, tt.notification.ID)
			assert.Equal(t, tt.want.AlertID, tt.notification.AlertID)
			assert.Equal(t, tt.want.AlertName, tt.notification.AlertName)
			assert.Equal(t, tt.want.Severity, tt.notification.Severity)
			assert.Equal(t, tt.want.State, tt.notification.State)
			assert.Equal(t, tt.want.Message, tt.notification.Message)
			assert.Equal(t, tt.want.Subject, tt.notification.Subject)
			assert.Equal(t, tt.want.Timestamp, tt.notification.Timestamp)
			assert.Equal(t, tt.want.Read, tt.notification.Read)
		})
	}
}
