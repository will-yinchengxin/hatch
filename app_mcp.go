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

type MCPRecommendationCandidate struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Source          string            `json:"source"`
	InstallMethod   string            `json:"installMethod"`
	Version         string            `json:"version"`
	VersionNote     string            `json:"versionNote"`
	Recommendation  string            `json:"recommendation"`
	Fit             string            `json:"fit"`
	Risk            string            `json:"risk"`
	InstallSteps    []string          `json:"installSteps"`
	Server          config.MCPServer  `json:"server"`
	RequiredEnv     []string          `json:"requiredEnv"`
	OptionalEnv     []string          `json:"optionalEnv"`
	QuickEnv        map[string]string `json:"quickEnv"`
	ClaudeSupported bool              `json:"claudeSupported"`
	CodexSupported  bool              `json:"codexSupported"`
}

type MCPRecommendationResult struct {
	Query      string                       `json:"query"`
	Summary    string                       `json:"summary"`
	SourceNote string                       `json:"sourceNote"`
	Candidates []MCPRecommendationCandidate `json:"candidates"`
}

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
		// First-time enable must not clobber a same-named entry the user wrote by
		// hand. Re-syncing an already-enabled MCP (edit/update) is allowed to overwrite.
		if enabled && !server.EnabledInClaude {
			exists, derr := claudeMCPEntryExists(server.ID)
			if derr != nil {
				return derr
			}
			if exists {
				return fmt.Errorf("Claude Code 配置中已存在名为「%s」的 MCP，已跳过以避免覆盖手工配置；如需 Hatch 接管请先在 ~/.claude.json 中删除该条目", server.ID)
			}
		}
		err = writeClaudeMCPConfig(server, enabled)
	case mcpTargetCodex:
		if enabled && !server.EnabledInCodex {
			if codexMCPEntryExists(server.ID) {
				return fmt.Errorf("Codex 配置中已存在名为「%s」的 MCP，已跳过以避免覆盖手工配置；如需 Hatch 接管请先在 ~/.codex/config.toml 中删除该条目", server.ID)
			}
		}
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
		{ID: "elastic-agent-builder", Name: "Elastic Agent Builder MCP", Description: "Elastic 9.2+ / Serverless 内置 MCP endpoint；可用时优先于自托管 MCP", SourceType: "remote", Transport: mcpTransportHTTP, URL: "https://<your-elastic-host>/mcp", BearerTokenEnvVar: "ES_API_KEY", Tags: []string{"elastic", "official", "agent-builder"}},
		{ID: "elasticsearch", Name: "Elasticsearch", Description: "Elastic 官方 Docker MCP（推荐自托管）：支持 ES 8.x/9.x，stdio 直连 Claude Code / Codex", SourceType: "remote", Transport: mcpTransportStdio, Command: "docker", Args: []string{"run", "-i", "--rm", "-e", "ES_URL", "-e", "ES_API_KEY", "-e", "ES_USERNAME", "-e", "ES_PASSWORD", "-e", "ES_SSL_SKIP_VERIFY", "docker.elastic.co/mcp/elasticsearch", "stdio"}, Env: map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": "", "ES_SSL_SKIP_VERIFY": "false"}, Tags: []string{"elastic", "official", "docker"}},
		{ID: "elasticsearch-npx", Name: "Elasticsearch npx", Description: "第三方 npm/npx MCP，适合不想安装 Docker 的本机轻量试用", SourceType: "npm", PackageName: "@tocharianou/elasticsearch-mcp", Transport: mcpTransportStdio, Command: "npx", Args: []string{"-y", "@tocharianou/elasticsearch-mcp"}, Env: map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": ""}, Tags: []string{"elastic", "npm"}},
		{ID: "prometheus", Name: "Prometheus", Description: "Prometheus 指标查询", SourceType: "git", Transport: mcpTransportStdio, Command: "prometheus-mcp-server", Args: []string{"--prometheus.url", "http://127.0.0.1:9090"}},
	}
}

