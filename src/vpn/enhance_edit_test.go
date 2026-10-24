package vpn

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func suggestion(t *testing.T, raw string) RuleSuggestion {
	t.Helper()
	s, err := ParseRuleSuggestion(raw)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestVerifyRuleSuggestion(t *testing.T) {
	m, _ := setupClashDir(t, "/nonexistent/hatch-test.sock")
	st, _ := m.ReadRules()

	ok := suggestion(t, "```json\n"+`{"summary":"notion 走代理","add":{"rules":[{"rule":"DOMAIN-SUFFIX,notion.so,Proxy","note":"Notion"}],"tun":["192.168.50.0/24"]},
	  "remove":{"rules":["DOMAIN,old.com,DIRECT"]},
	  "tests":[{"input":"www.notion.so","expect":"Proxy"},{"input":"192.168.50.7","expect":"BYPASS"}]}`+"\n```")
	p, err := m.VerifyRuleSuggestion(st.Draft, ok)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Draft.Prepend) != 1 || p.Draft.Prepend[0].String() != "DOMAIN-SUFFIX,notion.so,Proxy" || len(p.Draft.TunExclude) != 1 {
		t.Fatalf("draft: %+v", p.Draft)
	}
	if len(p.Changes) != 3 || len(p.Tests) != 2 || !p.Tests[0].Pass || !p.Tests[1].Pass {
		t.Fatalf("preview: %+v", p)
	}

	cases := map[string]string{
		"wrong expectation": `{"add":{"rules":[{"rule":"DOMAIN-SUFFIX,x.com,DIRECT"}]},"tests":[{"input":"x.com","expect":"Proxy"}]}`,
		"no change":         `{"add":{"rules":[{"rule":"DOMAIN,old.com,DIRECT"}]}}`,
		"invalid entry":     `{"add":{"rules":[{"rule":"IP-CIDR,999.1.1.1,DIRECT"}]}}`,
	}
	for name, raw := range cases {
		if _, err := m.VerifyRuleSuggestion(st.Draft, suggestion(t, raw)); err == nil {
			t.Fatalf("%s: expected rejection", name)
		}
	}
}

