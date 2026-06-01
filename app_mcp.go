package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"hatch/internal/config"
)

const (
	mcpTargetClaude = "claude"
	mcpTargetCodex  = "codex"

	mcpTransportStdio = "stdio"
	mcpTransportHTTP  = "http"
)

var mcpIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var mcpExecCommand = exec.CommandContext

type MCPSettings struct {
	MCPDir        string `json:"mcpDir"`
	MCPDirIsSet   bool   `json:"mcpDirIsSet"`
	DefaultMCPDir string `json:"defaultMCPDir"`
}

type MCPConnectionResult = config.MCPTestResult

func defaultMCPDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "MyHub", "mcp")
}

func (a *App) mcpDir() (string, error) {
	a.mu.RLock()
	custom := strings.TrimSpace(a.cfg.MCPDir)
	a.mu.RUnlock()
	if custom == "" {
		custom = defaultMCPDir()
	}
	return custom, os.MkdirAll(custom, 0o755)
}

func (a *App) GetMCPSettings() MCPSettings {
	a.mu.RLock()
	custom := strings.TrimSpace(a.cfg.MCPDir)
	a.mu.RUnlock()
	def := defaultMCPDir()
	return MCPSettings{
		MCPDir:        orDefault(custom, def),
		MCPDirIsSet:   custom != "",
		DefaultMCPDir: def,
	}
}

