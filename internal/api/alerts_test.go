package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"argus/internal/database"
	"argus/internal/models"
)

// Original test setup - keeping for reference but enhanced with our mock approach
func setupTestEnvironment(t *testing.T) (*gin.Engine, *AlertsHandler, *database.AlertStore, func()) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "argus-test-*")
	require.NoError(t, err)

	// Create a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Create alert store
	alertStore, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create alerts handler with direct dependencies
	handler := &AlertsHandler{
		alertStore: alertStore,
	}

	// Set up the router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	apiGroup := router.Group("/api")
	handler.RegisterRoutes(apiGroup)

	return router, handler, alertStore, cleanup
}

// MockAlertStore is a mock implementation of the AlertStore
type MockAlertStore struct {
	mock.Mock
}

func (m *MockAlertStore) GetAlerts() ([]models.AlertConfig, error) {
	args := m.Called()
	return args.Get(0).([]models.AlertConfig), args.Error(1)
}

func (m *MockAlertStore) GetAlert(id string) (*models.AlertConfig, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AlertConfig), args.Error(1)
}

func (m *MockAlertStore) CreateAlert(alert models.AlertConfig) error {
	args := m.Called(alert)
	return args.Error(0)
}

func (m *MockAlertStore) UpdateAlert(id string, alert models.AlertConfig) error {
	args := m.Called(id, alert)
	return args.Error(0)
}

func (m *MockAlertStore) DeleteAlert(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAlertStore) GetAlertStatus(id string) (*models.AlertStatus, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AlertStatus), args.Error(1)
}

func (m *MockAlertStore) GetAllAlertStatus() (map[string]models.AlertStatus, error) {
	args := m.Called()
	return args.Get(0).(map[string]models.AlertStatus), args.Error(1)
}

// MockEvaluator is a mock implementation of the alert evaluator interface
type MockEvaluator struct {
	mock.Mock
}

func (m *MockEvaluator) EvaluateAlert(alertID string, metricValue float64) (bool, error) {
	args := m.Called(alertID, metricValue)
	return args.Bool(0), args.Error(1)
}

// MockNotifier is a mock implementation of the alert notifier interface
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Notify(alertID string, status models.AlertStatus) error {
	args := m.Called(alertID, status)
	return args.Error(0)
}

func (m *MockNotifier) GetNotifications(limit int) ([]models.Notification, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotifier) MarkNotificationRead(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockNotifier) MarkAllNotificationsRead() error {
	args := m.Called()
	return args.Error(0)
}

// setupMockAPI sets up a test environment with mocked dependencies
func setupMockAPI(t *testing.T) (*gin.Engine, *AlertsHandler, *MockAlertStore, *MockEvaluator, *MockNotifier) {
	gin.SetMode(gin.TestMode)

	mockStore := new(MockAlertStore)
	mockEval := new(MockEvaluator)
	mockNotifier := new(MockNotifier)

	handler := &AlertsHandler{
		alertStore: mockStore,
		evaluator:  mockEval,
		notifier:   mockNotifier,
	}

	router := gin.New()
	router.Use(gin.Recovery())

	apiGroup := router.Group("/api")
	handler.RegisterRoutes(apiGroup)

	return router, handler, mockStore, mockEval, mockNotifier
}

// Test getting all alerts
func TestListAlerts(t *testing.T) {
	router, _, mockStore, _, _ := setupMockAPI(t)

	// Mock the GetAlerts method
	alerts := []models.AlertConfig{
		{
			ID:          uuid.New().String(),
			Name:        "Test Alert 1",
			Description: "Test Description 1",
			Enabled:     true,
			Severity:    models.SeverityWarning,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Test Alert 2",
			Description: "Test Description 2",
			Enabled:     false,
			Severity:    models.SeverityCritical,
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		},
	}
	mockStore.On("GetAlerts").Return(alerts, nil)

	// Perform the request
	req := httptest.NewRequest(http.MethodGet, "/api/alerts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool                 `json:"success"`
		Data    []models.AlertConfig `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, alerts[0].Name, response.Data[0].Name)
	assert.Equal(t, alerts[1].Name, response.Data[1].Name)

	// Verify expectations
	mockStore.AssertExpectations(t)
}

// Test getting a specific alert
func TestGetAlert(t *testing.T) {
	router, _, mockStore, _, _ := setupMockAPI(t)

	// Create a test alert
	alertID := uuid.New().String()
	alert := &models.AlertConfig{
		ID:          alertID,
		Name:        "Test Alert",
		Description: "Test Description",
		Enabled:     true,
		Severity:    models.SeverityWarning,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
	}

	// Mock the GetAlert method
	mockStore.On("GetAlert", alertID).Return(alert, nil)

	// Perform the request
	req := httptest.NewRequest(http.MethodGet, "/api/alerts/"+alertID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool               `json:"success"`
		Data    models.AlertConfig `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, alertID, response.Data.ID)
	assert.Equal(t, alert.Name, response.Data.Name)

	// Verify expectations
	mockStore.AssertExpectations(t)
}

// Test getting a non-existent alert
func TestGetAlertNotFound(t *testing.T) {
	router, _, mockStore, _, _ := setupMockAPI(t)

	alertID := uuid.New().String()

	// Mock the GetAlert method to return an error
	mockStore.On("GetAlert", alertID).Return(nil, fmt.Errorf("alert not found"))

	// Perform the request
	req := httptest.NewRequest(http.MethodGet, "/api/alerts/"+alertID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "not found")

	// Verify expectations
	mockStore.AssertExpectations(t)
}

