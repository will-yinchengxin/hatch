package mcp

import "hatch/core/config"

// Templates returns the built-in MCP starter gallery.
func Templates() []config.MCPServer {
	return []config.MCPServer{
		{ID: "mysql", Name: "MySQL", Description: "MySQL 查询与可选写操作", SourceType: "npm", PackageName: "@benborla29/mcp-server-mysql", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@benborla29/mcp-server-mysql"}, Env: map[string]string{"MYSQL_HOST": "127.0.0.1", "MYSQL_PORT": "3306", "MYSQL_USER": "root", "MYSQL_PASS": "", "MYSQL_DB": "", "ALLOW_INSERT_OPERATION": "false", "ALLOW_UPDATE_OPERATION": "false", "ALLOW_DELETE_OPERATION": "false"}},
		{ID: "filesystem", Name: "Filesystem", Description: "受限目录文件读写", SourceType: "npm", PackageName: "@modelcontextprotocol/server-filesystem", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"}},
		{ID: "git", Name: "Git", Description: "本地 Git 仓库读取与操作", SourceType: "npm", PackageName: "mcp-server-git", Transport: TransportStdio, Command: "uvx", Args: []string{"mcp-server-git", "--repository", "/path/to/git/repo"}},
		{ID: "github", Name: "GitHub", Description: "GitHub 仓库、Issue、PR 与 API", SourceType: "npm", PackageName: "@modelcontextprotocol/server-github", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}, Env: map[string]string{"GITHUB_PERSONAL_ACCESS_TOKEN": ""}},
		{ID: "github-remote", Name: "GitHub Remote", Description: "GitHub 官方远程 HTTP MCP", SourceType: "remote", Transport: TransportHTTP, URL: "https://api.githubcopilot.com/mcp/"},
		{ID: "postgres", Name: "Postgres", Description: "PostgreSQL 查询", SourceType: "npm", PackageName: "@modelcontextprotocol/server-postgres", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/mydb"}},
		{ID: "memory", Name: "Memory", Description: "持久化知识图谱记忆", SourceType: "npm", PackageName: "@modelcontextprotocol/server-memory", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-memory"}},
		{ID: "fetch", Name: "Fetch", Description: "网页内容抓取", SourceType: "npm", PackageName: "@modelcontextprotocol/server-fetch", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-fetch"}},
		{ID: "playwright", Name: "Playwright", Description: "浏览器自动化", SourceType: "npm", PackageName: "@playwright/mcp@latest", Transport: TransportStdio, Command: "npx", Args: []string{"@playwright/mcp@latest"}},
		{ID: "context7", Name: "Context7", Description: "实时库文档查询", SourceType: "npm", PackageName: "@upstash/context7-mcp", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@upstash/context7-mcp"}, Env: map[string]string{"CONTEXT7_API_KEY": ""}},
		{ID: "firecrawl", Name: "Firecrawl", Description: "网页搜索、抓取与结构化提取", SourceType: "npm", PackageName: "firecrawl-mcp", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "firecrawl-mcp"}, Env: map[string]string{"FIRECRAWL_API_KEY": ""}},
		{ID: "redis", Name: "Redis", Description: "Redis 数据访问", SourceType: "npm", PackageName: "@modelcontextprotocol/server-redis", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-redis"}},
		{ID: "elastic-agent-builder", Name: "Elastic Agent Builder MCP", Description: "Elastic 9.2+ / Serverless 内置 MCP endpoint；可用时优先于自托管 MCP", SourceType: "remote", Transport: TransportHTTP, URL: "https://<your-elastic-host>/mcp", BearerTokenEnvVar: "ES_API_KEY", Tags: []string{"elastic", "official", "agent-builder"}},
		{ID: "elasticsearch", Name: "Elasticsearch", Description: "Elastic 官方 Docker MCP（推荐自托管）：支持 ES 8.x/9.x，stdio 直连 Claude Code / Codex", SourceType: "remote", Transport: TransportStdio, Command: "docker", Args: []string{"run", "-i", "--rm", "-e", "ES_URL", "-e", "ES_API_KEY", "-e", "ES_USERNAME", "-e", "ES_PASSWORD", "-e", "ES_SSL_SKIP_VERIFY", "docker.elastic.co/mcp/elasticsearch", "stdio"}, Env: map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": "", "ES_SSL_SKIP_VERIFY": "false"}, Tags: []string{"elastic", "official", "docker"}},
		{ID: "elasticsearch-npx", Name: "Elasticsearch npx", Description: "第三方 npm/npx MCP，适合不想安装 Docker 的本机轻量试用", SourceType: "npm", PackageName: "@tocharianou/elasticsearch-mcp", Transport: TransportStdio, Command: "npx", Args: []string{"-y", "@tocharianou/elasticsearch-mcp"}, Env: map[string]string{"ES_URL": "http://localhost:9200", "ES_API_KEY": "", "ES_USERNAME": "", "ES_PASSWORD": ""}, Tags: []string{"elastic", "npm"}},
		{ID: "prometheus", Name: "Prometheus", Description: "Prometheus 指标查询", SourceType: "git", Transport: TransportStdio, Command: "prometheus-mcp-server", Args: []string{"--prometheus.url", "http://127.0.0.1:9090"}},
	}
}
