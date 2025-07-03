// Package database provides functionality for persisting and retrieving configurations
package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"argus/internal/models"
)

const (
	// DefaultConfigDir is the default directory for storing configurations
	DefaultConfigDir = ".argus/config"

	// AlertsDir is the subdirectory for storing alert configurations
	AlertsDir = "alerts"

	// BackupDir is the subdirectory for storing configuration backups
	BackupDir = "backups"

	// DefaultFileMode is the default file permission mode for configuration files
	DefaultFileMode = 0644

	// DefaultDirMode is the default directory permission mode
	DefaultDirMode = 0755
)

var (
	// ErrAlertNotFound is returned when an alert configuration is not found
	ErrAlertNotFound = errors.New("alert configuration not found")

	// ErrInvalidAlertID is returned when an alert ID is invalid
	ErrInvalidAlertID = errors.New("invalid alert ID")

	// ErrDirectoryCreation is returned when a directory cannot be created
	ErrDirectoryCreation = errors.New("failed to create directory")

	// ErrFileLocked is returned when a file is locked for writing
	ErrFileLocked = errors.New("file is locked for writing")
)

// AlertStore manages the storage of alert configurations
type AlertStore struct {
	configDir string
	alertsDir string
	backupDir string
	mu        sync.RWMutex
	fileLocks map[string]*sync.Mutex
	lockMu    sync.Mutex
}

// NewAlertStore creates a new AlertStore with the given configuration directory
func NewAlertStore(configDir string) (*AlertStore, error) {
	if configDir == "" {
		configDir = DefaultConfigDir
	}

	alertsDir := filepath.Join(configDir, AlertsDir)
	backupDir := filepath.Join(configDir, BackupDir)

	// Create directories if they don't exist
	if err := os.MkdirAll(alertsDir, DefaultDirMode); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrDirectoryCreation, alertsDir, err)
	}

	if err := os.MkdirAll(backupDir, DefaultDirMode); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrDirectoryCreation, backupDir, err)
	}

	return &AlertStore{
		configDir: configDir,
		alertsDir: alertsDir,
		backupDir: backupDir,
		fileLocks: make(map[string]*sync.Mutex),
	}, nil
}

// getFileLock returns a mutex for the given file path, creating one if it doesn't exist
func (s *AlertStore) getFileLock(path string) *sync.Mutex {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()

	if lock, exists := s.fileLocks[path]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	s.fileLocks[path] = lock
	return lock
}

// alertFilePath returns the file path for the given alert ID
func (s *AlertStore) alertFilePath(id string) string {
	return filepath.Join(s.alertsDir, fmt.Sprintf("%s.json", id))
}

// backupFilePath returns the backup file path for the given alert ID
func (s *AlertStore) backupFilePath(id string) string {
	timestamp := time.Now().Format("20060102-150405")
	return filepath.Join(s.backupDir, fmt.Sprintf("%s-%s.json", id, timestamp))
}

// CreateAlert stores a new alert configuration
func (s *AlertStore) CreateAlert(alert *models.AlertConfig) error {
	// Check if alert ID is valid
	if alert.ID == "" {
		// Generate a new UUID if ID is empty
		alert.ID = uuid.New().String()
	}

	// Check if file already exists
	filePath := s.alertFilePath(alert.ID)
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("alert with ID %s already exists", alert.ID)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("error checking file: %w", err)
	}

	// Update timestamps
	now := time.Now()
	alert.CreatedAt = now
	alert.UpdatedAt = now

	// Validate the alert configuration
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("invalid alert configuration: %w", err)
	}

	// Get file lock
	lock := s.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// Marshal the alert configuration to JSON
	data, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal alert configuration: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, DefaultFileMode); err != nil {
		return fmt.Errorf("failed to write alert configuration: %w", err)
	}

	return nil
}

// GetAlert retrieves an alert configuration by ID
func (s *AlertStore) GetAlert(id string) (*models.AlertConfig, error) {
	if id == "" {
		return nil, ErrInvalidAlertID
	}

	filePath := s.alertFilePath(id)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrAlertNotFound
		}
		return nil, fmt.Errorf("error checking file: %w", err)
	}

	// Get file lock for reading
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read alert configuration: %w", err)
	}

	// Unmarshal the JSON data
	alert := &models.AlertConfig{}
	if err := json.Unmarshal(data, alert); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert configuration: %w", err)
	}

	return alert, nil
}

