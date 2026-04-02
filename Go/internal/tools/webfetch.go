package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
)

const (
	maxFetchBytes   = 1_000_000 // 1MB
	fetchTimeout    = 30 * time.Second
)

// WebFetchTool fetches a URL and returns its text content.
type WebFetchTool struct {
	client *http.Client
}

func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		client: &http.Client{Timeout: fetchTimeout},
	}
}

func (t *WebFetchTool) Name() string        { return "WebFetch" }
func (t *WebFetchTool) NeedsPermission() bool { return false }

func (t *WebFetchTool) Description() string {
	return `Fetch a URL and return its text content (HTML stripped to readable text).
Use for reading documentation, APIs, or any web page.`
}

func (t *WebFetchTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to fetch.",
			},
			"prompt": map[string]any{
				"type":        "string",
				"description": "Optional: what specific information to extract from the page.",
			},
		},
		"required": []string{"url"},
	}
}

func (t *WebFetchTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{Name: t.Name(), Description: t.Description(), InputSchema: t.InputSchema()}
}

func (t *WebFetchTool) Execute(ctx context.Context, input map[string]any) (string, error) {
	url, ok := input["url"].(string)
	if !ok || url == "" {
		return "", fmt.Errorf("url is required")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	req.Header.Set("User-Agent", "tarra-claw/0.1 (AI agent CLI)")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchBytes))
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}

	content := stripHTML(string(body))
	if !utf8.ValidString(content) {
		return "", fmt.Errorf("response is not valid UTF-8 text")
	}

	// Trim to reasonable size
	if len(content) > 50_000 {
		content = content[:50_000] + "\n... (truncated)"
	}

	result := fmt.Sprintf("URL: %s\nStatus: %s\n\n%s", url, resp.Status, content)
	return result, nil
}

// stripHTML removes HTML tags and normalises whitespace.
func stripHTML(html string) string {
	var sb strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			sb.WriteRune(' ')
		case !inTag:
			sb.WriteRune(r)
		}
	}
	// Collapse whitespace
	lines := strings.Split(sb.String(), "\n")
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
