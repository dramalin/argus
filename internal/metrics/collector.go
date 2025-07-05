// File: internal/metrics/collector.go
// Brief: Centralized metrics collection system with caching for Argus
// Detailed: Implements a background metrics collector that caches CPU, memory, network, and process metrics to reduce HTTP response latency and system load.
// Author: drama.lin@aver.com
// Date: 2024-07-04

package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// CollectorConfig holds configuration for the metrics collector
type CollectorConfig struct {
	UpdateInterval time.Duration // How often to update metrics
	CacheTTL       time.Duration // How long cached metrics are valid
	ProcessLimit   int           // Maximum number of processes to collect
}

// DefaultConfig returns default configuration for the metrics collector
func DefaultConfig() CollectorConfig {
	return CollectorConfig{
		UpdateInterval: 5 * time.Second,
		CacheTTL:       10 * time.Second,
		ProcessLimit:   100,
	}
}

// CPUMetrics holds CPU-related metrics
type CPUMetrics struct {
	Load1        float64   `json:"load1"`
	Load5        float64   `json:"load5"`
	Load15       float64   `json:"load15"`
	UsagePercent float64   `json:"usage_percent"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MemoryMetrics holds memory-related metrics
type MemoryMetrics struct {
	Total       uint64    `json:"total"`
	Used        uint64    `json:"used"`
	Free        uint64    `json:"free"`
	UsedPercent float64   `json:"used_percent"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NetworkMetrics holds network-related metrics
type NetworkMetrics struct {
	BytesSent   uint64    `json:"bytes_sent"`
	BytesRecv   uint64    `json:"bytes_recv"`
	PacketsSent uint64    `json:"packets_sent"`
	PacketsRecv uint64    `json:"packets_recv"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProcessInfo holds information about a single process
type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float32 `json:"mem_percent"`
}

// ProcessMetrics holds process-related metrics
type ProcessMetrics struct {
	Processes []ProcessInfo `json:"processes"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// ProcessFilter defines filtering and pagination options for process metrics
type ProcessFilter struct {
	Limit        int     // Maximum number of processes to return
	Offset       int     // Number of processes to skip
	SortBy       string  // Sort field: cpu, memory, name, pid
	SortOrder    string  // Sort order: asc, desc
	MinCPU       float64 // Minimum CPU percentage filter
	MinMemory    float32 // Minimum memory percentage filter
	NameContains string  // Filter processes by name substring
	TopN         int     // Get top N processes (efficient heap-based selection)
}

// Collector manages centralized metrics collection with caching
type Collector struct {
	config CollectorConfig

	// Cached metrics with RWMutex for concurrent access
	cpuMutex   sync.RWMutex
	cpuMetrics *CPUMetrics

	memoryMutex   sync.RWMutex
	memoryMetrics *MemoryMetrics

	networkMutex   sync.RWMutex
	networkMetrics *NetworkMetrics

	processMutex   sync.RWMutex
	processMetrics *ProcessMetrics

	// Object pools for reducing allocations
	processInfoPool sync.Pool
	stringSlicePool sync.Pool

	// Control channels
	stopChan chan struct{}
	doneChan chan struct{}
}

// NewCollector creates a new metrics collector instance
func NewCollector(config CollectorConfig) *Collector {
	return &Collector{
		config:   config,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		processInfoPool: sync.Pool{
			New: func() interface{} {
				return make([]ProcessInfo, 0, config.ProcessLimit)
			},
		},
		stringSlicePool: sync.Pool{
			New: func() interface{} {
				return make([]string, 0, 16)
			},
		},
	}
}

// Start begins the background metrics collection
func (c *Collector) Start(ctx context.Context) error {
	slog.Info("Starting metrics collector", "update_interval", c.config.UpdateInterval)

	// Collect initial metrics
	c.collectAllMetrics(ctx)

	// Start background collection goroutine
	go c.collectLoop(ctx)

	return nil
}

// Stop stops the background metrics collection
func (c *Collector) Stop() {
	slog.Info("Stopping metrics collector")
	close(c.stopChan)
	<-c.doneChan
}

// collectLoop runs the background metrics collection
func (c *Collector) collectLoop(ctx context.Context) {
	defer close(c.doneChan)

	ticker := time.NewTicker(c.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Metrics collector stopped due to context cancellation")
			return
		case <-c.stopChan:
			slog.Info("Metrics collector stopped")
			return
		case <-ticker.C:
			c.collectAllMetrics(ctx)
		}
	}
}

// collectAllMetrics collects all types of metrics
func (c *Collector) collectAllMetrics(ctx context.Context) {
	// Use separate goroutines for parallel collection
	var wg sync.WaitGroup

	wg.Add(4)
	go func() {
		defer wg.Done()
		c.collectCPUMetrics(ctx)
	}()

	go func() {
		defer wg.Done()
		c.collectMemoryMetrics(ctx)
	}()

	go func() {
		defer wg.Done()
		c.collectNetworkMetrics(ctx)
	}()

	go func() {
		defer wg.Done()
		c.collectProcessMetrics(ctx)
	}()

	wg.Wait()
}

// collectCPUMetrics collects CPU metrics
func (c *Collector) collectCPUMetrics(ctx context.Context) {
	loadAvg, err := load.AvgWithContext(ctx)
	if err != nil {
		slog.Error("Failed to get load average", "error", err)
		return
	}

	cpuPercent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		slog.Error("Failed to get CPU percent", "error", err)
		return
	}

	var usage float64
	if len(cpuPercent) > 0 {
		usage = cpuPercent[0]
	}

	metrics := &CPUMetrics{
		Load1:        loadAvg.Load1,
		Load5:        loadAvg.Load5,
		Load15:       loadAvg.Load15,
		UsagePercent: usage,
		UpdatedAt:    time.Now(),
	}

	c.cpuMutex.Lock()
	c.cpuMetrics = metrics
	c.cpuMutex.Unlock()

	slog.Debug("CPU metrics updated", "usage_percent", usage, "load1", loadAvg.Load1)
}

