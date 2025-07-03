// Package repository provides functionality for persisting and retrieving task configurations
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"argus/internal/tasks"
)

const (
	// DefaultConfigDir is the default directory for storing task configurations
	DefaultConfigDir = ".argus/tasks"

	// TasksDir is the subdirectory for storing task configurations
	TasksDir = "configurations"

	// ExecutionsDir is the subdirectory for storing task execution records
	ExecutionsDir = "executions"

	// DefaultFileMode is the default file permission mode for configuration files
	DefaultFileMode = 0644

	// DefaultDirMode is the default directory permission mode
	DefaultDirMode = 0755
)

var (
	// ErrTaskNotFound is returned when a task configuration is not found
	ErrTaskNotFound = errors.New("task configuration not found")

	// ErrExecutionNotFound is returned when a task execution record is not found
	ErrExecutionNotFound = errors.New("task execution record not found")

	// ErrInvalidTaskID is returned when a task ID is invalid
	ErrInvalidTaskID = errors.New("invalid task ID")

	// ErrDirectoryCreation is returned when a directory cannot be created
	ErrDirectoryCreation = errors.New("failed to create directory")

	// ErrFileLocked is returned when a file is locked for writing
	ErrFileLocked = errors.New("file is locked for writing")
)

// TaskRepository defines the interface for task storage operations
type TaskRepository interface {
	// CreateTask saves a new task configuration
	CreateTask(ctx context.Context, task *tasks.TaskConfig) error

	// GetTask retrieves a task configuration by ID
	GetTask(ctx context.Context, id string) (*tasks.TaskConfig, error)

	// UpdateTask updates an existing task configuration
	UpdateTask(ctx context.Context, task *tasks.TaskConfig) error

	// DeleteTask removes a task configuration
	DeleteTask(ctx context.Context, id string) error

	// ListTasks retrieves all task configurations
	ListTasks(ctx context.Context) ([]*tasks.TaskConfig, error)

	// GetTasksByType retrieves task configurations of a specific type
	GetTasksByType(ctx context.Context, taskType tasks.TaskType) ([]*tasks.TaskConfig, error)

	// RecordExecution saves a task execution record
	RecordExecution(ctx context.Context, execution *tasks.TaskExecution) error

	// GetTaskExecutions retrieves execution records for a specific task
	GetTaskExecutions(ctx context.Context, taskID string, limit int) ([]*tasks.TaskExecution, error)

	// GetExecution retrieves a specific execution record by ID
	GetExecution(ctx context.Context, id string) (*tasks.TaskExecution, error)

	// GetExecutions retrieves all executions for a task
	GetExecutions(ctx context.Context, taskID string) ([]*tasks.TaskExecution, error)
}

// FileTaskRepository implements TaskRepository using the filesystem
type FileTaskRepository struct {
	baseDir       string
	tasksDir      string
	executionsDir string
	mutex         sync.RWMutex
	fileLocks     map[string]*sync.Mutex
	lockMu        sync.Mutex
}

// NewFileTaskRepository creates a new FileTaskRepository with the given base directory
func NewFileTaskRepository(baseDir string) (*FileTaskRepository, error) {
	if baseDir == "" {
		baseDir = DefaultConfigDir
	}

	tasksDir := filepath.Join(baseDir, TasksDir)
	executionsDir := filepath.Join(baseDir, ExecutionsDir)

	// Create directories if they don't exist
	if err := os.MkdirAll(tasksDir, DefaultDirMode); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrDirectoryCreation, tasksDir, err)
	}

	if err := os.MkdirAll(executionsDir, DefaultDirMode); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrDirectoryCreation, executionsDir, err)
	}

	return &FileTaskRepository{
		baseDir:       baseDir,
		tasksDir:      tasksDir,
		executionsDir: executionsDir,
		fileLocks:     make(map[string]*sync.Mutex),
	}, nil
}

// getFileLock returns a mutex for the given file path, creating one if it doesn't exist
func (r *FileTaskRepository) getFileLock(path string) *sync.Mutex {
	r.lockMu.Lock()
	defer r.lockMu.Unlock()

	if lock, exists := r.fileLocks[path]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	r.fileLocks[path] = lock
	return lock
}

// taskFilePath returns the file path for the given task ID
func (r *FileTaskRepository) taskFilePath(id string) string {
	return filepath.Join(r.tasksDir, fmt.Sprintf("%s.json", id))
}

