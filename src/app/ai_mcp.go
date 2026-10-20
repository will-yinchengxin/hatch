package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hatch/ai/mcp"
	"hatch/ai/prompts"
	"hatch/core/config"
	"hatch/core/strutil"
)

// ─── AI · MCP ────────────────────────────────────────────────────────────────
//
// The registry lives in config.MCPServers; internal/ai/mcp discovers, installs,
// renders and syncs it into ~/.claude.json and ~/.codex/config.toml.

// GetMCPSettings returns the MCP master directory configuration.
func (a *App) GetMCPSettings() mcp.Settings {
	a.mu.RLock()
	custom := a.cfg.MCPDir
	a.mu.RUnlock()
	return mcp.ResolveSettings(custom)
}

// SaveMCPSettings sets the MCP master directory ("" = default).
func (a *App) SaveMCPSettings(mcpDir string) error {
	mcpDir = strings.TrimSpace(mcpDir)
	if mcpDir != "" {
		if err := os.MkdirAll(mcpDir, 0o755); err != nil {
			return err
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.MCPDir = mcpDir
	return a.cfg.Save()
}

// GetMCPTemplates returns the built-in MCP starter gallery.
func (a *App) GetMCPTemplates() []config.MCPServer {
	return mcp.Templates()
}

// ListMCPServers returns the registry, newest first. On first call it seeds
// the registry by discovering subdirectories of the MCP master dir.
func (a *App) ListMCPServers() ([]config.MCPServer, error) {
	a.mu.Lock()
	if !a.cfg.MCPSeeded {
		if seeded, err := mcp.Discover(mcp.ResolveSettings(a.cfg.MCPDir).MCPDir); err == nil && len(seeded) > 0 {
			a.cfg.MCPServers = seeded
		}
		a.cfg.MCPSeeded = true
		_ = a.cfg.Save()
	}
	out := mcp.Clone(a.cfg.MCPServers)
	a.mu.Unlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt == out[j].UpdatedAt {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out, nil
}

// SaveMCPServer creates or updates a server (matched by ID), preserving sync
// flags and the last test results.
func (a *App) SaveMCPServer(server config.MCPServer) error {
	server, err := mcp.Normalize(server)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, s := range a.cfg.MCPServers {
		if s.ID == server.ID {
			server.EnabledInClaude = s.EnabledInClaude
			server.EnabledInCodex = s.EnabledInCodex
			if server.LastTestResults == nil {
				server.LastTestResults = s.LastTestResults
			}
			a.cfg.MCPServers[i] = server
			return a.cfg.Save()
		}
	}
	a.cfg.MCPServers = append(a.cfg.MCPServers, server)
	return a.cfg.Save()
}

// DeleteMCPServer removes a server from synced targets and the registry.
func (a *App) DeleteMCPServer(id string) error {
	_, err := a.removeMCPServer(id)
	return err
}

// UninstallMCPFromDisk is DeleteMCPServer plus deleting the local checkout
// (git/local sources with a localPath only).
func (a *App) UninstallMCPFromDisk(id string) error {
	removed, err := a.removeMCPServer(id)
	if err != nil || removed == nil {
		return err
	}
	if removed.LocalPath != "" && (removed.SourceType == "git" || removed.SourceType == "local") {
		if err := os.RemoveAll(removed.LocalPath); err != nil {
			return fmt.Errorf("从 Hatch 删除成功，但清理本地目录失败: %w", err)
		}
	}
	return nil
}

// removeMCPServer unsyncs the server from both targets, then drops it from
// the registry. It returns nil, nil when the id is unknown.
func (a *App) removeMCPServer(id string) (*config.MCPServer, error) {
	if !mcp.ValidID(id) {
		return nil, fmt.Errorf("invalid MCP id")
	}
	a.mu.RLock()
	removed, found := mcp.Find(a.cfg.MCPServers, id)
	a.mu.RUnlock()
	if !found {
		return nil, nil
	}
	if removed.EnabledInClaude {
		if err := mcp.WriteClaude(removed, false); err != nil {
			return nil, err
		}
	}
	if removed.EnabledInCodex {
		if err := mcp.WriteCodex(removed, false); err != nil {
			return nil, err
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	keep := a.cfg.MCPServers[:0]
	for _, s := range a.cfg.MCPServers {
		if s.ID != id {
			keep = append(keep, s)
		}
	}
	a.cfg.MCPServers = keep
	return &removed, a.cfg.Save()
}

// UpdateMCPFromGit runs git pull in a git-sourced server's checkout, re-infers
// its command and re-syncs enabled targets.
func (a *App) UpdateMCPFromGit(id string) (config.MCPServer, error) {
	if !mcp.ValidID(id) {
		return config.MCPServer{}, fmt.Errorf("invalid MCP id")
	}
	a.mu.RLock()
	server, found := mcp.Find(a.cfg.MCPServers, id)
	a.mu.RUnlock()
	if !found {
		return config.MCPServer{}, fmt.Errorf("MCP not found: %s", id)
	}
	if server.SourceType != "git" {
		return config.MCPServer{}, fmt.Errorf("只有 git 来源的 MCP 支持更新")
	}
	if server.LocalPath == "" {
		return config.MCPServer{}, fmt.Errorf("本地目录路径未设置")
	}
	if _, err := os.Stat(server.LocalPath); err != nil {
		return config.MCPServer{}, fmt.Errorf("本地目录不存在: %s", server.LocalPath)
	}
	if err := mcp.GitPull(server.LocalPath); err != nil {
		return config.MCPServer{}, err
	}
	updated := mcp.InferFromDir(server.LocalPath)
	updated.ID = server.ID
	updated.Name = server.Name
	updated.SourceType = server.SourceType
	updated.SourceURL = server.SourceURL
	updated.EnabledInClaude = server.EnabledInClaude
	updated.EnabledInCodex = server.EnabledInCodex
	updated.Tags = server.Tags
	updated.LastTestResults = server.LastTestResults
	if err := a.SaveMCPServer(updated); err != nil {
		return config.MCPServer{}, err
	}
	if updated.EnabledInClaude {
		_ = mcp.WriteClaude(updated, true)
	}
	if updated.EnabledInCodex {
		_ = mcp.WriteCodex(updated, true)
	}
	return updated, nil
}

// ImportMCPFromLocalDir registers an existing local checkout.
func (a *App) ImportMCPFromLocalDir(path string) (config.MCPServer, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return config.MCPServer{}, fmt.Errorf("local path is required")
	}
	fi, err := os.Stat(path)
	if err != nil {
		return config.MCPServer{}, err
	}
	if !fi.IsDir() {
		return config.MCPServer{}, fmt.Errorf("local path is not a directory")
	}
	server := mcp.InferFromDir(path)
	if err := a.SaveMCPServer(server); err != nil {
		return config.MCPServer{}, err
	}
	return server, nil
}

// InstallMCPFromGit clones url into <mcpDir>/<id> and registers it.
func (a *App) InstallMCPFromGit(url, name string) (config.MCPServer, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return config.MCPServer{}, fmt.Errorf("git URL is required")
	}
	id := strutil.SanitizeID(name)
	if id == "" {
		id = strutil.SanitizeID(strings.TrimSuffix(filepath.Base(url), ".git"))
	}
	if !mcp.ValidID(id) {
		return config.MCPServer{}, fmt.Errorf("invalid MCP id")
	}
	base := a.GetMCPSettings().MCPDir
	if err := os.MkdirAll(base, 0o755); err != nil {
		return config.MCPServer{}, err
	}
	dst := filepath.Join(base, id)
	if _, err := os.Stat(dst); err == nil {
		return config.MCPServer{}, fmt.Errorf("target directory already exists: %s", dst)
	}
	if err := mcp.GitClone(url, dst); err != nil {
		return config.MCPServer{}, err
	}
	server := mcp.InferFromDir(dst)
	server.ID = id
	server.Name = strutil.OrDefault(strings.TrimSpace(name), server.Name)
	server.SourceType = "git"
	server.SourceURL = url
	if err := a.SaveMCPServer(server); err != nil {
		return config.MCPServer{}, err
	}
	return server, nil
}

// SaveNpmMCPServer registers an `npx -y <package>` server.
func (a *App) SaveNpmMCPServer(packageName, name string) (config.MCPServer, error) {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return config.MCPServer{}, fmt.Errorf("npm package is required")
	}
	id := strutil.SanitizeID(name)
	if id == "" {
		id = strutil.SanitizeID(packageName)
	}
	server := config.MCPServer{
		ID:          id,
		Name:        strutil.OrDefault(strings.TrimSpace(name), packageName),
		SourceType:  "npm",
		PackageName: packageName,
		Transport:   mcp.TransportStdio,
		Command:     "npx",
		Args:        []string{"-y", packageName},
		UpdatedAt:   time.Now().Unix(),
	}
	if err := a.SaveMCPServer(server); err != nil {
		return config.MCPServer{}, err
	}
	return server, nil
}

// SetMCPEnabled syncs (or removes) one server to/from target ("claude" | "codex").
// First-time enable refuses to clobber a same-named hand-written entry;
// re-syncing an already-enabled server may overwrite its own entry.
func (a *App) SetMCPEnabled(id, target string, enabled bool) error {
	if !mcp.ValidID(id) {
		return fmt.Errorf("invalid MCP id")
	}
	a.mu.RLock()
	server, found := mcp.Find(a.cfg.MCPServers, id)
	a.mu.RUnlock()
	if !found {
		return fmt.Errorf("MCP not found: %s", id)
	}

	var err error
	switch target {
	case mcp.TargetClaude:
		if enabled && !server.EnabledInClaude {
			exists, derr := mcp.ClaudeEntryExists(server.ID)
			if derr != nil {
				return derr
			}
			if exists {
				return fmt.Errorf("Claude Code 配置中已存在名为「%s」的 MCP，已跳过以避免覆盖手工配置；如需 Hatch 接管请先在 ~/.claude.json 中删除该条目", server.ID)
			}
		}
		err = mcp.WriteClaude(server, enabled)
	case mcp.TargetCodex:
		if enabled && !server.EnabledInCodex && mcp.CodexEntryExists(server.ID) {
			return fmt.Errorf("Codex 配置中已存在名为「%s」的 MCP，已跳过以避免覆盖手工配置；如需 Hatch 接管请先在 ~/.codex/config.toml 中删除该条目", server.ID)
		}
		err = mcp.WriteCodex(server, enabled)
	default:
		return fmt.Errorf("unknown MCP target: %s", target)
	}
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for i, s := range a.cfg.MCPServers {
		if s.ID != id {
			continue
		}
		if target == mcp.TargetClaude {
			a.cfg.MCPServers[i].EnabledInClaude = enabled
		} else {
			a.cfg.MCPServers[i].EnabledInCodex = enabled
		}
		a.cfg.MCPServers[i].UpdatedAt = time.Now().Unix()
		break
	}
	return a.cfg.Save()
}

// TestMCPConnection checks a server via `claude mcp get` / `codex mcp get` and
// stores the result on the server.
func (a *App) TestMCPConnection(id, target string) (config.MCPTestResult, error) {
	if !mcp.ValidID(id) {
		return config.MCPTestResult{}, fmt.Errorf("invalid MCP id")
	}
	target = strings.TrimSpace(target)
	if target != mcp.TargetClaude && target != mcp.TargetCodex {
		return config.MCPTestResult{}, fmt.Errorf("unknown MCP target: %s", target)
	}

	result := mcp.TestConnection(id, target)
	result.Target = target
	result.CheckedAt = time.Now().Unix()
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.cfg.MCPServers {
		if a.cfg.MCPServers[i].ID != id {
			continue
		}
		if a.cfg.MCPServers[i].LastTestResults == nil {
			a.cfg.MCPServers[i].LastTestResults = map[string]config.MCPTestResult{}
		}
		a.cfg.MCPServers[i].LastTestResults[target] = result
		a.cfg.MCPServers[i].UpdatedAt = time.Now().Unix()
		return result, a.cfg.Save()
	}
	return result, fmt.Errorf("MCP not found: %s", id)
}

// SuggestMCPRecommendations returns curated candidates for query, re-ranked
// and annotated by AI when configured; any AI failure falls back to the
// offline list. Candidates are only filled into the editor, never saved.
func (a *App) SuggestMCPRecommendations(query string) (mcp.RecommendationResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		query = "elasticsearch mcp for Claude Code and Codex"
	}
	fallback := mcp.CuratedRecommendations(query)

	ai, err := a.readyAI()
	if err != nil {
		fallback.SourceNote = "离线内置推荐；配置 AI 后可基于这些候选生成排序说明。"
		return fallback, nil
	}
	userBytes, _ := json.Marshal(map[string]any{
		"query":      query,
		"candidates": fallback.Candidates,
		"context": []string{
			"Claude Code 和 Codex 都可使用 stdio MCP；Hatch 会分别写入 ~/.claude.json 和 ~/.codex/config.toml。",
			"Elastic 官方 mcp-server-elasticsearch 0.4.0+ 使用 docker.elastic.co/mcp/elasticsearch，支持 stdio/http，旧 npm 线已弃用。",
			"用户通常需要 ES_URL，以及 ES_API_KEY 或 ES_USERNAME/ES_PASSWORD。",
		},
	})
	raw, err := a.callAIOp(prompts.OpSuggestMCPRecommendations, ai, prompts.MCPRecommendations, string(userBytes))
	if err != nil {
		fallback.SourceNote = "AI 推荐失败，已使用离线内置推荐。"
		return fallback, nil
	}
	merged, ok := mcp.MergeAIRanking(fallback, raw)
	if !ok {
		fallback.SourceNote = "AI 返回无法解析，已使用离线内置推荐。"
		return fallback, nil
	}
	merged.SourceNote = "AI 基于 Hatch 内置候选和当前 Elastic MCP 公开说明完成排序；请在填入密钥前复核来源。"
	return merged, nil
}
