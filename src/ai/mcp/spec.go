// Package mcp shows and edits the MCP servers that Claude Code and Codex
// actually load. The client config files are the only source of truth: Hatch
// scans them into one inventory, edits single entries in place (keeping keys
// it does not model), parses pasted snippets / registry metadata into drafts,
// and probes servers by speaking MCP itself. It never keeps servers running.
package mcp

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"hatch/core/config"
)

const (
	ClientClaude  = "claude"
	ClientCodex   = "codex"
	ClientDesktop = "desktop" // Claude Desktop: read-only import source

	ScopeUser      = "user"
	ScopeLocal     = "local"   // Claude: ~/.claude.json projects[path].mcpServers
	ScopeProject   = "project" // Claude: <path>/.mcp.json; Codex: <path>/.codex/config.toml
	ScopePlugin    = "plugin"
	ScopeConnector = "connector"

	TransportStdio = "stdio"
	TransportHTTP  = "http"
	TransportSSE   = "sse"
	TransportWS    = "ws"
)

// Location is one place an MCP server definition can live.
type Location struct {
	Client  string `json:"client"`
	Scope   string `json:"scope"`
	Project string `json:"project,omitempty"`
}

// Key is the stable string form used for maps and the frontend.
func (l Location) Key() string {
	k := l.Client + ":" + l.Scope
	if l.Project != "" {
		k += ":" + l.Project
	}
	return k
}

// ParseLocationKey is the inverse of Key.
func ParseLocationKey(key string) (Location, error) {
	parts := strings.SplitN(key, ":", 3)
	if len(parts) < 2 {
		return Location{}, fmt.Errorf("invalid MCP location: %s", key)
	}
	loc := Location{Client: parts[0], Scope: parts[1]}
	if len(parts) == 3 {
		loc.Project = parts[2]
	}
	return loc, nil
}

// Label is the short Chinese chip label.
func (l Location) Label() string {
	base := filepath.Base(l.Project)
	switch {
	case l.Client == ClientClaude && l.Scope == ScopeUser:
		return "Claude·全局"
	case l.Client == ClientClaude && l.Scope == ScopeLocal:
		return "Claude·" + base
	case l.Client == ClientClaude && l.Scope == ScopeProject:
		return "Claude·" + base + " (.mcp.json)"
	case l.Client == ClientClaude && l.Scope == ScopePlugin:
		return "Claude 插件"
	case l.Client == ClientClaude && l.Scope == ScopeConnector:
		return "claude.ai 连接器"
	case l.Client == ClientCodex && l.Scope == ScopeUser:
		return "Codex"
	case l.Client == ClientCodex && l.Scope == ScopeProject:
		return "Codex·" + base
	case l.Client == ClientCodex && l.Scope == ScopePlugin:
		return "Codex 插件"
	case l.Client == ClientDesktop:
		return "Claude Desktop"
	}
	return l.Key()
}

// Writable reports whether Hatch may add, edit, disable or remove entries here.
func (l Location) Writable() bool {
	switch l.Client {
	case ClientClaude:
		return l.Scope == ScopeUser || ((l.Scope == ScopeLocal || l.Scope == ScopeProject) && l.Project != "")
	case ClientCodex:
		return l.Scope == ScopeUser || (l.Scope == ScopeProject && l.Project != "")
	}
	return false
}

// Spec is the client-neutral server definition. Extra keeps every key Hatch
// does not model (Claude alwaysLoad, Codex enabled_tools / tools.*, ...) so
// rewriting an entry never drops the user's tuning.
type Spec struct {
	Transport         string            `json:"transport"`
	Command           string            `json:"command,omitempty"`
	Args              []string          `json:"args,omitempty"`
	Env               map[string]string `json:"env,omitempty"`
	Cwd               string            `json:"cwd,omitempty"`
	URL               string            `json:"url,omitempty"`
	Headers           map[string]string `json:"headers,omitempty"`
	BearerTokenEnvVar string            `json:"bearerTokenEnvVar,omitempty"`
	StartupTimeoutSec int               `json:"startupTimeoutSec,omitempty"`
	ToolTimeoutSec    int               `json:"toolTimeoutSec,omitempty"`
	Extra             map[string]any    `json:"extra,omitempty"`
}

// Entry is one server definition at one location.
type Entry struct {
	Name           string           `json:"name"`
	Location       Location         `json:"location"`
	LocationKey    string           `json:"locationKey"`
	Label          string           `json:"label"`
	File           string           `json:"file,omitempty"`
	Spec           Spec             `json:"spec"`
	Enabled        bool             `json:"enabled"`
	Parked         bool             `json:"parked,omitempty"` // disabled by Hatch; raw kept in MCPMeta
	Stale          bool             `json:"stale,omitempty"`  // project directory no longer exists
	ReadOnly       bool             `json:"readOnly,omitempty"`
	ReadOnlyReason string           `json:"readOnlyReason,omitempty"`
	CLIStatus      string           `json:"cliStatus,omitempty"` // status text reported by `claude mcp list`
	Raw            string           `json:"raw"`
	Probe          *config.MCPProbe `json:"probe,omitempty"`
}

