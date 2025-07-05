// File: internal/sync/evaluator.go
// Brief: Unified alert evaluation logic (migrated from internal/alerts/evaluator/)
// Detailed: Contains Evaluator, metricCollector, and all related logic for evaluating alert conditions and generating events.
// Author: drama.lin@aver.com
// Date: 2024-07-03

package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"argus/internal/database"
	"argus/internal/metrics"
	"argus/internal/models"
)

const (
	DefaultEvaluationInterval = 30 * time.Second
	DefaultAlertDebounceCount = 2
	DefaultAlertResolveCount  = 2
	DefaultEventChannelSize   = 1000
)

type EvaluatorConfig struct {
	EvaluationInterval time.Duration
	AlertDebounceCount int
	AlertResolveCount  int
	EventChannelSize   int
}

func DefaultEvaluatorConfig() *EvaluatorConfig {
	return &EvaluatorConfig{
		EvaluationInterval: DefaultEvaluationInterval,
		AlertDebounceCount: DefaultAlertDebounceCount,
		AlertResolveCount:  DefaultAlertResolveCount,
		EventChannelSize:   DefaultEventChannelSize,
	}
}

// AlertStatusMap represents a thread-safe map of alert statuses using atomic operations
type AlertStatusMap struct {
	data atomic.Value // stores map[string]*models.AlertStatus
}

// NewAlertStatusMap creates a new atomic alert status map
func NewAlertStatusMap() *AlertStatusMap {
	m := &AlertStatusMap{}
	m.data.Store(make(map[string]*models.AlertStatus))
	return m
}

// Get retrieves an alert status by ID
func (m *AlertStatusMap) Get(alertID string) (*models.AlertStatus, bool) {
	statusMap := m.data.Load().(map[string]*models.AlertStatus)
	status, ok := statusMap[alertID]
	return status, ok
}

// GetAll returns a copy of all alert statuses
func (m *AlertStatusMap) GetAll() map[string]*models.AlertStatus {
	statusMap := m.data.Load().(map[string]*models.AlertStatus)
	result := make(map[string]*models.AlertStatus, len(statusMap))
	for id, status := range statusMap {
		result[id] = status
	}
	return result
}

// Update atomically updates the alert status map using read-copy-update pattern
func (m *AlertStatusMap) Update(alertID string, status *models.AlertStatus) {
	for {
		oldMap := m.data.Load().(map[string]*models.AlertStatus)
		newMap := make(map[string]*models.AlertStatus, len(oldMap)+1)

		// Copy existing entries
		for id, s := range oldMap {
			newMap[id] = s
		}

		// Update the specific entry
		newMap[alertID] = status

		// Attempt atomic swap
		if m.data.CompareAndSwap(oldMap, newMap) {
			break
		}
		// If CAS failed, retry with new snapshot
	}
}

// Delete atomically removes an alert status
func (m *AlertStatusMap) Delete(alertID string) {
	for {
		oldMap := m.data.Load().(map[string]*models.AlertStatus)
		if _, exists := oldMap[alertID]; !exists {
			return // Nothing to delete
		}

		newMap := make(map[string]*models.AlertStatus, len(oldMap)-1)

		// Copy existing entries except the one to delete
		for id, s := range oldMap {
			if id != alertID {
				newMap[id] = s
			}
		}

		// Attempt atomic swap
		if m.data.CompareAndSwap(oldMap, newMap) {
			break
		}
		// If CAS failed, retry with new snapshot
	}
}

// Initialize atomically sets the initial alert status map
func (m *AlertStatusMap) Initialize(statusMap map[string]*models.AlertStatus) {
	m.data.Store(statusMap)
}

type Evaluator struct {
	config           *EvaluatorConfig
	alertStore       *database.AlertStore
	alertStatus      *AlertStatusMap
	metricsCollector *metrics.Collector
	eventCh          chan models.AlertEvent
	wg               sync.WaitGroup

	// Object pools for reducing allocations
	eventPool sync.Pool
}

