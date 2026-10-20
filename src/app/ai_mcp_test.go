package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hatch/ai/mcp"
	"hatch/core/config"
)

func TestSetMCPEnabledRejectsExistingClaudeEntry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{
  "mcpServers": {
    "mysql": {"type":"stdio","command":"existing","args":[],"env":{}}
  }
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	a := &App{cfg: &config.Config{OverrideMeta: map[string]config.ScriptMeta{}, MCPServers: []config.MCPServer{{
		ID:        "mysql",
		Name:      "MySQL",
		Transport: mcp.TransportStdio,
		Command:   "node",
	}}}}
	err := a.SetMCPEnabled("mysql", mcp.TargetClaude, true)
	if err == nil || !strings.Contains(err.Error(), "已跳过以避免覆盖") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestSetMCPEnabledRejectsExistingCodexEntry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	codexDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte("[mcp_servers.mysql]\ncommand = \"existing\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := &App{cfg: &config.Config{OverrideMeta: map[string]config.ScriptMeta{}, MCPServers: []config.MCPServer{{
		ID:        "mysql",
		Name:      "MySQL",
		Transport: mcp.TransportStdio,
		Command:   "node",
	}}}}
	err := a.SetMCPEnabled("mysql", mcp.TargetCodex, true)
	if err == nil || !strings.Contains(err.Error(), "已跳过以避免覆盖") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

// Re-syncing an MCP that Hatch already enabled must be allowed to overwrite its
// own entry rather than tripping the conflict guard.
func TestSetMCPEnabledAllowsResyncOfOwnEntry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{
  "mcpServers": {
    "mysql": {"type":"stdio","command":"node","args":[],"env":{}}
  }
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	a := &App{cfg: &config.Config{OverrideMeta: map[string]config.ScriptMeta{}, MCPServers: []config.MCPServer{{
		ID:              "mysql",
		Name:            "MySQL",
		Transport:       mcp.TransportStdio,
		Command:         "node",
		EnabledInClaude: true,
	}}}}
	if err := a.SetMCPEnabled("mysql", mcp.TargetClaude, true); err != nil {
		t.Fatalf("resync should be allowed, got %v", err)
	}
}

func TestDeleteMCPServerRemovesSyncedTargets(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{
  "mcpServers": {
    "mysql": {"type":"stdio","command":"node","args":[],"env":{}},
    "keep": {"type":"stdio","command":"keep","args":[],"env":{}}
  }
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	codexDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexConfig := `model = "gpt-5"

# Managed by Hatch MCP
[mcp_servers.mysql]
command = "node"

[mcp_servers.keep]
command = "keep"
`
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(codexConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	a := &App{cfg: &config.Config{OverrideMeta: map[string]config.ScriptMeta{}, MCPServers: []config.MCPServer{{
		ID:              "mysql",
		Name:            "MySQL",
		Transport:       mcp.TransportStdio,
		Command:         "node",
		EnabledInClaude: true,
		EnabledInCodex:  true,
	}}}}
	if err := a.DeleteMCPServer("mysql"); err != nil {
		t.Fatal(err)
	}
	claudeData, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(claudeData), `"mysql"`) || !strings.Contains(string(claudeData), `"keep"`) {
		t.Fatalf("unexpected claude config:\n%s", string(claudeData))
	}
	codexData, err := os.ReadFile(filepath.Join(codexDir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(codexData), `[mcp_servers.mysql]`) || !strings.Contains(string(codexData), `[mcp_servers.keep]`) {
		t.Fatalf("unexpected codex config:\n%s", string(codexData))
	}
	if len(a.cfg.MCPServers) != 0 {
		t.Fatalf("server remained in registry: %#v", a.cfg.MCPServers)
	}
}

func TestSuggestMCPRecommendationsFallsBackWithoutAI(t *testing.T) {
	a := &App{cfg: &config.Config{OverrideMeta: map[string]config.ScriptMeta{}}}
	result, err := a.SuggestMCPRecommendations("elasticsearch for codex")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) == 0 {
		t.Fatalf("expected fallback candidates")
	}
	if result.Candidates[0].ID != "elastic-agent-builder" || result.Candidates[1].Server.Command != "docker" {
		t.Fatalf("expected agent builder then docker fallback, got %#v", result.Candidates[:2])
	}
	if !strings.Contains(result.SourceNote, "离线") {
		t.Fatalf("expected offline source note, got %q", result.SourceNote)
	}
}
