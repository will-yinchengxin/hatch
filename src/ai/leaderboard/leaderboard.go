// Package leaderboard reads public, CC BY 4.0 model evaluation data from Arena
// and Epoch AI. No credentials, LLM or background scheduler are involved: data
// is fetched only when the UI asks, and cached on disk between launches.
package leaderboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	cacheVersion = 2
	keepEntries  = 50                 // rows kept per board; the UI shows 20 by default
	freshFor     = 12 * time.Hour     // Arena publishes roughly weekly, Epoch daily
	retryAfter   = time.Minute        // failed feeds are not retried sooner than this
	maxBody      = 16 << 20           // Epoch's archive is ~2.4 MiB and grows over time
	fetchTimeout = 90 * time.Second   // whole refresh, all feeds in parallel
)

type Entry struct {
	Rank     int     `json:"rank"`
	Model    string  `json:"model"`
	Variant  string  `json:"variant,omitempty"` // reasoning effort or snapshot, e.g. "High"
	Org      string  `json:"org"`
	Brand    string  `json:"brand"` // normalized vendor key for icon/color
	Score    float64 `json:"score"`
	Lower    float64 `json:"lower,omitempty"` // confidence interval as published by the source
	Upper    float64 `json:"upper,omitempty"`
	Votes    int     `json:"votes,omitempty"`
	Released string  `json:"released,omitempty"` // model release date when the source provides it
}

type Board struct {
	Definition
	Entries   []Entry `json:"entries"`
	Total     int     `json:"total"`     // models in the full source board
	Published string  `json:"published"` // source's own publish date, if any
	SyncedAt  int64   `json:"syncedAt"`
	Error     string  `json:"error"`
}

type Snapshot struct {
	Boards     []Board `json:"boards"`
	Refreshing bool    `json:"refreshing"`
	Stale      bool    `json:"stale"` // some feed was never synced or is older than freshFor
	SyncedAt   int64   `json:"syncedAt"`
}

type boardData struct {
	Entries   []Entry `json:"entries"`
	Total     int     `json:"total"`
	Published string  `json:"published"`
	Error     string  `json:"error,omitempty"`
}

type feedState struct {
	SyncedAt int64  `json:"syncedAt"`
	TriedAt  int64  `json:"triedAt"`
	Error    string `json:"error,omitempty"`
}

type cacheFile struct {
	Version int                  `json:"version"`
	Feeds   map[string]feedState `json:"feeds"`
	Boards  map[string]boardData `json:"boards"`
}

type Service struct {
	mu     sync.Mutex
	path   string
	client *http.Client
	now    func() time.Time
	urls   map[string]string // feed id → URL; overridable in tests
	cache  cacheFile
	flight chan struct{} // non-nil while a refresh runs
}

func New(path string) *Service {
	s := &Service{path: path, client: &http.Client{Timeout: 45 * time.Second}, now: time.Now, urls: map[string]string{}}
	for _, f := range feeds {
		s.urls[f.id] = f.url
	}
	s.cache = cacheFile{Version: cacheVersion, Feeds: map[string]feedState{}, Boards: map[string]boardData{}}
	if data, err := os.ReadFile(path); err == nil {
		var c cacheFile
		// Older cache layouts are discarded rather than migrated; the next refresh rebuilds them.
		if json.Unmarshal(data, &c) == nil && c.Version == cacheVersion && c.Feeds != nil && c.Boards != nil {
			s.cache = c
		}
	}
	return s
}

// Snapshot returns the cached boards without touching the network.
func (s *Service) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshot()
}

