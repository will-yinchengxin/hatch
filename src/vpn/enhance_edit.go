package vpn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"hatch/core/fsutil"

	"gopkg.in/yaml.v3"
)

// ─── AI edits of enhancement files (订阅 → 你对这个订阅做的自定义) ─────────────
//
// An edit is only offered when it is *effective*: Hatch replays the file's
// Clash Verge stage (merge / script / rules / proxies / groups) on the real
// config, describes what changes in the final config, rejects edits that change
// nothing, and validates the result with `verge-mihomo -t`. Applying hot-reloads
// the core when the replay is exact (otherwise the file is saved and Clash
// Verge applies it the next time it regenerates).

type EnhanceInfo struct {
	File     string   `json:"file"`
	Kind     string   `json:"kind"`  // merge | script | rules | proxies | groups
	Scope    string   `json:"scope"` // current | global | other
	Label    string   `json:"label"`
	Desc     string   `json:"desc"`
	Effect   string   `json:"effect"`
	Examples []string `json:"examples"`
}

type DiffLine struct {
	Op   string `json:"op"` // " " | "+" | "-"
	Text string `json:"text"`
}

type EnhanceEditPreview struct {
	File    string     `json:"file"`
	Kind    string     `json:"kind"`
	Content string     `json:"content"`
	Diff    []DiffLine `json:"diff"`
	Effects []string   `json:"effects"`
	Live    bool       `json:"live"` // applying hot-reloads the running core
	Note    string     `json:"note"`
}

// enhanceInfo is the plain-language explanation shown for each enhancement row.
var enhanceInfo = map[string]EnhanceInfo{
	"rules": {Label: "分流规则",
		Desc:     "在订阅规则之前插入你自己的规则，决定哪些网站、IP、应用直连、走代理或被拦截。",
		Effect:   "命中第一条规则就停止，所以这里的规则总是比订阅规则优先。",
		Examples: []string{"让 notion.so 和 figma.com 走代理", "删除所有指向 REJECT 的规则", "让 *.corp.com 直连"}},
	"merge": {Label: "Hosts / DNS 覆写",
		Desc:     "按字段覆盖订阅里的配置，常用来设置 hosts、DNS、Fake-IP 过滤、日志级别等。",
		Effect:   "映射会逐层合并，列表会整体替换订阅里的同名列表。",
		Examples: []string{"把日志级别改成 warning", "让 git.corp.com 解析到 10.0.0.8", "关闭 IPv6"}},
	"script": {Label: "脚本",
		Desc:     "用 JavaScript 在订阅加载后批量加工整份配置，适合按名字批量改名、筛选节点、动态生成策略组等复杂需求。",
		Effect:   "每次生成配置时运行 main(config)，返回的 config 即最终配置。",
		Examples: []string{"给所有节点名加上国旗 emoji", "删除名字包含「防失联」的节点", "新增一个只包含香港节点的策略组"}},
	"proxies": {Label: "额外节点",
		Desc:     "在订阅之外追加你自己的节点（如自建服务器），或删除订阅里不想要的节点。",
		Effect:   "prepend / append 添加节点，delete 按名字删除节点。",
		Examples: []string{"删除名字是「防失联 ftqfabu.com」的节点", "添加一个 SOCKS5 节点 127.0.0.1:1080，名字叫「本地代理」"}},
	"groups": {Label: "额外策略组",
		Desc:     "新增或删除策略组，例如「AI 专用」「流媒体」，把挑选出的节点组合起来给规则使用。",
		Effect:   "prepend / append 添加策略组，delete 按名字删除策略组。",
		Examples: []string{"新增「AI 专用」策略组，包含所有美国节点", "新增自动测速的「香港自动」策略组"}},
}

// enhanceTarget is one enhancement file resolved against profiles.yaml.
type enhanceTarget struct {
	dir     string
	item    profileItem
	path    string
	scope   string
	owner   profileItem // subscription owning it (scope current / other)
	runtime string      // clash-verge.yaml
}

