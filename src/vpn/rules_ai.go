package vpn

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"hatch/core/strutil"
)

// ─── Natural-language rule edits (自定义规则 → AI) ───────────────────────────
//
// The AI answers with a structured change set (add / remove per section plus
// test cases). VerifyRuleSuggestion turns it into a new draft and only accepts
// it when it is effective: something actually changes, every entry is valid,
// each test case routes as the AI claims, and the core accepts the result.

type RuleSuggestion struct {
	Summary string `json:"summary"`
	Add     struct {
		Rules  []struct{ Rule, Note string } `json:"rules"`
		Hosts  []HostEntry                   `json:"hosts"`
		FakeIP []FakeIPEntry                 `json:"fakeIp"`
		DNS    []DNSPolicyEntry              `json:"dns"`
		Tun    []string                      `json:"tun"`
	} `json:"add"`
	Remove struct {
		Rules  []string `json:"rules"`
		Hosts  []string `json:"hosts"`
		FakeIP []string `json:"fakeIp"`
		DNS    []string `json:"dns"`
		Tun    []string `json:"tun"`
	} `json:"remove"`
	Tests []struct{ Input, Expect string } `json:"tests"`
}

type RuleChange struct {
	Section string `json:"section"` // rules | hosts | fakeip | dns | tun
	Op      string `json:"op"`      // add | remove
	Text    string `json:"text"`
}

type RuleTestResult struct {
	Input  string `json:"input"`
	Expect string `json:"expect"`
	Actual string `json:"actual"`
	Rule   string `json:"rule"`
	Pass   bool   `json:"pass"`
}

type RuleAIPreview struct {
	Summary string           `json:"summary"`
	Draft   RulesDraft       `json:"draft"`
	Changes []RuleChange     `json:"changes"`
	Tests   []RuleTestResult `json:"tests"`
}

// ParseRuleSuggestion reads the AI's JSON answer.
func ParseRuleSuggestion(raw string) (RuleSuggestion, error) {
	raw = strings.TrimSpace(stripCodeFence(raw))
	if i, j := strings.Index(raw, "{"), strings.LastIndex(raw, "}"); i >= 0 && j > i {
		raw = raw[i : j+1]
	}
	var s RuleSuggestion
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return s, fmt.Errorf("AI 返回的内容不是有效 JSON：%w", err)
	}
	return s, nil
}

