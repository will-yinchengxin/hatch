package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"hatch/ai/agent"
	"hatch/core/config"
	"hatch/core/macos"
)

// Inventory is the unified view of every MCP server the clients know about.
type Inventory struct {
	Rows        []ServerRow  `json:"rows"`
	Projects    []ProjectRef `json:"projects"`
	Sources     []Source     `json:"sources"`
	CLIIncluded bool         `json:"cliIncluded"`
	Legacy      []Draft      `json:"legacy,omitempty"`
	ScannedAt   int64        `json:"scannedAt"`
}

// ProjectRef is a directory that can hold local / project scoped servers.
type ProjectRef struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	HasServers   bool   `json:"hasServers"`
	CodexTrusted bool   `json:"codexTrusted"`
}

// Source reports how reading one config source went.
type Source struct {
	Label  string `json:"label"`
	Path   string `json:"path,omitempty"`
	Status string `json:"status"` // ok | missing | error
	Error  string `json:"error,omitempty"`
	Count  int    `json:"count"`
}

// ScanOptions feeds Hatch-side state into a scan.
type ScanOptions struct {
	ExtraProjects []string                 // workbench dirs etc.
	IncludeCLI    bool                     // also run `claude mcp list` / `codex mcp list --json`
	Meta          []config.MCPMeta         // parked entries + probes
	Legacy        []config.LegacyMCPServer // pre-redesign registry
}

// Scan reads every source and groups entries by name.
func Scan(opt ScanOptions) Inventory {
	inv := Inventory{ScannedAt: time.Now().Unix(), CLIIncluded: opt.IncludeCLI}
	var entries []Entry
	add := func(name string, loc Location, file string, spec Spec, enabled bool, raw string) {
		entries = append(entries, Entry{
			Name: name, Location: loc, LocationKey: loc.Key(), Label: loc.Label(), File: file,
			Spec: spec, Enabled: enabled, Raw: raw, ReadOnly: !loc.Writable(),
		})
	}

	// Claude: user + local scopes in ~/.claude.json.
	projects := map[string]*ProjectRef{}
	ref := func(p string) *ProjectRef {
		if r, ok := projects[p]; ok {
			return r
		}
		r := &ProjectRef{Path: p, Name: filepath.Base(p)}
		projects[p] = r
		return r
	}
	st, claudePath, err := readClaudeState()
	src := Source{Label: "Claude Code (~/.claude.json)", Path: claudePath, Status: "ok"}
	if err != nil {
		src.Status, src.Error = statusOf(err), err.Error()
	}
	for name, body := range st.User {
		add(name, Location{Client: ClientClaude, Scope: ScopeUser}, claudePath, SpecFromClaude(body), true, prettyJSON(name, body))
		src.Count++
	}
	for p, proj := range st.Projects {
		exists := dirExists(p)
		if exists {
			ref(p)
		}
		for name, body := range proj.Servers {
			loc := Location{Client: ClientClaude, Scope: ScopeLocal, Project: p}
			add(name, loc, claudePath, SpecFromClaude(body), !proj.Disabled[name], prettyJSON(name, body))
			entries[len(entries)-1].Stale = !exists
			if exists {
				ref(p).HasServers = true
			}
			src.Count++
		}
	}
	inv.Sources = append(inv.Sources, src)
	for _, p := range opt.ExtraProjects {
		if p = strings.TrimSpace(p); p != "" && dirExists(p) {
			ref(p)
		}
	}

	// Codex user config + trusted projects.
	codexPath, _ := agent.CodexConfigPath()
	for _, p := range codexTrustedProjects(codexPath) {
		// A trusted $HOME would make ~/.codex/config.toml its own project config.
		if projFile, _ := CodexConfigFile(Location{Client: ClientCodex, Scope: ScopeProject, Project: p}); dirExists(p) && projFile != codexPath {
			ref(p).CodexTrusted = true
		}
	}

	// Claude project scope (.mcp.json) for every known project.
	mcpJSON := Source{Label: "项目 .mcp.json", Status: "ok"}
	for p := range projects {
		servers, err := readMCPJSON(p)
		if err != nil {
			if !os.IsNotExist(err) {
				mcpJSON.Status, mcpJSON.Error = "error", p+": "+err.Error()
			}
			continue
		}
		disabled := st.Projects[p].Disabled
		for name, body := range servers {
			loc := Location{Client: ClientClaude, Scope: ScopeProject, Project: p}
			add(name, loc, mcpJSONPath(p), SpecFromClaude(body), !disabled[name], prettyJSON(name, body))
			projects[p].HasServers = true
			mcpJSON.Count++
		}
	}
	inv.Sources = append(inv.Sources, mcpJSON)

	codexSrc := Source{Label: "Codex (~/.codex/config.toml)", Path: codexPath, Status: "ok"}
	if servers, err := readCodexServers(codexPath); err != nil {
		codexSrc.Status, codexSrc.Error = statusOf(err), err.Error()
	} else {
		for name, table := range servers {
			spec, enabled := SpecFromCodex(table)
			add(name, Location{Client: ClientCodex, Scope: ScopeUser}, codexPath, spec, enabled, RenderCodexBlock(name, table))
			codexSrc.Count++
		}
	}
	inv.Sources = append(inv.Sources, codexSrc)
	for p, r := range projects {
		if !r.CodexTrusted {
			continue
		}
		loc := Location{Client: ClientCodex, Scope: ScopeProject, Project: p}
		path, _ := CodexConfigFile(loc)
		servers, err := readCodexServers(path)
		if err != nil {
			continue
		}
		for name, table := range servers {
			spec, enabled := SpecFromCodex(table)
			add(name, loc, path, spec, enabled, RenderCodexBlock(name, table))
			r.HasServers = true
		}
	}

	// Claude Desktop: read-only, offered as a copy source.
	desktop := Source{Label: "Claude Desktop", Path: ClaudeDesktopConfigPath(), Status: "ok"}
	if servers, err := readDesktopServers(); err != nil {
		desktop.Status = statusOf(err)
		if desktop.Status == "error" {
			desktop.Error = err.Error()
		}
	} else {
		for name, body := range servers {
			add(name, Location{Client: ClientDesktop, Scope: ScopeUser}, desktop.Path, SpecFromClaude(body), true, prettyJSON(name, body))
			desktop.Count++
		}
	}
	inv.Sources = append(inv.Sources, desktop)

	// Entries Hatch disabled (Claude has no off switch, so they are parked).
	live := map[string]bool{}
	for _, e := range entries {
		live[e.Name+"\x00"+e.LocationKey] = true
	}
	metaByName := map[string]config.MCPMeta{}
	for _, m := range opt.Meta {
		metaByName[m.Name] = m
		for key, raw := range m.Parked {
			if live[m.Name+"\x00"+key] {
				continue
			}
			loc, err := ParseLocationKey(key)
			if err != nil {
				continue
			}
			var body map[string]any
			if json.Unmarshal([]byte(raw), &body) != nil {
				continue
			}
			e := Entry{
				Name: m.Name, Location: loc, LocationKey: key, Label: loc.Label(), File: ClaudeFile(loc),
				Spec: SpecFromClaude(body), Enabled: false, Parked: true, Raw: prettyJSON(m.Name, body), ReadOnly: !loc.Writable(),
			}
			entries = append(entries, e)
		}
	}

	if opt.IncludeCLI {
		entries = append(entries, cliEntries(entries)...)
	}

	for i := range entries {
		if entries[i].ReadOnly && entries[i].ReadOnlyReason == "" {
			entries[i].ReadOnlyReason = readOnlyReason(entries[i].Location)
		}
		if m, ok := metaByName[entries[i].Name]; ok {
			if p, ok := m.LastProbe[entries[i].LocationKey]; ok {
				probe := p
				entries[i].Probe = &probe
			}
		}
	}
	inv.Rows = groupRows(entries, metaByName)
	for _, r := range projects {
		inv.Projects = append(inv.Projects, *r)
	}
	sort.Slice(inv.Projects, func(i, j int) bool {
		if inv.Projects[i].HasServers != inv.Projects[j].HasServers {
			return inv.Projects[i].HasServers
		}
		return inv.Projects[i].Path < inv.Projects[j].Path
	})
	inv.Legacy = legacyDrafts(opt.Legacy, entries)
	return inv
}

