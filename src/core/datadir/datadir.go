// Package datadir describes every on-disk location Hatch writes to, measures
// them for the AI → 数据 page, and cleans up the generated ones.
package datadir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hatch/core/macos"
)

// PIDDir is where service-type scripts keep their PID files.
const PIDDir = "/tmp/hatch"

// Entry describes one user-data location managed by Hatch.
type Entry struct {
	Key         string `json:"key"`   // stable id: config / traces / runs / skills / pids / backups
	Label       string `json:"label"` // human-readable
	Path        string `json:"path"`  // absolute path
	Exists      bool   `json:"exists"`
	IsDir       bool   `json:"isDir"`
	TotalBytes  int64  `json:"totalBytes"`
	FileCount   int    `json:"fileCount"`
	OldestUnix  int64  `json:"oldestUnix"` // mtime of oldest file (0 if none)
	NewestUnix  int64  `json:"newestUnix"`
	Cleanable   bool   `json:"cleanable"`   // whether CleanupDataDir(key) does something
	Description string `json:"description"` // what's stored here
}

// Locations are the configurable directories the inventory needs.
type Locations struct {
	HatchDir  string // ~/Library/Application Support/Hatch
	TraceDir  string
	RunsDir   string
	SkillsDir string
	MemoFiles string
}

// Inventory lists and measures every location Hatch writes to.
func Inventory(loc Locations) []Entry {
	home, _ := os.UserHomeDir()
	entries := []Entry{
		{Key: "config", Label: "主配置", Path: filepath.Join(loc.HatchDir, "config.json"),
			Description: "Hatch 全局配置：目录列表、AI provider、prompt 模板、自定义 env、界面模块开关等。"},
		{Key: "store", Label: "脚本元信息缓存", Path: filepath.Join(loc.HatchDir, "scripts.json"),
			Description: "脚本扫描结果的本地缓存，下次启动可秒开。"},
		{Key: "traces", Label: "AI 调用日志 (JSONL)", Path: loc.TraceDir, Cleanable: true,
			Description: "每次 AI 调用的 system/user prompt + response。按日期切片，自动按 AI.traceRetainDays 清理。"},
		{Key: "runs", Label: "脚本运行历史", Path: loc.RunsDir, Cleanable: true,
			Description: "每次脚本执行的完整输出 + 元数据。受 Retention.runsRetainDays / runsKeepPerName 限制。"},
		{Key: "skills", Label: "Skills 主仓库", Path: loc.SkillsDir,
			Description: "Hatch 维护的 SKILL.md 主仓库。同步到 ~/.claude/skills 和 ~/.codex/skills。"},
		{Key: "memo-files", Label: "备忘录附件", Path: loc.MemoFiles, Cleanable: true,
			Description: "备忘录托管的图片和文件附件。受 Memo.retainDays / autoCleanup 限制。"},
		{Key: "pids", Label: "Service PID 文件", Path: PIDDir, Cleanable: true,
			Description: "service 类脚本运行时的 PID 文件。停止后会自动清理，stale 文件可手动清理。"},
		// backups live on Desktop, user manages them
		{Key: "backups", Label: "备份导出", Path: filepath.Join(home, "Desktop"),
			Description: "ExportBackup 默认输出目录（~/Desktop/hatch-backup-*.zip）。"},
	}
	for i := range entries {
		stat(&entries[i])
	}
	return entries
}

func stat(e *Entry) {
	fi, err := os.Stat(e.Path)
	if err != nil {
		return
	}
	e.Exists = true
	e.IsDir = fi.IsDir()
	if !e.IsDir {
		e.TotalBytes = fi.Size()
		e.FileCount = 1
		e.OldestUnix = fi.ModTime().Unix()
		e.NewestUnix = fi.ModTime().Unix()
		return
	}
	_ = filepath.WalkDir(e.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		// Match only "interesting" backup files at Desktop key
		if e.Key == "backups" && !strings.HasPrefix(filepath.Base(path), "hatch-backup-") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		e.FileCount++
		e.TotalBytes += info.Size()
		mt := info.ModTime().Unix()
		if e.OldestUnix == 0 || mt < e.OldestUnix {
			e.OldestUnix = mt
		}
		if mt > e.NewestUnix {
			e.NewestUnix = mt
		}
		return nil
	})
}

// CleanRunHistory deletes every archived run log/record under runsDir.
func CleanRunHistory(runsDir string) string {
	count := 0
	_ = filepath.WalkDir(runsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".log") || strings.HasSuffix(path, ".json") {
			if os.Remove(path) == nil {
				count++
			}
		}
		return nil
	})
	return fmt.Sprintf("脚本运行历史已清理 %d 个文件。", count)
}

// CleanStalePIDs deletes PID files whose process is no longer alive.
func CleanStalePIDs() string {
	entries, err := os.ReadDir(PIDDir)
	if err != nil {
		return "PID 目录不存在，无需清理。"
	}
	removed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pid") {
			continue
		}
		path := filepath.Join(PIDDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if !macos.PIDAlive(strings.TrimSpace(string(data))) {
			if os.Remove(path) == nil {
				removed++
			}
		}
	}
	return fmt.Sprintf("Stale PID 文件已清理 %d 个。", removed)
}
