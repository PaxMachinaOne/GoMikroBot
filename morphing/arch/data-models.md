# Go Nanobot – Data Models

> Architecture Agent – Data Model Definitions
> Per AGENTS.md Section 3.3

---

## Message Types

### InboundMessage
```go
// Messages from channels → agent
type InboundMessage struct {
    Channel   string            `json:"channel"`
    SenderID  string            `json:"sender_id"`
    ChatID    string            `json:"chat_id"`
    Content   string            `json:"content"`
    Media     []string          `json:"media,omitempty"`
    Metadata  map[string]any    `json:"metadata,omitempty"`
    Timestamp time.Time         `json:"timestamp"`
}
```

### OutboundMessage
```go
// Messages from agent → channels
type OutboundMessage struct {
    Channel string `json:"channel"`
    ChatID  string `json:"chat_id"`
    Content string `json:"content"`
}
```

---

## LLM Types

### Message (LLM format)
```go
type Message struct {
    Role       string     `json:"role"`  // system, user, assistant, tool
    Content    string     `json:"content"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
}
```

### ToolCall
```go
type ToolCall struct {
    ID        string         `json:"id"`
    Name      string         `json:"name"`
    Arguments map[string]any `json:"arguments"`
}
```

### ToolDefinition
```go
type ToolDefinition struct {
    Type     string       `json:"type"`  // "function"
    Function FunctionDef  `json:"function"`
}

type FunctionDef struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  map[string]any `json:"parameters"`  // JSON Schema
}
```

### LLMResponse
```go
type LLMResponse struct {
    Content      string
    ToolCalls    []ToolCall
    FinishReason string
    Usage        Usage
}

type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

---

## Session Types

### Session
```go
type Session struct {
    Key       string    `json:"key"`
    Messages  []Message `json:"messages"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Metadata  map[string]any `json:"metadata,omitempty"`
}

func NewSession(key string) *Session {
    now := time.Now()
    return &Session{
        Key:       key,
        Messages:  []Message{},
        CreatedAt: now,
        UpdatedAt: now,
        Metadata:  map[string]any{},
    }
}

func (s *Session) AddMessage(role, content string) {
    s.Messages = append(s.Messages, Message{Role: role, Content: content})
    s.UpdatedAt = time.Now()
}

func (s *Session) GetHistory(maxMessages int) []Message {
    if len(s.Messages) <= maxMessages {
        return s.Messages
    }
    return s.Messages[len(s.Messages)-maxMessages:]
}
```

---

## Cron Types

### CronJob
```go
type CronJob struct {
    ID             string       `json:"id"`
    Name           string       `json:"name"`
    Schedule       CronSchedule `json:"schedule"`
    Message        string       `json:"message"`
    Enabled        bool         `json:"enabled"`
    Deliver        bool         `json:"deliver"`
    Channel        string       `json:"channel,omitempty"`
    To             string       `json:"to,omitempty"`
    DeleteAfterRun bool         `json:"delete_after_run"`
    State          CronJobState `json:"state"`
}

type CronSchedule struct {
    Cron     string `json:"cron,omitempty"`      // "0 9 * * *"
    Every    int    `json:"every,omitempty"`     // seconds
    RunAt    int64  `json:"run_at,omitempty"`    // unix ms for one-shot
}

type CronJobState struct {
    NextRunMs   int64 `json:"next_run_ms"`
    LastRunMs   int64 `json:"last_run_ms"`
    RunCount    int   `json:"run_count"`
    LastResult  string `json:"last_result,omitempty"`
}
```

---

## Configuration Types

### Config (root)
```go
type Config struct {
    Agents    AgentsConfig    `json:"agents"`
    Channels  ChannelsConfig  `json:"channels"`
    Providers ProvidersConfig `json:"providers"`
    Gateway   GatewayConfig   `json:"gateway"`
    Tools     ToolsConfig     `json:"tools"`
}
```

### Nested configs
```go
type AgentsConfig struct {
    Defaults AgentDefaults `json:"defaults"`
}

type AgentDefaults struct {
    Workspace         string  `json:"workspace" envconfig:"WORKSPACE"`
    Model             string  `json:"model"`
    MaxTokens         int     `json:"maxTokens"`
    Temperature       float64 `json:"temperature"`
    MaxToolIterations int     `json:"maxToolIterations"`
}

type ChannelsConfig struct {
    Telegram TelegramConfig `json:"telegram"`
    Discord  DiscordConfig  `json:"discord"`
    WhatsApp WhatsAppConfig `json:"whatsapp"`
    Feishu   FeishuConfig   `json:"feishu"`
}

type TelegramConfig struct {
    Enabled   bool     `json:"enabled"`
    Token     string   `json:"token"`
    AllowFrom []string `json:"allowFrom"`
    Proxy     string   `json:"proxy,omitempty"`
}

type DiscordConfig struct {
    Enabled   bool     `json:"enabled"`
    Token     string   `json:"token"`
    AllowFrom []string `json:"allowFrom"`
}

type WhatsAppConfig struct {
    Enabled   bool     `json:"enabled"`
    BridgeURL string   `json:"bridgeUrl"`
    AllowFrom []string `json:"allowFrom"`
}

type FeishuConfig struct {
    Enabled           bool     `json:"enabled"`
    AppID             string   `json:"appId"`
    AppSecret         string   `json:"appSecret"`
    EncryptKey        string   `json:"encryptKey"`
    VerificationToken string   `json:"verificationToken"`
    AllowFrom         []string `json:"allowFrom"`
}

type ProvidersConfig struct {
    Anthropic  ProviderConfig `json:"anthropic"`
    OpenAI     ProviderConfig `json:"openai"`
    OpenRouter ProviderConfig `json:"openrouter"`
    DeepSeek   ProviderConfig `json:"deepseek"`
    Groq       ProviderConfig `json:"groq"`
    Gemini     ProviderConfig `json:"gemini"`
    VLLM       ProviderConfig `json:"vllm"`
}

type ProviderConfig struct {
    APIKey  string `json:"apiKey"`
    APIBase string `json:"apiBase,omitempty"`
}

type GatewayConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

type ToolsConfig struct {
    Exec ExecToolConfig `json:"exec"`
    Web  WebToolConfig  `json:"web"`
}

type ExecToolConfig struct {
    Timeout             int  `json:"timeout"`
    RestrictToWorkspace bool `json:"restrictToWorkspace"`
}

type WebToolConfig struct {
    Search SearchConfig `json:"search"`
}

type SearchConfig struct {
    APIKey     string `json:"apiKey"`
    MaxResults int    `json:"maxResults"`
}
```

---

## Default Values

```go
func DefaultConfig() *Config {
    return &Config{
        Agents: AgentsConfig{
            Defaults: AgentDefaults{
                Workspace:         "~/.nanobot/workspace",
                Model:             "anthropic/claude-sonnet-4-5",
                MaxTokens:         8192,
                Temperature:       0.7,
                MaxToolIterations: 20,
            },
        },
        Gateway: GatewayConfig{
            Host: "127.0.0.1",  // Secure default!
            Port: 18790,
        },
        Tools: ToolsConfig{
            Exec: ExecToolConfig{
                Timeout:             60,
                RestrictToWorkspace: true,  // Secure default!
            },
        },
    }
}
```
