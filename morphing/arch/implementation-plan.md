# Go Nanobot – Implementation Plan

> Per AGENTS.md Migration Workflow

---

## Overview

Rewrite nanobot from Python to Go in 8 migration phases:

| Phase | Component | Effort | Priority |
|-------|-----------|--------|----------|
| 1 | Core types & config | 1 day | 🔴 Critical |
| 2 | LLM Provider | 1 day | 🔴 Critical |
| 3 | Tool framework | 2 days | 🔴 Critical |
| 4 | Session manager | 0.5 day | 🟠 High |
| 5 | Message bus | 0.5 day | 🟠 High |
| 6 | Agent loop | 2 days | 🔴 Critical |
| 7 | Channels | 2 days | 🟠 High |
| 8 | CLI & Cron | 1 day | 🟡 Medium |

---

## Verification Strategy

### Parity Tests

For each module:
1. Create test cases from Python behavior
2. Capture Python output for reference
3. Run same input through Go implementation
4. Assert output matches

### Unit Tests

Each Go package includes `*_test.go`:
- Table-driven tests
- Interface mocks for dependencies
- Edge cases from Python error handling

### Integration Tests

End-to-end tests:
- CLI command execution
- Agent chat completion
- Tool execution chain

---

## Proposed Changes

### Phase 1: Core Types & Config

#### [NEW] gonanobot/go.mod
Initialize Go module with dependencies.

#### [NEW] gonanobot/internal/config/config.go
- Config struct with JSON tags
- Default values
- Validation

#### [NEW] gonanobot/internal/config/loader.go
- Load from `~/.nanobot/config.json`
- Environment variable overrides
- camelCase parsing

**Validation:**
```bash
cd gonanobot && go test ./internal/config/...
```

---

### Phase 2: LLM Provider

#### [NEW] gonanobot/internal/provider/provider.go
- LLMProvider interface
- ChatRequest/ChatResponse types

#### [NEW] gonanobot/internal/provider/openai.go
- OpenAI-compatible client
- Model prefix routing (openrouter/, gemini/, etc.)
- Tool call parsing

**Validation:**
```bash
go test ./internal/provider/...
# Integration: OPENROUTER_API_KEY=... go test -tags=integration ./internal/provider/...
```

---

### Phase 3: Tool Framework

#### [NEW] gonanobot/internal/tools/tool.go
- Tool interface
- Parameter validation

#### [NEW] gonanobot/internal/tools/registry.go
- Tool registration
- Execute by name

#### [NEW] gonanobot/internal/tools/filesystem.go
- read_file, write_file, edit_file, list_dir

#### [NEW] gonanobot/internal/tools/shell.go
- exec tool with safety guards
- Deny patterns (rm -rf, dd, etc.)

#### [NEW] gonanobot/internal/tools/web.go
- web_search (Brave API)
- web_fetch (HTTP GET)

**Validation:**
```bash
go test ./internal/tools/...
```

---

### Phase 4: Session Manager

#### [NEW] gonanobot/internal/session/session.go
- Session struct
- Message history

#### [NEW] gonanobot/internal/session/manager.go
- JSONL persistence
- In-memory cache

**Validation:**
```bash
go test ./internal/session/...
```

---

### Phase 5: Message Bus

#### [NEW] gonanobot/internal/bus/events.go
- InboundMessage, OutboundMessage

#### [NEW] gonanobot/internal/bus/bus.go
- Channel-based pub/sub
- Dispatcher goroutine

**Validation:**
```bash
go test ./internal/bus/...
```

---

### Phase 6: Agent Loop

#### [NEW] gonanobot/internal/agent/loop.go
- Main processing loop
- Tool execution cycle
- Context building

#### [NEW] gonanobot/internal/agent/context.go
- System prompt
- Skills loading
- Memory integration

**Validation:**
```bash
go test ./internal/agent/...
# E2E: Run agent with mock provider
```

---

### Phase 7: Channels

#### [NEW] gonanobot/internal/channels/channel.go
- Channel interface
- Base helper functions

#### [NEW] gonanobot/internal/channels/telegram.go
- Long polling client
- Message formatting

#### [NEW] gonanobot/internal/channels/discord.go
- WebSocket gateway
- Intent handling

#### [NEW] gonanobot/internal/channels/manager.go
- Multi-channel orchestration

**Validation:**
```bash
go test ./internal/channels/...
# Manual: Configure token, run gateway, send message
```

---

### Phase 8: CLI & Cron

#### [NEW] gonanobot/cmd/nanobot/main.go
- Cobra CLI setup

#### [NEW] gonanobot/cmd/nanobot/cmd/root.go
- onboard, status commands

#### [NEW] gonanobot/cmd/nanobot/cmd/agent.go
- agent command

#### [NEW] gonanobot/cmd/nanobot/cmd/gateway.go
- gateway command

#### [NEW] gonanobot/cmd/nanobot/cmd/cron.go
- cron subcommands

#### [NEW] gonanobot/internal/cron/service.go
- Job scheduler
- Timer management

**Validation:**
```bash
go build -o nanobot ./cmd/nanobot
./nanobot --help
./nanobot status
```

---

## Manual Verification (User)

After implementation:

1. **Config parity:**
   ```bash
   # Python creates config
   python -m nanobot onboard
   # Go reads same config
   ./gonanobot status
   ```

2. **Agent chat:**
   ```bash
   ./gonanobot agent -m "What is 2+2?"
   # Should return "4" via LLM
   ```

3. **Tool execution:**
   ```bash
   ./gonanobot agent -m "List files in current directory"
   # Should invoke list_dir tool
   ```

---

## Rollback Path

Each phase is independent. If Go component fails:
- Continue using Python for that component
- Hybrid operation possible (e.g., Go CLI, Python channels)

---

## Reference Files Created

| Artifact | Path |
|----------|------|
| Requirements | `morphing/requirements/functional-requirements.md` |
| Concept Mapping | `morphing/arch/python-to-go-mapping.md` |
| Package Design | `morphing/arch/component-map.md` |
| Data Models | `morphing/arch/data-models.md` |
| Interfaces | `morphing/arch/interfaces.md` |
| This Plan | `morphing/arch/implementation-plan.md` |
| Tasks | `gonanobot/tasks.md` |
