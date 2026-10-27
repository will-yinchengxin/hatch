package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hatch/core/config"
)

// TestMain doubles as a tiny stdio MCP server for the probe test.
func TestMain(m *testing.M) {
	if os.Getenv("HATCH_FAKE_MCP") == "1" {
		fakeStdioServer()
		return
	}
	os.Exit(m.Run())
}

func fakeStdioServer() {
	fmt.Println("this banner line is not JSON and must be ignored")
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		var msg map[string]any
		if json.Unmarshal(sc.Bytes(), &msg) != nil || msg["id"] == nil {
			continue
		}
		var result any
		switch msg["method"] {
		case "initialize":
			result = map[string]any{"protocolVersion": "2025-06-18", "serverInfo": map[string]any{"name": "fake", "version": "9.9"}, "capabilities": map[string]any{}}
		case "tools/list":
			result = map[string]any{"tools": []any{map[string]any{"name": "echo", "description": "Echo back\nsecond line"}, map[string]any{"name": "sum"}}}
		}
		out, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": msg["id"], "result": result})
		fmt.Println(string(out))
	}
}

func TestParseClaudeJSONSnippet(t *testing.T) {
	res, err := Parse(`{
  "mcpServers": {
    "context7": { "command": "npx", "args": ["-y", "@upstash/context7-mcp"], "env": { "CONTEXT7_API_KEY": "YOUR_API_KEY" }, "alwaysLoad": true },
    "sentry": { "type": "http", "url": "https://mcp.sentry.dev/mcp" },
  }
}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Drafts) != 2 || res.Drafts[0].Name != "context7" || res.Drafts[1].Spec.Transport != TransportHTTP {
		t.Fatalf("drafts = %+v", res.Drafts)
	}
	d := res.Drafts[0]
	if d.Spec.Extra["alwaysLoad"] != true || d.SourceClient != ClientClaude {
		t.Fatalf("extra not kept: %+v", d.Spec)
	}
	if len(d.Inputs) != 1 || d.Inputs[0].Key != "CONTEXT7_API_KEY" || !d.Inputs[0].Secret {
		t.Fatalf("inputs = %+v", d.Inputs)
	}
}

func TestParseFragmentAndVSCodeShapes(t *testing.T) {
	res, err := Parse(`"git": {"command": "uvx", "args": ["mcp-server-git"]}`)
	if err != nil || len(res.Drafts) != 1 || res.Drafts[0].Name != "git" {
		t.Fatalf("fragment: %+v %v", res, err)
	}
	res, err = Parse(`{"mcp": {"servers": {"fetch": {"command": "uvx", "args": ["mcp-server-fetch"]}}}}`)
	if err != nil || res.Drafts[0].Name != "fetch" {
		t.Fatalf("vscode: %+v %v", res, err)
	}
	res, err = Parse(`{"command": "npx", "args": ["-y", "@modelcontextprotocol/server-memory"]}`)
	if err != nil || res.Drafts[0].Name != "memory" {
		t.Fatalf("bare spec: %+v %v", res, err)
	}
}

func TestParseClaudeAddCommand(t *testing.T) {
	res, err := Parse(`claude mcp add my-server -s user -e API_KEY=xxx OTHER=1 -- npx -y my-mcp-server --flag "a b"`)
	if err != nil {
		t.Fatal(err)
	}
	d := res.Drafts[0]
	if d.Name != "my-server" || d.Scope != "user" || d.Spec.Command != "npx" {
		t.Fatalf("draft = %+v", d)
	}
	if strings.Join(d.Spec.Args, "|") != "-y|my-mcp-server|--flag|a b" {
		t.Fatalf("args = %q", d.Spec.Args)
	}
	if d.Spec.Env["API_KEY"] != "xxx" || d.Spec.Env["OTHER"] != "1" {
		t.Fatalf("env = %v", d.Spec.Env)
	}
	res, err = Parse(`claude mcp add --transport http corridor https://app.corridor.dev/api/mcp --header "Authorization: Bearer abc"`)
	if err != nil || res.Drafts[0].Spec.URL != "https://app.corridor.dev/api/mcp" || res.Drafts[0].Spec.Headers["Authorization"] != "Bearer abc" {
		t.Fatalf("http add: %+v %v", res, err)
	}
	res, err = Parse(`claude mcp add-json -s user weather '{"type":"stdio","command":"/bin/weather","args":[]}'`)
	if err != nil || res.Drafts[0].Name != "weather" || res.Drafts[0].Spec.Command != "/bin/weather" {
		t.Fatalf("add-json: %+v %v", res, err)
	}
}

func TestParseCodexAddCommandAndTOML(t *testing.T) {
	res, err := Parse(`codex mcp add docs --url https://developers.openai.com/mcp --bearer-token-env-var DOCS_TOKEN`)
	if err != nil || res.Drafts[0].Spec.URL == "" || res.Drafts[0].Spec.BearerTokenEnvVar != "DOCS_TOKEN" {
		t.Fatalf("codex add: %+v %v", res, err)
	}
	res, err = Parse("```toml\n[mcp_servers.excalidraw]\ncommand = \"node\"\nargs = [\"dist/index.js\"]\n\n[mcp_servers.excalidraw.tools.export_scene]\napproval_mode = \"approve\"\n```")
	if err != nil {
		t.Fatal(err)
	}
	d := res.Drafts[0]
	if d.Name != "excalidraw" || d.SourceClient != ClientCodex || d.Spec.Extra["tools"] == nil {
		t.Fatalf("toml draft = %+v", d)
	}
}

func TestParseShortForms(t *testing.T) {
	for in, want := range map[string]string{
		"https://mcp.linear.app/mcp":        "linear",
		"@playwright/mcp@latest":            "playwright",
		"uvx mcp-server-git --repository .": "git",
		"io.github.upstash/context7":        "",
	} {
		res, err := Parse(in)
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if want == "" {
			if res.RegistryName != in {
				t.Fatalf("%s: expected registry lookup, got %+v", in, res)
			}
			continue
		}
		if res.Drafts[0].Name != want {
			t.Fatalf("%s: name = %s", in, res.Drafts[0].Name)
		}
	}
}

const codexSample = `model = "gpt-5"
# keep this comment

[mcp_servers.docs]
url = "https://developers.openai.com/mcp"

# Managed by Hatch MCP
[mcp_servers.canvas]
command = "/opt/homebrew/bin/node"
args = ["/x/index.js"]

[mcp_servers.canvas.env]
CANVAS_PORT = "3100"

[mcp_servers.canvas.tools.export_scene]
approval_mode = "approve"

[projects."/Users/me/p"]
trust_level = "trusted"
`

func TestCodexUpsertPreservesOtherContentAndExtras(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(codexSample), 0o644); err != nil {
		t.Fatal(err)
	}
	servers, err := readCodexServers(path)
	if err != nil {
		t.Fatal(err)
	}
	spec, enabled := SpecFromCodex(servers["canvas"])
	if !enabled || spec.Env["CANVAS_PORT"] != "3100" || spec.Extra["tools"] == nil {
		t.Fatalf("spec = %+v", spec)
	}
	spec.Args = []string{"/x/v2.js"}
	table, err := CodexTable(spec, true, true)
	if err != nil {
		t.Fatal(err)
	}
	out, err := CodexUpsert(codexSample, "canvas", table)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# keep this comment", `model = "gpt-5"`, "[mcp_servers.docs]", `args = ["/x/v2.js"]`, "[mcp_servers.canvas.tools.export_scene]", `approval_mode = "approve"`, `[projects."/Users/me/p"]`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, legacyCodexMarker) || strings.Contains(out, "/x/index.js") {
		t.Fatalf("old block or marker left:\n%s", out)
	}
	if strings.Index(out, "[mcp_servers.canvas]") > strings.Index(out, "[projects.") {
		t.Fatalf("block moved instead of replaced in place:\n%s", out)
	}
}