func (m Manager) findEnhance(file string) (enhanceTarget, error) {
	if !safeProfilePath(file) {
		return enhanceTarget{}, fmt.Errorf("非法文件名: %s", file)
	}
	dir, meta, err := m.readProfilesMeta()
	if err != nil {
		return enhanceTarget{}, err
	}
	var item *profileItem
	for i := range meta.Items {
		if meta.Items[i].File == file {
			item = &meta.Items[i]
		}
	}
	if item == nil {
		return enhanceTarget{}, fmt.Errorf("profiles.yaml 中没有 %s", file)
	}
	if _, ok := enhanceInfo[item.Type]; !ok {
		return enhanceTarget{}, errors.New("只能修改订阅的增强文件（规则 / 覆写 / 脚本 / 节点 / 策略组）")
	}
	t := enhanceTarget{dir: dir, item: *item, path: filepath.Join(dir, "profiles", file), scope: "global",
		runtime: filepath.Join(dir, "clash-verge.yaml")}
	for _, sub := range meta.Items {
		o := sub.Option
		for _, uid := range []string{o.Merge, o.Script, o.Rules, o.Proxies, o.Groups} {
			if uid != "" && uid == item.UID {
				t.owner = sub
				t.scope = "other"
				if sub.UID == meta.Current {
					t.scope = "current"
				}
			}
		}
	}
	return t, nil
}

// EnhanceInfo explains one enhancement file (what it is, what it does).
func (m Manager) EnhanceInfo(file string) (EnhanceInfo, error) {
	t, err := m.findEnhance(file)
	if err != nil {
		return EnhanceInfo{}, err
	}
	info := enhanceInfo[t.item.Type]
	info.File, info.Kind, info.Scope = file, t.item.Type, t.scope
	return info, nil
}

// EnhanceAIContext is what the AI needs to write a valid edit.
type EnhanceAIContext struct {
	Info    EnhanceInfo
	Content string
	Groups  []string
	Proxies []string
}

func (m Manager) EnhanceAIContext(file string) (EnhanceAIContext, error) {
	t, err := m.findEnhance(file)
	if err != nil {
		return EnhanceAIContext{}, err
	}
	info, _ := m.EnhanceInfo(file)
	data, _ := os.ReadFile(t.path)
	ctx := EnhanceAIContext{Info: info, Content: string(data)}
	base, _ := loadConfigMap(t.runtime)
	if t.scope == "other" {
		base, _ = loadConfigMap(filepath.Join(t.dir, "profiles", t.owner.File))
	}
	for _, g := range seqOf(base["proxy-groups"]) {
		ctx.Groups = append(ctx.Groups, nameOf(g))
	}
	for _, p := range seqOf(base["proxies"]) {
		ctx.Proxies = append(ctx.Proxies, nameOf(p))
	}
	return ctx, nil
}

// PreviewEnhanceEdit validates new content and reports its real effect.
func (m Manager) PreviewEnhanceEdit(file, content string) (EnhanceEditPreview, error) {
	p, _, _, err := m.prepareEnhanceEdit(file, content)
	return p, err
}

// ApplyEnhanceEdit re-validates, hot-reloads when exact, then writes the file.
func (m Manager) ApplyEnhanceEdit(file, content string) (RulesSaveResult, error) {
	res := RulesSaveResult{Issues: []RuleIssue{}, Warnings: []string{}}
	p, after, t, err := m.prepareEnhanceEdit(file, content)
	if err != nil {
		return res, err
	}
	var runtimeBytes []byte
	if p.Live {
		var node yaml.Node
		if err := node.Encode(after); err != nil {
			return res, err
		}
		doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{&node}}
		if res.Applied, runtimeBytes, err = m.applyRuntime(doc); err != nil {
			return res, err
		}
	}
	if err := fsutil.WriteWithBackup(t.path, []byte(strings.TrimRight(p.Content, "\n")+"\n"), 0o644); err != nil {
		return res, fmt.Errorf("写入 %s 失败: %w", file, err)
	}
	if res.Applied && runtimeBytes != nil {
		_ = os.WriteFile(t.runtime, runtimeBytes, 0o644)
	}
	res.Saved = true
	switch {
	case res.Applied:
		res.Message = "已保存并立即生效"
	case p.Live:
		res.Message = "已保存。Clash Verge 内核未运行，下次启动时生效"
	default:
		res.Message = "已保存。" + p.Note
	}
	return res, nil
}

