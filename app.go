package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"hatch/internal/autostart"
	"hatch/internal/config"
	"hatch/internal/executor"
	"hatch/internal/scanner"
	"hatch/internal/store"
	"hatch/internal/zshrc"
)

type App struct {
	ctx      context.Context
	cfg      *config.Config
	store    *store.Store
	services map[string]*executor.ServiceManager
	mu       sync.RWMutex
}

// ZshrcInfo is returned by CheckZshrc.
type ZshrcInfo struct {
	Shell      string   `json:"shell"`
	IsZsh      bool     `json:"isZsh"`
	ZshrcPath  string   `json:"zshrcPath"`
	HatchLines []string `json:"hatchLines"`
}

type AIScriptSuggestion struct {
	Name      string                 `json:"name"`
	Desc      string                 `json:"desc"`
	Type      config.ScriptType      `json:"type"`
	Commands  []string               `json:"commands"`
	PidFile   string                 `json:"pidFile"`
	NeedsArgs bool                   `json:"needsArgs"`
	Tags      []string               `json:"tags"`
	Presets   []config.CommandPreset `json:"presets"`
}

type ZshrcOrganizationSuggestion struct {
	Summary []string `json:"summary"`
	Content string   `json:"content"`
}

type ZshrcApplyResult struct {
	ZshrcPath  string `json:"zshrcPath"`
	BackupPath string `json:"backupPath"`
}

func NewApp() *App {
	return &App{
		services: make(map[string]*executor.ServiceManager),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{OverrideMeta: map[string]config.ScriptMeta{}}
	}
	a.cfg = cfg

	st, err := store.New()
	if err != nil {
		st, _ = store.NewWithPath("/tmp/hatch-scripts.json")
	}
	a.store = st

	a.restoreServices()
}

// restoreServices rebuilds ServiceManager instances for service-type scripts
// found in the store, and probes stale PID files on app start.
func (a *App) restoreServices() {
	for _, s := range a.store.All() {
		if s.Type == config.TypeService {
			sm := executor.NewServiceManager(s.Path, s.PidFile, executor.NewEmitter(a.ctx))
			a.mu.Lock()
			a.services[s.Path] = sm
			a.mu.Unlock()
			// Probe liveness — cleans up stale pidfiles automatically.
			sm.Status()
		}
	}
}

// ─── Config & Directories ────────────────────────────────────────────────────

// GetConfig returns the current application configuration.
func (a *App) GetConfig() *config.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

// AddDirectory adds a directory to the config and optionally writes an export
// to ~/.zshrc.
func (a *App) AddDirectory(path, exportName string, isScriptDir bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, d := range a.cfg.Directories {
		if d.Path == path {
			return fmt.Errorf("directory already added: %s", path)
		}
	}
	a.cfg.Directories = append(a.cfg.Directories, config.Directory{
		Path:        path,
		IsScriptDir: isScriptDir,
		ExportName:  exportName,
	})

	if exportName != "" {
		found := false
		for i, e := range a.cfg.Exports {
			if e.Name == exportName {
				a.cfg.Exports[i].Value = path
				found = true
				break
			}
		}
		if !found {
			a.cfg.Exports = append(a.cfg.Exports, config.ExportVar{Name: exportName, Value: path})
		}
		if err := a.syncZshrc(); err != nil {
			return err
		}
	}
	return a.cfg.Save()
}

// RemoveDirectory removes a directory (and its export) from config.
func (a *App) RemoveDirectory(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var kept []config.Directory
	removedExport := ""
	for _, d := range a.cfg.Directories {
		if d.Path == path {
			removedExport = d.ExportName
		} else {
			kept = append(kept, d)
		}
	}
	a.cfg.Directories = kept

	if removedExport != "" {
		var exports []config.ExportVar
		for _, e := range a.cfg.Exports {
			if e.Name != removedExport {
				exports = append(exports, e)
			}
		}
		a.cfg.Exports = exports
		if err := a.syncZshrc(); err != nil {
			return err
		}
	}
	return a.cfg.Save()
}

// UpdateDirectoryExport sets (or clears) the export name for an existing directory
// and syncs ~/.zshrc immediately.
func (a *App) UpdateDirectoryExport(dirPath, exportName string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	found := false
	for i, d := range a.cfg.Directories {
		if d.Path != dirPath {
			continue
		}
		found = true
		// Remove old export entry if name changed
		if d.ExportName != "" && d.ExportName != exportName {
			var kept []config.ExportVar
			for _, e := range a.cfg.Exports {
				if e.Name != d.ExportName {
					kept = append(kept, e)
				}
			}
			a.cfg.Exports = kept
		}
		a.cfg.Directories[i].ExportName = exportName
		break
	}
	if !found {
		return fmt.Errorf("directory not found: %s", dirPath)
	}

	if exportName != "" {
		updated := false
		for i, e := range a.cfg.Exports {
			if e.Name == exportName {
				a.cfg.Exports[i].Value = dirPath
				updated = true
				break
			}
		}
		if !updated {
			a.cfg.Exports = append(a.cfg.Exports, config.ExportVar{Name: exportName, Value: dirPath})
		}
		if err := a.syncZshrc(); err != nil {
			return err
		}
	}
	return a.cfg.Save()
}

