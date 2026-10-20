package app

import (
	"hatch/core/config"
)

// ─── 界面模块开关 ────────────────────────────────────────────────────────────
//
// Which top-level tabs (工作台 / AI / VPN / 备忘录) are shown is controlled by
// the "modules" object in config.json. Hiding a module only hides its UI;
// background behavior (schedules, memo reminders, retention) keeps running.

// GetModuleSettings re-reads "modules" from config.json (so edits made in an
// external editor apply on the next refresh/focus) and returns the effective
// visibility. At least one module is always visible.
func (a *App) GetModuleSettings() config.Modules {
	mods, err := config.LoadModules()
	a.mu.Lock()
	defer a.mu.Unlock()
	if err == nil {
		a.cfg.Modules = mods
	}
	return a.cfg.Modules.Normalized()
}
