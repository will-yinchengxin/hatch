package app

import (
	"fmt"

	"hatch/core/backup"
	"hatch/core/config"
)

// ─── Backup / Restore ────────────────────────────────────────────────────────

// ExportBackup writes a zip containing config.json, scripts.json (if any) and
// the skills dir to targetDir (default ~/Desktop).
func (a *App) ExportBackup(targetDir string) (backup.Result, error) {
	return backup.Export(targetDir, a.GetAppDataPaths().SkillsDir)
}

// ImportBackup restores config.json / scripts.json and the skills dir from a
// zip, then reloads config.
func (a *App) ImportBackup(zipPath string) (string, error) {
	restored, err := backup.Import(zipPath, a.GetAppDataPaths().SkillsDir)
	if err != nil {
		return "", err
	}
	if cfg, err := config.Load(); err == nil {
		a.mu.Lock()
		a.cfg = cfg
		a.mu.Unlock()
	}
	return fmt.Sprintf("已恢复 %d 个文件，请重启应用以让所有变更生效。", restored), nil
}
