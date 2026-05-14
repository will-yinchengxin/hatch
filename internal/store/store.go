package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"hatch/internal/config"
)

// Store persists the scanned script list between app restarts.
type Store struct {
	mu   sync.RWMutex
	path string
	data storeData
}

type storeData struct {
	Scripts []config.ScriptMeta `json:"scripts"`
}

func defaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Hatch", "scripts.json"), nil
}

// New loads (or initialises) the store from the default path.
// Falls back to a temp path if the OS config dir is unavailable.
func New() (*Store, error) {
	path, err := defaultPath()
	if err != nil {
		path = filepath.Join(os.TempDir(), "hatch", "scripts.json")
	}
	return NewWithPath(path)
}

// NewWithPath creates a store backed by the given file path.
func NewWithPath(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	// Corrupt JSON → start fresh rather than crash.
	if err := json.Unmarshal(data, &s.data); err != nil {
		s.data = storeData{}
	}
	return nil
}

func (s *Store) save() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

// All returns a copy of all stored scripts.
func (s *Store) All() []config.ScriptMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]config.ScriptMeta, len(s.data.Scripts))
	copy(out, s.data.Scripts)
	return out
}

// Replace overwrites the entire script list and persists to disk.
func (s *Store) Replace(scripts []config.ScriptMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Scripts = scripts
	return s.save()
}

// UpdateMeta updates a single script's metadata by path and persists.
func (s *Store) UpdateMeta(meta config.ScriptMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sc := range s.data.Scripts {
		if sc.Path == meta.Path {
			s.data.Scripts[i] = meta
			return s.save()
		}
	}
	s.data.Scripts = append(s.data.Scripts, meta)
	return s.save()
}
