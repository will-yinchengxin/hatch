package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hatch/internal/autostart"
	"hatch/internal/config"
	"hatch/internal/executor"
	"hatch/internal/scanner"
	"hatch/internal/store"
	"hatch/internal/zshrc"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	cfg      *config.Config
	store    *store.Store
	services map[string]*executor.ServiceManager
	mu       sync.RWMutex

	tracesMu sync.Mutex
	traces   []AICallTrace
	traceSeq int64
}

// AICallTrace records one AI API call for inspection in the frontend.
type AICallTrace struct {
	ID           int64  `json:"id"`
	At           string `json:"at"`
	Op           string `json:"op"`
	Model        string `json:"model"`
	URL          string `json:"url"`
	SystemPrompt string `json:"systemPrompt"`
	UserPrompt   string `json:"userPrompt"`
	RawResponse  string `json:"rawResponse"`
	Error        string `json:"error"`
	DurationMs   int64  `json:"durationMs"`
}

const maxTraces = 50

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

type ClaudeCodeStatus struct {
	ClaudeInstalled bool     `json:"claudeInstalled"`
	ClaudeVersion   string   `json:"claudeVersion"`
	VPNConnected    bool     `json:"vpnConnected"`
	VPNInterfaces   []string `json:"vpnInterfaces"`
	HasUsageData    bool     `json:"hasUsageData"`
	FiveHourPct     float64  `json:"fiveHourPct"`
	FiveHourReset   string   `json:"fiveHourReset"`
	SevenDayPct     float64  `json:"sevenDayPct"`
	SevenDayReset   string   `json:"sevenDayReset"`
}

type CodexUsageStatus struct {
	CodexInstalled bool             `json:"codexInstalled"`
	CodexVersion   string           `json:"codexVersion"`
	VPNConnected   bool             `json:"vpnConnected"`
	VPNInterfaces  []string         `json:"vpnInterfaces"`
	HasUsageData   bool             `json:"hasUsageData"`
	PlanType       string           `json:"planType"`
	LimitReached   string           `json:"limitReached"`
	SnapshotAt     string           `json:"snapshotAt"`
	Primary        CodexUsageWindow `json:"primary"`
	Secondary      CodexUsageWindow `json:"secondary"`
}

type CodexUsageWindow struct {
	Label         string  `json:"label"`
	UsedPercent   float64 `json:"usedPercent"`
	LeftPercent   float64 `json:"leftPercent"`
	WindowMinutes int     `json:"windowMinutes"`
	ResetsAt      int64   `json:"resetsAt"`
	ResetIn       string  `json:"resetIn"`
	Available     bool    `json:"available"`
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
	raw, err := a.callAIOp("SuggestZshrcOrganization", ai, system, string(userBytes))
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
	return a.callAIOp("TestAIConfig", ai, "你是 Hatch 的配置检测助手。请只回复 pong。", "ping")
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
	raw, err := a.callAIOp("SuggestScriptMeta", ai, system, string(userBytes))
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
	raw, err := a.callAIOp("SuggestZshrcAnnotation", ai, system, string(userBytes))
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

// GetAITraces returns the in-memory trace log of recent AI calls (newest first).
func (a *App) GetAITraces() []AICallTrace {
	a.tracesMu.Lock()
	defer a.tracesMu.Unlock()
	out := make([]AICallTrace, len(a.traces))
	copy(out, a.traces)
	return out
}

// ClearAITraces empties the in-memory AI call log.
func (a *App) ClearAITraces() {
	a.tracesMu.Lock()
	defer a.tracesMu.Unlock()
	a.traces = nil
}

func (a *App) callAI(ai config.AIConfig, systemPrompt, userPrompt string) (string, error) {
	return a.callAIOp("unknown", ai, systemPrompt, userPrompt)
}

func (a *App) callAIOp(op string, ai config.AIConfig, systemPrompt, userPrompt string) (string, error) {
	if err := validateAIConfig(ai); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(ai.BaseURL), "/")
	apiURL := baseURL
	if !strings.HasSuffix(apiURL, "/chat/completions") {
		apiURL += "/chat/completions"
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(ai.APIKey))

	start := time.Now()
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	durationMs := time.Since(start).Milliseconds()

	trace := AICallTrace{
		At:           time.Now().Format("15:04:05"),
		Op:           op,
		Model:        strings.TrimSpace(ai.Model),
		URL:          apiURL,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		DurationMs:   durationMs,
	}
	a.tracesMu.Lock()
	a.traceSeq++
	trace.ID = a.traceSeq
	a.tracesMu.Unlock()

	if err != nil {
		trace.Error = err.Error()
		a.recordTrace(trace)
		return "", err
	}

	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	trace.RawResponse = string(data)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		trace.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		a.recordTrace(trace)
		return "", fmt.Errorf("AI 请求失败: HTTP %d %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if len(data) == 0 {
		trace.Error = fmt.Sprintf("空响应 (HTTP %d)", resp.StatusCode)
		a.recordTrace(trace)
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
		trace.Error = fmt.Sprintf("解析失败: %v", err)
		a.recordTrace(trace)
		return "", fmt.Errorf("解析 AI 响应失败: %w — 收到: %s", err, preview)
	}
	if len(decoded.Choices) == 0 {
		trace.Error = "choices 为空"
		a.recordTrace(trace)
		return "", fmt.Errorf("AI 返回为空")
	}
	result := strings.TrimSpace(decoded.Choices[0].Message.Content)
	a.recordTrace(trace)
	return result, nil
}

