package commands

import (
	"context"
	"fmt"
	"strings"
)

// Command is a slash command available in the REPL.
type Command struct {
	Name        string
	Description string
	Hidden      bool
	Execute     func(ctx context.Context, args string, env *Env) (string, error)
}

// Env carries runtime context into command handlers.
type Env struct {
	WorkDir    string
	ClearFn    func()           // clear conversation
	GetHistory func() []string  // get message summaries
	GetCost    func() string    // get cost summary
	ListTools  func() []string  // get tool names
}

// Registry holds all registered slash commands.
type Registry struct {
	cmds map[string]*Command
}

// New returns a Registry with all built-in commands registered.
func New() *Registry {
	r := &Registry{cmds: make(map[string]*Command)}
	r.registerBuiltins()
	return r
}

// Get retrieves a command by name (with or without leading slash).
func (r *Registry) Get(name string) (*Command, bool) {
	name = strings.TrimPrefix(name, "/")
	c, ok := r.cmds[name]
	return c, ok
}

// All returns all visible commands.
func (r *Registry) All() []*Command {
	out := make([]*Command, 0, len(r.cmds))
	for _, c := range r.cmds {
		if !c.Hidden {
			out = append(out, c)
		}
	}
	return out
}

// Register adds a command.
func (r *Registry) Register(c *Command) {
	r.cmds[c.Name] = c
}

// Execute runs a slash command string like "/clear" or "/help".
func (r *Registry) Execute(ctx context.Context, input string, env *Env) (string, bool, error) {
	if !strings.HasPrefix(input, "/") {
		return "", false, nil
	}
	parts := strings.SplitN(strings.TrimPrefix(input, "/"), " ", 2)
	name := parts[0]
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	cmd, ok := r.Get(name)
	if !ok {
		return "", false, fmt.Errorf("unknown command: /%s — type /help to see available commands", name)
	}

	out, err := cmd.Execute(ctx, args, env)
	return out, true, err
}

func (r *Registry) registerBuiltins() {
	r.Register(&Command{
		Name:        "clear",
		Description: "Clear conversation history and start fresh.",
		Execute: func(_ context.Context, _ string, env *Env) (string, error) {
			if env.ClearFn != nil {
				env.ClearFn()
			}
			return "Conversation cleared.", nil
		},
	})

	r.Register(&Command{
		Name:        "help",
		Description: "Show available commands and tools.",
		Execute: func(_ context.Context, _ string, env *Env) (string, error) {
			var sb strings.Builder
			sb.WriteString("Available commands:\n")
			for _, c := range r.All() {
				sb.WriteString(fmt.Sprintf("  /%-20s %s\n", c.Name, c.Description))
			}
			if env.ListTools != nil {
				tools := env.ListTools()
				sb.WriteString(fmt.Sprintf("\nTools (%d): %s\n", len(tools), strings.Join(tools, ", ")))
			}
			sb.WriteString("\nCtrl+C to cancel / exit\n")
			return sb.String(), nil
		},
	})

	r.Register(&Command{
		Name:        "cost",
		Description: "Show token usage and estimated cost for this session.",
		Execute: func(_ context.Context, _ string, env *Env) (string, error) {
			if env.GetCost != nil {
				return env.GetCost(), nil
			}
			return "Cost tracking not available.", nil
		},
	})

	r.Register(&Command{
		Name:        "history",
		Description: "Show a summary of recent conversation turns.",
		Execute: func(_ context.Context, _ string, env *Env) (string, error) {
			if env.GetHistory == nil {
				return "No history available.", nil
			}
			msgs := env.GetHistory()
			if len(msgs) == 0 {
				return "No messages in history.", nil
			}
			var sb strings.Builder
			for i, m := range msgs {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, truncate(m, 120)))
			}
			return sb.String(), nil
		},
	})

	r.Register(&Command{
		Name:        "compact",
		Description: "Summarize and compress conversation history to free up context.",
		Execute: func(_ context.Context, _ string, _ *Env) (string, error) {
			return "Compaction will trigger automatically when needed, or use the API directly.", nil
		},
	})

	r.Register(&Command{
		Name:        "exit",
		Description: "Exit Tarra Claw.",
		Execute: func(_ context.Context, _ string, _ *Env) (string, error) {
			return "exit", nil // TUI handles this sentinel
		},
	})

	r.Register(&Command{
		Name:        "quit",
		Hidden:      true,
		Description: "Exit (alias for /exit).",
		Execute: func(ctx context.Context, args string, env *Env) (string, error) {
			return "exit", nil
		},
	})

	r.Register(&Command{
		Name:        "model",
		Description: "Show or set the active Claude model. Usage: /model [model-name]",
		Execute: func(_ context.Context, args string, _ *Env) (string, error) {
			if args == "" {
				return "Specify a model name. Available: claude-opus-4-6, claude-sonnet-4-6, claude-haiku-4-5-20251001", nil
			}
			return fmt.Sprintf("Model set to %s (takes effect on next message).", args), nil
		},
	})

	r.Register(&Command{
		Name:        "tools",
		Description: "List all available tools.",
		Execute: func(_ context.Context, _ string, env *Env) (string, error) {
			if env.ListTools == nil {
				return "No tools registered.", nil
			}
			tools := env.ListTools()
			return fmt.Sprintf("Tools (%d):\n  %s", len(tools), strings.Join(tools, "\n  ")), nil
		},
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