func NewEvaluator(alertStore *database.AlertStore, config *EvaluatorConfig) *Evaluator {
	if config == nil {
		config = DefaultEvaluatorConfig()
	}

	return &Evaluator{
		config:      config,
		alertStore:  alertStore,
		alertStatus: NewAlertStatusMap(),
		eventCh:     make(chan models.AlertEvent, config.EventChannelSize),
		eventPool: sync.Pool{
			New: func() interface{} {
				return &models.AlertEvent{}
			},
		},
	}
}

// SetMetricsCollector sets the centralized metrics collector
func (e *Evaluator) SetMetricsCollector(collector *metrics.Collector) {
	e.metricsCollector = collector
}

// Start begins the evaluation process
func (e *Evaluator) Start(ctx context.Context) error {
	slog.Info("Starting alert evaluator",
		"evaluation_interval", e.config.EvaluationInterval,
		"debounce_count", e.config.AlertDebounceCount,
		"resolve_count", e.config.AlertResolveCount,
		"event_channel_size", e.config.EventChannelSize)

	// Initialize alert status from stored configurations
	if err := e.initAlertStatus(); err != nil {
		return fmt.Errorf("failed to initialize alert status: %w", err)
	}

	// Start the evaluation loop
	e.wg.Add(1)
	go e.evaluationLoop(ctx)

	return nil
}

func (e *Evaluator) Stop() {
	slog.Info("Stopping alert evaluator")
	e.wg.Wait()
	close(e.eventCh)
}

func (e *Evaluator) Events() <-chan models.AlertEvent {
	return e.eventCh
}

func (e *Evaluator) GetAlertStatus(alertID string) (*models.AlertStatus, bool) {
	return e.alertStatus.Get(alertID)
}

func (e *Evaluator) GetAllAlertStatus() map[string]*models.AlertStatus {
	return e.alertStatus.GetAll()
}

func (e *Evaluator) initAlertStatus() error {
	alertConfigs, err := e.alertStore.ListAlerts()
	if err != nil {
		return err
	}

	statusMap := make(map[string]*models.AlertStatus)
	for _, config := range alertConfigs {
		if config.Enabled {
			statusMap[config.ID] = &models.AlertStatus{
				AlertID: config.ID,
				State:   models.StateInactive,
				Message: fmt.Sprintf("Alert %s initialized", config.Name),
			}
		}
	}

	e.alertStatus.Initialize(statusMap)
	slog.Info("Initialized alert status", "alert_count", len(statusMap))
	return nil
}

func (e *Evaluator) evaluationLoop(ctx context.Context) {
	defer e.wg.Done()
	ticker := time.NewTicker(e.config.EvaluationInterval)
	defer ticker.Stop()

	// Persistent counters to avoid allocations
	pendingCounters := make(map[string]int)
	resolveCounters := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Evaluation loop stopped due to context cancellation")
			return
		case <-ticker.C:
			e.evaluateAlerts(pendingCounters, resolveCounters)
		}
	}
}

func (e *Evaluator) evaluateAlerts(pendingCounters, resolveCounters map[string]int) {
	alertConfigs, err := e.alertStore.ListAlerts()
	if err != nil {
		slog.Error("Failed to list alerts", "error", err)
		return
	}

	for _, config := range alertConfigs {
		if !config.Enabled {
			continue
		}

		currentValue, err := e.evaluateMetric(config.Threshold)
		if err != nil {
			slog.Error("Failed to evaluate metric",
				"alert_id", config.ID,
				"alert_name", config.Name,
				"error", err)
			continue
		}

		exceeded := e.compareValue(currentValue, config.Threshold.Value, config.Threshold.Operator)
		e.processAlertState(config, currentValue, exceeded, pendingCounters, resolveCounters)
	}
}