func (a *App) recordTrace(t AICallTrace) {
	a.tracesMu.Lock()
	// Prepend newest first, cap at maxTraces
	a.traces = append([]AICallTrace{t}, a.traces...)
	if len(a.traces) > maxTraces {
		a.traces = a.traces[:maxTraces]
	}
	a.tracesMu.Unlock()
	// Notify frontend in real-time
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "ai:trace", t)
	}
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

// ─── Script file I/O ─────────────────────────────────────────────────────────

// ReadScriptFile returns the raw text content of a script file.
func (a *App) ReadScriptFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteScriptFile writes new content to an existing script file, ensures it is
// executable, and sources ~/.zshrc so the Hatch process picks up any env changes.
func (a *App) WriteScriptFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	a.autoSource()
	return nil
}

// CreateScriptFile creates a new script file inside dirPath, makes it
// executable, and sources ~/.zshrc.  Returns the full path of the created file.
func (a *App) CreateScriptFile(dirPath, filename, content string) (string, error) {
	fullPath := filepath.Join(dirPath, filename)
	if _, err := os.Stat(fullPath); err == nil {
		return "", fmt.Errorf("文件已存在: %s", fullPath)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0755); err != nil {
		return "", err
	}
	a.autoSource()
	return fullPath, nil
}

// SuggestScriptContent asks the configured AI to write or improve a script.
// It returns the raw script text (no Markdown fences).
func (a *App) SuggestScriptContent(path, currentContent, instruction string) (string, error) {
	a.mu.RLock()
	ai := a.cfg.AI
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		return "", err
	}
	system := "你是 macOS shell 脚本编写助手。根据用户指令直接输出完整的脚本内容。\n" +
		"要求：\n" +
		"1. 只输出脚本文本，不要任何解释，不要 Markdown 代码块（不要用三个反引号包裹输出）。\n" +
		"2. 脚本必须以合适的 shebang 开头（如 #!/bin/zsh 或 #!/usr/bin/env python3）。\n" +
		"3. 保留已有内容中有价值的逻辑，除非指令要求完全重写。"
	userBytes, _ := json.Marshal(map[string]any{
		"path":           path,
		"currentContent": currentContent,
		"instruction":    instruction,
	})
	raw, err := a.callAIOp("SuggestScriptContent", ai, system, string(userBytes))
	if err != nil {
		return "", err
	}
	// Strip markdown fences if the model adds them anyway
	raw = strings.TrimSpace(raw)
	for _, fence := range []string{"```zsh", "```shell", "```sh", "```bash", "```python", "```ruby", "```node", "```"} {
		if strings.HasPrefix(raw, fence) {
			raw = strings.TrimPrefix(raw, fence)
			// Remove the language tag's trailing newline
			raw = strings.TrimLeft(raw, " \t\r\n")
			break
		}
	}
	raw = strings.TrimSuffix(strings.TrimSpace(raw), "```")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("AI 返回了空内容（原始响应可在 AI 调用日志中查看）")
	}
	return raw, nil
}

