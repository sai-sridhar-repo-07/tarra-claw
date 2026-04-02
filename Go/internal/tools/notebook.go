package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sai-sridhar-repo-07/forge/internal/api"
)

// NotebookEditTool edits Jupyter notebook cells.
type NotebookEditTool struct{}

func NewNotebookEditTool() *NotebookEditTool { return &NotebookEditTool{} }

func (t *NotebookEditTool) Name() string        { return "NotebookEdit" }
func (t *NotebookEditTool) NeedsPermission() bool { return true }
func (t *NotebookEditTool) Description() string {
	return `Edit a Jupyter notebook cell. Modes: replace (default), insert, delete.`
}
func (t *NotebookEditTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"notebook_path": map[string]any{"type": "string", "description": "Path to .ipynb file."},
			"cell_id":       map[string]any{"type": "integer", "description": "Zero-based cell index."},
			"new_source":    map[string]any{"type": "string", "description": "New cell source code."},
			"cell_type":     map[string]any{"type": "string", "enum": []string{"code", "markdown"}, "description": "Cell type for insert mode."},
			"edit_mode":     map[string]any{"type": "string", "enum": []string{"replace", "insert", "delete"}, "description": "Edit operation."},
		},
		"required": []string{"notebook_path", "cell_id"},
	}
}
func (t *NotebookEditTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{Name: t.Name(), Description: t.Description(), InputSchema: t.InputSchema()}
}

func (t *NotebookEditTool) Execute(_ context.Context, input map[string]any) (string, error) {
	path, _ := input["notebook_path"].(string)
	if path == "" {
		return "", fmt.Errorf("notebook_path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read notebook: %w", err)
	}

	var nb map[string]any
	if err := json.Unmarshal(data, &nb); err != nil {
		return "", fmt.Errorf("invalid notebook JSON: %w", err)
	}

	cells, ok := nb["cells"].([]any)
	if !ok {
		return "", fmt.Errorf("notebook has no cells")
	}

	cellID := toInt(input["cell_id"])
	mode, _ := input["edit_mode"].(string)
	if mode == "" {
		mode = "replace"
	}

	switch mode {
	case "replace":
		if cellID < 0 || cellID >= len(cells) {
			return "", fmt.Errorf("cell_id %d out of range (notebook has %d cells)", cellID, len(cells))
		}
		cell, ok := cells[cellID].(map[string]any)
		if !ok {
			return "", fmt.Errorf("invalid cell format")
		}
		newSource, _ := input["new_source"].(string)
		cell["source"] = newSource
		cells[cellID] = cell

	case "insert":
		newSource, _ := input["new_source"].(string)
		cellType, _ := input["cell_type"].(string)
		if cellType == "" {
			cellType = "code"
		}
		newCell := map[string]any{
			"cell_type": cellType,
			"source":    newSource,
			"metadata":  map[string]any{},
			"outputs":   []any{},
		}
		after := make([]any, 0, len(cells)+1)
		after = append(after, cells[:cellID]...)
		after = append(after, newCell)
		after = append(after, cells[cellID:]...)
		cells = after

	case "delete":
		if cellID < 0 || cellID >= len(cells) {
			return "", fmt.Errorf("cell_id %d out of range", cellID)
		}
		cells = append(cells[:cellID], cells[cellID+1:]...)
	}

	nb["cells"] = cells
	out, err := json.MarshalIndent(nb, "", " ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return "", err
	}
	return fmt.Sprintf("Notebook %s updated (mode: %s, cell: %d)", path, mode, cellID), nil
}
