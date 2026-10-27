package mcp

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml"

	"hatch/ai/agent"
	"hatch/core/fsutil"
)

// Codex config is edited block by block: only the [mcp_servers.<name>] table
// (and its sub-tables) being changed is rewritten, so comments, ordering and
// every other server stay byte-for-byte intact.

const legacyCodexMarker = "# Managed by Hatch MCP"

// CodexConfigFile returns the config.toml for a Codex location.
func CodexConfigFile(loc Location) (string, error) {
	if loc.Scope == ScopeProject {
		if loc.Project == "" {
			return "", fmt.Errorf("Codex 项目作用域缺少项目路径")
		}
		return filepath.Join(loc.Project, ".codex", "config.toml"), nil
	}
	return agent.CodexConfigPath()
}

// readCodexServers parses [mcp_servers.*] tables from a config.toml.
func readCodexServers(path string) (map[string]map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tree, err := toml.LoadBytes(data)
	if err != nil {
		return nil, fmt.Errorf("config.toml 语法无效: %w", err)
	}
	out := map[string]map[string]any{}
	servers, ok := tree.Get("mcp_servers").(*toml.Tree)
	if !ok {
		return out, nil
	}
	for _, name := range servers.Keys() {
		if t, ok := servers.Get(name).(*toml.Tree); ok {
			out[name] = t.ToMap()
		}
	}
	return out, nil
}