// ServerRow groups every entry that shares a name.
type ServerRow struct {
	Name    string          `json:"name"`
	Entries []Entry         `json:"entries"`
	Drift   bool            `json:"drift"`
	Meta    *config.MCPMeta `json:"meta,omitempty"`
}

// Normalize trims fields, drops empty maps and fills the transport.
func (s Spec) Normalize() Spec {
	s.Command = strings.TrimSpace(s.Command)
	s.URL = strings.TrimSpace(s.URL)
	s.Cwd = strings.TrimSpace(s.Cwd)
	s.BearerTokenEnvVar = strings.TrimSpace(s.BearerTokenEnvVar)
	if s.Transport == "" {
		if s.URL != "" {
			s.Transport = TransportHTTP
		} else {
			s.Transport = TransportStdio
		}
	}
	s.Env = cleanMap(s.Env)
	s.Headers = cleanMap(s.Headers)
	if len(s.Extra) == 0 {
		s.Extra = nil
	}
	return s
}

// Validate checks the minimum a client needs to start the server.
func (s Spec) Validate() error {
	switch s.Transport {
	case TransportStdio:
		if s.Command == "" {
			return fmt.Errorf("stdio MCP 需要填写启动命令")
		}
	case TransportHTTP, TransportSSE, TransportWS:
		if s.URL == "" {
			return fmt.Errorf("远程 MCP 需要填写 URL")
		}
	default:
		return fmt.Errorf("未知传输类型: %s", s.Transport)
	}
	return nil
}

// Fingerprint identifies the launch-relevant part of a spec for drift checks.
func (s Spec) Fingerprint() string {
	s = s.Normalize()
	var b strings.Builder
	b.WriteString(s.Transport + "|" + s.Command + "|" + strings.Join(s.Args, "\x00") + "|" + s.URL + "|" + s.Cwd)
	for _, m := range []map[string]string{s.Env, s.Headers} {
		b.WriteString("|")
		for _, k := range sortedKeys(m) {
			b.WriteString(k + "=" + m[k] + ";")
		}
	}
	return b.String()
}

// ValidName reports whether name is usable as a server name in both clients.
func ValidName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.') {
			return false
		}
	}
	return true
}

// SanitizeName turns a package / repo / host name into a server name.
func SanitizeName(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndex(s, "@"); i > 0 {
		s = s[:i] // drop @version
	}
	if i := strings.LastIndex(s, "/"); i >= 0 {
		pkg := s[i+1:]
		// "@playwright/mcp" → "playwright": a generic package part is useless.
		if generic := map[string]bool{"mcp": true, "server": true, "mcp-server": true}; generic[strings.ToLower(pkg)] {
			pkg = strings.TrimPrefix(s[:i], "@")
			if j := strings.LastIndex(pkg, "/"); j >= 0 {
				pkg = pkg[j+1:]
			}
		}
		s = pkg
	}
	s = strings.TrimPrefix(s, "@")
	for _, p := range []string{"mcp-server-", "server-"} {
		s = strings.TrimPrefix(s, p)
	}
	for _, p := range []string{"-mcp-server", "-mcp", "_mcp"} {
		s = strings.TrimSuffix(s, p)
	}
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ' || r == '.':
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-_")
	if out == "" {
		return "mcp"
	}
	return out
}

