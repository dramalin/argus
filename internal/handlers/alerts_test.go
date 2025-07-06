package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/database"
	"argus/internal/handlers"
	"argus/internal/models"
	"argus/internal/server"
	"argus/internal/services"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAlertsAPI(t *testing.T) {
	alertStore, err := database.NewAlertStore(":memory:")
	require.NoError(t, err)

	evaluator := services.NewEvaluator(alertStore, services.DefaultEvaluatorConfig())
	notifier := services.NewNotifier(services.DefaultConfig())
	alertsHandler := handlers.NewAlertsHandler(alertStore, evaluator, notifier)

	router := setupRouter()
	apiGroup := router.Group("/api")
	alertsHandler.RegisterRoutes(apiGroup)

	// Test Create Alert with Advanced Features
	t.Run("CreateAlertWithAdvancedFeatures", func(t *testing.T) {
		target := "test-process"
		alert := models.AlertConfig{
			Name:        "Test Process Alert",
			Enabled:     true,
			Severity:    models.SeverityCritical,
			Threshold: models.ThresholdConfig{
				MetricType:   models.MetricProcess,
				MetricName:   "cpu_percent",
				Operator:     models.OperatorGreaterThan,
				Value:        80,
				Target:       &target,
			},
			Notifications: []models.NotificationConfig{
				{
					Type:    models.NotificationEmail,
					Enabled: true,
					Settings: map[string]interface{}{
						"recipient": "test@example.com",
					},
				},
			},
		}

		body, _ := json.Marshal(alert)
		req, _ := http.NewRequest(http.MethodPost, "/api/alerts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response models.APIResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		var createdAlert models.AlertConfig
		dataBytes, _ := json.Marshal(response.Data)
		json.Unmarshal(dataBytes, &createdAlert)

		assert.Equal(t, "Test Process Alert", createdAlert.Name)
		assert.Equal(t, "test@example.com", createdAlert.Notifications[0].Settings["recipient"])
		assert.Equal(t, "test-process", *createdAlert.Threshold.Target)
	})

	// Test Update Alert with invalid settings
	t.Run("UpdateAlertWithInvalidSettings", func(t *testing.T) {
		// First create a valid alert
		validAlert := models.AlertConfig{
			Name: "Initial Valid Alert",
			Enabled: true,
			Severity: models.SeverityInfo,
			Threshold: models.ThresholdConfig{MetricType: models.MetricCPU, MetricName: "usage_percent", Operator: ">", Value: 10},
		}
		body, _ := json.Marshal(validAlert)
		req, _ := http.NewRequest(http.MethodPost, "/api/alerts", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)
		var createResponse models.APIResponse
		json.Unmarshal(rr.Body.Bytes(), &createResponse)
		var createdAlert models.AlertConfig
		dataBytes, _ := json.Marshal(createResponse.Data)
		json.Unmarshal(dataBytes, &createdAlert)
		
		// Now try to update it with invalid settings
		updateData := createdAlert
		updateData.Notifications = []models.NotificationConfig{
			{
				Type:    models.NotificationEmail,
				Enabled: true,
				Settings: map[string]interface{}{
					"recipient": "", // Invalid empty recipient
				},
			},
		}

		updateBody, _ := json.Marshal(updateData)
		req, _ = http.NewRequest(http.MethodPut, "/api/alerts/"+createdAlert.ID, bytes.NewBuffer(updateBody))
		req.Header.Set("Content-Type", "application/json")
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errorResponse models.APIResponse
		json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		assert.False(t, errorResponse.Success)
		assert.Contains(t, errorResponse.Error, "email recipient must be a non-empty string")
	})
} 