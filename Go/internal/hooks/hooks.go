package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Phase identifies when a hook runs.
type Phase string

const (
	PhasePreTool    Phase = "pre_tool"    // before any tool executes
	PhasePostTool   Phase = "post_tool"   // after any tool executes
	PhaseSessionStart Phase = "session_start"
	PhasePreCompact  Phase = "pre_compact"
	PhasePostCompact Phase = "post_compact"
	PhaseFileChanged Phase = "file_changed"
)

// HookDef defines a hook shell command triggered on a phase.
type HookDef struct {
	Phase   Phase    `json:"phase"`
	Command []string `json:"command"` // [executable, arg1, arg2, ...]
	Timeout int      `json:"timeout_ms"`
}

// Event is passed to hooks as JSON via stdin.
type Event struct {
	Phase    Phase          `json:"phase"`
	Tool     string         `json:"tool,omitempty"`
	Input    map[string]any `json:"input,omitempty"`
	Output   string         `json:"output,omitempty"`
	IsError  bool           `json:"is_error,omitempty"`
	FilePath string         `json:"file_path,omitempty"`
}

// Result is the hook's response (parsed from stdout JSON).
type Result struct {
	Allow   *bool  `json:"allow,omitempty"`   // nil = no opinion
	Reason  string `json:"reason,omitempty"`
	Mutated map[string]any `json:"mutated_input,omitempty"`
}

// Runner executes hooks for a given phase.
type Runner struct {
	hooks   []HookDef
	workDir string
}

// New creates a Runner from hook definitions.
func New(hooks []HookDef, workDir string) *Runner {
	return &Runner{hooks: hooks, workDir: workDir}
}

// LoadFromSettings reads hook definitions from .claude/settings.json.
func LoadFromSettings(projectDir string) []HookDef {
	path := filepath.Join(projectDir, ".claude", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var settings struct {
		Hooks map[string][]string `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil
	}

	var defs []HookDef
	for phase, cmds := range settings.Hooks {
		for _, cmd := range cmds {
			defs = append(defs, HookDef{
				Phase:   Phase(phase),
				Command: []string{"bash", "-c", cmd},
				Timeout: 30000,
			})
		}
	}
	return defs
}

// Run executes all hooks matching the given phase.
// Returns the last non-nil result from any hook.
func (r *Runner) Run(ctx context.Context, ev Event) (*Result, error) {
	payload, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}

	var lastResult *Result
	for _, h := range r.hooks {
		if h.Phase != ev.Phase {
			continue
		}
		res, err := r.runOne(ctx, h, payload)
		if err != nil {
			// Non-fatal: log and continue
			fmt.Fprintf(os.Stderr, "hook error [%s]: %v\n", h.Phase, err)
			continue
		}
		if res != nil {
			lastResult = res
		}
	}
	return lastResult, nil
}

func (r *Runner) runOne(ctx context.Context, h HookDef, payload []byte) (*Result, error) {
	timeout := time.Duration(h.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if len(h.Command) == 0 {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx, h.Command[0], h.Command[1:]...)
	cmd.Dir = r.workDir
	cmd.Stdin = bytes.NewReader(payload)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("hook timed out")
		}
		return nil, err
	}

	if out.Len() == 0 {
		return nil, nil
	}

	var result Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, nil // hook output not JSON — ignore
	}
	return &result, nil
}
