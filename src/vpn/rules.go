package vpn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"hatch/core/fsutil"
	"hatch/core/strutil"

	"gopkg.in/yaml.v3"
)

// ─── Custom rules (自定义规则) ────────────────────────────────────────────────
//
// The editor works on the *current* subscription's two enhancement files that
// Clash Verge Rev already links in profiles.yaml (`option.rules` and
// `option.merge`):
//
//   rules file  prepend / append (delete is preserved untouched)
//   merge file  hosts, dns.fake-ip-filter, dns.nameserver-policy (other keys kept)
//
// Saving patches the generated runtime (clash-verge.yaml) the same way Verge
// would and hot-reloads it through the core's Unix-socket API, so edits take
// effect immediately. The core validates the payload first: if it rejects the
// config nothing is written.

type RuleItem struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Target    string `json:"target"`
	NoResolve bool   `json:"noResolve"`
	Raw       string `json:"raw"` // non-empty → advanced rule kept verbatim (AND/OR, MATCH, extra params…)
	Note      string `json:"note"`
}

type HostEntry struct {
	Domain string `json:"domain"`
	Value  string `json:"value"` // one or more IPs / domains, comma separated
	Note   string `json:"note"`
}

type FakeIPEntry struct {
	Value string `json:"value"`
	Note  string `json:"note"`
}

type DNSPolicyEntry struct {
	Domain string `json:"domain"`
	Server string `json:"server"` // one or more DNS servers, comma separated
	Note   string `json:"note"`
}

// RulesDraft is everything the custom-rules page edits.
type RulesDraft struct {
	Prepend      []RuleItem       `json:"prepend"`
	Append       []RuleItem       `json:"append"`
	Hosts        []HostEntry      `json:"hosts"`
	FakeIPFilter []FakeIPEntry    `json:"fakeIpFilter"`
	DNSPolicy    []DNSPolicyEntry `json:"dnsPolicy"`
	TunExclude   []TunEntry       `json:"tunExclude"` // global: Clash Verge config.yaml tun.route-exclude-address
}

// TunEntry is one CIDR whose traffic bypasses the TUN device entirely.
type TunEntry struct {
	Value string `json:"value"`
}

type RuleTypeInfo struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Desc        string `json:"desc"`
	Placeholder string `json:"placeholder"`
	NoResolve   bool   `json:"noResolve"` // accepts the no-resolve flag
	Advanced    bool   `json:"advanced"`
}

type RulesState struct {
	Draft                 RulesDraft     `json:"draft"`
	ProfileUID            string         `json:"profileUid"`
	ProfileName           string         `json:"profileName"`
	RulesFile             string         `json:"rulesFile"`
	MergeFile             string         `json:"mergeFile"`
	Groups                []string       `json:"groups"`
	RuleTypes             []RuleTypeInfo `json:"ruleTypes"`
	SubscriptionRuleCount int            `json:"subscriptionRuleCount"`
	SubscriptionEndsMatch bool           `json:"subscriptionEndsMatch"`
	SubscriptionFakeIP    []string       `json:"subscriptionFakeIp"`
	FakeIPOverridden      bool           `json:"fakeIpOverridden"`
	MissingFakeIP         []string       `json:"missingFakeIp"`
	DeleteCount           int            `json:"deleteCount"`
	CoreReachable         bool           `json:"coreReachable"`
	TunEnabled            bool           `json:"tunEnabled"`
	Warnings              []string       `json:"warnings"`
}

// RuleIssue points at one invalid draft entry so the UI can highlight it.
type RuleIssue struct {
	Section string `json:"section"` // prepend | append | hosts | fakeip | dns
	Index   int    `json:"index"`
	Message string `json:"message"`
}

type RulesSaveResult struct {
	Saved    bool        `json:"saved"`
	Applied  bool        `json:"applied"` // hot-reloaded into the running core
	Message  string      `json:"message"`
	Issues   []RuleIssue `json:"issues"`
	Warnings []string    `json:"warnings"`
}

// ruleTypes is the catalog shown in the editor; order = dropdown order.
var ruleTypes = []RuleTypeInfo{
	{Type: "DOMAIN-SUFFIX", Label: "域名（含子域名）", Desc: "最常用。填 example.com，会同时匹配 www.example.com、api.example.com", Placeholder: "example.com"},
	{Type: "DOMAIN", Label: "完整域名（精确）", Desc: "只匹配这一个域名，不包含它的子域名", Placeholder: "www.example.com"},
	{Type: "DOMAIN-KEYWORD", Label: "域名关键词", Desc: "域名里包含这个词就命中，如 google 会匹配 google.com、googleapis.com", Placeholder: "google"},
	{Type: "GEOSITE", Label: "网站分类（规则库）", Desc: "用内置规则库按类别匹配，如 openai、github、google、cn（国内网站）、category-ads-all（广告）", Placeholder: "openai"},
	{Type: "IP-CIDR", Label: "IP / IP 段", Desc: "填单个 IP 或网段，如 1.2.3.4 或 10.0.0.0/8；单个 IP 会自动补成 /32", Placeholder: "192.168.1.0/24", NoResolve: true},
	{Type: "GEOIP", Label: "国家/地区 IP", Desc: "按 IP 归属地匹配，如 CN（中国大陆）、LAN（局域网）", Placeholder: "CN", NoResolve: true},
	{Type: "PROCESS-NAME", Label: "应用（进程名）", Desc: "按发起连接的应用匹配，如 WeChat、Telegram；进程名可在「活动监视器」里查看", Placeholder: "WeChat"},
	{Type: "DST-PORT", Label: "目标端口", Desc: "按端口匹配，如 22（SSH）；支持 8000-9000 范围，多个用 / 分隔", Placeholder: "22"},
	{Type: "IP-CIDR6", Label: "IPv6 段", Desc: "IPv6 网段，如 2001:db8::/32", Placeholder: "2001:db8::/32", NoResolve: true, Advanced: true},
	{Type: "DOMAIN-REGEX", Label: "域名正则", Desc: "用正则表达式匹配域名（不能包含英文逗号）", Placeholder: `^.*\.example\.com$`, Advanced: true},
	{Type: "PROCESS-PATH", Label: "应用路径", Desc: "按应用可执行文件完整路径匹配", Placeholder: "/Applications/WeChat.app/Contents/MacOS/WeChat", Advanced: true},
	{Type: "SRC-IP-CIDR", Label: "来源 IP 段", Desc: "按发起连接的设备 IP 匹配（局域网共享时有用）", Placeholder: "192.168.1.10/32", Advanced: true},
	{Type: "SRC-PORT", Label: "来源端口", Desc: "按本地来源端口匹配", Placeholder: "50000-60000", Advanced: true},
	{Type: "NETWORK", Label: "传输协议", Desc: "tcp 或 udp", Placeholder: "udp", Advanced: true},
	{Type: "IP-ASN", Label: "IP 运营商 (ASN)", Desc: "按 IP 所属自治系统号匹配，如 13335（Cloudflare）", Placeholder: "13335", NoResolve: true, Advanced: true},
}