func (a *App) SuggestMCPRecommendations(query string) (MCPRecommendationResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		query = "elasticsearch mcp for Claude Code and Codex"
	}
	fallback := curatedMCPRecommendations(query)

	a.mu.RLock()
	ai := a.activeAIConfig()
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		fallback.SourceNote = "离线内置推荐；配置 AI 后可基于这些候选生成排序说明。"
		return fallback, nil
	}

	system := `你是 Hatch 的 MCP 选型助手。根据用户需求和候选 MCP，给出适配 Claude Code 与 Codex 的推荐排序。
只输出 JSON 对象，不要 Markdown。字段：
summary: 80 字以内中文总结；
candidates: 数组，每项包含 id, recommendation, fit, risk。不要新增候选，不要改 id。
排序原则：官方/维护活跃/Claude Code 与 Codex stdio 兼容优先；弃用方案必须说明风险。`
	userBytes, _ := json.Marshal(map[string]any{
		"query":      query,
		"candidates": fallback.Candidates,
		"context": []string{
			"Claude Code 和 Codex 都可使用 stdio MCP；Hatch 会分别写入 ~/.claude.json 和 ~/.codex/config.toml。",
			"Elastic 官方 mcp-server-elasticsearch 0.4.0+ 使用 docker.elastic.co/mcp/elasticsearch，支持 stdio/http，旧 npm 线已弃用。",
			"用户通常需要 ES_URL，以及 ES_API_KEY 或 ES_USERNAME/ES_PASSWORD。",
		},
	})
	raw, err := a.callAIOp("SuggestMCPRecommendations", ai, system, string(userBytes))
	if err != nil {
		fallback.SourceNote = "AI 推荐失败，已使用离线内置推荐。"
		return fallback, nil
	}
	var aiOut struct {
		Summary    string `json:"summary"`
		Candidates []struct {
			ID             string `json:"id"`
			Recommendation string `json:"recommendation"`
			Fit            string `json:"fit"`
			Risk           string `json:"risk"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &aiOut); err != nil {
		fallback.SourceNote = "AI 返回无法解析，已使用离线内置推荐。"
		return fallback, nil
	}
	byID := map[string]MCPRecommendationCandidate{}
	for _, c := range fallback.Candidates {
		byID[c.ID] = c
	}
	var merged []MCPRecommendationCandidate
	for _, c := range aiOut.Candidates {
		base, ok := byID[c.ID]
		if !ok {
			continue
		}
		if strings.TrimSpace(c.Recommendation) != "" {
			base.Recommendation = c.Recommendation
		}
		if strings.TrimSpace(c.Fit) != "" {
			base.Fit = c.Fit
		}
		if strings.TrimSpace(c.Risk) != "" {
			base.Risk = c.Risk
		}
		merged = append(merged, base)
		delete(byID, c.ID)
	}
	for _, c := range fallback.Candidates {
		if _, ok := byID[c.ID]; ok {
			merged = append(merged, c)
		}
	}
	if strings.TrimSpace(aiOut.Summary) != "" {
		fallback.Summary = strings.TrimSpace(aiOut.Summary)
	}
	fallback.Candidates = merged
	fallback.SourceNote = "AI 基于 Hatch 内置候选和当前 Elastic MCP 公开说明完成排序；请在填入密钥前复核来源。"
	return fallback, nil
}

func curatedMCPRecommendations(query string) MCPRecommendationResult {
	officialStdio := config.MCPServer{
		ID:          "elasticsearch",
		Name:        "Elasticsearch",
		Description: "Elastic 官方 Docker MCP（推荐）：连接 Elasticsearch 8.x/9.x，提供索引、mapping、search、ES|QL、shard 查询",
		SourceType:  "remote",
		Transport:   mcpTransportStdio,
		Command:     "docker",
		Args:        []string{"run", "-i", "--rm", "-e", "ES_URL", "-e", "ES_API_KEY", "-e", "ES_USERNAME", "-e", "ES_PASSWORD", "-e", "ES_SSL_SKIP_VERIFY", "docker.elastic.co/mcp/elasticsearch", "stdio"},
		Env:         map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": "", "ES_SSL_SKIP_VERIFY": "false"},
		Tags:        []string{"elastic", "official", "docker", "recommended"},
	}
	officialHTTP := config.MCPServer{
		ID:          "elasticsearch-http",
		Name:        "Elasticsearch HTTP",
		Description: "Elastic 官方 Docker MCP 的 streamable-HTTP 模式，适合共享 MCP 服务或多客户端",
		SourceType:  "remote",
		Transport:   mcpTransportHTTP,
		URL:         "http://127.0.0.1:8080/mcp",
		HTTPHeaders: map[string]string{},
		Tags:        []string{"elastic", "official", "http"},
	}
	npx := config.MCPServer{
		ID:          "elasticsearch-npx",
		Name:        "Elasticsearch npx",
		Description: "第三方 npm/npx MCP，适合本机快速试用；生产建议优先官方 Docker 或 Elastic Agent Builder MCP endpoint",
		SourceType:  "npm",
		PackageName: "@tocharianou/elasticsearch-mcp",
		Transport:   mcpTransportStdio,
		Command:     "npx",
		Args:        []string{"-y", "@tocharianou/elasticsearch-mcp"},
		Env:         map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": ""},
		Tags:        []string{"elastic", "npm", "npx"},
	}
	local := config.MCPServer{
		ID:          "elasticsearch-local",
		Name:        "Elasticsearch Local Build",
		Description: "从 elastic/mcp-server-elasticsearch 源码本地构建后运行；适合需要改源码或审计镜像的场景",
		SourceType:  "git",
		SourceURL:   "https://github.com/elastic/mcp-server-elasticsearch.git",
		Transport:   mcpTransportStdio,
		Command:     "docker",
		Args:        []string{"run", "-i", "--rm", "-e", "ES_URL", "-e", "ES_API_KEY", "-e", "ES_USERNAME", "-e", "ES_PASSWORD", "-e", "ES_SSL_SKIP_VERIFY", "docker.elastic.co/mcp/elasticsearch", "stdio"},
		Env:         map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": "", "ES_SSL_SKIP_VERIFY": "false"},
		Tags:        []string{"elastic", "git", "local-build"},
	}
	return MCPRecommendationResult{
		Query:      query,
		Summary:    "如果你的 Elastic 是 9.2+ 或 Serverless，优先用 Agent Builder MCP endpoint；否则推荐 Elastic 官方 Docker stdio MCP。",
		SourceNote: "离线内置推荐，依据 Elastic 官方公开说明：0.4.0+ 使用 Docker 镜像，旧 npm 线已弃用。",
		Candidates: []MCPRecommendationCandidate{
			{
				ID:             "elastic-agent-builder",
				Name:           "Elastic Agent Builder MCP endpoint",
				Source:         "Elastic 9.2+ / Elasticsearch Serverless",
				InstallMethod:  "remote-http",
				Version:        "Elastic 9.2+ / Serverless",
				VersionNote:    "官方说明中标记为 mcp-server-elasticsearch 的继任方案；需要你的 Elastic 环境已提供 MCP endpoint。",
				Recommendation: "可用时最优",
				Fit:            "无需本机 Docker/npm 进程，Codex 可直接写入远程 MCP endpoint；Claude Code 需要按 endpoint 认证方式配置 headers 或通过 mcp-proxy 桥接。",
				Risk:           "只有 Elastic 9.2+ 或 Serverless 项目可用；Claude Code 的远程认证可能需要手工 header 配置。",
				InstallSteps: []string{
					"在 Elastic 9.2+ 或 Serverless 项目中启用/获取 Agent Builder MCP endpoint。",
					"把 endpoint URL 填入 Hatch 的 HTTP URL。",
					"把认证 token 的环境变量名填为 ES_API_KEY，或按平台要求配置 Authorization header。",
					"保存后优先同步到 Codex；Claude Code 按 endpoint 认证要求补 header，或使用 mcp-proxy。",
				},
				Server: config.MCPServer{
					ID:                "elastic-agent-builder",
					Name:              "Elastic Agent Builder MCP",
					Description:       "Elastic 9.2+ / Serverless 内置 MCP endpoint",
					SourceType:        "remote",
					Transport:         mcpTransportHTTP,
					URL:               "https://<your-elastic-host>/mcp",
					BearerTokenEnvVar: "ES_API_KEY",
					Tags:              []string{"elastic", "official", "agent-builder"},
				},
				RequiredEnv:     []string{"MCP endpoint URL", "ES_API_KEY 或平台要求的认证 header"},
				OptionalEnv:     []string{"mcp-proxy（Claude Code stdio 桥接）"},
				QuickEnv:        map[string]string{"URL": "https://<your-elastic-host>/mcp", "BearerTokenEnvVar": "ES_API_KEY"},
				ClaudeSupported: true,
				CodexSupported:  true,
			},
			{
				ID:             "elastic-official-stdio",
				Name:           "Elastic 官方 Docker stdio",
				Source:         "elastic/mcp-server-elasticsearch",
				InstallMethod:  "docker",
				Version:        "0.4.0+ Docker image",
				VersionNote:    "官方自托管 MCP server 线；0.3.1 及更早 npm 安装已弃用，0.4.0+ 改为 Docker，支持 ES 8.x/9.x。",
				Recommendation: "最优推荐",
				Fit:            "Claude Code 与 Codex 都支持 stdio；Hatch 可直接写 command/args/env 到两个目标。",
				Risk:           "依赖 Docker；ES_API_KEY 或用户名密码会以本机环境变量形式保存。",
				InstallSteps: []string{
					"安装并启动 Docker Desktop。",
					"确认 Elasticsearch 8.x/9.x 地址可从本机访问。",
					"填写 ES_URL，并选择 ES_API_KEY 或 ES_USERNAME/ES_PASSWORD。",
					"保存后打开 Claude Code 和 Codex 开关，再分别测试。",
				},
				Server:          officialStdio,
				RequiredEnv:     []string{"ES_URL", "ES_API_KEY 或 ES_USERNAME/ES_PASSWORD"},
				OptionalEnv:     []string{"ES_SSL_SKIP_VERIFY"},
				QuickEnv:        officialStdio.Env,
				ClaudeSupported: true,
				CodexSupported:  true,
			},
			{
				ID:             "elastic-official-http",
				Name:           "Elastic 官方 streamable-HTTP",
				Source:         "docker.elastic.co/mcp/elasticsearch http",
				InstallMethod:  "docker-http",
				Version:        "0.4.0+ Docker image",
				VersionNote:    "同官方 Docker MCP server，HTTP 模式支持 streamable-HTTP；SSE 已不推荐。",
				Recommendation: "共享服务备选",
				Fit:            "适合把 MCP 作为常驻 HTTP 服务给多个客户端复用；Codex 可写 HTTP MCP URL。",
				Risk:           "Claude Code 若仅支持 stdio，需要 mcp-proxy 桥接；还要自行运行容器和端口。",
				InstallSteps: []string{
					"运行 docker.elastic.co/mcp/elasticsearch http 并映射 8080。",
					"确认 http://127.0.0.1:8080/ping 返回 pong。",
					"在 Hatch 中保存 HTTP URL: http://127.0.0.1:8080/mcp。",
					"Claude Code 如需 stdio 桥接，使用 mcp-proxy。",
				},
				Server:          officialHTTP,
				RequiredEnv:     []string{"HTTP 服务启动时设置 ES_URL 与认证变量"},
				OptionalEnv:     []string{"Authorization header", "mcp-proxy"},
				QuickEnv:        map[string]string{"URL": "http://127.0.0.1:8080/mcp"},
				ClaudeSupported: true,
				CodexSupported:  true,
			},
			{
				ID:             "third-party-npx",
				Name:           "第三方 npx 包",
				Source:         "@tocharianou/elasticsearch-mcp",
				InstallMethod:  "npx",
				Version:        "0.8.0",
				VersionNote:    "本地 npx smoke test 验证到 serverInfo.version=0.8.0；第三方包宣称支持 ES 5.x-9.x。",
				Recommendation: "轻量试用",
				Fit:            "不需要 Docker，Hatch 可直接保存 npx -y 包名；适合本地临时验证。",
				Risk:           "第三方维护与安全性需自行审查；功能和官方 Elastic MCP 不完全一致。",
				InstallSteps: []string{
					"确认 Node.js 和 npx 可用。",
					"保存 npx -y @tocharianou/elasticsearch-mcp。",
					"填写 ES_URL 和认证变量。",
					"同步到 Claude Code / Codex 后测试。",
				},
				Server:          npx,
				RequiredEnv:     []string{"ES_URL", "ES_API_KEY 或 ES_USERNAME/ES_PASSWORD"},
				OptionalEnv:     []string{"NODE_TLS_REJECT_UNAUTHORIZED", "ES_CA_CERT"},
				QuickEnv:        npx.Env,
				ClaudeSupported: true,
				CodexSupported:  true,
			},
			{
				ID:             "local-build",
				Name:           "GitHub 源码本地编译",
				Source:         "https://github.com/elastic/mcp-server-elasticsearch",
				InstallMethod:  "git/local",
				Version:        "跟随源码 tag",
				VersionNote:    "建议固定到官方 release/tag；0.4.0+ 是 Docker/Rust SDK 架构线。",
				Recommendation: "审计/定制",
				Fit:            "适合需要固定源码版本、内部审计或二次开发；最终仍建议按官方 Docker/二进制入口运行。",
				Risk:           "构建步骤随上游变化，Hatch 只能登记最终 command/args，不能保证源码构建成功。",
				InstallSteps: []string{
					"克隆 https://github.com/elastic/mcp-server-elasticsearch。",
					"按上游 README 构建或使用官方 Dockerfile。",
					"把最终可执行命令填入 Hatch，或直接使用官方 Docker stdio 命令。",
					"保存并同步到 Claude Code / Codex。",
				},
				Server:          local,
				RequiredEnv:     []string{"ES_URL", "ES_API_KEY 或 ES_USERNAME/ES_PASSWORD"},
				OptionalEnv:     []string{"ES_SSL_SKIP_VERIFY"},
				QuickEnv:        local.Env,
				ClaudeSupported: true,
				CodexSupported:  true,
			},
		},
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

// claudeMCPEntryExists reports whether ~/.claude.json already defines an MCP
// server under the given id. Used to avoid clobbering hand-written entries.
func claudeMCPEntryExists(id string) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return false, nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return false, fmt.Errorf("invalid ~/.claude.json: %w", err)
	}
	servers, _ := raw["mcpServers"].(map[string]interface{})
	_, ok := servers[id]
	return ok, nil
}

// codexMCPEntryExists reports whether ~/.codex/config.toml already defines a
// [mcp_servers.<id>] section. Missing file/dir is treated as "not present".
func codexMCPEntryExists(id string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
			continue
		}
		if strings.Trim(trimmed, "[]") == "mcp_servers."+id {
			return true
		}
	}
	return false
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