func TestCodexSetEnabledTouchesOneLine(t *testing.T) {
	off, err := CodexSetEnabled(codexSample, "canvas", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(off, "[mcp_servers.canvas]\nenabled = false\ncommand") {
		t.Fatalf("disable:\n%s", off)
	}
	on, err := CodexSetEnabled(off, "canvas", true)
	if err != nil {
		t.Fatal(err)
	}
	if on != codexSample {
		t.Fatalf("round trip changed content:\n%s", on)
	}
}

func TestCodexRemoveDropsSubtables(t *testing.T) {
	out := CodexRemove(codexSample, "canvas")
	if strings.Contains(out, "canvas") || !strings.Contains(out, "[mcp_servers.docs]") || !strings.Contains(out, "trust_level") {
		t.Fatalf("remove:\n%s", out)
	}
}

func TestClaudeAndCodexRendering(t *testing.T) {
	s := Spec{Transport: TransportHTTP, URL: "https://x/mcp", Headers: map[string]string{"Authorization": "Bearer ${GH_TOKEN}", "X-Team": "a"}}
	table, err := CodexTable(s, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if table["bearer_token_env_var"] != "GH_TOKEN" || table["http_headers"].(map[string]any)["Authorization"] != nil {
		t.Fatalf("codex table = %v", table)
	}
	if _, err := CodexTable(Spec{Transport: TransportSSE, URL: "https://x/sse"}, true, false); err == nil {
		t.Fatal("codex must reject sse")
	}
	body := ClaudeJSON(Spec{Transport: TransportHTTP, URL: "https://x", BearerTokenEnvVar: "TOK"}, false)
	if body["headers"].(map[string]string)["Authorization"] != "Bearer ${TOK}" {
		t.Fatalf("claude body = %v", body)
	}
}

func TestDraftsFromRegistry(t *testing.T) {
	var s registryServer
	raw := `{"name":"io.github.upstash/context7","packages":[{"registryType":"npm","identifier":"@upstash/context7-mcp","transport":{"type":"stdio"},"environmentVariables":[{"name":"CONTEXT7_API_KEY","isSecret":true}]},{"registryType":"oci","identifier":"ghcr.io/x/y","version":"1.2","environmentVariables":[{"name":"TOKEN","isRequired":true}]},{"registryType":"mcpb","identifier":"https://x.mcpb"}],"remotes":[{"type":"streamable-http","url":"https://mcp.context7.com/mcp","headers":[{"name":"Authorization","isSecret":true}]}]}`
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	drafts := DraftsFromRegistry(s)
	if len(drafts) != 3 {
		t.Fatalf("drafts = %+v", drafts)
	}
	if drafts[0].Spec.URL == "" || !drafts[0].Inputs[0].Secret {
		t.Fatalf("remote = %+v", drafts[0])
	}
	if drafts[1].Spec.Command != "npx" || drafts[1].Name != "context7" {
		t.Fatalf("npm = %+v", drafts[1])
	}
	if got := strings.Join(drafts[2].Spec.Args, " "); got != "run -i --rm -e TOKEN ghcr.io/x/y:1.2" {
		t.Fatalf("oci args = %s", got)
	}
}

func TestCatalogIsValid(t *testing.T) {
	items := Catalog()
	if len(items) < 10 {
		t.Fatalf("catalog has %d items", len(items))
	}
	for _, it := range items {
		d := it.Draft()
		if !ValidName(d.Name) || d.Spec.Validate() != nil {
			t.Fatalf("invalid catalog item %+v", it)
		}
	}
}

func TestScanGroupsEntriesAndDetectsDrift(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	proj := filepath.Join(home, "proj")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	claude := map[string]any{
		"mcpServers": map[string]any{"git": map[string]any{"type": "stdio", "command": "uvx", "args": []string{"mcp-server-git"}}},
		"projects":   map[string]any{proj: map[string]any{"mcpServers": map[string]any{"prom": map[string]any{"command": "uvx", "args": []string{"prometheus-mcp-server"}}}}},
	}
	data, _ := json.Marshal(claude)
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, ".mcp.json"), []byte(`{"mcpServers":{"shared":{"type":"http","url":"https://x/mcp"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	codex := "[mcp_servers.git]\ncommand = \"uvx\"\nargs = [\"mcp-server-git\", \"--repository\", \".\"]\nenabled = false\n"
	if err := os.WriteFile(filepath.Join(home, ".codex", "config.toml"), []byte(codex), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := []config.MCPMeta{{Name: "parkedone", Parked: map[string]string{"claude:user": `{"type":"stdio","command":"echo"}`}}}
	legacy := []config.LegacyMCPServer{{ID: "mysql", Transport: "stdio", Command: "npx", Env: map[string]string{"MYSQL_PASS": ""}}, {ID: "git", Command: "x"}}
	inv := Scan(ScanOptions{Meta: meta, Legacy: legacy})

	rows := map[string]ServerRow{}
	for _, r := range inv.Rows {
		rows[r.Name] = r
	}
	git := rows["git"]
	if len(git.Entries) != 2 || !git.Drift {
		t.Fatalf("git row = %+v", git)
	}
	if git.Entries[1].Location.Client != ClientCodex || git.Entries[1].Enabled {
		t.Fatalf("codex git entry = %+v", git.Entries[1])
	}
	if e := rows["prom"].Entries[0]; e.Location.Scope != ScopeLocal || e.Location.Project != proj {
		t.Fatalf("prom = %+v", e)
	}
	if e := rows["shared"].Entries[0]; e.Location.Scope != ScopeProject || e.Spec.Transport != TransportHTTP {
		t.Fatalf("shared = %+v", e)
	}
	if e := rows["parkedone"].Entries[0]; !e.Parked || e.Enabled {
		t.Fatalf("parked = %+v", e)
	}
	if len(inv.Legacy) != 1 || inv.Legacy[0].Name != "mysql" || len(inv.Legacy[0].Inputs) != 1 {
		t.Fatalf("legacy = %+v", inv.Legacy)
	}
	if len(inv.Projects) == 0 || inv.Projects[0].Path != proj {
		t.Fatalf("projects = %+v", inv.Projects)
	}
}

func TestApplyAndToggleCodexEntry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(codexSample), 0o644); err != nil {
		t.Fatal(err)
	}
	loc := Location{Client: ClientCodex, Scope: ScopeUser}
	req := ApplyRequest{Name: "context7", Spec: Spec{Command: "npx", Args: []string{"-y", "@upstash/context7-mcp"}, Extra: map[string]any{"alwaysLoad": true}}, Targets: []Location{loc}, SourceClient: ClientClaude}
	plan := Plan(req)
	if !plan.CanApply || plan.Items[0].Action != "add" || strings.Contains(plan.Items[0].After, "alwaysLoad") {
		t.Fatalf("plan = %+v", plan)
	}
	if err := Apply(req); err != nil {
		t.Fatal(err)
	}
	if plan := Plan(req); plan.Items[0].Action != "unchanged" {
		t.Fatalf("re-plan = %+v", plan.Items[0])
	}
	if _, err := SetEnabled("context7", loc, false, ""); err != nil {
		t.Fatal(err)
	}
	if err := SetCodexDisabledTools("canvas", loc, []string{"export_scene"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	out := string(data)
	if !strings.Contains(out, "[mcp_servers.context7]\nenabled = false") || !strings.Contains(out, `disabled_tools = ["export_scene"]`) || !strings.Contains(out, "approval_mode") {
		t.Fatalf("config:\n%s", out)
	}
	if err := Remove("context7", loc); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if strings.Contains(string(data), "context7") {
		t.Fatalf("not removed:\n%s", data)
	}
}

func TestProbeStdioAndHTTP(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	e := Entry{Name: "fake", Location: Location{Client: ClientClaude, Scope: ScopeUser}, Spec: Spec{Transport: TransportStdio, Command: exe, Env: map[string]string{"HATCH_FAKE_MCP": "1"}}}
	res := Probe(e)
	if res.Status != "ok" || res.ServerName != "fake" || len(res.Tools) != 2 || res.Tools[0].Description != "Echo back" {
		t.Fatalf("stdio probe = %+v", res)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var msg map[string]any
		_ = json.NewDecoder(r.Body).Decode(&msg)
		if msg["id"] == nil {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Mcp-Session-Id", "s1")
		var result any = map[string]any{"protocolVersion": "2025-06-18", "serverInfo": map[string]any{"name": "remote"}}
		if msg["method"] == "tools/list" {
			result = map[string]any{"tools": []any{map[string]any{"name": "search"}}}
		}
		out, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": msg["id"], "result": result})
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", out)
	}))
	defer srv.Close()
	h := Entry{Name: "remote", Location: Location{Client: ClientClaude, Scope: ScopeUser}, Spec: Spec{Transport: TransportHTTP, URL: srv.URL, Headers: map[string]string{"Authorization": "Bearer good"}}}
	if res := Probe(h); res.Status != "ok" || len(res.Tools) != 1 {
		t.Fatalf("http probe = %+v", res)
	}
	h.Spec.Headers["Authorization"] = "Bearer bad"
	if res := Probe(h); res.Status != "auth_required" {
		t.Fatalf("auth probe = %+v", res)
	}
}
