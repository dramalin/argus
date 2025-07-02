// Package evaluator provides functionality for evaluating alert conditions against system metrics
package evaluator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	"argus/internal/alerts"
	"argus/internal/storage"
)

// Default configuration values
const (
	// DefaultEvaluationInterval is the default interval between metric evaluations
	DefaultEvaluationInterval = 30 * time.Second

	// DefaultAlertDebounceCount is the default number of consecutive evaluations before triggering an alert
	DefaultAlertDebounceCount = 2

	// DefaultAlertResolveCount is the default number of consecutive evaluations before resolving an alert
	DefaultAlertResolveCount = 2
)

// AlertEvent represents an alert state change event
type AlertEvent struct {
	AlertID      string              // ID of the alert that changed state
	OldState     alerts.AlertState   // Previous state
	NewState     alerts.AlertState   // New state
	CurrentValue float64             // Current metric value
	Threshold    float64             // Alert threshold value
	Timestamp    time.Time           // When the state change occurred
	Message      string              // Human-readable message
	Alert        *alerts.AlertConfig // The full alert configuration
	Status       *alerts.AlertStatus // The current alert status
}

// Config holds the configuration for the evaluator
type Config struct {
	EvaluationInterval time.Duration // Interval between evaluations
	AlertDebounceCount int           // Number of consecutive evaluations before triggering an alert
	AlertResolveCount  int           // Number of consecutive evaluations before resolving an alert
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		EvaluationInterval: DefaultEvaluationInterval,
		AlertDebounceCount: DefaultAlertDebounceCount,
		AlertResolveCount:  DefaultAlertResolveCount,
	}
}

// Evaluator is responsible for evaluating alert conditions
type Evaluator struct {
	config      *Config
	alertStore  *storage.AlertStore
	alertStatus map[string]*alerts.AlertStatus
	statusMu    sync.RWMutex
	eventCh     chan AlertEvent
	wg          sync.WaitGroup
	metrics     *metricCollector
}

// metricCollector collects system metrics
type metricCollector struct {
	cpuSampleInterval time.Duration
}

// NewEvaluator creates a new alert evaluator
func NewEvaluator(alertStore *storage.AlertStore, config *Config) *Evaluator {
	if config == nil {
		config = DefaultConfig()
	}

	return &Evaluator{
		config:      config,
		alertStore:  alertStore,
		alertStatus: make(map[string]*alerts.AlertStatus),
		eventCh:     make(chan AlertEvent, 100), // Buffer for 100 events
		metrics: &metricCollector{
			cpuSampleInterval: 1 * time.Second, // 1 second sample for CPU
		},
	}
}

// Start begins the evaluation process
func (e *Evaluator) Start(ctx context.Context) error {
	slog.Info("Starting alert evaluator",
		"evaluation_interval", e.config.EvaluationInterval,
		"debounce_count", e.config.AlertDebounceCount,
		"resolve_count", e.config.AlertResolveCount)

	// Initialize alert status from stored configurations
	if err := e.initAlertStatus(); err != nil {
		return fmt.Errorf("failed to initialize alert status: %w", err)
	}

	// Start the evaluation loop
	e.wg.Add(1)
	go e.evaluationLoop(ctx)

	return nil
}

// Stop gracefully stops the evaluator
func (e *Evaluator) Stop() {
	slog.Info("Stopping alert evaluator")
	e.wg.Wait()
	close(e.eventCh)
}

// Events returns the channel for alert events
func (e *Evaluator) Events() <-chan AlertEvent {
	return e.eventCh
}

// GetAlertStatus returns the current status of an alert
func (e *Evaluator) GetAlertStatus(alertID string) (*alerts.AlertStatus, bool) {
	e.statusMu.RLock()
	defer e.statusMu.RUnlock()

	status, ok := e.alertStatus[alertID]
	return status, ok
}

