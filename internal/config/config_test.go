package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got %s", cfg.Server.Host)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	os.Setenv("ARGUS_SERVER_PORT", "9090")
	os.Setenv("ARGUS_SERVER_HOST", "0.0.0.0")
	defer os.Unsetenv("ARGUS_SERVER_PORT")
	defer os.Unsetenv("ARGUS_SERVER_HOST")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected env override port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected env override host '0.0.0.0', got %s", cfg.Server.Host)
	}
}

func TestValidateConfig_InvalidPort(t *testing.T) {
	cfg := defaultConfig()
	cfg.Server.Port = -1
	if err := validateConfig(cfg); err == nil {
		t.Error("expected error for invalid port, got nil")
	}
}

func TestValidateConfig_InvalidTimeout(t *testing.T) {
	cfg := defaultConfig()
	cfg.Server.ReadTimeout = "notaduration"
	if err := validateConfig(cfg); err == nil {
		t.Error("expected error for invalid read_timeout, got nil")
	}
}
