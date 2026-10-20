package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hatch/core/config"
)

func TestWriteCodexMCPConfigUpdatesManagedBlock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	codexDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(codexDir, "config.toml")
	initial := `model = "gpt-5.5"

[mcp_servers.keep]
command = "keep"

# Managed by Hatch MCP
[mcp_servers.mysql]
command = "old"

[mcp_servers.mysql.env]
MYSQL_HOST = "old"
`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	server := config.MCPServer{
		ID:        "mysql",
		Transport: TransportStdio,
		Command:   "node",
		Args:      []string{"/tmp/mysql/dist/index.js"},
		Cwd:       "/tmp/mysql",
		Env:       map[string]string{"MYSQL_HOST": "127.0.0.1", "MYSQL_PASS": "secret"},
	}
	if err := WriteCodex(server, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		`model = "gpt-5.5"`,
		`[mcp_servers.keep]`,
		`[mcp_servers.mysql]`,
		`command = "node"`,
		`args = ["/tmp/mysql/dist/index.js"]`,
		`cwd = "/tmp/mysql"`,
		`[mcp_servers.mysql.env]`,
		`MYSQL_HOST = "127.0.0.1"`,
		`MYSQL_PASS = "secret"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("config missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(content, `command = "old"`) {
		t.Fatalf("stale managed block remained:\n%s", content)
	}
}

func TestWriteClaudeMCPConfigUpdatesGlobalFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".claude.json")
	initial := `{"theme":"dark","mcpServers":{"keep":{"type":"stdio","command":"keep","args":[],"env":{}}}}`
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}

	server := config.MCPServer{
		ID:        "consul",
		Transport: TransportStdio,
		Command:   "node",
		Args:      []string{"/tmp/consul/build/index.js"},
		Cwd:       "/tmp/consul",
		Env:       map[string]string{"CONSUL_HOST": "localhost"},
	}
	if err := WriteClaude(server, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		`"theme": "dark"`,
		`"keep"`,
		`"consul"`,
		`"type": "stdio"`,
		`"command": "node"`,
		`"/tmp/consul/build/index.js"`,
		`"CONSUL_HOST": "localhost"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("settings missing %q:\n%s", want, content)
		}
	}

	if err := WriteClaude(server, false); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = string(data)
	if strings.Contains(content, `"consul"`) {
		t.Fatalf("consul was not removed:\n%s", content)
	}
	if !strings.Contains(content, `"keep"`) {
		t.Fatalf("unrelated server removed:\n%s", content)
	}
}

func TestInferMCPServerFromPackageJSON(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "example-mcp-server")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
  "name": "@scope/example-mcp",
  "bin": {"example-mcp": "dist/index.js"}
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	server := InferFromDir(dir)
	if server.ID != "example" {
		t.Fatalf("unexpected id: %s", server.ID)
	}
	if server.Command != "node" {
		t.Fatalf("command = %q", server.Command)
	}
	if len(server.Args) != 1 || !strings.HasSuffix(server.Args[0], "dist/index.js") {
		t.Fatalf("args = %#v", server.Args)
	}
}

func TestParseClaudeMCPTestOutputFailure(t *testing.T) {
	out := `gbrain:
  Scope: User config (available in all your projects)
  Status: ✗ Failed to connect
  Type: stdio`
	result := ParseTestOutput(TargetClaude, out, 0, nil)
	if result.Status != "failed" {
		t.Fatalf("status = %s", result.Status)
	}
	if !strings.Contains(result.Summary, "Failed to connect") {
		t.Fatalf("summary = %q", result.Summary)
	}
}

func TestCloneMCPServersDeepCopiesMutableFields(t *testing.T) {
	servers := []config.MCPServer{{
		ID:              "mysql",
		Args:            []string{"one"},
		Tags:            []string{"db"},
		Env:             map[string]string{"MYSQL_HOST": "127.0.0.1"},
		HTTPHeaders:     map[string]string{"X-Test": "1"},
		LastTestResults: map[string]config.MCPTestResult{"codex": {Status: "unknown"}},
	}}
	cloned := Clone(servers)
	cloned[0].Args[0] = "two"
	cloned[0].Tags[0] = "cache"
	cloned[0].Env["MYSQL_HOST"] = "localhost"
	cloned[0].HTTPHeaders["X-Test"] = "2"
	cloned[0].LastTestResults["codex"] = config.MCPTestResult{Status: "failed"}
	if servers[0].Args[0] != "one" ||
		servers[0].Tags[0] != "db" ||
		servers[0].Env["MYSQL_HOST"] != "127.0.0.1" ||
		servers[0].HTTPHeaders["X-Test"] != "1" ||
		servers[0].LastTestResults["codex"].Status != "unknown" {
		t.Fatalf("clone shared mutable fields: %#v", servers[0])
	}
}

func TestCuratedMCPRecommendationsPreferElasticOfficialOptions(t *testing.T) {
	result := CuratedRecommendations("es mcp")
	if result.Query != "es mcp" {
		t.Fatalf("query not preserved: %q", result.Query)
	}
	if len(result.Candidates) < 4 {
		t.Fatalf("expected multiple candidates, got %#v", result.Candidates)
	}
	first := result.Candidates[0]
	if first.ID != "elastic-agent-builder" || first.Version == "" {
		t.Fatalf("first candidate = %#v, want agent builder with version line", first)
	}
	stdio := result.Candidates[1]
	if stdio.ID != "elastic-official-stdio" {
		t.Fatalf("second candidate = %q, want official stdio", stdio.ID)
	}
	if !stdio.ClaudeSupported || !stdio.CodexSupported {
		t.Fatalf("official stdio should support both targets: %#v", stdio)
	}
	if stdio.Server.Command != "docker" || stdio.Server.Transport != TransportStdio {
		t.Fatalf("unexpected official server config: %#v", stdio.Server)
	}
	if !containsString(stdio.Server.Args, "docker.elastic.co/mcp/elasticsearch") || !containsString(stdio.Server.Args, "stdio") {
		t.Fatalf("official args missing image or stdio: %#v", stdio.Server.Args)
	}
	if stdio.QuickEnv["ES_URL"] == "" {
		t.Fatalf("ES_URL quick env missing: %#v", stdio.QuickEnv)
	}
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
