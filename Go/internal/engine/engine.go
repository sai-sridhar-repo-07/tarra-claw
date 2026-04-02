package engine

import (
	"context"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/tools"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/cost"
)

// Engine orchestrates the AI query loop with tool execution.
type Engine struct {
	client   *api.Client
	registry *tools.Registry
	cfg      *config.Config
	history  []anthropic.MessageParam
	cost     *cost.Session
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

// New creates a new Engine.
func New(cfg *config.Config) (*Engine, error) {
	client, err := api.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Engine{
		client:   client,
		registry: tools.DefaultRegistry(cfg.WorkingDir),
		cost:     cost.New(cfg.Model),
		cfg:      cfg,
	}, nil
}

// OnEvent sets a handler for streaming events (used by TUI).
func (e *Engine) OnEvent(fn func(Event)) {
	e.onEvent = fn
}

// emit sends an event to the registered handler.
func (e *Engine) emit(ev Event) {
	if e.onEvent != nil {
		e.onEvent(ev)
	}
}

// RunOnce sends a single user prompt and runs the full tool loop, printing to stdout.
func (e *Engine) RunOnce(ctx context.Context, prompt string) error {
	e.onEvent = func(ev Event) {
		switch ev.Type {
		case EventAssistantText:
			fmt.Print(ev.Text)
		case EventToolStart:
			fmt.Printf("\n[tool: %s]\n", ev.Tool)
		case EventToolResult:
			if ev.IsError {
				fmt.Printf("[error]: %s\n", ev.Result)
			} else {
				fmt.Printf("[result]: %s\n", ev.Result)
			}
		case EventDone:
			fmt.Println()
		}
	}
	return e.Send(ctx, prompt)
}

// Send adds a user message and runs the full agentic loop until the model stops.
func (e *Engine) Send(ctx context.Context, userMessage string) error {
	e.history = append(e.history, anthropic.NewUserMessage(
		anthropic.NewTextBlock(userMessage),
	))

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

// step runs one round of the query loop. Returns true when the model stops.
func (e *Engine) step(ctx context.Context) (bool, error) {
	toolDefs := e.registry.Definitions()
	toolParams := api.BuildToolParams(toolDefs)

	stream := e.client.Stream(ctx, e.history, toolParams, e.cfg.SystemPrompt)

	var (
		assistantText strings.Builder
		toolCalls     []pendingToolCall
		currentTool   *pendingToolCall
	)

	for ev := range stream {
		switch ev.Type {
		case api.EventTextDelta:
			assistantText.WriteString(ev.Text)
			e.emit(Event{Type: EventAssistantText, Text: ev.Text})

		case api.EventToolUseStart:
			tc := &pendingToolCall{id: ev.Tool.ID, name: ev.Tool.Name}
			toolCalls = append(toolCalls, *tc)
			currentTool = &toolCalls[len(toolCalls)-1]
			e.emit(Event{Type: EventToolStart, Tool: ev.Tool.Name, Input: ev.Tool.Input})

		case api.EventToolUseDelta:
			// input chunks accumulate in EventToolUseEnd

		case api.EventToolUseEnd:
			if currentTool != nil {
				currentTool.input = ev.Tool.Input
				// update the last entry
				toolCalls[len(toolCalls)-1] = *currentTool
			}

		case api.EventDone:
			e.emit(Event{Type: EventDone, Usage: ev.Usage})

		case api.EventError:
			e.emit(Event{Type: EventError, Text: ev.Err.Error()})
			return false, ev.Err
		}
	}

	// Build assistant message content blocks
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

	// No tool calls = model is done
	if len(toolCalls) == 0 {
		return true, nil
	}

	// Execute all tool calls (concurrent where safe)
	results := e.executeTools(ctx, toolCalls)

	// Add tool results as a single user message
	var resultBlocks []anthropic.ContentBlockParamUnion
	for _, r := range results {
		resultBlocks = append(resultBlocks, anthropic.NewToolResultBlock(r.id, r.content, r.isError))
	}
	e.history = append(e.history, anthropic.NewUserMessage(resultBlocks...))

	return false, nil
}

type pendingToolCall struct {
	id    string
	name  string
	input map[string]any
}

type toolResult struct {
	id      string
	content string
	isError bool
}

func (e *Engine) executeTools(ctx context.Context, calls []pendingToolCall) []toolResult {
	results := make([]toolResult, len(calls))

	// Use goroutines for read-only tools; serialize write tools
	type job struct {
		idx  int
		call pendingToolCall
	}

	ch := make(chan job, len(calls))
	for i, c := range calls {
		ch <- job{i, c}
	}
	close(ch)

	// For simplicity in v1, execute sequentially (concurrent in v2)
	for j := range ch {
		tool, ok := e.registry.Get(j.call.name)
		if !ok {
			results[j.idx] = toolResult{
				id:      j.call.id,
				content: fmt.Sprintf("unknown tool: %s", j.call.name),
				isError: true,
			}
			e.emit(Event{Type: EventToolResult, Tool: j.call.name, Result: results[j.idx].content, IsError: true})
			continue
		}

		e.emit(Event{Type: EventToolStart, Tool: j.call.name, Input: j.call.input})
		out, err := tool.Execute(ctx, j.call.input)

		if err != nil {
			results[j.idx] = toolResult{id: j.call.id, content: err.Error(), isError: true}
			e.emit(Event{Type: EventToolResult, Tool: j.call.name, Result: err.Error(), IsError: true})
		} else {
			results[j.idx] = toolResult{id: j.call.id, content: out}
			e.emit(Event{Type: EventToolResult, Tool: j.call.name, Result: out})
		}
	}

	return results
}

// ClearHistory resets conversation history (for /clear command).
func (e *Engine) ClearHistory() {
	e.history = nil
}

// Registry returns the tool registry (for /help, skill listing).
func (e *Engine) Registry() *tools.Registry {
	return e.registry
}

// CostSummary returns a human-readable cost/usage string.
func (e *Engine) CostSummary() string {
	if e.cost == nil {
		return "no cost data"
	}
	return e.cost.Summary()
}
