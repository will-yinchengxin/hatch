package app

import (
	"strings"

	"hatch/core/macos"
	"hatch/vpn"
)

// ─── 🔒 VPN（Clash Verge Rev） ───────────────────────────────────────────────
//
// Every method delegates to vpn.Manager built from the two configured dirs.

func (a *App) vpn() vpn.Manager {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return vpn.Manager{AppDir: a.cfg.VPNAppDir, BackupDir: a.cfg.VPNBackupDir}
}

func (a *App) GetVPNSettings() vpn.Settings { return a.vpn().ResolvedSettings() }

// SaveVPNSettings sets the Clash data dir and backup dir ("" = auto/default).
func (a *App) SaveVPNSettings(appDir, backupDir string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.VPNAppDir = strings.TrimSpace(appDir)
	a.cfg.VPNBackupDir = strings.TrimSpace(backupDir)
	return a.cfg.Save()
}

func (a *App) ValidateVPNSettings(appDir, backupDir string) vpn.PathValidation {
	return a.vpn().ValidatePaths(appDir, backupDir)
}

func (a *App) GetVPNHealthReport() vpn.HealthReport { return a.vpn().CheckHealth() }

// GetVPNStatus reads proxy/TUN flags, traffic and live node selections.
func (a *App) GetVPNStatus() vpn.Status { return a.vpn().ReadStatus() }

// GetVPNCurrentIP fetches the exit IP via Clash's HTTP proxy (direct fallback).
func (a *App) GetVPNCurrentIP() vpn.CurrentIP { return a.vpn().CurrentIP() }

func (a *App) GetVPNAppDir() string { return a.vpn().EffectiveAppDir() }

func (a *App) GetVPNProfiles() ([]vpn.Profile, error) { return a.vpn().Profiles() }

func (a *App) GetVPNConfigFile(name string) (vpn.ConfigFile, error) {
	return a.vpn().ReadConfigFile(name)
}

// SaveVPNConfigFile validates YAML, backs up and writes one allowed config file.
func (a *App) SaveVPNConfigFile(name, content string) error {
	return a.vpn().WriteConfigFile(name, content)
}

func (a *App) GetVPNProfileContent(filename string) (string, error) {
	return a.vpn().ReadProfile(filename)
}

func (a *App) SaveVPNProfileContent(filename, content string) error {
	return a.vpn().WriteProfile(filename, content)
}

// ── 自定义规则 / 订阅 ──

// GetVPNRules loads the current subscription's custom rules, hosts and DNS overrides.
func (a *App) GetVPNRules() (vpn.RulesState, error) { return a.vpn().ReadRules() }

// SaveVPNRules validates the draft, hot-reloads the core and writes the enhancement files.
func (a *App) SaveVPNRules(draft vpn.RulesDraft) (vpn.RulesSaveResult, error) {
	return a.vpn().SaveRules(draft)
}

func (a *App) GetVPNRuleTemplates() []vpn.RuleTemplate { return vpn.RuleTemplates() }

// ExpandVPNRuleTemplate fills a scenario template; the result is only added to the editor draft.
func (a *App) ExpandVPNRuleTemplate(id string, inputs map[string]string) (vpn.RulesDraft, error) {
	return vpn.ExpandTemplate(id, inputs)
}

// ParseVPNRuleLines turns pasted lines / bare domains into rules pointing at target.
func (a *App) ParseVPNRuleLines(text, target string) []vpn.RuleItem {
	return vpn.ParseRuleLines(text, target)
}

// TestVPNRule reports which rule (of the unsaved draft) a URL / domain / IP would hit.
func (a *App) TestVPNRule(input string, draft vpn.RulesDraft) (vpn.RuleMatch, error) {
	return a.vpn().TestRule(input, draft)
}

func (a *App) GetVPNSubscriptions() (vpn.SubscriptionsView, error) { return a.vpn().Subscriptions() }

func (a *App) GetVPNSubscriptionURL(uid string) (string, error) { return a.vpn().SubscriptionURL(uid) }

// OpenVPNApp brings Clash Verge Rev to the front.
func (a *App) OpenVPNApp() error { return vpn.OpenApp() }

func (a *App) ExportVPNBackup() (vpn.BackupResult, error) { return a.vpn().ExportBackup() }

func (a *App) PreviewVPNBackup(zipPath string) (vpn.BackupPreview, error) {
	return a.vpn().PreviewBackup(zipPath)
}

// ImportVPNBackup restores a backup zip (taking a pre-restore backup first).
func (a *App) ImportVPNBackup(zipPath string) (string, error) { return a.vpn().ImportBackup(zipPath) }

func (a *App) ListVPNBackups() []string { return a.vpn().ListBackups() }

func (a *App) DeleteVPNBackup(zipPath string) error { return a.vpn().DeleteBackup(zipPath) }

// OpenVPNDir opens the Clash Verge Rev data directory in Finder.
func (a *App) OpenVPNDir() error {
	dir, err := a.vpn().DataDir()
	if err != nil {
		return err
	}
	return macos.Open(dir)
}

// OpenVPNBackupDir opens the VPN backup directory in Finder.
func (a *App) OpenVPNBackupDir() error {
	dir, err := a.vpn().EnsureBackupDir()
	if err != nil {
		return err
	}
	return macos.Open(dir)
}