func (e *Evaluator) processAlertState(config *models.AlertConfig, currentValue float64, exceeded bool, pendingCounters, resolveCounters map[string]int) {
	// Get current status or create new one
	status, exists := e.alertStatus.Get(config.ID)
	if !exists {
		status = &models.AlertStatus{
			AlertID: config.ID,
			State:   models.StateInactive,
		}
	}

	// Create a copy for modification to avoid race conditions
	newStatus := *status
	newStatus.CurrentValue = currentValue

	switch status.State {
	case models.StateInactive:
		if exceeded {
			pendingCounters[config.ID]++
			if pendingCounters[config.ID] >= e.config.AlertDebounceCount {
				oldState := newStatus.State
				newStatus.State = models.StatePending
				delete(pendingCounters, config.ID)
				e.alertStatus.Update(config.ID, &newStatus)
				e.generateEvent(oldState, newStatus.State, currentValue, config, &newStatus)
			}
		} else {
			// Reset pending counter if condition is no longer met
			delete(pendingCounters, config.ID)
		}

	case models.StatePending:
		if !exceeded {
			resolveCounters[config.ID]++
			if resolveCounters[config.ID] >= e.config.AlertResolveCount {
				oldState := newStatus.State
				newStatus.State = models.StateResolved
				delete(resolveCounters, config.ID)
				e.alertStatus.Update(config.ID, &newStatus)
				e.generateEvent(oldState, newStatus.State, currentValue, config, &newStatus)
			}
		} else {
			// Reset resolve counter if condition is still met
			delete(resolveCounters, config.ID)
		}

	case models.StateResolved:
		if exceeded {
			pendingCounters[config.ID]++
			if pendingCounters[config.ID] >= e.config.AlertDebounceCount {
				oldState := newStatus.State
				newStatus.State = models.StatePending
				delete(pendingCounters, config.ID)
				e.alertStatus.Update(config.ID, &newStatus)
				e.generateEvent(oldState, newStatus.State, currentValue, config, &newStatus)
			}
		} else {
			// Reset pending counter if condition is no longer met
			delete(pendingCounters, config.ID)
		}
	}

	// Update current value even if state didn't change
	if exists {
		e.alertStatus.Update(config.ID, &newStatus)
	}
}

// generateEvent creates and sends an alert event using object pooling
func (e *Evaluator) generateEvent(oldState, newState models.AlertState, currentValue float64, config *models.AlertConfig, status *models.AlertStatus) {
	// Get event from pool
	event := e.eventPool.Get().(*models.AlertEvent)

	// Reset and populate event
	*event = models.AlertEvent{
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
	case e.eventCh <- *event:
		// Event sent successfully
		slog.Debug("Alert event generated",
			"alert_id", config.ID,
			"alert_name", config.Name,
			"old_state", oldState,
			"new_state", newState,
			"current_value", currentValue)
	default:
		slog.Warn("Event channel full, dropping alert event",
			"alert_id", config.ID,
			"alert_name", config.Name,
			"old_state", oldState,
			"new_state", newState)
	}

	// Return event to pool
	e.eventPool.Put(event)
}

func (e *Evaluator) evaluateMetric(threshold models.ThresholdConfig) (float64, error) {
	// Use centralized metrics collector if available
	if e.metricsCollector != nil {
		return e.evaluateMetricFromCollector(threshold)
	}

	// Fallback to direct evaluation (for backward compatibility)
	return e.evaluateMetricDirect(threshold)
}

