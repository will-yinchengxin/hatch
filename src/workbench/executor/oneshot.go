package executor

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// OneshotResult is returned when the process exits.
type OneshotResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// NotifyHook is invoked after a oneshot completes (used for OS notifications).
type NotifyHook func(exitCode int, durationMs int64)

// RunOneshot executes scriptPath with args in a login zsh.
// Backwards-compatible wrapper around RunOneshotWithEnv.
func RunOneshot(id, scriptPath string, args []string, em *Emitter) (*OneshotResult, error) {
	return RunOneshotWithEnv(id, scriptPath, args, nil, em, nil)
}

// RunOneshotWithEnv adds env injection, run-history archiving, and a completion hook.
func RunOneshotWithEnv(id, scriptPath string, args []string, env map[string]string, em *Emitter, hook NotifyHook) (*OneshotResult, error) {
	var logFn func(line, stream string)
	var doneFn func(code int)

	rec, _ := newRunRecorder(scriptPath, args) // best-effort archive

	if em != nil {
		em.Status(id, "running")
	}
	logFn = func(line, stream string) {
		if em != nil {
			em.Log(id, line, stream)
		}
		rec.write(line, stream)
	}
	doneFn = func(code int) {
		if em != nil {
			em.Done(id, code)
			if code == 0 {
				em.Status(id, "stopped")
			} else {
				em.Status(id, "error")
			}
		}
		var durMs int64
		if rec != nil {
			rec.finalize(code)
			durMs = rec.record.DurationMs
		}
		if hook != nil {
			hook(code, durMs)
		}
	}

	return runOneshotInternal(id, scriptPath, args, env, logFn, doneFn)
}

func runOneshotInternal(
	_ string,
	scriptPath string,
	args []string,
	env map[string]string,
	logFn func(line, stream string),
	doneFn func(code int),
) (*OneshotResult, error) {
	cmdArgs := buildZshArgs(scriptPath, args)
	cmd := exec.Command("/bin/zsh", cmdArgs...)
	if len(env) > 0 {
		cmd.Env = mergeEnv(env)
	}
	if logFn != nil {
		logFn("[hatch] command: "+buildScriptCommandForLog(scriptPath, args), "stdout")
		if len(env) > 0 {
			for k, v := range env {
				logFn("[hatch] env: "+k+"="+v, "stdout")
			}
		}
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

// mergeEnv combines os.Environ() with override map (override wins).
func mergeEnv(override map[string]string) []string {
	out := os.Environ()
	for k, v := range override {
		out = append(out, k+"="+v)
	}
	return out
}

// buildZshArgs constructs the argv for a login zsh that sources ~/.zshrc.
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
