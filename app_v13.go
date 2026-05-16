package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hatch/internal/config"
	"hatch/internal/executor"
)

// ─── v1.3-2: PATH health ─────────────────────────────────────────────────────

type PathEntry struct {
	Path        string `json:"path"`
	Exists      bool   `json:"exists"`
	Duplicate   bool   `json:"duplicate"`  // appears more than once
	ShadowsName string `json:"shadowsName"` // if a binary in this dir shadows one earlier on PATH
}

type PathHealth struct {
	Entries        []PathEntry `json:"entries"`
	MissingCount   int         `json:"missingCount"`
	DuplicateCount int         `json:"duplicateCount"`
}

// CheckPathHealth inspects $PATH from a login zsh and reports missing/duplicate dirs.
func (a *App) CheckPathHealth() PathHealth {
	out, err := exec.Command("/bin/zsh", "-l", "-c", "echo $PATH").Output()
	if err != nil {
		return PathHealth{}
	}
	rawPath := strings.TrimSpace(string(out))
	parts := strings.Split(rawPath, ":")

	seen := map[string]int{}
	binToFirstDir := map[string]string{}

	h := PathHealth{}
	for _, p := range parts {
		if p == "" {
			continue
		}
		entry := PathEntry{Path: p}
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			entry.Exists = true
		} else {
			h.MissingCount++
		}
		seen[p]++
		if seen[p] > 1 {
			entry.Duplicate = true
			h.DuplicateCount++
		}
		// Detect shadowing: list binaries in this dir
		if entry.Exists {
			if files, err := os.ReadDir(p); err == nil {
				for _, f := range files {
					if f.IsDir() {
						continue
					}
					name := f.Name()
					if prevDir, ok := binToFirstDir[name]; ok && prevDir != p && entry.ShadowsName == "" {
						entry.ShadowsName = name + " (also in " + prevDir + ")"
					} else if !ok {
						binToFirstDir[name] = p
					}
				}
			}
		}
		h.Entries = append(h.Entries, entry)
	}
	return h
}

// ─── v1.3-3: Backup / Restore ────────────────────────────────────────────────

type BackupResult struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	NumFiles int    `json:"numFiles"`
}

// ExportBackup writes a zip containing config.json, scripts.json (if any), and the skills dir.
// Returns the absolute path of the zip.
func (a *App) ExportBackup(targetDir string) (BackupResult, error) {
	if strings.TrimSpace(targetDir) == "" {
		home, _ := os.UserHomeDir()
		targetDir = filepath.Join(home, "Desktop")
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return BackupResult{}, err
	}
	stamp := time.Now().Format("20060102-150405")
	zipPath := filepath.Join(targetDir, "hatch-backup-"+stamp+".zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return BackupResult{}, err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	numFiles := 0

	addFile := func(srcPath, rel string) error {
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return nil // skip missing
		}
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		if err != nil {
			return err
		}
		numFiles++
		return nil
	}

	cfgDir, _ := os.UserConfigDir()
	if cfgDir != "" {
		_ = addFile(filepath.Join(cfgDir, "Hatch", "config.json"), "config.json")
		_ = addFile(filepath.Join(cfgDir, "Hatch", "scripts.json"), "scripts.json")
	}

	// skills dir (best-effort, recursively)
	skillsDir := a.GetAppDataPaths().SkillsDir
	if skillsDir != "" {
		_ = filepath.WalkDir(skillsDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(filepath.Dir(skillsDir), path)
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			w, err := zw.Create(filepath.Join("skills", rel))
			if err != nil {
				return err
			}
			_, _ = w.Write(data)
			numFiles++
			return nil
		})
	}

	if err := zw.Close(); err != nil {
		return BackupResult{}, err
	}
	fi, _ := f.Stat()
	return BackupResult{Path: zipPath, Size: fi.Size(), NumFiles: numFiles}, nil
}

