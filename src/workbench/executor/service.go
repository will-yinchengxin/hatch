package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ServiceStatus represents the runtime state of a service script.
type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusStopped ServiceStatus = "stopped"
	StatusUnknown ServiceStatus = "unknown"
)

// ServiceManager controls start/stop/restart for a single service script.
type ServiceManager struct {
	scriptPath string
	pidFile    string
	em         *Emitter
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

// Start launches the service in the background.
// Returns an error if the service is already running.
func (s *ServiceManager) Start(args []string) error {
	return s.StartWithEnv(args, nil)
}

// StartWithEnv launches the service with optional env overrides.
func (s *ServiceManager) StartWithEnv(args []string, env map[string]string) error {
	if st, _ := s.Status(); st == StatusRunning {
		return fmt.Errorf("service already running (pid %s)", s.readPid())
	}

	cmdArgs := buildZshArgs(s.scriptPath, args)
	cmd := newBackgroundCmd("/bin/zsh", cmdArgs...)
	if len(env) > 0 {
		cmd.Env = mergeEnv(env)
	}
	if s.em != nil {
		s.em.Log(s.scriptPath, "[hatch] command: "+buildScriptCommandForLog(s.scriptPath, args), "stdout")
	}

	if err := cmd.Start(); err != nil {
		if s.em != nil {
			s.em.Status(s.scriptPath, "error")
		}
		return fmt.Errorf("start: %w", err)
	}

	if err := s.writePid(cmd.Process.Pid); err != nil {
		return fmt.Errorf("write pidfile: %w", err)
	}

	if s.em != nil {
		s.em.Status(s.scriptPath, "running")
	}

	// Detach: let the child own its process group.
	go func() { _ = cmd.Wait() }()
	return nil
}

// Stop sends SIGTERM to the process, then cleans up the PID file.
func (s *ServiceManager) Stop() error {
	pid, err := s.pidFromFile()
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		s.removePid()
		return nil
	}
	if s.em != nil {
		s.em.Log(s.scriptPath, fmt.Sprintf("[hatch] command: SIGTERM pid %d", pid), "stdout")
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		s.removePid()
		if err == os.ErrProcessDone {
			if s.em != nil {
				s.em.Status(s.scriptPath, "stopped")
			}
			return nil
		}
		return fmt.Errorf("signal: %w", err)
	}
	s.removePid()
	if s.em != nil {
		s.em.Status(s.scriptPath, "stopped")
	}
	return nil
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

// Status returns the current runtime status by probing the PID file.
func (s *ServiceManager) Status() (ServiceStatus, int) {
	pid, err := s.pidFromFile()
	if err != nil {
		return StatusStopped, 0
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		s.removePid()
		return StatusStopped, 0
	}
	// kill -0 checks liveness without sending a real signal
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		s.removePid()
		return StatusStopped, 0
	}
	return StatusRunning, pid
}

// --- PID file helpers ---

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
	return filepath.Join("/tmp/hatch", base+".pid")
}
