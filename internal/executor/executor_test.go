package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- shellQuote ---

func TestShellQuote(t *testing.T) {
	cases := []struct{ in, want string }{
		{"hello", "'hello'"},
		{"hello world", "'hello world'"},
		{"it's", "'it'\\''s'"},
		{"/usr/local/bin/tool", "'/usr/local/bin/tool'"},
	}
	for _, c := range cases {
		if got := shellQuote(c.in); got != c.want {
			t.Errorf("shellQuote(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- oneshot ---

func TestRunOneshot_Echo(t *testing.T) {
	res, err := RunOneshot("test-id", "/bin/echo", []string{"hello hatch"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code: %d", res.ExitCode)
	}
	if !strings.Contains(res.Stdout, "hello hatch") {
		t.Errorf("stdout: %q", res.Stdout)
	}
}

func TestRunOneshot_ExitCode(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fail.sh")
	if err := os.WriteFile(script, []byte("#!/bin/bash\nexit 42\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := RunOneshot("fail", script, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", res.ExitCode)
	}
}

func TestRunOneshot_EmitsEvents(t *testing.T) {
	var logs []LogEvent
	var dones []DoneEvent

	// Use a fake emitter by wrapping the functions directly
	em := &testEmitter{
		onLog:  func(e LogEvent) { logs = append(logs, e) },
		onDone: func(e DoneEvent) { dones = append(dones, e) },
	}

	res, err := runOneshotWithTestEmitter("emit-id", "/bin/echo", []string{"line1"}, em)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code: %d", res.ExitCode)
	}
	if len(logs) == 0 {
		t.Error("expected at least one log event")
	}
	if len(dones) != 1 {
		t.Errorf("expected 1 done event, got %d", len(dones))
	}
	if dones[0].ID != "emit-id" {
		t.Errorf("done event id: %q", dones[0].ID)
	}
}

func TestRunOneshot_Stderr(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "errout.sh")
	if err := os.WriteFile(script, []byte("#!/bin/bash\necho out\n>&2 echo err\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := RunOneshot("stderr-test", script, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Stdout, "out") {
		t.Errorf("stdout: %q", res.Stdout)
	}
	if !strings.Contains(res.Stderr, "err") {
		t.Errorf("stderr: %q", res.Stderr)
	}
}

// --- service manager ---

func TestServiceManager_StartStop(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "daemon.sh")
	pidFile := filepath.Join(dir, "daemon.pid")
	// Long-running background process
	if err := os.WriteFile(script, []byte("#!/bin/bash\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	svc := NewServiceManager(script, pidFile, nil)

	if err := svc.Start(nil); err != nil {
		t.Fatalf("start: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	status, pid := svc.Status()
	if status != StatusRunning {
		t.Errorf("expected running, got %s", status)
	}
	if pid == 0 {
		t.Error("expected non-zero pid")
	}

	if err := svc.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	status, _ = svc.Status()
	if status != StatusStopped {
		t.Errorf("expected stopped, got %s", status)
	}
}

func TestServiceManager_DoubleStart(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "daemon.sh")
	pidFile := filepath.Join(dir, "daemon.pid")
	if err := os.WriteFile(script, []byte("#!/bin/bash\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	svc := NewServiceManager(script, pidFile, nil)
	if err := svc.Start(nil); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	defer svc.Stop()

	if err := svc.Start(nil); err == nil {
		t.Error("expected error on double start")
	}
}

func TestServiceManager_StaleStopCleansUp(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "stale.pid")
	// Write a PID that definitely doesn't exist
	if err := os.WriteFile(pidFile, []byte("999999999"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewServiceManager("/nonexistent", pidFile, nil)
	status, _ := svc.Status()
	if status != StatusStopped {
		t.Errorf("stale pid should report stopped, got %s", status)
	}
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Error("stale pidfile should have been removed")
	}
}

func TestServiceManager_Restart(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "daemon.sh")
	pidFile := filepath.Join(dir, "daemon.pid")
	if err := os.WriteFile(script, []byte("#!/bin/bash\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	svc := NewServiceManager(script, pidFile, nil)
	if err := svc.Start(nil); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	_, pid1 := svc.Status()
	if err := svc.Restart(nil); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	defer svc.Stop()

	status, pid2 := svc.Status()
	if status != StatusRunning {
		t.Errorf("expected running after restart, got %s", status)
	}
	if pid1 == pid2 {
		t.Error("expected different pid after restart")
	}
}

// --- test helpers (avoid Wails context dependency in tests) ---

type testEmitter struct {
	onLog    func(LogEvent)
	onDone   func(DoneEvent)
	onStatus func(StatusEvent)
}

// runOneshotWithTestEmitter mirrors RunOneshot but uses testEmitter instead of *Emitter.
func runOneshotWithTestEmitter(id, scriptPath string, args []string, em *testEmitter) (*OneshotResult, error) {
	// Delegate to RunOneshot with nil Emitter, then re-emit via testEmitter.
	// To actually test emission we wire a thin shim.
	return runOneshotInternal(id, scriptPath, args, func(line, stream string) {
		if em != nil && em.onLog != nil {
			em.onLog(LogEvent{ID: id, Line: line, Stream: stream})
		}
	}, func(code int) {
		if em != nil && em.onDone != nil {
			em.onDone(DoneEvent{ID: id, ExitCode: code})
		}
	})
}
