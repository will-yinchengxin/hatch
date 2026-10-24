package vpn

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseRuleRoundTrip(t *testing.T) {
	cases := map[string]RuleItem{
		"DOMAIN-SUFFIX,a.com,DIRECT":           {Type: "DOMAIN-SUFFIX", Value: "a.com", Target: "DIRECT"},
		"IP-CIDR,1.2.3.4/32,DIRECT,no-resolve": {Type: "IP-CIDR", Value: "1.2.3.4/32", Target: "DIRECT", NoResolve: true},
		" DOMAIN , b.com , Proxy ":             {Type: "DOMAIN", Value: "b.com", Target: "Proxy"},
	}
	for in, want := range cases {
		got := ParseRule(in)
		if got != want {
			t.Fatalf("ParseRule(%q) = %+v, want %+v", in, got, want)
		}
	}
	for _, raw := range []string{"AND,((DOMAIN,a.com),(NETWORK,UDP)),REJECT", "MATCH,Proxy", "DOMAIN,a.com,DIRECT,extra"} {
		got := ParseRule(raw)
		if got.Raw == "" || got.String() != canonicalRule(raw) {
			t.Fatalf("complex rule %q should stay raw, got %+v", raw, got)
		}
	}
	if ParseRule("IP-CIDR,1.2.3.4/32,DIRECT,no-resolve").String() != "IP-CIDR,1.2.3.4/32,DIRECT,no-resolve" {
		t.Fatal("no-resolve lost on format")
	}
}

func TestNormalizeRule(t *testing.T) {
	ok := []struct {
		in   RuleItem
		want string
	}{
		{RuleItem{Type: "DOMAIN-SUFFIX", Value: "https://www.Example.com:8443/path?q=1", Target: "DIRECT"}, "DOMAIN-SUFFIX,www.example.com,DIRECT"},
		{RuleItem{Type: "DOMAIN-SUFFIX", Value: "*.example.com", Target: "DIRECT"}, "DOMAIN-SUFFIX,example.com,DIRECT"},
		{RuleItem{Type: "IP-CIDR", Value: "10.1.2.3", Target: "DIRECT", NoResolve: true}, "IP-CIDR,10.1.2.3/32,DIRECT,no-resolve"},
		{RuleItem{Type: "IP-CIDR", Value: "2001:db8::1", Target: "DIRECT"}, "IP-CIDR6,2001:db8::1/128,DIRECT"},
		{RuleItem{Type: "DOMAIN", Value: "172.16.0.9", Target: "DIRECT"}, "IP-CIDR,172.16.0.9/32,DIRECT,no-resolve"},
		{RuleItem{Type: "GEOIP", Value: "cn", Target: "DIRECT", NoResolve: true}, "GEOIP,CN,DIRECT,no-resolve"},
		{RuleItem{Type: "DOMAIN-KEYWORD", Value: "Google", Target: "P", NoResolve: true}, "DOMAIN-KEYWORD,google,P"},
		{RuleItem{Raw: "AND,((DOMAIN,a.com),(NETWORK,UDP)),REJECT"}, "AND,((DOMAIN,a.com),(NETWORK,UDP)),REJECT"},
		{RuleItem{Raw: "domain-suffix, x.com ,DIRECT"}, "DOMAIN-SUFFIX,x.com,DIRECT"},
	}
	for _, c := range ok {
		got, err := normalizeRule(c.in)
		if err != nil || got.String() != c.want {
			t.Fatalf("normalizeRule(%+v) = %q, %v; want %q", c.in, got.String(), err, c.want)
		}
	}
	bad := []RuleItem{
		{Type: "DOMAIN-SUFFIX", Value: "", Target: "DIRECT"},
		{Type: "DOMAIN-SUFFIX", Value: "a.com", Target: ""},
		{Type: "IP-CIDR", Value: "1.2.3", Target: "DIRECT"},
		{Type: "DST-PORT", Value: "abc", Target: "DIRECT"},
		{Type: "NOPE", Value: "x", Target: "DIRECT"},
		{Type: "DOMAIN", Value: "a b.com", Target: "DIRECT"},
		{Raw: "justone"},
	}
	for _, r := range bad {
		if _, err := normalizeRule(r); err == nil {
			t.Fatalf("expected error for %+v", r)
		}
	}
}