// ─── Script scanning ─────────────────────────────────────────────────────────

// ScanAll rescans every configured script directory and returns the full list.
func (a *App) ScanAll() ([]config.ScriptMeta, error) {
	a.mu.RLock()
	dirs := make([]config.Directory, len(a.cfg.Directories))
	copy(dirs, a.cfg.Directories)
	overrides := a.cfg.OverrideMeta
	a.mu.RUnlock()

	var all []config.ScriptMeta
	for _, d := range dirs {
		if !d.IsScriptDir {
			continue
		}
		scripts, err := scanner.ScanDir(d.Path, overrides)
		if err != nil {
			continue
		}
		all = append(all, scripts...)
	}

	// Ensure a ServiceManager exists for every service-type script.
	a.mu.Lock()
	for _, s := range all {
		if s.Type == config.TypeService {
			if _, ok := a.services[s.Path]; !ok {
				a.services[s.Path] = executor.NewServiceManager(
					s.Path, s.PidFile, executor.NewEmitter(a.ctx),
				)
			}
		}
	}
	a.mu.Unlock()

	if a.store != nil {
		_ = a.store.Replace(all)
	}
	return all, nil
}

// GetScriptsByDir scans a single directory and returns its scripts.
func (a *App) GetScriptsByDir(dirPath string) ([]config.ScriptMeta, error) {
	a.mu.RLock()
	overrides := a.cfg.OverrideMeta
	a.mu.RUnlock()

	scripts, err := scanner.ScanDir(dirPath, overrides)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	for _, s := range scripts {
		if s.Type == config.TypeService {
			if _, ok := a.services[s.Path]; !ok {
				a.services[s.Path] = executor.NewServiceManager(
					s.Path, s.PidFile, executor.NewEmitter(a.ctx),
				)
			}
		}
	}
	a.mu.Unlock()

	if a.store != nil {
		for _, s := range scripts {
			_ = a.store.UpdateMeta(s)
		}
	}
	return scripts, nil
}

// SaveScriptMeta stores user-edited metadata into config.OverrideMeta and
// persists it.  The override is applied on top of auto-detected metadata on
// every subsequent ScanDir / GetScriptsByDir call.
func (a *App) SaveScriptMeta(meta config.ScriptMeta) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cfg.OverrideMeta == nil {
		a.cfg.OverrideMeta = make(map[string]config.ScriptMeta)
	}
	a.cfg.OverrideMeta[meta.Path] = meta
	return a.cfg.Save()
}

// DeleteScriptMeta removes a user override so auto-detection resumes.
func (a *App) DeleteScriptMeta(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.cfg.OverrideMeta, path)
	return a.cfg.Save()
}

// ─── Script execution ────────────────────────────────────────────────────────

// RunScript launches a oneshot script in a goroutine.
// Progress and result arrive via executor:log / executor:done / executor:status events.
func (a *App) RunScript(id, path string, args []string) {
	em := executor.NewEmitter(a.ctx)
	go func() {
		_, _ = executor.RunOneshot(id, path, args, em)
	}()
}

// ServiceAction executes start / stop / restart on a service script.
func (a *App) ServiceAction(path, action string, args []string) error {
	a.mu.Lock()
	sm, ok := a.services[path]
	if !ok {
		pidFile := ""
		if a.store != nil {
			for _, s := range a.store.All() {
				if s.Path == path {
					pidFile = s.PidFile
					break
				}
			}
		}
		sm = executor.NewServiceManager(path, pidFile, executor.NewEmitter(a.ctx))
		a.services[path] = sm
	}
	a.mu.Unlock()

	switch action {
	case "start":
		return sm.Start(args)
	case "stop":
		return sm.Stop()
	case "restart":
		return sm.Restart(args)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// GetAllServiceStatuses returns path → "running"|"stopped" for all tracked services.
func (a *App) GetAllServiceStatuses() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]string, len(a.services))
	for path, sm := range a.services {
		status, _ := sm.Status()
		out[path] = string(status)
	}
	return out
}

// ─── Zshrc & exports ─────────────────────────────────────────────────────────

// SaveExports persists the exports list to config and rewrites the zshrc block.
func (a *App) SaveExports(exports []config.ExportVar) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.Exports = exports
	if err := a.syncZshrc(); err != nil {
		return err
	}
	return a.cfg.Save()
}