func (m Manager) prepareEnhanceEdit(file, content string) (EnhanceEditPreview, map[string]any, enhanceTarget, error) {
	t, err := m.findEnhance(file)
	if err != nil {
		return EnhanceEditPreview{}, nil, t, err
	}
	kind := t.item.Type
	oldData, _ := os.ReadFile(t.path)
	oldContent := string(oldData)
	content = strings.TrimSpace(stripCodeFence(content))
	p := EnhanceEditPreview{File: file, Kind: kind, Content: content}
	if content == strings.TrimSpace(oldContent) {
		return p, nil, t, errors.New("新内容和当前文件完全一样，没有产生任何修改")
	}
	if err := checkEnhanceSyntax(kind, content); err != nil {
		return p, nil, t, err
	}

	var before, after map[string]any
	precise := true
	runtime, rtErr := loadConfigMap(t.runtime)
	if t.scope != "other" && rtErr == nil {
		before = runtime
		after, precise, err = replayStage(kind, cloneMap(runtime), oldContent, content)
	} else {
		base, err2 := loadConfigMap(filepath.Join(t.dir, "profiles", t.owner.File))
		if err2 != nil {
			return p, nil, t, fmt.Errorf("读取订阅配置失败: %w", err2)
		}
		if before, err = applyStage(kind, cloneMap(base), oldContent); err == nil {
			after, err = applyStage(kind, cloneMap(base), content)
		}
	}
	if err != nil {
		return p, nil, t, err
	}

	p.Effects = describeEffects(before, after)
	if len(p.Effects) == 0 {
		return p, nil, t, errors.New("这次修改不会让最终生效的配置发生任何变化（可能引用了不存在的节点 / 策略组，或与现有内容重复），请换个说法或补充细节")
	}
	afterBytes, err := yaml.Marshal(after)
	if err != nil {
		return p, nil, t, err
	}
	if err := m.testConfig(afterBytes); err != nil {
		return p, nil, t, err
	}
	p.Diff = lineDiff(oldContent, content)
	p.Live = t.scope != "other" && rtErr == nil && precise
	switch {
	case t.scope == "other":
		p.Note = "这个文件属于未在使用的订阅，切换到该订阅后生效"
	case !precise:
		p.Note = "这次修改删除了之前的设置，需要在 Clash Verge 中重新激活当前订阅后才会完全生效"
	}
	return p, after, t, nil
}

// ─── Stage replay (mirrors Clash Verge's enhance chain) ──────────────────────

type seqStage struct {
	Prepend []any `yaml:"prepend"`
	Append  []any `yaml:"append"`
	Delete  []any `yaml:"delete"`
}

var seqField = map[string]string{"rules": "rules", "proxies": "proxies", "groups": "proxy-groups"}

func checkEnhanceSyntax(kind, content string) error {
	if kind == "script" {
		if !regexp.MustCompile(`function\s+main\s*\(`).MatchString(content) {
			return errors.New("脚本必须定义 function main(config, profileName) 并返回 config")
		}
		return nil
	}
	var root map[string]any
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return fmt.Errorf("YAML 格式错误: %w", err)
	}
	field, isSeq := seqField[kind]
	if !isSeq {
		return nil
	}
	for k, v := range root {
		if k != "prepend" && k != "append" && k != "delete" {
			return fmt.Errorf("%s 文件只能包含 prepend / append / delete，发现了「%s」", enhanceInfo[kind].Label, k)
		}
		if v != nil {
			if _, ok := v.([]any); !ok {
				return fmt.Errorf("%s 必须是列表", k)
			}
		}
	}
	st, _ := parseSeqStage(content)
	for _, list := range [][]any{st.Prepend, st.Append} {
		for _, it := range list {
			switch field {
			case "rules":
				s, ok := it.(string)
				if !ok {
					return fmt.Errorf("规则必须是字符串，如 DOMAIN-SUFFIX,example.com,DIRECT")
				}
				if _, err := normalizeRule(ParseRule(s)); err != nil {
					return fmt.Errorf("规则「%s」无效: %v", s, err)
				}
			default:
				mp, ok := it.(map[string]any)
				if !ok || nameOf(mp) == "" || mp["type"] == nil {
					return fmt.Errorf("%s 的每一项都必须包含 name 和 type", enhanceInfo[kind].Label)
				}
			}
		}
	}
	return nil
}

func parseSeqStage(content string) (seqStage, error) {
	var st seqStage
	err := yaml.Unmarshal([]byte(content), &st)
	return st, err
}

func seqKey(field string, it any) string {
	if field == "rules" {
		if s, ok := it.(string); ok {
			return canonicalRule(s)
		}
		return ""
	}
	if s, ok := it.(string); ok { // delete lists hold names
		return s
	}
	return nameOf(it)
}

// applySeq returns prepend + (list − delete) + append.
func applySeq(field string, list []any, st seqStage) []any {
	del := map[string]bool{}
	for _, d := range st.Delete {
		del[seqKey(field, d)] = true
	}
	out := append([]any{}, st.Prepend...)
	for _, it := range list {
		if !del[seqKey(field, it)] {
			out = append(out, it)
		}
	}
	return append(out, st.Append...)
}

