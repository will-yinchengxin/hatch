// Package vpn reads and edits Clash Verge Rev state for the VPN page: proxy /
// TUN status, live node selections via the core's Unix-socket API, profiles,
// raw config files and zip backups. All paths derive from Manager.
package vpn

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hatch/core/fsutil"
	"hatch/core/strutil"

	"gopkg.in/yaml.v3"
)

// clashVergeAppID is the Clash Verge Rev data dir name under ~/Library/Application Support.
const clashVergeAppID = "io.github.clash-verge-rev.clash-verge-rev"

// ─── Types ───────────────────────────────────────────────────────────────────

type Profile struct {
	UID       string `json:"uid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	File      string `json:"file"`
	URL       string `json:"url,omitempty"`
	Updated   int64  `json:"updated"`
	IsCurrent bool   `json:"isCurrent"`
}

type ConfigFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type BackupResult struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	NumFiles int    `json:"numFiles"`
}

type BackupPreview struct {
	Path     string   `json:"path"`
	Valid    bool     `json:"valid"`
	NumFiles int      `json:"numFiles"`
	KeyFiles []string `json:"keyFiles"`
	Warnings []string `json:"warnings"`
}

type Settings struct {
	AppDir      string `json:"appDir"`
	BackupDir   string `json:"backupDir"`
	AppDirIsSet bool   `json:"appDirIsSet"`
	BackupIsSet bool   `json:"backupIsSet"`
}

type PathValidation struct {
	AppDirOK    bool     `json:"appDirOk"`
	BackupDirOK bool     `json:"backupDirOk"`
	Errors      []string `json:"errors"`
	Warnings    []string `json:"warnings"`
}

type ProxyGroup struct {
	Name  string `json:"name"`
	Node  string `json:"node"`
	Delay int    `json:"delay"` // ms of the selected node's last probe; 0 unknown, -1 failed
	Type  string `json:"type"`  // select | url-test | fallback | load-balance
}

type Status struct {
	SystemProxyEnabled bool         `json:"systemProxyEnabled"`
	TUNEnabled         bool         `json:"tunEnabled"`
	TUNInterface       string       `json:"tunInterface"`     // e.g. "utun3: 198.18.0.1"
	TUNExcludedCIDRs   []string     `json:"tunExcludedCidrs"` // from clash-verge.yaml tun config
	CurrentGroup       string       `json:"currentGroup"`     // primary proxy group name
	CurrentNode        string       `json:"currentNode"`      // selected node label
	ProxyGroups        []ProxyGroup `json:"proxyGroups"`      // all proxy group selections
	ExpireAt           int64        `json:"expireAt"`
	Upload             int64        `json:"upload"`
	Download           int64        `json:"download"`
	Total              int64        `json:"total"`
	// Live core data (empty when the core API is unreachable)
	CoreOnline      bool   `json:"coreOnline"`
	CoreVersion     string `json:"coreVersion"`
	Mode            string `json:"mode"` // rule | global | direct
	Connections     int    `json:"connections"`
	SessionUpload   int64  `json:"sessionUpload"` // since the core started
	SessionDownload int64  `json:"sessionDownload"`
	NodeDelay       int    `json:"nodeDelay"` // current node, same convention as ProxyGroup.Delay
}

type CurrentIP struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	Org     string `json:"org"`
	Error   string `json:"error,omitempty"`
}

type HealthCheck struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Status     string `json:"status"` // ok | warn | error
	Detail     string `json:"detail"`
	Suggestion string `json:"suggestion,omitempty"`
}

type HealthAction struct {
	Label string `json:"label"`
	Tab   string `json:"tab,omitempty"`
}

type HealthReport struct {
	Overall    string         `json:"overall"` // ok | warn | error
	Summary    string         `json:"summary"`
	Checks     []HealthCheck  `json:"checks"`
	Actions    []HealthAction `json:"actions"`
	LastBackup string         `json:"lastBackup,omitempty"`
}

// ─── Manager ─────────────────────────────────────────────────────────────────

// Manager is configured from config.VPNAppDir / config.VPNBackupDir; empty
// fields mean auto-detect / default.
type Manager struct {
	AppDir    string
	BackupDir string
	sockets   []string // tests only: replaces core socket discovery
	// tests only: skip `verge-mihomo -t` (fixtures have no geodata)
	skipCoreCheck bool
}

// DefaultAppDir is the auto-detected Clash Verge Rev data dir.
func DefaultAppDir() string {
	cfgDir, _ := os.UserConfigDir()
	return filepath.Join(cfgDir, clashVergeAppID)
}

// DefaultBackupDir is ~/Library/Application Support/Hatch/vpn-backups.
func DefaultBackupDir() string {
	return fsutil.AppDataPath("vpn-backups")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// DataDir returns the configured data dir when it exists, else the
// auto-detected one; it errors when neither is present.
func (m Manager) DataDir() (string, error) {
	if m.AppDir != "" {
		if _, err := os.Stat(m.AppDir); err == nil {
			return m.AppDir, nil
		}
	}
	if _, err := os.UserConfigDir(); err != nil {
		return "", fmt.Errorf("无法定位用户配置目录: %w", err)
	}
	dir := DefaultAppDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("未找到 Clash Verge Rev 数据目录: %s", dir)
	}
	return dir, nil
}

// EnsureBackupDir returns (and creates) the configured or default backup dir.
func (m Manager) EnsureBackupDir() (string, error) {
	return fsutil.EnsureDir(strutil.OrDefault(m.BackupDir, DefaultBackupDir()))
}

func safeProfilePath(uid string) bool {
	// Block path traversal; allow dots (extensions like .yaml, .js are fine)
	return !strings.Contains(uid, "..") && !strings.ContainsAny(uid, "/\\")
}

// validateYAML rejects content that is not parseable YAML, so a stray typo
// cannot be written back and silently break the Clash core on next reload.
func validateYAML(content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	var v interface{}
	if err := yaml.Unmarshal([]byte(content), &v); err != nil {
		return fmt.Errorf("YAML 格式错误，未保存: %w", err)
	}
	return nil
}

// ─── Settings ────────────────────────────────────────────────────────────────

// ResolvedSettings returns the effective dirs and whether each was set by hand.
func (m Manager) ResolvedSettings() Settings {
	return Settings{
		AppDir:      strutil.OrDefault(m.AppDir, DefaultAppDir()),
		BackupDir:   strutil.OrDefault(m.BackupDir, DefaultBackupDir()),
		AppDirIsSet: m.AppDir != "",
		BackupIsSet: m.BackupDir != "",
	}
}

// ValidatePaths checks candidate dirs (empty = current effective value).
func (m Manager) ValidatePaths(appDir, backupDir string) PathValidation {
	appDir = strings.TrimSpace(appDir)
	backupDir = strings.TrimSpace(backupDir)
	settings := m.ResolvedSettings()
	if appDir == "" {
		appDir = settings.AppDir
	}
	if backupDir == "" {
		backupDir = settings.BackupDir
	}

	var result PathValidation
	if fi, err := os.Stat(appDir); err != nil {
		result.Errors = append(result.Errors, "Clash Verge Rev 数据目录不存在: "+appDir)
	} else if !fi.IsDir() {
		result.Errors = append(result.Errors, "Clash Verge Rev 数据目录不是文件夹: "+appDir)
	} else {
		result.AppDirOK = true
		for _, name := range []string{"verge.yaml", "profiles.yaml"} {
			if _, err := os.Stat(filepath.Join(appDir, name)); err != nil {
				result.Warnings = append(result.Warnings, "未找到 "+name+"，请确认这是 Clash Verge Rev 数据目录。")
			}
		}
	}

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		result.Errors = append(result.Errors, "备份目录不可创建: "+err.Error())
	} else {
		testPath := filepath.Join(backupDir, ".hatch-write-test")
		if err := os.WriteFile(testPath, []byte("ok"), 0o600); err != nil {
			result.Errors = append(result.Errors, "备份目录不可写: "+err.Error())
		} else {
			_ = os.Remove(testPath)
			result.BackupDirOK = true
		}
	}

	return result
}

// CheckHealth runs the VPN 体检 checks: data dir, profile, mode, core API,
// node, subscription expiry and backups.
func (m Manager) CheckHealth() HealthReport {
	report := HealthReport{Overall: "ok"}
	addCheck := func(check HealthCheck) {
		report.Checks = append(report.Checks, check)
		switch check.Status {
		case "error":
			report.Overall = "error"
		case "warn":
			if report.Overall != "error" {
				report.Overall = "warn"
			}
		}
	}

	dir, dirErr := m.DataDir()
	if m.AppDir != "" {
		if fi, err := os.Stat(m.AppDir); err != nil || !fi.IsDir() {
			addCheck(HealthCheck{
				ID:         "app-dir",
				Label:      "Clash 数据目录",
				Status:     "error",
				Detail:     "手动设置的数据目录不可用: " + m.AppDir,
				Suggestion: "在“高级”里清空该路径使用自动检测，或填写正确的数据目录。",
			})
			report.Summary = "手动设置的 Clash Verge Rev 数据目录不可用。"
			report.Actions = []HealthAction{{Label: "检查目录设置", Tab: "advanced"}}
			return report
		}
	}
	if dirErr != nil {
		addCheck(HealthCheck{
			ID:         "app-dir",
			Label:      "Clash 数据目录",
			Status:     "error",
			Detail:     dirErr.Error(),
			Suggestion: "先安装并启动 Clash Verge Rev，或在“高级”里填写正确的数据目录。",
		})
		report.Summary = "未找到 Clash Verge Rev 配置，Hatch 还不能读取 VPN 状态。"
		report.Actions = []HealthAction{{Label: "检查目录设置", Tab: "advanced"}}
		return report
	}
	addCheck(HealthCheck{
		ID:     "app-dir",
		Label:  "Clash 数据目录",
		Status: "ok",
		Detail: dir,
	})

	profiles, profilesErr := m.Profiles()
	if profilesErr != nil {
		addCheck(HealthCheck{
			ID:         "profile",
			Label:      "当前订阅",
			Status:     "error",
			Detail:     profilesErr.Error(),
			Suggestion: "打开 Clash Verge Rev 检查 Profile 列表，或从备份恢复配置。",
		})
	} else {
		current := ""
		for _, p := range profiles {
			if p.IsCurrent {
				current = strutil.OrDefault(p.Name, p.UID)
				break
			}
		}
		if current == "" {
			addCheck(HealthCheck{
				ID:         "profile",
				Label:      "当前订阅",
				Status:     "warn",
				Detail:     fmt.Sprintf("检测到 %d 个 Profile，但没有激活项。", len(profiles)),
				Suggestion: "在 Clash Verge Rev 里选择一个可用 Profile。",
			})
		} else {
			addCheck(HealthCheck{
				ID:     "profile",
				Label:  "当前订阅",
				Status: "ok",
				Detail: current,
			})
		}
	}

	status := m.ReadStatus()
	if !status.SystemProxyEnabled && !status.TUNEnabled {
		addCheck(HealthCheck{
			ID:         "mode",
			Label:      "代理模式",
			Status:     "warn",
			Detail:     "系统代理和 TUN 都未开启。",
			Suggestion: "如果浏览器无法访问外网，请在 Clash Verge Rev 中开启系统代理或 TUN。",
		})
	} else {
		mode := []string{}
		if status.SystemProxyEnabled {
			mode = append(mode, "系统代理")
		}
		if status.TUNEnabled {
			mode = append(mode, "TUN")
		}
		addCheck(HealthCheck{
			ID:     "mode",
			Label:  "代理模式",
			Status: "ok",
			Detail: strings.Join(mode, " + "),
		})
	}

	if _, err := m.queryAPI("/version"); err != nil {
		addCheck(HealthCheck{
			ID:         "core",
			Label:      "Clash 核心连接",
			Status:     "warn",
			Detail:     "无法连接 Clash 核心 API。",
			Suggestion: "如果 Clash 正在运行，请等待几秒后重新体检；节点信息可能只能显示缓存值。",
		})
	} else {
		addCheck(HealthCheck{
			ID:     "core",
			Label:  "Clash 核心连接",
			Status: "ok",
			Detail: "可读取实时节点和配置。",
		})
	}

	if status.CurrentNode == "" {
		addCheck(HealthCheck{
			ID:         "node",
			Label:      "当前节点",
			Status:     "warn",
			Detail:     "未读到当前代理节点。",
			Suggestion: "打开 Clash Verge Rev 确认节点选择是否正常。",
		})
	} else {
		addCheck(HealthCheck{
			ID:     "node",
			Label:  "当前节点",
			Status: "ok",
			Detail: status.CurrentNode,
		})
	}

	if status.ExpireAt > 0 {
		days := int(time.Until(time.Unix(status.ExpireAt, 0)).Hours() / 24)
		switch {
		case days < 0:
			addCheck(HealthCheck{ID: "subscription", Label: "订阅到期", Status: "error", Detail: "订阅已过期。", Suggestion: "续费或更换可用订阅。"})
		case days <= 7:
			addCheck(HealthCheck{ID: "subscription", Label: "订阅到期", Status: "warn", Detail: fmt.Sprintf("约 %d 天后到期。", days), Suggestion: "建议提前续费，避免突然不可用。"})
		default:
			addCheck(HealthCheck{ID: "subscription", Label: "订阅到期", Status: "ok", Detail: fmt.Sprintf("约 %d 天后到期。", days)})
		}
	}

	backups := m.ListBackups()
	if len(backups) == 0 {
		addCheck(HealthCheck{
			ID:         "backup",
			Label:      "配置备份",
			Status:     "warn",
			Detail:     "还没有 Hatch 管理的 VPN 备份。",
			Suggestion: "建议先创建一份备份，后续改错配置可以快速恢复。",
		})
		report.Actions = append(report.Actions, HealthAction{Label: "创建备份", Tab: "backup"})
	} else {
		report.LastBackup = backups[0]
		addCheck(HealthCheck{ID: "backup", Label: "配置备份", Status: "ok", Detail: filepath.Base(backups[0])})
	}

	switch report.Overall {
	case "ok":
		report.Summary = "VPN 配置看起来正常。"
	case "warn":
		report.Summary = "VPN 可读取，但有几项建议检查。"
	default:
		report.Summary = "VPN 配置存在需要处理的问题。"
	}
	report.Actions = append(report.Actions, HealthAction{Label: "重新体检", Tab: "overview"})
	return report
}

// ─── Status ──────────────────────────────────────────────────────────────────

// GetVPNStatus reads verge.yaml + profiles.yaml to surface proxy/TUN state and active node.
func (m Manager) ReadStatus() Status {
	dir, err := m.DataDir()
	if err != nil {
		return Status{}
	}

	var status Status

	// Parse verge.yaml for proxy/TUN flags
	if data, err := os.ReadFile(filepath.Join(dir, "verge.yaml")); err == nil {
		var verge struct {
			EnableSystemProxy bool `yaml:"enable_system_proxy"`
			EnableTUNMode     bool `yaml:"enable_tun_mode"`
		}
		if yaml.Unmarshal(data, &verge) == nil {
			status.SystemProxyEnabled = verge.EnableSystemProxy
			status.TUNEnabled = verge.EnableTUNMode
		}
	}

	// Parse profiles.yaml for traffic stats (upload/download/total/expire)
	if data, err := os.ReadFile(filepath.Join(dir, "profiles.yaml")); err == nil {
		var raw struct {
			Current string `yaml:"current"`
			Items   []struct {
				UID   string `yaml:"uid"`
				Extra struct {
					Upload   int64 `yaml:"upload"`
					Download int64 `yaml:"download"`
					Total    int64 `yaml:"total"`
					Expire   int64 `yaml:"expire"`
				} `yaml:"extra"`
			} `yaml:"items"`
		}
		if yaml.Unmarshal(data, &raw) == nil {
			for _, item := range raw.Items {
				if item.UID == raw.Current {
					status.Upload = item.Extra.Upload
					status.Download = item.Extra.Download
					status.Total = item.Extra.Total
					status.ExpireAt = item.Extra.Expire
					break
				}
			}
		}
	}

	// Live proxy selections: query Clash core API via Unix socket.
	// Falls back to profiles.yaml selected[] if API is unavailable.
	if apiData, err := m.queryAPI("/proxies"); err == nil {
		var result struct {
			Proxies map[string]struct {
				Now     string `json:"now"`
				Type    string `json:"type"`
				Alive   *bool  `json:"alive"`
				History []struct {
					Delay int `json:"delay"`
				} `json:"history"`
			} `json:"proxies"`
		}
		if json.Unmarshal(apiData, &result) == nil {
			// Build set of all group names to distinguish nodes from sub-groups
			groupNames := map[string]bool{}
			for name, p := range result.Proxies {
				switch p.Type {
				case "Selector", "URLTest", "Fallback", "LoadBalance":
					groupNames[name] = true
				}
			}

			// delayOf follows group → selected member until a real node, then
			// reads its last probe.
			delayOf := func(name string) int {
				for i := 0; i < 8; i++ {
					p, ok := result.Proxies[name]
					if !ok {
						return 0
					}
					if groupNames[name] && p.Now != "" {
						name = p.Now
						continue
					}
					if p.Alive != nil && !*p.Alive {
						return -1
					}
					if n := len(p.History); n > 0 {
						if d := p.History[n-1].Delay; d > 0 {
							return d
						}
						return -1
					}
					return 0
				}
				return 0
			}
			groupType := map[string]string{"Selector": "select", "URLTest": "url-test", "Fallback": "fallback", "LoadBalance": "load-balance"}
			for name, p := range result.Proxies {
				if !groupNames[name] || name == "GLOBAL" {
					continue
				}
				status.ProxyGroups = append(status.ProxyGroups, ProxyGroup{
					Name:  name,
					Node:  p.Now,
					Delay: delayOf(p.Now),
					Type:  groupType[p.Type],
				})
			}
			// Sort: real server selections first, then meta (DIRECT/chain), alpha within each tier
			isRealNode := func(node string) bool {
				return node != "DIRECT" && node != "REJECT" && !groupNames[node]
			}
			// Manual (select) groups come first: the first one is "the" current node.
			sort.Slice(status.ProxyGroups, func(i, j int) bool {
				iSel, jSel := status.ProxyGroups[i].Type == "select", status.ProxyGroups[j].Type == "select"
				if iSel != jSel {
					return iSel
				}
				iReal := isRealNode(status.ProxyGroups[i].Node)
				jReal := isRealNode(status.ProxyGroups[j].Node)
				if iReal != jReal {
					return iReal
				}
				return status.ProxyGroups[i].Name < status.ProxyGroups[j].Name
			})
			if len(status.ProxyGroups) > 0 {
				status.CurrentGroup = status.ProxyGroups[0].Name
				status.CurrentNode = status.ProxyGroups[0].Node
				status.NodeDelay = status.ProxyGroups[0].Delay
			}
		}
		m.readCoreLive(&status)
	} else {
		// Fallback: read from profiles.yaml selected[] (may be stale)
		if data, err := os.ReadFile(filepath.Join(dir, "profiles.yaml")); err == nil {
			var raw struct {
				Current string `yaml:"current"`
				Items   []struct {
					UID      string `yaml:"uid"`
					Selected []struct {
						Name string `yaml:"name"`
						Now  string `yaml:"now"`
					} `yaml:"selected"`
				} `yaml:"items"`
			}
			if yaml.Unmarshal(data, &raw) == nil {
				for _, item := range raw.Items {
					if item.UID == raw.Current {
						for _, sel := range item.Selected {
							status.ProxyGroups = append(status.ProxyGroups, ProxyGroup{
								Name: sel.Name, Node: sel.Now,
							})
						}
						if len(item.Selected) > 0 {
							status.CurrentGroup = item.Selected[0].Name
							status.CurrentNode = item.Selected[0].Now
						}
						break
					}
				}
			}
		}
	}

	// Live TUN excluded routes from Clash API
	if apiData, err := m.queryAPI("/configs"); err == nil {
		var result struct {
			Mode string `json:"mode"`
			TUN  struct {
				RouteExcludeAddress []string `json:"route-exclude-address"`
			} `json:"tun"`
		}
		if json.Unmarshal(apiData, &result) == nil {
			status.Mode = strings.ToLower(result.Mode)
			if len(result.TUN.RouteExcludeAddress) > 0 {
				status.TUNExcludedCIDRs = result.TUN.RouteExcludeAddress
			}
		}
	}
	// Fallback: read excluded CIDRs from clash-verge.yaml (all known key variants)
	if len(status.TUNExcludedCIDRs) == 0 {
		if data, err := os.ReadFile(filepath.Join(dir, "clash-verge.yaml")); err == nil {
			var cvConfig struct {
				TUN struct {
					RouteExcludeAddress      []string `yaml:"route-exclude-address"`
					Inet4RouteExcludeAddress []string `yaml:"inet4-route-exclude-address"`
					ExcludeAddress           []string `yaml:"exclude-address"`
				} `yaml:"tun"`
			}
			if yaml.Unmarshal(data, &cvConfig) == nil {
				switch {
				case len(cvConfig.TUN.RouteExcludeAddress) > 0:
					status.TUNExcludedCIDRs = cvConfig.TUN.RouteExcludeAddress
				case len(cvConfig.TUN.Inet4RouteExcludeAddress) > 0:
					status.TUNExcludedCIDRs = cvConfig.TUN.Inet4RouteExcludeAddress
				case len(cvConfig.TUN.ExcludeAddress) > 0:
					status.TUNExcludedCIDRs = cvConfig.TUN.ExcludeAddress
				}
			}
		}
	}

	// Detect TUN virtual interface (Clash uses 198.18.0.0/15 subnet)
	if status.TUNEnabled {
		if iface, ip := detectTUNInterface(); iface != "" {
			status.TUNInterface = iface + ": " + ip
		}
	}

	return status
}

// readCoreLive fills core version, connection count and session traffic.
func (m Manager) readCoreLive(status *Status) {
	if data, err := m.queryAPI("/version"); err == nil {
		var v struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(data, &v) == nil && v.Version != "" {
			status.CoreOnline = true
			status.CoreVersion = v.Version
		}
	}
	if data, err := m.queryAPI("/connections"); err == nil {
		var c struct {
			UploadTotal   int64             `json:"uploadTotal"`
			DownloadTotal int64             `json:"downloadTotal"`
			Connections   []json.RawMessage `json:"connections"`
		}
		if json.Unmarshal(data, &c) == nil {
			status.Connections = len(c.Connections)
			status.SessionUpload = c.UploadTotal
			status.SessionDownload = c.DownloadTotal
		}
	}
}

// TestNodeDelay asks the core to probe a proxy or group now and returns the
// delay in ms (-1 when the probe failed).
func (m Manager) TestNodeDelay(name string) (int, error) {
	q := url.Values{"url": {"https://www.gstatic.com/generate_204"}, "timeout": {"2500"}}
	data, err := m.queryAPI("/proxies/" + url.PathEscape(name) + "/delay?" + q.Encode())
	if err != nil {
		return 0, fmt.Errorf("Clash 内核未连接: %w", err)
	}
	var r struct {
		Delay   int    `json:"delay"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return 0, err
	}
	if r.Delay <= 0 {
		return -1, nil
	}
	return r.Delay, nil
}

