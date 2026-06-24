package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hatch/internal/config"

	"github.com/pelletier/go-toml"
)

const maxWorkspaceSnapshots = 20

type AIWorkspacePreview struct {
	Profile  config.AIWorkspaceProfile `json:"profile"`
	Changes  []string                  `json:"changes"`
	Errors   []string                  `json:"errors"`
	Warnings []string                  `json:"warnings"`
	CanApply bool                      `json:"canApply"`
}

type AIWorkspaceSnapshotInfo struct {
	ID          string `json:"id"`
	ProfileID   string `json:"profileId"`
	ProfileName string `json:"profileName"`
	CreatedAt   int64  `json:"createdAt"`
}

type AIEnvironmentHealth struct {
	Overall        string                     `json:"overall"`
	CurrentProfile string                     `json:"currentProfile"`
	ProfileState   string                     `json:"profileState"`
	Checks         []AIEnvironmentHealthCheck `json:"checks"`
}

type AIEnvironmentHealthCheck struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Status     string `json:"status"` // ok | warn | error
	Detail     string `json:"detail"`
	Suggestion string `json:"suggestion,omitempty"`
}

type workspaceFileSnapshot struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Mode   uint32 `json:"mode,omitempty"`
	Data   string `json:"data,omitempty"`
}

type workspaceDirSnapshot struct {
	Path    string                  `json:"path"`
	Exists  bool                    `json:"exists"`
	Entries []workspaceFileSnapshot `json:"entries,omitempty"`
}

type workspaceSnapshot struct {
	AIWorkspaceSnapshotInfo
	Files       []workspaceFileSnapshot         `json:"files"`
	SkillDirs   []workspaceDirSnapshot          `json:"skillDirs"`
	MCPStates   []config.AgentResourceSelection `json:"mcpStates"`
	HookStates  []config.AgentResourceSelection `json:"hookStates"`
	LastProfile string                          `json:"lastProfile"`
}

func (a *App) ListAIWorkspaceProfiles() []config.AIWorkspaceProfile {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := append([]config.AIWorkspaceProfile(nil), a.cfg.AIWorkspaceProfiles...)
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out
}

func (a *App) SaveAIWorkspaceProfile(profile config.AIWorkspaceProfile) (config.AIWorkspaceProfile, error) {
	profile.Name = strings.TrimSpace(profile.Name)
	profile.RootPath = strings.TrimSpace(profile.RootPath)
	if profile.Name == "" || profile.RootPath == "" {
		return config.AIWorkspaceProfile{}, fmt.Errorf("工作区名称和目录不能为空")
	}
	abs, err := filepath.Abs(profile.RootPath)
	if err != nil {
		return config.AIWorkspaceProfile{}, err
	}
	profile.RootPath = filepath.Clean(abs)
	if profile.ID == "" {
		profile.ID = fmt.Sprintf("ws-%x", time.Now().UnixNano())
	}
	if !validWorkspaceID(profile.ID) {
		return config.AIWorkspaceProfile{}, fmt.Errorf("invalid workspace profile id")
	}
	profile.Skills = normalizeSelections(profile.Skills)
	profile.MCPServers = normalizeSelections(profile.MCPServers)
	profile.Hooks = normalizeSelections(profile.Hooks)
	profile.UpdatedAt = time.Now().Unix()

	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.cfg.AIWorkspaceProfiles {
		if a.cfg.AIWorkspaceProfiles[i].ID == profile.ID {
			a.cfg.AIWorkspaceProfiles[i] = profile
			return profile, a.cfg.Save()
		}
	}
	a.cfg.AIWorkspaceProfiles = append(a.cfg.AIWorkspaceProfiles, profile)
	return profile, a.cfg.Save()
}

