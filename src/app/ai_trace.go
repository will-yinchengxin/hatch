package app

import (
	"os"

	"hatch/ai/llm"
)

// ─── AI · 调用日志 ───────────────────────────────────────────────────────────

// recordTrace keeps the trace in memory, appends it to today's JSONL file and
// emits "ai:trace" to the frontend.
func (a *App) recordTrace(t llm.Trace) {
	a.tracesMu.Lock()
	a.traces = append([]llm.Trace{t}, a.traces...)
	if len(a.traces) > maxTraces {
		a.traces = a.traces[:maxTraces]
	}
	a.tracesMu.Unlock()

	go func() {
		if dir, err := a.traceDir(); err == nil {
			llm.AppendTrace(dir, t)
			llm.PruneTraces(dir, a.traceRetainDays())
		}
	}()
	a.emit("ai:trace", t)
}

// traceDir returns (and creates) the configured or default trace directory.
func (a *App) traceDir() (string, error) {
	a.mu.RLock()
	custom := a.cfg.TraceLogDir
	a.mu.RUnlock()
	if custom != "" {
		return custom, os.MkdirAll(custom, 0o755)
	}
	if _, err := os.UserConfigDir(); err != nil {
		return "", err
	}
	dir := llm.DefaultTraceDir()
	return dir, os.MkdirAll(dir, 0o755)
}

func (a *App) traceRetainDays() int {
	a.mu.RLock()
	d := a.cfg.AI.TraceRetainDays
	a.mu.RUnlock()
	if d <= 0 {
		return llm.DefaultTraceRetainDays
	}
	return d
}

// GetAITraces returns all persisted traces within the retain window (newest first).
func (a *App) GetAITraces() []llm.Trace {
	dir, err := a.traceDir()
	if err != nil {
		a.tracesMu.Lock()
		out := make([]llm.Trace, len(a.traces))
		copy(out, a.traces)
		a.tracesMu.Unlock()
		return out
	}
	return llm.ReadTraces(dir, a.traceRetainDays())
}

// ClearAITraces drops in-memory traces and deletes every trace file.
func (a *App) ClearAITraces() {
	a.tracesMu.Lock()
	a.traces = nil
	a.tracesMu.Unlock()
	if dir, err := a.traceDir(); err == nil {
		llm.ClearTraces(dir)
	}
}

// GetAITraceStorageInfo summarizes the trace directory.
func (a *App) GetAITraceStorageInfo() llm.TraceStorageInfo {
	retainDays := a.traceRetainDays()
	dir, err := a.traceDir()
	if err != nil {
		return llm.TraceStorageInfo{Dir: dir, RetainDays: retainDays}
	}
	return llm.TraceStorage(dir, retainDays)
}