// CheckZshrc returns current shell info and the Hatch managed block contents.
func (a *App) CheckZshrc() (ZshrcInfo, error) {
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return ZshrcInfo{}, err
	}
	shell := zshrc.CurrentShell()
	lines, _ := zshrc.ReadBlock(p)
	return ZshrcInfo{
		Shell:      shell,
		IsZsh:      strings.HasSuffix(shell, "zsh"),
		ZshrcPath:  p,
		HatchLines: lines,
	}, nil
}

// ParseZshrc reads ~/.zshrc and returns a structured list of entries for display.
func (a *App) ParseZshrc() ([]zshrc.Entry, error) {
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	a.mu.RLock()
	ann := a.cfg.ZshrcAnnotations
	a.mu.RUnlock()
	return zshrc.ParseContent(string(data), ann), nil
}

// SaveZshrcAnnotation persists a user note for a zshrc entry key.
// Passing an empty note removes the annotation.
func (a *App) SaveZshrcAnnotation(key, note string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cfg.ZshrcAnnotations == nil {
		a.cfg.ZshrcAnnotations = make(map[string]string)
	}
	if note == "" {
		delete(a.cfg.ZshrcAnnotations, key)
	} else {
		a.cfg.ZshrcAnnotations[key] = note
	}
	return a.cfg.Save()
}

// SuggestZshrcOrganization asks AI to reorganize ~/.zshrc and returns a preview.
// It does not write any files; ApplyZshrcOrganization performs backup + write.
func (a *App) SuggestZshrcOrganization(mode string) (ZshrcOrganizationSuggestion, error) {
	a.mu.RLock()
	ai := a.cfg.AI
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		return ZshrcOrganizationSuggestion{}, err
	}
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return ZshrcOrganizationSuggestion{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return ZshrcOrganizationSuggestion{}, err
	}
	content := string(data)
	if len(content) > 24000 {
		content = content[:24000]
	}
	entries := zshrc.ParseContent(content, nil)
	system := `你是谨慎的 zsh 配置整理助手。请重排 ~/.zshrc，使其结构清晰，同时保持语义等价。
只输出 JSON 对象，不要 Markdown，不要解释。字段：
summary: 字符串数组，列出主要整理动作；
content: 完整的新 ~/.zshrc 文件内容。
硬性要求：
1. 不要删除、改写或伪造任何命令、路径、token、key、alias、函数体、source、eval、conda/nvm/pyenv/go/ruby/node 初始化语句。
2. 可以移动相关配置到同一分组，添加简短中文分组注释。
3. Hatch managed block 必须原样保留为一个连续块，不要拆开、不要改 block 内内容。
4. 函数体内部内容和行序必须保持原样。
5. 如果不确定某行用途，放到“其他/未分类”分组，保留原文。
6. 输出 content 必须是可直接写回的完整文件。`
	if strings.TrimSpace(mode) == "" {
		mode = "按功能模块优先分组；同类 export/PATH/alias/function/source 放在相近区域。"
	}
	userBytes, _ := json.Marshal(map[string]any{
		"mode":    mode,
		"path":    p,
		"entries": entries,
		"content": content,
	})
	raw, err := a.callAI(ai, system, string(userBytes))
	if err != nil {
		return ZshrcOrganizationSuggestion{}, err
	}
	var out ZshrcOrganizationSuggestion
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &out); err != nil {
		return ZshrcOrganizationSuggestion{}, fmt.Errorf("AI 返回无法解析: %w", err)
	}
	if strings.TrimSpace(out.Content) == "" {
		return ZshrcOrganizationSuggestion{}, fmt.Errorf("AI 未返回整理后的 zshrc 内容")
	}
	return out, nil
}

// ApplyZshrcOrganization backs up ~/.zshrc and writes the approved content.
func (a *App) ApplyZshrcOrganization(content string) (ZshrcApplyResult, error) {
	if strings.TrimSpace(content) == "" {
		return ZshrcApplyResult{}, fmt.Errorf("整理内容为空")
	}
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return ZshrcApplyResult{}, err
	}
	backup, err := zshrc.Backup(p)
	if err != nil {
		return ZshrcApplyResult{}, err
	}
	if err := zshrc.WriteContent(p, content); err != nil {
		return ZshrcApplyResult{}, err
	}
	return ZshrcApplyResult{ZshrcPath: p, BackupPath: backup}, nil
}

func (a *App) GetAutoLaunchEnabled() bool {
	return autostart.Enabled()
}

func (a *App) SetAutoLaunchEnabled(enabled bool) error {
	if err := autostart.SetEnabled(enabled); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.AutoLaunch = enabled
	err := a.cfg.Save()
	a.mu.Unlock()
	return err
}

// SaveAIConfig persists OpenAI-compatible model settings.
func (a *App) SaveAIConfig(ai config.AIConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.AI = ai
	return a.cfg.Save()
}

