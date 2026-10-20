package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hatch/ai/agent"
	"hatch/ai/hooks"
	"hatch/ai/mcp"
	"hatch/ai/skills"
	"hatch/ai/workspace"
	"hatch/core/config"
	"hatch/core/macos"
)

// ─── AI · 工作区画像 ─────────────────────────────────────────────────────────
//
// A profile is a project-scoped combination of agent model settings plus which
// Skills / MCP servers / Hooks are synced to Claude Code and Codex. Profiles
// are only applied manually: preview → snapshot → apply (rollback on failure).
// This file orchestrates the other AI modules; snapshot storage and data types
// live in internal/ai/workspace.

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
	if !workspace.ValidID(profile.ID) {
		return config.AIWorkspaceProfile{}, fmt.Errorf("invalid workspace profile id")
	}
	profile.Skills = workspace.NormalizeSelections(profile.Skills)
	profile.MCPServers = workspace.NormalizeSelections(profile.MCPServers)
	profile.Hooks = workspace.NormalizeSelections(profile.Hooks)
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
	if !workspace.ValidID(id) {
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

// CaptureCurrentAIWorkspaceProfile builds and saves a profile from the current
// Hatch-managed state.
func (a *App) CaptureCurrentAIWorkspaceProfile(name, rootPath string) (config.AIWorkspaceProfile, error) {
	profile := config.AIWorkspaceProfile{Name: name, RootPath: rootPath}
	claude := agent.ReadClaudeCodeConfig()
	codex := agent.ReadCodexModelConfig()
	profile.ClaudeModel, profile.ClaudeEffort = claude.Model, claude.EffortLevel
	profile.CodexModel, profile.CodexEffort, profile.CodexGoals = codex.Model, codex.EffortLevel, codex.GoalsEnabled
	list, err := a.GetSkills()
	if err != nil {
		return profile, err
	}
	for _, s := range list {
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

func (a *App) PreviewAIWorkspaceProfileApply(id string) (workspace.Preview, error) {
	p, err := a.workspaceProfile(id)
	if err != nil {
		return workspace.Preview{}, err
	}
	return a.previewWorkspaceProfile(p), nil
}

// ApplyAIWorkspaceProfile snapshots current state, applies the profile and
// rolls back to the snapshot if any step fails.
func (a *App) ApplyAIWorkspaceProfile(id string) (workspace.SnapshotInfo, error) {
	p, err := a.workspaceProfile(id)
	if err != nil {
		return workspace.SnapshotInfo{}, err
	}
	preview := a.previewWorkspaceProfile(p)
	if !preview.CanApply {
		return workspace.SnapshotInfo{}, fmt.Errorf("无法应用工作区画像: %s", strings.Join(preview.Errors, "；"))
	}
	snap, err := a.captureWorkspaceSnapshot(p)
	if err != nil {
		return workspace.SnapshotInfo{}, err
	}
	if err := workspace.WriteSnapshot(snap); err != nil {
		return workspace.SnapshotInfo{}, err
	}
	if err := a.applyWorkspaceProfile(p); err != nil {
		_ = a.restoreWorkspaceSnapshot(snap)
		return workspace.SnapshotInfo{}, fmt.Errorf("应用失败，已恢复应用前状态: %w", err)
	}
	a.mu.Lock()
	a.cfg.LastAppliedWorkspaceID = p.ID
	err = a.cfg.Save()
	a.mu.Unlock()
	if err != nil {
		return workspace.SnapshotInfo{}, err
	}
	workspace.PruneSnapshots()
	return snap.SnapshotInfo, nil
}

func (a *App) ListAIWorkspaceSnapshots() []workspace.SnapshotInfo {
	return workspace.ListSnapshots()
}

func (a *App) RollbackAIWorkspaceSnapshot(id string) error {
	if !workspace.ValidID(id) {
		return fmt.Errorf("invalid workspace snapshot id")
	}
	snap, err := workspace.ReadSnapshot(id)
	if err != nil {
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

// GetAIEnvironmentHealth checks agent CLIs, config file syntax, commands used
// by synced MCP servers/hooks, and drift from the last applied profile.
func (a *App) GetAIEnvironmentHealth() workspace.Health {
	health := workspace.Health{Overall: "ok", Checks: []workspace.HealthCheck{}}
	if macos.FindCLI("claude") == "" {
		health.Add(workspace.HealthCheck{ID: "claude-cli", Label: "Claude Code CLI", Status: "warn", Detail: "未在登录 shell PATH 中找到 claude", Suggestion: "安装 Claude Code 或检查 ~/.zshrc PATH"})
	} else {
		health.Add(workspace.HealthCheck{ID: "claude-cli", Label: "Claude Code CLI", Status: "ok", Detail: "已可用"})
	}
	if macos.FindCLI("codex") == "" {
		health.Add(workspace.HealthCheck{ID: "codex-cli", Label: "Codex CLI", Status: "warn", Detail: "未在登录 shell PATH 中找到 codex", Suggestion: "安装 Codex 或检查 ~/.zshrc PATH"})
	} else {
		health.Add(workspace.HealthCheck{ID: "codex-cli", Label: "Codex CLI", Status: "ok", Detail: "已可用"})
	}
	if p, err := agent.ClaudeSettingsPath(); err == nil {
		health.AddJSONCheck(p, "claude-settings", "Claude 设置")
	}
	if p, err := agent.ClaudeJSONPath(); err == nil {
		health.AddJSONCheck(p, "claude-mcp", "Claude MCP 配置")
	}
	if p, err := agent.CodexHooksPath(); err == nil {
		health.AddJSONCheck(p, "codex-hooks", "Codex Hooks 配置")
	}
	if p, err := agent.CodexConfigPath(); err == nil {
		if raw, err := os.ReadFile(p); err == nil {
			if agent.ValidTOML(string(raw)) != nil {
				health.Add(workspace.HealthCheck{ID: "codex-config", Label: "Codex 配置", Status: "error", Detail: "config.toml 语法无效", Suggestion: "在助手页修复 TOML 后再应用画像"})
			} else {
				health.Add(workspace.HealthCheck{ID: "codex-config", Label: "Codex 配置", Status: "ok", Detail: "TOML 有效"})
			}
		}
	}
	a.mu.RLock()
	managedMCP := append([]config.MCPServer(nil), a.cfg.MCPServers...)
	managedHooks := hooks.Clone(a.cfg.Hooks)
	last := a.cfg.LastAppliedWorkspaceID
	a.mu.RUnlock()
	for _, server := range managedMCP {
		if !(server.EnabledInClaude || server.EnabledInCodex) || server.Transport != mcp.TransportStdio {
			continue
		}
		if macos.FindCLI(server.Command) == "" {
			health.Add(workspace.HealthCheck{ID: "mcp-" + server.ID, Label: "MCP: " + server.Name, Status: "warn", Detail: "命令不可用: " + server.Command, Suggestion: "安装命令或在 MCP 中改为可执行入口"})
		}
	}
	for _, hook := range managedHooks {
		if !(hook.EnabledInClaude || hook.EnabledInCodex) {
			continue
		}
		command := workspace.FirstExecutableToken(hook.Command)
		if command != "" && macos.FindCLI(command) == "" {
			health.Add(workspace.HealthCheck{ID: "hook-" + hook.ID, Label: "Hook: " + hook.Name, Status: "warn", Detail: "命令可能不可用: " + command, Suggestion: "检查 Hook 命令或登录 shell PATH"})
		}
	}
	if last == "" {
		health.ProfileState = "none"
		return health
	}
	p, err := a.workspaceProfile(last)
	if err != nil {
		health.ProfileState = "missing"
		health.Add(workspace.HealthCheck{ID: "workspace-profile", Label: "工作区画像", Status: "warn", Detail: "最近应用的画像已被删除"})
		return health
	}
	health.CurrentProfile = p.Name
	if preview := a.previewWorkspaceProfile(p); len(preview.Changes) > 0 {
		health.ProfileState = "drifted"
		health.Add(workspace.HealthCheck{ID: "workspace-profile", Label: "工作区画像", Status: "warn", Detail: "当前全局配置与「" + p.Name + "」不一致", Suggestion: "打开工作区页预览并重新应用，或回滚快照"})
	} else {
		health.ProfileState = "applied"
		health.Add(workspace.HealthCheck{ID: "workspace-profile", Label: "工作区画像", Status: "ok", Detail: "「" + p.Name + "」仍与当前 Hatch 管理状态一致"})
	}
	return health
}

// previewWorkspaceProfile dry-runs applying p and lists changes, blocking
// errors (e.g. overwriting a non-Hatch skill) and warnings.
func (a *App) previewWorkspaceProfile(p config.AIWorkspaceProfile) workspace.Preview {
	out := workspace.Preview{Profile: p, CanApply: true}
	if fi, err := os.Stat(p.RootPath); err != nil || !fi.IsDir() {
		out.Errors = append(out.Errors, "工作区目录不存在或不是目录: "+p.RootPath)
	}
	claude, codex := agent.ReadClaudeCodeConfig(), agent.ReadCodexModelConfig()
	if claude.Model != p.ClaudeModel || claude.EffortLevel != p.ClaudeEffort {
		out.Changes = append(out.Changes, "Claude Code 模型或推理档位将更新")
	}
	if codex.Model != p.CodexModel || codex.EffortLevel != p.CodexEffort || codex.GoalsEnabled != p.CodexGoals {
		out.Changes = append(out.Changes, "Codex 模型、推理档位或 Goals 将更新")
	}
	a.mu.RLock()
	servers := append([]config.MCPServer(nil), a.cfg.MCPServers...)
	hookList := hooks.Clone(a.cfg.Hooks)
	a.mu.RUnlock()

	if st, err := a.skillStore(); err != nil {
		out.Errors = append(out.Errors, err.Error())
	} else {
		for _, s := range p.Skills {
			if !st.Exists(s.ID) {
				out.Errors = append(out.Errors, "Skill 不存在: "+s.ID)
			}
			if s.Claude && !st.TargetMatchesMaster(s.ID, skills.TargetClaude) {
				out.Errors = append(out.Errors, "Claude 中同名 Skill 与 Hatch 主仓不同，拒绝覆盖: "+s.ID)
			}
			if s.Codex && !st.TargetMatchesMaster(s.ID, skills.TargetCodex) {
				out.Errors = append(out.Errors, "Codex 中同名 Skill 与 Hatch 主仓不同，拒绝覆盖: "+s.ID)
			}
		}
		if list, err := st.List(); err == nil {
			for _, skill := range list {
				desired := workspace.SelectionFor(p.Skills, skill.Name)
				if skill.InClaude && !desired.Claude && !st.TargetMatchesMaster(skill.Name, skills.TargetClaude) {
					out.Warnings = append(out.Warnings, "Skill「"+skill.Name+"」在 Claude 中不是 Hatch 当前版本，将保留不变")
				}
				if skill.InCodex && !desired.Codex && !st.TargetMatchesMaster(skill.Name, skills.TargetCodex) {
					out.Warnings = append(out.Warnings, "Skill「"+skill.Name+"」在 Codex 中不是 Hatch 当前版本，将保留不变")
				}
			}
		}
	}
	for _, s := range p.MCPServers {
		server, ok := mcp.Find(servers, s.ID)
		if !ok {
			out.Errors = append(out.Errors, "MCP 不存在: "+s.ID)
			continue
		}
		if server.Transport == mcp.TransportStdio && strings.TrimSpace(server.Command) == "" {
			out.Errors = append(out.Errors, "MCP 未配置命令: "+s.ID)
		} else if server.Transport == mcp.TransportStdio && macos.FindCLI(server.Command) == "" {
			out.Warnings = append(out.Warnings, "MCP 命令暂不可用: "+server.Command)
		}
		if server.Transport == mcp.TransportHTTP && strings.TrimSpace(server.URL) == "" {
			out.Errors = append(out.Errors, "HTTP MCP 未配置 URL: "+s.ID)
		}
		if (s.Claude != server.EnabledInClaude) || (s.Codex != server.EnabledInCodex) {
			out.Changes = append(out.Changes, "MCP「"+server.Name+"」同步状态将更新")
		}
	}
	for _, s := range p.Hooks {
		h, ok := hooks.Find(hookList, s.ID)
		if !ok {
			out.Errors = append(out.Errors, "Hook 不存在: "+s.ID)
		} else if s.Codex && !hooks.SupportedBy(h.Event, hooks.TargetCodex) {
			out.Errors = append(out.Errors, "Hook 不支持 Codex: "+h.Name)
		} else if s.Claude != h.EnabledInClaude || s.Codex != h.EnabledInCodex {
			out.Changes = append(out.Changes, "Hook「"+h.Name+"」同步状态将更新")
		}
	}
	out.CanApply = len(out.Errors) == 0
	return out
}

// applyWorkspaceProfile writes agent model settings, then toggles Skill / MCP /
// Hook sync to match p. Skills that differ from the master copy are never removed.
func (a *App) applyWorkspaceProfile(p config.AIWorkspaceProfile) error {
	if err := agent.WriteClaudeCodeConfig(p.ClaudeModel, p.ClaudeEffort); err != nil {
		return err
	}
	if err := agent.WriteCodexModelConfig(p.CodexModel, p.CodexEffort, p.CodexGoals); err != nil {
		return err
	}
	st, err := a.skillStore()
	if err != nil {
		return err
	}
	list, err := st.List()
	if err != nil {
		return err
	}
	for _, s := range list {
		desired := workspace.SelectionFor(p.Skills, s.Name)
		for _, t := range []struct {
			target  string
			current bool
			want    bool
		}{
			{skills.TargetClaude, s.InClaude, desired.Claude},
			{skills.TargetCodex, s.InCodex, desired.Codex},
		} {
			if t.want == t.current {
				continue
			}
			if !t.want && !st.TargetMatchesMaster(s.Name, t.target) {
				continue
			}
			if err := st.SetSync(s.Name, t.target, t.want); err != nil {
				return err
			}
		}
	}
	a.mu.RLock()
	servers := append([]config.MCPServer(nil), a.cfg.MCPServers...)
	hookList := hooks.Clone(a.cfg.Hooks)
	a.mu.RUnlock()
	for _, s := range servers {
		desired := workspace.SelectionFor(p.MCPServers, s.ID)
		if desired.Claude != s.EnabledInClaude {
			if err := a.SetMCPEnabled(s.ID, mcp.TargetClaude, desired.Claude); err != nil {
				return err
			}
		}
		if desired.Codex != s.EnabledInCodex {
			if err := a.SetMCPEnabled(s.ID, mcp.TargetCodex, desired.Codex); err != nil {
				return err
			}
		}
	}
	for _, h := range hookList {
		desired := workspace.SelectionFor(p.Hooks, h.ID)
		if desired.Claude != h.EnabledInClaude {
			if err := a.SetHookEnabled(h.ID, hooks.TargetClaude, desired.Claude); err != nil {
				return err
			}
		}
		if desired.Codex != h.EnabledInCodex {
			if err := a.SetHookEnabled(h.ID, hooks.TargetCodex, desired.Codex); err != nil {
				return err
			}
		}
	}
	return nil
}

// captureWorkspaceSnapshot records every agent config file, every synced skill
// dir and the registry sync flags before an apply.
func (a *App) captureWorkspaceSnapshot(p config.AIWorkspaceProfile) (workspace.Snapshot, error) {
	now := time.Now()
	snap := workspace.Snapshot{SnapshotInfo: workspace.SnapshotInfo{
		ID: fmt.Sprintf("snapshot-%x", now.UnixNano()), ProfileID: p.ID, ProfileName: p.Name, CreatedAt: now.Unix(),
	}}
	files, err := agent.ManagedConfigFiles()
	if err != nil {
		return snap, err
	}
	for _, path := range files {
		f, err := workspace.SnapshotFile(path)
		if err != nil {
			return snap, err
		}
		snap.Files = append(snap.Files, f)
	}
	list, err := a.GetSkills()
	if err != nil {
		return snap, err
	}
	for _, s := range list {
		for _, target := range []string{skills.TargetClaude, skills.TargetCodex} {
			base, _ := skills.TargetDir(target)
			d, err := workspace.SnapshotDir(filepath.Join(base, s.Name))
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

// restoreWorkspaceSnapshot writes back captured files/dirs and sync flags.
func (a *App) restoreWorkspaceSnapshot(s workspace.Snapshot) error {
	for _, f := range s.Files {
		if err := workspace.RestoreFile(f); err != nil {
			return err
		}
	}
	for _, d := range s.SkillDirs {
		if err := workspace.RestoreDir(d); err != nil {
			return err
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.cfg.MCPServers {
		state := workspace.SelectionFor(s.MCPStates, a.cfg.MCPServers[i].ID)
		a.cfg.MCPServers[i].EnabledInClaude, a.cfg.MCPServers[i].EnabledInCodex = state.Claude, state.Codex
	}
	for i := range a.cfg.Hooks {
		state := workspace.SelectionFor(s.HookStates, a.cfg.Hooks[i].ID)
		a.cfg.Hooks[i].EnabledInClaude, a.cfg.Hooks[i].EnabledInCodex = state.Claude, state.Codex
	}
	return a.cfg.Save()
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