func cleanMap(m map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range m {
		if k = strings.TrimSpace(k); k != "" {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func stringMap(v any) map[string]string {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]string{}
	for k, val := range m {
		out[k] = fmt.Sprint(val)
	}
	return out
}

func stringSlice(v any) []string {
	switch arr := v.(type) {
	case []any:
		out := make([]string, 0, len(arr))
		for _, x := range arr {
			out = append(out, fmt.Sprint(x))
		}
		return out
	case []string:
		return append([]string(nil), arr...)
	}
	return nil
}

func intValue(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// ── Claude JSON form ────────────────────────────────────────────────────────

var claudeKnownKeys = map[string]bool{"type": true, "command": true, "args": true, "env": true, "cwd": true, "url": true, "headers": true}

// SpecFromClaude converts one Claude mcpServers entry.
func SpecFromClaude(m map[string]any) Spec {
	s := Spec{
		Transport: str(m["type"]),
		Command:   str(m["command"]),
		Args:      stringSlice(m["args"]),
		Env:       stringMap(m["env"]),
		Cwd:       str(m["cwd"]),
		URL:       str(m["url"]),
		Headers:   stringMap(m["headers"]),
	}
	if s.Transport == "streamable-http" || s.Transport == "streamable_http" {
		s.Transport = TransportHTTP
	}
	for k, v := range m {
		if !claudeKnownKeys[k] {
			if s.Extra == nil {
				s.Extra = map[string]any{}
			}
			s.Extra[k] = v
		}
	}
	return s.Normalize()
}

// ClaudeJSON renders s as a Claude mcpServers entry. Extra keys are kept only
// when keepExtra is set (same client); they never override modelled fields.
func ClaudeJSON(s Spec, keepExtra bool) map[string]any {
	s = s.Normalize()
	out := map[string]any{}
	if keepExtra {
		for k, v := range s.Extra {
			out[k] = v
		}
	}
	if s.Transport == TransportStdio {
		out["type"] = "stdio"
		out["command"] = s.Command
		out["args"] = nonNilSlice(s.Args)
		env := map[string]string{}
		for k, v := range s.Env {
			env[k] = v
		}
		out["env"] = env
		if s.Cwd != "" {
			out["cwd"] = s.Cwd
		}
		return out
	}
	out["type"] = s.Transport
	out["url"] = s.URL
	headers := map[string]string{}
	for k, v := range s.Headers {
		headers[k] = v
	}
	if s.BearerTokenEnvVar != "" {
		if _, ok := headers["Authorization"]; !ok {
			headers["Authorization"] = "Bearer ${" + s.BearerTokenEnvVar + "}"
		}
	}
	if len(headers) > 0 {
		out["headers"] = headers
	}
	return out
}

func nonNilSlice(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

// ── Codex TOML form ─────────────────────────────────────────────────────────

var codexKnownKeys = map[string]bool{
	"command": true, "args": true, "env": true, "cwd": true, "url": true, "http_headers": true,
	"bearer_token_env_var": true, "startup_timeout_sec": true, "tool_timeout_sec": true, "enabled": true,
}

// SpecFromCodex converts one [mcp_servers.<name>] table and reports enabled.
func SpecFromCodex(m map[string]any) (Spec, bool) {
	s := Spec{
		Command:           str(m["command"]),
		Args:              stringSlice(m["args"]),
		Env:               stringMap(m["env"]),
		Cwd:               str(m["cwd"]),
		URL:               str(m["url"]),
		Headers:           stringMap(m["http_headers"]),
		BearerTokenEnvVar: str(m["bearer_token_env_var"]),
		StartupTimeoutSec: intValue(m["startup_timeout_sec"]),
		ToolTimeoutSec:    intValue(m["tool_timeout_sec"]),
	}
	if s.URL != "" {
		s.Transport = TransportHTTP
	} else {
		s.Transport = TransportStdio
	}
	for k, v := range m {
		if !codexKnownKeys[k] {
			if s.Extra == nil {
				s.Extra = map[string]any{}
			}
			s.Extra[k] = v
		}
	}
	enabled := true
	if b, ok := m["enabled"].(bool); ok {
		enabled = b
	}
	return s.Normalize(), enabled
}

// CodexTable renders s as a Codex server table. Codex has no SSE/WebSocket
// client, so those transports are rejected.
func CodexTable(s Spec, enabled, keepExtra bool) (map[string]any, error) {
	s = s.Normalize()
	out := map[string]any{}
	if keepExtra {
		for k, v := range s.Extra {
			out[k] = v
		}
	}
	switch s.Transport {
	case TransportStdio:
		out["command"] = s.Command
		if len(s.Args) > 0 {
			out["args"] = s.Args
		}
		if s.Cwd != "" {
			out["cwd"] = s.Cwd
		}
		if len(s.Env) > 0 {
			out["env"] = toAnyMap(s.Env)
		}
	case TransportHTTP:
		out["url"] = s.URL
		headers := strings.TrimSpace(s.Headers["Authorization"])
		if s.BearerTokenEnvVar != "" {
			out["bearer_token_env_var"] = s.BearerTokenEnvVar
		} else if v, ok := envRefBearer(headers); ok {
			out["bearer_token_env_var"] = v
		}
		rest := map[string]string{}
		for k, v := range s.Headers {
			if k == "Authorization" && out["bearer_token_env_var"] != nil {
				continue
			}
			rest[k] = v
		}
		if len(rest) > 0 {
			out["http_headers"] = toAnyMap(rest)
		}
	default:
		return nil, fmt.Errorf("Codex 不支持 %s 传输，仅支持 stdio 和 streamable HTTP", s.Transport)
	}
	if s.StartupTimeoutSec > 0 {
		out["startup_timeout_sec"] = int64(s.StartupTimeoutSec)
	}
	if s.ToolTimeoutSec > 0 {
		out["tool_timeout_sec"] = int64(s.ToolTimeoutSec)
	}
	if enabled {
		delete(out, "enabled")
	} else {
		out["enabled"] = false
	}
	return out, nil
}

// envRefBearer turns "Bearer ${VAR}" into VAR so Claude-style env references
// map onto Codex's bearer_token_env_var.
func envRefBearer(v string) (string, bool) {
	v = strings.TrimSpace(strings.TrimPrefix(v, "Bearer "))
	if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") && !strings.Contains(v, ":-") {
		return v[2 : len(v)-1], true
	}
	return "", false
}

func toAnyMap(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