// autoSource runs `source ~/.zshrc` in a background login shell so that any
// new exports/PATH entries take effect for scripts launched by Hatch.
// Errors are silently ignored — this is a best-effort operation.
func (a *App) autoSource() {
	go func() {
		_ = exec.Command("/bin/zsh", "-l", "-c", "source ~/.zshrc 2>/dev/null || true").Run()
	}()
}

// ─── Claude Code status ───────────────────────────────────────────────────────

// GetClaudeCodeStatus checks whether claude CLI is installed, whether a VPN is
// active, and reads cached rate-limit usage from the statusline cache file.
func (a *App) GetClaudeCodeStatus() ClaudeCodeStatus {
	st := ClaudeCodeStatus{}

	claudePath, err := exec.LookPath("claude")
	if err == nil && claudePath != "" {
		st.ClaudeInstalled = true
		out, _ := exec.Command(claudePath, "--version").Output()
		st.ClaudeVersion = strings.TrimSpace(string(out))
	}

	st.VPNConnected, st.VPNInterfaces = detectVPNInterfaces()

	home, _ := os.UserHomeDir()
	cachePath := home + "/.cache/will-statusline/last.json"
	data, err := os.ReadFile(cachePath)
	if err == nil {
		var cache struct {
			RateLimits struct {
				FiveHour struct {
					UsedPercentage float64 `json:"used_percentage"`
					ResetsAt       int64   `json:"resets_at"`
				} `json:"five_hour"`
				SevenDay struct {
					UsedPercentage float64 `json:"used_percentage"`
					ResetsAt       int64   `json:"resets_at"`
				} `json:"seven_day"`
			} `json:"rate_limits"`
		}
		if json.Unmarshal(data, &cache) == nil {
			st.HasUsageData = true
			st.FiveHourPct = cache.RateLimits.FiveHour.UsedPercentage
			st.FiveHourReset = formatResetCountdown(cache.RateLimits.FiveHour.ResetsAt)
			st.SevenDayPct = cache.RateLimits.SevenDay.UsedPercentage
			st.SevenDayReset = formatResetCountdown(cache.RateLimits.SevenDay.ResetsAt)
		}
	}
	return st
}

// ─── Codex usage status ───────────────────────────────────────────────────────

// GetCodexUsageStatus checks whether codex CLI is installed, whether a VPN is
// active, and reads the latest local Codex rate-limit snapshot.
func (a *App) GetCodexUsageStatus() CodexUsageStatus {
	st := CodexUsageStatus{}

	codexPath := findCodexCLI()
	if codexPath != "" {
		st.CodexInstalled = true
		out, _ := exec.Command(codexPath, "--version").Output()
		st.CodexVersion = strings.TrimSpace(string(out))
	}

	st.VPNConnected, st.VPNInterfaces = detectVPNInterfaces()

	home, err := os.UserHomeDir()
	if err != nil {
		return st
	}
	snapshot, ok := loadLatestCodexRateLimitSnapshot(filepath.Join(home, ".codex", "sessions"))
	if !ok {
		return st
	}

	st.HasUsageData = true
	st.PlanType = snapshot.PlanType
	st.LimitReached = snapshot.LimitReached
	st.SnapshotAt = snapshot.Timestamp
	st.Primary = codexWindowFromSnapshot("5h", snapshot.Primary)
	st.Secondary = codexWindowFromSnapshot("7d", snapshot.Secondary)
	return st
}

func findCodexCLI() string {
	if path, err := exec.LookPath("codex"); err == nil && path != "" {
		return path
	}
	out, err := exec.Command("/bin/zsh", "-l", "-c", "command -v codex").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func detectVPNInterfaces() (bool, []string) {
	ifOut, _ := exec.Command("ifconfig").Output()
	var currentIface string
	var interfaces []string
	for _, line := range strings.Split(string(ifOut), "\n") {
		if strings.HasPrefix(line, "utun") {
			currentIface = strings.SplitN(line, ":", 2)[0]
		} else if currentIface != "" && strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				interfaces = append(interfaces, currentIface+": "+fields[1])
			}
			currentIface = ""
		} else if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
			currentIface = ""
		}
	}
	return len(interfaces) > 0, interfaces
}

