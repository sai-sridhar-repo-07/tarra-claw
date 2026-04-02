package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sai-sridhar-repo-07/forge/internal/api"
)

// AskUserQuestionTool prompts the user for input during a session.
type AskUserQuestionTool struct{}

func NewAskUserQuestionTool() *AskUserQuestionTool { return &AskUserQuestionTool{} }

func (t *AskUserQuestionTool) Name() string        { return "AskUserQuestion" }
func (t *AskUserQuestionTool) NeedsPermission() bool { return false }

func (t *AskUserQuestionTool) Description() string {
	return `Ask the user a question and wait for their response.
Use when you genuinely need the user's input to proceed — not for confirmation of obvious actions.`
}

func (t *AskUserQuestionTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"question": map[string]any{
				"type":        "string",
				"description": "The question to ask the user.",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Brief context for why you're asking.",
			},
		},
		"required": []string{"question"},
	}
}

func (t *AskUserQuestionTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{Name: t.Name(), Description: t.Description(), InputSchema: t.InputSchema()}
}

func (t *AskUserQuestionTool) Execute(_ context.Context, input map[string]any) (string, error) {
	question, _ := input["question"].(string)
	if question == "" {
		return "", fmt.Errorf("question is required")
	}
	desc, _ := input["description"].(string)

	if desc != "" {
		fmt.Fprintf(os.Stderr, "\n[%s]\n", desc)
	}
	fmt.Fprintf(os.Stderr, "\n? %s\n> ", question)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no input received")
}