// TestAIConfig sends a tiny request to the configured OpenAI-compatible API.
func (a *App) TestAIConfig(ai config.AIConfig) (string, error) {
	if !ai.Enabled {
		ai.Enabled = true
	}
	return a.callAI(ai, "你是 Hatch 的配置检测助手。请只回复 pong。", "ping")
}

// SuggestScriptMeta reads a script and asks the configured model for metadata.
func (a *App) SuggestScriptMeta(meta config.ScriptMeta) (AIScriptSuggestion, error) {
	a.mu.RLock()
	ai := a.cfg.AI
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		return AIScriptSuggestion{}, err
	}

	data, err := os.ReadFile(meta.Path)
	if err != nil {
		return AIScriptSuggestion{}, err
	}
	content := string(data)
	if len(content) > 12000 {
		content = content[:12000]
	}

	system := `你是 macOS 脚本管理工具 Hatch 的脚本分析助手。请根据脚本内容生成适合直接展示给用户的元信息。
只输出一个 JSON 对象，不要 Markdown，不要解释。字段：
name: 简短中文显示名；
desc: 一句话说明脚本做什么；
type: "oneshot" 或 "service"；
commands: service 子命令数组，没有则 []；
pidFile: service 的 pid 文件建议，没有则 ""；
needsArgs: oneshot 运行时是否需要参数输入；
tags: 1-5 个短标签；
presets: 常用命令预设数组，元素为 {"name":"", "args":""}，没有可靠建议则 []。`
	userBytes, _ := json.Marshal(map[string]any{
		"path":     meta.Path,
		"name":     meta.Name,
		"desc":     meta.Desc,
		"type":     meta.Type,
		"language": meta.Language,
		"commands": meta.Commands,
		"pidFile":  meta.PidFile,
		"tags":     meta.Tags,
		"content":  content,
	})
	raw, err := a.callAI(ai, system, string(userBytes))
	if err != nil {
		return AIScriptSuggestion{}, err
	}
	var out AIScriptSuggestion
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &out); err != nil {
		return AIScriptSuggestion{}, fmt.Errorf("AI 返回无法解析: %w", err)
	}
	if out.Type != config.TypeService {
		out.Type = config.TypeOneshot
		out.Commands = nil
		out.PidFile = ""
	}
	return out, nil
}

// SuggestZshrcAnnotation asks the configured model for a short note for one zshrc entry.
func (a *App) SuggestZshrcAnnotation(entry zshrc.Entry) (string, error) {
	a.mu.RLock()
	ai := a.cfg.AI
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		return "", err
	}
	system := `你是 shell 配置解释助手。请为一条 zshrc 配置生成简短中文备注，说明它的作用、影响或风险。
只输出备注正文，不要 Markdown，不要加引号，控制在 60 个中文字符以内。`
	userBytes, _ := json.Marshal(entry)
	raw, err := a.callAI(ai, system, string(userBytes))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}

func validateAIConfig(ai config.AIConfig) error {
	if !ai.Enabled {
		return fmt.Errorf("AI 配置未启用")
	}
	if strings.TrimSpace(ai.BaseURL) == "" {
		return fmt.Errorf("请先配置 API Base URL")
	}
	if strings.TrimSpace(ai.APIKey) == "" {
		return fmt.Errorf("请先配置 API Key")
	}
	if strings.TrimSpace(ai.Model) == "" {
		return fmt.Errorf("请先配置模型名")
	}
	return nil
}

func (a *App) callAI(ai config.AIConfig, systemPrompt, userPrompt string) (string, error) {
	if err := validateAIConfig(ai); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(ai.BaseURL), "/")
	url := baseURL
	if !strings.HasSuffix(url, "/chat/completions") {
		url += "/chat/completions"
	}
	body := map[string]any{
		"model": strings.TrimSpace(ai.Model),
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(ai.APIKey))

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI 请求失败: HTTP %d %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if len(data) == 0 {
		return "", fmt.Errorf("AI 返回空响应 (HTTP %d)", resp.StatusCode)
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		preview := string(data)
		if len(preview) > 300 {
			preview = preview[:300] + "…"
		}
		return "", fmt.Errorf("解析 AI 响应失败: %w — 收到: %s", err, preview)
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("AI 返回为空")
	}
	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}

func extractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

// syncZshrc rebuilds the managed block from cfg.Exports. Caller must hold mu.
func (a *App) syncZshrc() error {
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return err
	}
	exports := make(map[string]string, len(a.cfg.Exports))
	for _, e := range a.cfg.Exports {
		if e.Name != "" {
			exports[e.Name] = e.Value
		}
	}
	if len(exports) == 0 {
		return zshrc.RemoveBlock(p)
	}
	return zshrc.WriteBlock(p, exports)
}
