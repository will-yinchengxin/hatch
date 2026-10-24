package vpn

import (
	"net"
	"regexp"
	"strconv"
	"strings"
)

// ─── Rule tester (测试网址) ──────────────────────────────────────────────────
//
// TestRule walks the would-be rule list (draft prepend + subscription rules +
// draft append) the way the core does and reports the first rule that surely
// matches. Rules that need data Hatch does not have locally (GEOSITE, process,
// DNS lookups…) are listed as "uncertain" instead of being guessed.

type RuleMatch struct {
	Input          string   `json:"input"`
	Host           string   `json:"host"`
	IP             string   `json:"ip"`
	Port           string   `json:"port"`
	HostsValue     string   `json:"hostsValue"`
	FakeIPBypass   string   `json:"fakeIpBypass"` // matching fake-ip-filter entry
	DNSServer      string   `json:"dnsServer"`    // matching nameserver-policy servers
	Matched        bool     `json:"matched"`
	Rule           string   `json:"rule"`
	Target         string   `json:"target"`
	Source         string   `json:"source"`   // custom | subscription | append
	Position       int      `json:"position"` // 1-based within Source
	Uncertain      []string `json:"uncertain"`
	UncertainCount int      `json:"uncertainCount"`
}

// TestRule tests input ("chat.openai.com", "https://a.com/x", "1.2.3.4:22")
// against the draft (unsaved edits included).
func (m Manager) TestRule(input string, d RulesDraft) (RuleMatch, error) {
	ctx, err := m.currentEnhance()
	if err != nil {
		return RuleMatch{}, err
	}
	old, err := readEnhanceFiles(ctx)
	if err != nil {
		return RuleMatch{}, err
	}
	sub := readSubscription(ctx.subPath)
	_, runtime, _ := readRuntime(ctx.runtimePath)
	middle := sub.rules
	if len(runtime) > 0 {
		middle = stripCustom(runtime, ruleStrings(old.draft.Prepend), ruleStrings(old.draft.Append))
	}
	nd, _ := NormalizeDraft(d)
	return matchRules(input, nd, middle), nil
}

func parseTestInput(input string) (host, port string) {
	s := strings.TrimSpace(input)
	scheme := ""
	if i := strings.Index(s, "://"); i >= 0 {
		scheme, s = strings.ToLower(s[:i]), s[i+3:]
	}
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	if i := strings.LastIndex(s, "@"); i >= 0 {
		s = s[i+1:]
	}
	if h, p, err := net.SplitHostPort(s); err == nil {
		s, port = h, p
	}
	if port == "" {
		switch scheme {
		case "https", "wss":
			port = "443"
		case "http", "ws":
			port = "80"
		}
	}
	return strings.ToLower(strings.Trim(strings.Trim(s, "[]"), ".")), port
}

func matchRules(input string, d RulesDraft, middle []string) RuleMatch {
	res := RuleMatch{Input: input, Uncertain: []string{}}
	res.Host, res.Port = parseTestInput(input)
	var ip net.IP
	if parsed := net.ParseIP(res.Host); parsed != nil {
		ip = parsed
		res.IP = parsed.String()
	} else {
		for _, h := range d.Hosts {
			if matchDomainPattern(h.Domain, res.Host) {
				res.HostsValue = h.Value
				if first := net.ParseIP(firstOf(splitList(h.Value))); first != nil {
					ip = first
					res.IP = first.String()
				}
				break
			}
		}
		for _, f := range d.FakeIPFilter {
			if matchDomainPattern(f.Value, res.Host) {
				res.FakeIPBypass = f.Value
				break
			}
		}
		for _, p := range d.DNSPolicy {
			if matchDomainPattern(p.Domain, res.Host) {
				res.DNSServer = p.Server
				break
			}
		}
	}

	type sourced struct {
		rule   string
		source string
		pos    int
	}
	var all []sourced
	for i, r := range ruleStrings(d.Prepend) {
		all = append(all, sourced{r, "custom", i + 1})
	}
	for i, r := range middle {
		all = append(all, sourced{r, "subscription", i + 1})
	}
	for i, r := range ruleStrings(d.Append) {
		all = append(all, sourced{r, "append", i + 1})
	}

	for _, r := range all {
		hit, certain := evalRule(r.rule, res.Host, ip, res.Port)
		if !certain {
			res.UncertainCount++
			if len(res.Uncertain) < 5 {
				res.Uncertain = append(res.Uncertain, r.rule)
			}
			continue
		}
		if hit {
			res.Matched = true
			res.Rule = r.rule
			res.Source = r.source
			res.Position = r.pos
			parts := splitRule(r.rule)
			res.Target = parts[len(parts)-1]
			if strings.EqualFold(res.Target, "no-resolve") && len(parts) >= 2 {
				res.Target = parts[len(parts)-2]
			}
			break
		}
	}
	return res
}

