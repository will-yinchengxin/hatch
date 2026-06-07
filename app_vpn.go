package main

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

	"gopkg.in/yaml.v3"
)

const clashVergeAppID = "io.github.clash-verge-rev.clash-verge-rev"

// ─── Types ───────────────────────────────────────────────────────────────────

type VPNProfile struct {
	UID       string `json:"uid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	File      string `json:"file"`
	URL       string `json:"url,omitempty"`
	Updated   int64  `json:"updated"`
	IsCurrent bool   `json:"isCurrent"`
}

type VPNConfigFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type VPNBackupResult struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	NumFiles int    `json:"numFiles"`
}

type VPNSettings struct {
	AppDir      string `json:"appDir"`
	BackupDir   string `json:"backupDir"`
	AppDirIsSet bool   `json:"appDirIsSet"`
	BackupIsSet bool   `json:"backupIsSet"`
}

type VPNProxyGroup struct {
	Name string `json:"name"`
	Node string `json:"node"`
}

type VPNEnhanceData struct {
	PrependRules    []string          `json:"prependRules"`
	Hosts           map[string]string `json:"hosts"`
	DNSFakeIPFilter []string          `json:"dnsFakeIPFilter"`
	HasData         bool              `json:"hasData"`
	SourceFiles     []string          `json:"sourceFiles"`
}

type VPNStatus struct {
	SystemProxyEnabled bool            `json:"systemProxyEnabled"`
	TUNEnabled         bool            `json:"tunEnabled"`
	TUNInterface       string          `json:"tunInterface"`     // e.g. "utun3: 198.18.0.1"
	TUNExcludedCIDRs   []string        `json:"tunExcludedCidrs"` // from clash-verge.yaml tun config
	CurrentGroup       string          `json:"currentGroup"`     // primary proxy group name
	CurrentNode        string          `json:"currentNode"`      // selected node label
	ProxyGroups        []VPNProxyGroup `json:"proxyGroups"`      // all proxy group selections
	ExpireAt           int64           `json:"expireAt"`
	Upload             int64           `json:"upload"`
	Download           int64           `json:"download"`
	Total              int64           `json:"total"`
}

type VPNCurrentIP struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	Org     string `json:"org"`
	Error   string `json:"error,omitempty"`
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (a *App) clashVergeDir() (string, error) {
	// User-configured path takes priority
	if a.cfg.VPNAppDir != "" {
		if _, err := os.Stat(a.cfg.VPNAppDir); err == nil {
			return a.cfg.VPNAppDir, nil
		}
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("无法定位用户配置目录: %w", err)
	}
	dir := filepath.Join(cfgDir, clashVergeAppID)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("未找到 Clash Verge Rev 数据目录: %s", dir)
	}
	return dir, nil
}

func (a *App) vpnBackupDir() (string, error) {
	if a.cfg.VPNBackupDir != "" {
		if err := os.MkdirAll(a.cfg.VPNBackupDir, 0o755); err != nil {
			return "", err
		}
		return a.cfg.VPNBackupDir, nil
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfgDir, "Hatch", "vpn-backups")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
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

func (a *App) GetVPNSettings() VPNSettings {
	cfgDir, _ := os.UserConfigDir()
	defaultBackup := filepath.Join(cfgDir, "Hatch", "vpn-backups")
	defaultApp := filepath.Join(cfgDir, clashVergeAppID)

	return VPNSettings{
		AppDir:      orDefault(a.cfg.VPNAppDir, defaultApp),
		BackupDir:   orDefault(a.cfg.VPNBackupDir, defaultBackup),
		AppDirIsSet: a.cfg.VPNAppDir != "",
		BackupIsSet: a.cfg.VPNBackupDir != "",
	}
}

func (a *App) SaveVPNSettings(appDir, backupDir string) error {
	a.cfg.VPNAppDir = strings.TrimSpace(appDir)
	a.cfg.VPNBackupDir = strings.TrimSpace(backupDir)
	return a.cfg.Save()
}

// ─── Status ──────────────────────────────────────────────────────────────────

// GetVPNStatus reads verge.yaml + profiles.yaml to surface proxy/TUN state and active node.
func (a *App) GetVPNStatus() VPNStatus {
	dir, err := a.clashVergeDir()
	if err != nil {
		return VPNStatus{}
	}

	var status VPNStatus

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
	if apiData, err := a.queryClashUnixAPI("/proxies"); err == nil {
		var result struct {
			Proxies map[string]struct {
				Now  string `json:"now"`
				Type string `json:"type"`
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

			for name, p := range result.Proxies {
				if p.Type != "Selector" || name == "GLOBAL" {
					continue
				}
				status.ProxyGroups = append(status.ProxyGroups, VPNProxyGroup{
					Name: name,
					Node: p.Now,
				})
			}
			// Sort: real server selections first, then meta (DIRECT/chain), alpha within each tier
			isRealNode := func(node string) bool {
				return node != "DIRECT" && node != "REJECT" && !groupNames[node]
			}
			sort.Slice(status.ProxyGroups, func(i, j int) bool {
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
			}
		}
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
							status.ProxyGroups = append(status.ProxyGroups, VPNProxyGroup{
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
	if apiData, err := a.queryClashUnixAPI("/configs"); err == nil {
		var result struct {
			TUN struct {
				RouteExcludeAddress []string `json:"route-exclude-address"`
			} `json:"tun"`
		}
		if json.Unmarshal(apiData, &result) == nil && len(result.TUN.RouteExcludeAddress) > 0 {
			status.TUNExcludedCIDRs = result.TUN.RouteExcludeAddress
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

// clashUnixSocketPath reads external-controller-unix from clash-verge.yaml,
// falling back to the default Clash Verge Rev socket path.
func (a *App) clashUnixSocketPath() string {
	dir, _ := a.clashVergeDir()
	if dir != "" {
		if data, err := os.ReadFile(filepath.Join(dir, "clash-verge.yaml")); err == nil {
			var cfg struct {
				ExternalControllerUnix string `yaml:"external-controller-unix"`
			}
			if yaml.Unmarshal(data, &cfg) == nil && cfg.ExternalControllerUnix != "" {
				return cfg.ExternalControllerUnix
			}
		}
	}
	return "/tmp/verge/verge-mihomo.sock"
}

// queryClashUnixAPI makes a GET request to the Clash core API via Unix socket.
func (a *App) queryClashUnixAPI(path string) ([]byte, error) {
	socketPath := a.clashUnixSocketPath()
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
func (a *App) clashHTTPPort() int {
	dir, err := a.clashVergeDir()
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
func (a *App) GetVPNCurrentIP() VPNCurrentIP {
	port := a.clashHTTPPort()
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
		return VPNCurrentIP{Error: err.Error()}
	}
	return result
}

func fetchIPInfo(client *http.Client) (VPNCurrentIP, error) {
	req, _ := http.NewRequest("GET", "https://ipinfo.io/json", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return VPNCurrentIP{}, err
	}
	defer resp.Body.Close()

	var result VPNCurrentIP
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return VPNCurrentIP{}, err
	}
	return result, nil
}

// ─── Bound Methods ────────────────────────────────────────────────────────────

// GetVPNAppDir returns the effective Clash Verge Rev application data directory path.
func (a *App) GetVPNAppDir() string {
	dir, err := a.clashVergeDir()
	if err != nil {
		return ""
	}
	return dir
}

// GetVPNProfiles parses profiles.yaml and returns all profiles with isCurrent marked.
func (a *App) GetVPNProfiles() ([]VPNProfile, error) {
	dir, err := a.clashVergeDir()
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

	profiles := make([]VPNProfile, 0, len(raw.Items))
	for _, item := range raw.Items {
		profiles = append(profiles, VPNProfile{
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
func (a *App) GetVPNConfigFile(name string) (VPNConfigFile, error) {
	allowed := map[string]bool{
		"verge.yaml":       true,
		"config.yaml":      true,
		"dns_config.yaml":  true,
		"profiles.yaml":    true,
		"clash-verge.yaml": true,
	}
	if !allowed[name] {
		return VPNConfigFile{}, fmt.Errorf("不允许访问的配置文件: %s", name)
	}

	dir, err := a.clashVergeDir()
	if err != nil {
		return VPNConfigFile{}, err
	}

	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VPNConfigFile{Name: name, Path: path, Content: ""}, nil
		}
		return VPNConfigFile{}, fmt.Errorf("读取 %s 失败: %w", name, err)
	}
	return VPNConfigFile{Name: name, Path: path, Content: string(data)}, nil
}

// SaveVPNConfigFile writes content back to a named config file, creating a .bak first.
func (a *App) SaveVPNConfigFile(name, content string) error {
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

	dir, err := a.clashVergeDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, name)
	if err := backupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// GetVPNProfileContent reads the raw content of a profile file by its file name.
func (a *App) GetVPNProfileContent(filename string) (string, error) {
	if !safeProfilePath(filename) {
		return "", fmt.Errorf("非法 profile 文件名: %s", filename)
	}

	dir, err := a.clashVergeDir()
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
func (a *App) SaveVPNProfileContent(filename, content string) error {
	if !safeProfilePath(filename) {
		return fmt.Errorf("非法 profile 文件名: %s", filename)
	}

	dir, err := a.clashVergeDir()
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

	if err := backupIfExists(targetPath); err != nil {
		return err
	}
	return os.WriteFile(targetPath, []byte(content), 0o644)
}

// ExportVPNBackup zips the Clash Verge Rev config directory.
func (a *App) ExportVPNBackup() (VPNBackupResult, error) {
	clashDir, err := a.clashVergeDir()
	if err != nil {
		return VPNBackupResult{}, err
	}

	backupDir, err := a.vpnBackupDir()
	if err != nil {
		return VPNBackupResult{}, err
	}

	stamp := time.Now().Format("20060102-150405")
	zipPath := filepath.Join(backupDir, "clash-verge-"+stamp+".zip")

	f, err := os.Create(zipPath)
	if err != nil {
		return VPNBackupResult{}, err
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
		return VPNBackupResult{}, err
	}

	if err := zw.Close(); err != nil {
		return VPNBackupResult{}, err
	}
	fi, _ := f.Stat()
	return VPNBackupResult{Path: zipPath, Size: fi.Size(), NumFiles: numFiles}, nil
}

// ImportVPNBackup restores Clash Verge Rev configs from a backup zip.
func (a *App) ImportVPNBackup(zipPath string) (string, error) {
	clashDir, err := a.clashVergeDir()
	if err != nil {
		cfgDir, _ := os.UserConfigDir()
		clashDir = filepath.Join(cfgDir, clashVergeAppID)
		if mkErr := os.MkdirAll(clashDir, 0o755); mkErr != nil {
			return "", fmt.Errorf("无法创建目标目录: %w", mkErr)
		}
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("打开 zip 失败: %w", err)
	}
	defer r.Close()

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

	return fmt.Sprintf("已恢复 %d 个文件", restored), nil
}

// ListVPNBackups returns paths of all clash-verge-*.zip backups, newest first.
func (a *App) ListVPNBackups() []string {
	backupDir, err := a.vpnBackupDir()
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
func (a *App) DeleteVPNBackup(zipPath string) error {
	backupDir, err := a.vpnBackupDir()
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

// OpenVPNDir opens the Clash Verge Rev data directory in Finder.
func (a *App) OpenVPNDir() error {
	dir, err := a.clashVergeDir()
	if err != nil {
		return err
	}
	return exec.Command("open", dir).Run()
}

// OpenVPNBackupDir opens the VPN backup directory in Finder.
func (a *App) OpenVPNBackupDir() error {
	dir, err := a.vpnBackupDir()
	if err != nil {
		return err
	}
	return exec.Command("open", dir).Run()
}

// GetVPNEnhanceData reads all "merge"-type profiles and extracts prepend rules,
// hosts overrides, and DNS fake-ip-filter entries added by the user.
func (a *App) GetVPNEnhanceData() (VPNEnhanceData, error) {
	dir, err := a.clashVergeDir()
	if err != nil {
		return VPNEnhanceData{}, err
	}

	data, err := os.ReadFile(filepath.Join(dir, "profiles.yaml"))
	if err != nil {
		return VPNEnhanceData{}, fmt.Errorf("读取 profiles.yaml 失败: %w", err)
	}

	var raw struct {
		Items []struct {
			UID  string `yaml:"uid"`
			Type string `yaml:"type"`
			Name string `yaml:"name"`
			File string `yaml:"file"`
		} `yaml:"items"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return VPNEnhanceData{}, fmt.Errorf("解析 profiles.yaml 失败: %w", err)
	}

	result := VPNEnhanceData{
		Hosts:           make(map[string]string),
		PrependRules:    []string{},
		DNSFakeIPFilter: []string{},
		SourceFiles:     []string{},
	}

	for _, item := range raw.Items {
		// Both "merge" and "rules" profile types can carry prepend/append/hosts/dns overrides.
		if item.Type != "merge" && item.Type != "rules" {
			continue
		}
		content, err := a.GetVPNProfileContent(item.File)
		if err != nil {
			continue
		}

		var mc struct {
			Prepend []string          `yaml:"prepend"`
			Append  []string          `yaml:"append"`
			Hosts   map[string]string `yaml:"hosts"`
			DNS     struct {
				FakeIPFilter []string `yaml:"fake-ip-filter"`
			} `yaml:"dns"`
		}
		if yaml.Unmarshal([]byte(content), &mc) != nil {
			continue
		}

		if len(mc.Prepend) > 0 || len(mc.Append) > 0 || len(mc.Hosts) > 0 || len(mc.DNS.FakeIPFilter) > 0 {
			label := item.Name
			if label == "" {
				label = item.File
			}
			result.SourceFiles = append(result.SourceFiles, label)
			result.HasData = true
		}
		result.PrependRules = append(result.PrependRules, mc.Prepend...)
		result.PrependRules = append(result.PrependRules, mc.Append...)
		for k, v := range mc.Hosts {
			result.Hosts[k] = v
		}
		result.DNSFakeIPFilter = append(result.DNSFakeIPFilter, mc.DNS.FakeIPFilter...)
	}

	return result, nil
}