func ruleTypeInfo(t string) (RuleTypeInfo, bool) {
	for _, info := range ruleTypes {
		if info.Type == t {
			return info, true
		}
	}
	return RuleTypeInfo{}, false
}

// ─── Parse / format ──────────────────────────────────────────────────────────

func splitRule(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// canonicalRule removes spacing differences so file and runtime copies compare equal.
func canonicalRule(s string) string { return strings.Join(splitRule(s), ",") }

// ParseRule turns "TYPE,value,target[,no-resolve]" into a RuleItem. Anything
// that does not fit the simple form is kept verbatim in Raw.
func ParseRule(s string) RuleItem {
	s = strings.TrimSpace(s)
	parts := splitRule(s)
	item := RuleItem{Raw: s}
	if len(parts) == 0 {
		return item
	}
	t := strings.ToUpper(parts[0])
	item.Type = t
	if len(parts) >= 2 {
		item.Target = parts[len(parts)-1]
		if strings.EqualFold(item.Target, "no-resolve") && len(parts) >= 3 {
			item.Target = parts[len(parts)-2]
		}
	}
	info, known := ruleTypeInfo(t)
	if !known {
		return item
	}
	switch {
	case len(parts) == 3:
		return RuleItem{Type: t, Value: parts[1], Target: parts[2]}
	case len(parts) == 4 && strings.EqualFold(parts[3], "no-resolve") && info.NoResolve:
		return RuleItem{Type: t, Value: parts[1], Target: parts[2], NoResolve: true}
	}
	return item
}

// String renders the rule as Clash expects it.
func (r RuleItem) String() string {
	if strings.TrimSpace(r.Raw) != "" {
		return canonicalRule(r.Raw)
	}
	s := r.Type + "," + r.Value + "," + r.Target
	if info, ok := ruleTypeInfo(r.Type); ok && info.NoResolve && r.NoResolve {
		s += ",no-resolve"
	}
	return s
}

// ─── Normalize / validate ────────────────────────────────────────────────────

var (
	portSpecRe   = regexp.MustCompile(`^\d+(-\d+)?(/\d+(-\d+)?)*$`)
	geoCodeRe    = regexp.MustCompile(`^!?[A-Za-z0-9_@-]+$`)
	digitsRe     = regexp.MustCompile(`^\d+$`)
	whitespaceRe = regexp.MustCompile(`\s`)
)

// cleanHost turns pasted URLs like "https://www.a.com:8443/x" into "www.a.com".
func cleanHost(v string) string {
	v = strings.TrimSpace(v)
	if i := strings.Index(v, "://"); i >= 0 {
		v = v[i+3:]
	}
	if i := strings.IndexAny(v, "/?#"); i >= 0 {
		v = v[:i]
	}
	if i := strings.LastIndex(v, "@"); i >= 0 {
		v = v[i+1:]
	}
	if h, _, err := net.SplitHostPort(v); err == nil {
		v = h
	}
	v = strings.Trim(v, "[]")
	return strings.ToLower(strings.Trim(v, "."))
}

func plainIP(v string) net.IP {
	return net.ParseIP(strings.Trim(strings.TrimSpace(v), "[]"))
}

// normalizeCIDR accepts "1.2.3.4" → "1.2.3.4/32" and reports whether it is IPv6.
func normalizeCIDR(v string) (string, bool, error) {
	v = strings.TrimSpace(v)
	if ip := plainIP(v); ip != nil {
		if ip.To4() != nil {
			return ip.String() + "/32", false, nil
		}
		return ip.String() + "/128", true, nil
	}
	ip, n, err := net.ParseCIDR(v)
	if err != nil {
		return "", false, fmt.Errorf("「%s」不是有效的 IP 或网段，例：192.168.1.0/24", v)
	}
	// Keep the address as written (e.g. 172.19.0.0/12): the core accepts it,
	// and rewriting it to the network address would silently change user data.
	ones, _ := n.Mask.Size()
	return fmt.Sprintf("%s/%d", ip.String(), ones), ip.To4() == nil, nil
}

func checkDomainLike(v string) error {
	if v == "" {
		return errors.New("域名不能为空")
	}
	if whitespaceRe.MatchString(v) || strings.ContainsAny(v, ",:/") {
		return fmt.Errorf("「%s」不是有效的域名", v)
	}
	return nil
}

// normalizeRule cleans user input (pasted URLs, bare IPs…) and validates it.
func normalizeRule(r RuleItem) (RuleItem, error) {
	r.Note = strings.TrimSpace(strings.ReplaceAll(r.Note, "\n", " "))
	if raw := strings.TrimSpace(r.Raw); raw != "" {
		if strings.ContainsAny(raw, "\n\r") {
			return r, errors.New("一条规则只能写一行")
		}
		if len(splitRule(raw)) < 2 {
			return r, fmt.Errorf("「%s」格式不对，应为 类型,值,目标", raw)
		}
		parsed := ParseRule(raw)
		if parsed.Raw == "" {
			parsed.Note = r.Note
			return normalizeRule(parsed)
		}
		r.Raw = canonicalRule(raw)
		return r, nil
	}

	r.Type = strings.ToUpper(strings.TrimSpace(r.Type))
	r.Value = strings.TrimSpace(r.Value)
	r.Target = strings.TrimSpace(r.Target)
	info, ok := ruleTypeInfo(r.Type)
	if !ok {
		return r, fmt.Errorf("不支持的规则类型「%s」，复杂规则请用「原样编辑」", r.Type)
	}
	if r.Target == "" {
		return r, errors.New("请选择命中后去哪里（直连 / 代理组 / 拦截）")
	}
	if strings.ContainsAny(r.Target, ",\n") {
		return r, fmt.Errorf("目标「%s」包含非法字符", r.Target)
	}
	if r.Value == "" {
		return r, fmt.Errorf("请填写%s", info.Label)
	}
	if strings.Contains(r.Value, ",") {
		return r, errors.New("值里不能包含英文逗号，多个值请拆成多条规则")
	}

	switch r.Type {
	case "DOMAIN", "DOMAIN-SUFFIX":
		r.Value = cleanHost(r.Value)
		if r.Type == "DOMAIN-SUFFIX" {
			r.Value = strings.TrimPrefix(strings.TrimPrefix(r.Value, "+."), "*.")
		}
		if err := checkDomainLike(r.Value); err != nil {
			return r, err
		}
		if plainIP(r.Value) != nil {
			r.Type, r.NoResolve = "IP-CIDR", true
			return normalizeRule(r)
		}
	case "DOMAIN-KEYWORD":
		r.Value = strings.ToLower(r.Value)
		if whitespaceRe.MatchString(r.Value) {
			return r, errors.New("关键词不能包含空格")
		}
	case "DOMAIN-REGEX":
		if _, err := regexp.Compile(r.Value); err != nil {
			return r, fmt.Errorf("正则表达式无效: %v", err)
		}
	case "GEOSITE":
		r.Value = strings.ToLower(r.Value)
		if !geoCodeRe.MatchString(r.Value) {
			return r, fmt.Errorf("「%s」不是有效的分类名，例：openai、cn", r.Value)
		}
	case "GEOIP":
		if !geoCodeRe.MatchString(r.Value) {
			return r, fmt.Errorf("「%s」不是有效的地区代码，例：CN、LAN", r.Value)
		}
		if len(r.Value) == 2 {
			r.Value = strings.ToUpper(r.Value)
		}
	case "IP-CIDR", "IP-CIDR6":
		cidr, v6, err := normalizeCIDR(r.Value)
		if err != nil {
			return r, err
		}
		r.Value = cidr
		r.Type = "IP-CIDR"
		if v6 {
			r.Type = "IP-CIDR6"
		}
	case "SRC-IP-CIDR":
		cidr, _, err := normalizeCIDR(r.Value)
		if err != nil {
			return r, err
		}
		r.Value = cidr
	case "DST-PORT", "SRC-PORT":
		if !portSpecRe.MatchString(r.Value) {
			return r, fmt.Errorf("「%s」不是有效的端口，例：22、8000-9000、80/443", r.Value)
		}
	case "NETWORK":
		r.Value = strings.ToLower(r.Value)
		if r.Value != "tcp" && r.Value != "udp" {
			return r, errors.New("协议只能是 tcp 或 udp")
		}
	case "IP-ASN":
		if !digitsRe.MatchString(r.Value) {
			return r, fmt.Errorf("「%s」不是有效的 ASN 号", r.Value)
		}
	case "PROCESS-NAME", "PROCESS-PATH":
		// free text, comma already rejected
	}
	if !info.NoResolve {
		r.NoResolve = false
	}
	return r, nil
}

func normalizeHostEntry(h HostEntry) (HostEntry, error) {
	h.Domain = strings.TrimSpace(h.Domain)
	if !strings.HasPrefix(h.Domain, "*") && !strings.HasPrefix(h.Domain, "+.") {
		h.Domain = cleanHost(h.Domain)
	}
	h.Note = strings.TrimSpace(h.Note)
	if err := checkDomainLike(h.Domain); err != nil {
		return h, err
	}
	vals := splitList(h.Value)
	if len(vals) == 0 {
		return h, fmt.Errorf("请填写「%s」要解析到的 IP", h.Domain)
	}
	for _, v := range vals {
		if plainIP(v) == nil && checkDomainLike(v) != nil {
			return h, fmt.Errorf("「%s」不是有效的 IP 或域名", v)
		}
	}
	h.Value = strings.Join(vals, ", ")
	return h, nil
}

func normalizeFakeIP(f FakeIPEntry) (FakeIPEntry, error) {
	f.Value = strings.ToLower(strings.TrimSpace(f.Value))
	if strings.Contains(f.Value, "://") {
		f.Value = cleanHost(f.Value)
	}
	f.Note = strings.TrimSpace(f.Note)
	if f.Value == "" || whitespaceRe.MatchString(f.Value) || strings.ContainsAny(f.Value, ",/") {
		return f, fmt.Errorf("「%s」不是有效的域名规则，例：*.lan、+.example.com", f.Value)
	}
	return f, nil
}

func normalizeDNSPolicy(p DNSPolicyEntry) (DNSPolicyEntry, error) {
	p.Domain = strings.TrimSpace(p.Domain)
	p.Note = strings.TrimSpace(p.Note)
	if p.Domain == "" || whitespaceRe.MatchString(p.Domain) {
		return p, errors.New("请填写域名，例：+.corp.example.com")
	}
	servers := splitList(p.Server)
	if len(servers) == 0 {
		return p, fmt.Errorf("请填写「%s」使用的 DNS 服务器", p.Domain)
	}
	for _, s := range servers {
		if whitespaceRe.MatchString(s) {
			return p, fmt.Errorf("「%s」不是有效的 DNS 地址", s)
		}
	}
	p.Server = strings.Join(servers, ", ")
	return p, nil
}

// splitList splits on commas / newlines / whitespace and drops blanks.
func splitList(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '，' || r == '\n' || r == '\r' || r == ' ' || r == '\t' || r == '、' || r == ';'
	})
}

