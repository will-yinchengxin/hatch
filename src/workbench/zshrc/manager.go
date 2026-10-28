package zshrc

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"hatch/core/fsutil"
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

// Var is one export in the managed block. InPath appends $Name to PATH.
type Var struct {
	Name   string
	Value  string
	InPath bool
}

// WriteBlock replaces (or creates) the Hatch managed block with the given
// export lines, adding every variable to PATH. Lines outside the block are
// never touched.
func WriteBlock(zshrcPath string, exports map[string]string) error {
	vars := make([]Var, 0, len(exports))
	for name, value := range exports {
		vars = append(vars, Var{Name: name, Value: value, InPath: true})
	}
	return WriteVars(zshrcPath, vars)
}

// WriteVars replaces (or creates) the Hatch managed block. Only variables
// with InPath set are appended to PATH; the rest are plain exports.
func WriteVars(zshrcPath string, vars []Var) error {
	data, err := os.ReadFile(zshrcPath)
	if err != nil && !os.IsNotExist(err) { // macOS doesn't create ~/.zshrc by default
		return err
	}
	content := string(data)

	block := buildBlock(vars)

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
	return fsutil.Backup(zshrcPath)
}

// WriteContent replaces the whole zshrc file with content.
func WriteContent(zshrcPath, content string) error {
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(zshrcPath, []byte(content), 0o644)
}

func buildBlock(vars []Var) string {
	sorted := make([]Var, 0, len(vars))
	for _, v := range vars {
		if v.Name != "" {
			sorted = append(sorted, v)
		}
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	var sb strings.Builder
	sb.WriteString(blockBegin + "\n")
	pathVars := make([]string, 0, len(sorted))
	for _, v := range sorted {
		sb.WriteString(fmt.Sprintf("export %s=%s\n", v.Name, shellValue(v.Value)))
		if v.InPath {
			pathVars = append(pathVars, "$"+v.Name)
		}
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
