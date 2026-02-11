# Nanobot Hardening Guide

> Practical steps to secure your nanobot deployment

---

## 🔒 Essential Configuration (Do These First!)

### 1. Restrict Shell Commands to Workspace

Edit `~/.nanobot/config.json`:

```json
{
  "tools": {
    "exec": {
      "timeout": 30,
      "restrictToWorkspace": true
    }
  }
}
```

This blocks commands that reference paths outside the workspace directory.

---

### 2. Configure Channel Allowlists

**Never run with empty allowlists in production!**

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allowFrom": ["123456789", "your_username"]
    },
    "discord": {
      "enabled": true,
      "token": "YOUR_DISCORD_TOKEN",
      "allowFrom": ["YOUR_DISCORD_USER_ID"]
    },
    "whatsapp": {
      "enabled": true,
      "allowFrom": ["+1234567890"]
    }
  }
}
```

Get your Telegram user ID from `@userinfobot`. Get your Discord user ID by enabling Developer Mode.

---

### 3. Bind Gateway to Localhost Only

```json
{
  "gateway": {
    "host": "127.0.0.1",
    "port": 18790
  }
}
```

---

### 4. Protect Your Config File

```bash
chmod 600 ~/.nanobot/config.json
```

---

## 🐳 Docker Secure Execution

### Recommended Docker Run Command

```bash
docker run -d \
  --name nanobot \
  --read-only \
  --tmpfs /tmp:noexec,nosuid,size=100m \
  --security-opt no-new-privileges \
  --cap-drop ALL \
  -v ~/.nanobot:/root/.nanobot:ro \
  -v ~/.nanobot/workspace:/root/.nanobot/workspace:rw \
  -p 127.0.0.1:18790:18790 \
  nanobot gateway
```

**Key flags**:
- `--read-only`: Filesystem is read-only except explicit mounts
- `--tmpfs /tmp:noexec`: Temp files can't be executed
- `--no-new-privileges`: Prevents privilege escalation
- `--cap-drop ALL`: Drops all Linux capabilities
- `-p 127.0.0.1:...`: Binds only to localhost

### Docker Compose (Secure)

Create `docker-compose.secure.yml`:

```yaml
version: '3.8'
services:
  nanobot:
    build: .
    read_only: true
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    tmpfs:
      - /tmp:noexec,nosuid,size=100m
    volumes:
      - ~/.nanobot:/root/.nanobot:ro
      - ~/.nanobot/workspace:/root/.nanobot/workspace:rw
    ports:
      - "127.0.0.1:18790:18790"
    command: gateway
```

Run with:
```bash
docker-compose -f docker-compose.secure.yml up -d
```

---

## 🔐 Advanced Hardening

### Use Environment Variables for Secrets

Instead of storing keys in config.json:

```bash
export NANOBOT_PROVIDERS__OPENROUTER__API_KEY="sk-or-v1-xxx"
export NANOBOT_CHANNELS__TELEGRAM__TOKEN="123456:ABC..."
```

Nanobot uses `pydantic-settings` which reads `NANOBOT_` prefixed env vars.

### Run as Unprivileged User

```bash
# Create dedicated user
sudo useradd -r -s /bin/false nanobot
sudo mkdir -p /home/nanobot/.nanobot
sudo chown -R nanobot:nanobot /home/nanobot

# Run as that user
sudo -u nanobot nanobot gateway
```

### Firewall Rules (UFW)

```bash
# Block external access to gateway
sudo ufw deny 18790

# If needed externally, allow specific IPs only
sudo ufw allow from 192.168.1.100 to any port 18790
```

---

## 📋 Security Checklist

```
[ ] restrictToWorkspace: true in config
[ ] Non-empty allowFrom for all enabled channels
[ ] Gateway bound to 127.0.0.1
[ ] Config file permissions set to 600
[ ] Running in Docker with security flags
[ ] Secrets in environment variables (not config file)
[ ] Running as non-root/unprivileged user
[ ] Firewall rules blocking external gateway access
[ ] Regular review of ~/.nanobot/workspace contents
[ ] Monitoring agent tool usage (check logs)
```

---

## 🚨 Incident Response

If you suspect compromise:

1. **Stop immediately**: `docker stop nanobot` or `pkill -f nanobot`
2. **Rotate all API keys** in OpenRouter, Telegram, etc.
3. **Check ~/.nanobot/workspace** for unexpected files
4. **Review logs**: `~/.nanobot/logs/` if enabled
5. **Check command history**: Look for suspicious shell commands

---

## 📚 References

- [SECURITY-RISKS.md](./SECURITY-RISKS.md) - Full risk assessment
- [SECURE-RUN.sh](./SECURE-RUN.sh) - Ready-to-use secure startup script
