package tools

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
)

const maxReadLines = 2000

// ReadTool reads file contents with optional line range.
type ReadTool struct{}

func NewReadTool() *ReadTool { return &ReadTool{} }

func (t *ReadTool) Name() string        { return "Read" }
func (t *ReadTool) NeedsPermission() bool { return false }

func (t *ReadTool) Description() string {
	return `Read the contents of a file. Supports optional offset and limit for large files.
Results include line numbers in cat -n format.`
}

func (t *ReadTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "Absolute path to the file to read.",
			},
			"offset": map[string]any{
				"type":        "integer",
				"description": "Line number to start reading from (1-based).",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of lines to read.",
			},
		},
		"required": []string{"file_path"},
	}
}

func (t *ReadTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

func (t *ReadTool) Execute(_ context.Context, input map[string]any) (string, error) {
	path, ok := input["file_path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("file_path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")

	offset := 0
	if v, ok := input["offset"]; ok {
		offset = toInt(v) - 1
		if offset < 0 {
			offset = 0
		}
	}

	limit := maxReadLines
	if v, ok := input["limit"]; ok {
		if l := toInt(v); l > 0 {
			limit = l
		}
	}

	end := offset + limit
	if end > len(lines) {
		end = len(lines)
	}
	if offset > len(lines) {
		return "(offset beyond end of file)", nil
	}

	lines = lines[offset:end]

	var sb strings.Builder
	for i, line := range lines {
		lineNum := offset + i + 1
		sb.WriteString(fmt.Sprintf("%6d\t%s\n", lineNum, line))
	}

	return sb.String(), nil
}

func toInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}
