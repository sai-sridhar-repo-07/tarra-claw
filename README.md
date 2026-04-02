<div align="center">

```
  ████████╗ █████╗ ██████╗ ██████╗  █████╗      ██████╗██╗      █████╗ ██╗    ██╗
     ██╔══╝██╔══██╗██╔══██╗██╔══██╗██╔══██╗    ██╔════╝██║     ██╔══██╗██║    ██║
     ██║   ███████║██████╔╝██████╔╝███████║    ██║     ██║     ███████║██║ █╗ ██║
     ██║   ██╔══██║██╔══██╗██╔══██╗██╔══██║    ██║     ██║     ██╔══██║██║███╗██║
     ██║   ██║  ██║██║  ██║██║  ██║██║  ██║    ╚██████╗███████╗██║  ██║╚███╔███╔╝
     ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝    ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
```

### AI coding agent for your terminal — built in Go

[![Build](https://github.com/sai-sridhar-repo-07/tarra-claw/actions/workflows/build.yml/badge.svg)](https://github.com/sai-sridhar-repo-07/tarra-claw/actions)
[![Release](https://img.shields.io/github/v/release/sai-sridhar-repo-07/tarra-claw?color=6ee7b7&label=latest)](https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-a78bfa)](LICENSE)
[![Stars](https://img.shields.io/github/stars/sai-sridhar-repo-07/tarra-claw?style=social)](https://github.com/sai-sridhar-repo-07/tarra-claw/stargazers)

![Demo](docs/demo.gif)

</div>

---

<div align="center">

### ⚡ 50ms startup &nbsp;·&nbsp; 🔒 100% offline &nbsp;·&nbsp; 🆓 Free with Ollama &nbsp;·&nbsp; 📦 16MB binary

</div>

---

## What is Tarra Claw?

An open-source AI agent you run in your terminal. Give it a task — it reads your files, runs commands, searches your codebase, fetches URLs, and gets it done. Like Claude Code, but:

| | Claude Code | **Tarra Claw** |
|--|--|--|
| Requires API key | ✅ Always | ❌ **Not needed** |
| Runs offline | ❌ No | ✅ **Yes** |
| Startup time | ~2 seconds | **< 50ms** |
| Install size | 200MB (Node.js) | **16MB** |
| Code privacy | Sent to Anthropic | **Stays on your machine** |
| Cost | Pay per token | **Free** (Ollama) |

---

## Install

<details open>
<summary><b>Mac — Apple Silicon (M1/M2/M3/M4)</b></summary>

```bash
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.1_darwin_arm64.tar.gz \
  | tar xz && sudo mv claw_v0.1.1_darwin_arm64 /usr/local/bin/claw
```
</details>

<details>
<summary><b>Mac — Intel</b></summary>

```bash
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.1_darwin_amd64.tar.gz \
  | tar xz && sudo mv claw_v0.1.1_darwin_amd64 /usr/local/bin/claw
```
</details>

<details>
<summary><b>Linux — x64</b></summary>

```bash
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.1_linux_amd64.tar.gz \
  | tar xz && sudo mv claw_v0.1.1_linux_amd64 /usr/local/bin/claw
```
</details>

<details>
<summary><b>Windows</b></summary>

Download [`claw_v0.1.1_windows_amd64.zip`](https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest) → extract → run `claw.exe`
</details>

<details>
<summary><b>Build from source</b></summary>

```bash
git clone https://github.com/sai-sridhar-repo-07/tarra-claw.git
cd tarra-claw/Go
go build -o claw ./cmd/claw
sudo mv claw /usr/local/bin/
```
</details>

```bash
claw --help   # verify install
```

---

## Quick Start

### 🆓 Option A — Free with Ollama (no account, works offline)

```bash
# 1. Install Ollama
brew install ollama                    # Mac
curl -fsSL https://ollama.com/install.sh | sh  # Linux

# 2. Pull a model and start
ollama serve &
ollama pull llama3.2                   # 2GB — start here

# 3. Launch
claw
```

> The header will show `ollama · llama3.2` — AI is running on your machine.

### ⚡ Option B — Anthropic Claude (more powerful)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
claw
```

**Auto-detection:** no key = Ollama free · key present = Anthropic Claude.  
Override anytime: `TARRA_PROVIDER=ollama claw`

---

## Usage

```
claw                              interactive chat (default)
claw run "fix the bug in main.go" one-shot task, exits when done
claw models                       list available models
```

### ⌨️ Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `↑` `↓` | Browse message history |
| `Ctrl+A` / `Ctrl+E` | Jump to line start / end |
| `Ctrl+U` | Clear input |
| `Ctrl+C` | Cancel / exit |

### `/` Slash Commands

```
/clear      clear conversation history
/cost       show token usage and cost
/tools      list all available tools
/model      switch model mid-session
/help       show all commands
/exit       quit
```

---

## 🛠 Built-in Tools

The AI asks your permission before using any tool:

```
Bash          ·  run shell commands
Read          ·  read files with line numbers
Write         ·  create or overwrite files
Edit          ·  precise find-and-replace in files
Glob          ·  find files by pattern  **/*.go
Grep          ·  search file contents with regex
LS            ·  list directory contents
WebFetch      ·  fetch and read any URL
AskUser       ·  ask you for input mid-task
TodoWrite     ·  manage task checklists
Task*         ·  create/list/stop background tasks
NotebookEdit  ·  edit Jupyter notebook cells
```

---

## 🤖 Choosing an Ollama Model

```bash
ollama pull llama3.2             # 2GB  · fast, good for general tasks
ollama pull mistral              # 4GB  · balanced speed + quality
ollama pull qwen2.5-coder:7b    # 4GB  · best for coding  ← recommended
ollama pull deepseek-coder-v2   # 8GB  · most capable for large codebases

ollama list                      # see what you have installed
```

> **Rule of thumb:** 8GB RAM → `llama3.2` · 16GB+ → `qwen2.5-coder`

---

## ⚙️ Configuration

Create `~/.config/tarra-claw/config.yaml` (optional):

```yaml
provider: ollama                  # "ollama" or "anthropic"
ollama_host: http://localhost:11434
ollama_model: qwen2.5-coder

# anthropic:
# api_key: sk-ant-...
# model: claude-opus-4-6

max_tokens: 8096
auto_approve: false               # true = skip permission prompts
```

Or use env vars — no config file needed:
```bash
TARRA_PROVIDER=ollama TARRA_OLLAMA_MODEL=mistral claw
```

---

## 🏗 Architecture

```
Go/
├── cmd/claw/          entry point
└── internal/
    ├── api/           provider interface  ·  Anthropic  ·  Ollama
    ├── engine/        agentic loop — send → stream → tools → repeat
    ├── tools/         16 built-in tools
    ├── tui/           Bubble Tea interactive REPL
    ├── commands/      slash command registry
    ├── config/        Viper config + auto-detection
    ├── cost/          token usage tracking
    ├── history/       session persistence
    ├── hooks/         pre/post tool hooks
    ├── mcp/           MCP protocol client
    ├── memory/        auto-memory system
    ├── permissions/   tool permission engine
    ├── tasks/         background task registry
    └── compact/       context compaction
```

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) · [Lip Gloss](https://github.com/charmbracelet/lipgloss) · [Cobra](https://github.com/spf13/cobra) · [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go)

---

## Contributing

```bash
git clone https://github.com/sai-sridhar-repo-07/tarra-claw.git
cd tarra-claw/Go && go run ./cmd/claw
```

PRs and issues welcome.

---

<div align="center">

**MIT License** · Free to use, modify, and distribute

[⬇ Download latest release](https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest) &nbsp;·&nbsp; [⭐ Star this repo](https://github.com/sai-sridhar-repo-07/tarra-claw)

</div>
