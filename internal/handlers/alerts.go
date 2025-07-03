// Package handlers provides HTTP API handlers for the Argus System Monitor
package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"argus/internal/database"
	"argus/internal/models"
	"argus/internal/services"
)

// AlertsHandler manages alert-related API endpoints
type AlertsHandler struct {
	alertStore *database.AlertStore
	evaluator  *services.Evaluator
	notifier   *services.Notifier
}

// NewAlertsHandler creates a new alerts API handler
func NewAlertsHandler(alertStore *database.AlertStore, evaluator *services.Evaluator, notifier *services.Notifier) *AlertsHandler {
	return &AlertsHandler{
		alertStore: alertStore,
		evaluator:  evaluator,
		notifier:   notifier,
	}
}

// RegisterRoutes registers all alert-related routes to the given router group
func (h *AlertsHandler) RegisterRoutes(router *gin.RouterGroup) {
	alerts := router.Group("/alerts")
	{
		// Alert configuration endpoints
		alerts.GET("", h.ListAlerts)
		alerts.GET("/:id", h.GetAlert)
		alerts.POST("", h.CreateAlert)
		alerts.PUT("/:id", h.UpdateAlert)
		alerts.DELETE("/:id", h.DeleteAlert)

		// Alert status endpoints
		alerts.GET("/status", h.GetAllAlertStatus)
		alerts.GET("/status/:id", h.GetAlertStatus)

		// Notification endpoints
		alerts.GET("/notifications", h.GetNotifications)
		alerts.POST("/notifications/:id/read", h.MarkNotificationRead)
		alerts.POST("/notifications/read-all", h.MarkAllNotificationsRead)
		alerts.DELETE("/notifications", h.ClearNotifications)

		// Test endpoint
		alerts.POST("/test/:id", h.TestAlert)
	}
}

// ListAlerts returns all alert configurations
func (h *AlertsHandler) ListAlerts(c *gin.Context) {
	slog.Debug("Fetching all alert configurations")

	alerts, err := h.alertStore.ListAlerts()
	if err != nil {
		slog.Error("Failed to list alerts", "error", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to list alerts: " + err.Error()})
		return
	}

	slog.Debug("Alert configurations retrieved successfully", "count", len(alerts))
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: alerts})
}

// GetAlert returns a specific alert configuration by ID
func (h *AlertsHandler) GetAlert(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Fetching alert configuration", "id", id)

	alert, err := h.alertStore.GetAlert(id)
	if err != nil {
		if err == database.ErrAlertNotFound {
			slog.Debug("Alert not found", "id", id)
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "Alert not found"})
			return
		}
		slog.Error("Failed to get alert", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to get alert: " + err.Error()})
		return
	}

	slog.Debug("Alert configuration retrieved successfully", "id", id)
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: alert})
}

// CreateAlert creates a new alert configuration
func (h *AlertsHandler) CreateAlert(c *gin.Context) {
	var alert models.AlertConfig
	if err := c.ShouldBindJSON(&alert); err != nil {
		slog.Debug("Invalid alert configuration data", "error", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Invalid alert configuration: " + err.Error()})
		return
	}

	// Generate a new UUID if ID is empty
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	alert.CreatedAt = now
	alert.UpdatedAt = now

	// Validate the alert configuration
	if err := alert.Validate(); err != nil {
		slog.Debug("Invalid alert configuration", "error", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Invalid alert configuration: " + err.Error()})
		return
	}

	// Store the alert
	if err := h.alertStore.CreateAlert(&alert); err != nil {
		slog.Error("Failed to create alert", "error", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to create alert: " + err.Error()})
		return
	}

	slog.Info("Alert created successfully", "id", alert.ID, "name", alert.Name)
	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: alert})
}

// UpdateAlert updates an existing alert configuration
func (h *AlertsHandler) UpdateAlert(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Updating alert configuration", "id", id)

	// Check if alert exists
	existingAlert, err := h.alertStore.GetAlert(id)
	if err != nil {
		if err == database.ErrAlertNotFound {
			slog.Debug("Alert not found for update", "id", id)
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "Alert not found"})
			return
		}
		slog.Error("Failed to get alert for update", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to get alert: " + err.Error()})
		return
	}

	// Parse update data
	var alert models.AlertConfig
	if err := c.ShouldBindJSON(&alert); err != nil {
		slog.Debug("Invalid alert update data", "error", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Invalid alert configuration: " + err.Error()})
		return
	}

	// Ensure ID matches
	alert.ID = id
	alert.CreatedAt = existingAlert.CreatedAt
	alert.UpdatedAt = time.Now()

	// Validate the alert configuration
	if err := alert.Validate(); err != nil {
		slog.Debug("Invalid alert update", "error", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Invalid alert configuration: " + err.Error()})
		return
	}

	// Update the alert
	if err := h.alertStore.UpdateAlert(&alert); err != nil {
		slog.Error("Failed to update alert", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to update alert: " + err.Error()})
		return
	}

	slog.Info("Alert updated successfully", "id", id, "name", alert.Name)
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: alert})
}

// DeleteAlert deletes an alert configuration
func (h *AlertsHandler) DeleteAlert(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Deleting alert configuration", "id", id)

	// Check if alert exists
	_, err := h.alertStore.GetAlert(id)
	if err != nil {
		if err == database.ErrAlertNotFound {
			slog.Debug("Alert not found for deletion", "id", id)
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "Alert not found"})
			return
		}
		slog.Error("Failed to get alert for deletion", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to get alert: " + err.Error()})
		return
	}

	// Delete the alert
	if err := h.alertStore.DeleteAlert(id); err != nil {
		slog.Error("Failed to delete alert", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to delete alert: " + err.Error()})
		return
	}

	slog.Info("Alert deleted successfully", "id", id)
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: gin.H{"message": "Alert deleted successfully"}})
}