// collectMemoryMetrics collects memory metrics
func (c *Collector) collectMemoryMetrics(ctx context.Context) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		slog.Error("Failed to get memory info", "error", err)
		return
	}

	metrics := &MemoryMetrics{
		Total:       vm.Total,
		Used:        vm.Used,
		Free:        vm.Free,
		UsedPercent: vm.UsedPercent,
		UpdatedAt:   time.Now(),
	}

	c.memoryMutex.Lock()
	c.memoryMetrics = metrics
	c.memoryMutex.Unlock()

	slog.Debug("Memory metrics updated", "used_percent", vm.UsedPercent, "total", vm.Total)
}

// collectNetworkMetrics collects network metrics
func (c *Collector) collectNetworkMetrics(ctx context.Context) {
	ioCounters, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		slog.Error("Failed to get network stats", "error", err)
		return
	}

	if len(ioCounters) == 0 {
		slog.Warn("No network interfaces found")
		return
	}

	io := ioCounters[0]
	metrics := &NetworkMetrics{
		BytesSent:   io.BytesSent,
		BytesRecv:   io.BytesRecv,
		PacketsSent: io.PacketsSent,
		PacketsRecv: io.PacketsRecv,
		UpdatedAt:   time.Now(),
	}

	c.networkMutex.Lock()
	c.networkMetrics = metrics
	c.networkMutex.Unlock()

	slog.Debug("Network metrics updated", "bytes_sent", io.BytesSent, "bytes_recv", io.BytesRecv)
}

