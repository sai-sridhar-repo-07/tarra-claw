package tools

import (
	"context"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
)

// Tool is the interface every tool must implement.
type Tool interface {
	// Name returns the tool name exposed to the model.
	Name() string
	// Description returns a concise description for the model.
	Description() string
	// InputSchema returns the JSON schema for tool input.
	InputSchema() map[string]any
	// Execute runs the tool with the given input and returns a result string.
	Execute(ctx context.Context, input map[string]any) (string, error)
	// Definition returns the api.ToolDefinition for this tool.
	Definition() api.ToolDefinition
	// NeedsPermission returns true if this tool requires user approval.
	NeedsPermission() bool
}

// Result wraps a tool execution outcome.
type Result struct {
	ToolID  string
	Content string
	IsError bool
}

// Registry holds all registered tools.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// All returns all registered tools.
func (r *Registry) All() []Tool {
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

// Definitions returns api.ToolDefinition for all tools.
func (r *Registry) Definitions() []api.ToolDefinition {
	all := r.All()
	defs := make([]api.ToolDefinition, len(all))
	for i, t := range all {
		defs[i] = t.Definition()
	}
	return defs
}

// DefaultRegistry builds and returns a registry with all built-in tools.
func DefaultRegistry(workDir string) *Registry {
	r := NewRegistry()
	r.Register(NewBashTool(workDir))
	r.Register(NewReadTool())
	r.Register(NewWriteTool())
	r.Register(NewEditTool())
	r.Register(NewGlobTool(workDir))
	r.Register(NewGrepTool(workDir))
	r.Register(NewLSTool(workDir))
	return r
}
