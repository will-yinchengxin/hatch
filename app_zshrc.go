package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"hatch/internal/config"
	"hatch/internal/zshrc"
)

// ZshrcInfo is returned by CheckZshrc.
type ZshrcInfo struct {
	Shell      string   `json:"shell"`
	IsZsh      bool     `json:"isZsh"`
	ZshrcPath  string   `json:"zshrcPath"`
	HatchLines []string `json:"hatchLines"`
}

type ZshrcOrganizationSuggestion struct {
	Summary []string `json:"summary"`
	Content string   `json:"content"`
}

type ZshrcApplyResult struct {
	ZshrcPath  string `json:"zshrcPath"`
	BackupPath string `json:"backupPath"`
}

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
func (a *App) SuggestZshrcOrganization(mode string) (ZshrcOrganizationSuggestion, error) {
	a.mu.RLock()
	ai := a.activeAIConfig()
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
