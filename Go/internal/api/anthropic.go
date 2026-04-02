package api

import (
	"context"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sai-sridhar-repo-07/forge/internal/config"
)

// AnthropicProvider calls the Anthropic Claude API.
type AnthropicProvider struct {
	client *anthropic.Client
	cfg    *config.Config
}

// NewAnthropic creates an Anthropic provider.
func NewAnthropic(cfg *config.Config) (*AnthropicProvider, error) {
	c := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))
	return &AnthropicProvider{client: c, cfg: cfg}, nil
}

func (p *AnthropicProvider) Name() string  { return "anthropic" }
func (p *AnthropicProvider) Model() string { return p.cfg.Model }

func (p *AnthropicProvider) Stream(
	ctx context.Context,
	messages []anthropic.MessageParam,
	tools []ToolDefinition,
	systemPrompt string,
) <-chan StreamEvent {
	ch := make(chan StreamEvent, 32)
	go func() {
		defer close(ch)

		params := anthropic.MessageNewParams{
			Model:     anthropic.F(anthropic.Model(p.cfg.Model)),
			MaxTokens: anthropic.F(int64(p.cfg.MaxTokens)),
			System:    anthropic.F([]anthropic.TextBlockParam{{Type: anthropic.F(anthropic.TextBlockParamTypeText), Text: anthropic.F(systemPrompt)}}),
			Messages:  anthropic.F(messages),
		}

		if len(tools) > 0 {
			sdkTools := BuildToolParams(tools)
			toolsUnion := make([]anthropic.ToolUnionUnionParam, len(sdkTools))
			for i, t := range sdkTools {
				toolsUnion[i] = t
			}
			params.Tools = anthropic.F(toolsUnion)
		}

		stream := p.client.Messages.NewStreaming(ctx, params)
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
					curToolID, curToolName, toolInputBuf = "", "", ""
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
