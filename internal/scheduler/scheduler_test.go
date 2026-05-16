package scheduler

import (
	"testing"
	"time"
)

func TestParseEvery(t *testing.T) {
	next, err := ParseSchedule("@every 30s")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if got := next(now); got.Sub(now) != 30*time.Second {
		t.Fatalf("expected +30s, got %v", got.Sub(now))
	}
}

func TestParseDaily(t *testing.T) {
	next, err := ParseSchedule("@daily 09:30")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 1, 1, 8, 0, 0, 0, time.Local)
	if got := next(now); got.Hour() != 9 || got.Minute() != 30 || got.Day() != 1 {
		t.Fatalf("expected 09:30 today, got %v", got)
	}
	now2 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.Local)
	if got := next(now2); got.Day() != 2 || got.Hour() != 9 || got.Minute() != 30 {
		t.Fatalf("expected 09:30 next day, got %v", got)
	}
}

func TestParseHourly(t *testing.T) {
	next, err := ParseSchedule("@hourly :15")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.Local)
	if got := next(now); got.Hour() != 10 || got.Minute() != 15 {
		t.Fatalf("expected 10:15, got %v", got)
	}
	now2 := time.Date(2026, 1, 1, 10, 20, 0, 0, time.Local)
	if got := next(now2); got.Hour() != 11 || got.Minute() != 15 {
		t.Fatalf("expected 11:15, got %v", got)
	}
}

func TestParseErrors(t *testing.T) {
	for _, s := range []string{"", "@every", "@every 100ms", "@daily 25:00", "@hourly :99", "garbage"} {
		if _, err := ParseSchedule(s); err == nil {
			t.Errorf("%q should have errored", s)
		}
	}
}

func TestScheduleRuns(t *testing.T) {
	s := New()
	defer s.Remove("t")
	ch := make(chan struct{}, 1)
	if err := s.Schedule("t", "@every 1s", func() { ch <- struct{}{} }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("job did not fire within 3s")
	}
}
