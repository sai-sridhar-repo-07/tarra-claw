package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
)

// EditTool performs exact string replacements in files.
type EditTool struct{}

func NewEditTool() *EditTool { return &EditTool{} }

func (t *EditTool) Name() string        { return "Edit" }
func (t *EditTool) NeedsPermission() bool { return true }

func (t *EditTool) Description() string {
	return `Perform an exact string replacement in a file.
old_string must appear exactly once in the file (provide enough context to be unique).
Use replace_all=true to replace every occurrence.`
}

func (t *EditTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "Absolute path to the file to edit.",
			},
			"old_string": map[string]any{
				"type":        "string",
				"description": "The exact text to replace. Must be unique in the file.",
			},
			"new_string": map[string]any{
				"type":        "string",
				"description": "The text to replace old_string with.",
			},
			"replace_all": map[string]any{
				"type":        "boolean",
				"description": "If true, replace all occurrences. Default: false.",
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

func (t *EditTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

func (t *EditTool) Execute(_ context.Context, input map[string]any) (string, error) {
	path, ok := input["file_path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("file_path is required")
	}
	oldStr, ok := input["old_string"].(string)
	if !ok {
		return "", fmt.Errorf("old_string is required")
	}
	newStr, _ := input["new_string"].(string)
	replaceAll, _ := input["replace_all"].(bool)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", path, err)
	}
	content := string(data)

	count := strings.Count(content, oldStr)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in %s", path)
	}
	if !replaceAll && count > 1 {
		return "", fmt.Errorf("old_string appears %d times in %s — provide more context to make it unique, or use replace_all=true", count, path)
	}

	var updated string
	if replaceAll {
		updated = strings.ReplaceAll(content, oldStr, newStr)
	} else {
		updated = strings.Replace(content, oldStr, newStr, 1)
	}

	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return "", fmt.Errorf("cannot write %s: %w", path, err)
	}

	replaced := 1
	if replaceAll {
		replaced = count
	}
	return fmt.Sprintf("Replaced %d occurrence(s) in %s", replaced, path), nil
}