// NormalizeDraft cleans every entry; issues lists the entries that are invalid.
func NormalizeDraft(d RulesDraft) (RulesDraft, []RuleIssue) {
	var out RulesDraft
	var issues []RuleIssue
	add := func(section string, i int, err error) {
		issues = append(issues, RuleIssue{Section: section, Index: i, Message: err.Error()})
	}
	for _, list := range []struct {
		name string
		in   []RuleItem
		out  *[]RuleItem
	}{{"prepend", d.Prepend, &out.Prepend}, {"append", d.Append, &out.Append}} {
		*list.out = []RuleItem{}
		for i, r := range list.in {
			n, err := normalizeRule(r)
			if err != nil {
				add(list.name, i, err)
			}
			*list.out = append(*list.out, n)
		}
	}
	out.Hosts = []HostEntry{}
	seenHost := map[string]int{}
	for i, h := range d.Hosts {
		n, err := normalizeHostEntry(h)
		if err == nil {
			if j, dup := seenHost[n.Domain]; dup {
				err = fmt.Errorf("「%s」与第 %d 条重复", n.Domain, j+1)
			}
			seenHost[n.Domain] = i
		}
		if err != nil {
			add("hosts", i, err)
		}
		out.Hosts = append(out.Hosts, n)
	}
	out.FakeIPFilter = []FakeIPEntry{}
	for i, f := range d.FakeIPFilter {
		n, err := normalizeFakeIP(f)
		if err != nil {
			add("fakeip", i, err)
		}
		out.FakeIPFilter = append(out.FakeIPFilter, n)
	}
	out.DNSPolicy = []DNSPolicyEntry{}
	seenPolicy := map[string]int{}
	for i, p := range d.DNSPolicy {
		n, err := normalizeDNSPolicy(p)
		if err == nil {
			if j, dup := seenPolicy[n.Domain]; dup {
				err = fmt.Errorf("「%s」与第 %d 条重复", n.Domain, j+1)
			}
			seenPolicy[n.Domain] = i
		}
		if err != nil {
			add("dns", i, err)
		}
		out.DNSPolicy = append(out.DNSPolicy, n)
	}
	out.TunExclude = []TunEntry{}
	seenTun := map[string]int{}
	for i, t := range d.TunExclude {
		cidr, _, err := normalizeCIDR(t.Value)
		if err == nil {
			if j, dup := seenTun[cidr]; dup {
				err = fmt.Errorf("「%s」与第 %d 条重复", cidr, j+1)
			}
			seenTun[cidr] = i
		} else {
			cidr = strings.TrimSpace(t.Value)
		}
		if err != nil {
			add("tun", i, err)
		}
		out.TunExclude = append(out.TunExclude, TunEntry{Value: cidr})
	}
	return out, issues
}