func TestParseRuleLines(t *testing.T) {
	got := ParseRuleLines("# comment\nhttps://foo.com/x\n10.0.0.1\n- DOMAIN,bar.com,REJECT # note\na.com, b.com\n", "Proxy")
	var strs []string
	for _, r := range got {
		strs = append(strs, r.String())
	}
	want := []string{"DOMAIN-SUFFIX,foo.com,Proxy", "IP-CIDR,10.0.0.1/32,Proxy,no-resolve", "DOMAIN,bar.com,REJECT"}
	if strings.Join(strs[:3], "|") != strings.Join(want, "|") {
		t.Fatalf("got %v", strs)
	}
	// "a.com, b.com" contains a comma, so it is parsed as a (broken) raw rule, surfaced by validation.
	if len(strs) != 4 {
		t.Fatalf("got %d rules: %v", len(strs), strs)
	}
}

func TestMatchDomainPattern(t *testing.T) {
	cases := []struct {
		pat, host string
		want      bool
	}{
		{"+.a.com", "a.com", true}, {"+.a.com", "x.y.a.com", true}, {"+.a.com", "ba.com", false},
		{"*.a.com", "x.a.com", true}, {"*.a.com", "x.y.a.com", false}, {"*.a.com", "a.com", false},
		{".a.com", "x.y.a.com", true}, {".a.com", "a.com", false},
		{"time.*.com", "time.apple.com", true}, {"+.stun.*.*", "stun.l.google.com", false}, {"+.stun.*.*", "stun.qq.com", true},
		{"a.com", "a.com", true}, {"a.com", "b.a.com", false},
	}
	for _, c := range cases {
		if got := matchDomainPattern(c.pat, c.host); got != c.want {
			t.Fatalf("matchDomainPattern(%q,%q) = %v", c.pat, c.host, got)
		}
	}
}

func TestMatchRules(t *testing.T) {
	d := RulesDraft{
		Prepend:      []RuleItem{{Type: "DOMAIN-SUFFIX", Value: "corp.com", Target: "DIRECT"}, {Type: "IP-CIDR", Value: "172.16.0.0/12", Target: "DIRECT", NoResolve: true}},
		Hosts:        []HostEntry{{Domain: "*.lab.io", Value: "172.16.1.5"}},
		FakeIPFilter: []FakeIPEntry{{Value: "+.corp.com"}},
		DNSPolicy:    []DNSPolicyEntry{{Domain: "+.corp.com", Server: "10.0.0.53"}},
	}
	middle := []string{"GEOSITE,openai,Proxy", "DOMAIN-KEYWORD,google,Proxy", "GEOIP,CN,DIRECT", "MATCH,Proxy"}

	r := matchRules("https://git.corp.com/x", d, middle)
	if !r.Matched || r.Source != "custom" || r.Position != 1 || r.Target != "DIRECT" || r.FakeIPBypass != "+.corp.com" || r.DNSServer != "10.0.0.53" || r.Port != "443" {
		t.Fatalf("corp: %+v", r)
	}
	r = matchRules("web.lab.io", d, middle)
	if !r.Matched || r.Rule != "IP-CIDR,172.16.0.0/12,DIRECT,no-resolve" || r.IP != "172.16.1.5" {
		t.Fatalf("hosts → ip rule: %+v", r)
	}
	r = matchRules("www.google.com", d, middle)
	if r.Source != "subscription" || r.Target != "Proxy" || r.UncertainCount != 1 || r.Uncertain[0] != "GEOSITE,openai,Proxy" {
		t.Fatalf("google: %+v", r)
	}
	r = matchRules("example.org", d, middle)
	if r.Rule != "MATCH,Proxy" || r.UncertainCount != 2 {
		t.Fatalf("fallback: %+v", r)
	}
}

