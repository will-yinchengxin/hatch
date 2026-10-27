package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"hatch/ai/agent"
	"hatch/core/fsutil"
	"hatch/core/macos"
)

// Claude Code keeps user and local (per-project) servers in ~/.claude.json, a
// large file it rewrites constantly while sessions run. Hatch therefore reads
// it directly but writes through `claude mcp add-json/remove --scope`, which
// is the client's own write path; a careful single-key patch is only the
// fallback when the CLI is missing.

type claudeProject struct {
	Servers  map[string]map[string]any
	Disabled map[string]bool
}

type claudeState struct {
	User     map[string]map[string]any
	Projects map[string]claudeProject
}

func readClaudeState() (claudeState, string, error) {
	path, err := agent.ClaudeJSONPath()
	if err != nil {
		return claudeState{}, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return claudeState{}, path, err
	}
	var raw struct {
		MCPServers map[string]map[string]any `json:"mcpServers"`
		Projects   map[string]struct {
			MCPServers         map[string]map[string]any `json:"mcpServers"`
			DisabledMcpServers []string                  `json:"disabledMcpServers"`
		} `json:"projects"`
	}
	if len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return claudeState{}, path, fmt.Errorf("~/.claude.json 不是有效 JSON: %w", err)
		}
	}
	st := claudeState{User: raw.MCPServers, Projects: map[string]claudeProject{}}
	for p, proj := range raw.Projects {
		cp := claudeProject{Servers: proj.MCPServers, Disabled: map[string]bool{}}
		for _, n := range proj.DisabledMcpServers {
			cp.Disabled[n] = true
		}
		st.Projects[p] = cp
	}
	return st, path, nil
}

func mcpJSONPath(project string) string { return filepath.Join(project, ".mcp.json") }

func readMCPJSON(project string) (map[string]map[string]any, error) {
	data, err := os.ReadFile(mcpJSONPath(project))
	if err != nil {
		return nil, err
	}
	var raw struct {
		MCPServers map[string]map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf(".mcp.json 不是有效 JSON: %w", err)
	}
	return raw.MCPServers, nil
}

// ClaudeDesktopConfigPath is Claude Desktop's config (import source only).
func ClaudeDesktopConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
}

func readDesktopServers() (map[string]map[string]any, error) {
	data, err := os.ReadFile(ClaudeDesktopConfigPath())
	if err != nil {
		return nil, err
	}
	var raw struct {
		MCPServers map[string]map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw.MCPServers, nil
}

// ClaudeFile returns the file a Claude location lives in.
func ClaudeFile(loc Location) string {
	if loc.Scope == ScopeProject {
		return mcpJSONPath(loc.Project)
	}
	p, _ := agent.ClaudeJSONPath()
	return p
}

// claudeEntry returns the raw entry at loc, if any.
func claudeEntry(loc Location, name string) (map[string]any, bool) {
	switch loc.Scope {
	case ScopeProject:
		servers, err := readMCPJSON(loc.Project)
		if err != nil {
			return nil, false
		}
		m, ok := servers[name]
		return m, ok
	case ScopeUser, ScopeLocal:
		st, _, err := readClaudeState()
		if err != nil {
			return nil, false
		}
		if loc.Scope == ScopeUser {
			m, ok := st.User[name]
			return m, ok
		}
		m, ok := st.Projects[loc.Project].Servers[name]
		return m, ok
	}
	return nil, false
}

// ClaudeWrite adds or replaces name at loc with body.
func ClaudeWrite(loc Location, name string, body map[string]any) error {
	old, exists := claudeEntry(loc, name)
	if useClaudePatch(loc) {
		return claudePatch(loc, func(servers map[string]any) { servers[name] = body })
	}
	if exists {
		if err := claudeCLIRemove(loc, name); err != nil {
			return err
		}
	}
	if err := claudeCLIAdd(loc, name, body); err != nil {
		if exists {
			_ = claudeCLIAdd(loc, name, old)
		}
		return err
	}
	return nil
}

// ClaudeRemove deletes name at loc.
func ClaudeRemove(loc Location, name string) error {
	if _, exists := claudeEntry(loc, name); !exists {
		return nil
	}
	if useClaudePatch(loc) {
		return claudePatch(loc, func(servers map[string]any) { delete(servers, name) })
	}
	return claudeCLIRemove(loc, name)
}

// useClaudePatch: the CLI keys local/project scope by its working directory,
// so it cannot reach entries of a project whose directory is gone.
func useClaudePatch(loc Location) bool {
	if macos.FindCLI("claude") == "" {
		return true
	}
	return loc.Project != "" && !dirExists(loc.Project)
}

func claudeCLIAdd(loc Location, name string, body map[string]any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = runClaudeCLI(loc, "mcp", "add-json", "-s", loc.Scope, name, string(payload))
	return err
}

func claudeCLIRemove(loc Location, name string) error {
	_, err := runClaudeCLI(loc, "mcp", "remove", "-s", loc.Scope, name)
	return err
}

// runClaudeCLI runs claude in the location's project dir (local/project scope
// are keyed by the working directory) through a login shell for PATH.
func runClaudeCLI(loc Location, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := runLoginCommand(ctx, claudeDir(loc), macos.FindCLI("claude"), args...)
	if err != nil {
		return out, fmt.Errorf("claude %s 失败: %s", strings.Join(args[:2], " "), firstNonEmptyLine(out, err.Error()))
	}
	return out, nil
}

func claudeDir(loc Location) string {
	if loc.Project != "" {
		return loc.Project
	}
	return os.TempDir()
}

// runLoginCommand runs bin with args through `zsh -l` (PATH from the login
// shell, args passed verbatim) and returns combined output.
func runLoginCommand(ctx context.Context, dir, bin string, args ...string) (string, error) {
	full := append([]string{"-l", "-c", `exec "$0" "$@"`, bin}, args...)
	cmd := exec.CommandContext(ctx, "/bin/zsh", full...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if ctx.Err() != nil {
		return buf.String(), fmt.Errorf("超时")
	}
	return buf.String(), err
}

// claudePatch is the no-CLI fallback: re-read the file immediately before
// writing and touch only the target mcpServers map.
func claudePatch(loc Location, edit func(servers map[string]any)) error {
	path := ClaudeFile(loc)
	root := map[string]any{}
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("%s 不是有效 JSON: %w", path, err)
		}
	}
	container := root
	if loc.Scope == ScopeLocal {
		projects, _ := root["projects"].(map[string]any)
		if projects == nil {
			projects = map[string]any{}
			root["projects"] = projects
		}
		proj, _ := projects[loc.Project].(map[string]any)
		if proj == nil {
			proj = map[string]any{}
			projects[loc.Project] = proj
		}
		container = proj
	}
	servers, _ := container["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	edit(servers)
	container["mcpServers"] = servers
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteWithBackup(path, append(out, '\n'), 0o600)
}

func firstNonEmptyLine(s, fallback string) string {
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return fallback
}

// prettyJSON renders one entry the way it appears in the client file.
func prettyJSON(name string, body map[string]any) string {
	out, _ := json.MarshalIndent(map[string]any{name: body}, "", "  ")
	return string(out)
}
