package compact

import (
	"context"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

const (
	TriggerRatio     = 0.85
	MaxContextTokens = 200_000
)

type Compactor struct {
	summarizer Summarizer
}

type Summarizer func(ctx context.Context, messages []anthropic.MessageParam) (string, error)

func New(s Summarizer) *Compactor {
	return &Compactor{summarizer: s}
}

func ShouldCompact(usedTokens, contextWindow int64) bool {
	if contextWindow <= 0 {
		contextWindow = MaxContextTokens
	}
	return float64(usedTokens)/float64(contextWindow) >= TriggerRatio
}

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
	summaryMsg := anthropic.NewUserMessage(
		anthropic.NewTextBlock(fmt.Sprintf("<compact_summary>\n%s\n</compact_summary>", summary)),
	)
	result := make([]anthropic.MessageParam, 0, 1+len(recent))
	result = append(result, summaryMsg)
	result = append(result, recent...)
	return result, nil
}

// EstimateTokens gives a rough estimate (4 chars ≈ 1 token).
func EstimateTokens(messages []anthropic.MessageParam) int64 {
	var total int64
	for _, m := range messages {
		for _, block := range m.Content.Value {
			switch b := block.(type) {
			case anthropic.TextBlockParam:
				total += int64(len(b.Text.Value)) / 4
			default:
				// conservative estimate for other block types
				total += 100
				_ = b
			}
		}
	}
	return total
}

func BuildSummarizationPrompt(messages []anthropic.MessageParam) string {
	var sb strings.Builder
	sb.WriteString("Summarise the following conversation concisely, preserving all key facts, ")
	sb.WriteString("decisions, code changes, and context needed to continue the work:\n\n")
	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("[%s]\n", m.Role))
		for _, b := range m.Content.Value {
			if tb, ok := b.(anthropic.TextBlockParam); ok {
				sb.WriteString(tb.Text.Value)
			}
		}
		sb.WriteString("\n\n")
	}
	return sb.String()
}
