package permissions

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Mode controls how permissions are enforced.
type Mode string

const (
	ModeDefault Mode = "default"  // prompt user for risky ops
	ModeBypass  Mode = "bypass"   // allow everything (--dangerously-skip-permissions)
	ModeAuto    Mode = "auto"     // auto-approve based on rules
)

// Decision is the outcome of a permission check.
type Decision string

const (
	Allow  Decision = "allow"
	Deny   Decision = "deny"
	Ask    Decision = "ask"
)

// Rule defines a permission rule for a tool.
type Rule struct {
	Tool      string   // tool name, "*" = all
	Pattern   string   // glob pattern for path/command, "" = all
	Decision  Decision
}

// Checker evaluates tool permission rules.
type Checker struct {
	mode       Mode
	rules      []Rule
	denialCount map[string]int // track repeated denials per tool
}

// New creates a Checker with default mode.
func New(mode Mode) *Checker {
	return &Checker{
		mode:        mode,
		denialCount: make(map[string]int),
	}
}

// AddRule appends a rule.
func (c *Checker) AddRule(r Rule) {
	c.rules = append(c.rules, r)
}

// Check evaluates whether a tool call is permitted.
// subject is the key value to match (file path, command, URL, etc.)
func (c *Checker) Check(toolName, subject string) Decision {
	if c.mode == ModeBypass {
		return Allow
	}

	// Check explicit rules first (last-match wins)
	result := Ask
	for _, r := range c.rules {
		if r.Tool != "*" && r.Tool != toolName {
			continue
		}
		if r.Pattern == "" || matchPattern(r.Pattern, subject) {
			result = r.Decision
		}
	}

	// Auto-deny after 3 repeated denials for same tool
	if result == Ask && c.mode == ModeAuto {
		if c.denialCount[toolName] >= 3 {
			return Deny
		}
	}

	return result
}

// RecordDenial tracks that a tool was denied (for auto-deny threshold).
func (c *Checker) RecordDenial(toolName string) {
	c.denialCount[toolName]++
}

// RecordAllow resets the denial count for a tool.
func (c *Checker) RecordAllow(toolName string) {
	c.denialCount[toolName] = 0
}

// Describe returns a human-readable reason for a decision.
func (c *Checker) Describe(toolName, subject string) string {
	d := c.Check(toolName, subject)
	switch d {
	case Allow:
		return fmt.Sprintf("%s: auto-approved", toolName)
	case Deny:
		return fmt.Sprintf("%s: denied by rule", toolName)
	default:
		return fmt.Sprintf("%s: requires approval", toolName)
	}
}

// IsDangerousCommand returns true if the bash command looks destructive.
func IsDangerousCommand(cmd string) bool {
	dangerous := []string{
		"rm -rf", "rm -f", "dd if=", "mkfs", "> /dev/",
		"chmod 777", "sudo rm", "git push --force",
		"DROP TABLE", "DELETE FROM", "truncate",
	}
	lower := strings.ToLower(cmd)
	for _, d := range dangerous {
		if strings.Contains(lower, strings.ToLower(d)) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, subject string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	// Try glob match
	matched, err := filepath.Match(pattern, subject)
	if err == nil && matched {
		return true
	}
	// Substring match as fallback
	return strings.Contains(subject, pattern)
}
