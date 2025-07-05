// File: internal/handlers/metrics.go
// Brief: HTTP handlers for metrics endpoints using centralized collector
// Detailed: Implements Gin HTTP handlers for CPU, memory, network, and process metrics that use the centralized metrics collector for improved performance.
// Author: drama.lin@aver.com
// Date: 2024-07-04

package handlers

import (
	"log/slog"
	"net/http"

	"argus/internal/metrics"

	"github.com/gin-gonic/gin"
)

// MetricsHandler provides HTTP handlers for metrics endpoints
type MetricsHandler struct {
	collector *metrics.Collector
}

// NewMetricsHandler creates a new metrics handler instance
func NewMetricsHandler(collector *metrics.Collector) *MetricsHandler {
	return &MetricsHandler{
		collector: collector,
	}
}

// GetCPU handles CPU metrics requests
func (h *MetricsHandler) GetCPU(c *gin.Context) {
	slog.Debug("Fetching cached CPU metrics")

	cpuMetrics := h.collector.GetCPUMetrics()
	if cpuMetrics == nil {
		slog.Error("CPU metrics not available")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "CPU metrics not available",
		})
		return
	}

	slog.Debug("CPU metrics retrieved from cache",
		"load1", cpuMetrics.Load1,
		"load5", cpuMetrics.Load5,
		"load15", cpuMetrics.Load15,
		"usage_percent", cpuMetrics.UsagePercent,
		"updated_at", cpuMetrics.UpdatedAt)

	c.JSON(http.StatusOK, gin.H{
		"load1":         cpuMetrics.Load1,
		"load5":         cpuMetrics.Load5,
		"load15":        cpuMetrics.Load15,
		"usage_percent": cpuMetrics.UsagePercent,
	})
}

// GetMemory handles memory metrics requests
func (h *MetricsHandler) GetMemory(c *gin.Context) {
	slog.Debug("Fetching cached memory metrics")

	memoryMetrics := h.collector.GetMemoryMetrics()
	if memoryMetrics == nil {
		slog.Error("Memory metrics not available")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Memory metrics not available",
		})
		return
	}

	slog.Debug("Memory metrics retrieved from cache",
		"total", memoryMetrics.Total,
		"used", memoryMetrics.Used,
		"free", memoryMetrics.Free,
		"used_percent", memoryMetrics.UsedPercent,
		"updated_at", memoryMetrics.UpdatedAt)

	c.JSON(http.StatusOK, gin.H{
		"total":        memoryMetrics.Total,
		"used":         memoryMetrics.Used,
		"free":         memoryMetrics.Free,
		"used_percent": memoryMetrics.UsedPercent,
	})
}

// GetNetwork handles network metrics requests
func (h *MetricsHandler) GetNetwork(c *gin.Context) {
	slog.Debug("Fetching cached network metrics")

	networkMetrics := h.collector.GetNetworkMetrics()
	if networkMetrics == nil {
		slog.Error("Network metrics not available")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Network metrics not available",
		})
		return
	}

	slog.Debug("Network metrics retrieved from cache",
		"bytes_sent", networkMetrics.BytesSent,
		"bytes_recv", networkMetrics.BytesRecv,
		"packets_sent", networkMetrics.PacketsSent,
		"packets_recv", networkMetrics.PacketsRecv,
		"updated_at", networkMetrics.UpdatedAt)

	c.JSON(http.StatusOK, gin.H{
		"bytes_sent":   networkMetrics.BytesSent,
		"bytes_recv":   networkMetrics.BytesRecv,
		"packets_sent": networkMetrics.PacketsSent,
		"packets_recv": networkMetrics.PacketsRecv,
	})
}

// ProcessQueryParams holds query parameters for process filtering and pagination
type ProcessQueryParams struct {
	Limit        int     `form:"limit"`         // Maximum number of processes to return (default: 50)
	Offset       int     `form:"offset"`        // Number of processes to skip (default: 0)
	SortBy       string  `form:"sort_by"`       // Sort field: cpu, memory, name, pid (default: cpu)
	SortOrder    string  `form:"sort_order"`    // Sort order: asc, desc (default: desc)
	MinCPU       float64 `form:"min_cpu"`       // Minimum CPU percentage filter
	MinMemory    float32 `form:"min_memory"`    // Minimum memory percentage filter
	NameContains string  `form:"name_contains"` // Filter processes by name substring
	TopN         int     `form:"top_n"`         // Get top N processes (efficient heap-based selection)
}

