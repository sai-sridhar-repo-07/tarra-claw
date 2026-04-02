package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/engine"
)

// --- Styles ---

var (
	stylePrompt    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	styleUser      = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	styleAssistant = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	styleTool      = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true)
	styleError     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	styleStatus    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleDim       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleBorder    = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)
)

// --- Messages ---

type engineEventMsg engine.Event
type engineDoneMsg struct{}
type errMsg struct{ err error }

// --- Model ---

// Model is the Bubble Tea model for the interactive TUI.
type Model struct {
	eng        *engine.Engine
	cfg        *config.Config
	ctx        context.Context
	cancel     context.CancelFunc
	input      string
	cursor     int
	output     []outputLine
	status     string
	width      int
	height     int
	ready      bool
	processing bool
}

type outputLine struct {
	kind lineKind
	text string
}

type lineKind int

const (
	lineUser lineKind = iota
	lineAssistant
	lineTool
	lineToolResult
	lineError
	lineSystem
)

// New creates a new TUI Model.
func New(eng *engine.Engine, cfg *config.Config) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Model{
		eng:    eng,
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		status: "ready",
	}
	eng.OnEvent(m.handleEngineEvent)
	return m
}

// Run starts the Bubble Tea program.
func (m *Model) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// handleEngineEvent is called by the engine goroutine; sends msgs to tea.
// We store the program reference to send messages cross-goroutine.
var globalProgram *tea.Program

func (m *Model) handleEngineEvent(ev engine.Event) {
	if globalProgram == nil {
		return
	}
	globalProgram.Send(engineEventMsg(ev))
}

// --- Bubble Tea interface ---

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tea.KeyMsg:
		return m.handleKey(msg)

	case engineEventMsg:
		return m.handleEngineEvent2(engine.Event(msg))

	case engineDoneMsg:
		m.processing = false
		m.status = "ready"

	case errMsg:
		m.processing = false
		m.status = "error: " + msg.err.Error()
		m.addLine(lineError, msg.err.Error())
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		if m.processing {
			m.cancel()
			ctx, cancel := context.WithCancel(context.Background())
			m.ctx = ctx
			m.cancel = cancel
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

	case tea.KeyBackspace, tea.KeyDelete:
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
		m.input = m.input[m.cursor:]
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
	m.input = ""
	m.cursor = 0

	// Built-in slash commands
	switch prompt {
	case "/clear":
		m.eng.ClearHistory()
		m.output = nil
		m.addLine(lineSystem, "Conversation cleared.")
		return m, nil
	case "/quit", "/exit":
		return m, tea.Quit
	case "/help":
		m.addLine(lineSystem, helpText(m.eng))
		return m, nil
	}

	m.addLine(lineUser, prompt)
	m.processing = true
	m.status = "thinking..."

	return m, m.runPrompt(prompt)
}

func (m *Model) runPrompt(prompt string) tea.Cmd {
	return func() tea.Msg {
		err := m.eng.Send(m.ctx, prompt)
		if err != nil {
			return errMsg{err}
		}
		return engineDoneMsg{}
	}
}

func (m *Model) handleEngineEvent2(ev engine.Event) (tea.Model, tea.Cmd) {
	switch ev.Type {
	case engine.EventAssistantText:
		m.appendToLastAssistant(ev.Text)
	case engine.EventToolStart:
		m.addLine(lineTool, fmt.Sprintf("  %s %v", ev.Tool, formatInput(ev.Input)))
		m.status = fmt.Sprintf("running %s...", ev.Tool)
	case engine.EventToolResult:
		if ev.IsError {
			m.addLine(lineError, "  "+truncate(ev.Result, 200))
		} else {
			m.addLine(lineToolResult, "  "+truncate(ev.Result, 200))
		}
	case engine.EventError:
		m.addLine(lineError, ev.Text)
	case engine.EventDone:
		if ev.Usage != nil {
			m.status = fmt.Sprintf("ready  [in:%d out:%d]", ev.Usage.InputTokens, ev.Usage.OutputTokens)
		}
	}
	return m, nil
}

func (m *Model) appendToLastAssistant(text string) {
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

// View renders the TUI.
func (m *Model) View() string {
	if !m.ready {
		return "initializing..."
	}

	var sb strings.Builder

	// Header
	header := stylePrompt.Render("  Tarra Claw") + styleDim.Render(fmt.Sprintf("  %s", m.cfg.Model))
	sb.WriteString(header + "\n")
	sb.WriteString(strings.Repeat("─", m.width) + "\n")

	// Output area
	outputHeight := m.height - 5
	lines := m.renderOutput()
	if len(lines) > outputHeight {
		lines = lines[len(lines)-outputHeight:]
	}
	for _, l := range lines {
		sb.WriteString(l + "\n")
	}

	// Fill remaining space
	rendered := len(lines)
	for i := rendered; i < outputHeight; i++ {
		sb.WriteString("\n")
	}

	// Status bar
	sb.WriteString(strings.Repeat("─", m.width) + "\n")

	// Input line
	inputDisplay := m.input[:m.cursor] + "█" + m.input[m.cursor:]
	if m.processing {
		inputDisplay = styleStatus.Render("processing... (Ctrl+C to cancel)")
	}
	sb.WriteString(stylePrompt.Render(" > ") + inputDisplay + "\n")
	sb.WriteString(styleStatus.Render(fmt.Sprintf("  %s", m.status)))

	return sb.String()
}

func (m *Model) renderOutput() []string {
	var lines []string
	for _, ol := range m.output {
		switch ol.kind {
		case lineUser:
			lines = append(lines, styleUser.Render(" > "+ol.text))
		case lineAssistant:
			for _, l := range strings.Split(ol.text, "\n") {
				lines = append(lines, styleAssistant.Render("   "+l))
			}
		case lineTool:
			lines = append(lines, styleTool.Render(ol.text))
		case lineToolResult:
			lines = append(lines, styleDim.Render(ol.text))
		case lineError:
			lines = append(lines, styleError.Render("   ✗ "+ol.text))
		case lineSystem:
			lines = append(lines, styleDim.Render("   "+ol.text))
		}
	}
	return lines
}

func helpText(eng *engine.Engine) string {
	tools := eng.Registry().All()
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name()
	}
	return fmt.Sprintf(
		"Commands: /clear /help /quit\nTools: %s\nCtrl+C: cancel / exit",
		strings.Join(names, ", "),
	)
}

func formatInput(input map[string]any) string {
	if len(input) == 0 {
		return ""
	}
	var parts []string
	for k, v := range input {
		parts = append(parts, fmt.Sprintf("%s=%v", k, truncate(fmt.Sprint(v), 40)))
	}
	return strings.Join(parts, " ")
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", "↵")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
