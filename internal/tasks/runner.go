// Package tasks provides functionality for scheduling and managing system maintenance tasks
package tasks

// All runner and business logic has been migrated to internal/sync/service.go as part of the architecture migration.
// This file remains as a stub for compatibility and to avoid breaking imports during transition.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"argus/internal/models"
	"argus/internal/services"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	// ErrUnsupportedTaskType is returned when a task type is not supported
	ErrUnsupportedTaskType = errors.New("unsupported task type")

	// ErrTaskCancelled is returned when a task is cancelled during execution
	ErrTaskCancelled = errors.New("task cancelled")

	// ErrInvalidParameter is returned when a task parameter is invalid
	ErrInvalidParameter = errors.New("invalid task parameter")
)

// LogRotationRunner implements TaskRunner for log rotation tasks
type LogRotationRunner struct {
	services.BaseTaskRunner
}

// GetType returns the task type that this runner handles
func (r *LogRotationRunner) GetType() models.TaskType {
	return models.TaskLogRotation
}

// Execute executes a log rotation task
func (r *LogRotationRunner) Execute(ctx context.Context, task models.TaskConfig) (*models.TaskResult, error) {
	slog.Info("Starting log rotation task", "task_id", task.ID, "name", task.Name)

	result := &models.TaskResult{
		ExecutionID: models.GenerateID(),
		StartTime:   time.Now(),
		Status:      models.StatusRunning,
	}

	// Get parameters with defaults
	logDir := "/var/log"
	if task.Parameters != nil {
		if val, ok := task.Parameters["log_dir"]; ok {
			logDir = val
		}
	}

	maxSizeMB := 10
	if task.Parameters != nil {
		if val, ok := task.Parameters["max_size_mb"]; ok {
			maxSize, err := strconv.Atoi(val)
			if err != nil {
				result.Status = models.StatusFailed
				result.Output = fmt.Sprintf("Invalid max_size_mb parameter: %v", err)
				result.EndTime = time.Now()
				return result, fmt.Errorf("invalid max_size_mb parameter: %w", err)
			}
			maxSizeMB = maxSize
		}
	}

	keepCount := 5
	if task.Parameters != nil {
		if val, ok := task.Parameters["keep_count"]; ok {
			keep, err := strconv.Atoi(val)
			if err != nil {
				result.Status = models.StatusFailed
				result.Output = fmt.Sprintf("Invalid keep_count parameter: %v", err)
				result.EndTime = time.Now()
				return result, fmt.Errorf("invalid keep_count parameter: %w", err)
			}
			keepCount = keep
		}
	}

	// Process log files
	var output strings.Builder
	rotatedCount := 0
	processedCount := 0
	maxSizeBytes := int64(maxSizeMB * 1024 * 1024)

	// Walk through the log directory
	var err error
	err = filepath.WalkDir(logDir, func(path string, d fs.DirEntry, err error) error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		if err != nil {
			fmt.Fprintf(&output, "Error accessing path %s: %v\n", path, err)
			return nil
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .log files
		if !strings.HasSuffix(d.Name(), ".log") {
			return nil
		}

		processedCount++

		// Get file info
		info, err := d.Info()
		if err != nil {
			fmt.Fprintf(&output, "Error getting file info for %s: %v\n", path, err)
			return nil
		}

		// Check if file needs rotation
		if info.Size() > maxSizeBytes {
			if err := rotateLogFile(path, keepCount); err != nil {
				fmt.Fprintf(&output, "Error rotating log file %s: %v\n", path, err)
			} else {
				rotatedCount++
				fmt.Fprintf(&output, "Rotated log file: %s\n", path)
			}
		}

		return nil
	})

	// Handle errors or context cancellation
	if err != nil {
		if errors.Is(err, context.Canceled) {
			result.Status = models.StatusFailed
			result.Output = fmt.Sprintf("Log rotation task cancelled: %v", err)
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, err)
		}
		result.Status = models.StatusFailed
		result.Output = fmt.Sprintf("Error during log rotation: %v", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Record successful execution
	fmt.Fprintf(&output, "Log rotation completed. Processed %d files, rotated %d files.\n", processedCount, rotatedCount)
	result.Status = models.StatusCompleted
	result.Output = output.String()
	result.EndTime = time.Now()

	slog.Info("Log rotation task completed",
		"task_id", task.ID,
		"processed", processedCount,
		"rotated", rotatedCount)

	return result, nil
}

// MetricsAggregationRunner implements TaskRunner for metrics aggregation tasks
type MetricsAggregationRunner struct {
	services.BaseTaskRunner
}

// GetType returns the task type that this runner handles
func (r *MetricsAggregationRunner) GetType() models.TaskType {
	return models.TaskMetricsAggregation
}

// Execute executes a metrics aggregation task
func (r *MetricsAggregationRunner) Execute(ctx context.Context, task models.TaskConfig) (*models.TaskResult, error) {
	slog.Info("Starting metrics aggregation task", "task_id", task.ID, "name", task.Name)

	result := &models.TaskResult{
		ExecutionID: models.GenerateID(),
		StartTime:   time.Now(),
		Status:      models.StatusRunning,
	}

	// Get parameters with defaults
	outputDir := ".argus/metrics"
	if val, ok := task.Parameters["output_dir"]; ok {
		outputDir = val
	}

	includeCPU := true
	if val, ok := task.Parameters["include_cpu"]; ok {
		includeCPU = val == "true"
	}

	includeMemory := true
	if val, ok := task.Parameters["include_memory"]; ok {
		includeMemory = val == "true"
	}

	includeDisk := true
	if val, ok := task.Parameters["include_disk"]; ok {
		includeDisk = val == "true"
	}

	includeNetwork := true
	if val, ok := task.Parameters["include_network"]; ok {
		includeNetwork = val == "true"
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		result.Status = models.StatusFailed
		result.Output = fmt.Sprintf("Failed to create output directory: %v", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Timestamp for this metrics collection
	timestamp := time.Now()
	filename := filepath.Join(outputDir, fmt.Sprintf("metrics-%s.txt", timestamp.Format("20060102-150405")))

	// Collect and write metrics
	var output strings.Builder
	fmt.Fprintf(&output, "System Metrics Collection - %s\n", timestamp.Format(time.RFC3339))
	fmt.Fprintf(&output, "========================================\n\n")

	// Collect CPU metrics if requested
	if includeCPU {
		select {
		case <-ctx.Done():
			result.Status = models.StatusFailed
			result.Output = "Task cancelled during CPU metrics collection"
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, ctx.Err())
		default:
			if err := collectCPUMetrics(&output); err != nil {
				fmt.Fprintf(&output, "Error collecting CPU metrics: %v\n", err)
			}
		}
	}

	// Collect memory metrics if requested
	if includeMemory {
		select {
		case <-ctx.Done():
			result.Status = models.StatusFailed
			result.Output = "Task cancelled during memory metrics collection"
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, ctx.Err())
		default:
			if err := collectMemoryMetrics(&output); err != nil {
				fmt.Fprintf(&output, "Error collecting memory metrics: %v\n", err)
			}
		}
	}

	// Collect disk metrics if requested
	if includeDisk {
		select {
		case <-ctx.Done():
			result.Status = models.StatusFailed
			result.Output = "Task cancelled during disk metrics collection"
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, ctx.Err())
		default:
			if err := collectDiskMetrics(&output); err != nil {
				fmt.Fprintf(&output, "Error collecting disk metrics: %v\n", err)
			}
		}
	}

	// Collect network metrics if requested
	if includeNetwork {
		select {
		case <-ctx.Done():
			result.Status = models.StatusFailed
			result.Output = "Task cancelled during network metrics collection"
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, ctx.Err())
		default:
			if err := collectNetworkMetrics(&output); err != nil {
				fmt.Fprintf(&output, "Error collecting network metrics: %v\n", err)
			}
		}
	}

	// Write output to file
	if err := ioutil.WriteFile(filename, []byte(output.String()), 0644); err != nil {
		result.Status = models.StatusFailed
		result.Output = fmt.Sprintf("Failed to write metrics to file: %v", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Record successful execution
	resultSummary := fmt.Sprintf("Metrics collected and saved to %s", filename)
	result.Status = models.StatusCompleted
	result.Output = resultSummary
	result.EndTime = time.Now()

	slog.Info("Metrics aggregation task completed", "task_id", task.ID, "output_file", filename)

	return result, nil
}

// HealthCheckRunner implements TaskRunner for system health check tasks
type HealthCheckRunner struct {
	services.BaseTaskRunner
}

// GetType returns the task type that this runner handles
func (r *HealthCheckRunner) GetType() models.TaskType {
	return models.TaskHealthCheck
}

// Execute executes a health check task
func (r *HealthCheckRunner) Execute(ctx context.Context, task models.TaskConfig) (*models.TaskResult, error) {
	slog.Info("Starting health check task", "task_id", task.ID, "name", task.Name)

	result := &models.TaskResult{
		ExecutionID: models.GenerateID(),
		StartTime:   time.Now(),
		Status:      models.StatusRunning,
	}

	// Get parameters with defaults
	checkDiskSpace := true
	if val, ok := task.Parameters["check_disk_space"]; ok {
		checkDiskSpace = val == "true"
	}

	checkCPULoad := true
	if val, ok := task.Parameters["check_cpu_load"]; ok {
		checkCPULoad = val == "true"
	}

	checkMemory := true
	if val, ok := task.Parameters["check_memory"]; ok {
		checkMemory = val == "true"
	}

	checkEndpoints := ""
	if val, ok := task.Parameters["check_endpoints"]; ok {
		checkEndpoints = val
	}

	var output strings.Builder
	fmt.Fprintf(&output, "Health Check Results - %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&output, "========================================\n\n")

	allHealthy := true

	// Check disk space
	if checkDiskSpace {
		select {
		case <-ctx.Done():
			result.Status = models.StatusFailed
			result.Output = "Task cancelled during disk space check"
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, ctx.Err())
		default:
			healthy, results := checkDiskSpaceHealth()
			fmt.Fprintf(&output, "Disk Space Check: %s\n%s\n", healthStatusString(healthy), results)
			allHealthy = allHealthy && healthy
		}
	}

	// Check CPU load
	if checkCPULoad {
		select {
		case <-ctx.Done():
			result.Status = models.StatusFailed
			result.Output = "Task cancelled during CPU load check"
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, ctx.Err())
		default:
			healthy, results := checkCPULoadHealth()
			fmt.Fprintf(&output, "CPU Load Check: %s\n%s\n", healthStatusString(healthy), results)
			allHealthy = allHealthy && healthy
		}
	}

	// Check memory
	if checkMemory {
		select {
		case <-ctx.Done():
			result.Status = models.StatusFailed
			result.Output = "Task cancelled during memory check"
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, ctx.Err())
		default:
			healthy, results := checkMemoryHealth()
			fmt.Fprintf(&output, "Memory Check: %s\n%s\n", healthStatusString(healthy), results)
			allHealthy = allHealthy && healthy
		}
	}

	// Check HTTP endpoints if specified
	if checkEndpoints != "" {
		endpoints := strings.Split(checkEndpoints, ",")
		for _, endpoint := range endpoints {
			endpoint = strings.TrimSpace(endpoint)
			if endpoint == "" {
				continue
			}

			select {
			case <-ctx.Done():
				result.Status = models.StatusFailed
				result.Output = "Task cancelled during endpoint check"
				result.EndTime = time.Now()
				return result, fmt.Errorf("%w: %v", ErrTaskCancelled, ctx.Err())
			default:
				healthy, results := checkEndpointHealth(endpoint)
				fmt.Fprintf(&output, "Endpoint Check (%s): %s\n%s\n", endpoint, healthStatusString(healthy), results)
				allHealthy = allHealthy && healthy
			}
		}
	}

	// Set overall health status
	fmt.Fprintf(&output, "\nOverall System Health: %s\n", healthStatusString(allHealthy))

	// Record execution result
	result.Status = models.StatusCompleted
	result.Output = output.String()
	result.EndTime = time.Now()
	result.Metadata = map[string]string{
		"all_healthy": strconv.FormatBool(allHealthy),
	}

	slog.Info("Health check task completed", "task_id", task.ID, "healthy", allHealthy)

	return result, nil
}