// executionFilePath returns the file path for the given execution ID
func (r *FileTaskRepository) executionFilePath(id string) string {
	return filepath.Join(r.executionsDir, fmt.Sprintf("%s.json", id))
}

// taskExecutionsDir returns the directory for a task's execution records
func (r *FileTaskRepository) taskExecutionsDir(taskID string) string {
	return filepath.Join(r.executionsDir, taskID)
}

// CreateTask saves a new task configuration
func (r *FileTaskRepository) CreateTask(ctx context.Context, task *tasks.TaskConfig) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	if task.ID == "" {
		task.ID = tasks.GenerateID()
	}

	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	task.UpdatedAt = time.Now()

	// Validate the task configuration
	if err := task.Validate(); err != nil {
		return err
	}

	// Lock for writing
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if task already exists
	filePath := r.taskFilePath(task.ID)
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check if task exists: %w", err)
	}

	// Write to file atomically
	return r.writeTaskToFile(task, filePath)
}

// GetTask retrieves a task configuration by ID
func (r *FileTaskRepository) GetTask(ctx context.Context, id string) (*tasks.TaskConfig, error) {
	if id == "" {
		return nil, ErrInvalidTaskID
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	filePath := r.taskFilePath(id)
	task, err := r.readTaskFromFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to read task: %w", err)
	}

	return task, nil
}

// UpdateTask updates an existing task configuration
func (r *FileTaskRepository) UpdateTask(ctx context.Context, task *tasks.TaskConfig) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	if task.ID == "" {
		return ErrInvalidTaskID
	}

	// Validate the task configuration
	if err := task.Validate(); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if task exists
	filePath := r.taskFilePath(task.ID)
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("failed to check if task exists: %w", err)
	}

	task.UpdatedAt = time.Now()

	// Write to file atomically
	return r.writeTaskToFile(task, filePath)
}

// DeleteTask removes a task configuration
func (r *FileTaskRepository) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidTaskID
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	filePath := r.taskFilePath(id)

	// Check if task exists
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("failed to check if task exists: %w", err)
	}

	// Get file lock
	lock := r.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// Remove the task file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// ListTasks retrieves all task configurations
func (r *FileTaskRepository) ListTasks(ctx context.Context) ([]*tasks.TaskConfig, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var tasksList []*tasks.TaskConfig

	// Read all files in the tasks directory
	files, err := filepath.Glob(filepath.Join(r.tasksDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list task files: %w", err)
	}

	for _, file := range files {
		task, err := r.readTaskFromFile(file)
		if err != nil {
			// Log error but continue with other tasks
			continue
		}
		tasksList = append(tasksList, task)
	}

	// Sort tasks by creation time (newest first)
	sort.Slice(tasksList, func(i, j int) bool {
		return tasksList[i].CreatedAt.After(tasksList[j].CreatedAt)
	})

	return tasksList, nil
}

// GetTasksByType retrieves task configurations of a specific type
func (r *FileTaskRepository) GetTasksByType(ctx context.Context, taskType tasks.TaskType) ([]*tasks.TaskConfig, error) {
	allTasks, err := r.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	var tasksOfType []*tasks.TaskConfig
	for _, task := range allTasks {
		if task.Type == taskType {
			tasksOfType = append(tasksOfType, task)
		}
	}

	return tasksOfType, nil
}

// RecordExecution saves a task execution record
func (r *FileTaskRepository) RecordExecution(ctx context.Context, execution *tasks.TaskExecution) error {
	if execution == nil {
		return errors.New("execution cannot be nil")
	}

	if execution.ID == "" {
		execution.ID = tasks.GenerateID()
	}

	if execution.TaskID == "" {
		return errors.New("task ID is required for execution record")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Ensure task executions directory exists
	taskExecDir := r.taskExecutionsDir(execution.TaskID)
	if err := os.MkdirAll(taskExecDir, DefaultDirMode); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrDirectoryCreation, taskExecDir, err)
	}

	// Create execution file path
	filePath := r.executionFilePath(execution.ID)

	// Write to file atomically
	return r.writeExecutionToFile(execution, filePath)
}

