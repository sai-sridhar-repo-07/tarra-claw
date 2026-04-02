package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

// Entry is a saved conversation session.
type Entry struct {
	ID        string                   `json:"id"`
	Model     string                   `json:"model"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
	Messages  []anthropic.MessageParam `json:"messages"`
	Summary   string                   `json:"summary,omitempty"`
}

// Store persists conversation history to disk.
type Store struct {
	dir string
}

// New creates a Store backed by dir.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create history dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Save persists an entry.
func (s *Store) Save(e *Entry) error {
	e.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, e.ID+".json")
	return os.WriteFile(path, data, 0644)
}

// Load retrieves an entry by ID.
func (s *Store) Load(id string) (*Entry, error) {
	path := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("session %s not found: %w", id, err)
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// List returns all saved sessions, newest first.
func (s *Store) List() ([]*Entry, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var out []*Entry
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		entry, err := s.Load(id)
		if err != nil {
			continue
		}
		out = append(out, entry)
	}
	return out, nil
}

// Delete removes an entry.
func (s *Store) Delete(id string) error {
	return os.Remove(filepath.Join(s.dir, id+".json"))
}

// NewEntry creates a new history entry with a random ID.
func NewEntry(model string) *Entry {
	return &Entry{
		ID:        newID(),
		Model:     model,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func newID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
