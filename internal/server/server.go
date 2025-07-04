package server

import (
	"argus/internal/config"
	"argus/internal/handlers"

	"github.com/gin-gonic/gin"
)

// IRoutesRegister is an interface for registering routes (for handlers).
type IRoutesRegister interface {
	RegisterRoutes(*gin.RouterGroup)
}

// NewServer sets up the Gin engine, middleware, and routes.
// Accepts configuration, alert/task handlers, and returns the *gin.Engine.
func NewServer(cfg *config.Config, alertsHandler IRoutesRegister, tasksHandler IRoutesRegister, getCPU, getMemory, getNetwork, getProcess gin.HandlerFunc) *gin.Engine {
	// Set Gin to release mode and disable default console logging
	// gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(LoggingMiddleware()) // Using our custom selective logging middleware
	router.Use(CORSMiddleware())
	router.Use(gin.Recovery())

	// Serve static assets from the release folder
	router.Static("/assets", "./web/assets")

	// Explicitly serve specific files from root
	router.StaticFile("/", "./web/index.html")
	router.StaticFile("/vite.svg", "./web/vite.svg")

	// Serve index.html for all other routes (SPA fallback)
	router.NoRoute(func(c *gin.Context) {
		// Skip already handled routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			return
		}
		c.File("./web/index.html")
	})

	// API routes
	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/cpu", getCPU)
		apiGroup.GET("/memory", getMemory)
		apiGroup.GET("/network", getNetwork)
		apiGroup.GET("/process", getProcess)
		handlers.RegisterHealthRoutes(apiGroup)
		alertsHandler.RegisterRoutes(apiGroup)
		tasksHandler.RegisterRoutes(apiGroup)
	}

	return router
}
