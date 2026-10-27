package scanner

import (
	"testing"

	"hatch/core/config"
)

func TestParseHatchAnnotations_ArgsAndPresets(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "report.sh", `#!/bin/zsh
# @hatch.name: 日报
# @hatch.arg: 日期 | --date | today | required
# @hatch.arg: 输出目录
# @hatch.preset: 本周 | --date week
`)
	meta, found := parseHatchAnnotations(path)
	if !found {
		t.Fatal("annotations not found")
	}
	want := []config.ArgDef{
		{Name: "日期", Flag: "--date", Default: "today", Required: true},
		{Name: "输出目录"},
	}
	if len(meta.Args) != len(want) {
		t.Fatalf("args = %+v", meta.Args)
	}
	for i := range want {
		if meta.Args[i] != want[i] {
			t.Errorf("arg %d = %+v, want %+v", i, meta.Args[i], want[i])
		}
	}
	if len(meta.Presets) != 1 || meta.Presets[0].Name != "本周" || meta.Presets[0].Args != "--date week" {
		t.Errorf("presets = %+v", meta.Presets)
	}
}

func TestMergeOverride_KeepsRunSettings(t *testing.T) {
	base := config.ScriptMeta{Path: "/x/a.sh", Name: "a", Schedule: "@daily 09:00"}
	ov := config.ScriptMeta{
		Path:         "/x/a.sh",
		Env:          map[string]string{"TOKEN": "t"},
		NotifyOnDone: true,
		Requires:     []string{"jq"},
		HealthURL:    "http://localhost/healthz",
		PreScripts:   []string{"/x/pre.sh"},
		Favorite:     true,
	}
	got := mergeOverride(base, ov)
	if got.Env["TOKEN"] != "t" || !got.NotifyOnDone || got.Requires[0] != "jq" ||
		got.HealthURL == "" || got.PreScripts[0] != "/x/pre.sh" || !got.Favorite {
		t.Fatalf("override fields lost: %+v", got)
	}
	if got.Schedule != "@daily 09:00" || got.Name != "a" {
		t.Fatalf("empty override fields must keep scanned values: %+v", got)
	}
}

func TestHeaderComment(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"a.sh": "#!/bin/zsh\n# -*- coding: utf-8 -*-\n# ------\n# 备份笔记到 iCloud\nset -e\n",
		"b.py": "#!/usr/bin/env python3\n\"\"\"Sync notes.\"\"\"\nimport os\n",
		"c.sh": "#!/bin/zsh\necho hi\n# not a header\n",
		"d.sh": "#!/bin/zsh\n# @hatch.name: x\n# Real desc\n",
	}
	want := map[string]string{"a.sh": "备份笔记到 iCloud", "b.py": "Sync notes.", "c.sh": "", "d.sh": "Real desc"}
	for name, body := range cases {
		if got := headerComment(writeScript(t, dir, name, body)); got != want[name] {
			t.Errorf("%s: got %q, want %q", name, got, want[name])
		}
	}
}