// ParseRuleLines turns pasted text into rules: full "TYPE,value,target" lines
// are parsed, bare domains / IPs / URLs become DOMAIN-SUFFIX / IP-CIDR rules
// pointing at target.
func ParseRuleLines(text, target string) []RuleItem {
	out := []RuleItem{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, " #"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "- "))
		line = strings.Trim(line, `'"`)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, ",") {
			out = append(out, ParseRule(line))
			continue
		}
		for _, v := range splitList(line) {
			r := RuleItem{Type: "DOMAIN-SUFFIX", Value: v, Target: target}
			if _, _, err := normalizeCIDR(v); err == nil {
				r.Type, r.NoResolve = "IP-CIDR", true
			}
			if n, err := normalizeRule(r); err == nil {
				r = n
			}
			out = append(out, r)
		}
	}
	return out
}

// ─── Current subscription + its enhancement files ────────────────────────────

type profileItem struct {
	UID      string `yaml:"uid"`
	Type     string `yaml:"type"`
	Name     string `yaml:"name"`
	File     string `yaml:"file"`
	URL      string `yaml:"url"`
	Updated  int64  `yaml:"updated"`
	Selected []struct {
		Name string `yaml:"name"`
		Now  string `yaml:"now"`
	} `yaml:"selected"`
	Extra struct {
		Upload   int64 `yaml:"upload"`
		Download int64 `yaml:"download"`
		Total    int64 `yaml:"total"`
		Expire   int64 `yaml:"expire"`
	} `yaml:"extra"`
	Option struct {
		UpdateInterval  int    `yaml:"update_interval"`
		AllowAutoUpdate *bool  `yaml:"allow_auto_update"`
		Merge           string `yaml:"merge"`
		Script          string `yaml:"script"`
		Rules           string `yaml:"rules"`
		Proxies         string `yaml:"proxies"`
		Groups          string `yaml:"groups"`
	} `yaml:"option"`
}

type profilesMeta struct {
	Current string        `yaml:"current"`
	Items   []profileItem `yaml:"items"`
}

func (p profilesMeta) find(uid string) (profileItem, bool) {
	for _, it := range p.Items {
		if it.UID == uid {
			return it, true
		}
	}
	return profileItem{}, false
}

func (m Manager) readProfilesMeta() (string, profilesMeta, error) {
	dir, err := m.DataDir()
	if err != nil {
		return "", profilesMeta{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "profiles.yaml"))
	if err != nil {
		return "", profilesMeta{}, fmt.Errorf("读取 profiles.yaml 失败: %w", err)
	}
	var meta profilesMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return "", profilesMeta{}, fmt.Errorf("解析 profiles.yaml 失败: %w", err)
	}
	return dir, meta, nil
}

// enhanceCtx locates the current subscription and its linked enhancement files.
type enhanceCtx struct {
	dir         string
	profile     profileItem
	subPath     string
	rulesPath   string // "" when the subscription has no rules enhancement
	mergePath   string
	runtimePath string
	clashPath   string // Clash Verge base config (config.yaml): TUN settings live here
}

func (m Manager) currentEnhance() (enhanceCtx, error) {
	dir, meta, err := m.readProfilesMeta()
	if err != nil {
		return enhanceCtx{}, err
	}
	cur, ok := meta.find(meta.Current)
	if !ok || meta.Current == "" {
		return enhanceCtx{}, errors.New("当前没有激活的订阅，请先在 Clash Verge 中选择一个订阅")
	}
	ctx := enhanceCtx{
		dir:         dir,
		profile:     cur,
		subPath:     filepath.Join(dir, "profiles", cur.File),
		runtimePath: filepath.Join(dir, "clash-verge.yaml"),
		clashPath:   filepath.Join(dir, "config.yaml"),
	}
	resolve := func(uid string) string {
		if it, ok := meta.find(uid); ok && it.File != "" && safeProfilePath(it.File) {
			return filepath.Join(dir, "profiles", it.File)
		}
		return ""
	}
	ctx.rulesPath = resolve(cur.Option.Rules)
	ctx.mergePath = resolve(cur.Option.Merge)
	return ctx, nil
}

// enhanceFiles is what the two enhancement files currently say.
type enhanceFiles struct {
	rulesDoc, mergeDoc, clashDoc *yaml.Node
	draft                        RulesDraft
	fakeIPSet                    bool // merge file sets dns.fake-ip-filter (replacing the subscription list)
	deleteCount                  int
}

func noteOf(n *yaml.Node) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(n.LineComment), "#"))
}

func scalarList(n *yaml.Node) []string {
	if n == nil {
		return nil
	}
	if n.Kind == yaml.ScalarNode {
		if n.Value == "" {
			return nil
		}
		return []string{n.Value}
	}
	var out []string
	for _, c := range n.Content {
		if c.Kind == yaml.ScalarNode {
			out = append(out, c.Value)
		}
	}
	return out
}

