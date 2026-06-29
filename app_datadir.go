package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"hatch/internal/autostart"
	"hatch/internal/config"
	"hatch/internal/executor"
)

// ─── Data directory inventory ────────────────────────────────────────────────

// DataDirEntry describes one user-data location managed by Hatch.
type DataDirEntry struct {
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

// GetAppDataInventory returns every directory/file Hatch writes to.
// Use this in the UI to show users exactly what's on disk and let them clean up.
func (a *App) GetAppDataInventory() []DataDirEntry {
	home, _ := os.UserHomeDir()
	cfgDir, _ := os.UserConfigDir()
	hatchDir := filepath.Join(cfgDir, "Hatch")

	paths := a.GetAppDataPaths()

	entries := []DataDirEntry{
		{
			Key:         "config",
			Label:       "主配置",
			Path:        filepath.Join(hatchDir, "config.json"),
			Cleanable:   false,
			Description: "Hatch 全局配置：目录列表、AI provider、prompt 模板、自定义 env 等。",
		},
		{
			Key:         "store",
			Label:       "脚本元信息缓存",
			Path:        filepath.Join(hatchDir, "scripts.json"),
			Cleanable:   false,
			Description: "脚本扫描结果的本地缓存，下次启动可秒开。",
		},
		{
			Key:         "traces",
			Label:       "AI 调用日志 (JSONL)",
			Path:        paths.TraceLogDir,
			Cleanable:   true,
			Description: "每次 AI 调用的 system/user prompt + response。按日期切片，自动按 AI.traceRetainDays 清理。",
		},
		{
			Key:         "runs",
			Label:       "脚本运行历史",
			Path:        executor.HistoryRoot(),
			Cleanable:   true,
			Description: "每次脚本执行的完整输出 + 元数据。受 Retention.runsRetainDays / runsKeepPerName 限制。",
		},
		{
			Key:         "skills",
			Label:       "Skills 主仓库",
			Path:        paths.SkillsDir,
			Cleanable:   false,
			Description: "Hatch 维护的 SKILL.md 主仓库。同步到 ~/.claude/skills 和 ~/.codex/skills。",
		},
		{
			Key:         "memo-files",
			Label:       "备忘录附件",
			Path:        a.memoFilesDir(),
			Cleanable:   true,
			Description: "备忘录托管的图片和文件附件。受 Memo.retainDays / autoCleanup 限制。",
		},
		{
			Key:         "pids",
			Label:       "Service PID 文件",
			Path:        "/tmp/hatch",
			Cleanable:   true,
			Description: "service 类脚本运行时的 PID 文件。停止后会自动清理，stale 文件可手动清理。",
		},
		{
			Key:         "backups",
			Label:       "备份导出",
			Path:        filepath.Join(home, "Desktop"),
			Cleanable:   false, // backups live on Desktop, user manages them
			Description: "ExportBackup 默认输出目录（~/Desktop/hatch-backup-*.zip）。",
		},
	}

	for i := range entries {
		statEntry(&entries[i])
	}
	return entries
}

func statEntry(e *DataDirEntry) {
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

// CleanupDataDir cleans up generated data by key.
//
//	key="traces"  → delete all .jsonl in traces dir
//	key="runs"    → delete all .log/.json in runs subdirs
//	key="pids"    → delete stale (process-dead) /tmp/hatch/*.pid files
//
// Returns a short summary.
func (a *App) CleanupDataDir(key string) (string, error) {
	switch key {
	case "traces":
		a.ClearAITraces()
		return "AI 调用日志已全部清理。", nil
	case "runs":
		root := executor.HistoryRoot()
		count := 0
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
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
		return fmt.Sprintf("脚本运行历史已清理 %d 个文件。", count), nil
	case "pids":
		entries, err := os.ReadDir("/tmp/hatch")
		if err != nil {
			return "PID 目录不存在，无需清理。", nil
		}
		removed := 0
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".pid") {
				continue
			}
			path := filepath.Join("/tmp/hatch", e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			pidStr := strings.TrimSpace(string(data))
			// quick liveness check via /proc-equivalent: just try a sig 0
			if !pidAlive(pidStr) {
				if os.Remove(path) == nil {
					removed++
				}
			}
		}
		return fmt.Sprintf("Stale PID 文件已清理 %d 个。", removed), nil
	case "memo-files":
		count, err := a.CleanupMemoFiles()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("备忘录附件已清理 %d 个文件。", count), nil
	default:
		return "", fmt.Errorf("不支持的清理目标: %s", key)
	}
}

func pidAlive(pidStr string) bool {
	if pidStr == "" {
		return false
	}
	pid := 0
	for _, c := range pidStr {
		if c < '0' || c > '9' {
			return false
		}
		pid = pid*10 + int(c-'0')
	}
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Sending signal 0 on darwin checks existence without delivering.
	return p.Signal(syscall.Signal(0)) == nil
}

// ─── Retention policy ────────────────────────────────────────────────────────

// GetRetentionPolicy returns the current effective retention settings.
type RetentionView struct {
	TraceRetainDays int `json:"traceRetainDays"`
	RunsRetainDays  int `json:"runsRetainDays"`
	RunsKeepPerName int `json:"runsKeepPerName"`
}

func (a *App) GetRetentionPolicy() RetentionView {
	a.mu.RLock()
	defer a.mu.RUnlock()
	v := RetentionView{
		TraceRetainDays: a.cfg.AI.TraceRetainDays,
		RunsRetainDays:  a.cfg.Retention.RunsRetainDays,
		RunsKeepPerName: a.cfg.Retention.RunsKeepPerName,
	}
	if v.TraceRetainDays <= 0 {
		v.TraceRetainDays = defaultTraceRetainDays
	}
	if v.RunsRetainDays <= 0 {
		v.RunsRetainDays = 14
	}
	if v.RunsKeepPerName <= 0 {
		v.RunsKeepPerName = 50
	}
	return v
}

// SaveRetentionPolicy persists new retention values and immediately applies them.
func (a *App) SaveRetentionPolicy(traceDays, runsDays, runsKeepPerName int) error {
	a.mu.Lock()
	a.cfg.AI.TraceRetainDays = traceDays
	a.cfg.Retention.RunsRetainDays = runsDays
	a.cfg.Retention.RunsKeepPerName = runsKeepPerName
	err := a.cfg.Save()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	go a.ApplyRetention()
	return nil
}

// ApplyRetention runs both pruners according to the current policy.
// Safe to call at any time; called on startup and after SaveRetentionPolicy.
func (a *App) ApplyRetention() {
	v := a.GetRetentionPolicy()
	if dir, err := a.traceDir(); err == nil {
		pruneTraceFiles(dir, v.TraceRetainDays)
	}
	executor.PruneHistoryOlderThan(v.RunsRetainDays)
	executor.PruneAllHistory(v.RunsKeepPerName)
	a.applyMemoRetention()
}

// ─── Skill AI optimization ───────────────────────────────────────────────────

// SuggestSkillContent asks the configured AI to optimize a SKILL.md.
// Returns the rewritten content as plain markdown (no fences).
func (a *App) SuggestSkillContent(name, currentContent, instruction string) (string, error) {
	a.mu.RLock()
	ai := a.activeAIConfig()
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		return "", err
	}

	system := `你是 Hatch 的 Skill 优化助手。Skill 是给 Claude Code / Codex 等 agent 使用的指令文档（SKILL.md）。
请根据用户指令改进或重写 SKILL.md。要求：
1. 只输出完整 SKILL.md 文本，不要解释，不要 Markdown 代码块（不要用三个反引号包裹整个输出）。
2. 必须保留或补全 frontmatter:
   ---
   name: <skill-name>
   description: <一句话触发条件描述>
   ---
3. description 字段决定 agent 何时触发本 skill —— 要包含具体的触发词或场景，越精准越好。
4. 正文用清晰的 Markdown 二级标题分块（## 何时使用 / ## 工作流程 / ## 注意事项 等）。
5. 保留已有的关键信息（具体命令、文件路径、特殊步骤），除非指令明确要求删除。`

	userBytes, _ := json.Marshal(map[string]any{
		"name":           name,
		"currentContent": currentContent,
		"instruction":    instruction,
	})
	raw, err := a.callAIOp("SuggestSkillContent", ai, system, string(userBytes))
	if err != nil {
		return "", err
	}
	raw = strings.TrimSpace(raw)
	for _, fence := range []string{"```markdown", "```md", "```"} {
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

// SuggestSkillFromDescription generates a brand-new SKILL.md from a one-line desc.
func (a *App) SuggestSkillFromDescription(name, description string) (string, error) {
	return a.SuggestSkillContent(name, "", "为这个新 skill 起草一份完整的 SKILL.md。description: "+description)
}

// ListRecentRunsSorted returns a flat list of newest runs across all scripts.
func (a *App) ListRecentRunsSorted(limit int) []executor.RunRecord {
	root := executor.HistoryRoot()
	var all []executor.RunRecord
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var r executor.RunRecord
		if json.Unmarshal(data, &r) == nil {
			all = append(all, r)
		}
		return nil
	})
	sort.Slice(all, func(i, j int) bool { return all[i].StartedAt.After(all[j].StartedAt) })
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all
}

