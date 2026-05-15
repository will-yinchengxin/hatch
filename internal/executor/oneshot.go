package executor

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"
)

// OneshotResult is returned when the process exits.
type OneshotResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// RunOneshot executes scriptPath with args in a login zsh so ~/.zshrc is sourced.
// Each stdout/stderr line is forwarded via emitter in real time.
// id is an opaque caller-supplied correlation token (e.g. script path or UUID).
func RunOneshot(id, scriptPath string, args []string, em *Emitter) (*OneshotResult, error) {
	var logFn func(line, stream string)
	var doneFn func(code int)

	if em != nil {
		em.Status(id, "running")
		logFn = func(line, stream string) { em.Log(id, line, stream) }
		doneFn = func(code int) {
			em.Done(id, code)
			if code == 0 {
				em.Status(id, "stopped")
			} else {
				em.Status(id, "error")
			}
		}
	}

	return runOneshotInternal(id, scriptPath, args, logFn, doneFn)
}

// runOneshotInternal is the testable core: it accepts plain callbacks instead
// of a Wails-context-bound Emitter.
func runOneshotInternal(
	_ string,
	scriptPath string,
	args []string,
	logFn func(line, stream string),
	doneFn func(code int),
) (*OneshotResult, error) {
	cmdArgs := buildZshArgs(scriptPath, args)
	cmd := exec.Command("/bin/zsh", cmdArgs...)
	if logFn != nil {
		logFn("[hatch] command: "+buildScriptCommandForLog(scriptPath, args), "stdout")
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	var (
		stdoutBuf, stderrBuf []byte
		mu                   sync.Mutex
		wg                   sync.WaitGroup
	)

	collect := func(scanner *bufio.Scanner, stream string, buf *[]byte) {
		defer wg.Done()
		for scanner.Scan() {
			line := scanner.Text()
			mu.Lock()
			*buf = append(*buf, []byte(line+"\n")...)
			mu.Unlock()
			if logFn != nil {
				logFn(line, stream)
			}
		}
	}

	wg.Add(2)
	go collect(bufio.NewScanner(stdoutPipe), "stdout", &stdoutBuf)
	go collect(bufio.NewScanner(stderrPipe), "stderr", &stderrBuf)
	wg.Wait()

	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}

	if doneFn != nil {
		doneFn(exitCode)
	}

	return &OneshotResult{
		ExitCode: exitCode,
		Stdout:   string(stdoutBuf),
		Stderr:   string(stderrBuf),
	}, nil
}

// buildZshArgs constructs the argv for a login zsh command that explicitly
// sources ~/.zshrc before executing the script. Wails apps launched from Finder
// do not inherit an interactive terminal environment, and zsh -l -c does not
// read ~/.zshrc by itself.
func buildZshArgs(scriptPath string, args []string) []string {
	cmd := `if [[ -r "${ZDOTDIR:-$HOME}/.zshrc" ]]; then source "${ZDOTDIR:-$HOME}/.zshrc"; fi; exec ` + shellQuote(scriptPath)
	for _, a := range args {
		cmd += " " + shellQuote(a)
	}
	return []string{"-l", "-c", cmd}
}

func buildScriptCommandForLog(scriptPath string, args []string) string {
	cmd := shellQuote(scriptPath)
	for _, a := range args {
		cmd += " " + shellQuote(a)
	}
	return cmd
}

// shellQuote wraps s in single quotes, escaping any embedded single quotes.
func shellQuote(s string) string {
	result := "'"
	for _, c := range s {
		if c == '\'' {
			result += "'\\''"
		} else {
			result += string(c)
		}
	}
	return result + "'"
}