func readEnhanceFiles(ctx enhanceCtx) (enhanceFiles, error) {
	ef := enhanceFiles{draft: RulesDraft{
		Prepend: []RuleItem{}, Append: []RuleItem{}, Hosts: []HostEntry{},
		FakeIPFilter: []FakeIPEntry{}, DNSPolicy: []DNSPolicyEntry{}, TunExclude: []TunEntry{},
	}}
	var err error
	if ef.rulesDoc, err = loadYAMLDoc(ctx.rulesPath); err != nil {
		return ef, err
	}
	if ef.mergeDoc, err = loadYAMLDoc(ctx.mergePath); err != nil {
		return ef, err
	}
	if ef.clashDoc, err = loadYAMLDoc(ctx.clashPath); err != nil {
		return ef, err
	}
	for _, v := range scalarList(mapGet(mapGet(ef.clashDoc.Content[0], "tun"), "route-exclude-address")) {
		ef.draft.TunExclude = append(ef.draft.TunExclude, TunEntry{Value: v})
	}

	rules := ef.rulesDoc.Content[0]
	for _, sec := range []struct {
		key string
		out *[]RuleItem
	}{{"prepend", &ef.draft.Prepend}, {"append", &ef.draft.Append}} {
		if seq := mapGet(rules, sec.key); seq != nil && seq.Kind == yaml.SequenceNode {
			for _, c := range seq.Content {
				if c.Kind != yaml.ScalarNode {
					continue
				}
				item := ParseRule(c.Value)
				item.Note = noteOf(c)
				*sec.out = append(*sec.out, item)
			}
		}
	}
	if del := mapGet(rules, "delete"); del != nil {
		ef.deleteCount = len(del.Content)
	}

	merge := ef.mergeDoc.Content[0]
	if hosts := mapGet(merge, "hosts"); hosts != nil && hosts.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(hosts.Content); i += 2 {
			k, v := hosts.Content[i], hosts.Content[i+1]
			note := noteOf(k)
			if note == "" {
				note = noteOf(v)
			}
			ef.draft.Hosts = append(ef.draft.Hosts, HostEntry{Domain: k.Value, Value: strings.Join(scalarList(v), ", "), Note: note})
		}
	}
	if dns := mapGet(merge, "dns"); dns != nil && dns.Kind == yaml.MappingNode {
		if f := mapGet(dns, "fake-ip-filter"); f != nil {
			ef.fakeIPSet = true
			for _, c := range f.Content {
				ef.draft.FakeIPFilter = append(ef.draft.FakeIPFilter, FakeIPEntry{Value: c.Value, Note: noteOf(c)})
			}
		}
		if pol := mapGet(dns, "nameserver-policy"); pol != nil && pol.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(pol.Content); i += 2 {
				k, v := pol.Content[i], pol.Content[i+1]
				note := noteOf(k)
				if note == "" {
					note = noteOf(v)
				}
				ef.draft.DNSPolicy = append(ef.draft.DNSPolicy, DNSPolicyEntry{Domain: k.Value, Server: strings.Join(scalarList(v), ", "), Note: note})
			}
		}
	}
	return ef, nil
}

// subscriptionInfo reads what the subscription itself ships: rules, groups and
// its own fake-ip-filter list (which a merge override replaces wholesale).
type subscriptionInfo struct {
	rules     []string
	groups    []string
	fakeIP    []string
	hasFakeIP bool
}

func readSubscription(path string) subscriptionInfo {
	var info subscriptionInfo
	data, err := os.ReadFile(path)
	if err != nil {
		return info
	}
	var raw struct {
		Rules       []string `yaml:"rules"`
		ProxyGroups []struct {
			Name string `yaml:"name"`
		} `yaml:"proxy-groups"`
		DNS struct {
			FakeIPFilter []string `yaml:"fake-ip-filter"`
		} `yaml:"dns"`
	}
	if yaml.Unmarshal(data, &raw) != nil {
		return info
	}
	var probe struct {
		DNS map[string]interface{} `yaml:"dns"`
	}
	_ = yaml.Unmarshal(data, &probe)
	_, info.hasFakeIP = probe.DNS["fake-ip-filter"]
	info.rules = raw.Rules
	info.fakeIP = raw.DNS.FakeIPFilter
	for _, g := range raw.ProxyGroups {
		if g.Name != "" {
			info.groups = append(info.groups, g.Name)
		}
	}
	return info
}

// runtimeRules returns the rules list of the generated clash-verge.yaml.
func readRuntime(path string) (*yaml.Node, []string, []string) {
	doc, err := loadYAMLDoc(path)
	if err != nil {
		return nil, nil, nil
	}
	root := doc.Content[0]
	rules := scalarList(mapGet(root, "rules"))
	var groups []string
	if pg := mapGet(root, "proxy-groups"); pg != nil {
		for _, g := range pg.Content {
			if n := mapGet(g, "name"); n != nil && n.Value != "" {
				groups = append(groups, n.Value)
			}
		}
	}
	return doc, rules, groups
}

// stripCustom removes the old prepend prefix / append suffix from the runtime
// rules, leaving what the subscription (plus global enhancements) contributes.
func stripCustom(runtime []string, oldPrepend, oldAppend []string) []string {
	out := runtime
	if hasPrefixRules(out, oldPrepend) {
		out = out[len(oldPrepend):]
	}
	if len(oldAppend) > 0 && len(out) >= len(oldAppend) && hasPrefixRules(out[len(out)-len(oldAppend):], oldAppend) {
		out = out[:len(out)-len(oldAppend)]
	}
	return out
}

func hasPrefixRules(list, prefix []string) bool {
	if len(prefix) > len(list) {
		return false
	}
	for i, p := range prefix {
		if canonicalRule(list[i]) != canonicalRule(p) {
			return false
		}
	}
	return true
}

func ruleStrings(items []RuleItem) []string {
	out := make([]string, 0, len(items))
	for _, r := range items {
		out = append(out, r.String())
	}
	return out
}