func (a *App) DeleteAIWorkspaceProfile(id string) error {
	if !validWorkspaceID(id) {
		return fmt.Errorf("invalid workspace profile id")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	keep := a.cfg.AIWorkspaceProfiles[:0]
	for _, p := range a.cfg.AIWorkspaceProfiles {
		if p.ID != id {
			keep = append(keep, p)
		}
	}
	a.cfg.AIWorkspaceProfiles = keep
	if a.cfg.LastAppliedWorkspaceID == id {
		a.cfg.LastAppliedWorkspaceID = ""
	}
	return a.cfg.Save()
}

// CaptureCurrentAIWorkspaceProfile builds an editable profile from current Hatch-managed state.
func (a *App) CaptureCurrentAIWorkspaceProfile(name, rootPath string) (config.AIWorkspaceProfile, error) {
	profile := config.AIWorkspaceProfile{Name: name, RootPath: rootPath}
	claude := a.GetClaudeCodeConfig()
	codex := a.GetCodexModelConfig()
	profile.ClaudeModel, profile.ClaudeEffort = claude.Model, claude.EffortLevel
	profile.CodexModel, profile.CodexEffort, profile.CodexGoals = codex.Model, codex.EffortLevel, codex.GoalsEnabled
	skills, err := a.GetSkills()
	if err != nil {
		return profile, err
	}
	for _, s := range skills {
		if s.InClaude || s.InCodex {
			profile.Skills = append(profile.Skills, config.AgentResourceSelection{ID: s.Name, Claude: s.InClaude, Codex: s.InCodex})
		}
	}
	a.mu.RLock()
	for _, s := range a.cfg.MCPServers {
		if s.EnabledInClaude || s.EnabledInCodex {
			profile.MCPServers = append(profile.MCPServers, config.AgentResourceSelection{ID: s.ID, Claude: s.EnabledInClaude, Codex: s.EnabledInCodex})
		}
	}
	for _, h := range a.cfg.Hooks {
		if h.EnabledInClaude || h.EnabledInCodex {
			profile.Hooks = append(profile.Hooks, config.AgentResourceSelection{ID: h.ID, Claude: h.EnabledInClaude, Codex: h.EnabledInCodex})
		}
	}
	a.mu.RUnlock()
	return a.SaveAIWorkspaceProfile(profile)
}

func (a *App) PreviewAIWorkspaceProfileApply(id string) (AIWorkspacePreview, error) {
	p, err := a.workspaceProfile(id)
	if err != nil {
		return AIWorkspacePreview{}, err
	}
	return a.previewWorkspaceProfile(p), nil
}

func (a *App) ApplyAIWorkspaceProfile(id string) (AIWorkspaceSnapshotInfo, error) {
	p, err := a.workspaceProfile(id)
	if err != nil {
		return AIWorkspaceSnapshotInfo{}, err
	}
	preview := a.previewWorkspaceProfile(p)
	if !preview.CanApply {
		return AIWorkspaceSnapshotInfo{}, fmt.Errorf("无法应用工作区画像: %s", strings.Join(preview.Errors, "；"))
	}
	snap, err := a.captureWorkspaceSnapshot(p)
	if err != nil {
		return AIWorkspaceSnapshotInfo{}, err
	}
	if err := a.writeWorkspaceSnapshot(snap); err != nil {
		return AIWorkspaceSnapshotInfo{}, err
	}
	if err := a.applyWorkspaceProfile(p); err != nil {
		_ = a.restoreWorkspaceSnapshot(snap)
		return AIWorkspaceSnapshotInfo{}, fmt.Errorf("应用失败，已恢复应用前状态: %w", err)
	}
	a.mu.Lock()
	a.cfg.LastAppliedWorkspaceID = p.ID
	err = a.cfg.Save()
	a.mu.Unlock()
	if err != nil {
		return AIWorkspaceSnapshotInfo{}, err
	}
	_ = a.pruneWorkspaceSnapshots()
	return snap.AIWorkspaceSnapshotInfo, nil
}

func (a *App) ListAIWorkspaceSnapshots() []AIWorkspaceSnapshotInfo {
	dir, err := a.workspaceSnapshotDir()
	if err != nil {
		return nil
	}
	entries, _ := os.ReadDir(dir)
	var out []AIWorkspaceSnapshotInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var s workspaceSnapshot
		if json.Unmarshal(data, &s) == nil && s.ID != "" {
			out = append(out, s.AIWorkspaceSnapshotInfo)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out
}

func (a *App) RollbackAIWorkspaceSnapshot(id string) error {
	if !validWorkspaceID(id) {
		return fmt.Errorf("invalid workspace snapshot id")
	}
	data, err := os.ReadFile(filepath.Join(mustWorkspaceSnapshotDir(), id+".json"))
	if err != nil {
		return err
	}
	var snap workspaceSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	if err := a.restoreWorkspaceSnapshot(snap); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.LastAppliedWorkspaceID = snap.LastProfile
	err = a.cfg.Save()
	a.mu.Unlock()
	return err
}

func (a *App) GetAIEnvironmentHealth() AIEnvironmentHealth {
	health := AIEnvironmentHealth{Overall: "ok", Checks: []AIEnvironmentHealthCheck{}}
	add := func(c AIEnvironmentHealthCheck) {
		if c.Status == "error" {
			health.Overall = "error"
		} else if c.Status == "warn" && health.Overall == "ok" {
			health.Overall = "warn"
		}
		health.Checks = append(health.Checks, c)
	}
	if findCLI("claude") == "" {
		add(AIEnvironmentHealthCheck{ID: "claude-cli", Label: "Claude Code CLI", Status: "warn", Detail: "未在登录 shell PATH 中找到 claude", Suggestion: "安装 Claude Code 或检查 ~/.zshrc PATH"})
	} else {
		add(AIEnvironmentHealthCheck{ID: "claude-cli", Label: "Claude Code CLI", Status: "ok", Detail: "已可用"})
	}
	if findCodexCLI() == "" {
		add(AIEnvironmentHealthCheck{ID: "codex-cli", Label: "Codex CLI", Status: "warn", Detail: "未在登录 shell PATH 中找到 codex", Suggestion: "安装 Codex 或检查 ~/.zshrc PATH"})
	} else {
		add(AIEnvironmentHealthCheck{ID: "codex-cli", Label: "Codex CLI", Status: "ok", Detail: "已可用"})
	}
	home, _ := os.UserHomeDir()
	health.addJSONCheck(filepath.Join(home, ".claude", "settings.json"), "claude-settings", "Claude 设置")
	health.addJSONCheck(filepath.Join(home, ".claude.json"), "claude-mcp", "Claude MCP 配置")
	health.addJSONCheck(filepath.Join(home, ".codex", "hooks.json"), "codex-hooks", "Codex Hooks 配置")
	if raw, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml")); err == nil {
		if _, err := parseTOML(string(raw)); err != nil {
			add(AIEnvironmentHealthCheck{ID: "codex-config", Label: "Codex 配置", Status: "error", Detail: "config.toml 语法无效", Suggestion: "在助手页修复 TOML 后再应用画像"})
		} else {
			add(AIEnvironmentHealthCheck{ID: "codex-config", Label: "Codex 配置", Status: "ok", Detail: "TOML 有效"})
		}
	}
	a.mu.RLock()
	managedMCP := append([]config.MCPServer(nil), a.cfg.MCPServers...)
	managedHooks := cloneHooks(a.cfg.Hooks)
	a.mu.RUnlock()
	for _, server := range managedMCP {
		if !(server.EnabledInClaude || server.EnabledInCodex) || server.Transport != mcpTransportStdio {
			continue
		}
		if findCLI(server.Command) == "" {
			add(AIEnvironmentHealthCheck{ID: "mcp-" + server.ID, Label: "MCP: " + server.Name, Status: "warn", Detail: "命令不可用: " + server.Command, Suggestion: "安装命令或在 MCP 中改为可执行入口"})
		}
	}
	for _, hook := range managedHooks {
		if !(hook.EnabledInClaude || hook.EnabledInCodex) {
			continue
		}
		command := firstExecutableToken(hook.Command)
		if command != "" && findCLI(command) == "" {
			add(AIEnvironmentHealthCheck{ID: "hook-" + hook.ID, Label: "Hook: " + hook.Name, Status: "warn", Detail: "命令可能不可用: " + command, Suggestion: "检查 Hook 命令或登录 shell PATH"})
		}
	}
	a.mu.RLock()
	last := a.cfg.LastAppliedWorkspaceID
	a.mu.RUnlock()
	if last == "" {
		health.ProfileState = "none"
		return health
	}
	p, err := a.workspaceProfile(last)
	if err != nil {
		health.ProfileState = "missing"
		add(AIEnvironmentHealthCheck{ID: "workspace-profile", Label: "工作区画像", Status: "warn", Detail: "最近应用的画像已被删除"})
		return health
	}
	health.CurrentProfile = p.Name
	preview := a.previewWorkspaceProfile(p)
	if len(preview.Changes) > 0 {
		health.ProfileState = "drifted"
		add(AIEnvironmentHealthCheck{ID: "workspace-profile", Label: "工作区画像", Status: "warn", Detail: "当前全局配置与「" + p.Name + "」不一致", Suggestion: "打开工作区页预览并重新应用，或回滚快照"})
	} else {
		health.ProfileState = "applied"
		add(AIEnvironmentHealthCheck{ID: "workspace-profile", Label: "工作区画像", Status: "ok", Detail: "「" + p.Name + "」仍与当前 Hatch 管理状态一致"})
	}
	return health
}

func (h *AIEnvironmentHealth) addJSONCheck(path, id, label string) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		h.Checks = append(h.Checks, AIEnvironmentHealthCheck{ID: id, Label: label, Status: "ok", Detail: "尚未创建"})
		return
	}
	if err != nil || !json.Valid(data) {
		h.Overall = "error"
		h.Checks = append(h.Checks, AIEnvironmentHealthCheck{ID: id, Label: label, Status: "error", Detail: "JSON 无效", Suggestion: "修复原始配置后再同步"})
		return
	}
	h.Checks = append(h.Checks, AIEnvironmentHealthCheck{ID: id, Label: label, Status: "ok", Detail: "JSON 有效"})
}