func addEnhanceFile(t *testing.T, dir, uid, typ, file, content string, link bool) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "profiles", file), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	pf := filepath.Join(dir, "profiles.yaml")
	data, _ := os.ReadFile(pf)
	s := strings.Replace(string(data), "items:\n", "items:\n- {uid: "+uid+", type: "+typ+", file: "+file+"}\n", 1)
	if link {
		s = strings.Replace(s, "rules: RULES}", "rules: RULES, "+typ+": "+uid+"}", 1)
	}
	if err := os.WriteFile(pf, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnhanceEditGroupsHotReload(t *testing.T) {
	sock, payload := fakeCore(t, http.StatusNoContent)
	m, dir := setupClashDir(t, sock)
	addEnhanceFile(t, dir, "GROUPS", "groups", "groups.yaml", "prepend: []\nappend: []\ndelete: []\n", true)

	info, err := m.EnhanceInfo("groups.yaml")
	if err != nil || info.Scope != "current" || info.Desc == "" {
		t.Fatalf("info: %+v %v", info, err)
	}
	newContent := "prepend:\n  - {name: AI 专用, type: select, proxies: [Proxy, DIRECT]}\nappend: []\ndelete: []\n"
	p, err := m.PreviewEnhanceEdit("groups.yaml", "```yaml\n"+newContent+"```")
	if err != nil {
		t.Fatal(err)
	}
	if !p.Live || len(p.Effects) != 1 || !strings.Contains(p.Effects[0], "新增策略组 1 个：AI 专用") {
		t.Fatalf("preview: %+v", p)
	}
	res, err := m.ApplyEnhanceEdit("groups.yaml", newContent)
	if err != nil || !res.Applied {
		t.Fatalf("apply: %+v %v", res, err)
	}
	var rt struct {
		Groups []struct{ Name string } `yaml:"proxy-groups"`
	}
	_ = yaml.Unmarshal([]byte(*payload), &rt)
	if len(rt.Groups) != 2 || rt.Groups[0].Name != "AI 专用" {
		t.Fatalf("runtime groups: %+v", rt.Groups)
	}
	disk, _ := os.ReadFile(filepath.Join(dir, "profiles", "groups.yaml"))
	if !strings.Contains(string(disk), "AI 专用") {
		t.Fatal("file not written")
	}
	// The same content again is no change at all.
	if _, err := m.PreviewEnhanceEdit("groups.yaml", newContent); err == nil {
		t.Fatal("identical content must be rejected")
	}
}

func TestEnhanceEditRejectsIneffectiveAndInvalid(t *testing.T) {
	m, dir := setupClashDir(t, "/nonexistent/hatch-test.sock")
	addEnhanceFile(t, dir, "PROXIES", "proxies", "proxies.yaml", "prepend: []\nappend: []\ndelete: []\n", true)
	// deleting a node that does not exist changes nothing
	if _, err := m.PreviewEnhanceEdit("proxies.yaml", "prepend: []\nappend: []\ndelete: [不存在]\n"); err == nil || !strings.Contains(err.Error(), "不会让最终生效的配置发生任何变化") {
		t.Fatalf("expected ineffective, got %v", err)
	}
	if _, err := m.PreviewEnhanceEdit("proxies.yaml", "foo: 1\n"); err == nil {
		t.Fatal("unknown key must be rejected")
	}
	if _, err := m.PreviewEnhanceEdit("proxies.yaml", "prepend:\n  - {server: 1.1.1.1}\n"); err == nil {
		t.Fatal("node without name/type must be rejected")
	}
	if _, err := m.PreviewEnhanceEdit("sub.yaml", "x: 1"); err == nil {
		t.Fatal("subscription file itself is not an enhancement")
	}
}

func TestEnhanceEditScript(t *testing.T) {
	if _, err := os.Stat("/usr/bin/osascript"); err != nil {
		t.Skip("osascript unavailable")
	}
	m, dir := setupClashDir(t, "/nonexistent/hatch-test.sock")
	addEnhanceFile(t, dir, "SCRIPT", "script", "script.js", "function main(config, profileName) {\n  return config;\n}\n", true)

	addRule := "function main(config, profileName) {\n  config.rules = ['DOMAIN,hatch.test,DIRECT', ...config.rules];\n  return config;\n}\n"
	p, err := m.PreviewEnhanceEdit("script.js", addRule)
	if err != nil {
		t.Fatal(err)
	}
	if !p.Live || len(p.Effects) != 1 || !strings.Contains(p.Effects[0], "DOMAIN,hatch.test,DIRECT") {
		t.Fatalf("preview: %+v", p)
	}
	if _, err := m.PreviewEnhanceEdit("script.js", "function main(config) { return config.nope.x; }"); err == nil || !strings.Contains(err.Error(), "脚本运行出错") {
		t.Fatalf("expected runtime error, got %v", err)
	}
	if _, err := m.PreviewEnhanceEdit("script.js", "function main(config) { return config; } // no-op"); err == nil {
		t.Fatal("no-op script must be rejected")
	}
}

func TestEnhanceEditMergeOtherSubscription(t *testing.T) {
	m, dir := setupClashDir(t, "/nonexistent/hatch-test.sock")
	// a second, inactive subscription with its own merge file
	if err := os.WriteFile(filepath.Join(dir, "profiles", "sub2.yaml"), []byte(testSubFile), 0o644); err != nil {
		t.Fatal(err)
	}
	addEnhanceFile(t, dir, "M2", "merge", "m2.yaml", "profile: {store-selected: true}\n", false)
	pf := filepath.Join(dir, "profiles.yaml")
	data, _ := os.ReadFile(pf)
	_ = os.WriteFile(pf, []byte(string(data)+"- {uid: SUB2, type: local, file: sub2.yaml, option: {merge: M2}}\n"), 0o644)

	p, err := m.PreviewEnhanceEdit("m2.yaml", "profile: {store-selected: true}\nlog-level: warning\n")
	if err != nil {
		t.Fatal(err)
	}
	if p.Live || p.Note == "" || !strings.Contains(strings.Join(p.Effects, ""), "log-level") {
		t.Fatalf("preview: %+v", p)
	}
}

func TestLineDiff(t *testing.T) {
	d := lineDiff("a\nb\nc\n", "a\nc\nd\n")
	var ops []string
	for _, l := range d {
		ops = append(ops, l.Op+l.Text)
	}
	if strings.Join(ops, "|") != " a|-b| c|+d" {
		t.Fatalf("diff: %v", ops)
	}
}