// SystemCleanupRunner implements TaskRunner for system cleanup tasks
type SystemCleanupRunner struct {
	services.BaseTaskRunner
}

// GetType returns the task type that this runner handles
func (r *SystemCleanupRunner) GetType() models.TaskType {
	return models.TaskSystemCleanup
}

// Execute executes a system cleanup task
func (r *SystemCleanupRunner) Execute(ctx context.Context, task models.TaskConfig) (*models.TaskResult, error) {
	slog.Info("Starting system cleanup task", "task_id", task.ID, "name", task.Name)

	result := &models.TaskResult{
		ExecutionID: models.GenerateID(),
		StartTime:   time.Now(),
		Status:      models.StatusRunning,
	}

	// Get parameters with defaults
	tmpDir := "/tmp"
	if val, ok := task.Parameters["tmp_dir"]; ok {
		tmpDir = val
	}

	oldestDays := 7
	if val, ok := task.Parameters["oldest_days"]; ok {
		days, err := strconv.Atoi(val)
		if err != nil {
			result.Status = models.StatusFailed
			result.Output = fmt.Sprintf("Invalid oldest_days parameter: %v", err)
			result.EndTime = time.Now()
			return result, fmt.Errorf("invalid oldest_days parameter: %w", err)
		}
		oldestDays = days
	}

	excludePatterns := []string{}
	if val, ok := task.Parameters["exclude_patterns"]; ok {
		patterns := strings.Split(val, ",")
		for _, pattern := range patterns {
			if trimmed := strings.TrimSpace(pattern); trimmed != "" {
				excludePatterns = append(excludePatterns, trimmed)
			}
		}
	}

	// Calculate the cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -oldestDays)

	var output strings.Builder
	removedCount := 0
	processedCount := 0
	totalBytes := int64(0)

	// Walk through the temporary directory
	var walkErr error
	walkErr = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		if err != nil {
			fmt.Fprintf(&output, "Error accessing path %s: %v\n", path, err)
			return nil
		}

		// Skip the root directory
		if path == tmpDir {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			fmt.Fprintf(&output, "Error getting file info for %s: %v\n", path, err)
			return nil
		}

		processedCount++

		// Skip if file/dir is newer than cutoff
		if info.ModTime().After(cutoffTime) {
			return nil
		}

		// Check if path matches any exclude patterns
		for _, pattern := range excludePatterns {
			if pattern != "" && strings.Contains(path, pattern) {
				return nil
			}
		}

		// Remove the file or directory
		fileSize := info.Size()
		if err := os.RemoveAll(path); err != nil {
			fmt.Fprintf(&output, "Error removing %s: %v\n", path, err)
		} else {
			removedCount++
			totalBytes += fileSize
			fmt.Fprintf(&output, "Removed: %s (size: %d bytes)\n", path, fileSize)

			// If it's a directory, don't process its contents
			if d.IsDir() {
				return filepath.SkipDir
			}
		}

		return nil
	})

	// Handle errors or context cancellation
	if walkErr != nil {
		if errors.Is(walkErr, context.Canceled) {
			result.Status = models.StatusFailed
			result.Output = fmt.Sprintf("System cleanup task cancelled: %v", walkErr)
			result.EndTime = time.Now()
			return result, fmt.Errorf("%w: %v", ErrTaskCancelled, walkErr)
		}
		result.Status = models.StatusFailed
		result.Output = fmt.Sprintf("Error during system cleanup: %v", walkErr)
		result.EndTime = time.Now()
		return result, walkErr
	}

	// Record successful execution
	resultSummary := fmt.Sprintf("Cleanup completed. Processed %d items, removed %d items, freed %d bytes.",
		processedCount, removedCount, totalBytes)
	result.Status = models.StatusCompleted
	result.Output = output.String() + "\n\n" + resultSummary
	result.EndTime = time.Now()
	result.Metadata = map[string]string{
		"processed_count": strconv.Itoa(processedCount),
		"removed_count":   strconv.Itoa(removedCount),
		"freed_bytes":     strconv.FormatInt(totalBytes, 10),
	}

	slog.Info("System cleanup task completed",
		"task_id", task.ID,
		"processed", processedCount,
		"removed", removedCount,
		"freed_bytes", totalBytes)

	return result, nil
}

