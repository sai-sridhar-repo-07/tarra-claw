package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
)

// TodoStatus represents the state of a todo item.
type TodoStatus string

const (
	TodoPending    TodoStatus = "pending"
	TodoInProgress TodoStatus = "in_progress"
	TodoCompleted  TodoStatus = "completed"
)

// TodoItem is a single task in the todo list.
type TodoItem struct {
	ID      string     `json:"id"`
	Subject string     `json:"subject"`
	Status  TodoStatus `json:"status"`
	Notes   string     `json:"notes,omitempty"`
}

// TodoWriteTool manages the agent's task list.
type TodoWriteTool struct {
	stateDir string
}

func NewTodoWriteTool(workDir string) *TodoWriteTool {
	return &TodoWriteTool{stateDir: filepath.Join(workDir, ".claude")}
}

func (t *TodoWriteTool) Name() string        { return "TodoWrite" }
func (t *TodoWriteTool) NeedsPermission() bool { return false }

func (t *TodoWriteTool) Description() string {
	return `Manage the current task list. Write the complete updated todo list.
Use to track progress on multi-step tasks. Always include all todos (completed and pending).`
}

func (t *TodoWriteTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"todos": map[string]any{
				"type":        "array",
				"description": "Complete list of todos.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":      map[string]any{"type": "string"},
						"subject": map[string]any{"type": "string"},
						"status":  map[string]any{"type": "string", "enum": []string{"pending", "in_progress", "completed"}},
						"notes":   map[string]any{"type": "string"},
					},
					"required": []string{"id", "subject", "status"},
				},
			},
		},
		"required": []string{"todos"},
	}
}

func (t *TodoWriteTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{Name: t.Name(), Description: t.Description(), InputSchema: t.InputSchema()}
}

func (t *TodoWriteTool) Execute(_ context.Context, input map[string]any) (string, error) {
	raw, ok := input["todos"]
	if !ok {
		return "", fmt.Errorf("todos is required")
	}

	// Re-marshal to get proper typed slice
	b, _ := json.Marshal(raw)
	var todos []TodoItem
	if err := json.Unmarshal(b, &todos); err != nil {
		return "", fmt.Errorf("invalid todos format: %w", err)
	}

	if err := os.MkdirAll(t.stateDir, 0755); err != nil {
		return "", err
	}

	data, _ := json.MarshalIndent(todos, "", "  ")
	path := filepath.Join(t.stateDir, "todos.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	// Format summary
	var sb strings.Builder
	pending, inProg, done := 0, 0, 0
	for _, td := range todos {
		switch td.Status {
		case TodoPending:
			pending++
		case TodoInProgress:
			inProg++
		case TodoCompleted:
			done++
		}
	}
	sb.WriteString(fmt.Sprintf("Todo list updated: %d pending, %d in progress, %d completed\n", pending, inProg, done))
	for _, td := range todos {
		mark := "[ ]"
		switch td.Status {
		case TodoInProgress:
			mark = "[~]"
		case TodoCompleted:
			mark = "[x]"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", mark, td.Subject))
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}
