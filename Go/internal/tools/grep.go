package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sai-sridhar-repo-07/forge/internal/api"
)

// GrepTool searches file contents using ripgrep.
type GrepTool struct {
	workDir string
}

func NewGrepTool(workDir string) *GrepTool { return &GrepTool{workDir: workDir} }

func (t *GrepTool) Name() string        { return "Grep" }
func (t *GrepTool) NeedsPermission() bool { return false }

func (t *GrepTool) Description() string {
	return `Search file contents using ripgrep. Supports regex patterns.
Returns matching lines with file path and line number.
Filter by file type with the "type" param (e.g. "go", "ts", "py").`
}

func (t *GrepTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Regex pattern to search for.",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "File or directory to search. Defaults to working directory.",
			},
			"glob": map[string]any{
				"type":        "string",
				"description": "Glob pattern to filter files (e.g. \"*.go\").",
			},
			"type": map[string]any{
				"type":        "string",
				"description": "File type filter (e.g. \"go\", \"ts\", \"py\").",
			},
			"case_insensitive": map[string]any{
				"type":        "boolean",
				"description": "Case-insensitive search.",
			},
			"context": map[string]any{
				"type":        "integer",
				"description": "Lines of context to show around each match.",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

func (t *GrepTool) Execute(ctx context.Context, input map[string]any) (string, error) {
	pattern, ok := input["pattern"].(string)
	if !ok || pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	args := []string{"--line-number", "--with-filename", "--no-heading"}

	if ci, ok := input["case_insensitive"].(bool); ok && ci {
		args = append(args, "--ignore-case")
	}
	if g, ok := input["glob"].(string); ok && g != "" {
		args = append(args, "--glob", g)
	}
	if typ, ok := input["type"].(string); ok && typ != "" {
		args = append(args, "--type", typ)
	}
	if c := toInt(input["context"]); c > 0 {
		args = append(args, fmt.Sprintf("-C%d", c))
	}

	args = append(args, pattern)

	searchPath := t.workDir
	if p, ok := input["path"].(string); ok && p != "" {
		searchPath = p
	}
	args = append(args, searchPath)

	cmd := exec.CommandContext(ctx, "rg", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())

	if err != nil {
		if stdout.Len() == 0 {
			// Exit code 1 = no matches, not an error
			if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
				return "(no matches found)", nil
			}
			// rg not installed, fall back
			return t.fallback(ctx, pattern, searchPath, input)
		}
	}

	if out == "" {
		return "(no matches found)", nil
	}

	// Limit output size
	lines := strings.Split(out, "\n")
	if len(lines) > 500 {
		lines = lines[:500]
		return strings.Join(lines, "\n") + fmt.Sprintf("\n... (truncated, %d total lines)", len(lines)), nil
	}
	return out, nil
}

// fallback uses grep if rg is not available.
func (t *GrepTool) fallback(ctx context.Context, pattern, path string, input map[string]any) (string, error) {
	args := []string{"-rn", "--include=*"}
	if ci, ok := input["case_insensitive"].(bool); ok && ci {
		args = append(args, "-i")
	}
	args = append(args, pattern, path)

	cmd := exec.CommandContext(ctx, "grep", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()

	result := strings.TrimSpace(out.String())
	if result == "" {
		return "(no matches found)", nil
	}
	return result, nil
}
