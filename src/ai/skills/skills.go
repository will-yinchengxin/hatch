// Package skills manages the Hatch SKILL.md master store and its one-way sync
// into ~/.claude/skills and ~/.codex/skills.
package skills

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hatch/ai/agent"
	"hatch/core/fsutil"
)

const (
	TargetClaude = agent.TargetClaude
	TargetCodex  = agent.TargetCodex
)

// Entry describes one skill in the Hatch master store.
type Entry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InClaude    bool   `json:"inClaude"`
	InCodex     bool   `json:"inCodex"`
	HasAgents   bool   `json:"hasAgents"`
	HasRefs     bool   `json:"hasRefs"`
	ModTime     int64  `json:"modTime"` // unix seconds
}

// ImportResult reports what ImportFrom did.
type ImportResult struct {
	Imported  []string         `json:"imported"`
	Skipped   []string         `json:"skipped"`
	Conflicts []ImportConflict `json:"conflicts"`
}

// ImportConflict reports a same-name skill whose source and master content differ.
type ImportConflict struct {
	Name          string `json:"name"`
	Source        string `json:"source"`
	MasterModTime int64  `json:"masterModTime"`
	SourceModTime int64  `json:"sourceModTime"`
	MasterPath    string `json:"masterPath"`
	SourcePath    string `json:"sourcePath"`
}

// Stats are the quick counts for the hero stats.
type Stats struct {
	Total  int `json:"total"`
	Claude int `json:"claude"`
	Codex  int `json:"codex"`
	Both   int `json:"both"`
}

// DefaultDir is ~/Library/Application Support/Hatch/skills (not created).
func DefaultDir() string {
	return fsutil.AppDataPath("skills")
}

// Store is the master store rooted at Dir.
type Store struct {
	Dir string
}

// Open returns the store at dir (or DefaultDir when empty) and ensures it exists.
func Open(dir string) (Store, error) {
	if dir == "" {
		d, err := fsutil.AppDataDir("skills")
		if err != nil {
			return Store{}, err
		}
		dir = d
	}
	return Store{Dir: dir}, os.MkdirAll(dir, 0o755)
}

// TargetDir returns (and creates) ~/.claude/skills or ~/.codex/skills.
func TargetDir(target string) (string, error) {
	var dir string
	var err error
	switch target {
	case TargetClaude:
		dir, err = agent.ClaudeSkillsDir()
	case TargetCodex:
		dir, err = agent.CodexSkillsDir()
	default:
		return "", fmt.Errorf("unknown target: %s (use 'claude' or 'codex')", target)
	}
	if err != nil {
		return "", err
	}
	return dir, os.MkdirAll(dir, 0o755)
}

// InvalidName rejects empty names and anything that could escape the store.
func InvalidName(name string) bool {
	return strings.Contains(name, "..") || strings.Contains(name, "/") || strings.TrimSpace(name) == ""
}

func checkName(name string) error {
	if InvalidName(name) {
		return fmt.Errorf("invalid skill name")
	}
	return nil
}

