package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/sai-sridhar-repo-07/forge/internal/api"
	"github.com/sai-sridhar-repo-07/forge/internal/tasks"
)

// TaskCreateTool creates a new task in the registry.
type TaskCreateTool struct{ registry *tasks.Registry }

func NewTaskCreateTool(r *tasks.Registry) *TaskCreateTool { return &TaskCreateTool{registry: r} }

func (t *TaskCreateTool) Name() string        { return "TaskCreate" }
func (t *TaskCreateTool) NeedsPermission() bool { return false }
func (t *TaskCreateTool) Description() string {
	return "Create a new background task. Returns the task ID for tracking."
}
func (t *TaskCreateTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"subject":     map[string]any{"type": "string", "description": "Short task name."},
			"description": map[string]any{"type": "string", "description": "Detailed description."},
			"type":        map[string]any{"type": "string", "enum": []string{"local_bash", "local_agent"}, "description": "Task type."},
		},
		"required": []string{"subject"},
	}
}
func (t *TaskCreateTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{Name: t.Name(), Description: t.Description(), InputSchema: t.InputSchema()}
}
func (t *TaskCreateTool) Execute(_ context.Context, input map[string]any) (string, error) {
	subject, _ := input["subject"].(string)
	if subject == "" {
		return "", fmt.Errorf("subject is required")
	}
	desc, _ := input["description"].(string)
	typ := tasks.TypeLocalBash
	if tv, ok := input["type"].(string); ok {
		typ = tasks.Type(tv)
	}
	task := t.registry.Create(subject, desc, typ)
	return fmt.Sprintf("Task created: %s (id: %s)", task.Subject, task.ID), nil
}

// TaskListTool lists all tasks in the registry.
type TaskListTool struct{ registry *tasks.Registry }

func NewTaskListTool(r *tasks.Registry) *TaskListTool { return &TaskListTool{registry: r} }

func (t *TaskListTool) Name() string        { return "TaskList" }
func (t *TaskListTool) NeedsPermission() bool { return false }
func (t *TaskListTool) Description() string  { return "List all tasks and their current status." }
func (t *TaskListTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}
func (t *TaskListTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{Name: t.Name(), Description: t.Description(), InputSchema: t.InputSchema()}
}
func (t *TaskListTool) Execute(_ context.Context, _ map[string]any) (string, error) {
	all := t.registry.List()
	if len(all) == 0 {
		return "No tasks.", nil
	}
	var sb strings.Builder
	for _, task := range all {
		sb.WriteString(fmt.Sprintf("[%s] %-12s %-10s %s\n", task.ID, task.Status, task.Type, task.Subject))
	}
	return sb.String(), nil
}

// TaskGetTool retrieves a single task's details.
type TaskGetTool struct{ registry *tasks.Registry }

func NewTaskGetTool(r *tasks.Registry) *TaskGetTool { return &TaskGetTool{registry: r} }

func (t *TaskGetTool) Name() string        { return "TaskGet" }
func (t *TaskGetTool) NeedsPermission() bool { return false }
func (t *TaskGetTool) Description() string  { return "Get details and output for a specific task." }
func (t *TaskGetTool) InputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{"task_id": map[string]any{"type": "string"}},
		"required":   []string{"task_id"},
	}
}
func (t *TaskGetTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{Name: t.Name(), Description: t.Description(), InputSchema: t.InputSchema()}
}
func (t *TaskGetTool) Execute(_ context.Context, input map[string]any) (string, error) {
	id, _ := input["task_id"].(string)
	task, ok := t.registry.Get(id)
	if !ok {
		return "", fmt.Errorf("task %s not found", id)
	}
	output := strings.Join(task.Output, "\n")
	if output == "" {
		output = "(no output yet)"
	}
	return fmt.Sprintf("ID: %s\nSubject: %s\nStatus: %s\nType: %s\n\nOutput:\n%s",
		task.ID, task.Subject, task.Status, task.Type, output), nil
}

// TaskStopTool cancels a running task.
type TaskStopTool struct{ registry *tasks.Registry }

func NewTaskStopTool(r *tasks.Registry) *TaskStopTool { return &TaskStopTool{registry: r} }

func (t *TaskStopTool) Name() string        { return "TaskStop" }
func (t *TaskStopTool) NeedsPermission() bool { return false }
func (t *TaskStopTool) Description() string  { return "Stop/cancel a running task." }
func (t *TaskStopTool) InputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{"task_id": map[string]any{"type": "string"}},
		"required":   []string{"task_id"},
	}
}
func (t *TaskStopTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{Name: t.Name(), Description: t.Description(), InputSchema: t.InputSchema()}
}
func (t *TaskStopTool) Execute(_ context.Context, input map[string]any) (string, error) {
	id, _ := input["task_id"].(string)
	if err := t.registry.Stop(id); err != nil {
		return "", err
	}
	return fmt.Sprintf("Task %s stopped.", id), nil
}
