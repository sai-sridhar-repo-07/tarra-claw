package tasks

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Type categorises what a task does.
type Type string

const (
	TypeLocalBash      Type = "local_bash"
	TypeLocalAgent     Type = "local_agent"
	TypeRemoteAgent    Type = "remote_agent"
	TypeInProcessAgent Type = "in_process_teammate"
	TypeWorkflow       Type = "local_workflow"
)

// Status tracks a task's lifecycle.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusKilled    Status = "killed"
)

// Task holds all state for a single task.
type Task struct {
	ID          string
	Type        Type
	Subject     string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Output      []string
	Err         error
	cancel      context.CancelFunc
	mu          sync.Mutex
}

// Registry manages all active tasks.
type Registry struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{tasks: make(map[string]*Task)}
}

// Create adds a new pending task.
func (r *Registry) Create(subject, description string, t Type) *Task {
	task := &Task{
		ID:          newTaskID(),
		Type:        t,
		Subject:     subject,
		Description: description,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.mu.Lock()
	r.tasks[task.ID] = task
	r.mu.Unlock()
	return task
}

// Get retrieves a task by ID.
func (r *Registry) Get(id string) (*Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	return t, ok
}

// List returns all tasks.
func (r *Registry) List() []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		out = append(out, t)
	}
	return out
}

// Update modifies task status.
func (r *Registry) Update(id string, status Status, err error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return fmt.Errorf("task %s not found", id)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
	t.UpdatedAt = time.Now()
	if err != nil {
		t.Err = err
	}
	return nil
}

// AppendOutput adds a line to a task's output buffer.
func (r *Registry) AppendOutput(id, line string) {
	r.mu.RLock()
	t, ok := r.tasks[id]
	r.mu.RUnlock()
	if !ok {
		return
	}
	t.mu.Lock()
	t.Output = append(t.Output, line)
	t.mu.Unlock()
}

// Stop cancels a running task.
func (r *Registry) Stop(id string) error {
	r.mu.RLock()
	t, ok := r.tasks[id]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("task %s not found", id)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancel != nil {
		t.cancel()
	}
	t.Status = StatusKilled
	t.UpdatedAt = time.Now()
	return nil
}

// SetCancel attaches a cancel function to a task (called when task starts).
func (r *Registry) SetCancel(id string, cancel context.CancelFunc) {
	r.mu.RLock()
	t, ok := r.tasks[id]
	r.mu.RUnlock()
	if ok {
		t.mu.Lock()
		t.cancel = cancel
		t.Status = StatusRunning
		t.mu.Unlock()
	}
}

func newTaskID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
