package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"argus/internal/models"
)

// MockTaskRepository is a mock implementation of models.TaskRepository
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, task *models.TaskConfig) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetTask(ctx context.Context, id string) (*models.TaskConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaskConfig), args.Error(1)
}

func (m *MockTaskRepository) UpdateTask(ctx context.Context, task *models.TaskConfig) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) ListTasks(ctx context.Context) ([]*models.TaskConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaskConfig), args.Error(1)
}

func (m *MockTaskRepository) GetTasksByType(ctx context.Context, taskType models.TaskType) ([]*models.TaskConfig, error) {
	args := m.Called(ctx, taskType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaskConfig), args.Error(1)
}

func (m *MockTaskRepository) RecordExecution(ctx context.Context, execution *models.TaskExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockTaskRepository) GetTaskExecutions(ctx context.Context, taskID string, limit int) ([]*models.TaskExecution, error) {
	args := m.Called(ctx, taskID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaskExecution), args.Error(1)
}

func (m *MockTaskRepository) GetExecution(ctx context.Context, id string) (*models.TaskExecution, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaskExecution), args.Error(1)
}

func (m *MockTaskRepository) GetExecutions(ctx context.Context, taskID string) ([]*models.TaskExecution, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaskExecution), args.Error(1)
}

// MockTaskScheduler is a mock implementation of TaskScheduler
type MockTaskScheduler struct {
	mock.Mock
}

func (m *MockTaskScheduler) ScheduleTask(ctx context.Context, task *models.TaskConfig) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskScheduler) CancelTask(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskScheduler) ExecuteTask(ctx context.Context, id string) (*models.TaskExecution, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaskExecution), args.Error(1)
}

func (m *MockTaskScheduler) RunTaskNow(taskID string) (*models.TaskExecution, error) {
	args := m.Called(taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaskExecution), args.Error(1)
}

func (m *MockTaskScheduler) Start() error {
	m.Called()
	return nil
}

func (m *MockTaskScheduler) Stop() {
	m.Called()
}

func setupTasksTest() (*gin.Engine, *MockTaskRepository, *MockTaskScheduler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockRepo := new(MockTaskRepository)
	mockScheduler := new(MockTaskScheduler)

	handler := NewTasksHandler(mockRepo, mockScheduler)

	tasksGroup := r.Group("/api")
	handler.RegisterRoutes(tasksGroup)

	return r, mockRepo, mockScheduler
}

func TestListTasks(t *testing.T) {
	r, mockRepo, _ := setupTasksTest()

	now := time.Now()

	// Setup mock
	tasks := []*models.TaskConfig{
		{
			ID:          "1",
			Name:        "Task 1",
			Description: "Description 1",
			Type:        models.TaskLogRotation,
			Enabled:     true,
			Schedule: models.Schedule{
				CronExpression: "* * * * *",
				NextRunTime:    now,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "2",
			Name:        "Task 2",
			Description: "Description 2",
			Type:        models.TaskHealthCheck,
			Enabled:     true,
			Schedule: models.Schedule{
				CronExpression: "0 * * * *",
				NextRunTime:    now,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockRepo.On("ListTasks", mock.Anything).Return(tasks, nil)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tasks", nil)
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []*models.TaskConfig
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 2)
	assert.Equal(t, "Task 1", response[0].Name)
	assert.Equal(t, "Task 2", response[1].Name)

	mockRepo.AssertExpectations(t)
}

func TestGetTask(t *testing.T) {
	r, mockRepo, _ := setupTasksTest()

	now := time.Now()
	taskID := "1"
	task := &models.TaskConfig{
		ID:          taskID,
		Name:        "Test Task",
		Description: "Test Description",
		Type:        models.TaskLogRotation,
		Enabled:     true,
		Schedule: models.Schedule{
			CronExpression: "* * * * *",
			NextRunTime:    now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo.On("GetTask", mock.Anything, taskID).Return(task, nil)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tasks/"+taskID, nil)
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response *models.TaskConfig
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, taskID, response.ID)
	assert.Equal(t, "Test Task", response.Name)

	mockRepo.AssertExpectations(t)
}

func TestGetTaskNotFound(t *testing.T) {
	r, mockRepo, _ := setupTasksTest()

	taskID := "nonexistent"
	mockRepo.On("GetTask", mock.Anything, taskID).Return(nil, errors.New("task not found"))

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tasks/"+taskID, nil)
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockRepo.AssertExpectations(t)
}