func TestExpandTemplate(t *testing.T) {
	d, err := ExpandTemplate("intranet", map[string]string{"domains": "https://corp.example.com/login\nlab.example.com", "ip": "172.16.27.50"})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Prepend) != 3 || d.Prepend[2].String() != "IP-CIDR,172.16.27.50/32,DIRECT,no-resolve" {
		t.Fatalf("rules: %+v", d.Prepend)
	}
	if len(d.Hosts) != 2 || d.Hosts[0].Domain != "corp.example.com" || len(d.FakeIPFilter) != 2 || d.FakeIPFilter[1].Value != "+.lab.example.com" {
		t.Fatalf("hosts/fakeip: %+v %+v", d.Hosts, d.FakeIPFilter)
	}
	if len(d.DNSPolicy) != 0 {
		t.Fatalf("optional dns should be skipped: %+v", d.DNSPolicy)
	}
	if _, err := ExpandTemplate("site-proxy", map[string]string{"domains": "a.com"}); err == nil {
		t.Fatal("missing target should fail")
	}
	for _, tpl := range RuleTemplates() {
		in := map[string]string{"domains": "a.com", "apps": "App", "target": "Proxy", "ip": "10.0.0.1", "dns": "10.0.0.53", "cidr": "10.0.0.0/8"}
		if _, err := ExpandTemplate(tpl.ID, in); err != nil {
			t.Fatalf("template %s: %v", tpl.ID, err)
		}
	}
}

// ─── End-to-end save against a fake Clash Verge dir ──────────────────────────

const testRulesFile = `# Profile Enhancement Rules Template for Clash Verge

prepend:
  - DOMAIN,old.com,DIRECT

append: []

delete:
  - DOMAIN-KEYWORD,ads,REJECT
`

const testMergeFile = `# Profile Enhancement Merge Template for Clash Verge

profile:
  store-selected: true
hosts:
  old.com: 1.1.1.1
`

const testSubFile = `proxies:
  - {name: 美国 01, type: ss}
  - {name: 美国 02, type: ss}
  - {name: 香港 01, type: ss}
proxy-groups:
  - {name: Proxy, type: select, proxies: [美国 01, 美国 02, 香港 01]}
dns:
  enhanced-mode: fake-ip
  fake-ip-filter:
    - '*.lan'
    - '+.stun.*.*'
rules:
  - GEOIP,CN,DIRECT
  - MATCH,Proxy
`

func setupClashDir(t *testing.T, socket string) (Manager, string) {
	t.Helper()
	dir := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(dir, "profiles"), 0o755))
	must(os.WriteFile(filepath.Join(dir, "profiles.yaml"), []byte(`current: SUB
items:
- {uid: RULES, type: rules, file: rules.yaml}
- {uid: MERGE, type: merge, file: merge.yaml}
- {uid: Merge, type: merge, file: GlobalMerge.yaml}
- uid: SUB
  type: remote
  name: 测试订阅
  file: sub.yaml
  url: https://sub.example.com/api?token=secret
  extra: {upload: 1, download: 2, total: 10, expire: 1900000000}
  option: {update_interval: 1440, merge: MERGE, rules: RULES}
`), 0o644))
	must(os.WriteFile(filepath.Join(dir, "profiles", "rules.yaml"), []byte(testRulesFile), 0o644))
	must(os.WriteFile(filepath.Join(dir, "profiles", "merge.yaml"), []byte(testMergeFile), 0o644))
	must(os.WriteFile(filepath.Join(dir, "profiles", "GlobalMerge.yaml"), []byte("profile:\n  store-selected: true\n"), 0o644))
	must(os.WriteFile(filepath.Join(dir, "profiles", "sub.yaml"), []byte(testSubFile), 0o644))
	runtime := `# Generated by Clash Verge
external-controller-unix: ` + socket + `
hosts:
  old.com: 1.1.1.1
dns:
  enhanced-mode: fake-ip
  fake-ip-filter: ['*.lan', '+.stun.*.*']
proxy-groups:
  - {name: Proxy, type: select, proxies: [DIRECT]}
rules:
  - DOMAIN,old.com,DIRECT
  - GEOIP,CN,DIRECT
  - MATCH,Proxy
`
	must(os.WriteFile(filepath.Join(dir, "clash-verge.yaml"), []byte(runtime), 0o644))
	return Manager{AppDir: dir, sockets: []string{socket}}, dir
}

func TestReadRules(t *testing.T) {
	m, _ := setupClashDir(t, "/nonexistent/hatch-test.sock")
	st, err := m.ReadRules()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Draft.Prepend) != 1 || st.Draft.Prepend[0].Value != "old.com" || st.DeleteCount != 1 {
		t.Fatalf("prepend: %+v", st)
	}
	if st.FakeIPOverridden || len(st.Draft.FakeIPFilter) != 2 || st.SubscriptionRuleCount != 2 || !st.SubscriptionEndsMatch {
		t.Fatalf("fake-ip/sub: %+v", st)
	}
	if len(st.Groups) != 1 || st.Groups[0] != "Proxy" || st.ProfileName != "测试订阅" || st.CoreReachable {
		t.Fatalf("meta: %+v", st)
	}
}

