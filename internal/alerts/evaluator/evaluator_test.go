package evaluator

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/alerts"
	"argus/internal/storage"
)

func setupTestAlertStore(t *testing.T) (*storage.AlertStore, string) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "evaluator_test")
	require.NoError(t, err)

	// Create a new alert store
	alertStore, err := storage.NewAlertStore(tempDir)
	require.NoError(t, err)

	return alertStore, tempDir
}

func createTestAlertConfig(t *testing.T, alertStore *storage.AlertStore, metricType alerts.MetricType, operator alerts.ComparisonOperator, value float64) *alerts.AlertConfig {
	// Create a test alert configuration
	alert := &alerts.AlertConfig{
		ID:          "test-alert-1",
		Name:        "Test Alert",
		Description: "Test alert for evaluator",
		Enabled:     true,
		Severity:    alerts.SeverityWarning,
		Threshold: alerts.ThresholdConfig{
			MetricType: metricType,
			MetricName: "usage_percent",
			Operator:   operator,
			Value:      value,
		},
		Notifications: []alerts.NotificationConfig{
			{
				Type:    alerts.NotificationInApp,
				Enabled: true,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store the alert
	err := alertStore.CreateAlert(alert)
	require.NoError(t, err)

	return alert
}

func TestNewEvaluator(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)

	// Test with default config
	evaluator := NewEvaluator(alertStore, nil)
	assert.NotNil(t, evaluator)
	assert.Equal(t, DefaultEvaluationInterval, evaluator.config.EvaluationInterval)
	assert.Equal(t, DefaultAlertDebounceCount, evaluator.config.AlertDebounceCount)
	assert.Equal(t, DefaultAlertResolveCount, evaluator.config.AlertResolveCount)

	// Test with custom config
	customConfig := &Config{
		EvaluationInterval: 10 * time.Second,
		AlertDebounceCount: 3,
		AlertResolveCount:  4,
	}
	evaluator = NewEvaluator(alertStore, customConfig)
	assert.NotNil(t, evaluator)
	assert.Equal(t, customConfig.EvaluationInterval, evaluator.config.EvaluationInterval)
	assert.Equal(t, customConfig.AlertDebounceCount, evaluator.config.AlertDebounceCount)
	assert.Equal(t, customConfig.AlertResolveCount, evaluator.config.AlertResolveCount)
}

func TestEvaluator_InitAlertStatus(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)

	// Create a test alert
	alert := createTestAlertConfig(t, alertStore, alerts.MetricCPU, alerts.OperatorGreaterThan, 90.0)

	// Create an evaluator
	evaluator := NewEvaluator(alertStore, nil)

	// Initialize alert status
	err := evaluator.initAlertStatus()
	require.NoError(t, err)

	// Check that the alert status was initialized
	status, exists := evaluator.GetAlertStatus(alert.ID)
	assert.True(t, exists)
	assert.Equal(t, alert.ID, status.AlertID)
	assert.Equal(t, alerts.StateInactive, status.State)
}

func TestEvaluator_CompareValue(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)
	evaluator := NewEvaluator(alertStore, nil)

	// Test greater than
	assert.True(t, evaluator.compareValue(91.0, 90.0, alerts.OperatorGreaterThan))
	assert.False(t, evaluator.compareValue(90.0, 90.0, alerts.OperatorGreaterThan))
	assert.False(t, evaluator.compareValue(89.0, 90.0, alerts.OperatorGreaterThan))

	// Test greater than or equal
	assert.True(t, evaluator.compareValue(91.0, 90.0, alerts.OperatorGreaterThanOrEqual))
	assert.True(t, evaluator.compareValue(90.0, 90.0, alerts.OperatorGreaterThanOrEqual))
	assert.False(t, evaluator.compareValue(89.0, 90.0, alerts.OperatorGreaterThanOrEqual))

	// Test less than
	assert.False(t, evaluator.compareValue(91.0, 90.0, alerts.OperatorLessThan))
	assert.False(t, evaluator.compareValue(90.0, 90.0, alerts.OperatorLessThan))
	assert.True(t, evaluator.compareValue(89.0, 90.0, alerts.OperatorLessThan))

	// Test less than or equal
	assert.False(t, evaluator.compareValue(91.0, 90.0, alerts.OperatorLessThanOrEqual))
	assert.True(t, evaluator.compareValue(90.0, 90.0, alerts.OperatorLessThanOrEqual))
	assert.True(t, evaluator.compareValue(89.0, 90.0, alerts.OperatorLessThanOrEqual))

	// Test equal
	assert.False(t, evaluator.compareValue(91.0, 90.0, alerts.OperatorEqual))
	assert.True(t, evaluator.compareValue(90.0, 90.0, alerts.OperatorEqual))
	assert.False(t, evaluator.compareValue(89.0, 90.0, alerts.OperatorEqual))

	// Test not equal
	assert.True(t, evaluator.compareValue(91.0, 90.0, alerts.OperatorNotEqual))
	assert.False(t, evaluator.compareValue(90.0, 90.0, alerts.OperatorNotEqual))
	assert.True(t, evaluator.compareValue(89.0, 90.0, alerts.OperatorNotEqual))

	// Test unknown operator
	assert.False(t, evaluator.compareValue(90.0, 90.0, "unknown"))
}

