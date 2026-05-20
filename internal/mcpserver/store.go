package mcpserver

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Project is a named workspace that scopes skill availability and carries
// a system prompt delivered to agents via workbench-system-prompt.
type Project struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	SystemPrompt string    `json:"system_prompt"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ErrProjectNotFound is returned when a project does not exist.
var ErrProjectNotFound = errors.New("project not found")

// ProjectStore is the storage contract for projects.
type ProjectStore interface {
	Create(name, description, systemPrompt string) (Project, error)
	Get(id uuid.UUID) (Project, error)
	List() ([]Project, error)
	Delete(id uuid.UUID) (bool, error)
}

// ---- MemProjectStore (tests) ----

// MemProjectStore is an in-memory ProjectStore. Restart loses all state.
type MemProjectStore struct {
	mu       sync.Mutex
	projects map[uuid.UUID]Project
}

// NewMemProjectStore returns an empty MemProjectStore.
func NewMemProjectStore() *MemProjectStore {
	return &MemProjectStore{projects: make(map[uuid.UUID]Project)}
}

func (s *MemProjectStore) Create(name, description, systemPrompt string) (Project, error) {
	now := time.Now().UTC()
	p := Project{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.mu.Lock()
	s.projects[p.ID] = p
	s.mu.Unlock()
	return p, nil
}

func (s *MemProjectStore) Get(id uuid.UUID) (Project, error) {
	s.mu.Lock()
	p, ok := s.projects[id]
	s.mu.Unlock()
	if !ok {
		return Project{}, ErrProjectNotFound
	}
	return p, nil
}

func (s *MemProjectStore) List() ([]Project, error) {
	s.mu.Lock()
	out := make([]Project, 0, len(s.projects))
	for _, p := range s.projects {
		out = append(out, p)
	}
	s.mu.Unlock()
	return out, nil
}

func (s *MemProjectStore) Delete(id uuid.UUID) (bool, error) {
	s.mu.Lock()
	_, ok := s.projects[id]
	if ok {
		delete(s.projects, id)
	}
	s.mu.Unlock()
	return ok, nil
}

// ---- FileProjectStore (production) ----

// FileProjectStore persists projects to a JSON file so they survive process
// restarts and are visible to parallel agent processes on the same machine.
type FileProjectStore struct {
	mu   sync.Mutex
	path string
}

// NewFileProjectStore returns a FileProjectStore backed by path.
// The file and its parent directory are created if they do not exist.
func NewFileProjectStore(path string) (*FileProjectStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	s := &FileProjectStore{path: path}
	// Initialize with an empty store if the file is missing.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := s.write(make(map[uuid.UUID]Project)); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *FileProjectStore) read() (map[uuid.UUID]Project, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[uuid.UUID]Project), nil
		}
		return nil, err
	}
	var m map[uuid.UUID]Project
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *FileProjectStore) write(m map[uuid.UUID]Project) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *FileProjectStore) Create(name, description, systemPrompt string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.read()
	if err != nil {
		return Project{}, err
	}
	now := time.Now().UTC()
	p := Project{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	m[p.ID] = p
	return p, s.write(m)
}

func (s *FileProjectStore) Get(id uuid.UUID) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.read()
	if err != nil {
		return Project{}, err
	}
	p, ok := m[id]
	if !ok {
		return Project{}, ErrProjectNotFound
	}
	return p, nil
}

func (s *FileProjectStore) List() ([]Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.read()
	if err != nil {
		return nil, err
	}
	out := make([]Project, 0, len(m))
	for _, p := range m {
		out = append(out, p)
	}
	return out, nil
}

func (s *FileProjectStore) Delete(id uuid.UUID) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.read()
	if err != nil {
		return false, err
	}
	_, ok := m[id]
	if !ok {
		return false, nil
	}
	delete(m, id)
	return true, s.write(m)
}
