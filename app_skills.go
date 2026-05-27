package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ─── Skills management ────────────────────────────────────────────────────────

// SkillEntry describes one skill in the Hatch master store.
type SkillEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InClaude    bool   `json:"inClaude"`
	InCodex     bool   `json:"inCodex"`
	HasAgents   bool   `json:"hasAgents"`
	HasRefs     bool   `json:"hasRefs"`
	ModTime     int64  `json:"modTime"` // unix seconds
}

// SkillImportResult reports what was imported.
type SkillImportResult struct {
	Imported  []string              `json:"imported"`
	Skipped   []string              `json:"skipped"`
	Conflicts []SkillImportConflict `json:"conflicts"`
}

// SkillImportConflict reports a same-name skill whose source and master content differ.
type SkillImportConflict struct {
	Name          string `json:"name"`
	Source        string `json:"source"`
	MasterModTime int64  `json:"masterModTime"`
	SourceModTime int64  `json:"sourceModTime"`
	MasterPath    string `json:"masterPath"`
	SourcePath    string `json:"sourcePath"`
}

const (
	skillTargetClaude = "claude"
	skillTargetCodex  = "codex"
)

// masterSkillsDir returns the skills master directory (from config or default) and ensures it exists.
func (a *App) masterSkillsDir() (string, error) {
	a.mu.RLock()
	custom := a.cfg.SkillsDir
	a.mu.RUnlock()
	if custom != "" {
		return custom, os.MkdirAll(custom, 0o755)
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfgDir, "Hatch", "skills")
	return dir, os.MkdirAll(dir, 0o755)
}

// defaultSkillsDir returns the default skills dir path (not creating it).
func defaultSkillsDir() string {
	cfgDir, _ := os.UserConfigDir()
	return filepath.Join(cfgDir, "Hatch", "skills")
}

func claudeSkillsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".claude", "skills")
	return dir, os.MkdirAll(dir, 0o755)
}

func codexSkillsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".codex", "skills")
	return dir, os.MkdirAll(dir, 0o755)
}

// parseSkillDescription reads the description field from SKILL.md YAML frontmatter.
func parseSkillDescription(skillDir string) string {
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return ""
	}
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Scan() // consume opening "---"
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		if strings.HasPrefix(line, "description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	return ""
}

// dirExists returns true if path is an existing directory.
func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// copyDirAll copies src directory to dst recursively (dst is created if needed).
func copyDirAll(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

func skillNameInvalid(name string) bool {
	return strings.Contains(name, "..") || strings.Contains(name, "/") || strings.TrimSpace(name) == ""
}

func skillDirFingerprint(dir string) (string, int64, error) {
	var files []string
	var latest int64
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		if mt := info.ModTime().Unix(); mt > latest {
			latest = mt
		}
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	sort.Strings(files)
	h := md5.New()
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			return "", 0, err
		}
		h.Write([]byte(rel))
		h.Write([]byte{0})
		h.Write(data)
		h.Write([]byte{0})
	}
	if len(files) == 0 {
		if fi, err := os.Stat(dir); err == nil {
			latest = fi.ModTime().Unix()
		}
	}
	return hex.EncodeToString(h.Sum(nil)), latest, nil
}

// GetSkills returns all skills in the master store with their sync status.
func (a *App) GetSkills() ([]SkillEntry, error) {
	master, err := a.masterSkillsDir()
	if err != nil {
		return nil, err
	}
	claudeDir, _ := claudeSkillsDir()
	codexDir, _ := codexSkillsDir()

	entries, err := os.ReadDir(master)
	if err != nil {
		return nil, err
	}

	var skills []SkillEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillPath := filepath.Join(master, e.Name())
		// skip dirs that don't have SKILL.md — they're not valid skills
		if _, err := os.Stat(filepath.Join(skillPath, "SKILL.md")); err != nil {
			continue
		}
		fi, _ := e.Info()
		modTime := int64(0)
		if fi != nil {
			modTime = fi.ModTime().Unix()
		}
		skill := SkillEntry{
			Name:        e.Name(),
			Description: parseSkillDescription(skillPath),
			InClaude:    dirExists(filepath.Join(claudeDir, e.Name())),
			InCodex:     dirExists(filepath.Join(codexDir, e.Name())),
			HasAgents:   dirExists(filepath.Join(skillPath, "agents")),
			HasRefs:     dirExists(filepath.Join(skillPath, "references")),
			ModTime:     modTime,
		}
		skills = append(skills, skill)
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].ModTime > skills[j].ModTime
	})
	return skills, nil
}

// GetSkillContent returns the SKILL.md content for one skill in the master store.
func (a *App) GetSkillContent(name string) (string, error) {
	if skillNameInvalid(name) {
		return "", fmt.Errorf("invalid skill name")
	}
	master, err := a.masterSkillsDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(master, name, "SKILL.md"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveSkillContent creates or updates the SKILL.md in the master store.
// If the skill dir doesn't exist it is created. Existing sub-dirs are preserved.
func (a *App) SaveSkillContent(name, content string) error {
	if skillNameInvalid(name) {
		return fmt.Errorf("invalid skill name")
	}
	master, err := a.masterSkillsDir()
	if err != nil {
		return err
	}
	skillDir := filepath.Join(master, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)
}

// DeleteSkill removes a skill from the master store.
// It does NOT touch ~/.claude/skills or ~/.codex/skills.
func (a *App) DeleteSkill(name string) error {
	if skillNameInvalid(name) {
		return fmt.Errorf("invalid skill name")
	}
	master, err := a.masterSkillsDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(master, name))
}

// DeleteSkillEverywhere removes a skill from the master store and both sync targets.
func (a *App) DeleteSkillEverywhere(name string) error {
	if skillNameInvalid(name) {
		return fmt.Errorf("invalid skill name")
	}
	master, err := a.masterSkillsDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(master, name)); err != nil {
		return err
	}
	claudeDir, err := claudeSkillsDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(claudeDir, name)); err != nil {
		return err
	}
	codexDir, err := codexSkillsDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(codexDir, name))
}