func TestEvaluator_StartStop(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)

	// Create a test alert
	createTestAlertConfig(t, alertStore, alerts.MetricCPU, alerts.OperatorGreaterThan, 90.0)

	// Create an evaluator with a short evaluation interval for testing
	evaluator := NewEvaluator(alertStore, &Config{
		EvaluationInterval: 100 * time.Millisecond,
		AlertDebounceCount: 1,
		AlertResolveCount:  1,
	})

	// Start the evaluator
	ctx, cancel := context.WithCancel(context.Background())
	err := evaluator.Start(ctx)
	require.NoError(t, err)

	// Wait for a short time to allow some evaluations to occur
	time.Sleep(300 * time.Millisecond)

	// Cancel the context and stop the evaluator
	cancel()
	evaluator.Stop()

	// The evaluator should have stopped gracefully
}

func TestEvaluator_GetAllAlertStatus(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)

	// Create multiple test alerts
	alert1 := createTestAlertConfig(t, alertStore, alerts.MetricCPU, alerts.OperatorGreaterThan, 90.0)
	alert1.ID = "test-alert-1"
	err := alertStore.UpdateAlert(alert1)
	require.NoError(t, err)

	alert2 := createTestAlertConfig(t, alertStore, alerts.MetricMemory, alerts.OperatorGreaterThan, 80.0)
	alert2.ID = "test-alert-2"
	err = alertStore.UpdateAlert(alert2)
	require.NoError(t, err)

	// Create an evaluator
	evaluator := NewEvaluator(alertStore, nil)

	// Initialize alert status
	err = evaluator.initAlertStatus()
	require.NoError(t, err)

	// Get all alert status
	allStatus := evaluator.GetAllAlertStatus()
	assert.Equal(t, 2, len(allStatus))
	assert.Contains(t, allStatus, "test-alert-1")
	assert.Contains(t, allStatus, "test-alert-2")
}

// MockMetricEvaluator is a mock implementation of the evaluator for testing
type MockMetricEvaluator struct {
	*Evaluator
	mockValue float64
	mockErr   error
}

func NewMockEvaluator(alertStore *storage.AlertStore, mockValue float64, mockErr error) *MockMetricEvaluator {
	return &MockMetricEvaluator{
		Evaluator: NewEvaluator(alertStore, nil),
		mockValue: mockValue,
		mockErr:   mockErr,
	}
}

// Override the evaluateMetric method for testing
func (m *MockMetricEvaluator) evaluateMetric(threshold alerts.ThresholdConfig) (float64, error) {
	return m.mockValue, m.mockErr
}

func TestEvaluator_StateTransitions(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)

	// Create a test alert with a threshold of 90.0
	alert := createTestAlertConfig(t, alertStore, alerts.MetricCPU, alerts.OperatorGreaterThan, 90.0)

	// Create a mock evaluator that always returns a value of 95.0 (above threshold)
	evaluator := NewMockEvaluator(alertStore, 95.0, nil)

	// Initialize alert status
	err := evaluator.initAlertStatus()
	require.NoError(t, err)

	// Get the initial status
	status, exists := evaluator.GetAlertStatus(alert.ID)
	require.True(t, exists)
	assert.Equal(t, alerts.StateInactive, status.State)

	// Create a test context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start collecting events
	events := make([]AlertEvent, 0)
	go func() {
		for event := range evaluator.Events() {
			events = append(events, event)
		}
	}()

	// Manually trigger the evaluation loop once
	evaluator.evaluationLoop(ctx)

	// Check that the state has changed to pending
	status, exists = evaluator.GetAlertStatus(alert.ID)
	require.True(t, exists)
	assert.Equal(t, alerts.StatePending, status.State)

	// Change the mock value to below threshold
	evaluator.mockValue = 85.0

	// Trigger the evaluation loop again
	evaluator.evaluationLoop(ctx)

	// Check that the state has changed back to inactive
	status, exists = evaluator.GetAlertStatus(alert.ID)
	require.True(t, exists)
	assert.Equal(t, alerts.StateInactive, status.State)

	// Change the mock value back to above threshold
	evaluator.mockValue = 95.0

	// Set the threshold duration to 0 for immediate activation
	alert.Threshold.Duration = 0
	err = alertStore.UpdateAlert(alert)
	require.NoError(t, err)

	// Trigger the evaluation loop again
	evaluator.evaluationLoop(ctx)

	// Check that the state has changed to pending
	status, exists = evaluator.GetAlertStatus(alert.ID)
	require.True(t, exists)
	assert.Equal(t, alerts.StatePending, status.State)

	// Trigger the evaluation loop again to activate the alert
	evaluator.evaluationLoop(ctx)

	// Check that the state has changed to active
	status, exists = evaluator.GetAlertStatus(alert.ID)
	require.True(t, exists)
	assert.Equal(t, alerts.StateActive, status.State)
	assert.NotNil(t, status.TriggeredAt)

	// Change the mock value to below threshold
	evaluator.mockValue = 85.0

	// Trigger the evaluation loop again
	evaluator.evaluationLoop(ctx)

	// Check that the state is still active (due to debouncing)
	status, exists = evaluator.GetAlertStatus(alert.ID)
	require.True(t, exists)
	assert.Equal(t, alerts.StateActive, status.State)

	// Trigger the evaluation loop again to resolve the alert
	evaluator.evaluationLoop(ctx)

	// Check that the state has changed to inactive
	status, exists = evaluator.GetAlertStatus(alert.ID)
	require.True(t, exists)
	assert.Equal(t, alerts.StateInactive, status.State)
	assert.NotNil(t, status.ResolvedAt)
}
