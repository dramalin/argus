package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"

	"argus/internal/config"
	"argus/internal/database"
	"argus/internal/handlers"
	"argus/internal/models"
	"argus/internal/server"
	"argus/internal/services"
)

// setupLogger configures structured logging
func setupLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		slog.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)
		return ""
	})
}

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getCPU(c *gin.Context) {
	slog.Debug("Fetching CPU metrics")

	loadAvg, err := load.Avg()
	if err != nil {
		slog.Error("Failed to get load average", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get load average: " + err.Error()})
		return
	}

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		slog.Error("Failed to get CPU usage", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get CPU usage: " + err.Error()})
		return
	}

	usage := 0.0
	if len(cpuPercent) > 0 {
		usage = cpuPercent[0]
	}

	slog.Debug("CPU metrics retrieved successfully",
		"load1", loadAvg.Load1,
		"load5", loadAvg.Load5,
		"load15", loadAvg.Load15,
		"usage_percent", usage)

	c.JSON(http.StatusOK, gin.H{
		"load1":         loadAvg.Load1,
		"load5":         loadAvg.Load5,
		"load15":        loadAvg.Load15,
		"usage_percent": usage,
	})
}

func getMemory(c *gin.Context) {
	slog.Debug("Fetching memory metrics")

	vm, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("Failed to get memory info", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get memory info: " + err.Error()})
		return
	}

	slog.Debug("Memory metrics retrieved successfully",
		"total", vm.Total,
		"used", vm.Used,
		"free", vm.Free,
		"used_percent", vm.UsedPercent)

	c.JSON(http.StatusOK, gin.H{
		"total":        vm.Total,
		"used":         vm.Used,
		"free":         vm.Free,
		"used_percent": vm.UsedPercent,
	})
}

func getNetwork(c *gin.Context) {
	slog.Debug("Fetching network metrics")

	ioCounters, err := net.IOCounters(false)
	if err != nil {
		slog.Error("Failed to get network stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get network stats: " + err.Error()})
		return
	}

	if len(ioCounters) == 0 {
		slog.Warn("No network interfaces found")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No network interfaces found"})
		return
	}

	io := ioCounters[0]

	slog.Debug("Network metrics retrieved successfully",
		"bytes_sent", io.BytesSent,
		"bytes_recv", io.BytesRecv,
		"packets_sent", io.PacketsSent,
		"packets_recv", io.PacketsRecv)

	c.JSON(http.StatusOK, gin.H{
		"bytes_sent":   io.BytesSent,
		"bytes_recv":   io.BytesRecv,
		"packets_sent": io.PacketsSent,
		"packets_recv": io.PacketsRecv,
	})
}

func getProcess(c *gin.Context) {
	slog.Debug("Fetching process metrics")

	procs, err := process.Processes()
	if err != nil {
		slog.Error("Failed to get process list", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get process list: " + err.Error()})
		return
	}

	result := []gin.H{}
	count := 0

	// Limit to top 20 processes to avoid overwhelming the frontend
	for _, p := range procs {
		if count >= 20 {
			break
		}

		name, err := p.Name()
		if err != nil {
			continue
		}

		cpuP, err := p.CPUPercent()
		if err != nil {
			cpuP = 0.0
		}

		memP, err := p.MemoryPercent()
		if err != nil {
			memP = 0.0
		}

		result = append(result, gin.H{
			"pid":         p.Pid,
			"name":        name,
			"cpu_percent": cpuP,
			"mem_percent": memP,
		})
		count++
	}

	slog.Debug("Process metrics retrieved successfully", "process_count", len(result))

	c.JSON(http.StatusOK, result)
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
		slog.Warn("Invalid integer value for environment variable", "key", key, "value", value, "using_default", defaultVal)
	}
	return defaultVal
}

