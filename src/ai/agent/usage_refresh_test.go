package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseClaudeRateLimitStream(t *testing.T) {
	stream := strings.Join([]string{
		`{"type":"system","subtype":"init"}`,
		`{"type":"assistant","message":{}}`,
		`{"type":"rate_limit_event","rate_limit_info":{"status":"allowed","unifiedWindows":{"five_hour":{"utilization":0.03,"resetsAt":1792741200},"seven_day":{"utilization":0.13,"resetsAt":1793037600}}}}`,
		`{"type":"result","subtype":"success","is_error":false,"result":"Hi"}`,
	}, "\n")
	five, seven, resultErr, ok := ParseClaudeRateLimitStream(strings.NewReader(stream))
	if !ok || resultErr != "" {
		t.Fatalf("ok=%v resultErr=%q", ok, resultErr)
	}
	if five.UsedPercentage != 3 || five.ResetsAt != 1792741200 {
		t.Fatalf("five_hour = %+v", five)
	}
	if seven.UsedPercentage != 13 || seven.ResetsAt != 1793037600 {
		t.Fatalf("seven_day = %+v", seven)
	}
}

func TestParseClaudeRateLimitStreamReportsResultError(t *testing.T) {
	stream := `{"type":"result","subtype":"success","is_error":true,"result":"Not logged in · Please run /login"}`
	_, _, resultErr, ok := ParseClaudeRateLimitStream(strings.NewReader(stream))
	if ok || resultErr != "Not logged in · Please run /login" {
		t.Fatalf("ok=%v resultErr=%q", ok, resultErr)
	}
}

func TestWriteClaudeUsageCacheKeepsPlanType(t *testing.T) {
	path := filepath.Join(t.TempDir(), "will-statusline", "last.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"rate_limits":{},"plan_type":"max"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteClaudeUsageCache(path, ClaudeUsageWindow{UsedPercentage: 42.5, ResetsAt: 1}, ClaudeUsageWindow{UsedPercentage: 8, ResetsAt: 2}); err != nil {
		t.Fatal(err)
	}
	var got struct {
		PlanType   string `json:"plan_type"`
		RateLimits struct {
			FiveHour ClaudeUsageWindow `json:"five_hour"`
			SevenDay ClaudeUsageWindow `json:"seven_day"`
		} `json:"rate_limits"`
	}
	data, _ := os.ReadFile(path)
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.PlanType != "max" || got.RateLimits.FiveHour.UsedPercentage != 42.5 || got.RateLimits.SevenDay.ResetsAt != 2 {
		t.Fatalf("cache = %s", data)
	}
}
