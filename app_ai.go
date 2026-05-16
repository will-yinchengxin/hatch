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
	"time"

	"hatch/internal/config"
	"hatch/internal/zshrc"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

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

type AIProvidersResult struct {
	Providers        []config.AIProvider `json:"providers"`
	ActiveProviderID string              `json:"activeProviderId"`
}

// SaveAIConfig persists OpenAI-compatible model settings.
func (a *App) SaveAIConfig(ai config.AIConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.AI = ai
	return a.cfg.Save()
}

func (a *App) GetAIProviders() AIProvidersResult {
	a.mu.RLock()
	defer a.mu.RUnlock()
	providers := make([]config.AIProvider, len(a.cfg.AIProviders))
	copy(providers, a.cfg.AIProviders)
	return AIProvidersResult{Providers: providers, ActiveProviderID: a.cfg.ActiveProviderID}
}

func (a *App) SaveAIProviders(providers []config.AIProvider, activeID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.AIProviders = providers
	a.cfg.ActiveProviderID = activeID
	for _, p := range providers {
		if p.ID == activeID {
			a.cfg.AI.BaseURL = p.BaseURL
			a.cfg.AI.APIKey = p.APIKey
			a.cfg.AI.Model = p.Model
			break
		}
	}
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
	ai := a.activeAIConfig()
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
	ai := a.activeAIConfig()
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

// SuggestScriptContent asks the configured AI to write or improve a script.
func (a *App) SuggestScriptContent(path, currentContent, instruction string) (string, error) {
	a.mu.RLock()
	ai := a.activeAIConfig()
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
	raw = strings.TrimSpace(raw)
	for _, fence := range []string{"```zsh", "```shell", "```sh", "```bash", "```python", "```ruby", "```node", "```"} {
		if strings.HasPrefix(raw, fence) {
			raw = strings.TrimPrefix(raw, fence)
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

// validateAIConfig ensures the AI config has the minimum required fields.
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
	a.traces = append([]AICallTrace{t}, a.traces...)
	if len(a.traces) > maxTraces {
		a.traces = a.traces[:maxTraces]
	}
	a.tracesMu.Unlock()

	go func() {
		if dir, err := a.traceDir(); err == nil {
			appendTraceFile(dir, t)
			pruneTraceFiles(dir, a.traceRetainDays())
		}
	}()

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