// GetAllAlertStatus returns the current status of all alerts
func (e *Evaluator) GetAllAlertStatus() map[string]*alerts.AlertStatus {
	e.statusMu.RLock()
	defer e.statusMu.RUnlock()

	// Create a copy to avoid concurrent map access
	statusCopy := make(map[string]*alerts.AlertStatus, len(e.alertStatus))
	for id, status := range e.alertStatus {
		statusCopy[id] = status
	}

	return statusCopy
}

// initAlertStatus initializes the alert status map from stored configurations
func (e *Evaluator) initAlertStatus() error {
	alertConfigs, err := e.alertStore.ListAlerts()
	if err != nil {
		return err
	}

	e.statusMu.Lock()
	defer e.statusMu.Unlock()

	for _, config := range alertConfigs {
		if config.Enabled {
			e.alertStatus[config.ID] = &alerts.AlertStatus{
				AlertID: config.ID,
				State:   alerts.StateInactive,
				Message: fmt.Sprintf("Alert %s initialized", config.Name),
			}
		}
	}

	slog.Info("Initialized alert status", "alert_count", len(e.alertStatus))
	return nil
}

// evaluationLoop runs the continuous evaluation process
func (e *Evaluator) evaluationLoop(ctx context.Context) {
	defer e.wg.Done()

	ticker := time.NewTicker(e.config.EvaluationInterval)
	defer ticker.Stop()

	// Counters for debouncing
	pendingCounters := make(map[string]int)
	resolveCounters := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Evaluation loop stopped due to context cancellation")
			return
		case <-ticker.C:
			alertConfigs, err := e.alertStore.ListAlerts()
			if err != nil {
				slog.Error("Failed to list alerts", "error", err)
				continue
			}

			for _, config := range alertConfigs {
				if !config.Enabled {
					continue
				}

				// Evaluate the alert condition
				currentValue, err := e.evaluateMetric(config.Threshold)
				if err != nil {
					slog.Error("Failed to evaluate metric",
						"alert_id", config.ID,
						"alert_name", config.Name,
						"error", err)
					continue
				}

				// Check if the threshold is exceeded
				exceeded := e.compareValue(currentValue, config.Threshold.Value, config.Threshold.Operator)

				e.statusMu.Lock()
				status, exists := e.alertStatus[config.ID]
				if !exists {
					// Initialize status if it doesn't exist
					status = &alerts.AlertStatus{
						AlertID: config.ID,
						State:   alerts.StateInactive,
					}
					e.alertStatus[config.ID] = status
				}

				// Update current value
				status.CurrentValue = currentValue

				// Handle state transitions with debouncing
				switch status.State {
				case alerts.StateInactive:
					if exceeded {
						// Potential transition to pending
						pendingCounters[config.ID]++
						if pendingCounters[config.ID] >= 1 {
							// Transition to pending on first detection
							oldState := status.State
							status.State = alerts.StatePending
							delete(pendingCounters, config.ID)

							slog.Info("Alert state changed to pending",
								"alert_id", config.ID,
								"alert_name", config.Name,
								"value", currentValue,
								"threshold", config.Threshold.Value)

							// Generate event for pending state
							e.generateEvent(oldState, status.State, currentValue, config, status)
						}
					} else {
						// Reset counter if condition is no longer met
						delete(pendingCounters, config.ID)
					}

				case alerts.StatePending:
					if !exceeded {
						// Condition no longer met, go back to inactive
						oldState := status.State
						status.State = alerts.StateInactive
						status.Message = fmt.Sprintf("Alert condition no longer met: %v %s %v",
							currentValue, config.Threshold.Operator, config.Threshold.Value)

						slog.Info("Alert returned to inactive state",
							"alert_id", config.ID,
							"alert_name", config.Name,
							"value", currentValue,
							"threshold", config.Threshold.Value)

						// Generate event for return to inactive
						e.generateEvent(oldState, status.State, currentValue, config, status)
					} else {
						// Check if duration requirement is met
						if config.Threshold.Duration > 0 {
							// If no triggered time is set, set it now
							if status.TriggeredAt == nil {
								now := time.Now()
								status.TriggeredAt = &now
							}

							// Check if duration has elapsed
							durationElapsed := time.Since(*status.TriggeredAt) >= config.Threshold.Duration
							if durationElapsed {
								oldState := status.State
								status.State = alerts.StateActive
								status.Message = fmt.Sprintf("Alert triggered: %v %s %v for %v",
									currentValue, config.Threshold.Operator, config.Threshold.Value, config.Threshold.Duration)

								slog.Warn("Alert activated",
									"alert_id", config.ID,
									"alert_name", config.Name,
									"severity", config.Severity,
									"value", currentValue,
									"threshold", config.Threshold.Value)

								// Generate event for active state
								e.generateEvent(oldState, status.State, currentValue, config, status)
							}
						} else if config.Threshold.SustainedFor > 0 {
							// Use sustained_for count instead of duration
							pendingCounters[config.ID]++
							if pendingCounters[config.ID] >= config.Threshold.SustainedFor {
								oldState := status.State
								now := time.Now()
								status.State = alerts.StateActive
								status.TriggeredAt = &now
								status.Message = fmt.Sprintf("Alert triggered: %v %s %v for %d checks",
									currentValue, config.Threshold.Operator, config.Threshold.Value, config.Threshold.SustainedFor)
								delete(pendingCounters, config.ID)

								slog.Warn("Alert activated",
									"alert_id", config.ID,
									"alert_name", config.Name,
									"severity", config.Severity,
									"value", currentValue,
									"threshold", config.Threshold.Value)

								// Generate event for active state
								e.generateEvent(oldState, status.State, currentValue, config, status)
							}
						} else {
							// No duration or sustained_for specified, activate immediately
							oldState := status.State
							now := time.Now()
							status.State = alerts.StateActive
							status.TriggeredAt = &now
							status.Message = fmt.Sprintf("Alert triggered: %v %s %v",
								currentValue, config.Threshold.Operator, config.Threshold.Value)

							slog.Warn("Alert activated",
								"alert_id", config.ID,
								"alert_name", config.Name,
								"severity", config.Severity,
								"value", currentValue,
								"threshold", config.Threshold.Value)

							// Generate event for active state
							e.generateEvent(oldState, status.State, currentValue, config, status)
						}
					}

				case alerts.StateActive:
					if !exceeded {
						// Potential resolution
						resolveCounters[config.ID]++
						if resolveCounters[config.ID] >= e.config.AlertResolveCount {
							oldState := status.State
							now := time.Now()
							status.State = alerts.StateInactive
							status.ResolvedAt = &now
							status.TriggeredAt = nil
							status.Message = fmt.Sprintf("Alert resolved: %v %s %v",
								currentValue, config.Threshold.Operator, config.Threshold.Value)
							delete(resolveCounters, config.ID)

							slog.Info("Alert resolved",
								"alert_id", config.ID,
								"alert_name", config.Name,
								"value", currentValue,
								"threshold", config.Threshold.Value)

							// Generate event for resolution
							e.generateEvent(oldState, status.State, currentValue, config, status)
						}
					} else {
						// Condition still met, reset resolve counter
						delete(resolveCounters, config.ID)
					}
				}

				e.statusMu.Unlock()
			}
		}
	}
}