func (a *App) previewWorkspaceProfile(p config.AIWorkspaceProfile) AIWorkspacePreview {
	out := AIWorkspacePreview{Profile: p, CanApply: true}
	if fi, err := os.Stat(p.RootPath); err != nil || !fi.IsDir() {
		out.Errors = append(out.Errors, "工作区目录不存在或不是目录: "+p.RootPath)
	}
	claude, codex := a.GetClaudeCodeConfig(), a.GetCodexModelConfig()
	if claude.Model != p.ClaudeModel || claude.EffortLevel != p.ClaudeEffort {
		out.Changes = append(out.Changes, "Claude Code 模型或推理档位将更新")
	}
	if codex.Model != p.CodexModel || codex.EffortLevel != p.CodexEffort || codex.GoalsEnabled != p.CodexGoals {
		out.Changes = append(out.Changes, "Codex 模型、推理档位或 Goals 将更新")
	}
	a.mu.RLock()
	mcp := append([]config.MCPServer(nil), a.cfg.MCPServers...)
	hooks := cloneHooks(a.cfg.Hooks)
	a.mu.RUnlock()
	for _, s := range p.Skills {
		if err := a.checkWorkspaceSkill(s); err != nil {
			out.Errors = append(out.Errors, err.Error())
		}
		if s.Claude && !a.skillTargetMatchesMaster(s.ID, skillTargetClaude) {
			out.Errors = append(out.Errors, "Claude 中同名 Skill 与 Hatch 主仓不同，拒绝覆盖: "+s.ID)
		}
		if s.Codex && !a.skillTargetMatchesMaster(s.ID, skillTargetCodex) {
			out.Errors = append(out.Errors, "Codex 中同名 Skill 与 Hatch 主仓不同，拒绝覆盖: "+s.ID)
		}
	}
	if skills, err := a.GetSkills(); err == nil {
		for _, skill := range skills {
			desired := selectionFor(p.Skills, skill.Name)
			if skill.InClaude && !desired.Claude && !a.skillTargetMatchesMaster(skill.Name, skillTargetClaude) {
				out.Warnings = append(out.Warnings, "Skill「"+skill.Name+"」在 Claude 中不是 Hatch 当前版本，将保留不变")
			}
			if skill.InCodex && !desired.Codex && !a.skillTargetMatchesMaster(skill.Name, skillTargetCodex) {
				out.Warnings = append(out.Warnings, "Skill「"+skill.Name+"」在 Codex 中不是 Hatch 当前版本，将保留不变")
			}
		}
	}
	for _, s := range p.MCPServers {
		server, ok := findMCP(mcp, s.ID)
		if !ok {
			out.Errors = append(out.Errors, "MCP 不存在: "+s.ID)
			continue
		}
		if server.Transport == mcpTransportStdio && strings.TrimSpace(server.Command) == "" {
			out.Errors = append(out.Errors, "MCP 未配置命令: "+s.ID)
		} else if server.Transport == mcpTransportStdio && findCLI(server.Command) == "" {
			out.Warnings = append(out.Warnings, "MCP 命令暂不可用: "+server.Command)
		}
		if server.Transport == mcpTransportHTTP && strings.TrimSpace(server.URL) == "" {
			out.Errors = append(out.Errors, "HTTP MCP 未配置 URL: "+s.ID)
		}
		if (s.Claude != server.EnabledInClaude) || (s.Codex != server.EnabledInCodex) {
			out.Changes = append(out.Changes, "MCP「"+server.Name+"」同步状态将更新")
		}
	}
	for _, s := range p.Hooks {
		h, ok := findHook(hooks, s.ID)
		if !ok {
			out.Errors = append(out.Errors, "Hook 不存在: "+s.ID)
		} else if s.Codex && !hookEventSupportedBy(h.Event, hookTargetCodex) {
			out.Errors = append(out.Errors, "Hook 不支持 Codex: "+h.Name)
		} else if s.Claude != h.EnabledInClaude || s.Codex != h.EnabledInCodex {
			out.Changes = append(out.Changes, "Hook「"+h.Name+"」同步状态将更新")
		}
	}
	out.CanApply = len(out.Errors) == 0
	return out
}

