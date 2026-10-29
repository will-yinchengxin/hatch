package agent

import "testing"

func TestFormatBeijingTime(t *testing.T) {
	if got := formatBeijingTime(0); got != "" {
		t.Fatalf("formatBeijingTime(0) = %q, want empty", got)
	}
	// 2026-10-30 02:00:00 UTC is 10:00:00 in Beijing.
	if got, want := formatBeijingTime(1793325600), "2026-10-30 10:00:00"; got != want {
		t.Fatalf("formatBeijingTime = %q, want %q", got, want)
	}
}
