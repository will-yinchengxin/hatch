package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"hatch/core/macos"
)

// claudeUsageProbeArgs is the cheapest print-mode turn that still makes the
// API return the account's rate-limit headers: Haiku, no tools, no MCP, no
// hooks, no saved session. stream-json is required for the rate_limit_event.
var claudeUsageProbeArgs = []string{
	"-p",
	"--model", "haiku",
	"--output-format", "stream-json",
	"--verbose",
	"--no-session-persistence",
	"--strict-mcp-config",
	"--tools", "",
	"--settings", `{"disableAllHooks":true}`,
	"hello",
}

// ClaudeUsageWindow is one window from a rate_limit_event, as a percentage.
type ClaudeUsageWindow struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

// RefreshClaudeUsage runs a one-line Haiku prompt through the claude CLI and
// writes the returned 5h/7d rate limits into the statusline cache, so usage
// stays current even when the user never opens an interactive CLI session.
func RefreshClaudeUsage(ctx context.Context) error {
	claudePath := macos.FindCLI("claude")
	if claudePath == "" {
		return errors.New("claude CLI 未安装")
	}
	// Run through a login zsh that sources ~/.zshrc so proxy/ANTHROPIC_* env
	// matches the user's terminal (Finder/Dock launches do not inherit it).
	args := append([]string{"-l", "-c", `source ~/.zshrc >/dev/null 2>&1; exec "$0" "$@"`, claudePath}, claudeUsageProbeArgs...)
	cmd := exec.CommandContext(ctx, "/bin/zsh", args...)
	// A neutral cwd keeps project CLAUDE.md / settings out of the probe.
	cmd.Dir = os.TempDir()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	fiveHour, sevenDay, resultErr, ok := ParseClaudeRateLimitStream(&stdout)
	if !ok {
		if ctx.Err() != nil {
			return fmt.Errorf("claude 调用超时: %w", ctx.Err())
		}
		if resultErr != "" {
			return fmt.Errorf("claude 调用失败: %s", resultErr)
		}
		if runErr != nil {
			return fmt.Errorf("claude 调用失败: %v %s", runErr, lastLine(stderr.String()))
		}
		return errors.New("claude 未返回用量信息（仅 Pro/Max 订阅登录会返回 5h/7d 额度）")
	}

	cachePath, err := ClaudeUsageCachePath()
	if err != nil {
		return err
	}
	return WriteClaudeUsageCache(cachePath, fiveHour, sevenDay)
}

// ParseClaudeRateLimitStream scans claude stream-json output for the last
// rate_limit_event carrying both unified windows. When none is found it also
// returns the error text of a failed result line, if any.
func ParseClaudeRateLimitStream(r io.Reader) (fiveHour, sevenDay ClaudeUsageWindow, resultErr string, ok bool) {
	type rawWindow struct {
		Utilization *float64 `json:"utilization"`
		ResetsAt    int64    `json:"resetsAt"`
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		var line struct {
			Type          string `json:"type"`
			IsError       bool   `json:"is_error"`
			Result        string `json:"result"`
			RateLimitInfo struct {
				UnifiedWindows struct {
					FiveHour *rawWindow `json:"five_hour"`
					SevenDay *rawWindow `json:"seven_day"`
				} `json:"unifiedWindows"`
			} `json:"rate_limit_info"`
		}
		if json.Unmarshal(sc.Bytes(), &line) != nil {
			continue
		}
		switch line.Type {
		case "rate_limit_event":
			w := line.RateLimitInfo.UnifiedWindows
			if w.FiveHour == nil || w.SevenDay == nil || w.FiveHour.Utilization == nil || w.SevenDay.Utilization == nil {
				continue
			}
			fiveHour = ClaudeUsageWindow{UsedPercentage: utilizationPct(*w.FiveHour.Utilization), ResetsAt: w.FiveHour.ResetsAt}
			sevenDay = ClaudeUsageWindow{UsedPercentage: utilizationPct(*w.SevenDay.Utilization), ResetsAt: w.SevenDay.ResetsAt}
			ok = true
		case "result":
			if line.IsError {
				resultErr = strings.TrimSpace(line.Result)
			}
		}
	}
	return fiveHour, sevenDay, resultErr, ok
}

// WriteClaudeUsageCache writes the statusline cache format, keeping any
// plan_type already recorded there.
func WriteClaudeUsageCache(path string, fiveHour, sevenDay ClaudeUsageWindow) error {
	var planType any
	if data, err := os.ReadFile(path); err == nil {
		var prev struct {
			PlanType any `json:"plan_type"`
		}
		if json.Unmarshal(data, &prev) == nil {
			planType = prev.PlanType
		}
	}
	out := map[string]any{
		"rate_limits": map[string]ClaudeUsageWindow{
			"five_hour": fiveHour,
			"seven_day": sevenDay,
		},
		"plan_type": planType,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// utilizationPct converts a 0–1 utilization into a percentage with one decimal.
func utilizationPct(u float64) float64 { return math.Round(u*1000) / 10 }

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}
