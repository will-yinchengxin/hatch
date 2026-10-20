package app

import (
	"fmt"
	"os"
	"strings"

	"hatch/workbench/zshrc"
)

// ─── 工作台 · ~/.zshrc ───────────────────────────────────────────────────────

// ZshrcInfo is returned by CheckZshrc.
type ZshrcInfo struct {
	Shell      string   `json:"shell"`
	IsZsh      bool     `json:"isZsh"`
	ZshrcPath  string   `json:"zshrcPath"`
	HatchLines []string `json:"hatchLines"`
}

// ZshrcApplyResult reports where an approved reorganization was written.
type ZshrcApplyResult struct {
	ZshrcPath  string `json:"zshrcPath"`
	BackupPath string `json:"backupPath"`
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
	content, err := readZshrc()
	if err != nil {
		return nil, err
	}
	a.mu.RLock()
	ann := a.cfg.ZshrcAnnotations
	a.mu.RUnlock()
	return zshrc.ParseContent(content, ann), nil
}

// SaveZshrcAnnotation persists a user note for a zshrc entry key ("" deletes it).
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

// ApplyZshrcOrganization backs up ~/.zshrc and writes the user-approved content
// (typically from SuggestZshrcOrganization).
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

// ListAliasInventory groups ~/.zshrc entries by type (alias / export / function …).
func (a *App) ListAliasInventory() []zshrc.Group {
	entries, err := a.ParseZshrc()
	if err != nil {
		return nil
	}
	return zshrc.GroupByType(entries)
}

func readZshrc() (string, error) {
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