func (s *Service) snapshot() Snapshot {
	now := s.now().Unix()
	out := Snapshot{Boards: make([]Board, 0, len(definitions)), Refreshing: s.flight != nil}
	for _, f := range feeds {
		st := s.cache.Feeds[f.id]
		if st.SyncedAt > out.SyncedAt {
			out.SyncedAt = st.SyncedAt
		}
		if s.due(st, now) {
			out.Stale = true
		}
	}
	for _, d := range definitions {
		st, data := s.cache.Feeds[d.feed], s.cache.Boards[d.ID]
		b := Board{Definition: d, Entries: append([]Entry{}, data.Entries...), Total: data.Total, Published: data.Published, SyncedAt: st.SyncedAt}
		b.Error = st.Error
		if b.Error == "" {
			b.Error = data.Error
		}
		out.Boards = append(out.Boards, b)
	}
	return out
}

func (s *Service) due(st feedState, now int64) bool {
	if st.Error != "" && now-st.TriedAt < int64(retryAfter/time.Second) {
		return false
	}
	return st.SyncedAt == 0 || now-st.SyncedAt >= int64(freshFor/time.Second)
}

// Refresh fetches every feed that is stale (or all of them when force is set)
// in parallel. Concurrent callers share one in-flight refresh. A failing feed
// keeps its previous boards and records the error.
func (s *Service) Refresh(ctx context.Context, force bool) Snapshot {
	s.mu.Lock()
	if wait := s.flight; wait != nil {
		s.mu.Unlock()
		select {
		case <-ctx.Done():
		case <-wait:
		}
		return s.Snapshot()
	}
	now := s.now().Unix()
	var todo []feed
	for _, f := range feeds {
		st := s.cache.Feeds[f.id]
		if force || s.due(st, now) {
			st.TriedAt = now
			s.cache.Feeds[f.id] = st
			todo = append(todo, f)
		}
	}
	if len(todo) == 0 {
		defer s.mu.Unlock()
		return s.snapshot()
	}
	done := make(chan struct{})
	s.flight = done
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	type result struct {
		boards map[string]boardData
		err    error
	}
	results := make([]result, len(todo))
	var wg sync.WaitGroup
	for i, f := range todo {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i].boards, results[i].err = s.load(ctx, f.id)
		}()
	}
	wg.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, f := range todo {
		st := s.cache.Feeds[f.id]
		if err := results[i].err; err != nil {
			st.Error = err.Error()
		} else {
			st.Error, st.SyncedAt = "", s.now().Unix()
			for id, data := range results[i].boards {
				s.cache.Boards[id] = data
			}
		}
		s.cache.Feeds[f.id] = st
	}
	if err := s.save(); err != nil {
		for _, f := range todo {
			st := s.cache.Feeds[f.id]
			if st.Error == "" {
				st.Error = "缓存写入失败：" + err.Error()
				s.cache.Feeds[f.id] = st
			}
		}
	}
	s.flight = nil
	close(done)
	return s.snapshot()
}

// load downloads one feed and parses every board defined on it.
func (s *Service) load(ctx context.Context, feedID string) (map[string]boardData, error) {
	body, err := s.get(ctx, s.urls[feedID])
	if err != nil {
		return nil, err
	}
	picks := map[string]string{} // board id → category / csv
	for _, d := range definitions {
		if d.feed == feedID {
			picks[d.ID] = d.pick
		}
	}
	if feedID == "epoch" {
		return parseEpoch(body, picks)
	}
	return parseArena(body, picks)
}

func (s *Service) save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(s.cache)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".leaderboards-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.path)
}

func (s *Service) get(ctx context.Context, url string) ([]byte, error) {
	var last error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Hatch model-leaderboards")
		resp, err := s.client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, errors.New("请求超时")
			}
			last = fmt.Errorf("网络请求失败：%w", err)
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			last = fmt.Errorf("数据源返回 HTTP %d", resp.StatusCode)
			if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
				return nil, last
			}
			continue
		}
		if readErr != nil {
			last = fmt.Errorf("读取响应失败：%w", readErr)
			continue
		}
		if len(body) > maxBody {
			return nil, fmt.Errorf("响应超过 %d MiB 上限", maxBody>>20)
		}
		return body, nil
	}
	return nil, last
}
