package compact

import (
	"context"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

const (
	// TriggerRatio — compact when used tokens exceed this fraction of context window.
	TriggerRatio = 0.85
	// MaxContextTokens — default context window size.
	MaxContextTokens = 200_000
)

// Compactor reduces conversation history to fit within token limits.
type Compactor struct {
	summarizer Summarizer
}

// Summarizer is called to produce a summary of old messages.
type Summarizer func(ctx context.Context, messages []anthropic.MessageParam) (string, error)

// New creates a Compactor with the given summarizer.
func New(s Summarizer) *Compactor {
	return &Compactor{summarizer: s}
}

// ShouldCompact returns true when token usage exceeds the trigger threshold.
func ShouldCompact(usedTokens, contextWindow int64) bool {
	if contextWindow <= 0 {
		contextWindow = MaxContextTokens
	}
	return float64(usedTokens)/float64(contextWindow) >= TriggerRatio
}

// Compact reduces messages to a summary + recent messages.
// It keeps the last keepRecent messages verbatim and summarizes the rest.
func (c *Compactor) Compact(
	ctx context.Context,
	messages []anthropic.MessageParam,
	keepRecent int,
) ([]anthropic.MessageParam, error) {
	if len(messages) <= keepRecent {
		return messages, nil
	}

	toSummarise := messages[:len(messages)-keepRecent]
	recent := messages[len(messages)-keepRecent:]

	summary, err := c.summarizer(ctx, toSummarise)
	if err != nil {
		return messages, fmt.Errorf("compaction failed: %w", err)
	}

	// Build a synthetic assistant message with the summary
	summaryMsg := anthropic.NewUserMessage(
		anthropic.NewTextBlock(
			fmt.Sprintf("<compact_summary>\n%s\n</compact_summary>", summary),
		),
	)

	result := make([]anthropic.MessageParam, 0, 1+len(recent))
	result = append(result, summaryMsg)
	result = append(result, recent...)
	return result, nil
}

// EstimateTokens gives a rough token estimate (4 chars ≈ 1 token).
func EstimateTokens(messages []anthropic.MessageParam) int64 {
	var total int64
	for _, m := range messages {
		// Rough: marshal to string and count
		for _, block := range m.Content.Value {
			switch b := block.(type) {
			case anthropic.TextBlockParam:
				total += int64(len(b.Text.Value)) / 4
			case anthropic.ToolResultBlockParam:
				for _, c := range b.Content.Value {
					if tb, ok := c.(anthropic.ToolResultBlockParamContentUnionMember0); ok {
						total += int64(len(tb.Text.Value)) / 4
					}
				}
			}
		}
	}
	return total
}

// BuildSummarizationPrompt creates a prompt asking the model to summarise messages.
func BuildSummarizationPrompt(messages []anthropic.MessageParam) string {
	var sb strings.Builder
	sb.WriteString("Summarise the following conversation concisely, preserving all key facts, ")
	sb.WriteString("decisions, code changes, and context needed to continue the work:\n\n")
	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("[%s]\n", m.Role))
		for _, b := range m.Content.Value {
			switch bl := b.(type) {
			case anthropic.TextBlockParam:
				sb.WriteString(bl.Text.Value)
			}
		}
		sb.WriteString("\n\n")
	}
	return sb.String()
}
