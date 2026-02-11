#!/bin/bash
# SECURE-RUN.sh - Secure local execution of nanobot
# Usage: ./SECURE-RUN.sh [command]
# Examples:
#   ./SECURE-RUN.sh gateway     # Start the gateway securely
#   ./SECURE-RUN.sh agent       # Interactive agent mode
#   ./SECURE-RUN.sh status      # Check status

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="${HOME}/.nanobot"
CONFIG_FILE="${CONFIG_DIR}/config.json"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔒 Nanobot Secure Execution"
echo "==========================="

# --- Pre-flight Security Checks ---

check_passed=true

# 1. Check config file permissions
if [[ -f "$CONFIG_FILE" ]]; then
    perms=$(stat -f "%OLp" "$CONFIG_FILE" 2>/dev/null || stat -c "%a" "$CONFIG_FILE" 2>/dev/null)
    if [[ "$perms" != "600" ]]; then
        echo -e "${YELLOW}⚠️  Config file permissions are $perms (should be 600)${NC}"
        echo "   Fix with: chmod 600 $CONFIG_FILE"
        check_passed=false
    else
        echo -e "${GREEN}✓${NC} Config file permissions OK (600)"
    fi
else
    echo -e "${RED}✗ Config file not found at $CONFIG_FILE${NC}"
    echo "  Run: nanobot onboard"
    exit 1
fi

# 2. Check restrictToWorkspace setting
if grep -q '"restrictToWorkspace":\s*true' "$CONFIG_FILE" 2>/dev/null || \
   grep -q '"restrict_to_workspace":\s*true' "$CONFIG_FILE" 2>/dev/null; then
    echo -e "${GREEN}✓${NC} restrictToWorkspace is enabled"
else
    echo -e "${YELLOW}⚠️  restrictToWorkspace is NOT enabled${NC}"
    echo "   Add to config: \"tools\": { \"exec\": { \"restrictToWorkspace\": true } }"
    check_passed=false
fi

# 3. Check allowFrom lists
for channel in telegram discord whatsapp feishu; do
    if grep -q "\"$channel\"" "$CONFIG_FILE" 2>/dev/null; then
        if grep -A5 "\"$channel\"" "$CONFIG_FILE" | grep -q '"enabled":\s*true'; then
            if grep -A10 "\"$channel\"" "$CONFIG_FILE" | grep -q '"allowFrom":\s*\[\s*\]'; then
                echo -e "${YELLOW}⚠️  $channel channel has empty allowFrom (anyone can message!)${NC}"
                check_passed=false
            elif grep -A10 "\"$channel\"" "$CONFIG_FILE" | grep -q '"allowFrom"'; then
                echo -e "${GREEN}✓${NC} $channel channel has allowFrom configured"
            fi
        fi
    fi
done

# 4. Check gateway binding
if grep -q '"host":\s*"0.0.0.0"' "$CONFIG_FILE" 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Gateway binds to 0.0.0.0 (accessible from network)${NC}"
    echo "   Consider: \"gateway\": { \"host\": \"127.0.0.1\" }"
    check_passed=false
elif grep -q '"host":\s*"127.0.0.1"' "$CONFIG_FILE" 2>/dev/null; then
    echo -e "${GREEN}✓${NC} Gateway bound to localhost only"
fi

echo ""

# --- Execution Decision ---

if [[ "$check_passed" == "false" ]]; then
    echo -e "${YELLOW}Security warnings detected. Continue anyway? (y/N)${NC}"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Aborted. Please fix security issues first."
        echo "See: $SCRIPT_DIR/HARDENING-GUIDE.md"
        exit 1
    fi
fi

# --- Run nanobot ---

COMMAND="${1:-status}"

echo "Starting nanobot with command: $COMMAND"
echo ""

# Set secure umask for any files created
umask 077

# Export security-related env vars if not already set
export NANOBOT_GATEWAY__HOST="${NANOBOT_GATEWAY__HOST:-127.0.0.1}"

# Run nanobot
cd "$PROJECT_DIR"

if command -v nanobot &> /dev/null; then
    exec nanobot "$COMMAND" "${@:2}"
elif [[ -f "$PROJECT_DIR/pyproject.toml" ]]; then
    exec python -m nanobot "$COMMAND" "${@:2}"
else
    echo -e "${RED}Error: nanobot not found. Install with: pip install -e .${NC}"
    exit 1
fi
