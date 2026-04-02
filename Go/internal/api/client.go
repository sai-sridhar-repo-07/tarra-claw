package api

import (
	"context"
	"encoding/json"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
)

// Client wraps the Anthropic SDK with streaming support.
type Client struct {
	inner *anthropic.Client
	cfg   *config.Config
}

// StreamEvent is emitted during streaming responses.
type StreamEvent struct {
	Type  StreamEventType
	Text  string
	Tool  *ToolUse
	Usage *UsageStats
	Err   error
}

type StreamEventType int

const (
	EventTextDelta StreamEventType = iota
	EventToolUseStart
	EventToolUseDelta
	EventToolUseEnd
	EventDone
	EventError
)

// ToolUse represents a tool call from the model.
type ToolUse struct {
	ID    string
	Name  string
	Input map[string]any
}

// UsageStats tracks token usage per request.
type UsageStats struct {
	InputTokens  int64
	OutputTokens int64
	CacheRead    int64
	CacheWrite   int64
}

// ToolDefinition describes a tool exposed to the model.
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]any
}

// New creates a new API client.
func New(cfg *config.Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	c := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))
	return &Client{inner: &c, cfg: cfg}, nil
}

// Stream sends messages and streams the response, emitting events on the returned channel.
// The channel is closed when streaming completes or errors.
func (c *Client) Stream(
	ctx context.Context,
	messages []anthropic.MessageParam,
	tools []anthropic.ToolParam,
	systemPrompt string,
) <-chan StreamEvent {
	ch := make(chan StreamEvent, 32)

	go func() {
		defer close(ch)

		params := anthropic.MessageNewParams{
			Model:     anthropic.F(anthropic.Model(c.cfg.Model)),
			MaxTokens: anthropic.F(int64(c.cfg.MaxTokens)),
			System: anthropic.F([]anthropic.TextBlockParam{
				{Text: anthropic.F(systemPrompt)},
			}),
			Messages: anthropic.F(messages),
		}

		if len(tools) > 0 {
			params.Tools = anthropic.F(tools)
		}

		stream := c.inner.Messages.NewStreaming(ctx, params)

		var currentTool *ToolUse
		var toolInputBuf string

		for stream.Next() {
			event := stream.Current()

			switch e := event.AsUnion().(type) {
			case anthropic.ContentBlockStartEvent:
				switch b := e.ContentBlock.AsUnion().(type) {
				case anthropic.TextBlock:
					_ = b
				case anthropic.ToolUseBlock:
					currentTool = &ToolUse{ID: b.ID, Name: b.Name}
					toolInputBuf = ""
					ch <- StreamEvent{Type: EventToolUseStart, Tool: currentTool}
				}

			case anthropic.ContentBlockDeltaEvent:
				switch d := e.Delta.AsUnion().(type) {
				case anthropic.TextDelta:
					ch <- StreamEvent{Type: EventTextDelta, Text: d.Text}
				case anthropic.InputJSONDelta:
					toolInputBuf += d.PartialJSON
					ch <- StreamEvent{Type: EventToolUseDelta, Text: d.PartialJSON}
				}

			case anthropic.ContentBlockStopEvent:
				if currentTool != nil {
					if input, err := parseJSON(toolInputBuf); err == nil {
						currentTool.Input = input
					}
					ch <- StreamEvent{Type: EventToolUseEnd, Tool: currentTool}
					currentTool = nil
					toolInputBuf = ""
				}

			case anthropic.MessageDeltaEvent:
				_ = e
			}
		}

		if err := stream.Err(); err != nil {
			if ctx.Err() == nil {
				ch <- StreamEvent{Type: EventError, Err: err}
			}
			return
		}

		msg := stream.Message()
		ch <- StreamEvent{
			Type: EventDone,
			Usage: &UsageStats{
				InputTokens:  msg.Usage.InputTokens,
				OutputTokens: msg.Usage.OutputTokens,
				CacheRead:    msg.Usage.CacheReadInputTokens,
				CacheWrite:   msg.Usage.CacheCreationInputTokens,
			},
		}
	}()

	return ch
}

// BuildToolParams converts ToolDefinition slice to Anthropic SDK format.
func BuildToolParams(tools []ToolDefinition) []anthropic.ToolParam {
	params := make([]anthropic.ToolParam, len(tools))
	for i, t := range tools {
		props, _ := t.InputSchema["properties"]
		params[i] = anthropic.ToolParam{
			Name:        anthropic.F(t.Name),
			Description: anthropic.F(t.Description),
			InputSchema: anthropic.F(anthropic.ToolInputSchemaParam{
				Type:       anthropic.F(anthropic.ToolInputSchemaTypeObject),
				Properties: anthropic.F[any](props),
			}),
		}
	}
	return params
}

func parseJSON(raw string) (map[string]any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	return m, nil
}
