// Package scheduler is a tiny in-process scheduler for Hatch.
//
// Supported schedule formats (parsed by ParseSchedule):
//   - "@every 30s"   — interval (s, m, h)
//   - "@daily 09:00" — every day at HH:MM (24h, local time)
//   - "@hourly :15"  — every hour at :MM
//
// This is intentionally minimal — we don't want a real cron dep for v1.2.
// If users need cron expressions later, we can swap in robfig/cron.
package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Job is one scheduled invocation.
type Job struct {
	ID       string
	Schedule string
	Run      func()
}

// nextFireFn returns the next absolute time the job should fire.
type nextFireFn func(now time.Time) time.Time

type runningJob struct {
	job      Job
	nextFire nextFireFn
	stop     chan struct{}
}

// Scheduler manages a set of jobs in their own goroutines.
type Scheduler struct {
	mu   sync.Mutex
	jobs map[string]*runningJob
}

func New() *Scheduler {
	return &Scheduler{jobs: make(map[string]*runningJob)}
}

// Schedule adds (or replaces) a job. Schedule is parsed via ParseSchedule.
func (s *Scheduler) Schedule(id, schedule string, run func()) error {
	next, err := ParseSchedule(schedule)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.jobs[id]; ok {
		close(existing.stop)
	}
	rj := &runningJob{
		job:      Job{ID: id, Schedule: schedule, Run: run},
		nextFire: next,
		stop:     make(chan struct{}),
	}
	s.jobs[id] = rj
	go s.loop(rj)
	return nil
}

// Remove stops and forgets a job.
func (s *Scheduler) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rj, ok := s.jobs[id]; ok {
		close(rj.stop)
		delete(s.jobs, id)
	}
}

// List returns the registered jobs (for inspection).
func (s *Scheduler) List() []Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Job, 0, len(s.jobs))
	for _, rj := range s.jobs {
		out = append(out, rj.job)
	}
	return out
}

// NextRun returns the next scheduled fire time for one job.
func (s *Scheduler) NextRun(id string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rj, ok := s.jobs[id]
	if !ok {
		return time.Time{}, false
	}
	return rj.nextFire(time.Now()), true
}

func (s *Scheduler) loop(rj *runningJob) {
	for {
		next := rj.nextFire(time.Now())
		delay := time.Until(next)
		if delay < 0 {
			delay = time.Second
		}
		timer := time.NewTimer(delay)
		select {
		case <-rj.stop:
			timer.Stop()
			return
		case <-timer.C:
			func() {
				defer func() { _ = recover() }()
				rj.job.Run()
			}()
		}
	}
}

// ParseSchedule parses a Hatch schedule string and returns a "next fire" function.
func ParseSchedule(s string) (nextFireFn, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty schedule")
	}
	switch {
	case strings.HasPrefix(s, "@every "):
		rest := strings.TrimSpace(strings.TrimPrefix(s, "@every "))
		d, err := time.ParseDuration(rest)
		if err != nil {
			return nil, fmt.Errorf("invalid duration: %w", err)
		}
		if d < time.Second {
			return nil, fmt.Errorf("interval too small (min 1s)")
		}
		return func(now time.Time) time.Time { return now.Add(d) }, nil
	case strings.HasPrefix(s, "@daily "):
		hh, mm, err := parseHHMM(strings.TrimSpace(strings.TrimPrefix(s, "@daily ")))
		if err != nil {
			return nil, err
		}
		return func(now time.Time) time.Time {
			t := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, 0, 0, now.Location())
			if !t.After(now) {
				t = t.Add(24 * time.Hour)
			}
			return t
		}, nil
	case strings.HasPrefix(s, "@hourly :"):
		mmStr := strings.TrimPrefix(s, "@hourly :")
		mm, err := strconv.Atoi(strings.TrimSpace(mmStr))
		if err != nil || mm < 0 || mm > 59 {
			return nil, fmt.Errorf("invalid minute: %q", mmStr)
		}
		return func(now time.Time) time.Time {
			t := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), mm, 0, 0, now.Location())
			if !t.After(now) {
				t = t.Add(time.Hour)
			}
			return t
		}, nil
	}
	return nil, fmt.Errorf("unrecognized schedule (supported: @every Ns/Nm/Nh, @daily HH:MM, @hourly :MM)")
}

func parseHHMM(s string) (int, int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid HH:MM: %q", s)
	}
	hh, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	mm, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return 0, 0, fmt.Errorf("invalid HH:MM: %q", s)
	}
	return hh, mm, nil
}
