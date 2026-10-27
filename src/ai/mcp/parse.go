package mcp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml"
)

// Draft is a server definition waiting for the user to fill inputs and pick
// targets. It is never persisted on its own.
type Draft struct {
	Name         string     `json:"name"`
	Title        string     `json:"title,omitempty"`
	Description  string     `json:"description,omitempty"`
	Variant      string     `json:"variant,omitempty"` // e.g. "npm · stdio", "远程 · HTTP"
	Spec         Spec       `json:"spec"`
	Inputs       []InputDef `json:"inputs,omitempty"`
	Origin       string     `json:"origin"`
	SourceClient string     `json:"sourceClient,omitempty"` // dialect of Spec.Extra
	Scope        string     `json:"scope,omitempty"`        // scope hinted by a pasted command (-s user)
	Notes        []string   `json:"notes,omitempty"`
	Homepage     string     `json:"homepage,omitempty"`
}

// InputDef is a value the user should fill before applying.
type InputDef struct {
	Kind        string `json:"kind"` // env | header
	Key         string `json:"key"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
	Default     string `json:"default,omitempty"`
}

// ParseResult is what a paste turned into.
type ParseResult struct {
	Format string   `json:"format"`
	Drafts []Draft  `json:"drafts"`
	Notes  []string `json:"notes,omitempty"`
	// RegistryName is set when the input looked like an MCP Registry name;
	// the caller resolves it over the network.
	RegistryName string `json:"registryName,omitempty"`
}

var (
	secretKeyRe   = regexp.MustCompile(`(?i)(key|token|secret|passw|pass$|_pass|auth|credential|cookie)`)
	placeholderRe = regexp.MustCompile(`(?i)^(<.*>|your[-_ ].*|xxx+|\.\.\.|changeme|todo|replace[-_ ]?me)$`)
	registryRe    = regexp.MustCompile(`^[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+/[a-zA-Z0-9._-]+$`)
	npmRe         = regexp.MustCompile(`^(@[a-z0-9][a-z0-9._-]*/)?[a-z0-9][a-z0-9._-]*(@[a-zA-Z0-9._-]+)?$`)
	lineComment   = regexp.MustCompile(`(?m)^\s*//.*$`)
	trailingComma = regexp.MustCompile(`,(\s*[}\]])`)
)

// Parse turns pasted text into drafts.
func Parse(text string) (ParseResult, error) {
	text = strings.TrimSpace(strings.TrimPrefix(text, "\uFEFF"))
	text = stripFence(text)
	if text == "" {
		return ParseResult{}, fmt.Errorf("请粘贴 MCP 配置、安装命令、URL 或包名")
	}
	first := strings.Fields(text)[0]
	switch {
	case strings.HasPrefix(text, "{") || strings.HasPrefix(text, `"`):
		return parseJSON(text)
	case strings.Contains(text, "[mcp_servers"):
		return parseTOML(text)
	case strings.HasPrefix(text, "claude mcp add") || strings.HasPrefix(text, "codex mcp add"):
		return parseCLI(text)
	case !strings.ContainsAny(text, " \n") && (strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://")):
		d := draftFromURL(text)
		return ParseResult{Format: "url", Drafts: []Draft{d}}, nil
	case !strings.ContainsAny(text, " \n") && registryRe.MatchString(text) && !strings.HasPrefix(text, "@"):
		return ParseResult{Format: "registry", RegistryName: text}, nil
	case !strings.ContainsAny(text, " \n") && npmRe.MatchString(text):
		spec := Spec{Transport: TransportStdio, Command: "npx", Args: []string{"-y", text}}
		d := Draft{Name: SanitizeName(text), Spec: spec, Origin: "paste", Variant: "npm · stdio"}
		d.Notes = append(d.Notes, "按 npm 包处理（npx -y）；Python 包请改为 uvx 命令")
		return ParseResult{Format: "npm", Drafts: []Draft{withInputs(d)}}, nil
	case isLauncher(first):
		args := shellSplit(text)
		spec := Spec{Transport: TransportStdio, Command: args[0], Args: args[1:]}
		return ParseResult{Format: "command", Drafts: []Draft{withInputs(Draft{Name: nameFromArgs(args), Spec: spec, Origin: "paste"})}}, nil
	}
	return ParseResult{}, fmt.Errorf("无法识别：支持 JSON 片段、claude/codex mcp add 命令、TOML、URL、npm 包名、Registry 名称或 npx/uvx/docker 命令")
}

func stripFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return s
	}
	lines = lines[1:]
	if n := len(lines); n > 0 && strings.HasPrefix(strings.TrimSpace(lines[n-1]), "```") {
		lines = lines[:n-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func isLauncher(cmd string) bool {
	switch cmd {
	case "npx", "uvx", "uv", "docker", "node", "python", "python3", "bunx", "pnpm", "deno", "go", "java":
		return true
	}
	return strings.HasPrefix(cmd, "/") || strings.HasPrefix(cmd, "./") || strings.HasPrefix(cmd, "~/")
}

// ── JSON ────────────────────────────────────────────────────────────────────

func parseJSON(text string) (ParseResult, error) {
	clean := trailingComma.ReplaceAllString(lineComment.ReplaceAllString(text, ""), "$1")
	var root map[string]any
	if err := json.Unmarshal([]byte(clean), &root); err != nil {
		// README fragments often omit the outer braces: "name": { ... }
		if err2 := json.Unmarshal([]byte("{"+clean+"}"), &root); err2 != nil {
			return ParseResult{}, fmt.Errorf("JSON 解析失败: %w", err)
		}
	}
	servers := findServersMap(root)
	res := ParseResult{Format: "json"}
	if servers == nil {
		if looksLikeSpec(root) {
			servers = map[string]any{"": root}
		} else {
			return res, fmt.Errorf("没有找到 mcpServers / servers 或带 command / url 的条目")
		}
	}
	for _, name := range sortedKeys(servers) {
		body, ok := servers[name].(map[string]any)
		if !ok || !looksLikeSpec(body) {
			continue
		}
		if u, ok := body["serverUrl"].(string); ok && body["url"] == nil {
			body["url"] = u
			delete(body, "serverUrl")
		}
		spec := SpecFromClaude(body)
		if name == "" {
			name = nameFromSpec(spec)
		}
		res.Drafts = append(res.Drafts, withInputs(Draft{Name: SanitizeKeep(name), Spec: spec, Origin: "paste", SourceClient: ClientClaude}))
	}
	if len(res.Drafts) == 0 {
		return res, fmt.Errorf("没有找到可用的 MCP 条目")
	}
	if len(res.Drafts) > 1 {
		res.Notes = append(res.Notes, fmt.Sprintf("识别到 %d 个服务，可逐个切换后应用", len(res.Drafts)))
	}
	return res, nil
}

// findServersMap accepts Claude/Cursor ({"mcpServers":…}), VS Code
// ({"servers":…} or {"mcp":{"servers":…}}) and bare {"name": spec} maps.
func findServersMap(root map[string]any) map[string]any {
	for _, k := range []string{"mcpServers", "servers", "mcp_servers"} {
		if m, ok := root[k].(map[string]any); ok {
			return m
		}
	}
	if mcp, ok := root["mcp"].(map[string]any); ok {
		if m, ok := mcp["servers"].(map[string]any); ok {
			return m
		}
	}
	if looksLikeSpec(root) {
		return nil
	}
	for _, v := range root {
		if m, ok := v.(map[string]any); ok && looksLikeSpec(m) {
			return root
		}
	}
	return nil
}

func looksLikeSpec(m map[string]any) bool {
	_, hasCmd := m["command"].(string)
	_, hasURL := m["url"].(string)
	_, hasServerURL := m["serverUrl"].(string)
	return hasCmd || hasURL || hasServerURL
}

// ── TOML ────────────────────────────────────────────────────────────────────

func parseTOML(text string) (ParseResult, error) {
	tree, err := toml.Load(text)
	if err != nil {
		return ParseResult{}, fmt.Errorf("TOML 解析失败: %w", err)
	}
	res := ParseResult{Format: "toml"}
	servers, ok := tree.Get("mcp_servers").(*toml.Tree)
	if !ok {
		return res, fmt.Errorf("没有找到 [mcp_servers.*]")
	}
	for _, name := range servers.Keys() {
		t, ok := servers.Get(name).(*toml.Tree)
		if !ok {
			continue
		}
		spec, _ := SpecFromCodex(t.ToMap())
		res.Drafts = append(res.Drafts, withInputs(Draft{Name: name, Spec: spec, Origin: "paste", SourceClient: ClientCodex}))
	}
	if len(res.Drafts) == 0 {
		return res, fmt.Errorf("没有找到可用的 MCP 条目")
	}
	return res, nil
}

// ── CLI commands ────────────────────────────────────────────────────────────

func parseCLI(text string) (ParseResult, error) {
	text = strings.ReplaceAll(text, "\\\n", " ")
	args := shellSplit(text)
	if len(args) < 4 {
		return ParseResult{}, fmt.Errorf("命令不完整")
	}
	client := args[0]
	if args[2] == "add-json" {
		if len(args) < 5 {
			return ParseResult{}, fmt.Errorf("add-json 需要名称和 JSON")
		}
		res, err := parseJSON(args[len(args)-1])
		if err != nil {
			return res, err
		}
		name := ""
		for i := 3; i < len(args)-1; i++ {
			if args[i] == "-s" || args[i] == "--scope" {
				i++
				continue
			}
			if !strings.HasPrefix(args[i], "-") {
				name = args[i]
				break
			}
		}
		if name != "" && len(res.Drafts) == 1 {
			res.Drafts[0].Name = name
		}
		res.Format = "claude-cli"
		return res, nil
	}
	rest := args[3:]
	spec := Spec{Env: map[string]string{}, Headers: map[string]string{}}
	d := Draft{Origin: "paste"}
	var positional []string
	for i := 0; i < len(rest); i++ {
		a := rest[i]
		next := func() string {
			if i+1 < len(rest) {
				i++
				return rest[i]
			}
			return ""
		}
		if a == "--" {
			positional = append(positional, rest[i+1:]...)
			break
		}
		key, inline, hasInline := strings.Cut(a, "=")
		val := func() string {
			if hasInline && strings.HasPrefix(key, "--") {
				return inline
			}
			return next()
		}
		switch key {
		case "-t", "--transport":
			spec.Transport = normalizeTransport(val())
		case "-s", "--scope":
			d.Scope = val()
		case "-e", "--env":
			// claude -e accepts several KEY=VALUE values in a row.
			for {
				kv := val()
				if k, v, ok := strings.Cut(kv, "="); ok {
					spec.Env[k] = v
				}
				if client != "claude" || i+1 >= len(rest) || !looksLikeEnvPair(rest[i+1]) {
					break
				}
				hasInline = false
			}
		case "-H", "--header":
			h := val()
			if k, v, ok := strings.Cut(h, ":"); ok {
				spec.Headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		case "--url":
			spec.URL, spec.Transport = val(), TransportHTTP
		case "--bearer-token-env-var":
			spec.BearerTokenEnvVar = val()
		case "--client-id", "--client-secret", "--callback-port", "-c", "--config", "--enable", "--disable":
			if !hasInline && key != "--client-secret" {
				next()
			}
		default:
			if strings.HasPrefix(a, "-") {
				d.Notes = append(d.Notes, "忽略未识别参数 "+a)
				continue
			}
			positional = append(positional, a)
		}
	}
	if len(positional) == 0 {
		return ParseResult{}, fmt.Errorf("缺少服务名称")
	}
	d.Name = positional[0]
	cmd := positional[1:]
	if spec.URL == "" && len(cmd) > 0 && (strings.HasPrefix(cmd[0], "http://") || strings.HasPrefix(cmd[0], "https://")) {
		spec.URL = cmd[0]
		if spec.Transport == "" || spec.Transport == TransportStdio {
			spec.Transport = TransportHTTP
		}
	} else if spec.URL == "" {
		if len(cmd) == 0 {
			return ParseResult{}, fmt.Errorf("缺少启动命令或 URL")
		}
		spec.Transport, spec.Command, spec.Args = TransportStdio, cmd[0], cmd[1:]
	}
	d.Spec = spec.Normalize()
	if client == "claude" {
		d.SourceClient = ClientClaude
	} else {
		d.SourceClient = ClientCodex
	}
	return ParseResult{Format: client + "-cli", Drafts: []Draft{withInputs(d)}}, nil
}

func looksLikeEnvPair(s string) bool {
	k, _, ok := strings.Cut(s, "=")
	return ok && !strings.HasPrefix(s, "-") && regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(k)
}

func normalizeTransport(t string) string {
	switch strings.ToLower(t) {
	case "http", "streamable-http", "streamable_http":
		return TransportHTTP
	case "sse":
		return TransportSSE
	case "ws", "websocket":
		return TransportWS
	}
	return TransportStdio
}

// shellSplit splits a command line honouring single/double quotes and \.
func shellSplit(s string) []string {
	var out []string
	var cur strings.Builder
	inWord := false
	quote := rune(0)
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case quote == '"':
			if r == '"' {
				quote = 0
			} else if r == '\\' && i+1 < len(runes) && strings.ContainsRune(`"\$`+"`", runes[i+1]) {
				i++
				cur.WriteRune(runes[i])
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, inWord = r, true
		case r == '\\' && i+1 < len(runes):
			i++
			if runes[i] != '\n' {
				cur.WriteRune(runes[i])
				inWord = true
			}
		case r == ' ' || r == '\t' || r == '\n':
			if inWord {
				out = append(out, cur.String())
				cur.Reset()
				inWord = false
			}
		default:
			cur.WriteRune(r)
			inWord = true
		}
	}
	if inWord {
		out = append(out, cur.String())
	}
	return out
}

// ── names / inputs ──────────────────────────────────────────────────────────

func draftFromURL(u string) Draft {
	name := "remote"
	if parsed, err := url.Parse(u); err == nil {
		host := strings.TrimPrefix(parsed.Hostname(), "www.")
		host = strings.TrimPrefix(host, "mcp.")
		host = strings.TrimPrefix(host, "api.")
		if i := strings.Index(host, "."); i > 0 {
			host = host[:i]
		}
		name = SanitizeName(host)
	}
	transport := TransportHTTP
	if strings.HasSuffix(strings.TrimRight(u, "/"), "/sse") {
		transport = TransportSSE
	}
	d := Draft{Name: name, Spec: Spec{Transport: transport, URL: u}, Origin: "paste", Variant: "远程 · " + strings.ToUpper(transport)}
	d.Notes = append(d.Notes, "需要 OAuth 的远程服务：应用后运行 `claude mcp login "+name+"` / `codex mcp login "+name+"` 完成授权")
	return d
}

func nameFromSpec(s Spec) string {
	if s.URL != "" {
		return draftFromURL(s.URL).Name
	}
	return nameFromArgs(append([]string{s.Command}, s.Args...))
}

// nameFromArgs picks the package / image out of a launcher command line.
func nameFromArgs(args []string) string {
	for i := 1; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") || a == "run" || a == "--from" || strings.Contains(a, "=") {
			continue
		}
		if args[0] == "docker" && i > 0 && (args[i-1] == "-e" || args[i-1] == "--env" || args[i-1] == "-v") {
			continue
		}
		if strings.HasPrefix(a, "/") || strings.HasSuffix(a, ".js") || strings.HasSuffix(a, ".py") {
			return SanitizeName(strings.TrimSuffix(strings.TrimSuffix(a[strings.LastIndex(a, "/")+1:], ".js"), ".py"))
		}
		return SanitizeName(a)
	}
	return SanitizeName(args[0])
}

// SanitizeKeep keeps a pasted name when it is already valid.
func SanitizeKeep(name string) string {
	if ValidName(name) {
		return name
	}
	return SanitizeName(name)
}

func withInputs(d Draft) Draft {
	d.Spec = d.Spec.Normalize()
	d.Inputs = detectInputs(d.Spec)
	if d.Spec.Transport == TransportSSE || d.Spec.Transport == TransportWS {
		d.Notes = append(d.Notes, "Codex 不支持 "+strings.ToUpper(d.Spec.Transport)+" 传输，只能写入 Claude")
	}
	return d
}

// detectInputs flags env / header values that are empty or placeholders.
func detectInputs(s Spec) []InputDef {
	var out []InputDef
	for _, kind := range []struct {
		name string
		m    map[string]string
	}{{"env", s.Env}, {"header", s.Headers}} {
		for _, k := range sortedKeys(kind.m) {
			v := strings.TrimSpace(kind.m[k])
			if kind.name == "header" && strings.HasPrefix(v, "Bearer ") {
				v = strings.TrimSpace(strings.TrimPrefix(v, "Bearer "))
			}
			if v != "" && !placeholderRe.MatchString(v) {
				continue
			}
			out = append(out, InputDef{Kind: kind.name, Key: k, Required: true, Secret: secretKeyRe.MatchString(k)})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}
