// Package strutil holds string/slice/map helpers shared across Hatch modules.
package strutil

import (
	"regexp"
	"sort"
	"strings"
)

var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// OrDefault returns v unless it is empty.
func OrDefault(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

// ValidID reports whether id (trimmed) is a safe registry id: [A-Za-z0-9_-]+.
func ValidID(id string) bool {
	return idPattern.MatchString(strings.TrimSpace(id))
}

// SanitizeID turns a package name, repo name or display name into a lowercase
// [a-z0-9_-] id, dropping "@", "-mcp" and "-server" decorations.
func SanitizeID(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "@")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.TrimSuffix(s, "-mcp")
	s = strings.TrimSuffix(s, "-server")
	s = strings.Trim(s, "-_ ")
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-'
		if !ok {
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
			continue
		}
		b.WriteRune(r)
		lastDash = r == '-'
	}
	return strings.Trim(b.String(), "-_")
}

// CleanSlice trims every element and drops empty ones.
func CleanSlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

// CloneMap returns a shallow copy of a string map (nil stays nil).
func CloneMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// SortedKeys returns the non-blank keys of a string map in sorted order.
func SortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for k := range values {
		if strings.TrimSpace(k) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// EnsureTrailingNewline appends "\n" to non-empty content that lacks one.
func EnsureTrailingNewline(content string) string {
	if content == "" || strings.HasSuffix(content, "\n") {
		return content
	}
	return content + "\n"
}