// UpdateAlert updates an existing alert configuration
func (s *AlertStore) UpdateAlert(alert *models.AlertConfig) error {
	if alert.ID == "" {
		return ErrInvalidAlertID
	}

	filePath := s.alertFilePath(alert.ID)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrAlertNotFound
		}
		return fmt.Errorf("error checking file: %w", err)
	}

	// Create a backup of the existing file
	if err := s.backupAlert(alert.ID); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Update timestamp
	alert.UpdatedAt = time.Now()

	// Validate the alert configuration
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("invalid alert configuration: %w", err)
	}

	// Get file lock
	lock := s.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// Marshal the alert configuration to JSON
	data, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal alert configuration: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, DefaultFileMode); err != nil {
		return fmt.Errorf("failed to write alert configuration: %w", err)
	}

	return nil
}

// DeleteAlert removes an alert configuration
func (s *AlertStore) DeleteAlert(id string) error {
	if id == "" {
		return ErrInvalidAlertID
	}

	filePath := s.alertFilePath(id)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrAlertNotFound
		}
		return fmt.Errorf("error checking file: %w", err)
	}

	// Create a backup of the file before deletion
	if err := s.backupAlert(id); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Get file lock
	lock := s.getFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete alert configuration: %w", err)
	}

	return nil
}

// ListAlerts returns a list of all alert configurations
func (s *AlertStore) ListAlerts() ([]*models.AlertConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alertConfigs []*models.AlertConfig

	// Read all JSON files in the alerts directory
	files, err := os.ReadDir(s.alertsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read alerts directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(s.alertsDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read alert configuration %s: %w", file.Name(), err)
		}

		alert := &models.AlertConfig{}
		if err := json.Unmarshal(data, alert); err != nil {
			return nil, fmt.Errorf("failed to unmarshal alert configuration %s: %w", file.Name(), err)
		}

		alertConfigs = append(alertConfigs, alert)
	}

	return alertConfigs, nil
}

// backupAlert creates a backup of an alert configuration
func (s *AlertStore) backupAlert(id string) error {
	srcPath := s.alertFilePath(id)
	dstPath := s.backupFilePath(id)

	// Read the source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read alert configuration for backup: %w", err)
	}

	// Write the backup file
	if err := os.WriteFile(dstPath, data, DefaultFileMode); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// RestoreAlert restores an alert configuration from a backup
func (s *AlertStore) RestoreAlert(id string, timestamp string) error {
	if id == "" {
		return ErrInvalidAlertID
	}

	backupPath := filepath.Join(s.backupDir, fmt.Sprintf("%s-%s.json", id, timestamp))
	destPath := s.alertFilePath(id)

	// Check if backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("backup file not found: %s", backupPath)
		}
		return fmt.Errorf("error checking backup file: %w", err)
	}

	// Get file lock
	lock := s.getFileLock(destPath)
	lock.Lock()
	defer lock.Unlock()

	// Read the backup file
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Validate the backup data
	alert := &models.AlertConfig{}
	if err := json.Unmarshal(data, alert); err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Write the file
	if err := os.WriteFile(destPath, data, DefaultFileMode); err != nil {
		return fmt.Errorf("failed to restore alert configuration: %w", err)
	}

	return nil
}

// ListBackups returns a list of available backups for an alert
func (s *AlertStore) ListBackups(id string) ([]string, error) {
	if id == "" {
		return nil, ErrInvalidAlertID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var backups []string
	prefix := fmt.Sprintf("%s-", id)

	// Read all files in the backup directory
	files, err := os.ReadDir(s.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backups directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		if len(file.Name()) > len(prefix) && file.Name()[:len(prefix)] == prefix {
			// Extract timestamp from filename
			timestamp := file.Name()[len(prefix) : len(file.Name())-5] // Remove prefix and .json extension
			backups = append(backups, timestamp)
		}
	}

	return backups, nil
}