// _ ensures time is imported (used in some retention helpers above)
var _ = time.Now

// ─── App data paths ──────────────────────────────────────────────────────────

// AppDataPaths holds all configurable directory paths.
type AppDataPaths struct {
	TraceLogDir      string `json:"traceLogDir"`
	TraceLogDirIsSet bool   `json:"traceLogDirIsSet"`
	SkillsDir        string `json:"skillsDir"`
	SkillsDirIsSet   bool   `json:"skillsDirIsSet"`
	DefaultTraceDir  string `json:"defaultTraceDir"`
	DefaultSkillsDir string `json:"defaultSkillsDir"`
}

// GetAppDataPaths returns current path configuration with effective values.
func (a *App) GetAppDataPaths() AppDataPaths {
	a.mu.RLock()
	traceCfg := a.cfg.TraceLogDir
	skillsCfg := a.cfg.SkillsDir
	a.mu.RUnlock()

	defTrace := defaultTraceDir()
	defSkills := defaultSkillsDir()

	return AppDataPaths{
		TraceLogDir:      orDefault(traceCfg, defTrace),
		TraceLogDirIsSet: traceCfg != "",
		SkillsDir:        orDefault(skillsCfg, defSkills),
		SkillsDirIsSet:   skillsCfg != "",
		DefaultTraceDir:  defTrace,
		DefaultSkillsDir: defSkills,
	}
}

