package zshrc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	blockBegin = "# >>> Hatch managed block >>>"
	blockEnd   = "# <<< Hatch managed block <<<"
)

var blockRe = regexp.MustCompile(
	`(?m)` + regexp.QuoteMeta(blockBegin) + `\n[\s\S]*?\n` + regexp.QuoteMeta(blockEnd) + `\n?`,
)

// ZshrcPath returns the path to ~/.zshrc, creating the file if absent.
func ZshrcPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(home, ".zshrc")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := os.WriteFile(p, []byte{}, 0o644); err != nil {
			return "", err
		}
	}
	return p, nil
}

// CurrentShell reports whether the running shell is zsh.
func CurrentShell() string {
	return os.Getenv("SHELL")
}

// ReadBlock returns the lines currently inside the Hatch managed block.
// Returns nil if no block exists yet.
func ReadBlock(zshrcPath string) ([]string, error) {
	data, err := os.ReadFile(zshrcPath)
	if err != nil {
		return nil, err
	}
	content := string(data)
	start := strings.Index(content, blockBegin)
	if start == -1 {
		return nil, nil
	}
	end := strings.Index(content, blockEnd)
	if end == -1 {
		return nil, nil
	}
	inner := content[start+len(blockBegin) : end]
	var lines []string
	for _, l := range strings.Split(strings.TrimLeft(inner, "\n"), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

// WriteBlock replaces (or creates) the Hatch managed block with the given
// export lines. Lines outside the block are never touched.
func WriteBlock(zshrcPath string, exports map[string]string) error {
	data, err := os.ReadFile(zshrcPath)
	if err != nil {
		return err
	}
	content := string(data)

	block := buildBlock(exports)

	if blockRe.MatchString(content) {
		content = blockRe.ReplaceAllString(content, strings.ReplaceAll(block, "$", "$$"))
	} else {
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += block
	}

	return os.WriteFile(zshrcPath, []byte(content), 0o644)
}

// RemoveBlock removes the entire Hatch managed block from ~/.zshrc.
func RemoveBlock(zshrcPath string) error {
	data, err := os.ReadFile(zshrcPath)
	if err != nil {
		return err
	}
	cleaned := blockRe.ReplaceAllString(string(data), "")
	return os.WriteFile(zshrcPath, []byte(cleaned), 0o644)
}

// Backup creates a non-overwriting timestamped copy beside ~/.zshrc.
func Backup(zshrcPath string) (string, error) {
	in, err := os.Open(zshrcPath)
	if err != nil {
		return "", err
	}
	defer in.Close()

	stamp := time.Now().Format("20060102-150405")
	for i := 0; ; i++ {
		suffix := stamp
		if i > 0 {
			suffix = fmt.Sprintf("%s-%d", stamp, i)
		}
		dst := zshrcPath + ".hatch-backup-" + suffix
		out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return "", err
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return dst, nil
	}
}

// WriteContent replaces the whole zshrc file with content.
func WriteContent(zshrcPath, content string) error {
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(zshrcPath, []byte(content), 0o644)
}

func buildBlock(exports map[string]string) string {
	keys := make([]string, 0, len(exports))
	for k := range exports {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(blockBegin + "\n")
	pathVars := make([]string, 0, len(keys))
	for _, name := range keys {
		if name == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("export %s=%s\n", name, shellValue(exports[name])))
		pathVars = append(pathVars, "$"+name)
	}
	if len(pathVars) > 0 {
		sb.WriteString(fmt.Sprintf("export PATH=%s:$PATH\n", strings.Join(pathVars, ":")))
	}
	sb.WriteString(blockEnd + "\n")
	return sb.String()
}

func shellValue(value string) string {
	if value == "" {
		return `""`
	}
	if isShellBareValue(value) {
		return value
	}
	return shellSingleQuote(value)
}

func isShellBareValue(value string) bool {
	for _, c := range value {
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			strings.ContainsRune("/._-+:", c) {
			continue
		}
		return false
	}
	return true
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
