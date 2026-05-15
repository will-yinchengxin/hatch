package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"hatch/internal/config"
)

// writeScript creates an executable script file in dir with the given content.
func writeScript(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// --- parser tests ---

func TestParseHatchAnnotations_Full(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "redis_sys.sh", `#!/bin/bash
# @hatch.name: Redis 系统管理
# @hatch.desc: 启停本地 Redis 服务
# @hatch.type: service        # inline comment
# @hatch.commands: start, stop, restart, status
# @hatch.pidfile: /tmp/hatch/redis.pid
# @hatch.tags: db, local
`)
	path := filepath.Join(dir, "redis_sys.sh")
	meta, found := parseHatchAnnotations(path)
	if !found {
		t.Fatal("expected annotations to be found")
	}
	if meta.Name != "Redis 系统管理" {
		t.Errorf("name: got %q", meta.Name)
	}
	if meta.Desc != "启停本地 Redis 服务" {
		t.Errorf("desc: got %q", meta.Desc)
	}
	if meta.Type != config.TypeService {
		t.Errorf("type: got %q", meta.Type)
	}
	if len(meta.Commands) != 4 {
		t.Errorf("commands: got %v", meta.Commands)
	}
	if meta.PidFile != "/tmp/hatch/redis.pid" {
		t.Errorf("pidfile: got %q", meta.PidFile)
	}
	if len(meta.Tags) != 2 {
		t.Errorf("tags: got %v", meta.Tags)
	}
}

func TestParseHatchAnnotations_None(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "plain.sh", "#!/bin/bash\necho hello\n")
	_, found := parseHatchAnnotations(filepath.Join(dir, "plain.sh"))
	if found {
		t.Error("should not find annotations in plain script")
	}
}

func TestParseHatchAnnotations_OneshotDefault(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "s.sh", "#!/bin/bash\n# @hatch.name: Simple\n")
	meta, _ := parseHatchAnnotations(filepath.Join(dir, "s.sh"))
	// type not set → stays zero value; scanner fills in default
	if meta.Type != "" && meta.Type != config.TypeOneshot {
		t.Errorf("unexpected type %q", meta.Type)
	}
}

// --- language detection tests ---

func TestDetectLanguage_Extension(t *testing.T) {
	cases := []struct{ name, want string }{
		{"foo.py", "python"},
		{"bar.sh", "shell"},
		{"baz.rb", "ruby"},
	}
	for _, c := range cases {
		if got := detectLanguage(c.name); got != c.want {
			t.Errorf("detectLanguage(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestDetectLanguage_Shebang(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name, shebang, want string
	}{
		{"myscript", "#!/usr/bin/env python3", "python"},
		{"runner", "#!/bin/bash", "shell"},
		{"tool", "#!/usr/bin/env node", "node"},
	}
	for _, c := range cases {
		path := writeScript(t, dir, c.name, c.shebang+"\necho hi\n")
		if got := detectLanguage(path); got != c.want {
			t.Errorf("detectLanguage(%q) shebang %q = %q, want %q", c.name, c.shebang, got, c.want)
		}
	}
}

// --- scanner tests ---

func TestScanDir_FindsExecutables(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "tool.sh", "#!/bin/bash\necho hi\n")
	writeScript(t, dir, "helper.py", "#!/usr/bin/env python3\nprint('hi')\n")
	// non-executable, no script extension → should be skipped
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0o644); err != nil {
		t.Fatal(err)
	}

	scripts, err := ScanDir(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 2 {
		t.Errorf("expected 2 scripts, got %d: %v", len(scripts), scripts)
	}
}

func TestScanDir_AnnotationsWin(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "svc.sh", `#!/bin/bash
# @hatch.name: My Service
# @hatch.type: service
# @hatch.commands: start, stop
`)
	scripts, err := ScanDir(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 1 {
		t.Fatalf("expected 1, got %d", len(scripts))
	}
	s := scripts[0]
	if s.Name != "My Service" {
		t.Errorf("name: %q", s.Name)
	}
	if s.Type != config.TypeService {
		t.Errorf("type: %q", s.Type)
	}
	if len(s.Commands) != 2 {
		t.Errorf("commands: %v", s.Commands)
	}
}

func TestScanDir_OverrideApplied(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "plain.sh", "#!/bin/bash\necho hi\n")
	overrides := map[string]config.ScriptMeta{
		path: {
			Name: "Overridden",
			Desc: "custom desc",
			Args: []config.ArgDef{},
			Presets: []config.CommandPreset{
				{Name: "Today", Args: "2026-05-15"},
			},
		},
	}
	scripts, err := ScanDir(dir, overrides)
	if err != nil {
		t.Fatal(err)
	}
	if scripts[0].Name != "Overridden" {
		t.Errorf("override not applied, got name %q", scripts[0].Name)
	}
	if scripts[0].Desc != "custom desc" {
		t.Errorf("override desc not applied, got %q", scripts[0].Desc)
	}
	if scripts[0].Args == nil {
		t.Errorf("override empty args should be preserved as non-nil slice")
	}
	if len(scripts[0].Presets) != 1 {
		t.Fatalf("override presets not applied, got %#v", scripts[0].Presets)
	}
	if scripts[0].Presets[0].Name != "Today" || scripts[0].Presets[0].Args != "2026-05-15" {
		t.Errorf("unexpected preset: %#v", scripts[0].Presets[0])
	}
}

