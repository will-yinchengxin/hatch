package app

import (
	"fmt"
	"strings"

	"hatch/ai/llm"
	"hatch/ai/skills"
	"hatch/core/autostart"
	"hatch/core/config"
	"hatch/core/datadir"
	"hatch/core/fsutil"
	"hatch/core/strutil"
	"hatch/workbench/executor"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// GetConfig returns the current application configuration.
func (a *App) GetConfig() *config.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

// ─── Data directory inventory ────────────────────────────────────────────────

// GetAppDataInventory returns every directory/file Hatch writes to, so the UI
// can show exactly what's on disk and offer cleanup.
func (a *App) GetAppDataInventory() []datadir.Entry {
	paths := a.GetAppDataPaths()
	return datadir.Inventory(datadir.Locations{
		HatchDir:  fsutil.AppDataPath(),
		TraceDir:  paths.TraceLogDir,
		RunsDir:   executor.HistoryRoot(),
		SkillsDir: paths.SkillsDir,
		MemoFiles: a.memoFilesDir(),
	})
}

// CleanupDataDir cleans up generated data by key (traces / runs / pids /
// memo-files) and returns a short summary.
func (a *App) CleanupDataDir(key string) (string, error) {
	switch key {
	case "traces":
		a.ClearAITraces()
		return "AI 调用日志已全部清理。", nil
	case "runs":
		return datadir.CleanRunHistory(executor.HistoryRoot()), nil
	case "pids":
		return datadir.CleanStalePIDs(), nil
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

// ListRecentRunsSorted returns a flat list of newest runs across all scripts.
func (a *App) ListRecentRunsSorted(limit int) []executor.RunRecord {
	return executor.ListRecentRuns(limit)
}

// ─── Retention policy ────────────────────────────────────────────────────────

// RetentionView is the effective retention settings (defaults filled in).
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
		v.TraceRetainDays = llm.DefaultTraceRetainDays
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

// ApplyRetention prunes AI traces, run history and (when enabled) memo
// attachments. Called on startup, hourly and after SaveRetentionPolicy.
func (a *App) ApplyRetention() {
	v := a.GetRetentionPolicy()
	if dir, err := a.traceDir(); err == nil {
		llm.PruneTraces(dir, v.TraceRetainDays)
	}
	executor.PruneHistoryOlderThan(v.RunsRetainDays)
	executor.PruneAllHistory(v.RunsKeepPerName)
	a.applyMemoRetention()
}

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

	defTrace := llm.DefaultTraceDir()
	defSkills := skills.DefaultDir()
	return AppDataPaths{
		TraceLogDir:      strutil.OrDefault(traceCfg, defTrace),
		TraceLogDirIsSet: traceCfg != "",
		SkillsDir:        strutil.OrDefault(skillsCfg, defSkills),
		SkillsDirIsSet:   skillsCfg != "",
		DefaultTraceDir:  defTrace,
		DefaultSkillsDir: defSkills,
	}
}

// SaveAppDataPaths saves custom directory overrides (empty string = reset to default).
func (a *App) SaveAppDataPaths(traceLogDir, skillsDir string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.TraceLogDir = strings.TrimSpace(traceLogDir)
	a.cfg.SkillsDir = strings.TrimSpace(skillsDir)
	return a.cfg.Save()
}

// ─── Auto launch ─────────────────────────────────────────────────────────────

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
