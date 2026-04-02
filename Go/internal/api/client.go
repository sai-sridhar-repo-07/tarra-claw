// Package api defines the Provider interface and shared streaming types.
// Implementations: AnthropicProvider (anthropic.go), OllamaProvider (ollama.go)
package api

import (
	"encoding/json"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

// StreamEventType identifies the kind of event being emitted.
type StreamEventType int

const (
	EventTextDelta    StreamEventType = iota
	EventToolUseStart                 // model wants to call a tool
	EventToolUseDelta                 // partial tool input JSON
	EventToolUseEnd                   // tool input complete, ready to execute
	EventDone                         // stream finished, usage stats attached
	EventError                        // unrecoverable error
)

// StreamEvent is emitted on the channel returned by Provider.Stream().
type StreamEvent struct {
	Type  StreamEventType
	Text  string
	Tool  *ToolUse
	Usage *UsageStats
	Err   error
}

// ToolUse represents a single tool call from the model.
type ToolUse struct {
	ID    string
	Name  string
	Input map[string]any
}

// UsageStats holds token counts from one API call.
type UsageStats struct {
	InputTokens  int64
	OutputTokens int64
	CacheRead    int64
	CacheWrite   int64
}

// ToolDefinition describes a tool to the model.
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]any
}

// BuildToolParams converts ToolDefinitions to Anthropic SDK ToolParam slice.
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
