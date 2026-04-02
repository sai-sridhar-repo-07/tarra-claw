package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/commands"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/engine"
)

// ── Styles ────────────────────────────────────────────────────────────────────
var (
	stylePrompt    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	styleUser      = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	styleAssistant = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	styleTool      = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true)
	styleError     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	styleStatus    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleDim       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleBanner    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
)

// ── Tea messages ──────────────────────────────────────────────────────────────
type engineEventMsg engine.Event
type engineDoneMsg struct{ err error }

// ── Output lines ──────────────────────────────────────────────────────────────
type lineKind int

const (
	lineUser lineKind = iota
	lineAssistant
	lineTool
	lineToolResult
	lineError
	lineSystem
)

type outputLine struct {
	kind lineKind
	text string
}

// ── Model ─────────────────────────────────────────────────────────────────────
type Model struct {
	eng     *engine.Engine
	cfg     *config.Config
	cmds    *commands.Registry
	ctx     context.Context
	cancel  context.CancelFunc
	program *tea.Program

	input      string
	cursor     int
	output     []outputLine
	history    []string // command history
	histIdx    int
	status     string
	width      int
	height     int
	ready      bool
	processing bool
}

func New(eng *engine.Engine, cfg *config.Config) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Model{
		eng:    eng,
		cfg:    cfg,
		cmds:   commands.New(),
		ctx:    ctx,
		cancel: cancel,
		status: "ready",
	}
	return m
}

func (m *Model) Run() error {
	// Wire engine events back to the tea program
	eng := m.eng
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p
	eng.OnEvent(func(ev engine.Event) { p.Send(engineEventMsg(ev)) })
	_, err := p.Run()
	return err
}

// ── Bubble Tea interface ───────────────────────────────────────────────────────
func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tea.KeyMsg:
		return m.handleKey(msg)

	case engineEventMsg:
		return m.handleEngineEvent(engine.Event(msg))

	case engineDoneMsg:
		m.processing = false
		if msg.err != nil {
			m.addLine(lineError, msg.err.Error())
			m.status = "error"
		} else {
			m.status = "ready"
		}
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		if m.processing {
			m.cancel()
			ctx, cancel := context.WithCancel(context.Background())
			m.ctx = ctx
			m.cancel = cancel
			m.eng.OnEvent(func(ev engine.Event) { m.program.Send(engineEventMsg(ev)) })
			m.processing = false
			m.status = "cancelled"
			return m, nil
		}
		m.cancel()
		return m, tea.Quit

	case tea.KeyEnter:
		if m.processing || strings.TrimSpace(m.input) == "" {
			return m, nil
		}
		return m.submit()

	case tea.KeyUp:
		if len(m.history) > 0 {
			if m.histIdx < len(m.history) {
				m.histIdx++
			}
			m.input = m.history[len(m.history)-m.histIdx]
			m.cursor = len(m.input)
		}

	case tea.KeyDown:
		if m.histIdx > 0 {
			m.histIdx--
			if m.histIdx == 0 {
				m.input = ""
			} else {
				m.input = m.history[len(m.history)-m.histIdx]
			}
			m.cursor = len(m.input)
		}

	case tea.KeyBackspace:
		if m.cursor > 0 {
			m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
			m.cursor--
		}

	case tea.KeyLeft:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyRight:
		if m.cursor < len(m.input) {
			m.cursor++
		}
	case tea.KeyCtrlA:
		m.cursor = 0
	case tea.KeyCtrlE:
		m.cursor = len(m.input)
	case tea.KeyCtrlU:
		m.input = ""
		m.cursor = 0

	default:
		if msg.Type == tea.KeyRunes {
			m.input = m.input[:m.cursor] + string(msg.Runes) + m.input[m.cursor:]
			m.cursor += len(msg.Runes)
		}
	}
	return m, nil
}

func (m *Model) submit() (tea.Model, tea.Cmd) {
	prompt := strings.TrimSpace(m.input)
	m.history = append(m.history, prompt)
	m.histIdx = 0
	m.input = ""
	m.cursor = 0

	// Try slash commands first
	cmdEnv := &commands.Env{
		WorkDir: m.cfg.WorkingDir,
		ClearFn: func() {
			m.eng.ClearHistory()
			m.output = nil
		},
		GetCost:   func() string { return m.eng.CostSummary() },
		ListTools: func() []string { return m.eng.Registry().Names() },
	}

	out, matched, err := m.cmds.Execute(m.ctx, prompt, cmdEnv)
	if matched {
		if err != nil {
			m.addLine(lineError, err.Error())
		} else if out == "exit" {
			return m, tea.Quit
		} else if out != "" {
			m.addLine(lineSystem, out)
		}
		return m, nil
	}

	m.addLine(lineUser, prompt)
	m.processing = true
	m.status = "thinking…"

	return m, func() tea.Msg {
		err := m.eng.Send(m.ctx, prompt)
		return engineDoneMsg{err: err}
	}
}

