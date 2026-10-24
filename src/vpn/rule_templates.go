package vpn

import (
	"fmt"
	"strings"
)

// ─── Scenario templates (常用场景) ───────────────────────────────────────────
//
// Each template is a ready-made fix for a common problem. Patterns use
// placeholders filled from the template inputs: a line containing a "list"
// input ({domains}, {apps}) is repeated once per value; a line whose
// placeholder input is left empty is skipped (optional inputs).

type TemplateInput struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Kind        string `json:"kind"` // list | text | target
	Placeholder string `json:"placeholder"`
	Default     string `json:"default"`
	Optional    bool   `json:"optional"`
	Hint        string `json:"hint"`
}

type RuleTemplate struct {
	ID       string          `json:"id"`
	Icon     string          `json:"icon"`
	Title    string          `json:"title"`
	Category string          `json:"category"` // fix | speed | direct | block | custom
	Problem  string          `json:"problem"`  // the symptom it solves, in user words
	Desc     string          `json:"desc"`     // what it adds
	Inputs   []TemplateInput `json:"inputs"`
	Rules    []string        `json:"rules"`
	Hosts    []string        `json:"hosts"`  // "domain=value"
	FakeIP   []string        `json:"fakeIp"` // fake-ip-filter entries
	DNS      []string        `json:"dns"`    // "domain=server"
}

var proxyTargetInput = TemplateInput{Key: "target", Label: "走哪个代理组", Kind: "target", Hint: "一般选订阅的主代理组即可"}

func domainRules(ruleType, target string, domains ...string) []string {
	out := make([]string, 0, len(domains))
	for _, d := range domains {
		out = append(out, ruleType+","+d+","+target)
	}
	return out
}

