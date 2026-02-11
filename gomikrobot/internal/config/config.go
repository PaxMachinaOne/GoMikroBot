// Package config provides configuration types and loading for gonanobot.
package config

import "time"

// Config is the root configuration struct.
type Config struct {
	Agents    AgentsConfig    `json:"agents"`
	Channels  ChannelsConfig  `json:"channels"`
	Providers ProvidersConfig `json:"providers"`
	Gateway   GatewayConfig   `json:"gateway"`
	Tools     ToolsConfig     `json:"tools"`
}

// AgentsConfig contains agent-related settings.
type AgentsConfig struct {
	Defaults AgentDefaults `json:"defaults"`
}

// AgentDefaults contains default agent settings.
type AgentDefaults struct {
	Workspace         string  `json:"workspace" envconfig:"WORKSPACE"`
	Model             string  `json:"model" envconfig:"MODEL"`
	MaxTokens         int     `json:"maxTokens" envconfig:"MAX_TOKENS"`
	Temperature       float64 `json:"temperature" envconfig:"TEMPERATURE"`
	MaxToolIterations int     `json:"maxToolIterations" envconfig:"MAX_TOOL_ITERATIONS"`
}

// ChannelsConfig contains all channel configurations.
type ChannelsConfig struct {
	Telegram TelegramConfig `json:"telegram"`
	Discord  DiscordConfig  `json:"discord"`
	WhatsApp WhatsAppConfig `json:"whatsapp"`
	Feishu   FeishuConfig   `json:"feishu"`
}

// TelegramConfig configures the Telegram channel.
type TelegramConfig struct {
	Enabled   bool     `json:"enabled" envconfig:"TELEGRAM_ENABLED"`
	Token     string   `json:"token" envconfig:"TELEGRAM_TOKEN"`
	AllowFrom []string `json:"allowFrom"`
	Proxy     string   `json:"proxy,omitempty" envconfig:"TELEGRAM_PROXY"`
}

// DiscordConfig configures the Discord channel.
type DiscordConfig struct {
	Enabled   bool     `json:"enabled" envconfig:"DISCORD_ENABLED"`
	Token     string   `json:"token" envconfig:"DISCORD_TOKEN"`
	AllowFrom []string `json:"allowFrom"`
}

// WhatsAppConfig configures the WhatsApp channel.
type WhatsAppConfig struct {
	Enabled   bool     `json:"enabled" envconfig:"WHATSAPP_ENABLED"`
	BridgeURL string   `json:"bridgeUrl" envconfig:"WHATSAPP_BRIDGE_URL"`
	AllowFrom []string `json:"allowFrom"`
}

// FeishuConfig configures the Feishu channel.
type FeishuConfig struct {
	Enabled           bool     `json:"enabled" envconfig:"FEISHU_ENABLED"`
	AppID             string   `json:"appId" envconfig:"FEISHU_APP_ID"`
	AppSecret         string   `json:"appSecret" envconfig:"FEISHU_APP_SECRET"`
	EncryptKey        string   `json:"encryptKey" envconfig:"FEISHU_ENCRYPT_KEY"`
	VerificationToken string   `json:"verificationToken" envconfig:"FEISHU_VERIFICATION_TOKEN"`
	AllowFrom         []string `json:"allowFrom"`
}

// ProvidersConfig contains LLM provider configurations.
type ProvidersConfig struct {
	Anthropic    ProviderConfig     `json:"anthropic"`
	OpenAI       ProviderConfig     `json:"openai"`
	LocalWhisper LocalWhisperConfig `json:"localWhisper"`
	OpenRouter   ProviderConfig     `json:"openrouter"`
	DeepSeek     ProviderConfig     `json:"deepseek"`
	Groq         ProviderConfig     `json:"groq"`
	Gemini       ProviderConfig     `json:"gemini"`
	VLLM         ProviderConfig     `json:"vllm"`
}

// ProviderConfig contains settings for a single LLM provider.
type ProviderConfig struct {
	APIKey  string `json:"apiKey" envconfig:"API_KEY"`
	APIBase string `json:"apiBase,omitempty" envconfig:"API_BASE"`
}

// LocalWhisperConfig contains settings for local Whisper transcription.
type LocalWhisperConfig struct {
	Enabled    bool   `json:"enabled" envconfig:"WHISPER_ENABLED"`
	Model      string `json:"model" envconfig:"WHISPER_MODEL"`
	BinaryPath string `json:"binaryPath" envconfig:"WHISPER_BINARY_PATH"`
}

// GatewayConfig contains gateway server settings.
type GatewayConfig struct {
	Host          string `json:"host" envconfig:"HOST"`
	Port          int    `json:"port" envconfig:"PORT"`
	DashboardPort int    `json:"dashboardPort" envconfig:"DASHBOARD_PORT"`

	// Optional API token for local-network API.
	APIToken string `json:"apiToken,omitempty" envconfig:"API_TOKEN"`

	// Enterprise hardening.
	RateLimitRPS    float64       `json:"rateLimitRps" envconfig:"RATE_LIMIT_RPS"`
	RateLimitBurst  int           `json:"rateLimitBurst" envconfig:"RATE_LIMIT_BURST"`
	MaxBodyBytes    int64         `json:"maxBodyBytes" envconfig:"MAX_BODY_BYTES"`
	ShutdownTimeout time.Duration `json:"shutdownTimeout" envconfig:"SHUTDOWN_TIMEOUT"`
}

// ToolsConfig contains tool-specific settings.
type ToolsConfig struct {
	Exec ExecToolConfig `json:"exec"`
	Web  WebToolConfig  `json:"web"`
}

// ExecToolConfig contains shell execution tool settings.
type ExecToolConfig struct {
	Timeout             time.Duration `json:"timeout"`
	RestrictToWorkspace bool          `json:"restrictToWorkspace" envconfig:"EXEC_RESTRICT_WORKSPACE"`
}

// WebToolConfig contains web tool settings.
type WebToolConfig struct {
	Search SearchConfig `json:"search"`
}

// SearchConfig contains web search settings.
type SearchConfig struct {
	APIKey     string `json:"apiKey" envconfig:"BRAVE_API_KEY"`
	MaxResults int    `json:"maxResults"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Workspace:         "~/.gomikrobot/workspace",
				Model:             "gpt-4o",
				MaxTokens:         8192,
				Temperature:       0.7,
				MaxToolIterations: 20,
			},
		},
		Providers: ProvidersConfig{
			LocalWhisper: LocalWhisperConfig{
				Enabled:    true,
				Model:      "base",
				BinaryPath: "/opt/homebrew/bin/whisper",
			},
		},
		Gateway: GatewayConfig{
			Host:           "127.0.0.1", // Secure default
			Port:           18790,
			DashboardPort:  18791,
			RateLimitRPS:   5,            // 5 req/sec per client IP
			RateLimitBurst: 10,           // allow short bursts
			MaxBodyBytes:   10 << 20,     // 10 MiB
			ShutdownTimeout: 10 * time.Second, // graceful drain
		},
		Tools: ToolsConfig{
			Exec: ExecToolConfig{
				Timeout:             60 * time.Second,
				RestrictToWorkspace: true, // Secure default
			},
			Web: WebToolConfig{
				Search: SearchConfig{
					MaxResults: 10,
				},
			},
		},
	}
}