// Test creating a new alert
func TestCreateAlert(t *testing.T) {
	router, _, mockStore, _, _ := setupMockAPI(t)

	// Prepare the request data
	newAlert := models.AlertConfig{
		Name:        "New Test Alert",
		Description: "Test Description",
		Enabled:     true,
		Severity:    models.SeverityWarning,
		Threshold: models.ThresholdConfig{
			MetricType: models.MetricCPU,
			MetricName: "usage_percent",
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

	// Mock the CreateAlert method
	mockStore.On("CreateAlert", mock.AnythingOfType("models.AlertConfig")).Return(nil)

	// Prepare the request
	reqBody, err := json.Marshal(newAlert)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/alerts", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		Success bool               `json:"success"`
		Data    models.AlertConfig `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, newAlert.Name, response.Data.Name)
	assert.Equal(t, newAlert.Description, response.Data.Description)
	assert.Equal(t, newAlert.Threshold.MetricType, response.Data.Threshold.MetricType)

	// Verify expectations
	mockStore.AssertExpectations(t)
}

// Test updating an alert
func TestUpdateAlert(t *testing.T) {
	router, _, mockStore, _, _ := setupMockAPI(t)

	// Create a test alert ID
	alertID := uuid.New().String()

	// Prepare the request data
	updatedAlert := models.AlertConfig{
		ID:          alertID,
		Name:        "Updated Test Alert",
		Description: "Updated Description",
		Enabled:     false,
		Severity:    models.SeverityCritical,
		Threshold: models.ThresholdConfig{
			MetricType: models.MetricCPU,
			MetricName: "usage_percent",
			Operator:   models.OperatorGreaterThan,
			Value:      95.0,
		},
		Notifications: []models.NotificationConfig{
			{
				Type:    models.NotificationInApp,
				Enabled: true,
			},
		},
	}

	// Mock the GetAlert method
	existingAlert := &models.AlertConfig{
		ID:          alertID,
		Name:        "Test Alert",
		Description: "Test Description",
		Enabled:     true,
		Severity:    models.SeverityWarning,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
	}
	mockStore.On("GetAlert", alertID).Return(existingAlert, nil)

	// Mock the UpdateAlert method
	mockStore.On("UpdateAlert", alertID, mock.AnythingOfType("models.AlertConfig")).Return(nil)

	// Prepare the request
	reqBody, err := json.Marshal(updatedAlert)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/alerts/"+alertID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool               `json:"success"`
		Data    models.AlertConfig `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, updatedAlert.Name, response.Data.Name)
	assert.Equal(t, updatedAlert.Description, response.Data.Description)

	// Verify expectations
	mockStore.AssertExpectations(t)
}

// Test deleting an alert
func TestDeleteAlert(t *testing.T) {
	router, _, mockStore, _, _ := setupMockAPI(t)

	// Create a test alert ID
	alertID := uuid.New().String()

	// Mock the DeleteAlert method
	mockStore.On("DeleteAlert", alertID).Return(nil)

	// Perform the request
	req := httptest.NewRequest(http.MethodDelete, "/api/alerts/"+alertID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "deleted")

	// Verify expectations
	mockStore.AssertExpectations(t)
}

// Test getting all alert statuses
func TestGetAllAlertStatus(t *testing.T) {
	router, _, mockStore, _, _ := setupMockAPI(t)

	// Create test alert statuses
	statuses := map[string]models.AlertStatus{
		"alert-1": {
			AlertID:      "alert-1",
			State:        models.StateActive,
			CurrentValue: 95.5,
			Message:      "CPU usage exceeded threshold",
		},
		"alert-2": {
			AlertID:      "alert-2",
			State:        models.StateInactive,
			CurrentValue: 45.2,
			Message:      "Memory usage normal",
		},
	}

	// Mock the GetAllAlertStatus method
	mockStore.On("GetAllAlertStatus").Return(statuses, nil)

	// Perform the request
	req := httptest.NewRequest(http.MethodGet, "/api/alerts/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool                          `json:"success"`
		Data    map[string]models.AlertStatus `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, models.StateActive, response.Data["alert-1"].State)
	assert.Equal(t, models.StateInactive, response.Data["alert-2"].State)

	// Verify expectations
	mockStore.AssertExpectations(t)
}

// Test getting notifications
func TestGetNotifications(t *testing.T) {
	router, _, _, _, mockNotifier := setupMockAPI(t)

	// Create test notifications
	notifications := []models.Notification{
		{
			ID:        "notification-1",
			AlertID:   "alert-1",
			Title:     "High CPU Usage",
			Message:   "CPU usage exceeded 90%",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Read:      false,
		},
		{
			ID:        "notification-2",
			AlertID:   "alert-2",
			Title:     "Memory Warning",
			Message:   "Memory usage approaching threshold",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Read:      true,
		},
	}

	// Mock the GetNotifications method
	mockNotifier.On("GetNotifications", 50).Return(notifications, nil)

	// Perform the request
	req := httptest.NewRequest(http.MethodGet, "/api/alerts/notifications", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool                  `json:"success"`
		Data    []models.Notification `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, "notification-1", response.Data[0].ID)
	assert.Equal(t, "notification-2", response.Data[1].ID)

	// Verify expectations
	mockNotifier.AssertExpectations(t)
}