func fakeIPValues(list []FakeIPEntry) []string {
	out := make([]string, 0, len(list))
	for _, f := range list {
		out = append(out, f.Value)
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (m Manager) vergeDNSOverride() bool {
	dir, err := m.DataDir()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(dir, "verge.yaml"))
	if err != nil {
		return false
	}
	var v struct {
		EnableDNSSettings bool `yaml:"enable_dns_settings"`
	}
	_ = yaml.Unmarshal(data, &v)
	return v.EnableDNSSettings
}

// ReadRules loads the editable state for the current subscription.
func (m Manager) ReadRules() (RulesState, error) {
	ctx, err := m.currentEnhance()
	if err != nil {
		return RulesState{}, err
	}
	ef, err := readEnhanceFiles(ctx)
	if err != nil {
		return RulesState{}, err
	}
	sub := readSubscription(ctx.subPath)
	runtimeDoc, runtime, runtimeGroups := readRuntime(ctx.runtimePath)

	st := RulesState{
		Draft:              ef.draft,
		ProfileUID:         ctx.profile.UID,
		ProfileName:        strutil.OrDefault(ctx.profile.Name, ctx.profile.File),
		RulesFile:          filepath.Base(ctx.rulesPath),
		MergeFile:          filepath.Base(ctx.mergePath),
		RuleTypes:          ruleTypes,
		SubscriptionFakeIP: append([]string{}, sub.fakeIP...),
		FakeIPOverridden:   ef.fakeIPSet,
		DeleteCount:        ef.deleteCount,
		Warnings:           []string{},
		MissingFakeIP:      []string{},
	}
	if ctx.rulesPath == "" {
		st.RulesFile = ""
		st.Warnings = append(st.Warnings, "当前订阅没有关联「规则增强」文件，分流规则暂不能保存。请在 Clash Verge 的订阅卡片上右键 →「编辑规则」打开一次后再回来。")
	}
	if ctx.mergePath == "" {
		st.MergeFile = ""
		st.Warnings = append(st.Warnings, "当前订阅没有关联「覆写 (Merge)」文件，Hosts / DNS 设置暂不能保存。请在 Clash Verge 的订阅卡片上右键 →「编辑覆写」打开一次后再回来。")
	}
	if !ef.fakeIPSet {
		for _, v := range sub.fakeIP {
			st.Draft.FakeIPFilter = append(st.Draft.FakeIPFilter, FakeIPEntry{Value: v})
		}
	} else {
		have := map[string]bool{}
		for _, f := range ef.draft.FakeIPFilter {
			have[f.Value] = true
		}
		for _, v := range sub.fakeIP {
			if !have[v] {
				st.MissingFakeIP = append(st.MissingFakeIP, v)
			}
		}
	}

	if runtimeDoc != nil {
		tun := mapGet(runtimeDoc.Content[0], "tun")
		st.TunEnabled = mapGet(tun, "enable") != nil && mapGet(tun, "enable").Value == "true"
	}
	st.Groups = runtimeGroups
	if len(st.Groups) == 0 {
		st.Groups = sub.groups
	}
	if st.Groups == nil {
		st.Groups = []string{}
	}
	middle := sub.rules
	if len(runtime) > 0 {
		middle = stripCustom(runtime, ruleStrings(ef.draft.Prepend), ruleStrings(ef.draft.Append))
	}
	st.SubscriptionRuleCount = len(middle)
	if n := len(middle); n > 0 {
		t := strings.ToUpper(splitRule(middle[n-1])[0])
		st.SubscriptionEndsMatch = t == "MATCH" || t == "FINAL"
	}
	if m.vergeDNSOverride() {
		st.Warnings = append(st.Warnings, "Clash Verge 已开启「DNS 覆写」，DNS 相关设置可能以 Clash Verge 设置页里的 DNS 配置为准。")
	}
	_, apiErr := m.queryAPI("/version")
	st.CoreReachable = apiErr == nil
	return st, nil
}

// ─── Save ────────────────────────────────────────────────────────────────────

// SaveRules validates the draft, hot-reloads it into the running core and then
// writes the enhancement files (with backups). When the core is not running the
// files are still written and take effect on Clash Verge's next start.
func (m Manager) SaveRules(d RulesDraft) (RulesSaveResult, error) {
	res := RulesSaveResult{Issues: []RuleIssue{}, Warnings: []string{}}
	ctx, err := m.currentEnhance()
	if err != nil {
		return res, err
	}
	old, err := readEnhanceFiles(ctx)
	if err != nil {
		return res, err
	}
	nd, issues := NormalizeDraft(d)
	if len(issues) > 0 {
		res.Issues = issues
		res.Message = fmt.Sprintf("有 %d 处需要修改，未保存", len(issues))
		return res, nil
	}
	sub := readSubscription(ctx.subPath)

	rulesChanged := !sameStrings(ruleStrings(old.draft.Prepend), ruleStrings(nd.Prepend)) ||
		!sameStrings(ruleStrings(old.draft.Append), ruleStrings(nd.Append)) ||
		!sameNotes(old.draft.Prepend, nd.Prepend) || !sameNotes(old.draft.Append, nd.Append)
	if rulesChanged && ctx.rulesPath == "" {
		return res, errors.New("当前订阅没有关联规则增强文件：请先在 Clash Verge 中对该订阅「编辑规则」一次")
	}

	// fake-ip-filter: a list identical to the subscription's means "no override".
	newFake := fakeIPValues(nd.FakeIPFilter)
	writeFake := !(sub.hasFakeIP && sameStrings(newFake, sub.fakeIP)) && len(newFake) > 0
	mergeChanged := !sameHosts(old.draft.Hosts, nd.Hosts) || !samePolicy(old.draft.DNSPolicy, nd.DNSPolicy) ||
		writeFake != old.fakeIPSet || (writeFake && !sameFakeIP(old.draft.FakeIPFilter, nd.FakeIPFilter))
	if mergeChanged && ctx.mergePath == "" {
		return res, errors.New("当前订阅没有关联覆写文件：请先在 Clash Verge 中对该订阅「编辑覆写」一次")
	}
	tunChanged := !sameStrings(tunValues(old.draft.TunExclude), tunValues(nd.TunExclude))
	if !rulesChanged && !mergeChanged && !tunChanged {
		res.Saved = true
		res.Message = "没有需要保存的修改"
		return res, nil
	}

	var rulesBytes, mergeBytes, clashBytes []byte
	if tunChanged {
		tun := mapOf(old.clashDoc.Content[0], "tun")
		if len(nd.TunExclude) == 0 {
			mapDel(tun, "route-exclude-address")
		} else {
			mapSet(tun, "route-exclude-address", seqNode(tunValues(nd.TunExclude), nil))
		}
		if clashBytes, err = encodeYAMLDoc(old.clashDoc); err != nil {
			return res, err
		}
	}
	if rulesChanged {
		applyRulesDoc(old.rulesDoc, nd)
		if rulesBytes, err = encodeYAMLDoc(old.rulesDoc); err != nil {
			return res, err
		}
	}
	if mergeChanged {
		applyMergeDoc(old.mergeDoc, nd, writeFake)
		if mergeBytes, err = encodeYAMLDoc(old.mergeDoc); err != nil {
			return res, err
		}
	}

	// Hot reload: patch the generated runtime and hand it to the core first.
	runtimeDoc, _, _ := readRuntime(ctx.runtimePath)
	var runtimeBytes []byte
	if runtimeDoc != nil {
		patchRuntime(runtimeDoc.Content[0], old, nd, writeFake, sub)
		if res.Applied, runtimeBytes, err = m.applyRuntime(runtimeDoc); err != nil {
			return res, err
		}
	}

	if rulesBytes != nil {
		if err := fsutil.WriteWithBackup(ctx.rulesPath, rulesBytes, 0o644); err != nil {
			return res, fmt.Errorf("写入规则文件失败: %w", err)
		}
	}
	if mergeBytes != nil {
		if err := fsutil.WriteWithBackup(ctx.mergePath, mergeBytes, 0o644); err != nil {
			return res, fmt.Errorf("写入覆写文件失败: %w", err)
		}
	}
	if clashBytes != nil {
		if err := fsutil.WriteWithBackup(ctx.clashPath, clashBytes, 0o644); err != nil {
			return res, fmt.Errorf("写入 config.yaml 失败: %w", err)
		}
	}
	if res.Applied && runtimeBytes != nil {
		// Keep Verge's generated file in step with what the core is now running.
		_ = os.WriteFile(ctx.runtimePath, runtimeBytes, 0o644)
	}
	res.Saved = true
	if res.Applied {
		res.Message = "已保存并立即生效"
	} else {
		res.Message = "已保存。Clash Verge 内核未运行，下次启动 Clash Verge 时生效"
	}
	if tunChanged {
		res.Warnings = append(res.Warnings, "TUN 排除网段写入了 Clash Verge 的 config.yaml；若 Clash Verge 运行中又在其设置里改动 TUN，会以它为准")
	}
	if m.vergeDNSOverride() && mergeChanged {
		res.Warnings = append(res.Warnings, "Clash Verge 开启了「DNS 覆写」，DNS 设置可能被它覆盖")
	}
	return res, nil
}

func tunValues(list []TunEntry) []string {
	out := make([]string, 0, len(list))
	for _, t := range list {
		out = append(out, t.Value)
	}
	return out
}

// applyRuntime hands a patched runtime config to the running core (PUT
// /configs payload, socket pinned to the live one). When the core is not
// running the config is checked with `verge-mihomo -t` instead. It returns
// whether the core now runs it and the bytes to store as clash-verge.yaml.
func (m Manager) applyRuntime(runtimeDoc *yaml.Node) (bool, []byte, error) {
	runtimeBytes, err := encodeYAMLDoc(runtimeDoc)
	if err != nil {
		return false, nil, err
	}
	root := runtimeDoc.Content[0]
	socket := m.unixSocketPath()
	ctl := mapGet(root, "external-controller-unix")
	orig := ""
	if ctl != nil {
		orig, ctl.Value = ctl.Value, socket
	}
	payload, err := encodeYAMLDoc(runtimeDoc)
	if ctl != nil {
		ctl.Value = orig
	}
	if err != nil {
		return false, nil, err
	}
	switch err := reloadCore(socket, payload); {
	case err == nil:
		return true, runtimeBytes, nil
	case errors.Is(err, errCoreUnreachable):
		if err := m.testConfig(runtimeBytes); err != nil {
			return false, nil, err
		}
		return false, nil, nil
	default:
		return false, nil, err
	}
}

func sameNotes(a, b []RuleItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Note != b[i].Note {
			return false
		}
	}
	return true
}

