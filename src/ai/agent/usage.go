package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hatch/core/macos"
)

// ClaudeCodeStatus is the Claude Code card on the AI → 助手 page.
type ClaudeCodeStatus struct {
	ClaudeInstalled bool     `json:"claudeInstalled"`
	ClaudeVersion   string   `json:"claudeVersion"`
	VPNConnected    bool     `json:"vpnConnected"`
	VPNInterfaces   []string `json:"vpnInterfaces"`
	HasUsageData    bool     `json:"hasUsageData"`
	FiveHourPct     float64  `json:"fiveHourPct"`
	FiveHourReset   string   `json:"fiveHourReset"`
	FiveHourResetAt string   `json:"fiveHourResetAt"`
	SevenDayPct     float64  `json:"sevenDayPct"`
	SevenDayReset   string   `json:"sevenDayReset"`
	SevenDayResetAt string   `json:"sevenDayResetAt"`
	PlanType        string   `json:"planType"`
	UsageUpdatedAt  string   `json:"usageUpdatedAt"`
	UsageAge        string   `json:"usageAge"`
	UsageStale      bool     `json:"usageStale"`
}

// CodexUsageStatus is the Codex card on the AI → 助手 page.
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

// CodexUsageWindow is one rate-limit window (5h or 7d).
type CodexUsageWindow struct {
	Label         string  `json:"label"`
	UsedPercent   float64 `json:"usedPercent"`
	LeftPercent   float64 `json:"leftPercent"`
	WindowMinutes int     `json:"windowMinutes"`
	ResetsAt      int64   `json:"resetsAt"`
	ResetIn       string  `json:"resetIn"`
	ResetAt       string  `json:"resetAt"`
	Available     bool    `json:"available"`
}

// claudeUsageCacheMaxAge matches CACHE_MAX_AGE in the statusline script; beyond
// it the cached rate limits are treated as no longer representative.
const claudeUsageCacheMaxAge = 6 * time.Hour

// ClaudeStatus checks whether claude CLI is installed, whether a VPN is active,
// and reads cached rate-limit usage from the statusline cache file.
func ClaudeStatus() ClaudeCodeStatus {
	st := ClaudeCodeStatus{}
	if claudePath := macos.FindCLI("claude"); claudePath != "" {
		st.ClaudeInstalled = true
		st.ClaudeVersion = macos.CLIVersion(claudePath)
	}
	st.VPNConnected, st.VPNInterfaces = macos.UtunInterfaces()

	cachePath, err := ClaudeUsageCachePath()
	if err != nil {
		return st
	}
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return st
	}
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
	if json.Unmarshal(data, &cache) != nil {
		return st
	}
	st.HasUsageData = true
	st.PlanType = cache.PlanType
	st.FiveHourPct = cache.RateLimits.FiveHour.UsedPercentage
	st.FiveHourReset = formatResetCountdown(cache.RateLimits.FiveHour.ResetsAt)
	st.FiveHourResetAt = formatBeijingTime(cache.RateLimits.FiveHour.ResetsAt)
	st.SevenDayPct = cache.RateLimits.SevenDay.UsedPercentage
	st.SevenDayReset = formatResetCountdown(cache.RateLimits.SevenDay.ResetsAt)
	st.SevenDayResetAt = formatBeijingTime(cache.RateLimits.SevenDay.ResetsAt)
	// The cache is written only by the Claude Code statusline hook, so it
	// goes stale whenever the user works in the desktop app instead of the
	// CLI. Surface its age so the UI can say how old the numbers are.
	if info, statErr := os.Stat(cachePath); statErr == nil {
		st.UsageUpdatedAt = info.ModTime().In(beijingTZ).Format("2006-01-02 15:04")
		st.UsageAge = formatElapsed(info.ModTime())
		st.UsageStale = time.Since(info.ModTime()) > claudeUsageCacheMaxAge
	}
	return st
}