func (m *Model) handleEngineEvent(ev engine.Event) (tea.Model, tea.Cmd) {
	switch ev.Type {
	case engine.EventAssistantText:
		m.appendAssistant(ev.Text)
	case engine.EventToolStart:
		label := fmt.Sprintf("⚙ %s", ev.Tool)
		if cmd, ok := ev.Input["command"].(string); ok {
			label += fmt.Sprintf("  %s", truncate(cmd, 60))
		} else if path, ok := ev.Input["file_path"].(string); ok {
			label += fmt.Sprintf("  %s", path)
		}
		m.addLine(lineTool, label)
		m.status = fmt.Sprintf("running %s…", ev.Tool)
	case engine.EventToolResult:
		if ev.IsError {
			m.addLine(lineError, "  ✗ "+truncate(ev.Result, 200))
		} else {
			m.addLine(lineToolResult, "  ✓ "+truncate(ev.Result, 200))
		}
	case engine.EventError:
		m.addLine(lineError, ev.Text)
	case engine.EventDone:
		if ev.Usage != nil {
			m.status = fmt.Sprintf("ready  ·  in:%d out:%d", ev.Usage.InputTokens, ev.Usage.OutputTokens)
		}
	}
	return m, nil
}

func (m *Model) appendAssistant(text string) {
	for i := len(m.output) - 1; i >= 0; i-- {
		if m.output[i].kind == lineAssistant {
			m.output[i].text += text
			return
		}
	}
	m.addLine(lineAssistant, text)
}

func (m *Model) addLine(kind lineKind, text string) {
	m.output = append(m.output, outputLine{kind: kind, text: text})
}

// ── View ──────────────────────────────────────────────────────────────────────
func (m *Model) View() string {
	if !m.ready {
		return "  Starting Tarra Claw…"
	}

	var sb strings.Builder

	// Header
	header := styleBanner.Render("  ⊕ Tarra Claw") +
		styleDim.Render(fmt.Sprintf("  %s  ·  %s", m.eng.ProviderInfo(), m.status))
	sb.WriteString(header + "\n")
	sb.WriteString(styleDim.Render(strings.Repeat("─", m.width)) + "\n")

	// Output area
	outputHeight := m.height - 5
	rendered := m.renderOutput()
	if len(rendered) > outputHeight {
		rendered = rendered[len(rendered)-outputHeight:]
	}
	for _, l := range rendered {
		sb.WriteString(l + "\n")
	}
	for i := len(rendered); i < outputHeight; i++ {
		sb.WriteString("\n")
	}

	// Footer
	sb.WriteString(styleDim.Render(strings.Repeat("─", m.width)) + "\n")

	// Input
	var inputView string
	if m.processing {
		inputView = styleDim.Render("  processing… (Ctrl+C to cancel)")
	} else {
		before := m.input[:m.cursor]
		cursor := "█"
		after := m.input[m.cursor:]
		inputView = stylePrompt.Render(" › ") + before + cursor + after
	}
	sb.WriteString(inputView)

	return sb.String()
}

func (m *Model) renderOutput() []string {
	var lines []string
	for _, ol := range m.output {
		switch ol.kind {
		case lineUser:
			lines = append(lines, styleUser.Render(" › ")+ol.text)
		case lineAssistant:
			for _, l := range strings.Split(ol.text, "\n") {
				lines = append(lines, styleAssistant.Render("   "+l))
			}
		case lineTool:
			lines = append(lines, styleTool.Render("   "+ol.text))
		case lineToolResult:
			lines = append(lines, styleDim.Render("   "+ol.text))
		case lineError:
			lines = append(lines, styleError.Render("   "+ol.text))
		case lineSystem:
			for _, l := range strings.Split(ol.text, "\n") {
				lines = append(lines, styleDim.Render("   "+l))
			}
		}
	}
	return lines
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", "↵")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
