package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// RunRecord is one archived script execution.
type RunRecord struct {
	ID         string    `json:"id"`       // basename-timestamp
	Script     string    `json:"script"`   // absolute script path
	Name       string    `json:"name"`     // basename for grouping
	Args       []string  `json:"args"`
	StartedAt  time.Time `json:"startedAt"`
	EndedAt    time.Time `json:"endedAt"`
	ExitCode   int       `json:"exitCode"`
	DurationMs int64     `json:"durationMs"`
	StdoutTail string    `json:"stdoutTail"` // last ~50 lines for preview
	LogPath    string    `json:"logPath"`    // full log file path
}

// historyRoot returns the base directory for run-history archives.
func historyRoot() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "hatch-runs")
	}
	return filepath.Join(cfgDir, "Hatch", "runs")
}

func scriptHistoryDir(scriptPath string) string {
	base := filepath.Base(scriptPath)
	return filepath.Join(historyRoot(), base)
}

// runRecorder collects stdout/stderr to disk while a script runs.
type runRecorder struct {
	mu      sync.Mutex
	record  RunRecord
	file    *os.File
	tail    []string // last N stdout lines for preview
	maxTail int
}

func newRunRecorder(scriptPath string, args []string) (*runRecorder, error) {
	dir := scriptHistoryDir(scriptPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%s-%d", filepath.Base(scriptPath), time.Now().UnixNano())
	logPath := filepath.Join(dir, id+".log")
	f, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	return &runRecorder{
		record: RunRecord{
			ID:        id,
			Script:    scriptPath,
			Name:      filepath.Base(scriptPath),
			Args:      args,
			StartedAt: time.Now(),
			LogPath:   logPath,
		},
		file:    f,
		maxTail: 50,
	}, nil
}

func (r *runRecorder) write(line, stream string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	prefix := "[out] "
	if stream == "stderr" {
		prefix = "[err] "
	}
	_, _ = r.file.WriteString(prefix + line + "\n")
	if stream == "stdout" {
		r.tail = append(r.tail, line)
		if len(r.tail) > r.maxTail {
			r.tail = r.tail[len(r.tail)-r.maxTail:]
		}
	}
}

func (r *runRecorder) finalize(exitCode int) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.record.EndedAt = time.Now()
	r.record.ExitCode = exitCode
	r.record.DurationMs = r.record.EndedAt.Sub(r.record.StartedAt).Milliseconds()
	r.record.StdoutTail = strings.Join(r.tail, "\n")
	_ = r.file.Close()

	metaPath := strings.TrimSuffix(r.record.LogPath, ".log") + ".json"
	if data, err := json.MarshalIndent(r.record, "", "  "); err == nil {
		_ = os.WriteFile(metaPath, data, 0o644)
	}
}

// ListRuns returns archived runs for a script, newest first.
func ListRuns(scriptPath string) ([]RunRecord, error) {
	dir := scriptHistoryDir(scriptPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []RunRecord
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var r RunRecord
		if json.Unmarshal(data, &r) != nil {
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	return out, nil
}

// GetRunLog returns the full log file contents for one run.
func GetRunLog(scriptPath, runID string) (string, error) {
	logPath := filepath.Join(scriptHistoryDir(scriptPath), runID+".log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeleteRun removes a single archived run (both .log and .json).
func DeleteRun(scriptPath, runID string) error {
	dir := scriptHistoryDir(scriptPath)
	_ = os.Remove(filepath.Join(dir, runID+".log"))
	_ = os.Remove(filepath.Join(dir, runID+".json"))
	return nil
}

// PruneHistoryOlderThan removes records older than retainDays across all scripts.
// retainDays <= 0 means "no time-based pruning". Returns number of records removed.
func PruneHistoryOlderThan(retainDays int) int {
	if retainDays <= 0 {
		return 0
	}
	cutoff := time.Now().AddDate(0, 0, -retainDays)
	root := historyRoot()
	subs, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	removed := 0
	for _, s := range subs {
		if !s.IsDir() {
			continue
		}
		full := filepath.Join(root, s.Name())
		entries, err := os.ReadDir(full)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			fi, err := e.Info()
			if err != nil {
				continue
			}
			if fi.ModTime().Before(cutoff) {
				id := strings.TrimSuffix(e.Name(), ".json")
				_ = os.Remove(filepath.Join(full, id+".json"))
				_ = os.Remove(filepath.Join(full, id+".log"))
				removed++
			}
		}
	}
	return removed
}

// HistoryRoot exposes the on-disk root for the data-dir dashboard.
func HistoryRoot() string { return historyRoot() }

// PruneAllHistory removes oldest records per script when over keepPerScript.
func PruneAllHistory(keepPerScript int) {
	root := historyRoot()
	subs, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, s := range subs {
		if !s.IsDir() {
			continue
		}
		full := filepath.Join(root, s.Name())
		entries, err := os.ReadDir(full)
		if err != nil {
			continue
		}
		var metas []os.DirEntry
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".json") {
				metas = append(metas, e)
			}
		}
		if len(metas) <= keepPerScript {
			continue
		}
		sort.Slice(metas, func(i, j int) bool {
			fi, _ := metas[i].Info()
			fj, _ := metas[j].Info()
			return fi.ModTime().Before(fj.ModTime())
		})
		for i := 0; i < len(metas)-keepPerScript; i++ {
			id := strings.TrimSuffix(metas[i].Name(), ".json")
			_ = os.Remove(filepath.Join(full, id+".json"))
			_ = os.Remove(filepath.Join(full, id+".log"))
		}
	}
}
