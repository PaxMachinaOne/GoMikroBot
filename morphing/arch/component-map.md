# Go Nanobot – Package Design

> Architecture Agent – Component Map
> Per AGENTS.md Section 3.3

---

## Package Structure

```
gonanobot/
├── cmd/
│   └── nanobot/
│       └── main.go              # CLI entrypoint
├── internal/
│   ├── agent/
│   │   ├── loop.go              # Core agent loop
│   │   ├── context.go           # Prompt builder
│   │   ├── memory.go            # Persistent memory
│   │   └── subagent.go          # Background tasks
│   ├── bus/
│   │   ├── bus.go               # Message bus
│   │   └── events.go            # Message types
│   ├── channels/
│   │   ├── channel.go           # Base interface
│   │   ├── telegram.go          # Telegram impl
│   │   ├── discord.go           # Discord impl
│   │   ├── whatsapp.go          # WhatsApp bridge client
│   │   ├── feishu.go            # Feishu impl
│   │   └── manager.go           # Channel orchestrator
│   ├── config/
│   │   ├── config.go            # Configuration struct
│   │   └── loader.go            # Load/save config
│   ├── cron/
│   │   ├── service.go           # Cron scheduler
│   │   └── types.go             # Job definitions
│   ├── provider/
│   │   ├── provider.go          # LLM interface
│   │   ├── openai.go            # OpenAI-compatible client
│   │   └── response.go          # Response types
│   ├── session/
│   │   ├── session.go           # Session struct
│   │   └── manager.go           # Session persistence
│   ├── tools/
│   │   ├── tool.go              # Tool interface
│   │   ├── registry.go          # Tool registration
│   │   ├── filesystem.go        # File tools
│   │   ├── shell.go             # Exec tool
│   │   ├── web.go               # Search/fetch
│   │   ├── spawn.go             # Subagent spawn
│   │   ├── cron.go              # Cron tool
│   │   └── message.go           # Message tool
│   └── util/
│       ├── helpers.go           # Utility functions
│       └── safepath.go          # Path sanitization
├── pkg/
│   └── api/                     # Public API (if needed)
├── go.mod
├── go.sum
└── Makefile
```

---

## Core Interfaces

### LLM Provider
```go
// internal/provider/provider.go
type LLMProvider interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    DefaultModel() string
}

type ChatRequest struct {
    Messages    []Message
    Tools       []ToolDefinition
    Model       string
    MaxTokens   int
    Temperature float64
}

type ChatResponse struct {
    Content      string
    ToolCalls    []ToolCall
    FinishReason string
    Usage        Usage
}
```

### Tool
```go
// internal/tools/tool.go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]any  // JSON Schema
    Execute(ctx context.Context, params map[string]any) (string, error)
}
```

### Channel
```go
// internal/channels/channel.go
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop() error
    Send(msg *OutboundMessage) error
    IsAllowed(senderID string) bool
}
```

### Message Bus
```go
// internal/bus/bus.go
type MessageBus struct {
    inbound  chan *InboundMessage
    outbound chan *OutboundMessage
    subs     map[string][]func(*OutboundMessage)
}
```

---

## Concurrency Model

```
┌──────────────┐     inbound chan      ┌─────────────┐
│   Channels   │ ──────────────────▶   │  AgentLoop  │
│  (goroutines)│                       │ (goroutine) │
└──────────────┘                       └─────────────┘
       ▲                                      │
       │                                      │
       │           outbound chan              │
       └──────────────────────────────────────┘

Each channel runs in its own goroutine.
AgentLoop runs in main goroutine.
Context propagation for graceful shutdown.
```

---

## Error Strategy

1. **Tools return (string, error)** – never panic
2. **Wrap errors with context**: `fmt.Errorf("loading config: %w", err)`
3. **Use custom error types** for actionable errors:
   ```go
   type ConfigError struct {
       Path string
       Err  error
   }
   ```
4. **Log errors at boundaries** (CLI, channel handlers)

---

## Configuration Loading

```go
// Priority (highest to lowest):
// 1. Environment variables (NANOBOT_*)
// 2. Config file (~/.nanobot/config.json)
// 3. Defaults

func LoadConfig() (*Config, error) {
    cfg := DefaultConfig()
    
    // Load from file
    if data, err := os.ReadFile(configPath); err == nil {
        json.Unmarshal(data, cfg)
    }
    
    // Override from env
    envconfig.Process("NANOBOT", cfg)
    
    return cfg, nil
}
```

---

## Dependency Graph

```
cmd/nanobot
    └── internal/config
    └── internal/bus
    └── internal/provider
    └── internal/agent
            └── internal/tools
            └── internal/session
    └── internal/channels
            └── internal/bus
    └── internal/cron
```

---

## External Dependencies

| Purpose | Library | Justification |
|---------|---------|---------------|
| CLI | `spf13/cobra` | Standard, feature-rich |
| HTTP client | `net/http` | Stdlib sufficient |
| WebSocket | `nhooyr.io/websocket` | Better than gorilla |
| JSON | `encoding/json` | Stdlib |
| Cron | `robfig/cron/v3` | Battle-tested |
| Logging | `log/slog` | Go 1.21+ stdlib |
| Config | `kelseyhightower/envconfig` | Simple env binding |
| Validation | `go-playground/validator` | Struct validation |
| Console | `fatih/color` | Colored output |