// generateEvent creates and sends an alert event
func (e *Evaluator) generateEvent(oldState, newState alerts.AlertState, currentValue float64, config *alerts.AlertConfig, status *alerts.AlertStatus) {
	event := AlertEvent{
		AlertID:      config.ID,
		OldState:     oldState,
		NewState:     newState,
		CurrentValue: currentValue,
		Threshold:    config.Threshold.Value,
		Timestamp:    time.Now(),
		Message:      status.Message,
		Alert:        config,
		Status:       status,
	}

	// Send event non-blocking to avoid deadlocks if channel is full
	select {
	case e.eventCh <- event:
		// Event sent successfully
	default:
		slog.Warn("Event channel full, dropping alert event",
			"alert_id", config.ID,
			"alert_name", config.Name,
			"old_state", oldState,
			"new_state", newState)
	}
}

// evaluateMetric gets the current value for the specified metric
func (e *Evaluator) evaluateMetric(threshold alerts.ThresholdConfig) (float64, error) {
	switch threshold.MetricType {
	case alerts.MetricCPU:
		return e.evaluateCPUMetric(threshold.MetricName)
	case alerts.MetricMemory:
		return e.evaluateMemoryMetric(threshold.MetricName)
	case alerts.MetricLoad:
		return e.evaluateLoadMetric(threshold.MetricName)
	case alerts.MetricNetwork:
		return e.evaluateNetworkMetric(threshold.MetricName)
	default:
		return 0, fmt.Errorf("unsupported metric type: %s", threshold.MetricType)
	}
}

