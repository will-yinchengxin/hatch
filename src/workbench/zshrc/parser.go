package zshrc

import (
	"fmt"
	"sort"
	"strings"
)

// Entry represents a parsed element from a zshrc file.
type Entry struct {
	Type    string `json:"type"`  // function | alias | export | path | source | section | other
	Key     string `json:"key"`   // unique identifier
	Value   string `json:"value"` // meaningful content (function body, alias value, etc.)
	Note    string `json:"note"`  // user annotation
	LineNum int    `json:"lineNum"`
}

// ParseContent parses raw zshrc text into structured entries.
// annotations maps entry keys to user-supplied notes.
func ParseContent(content string, annotations map[string]string) []Entry {
	lines := strings.Split(content, "\n")
	var entries []Entry
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		lineNum := i + 1
		i++

		if trimmed == "" {
			continue
		}

		// Skip Hatch managed block entirely.
		if strings.Contains(trimmed, "Hatch managed block") {
			for i < len(lines) {
				l := strings.TrimSpace(lines[i])
				i++
				if strings.Contains(l, "Hatch managed block") {
					break
				}
			}
			continue
		}

		// Comments / section headers
		if strings.HasPrefix(trimmed, "#") {
			text := strings.TrimSpace(trimmed[1:])
			if text == "" {
				continue
			}
			key := fmt.Sprintf("comment:%d", lineNum)
			entries = append(entries, Entry{
				Type: "section", Key: key, Value: text,
				Note: ann(annotations, key), LineNum: lineNum,
			})
			continue
		}

		// alias name=value
		if strings.HasPrefix(trimmed, "alias ") {
			rest := strings.TrimPrefix(trimmed, "alias ")
			if eq := strings.Index(rest, "="); eq > 0 {
				name := strings.TrimSpace(rest[:eq])
				val := unquote(rest[eq+1:])
				entries = append(entries, Entry{
					Type: "alias", Key: name, Value: val,
					Note: ann(annotations, name), LineNum: lineNum,
				})
				continue
			}
		}

		// export VAR=value  or  export PATH=…
		if strings.HasPrefix(trimmed, "export ") {
			rest := strings.TrimPrefix(trimmed, "export ")
			if eq := strings.Index(rest, "="); eq > 0 {
				name := strings.TrimSpace(rest[:eq])
				val := rest[eq+1:]
				t := "export"
				if name == "PATH" {
					t = "path"
				}
				entries = append(entries, Entry{
					Type: t, Key: name, Value: val,
					Note: ann(annotations, name), LineNum: lineNum,
				})
				continue
			}
		}

		// Bare PATH=…
		if strings.HasPrefix(trimmed, "PATH=") {
			entries = append(entries, Entry{
				Type: "path", Key: "PATH",
				Value:   strings.TrimPrefix(trimmed, "PATH="),
				Note:    ann(annotations, "PATH"),
				LineNum: lineNum,
			})
			continue
		}

		// source /path or . /path
		if strings.HasPrefix(trimmed, "source ") {
			key := fmt.Sprintf("source:%d", lineNum)
			entries = append(entries, Entry{
				Type: "source", Key: key,
				Value:   strings.TrimPrefix(trimmed, "source "),
				Note:    ann(annotations, key),
				LineNum: lineNum,
			})
			continue
		}
		if len(trimmed) > 2 && trimmed[0] == '.' && trimmed[1] == ' ' {
			key := fmt.Sprintf("source:%d", lineNum)
			entries = append(entries, Entry{
				Type: "source", Key: key,
				Value:   trimmed[2:],
				Note:    ann(annotations, key),
				LineNum: lineNum,
			})
			continue
		}

		// Function definitions: function name { or name() {
		if funcName := detectFuncName(trimmed); funcName != "" {
			var bodyLines []string
			bodyLines = append(bodyLines, line)
			depth := strings.Count(line, "{") - strings.Count(line, "}")
			for depth > 0 && i < len(lines) {
				bodyLines = append(bodyLines, lines[i])
				depth += strings.Count(lines[i], "{") - strings.Count(lines[i], "}")
				i++
			}
			entries = append(entries, Entry{
				Type: "function", Key: funcName,
				Value:   strings.Join(bodyLines, "\n"),
				Note:    ann(annotations, funcName),
				LineNum: lineNum,
			})
			continue
		}

		// Everything else (nvm init, conda, eval, etc.)
		key := fmt.Sprintf("other:%d", lineNum)
		entries = append(entries, Entry{
			Type: "other", Key: key, Value: trimmed,
			Note: ann(annotations, key), LineNum: lineNum,
		})
	}

	return entries
}

func ann(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 &&
		((s[0] == '\'' && s[len(s)-1] == '\'') ||
			(s[0] == '"' && s[len(s)-1] == '"')) {
		return s[1 : len(s)-1]
	}
	return s
}

func detectFuncName(line string) string {
	// function name { or function name() {
	if strings.HasPrefix(line, "function ") {
		rest := strings.TrimPrefix(line, "function ")
		parts := strings.Fields(rest)
		if len(parts) > 0 {
			name := strings.TrimSuffix(parts[0], "()")
			name = strings.TrimSuffix(name, "(")
			if isValidFuncName(name) {
				return name
			}
		}
		return ""
	}
	// name() { or name ()
	parenIdx := strings.Index(line, "()")
	if parenIdx > 0 {
		candidate := strings.TrimSpace(line[:parenIdx])
		// Must not contain spaces or shell meta-chars to be a function name.
		if isValidFuncName(candidate) {
			return candidate
		}
	}
	return ""
}

func isValidFuncName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

// Group is a set of entries sharing the same Type.
type Group struct {
	Kind    string   `json:"kind"`    // function / alias / export / path / source / section / other
	Entries []string `json:"entries"` // entry keys
}

// GroupByType buckets entries by Type (sorted by type name) for quick review.
func GroupByType(entries []Entry) []Group {
	buckets := map[string][]string{}
	for _, e := range entries {
		kind := e.Type
		if kind == "" {
			kind = "other"
		}
		buckets[kind] = append(buckets[kind], e.Key)
	}
	kinds := make([]string, 0, len(buckets))
	for k := range buckets {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	out := make([]Group, 0, len(kinds))
	for _, k := range kinds {
		out = append(out, Group{Kind: k, Entries: buckets[k]})
	}
	return out
}
