# Python to Go Concept Mapping

> Architecture Agent – Mapping Document
> Per AGENTS.md Section 3.3

---

## Type System Mapping

| Python | Go Equivalent |
|--------|---------------|
| `str` | `string` |
| `int` | `int64` |
| `float` | `float64` |
| `bool` | `bool` |
| `None` | `nil` (pointers), zero values |
| `list[T]` | `[]T` |
| `dict[K, V]` | `map[K]V` |
| `tuple` | Struct or multiple returns |
| `Any` | `interface{}` or generics |
| `Optional[T]` | `*T` (pointer) |

---

## Class → Struct + Interface

### Python Pattern
```python
class AgentLoop:
    def __init__(self, bus, provider, workspace):
        self.bus = bus
        self.provider = provider
        
    async def run(self):
        ...
```

### Go Equivalent
```go
type AgentLoop struct {
    bus      *MessageBus
    provider LLMProvider  // interface
    workspace string
}

func NewAgentLoop(bus *MessageBus, provider LLMProvider, workspace string) *AgentLoop {
    return &AgentLoop{bus: bus, provider: provider, workspace: workspace}
}

func (a *AgentLoop) Run(ctx context.Context) error {
    ...
}
```

---

## Async → Goroutines + Channels

| Python | Go |
|--------|-----|
| `async def` | Regular func with goroutine |
| `await` | Channel receive `<-ch` |
| `asyncio.Queue` | `chan T` |
| `asyncio.gather` | `sync.WaitGroup` |
| `asyncio.create_task` | `go func()` |
| `asyncio.wait_for(timeout)` | `select` with `time.After` |

### Python Pattern
```python
async def consume_inbound(self):
    return await self.inbound.get()
```

### Go Equivalent
```go
func (b *MessageBus) ConsumeInbound(ctx context.Context) (*InboundMessage, error) {
    select {
    case msg := <-b.inbound:
        return msg, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

---

## Exception Handling → Error Returns

| Python | Go |
|--------|-----|
| `try: ... except:` | `if err != nil` |
| `raise Exception` | `return fmt.Errorf(...)` |
| Multiple except | Type switch on error |

### Python Pattern
```python
try:
    content = file_path.read_text()
    return content
except PermissionError:
    return f"Error: Permission denied"
except Exception as e:
    return f"Error: {str(e)}"
```

### Go Equivalent
```go
content, err := os.ReadFile(filePath)
if err != nil {
    if os.IsPermission(err) {
        return "", fmt.Errorf("permission denied: %s", filePath)
    }
    return "", fmt.Errorf("error reading file: %w", err)
}
return string(content), nil
```

---

## Dataclass → Struct

### Python
```python
@dataclass
class Session:
    key: str
    messages: list[dict] = field(default_factory=list)
    created_at: datetime = field(default_factory=datetime.now)
```

### Go
```go
type Session struct {
    Key       string         `json:"key"`
    Messages  []Message      `json:"messages"`
    CreatedAt time.Time      `json:"created_at"`
}

func NewSession(key string) *Session {
    return &Session{
        Key:       key,
        Messages:  []Message{},
        CreatedAt: time.Now(),
    }
}
```

---

## ABC/Interface → Go Interface

### Python
```python
class LLMProvider(ABC):
    @abstractmethod
    async def chat(self, messages, tools, model) -> LLMResponse:
        pass
```

### Go
```go
type LLMProvider interface {
    Chat(ctx context.Context, messages []Message, tools []Tool, model string) (*LLMResponse, error)
}
```

---

## Dependency Injection

### Python (runtime attributes)
```python
class AgentLoop:
    def __init__(self, provider: LLMProvider):
        self.provider = provider
```

### Go (constructor injection)
```go
func NewAgentLoop(provider LLMProvider) *AgentLoop {
    return &AgentLoop{provider: provider}
}
```

---

## Pydantic → Struct + Validation

| Python Pydantic | Go Approach |
|-----------------|-------------|
| Field validation | Custom validation method |
| JSON serialization | `encoding/json` with struct tags |
| Environment override | `envconfig` or `viper` library |

---

## Key Library Mappings

| Python Library | Go Equivalent |
|----------------|---------------|
| `typer` | `cobra` or `urfave/cli` |
| `litellm` | Direct OpenAI SDK or custom client |
| `loguru` | `log/slog` or `zerolog` |
| `pydantic` | Struct + `go-playground/validator` |
| `httpx` | `net/http` |
| `websockets` | `gorilla/websocket` or `nhooyr.io/websocket` |
| `croniter` | `robfig/cron` |
| `rich` | `fatih/color` + `olekukonko/tablewriter` |

---

## Context Propagation

Python lacks explicit context. Go uses `context.Context`:

```go
func (a *AgentLoop) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-a.bus.Inbound():
            a.processMessage(ctx, msg)
        }
    }
}
```

---

## File Handling

| Python | Go |
|--------|-----|
| `Path.home()` | `os.UserHomeDir()` |
| `path.exists()` | `os.Stat()` + error check |
| `path.read_text()` | `os.ReadFile()` |
| `path.write_text()` | `os.WriteFile()` |
| `path.mkdir(parents=True)` | `os.MkdirAll()` |