var ruleTemplates = []RuleTemplate{
	{
		ID: "intranet", Icon: "🏢", Title: "公司内网 / 内部服务直连", Category: "fix",
		Problem: "开着 VPN 打不开公司内网、测试环境、内部系统",
		Desc:    "让内网域名不走代理、不使用 Fake-IP；可选固定解析到指定 IP、或指定公司 DNS",
		Inputs: []TemplateInput{
			{Key: "domains", Label: "内网域名", Kind: "list", Placeholder: "corp.example.com\nlab.example.com", Hint: "填主域名即可，会自动包含所有子域名；多个用换行或逗号分隔"},
			{Key: "ip", Label: "固定解析到 IP（可选）", Kind: "text", Placeholder: "172.16.27.50", Optional: true, Hint: "内网 DNS 解析不到时才需要填"},
			{Key: "cidr", Label: "内网 IP 段（可选）", Kind: "text", Placeholder: "172.16.0.0/12", Optional: true, Hint: "直接用 IP 访问的内网服务也会直连"},
			{Key: "dns", Label: "公司 DNS（可选）", Kind: "text", Placeholder: "10.0.0.53", Optional: true, Hint: "让这些域名用公司 DNS 解析"},
		},
		Rules:  []string{"DOMAIN-SUFFIX,{domains},DIRECT", "IP-CIDR,{ip},DIRECT,no-resolve", "IP-CIDR,{cidr},DIRECT,no-resolve"},
		Hosts:  []string{"{domains}={ip}"},
		FakeIP: []string{"+.{domains}"},
		DNS:    []string{"+.{domains}={dns}"},
	},
	{
		ID: "lan", Icon: "🏠", Title: "局域网设备直连", Category: "fix",
		Problem: "开着 VPN 打不开路由器后台、NAS、打印机、投屏",
		Desc:    "私有网段和 .local / .lan 域名全部直连，并且不使用 Fake-IP",
		Rules: []string{
			"IP-CIDR,192.168.0.0/16,DIRECT,no-resolve", "IP-CIDR,10.0.0.0/8,DIRECT,no-resolve",
			"IP-CIDR,172.16.0.0/12,DIRECT,no-resolve", "IP-CIDR,127.0.0.0/8,DIRECT,no-resolve",
			"IP-CIDR,169.254.0.0/16,DIRECT,no-resolve", "DOMAIN-SUFFIX,local,DIRECT", "DOMAIN-SUFFIX,lan,DIRECT",
		},
		FakeIP: []string{"*.lan", "*.local", "+.home.arpa"},
	},
	{
		ID: "stun", Icon: "📞", Title: "语音通话 / 游戏联机修复", Category: "fix",
		Problem: "微信/QQ/腾讯会议通话失败，游戏 NAT 类型严格，系统提示无网络",
		Desc:    "让 STUN、主机联机和系统联网检测域名使用真实 IP（Fake-IP 过滤）",
		FakeIP: []string{
			"+.stun.*.*", "+.stun.*.*.*", "+.stun.*.*.*.*", "localhost.ptlogin2.qq.com",
			"*.srv.nintendo.net", "+.stun.playstation.net", "xbox.*.*.microsoft.com", "*.*.xboxlive.com",
			"+.msftncsi.com", "+.msftconnecttest.com", "time.*.com", "ntp.*.com",
		},
	},
	{
		ID: "ai", Icon: "🤖", Title: "AI 服务走代理", Category: "speed",
		Problem: "ChatGPT / Claude / Gemini 提示地区不可用或打不开",
		Desc:    "OpenAI、Anthropic、Google AI、Perplexity、Cursor 等域名强制走代理",
		Inputs:  []TemplateInput{proxyTargetInput},
		Rules: domainRules("DOMAIN-SUFFIX", "{target}",
			"openai.com", "chatgpt.com", "oaistatic.com", "oaiusercontent.com",
			"anthropic.com", "claude.ai", "claude.com", "claudeusercontent.com",
			"gemini.google.com", "generativelanguage.googleapis.com", "aistudio.google.com",
			"perplexity.ai", "x.ai", "grok.com", "cursor.com", "cursor.sh", "githubcopilot.com"),
	},
	{
		ID: "dev", Icon: "💻", Title: "开发工具加速", Category: "speed",
		Problem: "git clone、npm install、docker pull、pip install 很慢或超时",
		Desc:    "GitHub、npm、Docker Hub、PyPI、Go、Hugging Face 走代理",
		Inputs:  []TemplateInput{proxyTargetInput},
		Rules: domainRules("DOMAIN-SUFFIX", "{target}",
			"github.com", "githubusercontent.com", "githubassets.com", "github.io", "ghcr.io",
			"docker.io", "docker.com", "npmjs.org", "npmjs.com", "yarnpkg.com",
			"pypi.org", "pythonhosted.org", "go.dev", "golang.org", "huggingface.co", "hf.co"),
	},
	{
		ID: "media", Icon: "📺", Title: "流媒体走代理", Category: "speed",
		Problem: "YouTube / Netflix / Spotify 打不开或卡顿",
		Desc:    "常见海外流媒体域名走代理（建议选择速度快的代理组）",
		Inputs:  []TemplateInput{proxyTargetInput},
		Rules: domainRules("DOMAIN-SUFFIX", "{target}",
			"youtube.com", "googlevideo.com", "ytimg.com", "youtu.be", "netflix.com", "nflxvideo.net",
			"nflximg.net", "nflxext.com", "spotify.com", "scdn.co", "disneyplus.com", "twitch.tv"),
	},
	{
		ID: "telegram", Icon: "✈️", Title: "Telegram 走代理", Category: "speed",
		Problem: "Telegram 一直「连接中…」",
		Desc:    "Telegram 域名和官方 IP 段走代理",
		Inputs:  []TemplateInput{proxyTargetInput},
		Rules: append(domainRules("DOMAIN-SUFFIX", "{target}", "telegram.org", "t.me", "tdesktop.com", "telegra.ph"),
			"IP-CIDR,91.108.4.0/22,{target},no-resolve", "IP-CIDR,91.108.8.0/22,{target},no-resolve",
			"IP-CIDR,91.108.12.0/22,{target},no-resolve", "IP-CIDR,91.108.16.0/22,{target},no-resolve",
			"IP-CIDR,91.108.56.0/22,{target},no-resolve", "IP-CIDR,149.154.160.0/20,{target},no-resolve",
			"IP-CIDR6,2001:b28:f23d::/48,{target},no-resolve", "IP-CIDR6,2001:b28:f23f::/48,{target},no-resolve",
			"IP-CIDR6,2001:67c:4e8::/48,{target},no-resolve"),
	},
	{
		ID: "apple", Icon: "🍎", Title: "Apple 服务直连", Category: "direct",
		Problem: "App Store 下载慢、iCloud 同步异常、系统更新走了代理流量",
		Desc:    "Apple、iCloud 及其 CDN 域名直连",
		Rules: domainRules("DOMAIN-SUFFIX", "DIRECT",
			"apple.com", "icloud.com", "icloud-content.com", "mzstatic.com", "cdn-apple.com", "aaplimg.com", "apple-cloudkit.com"),
	},
	{
		ID: "cn", Icon: "🇨🇳", Title: "国内网站直连", Category: "direct",
		Problem: "国内网站（淘宝、B 站、银行）变慢或提示异地登录",
		Desc:    "用内置规则库把国内网站全部直连（GEOSITE,cn）",
		Rules:   []string{"GEOSITE,cn,DIRECT"},
	},
	{
		ID: "ads", Icon: "🛡️", Title: "屏蔽广告与跟踪", Category: "block",
		Problem: "网页、App 里广告太多",
		Desc:    "用内置广告规则库拦截广告域名（个别网站可能因此显示异常，可随时删除）",
		Rules:   []string{"GEOSITE,category-ads-all,REJECT"},
	},
	{
		ID: "site-direct", Icon: "➡️", Title: "指定网站直连", Category: "custom",
		Problem: "某个网站走代理后变慢、被风控或要求验证",
		Desc:    "让你填写的网站不走代理",
		Inputs:  []TemplateInput{{Key: "domains", Label: "网站域名", Kind: "list", Placeholder: "example.com", Hint: "可直接粘贴网址，会自动提取域名"}},
		Rules:   []string{"DOMAIN-SUFFIX,{domains},DIRECT"},
	},
	{
		ID: "site-proxy", Icon: "🌐", Title: "指定网站走代理", Category: "custom",
		Problem: "某个网站打不开，但订阅规则没覆盖它",
		Desc:    "让你填写的网站强制走代理",
		Inputs:  []TemplateInput{{Key: "domains", Label: "网站域名", Kind: "list", Placeholder: "example.com", Hint: "可直接粘贴网址，会自动提取域名"}, proxyTargetInput},
		Rules:   []string{"DOMAIN-SUFFIX,{domains},{target}"},
	},
	{
		ID: "site-block", Icon: "⛔", Title: "屏蔽指定网站", Category: "block",
		Problem: "想彻底阻止访问某些网站（防沉迷、拦截恶意域名）",
		Desc:    "访问这些网站会直接被拒绝",
		Inputs:  []TemplateInput{{Key: "domains", Label: "网站域名", Kind: "list", Placeholder: "example.com"}},
		Rules:   []string{"DOMAIN-SUFFIX,{domains},REJECT"},
	},
	{
		ID: "app-direct", Icon: "📱", Title: "指定应用直连", Category: "custom",
		Problem: "某个应用（企业微信、银行、远程桌面）走代理后无法登录",
		Desc:    "按应用进程名匹配，该应用的所有连接都不走代理",
		Inputs:  []TemplateInput{{Key: "apps", Label: "应用进程名", Kind: "list", Placeholder: "WeChat\n企业微信", Hint: "进程名可在「活动监视器」里看到，区分大小写"}},
		Rules:   []string{"PROCESS-NAME,{apps},DIRECT"},
	},
	{
		ID: "app-proxy", Icon: "🚀", Title: "指定应用走代理", Category: "custom",
		Problem: "某个应用（Discord、Steam、终端工具）需要全部走代理",
		Desc:    "按应用进程名匹配，该应用的所有连接都走代理",
		Inputs:  []TemplateInput{{Key: "apps", Label: "应用进程名", Kind: "list", Placeholder: "Discord", Hint: "进程名可在「活动监视器」里看到，区分大小写"}, proxyTargetInput},
		Rules:   []string{"PROCESS-NAME,{apps},{target}"},
	},
	{
		ID: "hosts", Icon: "📌", Title: "域名固定解析到 IP", Category: "custom",
		Problem: "开发测试时想让某个域名指向指定服务器（类似改 /etc/hosts）",
		Desc:    "添加 Hosts 映射，并让该域名使用真实 IP",
		Inputs: []TemplateInput{
			{Key: "domains", Label: "域名", Kind: "list", Placeholder: "api.test.local"},
			{Key: "ip", Label: "解析到 IP", Kind: "text", Placeholder: "192.168.1.20"},
		},
		Hosts:  []string{"{domains}={ip}"},
		FakeIP: []string{"{domains}"},
	},
	{
		ID: "dns", Icon: "🧭", Title: "指定域名用指定 DNS", Category: "custom",
		Problem: "某些域名需要专用 DNS 才能解析（公司 DNS、特定解析服务）",
		Desc:    "为这些域名单独指定 DNS 服务器",
		Inputs: []TemplateInput{
			{Key: "domains", Label: "域名", Kind: "list", Placeholder: "corp.example.com"},
			{Key: "dns", Label: "DNS 服务器", Kind: "text", Placeholder: "10.0.0.53 或 https://dns.alidns.com/dns-query"},
		},
		DNS: []string{"+.{domains}={dns}"},
	},
}