// collectProcessMetrics collects process metrics
func (c *Collector) collectProcessMetrics(ctx context.Context) {
	// Add timeout to prevent hanging
	processCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	procs, err := process.ProcessesWithContext(processCtx)
	if err != nil {
		slog.Error("Failed to get process list", "error", err)
		return
	}

	// Get process info slice from pool
	processes := c.processInfoPool.Get().([]ProcessInfo)
	processes = processes[:0] // Reset slice but keep capacity

	processedCount := 0
	errorCount := 0

	// Collect process information with error handling
	for _, p := range procs {
		// Check context cancellation
		select {
		case <-processCtx.Done():
			slog.Warn("Process metrics collection cancelled due to timeout")
			break
		default:
		}

		if p == nil || p.Pid <= 0 {
			continue
		}

		// Limit number of processes to prevent excessive memory usage
		if len(processes) >= c.config.ProcessLimit {
			break
		}

		// Process individual process with error recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Debug("Panic occurred processing process", "pid", p.Pid, "panic", r)
					errorCount++
				}
			}()

			processedCount++

			name, err := p.NameWithContext(processCtx)
			if err != nil {
				slog.Debug("Failed to get process name", "pid", p.Pid, "error", err)
				return
			}

			// Skip kernel threads and system processes
			if name == "" || (len(name) > 0 && name[0] == '[') {
				return
			}

			var cpuP float64 = 0.0
			var memP float32 = 0.0

			// Get CPU percentage with error handling
			if cpu, err := p.CPUPercentWithContext(processCtx); err == nil {
				cpuP = cpu
			}

			// Get memory percentage with error handling
			if mem, err := p.MemoryPercentWithContext(processCtx); err == nil {
				memP = mem
			}

			processes = append(processes, ProcessInfo{
				PID:        p.Pid,
				Name:       name,
				CPUPercent: cpuP,
				MemPercent: memP,
			})
		}()
	}

	// Sort by CPU percentage in descending order
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPUPercent > processes[j].CPUPercent
	})

	// Create metrics with copied slice to avoid pool interference
	processSlice := make([]ProcessInfo, len(processes))
	copy(processSlice, processes)

	metrics := &ProcessMetrics{
		Processes: processSlice,
		UpdatedAt: time.Now(),
	}

	c.processMutex.Lock()
	c.processMetrics = metrics
	c.processMutex.Unlock()

	// Return slice to pool
	c.processInfoPool.Put(processes)

	slog.Debug("Process metrics updated",
		"total_processed", processedCount,
		"successful", len(processSlice),
		"errors", errorCount)
}

// GetCPUMetrics returns cached CPU metrics
func (c *Collector) GetCPUMetrics() *CPUMetrics {
	c.cpuMutex.RLock()
	defer c.cpuMutex.RUnlock()

	if c.cpuMetrics == nil {
		return nil
	}

	// Check if cache is still valid
	if time.Since(c.cpuMetrics.UpdatedAt) > c.config.CacheTTL {
		slog.Debug("CPU metrics cache expired")
		return nil
	}

	// Return a copy to prevent race conditions
	metrics := *c.cpuMetrics
	return &metrics
}

// GetMemoryMetrics returns cached memory metrics
func (c *Collector) GetMemoryMetrics() *MemoryMetrics {
	c.memoryMutex.RLock()
	defer c.memoryMutex.RUnlock()

	if c.memoryMetrics == nil {
		return nil
	}

	if time.Since(c.memoryMetrics.UpdatedAt) > c.config.CacheTTL {
		slog.Debug("Memory metrics cache expired")
		return nil
	}

	metrics := *c.memoryMetrics
	return &metrics
}

// GetNetworkMetrics returns cached network metrics
func (c *Collector) GetNetworkMetrics() *NetworkMetrics {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()

	if c.networkMetrics == nil {
		return nil
	}

	if time.Since(c.networkMetrics.UpdatedAt) > c.config.CacheTTL {
		slog.Debug("Network metrics cache expired")
		return nil
	}

	metrics := *c.networkMetrics
	return &metrics
}

// GetProcessMetrics returns cached process metrics
func (c *Collector) GetProcessMetrics() *ProcessMetrics {
	c.processMutex.RLock()
	defer c.processMutex.RUnlock()

	if c.processMetrics == nil {
		return nil
	}

	if time.Since(c.processMetrics.UpdatedAt) > c.config.CacheTTL {
		slog.Debug("Process metrics cache expired")
		return nil
	}

	// Return a copy with copied slice to prevent race conditions
	processes := make([]ProcessInfo, len(c.processMetrics.Processes))
	copy(processes, c.processMetrics.Processes)

	metrics := &ProcessMetrics{
		Processes: processes,
		UpdatedAt: c.processMetrics.UpdatedAt,
	}

	return metrics
}

