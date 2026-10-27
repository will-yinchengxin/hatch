package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hatch/ai/mcp"
	"hatch/core/config"
	"hatch/core/fsutil"
)

func TestMCPMetaParkingMirrorAndPrune(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a := &App{cfg: &config.Config{LegacyMCPServers: []config.LegacyMCPServer{{ID: "mysql"}, {ID: "keep"}}}}

	if err := a.updateMCPMeta("ctx", func(m *config.MCPMeta) {
		m.Parked = map[string]string{"claude:user": `{"command":"npx"}`}
	}, "mysql"); err != nil {
		t.Fatal(err)
	}
	if got := a.mcpParked("ctx", "claude:user"); got != `{"command":"npx"}` {
		t.Fatalf("parked = %q", got)
	}
	mirror := filepath.Join(fsutil.AppDataPath("mcp-parked"), "ctx.json")
	if data, err := os.ReadFile(mirror); err != nil || !strings.Contains(string(data), "npx") {
		t.Fatalf("mirror = %s %v", data, err)
	}
	if len(a.cfg.LegacyMCPServers) != 1 || a.cfg.LegacyMCPServers[0].ID != "keep" {
		t.Fatalf("legacy = %+v", a.cfg.LegacyMCPServers)
	}

	if err := a.updateMCPMeta("ctx", func(m *config.MCPMeta) { delete(m.Parked, "claude:user") }, ""); err != nil {
		t.Fatal(err)
	}
	if len(a.cfg.MCPMeta) != 0 {
		t.Fatalf("empty overlay not pruned: %+v", a.cfg.MCPMeta)
	}
	if _, err := os.Stat(mirror); !os.IsNotExist(err) {
		t.Fatalf("mirror not removed: %v", err)
	}
}

func TestMCPCodexToggleThroughApp(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[mcp_servers.docs]\nurl = \"https://x/mcp\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := &App{cfg: &config.Config{}}
	key := mcp.Location{Client: mcp.ClientCodex, Scope: mcp.ScopeUser}.Key()
	if err := a.SetMCPEnabled("docs", key, false); err != nil {
		t.Fatal(err)
	}
	states := a.globalMCPStates()
	if len(states) != 1 || !states[0].HasCodex || states[0].Codex {
		t.Fatalf("states = %+v", states)
	}
	if len(a.cfg.MCPMeta) != 0 {
		t.Fatalf("codex toggle must not park: %+v", a.cfg.MCPMeta)
	}
	if _, err := a.findMCPEntry("docs", key); err != nil {
		t.Fatal(err)
	}
}
