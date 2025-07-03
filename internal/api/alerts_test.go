package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/alerts/evaluator"
	"argus/internal/alerts/notifier"
	"argus/internal/models"
	"argus/internal/storage"
)

func setupTestEnvironment(t *testing.T) (*gin.Engine, *AlertsHandler, *storage.AlertStore, func()) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "argus-test-*")
	require.NoError(t, err)

	// Create a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Create alert store
	alertStore, err := storage.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create evaluator
	evalConfig := evaluator.DefaultConfig()
	eval := evaluator.NewEvaluator(alertStore, evalConfig)

	// Create notifier
	notifierConfig := notifier.DefaultConfig()
	n := notifier.NewNotifier(notifierConfig)

	// Register in-app notification channel
	inAppChannel := notifier.NewInAppChannel(100)
	n.RegisterChannel(inAppChannel)

	// Create alerts handler
	handler := NewAlertsHandler(alertStore, eval, n)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Register routes
	api := router.Group("/api")
	handler.RegisterRoutes(api)

	return router, handler, alertStore, cleanup
}

func createTestAlert(t *testing.T, alertStore *storage.AlertStore) *models.AlertConfig {
	alert := &models.AlertConfig{
		ID:          uuid.New().String(),
		Name:        "Test Alert",
		Description: "Test alert for CPU usage",
		Enabled:     true,
		Severity:    models.SeverityWarning,
		Threshold: models.ThresholdConfig{
			MetricType: models.MetricCPU,
			MetricName: "usage_percent",
			Operator:   models.OperatorGreaterThan,
			Value:      80.0,
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

	err := alertStore.CreateAlert(alert)
	require.NoError(t, err)

	return alert
}

func TestListAlerts(t *testing.T) {
	router, _, alertStore, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test alerts
	alert1 := createTestAlert(t, alertStore)
	alert2 := createTestAlert(t, alertStore)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/alerts", nil)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []*models.AlertConfig
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response contains the created alerts
	assert.Len(t, response, 2)

	// Check if the response contains our alerts (by ID)
	foundAlert1 := false
	foundAlert2 := false

	for _, alert := range response {
		if alert.ID == alert1.ID {
			foundAlert1 = true
		}
		if alert.ID == alert2.ID {
			foundAlert2 = true
		}
	}

	assert.True(t, foundAlert1, "Alert 1 should be in the response")
	assert.True(t, foundAlert2, "Alert 2 should be in the response")
}

func TestGetAlert(t *testing.T) {
	router, _, alertStore, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test alert
	alert := createTestAlert(t, alertStore)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/alerts/"+alert.ID, nil)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.AlertConfig
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response contains the correct alert
	assert.Equal(t, alert.ID, response.ID)
	assert.Equal(t, alert.Name, response.Name)
}

func TestGetAlert_NotFound(t *testing.T) {
	router, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Make request with non-existent ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/alerts/non-existent-id", nil)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateAlert(t *testing.T) {
	router, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create alert payload
	alert := models.AlertConfig{
		Name:        "New Test Alert",
		Description: "Test alert for memory usage",
		Enabled:     true,
		Severity:    models.SeverityWarning,
		Threshold: models.ThresholdConfig{
			MetricType: models.MetricMemory,
			MetricName: "used_percent",
			Operator:   models.OperatorGreaterThan,
			Value:      90.0,
		},
		Notifications: []models.NotificationConfig{
			{
				Type:    models.NotificationInApp,
				Enabled: true,
			},
		},
	}

	payload, err := json.Marshal(alert)
	require.NoError(t, err)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/alerts", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.AlertConfig
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response contains the created alert
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, alert.Name, response.Name)
	assert.Equal(t, alert.Description, response.Description)
	assert.Equal(t, alert.Severity, response.Severity)
	assert.Equal(t, alert.Threshold.MetricType, response.Threshold.MetricType)
	assert.Equal(t, alert.Threshold.Value, response.Threshold.Value)
}

func TestCreateAlert_InvalidPayload(t *testing.T) {
	router, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create invalid alert payload (missing required fields)
	alert := models.AlertConfig{
		Name: "Invalid Alert",
		// Missing severity and threshold
	}

	payload, err := json.Marshal(alert)
	require.NoError(t, err)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/alerts", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateAlert(t *testing.T) {
	router, _, alertStore, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test alert
	alert := createTestAlert(t, alertStore)

	// Update alert
	alert.Name = "Updated Alert Name"
	alert.Description = "Updated description"
	alert.Threshold.Value = 95.0

	payload, err := json.Marshal(alert)
	require.NoError(t, err)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/alerts/"+alert.ID, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.AlertConfig
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response contains the updated alert
	assert.Equal(t, alert.ID, response.ID)
	assert.Equal(t, "Updated Alert Name", response.Name)
	assert.Equal(t, "Updated description", response.Description)
	assert.Equal(t, 95.0, response.Threshold.Value)
}

func TestUpdateAlert_NotFound(t *testing.T) {
	router, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create update payload
	alert := models.AlertConfig{
		Name:        "Updated Alert",
		Description: "This alert doesn't exist",
		Severity:    models.SeverityWarning,
		Threshold: models.ThresholdConfig{
			MetricType: models.MetricCPU,
			MetricName: "usage_percent",
			Operator:   models.OperatorGreaterThan,
			Value:      85.0,
		},
	}

	payload, err := json.Marshal(alert)
	require.NoError(t, err)

	// Make request with non-existent ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/alerts/non-existent-id", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteAlert(t *testing.T) {
	router, _, alertStore, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test alert
	alert := createTestAlert(t, alertStore)

	// Make delete request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/alerts/"+alert.ID, nil)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify alert was deleted
	_, err := alertStore.GetAlert(alert.ID)
	assert.Equal(t, storage.ErrAlertNotFound, err)
}

func TestDeleteAlert_NotFound(t *testing.T) {
	router, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Make request with non-existent ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/alerts/non-existent-id", nil)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetNotifications(t *testing.T) {
	router, handler, alertStore, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test alert
	alert := createTestAlert(t, alertStore)

	// Create a test event
	now := time.Now()
	testEvent := models.AlertEvent{
		AlertID:      alert.ID,
		OldState:     models.StateInactive,
		NewState:     models.StateActive,
		CurrentValue: 85.0,
		Threshold:    alert.Threshold.Value,
		Timestamp:    now,
		Message:      "Test alert notification",
		Alert:        alert,
		Status: &models.AlertStatus{
			AlertID:      alert.ID,
			State:        models.StateActive,
			CurrentValue: 85.0,
			TriggeredAt:  &now,
			Message:      "Test alert triggered",
		},
	}

	// Process the test event to create a notification
	handler.notifier.ProcessEvent(testEvent)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/alerts/notifications", nil)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.InAppNotification
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response contains the notification
	assert.Len(t, response, 1)
	assert.Equal(t, alert.ID, response[0].AlertID)
	assert.Equal(t, alert.Name, response[0].AlertName)
	assert.Equal(t, models.StateActive, response[0].State)
	assert.False(t, response[0].Read)
}

func TestMarkNotificationRead(t *testing.T) {
	router, handler, alertStore, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test alert
	alert := createTestAlert(t, alertStore)

	// Create a test event
	now := time.Now()
	testEvent := models.AlertEvent{
		AlertID:      alert.ID,
		OldState:     models.StateInactive,
		NewState:     models.StateActive,
		CurrentValue: 85.0,
		Threshold:    alert.Threshold.Value,
		Timestamp:    now,
		Message:      "Test alert notification",
		Alert:        alert,
		Status: &models.AlertStatus{
			AlertID:      alert.ID,
			State:        models.StateActive,
			CurrentValue: 85.0,
			TriggeredAt:  &now,
			Message:      "Test alert triggered",
		},
	}

	// Process the test event to create a notification
	handler.notifier.ProcessEvent(testEvent)

	// Get the notification ID
	channel, _ := handler.notifier.GetChannel(models.NotificationInApp)
	inAppChannel := channel.(*notifier.InAppChannel)
	notifications := inAppChannel.GetNotifications()
	require.Len(t, notifications, 1)
	notificationID := notifications[0].ID

	// Make request to mark as read
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/alerts/notifications/"+notificationID+"/read", nil)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify notification is marked as read
	notifications = inAppChannel.GetNotifications()
	require.Len(t, notifications, 1)
	assert.True(t, notifications[0].Read)
}

func TestTestAlert(t *testing.T) {
	router, _, alertStore, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test alert
	alert := createTestAlert(t, alertStore)

	// Make request to test alert
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/alerts/test/"+alert.ID, nil)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response contains success message
	assert.Contains(t, response, "message")
	assert.Equal(t, "Test alert sent successfully", response["message"])
}
