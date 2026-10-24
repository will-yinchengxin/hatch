package vpn

import (
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ─── Subscriptions (订阅) ────────────────────────────────────────────────────
//
// A readable summary of every subscription in profiles.yaml — traffic,
// expiry, nodes, groups and which enhancements are customised — instead of
// raw profile files. Raw editing lives in the advanced config page.

type GroupSummary struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Count int    `json:"count"`
	Now   string `json:"now"`
}

type RegionCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type EnhanceSummary struct {
	Kind       string `json:"kind"` // merge | script | rules | proxies | groups
	Label      string `json:"label"`
	File       string `json:"file"`
	Customized bool   `json:"customized"`
	Summary    string `json:"summary"`
	Desc       string `json:"desc"` // what it is and what it does, for ordinary users
}

type Subscription struct {
	UID            string           `json:"uid"`
	Name           string           `json:"name"`
	Type           string           `json:"type"` // remote | local
	File           string           `json:"file"`
	URLHost        string           `json:"urlHost"`
	IsCurrent      bool             `json:"isCurrent"`
	Updated        int64            `json:"updated"`
	Upload         int64            `json:"upload"`
	Download       int64            `json:"download"`
	Total          int64            `json:"total"`
	Expire         int64            `json:"expire"`
	UpdateInterval int              `json:"updateInterval"` // minutes
	AutoUpdate     bool             `json:"autoUpdate"`
	NodeCount      int              `json:"nodeCount"`
	RuleCount      int              `json:"ruleCount"`
	Groups         []GroupSummary   `json:"groups"`
	Regions        []RegionCount    `json:"regions"`
	Enhancements   []EnhanceSummary `json:"enhancements"`
	Error          string           `json:"error,omitempty"`
}

type SubscriptionsView struct {
	Subscriptions []Subscription   `json:"subscriptions"`
	Global        []EnhanceSummary `json:"global"` // global Merge / Script (apply to every subscription)
	Files         []ProfileFile    `json:"files"`  // every profile file, for the advanced editor
}

// ProfileFile labels one profile file for the advanced raw editor.
type ProfileFile struct {
	File   string `json:"file"`
	Label  string `json:"label"`
	Type   string `json:"type"`
	Remote bool   `json:"remote"` // overwritten on every subscription update
}

var enhanceLabels = map[string]string{
	"merge": "覆写 (Merge)", "script": "脚本 (Script)", "rules": "规则增强",
	"proxies": "节点增强", "groups": "分组增强",
}

func (m Manager) Subscriptions() (SubscriptionsView, error) {
	dir, meta, err := m.readProfilesMeta()
	if err != nil {
		return SubscriptionsView{}, err
	}
	view := SubscriptionsView{Subscriptions: []Subscription{}, Global: []EnhanceSummary{}, Files: []ProfileFile{}}
	owner := map[string]string{} // enhancement uid → subscription name
	for _, it := range meta.Items {
		if it.Type != "remote" && it.Type != "local" {
			continue
		}
		name := profileLabel(it)
		for _, uid := range []string{it.Option.Merge, it.Option.Script, it.Option.Rules, it.Option.Proxies, it.Option.Groups} {
			if uid != "" {
				owner[uid] = name
			}
		}
	}

	for _, it := range meta.Items {
		switch it.Type {
		case "remote", "local":
			view.Subscriptions = append(view.Subscriptions, m.summarize(dir, meta, it))
			view.Files = append(view.Files, ProfileFile{File: it.File, Label: "订阅 · " + profileLabel(it), Type: it.Type, Remote: it.Type == "remote"})
		default:
			label := enhanceLabels[it.Type]
			if label == "" {
				label = it.Type
			}
			if o, ok := owner[it.UID]; ok {
				view.Files = append(view.Files, ProfileFile{File: it.File, Label: o + " · " + label, Type: it.Type})
			} else {
				view.Files = append(view.Files, ProfileFile{File: it.File, Label: "全局 · " + label, Type: it.Type})
				if it.Type == "merge" || it.Type == "script" {
					view.Global = append(view.Global, summarizeEnhance(dir, it))
				}
			}
		}
	}
	return view, nil
}

func profileLabel(it profileItem) string {
	if it.Name != "" {
		return it.Name
	}
	return it.File
}

