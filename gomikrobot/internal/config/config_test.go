package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Agents.Defaults.Model != "anthropic/claude-sonnet-4-5" {
		t.Errorf("expected default model anthropic/claude-sonnet-4-5, got %s", cfg.Agents.Defaults.Model)
	}

	if cfg.Gateway.Host != "127.0.0.1" {
		t.Errorf("expected gateway host 127.0.0.1, got %s", cfg.Gateway.Host)
	}

	if cfg.Gateway.Port != 18790 {
		t.Errorf("expected gateway port 18790, got %d", cfg.Gateway.Port)
	}

	if !cfg.Tools.Exec.RestrictToWorkspace {
		t.Error("expected RestrictToWorkspace to be true by default")
	}

	if cfg.Tools.Exec.Timeout != 60*time.Second {
		t.Errorf("expected exec timeout 60s, got %v", cfg.Tools.Exec.Timeout)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Temporarily set HOME to a non-existent directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", "/tmp/nonexistent-nanobot-test")
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Agents.Defaults.MaxTokens != 8192 {
		t.Errorf("expected maxTokens 8192, got %d", cfg.Agents.Defaults.MaxTokens)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create temp config
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".nanobot")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "config.json")

	configJSON := `{
		"agents": {
			"defaults": {
				"model": "openai/gpt-4",
				"maxTokens": 4096
			}
		},
		"gateway": {
			"port": 9999
		}
	}`
	os.WriteFile(configFile, []byte(configJSON), 0600)

	// Temporarily set HOME
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Agents.Defaults.Model != "openai/gpt-4" {
		t.Errorf("expected model openai/gpt-4, got %s", cfg.Agents.Defaults.Model)
	}

	if cfg.Gateway.Port != 9999 {
		t.Errorf("expected port 9999, got %d", cfg.Gateway.Port)
	}
}

func TestEnvOverride(t *testing.T) {
	// Set env var with correct prefix for nested struct
	os.Setenv("NANOBOT_GATEWAY_HOST", "0.0.0.0")
	os.Setenv("NANOBOT_GATEWAY_PORT", "8080")
	defer func() {
		os.Unsetenv("NANOBOT_GATEWAY_HOST")
		os.Unsetenv("NANOBOT_GATEWAY_PORT")
	}()

	// Use temp home with no config file
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Gateway.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0 from env, got %s", cfg.Gateway.Host)
	}

	if cfg.Gateway.Port != 8080 {
		t.Errorf("expected port 8080 from env, got %d", cfg.Gateway.Port)
	}
}