type codexRateLimitSnapshot struct {
	Timestamp    string
	PlanType     string
	LimitReached string
	Primary      codexRateLimitWindow
	Secondary    codexRateLimitWindow
}

type codexRateLimitWindow struct {
	UsedPercent   float64 `json:"used_percent"`
	WindowMinutes int     `json:"window_minutes"`
	ResetsAt      int64   `json:"resets_at"`
	Available     bool
}

func loadLatestCodexRateLimitSnapshot(sessionsDir string) (codexRateLimitSnapshot, bool) {
	var latest codexRateLimitSnapshot
	var found bool

	_ = filepath.WalkDir(sessionsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasPrefix(d.Name(), "rollout-") || !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			snapshot, ok := parseCodexRateLimitLine(line)
			if !ok {
				continue
			}
			if !found || snapshot.Timestamp > latest.Timestamp {
				latest = snapshot
				found = true
			}
		}
		return nil
	})

	return latest, found
}

func parseCodexRateLimitLine(line string) (codexRateLimitSnapshot, bool) {
	type rawWindow struct {
		UsedPercent   *float64 `json:"used_percent"`
		WindowMinutes *int     `json:"window_minutes"`
		ResetsAt      *int64   `json:"resets_at"`
	}
	var entry struct {
		Timestamp string `json:"timestamp"`
		Payload   struct {
			Type      string `json:"type"`
			RateLimit struct {
				PlanType     string    `json:"plan_type"`
				LimitReached string    `json:"rate_limit_reached_type"`
				Primary      rawWindow `json:"primary"`
				Secondary    rawWindow `json:"secondary"`
			} `json:"rate_limits"`
		} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return codexRateLimitSnapshot{}, false
	}
	if entry.Payload.Type != "token_count" {
		return codexRateLimitSnapshot{}, false
	}
	fromRaw := func(raw rawWindow) codexRateLimitWindow {
		window := codexRateLimitWindow{Available: raw.UsedPercent != nil || raw.WindowMinutes != nil || raw.ResetsAt != nil}
		if raw.UsedPercent != nil {
			window.UsedPercent = *raw.UsedPercent
		}
		if raw.WindowMinutes != nil {
			window.WindowMinutes = *raw.WindowMinutes
		}
		if raw.ResetsAt != nil {
			window.ResetsAt = *raw.ResetsAt
		}
		return window
	}
	primary := fromRaw(entry.Payload.RateLimit.Primary)
	secondary := fromRaw(entry.Payload.RateLimit.Secondary)
	if !primary.Available && !secondary.Available {
		return codexRateLimitSnapshot{}, false
	}
	return codexRateLimitSnapshot{
		Timestamp:    entry.Timestamp,
		PlanType:     entry.Payload.RateLimit.PlanType,
		LimitReached: entry.Payload.RateLimit.LimitReached,
		Primary:      primary,
		Secondary:    secondary,
	}, true
}

func codexWindowFromSnapshot(label string, window codexRateLimitWindow) CodexUsageWindow {
	if !window.Available {
		return CodexUsageWindow{Label: label}
	}
	used := window.UsedPercent
	if window.ResetsAt > 0 && window.ResetsAt <= time.Now().Unix() {
		used = 0
	}
	return CodexUsageWindow{
		Label:         label,
		UsedPercent:   used,
		LeftPercent:   maxFloat64(0, 100-used),
		WindowMinutes: window.WindowMinutes,
		ResetsAt:      window.ResetsAt,
		ResetIn:       formatResetCountdown(window.ResetsAt),
		Available:     true,
	}
}

func formatResetCountdown(ts int64) string {
	if ts == 0 {
		return ""
	}
	diff := ts - time.Now().Unix()
	if diff <= 0 {
		return "已重置"
	}
	mins := diff / 60
	hours := diff / 3600
	days := diff / 86400
	switch {
	case days >= 1:
		return fmt.Sprintf("%dd%dh", days, hours%24)
	case hours >= 1:
		return fmt.Sprintf("%dh%dm", hours, mins%60)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
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