// Helper functions for task runners

// formatBytes converts bytes to a human-readable string (e.g., "1.5 GB")
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// healthStatusString returns a string representation of a health check status
func healthStatusString(healthy bool) string {
	if healthy {
		return "HEALTHY"
	}
	return "UNHEALTHY"
}

// rotateLogFile rotates a log file, keeping a specified number of backups
func rotateLogFile(filePath string, keepCount int) error {
	// Remove the oldest backup if it exists
	oldestBackup := fmt.Sprintf("%s.%d", filePath, keepCount)
	os.Remove(oldestBackup) // Ignore errors, file may not exist

	// Shift existing backups
	for i := keepCount - 1; i >= 1; i-- {
		oldFile := fmt.Sprintf("%s.%d", filePath, i)
		newFile := fmt.Sprintf("%s.%d", filePath, i+1)
		os.Rename(oldFile, newFile) // Ignore errors, file may not exist
	}

	// Rename current log to .1
	backupFile := fmt.Sprintf("%s.1", filePath)
	if err := os.Rename(filePath, backupFile); err != nil {
		return err
	}

	// Create new empty log file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	return file.Close()
}

// collectCPUMetrics collects and writes CPU metrics
func collectCPUMetrics(output *strings.Builder) error {
	fmt.Fprintf(output, "CPU Metrics:\n")
	fmt.Fprintf(output, "------------\n")

	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}

	fmt.Fprintf(output, "CPU Usage: %.2f%%\n", cpuPercent[0])

	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		fmt.Fprintf(output, "Error getting CPU info: %v\n", err)
	} else if len(cpuInfo) > 0 {
		fmt.Fprintf(output, "CPU Model: %s\n", cpuInfo[0].ModelName)
		fmt.Fprintf(output, "CPU Cores: %d\n", cpuInfo[0].Cores)
		fmt.Fprintf(output, "CPU MHz: %.2f\n", cpuInfo[0].Mhz)
	}

	fmt.Fprintf(output, "\n")
	return nil
}