// GetOptimizedProcessMetrics returns filtered and paginated process metrics with efficient algorithms
func (c *Collector) GetOptimizedProcessMetrics(filter ProcessFilter) ([]ProcessInfo, int, error) {
	c.processMutex.RLock()
	defer c.processMutex.RUnlock()

	if c.processMetrics == nil {
		return nil, 0, fmt.Errorf("process metrics not available")
	}

	if time.Since(c.processMetrics.UpdatedAt) > c.config.CacheTTL {
		return nil, 0, fmt.Errorf("process metrics cache expired")
	}

	processes := c.processMetrics.Processes

	// Apply filters first to reduce dataset size
	filtered := c.applyProcessFilters(processes, filter)
	totalCount := len(filtered)

	// Handle top-N selection with heap-based algorithm for efficiency
	if filter.TopN > 0 && filter.TopN < len(filtered) {
		topProcesses := c.selectTopNProcesses(filtered, filter.TopN, filter.SortBy, filter.SortOrder)
		return topProcesses, totalCount, nil
	}

	// Apply sorting
	c.sortProcesses(filtered, filter.SortBy, filter.SortOrder)

	// Apply pagination
	start := filter.Offset
	if start > len(filtered) {
		return []ProcessInfo{}, totalCount, nil
	}

	end := start + filter.Limit
	if end > len(filtered) {
		end = len(filtered)
	}

	result := make([]ProcessInfo, end-start)
	copy(result, filtered[start:end])

	return result, totalCount, nil
}

// applyProcessFilters applies filtering criteria to process list
func (c *Collector) applyProcessFilters(processes []ProcessInfo, filter ProcessFilter) []ProcessInfo {
	if filter.MinCPU == 0 && filter.MinMemory == 0 && filter.NameContains == "" {
		// No filters to apply, return copy of original slice
		result := make([]ProcessInfo, len(processes))
		copy(result, processes)
		return result
	}

	// Pre-allocate with estimated capacity to reduce allocations
	filtered := make([]ProcessInfo, 0, len(processes)/2)

	for _, p := range processes {
		// Apply CPU filter
		if filter.MinCPU > 0 && p.CPUPercent < filter.MinCPU {
			continue
		}

		// Apply memory filter
		if filter.MinMemory > 0 && p.MemPercent < filter.MinMemory {
			continue
		}

		// Apply name filter (case-insensitive substring match)
		if filter.NameContains != "" {
			if !strings.Contains(strings.ToLower(p.Name), strings.ToLower(filter.NameContains)) {
				continue
			}
		}

		filtered = append(filtered, p)
	}

	return filtered
}

// selectTopNProcesses uses a heap-based algorithm to efficiently select top N processes
func (c *Collector) selectTopNProcesses(processes []ProcessInfo, n int, sortBy, sortOrder string) []ProcessInfo {
	if n >= len(processes) {
		// If n is larger than available processes, sort and return all
		c.sortProcesses(processes, sortBy, sortOrder)
		return processes
	}

	// Use heap-based selection for efficiency - O(n log k) instead of O(n log n)
	heap := make([]ProcessInfo, 0, n)

	// Build min-heap for descending order, max-heap for ascending order
	isMinHeap := sortOrder == "desc"

	for _, p := range processes {
		if len(heap) < n {
			// Heap not full, add element
			heap = append(heap, p)
			if len(heap) == n {
				// Heapify when full
				c.heapifyProcesses(heap, sortBy, isMinHeap)
			}
		} else {
			// Heap full, check if current element should replace heap root
			if c.shouldReplaceHeapRoot(heap[0], p, sortBy, isMinHeap) {
				heap[0] = p
				c.heapifyDownProcesses(heap, 0, sortBy, isMinHeap)
			}
		}
	}

	// Sort the final heap for proper ordering
	c.sortProcesses(heap, sortBy, sortOrder)
	return heap
}

