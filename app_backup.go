package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hatch/internal/config"
)

// ─── Backup / Restore ────────────────────────────────────────────────────────

// BackupResult describes a created backup archive.
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