func sameHosts(a, b []HostEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Domain != b[i].Domain || strings.Join(splitList(a[i].Value), ",") != strings.Join(splitList(b[i].Value), ",") || a[i].Note != b[i].Note {
			return false
		}
	}
	return true
}

func samePolicy(a, b []DNSPolicyEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Domain != b[i].Domain || strings.Join(splitList(a[i].Server), ",") != strings.Join(splitList(b[i].Server), ",") || a[i].Note != b[i].Note {
			return false
		}
	}
	return true
}

func sameFakeIP(a, b []FakeIPEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ─── YAML node helpers ───────────────────────────────────────────────────────

// loadYAMLDoc parses path into a document whose first child is a mapping
// (created when the file is missing, empty or comment-only).
func loadYAMLDoc(path string) (*yaml.Node, error) {
	doc := &yaml.Node{Kind: yaml.DocumentNode}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("读取 %s 失败: %w", filepath.Base(path), err)
		}
		if len(bytes.TrimSpace(data)) > 0 {
			var parsed yaml.Node
			if err := yaml.Unmarshal(data, &parsed); err != nil {
				return nil, fmt.Errorf("%s 不是有效的 YAML: %w", filepath.Base(path), err)
			}
			if parsed.Kind == yaml.DocumentNode {
				doc = &parsed
			}
		}
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		if len(doc.Content) > 0 && !(doc.Content[0].Kind == yaml.ScalarNode && doc.Content[0].Tag == "!!null") {
			return nil, fmt.Errorf("%s 顶层不是键值结构", filepath.Base(path))
		}
		doc.Content = []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}
	}
	return doc, nil
}

func encodeYAMLDoc(doc *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	_ = enc.Close()
	return buf.Bytes(), nil
}

func mapGet(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func mapSet(m *yaml.Node, key string, val *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = val
			return
		}
	}
	m.Content = append(m.Content, strNode(key, ""), val)
}

func mapDel(m *yaml.Node, key string) {
	if m == nil {
		return
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			return
		}
	}
}

func mapOf(m *yaml.Node, key string) *yaml.Node {
	if n := mapGet(m, key); n != nil && n.Kind == yaml.MappingNode {
		return n
	}
	n := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	mapSet(m, key, n)
	return n
}

func strNode(v, note string) *yaml.Node {
	n := &yaml.Node{}
	_ = n.Encode(v)
	if note != "" {
		n.LineComment = "# " + note
	}
	return n
}

