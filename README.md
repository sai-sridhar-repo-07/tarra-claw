# Tarra Claw

> AI agent CLI harness in Go — streaming tools, MCP support, Bubble Tea TUI. Clean-room reimplementation with novelty.

**What Claude Code would look like if it were built in Go. Concurrent. Single binary. Fast.**

---

## Architecture

```
cmd/claw/         — Entry point (cobra CLI)
internal/
  api/            — Anthropic SDK wrapper + SSE streaming
  engine/         — Query loop + concurrent tool orchestration  
  tools/          — Bash, Read, Write, Edit, Glob, Grep, LS, Agent, MCP...
  tui/            — Bubble Tea interactive REPL
  commands/       — Slash command registry (/clear, /compact, /help...)
  hooks/          — Pre/post tool hook pipeline
  memory/         — Persistent memory system
  mcp/            — MCP protocol client
  config/         — Viper-based config (~/.config/tarra-claw/config.yaml)
  state/          — App state management
  permissions/    — Tool permission model
```

## Why Go?

| Claude Code (TypeScript) | Tarra Claw (Go)              |
|--------------------------|------------------------------|
| Ink (React TUI)          | Bubble Tea                   |
| chalk                    | Lip Gloss                    |
| marked                   | Glamour                      |
| Node streams             | goroutines + channels        |
| EventEmitter hooks       | channel middleware            |
| MCP JS SDK               | mcp-go                       |
| Bun subprocess           | os/exec                      |

- **Single static binary** — no Node, no Bun, no npm
- **Goroutines** — parallel tool execution natively
- **Fast startup** — <50ms cold start
- **Cross-platform** — compile anywhere, run anywhere

## Getting Started

```bash
# Install
go install github.com/sai-sridhar-repo-07/tarra-claw/cmd/claw@latest

# Set API key
export ANTHROPIC_API_KEY=your-key-here

# Interactive mode
claw

# Single prompt
claw run "explain this codebase"
```

## Configuration

```yaml
# ~/.config/tarra-claw/config.yaml
model: claude-opus-4-6
max_tokens: 8096
auto_approve: false
```

## Commands

| Command   | Description                        |
|-----------|------------------------------------|
| `/clear`  | Clear conversation history         |
| `/help`   | Show available tools and commands  |
| `/quit`   | Exit                               |
| `Ctrl+C`  | Cancel current operation / exit    |

## Tools

| Tool    | Permission | Description                        |
|---------|------------|------------------------------------|
| `Bash`  | Required   | Execute shell commands             |
| `Read`  | Auto       | Read file contents                 |
| `Write` | Required   | Create or overwrite files          |
| `Edit`  | Required   | Exact string replacement in files  |
| `Glob`  | Auto       | Find files by pattern              |
| `Grep`  | Auto       | Search file contents (ripgrep)     |
| `LS`    | Auto       | List directory contents            |

## Roadmap

- [ ] MCP server protocol support
- [ ] Agent sub-process tool
- [ ] Memory system (memdir)
- [ ] Hook pipeline (pre/post tool)
- [ ] Parallel tool execution
- [ ] `/compact` conversation compression
- [ ] Voice input
- [ ] Remote/server mode
- [ ] VSCode extension

## License

MIT