// GetAllAlertStatus returns the current status of all alerts
func (h *AlertsHandler) GetAllAlertStatus(c *gin.Context) {
	slog.Debug("Fetching all alert statuses")

	statuses := h.evaluator.GetAllAlertStatus()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: statuses})
}

// GetAlertStatus returns the current status of a specific alert
func (h *AlertsHandler) GetAlertStatus(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Fetching alert status", "id", id)

	status, found := h.evaluator.GetAlertStatus(id)
	if !found {
		slog.Debug("Alert status not found", "id", id)
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "Alert status not found"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: status})
}

// GetNotifications returns all in-app notifications
func (h *AlertsHandler) GetNotifications(c *gin.Context) {
	slog.Debug("Fetching in-app notifications")

	notifications := h.notifier.GetNotifications()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: notifications})
}

// MarkNotificationRead marks a notification as read
func (h *AlertsHandler) MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Marking notification as read", "id", id)

	if !h.notifier.MarkNotificationRead(id) {
		slog.Debug("Notification not found", "id", id)
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: gin.H{"message": "Notification marked as read"}})
}

// MarkAllNotificationsRead marks all notifications as read
func (h *AlertsHandler) MarkAllNotificationsRead(c *gin.Context) {
	slog.Debug("Marking all notifications as read")

	h.notifier.MarkAllNotificationsRead()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: gin.H{"message": "All notifications marked as read"}})
}

// ClearNotifications removes all notifications
func (h *AlertsHandler) ClearNotifications(c *gin.Context) {
	slog.Debug("Clearing all notifications")

	h.notifier.ClearNotifications()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: gin.H{"message": "All notifications cleared"}})
}

// TestAlert tests an alert by simulating an alert event
func (h *AlertsHandler) TestAlert(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Testing alert", "id", id)

	// Get the alert configuration
	alertConfig, err := h.alertStore.GetAlert(id)
	if err != nil {
		if err == database.ErrAlertNotFound {
			slog.Debug("Alert not found for testing", "id", id)
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "Alert not found"})
			return
		}
		slog.Error("Failed to get alert for testing", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to get alert: " + err.Error()})
		return
	}

	// Create a test alert event
	now := time.Now()
	testValue := alertConfig.Threshold.Value + 1 // Ensure it exceeds the threshold

	testEvent := models.AlertEvent{
		AlertID:      alertConfig.ID,
		OldState:     models.StateInactive,
		NewState:     models.StateActive,
		CurrentValue: testValue,
		Threshold:    alertConfig.Threshold.Value,
		Timestamp:    now,
		Message:      "This is a test alert notification",
		Alert:        alertConfig,
		Status: &models.AlertStatus{
			AlertID:      alertConfig.ID,
			State:        models.StateActive,
			CurrentValue: testValue,
			TriggeredAt:  &now,
			Message:      "Test alert triggered",
		},
	}

	// Process the test event
	h.notifier.ProcessEvent(testEvent)

	slog.Info("Test alert sent successfully", "id", id, "name", alertConfig.Name)
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: gin.H{
		"message": "Test alert sent successfully",
		"event":   testEvent,
	}})
}
