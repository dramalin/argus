package benchmarks

import (
	"net/http/httptest"
	"testing"

	"argus/internal/config"
	"argus/internal/handlers"
	"argus/internal/metrics"
	"argus/internal/server"

	"github.com/gin-gonic/gin"
)

// BenchmarkMiddlewareStack benchmarks the complete middleware stack
func BenchmarkMiddlewareStack(b *testing.B) {
	cfg := &config.Config{}
	cfg.Debug.Enabled = false // Production mode

	// Create minimal handlers for testing
	metricsCollector := metrics.NewCollector(metrics.DefaultConfig())
	metricsHandler := handlers.NewMetricsHandler(metricsCollector)

	// Create mock handlers
	alertsHandler := &mockRoutesRegister{}
	tasksHandler := &mockRoutesRegister{}

	router := server.NewServer(cfg, alertsHandler, tasksHandler, metricsHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/metrics/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkStaticFileServing benchmarks static file serving performance
func BenchmarkStaticFileServing(b *testing.B) {
	cfg := &config.Config{}
	cfg.Debug.Enabled = false

	metricsCollector := metrics.NewCollector(metrics.DefaultConfig())
	metricsHandler := handlers.NewMetricsHandler(metricsCollector)
	alertsHandler := &mockRoutesRegister{}
	tasksHandler := &mockRoutesRegister{}

	router := server.NewServer(cfg, alertsHandler, tasksHandler, metricsHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/vite.svg", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkMiddlewareOverhead benchmarks middleware overhead through HTTP requests
func BenchmarkMiddlewareOverhead(b *testing.B) {
	cfg := &config.Config{}
	cfg.Debug.Enabled = false

	metricsCollector := metrics.NewCollector(metrics.DefaultConfig())
	metricsHandler := handlers.NewMetricsHandler(metricsCollector)
	alertsHandler := &mockRoutesRegister{}
	tasksHandler := &mockRoutesRegister{}

	router := server.NewServer(cfg, alertsHandler, tasksHandler, metricsHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("OPTIONS", "/api/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIEndpoint benchmarks API endpoint performance
func BenchmarkAPIEndpoint(b *testing.B) {
	cfg := &config.Config{}
	cfg.Debug.Enabled = false

	metricsCollector := metrics.NewCollector(metrics.DefaultConfig())
	metricsHandler := handlers.NewMetricsHandler(metricsCollector)
	alertsHandler := &mockRoutesRegister{}
	tasksHandler := &mockRoutesRegister{}

	router := server.NewServer(cfg, alertsHandler, tasksHandler, metricsHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/cpu", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	cfg := &config.Config{}
	cfg.Debug.Enabled = false

	metricsCollector := metrics.NewCollector(metrics.DefaultConfig())
	metricsHandler := handlers.NewMetricsHandler(metricsCollector)
	alertsHandler := &mockRoutesRegister{}
	tasksHandler := &mockRoutesRegister{}

	router := server.NewServer(cfg, alertsHandler, tasksHandler, metricsHandler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/api/metrics/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// Mock implementations for testing

type mockRoutesRegister struct{}

func (m *mockRoutesRegister) RegisterRoutes(group *gin.RouterGroup) {
	// Mock implementation - no routes needed for benchmarks
}
