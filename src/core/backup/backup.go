// Package backup exports and restores Hatch's own data (config.json,
// scripts.json and the skills master store) as a single zip.
package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hatch/core/fsutil"
)

// Result describes a created backup archive.
type Result struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	NumFiles int    `json:"numFiles"`
}

// Export writes <targetDir>/hatch-backup-<ts>.zip (default ~/Desktop) holding
// config.json, scripts.json (if any) and every file under skillsDir.
func Export(targetDir, skillsDir string) (Result, error) {
	if strings.TrimSpace(targetDir) == "" {
		home, _ := os.UserHomeDir()
		targetDir = filepath.Join(home, "Desktop")
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Result{}, err
	}
	zipPath := filepath.Join(targetDir, "hatch-backup-"+time.Now().Format("20060102-150405")+".zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return Result{}, err
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
		if _, err := w.Write(data); err != nil {
			return err
		}
		numFiles++
		return nil
	}

	if hatchDir, err := fsutil.AppDataDir(); err == nil {
		_ = addFile(filepath.Join(hatchDir, "config.json"), "config.json")
		_ = addFile(filepath.Join(hatchDir, "scripts.json"), "scripts.json")
	}

	// skills dir (best-effort, recursively)
	if skillsDir != "" {
		_ = filepath.WalkDir(skillsDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			// Relative to skillsDir itself: entries are skills/<name>/…, which is
			// what Import expects (older builds wrote skills/skills/<name>/…).
			rel, _ := filepath.Rel(skillsDir, path)
			return addFile(path, filepath.ToSlash(filepath.Join("skills", rel)))
		})
	}

	if err := zw.Close(); err != nil {
		return Result{}, err
	}
	fi, _ := f.Stat()
	return Result{Path: zipPath, Size: fi.Size(), NumFiles: numFiles}, nil
}

// Import restores config.json / scripts.json into the Hatch data dir and
// skills/* into skillsDir. It returns how many files were written; the caller
// reloads config.
func Import(zipPath, skillsDir string) (int, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	hatchDir, err := fsutil.AppDataDir()
	if err != nil {
		return 0, fmt.Errorf("无法定位 user config dir")
	}
	if err := os.MkdirAll(hatchDir, 0o755); err != nil {
		return 0, err
	}

	legacy := legacySkillsPrefix(r.File)
	restored := 0
	for _, f := range r.File {
		var dst string
		switch {
		case f.Name == "config.json", f.Name == "scripts.json":
			dst = filepath.Join(hatchDir, f.Name)
			// Keep the current file restorable: a wrong zip shouldn't cost the config.
			if err := fsutil.BackupIfExists(dst); err != nil {
				return restored, err
			}
		case strings.HasPrefix(f.Name, "skills/"):
			dst = filepath.Join(skillsDir, strings.TrimPrefix(strings.TrimPrefix(f.Name, "skills/"), legacy))
			if !fsutil.IsPathInside(dst, skillsDir) { // zip-slip: "skills/../../x"
				continue
			}
			_ = os.MkdirAll(filepath.Dir(dst), 0o755)
		default:
			continue
		}
		if err := writeZipEntry(f, dst); err == nil {
			restored++
		}
	}
	return restored, nil
}

// legacySkillsPrefix detects zips from builds that nested skills one level too
// deep (skills/skills/<name>/…): every skills entry shares one top folder and
// none sits directly in it as SKILL.md. Returns "<folder>/" to strip, or "".
func legacySkillsPrefix(files []*zip.File) string {
	top, seen := "", false
	for _, f := range files {
		rest, ok := strings.CutPrefix(f.Name, "skills/")
		if !ok || f.FileInfo().IsDir() {
			continue
		}
		parts := strings.SplitN(rest, "/", 3)
		if len(parts) < 3 {
			return "" // skills/<name>/<file>: current layout
		}
		if seen && parts[0] != top {
			return ""
		}
		top, seen = parts[0], true
	}
	if !seen {
		return ""
	}
	return top + "/"
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