// codexTrustedProjects lists projects marked trust_level = "trusted" in the
// user config; only those may carry a project-level .codex/config.toml.
func codexTrustedProjects(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	tree, err := toml.LoadBytes(data)
	if err != nil {
		return nil
	}
	projects, ok := tree.Get("projects").(*toml.Tree)
	if !ok {
		return nil
	}
	var out []string
	for _, p := range projects.Keys() {
		if t, ok := projects.Get(p).(*toml.Tree); ok && t.Get("trust_level") == "trusted" {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// tomlHeaderPath parses "[a.b.\"c d\"]" into its key path. Array-of-table
// headers ("[[x]]") also count as headers so they end a block.
func tomlHeaderPath(line string) ([]string, bool) {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "[") {
		return nil, false
	}
	if i := strings.Index(t, "#"); i > 0 && !strings.Contains(t[:i], `"`) && !strings.Contains(t[:i], "'") {
		t = strings.TrimSpace(t[:i])
	}
	if !strings.HasSuffix(t, "]") {
		return nil, false
	}
	t = strings.TrimSpace(strings.Trim(t, "[]"))
	var parts []string
	var cur strings.Builder
	quote := byte(0)
	for i := 0; i < len(t); i++ {
		c := t[i]
		switch {
		case quote != 0:
			if c == '\\' && quote == '"' && i+1 < len(t) {
				i++
				cur.WriteByte(t[i])
			} else if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '.':
			parts = append(parts, strings.TrimSpace(cur.String()))
			cur.Reset()
		case c == ' ' || c == '\t':
		default:
			cur.WriteByte(c)
		}
	}
	parts = append(parts, strings.TrimSpace(cur.String()))
	return parts, true
}

func isServerPath(path []string, name string) bool {
	return len(path) >= 2 && path[0] == "mcp_servers" && path[1] == name
}

// splitCodexServer removes the server's tables from content and returns the
// remaining lines plus the index where the main table used to start (-1 when
// absent). A legacy "# Managed by Hatch MCP" line right above is dropped too.
func splitCodexServer(content, name string) ([]string, int) {
	lines := strings.Split(content, "\n")
	var kept []string
	insertAt := -1
	inTarget := false
	for _, line := range lines {
		if path, ok := tomlHeaderPath(line); ok {
			inTarget = isServerPath(path, name)
			if inTarget && insertAt < 0 {
				if n := len(kept); n > 0 && strings.TrimSpace(kept[n-1]) == legacyCodexMarker {
					kept = kept[:n-1]
				}
				insertAt = len(kept)
			}
		}
		if !inTarget {
			kept = append(kept, line)
		}
	}
	return kept, insertAt
}

// CodexUpsert replaces (or appends) the server's tables with table.
func CodexUpsert(content, name string, table map[string]any) (string, error) {
	kept, at := splitCodexServer(content, name)
	block := strings.Split(strings.TrimRight(renderCodexTable(name, table), "\n"), "\n")
	var lines []string
	if at < 0 {
		lines = trimTrailingBlank(kept)
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, block...)
	} else {
		lines = append(lines, kept[:at]...)
		lines = append(lines, block...)
		rest := kept[at:]
		if len(rest) > 0 && strings.TrimSpace(rest[0]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, rest...)
	}
	out := strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
	if _, err := toml.Load(out); err != nil {
		return "", fmt.Errorf("生成的 config.toml 无效（可能存在以点号键或内联表定义的同名 MCP）: %w", err)
	}
	return out, nil
}

// CodexRemove deletes the server's tables.
func CodexRemove(content, name string) string {
	kept, at := splitCodexServer(content, name)
	if at >= 0 && at < len(kept) && at > 0 && strings.TrimSpace(kept[at-1]) == "" && strings.TrimSpace(kept[at]) == "" {
		kept = append(kept[:at], kept[at+1:]...)
	}
	return strings.TrimRight(strings.Join(kept, "\n"), "\n") + "\n"
}

// CodexSetEnabled flips only the enabled key of the main table, leaving every
// other line untouched (enabled defaults to true, so enabling removes it).
func CodexSetEnabled(content, name string, enabled bool) (string, error) {
	lines := strings.Split(content, "\n")
	header := -1
	end := len(lines)
	for i, line := range lines {
		path, ok := tomlHeaderPath(line)
		if !ok {
			continue
		}
		if header >= 0 {
			end = i
			break
		}
		if len(path) == 2 && isServerPath(path, name) {
			header = i
		}
	}
	if header < 0 {
		return "", fmt.Errorf("Codex 配置中没有 [mcp_servers.%s]", name)
	}
	var out []string
	out = append(out, lines[:header+1]...)
	if !enabled {
		out = append(out, "enabled = false")
	}
	for _, line := range lines[header+1 : end] {
		if k, _, ok := strings.Cut(strings.TrimSpace(line), "="); ok && strings.TrimSpace(k) == "enabled" {
			continue
		}
		out = append(out, line)
	}
	out = append(out, lines[end:]...)
	result := strings.Join(out, "\n")
	if _, err := toml.Load(result); err != nil {
		return "", err
	}
	return result, nil
}

// writeCodexFile validates and writes content with a backup.
func writeCodexFile(path, content string) error {
	if _, err := toml.Load(content); err != nil {
		return fmt.Errorf("拒绝写入无效的 TOML: %w", err)
	}
	return fsutil.WriteWithBackup(path, []byte(content), 0o644)
}

func trimTrailingBlank(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// ── rendering ───────────────────────────────────────────────────────────────

var codexKeyOrder = []string{"command", "args", "cwd", "url", "bearer_token_env_var", "http_headers", "env_vars", "env_http_headers", "startup_timeout_sec", "tool_timeout_sec", "enabled", "required"}

func renderCodexTable(name string, table map[string]any) string {
	var b strings.Builder
	writeTOMLTable(&b, []string{"mcp_servers", name}, table)
	return b.String()
}

// RenderCodexBlock renders a full [mcp_servers.<name>] block (for previews).
func RenderCodexBlock(name string, table map[string]any) string { return renderCodexTable(name, table) }

func writeTOMLTable(b *strings.Builder, path []string, table map[string]any) {
	quoted := make([]string, len(path))
	for i, p := range path {
		quoted[i] = tomlKey(p)
	}
	fmt.Fprintf(b, "[%s]\n", strings.Join(quoted, "."))
	var scalars, tables []string
	for k, v := range table {
		if _, ok := v.(map[string]any); ok {
			tables = append(tables, k)
		} else {
			scalars = append(scalars, k)
		}
	}
	sortByOrder(scalars, codexKeyOrder)
	sortByOrder(tables, []string{"env"})
	for _, k := range scalars {
		fmt.Fprintf(b, "%s = %s\n", tomlKey(k), tomlValue(table[k]))
	}
	for _, k := range tables {
		sub := table[k].(map[string]any)
		if len(sub) == 0 {
			continue
		}
		b.WriteString("\n")
		writeTOMLTable(b, append(append([]string(nil), path...), k), sub)
	}
}

func sortByOrder(keys []string, order []string) {
	rank := map[string]int{}
	for i, k := range order {
		rank[k] = i + 1
	}
	sort.SliceStable(keys, func(i, j int) bool {
		ri, rj := rank[keys[i]], rank[keys[j]]
		if ri != 0 || rj != 0 {
			if ri == 0 {
				return false
			}
			if rj == 0 {
				return true
			}
			return ri < rj
		}
		return keys[i] < keys[j]
	})
}

func tomlKey(k string) string {
	if k == "" {
		return `""`
	}
	for _, r := range k {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return tomlQuote(k)
		}
	}
	return k
}

func tomlQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&b, `\u%04X`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

func tomlValue(v any) string {
	switch x := v.(type) {
	case string:
		return tomlQuote(x)
	case bool:
		return strconv.FormatBool(x)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		if x == math.Trunc(x) && math.Abs(x) < 1e15 {
			return strconv.FormatFloat(x, 'f', 1, 64)
		}
		return strconv.FormatFloat(x, 'g', -1, 64)
	case time.Time:
		return x.Format(time.RFC3339)
	case []string:
		parts := make([]string, len(x))
		for i, s := range x {
			parts[i] = tomlQuote(s)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case []any:
		parts := make([]string, len(x))
		for i, s := range x {
			parts[i] = tomlValue(s)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]any:
		keys := sortedKeys(x)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = tomlKey(k) + " = " + tomlValue(x[k])
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	case map[string]string:
		return tomlValue(toAnyMap(x))
	}
	return tomlQuote(fmt.Sprint(v))
}