// evaluateCPUMetric gets the current CPU metric value
func (e *Evaluator) evaluateCPUMetric(metricName string) (float64, error) {
	switch metricName {
	case "usage_percent":
		cpuPercent, err := cpu.Percent(e.metrics.cpuSampleInterval, false)
		if err != nil {
			return 0, err
		}
		if len(cpuPercent) == 0 {
			return 0, fmt.Errorf("no CPU usage data available")
		}
		return cpuPercent[0], nil
	case "load1", "load5", "load15":
		loadAvg, err := load.Avg()
		if err != nil {
			return 0, err
		}
		switch metricName {
		case "load1":
			return loadAvg.Load1, nil
		case "load5":
			return loadAvg.Load5, nil
		case "load15":
			return loadAvg.Load15, nil
		}
	}
	return 0, fmt.Errorf("unsupported CPU metric: %s", metricName)
}

// evaluateMemoryMetric gets the current memory metric value
func (e *Evaluator) evaluateMemoryMetric(metricName string) (float64, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	switch metricName {
	case "used_percent":
		return vm.UsedPercent, nil
	case "used":
		return float64(vm.Used), nil
	case "free":
		return float64(vm.Free), nil
	default:
		return 0, fmt.Errorf("unsupported memory metric: %s", metricName)
	}
}

// evaluateLoadMetric gets the current load metric value
func (e *Evaluator) evaluateLoadMetric(metricName string) (float64, error) {
	loadAvg, err := load.Avg()
	if err != nil {
		return 0, err
	}

	switch metricName {
	case "load1":
		return loadAvg.Load1, nil
	case "load5":
		return loadAvg.Load5, nil
	case "load15":
		return loadAvg.Load15, nil
	default:
		return 0, fmt.Errorf("unsupported load metric: %s", metricName)
	}
}

// evaluateNetworkMetric gets the current network metric value
func (e *Evaluator) evaluateNetworkMetric(metricName string) (float64, error) {
	ioCounters, err := net.IOCounters(false)
	if err != nil {
		return 0, err
	}

	if len(ioCounters) == 0 {
		return 0, fmt.Errorf("no network interfaces found")
	}

	io := ioCounters[0]

	switch metricName {
	case "bytes_sent":
		return float64(io.BytesSent), nil
	case "bytes_recv":
		return float64(io.BytesRecv), nil
	case "packets_sent":
		return float64(io.PacketsSent), nil
	case "packets_recv":
		return float64(io.PacketsRecv), nil
	default:
		return 0, fmt.Errorf("unsupported network metric: %s", metricName)
	}
}

// compareValue compares the current value against the threshold using the specified operator
func (e *Evaluator) compareValue(current, threshold float64, operator alerts.ComparisonOperator) bool {
	switch operator {
	case alerts.OperatorGreaterThan:
		return current > threshold
	case alerts.OperatorGreaterThanOrEqual:
		return current >= threshold
	case alerts.OperatorLessThan:
		return current < threshold
	case alerts.OperatorLessThanOrEqual:
		return current <= threshold
	case alerts.OperatorEqual:
		return current == threshold
	case alerts.OperatorNotEqual:
		return current != threshold
	default:
		slog.Error("Unknown comparison operator", "operator", operator)
		return false
	}
}
