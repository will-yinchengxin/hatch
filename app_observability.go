package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ClaudeCodeStatus struct {
	ClaudeInstalled bool     `json:"claudeInstalled"`
	ClaudeVersion   string   `json:"claudeVersion"`
	VPNConnected    bool     `json:"vpnConnected"`
	VPNInterfaces   []string `json:"vpnInterfaces"`
	HasUsageData    bool     `json:"hasUsageData"`
	FiveHourPct     float64  `json:"fiveHourPct"`
	FiveHourReset   string   `json:"fiveHourReset"`
	SevenDayPct     float64  `json:"sevenDayPct"`
	SevenDayReset   string   `json:"sevenDayReset"`
	PlanType        string   `json:"planType"`
}

type CodexUsageStatus struct {
	CodexInstalled bool             `json:"codexInstalled"`
	CodexVersion   string           `json:"codexVersion"`
	VPNConnected   bool             `json:"vpnConnected"`
	VPNInterfaces  []string         `json:"vpnInterfaces"`
	HasUsageData   bool             `json:"hasUsageData"`
	PlanType       string           `json:"planType"`
	LimitReached   string           `json:"limitReached"`
	SnapshotAt     string           `json:"snapshotAt"`
	Primary        CodexUsageWindow `json:"primary"`
	Secondary      CodexUsageWindow `json:"secondary"`
}

type CodexUsageWindow struct {
	Label         string  `json:"label"`
	UsedPercent   float64 `json:"usedPercent"`
	LeftPercent   float64 `json:"leftPercent"`
	WindowMinutes int     `json:"windowMinutes"`
	ResetsAt      int64   `json:"resetsAt"`
	ResetIn       string  `json:"resetIn"`
	Available     bool    `json:"available"`
}

// GetClaudeCodeStatus checks whether claude CLI is installed, whether a VPN is
// active, and reads cached rate-limit usage from the statusline cache file.
func (a *App) GetClaudeCodeStatus() ClaudeCodeStatus {
	st := ClaudeCodeStatus{}

	claudePath := findCLI("claude")
	if claudePath != "" {
		st.ClaudeInstalled = true
		out, _ := exec.Command(claudePath, "--version").Output()
		st.ClaudeVersion = strings.TrimSpace(string(out))
	}

	st.VPNConnected, st.VPNInterfaces = detectVPNInterfaces()

	home, _ := os.UserHomeDir()
	cachePath := home + "/.cache/will-statusline/last.json"
	data, err := os.ReadFile(cachePath)
	if err == nil {
		var cache struct {
			PlanType   string `json:"plan_type"`
			RateLimits struct {
				FiveHour struct {
					UsedPercentage float64 `json:"used_percentage"`
					ResetsAt       int64   `json:"resets_at"`
				} `json:"five_hour"`
				SevenDay struct {
					UsedPercentage float64 `json:"used_percentage"`
					ResetsAt       int64   `json:"resets_at"`
				} `json:"seven_day"`
			} `json:"rate_limits"`
		}
		if json.Unmarshal(data, &cache) == nil {
			st.HasUsageData = true
			st.PlanType = cache.PlanType
			st.FiveHourPct = cache.RateLimits.FiveHour.UsedPercentage
			st.FiveHourReset = formatResetCountdown(cache.RateLimits.FiveHour.ResetsAt)
			st.SevenDayPct = cache.RateLimits.SevenDay.UsedPercentage
			st.SevenDayReset = formatResetCountdown(cache.RateLimits.SevenDay.ResetsAt)
		}
	}
	return st
}

// GetCodexUsageStatus checks whether codex CLI is installed and reads the latest local snapshot.
func (a *App) GetCodexUsageStatus() CodexUsageStatus {
	st := CodexUsageStatus{}

	codexPath := findCodexCLI()
	if codexPath != "" {
		st.CodexInstalled = true
		out, _ := exec.Command(codexPath, "--version").Output()
		st.CodexVersion = strings.TrimSpace(string(out))
	}

	st.VPNConnected, st.VPNInterfaces = detectVPNInterfaces()

	home, err := os.UserHomeDir()
	if err != nil {
		return st
	}
	snapshot, ok := loadLatestCodexRateLimitSnapshot(filepath.Join(home, ".codex", "sessions"))
	if !ok {
		return st
	}

	st.HasUsageData = true
	st.PlanType = snapshot.PlanType
	st.LimitReached = snapshot.LimitReached
	st.SnapshotAt = snapshot.Timestamp
	st.Primary = codexWindowFromSnapshot("5h", snapshot.Primary)
	st.Secondary = codexWindowFromSnapshot("7d", snapshot.Secondary)
	return st
}