// stripSeq removes a previously applied prepend prefix / append suffix.
func stripSeq(field string, list []any, st seqStage) []any {
	match := func(part []any, want []any) bool {
		if len(want) > len(part) {
			return false
		}
		for i := range want {
			if seqKey(field, part[i]) != seqKey(field, want[i]) {
				return false
			}
		}
		return true
	}
	if match(list, st.Prepend) {
		list = list[len(st.Prepend):]
	}
	if n := len(st.Append); n > 0 && len(list) >= n && match(list[len(list)-n:], st.Append) {
		list = list[:len(list)-n]
	}
	return list
}

// applyStage applies one stage to a base config (non-current subscriptions).
func applyStage(kind string, cfg map[string]any, content string) (map[string]any, error) {
	switch kind {
	case "merge":
		var mp map[string]any
		if err := yaml.Unmarshal([]byte(content), &mp); err != nil {
			return nil, err
		}
		deepMerge(cfg, mp)
		return cfg, nil
	case "script":
		return runScript(content, cfg)
	default:
		field := seqField[kind]
		st, err := parseSeqStage(content)
		if err != nil {
			return nil, err
		}
		cfg[field] = applySeq(field, seqOf(cfg[field]), st)
		return cfg, nil
	}
}

// replayStage swaps the old stage for the new one on the running config.
// precise=false means the old stage's effect cannot be fully undone here.
func replayStage(kind string, runtime map[string]any, oldContent, newContent string) (map[string]any, bool, error) {
	switch kind {
	case "merge":
		var oldMp, newMp map[string]any
		_ = yaml.Unmarshal([]byte(oldContent), &oldMp)
		if err := yaml.Unmarshal([]byte(newContent), &newMp); err != nil {
			return nil, false, err
		}
		precise := true
		newPaths := keyPaths(newMp, "")
		for path := range keyPaths(oldMp, "") {
			if !newPaths[path] {
				precise = false
			}
		}
		deepMerge(runtime, newMp)
		return runtime, precise, nil
	case "script":
		out, err := runScript(newContent, runtime)
		return out, scriptIsTemplate(oldContent), err
	default:
		field := seqField[kind]
		oldSt, _ := parseSeqStage(oldContent)
		newSt, err := parseSeqStage(newContent)
		if err != nil {
			return nil, false, err
		}
		keep := map[string]bool{}
		for _, d := range newSt.Delete {
			keep[seqKey(field, d)] = true
		}
		precise := true
		for _, d := range oldSt.Delete {
			if !keep[seqKey(field, d)] {
				precise = false // a previously deleted item cannot be restored here
			}
		}
		runtime[field] = applySeq(field, stripSeq(field, seqOf(runtime[field]), oldSt), newSt)
		return runtime, precise, nil
	}
}

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		sm, sok := v.(map[string]any)
		dm, dok := dst[k].(map[string]any)
		if sok && dok {
			deepMerge(dm, sm)
			continue
		}
		dst[k] = v
	}
}

func keyPaths(m map[string]any, prefix string) map[string]bool {
	out := map[string]bool{}
	for k, v := range m {
		p := prefix + "/" + k
		if sub, ok := v.(map[string]any); ok && len(sub) > 0 {
			for sp := range keyPaths(sub, p) {
				out[sp] = true
			}
			continue
		}
		out[p] = true
	}
	return out
}

func scriptIsTemplate(code string) bool {
	code = regexp.MustCompile(`//[^\n]*`).ReplaceAllString(code, "")
	code = strings.Join(strings.Fields(code), "")
	return code == "" || code == "functionmain(config,profileName){returnconfig;}" || code == "functionmain(config){returnconfig;}"
}

// ─── Effects ─────────────────────────────────────────────────────────────────

