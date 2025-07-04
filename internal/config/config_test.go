package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_FromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
server:
  host: "testhost"
  port: 9090
  read_timeout: "60s"
  write_timeout: "30s"
storage:
  base_path: "/tmp/test-storage"
  file_permissions: 0644
  backup_enabled: true
alerts:
  enabled: true
  storage_path: "/tmp/alerts"
  notification_interval: "30s"
monitoring:
  update_interval: "15s"
  metrics_retention: "7d"
  process_limit: 50
tasks:
  enabled: true
  storage_path: "/tmp/tasks"
  max_concurrent: 5
logging:
  level: "debug"
  format: "json"
  file: "/tmp/logs/argus.log"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "Failed to create test config file")

	// Load the config
	cfg, err := LoadConfig(configPath)
	require.NoError(t, err, "Failed to load config")

	// Verify the config was loaded correctly
	assert.Equal(t, "testhost", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "60s", cfg.Server.ReadTimeout)
	assert.Equal(t, "30s", cfg.Server.WriteTimeout)

	assert.Equal(t, "/tmp/test-storage", cfg.Storage.BasePath)
	assert.Equal(t, 0644, cfg.Storage.FilePermissions)
	assert.True(t, cfg.Storage.BackupEnabled)

	assert.True(t, cfg.Alerts.Enabled)
	assert.Equal(t, "/tmp/alerts", cfg.Alerts.StoragePath)
	assert.Equal(t, "30s", cfg.Alerts.NotificationInterval)

	assert.Equal(t, "15s", cfg.Monitoring.UpdateInterval)
	assert.Equal(t, "7d", cfg.Monitoring.MetricsRetention)
	assert.Equal(t, 50, cfg.Monitoring.ProcessLimit)

	assert.True(t, cfg.Tasks.Enabled)
	assert.Equal(t, "/tmp/tasks", cfg.Tasks.StoragePath)
	assert.Equal(t, 5, cfg.Tasks.MaxConcurrent)

	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "/tmp/logs/argus.log", cfg.Logging.File)
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	// Create a temporary invalid config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yaml")

	configContent := `
server:
  host: "testhost"
  port: "not-a-number" # This should cause an error
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "Failed to create test config file")

	// Load the config
	_, err = LoadConfig(configPath)
	assert.Error(t, err, "Should fail with invalid config")
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "Valid config",
			config: &Config{
				Server: struct {
					Port         int    `yaml:"port"`
					Host         string `yaml:"host"`
					ReadTimeout  string `yaml:"read_timeout"`
					WriteTimeout string `yaml:"write_timeout"`
				}{
					Host:         "localhost",
					Port:         8080,
					ReadTimeout:  "30s",
					WriteTimeout: "30s",
				},
			},
			expectError: false,
		},
		{
			name: "Invalid server config - invalid port",
			config: &Config{
				Server: struct {
					Port         int    `yaml:"port"`
					Host         string `yaml:"host"`
					ReadTimeout  string `yaml:"read_timeout"`
					WriteTimeout string `yaml:"write_timeout"`
				}{
					Host:         "localhost",
					Port:         -1,
					ReadTimeout:  "30s",
					WriteTimeout: "30s",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid server config - invalid read timeout",
			config: &Config{
				Server: struct {
					Port         int    `yaml:"port"`
					Host         string `yaml:"host"`
					ReadTimeout  string `yaml:"read_timeout"`
					WriteTimeout string `yaml:"write_timeout"`
				}{
					Host:         "localhost",
					Port:         8080,
					ReadTimeout:  "invalid",
					WriteTimeout: "30s",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadLocation(t *testing.T) {
	// Test with valid timezone
	tz := LoadLocation("America/New_York")
	assert.NotNil(t, tz)
	assert.Equal(t, "America/New_York", tz.String())

	// Test with invalid timezone should return UTC
	tz = LoadLocation("Invalid/TimeZone")
	assert.NotNil(t, tz)
	assert.Equal(t, "UTC", tz.String())

	// Test with empty timezone should return UTC
	tz = LoadLocation("")
	assert.NotNil(t, tz)
	assert.Equal(t, "UTC", tz.String())
}
