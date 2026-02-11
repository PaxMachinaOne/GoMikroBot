package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kamir/gomikrobot/internal/provider"
	"github.com/kamir/gomikrobot/internal/session"
	"github.com/kamir/gomikrobot/internal/tools"
)

var bootstrapFiles = []string{
	"AGENTS.md",
	"SOUL.md",
	"USER.md",
	"TOOLS.md",
	"IDENTITY.md",
}

// ContextBuilder assembles the system prompt and messages.
type ContextBuilder struct {
	workspace string
	registry  *tools.Registry
}

// NewContextBuilder creates a new ContextBuilder.
func NewContextBuilder(workspace string, registry *tools.Registry) *ContextBuilder {
	return &ContextBuilder{
		workspace: workspace,
		registry:  registry,
	}
}

// BuildSystemPrompt constructs the full system prompt from files and runtime info.
func (b *ContextBuilder) BuildSystemPrompt() string {
	var parts []string

	// 1. Core Identity & Runtime Info
	parts = append(parts, b.getIdentity())

	// 2. Bootstrap Files
	if bootstrap := b.loadBootstrapFiles(); bootstrap != "" {
		parts = append(parts, bootstrap)
	}

	// 3. Memory
	if memory := b.loadMemory(); memory != "" {
		parts = append(parts, "# Memory\n\n"+memory)
	}

	// 4. Skills (Summary)
	if skills := b.buildSkillsSummary(); skills != "" {
		parts = append(parts, "# Skills\n\n"+skills)
	}

	return strings.Join(parts, "\n\n---\n\n")
}

func (b *ContextBuilder) getIdentity() string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")

	// Expand workspace path
	wsPath := b.workspace
	if strings.HasPrefix(wsPath, "~") {
		home, _ := os.UserHomeDir()
		wsPath = filepath.Join(home, wsPath[1:])
	}
	if abs, err := filepath.Abs(wsPath); err == nil {
		wsPath = abs
	}

	runtimeInfo := fmt.Sprintf("%s %s, Go %s", runtime.GOOS, runtime.GOARCH, runtime.Version())

	return fmt.Sprintf(`# GoMikroBot 🤖

You are GoMikroBot, a helpful, efficient AI assistant.
You have access to tools that allow you to:
- Read, write, and edit files
- Execute shell commands
- Search the web and fetch web pages
- Send messages to users

## Current Time
%s

## Runtime
%s

## Workspace
Your workspace is at: %s
- Memory files: %s/memory/MEMORY.md
- Daily notes: %s/memory/YYYY-MM-DD.md
- Custom skills: %s/skills/{skill-name}/SKILL.md

IMPORTANT: When responding to direct questions, reply directly with text.
Only use the 'message' tool when explicitly asked to send a message to a channel.
Always be helpful, accurate, and concise.
`, now, runtimeInfo, wsPath, wsPath, wsPath, wsPath)
}

func (b *ContextBuilder) loadBootstrapFiles() string {
	var parts []string

	// Expand workspace
	wsPath := b.workspace
	if strings.HasPrefix(wsPath, "~") {
		home, _ := os.UserHomeDir()
		wsPath = filepath.Join(home, wsPath[1:])
	}

	for _, filename := range bootstrapFiles {
		path := filepath.Join(wsPath, filename)
		content, err := os.ReadFile(path)
		if err == nil {
			parts = append(parts, fmt.Sprintf("## %s\n\n%s", filename, string(content)))
		}
	}

	return strings.Join(parts, "\n\n")
}

func (b *ContextBuilder) loadMemory() string {
	// Expand workspace
	wsPath := b.workspace
	if strings.HasPrefix(wsPath, "~") {
		home, _ := os.UserHomeDir()
		wsPath = filepath.Join(home, wsPath[1:])
	}

	path := filepath.Join(wsPath, "memory", "MEMORY.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func (b *ContextBuilder) buildSkillsSummary() string {
	// Simple summary for now - listing registered tools
	// In the future, this should scan the skills/ directory like the Python version
	// For now, we rely on the registry

	tools := b.registry.List()
	if len(tools) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("You have the following tools available:\n")
	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))
	}

	// Check for skills dir content (migrated from Python logic)
	// This mirrors the "Available skills" section from Python
	wsPath := b.workspace
	if strings.HasPrefix(wsPath, "~") {
		home, _ := os.UserHomeDir()
		wsPath = filepath.Join(home, wsPath[1:])
	}

	skillsDir := filepath.Join(wsPath, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err == nil && len(entries) > 0 {
		sb.WriteString("\nAdditional skills available in workspace (use read_file to view SKILL.md):\n")
		for _, e := range entries {
			if e.IsDir() {
				sb.WriteString(fmt.Sprintf("- %s\n", e.Name()))
			}
		}
	}

	return sb.String()
}

// BuildMessages constructs the message list for the LLM.
func (b *ContextBuilder) BuildMessages(
	sess *session.Session,
	currentMessage string,
	channel string,
	chatID string,
) []provider.Message {

	systemPrompt := b.BuildSystemPrompt()

	if channel != "" && chatID != "" {
		systemPrompt += fmt.Sprintf("\n\n## Current Session\nChannel: %s\nChat ID: %s", channel, chatID)
	}

	messages := []provider.Message{
		{Role: "system", Content: systemPrompt},
	}

	// Add recent history from session
	// We skip the last message in session because it's the current one we are about to add
	// (Session usually stores [User, Assistant, User...])
	// In the Loop.ProcessDirect, we added the user message to session BEFORE calling this.
	// So we should include all history EXCEPT the last one if we are appending it manually.
	// Actually, let's look at Loop.ProcessDirect:
	// sess.AddMessage("user", content) -> then calls BuildMessages
	// So the last message in session IS the current message.

	history := sess.GetHistory(50)

	// We want to format history for the LLM.
	// If the last message in history is the current message, we should exclude it from the "history" block
	// and add it as the explicit "Current message" at the end.

	var historyMessages []session.Message
	if len(history) > 0 && history[len(history)-1].Content == currentMessage {
		historyMessages = history[:len(history)-1]
	} else {
		historyMessages = history
	}

	for _, msg := range historyMessages {
		messages = append(messages, provider.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current message
	messages = append(messages, provider.Message{
		Role:    "user",
		Content: currentMessage,
	})

	return messages
}