func seqNode(values []string, notes []string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	if len(values) == 0 {
		n.Style = yaml.FlowStyle
	}
	for i, v := range values {
		note := ""
		if i < len(notes) {
			note = notes[i]
		}
		n.Content = append(n.Content, strNode(v, note))
	}
	return n
}

// valuesNode writes one value as a scalar, several as a flow list.
func valuesNode(list string) *yaml.Node {
	vals := splitList(list)
	if len(vals) == 1 {
		return strNode(vals[0], "")
	}
	n := seqNode(vals, nil)
	n.Style = yaml.FlowStyle
	return n
}

func applyRulesDoc(doc *yaml.Node, d RulesDraft) {
	root := doc.Content[0]
	for _, sec := range []struct {
		key   string
		items []RuleItem
	}{{"prepend", d.Prepend}, {"append", d.Append}} {
		notes := make([]string, len(sec.items))
		for i, r := range sec.items {
			notes[i] = r.Note
		}
		mapSet(root, sec.key, seqNode(ruleStrings(sec.items), notes))
	}
	if mapGet(root, "delete") == nil {
		mapSet(root, "delete", seqNode(nil, nil))
	}
}

func applyMergeDoc(doc *yaml.Node, d RulesDraft, writeFake bool) {
	root := doc.Content[0]
	if len(d.Hosts) == 0 {
		mapDel(root, "hosts")
	} else {
		hosts := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, h := range d.Hosts {
			hosts.Content = append(hosts.Content, strNode(h.Domain, ""), valuesNode(h.Value))
			if h.Note != "" {
				hosts.Content[len(hosts.Content)-1].LineComment = "# " + h.Note
			}
		}
		mapSet(root, "hosts", hosts)
	}

	dns := mapGet(root, "dns")
	if dns == nil && (writeFake || len(d.DNSPolicy) > 0) {
		dns = mapOf(root, "dns")
	}
	if dns == nil {
		return
	}
	if writeFake {
		notes := make([]string, len(d.FakeIPFilter))
		for i, f := range d.FakeIPFilter {
			notes[i] = f.Note
		}
		mapSet(dns, "fake-ip-filter", seqNode(fakeIPValues(d.FakeIPFilter), notes))
	} else {
		mapDel(dns, "fake-ip-filter")
	}
	if len(d.DNSPolicy) == 0 {
		mapDel(dns, "nameserver-policy")
	} else {
		pol := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, p := range d.DNSPolicy {
			pol.Content = append(pol.Content, strNode(p.Domain, ""), valuesNode(p.Server))
			if p.Note != "" {
				pol.Content[len(pol.Content)-1].LineComment = "# " + p.Note
			}
		}
		mapSet(dns, "nameserver-policy", pol)
	}
	if len(dns.Content) == 0 {
		mapDel(root, "dns")
	}
}

// patchRuntime rewrites the generated runtime config the way Clash Verge would
// after the enhancement files change: swap the custom rule blocks, re-merge
// hosts / nameserver-policy keys and replace (or restore) fake-ip-filter.
func patchRuntime(root *yaml.Node, old enhanceFiles, nd RulesDraft, writeFake bool, sub subscriptionInfo) {
	if tun := mapGet(root, "tun"); tun != nil && tun.Kind == yaml.MappingNode {
		if len(nd.TunExclude) == 0 {
			mapDel(tun, "route-exclude-address")
		} else {
			mapSet(tun, "route-exclude-address", seqNode(tunValues(nd.TunExclude), nil))
		}
	}
	runtimeRules := scalarList(mapGet(root, "rules"))
	middle := stripCustom(runtimeRules, ruleStrings(old.draft.Prepend), ruleStrings(old.draft.Append))
	merged := append(append(append([]string{}, ruleStrings(nd.Prepend)...), middle...), ruleStrings(nd.Append)...)
	mapSet(root, "rules", seqNode(merged, nil))

	hosts := mapGet(root, "hosts")
	if hosts == nil || hosts.Kind != yaml.MappingNode {
		hosts = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	}
	for _, h := range old.draft.Hosts {
		mapDel(hosts, h.Domain)
	}
	for _, h := range nd.Hosts {
		mapSet(hosts, h.Domain, valuesNode(h.Value))
	}
	if len(hosts.Content) > 0 {
		mapSet(root, "hosts", hosts)
	} else {
		mapDel(root, "hosts")
	}

	dns := mapGet(root, "dns")
	if dns == nil || dns.Kind != yaml.MappingNode {
		if !writeFake && len(nd.DNSPolicy) == 0 {
			return
		}
		dns = mapOf(root, "dns")
	}
	switch {
	case writeFake:
		mapSet(dns, "fake-ip-filter", seqNode(fakeIPValues(nd.FakeIPFilter), nil))
	case sub.hasFakeIP:
		mapSet(dns, "fake-ip-filter", seqNode(sub.fakeIP, nil))
	default:
		mapDel(dns, "fake-ip-filter")
	}
	pol := mapGet(dns, "nameserver-policy")
	if pol == nil || pol.Kind != yaml.MappingNode {
		pol = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	}
	for _, p := range old.draft.DNSPolicy {
		mapDel(pol, p.Domain)
	}
	for _, p := range nd.DNSPolicy {
		mapSet(pol, p.Domain, valuesNode(p.Server))
	}
	if len(pol.Content) > 0 {
		mapSet(dns, "nameserver-policy", pol)
	} else {
		mapDel(dns, "nameserver-policy")
	}
}

// ─── Core hot reload ─────────────────────────────────────────────────────────

var errCoreUnreachable = errors.New("clash core unreachable")

// reloadCore replaces the running config via PUT /configs (payload mode). The
// core parses the payload first and keeps the old config when it is invalid.
func reloadCore(socketPath string, payload []byte) error {
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, "unix", socketPath)
			},
		},
	}
	body, _ := json.Marshal(map[string]string{"path": "", "payload": string(payload)})
	req, err := http.NewRequest(http.MethodPut, "http://localhost/configs?force=true", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) && opErr.Op == "dial" {
			return errCoreUnreachable
		}
		return fmt.Errorf("通知 Clash 内核重载失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		return nil
	}
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(msg, &parsed) == nil && parsed.Message != "" {
		msg = []byte(parsed.Message)
	}
	return fmt.Errorf("Clash 内核拒绝了新配置，未保存：%s", strings.TrimSpace(string(msg)))
}
