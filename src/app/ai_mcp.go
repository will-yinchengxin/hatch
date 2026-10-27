package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hatch/ai/mcp"
	"hatch/core/config"
	"hatch/core/fsutil"
)

// ─── AI · MCP ────────────────────────────────────────────────────────────────
//
// Claude Code / Codex config files are the source of truth (hatch/ai/mcp scans
// and edits them in place). Hatch only persists config.MCPMeta: origin, entries
// it parked while disabled (Claude has no off switch) and last probe results.

// GetMCPInventory scans every client config; includeCLI also asks the CLIs
// for plugin / connector servers (slow: claude health-checks each server).
func (a *App) GetMCPInventory(includeCLI bool) mcp.Inventory {
	a.mu.RLock()
	opt := mcp.ScanOptions{
		IncludeCLI: includeCLI,
		Meta:       append([]config.MCPMeta(nil), a.cfg.MCPMeta...),
		Legacy:     append([]config.LegacyMCPServer(nil), a.cfg.LegacyMCPServers...),
	}
	for _, d := range a.cfg.Directories {
		opt.ExtraProjects = append(opt.ExtraProjects, d.Path)
	}
	a.mu.RUnlock()
	return mcp.Scan(opt)
}

// GetMCPCatalog returns the curated starter list.
func (a *App) GetMCPCatalog() []mcp.CatalogItem { return mcp.Catalog() }

// ParseMCPInput turns pasted text into drafts; registry names are resolved
// online. Drafts only fill the form — nothing is written.
func (a *App) ParseMCPInput(text string) (mcp.ParseResult, error) {
	res, err := mcp.Parse(text)
	if err != nil || res.RegistryName == "" {
		return res, err
	}
	drafts, err := a.LookupMCPRegistry(res.RegistryName)
	if err != nil {
		return res, err
	}
	res.Drafts = drafts
	return res, nil
}

// SearchMCPRegistry searches the official MCP Registry by name.
func (a *App) SearchMCPRegistry(query string) ([]mcp.RegistryHit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return mcp.SearchRegistry(ctx, query)
}

// LookupMCPRegistry resolves an exact registry name into install drafts.
func (a *App) LookupMCPRegistry(name string) ([]mcp.Draft, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return mcp.LookupRegistry(ctx, name)
}

// PreviewMCPApply dry-runs an apply and returns per-location before/after.
func (a *App) PreviewMCPApply(req mcp.ApplyRequest) mcp.ApplyPlan {
	return mcp.Plan(req)
}

// ApplyMCP writes one server definition to the chosen locations.
func (a *App) ApplyMCP(req mcp.ApplyRequest) error {
	a.mcpOpMu.Lock()
	defer a.mcpOpMu.Unlock()
	if err := mcp.Apply(req); err != nil {
		return err
	}
	return a.updateMCPMeta(req.Name, func(m *config.MCPMeta) {
		if req.Origin != "" && req.Origin != "edit" {
			m.Origin = req.Origin
		}
		for _, loc := range req.Targets {
			delete(m.Parked, loc.Key())
		}
	}, req.Name)
}

// SetMCPEnabled turns one entry on/off without losing its definition.
func (a *App) SetMCPEnabled(name, locKey string, enabled bool) error {
	loc, err := mcp.ParseLocationKey(locKey)
	if err != nil {
		return err
	}
	a.mcpOpMu.Lock()
	defer a.mcpOpMu.Unlock()
	parked := a.mcpParked(name, locKey)
	newParked, err := mcp.SetEnabled(name, loc, enabled, parked)
	if err != nil {
		return err
	}
	return a.updateMCPMeta(name, func(m *config.MCPMeta) {
		if newParked == "" {
			delete(m.Parked, locKey)
		} else {
			if m.Parked == nil {
				m.Parked = map[string]string{}
			}
			m.Parked[locKey] = newParked
		}
	}, "")
}

// RemoveMCPEntry deletes one entry (and any parked copy) at a location.
func (a *App) RemoveMCPEntry(name, locKey string) error {
	loc, err := mcp.ParseLocationKey(locKey)
	if err != nil {
		return err
	}
	a.mcpOpMu.Lock()
	defer a.mcpOpMu.Unlock()
	if err := mcp.Remove(name, loc); err != nil {
		return err
	}
	return a.updateMCPMeta(name, func(m *config.MCPMeta) {
		delete(m.Parked, locKey)
		delete(m.LastProbe, locKey)
	}, "")
}

