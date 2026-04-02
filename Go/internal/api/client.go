package api

import (
	"context"
	"encoding/json"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
)

type Client struct {
	inner *anthropic.Client
	cfg   *config.Config
}

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

type ToolUse struct {
	ID    string
	Name  string
	Input map[string]any
}

type UsageStats struct {
	InputTokens  int64
	OutputTokens int64
	CacheRead    int64
	CacheWrite   int64
}

type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]any
}

func New(cfg *config.Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	c := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))
	return &Client{inner: c, cfg: cfg}, nil
}

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
			toolsUnion := make([]anthropic.ToolUnionUnionParam, len(tools))
			for i, t := range tools {
				toolsUnion[i] = t
			}
			params.Tools = anthropic.F(toolsUnion)
		}
		stream := c.inner.Messages.NewStreaming(ctx, params)
		defer stream.Close()
		var accumMsg anthropic.Message
		var curToolID, curToolName, toolInputBuf string
		for stream.Next() {
			event := stream.Current()
			_ = accumMsg.Accumulate(event)
			switch e := event.AsUnion().(type) {
			case anthropic.ContentBlockStartEvent:
				switch b := e.ContentBlock.AsUnion().(type) {
				case anthropic.TextBlock:
					_ = b
				case anthropic.ToolUseBlock:
					curToolID = b.ID
					curToolName = b.Name
					toolInputBuf = ""
					ch <- StreamEvent{Type: EventToolUseStart, Tool: &ToolUse{ID: b.ID, Name: b.Name}}
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
				if curToolID != "" {
					input, _ := parseJSON(toolInputBuf)
					ch <- StreamEvent{Type: EventToolUseEnd, Tool: &ToolUse{ID: curToolID, Name: curToolName, Input: input}}
					curToolID = ""
					curToolName = ""
					toolInputBuf = ""
				}
			case anthropic.MessageDeltaEvent:
				_ = e
			case anthropic.MessageStopEvent:
				_ = e
			}
		}
		if err := stream.Err(); err != nil {
			if ctx.Err() == nil {
				ch <- StreamEvent{Type: EventError, Err: err}
			}
			return
		}
		ch <- StreamEvent{
			Type: EventDone,
			Usage: &UsageStats{
				InputTokens:  accumMsg.Usage.InputTokens,
				OutputTokens: accumMsg.Usage.OutputTokens,
				CacheRead:    accumMsg.Usage.CacheReadInputTokens,
				CacheWrite:   accumMsg.Usage.CacheCreationInputTokens,
			},
		}
	}()
	return ch
}

func BuildToolParams(tools []ToolDefinition) []anthropic.ToolParam {
	params := make([]anthropic.ToolParam, len(tools))
	for i, t := range tools {
		schema := map[string]any{"type": "object"}
		if props, ok := t.InputSchema["properties"]; ok {
			schema["properties"] = props
		}
		if req, ok := t.InputSchema["required"]; ok {
			schema["required"] = req
		}
		params[i] = anthropic.ToolParam{
			Name:        anthropic.F(t.Name),
			Description: anthropic.F(t.Description),
			InputSchema: anthropic.F[any](schema),
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
