// Package macos wraps the macOS command-line integrations Hatch relies on:
// login-shell CLI lookup, Finder open, notifications, sounds and speech.
package macos

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// FindCLI locates a user-installed command. Finder/Dock launches do not inherit
// the interactive shell PATH, so it falls back to `zsh -l -c command -v`.
func FindCLI(name string) string {
	if path, err := exec.LookPath(name); err == nil && path != "" {
		return path
	}
	out, err := exec.Command("/bin/zsh", "-l", "-c", "command -v "+name).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// CLIVersion runs `<path> --version` and returns the trimmed output.
func CLIVersion(path string) string {
	out, _ := exec.Command(path, "--version").Output()
	return strings.TrimSpace(string(out))
}

// LoginShellPATH returns $PATH as seen by a login zsh.
func LoginShellPATH() (string, error) {
	out, err := exec.Command("/bin/zsh", "-l", "-c", "echo $PATH").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// SourceZshrcAsync runs `source ~/.zshrc` in a background login shell.
func SourceZshrcAsync() {
	go func() {
		_ = exec.Command("/bin/zsh", "-l", "-c", "source ~/.zshrc 2>/dev/null || true").Run()
	}()
}

// Open opens a file, directory or URL with the default macOS handler.
func Open(target string) error {
	return exec.Command("open", target).Start()
}

// Notify triggers a macOS notification banner via osascript. Best-effort:
// errors are swallowed (the user may have notifications disabled).
func Notify(title, body string) {
	esc := func(s string) string {
		return strings.ReplaceAll(s, `"`, `\"`)
	}
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, esc(body), esc(title))
	_ = exec.Command("osascript", "-e", script).Start()
}

// Beep plays the system alert sound.
func Beep() {
	_ = exec.Command("osascript", "-e", "beep").Start()
}

// Say speaks text with the macOS `say` command without blocking the caller.
func Say(text string) error {
	sayPath := "/usr/bin/say"
	if _, err := os.Stat(sayPath); err != nil {
		var lookErr error
		sayPath, lookErr = exec.LookPath("say")
		if lookErr != nil {
			return fmt.Errorf("macOS say command not found")
		}
	}
	cmd := exec.Command(sayPath, text)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}

// UtunInterfaces lists "utunN: <ipv4>" for every utun interface with an IPv4
// address, which is how Hatch detects that some VPN tunnel is up.
func UtunInterfaces() (bool, []string) {
	ifOut, _ := exec.Command("ifconfig").Output()
	var currentIface string
	var interfaces []string
	for _, line := range strings.Split(string(ifOut), "\n") {
		if strings.HasPrefix(line, "utun") {
			currentIface = strings.SplitN(line, ":", 2)[0]
		} else if currentIface != "" && strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				interfaces = append(interfaces, currentIface+": "+fields[1])
			}
			currentIface = ""
		} else if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
			currentIface = ""
		}
	}
	return len(interfaces) > 0, interfaces
}

// PIDAlive reports whether a decimal PID string names a running process.
func PIDAlive(pidStr string) bool {
	if pidStr == "" {
		return false
	}
	pid := 0
	for _, c := range pidStr {
		if c < '0' || c > '9' {
			return false
		}
		pid = pid*10 + int(c-'0')
	}
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Sending signal 0 on darwin checks existence without delivering.
	return p.Signal(syscall.Signal(0)) == nil
}