// VerifyRuleSuggestion applies a suggestion to the draft and proves it works.
func (m Manager) VerifyRuleSuggestion(d RulesDraft, s RuleSuggestion) (RuleAIPreview, error) {
	p := RuleAIPreview{Summary: strings.TrimSpace(s.Summary), Changes: []RuleChange{}, Tests: []RuleTestResult{}}
	base, _ := NormalizeDraft(d)
	nd := cloneDraft(base)

	// removals (matched loosely: canonical rule text / domain / value)
	if len(s.Remove.Rules) > 0 {
		drop := map[string]bool{}
		for _, r := range s.Remove.Rules {
			if n, err := normalizeRule(ParseRule(r)); err == nil {
				drop[n.String()] = true
			}
			drop[canonicalRule(r)] = true
		}
		filter := func(list []RuleItem) []RuleItem {
			out := []RuleItem{}
			for _, r := range list {
				if drop[r.String()] {
					p.Changes = append(p.Changes, RuleChange{"rules", "remove", r.String()})
					continue
				}
				out = append(out, r)
			}
			return out
		}
		nd.Prepend, nd.Append = filter(nd.Prepend), filter(nd.Append)
	}
	removeBy := func(section string, keys []string, n int, key func(int) string, del func(int)) {
		set := map[string]bool{}
		for _, k := range keys {
			set[strings.ToLower(strings.TrimSpace(k))] = true
		}
		for i := n - 1; i >= 0; i-- {
			if set[strings.ToLower(key(i))] {
				p.Changes = append(p.Changes, RuleChange{section, "remove", key(i)})
				del(i)
			}
		}
	}
	removeBy("hosts", s.Remove.Hosts, len(nd.Hosts), func(i int) string { return nd.Hosts[i].Domain },
		func(i int) { nd.Hosts = append(nd.Hosts[:i], nd.Hosts[i+1:]...) })
	removeBy("fakeip", s.Remove.FakeIP, len(nd.FakeIPFilter), func(i int) string { return nd.FakeIPFilter[i].Value },
		func(i int) { nd.FakeIPFilter = append(nd.FakeIPFilter[:i], nd.FakeIPFilter[i+1:]...) })
	removeBy("dns", s.Remove.DNS, len(nd.DNSPolicy), func(i int) string { return nd.DNSPolicy[i].Domain },
		func(i int) { nd.DNSPolicy = append(nd.DNSPolicy[:i], nd.DNSPolicy[i+1:]...) })
	var tunDrop []string
	for _, v := range s.Remove.Tun {
		if c, _, err := normalizeCIDR(v); err == nil {
			tunDrop = append(tunDrop, c)
		}
	}
	removeBy("tun", tunDrop, len(nd.TunExclude), func(i int) string { return nd.TunExclude[i].Value },
		func(i int) { nd.TunExclude = append(nd.TunExclude[:i], nd.TunExclude[i+1:]...) })

	// additions — every entry must be valid, otherwise the whole answer is rejected
	var problems []string
	have := map[string]bool{}
	for _, r := range append(append([]RuleItem{}, nd.Prepend...), nd.Append...) {
		have[r.String()] = true
	}
	var newRules []RuleItem
	for _, a := range s.Add.Rules {
		r := ParseRule(a.Rule)
		r.Note = a.Note
		n, err := normalizeRule(r)
		if err != nil {
			problems = append(problems, fmt.Sprintf("规则「%s」：%v", a.Rule, err))
			continue
		}
		if !have[n.String()] {
			have[n.String()] = true
			newRules = append(newRules, n)
			p.Changes = append(p.Changes, RuleChange{"rules", "add", n.String()})
		}
	}
	// new rules go first so they win over older, broader custom rules
	nd.Prepend = append(newRules, nd.Prepend...)
	for _, h := range s.Add.Hosts {
		n, err := normalizeHostEntry(h)
		if err != nil {
			problems = append(problems, fmt.Sprintf("Hosts「%s」：%v", h.Domain, err))
			continue
		}
		replaced := false
		for i := range nd.Hosts {
			if nd.Hosts[i].Domain == n.Domain {
				if nd.Hosts[i].Value != n.Value {
					nd.Hosts[i] = n
					p.Changes = append(p.Changes, RuleChange{"hosts", "add", n.Domain + " → " + n.Value})
				}
				replaced = true
			}
		}
		if !replaced {
			nd.Hosts = append(nd.Hosts, n)
			p.Changes = append(p.Changes, RuleChange{"hosts", "add", n.Domain + " → " + n.Value})
		}
	}
	for _, f := range s.Add.FakeIP {
		n, err := normalizeFakeIP(f)
		if err != nil {
			problems = append(problems, fmt.Sprintf("Fake-IP「%s」：%v", f.Value, err))
			continue
		}
		dup := false
		for _, x := range nd.FakeIPFilter {
			dup = dup || x.Value == n.Value
		}
		if !dup {
			nd.FakeIPFilter = append(nd.FakeIPFilter, n)
			p.Changes = append(p.Changes, RuleChange{"fakeip", "add", n.Value})
		}
	}
	for _, x := range s.Add.DNS {
		n, err := normalizeDNSPolicy(x)
		if err != nil {
			problems = append(problems, fmt.Sprintf("DNS「%s」：%v", x.Domain, err))
			continue
		}
		replaced := false
		for i := range nd.DNSPolicy {
			if nd.DNSPolicy[i].Domain == n.Domain {
				if nd.DNSPolicy[i].Server != n.Server {
					nd.DNSPolicy[i] = n
					p.Changes = append(p.Changes, RuleChange{"dns", "add", n.Domain + " → " + n.Server})
				}
				replaced = true
			}
		}
		if !replaced {
			nd.DNSPolicy = append(nd.DNSPolicy, n)
			p.Changes = append(p.Changes, RuleChange{"dns", "add", n.Domain + " → " + n.Server})
		}
	}
	for _, v := range s.Add.Tun {
		c, _, err := normalizeCIDR(v)
		if err != nil {
			problems = append(problems, fmt.Sprintf("TUN 排除「%s」：%v", v, err))
			continue
		}
		dup := false
		for _, x := range nd.TunExclude {
			dup = dup || x.Value == c
		}
		if !dup {
			nd.TunExclude = append(nd.TunExclude, TunEntry{Value: c})
			p.Changes = append(p.Changes, RuleChange{"tun", "add", c})
		}
	}
	if len(problems) > 0 {
		return p, errors.New("AI 给出的内容有误：" + strings.Join(problems, "；"))
	}
	if len(p.Changes) == 0 {
		return p, errors.New("AI 没有给出任何实际修改（可能要求的内容已经存在），请换个说法或说得更具体")
	}
	if _, issues := NormalizeDraft(nd); len(issues) > 0 {
		return p, fmt.Errorf("修改后的规则有误：%s", issues[0].Message)
	}

	// prove routing intent with the AI's own test cases
	ctx, err := m.currentEnhance()
	if err != nil {
		return p, err
	}
	old, err := readEnhanceFiles(ctx)
	if err != nil {
		return p, err
	}
	sub := readSubscription(ctx.subPath)
	runtimeDoc, runtime, _ := readRuntime(ctx.runtimePath)
	middle := sub.rules
	if len(runtime) > 0 {
		middle = stripCustom(runtime, ruleStrings(old.draft.Prepend), ruleStrings(old.draft.Append))
	}
	var failed []string
	for _, t := range s.Tests {
		if strings.TrimSpace(t.Input) == "" {
			continue
		}
		r := matchRules(t.Input, nd, middle)
		res := RuleTestResult{Input: t.Input, Expect: t.Expect, Actual: r.Target, Rule: r.Rule}
		res.Pass = testExpectMet(t.Expect, r, nd)
		if !res.Pass {
			failed = append(failed, fmt.Sprintf("%s 期望 %s，实际 %s（%s）", t.Input, t.Expect, strutil.OrDefault(r.Target, "未命中"), r.Rule))
		}
		p.Tests = append(p.Tests, res)
	}
	if len(failed) > 0 {
		return p, errors.New("验证没通过：" + strings.Join(failed, "；"))
	}

	// the core must accept the resulting configuration
	if runtimeDoc != nil {
		newFake := fakeIPValues(nd.FakeIPFilter)
		writeFake := !(sub.hasFakeIP && sameStrings(newFake, sub.fakeIP)) && len(newFake) > 0
		patchRuntime(runtimeDoc.Content[0], old, nd, writeFake, sub)
		data, err := encodeYAMLDoc(runtimeDoc)
		if err != nil {
			return p, err
		}
		if err := m.testConfig(data); err != nil {
			return p, err
		}
	}
	p.Draft = nd
	return p, nil
}

// testExpectMet: expect is a target (DIRECT / REJECT / group), "BYPASS" for TUN
// exclusion, or an IP the domain should resolve to via hosts.
func testExpectMet(expect string, r RuleMatch, d RulesDraft) bool {
	e := strings.TrimSpace(expect)
	switch {
	case e == "":
		return r.Matched
	case strings.EqualFold(e, "BYPASS"):
		ip := net.ParseIP(r.IP)
		for _, t := range d.TunExclude {
			if _, n, err := net.ParseCIDR(t.Value); err == nil && ip != nil && n.Contains(ip) {
				return true
			}
		}
		return false
	case net.ParseIP(e) != nil:
		return r.IP == e
	default:
		return strings.EqualFold(r.Target, e)
	}
}

func cloneDraft(d RulesDraft) RulesDraft {
	data, _ := json.Marshal(d)
	var out RulesDraft
	_ = json.Unmarshal(data, &out)
	return out
}