// GetProcess handles process metrics requests with pagination and filtering
func (h *MetricsHandler) GetProcess(c *gin.Context) {
	slog.Debug("Fetching cached process metrics with filters")

	// Parse query parameters from the HTTP request URL
	// This uses Gin's ShouldBindQuery to automatically parse URL query parameters
	// into the ProcessQueryParams struct based on the `form` tags
	var params ProcessQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		// If parsing fails (e.g., invalid data types), return a 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Set defaults for query parameters
	if params.Limit <= 0 || params.Limit > 500 {
		params.Limit = 50 // Default limit, max 500 for safety
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	if params.SortBy == "" {
		params.SortBy = "cpu"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// Validate sort parameters
	validSortFields := map[string]bool{
		"cpu": true, "memory": true, "name": true, "pid": true,
	}
	if !validSortFields[params.SortBy] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sort_by field. Valid values: cpu, memory, name, pid",
		})
		return
	}

	validSortOrders := map[string]bool{
		"asc": true, "desc": true,
	}
	if !validSortOrders[params.SortOrder] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sort_order. Valid values: asc, desc",
		})
		return
	}

	// Get process metrics from collector
	processMetrics := h.collector.GetProcessMetrics()
	if processMetrics == nil {
		slog.Error("Process metrics not available")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Process metrics not available",
		})
		return
	}

	// Get optimized process data based on query parameters
	result, totalCount, err := h.collector.GetOptimizedProcessMetrics(metrics.ProcessFilter{
		Limit:        params.Limit,
		Offset:       params.Offset,
		SortBy:       params.SortBy,
		SortOrder:    params.SortOrder,
		MinCPU:       params.MinCPU,
		MinMemory:    params.MinMemory,
		NameContains: params.NameContains,
		TopN:         params.TopN,
	})

	if err != nil {
		slog.Error("Failed to get optimized process metrics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process metrics",
			"details": err.Error(),
		})
		return
	}

	// Convert to the expected format
	processes := make([]gin.H, len(result))
	for i, p := range result {
		processes[i] = gin.H{
			"pid":         p.PID,
			"name":        p.Name,
			"cpu_percent": p.CPUPercent,
			"mem_percent": p.MemPercent,
		}
	}

	// Calculate pagination metadata
	totalPages := (totalCount + params.Limit - 1) / params.Limit
	currentPage := (params.Offset / params.Limit) + 1
	hasNext := params.Offset+params.Limit < totalCount
	hasPrev := params.Offset > 0

	response := gin.H{
		"processes":   processes,
		"total_count": totalCount,
		"pagination": gin.H{
			"total_count":  totalCount,
			"total_pages":  totalPages,
			"current_page": currentPage,
			"limit":        params.Limit,
			"offset":       params.Offset,
			"has_next":     hasNext,
			"has_previous": hasPrev,
		},
		"filters": gin.H{
			"sort_by":       params.SortBy,
			"sort_order":    params.SortOrder,
			"min_cpu":       params.MinCPU,
			"min_memory":    params.MinMemory,
			"name_contains": params.NameContains,
			"top_n":         params.TopN,
		},
		"updated_at": processMetrics.UpdatedAt,
	}

	slog.Debug("Process metrics retrieved with optimization",
		"total_processes", totalCount,
		"returned_processes", len(processes),
		"limit", params.Limit,
		"offset", params.Offset,
		"sort_by", params.SortBy,
		"updated_at", processMetrics.UpdatedAt)

	c.JSON(http.StatusOK, response)
}

// GetMetricsHealth returns health status of the metrics collector
func (h *MetricsHandler) GetMetricsHealth(c *gin.Context) {
	healthy := h.collector.IsHealthy()

	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"healthy": healthy,
	})
}
