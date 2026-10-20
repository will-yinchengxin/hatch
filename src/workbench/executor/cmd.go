package executor

import (
	"os/exec"
	"syscall"
)

// newBackgroundCmd creates a Cmd that runs in its own process group (detached).
// This prevents SIGINT to the parent from killing the child.
func newBackgroundCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}
