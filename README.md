# Tarra Claw

> AI agent CLI in Go — works free with Ollama (no API key) or with Anthropic Claude.

**What Claude Code would look like if it were built in Go. Concurrent. Single binary. Free.**

---

## Install

### Option 1 — Download binary (easiest)

Go to [Releases](https://github.com/sai-sridhar-repo-07/tarra-claw/releases) and download the binary for your platform:

| Platform | File |
|---|---|
| Mac (Apple Silicon) | `claw_*_darwin_arm64.tar.gz` |
| Mac (Intel) | `claw_*_darwin_amd64.tar.gz` |
| Linux | `claw_*_linux_amd64.tar.gz` |
| Windows | `claw_*_windows_amd64.zip` |

```bash
# Mac example
tar -xzf claw_*_darwin_arm64.tar.gz
sudo mv claw /usr/local/bin/
claw --help
```

### Option 2 — Go install

```bash
go install github.com/sai-sridhar-repo-07/tarra-claw/Go/cmd/claw@latest
```

---

## Quick Start (Free — no API key needed)

```bash
# 1. Install Ollama
brew install ollama        # Mac
# Linux: curl -fsSL https://ollama.com/install.sh | sh

# 2. Start Ollama and pull a model
ollama serve &
ollama pull llama3.2       # 2GB general model
# or: ollama pull qwen2.5-coder   (best for coding)

# 3. Run
claw
```

That's it. The AI runs **100% on your machine**. No internet after setup. No cost. Ever.

---

## Quick Start (Anthropic Claude)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
claw
```

Tarra Claw auto-detects which to use — Ollama if no API key is set, Anthropic if it is.

---

## Usage

```
claw                          # interactive TUI (default)
claw run "explain this repo"  # single prompt, exits after
claw models                   # list available models
```

### Inside the TUI

| Key | Action |
|---|---|
| Type + `Enter` | Send message |
| `↑` / `↓` | Browse message history |
| `Ctrl+A` / `Ctrl+E` | Jump to line start/end |
| `Ctrl+U` | Clear input |
| `Ctrl+C` | Cancel operation or exit |

### Slash commands

| Command | Description |
|---|---|
| `/clear` | Clear conversation history |
| `/cost` | Show token usage and cost |
| `/tools` | List all available tools |
| `/model <name>` | Switch model |
| `/help` | Show all commands |
| `/exit` | Quit |

---

## Configuration

Create `~/.config/tarra-claw/config.yaml`:

```yaml
# Provider: "ollama" or "anthropic" (auto-detected if not set)
provider: ollama

# Ollama settings
ollama_host: http://localhost:11434
ollama_model: qwen2.5-coder

# Anthropic settings (if using Claude)
# api_key: sk-ant-...
# model: claude-opus-4-6

max_tokens: 8096
auto_approve: false   # set true to skip permission prompts
```

Or use environment variables:
```bash
TARRA_PROVIDER=ollama claw
TARRA_OLLAMA_MODEL=mistral claw
ANTHROPIC_API_KEY=sk-ant-... claw
```

---

## Tools

| Tool | Description |
|---|---|
| `Bash` | Execute shell commands |
| `Read` | Read file contents with line numbers |
| `Write` | Create or overwrite files |
| `Edit` | Exact string replacement in files |
| `Glob` | Find files by pattern |
| `Grep` | Search file contents (ripgrep) |
| `LS` | List directory contents |
| `WebFetch` | Fetch and read any URL |
| `AskUserQuestion` | Ask user for input during a task |
| `TodoWrite` | Manage task lists |
| `TaskCreate/List/Get/Stop` | Background task management |
| `NotebookEdit` | Edit Jupyter notebook cells |

---

## Best Models for Coding (Ollama)

```bash
ollama pull qwen2.5-coder:7b      # best balance (4GB)
ollama pull deepseek-coder-v2     # most capable (8GB)
ollama pull mistral               # fast (4GB)
ollama pull llama3.2              # lightest (2GB)
```

---

## Architecture

```
Go/
├── cmd/claw/             Entry point
└── internal/
    ├── api/              Provider interface + Anthropic + Ollama
    ├── engine/           Agentic query loop + tool orchestration
    ├── tools/            16 built-in tools
    ├── tui/              Bubble Tea interactive REPL
    ├── commands/         Slash command registry
    ├── config/           Viper config + auto-detection
    ├── cost/             Token usage + cost tracking
    ├── history/          Session persistence
    ├── hooks/            Pre/post tool hook pipeline
    ├── mcp/              MCP protocol client
    ├── memory/           CLAUDE.md + auto-memory
    ├── permissions/      Tool permission rule engine
    ├── state/            Reactive app state
    ├── tasks/            Background task registry
    └── compact/          Context compaction
```

## Why Go?

| | Claude Code (TS) | Tarra Claw (Go) |
|---|---|---|
| Startup | ~2s | <50ms |
| API key required | Yes | No (Ollama) |
| Runs offline | No | Yes |
| Binary size | 200MB+ (Node) | 16MB |
| Language | TypeScript | Go |

---

## License

MIT — free to use, modify, and distribute.