// collectMemoryMetrics collects and writes memory metrics
func collectMemoryMetrics(output *strings.Builder) error {
	fmt.Fprintf(output, "Memory Metrics:\n")
	fmt.Fprintf(output, "---------------\n")

	// Get memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	fmt.Fprintf(output, "Total Memory: %d bytes (%.2f GB)\n", memInfo.Total, float64(memInfo.Total)/(1024*1024*1024))
	fmt.Fprintf(output, "Used Memory: %d bytes (%.2f GB)\n", memInfo.Used, float64(memInfo.Used)/(1024*1024*1024))
	fmt.Fprintf(output, "Free Memory: %d bytes (%.2f GB)\n", memInfo.Free, float64(memInfo.Free)/(1024*1024*1024))
	fmt.Fprintf(output, "Memory Usage: %.2f%%\n", memInfo.UsedPercent)

	fmt.Fprintf(output, "\n")
	return nil
}

// collectDiskMetrics collects and writes disk metrics
func collectDiskMetrics(output *strings.Builder) error {
	fmt.Fprintf(output, "Disk Metrics:\n")
	fmt.Fprintf(output, "-------------\n")

	// Get partitions
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			fmt.Fprintf(output, "Error getting usage for %s: %v\n", partition.Mountpoint, err)
			continue
		}

		fmt.Fprintf(output, "Partition: %s\n", partition.Mountpoint)
		fmt.Fprintf(output, "  Device: %s\n", partition.Device)
		fmt.Fprintf(output, "  Total Space: %d bytes (%.2f GB)\n", usage.Total, float64(usage.Total)/(1024*1024*1024))
		fmt.Fprintf(output, "  Used Space: %d bytes (%.2f GB)\n", usage.Used, float64(usage.Used)/(1024*1024*1024))
		fmt.Fprintf(output, "  Free Space: %d bytes (%.2f GB)\n", usage.Free, float64(usage.Free)/(1024*1024*1024))
		fmt.Fprintf(output, "  Usage: %.2f%%\n", usage.UsedPercent)
		fmt.Fprintf(output, "\n")
	}

	return nil
}

