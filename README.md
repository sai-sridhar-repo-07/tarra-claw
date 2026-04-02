<div align="center">

# 🦀 Tarra Claw

### AI coding agent CLI — built in Go. Fast, free, and offline-capable.

[![Build](https://github.com/sai-sridhar-repo-07/tarra-claw/actions/workflows/build.yml/badge.svg)](https://github.com/sai-sridhar-repo-07/tarra-claw/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/sai-sridhar-repo-07/tarra-claw?color=brightgreen)](https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platforms](https://img.shields.io/badge/platform-Mac%20%7C%20Linux%20%7C%20Windows-lightgrey)](#install)

**What Claude Code would look like if it were built in Go.**  
Single binary. 16MB. Starts in 50ms. Works 100% offline with Ollama — no API key needed.

![Tarra Claw Demo](docs/demo.gif)

[Install](#install) · [Quick Start](#quick-start) · [Tools](#built-in-tools) · [Config](#configuration) · [Why Go?](#why-go)

</div>

---

## What is this?

Tarra Claw is an open-source AI agent you run in your terminal. Give it a task — it reads your files, runs commands, searches your codebase, fetches URLs, and gets it done. Like Claude Code or Cursor, but:

- **Free** — use Ollama with local AI models (zero cost, no account needed)
- **Private** — your code never leaves your machine when using Ollama
- **Fast** — single Go binary, no Node.js, no 200MB install, starts instantly
- **Flexible** — switch to Anthropic Claude when you need more power

---

## Install

### One-liner (Mac/Linux)

```bash
# Mac — Apple Silicon (M1/M2/M3/M4)
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.1_darwin_arm64.tar.gz | tar xz && sudo mv claw_v0.1.1_darwin_arm64 /usr/local/bin/claw

# Mac — Intel
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.1_darwin_amd64.tar.gz | tar xz && sudo mv claw_v0.1.1_darwin_amd64 /usr/local/bin/claw

# Linux — x64
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.1_linux_amd64.tar.gz | tar xz && sudo mv claw_v0.1.1_linux_amd64 /usr/local/bin/claw
```

```bash
# Verify
claw --version
```

### Windows

Download `claw_v0.1.1_windows_amd64.zip` from [Releases](https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest), extract, and run `claw.exe`.

### Build from source (requires Go 1.22+)

```bash
git clone https://github.com/sai-sridhar-repo-07/tarra-claw.git
cd tarra-claw/Go
go build -o claw ./cmd/claw
sudo mv claw /usr/local/bin/
```

---

## Quick Start

### Option A — Free with Ollama (no account, no API key, fully offline)

```bash
# 1. Install Ollama
brew install ollama                     # Mac
# Linux: curl -fsSL https://ollama.com/install.sh | sh

# 2. Start Ollama + pull a model
ollama serve &
ollama pull llama3.2                    # 2GB — fast general use
# or: ollama pull qwen2.5-coder        # 4GB — best for coding

# 3. Run Tarra Claw
claw
```

> The header shows `ollama · llama3.2` — your AI is running 100% on your machine.

### Option B — Anthropic Claude (more powerful, requires API key)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
claw
```

**Auto-detection:** If `ANTHROPIC_API_KEY` is set → uses Claude. If not → uses Ollama free.  
You can override: `TARRA_PROVIDER=ollama claw` or `TARRA_PROVIDER=anthropic claw`

---

## Usage

```bash
claw                           # open interactive chat (default)
claw run "fix the bug in main.go"   # one-shot, exits when done
claw models                    # list available models
claw --help                    # all flags
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `↑` / `↓` | Navigate message history |
| `Ctrl+A` | Jump to start of input |
| `Ctrl+E` | Jump to end of input |
| `Ctrl+U` | Clear input |
| `Ctrl+C` | Cancel / exit |

### Slash commands

| Command | Description |
|---------|-------------|
| `/clear` | Clear conversation history |
| `/cost` | Show token usage and cost |
| `/tools` | List all available tools |
| `/model <name>` | Switch model mid-session |
| `/help` | Show all commands |
| `/exit` | Quit |

---

## Built-in Tools

The AI has access to these tools and will ask your permission before using them:

| Tool | What it does |
|------|-------------|
| `Bash` | Run shell commands in your terminal |
| `Read` | Read files with line numbers |
| `Write` | Create or overwrite files |
| `Edit` | Precise find-and-replace inside files |
| `Glob` | Find files by pattern (`**/*.go`) |
| `Grep` | Search file contents with regex |
| `LS` | List directory contents |
| `WebFetch` | Fetch and read any URL |
| `AskUserQuestion` | Ask you for input mid-task |
| `TodoWrite` | Manage task checklists |
| `TaskCreate/List/Get/Stop` | Manage background tasks |
| `NotebookEdit` | Edit Jupyter notebook cells |

---

## Configuration

Optional — create `~/.config/tarra-claw/config.yaml`:

```yaml
# Provider: "ollama" or "anthropic" (auto-detected if not set)
provider: ollama

# Ollama settings
ollama_host: http://localhost:11434
ollama_model: qwen2.5-coder          # any model you've pulled

# Anthropic (if using Claude)
# api_key: sk-ant-...
# model: claude-opus-4-6

max_tokens: 8096
auto_approve: false                   # true = skip permission prompts
```

Or just use environment variables — no config file needed:

```bash
TARRA_PROVIDER=ollama claw
TARRA_OLLAMA_MODEL=mistral claw
ANTHROPIC_API_KEY=sk-ant-... claw
```

---

## Choosing an Ollama Model

| Model | Size | Best for |
|-------|------|----------|
| `llama3.2` | 2GB | General chat, quick answers |
| `mistral` | 4GB | Balanced speed + quality |
| `qwen2.5-coder:7b` | 4GB | **Coding tasks** (recommended) |
| `deepseek-coder-v2` | 8GB | Complex code, large files |

```bash
ollama pull qwen2.5-coder    # install a model
ollama list                  # see what you have
```

> Pick based on your RAM. 8GB RAM → `llama3.2`. 16GB+ → `qwen2.5-coder`.

---

## Why Go?

| | Claude Code (TypeScript) | Tarra Claw (Go) |
|--|--|--|
| Startup time | ~2 seconds | < 50ms |
| Binary size | 200MB+ (needs Node.js) | **16MB** single file |
| Requires API key | Yes, always | **No** — free with Ollama |
| Runs offline | No | **Yes** |
| Install method | `npm install -g` | Download one file |
| Privacy | Code sent to Anthropic | **Stays on your machine** |

---

## Architecture

```
Go/
├── cmd/claw/             Entry point
└── internal/
    ├── api/              Provider interface — Anthropic + Ollama backends
    ├── engine/           Agentic loop (send → stream → tools → repeat)
    ├── tools/            16 built-in tools
    ├── tui/              Bubble Tea interactive REPL
    ├── commands/         Slash command registry
    ├── config/           Viper config + auto-detection logic
    ├── cost/             Token usage + cost tracking
    ├── history/          Session persistence
    ├── hooks/            Pre/post tool hook pipeline
    ├── mcp/              MCP protocol client
    ├── memory/           Auto-memory system
    ├── permissions/      Tool permission engine
    ├── tasks/            Background task registry
    └── compact/          Context compaction
```

**Built with:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) · [Lip Gloss](https://github.com/charmbracelet/lipgloss) · [Cobra](https://github.com/spf13/cobra) · [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go) · [MCP Go](https://github.com/mark3labs/mcp-go)

---

## Contributing

PRs welcome. To run locally:

```bash
git clone https://github.com/sai-sridhar-repo-07/tarra-claw.git
cd tarra-claw/Go
go run ./cmd/claw
```

---

## License

MIT — free to use, modify, and distribute.

---

<div align="center">
Made with Go · <a href="https://github.com/sai-sridhar-repo-07/tarra-claw/releases">Download latest release</a>
</div>