// RuleTemplates returns the scenario catalog.
func RuleTemplates() []RuleTemplate { return ruleTemplates }

// ExpandTemplate turns a template + user inputs into draft entries the UI
// appends to the editor (nothing is saved here).
func ExpandTemplate(id string, inputs map[string]string) (RulesDraft, error) {
	var tpl *RuleTemplate
	for i := range ruleTemplates {
		if ruleTemplates[i].ID == id {
			tpl = &ruleTemplates[i]
		}
	}
	if tpl == nil {
		return RulesDraft{}, fmt.Errorf("未知模板: %s", id)
	}
	values := map[string][]string{}
	for _, in := range tpl.Inputs {
		raw := strings.TrimSpace(inputs[in.Key])
		if raw == "" {
			raw = in.Default
		}
		var vals []string
		if in.Kind == "list" {
			for _, v := range splitList(raw) {
				if in.Key == "domains" {
					v = strings.TrimPrefix(strings.TrimPrefix(cleanHost(v), "+."), "*.")
				}
				if v != "" {
					vals = append(vals, v)
				}
			}
		} else if raw != "" {
			vals = []string{raw}
		}
		if len(vals) == 0 && !in.Optional {
			return RulesDraft{}, fmt.Errorf("请填写「%s」", in.Label)
		}
		values[in.Key] = vals
	}

	note := tpl.Title
	out := RulesDraft{Prepend: []RuleItem{}, Hosts: []HostEntry{}, FakeIPFilter: []FakeIPEntry{}, DNSPolicy: []DNSPolicyEntry{}}
	for _, line := range expandPatterns(tpl.Rules, values) {
		r := ParseRule(line)
		r.Note = note
		n, err := normalizeRule(r)
		if err != nil {
			return RulesDraft{}, err
		}
		out.Prepend = append(out.Prepend, n)
	}
	for _, line := range expandPatterns(tpl.Hosts, values) {
		domain, value, _ := strings.Cut(line, "=")
		out.Hosts = append(out.Hosts, HostEntry{Domain: domain, Value: value, Note: note})
	}
	for _, line := range expandPatterns(tpl.FakeIP, values) {
		out.FakeIPFilter = append(out.FakeIPFilter, FakeIPEntry{Value: line, Note: note})
	}
	for _, line := range expandPatterns(tpl.DNS, values) {
		domain, server, _ := strings.Cut(line, "=")
		out.DNSPolicy = append(out.DNSPolicy, DNSPolicyEntry{Domain: domain, Server: server, Note: note})
	}
	return out, nil
}

// expandPatterns substitutes {key} placeholders; a multi-valued key repeats
// the line, an empty key drops it.
func expandPatterns(patterns []string, values map[string][]string) []string {
	var out []string
	for _, p := range patterns {
		lines := []string{p}
		for key, vals := range values {
			ph := "{" + key + "}"
			var next []string
			for _, l := range lines {
				if !strings.Contains(l, ph) {
					next = append(next, l)
					continue
				}
				for _, v := range vals {
					next = append(next, strings.ReplaceAll(l, ph, v))
				}
			}
			lines = next
		}
		for _, l := range lines {
			if !strings.Contains(l, "{") {
				out = append(out, l)
			}
		}
	}
	return out
}