func findCLI(name string) string {
	if path, err := exec.LookPath(name); err == nil && path != "" {
		return path
	}
	out, err := exec.Command("/bin/zsh", "-l", "-c", "command -v "+name).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func findCodexCLI() string { return findCLI("codex") }

func detectVPNInterfaces() (bool, []string) {
	ifOut, _ := exec.Command("ifconfig").Output()
	var currentIface string
	var interfaces []string
	for _, line := range strings.Split(string(ifOut), "\n") {
		if strings.HasPrefix(line, "utun") {
			currentIface = strings.SplitN(line, ":", 2)[0]
		} else if currentIface != "" && strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				interfaces = append(interfaces, currentIface+": "+fields[1])
			}
			currentIface = ""
		} else if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
			currentIface = ""
		}
	}
	return len(interfaces) > 0, interfaces
}

type codexRateLimitSnapshot struct {
	Timestamp    string
	PlanType     string
	LimitReached string
	Primary      codexRateLimitWindow
	Secondary    codexRateLimitWindow
}

type codexRateLimitWindow struct {
	UsedPercent   float64 `json:"used_percent"`
	WindowMinutes int     `json:"window_minutes"`
	ResetsAt      int64   `json:"resets_at"`
	Available     bool
}

func loadLatestCodexRateLimitSnapshot(sessionsDir string) (codexRateLimitSnapshot, bool) {
	var latest codexRateLimitSnapshot
	var found bool

	_ = filepath.WalkDir(sessionsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasPrefix(d.Name(), "rollout-") || !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			snapshot, ok := parseCodexRateLimitLine(line)
			if !ok {
				continue
			}
			if !found || snapshot.Timestamp > latest.Timestamp {
				latest = snapshot
				found = true
			}
		}
		return nil
	})

	return latest, found
}

func parseCodexRateLimitLine(line string) (codexRateLimitSnapshot, bool) {
	type rawWindow struct {
		UsedPercent   *float64 `json:"used_percent"`
		WindowMinutes *int     `json:"window_minutes"`
		ResetsAt      *int64   `json:"resets_at"`
	}
	var entry struct {
		Timestamp string `json:"timestamp"`
		Payload   struct {
			Type      string `json:"type"`
			RateLimit struct {
				PlanType     string    `json:"plan_type"`
				LimitReached string    `json:"rate_limit_reached_type"`
				Primary      rawWindow `json:"primary"`
				Secondary    rawWindow `json:"secondary"`
			} `json:"rate_limits"`
		} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return codexRateLimitSnapshot{}, false
	}
	if entry.Payload.Type != "token_count" {
		return codexRateLimitSnapshot{}, false
	}
	fromRaw := func(raw rawWindow) codexRateLimitWindow {
		window := codexRateLimitWindow{Available: raw.UsedPercent != nil || raw.WindowMinutes != nil || raw.ResetsAt != nil}
		if raw.UsedPercent != nil {
			window.UsedPercent = *raw.UsedPercent
		}
		if raw.WindowMinutes != nil {
			window.WindowMinutes = *raw.WindowMinutes
		}
		if raw.ResetsAt != nil {
			window.ResetsAt = *raw.ResetsAt
		}
		return window
	}
	primary := fromRaw(entry.Payload.RateLimit.Primary)
	secondary := fromRaw(entry.Payload.RateLimit.Secondary)
	if !primary.Available && !secondary.Available {
		return codexRateLimitSnapshot{}, false
	}
	return codexRateLimitSnapshot{
		Timestamp:    entry.Timestamp,
		PlanType:     entry.Payload.RateLimit.PlanType,
		LimitReached: entry.Payload.RateLimit.LimitReached,
		Primary:      primary,
		Secondary:    secondary,
	}, true
}

func codexWindowFromSnapshot(label string, window codexRateLimitWindow) CodexUsageWindow {
	if !window.Available {
		return CodexUsageWindow{Label: label}
	}
	used := window.UsedPercent
	if window.ResetsAt > 0 && window.ResetsAt <= time.Now().Unix() {
		used = 0
	}
	return CodexUsageWindow{
		Label:         label,
		UsedPercent:   used,
		LeftPercent:   maxFloat64(0, 100-used),
		WindowMinutes: window.WindowMinutes,
		ResetsAt:      window.ResetsAt,
		ResetIn:       formatResetCountdown(window.ResetsAt),
		Available:     true,
	}
}