// collectNetworkMetrics collects and writes network metrics
func collectNetworkMetrics(output *strings.Builder) error {
	fmt.Fprintf(output, "Network Metrics:\n")
	fmt.Fprintf(output, "----------------\n")

	// Get network IO counters
	netStats, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	for _, stat := range netStats {
		fmt.Fprintf(output, "Interface: %s\n", stat.Name)
		fmt.Fprintf(output, "  Bytes Sent: %d\n", stat.BytesSent)
		fmt.Fprintf(output, "  Bytes Received: %d\n", stat.BytesRecv)
		fmt.Fprintf(output, "  Packets Sent: %d\n", stat.PacketsSent)
		fmt.Fprintf(output, "  Packets Received: %d\n", stat.PacketsRecv)
		fmt.Fprintf(output, "  Errors In: %d\n", stat.Errin)
		fmt.Fprintf(output, "  Errors Out: %d\n", stat.Errout)
		fmt.Fprintf(output, "\n")
	}

	return nil
}

// checkDiskSpaceHealth checks disk space health
func checkDiskSpaceHealth() (bool, string) {
	var output strings.Builder
	healthy := true

	partitions, err := disk.Partitions(false)
	if err != nil {
		return false, fmt.Sprintf("Error getting disk partitions: %v", err)
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			fmt.Fprintf(&output, "Error checking %s: %v\n", partition.Mountpoint, err)
			healthy = false
			continue
		}

		// Consider unhealthy if disk usage is over 90%
		isPartitionHealthy := usage.UsedPercent < 90.0
		healthy = healthy && isPartitionHealthy

		status := "OK"
		if !isPartitionHealthy {
			status = "WARNING: High disk usage"
		}

		fmt.Fprintf(&output, "%s: %s - %.2f%% used of %.2f GB\n",
			partition.Mountpoint, status, usage.UsedPercent, float64(usage.Total)/(1024*1024*1024))
	}

	return healthy, output.String()
}

