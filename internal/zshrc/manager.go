package zshrc

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
		content = blockRe.ReplaceAllString(content, block)
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

func buildBlock(exports map[string]string) string {
	keys := make([]string, 0, len(exports))
	for k := range exports {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(blockBegin + "\n")
	for _, name := range keys {
		sb.WriteString(fmt.Sprintf("export %s=%q\n", name, exports[name]))
	}
	sb.WriteString(blockEnd + "\n")
	return sb.String()
}
