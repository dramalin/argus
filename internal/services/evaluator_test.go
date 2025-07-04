// File: internal/sync/evaluator_test.go
// Brief: Tests for unified alert evaluation logic (migrated from internal/alerts/evaluator/)
// Detailed: Contains tests for Evaluator, metricCollector, and all related logic for evaluating alert conditions and generating events.
// Author: drama.lin@aver.com
// Date: 2024-07-03

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/database"
	"argus/internal/models"
)

func createTestAlertStore(t testing.TB) *database.AlertStore {
	t.Helper()
	tempDir := t.TempDir()
	store, err := database.NewAlertStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create alert store: %v", err)
	}
	return store
}

// Helper function to create a test alert config with a unique ID
func createTestAlertConfig(t testing.TB) models.AlertConfig {
	t.Helper()
	return models.AlertConfig{
		ID:          fmt.Sprintf("test-alert-%s", uuid.New().String()),
		Name:        "Test CPU Alert",
		Description: "Test alert for high CPU usage",
		Severity:    "critical",
		Enabled:     true,
		Threshold: models.ThresholdConfig{
			MetricType: models.MetricCPU,
			MetricName: "usage_percent",
			Operator:   models.OperatorGreaterThan,
			Value:      90.0,
		},
		Notifications: []models.NotificationConfig{
			{
				Type: models.NotificationInApp,
			},
		},
	}
}

func TestNewEvaluator(t *testing.T) {
	store := createTestAlertStore(t)
	config := &EvaluatorConfig{
		EvaluationInterval: 15 * time.Second,
		AlertDebounceCount: 3,
		AlertResolveCount:  2,
	}

	evaluator := NewEvaluator(store, config)
	assert.NotNil(t, evaluator)
	assert.Equal(t, config, evaluator.config)
	assert.NotNil(t, evaluator.alertStatus)
	assert.NotNil(t, evaluator.eventCh)
	assert.NotNil(t, evaluator.metrics)

	// Test default config
	evaluator = NewEvaluator(store, nil)
	assert.NotNil(t, evaluator)
	assert.Equal(t, DefaultEvaluationInterval, evaluator.config.EvaluationInterval)
	assert.Equal(t, DefaultAlertDebounceCount, evaluator.config.AlertDebounceCount)
	assert.Equal(t, DefaultAlertResolveCount, evaluator.config.AlertResolveCount)
}

func TestEvaluatorStart(t *testing.T) {
	store := createTestAlertStore(t)
	testAlert := createTestAlertConfig(t)

	// Create test alert
	err := store.CreateAlert(&testAlert)
	require.NoError(t, err)

	evaluator := NewEvaluator(store, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = evaluator.Start(ctx)
	require.NoError(t, err)

	// Wait for at least one evaluation cycle
	time.Sleep(50 * time.Millisecond)

	// Verify the alert status was initialized
	evaluator.statusMu.RLock()
	status, exists := evaluator.alertStatus[testAlert.ID]
	evaluator.statusMu.RUnlock()

	assert.True(t, exists)
	assert.NotNil(t, status)
	assert.Equal(t, testAlert.ID, status.AlertID)

	// Stop the evaluator
	evaluator.Stop()
}

func TestMetricCollection(t *testing.T) {
	store := createTestAlertStore(t)
	testAlert := createTestAlertConfig(t)

	// Create test alert
	err := store.CreateAlert(&testAlert)
	require.NoError(t, err)

	evaluator := NewEvaluator(store, &EvaluatorConfig{
		EvaluationInterval: 50 * time.Millisecond,
		AlertDebounceCount: 1,
		AlertResolveCount:  1,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = evaluator.Start(ctx)
	require.NoError(t, err)

	// Wait for events
	eventCount := 0
	timeout := time.After(150 * time.Millisecond)
	for {
		select {
		case event := <-evaluator.Events():
			eventCount++
			assert.Equal(t, testAlert.ID, event.AlertID)
			assert.NotZero(t, event.CurrentValue)
		case <-timeout:
			goto end
		}
	}
end:
	evaluator.Stop()
	assert.True(t, eventCount > 0, "Should have received at least one event")
}

func TestEvaluatorStopCleanup(t *testing.T) {
	store := createTestAlertStore(t)
	evaluator := NewEvaluator(store, nil)

	// Start and immediately stop
	ctx := context.Background()
	err := evaluator.Start(ctx)
	require.NoError(t, err)
	evaluator.Stop()

	// Verify channel is closed
	_, ok := <-evaluator.eventCh
	assert.False(t, ok, "Event channel should be closed")
}

func TestAlertStateTransitions(t *testing.T) {
	store := createTestAlertStore(t)
	testAlert := createTestAlertConfig(t)

	// Create test alert
	err := store.CreateAlert(&testAlert)
	require.NoError(t, err)

	evaluator := NewEvaluator(store, &EvaluatorConfig{
		EvaluationInterval: 50 * time.Millisecond,
		AlertDebounceCount: 2, // Require 2 consecutive violations to trigger
		AlertResolveCount:  2, // Require 2 consecutive recoveries to resolve
	})

	// Initialize with inactive state
	evaluator.alertStatus[testAlert.ID] = &models.AlertStatus{
		AlertID: testAlert.ID,
		State:   models.StateInactive,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err = evaluator.Start(ctx)
	require.NoError(t, err)

	// Wait and collect events
	var events []models.AlertEvent
	timeout := time.After(250 * time.Millisecond)
collectEvents:
	for {
		select {
		case event := <-evaluator.Events():
			events = append(events, event)
		case <-timeout:
			break collectEvents
		}
	}

	evaluator.Stop()

	// Verify we got some events
	assert.NotEmpty(t, events, "Should have received at least one event")
}
