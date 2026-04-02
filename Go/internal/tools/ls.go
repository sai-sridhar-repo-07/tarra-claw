package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sai-sridhar-repo-07/forge/internal/api"
)

// LSTool lists directory contents.
type LSTool struct {
	workDir string
}

func NewLSTool(workDir string) *LSTool { return &LSTool{workDir: workDir} }

func (t *LSTool) Name() string        { return "LS" }
func (t *LSTool) NeedsPermission() bool { return false }

func (t *LSTool) Description() string {
	return "List directory contents with file sizes and modification times."
}

func (t *LSTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Directory path to list. Defaults to working directory.",
			},
		},
	}
}

func (t *LSTool) Definition() api.ToolDefinition {
	return api.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

func (t *LSTool) Execute(_ context.Context, input map[string]any) (string, error) {
	dir := t.workDir
	if p, ok := input["path"].(string); ok && p != "" {
		dir = p
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("cannot list %s: %w", dir, err)
	}

	type entry struct {
		name    string
		size    int64
		modTime time.Time
		isDir   bool
	}

	var items []entry
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, entry{
			name:    e.Name(),
			size:    info.Size(),
			modTime: info.ModTime(),
			isDir:   e.IsDir(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].isDir != items[j].isDir {
			return items[i].isDir
		}
		return items[i].name < items[j].name
	})

	var sb strings.Builder
	sb.WriteString(dir + "\n")
	for _, item := range items {
		indicator := " "
		if item.isDir {
			indicator = "/"
		}
		sb.WriteString(fmt.Sprintf("%-40s %8s  %s\n",
			item.name+indicator,
			formatSize(item.size, item.isDir),
			item.modTime.Format("Jan 02 15:04"),
		))
	}
	return sb.String(), nil
}

func formatSize(size int64, isDir bool) string {
	if isDir {
		return "-"
	}
	switch {
	case size < 1024:
		return fmt.Sprintf("%dB", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.1fK", float64(size)/1024)
	default:
		return fmt.Sprintf("%.1fM", float64(size)/(1024*1024))
	}
}

var _ = filepath.Join // keep import