// List returns every skill (a subdir with SKILL.md) newest first.
func (s Store) List() ([]Entry, error) {
	claudeDir, _ := TargetDir(TargetClaude)
	codexDir, _ := TargetDir(TargetCodex)

	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, e := range entries {
		skillPath := filepath.Join(s.Dir, e.Name())
		if !isSkillDir(skillPath, e) {
			continue
		}
		modTime := int64(0)
		if fi, _ := e.Info(); fi != nil {
			modTime = fi.ModTime().Unix()
		}
		out = append(out, Entry{
			Name:        e.Name(),
			Description: parseDescription(skillPath),
			InClaude:    fsutil.DirExists(filepath.Join(claudeDir, e.Name())),
			InCodex:     fsutil.DirExists(filepath.Join(codexDir, e.Name())),
			HasAgents:   fsutil.DirExists(filepath.Join(skillPath, "agents")),
			HasRefs:     fsutil.DirExists(filepath.Join(skillPath, "references")),
			ModTime:     modTime,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ModTime > out[j].ModTime })
	return out, nil
}

// CountStats summarizes a skill list.
func CountStats(list []Entry) Stats {
	st := Stats{Total: len(list)}
	for _, s := range list {
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

// Exists reports whether name is a directory in the master store.
func (s Store) Exists(name string) bool {
	return fsutil.DirExists(filepath.Join(s.Dir, name))
}

// Content returns SKILL.md for one skill.
func (s Store) Content(name string) (string, error) {
	if err := checkName(name); err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(s.Dir, name, "SKILL.md"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Save creates or updates SKILL.md; existing sub-dirs are preserved.
func (s Store) Save(name, content string) error {
	if err := checkName(name); err != nil {
		return err
	}
	skillDir := filepath.Join(s.Dir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)
}

// Delete removes a skill from the master store only.
func (s Store) Delete(name string) error {
	if err := checkName(name); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(s.Dir, name))
}

// DeleteEverywhere removes a skill from the master store and both sync targets.
func (s Store) DeleteEverywhere(name string) error {
	if err := s.Delete(name); err != nil {
		return err
	}
	for _, target := range []string{TargetClaude, TargetCodex} {
		dir, err := TargetDir(target)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// SetSync copies master/name → target/name (enabled) or removes target/name.
func (s Store) SetSync(name, target string, enabled bool) error {
	if err := checkName(name); err != nil {
		return err
	}
	src := filepath.Join(s.Dir, name)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("skill not found in master store: %s", name)
	}
	base, err := TargetDir(target)
	if err != nil {
		return err
	}
	dst := filepath.Join(base, name)
	if !enabled {
		return os.RemoveAll(dst)
	}
	// Remove stale dst before copying to ensure a clean state
	_ = os.RemoveAll(dst)
	return fsutil.CopyDir(src, dst)
}

// ImportFrom copies new skill dirs from source ("claude" | "codex") into the
// master store. Same-name skills are fingerprinted; matching content is
// skipped, while differing content is returned as a conflict for user choice.
func (s Store) ImportFrom(source string) (ImportResult, error) {
	srcBase, err := sourceDir(source)
	if err != nil {
		return ImportResult{}, err
	}
	entries, err := os.ReadDir(srcBase)
	if err != nil {
		return ImportResult{}, err
	}

	var result ImportResult
	for _, e := range entries {
		src := filepath.Join(srcBase, e.Name())
		// Client bookkeeping dirs (.trash, .system, synced/) are not skills.
		if !isSkillDir(src, e) {
			continue
		}
		dst := filepath.Join(s.Dir, e.Name())
		if !fsutil.DirExists(dst) {
			if err := fsutil.CopyDir(src, dst); err != nil {
				result.Skipped = append(result.Skipped, e.Name()+" (error: "+err.Error()+")")
				continue
			}
			result.Imported = append(result.Imported, e.Name())
			continue
		}
		srcHash, srcMod, err := Fingerprint(src)
		if err != nil {
			result.Skipped = append(result.Skipped, e.Name()+" (error: "+err.Error()+")")
			continue
		}
		dstHash, dstMod, err := Fingerprint(dst)
		if err != nil {
			result.Skipped = append(result.Skipped, e.Name()+" (error: "+err.Error()+")")
			continue
		}
		if srcHash == dstHash {
			result.Skipped = append(result.Skipped, e.Name())
			continue
		}
		result.Conflicts = append(result.Conflicts, ImportConflict{
			Name:          e.Name(),
			Source:        source,
			MasterModTime: dstMod,
			SourceModTime: srcMod,
			MasterPath:    dst,
			SourcePath:    src,
		})
	}
	return result, nil
}

// ResolveConflict replaces the master copy with source's when useSource is true.
func (s Store) ResolveConflict(name, source string, useSource bool) error {
	if err := checkName(name); err != nil {
		return err
	}
	srcBase, err := sourceDir(source)
	if err != nil {
		return err
	}
	if !useSource {
		return nil
	}
	src := filepath.Join(srcBase, name)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("skill not found in %s: %s", source, name)
	}
	dst := filepath.Join(s.Dir, name)
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return fsutil.CopyDir(src, dst)
}

// TargetMatchesMaster reports whether target's copy of name is absent or
// byte-identical to the master copy (i.e. safe for Hatch to overwrite/remove).
func (s Store) TargetMatchesMaster(name, target string) bool {
	base, err := TargetDir(target)
	if err != nil {
		return false
	}
	dst := filepath.Join(base, name)
	if !fsutil.DirExists(dst) {
		return true
	}
	left, _, err := Fingerprint(filepath.Join(s.Dir, name))
	if err != nil {
		return false
	}
	right, _, err := Fingerprint(dst)
	return err == nil && left == right
}

// Template returns a starter SKILL.md for a new skill.
func Template(name string) string {
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

// Fingerprint hashes every file (relative path + content) under dir and
// returns the newest file mtime alongside.
func Fingerprint(dir string) (string, int64, error) {
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

// isSkillDir reports whether e is a visible directory holding a SKILL.md.
func isSkillDir(path string, e os.DirEntry) bool {
	if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
		return false
	}
	_, err := os.Stat(filepath.Join(path, "SKILL.md"))
	return err == nil
}

func sourceDir(source string) (string, error) {
	if source != TargetClaude && source != TargetCodex {
		return "", fmt.Errorf("unknown source: %s", source)
	}
	return TargetDir(source)
}

// parseDescription reads the description field from SKILL.md YAML frontmatter.
func parseDescription(skillDir string) string {
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
