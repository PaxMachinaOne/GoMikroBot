# Go Nanobot – Interface Definitions

> Architecture Agent – Interface Contracts
> Per AGENTS.md Section 3.3

---

## Core Interfaces

### LLMProvider
```go
// internal/provider/provider.go

// LLMProvider abstracts LLM API calls
type LLMProvider interface {
    // Chat sends a completion request and returns the response
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    
    // DefaultModel returns the configured default model
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
    FinishReason string  // "stop", "tool_calls", "error"
    Usage        Usage
}
```

---

### Tool
```go
// internal/tools/tool.go

// Tool is the interface all agent tools must implement
type Tool interface {
    // Name returns the tool identifier (used in function calls)
    Name() string
    
    // Description returns a human-readable description for the LLM
    Description() string
    
    // Parameters returns the JSON Schema for tool parameters
    Parameters() map[string]any
    
    // Execute runs the tool with validated parameters
    // Returns result string and error (error → return error message string)
    Execute(ctx context.Context, params map[string]any) (string, error)
}

// ValidatableParams is optional interface for parameter validation
type ValidatableParams interface {
    Validate(params map[string]any) []string  // Returns validation errors
}
```

---

### ToolRegistry
```go
// internal/tools/registry.go

// ToolRegistry manages tool registration and execution
type ToolRegistry interface {
    // Register adds a tool to the registry
    Register(tool Tool)
    
    // Get returns a tool by name
    Get(name string) (Tool, bool)
    
    // List returns all registered tools
    List() []Tool
    
    // Definitions returns tool definitions for LLM
    Definitions() []ToolDefinition
    
    // Execute runs a tool by name with given parameters
    Execute(ctx context.Context, name string, params map[string]any) (string, error)
}
```

---

### Channel
```go
// internal/channels/channel.go

// Channel represents a chat platform integration
type Channel interface {
    // Name returns the channel identifier (e.g., "telegram")
    Name() string
    
    // Start begins listening for messages (blocking)
    Start(ctx context.Context) error
    
    // Stop gracefully shuts down the channel
    Stop() error
    
    // Send delivers a message to the channel
    Send(msg *OutboundMessage) error
    
    // IsAllowed checks if a sender is permitted
    IsAllowed(senderID string) bool
}
```

---

### MessageBus
```go
// internal/bus/bus.go

// MessageBus decouples channels from the agent core
type MessageBus interface {
    // PublishInbound queues a message from a channel
    PublishInbound(msg *InboundMessage)
    
    // ConsumeInbound blocks until a message is available
    ConsumeInbound(ctx context.Context) (*InboundMessage, error)
    
    // PublishOutbound queues a message for channels
    PublishOutbound(msg *OutboundMessage)
    
    // Subscribe registers a callback for outbound messages
    Subscribe(channel string, callback func(*OutboundMessage))
    
    // DispatchOutbound runs the outbound dispatcher (blocking)
    DispatchOutbound(ctx context.Context) error
    
    // Stop signals shutdown
    Stop()
}
```

---

### SessionManager
```go
// internal/session/manager.go

// SessionManager handles conversation persistence
type SessionManager interface {
    // GetOrCreate returns existing session or creates new
    GetOrCreate(key string) *Session
    
    // Save persists a session to disk
    Save(session *Session) error
    
    // Delete removes a session
    Delete(key string) bool
    
    // List returns all session metadata
    List() []SessionInfo
}

type SessionInfo struct {
    Key       string
    CreatedAt time.Time
    UpdatedAt time.Time
    Path      string
}
```

---

### CronService
```go
// internal/cron/service.go

// CronService manages scheduled agent tasks
type CronService interface {
    // Start begins the scheduler
    Start(ctx context.Context) error
    
    // Stop halts the scheduler
    Stop()
    
    // AddJob creates a new scheduled job
    AddJob(job *CronJob) error
    
    // RemoveJob deletes a job by ID
    RemoveJob(id string) bool
    
    // ListJobs returns all jobs
    ListJobs(includeDisabled bool) []*CronJob
    
    // EnableJob toggles job enabled state
    EnableJob(id string, enabled bool) error
    
    // RunJob manually triggers a job
    RunJob(id string, force bool) error
    
    // Status returns service status
    Status() *CronStatus
}

type CronStatus struct {
    Running  bool
    JobCount int
    NextWake time.Time
}
```

---

### AgentLoop
```go
// internal/agent/loop.go

// AgentLoop is the core processing engine
type AgentLoop interface {
    // Run starts processing messages from the bus (blocking)
    Run(ctx context.Context) error
    
    // Stop signals graceful shutdown
    Stop()
    
    // ProcessDirect handles a single message synchronously (for CLI)
    ProcessDirect(ctx context.Context, content, sessionKey string) (string, error)
}
```

---

## Factory Functions

Each interface has a corresponding constructor:

```go
// Providers
func NewOpenAIProvider(apiKey, apiBase, model string) LLMProvider

// Tools
func NewToolRegistry() ToolRegistry
func NewReadFileTool() Tool
func NewWriteFileTool() Tool
func NewExecTool(config ExecToolConfig, workDir string) Tool

// Channels
func NewTelegramChannel(config TelegramConfig, bus MessageBus) Channel
func NewDiscordChannel(config DiscordConfig, bus MessageBus) Channel

// Core
func NewMessageBus() MessageBus
func NewSessionManager(workspace string) SessionManager
func NewCronService(storePath string, onJob func(*CronJob) string) CronService
func NewAgentLoop(opts AgentLoopOptions) AgentLoop

type AgentLoopOptions struct {
    Bus           MessageBus
    Provider      LLMProvider
    Workspace     string
    Model         string
    MaxIterations int
    ExecConfig    ExecToolConfig
    CronService   CronService
}
```

---

## Interface Segregation

Small, focused interfaces enable:
- Easy testing with mocks
- Swappable implementations
- Clear dependency boundaries