func firstOf(list []string) string {
	if len(list) == 0 {
		return ""
	}
	return list[0]
}

// evalRule reports (matched, certain). certain=false means Hatch cannot decide
// locally (needs GEOSITE data, DNS lookup, process info, …).
func evalRule(rule, host string, ip net.IP, port string) (bool, bool) {
	parts := splitRule(rule)
	t := strings.ToUpper(parts[0])
	if t == "MATCH" || t == "FINAL" {
		return true, true
	}
	if len(parts) < 3 {
		return false, false
	}
	v := parts[1]
	noResolve := strings.EqualFold(parts[len(parts)-1], "no-resolve")
	isIPInput := net.ParseIP(host) != nil

	switch t {
	case "DOMAIN":
		return !isIPInput && host == strings.ToLower(v), true
	case "DOMAIN-SUFFIX":
		v = strings.ToLower(v)
		return !isIPInput && (host == v || strings.HasSuffix(host, "."+v)), true
	case "DOMAIN-KEYWORD":
		return !isIPInput && strings.Contains(host, strings.ToLower(v)), true
	case "DOMAIN-REGEX":
		re, err := regexp.Compile(v)
		if err != nil {
			return false, true
		}
		return !isIPInput && re.MatchString(host), true
	case "DOMAIN-WILDCARD":
		return !isIPInput && matchDomainPattern(v, host), true
	case "IP-CIDR", "IP-CIDR6", "IP-SUFFIX":
		if ip == nil {
			return false, noResolve
		}
		_, n, err := net.ParseCIDR(v)
		if err != nil {
			return false, true
		}
		return n.Contains(ip), true
	case "GEOIP":
		code := strings.ToUpper(v)
		if ip == nil {
			return false, noResolve
		}
		if code == "LAN" || code == "PRIVATE" {
			return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast(), true
		}
		if ip.IsPrivate() || ip.IsLoopback() {
			return false, true
		}
		return false, false
	case "DST-PORT":
		if port == "" {
			return false, false
		}
		return portMatches(v, port), true
	}
	return false, false
}

func portMatches(spec, port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	for _, part := range strings.Split(spec, "/") {
		lo, hi, isRange := strings.Cut(part, "-")
		a, err1 := strconv.Atoi(lo)
		b := a
		var err2 error
		if isRange {
			b, err2 = strconv.Atoi(hi)
		}
		if err1 == nil && err2 == nil && p >= a && p <= b {
			return true
		}
	}
	return false
}

// matchDomainPattern implements Clash's domain-set syntax used by hosts,
// fake-ip-filter and nameserver-policy: "+.a.com" (a.com and all subdomains),
// ".a.com" (subdomains only), "*" (exactly one label), plain (exact).
func matchDomainPattern(pattern, host string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" || host == "" {
		return false
	}
	hostLabels := strings.Split(host, ".")
	switch {
	case strings.HasPrefix(pattern, "+."):
		return labelsSuffix(strings.Split(pattern[2:], "."), hostLabels, 0)
	case strings.HasPrefix(pattern, "."):
		return labelsSuffix(strings.Split(pattern[1:], "."), hostLabels, 1)
	}
	pat := strings.Split(pattern, ".")
	return len(pat) == len(hostLabels) && labelsSuffix(pat, hostLabels, 0)
}

// labelsSuffix: host ends with pat (per-label, "*" wildcard) and has at least
// minExtra more labels than pat.
func labelsSuffix(pat, host []string, minExtra int) bool {
	if len(host) < len(pat)+minExtra {
		return false
	}
	off := len(host) - len(pat)
	for i, p := range pat {
		if p != "*" && p != host[off+i] {
			return false
		}
	}
	return true
}
