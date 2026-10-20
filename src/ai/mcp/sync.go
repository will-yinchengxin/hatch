package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"hatch/ai/agent"
	"hatch/core/config"
	"hatch/core/fsutil"
	"hatch/core/strutil"
)

// ClaudeEntryExists reports whether ~/.claude.json already defines an MCP
// server under the given id. Used to avoid clobbering hand-written entries.
func ClaudeEntryExists(id string) (bool, error) {
	path, err := agent.ClaudeJSONPath()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return false, nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return false, fmt.Errorf("invalid ~/.claude.json: %w", err)
	}
	servers, _ := raw["mcpServers"].(map[string]interface{})
	_, ok := servers[id]
	return ok, nil
}

// CodexEntryExists reports whether ~/.codex/config.toml already defines a
// [mcp_servers.<id>] section. Missing file/dir is treated as "not present".
func CodexEntryExists(id string) bool {
	path, err := agent.CodexConfigPath()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
			continue
		}
		if strings.Trim(trimmed, "[]") == "mcp_servers."+id {
			return true
		}
	}
	return false
}

// WriteClaude adds (enabled) or removes server under "mcpServers" in ~/.claude.json.
func WriteClaude(server config.MCPServer, enabled bool) error {
	path, err := agent.ClaudeJSONPath()
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("invalid ~/.claude.json: %w", err)
		}
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}
	servers, _ := raw["mcpServers"].(map[string]interface{})
	if servers == nil {
		servers = map[string]interface{}{}
	}
	if enabled {
		servers[server.ID] = renderClaude(server)
	} else {
		delete(servers, server.ID)
	}
	raw["mcpServers"] = servers
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteWithBackup(path, append(out, '\n'), 0o600)
}

func renderClaude(server config.MCPServer) map[string]interface{} {
	out := map[string]interface{}{}
	if server.Transport == TransportHTTP {
		out["type"] = "http"
		out["url"] = server.URL
		if len(server.HTTPHeaders) > 0 {
			out["headers"] = server.HTTPHeaders
		}
		return out
	}
	out["type"] = "stdio"
	out["command"] = server.Command
	out["args"] = server.Args
	if len(server.Env) > 0 {
		out["env"] = server.Env
	} else {
		out["env"] = map[string]string{}
	}
	if server.Cwd != "" {
		out["cwd"] = server.Cwd
	}
	return out
}

// WriteCodex replaces the Hatch-managed [mcp_servers.<id>] block in
// ~/.codex/config.toml, removing it when enabled is false.
func WriteCodex(server config.MCPServer, enabled bool) error {
	path, err := agent.CodexConfigPath()
	if err != nil {
		return err
	}
	content := ""
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}
	content = removeCodexServer(content, server.ID)
	if enabled {
		content = strutil.EnsureTrailingNewline(content)
		if strings.TrimSpace(content) != "" {
			content += "\n"
		}
		content += renderCodex(server)
	}
	return fsutil.WriteWithBackup(path, []byte(strutil.EnsureTrailingNewline(content)), 0o644)
}

func renderCodex(server config.MCPServer) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Managed by Hatch MCP\n[mcp_servers.%s]\n", server.ID)
	if server.Transport == TransportHTTP {
		fmt.Fprintf(&b, "url = %s\n", strconv.Quote(server.URL))
		if server.BearerTokenEnvVar != "" {
			fmt.Fprintf(&b, "bearer_token_env_var = %s\n", strconv.Quote(server.BearerTokenEnvVar))
		}
		if len(server.HTTPHeaders) > 0 {
			fmt.Fprintf(&b, "http_headers = %s\n", tomlInlineMap(server.HTTPHeaders))
		}
	} else {
		fmt.Fprintf(&b, "command = %s\n", strconv.Quote(server.Command))
		if len(server.Args) > 0 {
			fmt.Fprintf(&b, "args = %s\n", tomlStringArray(server.Args))
		}
		if server.Cwd != "" {
			fmt.Fprintf(&b, "cwd = %s\n", strconv.Quote(server.Cwd))
		}
	}
	if server.StartupTimeoutSec > 0 {
		fmt.Fprintf(&b, "startup_timeout_sec = %d\n", server.StartupTimeoutSec)
	}
	if server.ToolTimeoutSec > 0 {
		fmt.Fprintf(&b, "tool_timeout_sec = %d\n", server.ToolTimeoutSec)
	}
	b.WriteString("enabled = true\n")
	if len(server.Env) > 0 && server.Transport == TransportStdio {
		fmt.Fprintf(&b, "\n[mcp_servers.%s.env]\n", server.ID)
		keys := strutil.SortedKeys(server.Env)
		for _, k := range keys {
			fmt.Fprintf(&b, "%s = %s\n", k, strconv.Quote(server.Env[k]))
		}
	}
	return b.String()
}

func removeCodexServer(content, id string) string {
	var kept []string
	inTarget := false
	for _, line := range strings.SplitAfter(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			name := strings.Trim(trimmed, "[]")
			inTarget = name == "mcp_servers."+id || strings.HasPrefix(name, "mcp_servers."+id+".")
		}
		if inTarget {
			continue
		}
		if strings.TrimSpace(line) == "# Managed by Hatch MCP" {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimRight(strings.Join(kept, ""), "\n") + "\n"
}

func tomlStringArray(values []string) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, strconv.Quote(v))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func tomlInlineMap(values map[string]string) string {
	keys := strutil.SortedKeys(values)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, strconv.Quote(k)+" = "+strconv.Quote(values[k]))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}
