// Package mcp manages Hatch's MCP server registry: discovering local checkouts,
// installing from git/npm, normalizing definitions, syncing them into Claude
// Code (~/.claude.json) and Codex (~/.codex/config.toml), and testing them.
// Hatch only writes configuration; it never keeps MCP processes running.
package mcp

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"hatch/ai/agent"
	"hatch/core/config"
	"hatch/core/strutil"
)

const (
	TargetClaude = agent.TargetClaude
	TargetCodex  = agent.TargetCodex

	TransportStdio = "stdio"
	TransportHTTP  = "http"
)

// ExecCommand is swappable in tests.
var ExecCommand = exec.CommandContext

// Settings is the MCP directory configuration shown on the MCP page.
type Settings struct {
	MCPDir        string `json:"mcpDir"`
	MCPDirIsSet   bool   `json:"mcpDirIsSet"`
	DefaultMCPDir string `json:"defaultMCPDir"`
}

// DefaultDir is ~/MyHub/mcp.
func DefaultDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "MyHub", "mcp")
}

// ResolveSettings combines the configured dir with the default.
func ResolveSettings(custom string) Settings {
	custom = strings.TrimSpace(custom)
	def := DefaultDir()
	return Settings{
		MCPDir:        strutil.OrDefault(custom, def),
		MCPDirIsSet:   custom != "",
		DefaultMCPDir: def,
	}
}

// ValidID reports whether id is a safe MCP id.
func ValidID(id string) bool {
	return strutil.ValidID(id)
}

// Find returns the server with the given id.
func Find(in []config.MCPServer, id string) (config.MCPServer, bool) {
	for _, s := range in {
		if s.ID == id {
			return s, true
		}
	}
	return config.MCPServer{}, false
}

// Discover infers one server per non-hidden subdirectory of base.
func Discover(base string) ([]config.MCPServer, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var servers []config.MCPServer
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		server := InferFromDir(filepath.Join(base, e.Name()))
		server.UpdatedAt = now
		servers = append(servers, server)
	}
	return servers, nil
}
