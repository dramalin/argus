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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware
	router.Use(LoggingMiddleware())
	router.Use(CORSMiddleware())
	router.Use(gin.Recovery())

	// Static files
	router.Static("/static", "./web/static")
	router.StaticFile("/", "./web/static/index.html")
	router.StaticFile("/app.js", "./web/static/js/app.js")
	router.StaticFile("/alerts.js", "./web/static/js/alerts.js")
	router.StaticFile("/alert-status.js", "./web/static/js/alert-status.js")
	router.StaticFile("/shared.js", "./web/static/js/shared.js")
	router.StaticFile("/css/main.css", "./web/static/css/main.css")

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
