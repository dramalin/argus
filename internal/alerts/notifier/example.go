package notifier

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"argus/internal/alerts"
	"argus/internal/alerts/evaluator"
	"argus/internal/storage"
)

// ExampleUsage demonstrates how to use the notifier in a main application
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

	// Create evaluator
	evaluator := evaluator.NewEvaluator(alertStore, nil)

	// Create notifier
	notifier := NewNotifier(nil)

	// Register notification channels
	inAppChannel := NewInAppChannel(100)
	notifier.RegisterChannel(inAppChannel)

	emailConfig := &EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "alerts@example.com",
		Password: "password",
		From:     "alerts@example.com",
	}
	emailChannel := NewEmailChannel(emailConfig)
	notifier.RegisterChannel(emailChannel)

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
			// Process the event with the notifier
			notifier.ProcessEvent(event)

			// For demonstration, show in-app notifications
			if event.NewState == alerts.StateActive {
				slog.Info("In-app notifications", "count", len(inAppChannel.GetNotifications()))
				for _, notification := range inAppChannel.GetNotifications() {
					slog.Info("Notification",
						"id", notification.ID,
						"subject", notification.Subject,
						"read", notification.Read)
				}
			}
		}
	}()

	slog.Info("Alert notifier started successfully")

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Cancel context and stop evaluator
	cancel()
	evaluator.Stop()

	slog.Info("Alert notifier stopped successfully")
}

// ExampleCreateAlert shows how to create an alert with notifications
func ExampleCreateAlert() {
	// Create a CPU usage alert with email notification
	alert := &alerts.AlertConfig{
		ID:          "cpu-usage-alert",
		Name:        "High CPU Usage",
		Description: "Alert when CPU usage exceeds 90% for 5 minutes",
		Enabled:     true,
		Severity:    alerts.SeverityCritical,
		Threshold: alerts.ThresholdConfig{
			MetricType: alerts.MetricCPU,
			MetricName: "usage_percent",
			Operator:   alerts.OperatorGreaterThan,
			Value:      90.0,
			Duration:   5 * time.Minute,
		},
		Notifications: []alerts.NotificationConfig{
			{
				Type:    alerts.NotificationInApp,
				Enabled: true,
			},
			{
				Type:    alerts.NotificationEmail,
				Enabled: true,
				Settings: map[string]any{
					"recipient": "admin@example.com",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validate the alert configuration
	if err := alert.Validate(); err != nil {
		slog.Error("Invalid alert configuration", "error", err)
		return
	}

	// In a real application, you would save this alert to storage
	alertStore, _ := storage.NewAlertStore(".argus/config")
	if err := alertStore.CreateAlert(alert); err != nil {
		slog.Error("Failed to save alert", "error", err)
		return
	}

	slog.Info("Alert created successfully", "id", alert.ID)
}