func TestSaveRulesOffline(t *testing.T) {
	m, dir := setupClashDir(t, "/nonexistent/hatch-test.sock")
	st, _ := m.ReadRules()
	d := st.Draft
	d.Prepend = append(d.Prepend, RuleItem{Type: "DOMAIN-SUFFIX", Value: "https://new.com/", Target: "Proxy", Note: "新加的"})
	d.Hosts = []HostEntry{{Domain: "*.lab.io", Value: "10.0.0.5"}}
	d.FakeIPFilter = append(d.FakeIPFilter, FakeIPEntry{Value: "+.lab.io", Note: "内网"})
	d.DNSPolicy = []DNSPolicyEntry{{Domain: "+.lab.io", Server: "10.0.0.53, 10.0.0.54"}}

	res, err := m.SaveRules(d)
	if err != nil || !res.Saved || res.Applied {
		t.Fatalf("save: %+v %v", res, err)
	}
	rules, _ := os.ReadFile(filepath.Join(dir, "profiles", "rules.yaml"))
	if !strings.Contains(string(rules), "DOMAIN-SUFFIX,new.com,Proxy # 新加的") || !strings.Contains(string(rules), "DOMAIN-KEYWORD,ads,REJECT") ||
		!strings.Contains(string(rules), "# Profile Enhancement Rules Template") {
		t.Fatalf("rules file:\n%s", rules)
	}
	merge, _ := os.ReadFile(filepath.Join(dir, "profiles", "merge.yaml"))
	var mm map[string]interface{}
	if err := yaml.Unmarshal(merge, &mm); err != nil {
		t.Fatal(err)
	}
	if mm["profile"] == nil || mm["hosts"].(map[string]interface{})["*.lab.io"] != "10.0.0.5" {
		t.Fatalf("merge file:\n%s", merge)
	}
	dns := mm["dns"].(map[string]interface{})
	if len(dns["fake-ip-filter"].([]interface{})) != 3 || len(dns["nameserver-policy"].(map[string]interface{})["+.lab.io"].([]interface{})) != 2 {
		t.Fatalf("merge dns:\n%s", merge)
	}
	// Offline: generated runtime is left for Clash Verge to regenerate.
	runtime, _ := os.ReadFile(filepath.Join(dir, "clash-verge.yaml"))
	if strings.Contains(string(runtime), "new.com") {
		t.Fatal("runtime should not change when core is unreachable")
	}

	// Reload round-trip: state reflects the save; restoring the subscription list drops the override.
	st, _ = m.ReadRules()
	if !st.FakeIPOverridden || len(st.Draft.Prepend) != 2 || st.Draft.Prepend[1].Note != "新加的" {
		t.Fatalf("reloaded: %+v", st)
	}
	d = st.Draft
	d.FakeIPFilter = []FakeIPEntry{{Value: "*.lan"}, {Value: "+.stun.*.*"}}
	if _, err := m.SaveRules(d); err != nil {
		t.Fatal(err)
	}
	merge, _ = os.ReadFile(filepath.Join(dir, "profiles", "merge.yaml"))
	if strings.Contains(string(merge), "fake-ip-filter") {
		t.Fatalf("fake-ip override should be removed:\n%s", merge)
	}
}

func TestSaveRulesValidation(t *testing.T) {
	m, dir := setupClashDir(t, "/nonexistent/hatch-test.sock")
	before, _ := os.ReadFile(filepath.Join(dir, "profiles", "rules.yaml"))
	res, err := m.SaveRules(RulesDraft{Prepend: []RuleItem{{Type: "IP-CIDR", Value: "300.1.1.1", Target: "DIRECT"}}})
	if err != nil || res.Saved || len(res.Issues) != 1 || res.Issues[0].Section != "prepend" {
		t.Fatalf("expected issue: %+v %v", res, err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, "profiles", "rules.yaml"))
	if string(before) != string(after) {
		t.Fatal("invalid draft must not write")
	}
}