// describeEffects lists, in plain words, how the final config changes.
func describeEffects(before, after map[string]any) []string {
	var out []string
	named := func(field, label string) {
		b, a := indexByName(seqOf(before[field])), indexByName(seqOf(after[field]))
		var added, removed, changed []string
		for _, n := range sortedKeys(a) {
			if _, ok := b[n]; !ok {
				added = append(added, n)
			} else if !reflect.DeepEqual(b[n], a[n]) {
				changed = append(changed, n)
			}
		}
		for _, n := range sortedKeys(b) {
			if _, ok := a[n]; !ok {
				removed = append(removed, n)
			}
		}
		if len(added) > 0 {
			out = append(out, fmt.Sprintf("新增%s %d 个：%s", label, len(added), listPreview(added)))
		}
		if len(removed) > 0 {
			out = append(out, fmt.Sprintf("删除%s %d 个：%s", label, len(removed), listPreview(removed)))
		}
		if len(changed) > 0 {
			out = append(out, fmt.Sprintf("修改%s %d 个：%s", label, len(changed), listPreview(changed)))
		}
	}
	named("proxies", "节点")
	named("proxy-groups", "策略组")

	br, ar := ruleSet(seqOf(before["rules"])), ruleSet(seqOf(after["rules"]))
	var addR, delR []string
	for _, r := range seqOf(after["rules"]) {
		if s := canonicalRule(fmt.Sprint(r)); !br[s] {
			addR = append(addR, s)
		}
	}
	for _, r := range seqOf(before["rules"]) {
		if s := canonicalRule(fmt.Sprint(r)); !ar[s] {
			delR = append(delR, s)
		}
	}
	if len(addR) > 0 {
		out = append(out, fmt.Sprintf("新增规则 %d 条：%s", len(addR), listPreview(addR)))
	}
	if len(delR) > 0 {
		out = append(out, fmt.Sprintf("删除规则 %d 条：%s", len(delR), listPreview(delR)))
	}
	if len(addR) == 0 && len(delR) == 0 && !reflect.DeepEqual(before["rules"], after["rules"]) {
		out = append(out, "调整了规则顺序")
	}

	keys := map[string]bool{}
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}
	var other []string
	for k := range keys {
		if k == "proxies" || k == "proxy-groups" || k == "rules" {
			continue
		}
		if !reflect.DeepEqual(before[k], after[k]) {
			other = append(other, k)
		}
	}
	sort.Strings(other)
	for _, k := range other {
		out = append(out, "修改配置项「"+k+"」"+fieldDiffHint(before[k], after[k]))
	}
	return out
}

func fieldDiffHint(b, a any) string {
	bm, bok := b.(map[string]any)
	am, aok := a.(map[string]any)
	if !bok || !aok {
		if !aok && a != nil {
			return fmt.Sprintf("：%v → %v", shortVal(b), shortVal(a))
		}
		return ""
	}
	var keys []string
	for k := range am {
		if !reflect.DeepEqual(bm[k], am[k]) {
			keys = append(keys, k)
		}
	}
	for k := range bm {
		if _, ok := am[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return "（" + listPreview(keys) + "）"
}

func shortVal(v any) string {
	if v == nil {
		return "无"
	}
	s := fmt.Sprint(v)
	if len([]rune(s)) > 40 {
		s = string([]rune(s)[:40]) + "…"
	}
	return s
}

func listPreview(items []string) string {
	const max = 5
	if len(items) <= max {
		return strings.Join(items, "、")
	}
	return strings.Join(items[:max], "、") + fmt.Sprintf(" 等 %d 个", len(items))
}

func ruleSet(list []any) map[string]bool {
	out := map[string]bool{}
	for _, r := range list {
		out[canonicalRule(fmt.Sprint(r))] = true
	}
	return out
}

func indexByName(list []any) map[string]any {
	out := map[string]any{}
	for _, it := range list {
		if n := nameOf(it); n != "" {
			out[n] = it
		}
	}
	return out
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ─── Generic config helpers ──────────────────────────────────────────────────

func loadConfigMap(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var mp map[string]any
	if err := yaml.Unmarshal(data, &mp); err != nil {
		return nil, err
	}
	if mp == nil {
		mp = map[string]any{}
	}
	return mp, nil
}

func cloneMap(m map[string]any) map[string]any {
	data, _ := json.Marshal(m)
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	if out == nil {
		out = map[string]any{}
	}
	return normalizeNumbers(out).(map[string]any)
}

// normalizeNumbers turns JSON float64s that are whole numbers back into ints so
// re-encoded YAML keeps "port: 7890" instead of "port: 7890.0"-like values and
// DeepEqual comparisons against YAML-decoded configs stay meaningful.
func normalizeNumbers(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			t[k] = normalizeNumbers(x)
		}
		return t
	case []any:
		for i, x := range t {
			t[i] = normalizeNumbers(x)
		}
		return t
	case float64:
		if t == float64(int64(t)) {
			return int(t)
		}
	}
	return v
}

func seqOf(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}

func nameOf(v any) string {
	if mp, ok := v.(map[string]any); ok {
		if n, ok := mp["name"].(string); ok {
			return n
		}
	}
	return ""
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s, "\n"); i >= 0 {
			s = s[i+1:]
		}
		s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	}
	return s
}

