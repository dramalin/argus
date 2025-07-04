// File: internal/sync/evaluator.go
// Brief: Unified alert evaluation logic (migrated from internal/alerts/evaluator/)
// Detailed: Contains Evaluator, metricCollector, and all related logic for evaluating alert conditions and generating events.
// Author: Argus Migration (AI)
// Date: 2024-07-03

package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"argus/internal/database"
	"argus/internal/models"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

const (
	DefaultEvaluationInterval = 30 * time.Second
	DefaultAlertDebounceCount = 2
	DefaultAlertResolveCount  = 2
)

type EvaluatorConfig struct {
	EvaluationInterval time.Duration
	AlertDebounceCount int
	AlertResolveCount  int
}

func DefaultEvaluatorConfig() *EvaluatorConfig {
	return &EvaluatorConfig{
		EvaluationInterval: DefaultEvaluationInterval,
		AlertDebounceCount: DefaultAlertDebounceCount,
		AlertResolveCount:  DefaultAlertResolveCount,
	}
}

type Evaluator struct {
	config      *EvaluatorConfig
	alertStore  *database.AlertStore
	alertStatus map[string]*models.AlertStatus
	statusMu    sync.RWMutex
	eventCh     chan models.AlertEvent
	wg          sync.WaitGroup
	metrics     *metricCollector
}

type metricCollector struct {
	cpuSampleInterval time.Duration
}

func NewEvaluator(alertStore *database.AlertStore, config *EvaluatorConfig) *Evaluator {
	if config == nil {
		config = DefaultEvaluatorConfig()
	}
	return &Evaluator{
		config:      config,
		alertStore:  alertStore,
		alertStatus: make(map[string]*models.AlertStatus),
		eventCh:     make(chan models.AlertEvent, 100),
		metrics: &metricCollector{
			cpuSampleInterval: 1 * time.Second,
		},
	}
}

// Start begins the evaluation process
func (e *Evaluator) Start(ctx context.Context) error {
	log.Printf("Starting alert evaluator, evaluation_interval: %v, debounce_count: %d, resolve_count: %d",
		e.config.EvaluationInterval, e.config.AlertDebounceCount, e.config.AlertResolveCount)

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
	log.Println("Stopping alert evaluator")
	e.wg.Wait()
	close(e.eventCh)
}

func (e *Evaluator) Events() <-chan models.AlertEvent {
	return e.eventCh
}

func (e *Evaluator) GetAlertStatus(alertID string) (*models.AlertStatus, bool) {
	e.statusMu.RLock()
	defer e.statusMu.RUnlock()
	status, ok := e.alertStatus[alertID]
	return status, ok
}

func (e *Evaluator) GetAllAlertStatus() map[string]*models.AlertStatus {
	e.statusMu.RLock()
	defer e.statusMu.RUnlock()
	statusCopy := make(map[string]*models.AlertStatus, len(e.alertStatus))
	for id, status := range e.alertStatus {
		statusCopy[id] = status
	}
	return statusCopy
}

func (e *Evaluator) initAlertStatus() error {
	alertConfigs, err := e.alertStore.ListAlerts()
	if err != nil {
		return err
	}
	e.statusMu.Lock()
	defer e.statusMu.Unlock()
	for _, config := range alertConfigs {
		if config.Enabled {
			e.alertStatus[config.ID] = &models.AlertStatus{
				AlertID: config.ID,
				State:   models.StateInactive,
				Message: fmt.Sprintf("Alert %s initialized", config.Name),
			}
		}
	}
	log.Printf("Initialized alert status, alert_count: %d", len(e.alertStatus))
	return nil
}

func (e *Evaluator) evaluationLoop(ctx context.Context) {
	defer e.wg.Done()
	ticker := time.NewTicker(e.config.EvaluationInterval)
	defer ticker.Stop()
	pendingCounters := make(map[string]int)
	resolveCounters := make(map[string]int)
	for {
		select {
		case <-ctx.Done():
			log.Println("Evaluation loop stopped due to context cancellation")
			return
		case <-ticker.C:
			alertConfigs, err := e.alertStore.ListAlerts()
			if err != nil {
				log.Printf("Failed to list alerts, error: %v", err)
				continue
			}
			for _, config := range alertConfigs {
				if !config.Enabled {
					continue
				}
				currentValue, err := e.evaluateMetric(config.Threshold)
				if err != nil {
					log.Printf("Failed to evaluate metric, alert_id: %s, alert_name: %s, error: %v",
						config.ID, config.Name, err)
					continue
				}
				exceeded := e.compareValue(currentValue, config.Threshold.Value, config.Threshold.Operator)
				e.statusMu.Lock()
				status, exists := e.alertStatus[config.ID]
				if !exists {
					status = &models.AlertStatus{
						AlertID: config.ID,
						State:   models.StateInactive,
					}
					e.alertStatus[config.ID] = status
				}
				status.CurrentValue = currentValue
				switch status.State {
				case models.StateInactive:
					if exceeded {
						pendingCounters[config.ID]++
						if pendingCounters[config.ID] >= 1 {
							oldState := status.State
							status.State = models.StatePending
							delete(pendingCounters, config.ID)
							e.generateEvent(oldState, status.State, currentValue, config, status)
						}
					}
				case models.StatePending:
					if !exceeded {
						resolveCounters[config.ID]++
						if resolveCounters[config.ID] >= 1 {
							oldState := status.State
							status.State = models.StateResolved
							delete(resolveCounters, config.ID)
							e.generateEvent(oldState, status.State, currentValue, config, status)
						}
					}
				case models.StateResolved:
					if exceeded {
						pendingCounters[config.ID]++
						if pendingCounters[config.ID] >= 1 {
							oldState := status.State
							status.State = models.StatePending
							delete(pendingCounters, config.ID)
							e.generateEvent(oldState, status.State, currentValue, config, status)
						}
					}
				}
				e.statusMu.Unlock()
			}
		}
	}
}

// generateEvent creates and sends an alert event
func (e *Evaluator) generateEvent(oldState, newState models.AlertState, currentValue float64, config *models.AlertConfig, status *models.AlertStatus) {
	event := models.AlertEvent{
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
		log.Printf("Event channel full, dropping alert event: alert_id=%s alert_name=%s old_state=%v new_state=%v", config.ID, config.Name, oldState, newState)
	}
}

func (e *Evaluator) evaluateMetric(threshold models.ThresholdConfig) (float64, error) {
	switch threshold.MetricType {
	case models.MetricCPU:
		return e.evaluateCPUMetric(threshold.MetricName)
	case models.MetricMemory:
		return e.evaluateMemoryMetric(threshold.MetricName)
	case models.MetricLoad:
		return e.evaluateLoadMetric(threshold.MetricName)
	case models.MetricNetwork:
		return e.evaluateNetworkMetric(threshold.MetricName)
	default:
		return 0, fmt.Errorf("unsupported metric type: %s", threshold.MetricType)
	}
}

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
		log.Printf("Unknown comparison operator: %v", operator)
		return false
	}
}
