package state

import (
	"sync"
	"time"
)

// AppState holds all mutable runtime state for a Tarra Claw session.
type AppState struct {
	mu sync.RWMutex

	// Session
	SessionID  string
	StartedAt  time.Time
	WorkingDir string
	Model      string

	// Conversation
	TurnCount    int
	IsProcessing bool
	LastError    string

	// Cost / usage
	TotalInputTokens  int64
	TotalOutputTokens int64
	TotalCostUSD      float64

	// Permission
	AutoApprove  bool
	DenialCounts map[string]int // per-tool denial counter

	// Tool state
	ActiveTools map[string]bool // tool names currently running

	// UI
	Theme  string
	Width  int
	Height int
}

// New creates a fresh AppState.
func New(workDir, model, sessionID string) *AppState {
	return &AppState{
		SessionID:    sessionID,
		StartedAt:    time.Now(),
		WorkingDir:   workDir,
		Model:        model,
		DenialCounts: make(map[string]int),
		ActiveTools:  make(map[string]bool),
		Theme:        "dark",
	}
}

// SetProcessing marks whether the engine is currently running.
func (s *AppState) SetProcessing(v bool) {
	s.mu.Lock()
	s.IsProcessing = v
	s.mu.Unlock()
}

// AddUsage records token/cost usage from one API call.
func (s *AppState) AddUsage(inputTokens, outputTokens int64, costUSD float64) {
	s.mu.Lock()
	s.TotalInputTokens += inputTokens
	s.TotalOutputTokens += outputTokens
	s.TotalCostUSD += costUSD
	s.TurnCount++
	s.mu.Unlock()
}

// SetError records the last error string.
func (s *AppState) SetError(err string) {
	s.mu.Lock()
	s.LastError = err
	s.mu.Unlock()
}

// MarkToolActive marks a tool as currently running.
func (s *AppState) MarkToolActive(name string, active bool) {
	s.mu.Lock()
	if active {
		s.ActiveTools[name] = true
	} else {
		delete(s.ActiveTools, name)
	}
	s.mu.Unlock()
}

// ActiveToolList returns names of currently running tools.
func (s *AppState) ActiveToolList() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.ActiveTools))
	for n := range s.ActiveTools {
		out = append(out, n)
	}
	return out
}

// Snapshot returns an immutable copy of key stats for display.
func (s *AppState) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Snapshot{
		TurnCount:         s.TurnCount,
		TotalInputTokens:  s.TotalInputTokens,
		TotalOutputTokens: s.TotalOutputTokens,
		TotalCostUSD:      s.TotalCostUSD,
		IsProcessing:      s.IsProcessing,
		ActiveTools:       s.ActiveToolList(),
		Uptime:            time.Since(s.StartedAt),
	}
}

// Snapshot is an immutable view of AppState for display/logging.
type Snapshot struct {
	TurnCount         int
	TotalInputTokens  int64
	TotalOutputTokens int64
	TotalCostUSD      float64
	IsProcessing      bool
	ActiveTools       []string
	Uptime            time.Duration
}
