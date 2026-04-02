<div align="center">

```
    __________  ____  ____________
   / ____/ __ \/ __ \/ ____/ ____/
  / /_  / / / / /_/ / / __/ __/
 / __/ / /_/ / _, _/ /_/ / /___
/_/    \____/_/ |_|\____/_____/
```

**Catch bugs before you commit. Write commit messages automatically.**

[![Build](https://github.com/sai-sridhar-repo-07/tarra-claw/actions/workflows/build.yml/badge.svg)](https://github.com/sai-sridhar-repo-07/tarra-claw/actions)
[![Release](https://img.shields.io/github/v/release/sai-sridhar-repo-07/tarra-claw?color=6ee7b7&label=latest)](https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-a78bfa)](LICENSE)
[![Stars](https://img.shields.io/github/stars/sai-sridhar-repo-07/tarra-claw?style=social)](https://github.com/sai-sridhar-repo-07/tarra-claw/stargazers)

<br/>

![Demo](docs/demo.gif)

</div>

---

## The two commands you'll use every day

### `forge review` — AI finds bugs in your code before you commit

```bash
forge review --staged
```

```
## Issues Found
1. divide-by-zero — line 11: no zero check, causes runtime panic
2. ignored error  — line 17: os.ReadFile() error silently discarded
3. SQL injection   — line 23: string concatenation with user input is dangerous

## Verdict
❌ Needs changes — has bugs or security issues that must be fixed
```

Run it before every `git commit`. Catches what you miss.

---

### `forge commit` — AI writes your commit message

```bash
git add .
forge commit
```

```
fix: add input validation and SQL injection protection

- Replace string concatenation in SQL query with parameterized query
- Add error handling for os.ReadFile() calls
- Add divide-by-zero guard in divide()
```

No more staring at the terminal thinking what to write.

---

## Install

<details open>
<summary>&nbsp;&nbsp;<b>🍎 &nbsp;Mac — Apple Silicon (M1 / M2 / M3 / M4)</b></summary>
<br/>

```bash
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/forge_v0.3.0_darwin_arm64.tar.gz \
  | tar xz && sudo mv forge_v0.3.0_darwin_arm64 /usr/local/bin/forge
```

</details>

<details>
<summary>&nbsp;&nbsp;<b>🍎 &nbsp;Mac — Intel</b></summary>
<br/>

```bash
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/forge_v0.3.0_darwin_amd64.tar.gz \
  | tar xz && sudo mv forge_v0.3.0_darwin_amd64 /usr/local/bin/forge
```

</details>

<details>
<summary>&nbsp;&nbsp;<b>🐧 &nbsp;Linux — x64</b></summary>
<br/>

```bash
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/forge_v0.3.0_linux_amd64.tar.gz \
  | tar xz && sudo mv forge_v0.3.0_linux_amd64 /usr/local/bin/forge
```

</details>

<details>
<summary>&nbsp;&nbsp;<b>🪟 &nbsp;Windows</b></summary>
<br/>

Download [`forge_v0.3.0_windows_amd64.zip`](https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest) → extract → run `forge.exe`

</details>

<details>
<summary>&nbsp;&nbsp;<b>🔧 &nbsp;Build from source</b></summary>
<br/>

```bash
git clone https://github.com/sai-sridhar-repo-07/tarra-claw.git
cd tarra-claw/Go
go build -o forge ./cmd/forge
sudo mv forge /usr/local/bin/
```

</details>

```bash
forge --version   # verify install
```

---

## Quick Start — get running in 60 seconds

### With Anthropic Claude (recommended to start)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
forge review --staged   # done. that's it.
```

Get a free API key at [console.anthropic.com](https://console.anthropic.com) — costs cents per review.

### With Ollama (free, runs on your machine)

```bash
brew install ollama && ollama serve &
ollama pull llama3.2
forge review --staged
```

**Auto-detection:** no API key set → uses Ollama free. API key set → uses Claude.

---

## More commands

```bash
forge                                      # full AI coding agent — interactive chat
forge run "explain what this repo does"    # one-shot task, exits when done
forge review --branch main                 # compare your branch vs main
forge models                               # list available AI models
```

Forge is also a full AI coding agent. Ask it to fix bugs, explain code, search files, run commands — it does all of it with your permission.

---

## ✨ Why Forge over Claude Code?

<table>
<tr>
<td width="25%" align="center"><br/><b>🎯 forge review</b><br/><br/>Structured bug report from your git diff. SQL injection, ignored errors, logic bugs — before they hit production.<br/><br/></td>
<td width="25%" align="center"><br/><b>✍️ forge commit</b><br/><br/>AI reads your staged changes and writes a proper conventional commit message. One command.<br/><br/></td>
<td width="25%" align="center"><br/><b>⚡ 16MB binary</b><br/><br/>No Node.js. No npm install. Download one file, run it. Starts in 50ms.<br/><br/></td>
<td width="25%" align="center"><br/><b>🔒 Works offline</b><br/><br/>Switch to Ollama and your code never leaves your machine. Perfect for private or company projects.<br/><br/></td>
</tr>
</table>

---

## 📊 Comparison

<div align="center">

|  | Claude Code | Copilot CLI | **Forge** |
|--|:--:|:--:|:--:|
| `review` command | ❌ | ❌ | ✅ |
| `commit` command | ❌ | ❌ | ✅ |
| Needs API key | ✅ | ✅ | ❌ optional |
| Works offline | ❌ | ❌ | ✅ |
| Install size | 200MB | 150MB | **16MB** |
| Startup time | ~2s | ~2s | **< 50ms** |

</div>

---

## ⌨️ Inside the chat

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `↑` / `↓` | Browse message history |
| `Ctrl+U` | Clear input |
| `Ctrl+C` | Cancel / exit |

| Command | Description |
|---------|-------------|
| `/clear` | Clear conversation history |
| `/cost` | Show token usage and cost |
| `/model <name>` | Switch AI model |
| `/help` | All commands |
| `/exit` | Quit |

---

## 🛠 Built-in Tools

The AI asks your permission before using any tool:

| Tool | What it does |
|------|-------------|
| `Bash` | Execute shell commands |
| `Read` | Read files with line numbers |
| `Write` | Create or overwrite files |
| `Edit` | Find-and-replace inside files |
| `Glob` | Find files by pattern (`**/*.go`) |
| `Grep` | Search file contents with regex |
| `WebFetch` | Fetch and read any URL |
| `TodoWrite` | Manage task checklists |

---

## 🤖 Ollama Model Guide

| Model | Size | Best for |
|-------|------|----------|
| `llama3.2` | 2 GB | General use, quick answers |
| `mistral` | 4 GB | Balanced speed + quality |
| `qwen2.5-coder:7b` | 4 GB | **Coding tasks** ← recommended |
| `deepseek-coder-v2` | 8 GB | Large codebases |

> 💡 8 GB RAM → `llama3.2` &nbsp;·&nbsp; 16 GB+ → `qwen2.5-coder`

---

## ⚙️ Configuration

`~/.config/forge/config.yaml` *(optional)*:

```yaml
provider: anthropic        # or "ollama"
ollama_model: qwen2.5-coder
max_tokens: 8096
auto_approve: false
```

Or env vars:
```bash
FORGE_PROVIDER=ollama forge review --staged
```

---

## 🏗 Architecture

```
Go/
├── cmd/forge/          entry point
└── internal/
    ├── api/            Anthropic + Ollama provider interface
    ├── engine/         agentic loop — send → stream → tools → repeat
    ├── tools/          16 built-in tools
    ├── tui/            Bubble Tea interactive REPL
    └── ...
```

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) · [Cobra](https://github.com/spf13/cobra) · [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go)

---

## 🤝 Contributing

```bash
git clone https://github.com/sai-sridhar-repo-07/tarra-claw.git
cd tarra-claw/Go && go run ./cmd/forge
```

---

<div align="center">

MIT License · Free to use, modify, and distribute

<br/>

[⬇️ Download](https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest) &nbsp;·&nbsp; [⭐ Star](https://github.com/sai-sridhar-repo-07/tarra-claw) &nbsp;·&nbsp; [🐛 Issues](https://github.com/sai-sridhar-repo-07/tarra-claw/issues)

*Made with ❤️ in Go*

</div>
