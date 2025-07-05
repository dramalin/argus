package benchmarks

import (
	"testing"
	"time"

	"argus/internal/models"
	"argus/internal/services"
)

// BenchmarkNotificationProcessing benchmarks overall notification processing performance
func BenchmarkNotificationProcessing(b *testing.B) {
	// Create notifier with pre-compiled templates
	notifier := services.NewNotifier(nil)

	// Register mock channel
	mockChannel := &benchmarkMockChannel{}
	notifier.RegisterChannel(mockChannel)

	// Create test alert event
	event := createBenchmarkAlertEvent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notifier.ProcessEvent(event)
	}
}

// BenchmarkRateLimiting benchmarks the new efficient rate limiting
func BenchmarkRateLimiting(b *testing.B) {
	config := &services.NotifierConfig{
		RateLimit:       1000,
		RateLimitWindow: 1 * time.Hour,
	}
	notifier := services.NewNotifier(config)

	// Create mock channel for testing
	mockChannel := &benchmarkMockChannel{}
	notifier.RegisterChannel(mockChannel)

	event := createBenchmarkAlertEvent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notifier.ProcessEvent(event)
	}
}

// BenchmarkEmailQueueing benchmarks non-blocking email sending
func BenchmarkEmailQueueing(b *testing.B) {
	emailConfig := &services.EmailConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		UseSSL:   false,
	}

	notifierConfig := &services.NotifierConfig{
		EmailWorkerCount: 3,
		EmailQueueSize:   1000,
	}

	emailChannel := services.NewEmailChannel(emailConfig, notifierConfig)
	defer emailChannel.Stop()

	event := createBenchmarkAlertEvent()
	// Add email notification config
	event.Alert.Notifications = append(event.Alert.Notifications, models.NotificationConfig{
		Type:    models.NotificationEmail,
		Enabled: true,
		Settings: map[string]any{
			"recipient": "test@example.com",
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := emailChannel.Send(event, "Test Subject", "Test Body")
		if err != nil && err.Error() != "email queue is full" {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentNotifications benchmarks concurrent notification processing
func BenchmarkConcurrentNotifications(b *testing.B) {
	notifier := services.NewNotifier(nil)

	// Register mock channel
	mockChannel := &benchmarkMockChannel{}
	notifier.RegisterChannel(mockChannel)

	event := createBenchmarkAlertEvent()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			notifier.ProcessEvent(event)
		}
	})
}

// Helper functions

func createBenchmarkAlertEvent() models.AlertEvent {
	alert := &models.AlertConfig{
		ID:          "benchmark-alert",
		Name:        "Benchmark Alert",
		Description: "Benchmark alert description",
		Enabled:     true,
		Severity:    models.SeverityCritical,
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

	status := &models.AlertStatus{
		AlertID:      alert.ID,
		State:        models.StateActive,
		CurrentValue: 95.0,
	}

	return models.AlertEvent{
		AlertID:      alert.ID,
		OldState:     models.StateInactive,
		NewState:     models.StateActive,
		CurrentValue: 95.0,
		Threshold:    90.0,
		Timestamp:    time.Now(),
		Message:      "CPU usage exceeded threshold",
		Alert:        alert,
		Status:       status,
	}
}

// Mock channel for benchmarking
type benchmarkMockChannel struct {
	sendCount int
}

func (m *benchmarkMockChannel) Send(event models.AlertEvent, subject, body string) error {
	m.sendCount++
	return nil
}

func (m *benchmarkMockChannel) Type() models.NotificationType {
	return models.NotificationInApp
}

func (m *benchmarkMockChannel) Name() string {
	return "Benchmark Mock Channel"
}
