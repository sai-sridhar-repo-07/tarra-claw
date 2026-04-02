package api

import (
	"context"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

// Provider is the interface every AI backend must implement.
// Both Anthropic and Ollama satisfy this interface.
type Provider interface {
	// Stream sends messages and streams back events.
	// systemPrompt is prepended as context.
	// tools is the list of tools available to the model.
	Stream(
		ctx context.Context,
		messages []anthropic.MessageParam,
		tools []ToolDefinition,
		systemPrompt string,
	) <-chan StreamEvent

	// Name returns the provider name (e.g. "anthropic", "ollama").
	Name() string

	// Model returns the active model name.
	Model() string
}
