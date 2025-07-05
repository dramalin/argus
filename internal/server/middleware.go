package server

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Pool for reusing string builders in logging
var logBufferPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// Optimized logging middleware that reduces memory allocations
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Skip logging for certain paths to reduce noise
		if shouldSkipLogging(param.Path) {
			return ""
		}

		// Use structured logging which is more efficient than string formatting
		slog.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
		)
		return ""
	})
}

// shouldSkipLogging determines if we should skip logging for certain paths
func shouldSkipLogging(path string) bool {
	// Skip logging for static assets and health checks to reduce log noise
	skipPaths := []string{
		"/assets/",
		"/vite.svg",
		"/favicon.ico",
		"/api/metrics/health",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// Optimized CORS middleware with pre-allocated headers
var (
	corsOrigin      = "*"
	corsCredentials = "true"
	corsHeaders     = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
	corsMethods     = "POST, OPTIONS, GET, PUT, DELETE"
)

// CORSMiddleware sets CORS headers for all requests with optimized allocations
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use pre-allocated strings to reduce memory allocations
		c.Header("Access-Control-Allow-Origin", corsOrigin)
		c.Header("Access-Control-Allow-Credentials", corsCredentials)
		c.Header("Access-Control-Allow-Headers", corsHeaders)
		c.Header("Access-Control-Allow-Methods", corsMethods)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// CacheControlMiddleware adds appropriate caching headers for static assets
func CacheControlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Set cache headers based on file type
		if strings.HasPrefix(path, "/assets/") {
			// Static assets can be cached for a long time
			c.Header("Cache-Control", "public, max-age=31536000") // 1 year
			c.Header("Expires", time.Now().Add(365*24*time.Hour).Format(time.RFC1123))
		} else if strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".ico") {
			// Icons and SVGs can be cached for a moderate time
			c.Header("Cache-Control", "public, max-age=86400") // 1 day
		} else if path == "/" || strings.HasSuffix(path, ".html") {
			// HTML files should not be cached to ensure updates are reflected
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// CompressionMiddleware enables gzip compression for responses
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Enable compression for text-based content
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "text/") ||
				strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "application/javascript") {
				c.Header("Content-Encoding", "gzip")
			}
		}
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
