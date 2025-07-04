package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlertEvent(t *testing.T) {
	now := time.Now()
	testAlert := &AlertConfig{
		ID:          "test-alert-1",
		Name:        "Test Alert",
		Description: "Test alert description",
		Severity:    "high",
		Enabled:     true,
		Threshold: ThresholdConfig{
			MetricType: MetricCPU,
			MetricName: "usage_percent",
			Operator:   OperatorGreaterThan,
			Value:      90.0,
		},
	}
	testStatus := &AlertStatus{
		AlertID:      "test-alert-1",
		State:        StateActive,
		CurrentValue: 95.0,
		Message:      "CPU usage is high",
		TriggeredAt:  &now,
	}

	tests := []struct {
		name  string
		event AlertEvent
		want  AlertEvent
	}{
		{
			name: "Create alert event with full data",
			event: AlertEvent{
				AlertID:      "test-alert-1",
				OldState:     StateInactive,
				NewState:     StateActive,
				CurrentValue: 95.0,
				Threshold:    90.0,
				Timestamp:    now,
				Message:      "CPU usage exceeded threshold",
				Alert:        testAlert,
				Status:       testStatus,
			},
			want: AlertEvent{
				AlertID:      "test-alert-1",
				OldState:     StateInactive,
				NewState:     StateActive,
				CurrentValue: 95.0,
				Threshold:    90.0,
				Timestamp:    now,
				Message:      "CPU usage exceeded threshold",
				Alert:        testAlert,
				Status:       testStatus,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that all fields are set correctly
			assert.Equal(t, tt.want.AlertID, tt.event.AlertID)
			assert.Equal(t, tt.want.OldState, tt.event.OldState)
			assert.Equal(t, tt.want.NewState, tt.event.NewState)
			assert.Equal(t, tt.want.CurrentValue, tt.event.CurrentValue)
			assert.Equal(t, tt.want.Threshold, tt.event.Threshold)
			assert.Equal(t, tt.want.Timestamp, tt.event.Timestamp)
			assert.Equal(t, tt.want.Message, tt.event.Message)
			assert.Equal(t, tt.want.Alert, tt.event.Alert)
			assert.Equal(t, tt.want.Status, tt.event.Status)
		})
	}
}
