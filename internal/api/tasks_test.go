package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"argus/internal/models"
)

// TestTasksAPIBasic tests the basic functionality of the tasks API
func TestTasksAPIBasic(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()
	router.Use(gin.Recovery())

	// Create a minimal test handler
	handler := &TasksHandler{}

	// Register routes
	apiGroup := router.Group("/api")
	handler.RegisterRoutes(apiGroup)

	// Test that the router has registered the task routes
	routes := router.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path] = true
	}

	// Verify that expected routes exist
	expectedPaths := []string{
		"/api/tasks",
		"/api/tasks/:id",
		"/api/tasks/:id/executions",
		"/api/tasks/:id/run",
	}

	for _, path := range expectedPaths {
		assert.True(t, routePaths[path], "Route %s should be registered", path)
	}
}

// TestTasksEndpointsReturnsStatusCodes tests that the task endpoints return the expected status codes
func TestTasksEndpointsReturnsStatusCodes(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a minimal test task
	task := &models.TaskConfig{
		ID:          "test-task-id",
		Name:        "Test Task",
		Description: "A test task",
		Type:        models.TaskLogRotation,
		Enabled:     true,
		Schedule: models.Schedule{
			CronExpression: "0 0 * * *",
		},
	}

	// Create task JSON
	taskJSON, err := json.Marshal(task)
	require.NoError(t, err)

	// Define test cases
	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "POST /tasks returns status code",
			method:         http.MethodPost,
			path:           "/api/tasks",
			body:           taskJSON,
			expectedStatus: http.StatusInternalServerError, // Will fail without a real repo
		},
		{
			name:           "GET /tasks returns status code",
			method:         http.MethodGet,
			path:           "/api/tasks",
			expectedStatus: http.StatusInternalServerError, // Will fail without a real repo
		},
		{
			name:           "GET /tasks/:id returns status code",
			method:         http.MethodGet,
			path:           "/api/tasks/test-task-id",
			expectedStatus: http.StatusInternalServerError, // Will fail without a real repo
		},
		{
			name:           "PUT /tasks/:id returns status code",
			method:         http.MethodPut,
			path:           "/api/tasks/test-task-id",
			body:           taskJSON,
			expectedStatus: http.StatusInternalServerError, // Will fail without a real repo
		},
		{
			name:           "DELETE /tasks/:id returns status code",
			method:         http.MethodDelete,
			path:           "/api/tasks/test-task-id",
			expectedStatus: http.StatusInternalServerError, // Will fail without a real repo
		},
		{
			name:           "GET /tasks/:id/executions returns status code",
			method:         http.MethodGet,
			path:           "/api/tasks/test-task-id/executions",
			expectedStatus: http.StatusInternalServerError, // Will fail without a real repo
		},
		{
			name:           "POST /tasks/:id/run returns status code",
			method:         http.MethodPost,
			path:           "/api/tasks/test-task-id/run",
			expectedStatus: http.StatusInternalServerError, // Will fail without a real repo
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new router and handler for each test
			router := gin.New()
			router.Use(gin.Recovery())

			// Create a minimal test handler
			handler := &TasksHandler{}

			// Register routes
			apiGroup := router.Group("/api")
			handler.RegisterRoutes(apiGroup)

			// Create request
			var req *http.Request
			if tc.body != nil {
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			// Create recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert status
			assert.Equal(t, tc.expectedStatus, w.Code, "Expected status %d for %s %s, got %d",
				tc.expectedStatus, tc.method, tc.path, w.Code)
		})
	}
}