func (a *App) applyWorkspaceProfile(p config.AIWorkspaceProfile) error {
	if err := a.SetClaudeCodeConfig(p.ClaudeModel, p.ClaudeEffort); err != nil {
		return err
	}
	if err := a.SetCodexModelConfig(p.CodexModel, p.CodexEffort, p.CodexGoals); err != nil {
		return err
	}
	skills, err := a.GetSkills()
	if err != nil {
		return err
	}
	for _, s := range skills {
		desired := selectionFor(p.Skills, s.Name)
		if desired.Claude != s.InClaude {
			if !desired.Claude && !a.skillTargetMatchesMaster(s.Name, skillTargetClaude) {
				continue
			}
			if err := a.SetSkillSync(s.Name, skillTargetClaude, desired.Claude); err != nil {
				return err
			}
		}
		if desired.Codex != s.InCodex {
			if !desired.Codex && !a.skillTargetMatchesMaster(s.Name, skillTargetCodex) {
				continue
			}
			if err := a.SetSkillSync(s.Name, skillTargetCodex, desired.Codex); err != nil {
				return err
			}
		}
	}
	a.mu.RLock()
	mcps := append([]config.MCPServer(nil), a.cfg.MCPServers...)
	hooks := cloneHooks(a.cfg.Hooks)
	a.mu.RUnlock()
	for _, s := range mcps {
		desired := selectionFor(p.MCPServers, s.ID)
		if desired.Claude != s.EnabledInClaude {
			if err := a.SetMCPEnabled(s.ID, mcpTargetClaude, desired.Claude); err != nil {
				return err
			}
		}
		if desired.Codex != s.EnabledInCodex {
			if err := a.SetMCPEnabled(s.ID, mcpTargetCodex, desired.Codex); err != nil {
				return err
			}
		}
	}
	for _, h := range hooks {
		desired := selectionFor(p.Hooks, h.ID)
		if desired.Claude != h.EnabledInClaude {
			if err := a.SetHookEnabled(h.ID, hookTargetClaude, desired.Claude); err != nil {
				return err
			}
		}
		if desired.Codex != h.EnabledInCodex {
			if err := a.SetHookEnabled(h.ID, hookTargetCodex, desired.Codex); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) captureWorkspaceSnapshot(p config.AIWorkspaceProfile) (workspaceSnapshot, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return workspaceSnapshot{}, err
	}
	snap := workspaceSnapshot{AIWorkspaceSnapshotInfo: AIWorkspaceSnapshotInfo{ID: fmt.Sprintf("snapshot-%x", time.Now().UnixNano()), ProfileID: p.ID, ProfileName: p.Name, CreatedAt: time.Now().Unix()}}
	for _, path := range []string{filepath.Join(home, ".claude", "settings.json"), filepath.Join(home, ".claude.json"), filepath.Join(home, ".codex", "config.toml"), filepath.Join(home, ".codex", "hooks.json")} {
		f, err := snapshotFile(path)
		if err != nil {
			return snap, err
		}
		snap.Files = append(snap.Files, f)
	}
	skills, err := a.GetSkills()
	if err != nil {
		return snap, err
	}
	for _, s := range skills {
		for _, target := range []string{skillTargetClaude, skillTargetCodex} {
			dir, _ := workspaceSkillTarget(s.Name, target)
			d, err := snapshotDir(dir)
			if err != nil {
				return snap, err
			}
			snap.SkillDirs = append(snap.SkillDirs, d)
		}
	}
	a.mu.RLock()
	snap.LastProfile = a.cfg.LastAppliedWorkspaceID
	for _, s := range a.cfg.MCPServers {
		snap.MCPStates = append(snap.MCPStates, config.AgentResourceSelection{ID: s.ID, Claude: s.EnabledInClaude, Codex: s.EnabledInCodex})
	}
	for _, h := range a.cfg.Hooks {
		snap.HookStates = append(snap.HookStates, config.AgentResourceSelection{ID: h.ID, Claude: h.EnabledInClaude, Codex: h.EnabledInCodex})
	}
	a.mu.RUnlock()
	return snap, nil
}

func (a *App) restoreWorkspaceSnapshot(s workspaceSnapshot) error {
	for _, f := range s.Files {
		if err := restoreFile(f); err != nil {
			return err
		}
	}
	for _, d := range s.SkillDirs {
		if err := restoreDir(d); err != nil {
			return err
		}
	}
	a.mu.Lock()
	for i := range a.cfg.MCPServers {
		state := selectionFor(s.MCPStates, a.cfg.MCPServers[i].ID)
		a.cfg.MCPServers[i].EnabledInClaude, a.cfg.MCPServers[i].EnabledInCodex = state.Claude, state.Codex
	}
	for i := range a.cfg.Hooks {
		state := selectionFor(s.HookStates, a.cfg.Hooks[i].ID)
		a.cfg.Hooks[i].EnabledInClaude, a.cfg.Hooks[i].EnabledInCodex = state.Claude, state.Codex
	}
	err := a.cfg.Save()
	a.mu.Unlock()
	return err
}

func (a *App) workspaceProfile(id string) (config.AIWorkspaceProfile, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, p := range a.cfg.AIWorkspaceProfiles {
		if p.ID == id {
			return p, nil
		}
	}
	return config.AIWorkspaceProfile{}, fmt.Errorf("工作区画像不存在")
}
func parseTOML(content string) (interface{}, error) { return toml.Load(content) }
func firstExecutableToken(command string) string {
	command = strings.TrimSpace(command)
	if command == "" || strings.ContainsAny(command, "|;&$`()") {
		return ""
	}
	return strings.Fields(command)[0]
}
func validWorkspaceID(id string) bool {
	if !(strings.HasPrefix(id, "ws-") || strings.HasPrefix(id, "snapshot-")) || len(id) > 80 {
		return false
	}
	for _, r := range id {
		if !(r == '-' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}
func normalizeSelections(in []config.AgentResourceSelection) []config.AgentResourceSelection {
	seen := map[string]bool{}
	var out []config.AgentResourceSelection
	for _, s := range in {
		s.ID = strings.TrimSpace(s.ID)
		if s.ID != "" && !seen[s.ID] && (s.Claude || s.Codex) {
			seen[s.ID] = true
			out = append(out, s)
		}
	}
	return out
}
func selectionFor(in []config.AgentResourceSelection, id string) config.AgentResourceSelection {
	for _, s := range in {
		if s.ID == id {
			return s
		}
	}
	return config.AgentResourceSelection{ID: id}
}
func findMCP(in []config.MCPServer, id string) (config.MCPServer, bool) {
	for _, x := range in {
		if x.ID == id {
			return x, true
		}
	}
	return config.MCPServer{}, false
}
func findHook(in []config.HookDef, id string) (config.HookDef, bool) {
	for _, x := range in {
		if x.ID == id {
			return x, true
		}
	}
	return config.HookDef{}, false
}
func (a *App) checkWorkspaceSkill(s config.AgentResourceSelection) error {
	master, err := a.masterSkillsDir()
	if err != nil {
		return err
	}
	if !dirExists(filepath.Join(master, s.ID)) {
		return fmt.Errorf("Skill 不存在: %s", s.ID)
	}
	return nil
}

func (a *App) skillTargetMatchesMaster(name, target string) bool {
	master, err := a.masterSkillsDir()
	if err != nil {
		return false
	}
	dst, err := workspaceSkillTarget(name, target)
	if err != nil {
		return false
	}
	if !dirExists(dst) {
		return true
	}
	left, _, err := skillDirFingerprint(filepath.Join(master, name))
	if err != nil {
		return false
	}
	right, _, err := skillDirFingerprint(dst)
	return err == nil && left == right
}

func workspaceSkillTarget(name, target string) (string, error) {
	if target == skillTargetClaude {
		base, e := claudeSkillsDir()
		return filepath.Join(base, name), e
	}
	base, e := codexSkillsDir()
	return filepath.Join(base, name), e
}
func snapshotFile(path string) (workspaceFileSnapshot, error) {
	f := workspaceFileSnapshot{Path: path}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return f, nil
	}
	if err != nil {
		return f, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return f, err
	}
	f.Exists = true
	f.Mode = uint32(info.Mode())
	f.Data = base64.StdEncoding.EncodeToString(data)
	return f, nil
}
func restoreFile(f workspaceFileSnapshot) error {
	if !f.Exists {
		if err := os.Remove(f.Path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(f.Data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(f.Path, data, os.FileMode(f.Mode))
}
func snapshotDir(path string) (workspaceDirSnapshot, error) {
	d := workspaceDirSnapshot{Path: path}
	if !dirExists(path) {
		return d, nil
	}
	d.Exists = true
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := snapshotFile(p)
		if err != nil {
			return err
		}
		d.Entries = append(d.Entries, f)
		return nil
	})
	return d, err
}
func restoreDir(d workspaceDirSnapshot) error {
	if err := os.RemoveAll(d.Path); err != nil {
		return err
	}
	if !d.Exists {
		return nil
	}
	if err := os.MkdirAll(d.Path, 0o755); err != nil {
		return err
	}
	for _, f := range d.Entries {
		if err := restoreFile(f); err != nil {
			return err
		}
	}
	return nil
}
func (a *App) workspaceSnapshotDir() (string, error) {
	dir := mustWorkspaceSnapshotDir()
	return dir, os.MkdirAll(dir, 0o700)
}
func mustWorkspaceSnapshotDir() string {
	base, _ := os.UserConfigDir()
	return filepath.Join(base, "Hatch", "ai-workspace-snapshots")
}
func (a *App) writeWorkspaceSnapshot(s workspaceSnapshot) error {
	dir, err := a.workspaceSnapshotDir()
	if err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, s.ID+".json"), data, 0o600)
}
func (a *App) pruneWorkspaceSnapshots() error {
	list := a.ListAIWorkspaceSnapshots()
	for i := maxWorkspaceSnapshots; i < len(list); i++ {
		_ = os.Remove(filepath.Join(mustWorkspaceSnapshotDir(), list[i].ID+".json"))
	}
	return nil
}