// detectTUNInterface finds the active Clash TUN virtual network interface by scanning
// for utun interfaces with IP in the 198.18.0.0/15 range (Clash Meta's TUN subnet).
func detectTUNInterface() (iface, ip string) {
	out, err := exec.Command("ifconfig").Output()
	if err != nil {
		return "", ""
	}
	lines := strings.Split(string(out), "\n")
	currentIface := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Interface header: "utun3: flags=..."
		if strings.HasPrefix(line, "utun") && strings.Contains(line, ":") {
			currentIface = strings.SplitN(line, ":", 2)[0]
			continue
		}
		if currentIface != "" && strings.HasPrefix(trimmed, "inet ") {
			fields := strings.Fields(trimmed)
			for i, f := range fields {
				if f == "inet" && i+1 < len(fields) {
					ipStr := fields[i+1]
					if strings.HasPrefix(ipStr, "198.18.") || strings.HasPrefix(ipStr, "198.19.") {
						return currentIface, ipStr
					}
				}
			}
		}
	}
	return "", ""
}

// unixSocketPath returns the first reachable core socket. In service mode
// Clash Verge Rev runs the core as root with its socket under
// /var/run/clash-verge-service/users/<uid>/, and the external-controller-unix
// in clash-verge.yaml is then stale, so every candidate is probed.
func (m Manager) unixSocketPath() string {
	if len(m.sockets) > 0 {
		return m.sockets[0]
	}
	var candidates []string
	if dir, _ := m.DataDir(); dir != "" {
		if data, err := os.ReadFile(filepath.Join(dir, "clash-verge.yaml")); err == nil {
			var cfg struct {
				ExternalControllerUnix string `yaml:"external-controller-unix"`
			}
			if yaml.Unmarshal(data, &cfg) == nil && cfg.ExternalControllerUnix != "" {
				candidates = append(candidates, cfg.ExternalControllerUnix)
			}
		}
	}
	candidates = append(candidates,
		fmt.Sprintf("/var/run/clash-verge-service/users/%d/verge-mihomo.sock", os.Getuid()),
		"/tmp/verge/verge-mihomo.sock")
	for _, p := range candidates {
		if conn, err := net.DialTimeout("unix", p, 300*time.Millisecond); err == nil {
			conn.Close()
			return p
		}
	}
	return candidates[0]
}

