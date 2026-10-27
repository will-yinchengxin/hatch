package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ApplyItem previews one target location of an apply.
type ApplyItem struct {
	Location    Location `json:"location"`
	LocationKey string   `json:"locationKey"`
	Label       string   `json:"label"`
	File        string   `json:"file"`
	Action      string   `json:"action"` // add | update | unchanged | error
	Before      string   `json:"before,omitempty"`
	After       string   `json:"after,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// ApplyPlan is the dry run shown before writing anything.
type ApplyPlan struct {
	Name     string      `json:"name"`
	Items    []ApplyItem `json:"items"`
	CanApply bool        `json:"canApply"`
}

// ApplyRequest writes one spec to several locations. SourceClient says which
// client's dialect Spec.Extra is in; Extra is only written to that client.
type ApplyRequest struct {
	Name         string     `json:"name"`
	Spec         Spec       `json:"spec"`
	Targets      []Location `json:"targets"`
	SourceClient string     `json:"sourceClient,omitempty"`
	Origin       string     `json:"origin,omitempty"`
}

// rendered is what will be written at one location.
type rendered struct {
	loc     Location
	file    string
	before  string
	after   string
	body    map[string]any // Claude entry
	content string         // Codex full file content
	changed bool
}

func (r ApplyRequest) validate() error {
	if !ValidName(r.Name) {
		return fmt.Errorf("名称只能包含字母、数字、-、_、.，且不超过 64 个字符")
	}
	if len(r.Targets) == 0 {
		return fmt.Errorf("请至少选择一个写入位置")
	}
	return r.Spec.Normalize().Validate()
}

func render(r ApplyRequest, loc Location) (rendered, error) {
	out := rendered{loc: loc}
	if !loc.Writable() {
		return out, fmt.Errorf("%s 为只读位置", loc.Label())
	}
	if loc.Project != "" {
		if fi, err := os.Stat(loc.Project); err != nil || !fi.IsDir() {
			return out, fmt.Errorf("项目目录不存在: %s", loc.Project)
		}
	}
	spec := r.Spec.Normalize()
	keep := r.SourceClient == "" || r.SourceClient == loc.Client
	switch loc.Client {
	case ClientClaude:
		out.file = ClaudeFile(loc)
		existing, ok := claudeEntry(loc, r.Name)
		if ok {
			out.before = prettyJSON(r.Name, existing)
			spec.Extra = mergeExtra(SpecFromClaude(existing).Extra, spec.Extra, keep)
			keep = true
		}
		out.body = ClaudeJSON(spec, keep)
		out.after = prettyJSON(r.Name, out.body)
		out.changed = !ok || canonicalJSON(existing) != canonicalJSON(out.body)
	case ClientCodex:
		path, err := CodexConfigFile(loc)
		if err != nil {
			return out, err
		}
		out.file = path
		content := ""
		if data, err := os.ReadFile(path); err == nil {
			content = string(data)
		}
		servers := map[string]map[string]any{}
		if content != "" {
			if servers, err = readCodexServers(path); err != nil {
				return out, err
			}
		}
		enabled := true
		if existing, ok := servers[r.Name]; ok {
			var prev Spec
			prev, enabled = SpecFromCodex(existing)
			out.before = RenderCodexBlock(r.Name, existing)
			spec.Extra = mergeExtra(prev.Extra, spec.Extra, keep)
			keep = true
		}
		table, err := CodexTable(spec, enabled, keep)
		if err != nil {
			return out, err
		}
		out.after = RenderCodexBlock(r.Name, table)
		out.changed = out.before != out.after
		if out.content, err = CodexUpsert(content, r.Name, table); err != nil {
			return out, err
		}
	default:
		return out, fmt.Errorf("未知客户端: %s", loc.Client)
	}
	return out, nil
}

// mergeExtra keeps the target's existing unknown keys and overlays the new
// spec's keys when they are in the target's dialect.
func mergeExtra(existing, incoming map[string]any, incomingSameClient bool) map[string]any {
	out := map[string]any{}
	for k, v := range existing {
		out[k] = v
	}
	if incomingSameClient {
		for k, v := range incoming {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func canonicalJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// Plan dry-runs r.
func Plan(r ApplyRequest) ApplyPlan {
	plan := ApplyPlan{Name: r.Name, CanApply: true}
	if err := r.validate(); err != nil {
		plan.CanApply = false
		plan.Items = append(plan.Items, ApplyItem{Action: "error", Error: err.Error()})
		return plan
	}
	for _, loc := range r.Targets {
		item := ApplyItem{Location: loc, LocationKey: loc.Key(), Label: loc.Label()}
		res, err := render(r, loc)
		item.File = res.file
		switch {
		case err != nil:
			item.Action, item.Error = "error", err.Error()
			plan.CanApply = false
		case res.before == "":
			item.Action, item.After = "add", res.after
		case !res.changed:
			item.Action, item.Before, item.After = "unchanged", res.before, res.after
		default:
			item.Action, item.Before, item.After = "update", res.before, res.after
		}
		plan.Items = append(plan.Items, item)
	}
	return plan
}

// Apply writes r to every target, stopping at the first failure.
func Apply(r ApplyRequest) error {
	if err := r.validate(); err != nil {
		return err
	}
	for _, loc := range r.Targets {
		res, err := render(r, loc)
		if err != nil {
			return fmt.Errorf("%s: %w", loc.Label(), err)
		}
		if !res.changed {
			continue
		}
		switch loc.Client {
		case ClientClaude:
			err = ClaudeWrite(loc, r.Name, res.body)
		case ClientCodex:
			err = writeCodexFile(res.file, res.content)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", loc.Label(), err)
		}
	}
	return nil
}

// SetEnabled toggles name at loc without losing its definition. Codex flips
// `enabled`; Claude has no file-level switch, so a disabled entry is removed
// and its raw JSON returned for the caller to park (and passed back to
// re-enable). The returned string is the new parked raw ("" when none).
func SetEnabled(name string, loc Location, enabled bool, parked string) (string, error) {
	if !loc.Writable() {
		return parked, fmt.Errorf("%s 为只读位置", loc.Label())
	}
	switch loc.Client {
	case ClientCodex:
		path, err := CodexConfigFile(loc)
		if err != nil {
			return parked, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return parked, err
		}
		content, err := CodexSetEnabled(string(data), name, enabled)
		if err != nil {
			return parked, err
		}
		return "", writeCodexFile(path, content)
	case ClientClaude:
		existing, exists := claudeEntry(loc, name)
		if !enabled {
			if !exists {
				return parked, nil
			}
			raw, _ := json.Marshal(existing)
			if err := ClaudeRemove(loc, name); err != nil {
				return parked, err
			}
			return string(raw), nil
		}
		if exists {
			return "", nil
		}
		if parked == "" {
			return "", fmt.Errorf("没有可恢复的停用配置")
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(parked), &body); err != nil {
			return parked, fmt.Errorf("停用暂存已损坏: %w", err)
		}
		if err := ClaudeWrite(loc, name, body); err != nil {
			return parked, err
		}
		return "", nil
	}
	return parked, fmt.Errorf("未知客户端: %s", loc.Client)
}

// Remove deletes name at loc.
func Remove(name string, loc Location) error {
	if !loc.Writable() {
		return fmt.Errorf("%s 为只读位置", loc.Label())
	}
	switch loc.Client {
	case ClientClaude:
		return ClaudeRemove(loc, name)
	case ClientCodex:
		path, err := CodexConfigFile(loc)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		return writeCodexFile(path, CodexRemove(string(data), name))
	}
	return fmt.Errorf("未知客户端: %s", loc.Client)
}

// SetCodexDisabledTools writes disabled_tools for a Codex entry (empty clears it).
func SetCodexDisabledTools(name string, loc Location, disabled []string) error {
	if loc.Client != ClientCodex || !loc.Writable() {
		return fmt.Errorf("仅 Codex 支持按工具开关")
	}
	path, err := CodexConfigFile(loc)
	if err != nil {
		return err
	}
	servers, err := readCodexServers(path)
	if err != nil {
		return err
	}
	table, ok := servers[name]
	if !ok {
		return fmt.Errorf("Codex 配置中没有 %s", name)
	}
	clean := []string{}
	for _, t := range disabled {
		if t = strings.TrimSpace(t); t != "" {
			clean = append(clean, t)
		}
	}
	sort.Strings(clean)
	if len(clean) == 0 {
		delete(table, "disabled_tools")
	} else {
		table["disabled_tools"] = clean
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := CodexUpsert(string(data), name, table)
	if err != nil {
		return err
	}
	return writeCodexFile(path, content)
}
