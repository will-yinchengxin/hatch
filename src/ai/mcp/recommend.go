package mcp

import (
	"encoding/json"
	"strings"

	"hatch/ai/llm"
	"hatch/core/config"
)

type RecommendationCandidate struct {
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

type RecommendationResult struct {
	Query      string                    `json:"query"`
	Summary    string                    `json:"summary"`
	SourceNote string                    `json:"sourceNote"`
	Candidates []RecommendationCandidate `json:"candidates"`
}

// CuratedRecommendations is the offline candidate list (currently focused on
// Elasticsearch, Hatch's most requested MCP), used directly or as AI input.
func CuratedRecommendations(query string) RecommendationResult {
	officialStdio := config.MCPServer{
		ID:          "elasticsearch",
		Name:        "Elasticsearch",
		Description: "Elastic 官方 Docker MCP（推荐）：连接 Elasticsearch 8.x/9.x，提供索引、mapping、search、ES|QL、shard 查询",
		SourceType:  "remote",
		Transport:   TransportStdio,
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
		Transport:   TransportHTTP,
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
		Transport:   TransportStdio,
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
		Transport:   TransportStdio,
		Command:     "docker",
		Args:        []string{"run", "-i", "--rm", "-e", "ES_URL", "-e", "ES_API_KEY", "-e", "ES_USERNAME", "-e", "ES_PASSWORD", "-e", "ES_SSL_SKIP_VERIFY", "docker.elastic.co/mcp/elasticsearch", "stdio"},
		Env:         map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": "", "ES_SSL_SKIP_VERIFY": "false"},
		Tags:        []string{"elastic", "git", "local-build"},
	}
	return RecommendationResult{
		Query:      query,
		Summary:    "如果你的 Elastic 是 9.2+ 或 Serverless，优先用 Agent Builder MCP endpoint；否则推荐 Elastic 官方 Docker stdio MCP。",
		SourceNote: "离线内置推荐，依据 Elastic 官方公开说明：0.4.0+ 使用 Docker 镜像，旧 npm 线已弃用。",
		Candidates: []RecommendationCandidate{
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
					Transport:         TransportHTTP,
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

// MergeAIRanking applies an AI re-ranking (ids + recommendation/fit/risk text)
// onto the curated candidates. Candidates the AI omitted keep their original
// order after the ranked ones; unknown ids are ignored. ok is false when the
// reply cannot be parsed.
func MergeAIRanking(fallback RecommendationResult, raw string) (RecommendationResult, bool) {
	var aiOut struct {
		Summary    string `json:"summary"`
		Candidates []struct {
			ID             string `json:"id"`
			Recommendation string `json:"recommendation"`
			Fit            string `json:"fit"`
			Risk           string `json:"risk"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(llm.ExtractJSONObject(raw)), &aiOut); err != nil {
		return fallback, false
	}
	byID := map[string]RecommendationCandidate{}
	for _, c := range fallback.Candidates {
		byID[c.ID] = c
	}
	var merged []RecommendationCandidate
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
	return fallback, true
}
