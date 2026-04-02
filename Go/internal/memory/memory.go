package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MemoryFile represents a loaded memory file.
type MemoryFile struct {
	Path    string
	Content string
	Type    MemType
	ModTime time.Time
}

// MemType categorises memory files.
type MemType string

const (
	TypeUser      MemType = "user"
	TypeProject   MemType = "project"
	TypeFeedback  MemType = "feedback"
	TypeReference MemType = "reference"
)

// Loader finds and loads memory files relevant to the current session.
type Loader struct {
	homeDir    string
	projectDir string
	memDir     string // ~/.claude/projects/.../memory/
}

// New creates a memory Loader.
func New(projectDir string) *Loader {
	home, _ := os.UserHomeDir()
	return &Loader{
		homeDir:    home,
		projectDir: projectDir,
		memDir:     memoryDir(home, projectDir),
	}
}

// LoadAll returns all memory files: global CLAUDE.md + project CLAUDE.md + auto-memory files.
func (l *Loader) LoadAll() ([]*MemoryFile, error) {
	var files []*MemoryFile

	// 1. Global CLAUDE.md
	if f := l.loadFile(filepath.Join(l.homeDir, ".claude", "CLAUDE.md"), TypeUser); f != nil {
		files = append(files, f)
	}

	// 2. Project CLAUDE.md (walk up from projectDir)
	for dir := l.projectDir; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
		p := filepath.Join(dir, "CLAUDE.md")
		if f := l.loadFile(p, TypeProject); f != nil {
			files = append(files, f)
			break
		}
	}

	// 3. Auto-memory files from memdir
	if entries, err := os.ReadDir(l.memDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
				continue
			}
			p := filepath.Join(l.memDir, e.Name())
			if f := l.loadFile(p, classifyMemFile(e.Name())); f != nil {
				files = append(files, f)
			}
		}
	}

	return files, nil
}

// BuildPrompt assembles all memory into a single system prompt section.
func (l *Loader) BuildPrompt() string {
	files, err := l.LoadAll()
	if err != nil || len(files) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("<memory>\n")
	for _, f := range files {
		sb.WriteString(fmt.Sprintf("<!-- %s -->\n", f.Path))
		sb.WriteString(f.Content)
		sb.WriteString("\n")
	}
	sb.WriteString("</memory>")
	return sb.String()
}

// WriteAutoMemory saves a memory note to the auto-memory directory.
func (l *Loader) WriteAutoMemory(name, content string, t MemType) error {
	if err := os.MkdirAll(l.memDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(l.memDir, name)
	return os.WriteFile(path, []byte(content), 0644)
}

// MEMORY.md index path.
func (l *Loader) MemoryIndexPath() string {
	return filepath.Join(l.memDir, "MEMORY.md")
}

func (l *Loader) loadFile(path string, t MemType) *MemoryFile {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return &MemoryFile{
		Path:    path,
		Content: string(data),
		Type:    t,
		ModTime: info.ModTime(),
	}
}

func memoryDir(home, projectDir string) string {
	// Mirror Claude Code's memdir path: ~/.claude/projects/<encoded-path>/memory/
	encoded := strings.ReplaceAll(projectDir, "/", "-")
	return filepath.Join(home, ".claude", "projects", encoded, "memory")
}

func classifyMemFile(name string) MemType {
	lower := strings.ToLower(name)
	switch {
	case strings.HasPrefix(lower, "user"):
		return TypeUser
	case strings.HasPrefix(lower, "feedback"):
		return TypeFeedback
	case strings.HasPrefix(lower, "reference"):
		return TypeReference
	default:
		return TypeProject
	}
}
