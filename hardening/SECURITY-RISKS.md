# Nanobot Security Risk Assessment

> Security analysis performed on 2026-02-06

## Executive Summary

Nanobot is an AI assistant framework with shell execution, filesystem access, and multi-channel messaging (Telegram, Discord, WhatsApp, Feishu). This document identifies security risks and provides hardening recommendations.

---

## 🔴 Critical Risks

### 1. Shell Command Execution (`nanobot/agent/tools/shell.py`)

**Risk Level**: CRITICAL

The `exec` tool uses `asyncio.create_subprocess_shell()` which executes arbitrary commands:

```python
process = await asyncio.create_subprocess_shell(
    command,
    stdout=asyncio.subprocess.PIPE,
    stderr=asyncio.subprocess.PIPE,
    cwd=cwd,
)
```

**Existing Mitigations**:
- Deny pattern blocklist for dangerous commands (`rm -rf`, `dd`, `format`, fork bombs)
- Optional `restrict_to_workspace` flag (disabled by default)
- Timeout protection (60s default)

**Gaps**:
- Blocklist is easily bypassed (encoding, aliases, symlinks)
- `restrict_to_workspace` is **OFF by default**
- No allowlist mode for production use

---

### 2. Unrestricted Filesystem Access (`nanobot/agent/tools/filesystem.py`)

**Risk Level**: CRITICAL

The `read_file`, `write_file`, `edit_file`, and `list_dir` tools have **no path restrictions**:

```python
async def execute(self, path: str, content: str, **kwargs: Any) -> str:
    file_path = Path(path).expanduser()
    file_path.parent.mkdir(parents=True, exist_ok=True)
    file_path.write_text(content, encoding="utf-8")
```

**Impact**: An AI agent can read/write ANY file the process has access to, including:
- `~/.ssh/` keys
- `~/.nanobot/config.json` (contains API keys!)
- System configuration files
- Source code and databases

---

### 3. Plaintext Credential Storage (`~/.nanobot/config.json`)

**Risk Level**: HIGH

All API keys and tokens are stored in plaintext JSON:

```json
{
  "providers": { "openrouter": { "apiKey": "sk-or-v1-xxx" } },
  "channels": { "telegram": { "token": "123456:ABC..." } }
}
```

**Impact**: File readable by any process running as the user. No encryption at rest.

---

## 🟠 High Risks

### 4. Unauthenticated WebSocket Bridge (`bridge/src/server.ts`)

**Risk Level**: HIGH

The Node.js WhatsApp bridge accepts ANY WebSocket connection without authentication:

```typescript
this.wss.on('connection', (ws) => {
    console.log('🔗 Python client connected');
    this.clients.add(ws);
    // No authentication check!
});
```

**Impact**: Any local process can connect and send WhatsApp messages.

---

### 5. Empty Default Allowlists

**Risk Level**: HIGH

All channel allowlists (`allow_from`) default to empty arrays, meaning **anyone can message the bot**:

```python
class TelegramConfig(BaseModel):
    allow_from: list[str] = Field(default_factory=list)  # Empty = allow everyone
```

From `channels/base.py`:
```python
def is_allowed(self, sender_id: str) -> bool:
    allow_list = getattr(self.config, "allow_from", [])
    if not allow_list:
        return True  # ⚠️ Allows everyone!
```

---

## 🟡 Medium Risks

### 6. Gateway Binds to 0.0.0.0

The gateway server binds to all interfaces by default:

```python
class GatewayConfig(BaseModel):
    host: str = "0.0.0.0"
    port: int = 18790
```

**Impact**: Accessible from the network, not just localhost.

---

### 7. Limited Test Coverage

Only 2 test files exist:
- `tests/test_tool_validation.py` (89 lines)
- `tests/test_docker.sh` (shell script)

No security-focused tests, no fuzzing, no integration tests for channel authentication.

---

### 8. Subagent Task Execution

Subagents can spawn additional background tasks with full tool access:

```python
# From nanobot/agent/tools/spawn.py
async def execute(self, task: str, label: str | None = None, **kwargs: Any) -> str:
    """Spawn a subagent to execute the given task."""
```

**Impact**: Recursive task spawning could bypass rate limits or user oversight.

---

## 📊 Risk Summary Table

| Risk | Severity | Default Exposure | Mitigation Available |
|------|----------|------------------|---------------------|
| Shell execution | 🔴 Critical | Yes | Partial (blocklist) |
| Filesystem access | 🔴 Critical | Yes | None |
| Plaintext credentials | 🔴 Critical | Yes | None |
| Unauthenticated bridge | 🟠 High | Local only | None |
| Empty allowlists | 🟠 High | Yes | Config option |
| 0.0.0.0 binding | 🟡 Medium | Yes | Config option |
| Limited testing | 🟡 Medium | N/A | N/A |
| Subagent spawning | 🟡 Medium | Yes | None |

---

## Next Steps

See [HARDENING-GUIDE.md](./HARDENING-GUIDE.md) for recommended mitigations.
