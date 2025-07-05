package benchmarks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"argus/internal/database"
	"argus/internal/models"
	"argus/internal/services"
)

// BenchmarkAlertEvaluator benchmarks the alert evaluation process
func BenchmarkAlertEvaluator(b *testing.B) {
	// Create a temporary alert store
	alertStore, err := database.NewAlertStore("./.test_alerts")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		// Clean up test directory
		// os.RemoveAll("./.test_alerts")
	}()

	// Create test alert configuration
	alertConfig := &models.AlertConfig{
		ID:          "test-alert",
		Name:        "Test Alert",
		Description: "Test alert for benchmarking",
		Enabled:     true,
		Severity:    models.SeverityWarning,
		Threshold: models.ThresholdConfig{
			MetricType: models.MetricCPU,
			MetricName: "usage_percent",
			Value:      80.0,
			Operator:   models.OperatorGreaterThan,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the alert configuration
	if err := alertStore.CreateAlert(alertConfig); err != nil {
		b.Fatal(err)
	}

	// Create evaluator
	evalConfig := services.DefaultEvaluatorConfig()
	evaluator := services.NewEvaluator(alertStore, evalConfig)

	// Initialize evaluator
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := evaluator.Start(ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate evaluation cycle
		status, exists := evaluator.GetAlertStatus("test-alert")
		if !exists {
			b.Fatal("Alert status not found")
		}
		_ = status
	}
}

// BenchmarkAlertStatusAccess benchmarks concurrent access to alert status
func BenchmarkAlertStatusAccess(b *testing.B) {
	alertStore, err := database.NewAlertStore("./.test_alerts_concurrent")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		// Clean up test directory
		// os.RemoveAll("./.test_alerts_concurrent")
	}()

	// Create multiple test alert configurations
	for i := 0; i < 10; i++ {
		alertConfig := &models.AlertConfig{
			ID:          fmt.Sprintf("test-alert-%d", i),
			Name:        fmt.Sprintf("Test Alert %d", i),
			Description: "Test alert for concurrent benchmarking",
			Enabled:     true,
			Severity:    models.SeverityWarning,
			Threshold: models.ThresholdConfig{
				MetricType: models.MetricCPU,
				MetricName: "usage_percent",
				Value:      80.0,
				Operator:   models.OperatorGreaterThan,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := alertStore.CreateAlert(alertConfig); err != nil {
			b.Fatal(err)
		}
	}

	evalConfig := services.DefaultEvaluatorConfig()
	evaluator := services.NewEvaluator(alertStore, evalConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := evaluator.Start(ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate concurrent access to different alerts
			for i := 0; i < 10; i++ {
				alertID := fmt.Sprintf("test-alert-%d", i%10)
				status, _ := evaluator.GetAlertStatus(alertID)
				_ = status
			}
		}
	})
}

// BenchmarkAlertEventProcessing benchmarks alert event processing
func BenchmarkAlertEventProcessing(b *testing.B) {
	// Create notifier
	notifierConfig := services.DefaultConfig()
	notifier := services.NewNotifier(notifierConfig)

	// Create in-app channel
	inAppChannel := services.NewInAppChannel(1000)
	notifier.RegisterChannel(inAppChannel)

	// Create test alert event
	alertEvent := models.AlertEvent{
		AlertID:      "test-alert",
		NewState:     models.StateActive,
		OldState:     models.StateInactive,
		CurrentValue: 85.0,
		Threshold:    80.0,
		Timestamp:    time.Now(),
		Message:      "Test alert triggered",
		Alert: &models.AlertConfig{
			ID:          "test-alert",
			Name:        "Test Alert",
			Description: "Test alert for benchmarking",
			Severity:    models.SeverityWarning,
			Threshold: models.ThresholdConfig{
				MetricType: models.MetricCPU,
				MetricName: "usage_percent",
				Value:      80.0,
				Operator:   models.OperatorGreaterThan,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notifier.ProcessEvent(alertEvent)
	}
}