// fakeCore serves PUT /configs on a unix socket and records the payload.
func fakeCore(t *testing.T, status int) (string, *string) {
	t.Helper()
	sockDir, err := os.MkdirTemp("/tmp", "hatchvpn")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(sockDir) })
	sock := filepath.Join(sockDir, "core.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	var got string
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/configs" {
			body, _ := io.ReadAll(r.Body)
			var req struct{ Payload string }
			_ = json.Unmarshal(body, &req)
			got = req.Payload
			w.WriteHeader(status)
			if status != http.StatusNoContent {
				_, _ = w.Write([]byte(`{"message":"rule parse error"}`))
			}
			return
		}
		_, _ = w.Write([]byte(`{"version":"test"}`))
	})}
	go srv.Serve(ln)
	t.Cleanup(func() { srv.Close() })
	return sock, &got
}

func TestSaveRulesHotReload(t *testing.T) {
	sock, payload := fakeCore(t, http.StatusNoContent)
	m, dir := setupClashDir(t, sock)
	st, _ := m.ReadRules()
	if !st.CoreReachable {
		t.Fatal("fake core should be reachable")
	}
	d := st.Draft
	d.Prepend = []RuleItem{{Type: "DOMAIN-SUFFIX", Value: "new.com", Target: "Proxy"}}
	d.Hosts = nil
	d.FakeIPFilter = []FakeIPEntry{{Value: "*.lan"}}
	res, err := m.SaveRules(d)
	if err != nil || !res.Applied {
		t.Fatalf("save: %+v %v", res, err)
	}
	var rt struct {
		Rules []string          `yaml:"rules"`
		Hosts map[string]string `yaml:"hosts"`
		DNS   struct {
			FakeIPFilter []string `yaml:"fake-ip-filter"`
		} `yaml:"dns"`
	}
	if err := yaml.Unmarshal([]byte(*payload), &rt); err != nil {
		t.Fatal(err)
	}
	if strings.Join(rt.Rules, "|") != "DOMAIN-SUFFIX,new.com,Proxy|GEOIP,CN,DIRECT|MATCH,Proxy" {
		t.Fatalf("runtime rules: %v", rt.Rules)
	}
	if len(rt.Hosts) != 0 || strings.Join(rt.DNS.FakeIPFilter, "|") != "*.lan" {
		t.Fatalf("runtime hosts/dns: %+v", rt)
	}
	disk, _ := os.ReadFile(filepath.Join(dir, "clash-verge.yaml"))
	if !strings.Contains(string(disk), "new.com") || !strings.Contains(string(disk), "# Generated by Clash Verge") {
		t.Fatalf("runtime file not synced:\n%s", disk)
	}
}

func TestSaveRulesCoreRejects(t *testing.T) {
	sock, _ := fakeCore(t, http.StatusBadRequest)
	m, dir := setupClashDir(t, sock)
	before, _ := os.ReadFile(filepath.Join(dir, "profiles", "rules.yaml"))
	st, _ := m.ReadRules()
	d := st.Draft
	d.Prepend = append(d.Prepend, RuleItem{Type: "GEOSITE", Value: "nope", Target: "Proxy"})
	_, err := m.SaveRules(d)
	if err == nil || !strings.Contains(err.Error(), "rule parse error") {
		t.Fatalf("expected core rejection, got %v", err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, "profiles", "rules.yaml"))
	if string(before) != string(after) {
		t.Fatal("rejected config must not be written")
	}
}

func TestSubscriptions(t *testing.T) {
	m, _ := setupClashDir(t, "/nonexistent/hatch-test.sock")
	v, err := m.Subscriptions()
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Subscriptions) != 1 {
		t.Fatalf("subs: %+v", v.Subscriptions)
	}
	s := v.Subscriptions[0]
	if s.URLHost != "sub.example.com" || !s.IsCurrent || s.NodeCount != 3 || s.RuleCount != 2 || !s.AutoUpdate {
		t.Fatalf("summary: %+v", s)
	}
	if len(s.Regions) != 2 || s.Regions[0].Name != "美国" || s.Regions[0].Count != 2 {
		t.Fatalf("regions: %+v", s.Regions)
	}
	if len(s.Enhancements) != 2 || !s.Enhancements[0].Customized || s.Enhancements[0].Summary != "前置 1 · 删除 1" {
		t.Fatalf("enhancements: %+v", s.Enhancements)
	}
	if len(v.Global) != 1 || !v.Global[0].Customized || len(v.Files) != 4 {
		t.Fatalf("global/files: %+v %+v", v.Global, v.Files)
	}
	if u, _ := m.SubscriptionURL("SUB"); !strings.Contains(u, "token=secret") {
		t.Fatal("url")
	}
}
