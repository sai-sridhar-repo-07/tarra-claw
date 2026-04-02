package cost

import (
	"fmt"
	"sync"
	"time"
)

// Pricing per model (per million tokens, USD).
var modelPricing = map[string]struct{ input, output, cacheRead, cacheWrite float64 }{
	"claude-opus-4-6":           {15.0, 75.0, 1.50, 18.75},
	"claude-sonnet-4-6":         {3.0, 15.0, 0.30, 3.75},
	"claude-haiku-4-5-20251001": {0.80, 4.0, 0.08, 1.00},
}

// Usage holds token counts for a single request.
type Usage struct {
	InputTokens  int64
	OutputTokens int64
	CacheRead    int64
	CacheWrite   int64
	Duration     time.Duration
}

// Session accumulates usage across all requests in a session.
type Session struct {
	mu      sync.Mutex
	model   string
	total   Usage
	started time.Time
	requests int
}

// New creates a new cost session.
func New(model string) *Session {
	return &Session{model: model, started: time.Now()}
}

// Add records usage from a single API call.
func (s *Session) Add(u Usage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.total.InputTokens += u.InputTokens
	s.total.OutputTokens += u.OutputTokens
	s.total.CacheRead += u.CacheRead
	s.total.CacheWrite += u.CacheWrite
	s.total.Duration += u.Duration
	s.requests++
}

// TotalCost returns the estimated USD cost for the session.
func (s *Session) TotalCost() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calcCost(s.total)
}

func (s *Session) calcCost(u Usage) float64 {
	p, ok := modelPricing[s.model]
	if !ok {
		// Default to Sonnet pricing if unknown
		p = modelPricing["claude-sonnet-4-6"]
	}
	const M = 1_000_000.0
	return (float64(u.InputTokens)*p.input +
		float64(u.OutputTokens)*p.output +
		float64(u.CacheRead)*p.cacheRead +
		float64(u.CacheWrite)*p.cacheWrite) / M
}

// Summary returns a formatted session summary.
func (s *Session) Summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	cost := s.calcCost(s.total)
	elapsed := time.Since(s.started).Round(time.Second)
	return fmt.Sprintf(
		"tokens: %d in / %d out | cache: %d read / %d write | cost: $%.4f | time: %s | requests: %d",
		s.total.InputTokens, s.total.OutputTokens,
		s.total.CacheRead, s.total.CacheWrite,
		cost, elapsed, s.requests,
	)
}

// FormatCost returns a human-readable cost string.
func FormatCost(usd float64) string {
	if usd < 0.01 {
		return fmt.Sprintf("$%.5f", usd)
	}
	return fmt.Sprintf("$%.4f", usd)
}

// Tokens returns the total token count.
func (s *Session) Tokens() (in, out int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.total.InputTokens, s.total.OutputTokens
}

// Requests returns the number of API calls made.
func (s *Session) Requests() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.requests
}
