# Tarra Claw

> AI agent CLI in Go — works free with Ollama (no API key) or with Anthropic Claude.

**What Claude Code would look like if it were built in Go. Concurrent. Single binary. Free.**

---

## Install

### Option 1 — One-liner (Mac/Linux)

```bash
# Mac (Apple Silicon — M1/M2/M3/M4)
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.0_darwin_arm64.tar.gz | tar xz && sudo mv claw /usr/local/bin/

# Mac (Intel)
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.0_darwin_amd64.tar.gz | tar xz && sudo mv claw /usr/local/bin/

# Linux (x64)
curl -L https://github.com/sai-sridhar-repo-07/tarra-claw/releases/latest/download/claw_v0.1.0_linux_amd64.tar.gz | tar xz && sudo mv claw /usr/local/bin/
```

Verify it works:
```bash
claw --version
```

### Option 2 — Download manually

Go to [**Releases**](https://github.com/sai-sridhar-repo-07/tarra-claw/releases) and pick your platform:

| Platform | File |
|---|---|
| Mac Apple Silicon (M1/M2/M3) | `claw_v0.1.0_darwin_arm64.tar.gz` |
| Mac Intel | `claw_v0.1.0_darwin_amd64.tar.gz` |
| Linux x64 | `claw_v0.1.0_linux_amd64.tar.gz` |
| Linux ARM | `claw_v0.1.0_linux_arm64.tar.gz` |
| Windows | `claw_v0.1.0_windows_amd64.zip` |

**Windows:** Extract the `.zip`, then run `claw.exe` from a terminal.

### Option 3 — Build from source (requires Go 1.22+)

```bash
git clone https://github.com/sai-sridhar-repo-07/tarra-claw.git
cd tarra-claw/Go
go build -o claw ./cmd/claw
sudo mv claw /usr/local/bin/
```

---

## Quick Start — Free with Ollama (no API key)

Ollama runs the AI model **100% on your machine**. No internet required after setup. No cost. Ever.

```bash
# Step 1: Install Ollama
brew install ollama                           # Mac
# Linux: curl -fsSL https://ollama.com/install.sh | sh

# Step 2: Start Ollama and pull a model
ollama serve &
ollama pull llama3.2                          # lightweight (2GB), good for general use
# or for coding:
ollama pull qwen2.5-coder                     # best for code tasks (4GB)

# Step 3: Run
claw
```

You'll see `ollama · llama3.2` in the header confirming it's running locally.

---

## Quick Start — Anthropic Claude

```bash
export ANTHROPIC_API_KEY=sk-ant-...
claw
```

Tarra Claw auto-detects which backend to use:
- **No API key set** → uses Ollama (free, local)
- **`ANTHROPIC_API_KEY` set** → uses Anthropic Claude

You can also force a provider:
```bash
TARRA_PROVIDER=ollama claw         # always use Ollama
TARRA_PROVIDER=anthropic claw      # always use Anthropic
```

---

## Usage

```bash
claw                          # open interactive chat (default)
claw run "explain this repo"  # one-shot prompt, exits when done
claw models                   # list available models
claw --help                   # all commands
```

### Inside the chat

Just type and press **Enter**. The AI can read your files, run commands, search code — it has full tool access.

| Key | Action |
|---|---|
| `Enter` | Send message |
| `↑` / `↓` | Browse previous messages |
| `Ctrl+A` | Jump to start of input |
| `Ctrl+E` | Jump to end of input |
| `Ctrl+U` | Clear input |
| `Ctrl+C` | Cancel current operation / exit |

### Slash commands (type inside chat)

| Command | What it does |
|---|---|
| `/clear` | Clear conversation history |
| `/cost` | Show token usage and estimated cost |
| `/tools` | List all available tools |
| `/model <name>` | Switch model mid-session |
| `/help` | Show all slash commands |
| `/exit` | Quit |

---

## Configuration

Optional. Create `~/.config/tarra-claw/config.yaml`:

```yaml
# Force a provider (auto-detected if not set)
provider: ollama              # or "anthropic"

# Ollama
ollama_host: http://localhost:11434
ollama_model: qwen2.5-coder   # change this to any model you've pulled

# Anthropic (only needed if not using env var)
# api_key: sk-ant-...
# model: claude-opus-4-6

max_tokens: 8096
auto_approve: false           # set true to skip tool permission prompts
```

---

## Built-in Tools

The AI can use these tools on your behalf (it will ask permission first):

| Tool | What it does |
|---|---|
| `Bash` | Run shell commands |
| `Read` | Read files with line numbers |
| `Write` | Create or overwrite files |
| `Edit` | Precise string replacement in files |
| `Glob` | Find files by pattern (`**/*.go`) |
| `Grep` | Search inside files (ripgrep-powered) |
| `LS` | List directory contents |
| `WebFetch` | Fetch and read any URL |
| `AskUserQuestion` | Ask you for input mid-task |
| `TodoWrite` | Manage task lists |
| `TaskCreate/List/Get/Stop` | Run and manage background tasks |
| `NotebookEdit` | Edit Jupyter notebook cells |

---

## Best Models for Ollama

```bash
ollama pull llama3.2              # 2GB — lightest, fast, good general use
ollama pull qwen2.5-coder:7b      # 4GB — best for coding tasks
ollama pull mistral               # 4GB — fast, balanced
ollama pull deepseek-coder-v2     # 8GB — most capable for code
```

Pick based on your RAM. 8GB RAM → use llama3.2. 16GB+ → qwen2.5-coder.

---

## Why Go?

| | Claude Code (TypeScript) | Tarra Claw (Go) |
|---|---|---|
| Startup time | ~2 seconds | < 50ms |
| Requires API key | Yes (always) | No — works free with Ollama |
| Runs offline | No | Yes |
| Binary size | 200MB+ (includes Node.js) | ~16MB |
| Install | `npm install -g` | Single file download |

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
    ├── state/            Reactive app state
    ├── tasks/            Background task registry
    └── compact/          Context compaction
```

---

## License

MIT — free to use, modify, and distribute.
