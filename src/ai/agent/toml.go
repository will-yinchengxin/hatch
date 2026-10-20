package agent

import (
	"strconv"
	"strings"

	"hatch/core/strutil"
)

// Line-based TOML editing for ~/.codex/config.toml. A full TOML round-trip
// would drop the user's comments and ordering, so Hatch only rewrites the
// specific top-level keys / section keys it owns.

// ReadTOMLStringKey reads a quoted top-level (pre-section) string key.
func ReadTOMLStringKey(content, key string) string {
	preamble, _ := splitTOMLPreamble(content)
	for _, line := range strings.Split(preamble, "\n") {
		if !isTOMLKeyLine(line, key) {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return ""
		}
		quoted := quotedTOMLStringPrefix(strings.TrimSpace(parts[1]))
		if quoted == "" {
			return ""
		}
		value, err := strconv.Unquote(quoted)
		if err != nil {
			return ""
		}
		return value
	}
	return ""
}

// ReadTOMLSectionBoolKey reads a bool key inside [section].
func ReadTOMLSectionBoolKey(content, section, key string) bool {
	inSection := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if isTOMLSectionLine(line, section) {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(trimmed, "[") {
			return false
		}
		if !inSection || !isTOMLKeyLine(line, key) {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return false
		}
		fields := strings.Fields(strings.TrimSpace(parts[1]))
		if len(fields) == 0 {
			return false
		}
		value, err := strconv.ParseBool(fields[0])
		if err != nil {
			return false
		}
		return value
	}
	return false
}

// SetTOMLStringKey sets (or with value=="" removes) a top-level string key,
// writing it at the top of the file.
func SetTOMLStringKey(content, key, value string) string {
	preamble, rest := splitTOMLPreamble(content)
	var kept []string
	for _, line := range strings.SplitAfter(preamble, "\n") {
		if line == "" || isTOMLKeyLine(line, key) {
			continue
		}
		kept = append(kept, line)
	}
	content = strings.TrimLeft(strings.Join(kept, "")+rest, "\n")
	if value == "" {
		return strutil.EnsureTrailingNewline(content)
	}
	line := key + " = " + strconv.Quote(value) + "\n"
	if content == "" {
		return line
	}
	return line + "\n" + strutil.EnsureTrailingNewline(content)
}

// SetTOMLSectionBoolKey sets a bool key inside [section], creating the section
// at the end of the file when missing.
func SetTOMLSectionBoolKey(content, section, key string, value bool) string {
	keyLine := key + " = " + strconv.FormatBool(value) + "\n"
	lines := strings.SplitAfter(content, "\n")
	var out []string
	inSection := false
	foundSection := false
	wroteKey := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			if inSection && !wroteKey {
				out = append(out, keyLine)
				wroteKey = true
			}
			inSection = false
			if isTOMLSectionLine(line, section) {
				inSection = true
				foundSection = true
			}
		}
		if inSection && isTOMLKeyLine(line, key) {
			if !wroteKey {
				out = append(out, keyLine)
				wroteKey = true
			}
			continue
		}
		out = append(out, line)
	}

	if inSection && !wroteKey {
		out = append(out, keyLine)
	}

	updated := strings.Join(out, "")
	if foundSection {
		return strutil.EnsureTrailingNewline(updated)
	}

	updated = strutil.EnsureTrailingNewline(updated)
	if strings.TrimSpace(updated) != "" {
		updated += "\n"
	}
	updated += "[" + section + "]\n" + keyLine
	return strutil.EnsureTrailingNewline(updated)
}

func splitTOMLPreamble(content string) (string, string) {
	offset := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			return content[:offset], content[offset:]
		}
		offset += len(line)
	}
	return content, ""
}

func isTOMLKeyLine(line, key string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, key) {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(trimmed, key))
	return strings.HasPrefix(rest, "=")
}

func isTOMLSectionLine(line, section string) bool {
	return strings.TrimSpace(line) == "["+section+"]"
}

func quotedTOMLStringPrefix(s string) string {
	if !strings.HasPrefix(s, `"`) {
		return ""
	}
	escaped := false
	for i := 1; i < len(s); i++ {
		switch {
		case escaped:
			escaped = false
		case s[i] == '\\':
			escaped = true
		case s[i] == '"':
			return s[:i+1]
		}
	}
	return ""
}