func main() {
	// Setup structured logging
	setupLogger()

	slog.Info("Starting Argus System Monitor")

	// Load configuration
	cfgPath := "config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfgPath = "config.example.yaml"
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("Configuration loaded successfully", "config_file", cfgPath)

	// Initialize alert storage
	alertStore, err := database.NewAlertStore(cfg.Alerts.StoragePath)
	if err != nil {
		slog.Error("Failed to initialize alert storage", "error", err)
		os.Exit(1)
	}
	slog.Info("Alert storage initialized successfully", "storage_path", cfg.Alerts.StoragePath)

	// Initialize alert evaluator
	evalConfig := services.DefaultEvaluatorConfig()
	alertEvaluator := services.NewEvaluator(alertStore, evalConfig)

	// Create a context for the evaluator
	evalCtx, evalCancel := context.WithCancel(context.Background())
	defer evalCancel()

	// Start the evaluator
	if err := alertEvaluator.Start(evalCtx); err != nil {
		slog.Error("Failed to start alert evaluator", "error", err)
		os.Exit(1)
	}
	slog.Info("Alert evaluator started successfully")

	// Initialize notification system
	notifierConfig := services.DefaultConfig()
	alertNotifier := services.NewNotifier(notifierConfig)

	// Register notification channels
	inAppChannel := services.NewInAppChannel(100) // Store up to 100 notifications
	alertNotifier.RegisterChannel(inAppChannel)

	// Register email notification if configured
	if os.Getenv("SMTP_HOST") != "" {
		emailConfig := &services.EmailConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     getEnvAsInt("SMTP_PORT", 587), // Convert string to int with default value
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		}
		emailChannel := services.NewEmailChannel(emailConfig)
		alertNotifier.RegisterChannel(emailChannel)
		slog.Info("Email notification channel registered successfully")
	}

	// Connect evaluator events to notifier
	go func() {
		for event := range alertEvaluator.Events() {
			alertNotifier.ProcessEvent(event)
		}
	}()
	slog.Info("Alert notification system initialized successfully")

	// Create API handlers
	alertsHandler := handlers.NewAlertsHandler(alertStore, alertEvaluator, alertNotifier)

	// Initialize task repository and scheduler
	taskRepo, err := database.NewFileTaskRepository(cfg.Tasks.StoragePath)
	if err != nil {
		slog.Error("Failed to initialize task repository", "error", err)
		os.Exit(1)
	}
	slog.Info("Task repository initialized successfully")
	taskScheduler := services.NewTaskScheduler(taskRepo, nil)

	// Register all task runners
	runners := []services.TaskRunner{}
	runnerTypes := []models.TaskType{
		models.TaskLogRotation,
		models.TaskMetricsAggregation,
		models.TaskHealthCheck,
		models.TaskSystemCleanup,
	}
	for _, t := range runnerTypes {
		runner, err := services.NewTaskRunner(t)
		if err != nil {
			slog.Error("Failed to create task runner", "type", t, "error", err)
			continue
		}
		taskScheduler.RegisterRunner(runner)
		runners = append(runners, runner)
	}

	if err := taskScheduler.Start(); err != nil {
		slog.Error("Failed to start task scheduler", "error", err)
		os.Exit(1)
	}
	slog.Info("Task scheduler started successfully")

	// Create tasks API handler
	tasksHandler := handlers.NewTasksHandler(taskRepo, taskScheduler)

	// --- Use the new server package for all server setup ---
	router := server.NewServer(cfg, alertsHandler, tasksHandler, getCPU, getMemory, getNetwork, getProcess)
	// Add WebSocket route
	router.GET("/ws", server.WebSocketHandler)

	slog.Info("API routes and static file serving configured via server package")

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting HTTP server", "address", srv.Addr, "url", fmt.Sprintf("http://%s%s", cfg.Server.Host, srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Cancel the evaluator context to stop it
	evalCancel()

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	// On shutdown, stop the scheduler
	taskScheduler.Stop()

	slog.Info("Server shutdown completed successfully")
}
