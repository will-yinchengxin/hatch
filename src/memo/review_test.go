package memo

import (
	"testing"
	"time"
)

// Mirrors the review buttons: 不认识 (hard) → tomorrow and reset, 认识 (known)
// → 3+2·level days, 很熟 (mastered) → 14+3·level days.
func TestApplyReviewIntervals(t *testing.T) {
	now := time.Date(2026, 10, 28, 15, 0, 0, 0, time.Local)
	days := func(e VocabEntry) int {
		return int(time.Unix(e.NextReviewAt, 0).Sub(now).Hours()/24 + 0.5)
	}
	if e := ApplyReview(VocabEntry{ReviewLevel: 3}, "hard", now); e.ReviewLevel != 0 || days(e) != 1 {
		t.Fatalf("hard: level %d, %d days", e.ReviewLevel, days(e))
	}
	if e := ApplyReview(VocabEntry{ReviewLevel: 0}, "known", now); e.ReviewLevel != 1 || days(e) != 5 {
		t.Fatalf("known: level %d, %d days", e.ReviewLevel, days(e))
	}
	if e := ApplyReview(VocabEntry{ReviewLevel: 0}, "mastered", now); e.ReviewLevel != 2 || days(e) != 20 {
		t.Fatalf("mastered: level %d, %d days", e.ReviewLevel, days(e))
	}
}
