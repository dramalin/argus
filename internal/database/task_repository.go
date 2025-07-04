// Package database provides task storage and repository logic for Argus
package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"argus/internal/models"
)

// FileTaskRepository implements models.TaskRepository using the filesystem
// Moved from internal/models/task.go as part of architecture migration
// @migration 2024-07-03
// @author Argus

const (
	TasksDir      = "configurations"
	ExecutionsDir = "executions"
)

var (
	ErrTaskNotFound      = errors.New("task configuration not found")
	ErrExecutionNotFound = errors.New("task execution record not found")
	ErrInvalidTaskID     = errors.New("invalid task ID")
)

type FileTaskRepository struct {
	baseDir       string
	tasksDir      string
	executionsDir string
	mutex         sync.RWMutex
	fileLocks     map[string]*sync.Mutex
	lockMu        sync.Mutex
}

func NewFileTaskRepository(baseDir string) (*FileTaskRepository, error) {
	if baseDir == "" {
		baseDir = DefaultConfigDir
	}
	tasksDir := filepath.Join(baseDir, TasksDir)
	executionsDir := filepath.Join(baseDir, ExecutionsDir)
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

func (r *FileTaskRepository) taskFilePath(id string) string {
	return filepath.Join(r.tasksDir, fmt.Sprintf("%s.json", id))
}

func (r *FileTaskRepository) executionFilePath(id string) string {
	return filepath.Join(r.executionsDir, fmt.Sprintf("%s.json", id))
}

func (r *FileTaskRepository) taskExecutionsDir(taskID string) string {
	return filepath.Join(r.executionsDir, taskID)
}

func (r *FileTaskRepository) CreateTask(ctx context.Context, task *models.TaskConfig) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}
	if task.ID == "" {
		task.ID = models.GenerateID()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.UpdatedAt = time.Now()
	if err := task.Validate(); err != nil {
		return err
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	filePath := r.taskFilePath(task.ID)
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check if task exists: %w", err)
	}
	return r.writeTaskToFile(task, filePath)
}

func (r *FileTaskRepository) GetTask(ctx context.Context, id string) (*models.TaskConfig, error) {
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

func (r *FileTaskRepository) UpdateTask(ctx context.Context, task *models.TaskConfig) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}
	if task.ID == "" {
		return ErrInvalidTaskID
	}
	if err := task.Validate(); err != nil {
		return err
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	filePath := r.taskFilePath(task.ID)
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("failed to check if task exists: %w", err)
	}
	task.UpdatedAt = time.Now()
	return r.writeTaskToFile(task, filePath)
}

func (r *FileTaskRepository) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidTaskID
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	filePath := r.taskFilePath(id)
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("failed to check if task exists: %w", err)
	}
	lock := r.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *FileTaskRepository) ListTasks(ctx context.Context) ([]*models.TaskConfig, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	var tasksList []*models.TaskConfig
	files, err := filepath.Glob(filepath.Join(r.tasksDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list task files: %w", err)
	}
	for _, file := range files {
		task, err := r.readTaskFromFile(file)
		if err != nil {
			continue
		}
		tasksList = append(tasksList, task)
	}
	sort.Slice(tasksList, func(i, j int) bool {
		return tasksList[i].CreatedAt.After(tasksList[j].CreatedAt)
	})
	return tasksList, nil
}

func (r *FileTaskRepository) GetTasksByType(ctx context.Context, taskType models.TaskType) ([]*models.TaskConfig, error) {
	allTasks, err := r.ListTasks(ctx)
	if err != nil {
		return nil, err
	}
	var tasksOfType []*models.TaskConfig
	for _, task := range allTasks {
		if task.Type == taskType {
			tasksOfType = append(tasksOfType, task)
		}
	}
	return tasksOfType, nil
}

func (r *FileTaskRepository) RecordExecution(ctx context.Context, execution *models.TaskExecution) error {
	if execution == nil {
		return errors.New("execution cannot be nil")
	}
	if execution.ID == "" {
		execution.ID = models.GenerateID()
	}
	if execution.TaskID == "" {
		return errors.New("task ID is required for execution record")
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	taskExecDir := r.taskExecutionsDir(execution.TaskID)
	if err := os.MkdirAll(taskExecDir, DefaultDirMode); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrDirectoryCreation, taskExecDir, err)
	}
	filePath := r.executionFilePath(execution.ID)
	return r.writeExecutionToFile(execution, filePath)
}

func (r *FileTaskRepository) GetTaskExecutions(ctx context.Context, taskID string, limit int) ([]*models.TaskExecution, error) {
	if taskID == "" {
		return nil, ErrInvalidTaskID
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	taskExecDir := r.taskExecutionsDir(taskID)
	if _, err := os.Stat(taskExecDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*models.TaskExecution{}, nil
		}
		return nil, fmt.Errorf("failed to check task executions directory: %w", err)
	}
	files, err := filepath.Glob(filepath.Join(taskExecDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list execution files: %w", err)
	}
	var executions []*models.TaskExecution
	for _, file := range files {
		exec, err := r.readExecutionFromFile(file)
		if err != nil {
			continue
		}
		executions = append(executions, exec)
	}
	if limit > 0 && len(executions) > limit {
		executions = executions[:limit]
	}
	return executions, nil
}

func (r *FileTaskRepository) GetExecution(ctx context.Context, id string) (*models.TaskExecution, error) {
	if id == "" {
		return nil, ErrInvalidTaskID
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	filePath := r.executionFilePath(id)
	exec, err := r.readExecutionFromFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrExecutionNotFound
		}
		return nil, fmt.Errorf("failed to read execution: %w", err)
	}
	return exec, nil
}

func (r *FileTaskRepository) GetExecutions(ctx context.Context, taskID string) ([]*models.TaskExecution, error) {
	return r.GetTaskExecutions(ctx, taskID, 0)
}

func (r *FileTaskRepository) writeTaskToFile(task *models.TaskConfig, filePath string) error {
	lock := r.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(task); err != nil {
		return fmt.Errorf("failed to encode task: %w", err)
	}
	return nil
}

func (r *FileTaskRepository) readTaskFromFile(filePath string) (*models.TaskConfig, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var task models.TaskConfig
	dec := json.NewDecoder(f)
	if err := dec.Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode task: %w", err)
	}
	return &task, nil
}

func (r *FileTaskRepository) writeExecutionToFile(execution *models.TaskExecution, filePath string) error {
	lock := r.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(execution); err != nil {
		return fmt.Errorf("failed to encode execution: %w", err)
	}
	return nil
}

func (r *FileTaskRepository) readExecutionFromFile(filePath string) (*models.TaskExecution, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var exec models.TaskExecution
	dec := json.NewDecoder(f)
	if err := dec.Decode(&exec); err != nil {
		return nil, fmt.Errorf("failed to decode execution: %w", err)
	}
	return &exec, nil
}
