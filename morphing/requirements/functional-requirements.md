# Python Nanobot – Functional Requirements

> Extracted by Requirements Tracker Agent
> Source: Python codebase analysis

---

## 1. Core Agent Behavior

### REQ-001 – LLM Chat Completion
**Type:** functional  
**Status:** to-preserve  

The agent must:
- Send messages to LLM providers via OpenAI-compatible API
- Support tool/function calling
- Parse tool call responses and execute tools
- Loop until LLM returns final response (max iterations limit)

**Python Implementation:** `agent/loop.py`, `providers/litellm_provider.py`

**Acceptance Criteria:**
- [ ] Supports Claude, GPT, Gemini, DeepSeek, Groq, vLLM models
- [ ] Tool calls are executed and results returned to LLM
- [ ] Configurable max iterations (default: 20)

---

### REQ-002 – Session/Conversation History
**Type:** functional  
**Status:** to-preserve  

The agent must:
- Maintain conversation history per session
- Persist sessions to disk (JSONL format)
- Support session listing, deletion, clearing

**Python Implementation:** `session/manager.py`

**Acceptance Criteria:**
- [ ] Sessions keyed by `channel:chat_id`
- [ ] History survives process restart
- [ ] Configurable max history length

---

### REQ-003 – Tool Execution Framework
**Type:** functional  
**Status:** to-preserve  

Built-in tools:
| Tool | Function |
|------|----------|
| `read_file` | Read file contents |
| `write_file` | Write/create files |
| `edit_file` | Replace text in file |
| `list_dir` | List directory contents |
| `exec` | Execute shell commands |
| `web_search` | Brave Search API |
| `web_fetch` | Fetch URL content |
| `spawn` | Background subagent tasks |
| `cron` | Schedule tasks |
| `message` | Send messages to channels |

**Python Implementation:** `agent/tools/*.py`

**Acceptance Criteria:**
- [ ] Tools register with parameter schema
- [ ] Parameter validation before execution
- [ ] Error handling returns error string (not exception)

---

## 2. Channel Integrations

### REQ-004 – Multi-Channel Messaging
**Type:** functional  
**Status:** to-preserve  

Supported channels:
| Channel | Protocol | Auth |
|---------|----------|------|
| Telegram | HTTP Long Polling | Bot token |
| Discord | WebSocket | Bot token |
| WhatsApp | WebSocket Bridge (Node.js) | QR code |
| Feishu | WebSocket | App ID/Secret |

**Python Implementation:** `channels/*.py`

**Acceptance Criteria:**
- [ ] Each channel has `start()`, `stop()`, `send()` methods
- [ ] Message allowlist per channel (`allow_from`)
- [ ] Messages forwarded to MessageBus

---

### REQ-005 – Message Bus
**Type:** functional  
**Status:** to-preserve  

Async pub/sub decoupling channels from agent:
- Inbound queue: channels → agent
- Outbound queue: agent → channels
- Subscriber pattern for outbound dispatch

**Python Implementation:** `bus/queue.py`

**Acceptance Criteria:**
- [ ] Async queues with blocking consume
- [ ] Channel subscription by name
- [ ] Dispatcher runs as background task

---

## 3. Scheduling

### REQ-006 – Cron Service
**Type:** functional  
**Status:** to-preserve  

Schedule agent tasks via:
- Cron expressions (`0 9 * * *`)
- Interval (every N seconds)
- One-shot (run once then delete)

**Python Implementation:** `cron/service.py`, `cron/types.py`

**Acceptance Criteria:**
- [ ] Jobs persist to JSON file
- [ ] Next run computed via croniter
- [ ] Jobs can deliver messages to channels

---

## 4. CLI Interface

### REQ-007 – Command Line Commands
**Type:** functional  
**Status:** to-preserve  

| Command | Function |
|---------|----------|
| `nanobot onboard` | Initialize config/workspace |
| `nanobot agent -m "..."` | Direct agent chat |
| `nanobot gateway` | Start gateway server |
| `nanobot status` | Show status |
| `nanobot cron list/add/remove/run` | Manage cron jobs |
| `nanobot channels status/login` | Manage channels |

**Python Implementation:** `cli/commands.py` (Typer)

---

## 5. Configuration

### REQ-008 – JSON Configuration
**Type:** functional  
**Status:** to-preserve  

Config file: `~/.nanobot/config.json`

Sections:
- `providers` – API keys per LLM provider
- `channels` – Channel tokens and allowlists
- `agents.defaults` – Model, workspace, limits
- `tools.exec` – Shell command restrictions
- `gateway` – Host/port binding

**Python Implementation:** `config/schema.py`, `config/loader.py`

**Acceptance Criteria:**
- [ ] Pydantic validation
- [ ] Environment variable override (NANOBOT_ prefix)
- [ ] camelCase ↔ snake_case conversion

---

## 6. Security Constraints

### REQ-009 – Shell Command Safety
**Type:** constraint  
**Status:** to-improve  

Current: Blocklist regex for dangerous patterns
Desired: Maintain blocklist, add optional allowlist mode

---

### REQ-010 – Channel Authorization
**Type:** constraint  
**Status:** to-preserve  

- Empty `allowFrom` = allow everyone (⚠️ insecure default)
- Non-empty = whitelist only

---

## 7. Performance Characteristics

### REQ-011 – Async Architecture
**Type:** non-functional  
**Status:** to-preserve  

All I/O is async:
- `asyncio` event loop
- `aiohttp`/`httpx` for HTTP
- `asyncio.Queue` for message bus

**Go Equivalent:** goroutines + channels

---

## 8. Behavior to Remove

### None identified

All current behaviors should be preserved or improved.

---

## 9. Implicit Contracts

| Contract | Location | Implication |
|----------|----------|-------------|
| Session key format | `channel:chat_id` | Must parse/format consistently |
| Tool response | Always string | Never raise, return error string |
| Config path | `~/.nanobot/` | Fixed location |
| JSONL sessions | Append metadata first | First line is metadata object |

---

## Traceability

Source: Python codebase `/Users/kamir/GITHUB.kamir/nanobot/nanobot/`
