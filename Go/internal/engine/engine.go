package engine

import (
	"context"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/cost"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/tools"
)

// Engine orchestrates the AI query loop with tool execution.
type Engine struct {
	provider api.Provider
	registry *tools.Registry
	cfg      *config.Config
	history  []anthropic.MessageParam
	costSess *cost.Session
	onEvent  func(Event)
}

// Event is emitted during the query loop for the TUI to consume.
type Event struct {
	Type    EventType
	Text    string
	Tool    string
	Input   map[string]any
	Result  string
	IsError bool
	Usage   *api.UsageStats
}

type EventType int

const (
	EventAssistantText EventType = iota
	EventToolStart
	EventToolResult
	EventError
	EventDone
)

// New creates an Engine, auto-selecting provider from config.
func New(cfg *config.Config) (*Engine, error) {
	var provider api.Provider
	var err error

	switch cfg.Provider {
	case "ollama":
		p := api.NewOllama(cfg.OllamaHost, cfg.OllamaModel)
		// Check availability
		if avErr := p.IsAvailable(context.Background()); avErr != nil {
			return nil, fmt.Errorf("ollama: %w\n\nQuick fix:\n  brew install ollama\n  ollama serve &\n  ollama pull %s", avErr, cfg.OllamaModel)
		}
		provider = p
	case "anthropic":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY not set.\n\nQuick fix:\n  export ANTHROPIC_API_KEY=sk-ant-...\n\nOr use Ollama (free, no key):\n  ollama serve & ollama pull llama3.2\n  TARRA_PROVIDER=ollama claw")
		}
		provider, err = api.NewAnthropic(cfg)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown provider: %s (use 'anthropic' or 'ollama')", cfg.Provider)
	}

	return &Engine{
		provider: provider,
		registry: tools.DefaultRegistry(cfg.WorkingDir),
		cfg:      cfg,
		costSess: cost.New(cfg.Model),
	}, nil
}

// OnEvent sets a handler for streaming events (used by TUI).
func (e *Engine) OnEvent(fn func(Event)) { e.onEvent = fn }

func (e *Engine) emit(ev Event) {
	if e.onEvent != nil {
		e.onEvent(ev)
	}
}

// RunOnce sends a prompt non-interactively and prints to stdout.
func (e *Engine) RunOnce(ctx context.Context, prompt string) error {
	e.onEvent = func(ev Event) {
		switch ev.Type {
		case EventAssistantText:
			fmt.Print(ev.Text)
		case EventToolStart:
			fmt.Printf("\n[%s] %v\n", ev.Tool, formatInput(ev.Input))
		case EventToolResult:
			if ev.IsError {
				fmt.Printf("[error] %s\n", ev.Result)
			}
		case EventDone:
			fmt.Println()
			if ev.Usage != nil {
				fmt.Printf("\n[%s · in:%d out:%d]\n", e.provider.Name(), ev.Usage.InputTokens, ev.Usage.OutputTokens)
			}
		}
	}
	return e.Send(ctx, prompt)
}

// RunOnceDirect sends a prompt without tools — clean text output only.
// Used for review, commit, and other single-shot analysis commands.
func (e *Engine) RunOnceDirect(ctx context.Context, prompt string) error {
	e.history = append(e.history, anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))

	stream := e.provider.Stream(ctx, e.history, nil, "You are a helpful assistant. Be concise and direct.")
	for ev := range stream {
		switch ev.Type {
		case api.EventTextDelta:
			fmt.Print(ev.Text)
		case api.EventDone:
			fmt.Println()
			if ev.Usage != nil {
				fmt.Printf("\n[%s · in:%d out:%d]\n", e.provider.Name(), ev.Usage.InputTokens, ev.Usage.OutputTokens)
			}
		case api.EventError:
			return ev.Err
		}
	}
	return nil
}

