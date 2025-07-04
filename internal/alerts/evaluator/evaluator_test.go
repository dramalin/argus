package evaluator

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/database"
	"argus/internal/models"
)

func setupTestAlertStore(t *testing.T) (*database.AlertStore, string) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "evaluator_test")
	require.NoError(t, err)

	// Create a new alert store
	alertStore, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	return alertStore, tempDir
}

func createTestAlertConfig(t *testing.T, alertStore *database.AlertStore, metricType models.MetricType, operator models.ComparisonOperator, value float64) *models.AlertConfig {
	// Create a test alert configuration
	alert := &models.AlertConfig{
		ID:          "test-alert-1",
		Name:        "Test Alert",
		Description: "Test alert for evaluator",
		Enabled:     true,
		Severity:    models.SeverityWarning,
		Threshold: models.ThresholdConfig{
			MetricType: metricType,
			MetricName: "usage_percent",
			Operator:   operator,
			Value:      value,
		},
		Notifications: []models.NotificationConfig{
			{
				Type:    models.NotificationInApp,
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
	alert := createTestAlertConfig(t, alertStore, models.MetricCPU, models.OperatorGreaterThan, 90.0)

	// Create an evaluator
	evaluator := NewEvaluator(alertStore, nil)

	// Initialize alert status
	err := evaluator.initAlertStatus()
	require.NoError(t, err)

	// Check that the alert status was initialized
	evaluator.statusMu.RLock()
	defer evaluator.statusMu.RUnlock()
	status, exists := evaluator.alertStatus[alert.ID]
	assert.True(t, exists)
	assert.NotNil(t, status)
	assert.Equal(t, alert.ID, status.AlertID)
	assert.Equal(t, models.StateInactive, status.State)
}

func TestEvaluator_CompareValue(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)

	evaluator := NewEvaluator(alertStore, nil)

	// Test comparison operators
	assert.True(t, evaluator.compareValue(91.0, 90.0, models.OperatorGreaterThan))
	assert.False(t, evaluator.compareValue(90.0, 90.0, models.OperatorGreaterThan))
	assert.False(t, evaluator.compareValue(89.0, 90.0, models.OperatorGreaterThan))

	// Test greater than or equal
	assert.True(t, evaluator.compareValue(91.0, 90.0, models.OperatorGreaterThanOrEqual))
	assert.True(t, evaluator.compareValue(90.0, 90.0, models.OperatorGreaterThanOrEqual))
	assert.False(t, evaluator.compareValue(89.0, 90.0, models.OperatorGreaterThanOrEqual))

	// Test less than
	assert.True(t, evaluator.compareValue(89.0, 90.0, models.OperatorLessThan))
	assert.False(t, evaluator.compareValue(90.0, 90.0, models.OperatorLessThan))
	assert.False(t, evaluator.compareValue(91.0, 90.0, models.OperatorLessThan))

	// Test less than or equal
	assert.True(t, evaluator.compareValue(89.0, 90.0, models.OperatorLessThanOrEqual))
	assert.True(t, evaluator.compareValue(90.0, 90.0, models.OperatorLessThanOrEqual))
	assert.False(t, evaluator.compareValue(91.0, 90.0, models.OperatorLessThanOrEqual))

	// Test equal
	assert.True(t, evaluator.compareValue(90.0, 90.0, models.OperatorEqual))
	assert.False(t, evaluator.compareValue(89.0, 90.0, models.OperatorEqual))
	assert.False(t, evaluator.compareValue(91.0, 90.0, models.OperatorEqual))

	// Test not equal
	assert.True(t, evaluator.compareValue(91.0, 90.0, models.OperatorNotEqual))
	assert.True(t, evaluator.compareValue(89.0, 90.0, models.OperatorNotEqual))
	assert.False(t, evaluator.compareValue(90.0, 90.0, models.OperatorNotEqual))

	// Test invalid operator
	assert.False(t, evaluator.compareValue(90.0, 90.0, "invalid"))
}

// Test evaluator start and stop
func TestEvaluator_StartStop(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)

	// Create a test alert
	createTestAlertConfig(t, alertStore, models.MetricCPU, models.OperatorGreaterThan, 90.0)

	// Create an evaluator
	evaluator := NewEvaluator(alertStore, nil)

	// Start the evaluator
	ctx, cancel := context.WithCancel(context.Background())
	err := evaluator.Start(ctx)
	require.NoError(t, err)

	// Short delay to allow evaluation to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to stop evaluation
	cancel()

	// Stop the evaluator
	evaluator.Stop()
}

// Test event channel access
func TestEvaluator_Events(t *testing.T) {
	alertStore, tempDir := setupTestAlertStore(t)
	defer os.RemoveAll(tempDir)

	evaluator := NewEvaluator(alertStore, nil)

	// Get the event channel
	eventCh := evaluator.Events()
	assert.NotNil(t, eventCh)
}