func readOnlyReason(loc Location) string {
	switch {
	case loc.Client == ClientDesktop:
		return "Claude Desktop 配置，仅可复制到 Claude Code / Codex"
	case loc.Scope == ScopePlugin:
		return "由插件提供，请在插件中管理"
	case loc.Scope == ScopeConnector:
		return "claude.ai 连接器，请在 claude.ai 设置中管理"
	}
	return "只读"
}

func groupRows(entries []Entry, meta map[string]config.MCPMeta) []ServerRow {
	byName := map[string]*ServerRow{}
	var names []string
	for _, e := range entries {
		r, ok := byName[e.Name]
		if !ok {
			r = &ServerRow{Name: e.Name}
			if m, ok := meta[e.Name]; ok {
				mm := m
				r.Meta = &mm
			}
			byName[e.Name] = r
			names = append(names, e.Name)
		}
		r.Entries = append(r.Entries, e)
	}
	sort.Slice(names, func(i, j int) bool { return strings.ToLower(names[i]) < strings.ToLower(names[j]) })
	rows := make([]ServerRow, 0, len(names))
	for _, n := range names {
		r := byName[n]
		sort.SliceStable(r.Entries, func(i, j int) bool { return locRank(r.Entries[i].Location) < locRank(r.Entries[j].Location) })
		prints := map[string]bool{}
		for _, e := range r.Entries {
			if !e.ReadOnly {
				prints[e.Spec.Fingerprint()] = true
			}
		}
		r.Drift = len(prints) > 1
		rows = append(rows, *r)
	}
	return rows
}

