// Package macos wraps the macOS command-line integrations Hatch relies on:
// login-shell CLI lookup, Finder open, notifications, sounds and speech.
package macos

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// FindCLI locates a user-installed command. Finder/Dock launches do not inherit
// the interactive shell PATH, so it falls back to a login zsh that also sources
// ~/.zshrc: PATH entries like ~/.local/bin usually live there, and `zsh -l`
// alone can resolve a stale copy (e.g. an old npm install in /usr/local/bin).
func FindCLI(name string) string {
	if path, err := exec.LookPath(name); err == nil && path != "" {
		return path
	}
	out, err := exec.Command("/bin/zsh", "-l", "-c", `source ~/.zshrc >/dev/null 2>&1; command -v "$0"`, name).Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
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

// Reveal selects target in a Finder window.
func Reveal(target string) error {
	return exec.Command("open", "-R", target).Start()
}

// MoveToTrash moves path into ~/.Trash (renaming on collision) so the user can
// still restore it from Finder. Returns the new location.
func MoveToTrash(path string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	trash := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trash, 0o700); err != nil {
		return "", err
	}
	base := filepath.Base(path)
	dst := filepath.Join(trash, base)
	if _, err := os.Stat(dst); err == nil {
		ext := filepath.Ext(base)
		dst = filepath.Join(trash, fmt.Sprintf("%s %s%s", strings.TrimSuffix(base, ext), time.Now().Format("150405"), ext))
	}
	if err := os.Rename(path, dst); err != nil {
		// A file on another volume can't be renamed into ~/.Trash; let Finder
		// move it to that volume's own trash instead.
		if errors.Is(err, syscall.EXDEV) {
			abs, _ := filepath.Abs(path)
			script := fmt.Sprintf(`tell application "Finder" to delete (POSIX file %q as alias)`, abs)
			if out, ferr := exec.Command("osascript", "-e", script).CombinedOutput(); ferr != nil {
				return "", fmt.Errorf("移到废纸篓失败: %s", strings.TrimSpace(string(out)))
			}
			return "", nil
		}
		return "", err
	}
	return dst, nil
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