// CodexStatus checks whether codex CLI is installed and reads the latest local
// rate-limit snapshot from ~/.codex/sessions.
func CodexStatus() CodexUsageStatus {
	st := CodexUsageStatus{}
	if codexPath := macos.FindCLI("codex"); codexPath != "" {
		st.CodexInstalled = true
		st.CodexVersion = macos.CLIVersion(codexPath)
	}
	st.VPNConnected, st.VPNInterfaces = macos.UtunInterfaces()

	sessionsDir, err := CodexSessionsDir()
	if err != nil {
		return st
	}
	snapshot, ok := LoadLatestCodexRateLimitSnapshot(sessionsDir)
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

// CodexRateLimitSnapshot is one token_count event from a Codex rollout file.
type CodexRateLimitSnapshot struct {
	Timestamp    string
	PlanType     string
	LimitReached string
	Primary      CodexRateLimitWindow
	Secondary    CodexRateLimitWindow
}

// CodexRateLimitWindow is the raw window inside a snapshot.
type CodexRateLimitWindow struct {
	UsedPercent   float64 `json:"used_percent"`
	WindowMinutes int     `json:"window_minutes"`
	ResetsAt      int64   `json:"resets_at"`
	Available     bool
}

// LoadLatestCodexRateLimitSnapshot scans rollout-*.jsonl files under
// sessionsDir and returns the newest rate-limit snapshot.
func LoadLatestCodexRateLimitSnapshot(sessionsDir string) (CodexRateLimitSnapshot, bool) {
	var latest CodexRateLimitSnapshot
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
			snapshot, ok := ParseCodexRateLimitLine(line)
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

// ParseCodexRateLimitLine parses one rollout JSONL line; only token_count
// payloads with at least one window are accepted.
func ParseCodexRateLimitLine(line string) (CodexRateLimitSnapshot, bool) {
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
		return CodexRateLimitSnapshot{}, false
	}
	if entry.Payload.Type != "token_count" {
		return CodexRateLimitSnapshot{}, false
	}
	fromRaw := func(raw rawWindow) CodexRateLimitWindow {
		window := CodexRateLimitWindow{Available: raw.UsedPercent != nil || raw.WindowMinutes != nil || raw.ResetsAt != nil}
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
	rawPrimary := fromRaw(entry.Payload.RateLimit.Primary)
	rawSecondary := fromRaw(entry.Payload.RateLimit.Secondary)
	if !rawPrimary.Available && !rawSecondary.Available {
		return CodexRateLimitSnapshot{}, false
	}
	primary, secondary := ClassifyCodexRateLimitWindows(rawPrimary, rawSecondary)
	return CodexRateLimitSnapshot{
		Timestamp:    entry.Timestamp,
		PlanType:     entry.Payload.RateLimit.PlanType,
		LimitReached: entry.Payload.RateLimit.LimitReached,
		Primary:      primary,
		Secondary:    secondary,
	}, true
}

// ClassifyCodexRateLimitWindows returns (5h, 7d) windows by their duration
// rather than their position in the payload.
func ClassifyCodexRateLimitWindows(primary, secondary CodexRateLimitWindow) (CodexRateLimitWindow, CodexRateLimitWindow) {
	var fiveHour, sevenDay CodexRateLimitWindow
	for index, window := range []CodexRateLimitWindow{primary, secondary} {
		if !window.Available {
			continue
		}
		switch window.WindowMinutes {
		case 300:
			fiveHour = window
		case 10080:
			sevenDay = window
		case 0:
			// Older snapshots may omit window_minutes; retain their positional meaning.
			if index == 0 && !fiveHour.Available {
				fiveHour = window
			} else if index == 1 && !sevenDay.Available {
				sevenDay = window
			}
		}
	}
	return fiveHour, sevenDay
}

func codexWindowFromSnapshot(label string, window CodexRateLimitWindow) CodexUsageWindow {
	if !window.Available {
		return CodexUsageWindow{Label: label}
	}
	used := window.UsedPercent
	if window.ResetsAt > 0 && window.ResetsAt <= time.Now().Unix() {
		used = 0
	}
	left := 100 - used
	if left < 0 {
		left = 0
	}
	return CodexUsageWindow{
		Label:         label,
		UsedPercent:   used,
		LeftPercent:   left,
		WindowMinutes: window.WindowMinutes,
		ResetsAt:      window.ResetsAt,
		ResetIn:       formatResetCountdown(window.ResetsAt),
		ResetAt:       formatBeijingTime(window.ResetsAt),
		Available:     true,
	}
}

// formatElapsed renders how long ago t was, in the same compact style as
// formatResetCountdown.
func formatElapsed(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return "刚刚"
	}
	mins := int(diff.Minutes())
	hours := int(diff.Hours())
	days := hours / 24
	switch {
	case days >= 1:
		return fmt.Sprintf("%d天前", days)
	case hours >= 1:
		return fmt.Sprintf("%d小时前", hours)
	default:
		return fmt.Sprintf("%d分钟前", mins)
	}
}

// beijingTZ is China Standard Time (UTC+8, no DST); usage timestamps are
// always shown in it regardless of the Mac's own time zone.
var beijingTZ = time.FixedZone("CST", 8*3600)

// formatBeijingTime renders a unix timestamp as "2006-01-02 15:04:05" in
// Beijing time, or "" when ts is unset.
func formatBeijingTime(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).In(beijingTZ).Format("2006-01-02 15:04:05")
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