// CopyMCPEntry copies the entry at fromKey to every toKeys location (also
// used to resolve drift: copy the version you trust over the others).
func (a *App) CopyMCPEntry(name, fromKey string, toKeys []string) error {
	src, err := a.findMCPEntry(name, fromKey)
	if err != nil {
		return err
	}
	req := mcp.ApplyRequest{Name: name, Spec: src.Spec, SourceClient: src.Location.Client, Origin: "copy:" + fromKey}
	if src.Location.Client == mcp.ClientDesktop {
		req.SourceClient = mcp.ClientClaude
	}
	for _, k := range toKeys {
		loc, err := mcp.ParseLocationKey(k)
		if err != nil {
			return err
		}
		req.Targets = append(req.Targets, loc)
	}
	return a.ApplyMCP(req)
}

// ProbeMCP connects to one entry as an MCP client and stores the result.
func (a *App) ProbeMCP(name, locKey string) (config.MCPProbe, error) {
	e, err := a.findMCPEntry(name, locKey)
	if err != nil {
		return config.MCPProbe{}, err
	}
	res := mcp.Probe(e)
	err = a.updateMCPMeta(name, func(m *config.MCPMeta) {
		if m.LastProbe == nil {
			m.LastProbe = map[string]config.MCPProbe{}
		}
		m.LastProbe[locKey] = res
	}, "")
	return res, err
}

// SetMCPDisabledTools writes Codex disabled_tools for one entry.
func (a *App) SetMCPDisabledTools(name, locKey string, disabled []string) error {
	loc, err := mcp.ParseLocationKey(locKey)
	if err != nil {
		return err
	}
	a.mcpOpMu.Lock()
	defer a.mcpOpMu.Unlock()
	return mcp.SetCodexDisabledTools(name, loc, disabled)
}

// DiscardLegacyMCP drops a pre-redesign registry entry that was never synced.
func (a *App) DiscardLegacyMCP(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	kept := a.cfg.LegacyMCPServers[:0]
	for _, s := range a.cfg.LegacyMCPServers {
		if s.ID != name {
			kept = append(kept, s)
		}
	}
	a.cfg.LegacyMCPServers = kept
	return a.cfg.Save()
}

// findMCPEntry locates a live (or parked) entry by name and location key.
func (a *App) findMCPEntry(name, locKey string) (mcp.Entry, error) {
	inv := a.GetMCPInventory(false)
	for _, r := range inv.Rows {
		if r.Name != name {
			continue
		}
		for _, e := range r.Entries {
			if e.LocationKey == locKey {
				return e, nil
			}
		}
	}
	return mcp.Entry{}, fmt.Errorf("找不到 MCP「%s」(%s)，可能已被其他程序修改，请刷新", name, locKey)
}

func (a *App) mcpParked(name, locKey string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, m := range a.cfg.MCPMeta {
		if m.Name == name {
			return m.Parked[locKey]
		}
	}
	return ""
}

// updateMCPMeta edits (creating if needed) the overlay for name, drops empty
// overlays, mirrors parked entries to disk and, when dropLegacy is set, removes
// the same-named legacy registry entry.
func (a *App) updateMCPMeta(name string, edit func(m *config.MCPMeta), dropLegacy string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := -1
	for i := range a.cfg.MCPMeta {
		if a.cfg.MCPMeta[i].Name == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		a.cfg.MCPMeta = append(a.cfg.MCPMeta, config.MCPMeta{Name: name})
		idx = len(a.cfg.MCPMeta) - 1
	}
	m := &a.cfg.MCPMeta[idx]
	edit(m)
	if len(m.Parked) == 0 {
		m.Parked = nil
	}
	if len(m.LastProbe) == 0 {
		m.LastProbe = nil
	}
	writeParkedMirror(*m)
	if m.Origin == "" && m.Notes == "" && m.Parked == nil && m.LastProbe == nil {
		a.cfg.MCPMeta = append(a.cfg.MCPMeta[:idx], a.cfg.MCPMeta[idx+1:]...)
	}
	if dropLegacy != "" {
		kept := a.cfg.LegacyMCPServers[:0]
		for _, s := range a.cfg.LegacyMCPServers {
			if s.ID != dropLegacy {
				kept = append(kept, s)
			}
		}
		a.cfg.LegacyMCPServers = kept
	}
	return a.cfg.Save()
}

// writeParkedMirror keeps a copy of parked entries outside config.json so a
// disabled Claude server survives even if Hatch's config is lost.
func writeParkedMirror(m config.MCPMeta) {
	dir := fsutil.AppDataPath("mcp-parked")
	path := filepath.Join(dir, strings.ReplaceAll(m.Name, string(os.PathSeparator), "_")+".json")
	if len(m.Parked) == 0 {
		_ = os.Remove(path)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(m.Parked, "", "  ")
	if err == nil {
		_ = os.WriteFile(path, data, 0o600)
	}
}