// checkCPULoadHealth checks CPU load health
func checkCPULoadHealth() (bool, string) {
	var output strings.Builder

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return false, fmt.Sprintf("Error getting CPU usage: %v", err)
	}

	// Consider unhealthy if CPU usage is over 90% for the sample
	healthy := cpuPercent[0] < 90.0

	status := "OK"
	if !healthy {
		status = "WARNING: High CPU usage"
	}

	fmt.Fprintf(&output, "CPU Usage: %s - %.2f%%\n", status, cpuPercent[0])

	return healthy, output.String()
}

// checkMemoryHealth checks memory health
func checkMemoryHealth() (bool, string) {
	var output strings.Builder

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return false, fmt.Sprintf("Error getting memory info: %v", err)
	}

	// Consider unhealthy if memory usage is over 90%
	healthy := memInfo.UsedPercent < 90.0

	status := "OK"
	if !healthy {
		status = "WARNING: High memory usage"
	}

	fmt.Fprintf(&output, "Memory Usage: %s - %.2f%% (%.2f GB used of %.2f GB total)\n",
		status, memInfo.UsedPercent, float64(memInfo.Used)/(1024*1024*1024), float64(memInfo.Total)/(1024*1024*1024))

	return healthy, output.String()
}

// checkEndpointHealth checks if an HTTP endpoint is healthy
func checkEndpointHealth(endpoint string) (bool, string) {
	var output strings.Builder

	// Add http:// prefix if not present
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make HTTP request
	startTime := time.Now()
	resp, err := client.Get(endpoint)
	duration := time.Since(startTime)

	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v\n", err)
	}
	defer resp.Body.Close()

	// Check status code (200-399 range is considered healthy)
	healthy := resp.StatusCode >= 200 && resp.StatusCode < 400

	status := "OK"
	if !healthy {
		status = fmt.Sprintf("ERROR: HTTP Status %d", resp.StatusCode)
	}

	fmt.Fprintf(&output, "Response: %s - Status %d, Response time: %v\n", status, resp.StatusCode, duration)

	return healthy, output.String()
}
