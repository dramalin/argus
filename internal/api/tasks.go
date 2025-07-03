// Package api provides HTTP API handlers for the Argus Task Management System
package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"argus/internal/tasks"
	"argus/internal/tasks/repository"
)

// TasksHandler manages task-related API endpoints
type TasksHandler struct {
	repo      repository.TaskRepository
	scheduler *tasks.TaskScheduler
}

// NewTasksHandler creates a new tasks API handler
func NewTasksHandler(repo repository.TaskRepository, scheduler *tasks.TaskScheduler) *TasksHandler {
	return &TasksHandler{
		repo:      repo,
		scheduler: scheduler,
	}
}

// RegisterRoutes registers all task-related routes to the given router group
func (h *TasksHandler) RegisterRoutes(router *gin.RouterGroup) {
	tasks := router.Group("/tasks")
	{
		tasks.GET("", h.ListTasks)
		tasks.GET("/:id", h.GetTask)
		tasks.POST("", h.CreateTask)
		tasks.PUT("/:id", h.UpdateTask)
		tasks.DELETE("/:id", h.DeleteTask)
		tasks.GET("/:id/executions", h.GetTaskExecutions)
		tasks.POST("/:id/run", h.RunTaskNow)
	}
}

// ListTasks returns all task configurations
func (h *TasksHandler) ListTasks(c *gin.Context) {
	slog.Debug("Fetching all task configurations")

	tasksList, err := h.repo.ListTasks(c.Request.Context())
	if err != nil {
		slog.Error("Failed to list tasks", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks: " + err.Error()})
		return
	}

	slog.Debug("Task configurations retrieved successfully", "count", len(tasksList))
	c.JSON(http.StatusOK, tasksList)
}

// GetTask returns a specific task configuration by ID
func (h *TasksHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Fetching task configuration", "id", id)

	task, err := h.repo.GetTask(c.Request.Context(), id)
	if err != nil {
		slog.Debug("Task not found", "id", id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	slog.Debug("Task configuration retrieved successfully", "id", id)
	c.JSON(http.StatusOK, task)
}

// CreateTask creates a new task configuration
func (h *TasksHandler) CreateTask(c *gin.Context) {
	var task tasks.TaskConfig
	if err := c.ShouldBindJSON(&task); err != nil {
		slog.Debug("Invalid task configuration data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task configuration: " + err.Error()})
		return
	}

	// Generate a new UUID if ID is empty
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	// Validate the task configuration
	if err := task.Validate(); err != nil {
		slog.Debug("Invalid task configuration", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task configuration: " + err.Error()})
		return
	}

	// Store the task
	if err := h.repo.CreateTask(c.Request.Context(), &task); err != nil {
		slog.Error("Failed to create task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task: " + err.Error()})
		return
	}

	slog.Info("Task created successfully", "id", task.ID, "name", task.Name, "type", task.Type)
	c.JSON(http.StatusCreated, task)
}

// UpdateTask updates an existing task configuration
func (h *TasksHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Updating task configuration", "id", id)

	// Check if task exists
	existing, err := h.repo.GetTask(c.Request.Context(), id)
	if err != nil {
		slog.Debug("Task not found for update", "id", id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Parse update data
	var task tasks.TaskConfig
	if err := c.ShouldBindJSON(&task); err != nil {
		slog.Debug("Invalid task update data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task configuration: " + err.Error()})
		return
	}

	// Ensure ID matches and preserve creation timestamp
	task.ID = id
	task.CreatedAt = existing.CreatedAt
	task.UpdatedAt = time.Now()

	// Validate the task configuration
	if err := task.Validate(); err != nil {
		slog.Debug("Invalid task update", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task configuration: " + err.Error()})
		return
	}

	// Update the task
	if err := h.repo.UpdateTask(c.Request.Context(), &task); err != nil {
		slog.Error("Failed to update task", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task: " + err.Error()})
		return
	}

	slog.Info("Task updated successfully", "id", id, "name", task.Name)
	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a task configuration
func (h *TasksHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Deleting task configuration", "id", id)

	// Delete the task
	if err := h.repo.DeleteTask(c.Request.Context(), id); err != nil {
		slog.Debug("Task not found for deletion", "id", id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	slog.Info("Task deleted successfully", "id", id)
	c.Status(http.StatusNoContent)
}

// GetTaskExecutions retrieves execution records for a specific task
func (h *TasksHandler) GetTaskExecutions(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Fetching task executions", "id", id)

	// Check if task exists
	_, err := h.repo.GetTask(c.Request.Context(), id)
	if err != nil {
		slog.Debug("Task not found for execution history", "id", id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Get limit parameter with default
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	executions, err := h.repo.GetTaskExecutions(c.Request.Context(), id, limit)
	if err != nil {
		slog.Error("Failed to get task executions", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get task executions: %v", err)})
		return
	}

	slog.Debug("Task executions retrieved successfully", "id", id, "count", len(executions))
	c.JSON(http.StatusOK, executions)
}

// RunTaskNow executes a task immediately
func (h *TasksHandler) RunTaskNow(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Running task immediately", "id", id)

	// Check if task exists
	_, err := h.repo.GetTask(c.Request.Context(), id)
	if err != nil {
		slog.Debug("Task not found for execution", "id", id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	execution, err := h.scheduler.RunTaskNow(id)
	if err != nil {
		slog.Error("Failed to run task", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run task: " + err.Error()})
		return
	}

	slog.Info("Task executed successfully", "id", id, "execution_id", execution.ID, "status", execution.Status)
	c.JSON(http.StatusOK, execution)
}
