package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

// OllamaProvider calls a local Ollama server (free, no API key).
type OllamaProvider struct {
	host   string
	model  string
	client *http.Client
}

// NewOllama creates an Ollama provider.
// host defaults to http://localhost:11434
func NewOllama(host, model string) *OllamaProvider {
	if host == "" {
		host = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaProvider{
		host:   host,
		model:  model,
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (p *OllamaProvider) Name() string  { return "ollama" }
func (p *OllamaProvider) Model() string { return p.model }

// ── Ollama wire types ──────────────────────────────────────────────────────────

type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaToolCall struct {
	Function ollamaFunction `json:"function"`
}

type ollamaFunction struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ollamaTool struct {
	Type     string          `json:"type"`
	Function ollamaToolDef   `json:"function"`
}

type ollamaToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaTool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

type ollamaChatResponse struct {
	Model   string        `json:"model"`
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
	// Usage stats (only in final message)
	PromptEvalCount int `json:"prompt_eval_count"`
	EvalCount       int `json:"eval_count"`
}

// ── Stream implementation ──────────────────────────────────────────────────────

func (p *OllamaProvider) Stream(
	ctx context.Context,
	messages []anthropic.MessageParam,
	tools []ToolDefinition,
	systemPrompt string,
) <-chan StreamEvent {
	ch := make(chan StreamEvent, 32)

	go func() {
		defer close(ch)

		// Convert messages to Ollama format
		ollamaMsgs := convertMessages(messages, systemPrompt)

		// Convert tools
		var ollamaTools []ollamaTool
		for _, t := range tools {
			params := map[string]any{"type": "object"}
			if props, ok := t.InputSchema["properties"]; ok {
				params["properties"] = props
			}
			if req, ok := t.InputSchema["required"]; ok {
				params["required"] = req
			}
			ollamaTools = append(ollamaTools, ollamaTool{
				Type: "function",
				Function: ollamaToolDef{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  params,
				},
			})
		}

		req := ollamaChatRequest{
			Model:    p.model,
			Messages: ollamaMsgs,
			Tools:    ollamaTools,
			Stream:   true,
		}

		body, _ := json.Marshal(req)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.host+"/api/chat", bytes.NewReader(body))
		if err != nil {
			ch <- StreamEvent{Type: EventError, Err: fmt.Errorf("ollama request error: %w", err)}
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(httpReq)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			ch <- StreamEvent{Type: EventError, Err: fmt.Errorf("ollama unreachable — is it running? (ollama serve): %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			ch <- StreamEvent{Type: EventError, Err: fmt.Errorf("ollama error %d — model '%s' may not be pulled yet (run: ollama pull %s)", resp.StatusCode, p.model, p.model)}
			return
		}

		var (
			inputTokens  int64
			outputTokens int64
		)

		// Track pending tool calls assembled from non-streaming final message
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var chunk ollamaChatResponse
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}

			if !chunk.Done {
				// Streaming text content
				if chunk.Message.Content != "" {
					ch <- StreamEvent{Type: EventTextDelta, Text: chunk.Message.Content}
				}
			} else {
				// Final chunk — may contain tool calls
				inputTokens = int64(chunk.PromptEvalCount)
				outputTokens = int64(chunk.EvalCount)

				// Handle tool calls in final message
				for _, tc := range chunk.Message.ToolCalls {
					id := fmt.Sprintf("call_%s_%d", tc.Function.Name, time.Now().UnixNano())
					ch <- StreamEvent{
						Type: EventToolUseStart,
						Tool: &ToolUse{ID: id, Name: tc.Function.Name},
					}
					ch <- StreamEvent{
						Type: EventToolUseEnd,
						Tool: &ToolUse{ID: id, Name: tc.Function.Name, Input: tc.Function.Arguments},
					}
				}
			}
		}

		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			ch <- StreamEvent{Type: EventError, Err: err}
			return
		}

		ch <- StreamEvent{
			Type:  EventDone,
			Usage: &UsageStats{InputTokens: inputTokens, OutputTokens: outputTokens},
		}
	}()

	return ch
}

// convertMessages converts Anthropic MessageParam slice to Ollama format.
func convertMessages(messages []anthropic.MessageParam, systemPrompt string) []ollamaMessage {
	var out []ollamaMessage

	if systemPrompt != "" {
		out = append(out, ollamaMessage{Role: "system", Content: systemPrompt})
	}

	for _, m := range messages {
		role := string(m.Role.Value)
		var contentParts []string
		var toolCalls []ollamaToolCall

		for _, block := range m.Content.Value {
			switch b := block.(type) {
			case anthropic.TextBlockParam:
				contentParts = append(contentParts, b.Text.Value)

			case anthropic.ToolUseBlockParam:
				argsJSON, _ := json.Marshal(b.Input.Value)
				var args map[string]any
				_ = json.Unmarshal(argsJSON, &args)
				toolCalls = append(toolCalls, ollamaToolCall{
					Function: ollamaFunction{Name: b.Name.Value, Arguments: args},
				})

			case anthropic.ToolResultBlockParam:
				// Tool results become "tool" role messages in Ollama
				for _, c := range b.Content.Value {
					if tb, ok := c.(anthropic.TextBlockParam); ok {
						out = append(out, ollamaMessage{
							Role:    "tool",
							Content: tb.Text.Value,
						})
					}
				}
				continue
			}
		}

		msg := ollamaMessage{
			Role:      role,
			Content:   strings.Join(contentParts, "\n"),
			ToolCalls: toolCalls,
		}
		out = append(out, msg)
	}

	return out
}

// IsAvailable checks if Ollama is running and the model is available.
func (p *OllamaProvider) IsAvailable(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", p.host+"/api/tags", nil)
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama not running — start it with: ollama serve")
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	modelBase := strings.Split(p.model, ":")[0]
	for _, m := range result.Models {
		if strings.HasPrefix(m.Name, modelBase) {
			return nil
		}
	}
	return fmt.Errorf("model '%s' not found — run: ollama pull %s", p.model, p.model)
}

// ListModels returns all locally available Ollama models.
func ListOllamaModels(host string) ([]string, error) {
	if host == "" {
		host = "http://localhost:11434"
	}
	resp, err := http.Get(host + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("ollama not running")
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	names := make([]string, len(result.Models))
	for i, m := range result.Models {
		names[i] = m.Name
	}
	return names, nil
}