// queryClashUnixAPI makes a GET request to the Clash core API via Unix socket.
func (m Manager) queryAPI(path string) ([]byte, error) {
	socketPath := m.unixSocketPath()
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
	}
	resp, err := client.Get("http://localhost" + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// clashHTTPPort reads the HTTP proxy port from the active Clash profile's YAML.
// Falls back to 7890 (Clash default).
func (m Manager) httpPort() int {
	dir, err := m.DataDir()
	if err != nil {
		return 7890
	}

	// Read port from profiles.yaml → current profile file
	data, err := os.ReadFile(filepath.Join(dir, "profiles.yaml"))
	if err != nil {
		return 7890
	}
	var meta struct {
		Current string `yaml:"current"`
		Items   []struct {
			UID  string `yaml:"uid"`
			File string `yaml:"file"`
		} `yaml:"items"`
	}
	if yaml.Unmarshal(data, &meta) != nil {
		return 7890
	}
	for _, item := range meta.Items {
		if item.UID == meta.Current {
			profileData, err := os.ReadFile(filepath.Join(dir, "profiles", item.File))
			if err != nil {
				break
			}
			var cfg struct {
				MixedPort int `yaml:"mixed-port"`
				Port      int `yaml:"port"`
			}
			if yaml.Unmarshal(profileData, &cfg) == nil {
				if cfg.MixedPort > 0 {
					return cfg.MixedPort
				}
				if cfg.Port > 0 {
					return cfg.Port
				}
			}
			break
		}
	}
	return 7890
}

// GetVPNCurrentIP fetches the real exit IP and geolocation via ipinfo.io,
// routing through Clash's HTTP proxy so the result reflects the actual VPN exit.
func (m Manager) CurrentIP() CurrentIP {
	port := m.httpPort()
	proxyURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))

	// Primary: route through Clash HTTP proxy (reflects true exit IP)
	proxyClient := &http.Client{
		Timeout: 7 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	result, err := fetchIPInfo(proxyClient)
	if err == nil {
		return result
	}

	// Fallback: direct (works for TUN mode where all traffic is captured)
	directClient := &http.Client{Timeout: 7 * time.Second}
	result, err = fetchIPInfo(directClient)
	if err != nil {
		return CurrentIP{Error: err.Error()}
	}
	return result
}

