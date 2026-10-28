package executor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ServiceStatus represents the runtime state of a service script.
type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusStopped ServiceStatus = "stopped"
	StatusUnknown ServiceStatus = "unknown"
)

// hatchPidDir holds Hatch-owned pid files and service logs.
const hatchPidDir = "/tmp/hatch"

// serviceActionTimeout bounds `<script> stop|status|reload` runs.
const serviceActionTimeout = 15 * time.Second

// ServiceManager controls start/stop/restart for a single service script.
//
// A service is started as `<script> start [args]` in its own process group and
// the group id is kept in pidFile. The service counts as running while any
// process of that group is alive, so start branches that background the real
// server (`nohup … &`) and exit stay tracked. Output goes to a log file next to
// the pid file — not a pipe, so the service survives Hatch quitting — and is
// tailed into the console while the group lives.
type ServiceManager struct {
	scriptPath string
	pidFile    string // Hatch-owned: process-group id of the last start
	em         *Emitter

	mu            sync.Mutex
	commands      []string // case arms detected in the script (start/stop/…)
	scriptPidFile string   // optional pid file the script maintains itself
}

func NewServiceManager(scriptPath, pidFile string, em *Emitter) *ServiceManager {
	if pidFile == "" {
		pidFile = defaultPidFile(scriptPath)
	}
	return &ServiceManager{
		scriptPath: scriptPath,
		pidFile:    pidFile,
		em:         em,
	}
}

// SplitPidFile separates a configured pid file into the Hatch-owned group file
// (empty = default) and a pid file the script writes itself: anything outside
// /tmp/hatch belongs to the script, and Hatch must not overwrite or delete it.
func SplitPidFile(pidFile string) (hatchFile, scriptFile string) {
	if pidFile == "" || strings.HasPrefix(filepath.Clean(pidFile), hatchPidDir+"/") {
		return pidFile, ""
	}
	return "", pidFile
}

// Configure records the script's detected commands and its own pid file.
// Action words are passed to the script only when it handles them.
func (s *ServiceManager) Configure(commands []string, scriptPidFile string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commands = append([]string(nil), commands...)
	s.scriptPidFile = scriptPidFile
}

func (s *ServiceManager) handles(cmd string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.commands {
		if c == cmd {
			return true
		}
	}
	return false
}

// Start launches the service in the background.
// Returns an error if the service is already running.
func (s *ServiceManager) Start(args []string) error {
	return s.StartWithEnv(args, nil)
}

// StartWithEnv launches the service with optional env overrides.
func (s *ServiceManager) StartWithEnv(args []string, env map[string]string) error {
	if st, pid := s.Status(); st == StatusRunning {
		return fmt.Errorf("service already running (pid %d)", pid)
	}
	if s.handles("start") {
		args = append([]string{"start"}, args...)
	}

	cmd := newBackgroundCmd("/bin/zsh", buildZshArgs(s.scriptPath, args)...)
	if len(env) > 0 {
		cmd.Env = mergeEnv(env)
	}
	logPath := s.logFile()
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}
	logf, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("open service log: %w", err)
	}
	cmd.Stdout, cmd.Stderr = logf, logf
	s.log("[hatch] command: "+buildScriptCommandForLog(s.scriptPath, args), "stdout")

	err = cmd.Start()
	logf.Close() // the child keeps its own descriptor
	if err != nil {
		s.status("error")
		return fmt.Errorf("start: %w", err)
	}
	pgid := cmd.Process.Pid // Setpgid: the leader's pid is the group id
	if err := s.writePid(pgid); err != nil {
		return fmt.Errorf("write pidfile: %w", err)
	}
	s.status("running")

	go func() { _ = cmd.Wait() }()
	go s.follow(pgid, logPath)
	return nil
}

