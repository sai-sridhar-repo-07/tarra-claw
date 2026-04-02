<div align="center">

```
███████╗ ██████╗ ██████╗  ██████╗ ███████╗
██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
█████╗  ██║   ██║██████╔╝██║  ███╗█████╗
██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝
██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝
```

**AI coding agent for your terminal — built in Go**

[![Build](https://github.com/sai-sridhar-repo-07/forge/actions/workflows/build.yml/badge.svg)](https://github.com/sai-sridhar-repo-07/forge/actions)
[![Release](https://img.shields.io/github/v/release/sai-sridhar-repo-07/forge?color=6ee7b7&label=latest)](https://github.com/sai-sridhar-repo-07/forge/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-a78bfa)](LICENSE)
[![Stars](https://img.shields.io/github/stars/sai-sridhar-repo-07/forge?style=social)](https://github.com/sai-sridhar-repo-07/forge/stargazers)

<br/>

![Demo](docs/demo.gif)

<br/>

> **Give it a task. It reads your files, runs commands, searches your code, and gets it done.**  
> Like Claude Code — but free, offline, and a single 16MB binary.

</div>

<br/>

---

## ✨ Why Forge?

<table>
<tr>
<td width="25%" align="center"><br/><b>🆓 Completely Free</b><br/><br/>Use Ollama — AI runs on your machine. No account. No API key. No cost. Ever.<br/><br/></td>
<td width="25%" align="center"><br/><b>🔒 Private by Default</b><br/><br/>With Ollama, your code never leaves your machine. Perfect for proprietary projects.<br/><br/></td>
<td width="25%" align="center"><br/><b>⚡ Instant Startup</b><br/><br/>50ms cold start. No Node.js runtime. One binary, drop it anywhere and run.<br/><br/></td>
<td width="25%" align="center"><br/><b>🔌 Dual Backend</b><br/><br/>Switch between free local Ollama and powerful Anthropic Claude with one env var.<br/><br/></td>
</tr>
</table>

<br/>

---

## 📊 How It Compares

<div align="center">

|  | Claude Code | Copilot CLI | **Forge** |
|--|:--:|:--:|:--:|
| Needs API key | ✅ | ✅ | ❌ **No** |
| Works offline | ❌ | ❌ | ✅ **Yes** |
| Code stays private | ❌ | ❌ | ✅ **Yes** |
| Startup time | ~2s | ~2s | **< 50ms** |
| Install size | 200MB | 150MB | **16MB** |
| Cost | $$ / token | $$ / month | **Free** |
| Language | TypeScript | TypeScript | **Go** |

</div>

<br/>

---

## 📥 Install

<details open>
<summary>&nbsp;&nbsp;<b>🍎 &nbsp;Mac — Apple Silicon (M1 / M2 / M3 / M4)</b></summary>
<br/>

```bash
curl -L https://github.com/sai-sridhar-repo-07/forge/releases/latest/download/forge_v0.2.0_darwin_arm64.tar.gz \
  | tar xz && sudo mv forge_v0.2.0_darwin_arm64 /usr/local/bin/claw
```

</details>

<details>
<summary>&nbsp;&nbsp;<b>🍎 &nbsp;Mac — Intel</b></summary>
<br/>

```bash
curl -L https://github.com/sai-sridhar-repo-07/forge/releases/latest/download/forge_v0.2.0_darwin_amd64.tar.gz \
  | tar xz && sudo mv forge_v0.2.0_darwin_amd64 /usr/local/bin/claw
```

</details>

<details>
<summary>&nbsp;&nbsp;<b>🐧 &nbsp;Linux — x64</b></summary>
<br/>

```bash
curl -L https://github.com/sai-sridhar-repo-07/forge/releases/latest/download/forge_v0.2.0_linux_amd64.tar.gz \
  | tar xz && sudo mv forge_v0.2.0_linux_amd64 /usr/local/bin/claw
```

</details>

<details>
<summary>&nbsp;&nbsp;<b>🪟 &nbsp;Windows</b></summary>
<br/>

1. Download [`forge_v0.2.0_windows_amd64.zip`](https://github.com/sai-sridhar-repo-07/forge/releases/latest)
2. Extract the zip
3. Run `claw.exe` from any terminal

</details>

<details>
<summary>&nbsp;&nbsp;<b>🔧 &nbsp;Build from source</b></summary>
<br/>

```bash
git clone https://github.com/sai-sridhar-repo-07/forge.git
cd forge/Go
go build -o forge ./cmd/claw
sudo mv forge /usr/local/bin/
```

</details>

<br/>

```bash
forge --help   # ✓ verify install
```

<br/>

---

## 🚀 Quick Start

### 🆓 Option A — Ollama (Free, no account, works offline)

```bash
# Step 1 — install Ollama
brew install ollama                              # Mac
curl -fsSL https://ollama.com/install.sh | sh   # Linux

# Step 2 — download a model
ollama serve &
ollama pull llama3.2          # 2GB · good starting point
# ollama pull qwen2.5-coder   # 4GB · best for coding tasks

# Step 3 — run
claw
```

The header shows **`ollama · llama3.2`** — AI is running 100% on your machine.

<br/>

### ⚡ Option B — Anthropic Claude (More powerful)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
claw
```

**Auto-detection** — Forge picks the right backend automatically:

```
No ANTHROPIC_API_KEY set  →  Ollama  (free, local)
ANTHROPIC_API_KEY is set  →  Claude  (Anthropic API)
```

Force a provider anytime: `FORGE_PROVIDER=ollama claw`

<br/>

---

## 💬 Usage

```bash
claw                                     # open interactive chat
forge run "explain what this repo does"   # one-shot task, auto-exits
forge models                              # list available AI models
```

### 🔍 Unique Commands (not in Claude Code)

**`forge review`** — AI reviews your git diff and finds bugs before you commit

```bash
forge review              # review all current changes
forge review --staged     # review only staged changes
forge review --branch main # compare your branch vs main
```

Output:
```
## Summary
Adds getUserInput() and constructs a SQL query from user input.

## Issues Found
1. divide-by-zero — line 11: no zero check, causes runtime panic
2. ignored error — line 17: os.ReadFile() error silently discarded
3. SQL injection — line 23: string concatenation with user input is dangerous

## Verdict
❌ Needs changes — has bugs or security issues that must be fixed
```

**`forge commit`** — AI writes your commit message from staged changes

```bash
git add .
forge commit
```

Output:
```
─────────────────────────────────
fix: add input validation and SQL injection protection

- Replace string concatenation in SQL query with parameterized query
- Add error handling for os.ReadFile() calls
- Add divide-by-zero guard in divide()
─────────────────────────────────

To use it:
  git commit -m "<paste message above>"
```

<br/>

### ⌨️ Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `↑` / `↓` | Browse message history |
| `Ctrl+A` | Jump to start of input |
| `Ctrl+E` | Jump to end of input |
| `Ctrl+U` | Clear input line |
| `Ctrl+C` | Cancel operation / exit |

<br/>

### 💡 Slash Commands

| Command | Description |
|---------|-------------|
| `/clear` | Clear conversation history |
| `/cost` | Show token usage and estimated cost |
| `/tools` | List all available tools |
| `/model <name>` | Switch AI model mid-session |
| `/help` | Show all slash commands |
| `/exit` | Quit |

<br/>

---

## 🛠 Built-in Tools

The AI uses tools to complete your tasks — it always asks permission first:

| Tool | What it does |
|------|-------------|
| `Bash` | Execute shell commands |
| `Read` | Read files with line numbers |
| `Write` | Create or overwrite files |
| `Edit` | Precise find-and-replace inside files |
| `Glob` | Find files by pattern (`**/*.go`) |
| `Grep` | Search file contents with regex |
| `LS` | List directory contents |
| `WebFetch` | Fetch and read any URL |
| `AskUser` | Ask you for input mid-task |
| `TodoWrite` | Manage task checklists |
| `Task*` | Create, list, and stop background tasks |
| `NotebookEdit` | Edit Jupyter notebook cells |

<br/>

---

## 🤖 Choosing an Ollama Model

| Model | Size | Best for |
|-------|------|----------|
| `llama3.2` | 2 GB | General chat, quick answers |
| `mistral` | 4 GB | Balanced speed and quality |
| `qwen2.5-coder:7b` | 4 GB | **Coding tasks** ← recommended |
| `deepseek-coder-v2` | 8 GB | Large codebases, complex reasoning |

```bash
ollama pull qwen2.5-coder    # download your chosen model
ollama list                  # see installed models
```

> 💡 **Pick by RAM:** 8 GB → `llama3.2` &nbsp; 16 GB+ → `qwen2.5-coder`

<br/>

---

## ⚙️ Configuration

Create `~/.config/forge/config.yaml` *(optional — everything works without it)*:

```yaml
# Provider: "ollama" or "anthropic" (auto-detected if omitted)
provider: ollama

# Ollama settings
ollama_host: http://localhost:11434
ollama_model: qwen2.5-coder

# Anthropic settings (uncomment to use Claude)
# api_key: sk-ant-...
# model: claude-opus-4-6

max_tokens: 8096
auto_approve: false    # set true to skip permission prompts
```

Or skip the file and use environment variables:

```bash
FORGE_PROVIDER=ollama FORGE_OLLAMA_MODEL=mistral claw
```

<br/>

---

## 🏗 Architecture

```
Go/
├── cmd/claw/           entry point
└── internal/
    ├── api/            provider interface · Anthropic · Ollama
    ├── engine/         agentic loop — send → stream → tools → repeat
    ├── tools/          16 built-in tools
    ├── tui/            Bubble Tea interactive REPL
    ├── commands/       slash command registry
    ├── config/         Viper config + auto-detection
    ├── cost/           token usage + cost tracking
    ├── history/        session persistence
    ├── hooks/          pre/post tool hook pipeline
    ├── mcp/            MCP protocol client
    ├── memory/         auto-memory system
    ├── permissions/    tool permission engine
    ├── tasks/          background task registry
    └── compact/        context compaction
```

**Built with:**
[Bubble Tea](https://github.com/charmbracelet/bubbletea) &nbsp;·&nbsp;
[Lip Gloss](https://github.com/charmbracelet/lipgloss) &nbsp;·&nbsp;
[Cobra](https://github.com/spf13/cobra) &nbsp;·&nbsp;
[Viper](https://github.com/spf13/viper) &nbsp;·&nbsp;
[Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go) &nbsp;·&nbsp;
[MCP Go](https://github.com/mark3labs/mcp-go)

<br/>

---

## 🤝 Contributing

Found a bug? Have an idea? PRs and issues are welcome.

```bash
git clone https://github.com/sai-sridhar-repo-07/forge.git
cd forge/Go
go run ./cmd/claw
```

<br/>

---

<div align="center">

MIT License &nbsp;·&nbsp; Free to use, modify, and distribute

<br/>

[⬇️ &nbsp; Download Latest Release](https://github.com/sai-sridhar-repo-07/forge/releases/latest) &nbsp;&nbsp;&nbsp; [⭐ &nbsp; Star on GitHub](https://github.com/sai-sridhar-repo-07/forge) &nbsp;&nbsp;&nbsp; [🐛 &nbsp; Report a Bug](https://github.com/sai-sridhar-repo-07/forge/issues)

<br/>

*Made with ❤️ in Go*

</div>
