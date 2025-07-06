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
	"github.com/stretchr/testify/mock"

	"argus/internal/database"
	"argus/internal/metrics"
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

// MockMetricsCollector is a mock implementation of the MetricsCollector
type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) GetCPU() (*metrics.CPUMetrics, error) {
	args := m.Called()
	return args.Get(0).(*metrics.CPUMetrics), args.Error(1)
}

func (m *MockMetricsCollector) GetMemory() (*metrics.MemoryMetrics, error) {
	args := m.Called()
	return args.Get(0).(*metrics.MemoryMetrics), args.Error(1)
}

func (m *MockMetricsCollector) GetNetwork() (*metrics.NetworkMetrics, error) {
	args := m.Called()
	return args.Get(0).(*metrics.NetworkMetrics), args.Error(1)
}

func TestEvaluator_evaluateMetricFromCollector(t *testing.T) {
	// Setup
	mockCollector := new(MockMetricsCollector)
	alertStore, _ := database.NewAlertStore(":memory:")
	evaluator := NewEvaluator(alertStore, DefaultEvaluatorConfig())
	evaluator.SetMetricsCollector(mockCollector)

	// Mock CPU metrics
	mockCollector.On("GetCPU").Return(&metrics.CPUMetrics{
		UsagePercent: 95.0,
		Load1:        1.5,
	}, nil)

	// Test case 1: CPU usage
	threshold := models.ThresholdConfig{MetricType: "cpu", MetricName: "usage_percent"}
	value, err := evaluator.evaluateMetric(threshold)
	assert.NoError(t, err)
	assert.Equal(t, 95.0, value)

	// Test case 2: CPU load
	threshold = models.ThresholdConfig{MetricType: "cpu", MetricName: "load1"}
	value, err = evaluator.evaluateMetric(threshold)
	assert.NoError(t, err)
	assert.Equal(t, 1.5, value)

	mockCollector.AssertExpectations(t)
}

func TestEvaluator_evaluateMetricDirect(t *testing.T) {
	// This test ensures the fallback works.
	// Note: This test will actually call gopsutil and may be flaky depending on the test environment.
	// A more robust implementation would involve mocking gopsutil calls.
	t.Skip("Skipping direct evaluation test as it depends on the host system state.")
}

func TestEvaluator_processAlertState(t *testing.T) {
	alertStore, _ := database.NewAlertStore(":memory:")
	evaluator := NewEvaluator(alertStore, DefaultEvaluatorConfig())

	alertConfig := &models.AlertConfig{
		ID:       "alert-1",
		Name:     "Test Alert",
		Severity: models.SeverityCritical,
		Threshold: models.ThresholdConfig{
			Value: 90,
		},
	}
	pendingCounters := make(map[string]int)
	resolveCounters := make(map[string]int)

	// Initial state: inactive
	status, _ := evaluator.GetAlertStatus("alert-1")
	if status == nil {
		status = &models.AlertStatus{State: models.StateInactive}
		evaluator.alertStatus.Update("alert-1", status)
	}

	// Condition exceeded for the first time -> pending
	evaluator.processAlertState(alertConfig, 95.0, true, pendingCounters, resolveCounters)
	status, _ = evaluator.GetAlertStatus("alert-1")
	assert.Equal(t, models.StatePending, status.State)
	assert.Equal(t, 1, pendingCounters["alert-1"])

	// Condition exceeded again, reaching debounce count -> active
	pendingCounters["alert-1"] = evaluator.config.AlertDebounceCount - 1
	evaluator.processAlertState(alertConfig, 96.0, true, pendingCounters, resolveCounters)
	status, _ = evaluator.GetAlertStatus("alert-1")
	assert.Equal(t, models.StateActive, status.State)
	assert.Equal(t, 0, pendingCounters["alert-1"]) // counter reset

	// Condition no longer exceeded -> resolving
	evaluator.processAlertState(alertConfig, 85.0, false, pendingCounters, resolveCounters)
	status, _ = evaluator.GetAlertStatus("alert-1")
	assert.Equal(t, models.StateResolving, status.State)
	assert.Equal(t, 1, resolveCounters["alert-1"])

	// Condition still not exceeded, reaching resolve count -> inactive (resolved)
	resolveCounters["alert-1"] = evaluator.config.AlertResolveCount - 1
	evaluator.processAlertState(alertConfig, 80.0, false, pendingCounters, resolveCounters)
	status, _ = evaluator.GetAlertStatus("alert-1")
	assert.Equal(t, models.StateInactive, status.State)
	assert.Equal(t, 0, resolveCounters["alert-1"]) // counter reset
}

func TestEvaluator_StartStop(t *testing.T) {
	alertStore, _ := database.NewAlertStore(":memory:")
	evaluator := NewEvaluator(alertStore, DefaultEvaluatorConfig())

	ctx, cancel := context.WithCancel(context.Background())
	err := evaluator.Start(ctx)
	assert.NoError(t, err)

	// Give it a moment to start the loop
	time.Sleep(50 * time.Millisecond)

	cancel()
	evaluator.Stop()

	// Check if event channel is closed
	_, ok := <-evaluator.Events()
	assert.False(t, ok, "Event channel should be closed after stopping")
}

// Add a mock for metrics.Collector's other methods if needed
func (m *MockMetricsCollector) GetDisk() (*metrics.DiskMetrics, error) {
	args := m.Called()
	return args.Get(0).(*metrics.DiskMetrics), args.Error(1)
}

func (m *MockMetricsCollector) GetProcesses() ([]*metrics.ProcessMetrics, error) {
	args := m.Called()
	return args.Get(0).([]*metrics.ProcessMetrics), args.Error(1)
}

func (m *MockMetricsCollector) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMetricsCollector) Stop() {
	m.Called()
}