func locRank(l Location) string {
	order := map[string]string{
		"claude:user": "0", "claude:local": "1", "claude:project": "2", "codex:user": "3", "codex:project": "4",
		"claude:plugin": "5", "claude:connector": "6", "codex:plugin": "7", "desktop:user": "8",
	}
	return order[l.Client+":"+l.Scope] + l.Project
}

func statusOf(err error) string {
	if os.IsNotExist(err) {
		return "missing"
	}
	return "error"
}

func dirExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// ── CLI-only entries (plugins / connectors) ────────────────────────────────

var claudeListLine = regexp.MustCompile(`^(.+?): (.*) - (.+)$`)

// cliEntries asks both CLIs for their effective lists and returns read-only
// entries the files do not contain; Claude user entries get a live status.
func cliEntries(fileEntries []Entry) []Entry {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	var claudeOut, codexOut string
	if bin := macos.FindCLI("claude"); bin != "" {
		wg.Add(1)
		go func() { defer wg.Done(); claudeOut, _ = runLoginCommand(ctx, os.TempDir(), bin, "mcp", "list") }()
	}
	if bin := macos.FindCLI("codex"); bin != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codexOut, _ = runLoginCommand(ctx, os.TempDir(), bin, "mcp", "list", "--json")
		}()
	}
	wg.Wait()

	known := map[string]bool{}
	for _, e := range fileEntries {
		known[e.Location.Client+"\x00"+e.Name] = true
	}
	var out []Entry
	for _, line := range strings.Split(claudeOut, "\n") {
		m := claudeListLine.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		name, target, status := m[1], strings.TrimSpace(m[2]), strings.TrimSpace(m[3])
		if known[ClientClaude+"\x00"+name] {
			for i := range fileEntries {
				if fileEntries[i].Name == name && fileEntries[i].Location.Client == ClientClaude && fileEntries[i].Location.Scope == ScopeUser {
					fileEntries[i].CLIStatus = status
				}
			}
			continue
		}
		scope := ScopePlugin
		if strings.HasPrefix(name, "claude.ai ") {
			scope = ScopeConnector
		}
		if i := strings.Index(target, " ("); i > 0 {
			target = target[:i]
		}
		spec := Spec{Transport: TransportStdio, Command: target}
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			spec = Spec{Transport: TransportHTTP, URL: target}
		}
		loc := Location{Client: ClientClaude, Scope: scope}
		out = append(out, Entry{Name: name, Location: loc, LocationKey: loc.Key(), Label: loc.Label(), Spec: spec.Normalize(), Enabled: true, ReadOnly: true, CLIStatus: status, Raw: line})
	}
	var codexList []struct {
		Name      string `json:"name"`
		Enabled   bool   `json:"enabled"`
		Transport struct {
			Type    string   `json:"type"`
			Command string   `json:"command"`
			Args    []string `json:"args"`
			URL     string   `json:"url"`
		} `json:"transport"`
	}
	if start := strings.Index(codexOut, "["); start >= 0 && json.Unmarshal([]byte(codexOut[start:]), &codexList) == nil {
		for _, c := range codexList {
			if known[ClientCodex+"\x00"+c.Name] {
				continue
			}
			spec := Spec{Transport: TransportStdio, Command: c.Transport.Command, Args: c.Transport.Args}
			if c.Transport.URL != "" {
				spec = Spec{Transport: TransportHTTP, URL: c.Transport.URL}
			}
			loc := Location{Client: ClientCodex, Scope: ScopePlugin}
			raw, _ := json.MarshalIndent(c, "", "  ")
			out = append(out, Entry{Name: c.Name, Location: loc, LocationKey: loc.Key(), Label: loc.Label(), Spec: spec.Normalize(), Enabled: c.Enabled, ReadOnly: true, Raw: string(raw)})
		}
	}
	return out
}

// ── legacy registry ─────────────────────────────────────────────────────────

func legacyDrafts(legacy []config.LegacyMCPServer, entries []Entry) []Draft {
	present := map[string]bool{}
	for _, e := range entries {
		present[e.Name] = true
	}
	var out []Draft
	for _, s := range legacy {
		if s.ID == "" || present[s.ID] {
			continue
		}
		spec := Spec{
			Transport: s.Transport, Command: s.Command, Args: s.Args, Cwd: s.Cwd, URL: s.URL, Env: s.Env,
			Headers: s.HTTPHeaders, BearerTokenEnvVar: s.BearerTokenEnvVar,
			StartupTimeoutSec: s.StartupTimeoutSec, ToolTimeoutSec: s.ToolTimeoutSec,
		}.Normalize()
		d := Draft{Name: s.ID, Title: strings.TrimSpace(s.Name), Description: s.Description, Spec: spec, Origin: "legacy"}
		d.Inputs = detectInputs(spec)
		out = append(out, d)
	}
	return out
}
