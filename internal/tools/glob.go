package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
)

// GlobTool finds files matching a glob pattern.
type GlobTool struct {
	workDir string
}

func NewGlobTool(workDir string) *GlobTool { return &GlobTool{workDir: workDir} }

func (t *GlobTool) Name() string        { return "Glob" }
func (t *GlobTool) NeedsPermission() bool { return false }

func (t *GlobTool) Description() string {
	return `Find files matching a glob pattern. Returns paths sorted by modification time.
Supports patterns like "**/*.go", "src/**/*.ts", "*.{json,yaml}".`
}

func (t *GlobTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Glob pattern to match files against.",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Directory to search in. Defaults to working directory.",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GlobTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

func (t *GlobTool) Execute(_ context.Context, input map[string]any) (string, error) {
	pattern, ok := input["pattern"].(string)
	if !ok || pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	base := t.workDir
	if p, ok := input["path"].(string); ok && p != "" {
		base = p
	}

	fullPattern := filepath.Join(base, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return "", fmt.Errorf("invalid pattern: %w", err)
	}

	if len(matches) == 0 {
		return "(no files found)", nil
	}

	sort.Strings(matches)

	var sb strings.Builder
	for _, m := range matches {
		sb.WriteString(m)
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}