func orDefault(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

// SaveAppDataPaths saves custom directory overrides (empty string = reset to default).
func (a *App) SaveAppDataPaths(traceLogDir, skillsDir string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.TraceLogDir = strings.TrimSpace(traceLogDir)
	a.cfg.SkillsDir = strings.TrimSpace(skillsDir)
	return a.cfg.Save()
}

// GetAutoLaunchEnabled reports whether Hatch is configured to launch at login.
func (a *App) GetAutoLaunchEnabled() bool {
	return autostart.Enabled()
}

// SetAutoLaunchEnabled enables or disables Hatch auto-launch at login.
func (a *App) SetAutoLaunchEnabled(enabled bool) error {
	if err := autostart.SetEnabled(enabled); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.AutoLaunch = enabled
	err := a.cfg.Save()
	a.mu.Unlock()
	return err
}

// ─── Doc links ───────────────────────────────────────────────────────────────

// GetDocLinks returns the configured quick-access doc links.
func (a *App) GetDocLinks() []config.DocLink {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]config.DocLink, len(a.cfg.DocLinks))
	copy(out, a.cfg.DocLinks)
	return out
}

// SaveDocLinks persists the doc links list.
func (a *App) SaveDocLinks(links []config.DocLink) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.DocLinks = links
	return a.cfg.Save()
}