func (m Manager) summarize(dir string, meta profilesMeta, it profileItem) Subscription {
	s := Subscription{
		UID: it.UID, Name: profileLabel(it), Type: it.Type, File: it.File,
		IsCurrent: it.UID == meta.Current, Updated: it.Updated,
		Upload: it.Extra.Upload, Download: it.Extra.Download, Total: it.Extra.Total, Expire: it.Extra.Expire,
		UpdateInterval: it.Option.UpdateInterval,
		AutoUpdate:     it.Option.AllowAutoUpdate == nil || *it.Option.AllowAutoUpdate,
		Groups:         []GroupSummary{}, Regions: []RegionCount{}, Enhancements: []EnhanceSummary{},
	}
	if u, err := url.Parse(it.URL); err == nil && u.Host != "" {
		s.URLHost = u.Host
	}
	if it.Type != "remote" {
		s.AutoUpdate = false
	}

	data, err := os.ReadFile(filepath.Join(dir, "profiles", it.File))
	if err != nil {
		s.Error = "读取订阅文件失败"
	} else {
		var raw struct {
			Proxies []struct {
				Name string `yaml:"name"`
			} `yaml:"proxies"`
			ProxyGroups []struct {
				Name    string   `yaml:"name"`
				Type    string   `yaml:"type"`
				Proxies []string `yaml:"proxies"`
			} `yaml:"proxy-groups"`
			Rules []string `yaml:"rules"`
		}
		if yaml.Unmarshal(data, &raw) != nil {
			s.Error = "订阅文件不是有效的 YAML"
		}
		s.NodeCount = len(raw.Proxies)
		s.RuleCount = len(raw.Rules)
		selected := map[string]string{}
		for _, sel := range it.Selected {
			selected[sel.Name] = sel.Now
		}
		for _, g := range raw.ProxyGroups {
			s.Groups = append(s.Groups, GroupSummary{Name: g.Name, Type: g.Type, Count: len(g.Proxies), Now: selected[g.Name]})
		}
		names := make([]string, 0, len(raw.Proxies))
		for _, p := range raw.Proxies {
			names = append(names, p.Name)
		}
		s.Regions = regionCounts(names)
	}

	for _, e := range []struct{ kind, uid string }{
		{"rules", it.Option.Rules}, {"merge", it.Option.Merge}, {"script", it.Option.Script},
		{"proxies", it.Option.Proxies}, {"groups", it.Option.Groups},
	} {
		if e.uid == "" {
			continue
		}
		if enh, ok := meta.find(e.uid); ok {
			s.Enhancements = append(s.Enhancements, summarizeEnhance(dir, enh))
		}
	}
	return s
}

// summarizeEnhance says whether an enhancement file holds anything beyond
// Clash Verge's empty template.
func summarizeEnhance(dir string, it profileItem) EnhanceSummary {
	out := EnhanceSummary{Kind: it.Type, Label: enhanceLabels[it.Type], File: it.File, Desc: enhanceInfo[it.Type].Desc}
	data, err := os.ReadFile(filepath.Join(dir, "profiles", it.File))
	if err != nil {
		out.Summary = "文件不存在"
		return out
	}
	switch it.Type {
	case "script":
		code := regexp.MustCompile(`//[^\n]*`).ReplaceAllString(string(data), "")
		code = strings.Join(strings.Fields(code), "")
		out.Customized = code != "" && code != "functionmain(config,profileName){returnconfig;}" && code != "functionmain(config){returnconfig;}"
		out.Summary = map[bool]string{true: "已自定义脚本", false: "未修改"}[out.Customized]
	case "rules", "proxies", "groups":
		var seq struct {
			Prepend []interface{} `yaml:"prepend"`
			Append  []interface{} `yaml:"append"`
			Delete  []interface{} `yaml:"delete"`
		}
		_ = yaml.Unmarshal(data, &seq)
		n := len(seq.Prepend) + len(seq.Append) + len(seq.Delete)
		out.Customized = n > 0
		out.Summary = "未修改"
		if n > 0 {
			parts := []string{}
			if len(seq.Prepend) > 0 {
				parts = append(parts, "前置 "+strconv.Itoa(len(seq.Prepend)))
			}
			if len(seq.Append) > 0 {
				parts = append(parts, "追加 "+strconv.Itoa(len(seq.Append)))
			}
			if len(seq.Delete) > 0 {
				parts = append(parts, "删除 "+strconv.Itoa(len(seq.Delete)))
			}
			out.Summary = strings.Join(parts, " · ")
		}
	default: // merge
		var mp map[string]interface{}
		_ = yaml.Unmarshal(data, &mp)
		keys := make([]string, 0, len(mp))
		for k := range mp {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out.Customized = len(keys) > 0
		out.Summary = "未修改"
		if len(keys) > 0 {
			out.Summary = "覆盖 " + strings.Join(keys, "、")
		}
	}
	return out
}

var trailingNumRe = regexp.MustCompile(`[\s\-_|·]*\d+[a-zA-Z]?$`)

// regionCounts groups node names by their label minus the trailing index
// ("美国 03" → "美国"), largest first, capped at 12.
func regionCounts(names []string) []RegionCount {
	counts := map[string]int{}
	order := []string{}
	for _, n := range names {
		key := strings.TrimSpace(trailingNumRe.ReplaceAllString(strings.TrimSpace(n), ""))
		if key == "" {
			key = n
		}
		if _, ok := counts[key]; !ok {
			order = append(order, key)
		}
		counts[key]++
	}
	out := make([]RegionCount, 0, len(order))
	for _, k := range order {
		out = append(out, RegionCount{Name: k, Count: counts[k]})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	if len(out) > 12 {
		out = out[:12]
	}
	return out
}

// SubscriptionURL returns the full subscription link (contains the token) for copying.
func (m Manager) SubscriptionURL(uid string) (string, error) {
	_, meta, err := m.readProfilesMeta()
	if err != nil {
		return "", err
	}
	if it, ok := meta.find(uid); ok && it.URL != "" {
		return it.URL, nil
	}
	return "", errors.New("该订阅没有链接")
}

// OpenApp brings Clash Verge Rev to the front (switching / updating
// subscriptions is done there so Verge regenerates its runtime).
func OpenApp() error {
	return exec.Command("open", "-b", clashVergeAppID).Run()
}
