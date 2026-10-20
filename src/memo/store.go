package memo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Attachment struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"` // link | image | file
	Label       string `json:"label"`
	URL         string `json:"url,omitempty"`
	SourcePath  string `json:"sourcePath,omitempty"`
	ManagedPath string `json:"managedPath,omitempty"`
	Mime        string `json:"mime,omitempty"`
	Size        int64  `json:"size,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
	CleanedAt   int64  `json:"cleanedAt,omitempty"`
}

type Item struct {
	ID             string       `json:"id"`
	Title          string       `json:"title"`
	Body           string       `json:"body,omitempty"`
	Date           string       `json:"date"`   // YYYY-MM-DD
	Status         string       `json:"status"` // todo | done
	Priority       string       `json:"priority,omitempty"`
	Tags           []string     `json:"tags,omitempty"`
	ReminderAt     int64        `json:"reminderAt,omitempty"`
	ReminderMode   string       `json:"reminderMode,omitempty"` // banner | sound | both
	ReminderSentAt int64        `json:"reminderSentAt,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty"`
	CreatedAt      int64        `json:"createdAt"`
	UpdatedAt      int64        `json:"updatedAt"`
}

type Collocation struct {
	Pattern string `json:"pattern"`
	Meaning string `json:"meaning,omitempty"`
	Example string `json:"example,omitempty"`
}

type VocabEntry struct {
	ID             string        `json:"id"`
	Word           string        `json:"word"`
	Chinese        string        `json:"chinese"`
	Phonetic       string        `json:"phonetic,omitempty"`
	PartOfSpeech   string        `json:"partOfSpeech,omitempty"`
	Collocations   []Collocation `json:"collocations,omitempty"`
	Example        string        `json:"example,omitempty"`
	AudioURL       string        `json:"audioUrl,omitempty"`
	Tags           []string      `json:"tags,omitempty"`
	ReviewLevel    int           `json:"reviewLevel,omitempty"`
	NextReviewAt   int64         `json:"nextReviewAt,omitempty"`
	LastReviewedAt int64         `json:"lastReviewedAt,omitempty"`
	CreatedAt      int64         `json:"createdAt"`
	UpdatedAt      int64         `json:"updatedAt"`
}

type Data struct {
	Memos []Item       `json:"memos"`
	Vocab []VocabEntry `json:"vocab"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	data Data
}

func New(path string) (*Store, error) {
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
	if err := json.Unmarshal(data, &s.data); err != nil {
		s.data = Data{}
	}
	return nil
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *Store) AllMemos() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Item, len(s.data.Memos))
	copy(out, s.data.Memos)
	return out
}

func (s *Store) SaveMemo(item Item) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Memos {
		if s.data.Memos[i].ID == item.ID {
			s.data.Memos[i] = item
			return item, s.saveLocked()
		}
	}
	s.data.Memos = append(s.data.Memos, item)
	return item, s.saveLocked()
}

func (s *Store) DeleteMemo(id string) (Item, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, item := range s.data.Memos {
		if item.ID == id {
			s.data.Memos = append(s.data.Memos[:i], s.data.Memos[i+1:]...)
			return item, true, s.saveLocked()
		}
	}
	return Item{}, false, nil
}

func (s *Store) Memo(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.data.Memos {
		if item.ID == id {
			return item, true
		}
	}
	return Item{}, false
}

func (s *Store) AllVocab() []VocabEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]VocabEntry, len(s.data.Vocab))
	copy(out, s.data.Vocab)
	return out
}

func (s *Store) SaveVocab(entry VocabEntry) (VocabEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Vocab {
		if s.data.Vocab[i].ID == entry.ID {
			s.data.Vocab[i] = entry
			return entry, s.saveLocked()
		}
	}
	s.data.Vocab = append(s.data.Vocab, entry)
	return entry, s.saveLocked()
}

func (s *Store) DeleteVocab(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Vocab {
		if s.data.Vocab[i].ID == id {
			s.data.Vocab = append(s.data.Vocab[:i], s.data.Vocab[i+1:]...)
			return s.saveLocked()
		}
	}
	return nil
}
