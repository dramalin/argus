package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes registers the health check endpoint.
func RegisterHealthRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", HealthHandler)
}

// HealthHandler responds with a simple health status.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
} 