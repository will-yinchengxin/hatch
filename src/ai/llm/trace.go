package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hatch/core/fsutil"
)

// DefaultTraceRetainDays is used when config.AI.TraceRetainDays is 0.
const DefaultTraceRetainDays = 3

// TraceFileInfo summarizes one daily trace file.
type TraceFileInfo struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Bytes int64  `json:"bytes"`
}

// TraceStorageInfo summarizes the trace directory for the data page.
type TraceStorageInfo struct {
	Dir         string          `json:"dir"`
	RetainDays  int             `json:"retainDays"`
	Files       []TraceFileInfo `json:"files"`
	TotalBytes  int64           `json:"totalBytes"`
	TotalTraces int             `json:"totalTraces"`
}

// DefaultTraceDir is ~/Library/Application Support/Hatch/traces.
func DefaultTraceDir() string {
	return fsutil.AppDataPath("traces")
}

func traceFilePath(dir, date string) string {
	return filepath.Join(dir, date+".jsonl")
}

// AppendTrace appends t to today's <dir>/<YYYY-MM-DD>.jsonl. Best-effort.
func AppendTrace(dir string, t Trace) {
	line, err := json.Marshal(t)
	if err != nil {
		return
	}
	f, err := os.OpenFile(traceFilePath(dir, time.Now().Format("2006-01-02")),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(line, '\n'))
}

// PruneTraces deletes daily trace files older than retainDays.
func PruneTraces(dir string, retainDays int) {
	cutoff := time.Now().AddDate(0, 0, -retainDays)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		t, err := time.Parse("2006-01-02", strings.TrimSuffix(e.Name(), ".jsonl"))
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// ReadTraces returns traces from the last retainDays daily files, newest first.
func ReadTraces(dir string, retainDays int) []Trace {
	var all []Trace
	for i := 0; i < retainDays; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		data, err := os.ReadFile(traceFilePath(dir, date))
		if err != nil {
			continue
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		for j := len(lines) - 1; j >= 0; j-- {
			line := strings.TrimSpace(lines[j])
			if line == "" {
				continue
			}
			var t Trace
			if json.Unmarshal([]byte(line), &t) == nil {
				all = append(all, t)
			}
		}
	}
	return all
}

// ClearTraces removes every .jsonl trace file in dir.
func ClearTraces(dir string) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".jsonl" {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// TraceStorage reports the size and count of every trace file in dir.
func TraceStorage(dir string, retainDays int) TraceStorageInfo {
	info := TraceStorageInfo{Dir: dir, RetainDays: retainDays}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		data, _ := os.ReadFile(filepath.Join(dir, e.Name()))
		count := 0
		for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if strings.TrimSpace(l) != "" {
				count++
			}
		}
		info.Files = append(info.Files, TraceFileInfo{
			Date:  strings.TrimSuffix(e.Name(), ".jsonl"),
			Count: count,
			Bytes: fi.Size(),
		})
		info.TotalBytes += fi.Size()
		info.TotalTraces += count
	}
	return info
}