// follow streams new log lines to the console until the process group exits,
// then reports the service stopped (unless a newer start replaced it).
func (s *ServiceManager) follow(pgid int, logPath string) {
	f, err := os.Open(logPath)
	if err != nil {
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	partial := ""
	for {
		alive := groupAlive(pgid)
		for {
			chunk, err := r.ReadString('\n')
			partial += chunk
			if err != nil {
				break
			}
			s.log(strings.TrimRight(partial, "\r\n"), "stdout")
			partial = ""
		}
		if !alive {
			if partial != "" {
				s.log(partial, "stdout")
			}
			if current := s.readPid(); current == "" || current == strconv.Itoa(pgid) {
				s.removePid()
				s.status("stopped")
			}
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// Stop runs the script's own `stop` branch when it has one, then sends SIGTERM
// to whatever is left of the service's process group.
func (s *ServiceManager) Stop() error {
	if s.handles("stop") {
		if err := s.RunAction("stop", nil); err != nil {
			s.log("[hatch] stop command: "+err.Error(), "stderr")
		}
	}
	if pid, err := s.pidFromFile(); err == nil {
		if groupAlive(pid) {
			s.log(fmt.Sprintf("[hatch] command: SIGTERM process group %d", pid), "stdout")
			if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
				return fmt.Errorf("signal: %w", err)
			}
		}
		s.removePid()
	}
	s.status("stopped")
	return nil
}

// RunAction runs `<script> <action>` (stop, status, reload, …) to completion
// and streams its output to the console.
func (s *ServiceManager) RunAction(action string, env map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), serviceActionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/zsh", buildZshArgs(s.scriptPath, []string{action})...)
	if len(env) > 0 {
		cmd.Env = mergeEnv(env)
	}
	s.log("[hatch] command: "+buildScriptCommandForLog(s.scriptPath, []string{action}), "stdout")
	out, err := cmd.CombinedOutput()
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line != "" {
			s.log(line, "stdout")
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("%s 超时（%s）", action, serviceActionTimeout)
	}
	return err
}

// Restart stops then starts the service.
func (s *ServiceManager) Restart(args []string) error {
	return s.RestartWithEnv(args, nil)
}

// RestartWithEnv stops then starts the service with env overrides.
func (s *ServiceManager) RestartWithEnv(args []string, env map[string]string) error {
	_ = s.Stop()
	return s.StartWithEnv(args, env)
}

// Status reports running while the started process group has a live member,
// or while the pid in the script's own pid file is alive (which also covers a
// service started outside Hatch).
func (s *ServiceManager) Status() (ServiceStatus, int) {
	if pid, err := s.pidFromFile(); err == nil {
		if groupAlive(pid) {
			return StatusRunning, pid
		}
		s.removePid()
	}
	s.mu.Lock()
	scriptPidFile := s.scriptPidFile
	s.mu.Unlock()
	if scriptPidFile != "" {
		if data, err := os.ReadFile(scriptPidFile); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
				if err := syscall.Kill(pid, 0); err == nil || err == syscall.EPERM {
					return StatusRunning, pid
				}
			}
		}
	}
	return StatusStopped, 0
}

// groupAlive reports whether any process in process group pgid exists.
func groupAlive(pgid int) bool {
	if pgid <= 0 {
		return false
	}
	err := syscall.Kill(-pgid, 0)
	return err == nil || err == syscall.EPERM
}

func (s *ServiceManager) log(line, stream string) {
	if s.em != nil {
		s.em.Log(s.scriptPath, line, stream)
	}
}

func (s *ServiceManager) status(st string) {
	if s.em != nil {
		s.em.Status(s.scriptPath, st)
	}
}

// --- PID file helpers ---

func (s *ServiceManager) logFile() string {
	return strings.TrimSuffix(s.pidFile, ".pid") + ".log"
}

func (s *ServiceManager) writePid(pid int) error {
	if err := os.MkdirAll(filepath.Dir(s.pidFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.pidFile, []byte(strconv.Itoa(pid)), 0o644)
}

func (s *ServiceManager) readPid() string {
	data, err := os.ReadFile(s.pidFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (s *ServiceManager) pidFromFile() (int, error) {
	raw := s.readPid()
	if raw == "" {
		return 0, fmt.Errorf("no pidfile at %s", s.pidFile)
	}
	pid, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid pid %q: %w", raw, err)
	}
	return pid, nil
}

func (s *ServiceManager) removePid() {
	_ = os.Remove(s.pidFile)
}

func defaultPidFile(scriptPath string) string {
	base := filepath.Base(scriptPath)
	return filepath.Join(hatchPidDir, base+".pid")
}