// sortProcesses sorts processes by the specified field and order
func (c *Collector) sortProcesses(processes []ProcessInfo, sortBy, sortOrder string) {
	less := c.getProcessComparator(processes, sortBy, sortOrder == "asc")

	// Use Go's optimized sort algorithm
	sort.Slice(processes, less)
}

// getProcessComparator returns a comparison function for the specified field and order
func (c *Collector) getProcessComparator(processes []ProcessInfo, sortBy string, ascending bool) func(i, j int) bool {
	return func(i, j int) bool {
		var result bool

		switch sortBy {
		case "cpu":
			result = processes[i].CPUPercent < processes[j].CPUPercent
		case "memory":
			result = processes[i].MemPercent < processes[j].MemPercent
		case "name":
			result = strings.ToLower(processes[i].Name) < strings.ToLower(processes[j].Name)
		case "pid":
			result = processes[i].PID < processes[j].PID
		default:
			// Default to CPU sorting
			result = processes[i].CPUPercent < processes[j].CPUPercent
		}

		if !ascending {
			result = !result
		}

		return result
	}
}

// heapifyProcesses converts slice to heap structure
func (c *Collector) heapifyProcesses(heap []ProcessInfo, sortBy string, isMinHeap bool) {
	n := len(heap)
	// Start from last non-leaf node
	for i := n/2 - 1; i >= 0; i-- {
		c.heapifyDownProcesses(heap, i, sortBy, isMinHeap)
	}
}

// heapifyDownProcesses maintains heap property downward from given index
func (c *Collector) heapifyDownProcesses(heap []ProcessInfo, i int, sortBy string, isMinHeap bool) {
	n := len(heap)
	for {
		largest := i
		left := 2*i + 1
		right := 2*i + 2

		// Compare with left child
		if left < n && c.compareProcesses(heap[left], heap[largest], sortBy, isMinHeap) {
			largest = left
		}

		// Compare with right child
		if right < n && c.compareProcesses(heap[right], heap[largest], sortBy, isMinHeap) {
			largest = right
		}

		if largest == i {
			break
		}

		// Swap and continue
		heap[i], heap[largest] = heap[largest], heap[i]
		i = largest
	}
}

// shouldReplaceHeapRoot checks if new process should replace heap root
func (c *Collector) shouldReplaceHeapRoot(root, candidate ProcessInfo, sortBy string, isMinHeap bool) bool {
	return c.compareProcesses(candidate, root, sortBy, !isMinHeap)
}

// compareProcesses compares two processes based on the specified field
func (c *Collector) compareProcesses(a, b ProcessInfo, sortBy string, aIsGreater bool) bool {
	var result bool

	switch sortBy {
	case "cpu":
		result = a.CPUPercent > b.CPUPercent
	case "memory":
		result = a.MemPercent > b.MemPercent
	case "name":
		result = strings.ToLower(a.Name) > strings.ToLower(b.Name)
	case "pid":
		result = a.PID > b.PID
	default:
		result = a.CPUPercent > b.CPUPercent
	}

	if !aIsGreater {
		result = !result
	}

	return result
}

// IsHealthy returns true if all metrics are being collected successfully
func (c *Collector) IsHealthy() bool {
	now := time.Now()

	c.cpuMutex.RLock()
	cpuHealthy := c.cpuMetrics != nil && now.Sub(c.cpuMetrics.UpdatedAt) < c.config.CacheTTL*2
	c.cpuMutex.RUnlock()

	c.memoryMutex.RLock()
	memoryHealthy := c.memoryMetrics != nil && now.Sub(c.memoryMetrics.UpdatedAt) < c.config.CacheTTL*2
	c.memoryMutex.RUnlock()

	c.networkMutex.RLock()
	networkHealthy := c.networkMetrics != nil && now.Sub(c.networkMetrics.UpdatedAt) < c.config.CacheTTL*2
	c.networkMutex.RUnlock()

	c.processMutex.RLock()
	processHealthy := c.processMetrics != nil && now.Sub(c.processMetrics.UpdatedAt) < c.config.CacheTTL*2
	c.processMutex.RUnlock()

	return cpuHealthy && memoryHealthy && networkHealthy && processHealthy
}