func fetchIPInfo(client *http.Client) (CurrentIP, error) {
	req, _ := http.NewRequest("GET", "https://ipinfo.io/json", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return CurrentIP{}, err
	}
	defer resp.Body.Close()

	var result CurrentIP
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return CurrentIP{}, err
	}
	return result, nil
}

// ─── Profiles, config files & backups ─────────────────────────────────────────

// GetVPNAppDir returns the effective Clash Verge Rev application data directory path.
func (m Manager) EffectiveAppDir() string {
	dir, err := m.DataDir()
	if err != nil {
		return ""
	}
	return dir
}

// GetVPNProfiles parses profiles.yaml and returns all profiles with isCurrent marked.
func (m Manager) Profiles() ([]Profile, error) {
	dir, err := m.DataDir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(dir, "profiles.yaml"))
	if err != nil {
		return nil, fmt.Errorf("读取 profiles.yaml 失败: %w", err)
	}

	var raw struct {
		Current string `yaml:"current"`
		Items   []struct {
			UID     string `yaml:"uid"`
			Type    string `yaml:"type"`
			Name    string `yaml:"name"`
			File    string `yaml:"file"`
			URL     string `yaml:"url"`
			Updated int64  `yaml:"updated"`
		} `yaml:"items"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析 profiles.yaml 失败: %w", err)
	}

	profiles := make([]Profile, 0, len(raw.Items))
	for _, item := range raw.Items {
		profiles = append(profiles, Profile{
			UID:       item.UID,
			Type:      item.Type,
			Name:      item.Name,
			File:      item.File,
			URL:       item.URL,
			Updated:   item.Updated,
			IsCurrent: item.UID == raw.Current,
		})
	}
	return profiles, nil
}

// GetVPNConfigFile reads one of the named config files from the Clash Verge Rev directory.
func (m Manager) ReadConfigFile(name string) (ConfigFile, error) {
	allowed := map[string]bool{
		"verge.yaml":       true,
		"config.yaml":      true,
		"dns_config.yaml":  true,
		"profiles.yaml":    true,
		"clash-verge.yaml": true,
	}
	if !allowed[name] {
		return ConfigFile{}, fmt.Errorf("不允许访问的配置文件: %s", name)
	}

	dir, err := m.DataDir()
	if err != nil {
		return ConfigFile{}, err
	}

	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ConfigFile{Name: name, Path: path, Content: ""}, nil
		}
		return ConfigFile{}, fmt.Errorf("读取 %s 失败: %w", name, err)
	}
	return ConfigFile{Name: name, Path: path, Content: string(data)}, nil
}

// SaveVPNConfigFile writes content back to a named config file, creating a .bak first.
func (m Manager) WriteConfigFile(name, content string) error {
	allowed := map[string]bool{
		"verge.yaml":       true,
		"config.yaml":      true,
		"dns_config.yaml":  true,
		"profiles.yaml":    true,
		"clash-verge.yaml": true,
	}
	if !allowed[name] {
		return fmt.Errorf("不允许写入的配置文件: %s", name)
	}
	// All managed config files are YAML; reject broken content before overwriting.
	if err := validateYAML(content); err != nil {
		return err
	}

	dir, err := m.DataDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, name)
	if err := fsutil.BackupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// GetVPNProfileContent reads the raw content of a profile file by its file name.
func (m Manager) ReadProfile(filename string) (string, error) {
	if !safeProfilePath(filename) {
		return "", fmt.Errorf("非法 profile 文件名: %s", filename)
	}

	dir, err := m.DataDir()
	if err != nil {
		return "", err
	}

	profilesDir := filepath.Join(dir, "profiles")

	// If caller passes the full filename (e.g. "rHWynz2rOL7z.yaml"), try directly first
	if data, err := os.ReadFile(filepath.Join(profilesDir, filename)); err == nil {
		return string(data), nil
	}
	// Fallback: try appending extensions (uid without extension)
	for _, ext := range []string{".yaml", ".js"} {
		if data, err := os.ReadFile(filepath.Join(profilesDir, filename+ext)); err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("未找到 profile 文件: %s", filename)
}

// SaveVPNProfileContent writes updated content to a profile file.
func (m Manager) WriteProfile(filename, content string) error {
	if !safeProfilePath(filename) {
		return fmt.Errorf("非法 profile 文件名: %s", filename)
	}

	dir, err := m.DataDir()
	if err != nil {
		return err
	}

	profilesDir := filepath.Join(dir, "profiles")
	targetPath := filepath.Join(profilesDir, filename)

	// If exact path doesn't exist, try with extensions
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		for _, ext := range []string{".yaml", ".js"} {
			candidate := filepath.Join(profilesDir, filename+ext)
			if _, err := os.Stat(candidate); err == nil {
				targetPath = candidate
				break
			}
		}
	}

	// Validate YAML profiles before writing; script (.js) profiles are skipped.
	lower := strings.ToLower(targetPath)
	if strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") {
		if err := validateYAML(content); err != nil {
			return err
		}
	}

	if err := fsutil.BackupIfExists(targetPath); err != nil {
		return err
	}
	return os.WriteFile(targetPath, []byte(content), 0o644)
}

func (m Manager) exportBackup(prefix string) (BackupResult, error) {
	clashDir, err := m.DataDir()
	if err != nil {
		return BackupResult{}, err
	}

	backupDir, err := m.EnsureBackupDir()
	if err != nil {
		return BackupResult{}, err
	}

	stamp := time.Now().Format("20060102-150405")
	zipPath := filepath.Join(backupDir, prefix+stamp+".zip")

	f, err := os.Create(zipPath)
	if err != nil {
		return BackupResult{}, err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	numFiles := 0

	skipNames := map[string]bool{
		"cache.db": true, "geoip.dat": true, "geosite.dat": true, "Country.mmdb": true,
	}

	err = filepath.WalkDir(clashDir, func(path string, d os.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "logs" {
				return filepath.SkipDir
			}
			return nil
		}
		if skipNames[d.Name()] || strings.HasSuffix(d.Name(), ".dat") {
			return nil
		}
		rel, _ := filepath.Rel(clashDir, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		_, _ = w.Write(data)
		numFiles++
		return nil
	})
	if err != nil {
		return BackupResult{}, err
	}

	if err := zw.Close(); err != nil {
		return BackupResult{}, err
	}
	fi, _ := f.Stat()
	return BackupResult{Path: zipPath, Size: fi.Size(), NumFiles: numFiles}, nil
}

// ExportVPNBackup zips the Clash Verge Rev config directory.
func (m Manager) ExportBackup() (BackupResult, error) {
	return m.exportBackup("clash-verge-")
}

// PreviewVPNBackup inspects a backup zip before restore, without writing files.
func (m Manager) PreviewBackup(zipPath string) (BackupPreview, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return BackupPreview{}, fmt.Errorf("打开 zip 失败: %w", err)
	}
	defer r.Close()

	preview := BackupPreview{Path: zipPath}
	keySet := map[string]bool{}
	hasConfig := false
	hasProfilesDir := false
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := filepath.Clean(f.Name)
		if name == "." || strings.HasPrefix(name, ".."+string(os.PathSeparator)) || filepath.IsAbs(name) {
			preview.Warnings = append(preview.Warnings, "跳过可疑路径: "+f.Name)
			continue
		}
		preview.NumFiles++
		switch name {
		case "verge.yaml", "profiles.yaml", "config.yaml", "dns_config.yaml", "clash-verge.yaml":
			keySet[name] = true
			hasConfig = true
		default:
			if strings.HasPrefix(name, "profiles"+string(os.PathSeparator)) {
				hasProfilesDir = true
			}
		}
	}
	for name := range keySet {
		preview.KeyFiles = append(preview.KeyFiles, name)
	}
	sort.Strings(preview.KeyFiles)
	if !keySet["profiles.yaml"] {
		preview.Warnings = append(preview.Warnings, "未发现 profiles.yaml，恢复后可能没有订阅列表。")
	}
	if !hasProfilesDir {
		preview.Warnings = append(preview.Warnings, "未发现 profiles/ 目录，恢复后可能缺少订阅内容。")
	}
	preview.Valid = preview.NumFiles > 0 && (hasConfig || hasProfilesDir)
	if !preview.Valid {
		preview.Warnings = append(preview.Warnings, "这不像 Clash Verge Rev 配置备份。")
	}
	return preview, nil
}

// ImportVPNBackup restores Clash Verge Rev configs from a backup zip.
func (m Manager) ImportBackup(zipPath string) (string, error) {
	clashDir, err := m.DataDir()
	canBackupCurrent := err == nil
	preRestoreBackup := ""
	if err != nil {
		clashDir = DefaultAppDir()
		if mkErr := os.MkdirAll(clashDir, 0o755); mkErr != nil {
			return "", fmt.Errorf("无法创建目标目录: %w", mkErr)
		}
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("打开 zip 失败: %w", err)
	}
	defer r.Close()

	if canBackupCurrent {
		if r, backupErr := m.exportBackup("clash-verge-pre-restore-"); backupErr == nil {
			preRestoreBackup = r.Path
		}
	}

	restored := 0
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		dest := filepath.Join(clashDir, f.Name)
		cleanRoot := filepath.Clean(clashDir)
		cleanDest := filepath.Clean(dest)
		if cleanDest != cleanRoot && !strings.HasPrefix(cleanDest, cleanRoot+string(os.PathSeparator)) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			continue
		}
		restored++
	}

	if preRestoreBackup != "" {
		return fmt.Sprintf("已恢复 %d 个文件；恢复前备份: %s", restored, preRestoreBackup), nil
	}
	return fmt.Sprintf("已恢复 %d 个文件", restored), nil
}

// ListVPNBackups returns paths of all clash-verge-*.zip backups, newest first.
func (m Manager) ListBackups() []string {
	backupDir, err := m.EnsureBackupDir()
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil
	}

	var paths []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "clash-verge-") && strings.HasSuffix(e.Name(), ".zip") {
			paths = append(paths, filepath.Join(backupDir, e.Name()))
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	return paths
}

// DeleteVPNBackup removes a single clash-verge-*.zip backup from the managed
// backup directory. Refuses paths outside that dir or with an unexpected name.
func (m Manager) DeleteBackup(zipPath string) error {
	backupDir, err := m.EnsureBackupDir()
	if err != nil {
		return err
	}
	clean := filepath.Clean(zipPath)
	if filepath.Dir(clean) != filepath.Clean(backupDir) {
		return fmt.Errorf("只能删除备份目录内的文件")
	}
	base := filepath.Base(clean)
	if !strings.HasPrefix(base, "clash-verge-") || !strings.HasSuffix(base, ".zip") {
		return fmt.Errorf("非法的备份文件名: %s", base)
	}
	return os.Remove(clean)
}