func (e *Evaluator) evaluateMetricFromCollector(threshold models.ThresholdConfig) (float64, error) {
	switch threshold.MetricType {
	case models.MetricCPU:
		cpuMetrics := e.metricsCollector.GetCPUMetrics()
		if cpuMetrics == nil {
			return 0, fmt.Errorf("CPU metrics not available from collector")
		}
		return e.extractCPUValue(cpuMetrics, threshold.MetricName)

	case models.MetricMemory:
		memoryMetrics := e.metricsCollector.GetMemoryMetrics()
		if memoryMetrics == nil {
			return 0, fmt.Errorf("memory metrics not available from collector")
		}
		return e.extractMemoryValue(memoryMetrics, threshold.MetricName)

	case models.MetricLoad:
		cpuMetrics := e.metricsCollector.GetCPUMetrics()
		if cpuMetrics == nil {
			return 0, fmt.Errorf("load metrics not available from collector")
		}
		return e.extractLoadValue(cpuMetrics, threshold.MetricName)

	case models.MetricNetwork:
		networkMetrics := e.metricsCollector.GetNetworkMetrics()
		if networkMetrics == nil {
			return 0, fmt.Errorf("network metrics not available from collector")
		}
		return e.extractNetworkValue(networkMetrics, threshold.MetricName)

	default:
		return 0, fmt.Errorf("unsupported metric type: %s", threshold.MetricType)
	}
}

func (e *Evaluator) extractCPUValue(cpuMetrics *metrics.CPUMetrics, metricName string) (float64, error) {
	switch metricName {
	case "usage_percent":
		return cpuMetrics.UsagePercent, nil
	case "load1":
		return cpuMetrics.Load1, nil
	case "load5":
		return cpuMetrics.Load5, nil
	case "load15":
		return cpuMetrics.Load15, nil
	default:
		return 0, fmt.Errorf("unsupported CPU metric: %s", metricName)
	}
}

func (e *Evaluator) extractMemoryValue(memoryMetrics *metrics.MemoryMetrics, metricName string) (float64, error) {
	switch metricName {
	case "used_percent":
		return memoryMetrics.UsedPercent, nil
	case "used":
		return float64(memoryMetrics.Used), nil
	case "free":
		return float64(memoryMetrics.Free), nil
	default:
		return 0, fmt.Errorf("unsupported memory metric: %s", metricName)
	}
}

func (e *Evaluator) extractLoadValue(cpuMetrics *metrics.CPUMetrics, metricName string) (float64, error) {
	switch metricName {
	case "load1":
		return cpuMetrics.Load1, nil
	case "load5":
		return cpuMetrics.Load5, nil
	case "load15":
		return cpuMetrics.Load15, nil
	default:
		return 0, fmt.Errorf("unsupported load metric: %s", metricName)
	}
}

func (e *Evaluator) extractNetworkValue(networkMetrics *metrics.NetworkMetrics, metricName string) (float64, error) {
	switch metricName {
	case "bytes_sent":
		return float64(networkMetrics.BytesSent), nil
	case "bytes_recv":
		return float64(networkMetrics.BytesRecv), nil
	case "packets_sent":
		return float64(networkMetrics.PacketsSent), nil
	case "packets_recv":
		return float64(networkMetrics.PacketsRecv), nil
	default:
		return 0, fmt.Errorf("unsupported network metric: %s", metricName)
	}
}

// Fallback direct metric evaluation (kept for backward compatibility)
func (e *Evaluator) evaluateMetricDirect(threshold models.ThresholdConfig) (float64, error) {
	// This would contain the original direct gopsutil calls
	// Omitted for brevity as it's now deprecated in favor of centralized collector
	return 0, fmt.Errorf("direct metric evaluation not supported, use centralized metrics collector")
}

func (e *Evaluator) compareValue(current, threshold float64, operator models.ComparisonOperator) bool {
	switch operator {
	case models.OperatorGreaterThan:
		return current > threshold
	case models.OperatorGreaterThanOrEqual:
		return current >= threshold
	case models.OperatorLessThan:
		return current < threshold
	case models.OperatorLessThanOrEqual:
		return current <= threshold
	case models.OperatorEqual:
		return current == threshold
	case models.OperatorNotEqual:
		return current != threshold
	default:
		slog.Warn("Unknown comparison operator", "operator", operator)
		return false
	}
}
