// Package agent implements the core agent loop.
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/kamir/gomikrobot/internal/bus"
	"github.com/kamir/gomikrobot/internal/provider"
	"github.com/kamir/gomikrobot/internal/session"
	"github.com/kamir/gomikrobot/internal/tools"
)

// LoopOptions contains configuration for the agent loop.
type LoopOptions struct {
	Bus           *bus.MessageBus
	Provider      provider.LLMProvider
	Workspace     string
	Model         string
	MaxIterations int
}

// Loop is the core agent processing engine.
type Loop struct {
	bus            *bus.MessageBus
	provider       provider.LLMProvider
	registry       *tools.Registry
	sessions       *session.Manager
	contextBuilder *ContextBuilder
	workspace      string
	model          string
	maxIterations  int
	running        bool
}

// NewLoop creates a new agent loop.
func NewLoop(opts LoopOptions) *Loop {
	maxIter := opts.MaxIterations
	if maxIter == 0 {
		maxIter = 20
	}

	registry := tools.NewRegistry()

	// Create context builder
	ctxBuilder := NewContextBuilder(opts.Workspace, registry)

	loop := &Loop{
		bus:            opts.Bus,
		provider:       opts.Provider,
		registry:       registry,
		sessions:       session.NewManager(opts.Workspace),
		contextBuilder: ctxBuilder,
		workspace:      opts.Workspace,
		model:          opts.Model,
		maxIterations:  maxIter,
	}

	// Register default tools
	loop.registerDefaultTools()

	return loop
}

func (l *Loop) registerDefaultTools() {
	l.registry.Register(tools.NewReadFileTool())
	l.registry.Register(tools.NewWriteFileTool())
	l.registry.Register(tools.NewEditFileTool())
	l.registry.Register(tools.NewListDirTool())
	l.registry.Register(tools.NewExecTool(0, true, l.workspace))
}

// Run starts the agent loop, processing messages from the bus.
func (l *Loop) Run(ctx context.Context) error {
	l.running = true
	slog.Info("Agent loop started")

	for l.running {
		msg, err := l.bus.ConsumeInbound(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // Context cancelled, normal shutdown
			}
			slog.Error("Failed to consume message", "error", err)
			continue
		}

		response, err := l.processMessage(ctx, msg)
		if err != nil {
			slog.Error("Failed to process message", "error", err)
			response = fmt.Sprintf("Error: %v", err)
		}

		if response != "" {
			l.bus.PublishOutbound(&bus.OutboundMessage{
				Channel: msg.Channel,
				ChatID:  msg.ChatID,
				Content: response,
			})
		}
	}

	return nil
}

// Stop signals the agent loop to stop.
func (l *Loop) Stop() {
	l.running = false
}

// ProcessDirect processes a message directly (for CLI usage).
func (l *Loop) ProcessDirect(ctx context.Context, content, sessionKey string) (string, error) {
	// Extract channel and chatID from key if possible
	parts := strings.SplitN(sessionKey, ":", 2)
	channel, chatID := "cli", "default"
	if len(parts) == 2 {
		channel, chatID = parts[0], parts[1]
	}

	// Get or create session
	sess := l.sessions.GetOrCreate(sessionKey)
	sess.AddMessage("user", content)

	// Build messages using the context builder
	messages := l.contextBuilder.BuildMessages(sess, content, channel, chatID)

	// Run the agentic loop
	response, err := l.runAgentLoop(ctx, messages)
	if err != nil {
		return "", err
	}

	// Save session with response
	sess.AddMessage("assistant", response)
	l.sessions.Save(sess)

	return response, nil
}

func (l *Loop) processMessage(ctx context.Context, msg *bus.InboundMessage) (string, error) {
	sessionKey := fmt.Sprintf("%s:%s", msg.Channel, msg.ChatID)
	return l.ProcessDirect(ctx, msg.Content, sessionKey)
}

func (l *Loop) runAgentLoop(ctx context.Context, messages []provider.Message) (string, error) {
	toolDefs := l.buildToolDefinitions()

	for i := 0; i < l.maxIterations; i++ {
		// Call LLM
		resp, err := l.provider.Chat(ctx, &provider.ChatRequest{
			Messages:    messages,
			Tools:       toolDefs,
			Model:       l.model,
			MaxTokens:   4096,
			Temperature: 0.7,
		})
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		// Check for tool calls
		if len(resp.ToolCalls) == 0 {
			// No tool calls, return the response
			return resp.Content, nil
		}

		// Add assistant message with tool calls
		messages = append(messages, provider.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// Execute each tool call
		for _, tc := range resp.ToolCalls {
			result, err := l.registry.Execute(ctx, tc.Name, tc.Arguments)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			// Add tool result
			messages = append(messages, provider.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})

			slog.Debug("Tool executed", "name", tc.Name, "result_length", len(result))
		}
	}

	return "Max iterations reached. Please try a simpler request.", nil
}

func (l *Loop) buildToolDefinitions() []provider.ToolDefinition {
	toolList := l.registry.List()
	defs := make([]provider.ToolDefinition, len(toolList))

	for i, tool := range toolList {
		defs[i] = provider.ToolDefinition{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}
	}

	return defs
}

// SessionKey builds a session key from channel and chat ID.
func SessionKey(channel, chatID string) string {
	return strings.Join([]string{channel, chatID}, ":")
}
