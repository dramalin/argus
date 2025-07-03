package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration loaded from YAML and environment variables.
type Config struct {
	Server struct {
		Port         int    `yaml:"port"`
		Host         string `yaml:"host"`
		ReadTimeout  string `yaml:"read_timeout"`
		WriteTimeout string `yaml:"write_timeout"`
	} `yaml:"server"`

	Monitoring struct {
		UpdateInterval   string `yaml:"update_interval"`
		MetricsRetention string `yaml:"metrics_retention"`
		ProcessLimit     int    `yaml:"process_limit"`
	} `yaml:"monitoring"`

	Alerts struct {
		Enabled              bool   `yaml:"enabled"`
		StoragePath          string `yaml:"storage_path"`
		NotificationInterval string `yaml:"notification_interval"`
	} `yaml:"alerts"`

	Tasks struct {
		Enabled       bool   `yaml:"enabled"`
		StoragePath   string `yaml:"storage_path"`
		MaxConcurrent int    `yaml:"max_concurrent"`
	} `yaml:"tasks"`

	Storage struct {
		BasePath        string `yaml:"base_path"`
		FilePermissions int    `yaml:"file_permissions"`
		BackupEnabled   bool   `yaml:"backup_enabled"`
	} `yaml:"storage"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		File   string `yaml:"file"`
	} `yaml:"logging"`

	WebSocket struct {
		Enabled         bool   `yaml:"enabled"`
		Path            string `yaml:"path"`
		ReadBufferSize  int    `yaml:"read_buffer_size"`
		WriteBufferSize int    `yaml:"write_buffer_size"`
	} `yaml:"websocket"`

	CORS struct {
		Enabled        bool     `yaml:"enabled"`
		AllowedOrigins []string `yaml:"allowed_origins"`
		AllowedMethods []string `yaml:"allowed_methods"`
		AllowedHeaders []string `yaml:"allowed_headers"`
	} `yaml:"cors"`
}

// LoadConfig loads configuration from a YAML file and applies environment variable overrides.
func LoadConfig(path string) (*Config, error) {
	cfg := defaultConfig()
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
		defer f.Close()
		decoder := yaml.NewDecoder(f)
		if err := decoder.Decode(cfg); err != nil {
			return nil, fmt.Errorf("failed to decode config yaml: %w", err)
		}
	}
	applyEnvOverrides(cfg)
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// defaultConfig returns a Config struct with default values.
func defaultConfig() *Config {
	return &Config{
		Server: struct {
			Port         int    `yaml:"port"`
			Host         string `yaml:"host"`
			ReadTimeout  string `yaml:"read_timeout"`
			WriteTimeout string `yaml:"write_timeout"`
		}{
			Port:         8080,
			Host:         "localhost",
			ReadTimeout:  "30s",
			WriteTimeout: "30s",
		},
		Monitoring: struct {
			UpdateInterval   string `yaml:"update_interval"`
			MetricsRetention string `yaml:"metrics_retention"`
			ProcessLimit     int    `yaml:"process_limit"`
		}{
			UpdateInterval:   "5s",
			MetricsRetention: "24h",
			ProcessLimit:     100,
		},
		Alerts: struct {
			Enabled              bool   `yaml:"enabled"`
			StoragePath          string `yaml:"storage_path"`
			NotificationInterval string `yaml:"notification_interval"`
		}{
			Enabled:              true,
			StoragePath:          "./.argus/alerts",
			NotificationInterval: "1m",
		},
		Tasks: struct {
			Enabled       bool   `yaml:"enabled"`
			StoragePath   string `yaml:"storage_path"`
			MaxConcurrent int    `yaml:"max_concurrent"`
		}{
			Enabled:       true,
			StoragePath:   "./.argus/tasks",
			MaxConcurrent: 5,
		},
		Storage: struct {
			BasePath        string `yaml:"base_path"`
			FilePermissions int    `yaml:"file_permissions"`
			BackupEnabled   bool   `yaml:"backup_enabled"`
		}{
			BasePath:        "./.argus",
			FilePermissions: 0644,
			BackupEnabled:   true,
		},
		Logging: struct {
			Level  string `yaml:"level"`
			Format string `yaml:"format"`
			File   string `yaml:"file"`
		}{
			Level:  "info",
			Format: "json",
			File:   "",
		},
		WebSocket: struct {
			Enabled         bool   `yaml:"enabled"`
			Path            string `yaml:"path"`
			ReadBufferSize  int    `yaml:"read_buffer_size"`
			WriteBufferSize int    `yaml:"write_buffer_size"`
		}{
			Enabled:         true,
			Path:            "/ws",
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		CORS: struct {
			Enabled        bool     `yaml:"enabled"`
			AllowedOrigins []string `yaml:"allowed_origins"`
			AllowedMethods []string `yaml:"allowed_methods"`
			AllowedHeaders []string `yaml:"allowed_headers"`
		}{
			Enabled:        true,
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
	}
}

// applyEnvOverrides applies environment variable overrides to the config struct.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ARGUS_SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	if v := os.Getenv("ARGUS_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	// Add more environment variable overrides as needed for other fields
}

// validateConfig checks for required fields and valid values.
func validateConfig(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return errors.New("invalid server port")
	}
	if _, err := time.ParseDuration(cfg.Server.ReadTimeout); err != nil {
		return fmt.Errorf("invalid server read_timeout: %w", err)
	}
	if _, err := time.ParseDuration(cfg.Server.WriteTimeout); err != nil {
		return fmt.Errorf("invalid server write_timeout: %w", err)
	}
	return nil
}