func formatResetCountdown(ts int64) string {
	if ts == 0 {
		return ""
	}
	diff := ts - time.Now().Unix()
	if diff <= 0 {
		return "已重置"
	}
	mins := diff / 60
	hours := diff / 3600
	days := diff / 86400
	switch {
	case days >= 1:
		return fmt.Sprintf("%dd%dh", days, hours%24)
	case hours >= 1:
		return fmt.Sprintf("%dh%dm", hours, mins%60)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ─── AI trace file persistence ───────────────────────────────────────────────

const defaultTraceRetainDays = 3

type AITraceFileInfo struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Bytes int64  `json:"bytes"`
}

type AITraceStorageInfo struct {
	Dir         string            `json:"dir"`
	RetainDays  int               `json:"retainDays"`
	Files       []AITraceFileInfo `json:"files"`
	TotalBytes  int64             `json:"totalBytes"`
	TotalTraces int               `json:"totalTraces"`
}

func (a *App) traceDir() (string, error) {
	a.mu.RLock()
	custom := a.cfg.TraceLogDir
	a.mu.RUnlock()
	if custom != "" {
		return custom, os.MkdirAll(custom, 0o755)
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfgDir, "Hatch", "traces")
	return dir, os.MkdirAll(dir, 0o755)
}

func defaultTraceDir() string {
	cfgDir, _ := os.UserConfigDir()
	return filepath.Join(cfgDir, "Hatch", "traces")
}

func traceFilePath(dir, date string) string {
	return filepath.Join(dir, date+".jsonl")
}

func (a *App) traceRetainDays() int {
	a.mu.RLock()
	d := a.cfg.AI.TraceRetainDays
	a.mu.RUnlock()
	if d <= 0 {
		return defaultTraceRetainDays
	}
	return d
}

func appendTraceFile(dir string, t AICallTrace) {
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

func pruneTraceFiles(dir string, retainDays int) {
	cutoff := time.Now().AddDate(0, 0, -retainDays)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		date := strings.TrimSuffix(e.Name(), ".jsonl")
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

func readTracesFromFiles(dir string, retainDays int) []AICallTrace {
	var all []AICallTrace
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
			var t AICallTrace
			if json.Unmarshal([]byte(line), &t) == nil {
				all = append(all, t)
			}
		}
	}
	return all
}

// GetAITraces returns all persisted traces within the retain window (newest first).
func (a *App) GetAITraces() []AICallTrace {
	dir, err := a.traceDir()
	if err != nil {
		a.tracesMu.Lock()
		out := make([]AICallTrace, len(a.traces))
		copy(out, a.traces)
		a.tracesMu.Unlock()
		return out
	}
	return readTracesFromFiles(dir, a.traceRetainDays())
}

func (a *App) ClearAITraces() {
	a.tracesMu.Lock()
	a.traces = nil
	a.tracesMu.Unlock()
	dir, err := a.traceDir()
	if err != nil {
		return
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".jsonl" {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

func (a *App) GetAITraceStorageInfo() AITraceStorageInfo {
	retainDays := a.traceRetainDays()
	dir, err := a.traceDir()
	info := AITraceStorageInfo{Dir: dir, RetainDays: retainDays}
	if err != nil {
		return info
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		date := strings.TrimSuffix(e.Name(), ".jsonl")
		data, _ := os.ReadFile(filepath.Join(dir, e.Name()))
		count := 0
		for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if strings.TrimSpace(l) != "" {
				count++
			}
		}
		info.Files = append(info.Files, AITraceFileInfo{
			Date:  date,
			Count: count,
			Bytes: fi.Size(),
		})
		info.TotalBytes += fi.Size()
		info.TotalTraces += count
	}
	return info
}

// sendMacNotification triggers a macOS notification banner via osascript.
// Best-effort: errors are swallowed (user may have notifications disabled).
func sendMacNotification(title, body string) {
	// Escape double quotes for AppleScript
	esc := func(s string) string {
		return strings.ReplaceAll(s, `"`, `\"`)
	}
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, esc(body), esc(title))
	_ = exec.Command("osascript", "-e", script).Start()
}