// GetTaskExecutions retrieves execution records for a specific task
func (r *FileTaskRepository) GetTaskExecutions(ctx context.Context, taskID string, limit int) ([]*tasks.TaskExecution, error) {
	if taskID == "" {
		return nil, ErrInvalidTaskID
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	taskExecDir := r.taskExecutionsDir(taskID)

	// Create directory if it doesn't exist (no executions yet)
	if _, err := os.Stat(taskExecDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*tasks.TaskExecution{}, nil
		}
		return nil, fmt.Errorf("failed to check task executions directory: %w", err)
	}

	// Read all execution files for the task
	files, err := filepath.Glob(filepath.Join(taskExecDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list execution files: %w", err)
	}

	var executions []*tasks.TaskExecution

	for _, file := range files {
		execution, err := r.readExecutionFromFile(file)
		if err != nil {
			// Log error but continue with other executions
			continue
		}
		executions = append(executions, execution)
	}

	// Sort executions by start time (newest first)
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartTime.After(executions[j].StartTime)
	})

	// Apply limit if specified
	if limit > 0 && len(executions) > limit {
		executions = executions[:limit]
	}

	return executions, nil
}

// GetExecution retrieves a specific execution record by ID
func (r *FileTaskRepository) GetExecution(ctx context.Context, id string) (*tasks.TaskExecution, error) {
	if id == "" {
		return nil, errors.New("execution ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	filePath := r.executionFilePath(id)
	execution, err := r.readExecutionFromFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrExecutionNotFound
		}
		return nil, fmt.Errorf("failed to read execution: %w", err)
	}

	return execution, nil
}

// GetExecutions retrieves all executions for a task
func (r *FileTaskRepository) GetExecutions(ctx context.Context, taskID string) ([]*tasks.TaskExecution, error) {
	if taskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// List all execution files in the executions directory
	executionsDir := filepath.Join(r.baseDir, ExecutionsDir)
	files, err := os.ReadDir(executionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list executions directory: %w", err)
	}

	var executions []*tasks.TaskExecution
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only consider JSON files
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Read the execution file
		filePath := filepath.Join(executionsDir, file.Name())
		execution, err := r.readExecutionFromFile(filePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		// Check if this execution is for the requested task
		if execution.TaskID == taskID {
			executions = append(executions, execution)
		}
	}

	// Sort executions by start time (newest first)
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartTime.After(executions[j].StartTime)
	})

	return executions, nil
}

// writeTaskToFile writes a task to a file atomically
func (r *FileTaskRepository) writeTaskToFile(task *tasks.TaskConfig, filePath string) error {
	// Get file lock
	lock := r.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// Create a temporary file
	tempFile := filePath + ".tmp"
	file, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer file.Close()

	// Marshal task to JSON
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Write data to temporary file
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write task data: %w", err)
	}

	// Ensure data is written to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync task data: %w", err)
	}

	// Close the file before renaming
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomically replace the old file with the new one
	if err := os.Rename(tempFile, filePath); err != nil {
		return fmt.Errorf("failed to replace task file: %w", err)
	}

	return nil
}

// readTaskFromFile reads a task from a file
func (r *FileTaskRepository) readTaskFromFile(filePath string) (*tasks.TaskConfig, error) {
	// Get file lock
	lock := r.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	// Unmarshal JSON to task
	var task tasks.TaskConfig
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task file: %w", err)
	}

	return &task, nil
}

// writeExecutionToFile writes an execution record to a file atomically
func (r *FileTaskRepository) writeExecutionToFile(execution *tasks.TaskExecution, filePath string) error {
	// Get file lock
	lock := r.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// Create a temporary file
	tempFile := filePath + ".tmp"
	file, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer file.Close()

	// Marshal execution to JSON
	data, err := json.MarshalIndent(execution, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal execution: %w", err)
	}

	// Write data to temporary file
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write execution data: %w", err)
	}

	// Ensure data is written to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync execution data: %w", err)
	}

	// Close the file before renaming
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomically replace the old file with the new one
	if err := os.Rename(tempFile, filePath); err != nil {
		return fmt.Errorf("failed to replace execution file: %w", err)
	}

	return nil
}

// readExecutionFromFile reads an execution record from a file
func (r *FileTaskRepository) readExecutionFromFile(filePath string) (*tasks.TaskExecution, error) {
	// Get file lock
	lock := r.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read execution file: %w", err)
	}

	// Unmarshal JSON to execution
	var execution tasks.TaskExecution
	if err := json.Unmarshal(data, &execution); err != nil {
		return nil, fmt.Errorf("failed to parse execution file: %w", err)
	}

	return &execution, nil
}