// ImportBackup restores config.json (and optionally scripts.json) from a zip.
// SkillsDir is restored side-by-side without overwriting unchanged files.
// Returns a brief summary.
func (a *App) ImportBackup(zipPath string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	cfgDir, _ := os.UserConfigDir()
	if cfgDir == "" {
		return "", fmt.Errorf("无法定位 user config dir")
	}
	hatchDir := filepath.Join(cfgDir, "Hatch")
	if err := os.MkdirAll(hatchDir, 0o755); err != nil {
		return "", err
	}

	restored := 0
	for _, f := range r.File {
		switch {
		case f.Name == "config.json":
			if err := writeZipEntry(f, filepath.Join(hatchDir, "config.json")); err == nil {
				restored++
			}
		case f.Name == "scripts.json":
			if err := writeZipEntry(f, filepath.Join(hatchDir, "scripts.json")); err == nil {
				restored++
			}
		case strings.HasPrefix(f.Name, "skills/"):
			target := filepath.Join(a.GetAppDataPaths().SkillsDir, strings.TrimPrefix(f.Name, "skills/"))
			_ = os.MkdirAll(filepath.Dir(target), 0o755)
			if err := writeZipEntry(f, target); err == nil {
				restored++
			}
		}
	}

	// Reload config
	if cfg, err := config.Load(); err == nil {
		a.mu.Lock()
		a.cfg = cfg
		a.mu.Unlock()
	}
	return fmt.Sprintf("已恢复 %d 个文件，请重启应用以让所有变更生效。", restored), nil
}

func writeZipEntry(f *zip.File, dst string) error {
	if f.FileInfo().IsDir() {
		return nil
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

// ─── v1.3-4: Script orchestration (pre-scripts) ──────────────────────────────

// RunWithPreScripts executes pre-scripts in order, then the target.
// pre-scripts are paths to oneshot scripts that must exit 0 before continuing.
// Emits a "[hatch] pre: " log line at the beginning of each step.
func (a *App) RunWithPreScripts(id, path string, args []string, preScripts []string) {
	go func() {
		em := executor.NewEmitter(a.ctx)
		for _, pre := range preScripts {
			pre = strings.TrimSpace(pre)
			if pre == "" {
				continue
			}
			em.Log(id, "[hatch] pre: "+pre, "stdout")
			res, err := executor.RunOneshotWithEnv(pre, pre, nil, a.scriptEnv(pre), nil, nil)
			if err != nil || res == nil || res.ExitCode != 0 {
				code := -1
				if res != nil {
					code = res.ExitCode
				}
				em.Log(id, fmt.Sprintf("[hatch] pre failed: %s (exit %d)", pre, code), "stderr")
				em.Done(id, 99)
				em.Status(id, "error")
				return
			}
		}
		_, _ = executor.RunOneshotWithEnv(id, path, args, a.scriptEnv(path), em, a.notifyHook(path))
	}()
}

// ─── v1.3-2: alias inventory (read-only) ─────────────────────────────────────

type AliasGroup struct {
	Kind    string   `json:"kind"`    // alias / export / source / function / other
	Entries []string `json:"entries"` // raw lines
}

// ListAliasInventory reads ~/.zshrc and groups entries by kind for quick review.
func (a *App) ListAliasInventory() []AliasGroup {
	entries, err := a.ParseZshrc()
	if err != nil {
		return nil
	}
	buckets := map[string][]string{}
	for _, e := range entries {
		// Entry struct exposes Kind/Raw in zshrc package; we use json round-trip to access fields generically.
		var anon struct {
			Kind string `json:"kind"`
			Raw  string `json:"raw"`
			Key  string `json:"key"`
		}
		data, _ := json.Marshal(e)
		_ = json.Unmarshal(data, &anon)
		kind := anon.Kind
		if kind == "" {
			kind = "other"
		}
		line := anon.Raw
		if line == "" {
			line = anon.Key
		}
		buckets[kind] = append(buckets[kind], line)
	}
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]AliasGroup, 0, len(keys))
	for _, k := range keys {
		out = append(out, AliasGroup{Kind: k, Entries: buckets[k]})
	}
	return out
}
