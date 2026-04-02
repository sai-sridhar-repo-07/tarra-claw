package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sai-sridhar-repo-07/forge/internal/api"
)

// WriteTool creates or overwrites a file with given content.
type WriteTool struct{}

func NewWriteTool() *WriteTool { return &WriteTool{} }

func (t *WriteTool) Name() string        { return "Write" }
func (t *WriteTool) NeedsPermission() bool { return true }

func (t *WriteTool) Description() string {
	return `Write content to a file, creating it (and any parent directories) if needed.
Overwrites the file completely. For partial edits, use the Edit tool instead.`
}

func (t *WriteTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "Absolute path to the file to write.",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content to write to the file.",
			},
		},
		"required": []string{"file_path", "content"},
	}
}

func (t *WriteTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

func (t *WriteTool) Execute(_ context.Context, input map[string]any) (string, error) {
	path, ok := input["file_path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("file_path is required")
	}
	content, _ := input["content"].(string)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("cannot create directories: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("cannot write file: %w", err)
	}

	return fmt.Sprintf("File written successfully: %s (%d bytes)", path, len(content)), nil
}