func (a *App) SaveMCPSettings(mcpDir string) error {
	mcpDir = strings.TrimSpace(mcpDir)
	if mcpDir != "" {
		if err := os.MkdirAll(mcpDir, 0o755); err != nil {
			return err
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.MCPDir = mcpDir
	return a.cfg.Save()
}

func (a *App) ListMCPServers() ([]config.MCPServer, error) {
	a.mu.Lock()
	if !a.cfg.MCPSeeded {
		if seeded, err := a.discoverMCPServersLocked(); err == nil && len(seeded) > 0 {
			a.cfg.MCPServers = seeded
		}
		a.cfg.MCPSeeded = true
		_ = a.cfg.Save()
	}
	out := cloneMCPServers(a.cfg.MCPServers)
	a.mu.Unlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt == out[j].UpdatedAt {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out, nil
}

func (a *App) SaveMCPServer(server config.MCPServer) error {
	server, err := normalizeMCPServer(server)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, s := range a.cfg.MCPServers {
		if s.ID == server.ID {
			server.EnabledInClaude = s.EnabledInClaude
			server.EnabledInCodex = s.EnabledInCodex
			if server.LastTestResults == nil {
				server.LastTestResults = s.LastTestResults
			}
			a.cfg.MCPServers[i] = server
			return a.cfg.Save()
		}
	}
	a.cfg.MCPServers = append(a.cfg.MCPServers, server)
	return a.cfg.Save()
}

func (a *App) DeleteMCPServer(id string) error {
	if !validMCPID(id) {
		return fmt.Errorf("invalid MCP id")
	}
	a.mu.Lock()
	var removed config.MCPServer
	found := false
	for _, s := range a.cfg.MCPServers {
		if s.ID == id {
			removed = s
			found = true
			break
		}
	}
	a.mu.Unlock()
	if !found {
		return nil
	}
	if removed.EnabledInClaude {
		if err := writeClaudeMCPConfig(removed, false); err != nil {
			return err
		}
	}
	if removed.EnabledInCodex {
		if err := writeCodexMCPConfig(removed, false); err != nil {
			return err
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	keep := a.cfg.MCPServers[:0]
	for _, s := range a.cfg.MCPServers {
		if s.ID != id {
			keep = append(keep, s)
		}
	}
	a.cfg.MCPServers = keep
	return a.cfg.Save()
}

// UninstallMCPFromDisk removes the MCP from Hatch registry and synced configs,
// then deletes the local directory (only for git/local source types with a localPath).
func (a *App) UninstallMCPFromDisk(id string) error {
	if !validMCPID(id) {
		return fmt.Errorf("invalid MCP id")
	}
	a.mu.Lock()
	var removed config.MCPServer
	found := false
	for _, s := range a.cfg.MCPServers {
		if s.ID == id {
			removed = s
			found = true
			break
		}
	}
	a.mu.Unlock()
	if !found {
		return nil
	}
	if removed.EnabledInClaude {
		if err := writeClaudeMCPConfig(removed, false); err != nil {
			return err
		}
	}
	if removed.EnabledInCodex {
		if err := writeCodexMCPConfig(removed, false); err != nil {
			return err
		}
	}
	a.mu.Lock()
	keep := a.cfg.MCPServers[:0]
	for _, s := range a.cfg.MCPServers {
		if s.ID != id {
			keep = append(keep, s)
		}
	}
	a.cfg.MCPServers = keep
	if err := a.cfg.Save(); err != nil {
		a.mu.Unlock()
		return err
	}
	a.mu.Unlock()
	if removed.LocalPath != "" && (removed.SourceType == "git" || removed.SourceType == "local") {
		if err := os.RemoveAll(removed.LocalPath); err != nil {
			return fmt.Errorf("从 Hatch 删除成功，但清理本地目录失败: %w", err)
		}
	}
	return nil
}

// UpdateMCPFromGit runs git pull in the MCP's local directory (git source type only),
// then re-infers the command configuration from the updated directory.
func (a *App) UpdateMCPFromGit(id string) (config.MCPServer, error) {
	if !validMCPID(id) {
		return config.MCPServer{}, fmt.Errorf("invalid MCP id")
	}
	a.mu.RLock()
	var server config.MCPServer
	found := false
	for _, s := range a.cfg.MCPServers {
		if s.ID == id {
			server = s
			found = true
			break
		}
	}
	a.mu.RUnlock()
	if !found {
		return config.MCPServer{}, fmt.Errorf("MCP not found: %s", id)
	}
	if server.SourceType != "git" {
		return config.MCPServer{}, fmt.Errorf("只有 git 来源的 MCP 支持更新")
	}
	if server.LocalPath == "" {
		return config.MCPServer{}, fmt.Errorf("本地目录路径未设置")
	}
	if _, err := os.Stat(server.LocalPath); err != nil {
		return config.MCPServer{}, fmt.Errorf("本地目录不存在: %s", server.LocalPath)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := mcpExecCommand(ctx, "git", "-C", server.LocalPath, "pull")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return config.MCPServer{}, fmt.Errorf("git pull 失败: %s", strings.TrimSpace(buf.String()))
	}
	updated := inferMCPServerFromDir(server.LocalPath)
	updated.ID = server.ID
	updated.Name = server.Name
	updated.SourceType = server.SourceType
	updated.SourceURL = server.SourceURL
	updated.EnabledInClaude = server.EnabledInClaude
	updated.EnabledInCodex = server.EnabledInCodex
	updated.Tags = server.Tags
	updated.LastTestResults = server.LastTestResults
	if err := a.SaveMCPServer(updated); err != nil {
		return config.MCPServer{}, err
	}
	if updated.EnabledInClaude {
		_ = writeClaudeMCPConfig(updated, true)
	}
	if updated.EnabledInCodex {
		_ = writeCodexMCPConfig(updated, true)
	}
	return updated, nil
}

func (a *App) ImportMCPFromLocalDir(path string) (config.MCPServer, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return config.MCPServer{}, fmt.Errorf("local path is required")
	}
	fi, err := os.Stat(path)
	if err != nil {
		return config.MCPServer{}, err
	}
	if !fi.IsDir() {
		return config.MCPServer{}, fmt.Errorf("local path is not a directory")
	}
	server := inferMCPServerFromDir(path)
	if err := a.SaveMCPServer(server); err != nil {
		return config.MCPServer{}, err
	}
	return server, nil
}

func (a *App) InstallMCPFromGit(url, name string) (config.MCPServer, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return config.MCPServer{}, fmt.Errorf("git URL is required")
	}
	id := sanitizeMCPID(name)
	if id == "" {
		id = sanitizeMCPID(strings.TrimSuffix(filepath.Base(url), ".git"))
	}
	if !validMCPID(id) {
		return config.MCPServer{}, fmt.Errorf("invalid MCP id")
	}
	base, err := a.mcpDir()
	if err != nil {
		return config.MCPServer{}, err
	}
	dst := filepath.Join(base, id)
	if _, err := os.Stat(dst); err == nil {
		return config.MCPServer{}, fmt.Errorf("target directory already exists: %s", dst)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := mcpExecCommand(ctx, "git", "clone", url, dst)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return config.MCPServer{}, fmt.Errorf("git clone failed: %s", strings.TrimSpace(buf.String()))
	}
	server := inferMCPServerFromDir(dst)
	server.ID = id
	server.Name = orDefault(strings.TrimSpace(name), server.Name)
	server.SourceType = "git"
	server.SourceURL = url
	if err := a.SaveMCPServer(server); err != nil {
		return config.MCPServer{}, err
	}
	return server, nil
}

func (a *App) SaveNpmMCPServer(packageName, name string) (config.MCPServer, error) {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return config.MCPServer{}, fmt.Errorf("npm package is required")
	}
	id := sanitizeMCPID(name)
	if id == "" {
		id = sanitizeMCPID(packageName)
	}
	server := config.MCPServer{
		ID:          id,
		Name:        orDefault(strings.TrimSpace(name), packageName),
		SourceType:  "npm",
		PackageName: packageName,
		Transport:   mcpTransportStdio,
		Command:     "npx",
		Args:        []string{"-y", packageName},
		UpdatedAt:   time.Now().Unix(),
	}
	if err := a.SaveMCPServer(server); err != nil {
		return config.MCPServer{}, err
	}
	return server, nil
}

func (a *App) SetMCPEnabled(id, target string, enabled bool) error {
	if !validMCPID(id) {
		return fmt.Errorf("invalid MCP id")
	}
	a.mu.Lock()
	idx := -1
	var server config.MCPServer
	for i, s := range a.cfg.MCPServers {
		if s.ID == id {
			idx = i
			server = s
			break
		}
	}
	a.mu.Unlock()
	if idx < 0 {
		return fmt.Errorf("MCP not found: %s", id)
	}

	var err error
	switch target {
	case mcpTargetClaude:
		err = writeClaudeMCPConfig(server, enabled)
	case mcpTargetCodex:
		err = writeCodexMCPConfig(server, enabled)
	default:
		return fmt.Errorf("unknown MCP target: %s", target)
	}
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for i, s := range a.cfg.MCPServers {
		if s.ID != id {
			continue
		}
		if target == mcpTargetClaude {
			a.cfg.MCPServers[i].EnabledInClaude = enabled
		} else {
			a.cfg.MCPServers[i].EnabledInCodex = enabled
		}
		a.cfg.MCPServers[i].UpdatedAt = time.Now().Unix()
		break
	}
	return a.cfg.Save()
}

func (a *App) TestMCPConnection(id, target string) (MCPConnectionResult, error) {
	if !validMCPID(id) {
		return MCPConnectionResult{}, fmt.Errorf("invalid MCP id")
	}
	target = strings.TrimSpace(target)
	if target != mcpTargetClaude && target != mcpTargetCodex {
		return MCPConnectionResult{}, fmt.Errorf("unknown MCP target: %s", target)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	var cmd *exec.Cmd
	if target == mcpTargetClaude {
		if findCLI("claude") == "" {
			return a.storeMCPTestResult(id, target, mcpTestFailure(target, "claude CLI 未安装或不在 PATH", "", -1, "connection"))
		}
		cmd = mcpExecCommand(ctx, "claude", "mcp", "get", id)
	} else {
		if findCodexCLI() == "" {
			return a.storeMCPTestResult(id, target, mcpTestFailure(target, "codex CLI 未安装或不在 PATH", "", -1, "config"))
		}
		cmd = mcpExecCommand(ctx, "codex", "mcp", "get", id)
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	detail := strings.TrimSpace(buf.String())
	exitCode := 0
	if err != nil {
		exitCode = commandExitCode(err)
		if ctx.Err() == context.DeadlineExceeded {
			exitCode = -1
			detail = strings.TrimSpace(detail + "\nconnection test timed out")
		}
	}

	result := parseMCPTestOutput(target, detail, exitCode, err)
	return a.storeMCPTestResult(id, target, result)
}

func (a *App) GetMCPTemplates() []config.MCPServer {
	return []config.MCPServer{
		{ID: "mysql", Name: "MySQL", Description: "MySQL 查询与可选写操作", SourceType: "npm", PackageName: "@benborla29/mcp-server-mysql", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@benborla29/mcp-server-mysql"}, Env: map[string]string{"MYSQL_HOST": "127.0.0.1", "MYSQL_PORT": "3306", "MYSQL_USER": "root", "MYSQL_PASS": "", "MYSQL_DB": "", "ALLOW_INSERT_OPERATION": "false", "ALLOW_UPDATE_OPERATION": "false", "ALLOW_DELETE_OPERATION": "false"}},
		{ID: "filesystem", Name: "Filesystem", Description: "受限目录文件读写", SourceType: "npm", PackageName: "@modelcontextprotocol/server-filesystem", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"}},
		{ID: "git", Name: "Git", Description: "本地 Git 仓库读取与操作", SourceType: "npm", PackageName: "mcp-server-git", Transport: mcpTransportStdio, Command: "uvx", Args: []string{"mcp-server-git", "--repository", "/path/to/git/repo"}},
		{ID: "github", Name: "GitHub", Description: "GitHub 仓库、Issue、PR 与 API", SourceType: "npm", PackageName: "@modelcontextprotocol/server-github", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}, Env: map[string]string{"GITHUB_PERSONAL_ACCESS_TOKEN": ""}},
		{ID: "github-remote", Name: "GitHub Remote", Description: "GitHub 官方远程 HTTP MCP", SourceType: "remote", Transport: mcpTransportHTTP, URL: "https://api.githubcopilot.com/mcp/"},
		{ID: "postgres", Name: "Postgres", Description: "PostgreSQL 查询", SourceType: "npm", PackageName: "@modelcontextprotocol/server-postgres", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/mydb"}},
		{ID: "memory", Name: "Memory", Description: "持久化知识图谱记忆", SourceType: "npm", PackageName: "@modelcontextprotocol/server-memory", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-memory"}},
		{ID: "fetch", Name: "Fetch", Description: "网页内容抓取", SourceType: "npm", PackageName: "@modelcontextprotocol/server-fetch", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-fetch"}},
		{ID: "playwright", Name: "Playwright", Description: "浏览器自动化", SourceType: "npm", PackageName: "@playwright/mcp@latest", Transport: mcpTransportStdio, Command: "npx", Args: []string{"@playwright/mcp@latest"}},
		{ID: "context7", Name: "Context7", Description: "实时库文档查询", SourceType: "npm", PackageName: "@upstash/context7-mcp", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@upstash/context7-mcp"}, Env: map[string]string{"CONTEXT7_API_KEY": ""}},
		{ID: "firecrawl", Name: "Firecrawl", Description: "网页搜索、抓取与结构化提取", SourceType: "npm", PackageName: "firecrawl-mcp", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "firecrawl-mcp"}, Env: map[string]string{"FIRECRAWL_API_KEY": ""}},
		{ID: "redis", Name: "Redis", Description: "Redis 数据访问", SourceType: "npm", PackageName: "@modelcontextprotocol/server-redis", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-redis"}},
		{ID: "elasticsearch", Name: "Elasticsearch", Description: "Elasticsearch 查询", SourceType: "npm", PackageName: "@tocharianou/elasticsearch-mcp", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@tocharianou/elasticsearch-mcp"}, Env: map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": ""}},
		{ID: "prometheus", Name: "Prometheus", Description: "Prometheus 指标查询", SourceType: "git", Transport: mcpTransportStdio, Command: "prometheus-mcp-server", Args: []string{"--prometheus.url", "http://127.0.0.1:9090"}},
	}
}

func (a *App) storeMCPTestResult(id, target string, result MCPConnectionResult) (MCPConnectionResult, error) {
	result.Target = target
	result.CheckedAt = time.Now().Unix()
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, s := range a.cfg.MCPServers {
		if s.ID != id {
			continue
		}
		if a.cfg.MCPServers[i].LastTestResults == nil {
			a.cfg.MCPServers[i].LastTestResults = map[string]config.MCPTestResult{}
		}
		a.cfg.MCPServers[i].LastTestResults[target] = result
		a.cfg.MCPServers[i].UpdatedAt = time.Now().Unix()
		return result, a.cfg.Save()
	}
	return result, fmt.Errorf("MCP not found: %s", id)
}

func (a *App) discoverMCPServersLocked() ([]config.MCPServer, error) {
	base := strings.TrimSpace(a.cfg.MCPDir)
	if base == "" {
		base = defaultMCPDir()
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var servers []config.MCPServer
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		server := inferMCPServerFromDir(filepath.Join(base, e.Name()))
		server.UpdatedAt = now
		servers = append(servers, server)
	}
	return servers, nil
}

func inferMCPServerFromDir(dir string) config.MCPServer {
	name := filepath.Base(dir)
	id := sanitizeMCPID(strings.TrimSuffix(strings.TrimSuffix(name, "-mcp-server"), "_mcp_server"))
	server := config.MCPServer{
		ID:         id,
		Name:       humanizeMCPName(name),
		SourceType: "local",
		LocalPath:  dir,
		Transport:  mcpTransportStdio,
		Cwd:        dir,
		UpdatedAt:  time.Now().Unix(),
	}
	if desc := readMCPReadmeSummary(dir); desc != "" {
		server.Description = desc
	}
	if pkg, ok := readPackageJSON(dir); ok {
		if pkg.Name != "" {
			server.PackageName = pkg.Name
		}
		if server.ID == "" {
			server.ID = sanitizeMCPID(pkg.Name)
		}
		if server.Name == "" {
			server.Name = pkg.Name
		}
		if binPath := firstPackageBin(pkg); binPath != "" {
			server.Command = "node"
			server.Args = []string{filepath.Join(dir, binPath)}
			return server
		}
		if start := strings.TrimSpace(pkg.Scripts["start"]); strings.HasPrefix(start, "node ") {
			server.Command = "node"
			server.Args = strings.Fields(strings.TrimPrefix(start, "node "))
			for i, arg := range server.Args {
				if !filepath.IsAbs(arg) && strings.HasSuffix(arg, ".js") {
					server.Args[i] = filepath.Join(dir, arg)
				}
			}
			return server
		}
		if pkg.Main != "" {
			server.Command = "node"
			server.Args = []string{filepath.Join(dir, pkg.Main)}
			return server
		}
		server.Command = "npm"
		server.Args = []string{"run", "start"}
		return server
	}
	if exe := firstExecutableFile(dir); exe != "" {
		server.Command = exe
		if strings.Contains(filepath.Base(exe), "prometheus") {
			server.Args = []string{"--prometheus.url", "http://127.0.0.1:9090"}
		}
		return server
	}
	return server
}

type packageJSON struct {
	Name    string            `json:"name"`
	Main    string            `json:"main"`
	Bin     json.RawMessage   `json:"bin"`
	Scripts map[string]string `json:"scripts"`
}

func readPackageJSON(dir string) (packageJSON, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return packageJSON{}, false
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return packageJSON{}, false
	}
	if pkg.Scripts == nil {
		pkg.Scripts = map[string]string{}
	}
	return pkg, true
}

func firstPackageBin(pkg packageJSON) string {
	if len(pkg.Bin) == 0 || string(pkg.Bin) == "null" {
		return ""
	}
	var one string
	if json.Unmarshal(pkg.Bin, &one) == nil {
		return one
	}
	var many map[string]string
	if json.Unmarshal(pkg.Bin, &many) != nil {
		return ""
	}
	keys := make([]string, 0, len(many))
	for k := range many {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return many[keys[0]]
}

func readMCPReadmeSummary(dir string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, "README*"))
	sort.Strings(matches)
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(strings.Trim(line, "#> "))
			if line == "" || strings.HasPrefix(line, "[!") || strings.HasPrefix(line, "<") {
				continue
			}
			if len(line) > 180 {
				line = line[:180] + "..."
			}
			return line
		}
	}
	return ""
}

func firstExecutableFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0o111 != 0 {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

func normalizeMCPServer(server config.MCPServer) (config.MCPServer, error) {
	server.ID = sanitizeMCPID(server.ID)
	if !validMCPID(server.ID) {
		return config.MCPServer{}, fmt.Errorf("invalid MCP id")
	}
	server.Name = orDefault(strings.TrimSpace(server.Name), server.ID)
	server.SourceType = orDefault(strings.TrimSpace(server.SourceType), "local")
	server.Transport = strings.TrimSpace(server.Transport)
	if server.Transport == "" {
		server.Transport = mcpTransportStdio
	}
	if server.Transport != mcpTransportStdio && server.Transport != mcpTransportHTTP {
		return config.MCPServer{}, fmt.Errorf("unsupported MCP transport: %s", server.Transport)
	}
	server.Command = strings.TrimSpace(server.Command)
	server.Cwd = strings.TrimSpace(server.Cwd)
	server.URL = strings.TrimSpace(server.URL)
	server.SourceURL = strings.TrimSpace(server.SourceURL)
	server.LocalPath = strings.TrimSpace(server.LocalPath)
	server.PackageName = strings.TrimSpace(server.PackageName)
	server.BearerTokenEnvVar = strings.TrimSpace(server.BearerTokenEnvVar)
	if server.Transport == mcpTransportStdio && server.Command == "" {
		return config.MCPServer{}, fmt.Errorf("stdio MCP command is required")
	}
	if server.Transport == mcpTransportHTTP && server.URL == "" {
		return config.MCPServer{}, fmt.Errorf("HTTP MCP URL is required")
	}
	server.Args = cleanStringSlice(server.Args)
	server.Tags = cleanStringSlice(server.Tags)
	if server.Env == nil {
		server.Env = map[string]string{}
	}
	if server.HTTPHeaders == nil {
		server.HTTPHeaders = map[string]string{}
	}
	if server.LastTestResults == nil {
		server.LastTestResults = map[string]config.MCPTestResult{}
	}
	server.UpdatedAt = time.Now().Unix()
	return server, nil
}

func cleanStringSlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

func validMCPID(id string) bool {
	return mcpIDPattern.MatchString(strings.TrimSpace(id))
}

func sanitizeMCPID(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "@")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.TrimSuffix(s, "-mcp")
	s = strings.TrimSuffix(s, "-server")
	s = strings.Trim(s, "-_ ")
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-'
		if !ok {
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
			continue
		}
		b.WriteRune(r)
		lastDash = r == '-'
	}
	return strings.Trim(b.String(), "-_")
}

func humanizeMCPName(name string) string {
	name = strings.TrimSuffix(strings.TrimSuffix(name, "-mcp-server"), "_mcp_server")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	parts := strings.Fields(name)
	for i, p := range parts {
		if p == "mcp" {
			parts[i] = "MCP"
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func cloneMCPServers(in []config.MCPServer) []config.MCPServer {
	out := make([]config.MCPServer, len(in))
	for i, s := range in {
		out[i] = s
		out[i].Args = append([]string(nil), s.Args...)
		out[i].Tags = append([]string(nil), s.Tags...)
		out[i].Env = cloneStringMap(s.Env)
		out[i].HTTPHeaders = cloneStringMap(s.HTTPHeaders)
		if s.LastTestResults != nil {
			out[i].LastTestResults = make(map[string]config.MCPTestResult, len(s.LastTestResults))
			for k, v := range s.LastTestResults {
				out[i].LastTestResults[k] = v
			}
		}
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func writeClaudeMCPConfig(server config.MCPServer, enabled bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".claude.json")
	var raw map[string]interface{}
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("invalid ~/.claude.json: %w", err)
		}
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}
	servers, _ := raw["mcpServers"].(map[string]interface{})
	if servers == nil {
		servers = map[string]interface{}{}
	}
	if enabled {
		servers[server.ID] = renderClaudeMCPServer(server)
	} else {
		delete(servers, server.ID)
	}
	raw["mcpServers"] = servers
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := backupIfExists(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o600)
}

func renderClaudeMCPServer(server config.MCPServer) map[string]interface{} {
	out := map[string]interface{}{}
	if server.Transport == mcpTransportHTTP {
		out["type"] = "http"
		out["url"] = server.URL
		if len(server.HTTPHeaders) > 0 {
			out["headers"] = server.HTTPHeaders
		}
		return out
	}
	out["type"] = "stdio"
	out["command"] = server.Command
	out["args"] = server.Args
	if len(server.Env) > 0 {
		out["env"] = server.Env
	} else {
		out["env"] = map[string]string{}
	}
	if server.Cwd != "" {
		out["cwd"] = server.Cwd
	}
	return out
}

func writeCodexMCPConfig(server config.MCPServer, enabled bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".codex", "config.toml")
	content := ""
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}
	content = removeCodexMCPServer(content, server.ID)
	if enabled {
		content = ensureTrailingNewline(content)
		if strings.TrimSpace(content) != "" {
			content += "\n"
		}
		content += renderCodexMCPServer(server)
	}
	if err := backupIfExists(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(ensureTrailingNewline(content)), 0o644)
}

func renderCodexMCPServer(server config.MCPServer) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Managed by Hatch MCP\n[mcp_servers.%s]\n", server.ID)
	if server.Transport == mcpTransportHTTP {
		fmt.Fprintf(&b, "url = %s\n", strconv.Quote(server.URL))
		if server.BearerTokenEnvVar != "" {
			fmt.Fprintf(&b, "bearer_token_env_var = %s\n", strconv.Quote(server.BearerTokenEnvVar))
		}
		if len(server.HTTPHeaders) > 0 {
			fmt.Fprintf(&b, "http_headers = %s\n", tomlInlineMap(server.HTTPHeaders))
		}
	} else {
		fmt.Fprintf(&b, "command = %s\n", strconv.Quote(server.Command))
		if len(server.Args) > 0 {
			fmt.Fprintf(&b, "args = %s\n", tomlStringArray(server.Args))
		}
		if server.Cwd != "" {
			fmt.Fprintf(&b, "cwd = %s\n", strconv.Quote(server.Cwd))
		}
	}
	if server.StartupTimeoutSec > 0 {
		fmt.Fprintf(&b, "startup_timeout_sec = %d\n", server.StartupTimeoutSec)
	}
	if server.ToolTimeoutSec > 0 {
		fmt.Fprintf(&b, "tool_timeout_sec = %d\n", server.ToolTimeoutSec)
	}
	b.WriteString("enabled = true\n")
	if len(server.Env) > 0 && server.Transport == mcpTransportStdio {
		fmt.Fprintf(&b, "\n[mcp_servers.%s.env]\n", server.ID)
		keys := sortedMapKeys(server.Env)
		for _, k := range keys {
			fmt.Fprintf(&b, "%s = %s\n", k, strconv.Quote(server.Env[k]))
		}
	}
	return b.String()
}

func removeCodexMCPServer(content, id string) string {
	var kept []string
	inTarget := false
	for _, line := range strings.SplitAfter(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			name := strings.Trim(trimmed, "[]")
			inTarget = name == "mcp_servers."+id || strings.HasPrefix(name, "mcp_servers."+id+".")
		}
		if inTarget {
			continue
		}
		if strings.TrimSpace(line) == "# Managed by Hatch MCP" {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimRight(strings.Join(kept, ""), "\n") + "\n"
}

func tomlStringArray(values []string) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, strconv.Quote(v))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func tomlInlineMap(values map[string]string) string {
	keys := sortedMapKeys(values)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, strconv.Quote(k)+" = "+strconv.Quote(values[k]))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for k := range values {
		if strings.TrimSpace(k) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func parseMCPTestOutput(target, detail string, exitCode int, err error) MCPConnectionResult {
	mode := "connection"
	if target == mcpTargetCodex {
		mode = "config"
	}
	if err != nil {
		return mcpTestFailure(target, summarizeMCPFailure(detail), detail, exitCode, mode)
	}
	if target == mcpTargetClaude {
		if strings.Contains(detail, "Status: ✓") || strings.Contains(detail, "Connected") {
			return MCPConnectionResult{Target: target, Status: "success", Summary: "Claude Code MCP 连接成功", Detail: detail, Mode: mode}
		}
		if line := firstLineContaining(detail, "Status:"); line != "" {
			return mcpTestFailure(target, strings.TrimSpace(strings.TrimPrefix(line, "Status:")), detail, exitCode, mode)
		}
		return MCPConnectionResult{Target: target, Status: "unknown", Summary: "Claude Code 返回了配置，但未包含连接状态", Detail: detail, Mode: mode}
	}
	if strings.Contains(detail, "enabled: true") || strings.Contains(detail, "Status") || strings.Contains(detail, "transport:") {
		return MCPConnectionResult{Target: target, Status: "unknown", Summary: "Codex 配置可读取；当前 CLI 未提供实时连接状态", Detail: detail, Mode: mode}
	}
	return MCPConnectionResult{Target: target, Status: "unknown", Summary: "Codex MCP 配置检查完成", Detail: detail, Mode: mode}
}

func mcpTestFailure(target, summary, detail string, exitCode int, mode string) MCPConnectionResult {
	return MCPConnectionResult{Target: target, Status: "failed", Summary: summary, Detail: detail, ExitCode: exitCode, Mode: mode}
}

func summarizeMCPFailure(detail string) string {
	detail = strings.TrimSpace(detail)
	if detail == "" {
		return "连接测试失败，但 CLI 未返回错误详情"
	}
	for _, line := range strings.Split(detail, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return "连接测试失败"
}

func firstLineContaining(text, needle string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, needle) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	return -1
}