func TestScanDir_SkipsHidden(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, ".hidden.sh", "#!/bin/bash\necho hi\n")
	scripts, err := ScanDir(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 0 {
		t.Errorf("hidden file should be skipped, got %v", scripts)
	}
}

func TestScanDir_DefaultsForPlainScript(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "backup.sh", "#!/bin/bash\necho backup\n")
	scripts, _ := ScanDir(dir, nil)
	s := scripts[0]
	if s.Type != config.TypeOneshot {
		t.Errorf("default type should be oneshot, got %q", s.Type)
	}
	if s.Language != "shell" {
		t.Errorf("language should be shell, got %q", s.Language)
	}
	if s.Name != "backup" {
		t.Errorf("name should strip ext, got %q", s.Name)
	}
}

func TestDetectServicePattern_CaseStatement(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "redis.sh", `#!/bin/bash
case "$1" in
  start)
    echo "starting"
    ;;
  stop)
    echo "stopping"
    ;;
  restart)
    echo "restarting"
    ;;
  status)
    echo "status"
    ;;
  *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
esac
`)
	scripts, err := ScanDir(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(scripts))
	}
	s := scripts[0]
	if s.Type != config.TypeService {
		t.Errorf("expected service type, got %q", s.Type)
	}
	if !containsStr(s.Commands, "start") || !containsStr(s.Commands, "stop") {
		t.Errorf("expected start/stop in commands, got %v", s.Commands)
	}
}

func TestDetectServicePattern_UsageHint(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "svc.sh", `#!/bin/bash
usage() { echo "Usage: $0 {start|stop|reload}"; exit 1; }
[ $# -eq 0 ] && usage
`)
	scripts, _ := ScanDir(dir, nil)
	if len(scripts) == 1 && scripts[0].Type != config.TypeService {
		t.Errorf("usage hint should trigger service detection, got type %q", scripts[0].Type)
	}
}

func TestDetectServicePattern_PlainScript(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "backup.sh", "#!/bin/bash\ntar czf backup.tar.gz /data\n")
	scripts, _ := ScanDir(dir, nil)
	if len(scripts) == 1 && scripts[0].Type != config.TypeOneshot {
		t.Errorf("plain script should be oneshot, got %q", scripts[0].Type)
	}
}

func TestScanDir_ServiceAutoFillsPidFile(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "daemon.sh", "#!/bin/bash\n# @hatch.type: service\n")
	scripts, _ := ScanDir(dir, nil)
	if scripts[0].PidFile == "" {
		t.Error("service type should auto-fill PidFile")
	}
}
