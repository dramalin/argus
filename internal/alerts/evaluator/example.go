// Package evaluator provides functionality for evaluating alert conditions against system metrics
package evaluator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"argus/internal/alerts"
	"argus/internal/storage"
)

// ExampleUsage demonstrates how to use the evaluator in a main application
func ExampleUsage() {
	// Set up structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Create alert store
	alertStore, err := storage.NewAlertStore(".argus/config")
	if err != nil {
		slog.Error("Failed to create alert store", "error", err)
		os.Exit(1)
	}

	// Create evaluator with custom configuration
	config := &Config{
		EvaluationInterval: 30 * time.Second,
		AlertDebounceCount: 2,
		AlertResolveCount:  3,
	}
	evaluator := NewEvaluator(alertStore, config)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the evaluator
	if err := evaluator.Start(ctx); err != nil {
		slog.Error("Failed to start evaluator", "error", err)
		os.Exit(1)
	}

	// Start a goroutine to process alert events
	go func() {
		for event := range evaluator.Events() {
			handleAlertEvent(event)
		}
	}()

	slog.Info("Alert evaluator started successfully")

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Cancel context and stop evaluator
	cancel()
	evaluator.Stop()

	slog.Info("Alert evaluator stopped successfully")
}

// handleAlertEvent processes alert events
func handleAlertEvent(event AlertEvent) {
	switch event.NewState {
	case "active":
		slog.Warn("Alert activated",
			"alert_id", event.AlertID,
			"alert_name", event.Alert.Name,
			"severity", event.Alert.Severity,
			"value", event.CurrentValue,
			"threshold", event.Threshold,
			"message", event.Message)

		// Here you would dispatch notifications based on the alert configuration
		// For example:
		for _, notification := range event.Alert.Notifications {
			if notification.Enabled {
				dispatchNotification(string(notification.Type), event)
			}
		}

	case "inactive":
		if event.OldState == "active" {
			slog.Info("Alert resolved",
				"alert_id", event.AlertID,
				"alert_name", event.Alert.Name,
				"value", event.CurrentValue,
				"threshold", event.Threshold,
				"message", event.Message)

			// Here you would send resolution notifications
		}

	case "pending":
		slog.Info("Alert pending",
			"alert_id", event.AlertID,
			"alert_name", event.Alert.Name,
			"value", event.CurrentValue,
			"threshold", event.Threshold)
	}
}

// dispatchNotification sends a notification based on the notification type
func dispatchNotification(notificationType string, event AlertEvent) {
	// This is a placeholder for the actual notification dispatch logic
	// In a real implementation, this would send emails, push notifications, etc.
	slog.Info("Dispatching notification",
		"type", notificationType,
		"alert_id", event.AlertID,
		"alert_name", event.Alert.Name,
		"message", event.Message)

	switch notificationType {
	case string(alerts.NotificationInApp):
		fmt.Println("Would send in-app notification:", event.Message)
	case string(alerts.NotificationEmail):
		recipient := "admin@example.com"
		if event.Alert.Notifications[0].Settings != nil {
			if r, ok := event.Alert.Notifications[0].Settings["recipient"].(string); ok {
				recipient = r
			}
		}
		fmt.Printf("Would send email to %s: %s\n", recipient, event.Message)
	}
}
