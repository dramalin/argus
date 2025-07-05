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

	"argus/internal/config"
	"argus/internal/database"
	"argus/internal/handlers"
	"argus/internal/metrics"
	"argus/internal/models"
	"argus/internal/server"
	"argus/internal/services"
)

// setupLogger configures structured logging
func setupLogger() {
	// Set up structured logging with JSON format for production
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Only log API requests and errors
		if param.StatusCode >= 400 || (len(param.Path) >= 4 && param.Path[:4] == "/api") {
			return fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %#v\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
				param.ErrorMessage,
			)
		}
		return ""
	})
}

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
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

	// Load configuration (with minimal logging)
	cfgPath := "config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfgPath = "config.example.yaml"
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize metrics collector
	metricsConfig := metrics.DefaultConfig()
	// Override with configuration if available
	if cfg.Monitoring.UpdateInterval != "" {
		if interval, err := time.ParseDuration(cfg.Monitoring.UpdateInterval); err == nil {
			metricsConfig.UpdateInterval = interval
		}
	}
	if cfg.Monitoring.ProcessLimit > 0 {
		metricsConfig.ProcessLimit = cfg.Monitoring.ProcessLimit
	}

	metricsCollector := metrics.NewCollector(metricsConfig)

	// Create a context for the metrics collector
	metricsCtx, metricsCancel := context.WithCancel(context.Background())
	defer metricsCancel()

	// Start the metrics collector
	if err := metricsCollector.Start(metricsCtx); err != nil {
		slog.Error("Failed to start metrics collector", "error", err)
		os.Exit(1)
	}
	slog.Info("Metrics collector started successfully")

	// Initialize alert storage
	alertStore, err := database.NewAlertStore(cfg.Alerts.StoragePath)
	if err != nil {
		slog.Error("Failed to initialize alert storage", "error", err)
		os.Exit(1)
	}

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
		emailChannel := services.NewEmailChannel(emailConfig, notifierConfig)
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
	metricsHandler := handlers.NewMetricsHandler(metricsCollector)

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
	router := server.NewServer(cfg, alertsHandler, tasksHandler, metricsHandler)
	// Add WebSocket route
	router.GET("/ws", server.WebSocketHandler)

	slog.Info("API routes and static file serving configured via server package")

	// Create optimized HTTP server with production settings
	srv := server.CreateOptimizedHTTPServer(router, fmt.Sprintf(":%d", cfg.Server.Port))

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

	// Cancel the metrics collector context to stop it
	metricsCancel()
	metricsCollector.Stop()

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