// lineDiff is a small LCS line diff for the preview (files here are short).
func lineDiff(a, b string) []DiffLine {
	al := strings.Split(strings.TrimRight(a, "\n"), "\n")
	bl := strings.Split(strings.TrimRight(b, "\n"), "\n")
	if len(al)*len(bl) > 4_000_000 {
		return []DiffLine{{Op: " ", Text: "（文件较长，未显示逐行差异）"}}
	}
	n, m := len(al), len(bl)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if al[i] == bl[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	var out []DiffLine
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case al[i] == bl[j]:
			out = append(out, DiffLine{" ", al[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, DiffLine{"-", al[i]})
			i++
		default:
			out = append(out, DiffLine{"+", bl[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, DiffLine{"-", al[i]})
	}
	for ; j < m; j++ {
		out = append(out, DiffLine{"+", bl[j]})
	}
	return out
}

// ─── Validation with the real core ───────────────────────────────────────────

func coreBinary() string {
	home, _ := os.UserHomeDir()
	for _, p := range []string{
		"/Applications/Clash Verge.app/Contents/MacOS/verge-mihomo",
		filepath.Join(home, "Applications/Clash Verge.app/Contents/MacOS/verge-mihomo"),
	} {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return ""
}

var coreErrRe = regexp.MustCompile(`level=(?:error|fatal) msg="([^"]+)"`)

// testConfig runs `verge-mihomo -t` on a full config, using the Clash Verge
// data dir for geodata. It is skipped when the core binary is not installed.
func (m Manager) testConfig(content []byte) error {
	if m.skipCoreCheck {
		return nil
	}
	bin := coreBinary()
	dir, err := m.DataDir()
	if bin == "" || err != nil {
		return nil
	}
	tmp, err := os.MkdirTemp("", "hatch-vpn-check")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	file := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(file, content, 0o600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, bin, "-t", "-d", dir, "-f", file).CombinedOutput()
	if err == nil {
		return nil
	}
	var msgs []string
	for _, mm := range coreErrRe.FindAllStringSubmatch(string(out), 3) {
		msgs = append(msgs, mm[1])
	}
	if len(msgs) == 0 {
		msgs = append(msgs, strings.TrimSpace(lastLine(string(out))))
	}
	return fmt.Errorf("Clash 内核校验未通过：%s", strings.Join(msgs, "；"))
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return lines[len(lines)-1]
}

// ─── Script runner (JavaScriptCore via osascript) ────────────────────────────

const scriptRunner = `ObjC.import('Foundation');
function readFile(p) { return $.NSString.stringWithContentsOfFileEncodingError(p, $.NSUTF8StringEncoding, null).js; }
function run(argv) {
  var config = JSON.parse(readFile(argv[0]));
  var src = readFile(argv[1]);
  var console = { log: function () {}, warn: function () {}, error: function () {}, info: function () {} };
  var main = new Function('console', src + '\n;return typeof main === "function" ? main : null;')(console);
  if (!main) throw new Error('脚本中没有找到 main 函数');
  var out = main(config, argv[2] || '');
  if (out === null || typeof out !== 'object' || Array.isArray(out)) throw new Error('main 必须返回配置对象 config');
  return JSON.stringify(out);
}`

var osaErrRe = regexp.MustCompile(`execution error: (?:Error: )?(.*?) \(-\d+\)`)

// runScript executes a Clash Verge script's main(config) on cfg the same way
// Verge does, so a broken or no-op script is caught before it is saved.
func runScript(code string, cfg map[string]any) (map[string]any, error) {
	tmp, err := os.MkdirTemp("", "hatch-vpn-script")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{"runner.js": []byte(scriptRunner), "config.json": cfgJSON, "script.js": []byte(code)}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(tmp, name), data, 0o600); err != nil {
			return nil, err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "osascript", "-l", "JavaScript", filepath.Join(tmp, "runner.js"),
		filepath.Join(tmp, "config.json"), filepath.Join(tmp, "script.js"), "hatch")
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if mm := osaErrRe.FindStringSubmatch(msg); mm != nil {
			msg = mm[1]
		}
		return nil, fmt.Errorf("脚本运行出错：%s", msg)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout.String())), &out); err != nil {
		return nil, fmt.Errorf("脚本返回的不是有效配置: %w", err)
	}
	return normalizeNumbers(out).(map[string]any), nil
}