// SetSkillSync syncs (or removes) one skill to/from the specified target ("claude" | "codex").
// enabled=true copies master/name → target/name; enabled=false removes target/name.
func (a *App) SetSkillSync(name, target string, enabled bool) error {
	if skillNameInvalid(name) {
		return fmt.Errorf("invalid skill name")
	}
	master, err := a.masterSkillsDir()
	if err != nil {
		return err
	}
	src := filepath.Join(master, name)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("skill not found in master store: %s", name)
	}

	var targetBase string
	switch target {
	case skillTargetClaude:
		targetBase, err = claudeSkillsDir()
	case skillTargetCodex:
		targetBase, err = codexSkillsDir()
	default:
		return fmt.Errorf("unknown target: %s (use 'claude' or 'codex')", target)
	}
	if err != nil {
		return err
	}

	dst := filepath.Join(targetBase, name)
	if !enabled {
		return os.RemoveAll(dst)
	}
	// Remove stale dst before copying to ensure a clean state
	_ = os.RemoveAll(dst)
	return copyDirAll(src, dst)
}

// ImportSkillsFrom copies new skill dirs from source ("claude" | "codex") into
// the master store. Same-name skills are fingerprinted; matching content is
// skipped, while differing content is returned as a conflict for user choice.
func (a *App) ImportSkillsFrom(source string) (SkillImportResult, error) {
	var srcBase string
	var err error
	switch source {
	case skillTargetClaude:
		srcBase, err = claudeSkillsDir()
	case skillTargetCodex:
		srcBase, err = codexSkillsDir()
	default:
		return SkillImportResult{}, fmt.Errorf("unknown source: %s", source)
	}
	if err != nil {
		return SkillImportResult{}, err
	}

	master, err := a.masterSkillsDir()
	if err != nil {
		return SkillImportResult{}, err
	}

	entries, err := os.ReadDir(srcBase)
	if err != nil {
		return SkillImportResult{}, err
	}

	var result SkillImportResult
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dst := filepath.Join(master, e.Name())
		if dirExists(dst) {
			src := filepath.Join(srcBase, e.Name())
			srcHash, srcMod, err := skillDirFingerprint(src)
			if err != nil {
				result.Skipped = append(result.Skipped, e.Name()+" (error: "+err.Error()+")")
				continue
			}
			dstHash, dstMod, err := skillDirFingerprint(dst)
			if err != nil {
				result.Skipped = append(result.Skipped, e.Name()+" (error: "+err.Error()+")")
				continue
			}
			if srcHash == dstHash {
				result.Skipped = append(result.Skipped, e.Name())
				continue
			}
			result.Conflicts = append(result.Conflicts, SkillImportConflict{
				Name:          e.Name(),
				Source:        source,
				MasterModTime: dstMod,
				SourceModTime: srcMod,
				MasterPath:    dst,
				SourcePath:    src,
			})
			continue
		}
		if err := copyDirAll(filepath.Join(srcBase, e.Name()), dst); err != nil {
			result.Skipped = append(result.Skipped, e.Name()+" (error: "+err.Error()+")")
			continue
		}
		result.Imported = append(result.Imported, e.Name())
	}
	return result, nil
}

// ResolveSkillImportConflict applies or dismisses one conflict returned by ImportSkillsFrom.
func (a *App) ResolveSkillImportConflict(name, source string, useSource bool) error {
	if skillNameInvalid(name) {
		return fmt.Errorf("invalid skill name")
	}
	var srcBase string
	var err error
	switch source {
	case skillTargetClaude:
		srcBase, err = claudeSkillsDir()
	case skillTargetCodex:
		srcBase, err = codexSkillsDir()
	default:
		return fmt.Errorf("unknown source: %s", source)
	}
	if err != nil {
		return err
	}
	if !useSource {
		return nil
	}
	master, err := a.masterSkillsDir()
	if err != nil {
		return err
	}
	src := filepath.Join(srcBase, name)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("skill not found in %s: %s", source, name)
	}
	dst := filepath.Join(master, name)
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return copyDirAll(src, dst)
}

// CreateSkillTemplate returns a starter SKILL.md template for a new skill.
func (a *App) CreateSkillTemplate(name string) string {
	return fmt.Sprintf(`---
name: %s
description:
triggers:
  -
---

# %s

在此描述 Skill 的功能和使用场景。

## 使用方式

## 示例
`, name, name)
}

// GetSkillsStats returns quick counts for the hero stats.
type SkillsStats struct {
	Total  int `json:"total"`
	Claude int `json:"claude"`
	Codex  int `json:"codex"`
	Both   int `json:"both"`
}

func (a *App) GetSkillsStats() SkillsStats {
	skills, err := a.GetSkills()
	if err != nil {
		return SkillsStats{}
	}
	st := SkillsStats{Total: len(skills)}
	for _, s := range skills {
		if s.InClaude {
			st.Claude++
		}
		if s.InCodex {
			st.Codex++
		}
		if s.InClaude && s.InCodex {
			st.Both++
		}
	}
	return st
}

// ─── Zshrc ~.zshrc IA fix ─────────────────────────────────────────────────────

func formatSkillModTime(unix int64) string {
	if unix == 0 {
		return ""
	}
	return time.Unix(unix, 0).Format("01-02")
}
