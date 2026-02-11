package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

const (
	// ConfigDir is the default config directory name.
	ConfigDir = ".gomikrobot"
	// ConfigFile is the default config file name.
	ConfigFile = "config.json"
)

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigDir, ConfigFile), nil
}

// Load loads the configuration from file and environment variables.
// Priority: environment > file > defaults
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load from file
	path, err := ConfigPath()
	if err != nil {
		return cfg, nil // Use defaults if we can't find config path
	}

	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}
	// If file doesn't exist, continue with defaults

	// Override with environment variables for each section
	envconfig.Process("MIKROBOT_OPENAI", &cfg.Providers.OpenAI)
	envconfig.Process("MIKROBOT_AGENTS", &cfg.Agents.Defaults)
	envconfig.Process("MIKROBOT_CHANNELS_TELEGRAM", &cfg.Channels.Telegram)
	envconfig.Process("MIKROBOT_CHANNELS_DISCORD", &cfg.Channels.Discord)
	envconfig.Process("MIKROBOT_CHANNELS_WHATSAPP", &cfg.Channels.WhatsApp)
	envconfig.Process("MIKROBOT_CHANNELS_FEISHU", &cfg.Channels.Feishu)
	envconfig.Process("MIKROBOT_GATEWAY", &cfg.Gateway)
	envconfig.Process("MIKROBOT_TOOLS_EXEC", &cfg.Tools.Exec)
	envconfig.Process("MIKROBOT_TOOLS_WEB_SEARCH", &cfg.Tools.Web.Search)

	// Fallback for API Key
	if cfg.Providers.OpenAI.APIKey == "" {
		if key := os.Getenv("OPENAI_API_KEY"); key != "" {
			cfg.Providers.OpenAI.APIKey = key
		} else if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
			cfg.Providers.OpenAI.APIKey = key
		}
	}

	// Expand ~ in workspace path
	if strings.HasPrefix(cfg.Agents.Defaults.Workspace, "~") {
		home, _ := os.UserHomeDir()
		cfg.Agents.Defaults.Workspace = filepath.Join(home, cfg.Agents.Defaults.Workspace[1:])
	}

	return cfg, nil
}

// Save writes the configuration to the config file.
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// EnsureDir ensures a directory exists with proper permissions.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