// Send adds a user message and runs the full agentic loop.
func (e *Engine) Send(ctx context.Context, userMessage string) error {
	e.history = append(e.history, anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))
	for {
		done, err := e.step(ctx)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

func (e *Engine) step(ctx context.Context) (bool, error) {
	toolDefs := e.registry.Definitions()

	stream := e.provider.Stream(ctx, e.history, toolDefs, e.cfg.SystemPrompt)

	var (
		assistantText strings.Builder
		toolCalls     []pendingCall
		curTool       *pendingCall
	)

	for ev := range stream {
		switch ev.Type {
		case api.EventTextDelta:
			assistantText.WriteString(ev.Text)
			e.emit(Event{Type: EventAssistantText, Text: ev.Text})

		case api.EventToolUseStart:
			toolCalls = append(toolCalls, pendingCall{id: ev.Tool.ID, name: ev.Tool.Name})
			curTool = &toolCalls[len(toolCalls)-1]
			_ = curTool

		case api.EventToolUseDelta:
			// accumulation handled at EventToolUseEnd

		case api.EventToolUseEnd:
			if len(toolCalls) > 0 {
				toolCalls[len(toolCalls)-1].input = ev.Tool.Input
				e.emit(Event{Type: EventToolStart, Tool: ev.Tool.Name, Input: ev.Tool.Input})
			}

		case api.EventDone:
			if ev.Usage != nil {
				e.costSess.Add(cost.Usage{
					InputTokens:  ev.Usage.InputTokens,
					OutputTokens: ev.Usage.OutputTokens,
					CacheRead:    ev.Usage.CacheRead,
					CacheWrite:   ev.Usage.CacheWrite,
				})
			}
			e.emit(Event{Type: EventDone, Usage: ev.Usage})

		case api.EventError:
			e.emit(Event{Type: EventError, Text: ev.Err.Error()})
			return false, ev.Err
		}
	}

	// Build assistant message
	var contentBlocks []anthropic.ContentBlockParamUnion
	if assistantText.Len() > 0 {
		contentBlocks = append(contentBlocks, anthropic.NewTextBlock(assistantText.String()))
	}
	for _, tc := range toolCalls {
		contentBlocks = append(contentBlocks, anthropic.NewToolUseBlockParam(tc.id, tc.name, tc.input))
	}
	if len(contentBlocks) > 0 {
		e.history = append(e.history, anthropic.NewAssistantMessage(contentBlocks...))
	}

	if len(toolCalls) == 0 {
		return true, nil
	}

	// Execute tools and collect results
	results := e.executeTools(ctx, toolCalls)

	var resultBlocks []anthropic.ContentBlockParamUnion
	for _, r := range results {
		resultBlocks = append(resultBlocks, anthropic.NewToolResultBlock(r.id, r.content, r.isError))
	}
	e.history = append(e.history, anthropic.NewUserMessage(resultBlocks...))

	return false, nil
}

type pendingCall struct {
	id    string
	name  string
	input map[string]any
}

type toolResult struct {
	id      string
	content string
	isError bool
}

func (e *Engine) executeTools(ctx context.Context, calls []pendingCall) []toolResult {
	results := make([]toolResult, len(calls))
	for i, c := range calls {
		tool, ok := e.registry.Get(c.name)
		if !ok {
			results[i] = toolResult{id: c.id, content: fmt.Sprintf("unknown tool: %s", c.name), isError: true}
			e.emit(Event{Type: EventToolResult, Tool: c.name, Result: results[i].content, IsError: true})
			continue
		}
		out, err := tool.Execute(ctx, c.input)
		if err != nil {
			results[i] = toolResult{id: c.id, content: err.Error(), isError: true}
			e.emit(Event{Type: EventToolResult, Tool: c.name, Result: err.Error(), IsError: true})
		} else {
			results[i] = toolResult{id: c.id, content: out}
			e.emit(Event{Type: EventToolResult, Tool: c.name, Result: out})
		}
	}
	return results
}

// ClearHistory resets the conversation.
func (e *Engine) ClearHistory() { e.history = nil }

// Registry returns the tool registry.
func (e *Engine) Registry() *tools.Registry { return e.registry }

// CostSummary returns a human-readable cost/usage string.
func (e *Engine) CostSummary() string { return e.costSess.Summary() }

// ProviderInfo returns "provider · model".
func (e *Engine) ProviderInfo() string {
	return fmt.Sprintf("%s · %s", e.provider.Name(), e.provider.Model())
}

func formatInput(input map[string]any) string {
	if v, ok := input["command"]; ok {
		return fmt.Sprint(v)
	}
	if v, ok := input["file_path"]; ok {
		return fmt.Sprint(v)
	}
	if v, ok := input["pattern"]; ok {
		return fmt.Sprint(v)
	}
	return ""
}
