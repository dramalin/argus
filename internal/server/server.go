package server

import (
	"net/http"
	"net/http/pprof"
	"time"

	"argus/internal/config"
	"argus/internal/handlers"

	"github.com/gin-gonic/gin"
)

// IRoutesRegister is an interface for registering routes (for handlers).
type IRoutesRegister interface {
	RegisterRoutes(*gin.RouterGroup)
}

// setupPprofRoutes sets up pprof debugging routes
func setupPprofRoutes(router *gin.Engine, pprofPath string) {
	// Create a group for pprof routes
	pprofGroup := router.Group(pprofPath)
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}
}

// NewServer sets up the Gin engine, middleware, and routes with production optimizations.
// Accepts configuration, alert/task handlers, and metrics handler, returns the *gin.Engine.
func NewServer(cfg *config.Config, alertsHandler IRoutesRegister, tasksHandler IRoutesRegister, metricsHandler *handlers.MetricsHandler) *gin.Engine {
	// Configure Gin for production or development
	if !cfg.Debug.Enabled {
		gin.SetMode(gin.ReleaseMode)
		// Disable Gin's default console logging in production
		gin.DefaultWriter = nil
	}

	router := gin.New()

	// Middleware stack order is important for performance
	// 1. Recovery middleware (should be first)
	router.Use(gin.Recovery())

	// 2. Security headers (early in the chain)
	router.Use(SecurityHeadersMiddleware())

	// 3. CORS middleware (before any request processing)
	router.Use(CORSMiddleware())

	// 4. Cache control for static assets
	router.Use(CacheControlMiddleware())

	// 5. Compression middleware (before logging to avoid compressing logs)
	if !cfg.Debug.Enabled {
		router.Use(CompressionMiddleware())
	}

	// 6. Logging middleware (last to capture all request details)
	router.Use(LoggingMiddleware())

	// Add pprof endpoints if debug mode is enabled
	if cfg.Debug.Enabled && cfg.Debug.PprofEnabled {
		setupPprofRoutes(router, cfg.Debug.PprofPath)
	}

	// Optimized static file serving with proper caching
	// Use gin.Static with custom file server for better control
	router.Static("/assets", "./web/assets")

	// Serve specific files with optimized handlers
	router.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.File("./web/index.html")
	})

	router.GET("/vite.svg", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400") // 1 day cache
		c.File("./web/vite.svg")
	})

/* 	// Serve index.html for all other routes (SPA fallback) with optimized handler
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip already handled routes
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		if len(path) >= 6 && path[:6] == "/debug" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debug endpoint not found"})
			return
		}

		// Serve SPA fallback with no-cache headers
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File("./web/index.html")
	}) */

	// API routes with optimized grouping
	apiGroup := router.Group("/api")
	{
		// Metrics endpoints using the centralized collector
		metricsGroup := apiGroup.Group("/metrics")
		{
			metricsGroup.GET("/cpu", metricsHandler.GetCPU)
			metricsGroup.GET("/memory", metricsHandler.GetMemory)
			metricsGroup.GET("/network", metricsHandler.GetNetwork)
			metricsGroup.GET("/process", metricsHandler.GetProcess)
			metricsGroup.GET("/health", metricsHandler.GetMetricsHealth)
		}

		// Legacy endpoints for backward compatibility
		apiGroup.GET("/cpu", metricsHandler.GetCPU)
		apiGroup.GET("/memory", metricsHandler.GetMemory)
		apiGroup.GET("/network", metricsHandler.GetNetwork)
		apiGroup.GET("/process", metricsHandler.GetProcess)

		// Other endpoints
		handlers.RegisterHealthRoutes(apiGroup)
		alertsHandler.RegisterRoutes(apiGroup)
		tasksHandler.RegisterRoutes(apiGroup)
	}

	return router
}

// CreateOptimizedHTTPServer creates an HTTP server with production-optimized settings
func CreateOptimizedHTTPServer(handler http.Handler, addr string) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,

		// Production timeouts
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,

		// Optimize for production
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}
