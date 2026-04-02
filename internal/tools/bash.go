package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
)

const bashTimeout = 2 * time.Minute

// BashTool executes shell commands in the working directory.
type BashTool struct {
	workDir string
}

func NewBashTool(workDir string) *BashTool {
	return &BashTool{workDir: workDir}
}

func (t *BashTool) Name() string { return "Bash" }

func (t *BashTool) Description() string {
	return `Execute a bash command in a persistent shell session.
Use for running tests, build commands, git operations, and any shell task.
Timeout: 2 minutes. For long-running processes, use run_in_background=true.`
}

func (t *BashTool) NeedsPermission() bool { return true }

func (t *BashTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The bash command to execute.",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Short description of what this command does (shown to user).",
			},
			"timeout": map[string]any{
				"type":        "integer",
				"description": "Optional timeout in milliseconds (max 600000).",
			},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

func (t *BashTool) Execute(ctx context.Context, input map[string]any) (string, error) {
	command, ok := input["command"].(string)
	if !ok || strings.TrimSpace(command) == "" {
		return "", fmt.Errorf("command is required")
	}

	timeout := bashTimeout
	if ms, ok := input["timeout"].(float64); ok && ms > 0 {
		d := time.Duration(ms) * time.Millisecond
		if d > 10*time.Minute {
			d = 10 * time.Minute
		}
		timeout = d
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = t.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var out strings.Builder
	if stdout.Len() > 0 {
		out.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString("<stderr>\n")
		out.WriteString(stderr.String())
		out.WriteString("</stderr>")
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return out.String(), fmt.Errorf("command timed out after %v", timeout)
		}
		// Return output even on non-zero exit; model handles error analysis
		if out.Len() == 0 {
			return fmt.Sprintf("exit status: %v", err), nil
		}
		return out.String(), nil
	}

	result := out.String()
	if result == "" {
		return "(no output)", nil
	}
	return result, nil
}
