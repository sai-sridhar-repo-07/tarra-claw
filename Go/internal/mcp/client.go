package mcp

import (
	"context"
	"fmt"
	"sync"

	mcpsdk "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// ServerConfig defines how to connect to an MCP server.
type ServerConfig struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
}

// Tool represents a tool exposed by an MCP server.
type Tool struct {
	Server      string
	Name        string
	Description string
	InputSchema map[string]any
}

// Resource represents a resource exposed by an MCP server.
type Resource struct {
	Server  string
	URI     string
	Name    string
	MimeType string
}

// Manager manages connections to multiple MCP servers.
type Manager struct {
	mu      sync.RWMutex
	servers map[string]*serverConn
}

type serverConn struct {
	name   string
	client *mcpsdk.StdioMCPClient
	tools  []mcp.Tool
}

// New creates an empty Manager.
func New() *Manager {
	return &Manager{servers: make(map[string]*serverConn)}
}

// Connect starts an MCP server process and initializes the connection.
func (m *Manager) Connect(ctx context.Context, cfg ServerConfig) error {
	client, err := mcpsdk.NewStdioMCPClient(cfg.Command, cfg.Args)
	if err != nil {
		return fmt.Errorf("mcp connect %s: %w", cfg.Name, err)
	}

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "forge",
		Version: "0.1.0",
	}

	if _, err := client.Initialize(ctx, initReq); err != nil {
		return fmt.Errorf("mcp init %s: %w", cfg.Name, err)
	}

	// List tools
	listReq := mcp.ListToolsRequest{}
	toolsResult, err := client.ListTools(ctx, listReq)
	if err != nil {
		return fmt.Errorf("mcp list-tools %s: %w", cfg.Name, err)
	}

	m.mu.Lock()
	m.servers[cfg.Name] = &serverConn{
		name:   cfg.Name,
		client: client,
		tools:  toolsResult.Tools,
	}
	m.mu.Unlock()

	return nil
}

// AllTools returns all tools from all connected servers.
func (m *Manager) AllTools() []Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Tool
	for _, s := range m.servers {
		for _, t := range s.tools {
			out = append(out, Tool{
				Server:      s.name,
				Name:        fmt.Sprintf("%s__%s", s.name, t.Name),
				Description: t.Description,
			})
		}
	}
	return out
}

// Call executes a tool on an MCP server.
func (m *Manager) Call(ctx context.Context, serverName, toolName string, args map[string]any) (string, error) {
	m.mu.RLock()
	s, ok := m.servers[serverName]
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("mcp server %q not connected", serverName)
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = args

	result, err := s.client.CallTool(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mcp call %s/%s: %w", serverName, toolName, err)
	}

	if result.IsError {
		return "", fmt.Errorf("mcp tool error: %v", result.Content)
	}

	var out string
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			out += tc.Text
		}
	}
	return out, nil
}

// Disconnect stops a server connection.
func (m *Manager) Disconnect(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.servers[name]; ok {
		s.client.Close()
		delete(m.servers, name)
	}
}

// Connected returns names of all connected servers.
func (m *Manager) Connected() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.servers))
	for n := range m.servers {
		names = append(names, n)
	}
	return names
}
