package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"hatch/internal/config"
)

func newSkillTestApp(t *testing.T) (*App, string, string) {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "home")
	master := filepath.Join(root, "master")
	t.Setenv("HOME", home)
	a := &App{cfg: &config.Config{SkillsDir: master}}
	return a, master, home
}

func writeSkillFile(t *testing.T, base, name, rel, content string, mod time.Time) {
	t.Helper()
	path := filepath.Join(base, name, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mod, mod); err != nil {
		t.Fatal(err)
	}
}

func readSkillContent(t *testing.T, base, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(base, name, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestImportSkillsFromImportsNewSkill(t *testing.T) {
	a, master, home := newSkillTestApp(t)
	codex := filepath.Join(home, ".codex", "skills")
	writeSkillFile(t, codex, "demo", "SKILL.md", "codex skill", time.Unix(100, 0))

	result, err := a.ImportSkillsFrom(skillTargetCodex)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Imported) != 1 || result.Imported[0] != "demo" {
		t.Fatalf("imported = %#v", result.Imported)
	}
	if len(result.Skipped) != 0 || len(result.Conflicts) != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if got := readSkillContent(t, master, "demo"); got != "codex skill" {
		t.Fatalf("master content = %q", got)
	}
}

func TestImportSkillsFromSkipsSameFingerprint(t *testing.T) {
	a, master, home := newSkillTestApp(t)
	codex := filepath.Join(home, ".codex", "skills")
	writeSkillFile(t, master, "demo", "SKILL.md", "same skill", time.Unix(100, 0))
	writeSkillFile(t, codex, "demo", "SKILL.md", "same skill", time.Unix(200, 0))

	result, err := a.ImportSkillsFrom(skillTargetCodex)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "demo" {
		t.Fatalf("skipped = %#v", result.Skipped)
	}
	if len(result.Imported) != 0 || len(result.Conflicts) != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if got := readSkillContent(t, master, "demo"); got != "same skill" {
		t.Fatalf("master content = %q", got)
	}
}

func TestImportSkillsFromConflictsOnDifferentSkillMD(t *testing.T) {
	a, master, home := newSkillTestApp(t)
	codex := filepath.Join(home, ".codex", "skills")
	writeSkillFile(t, master, "demo", "SKILL.md", "master skill", time.Unix(100, 0))
	writeSkillFile(t, codex, "demo", "SKILL.md", "codex skill", time.Unix(200, 0))

	result, err := a.ImportSkillsFrom(skillTargetCodex)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("conflicts = %#v", result.Conflicts)
	}
	c := result.Conflicts[0]
	if c.Name != "demo" || c.Source != skillTargetCodex || c.MasterModTime != 100 || c.SourceModTime != 200 {
		t.Fatalf("unexpected conflict: %#v", c)
	}
	if got := readSkillContent(t, master, "demo"); got != "master skill" {
		t.Fatalf("master should not be overwritten, got %q", got)
	}
}

func TestImportSkillsFromConflictsOnNestedFileDifference(t *testing.T) {
	a, master, home := newSkillTestApp(t)
	claude := filepath.Join(home, ".claude", "skills")
	writeSkillFile(t, master, "demo", "SKILL.md", "same skill", time.Unix(100, 0))
	writeSkillFile(t, master, "demo", "references/a.md", "master ref", time.Unix(110, 0))
	writeSkillFile(t, claude, "demo", "SKILL.md", "same skill", time.Unix(100, 0))
	writeSkillFile(t, claude, "demo", "references/a.md", "claude ref", time.Unix(220, 0))

	result, err := a.ImportSkillsFrom(skillTargetClaude)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("conflicts = %#v", result.Conflicts)
	}
	if result.Conflicts[0].MasterModTime != 110 || result.Conflicts[0].SourceModTime != 220 {
		t.Fatalf("mtime should use latest nested file: %#v", result.Conflicts[0])
	}
}

func TestResolveSkillImportConflictUsesSourceWhenRequested(t *testing.T) {
	a, master, home := newSkillTestApp(t)
	codex := filepath.Join(home, ".codex", "skills")
	writeSkillFile(t, master, "demo", "SKILL.md", "master skill", time.Unix(100, 0))
	writeSkillFile(t, master, "demo", "references/old.md", "old", time.Unix(100, 0))
	writeSkillFile(t, codex, "demo", "SKILL.md", "codex skill", time.Unix(200, 0))

	if err := a.ResolveSkillImportConflict("demo", skillTargetCodex, true); err != nil {
		t.Fatal(err)
	}
	if got := readSkillContent(t, master, "demo"); got != "codex skill" {
		t.Fatalf("master content = %q", got)
	}
	if _, err := os.Stat(filepath.Join(master, "demo", "references", "old.md")); !os.IsNotExist(err) {
		t.Fatalf("stale master file should be removed, stat err = %v", err)
	}
}

func TestResolveSkillImportConflictKeepsMasterWhenRequested(t *testing.T) {
	a, master, home := newSkillTestApp(t)
	codex := filepath.Join(home, ".codex", "skills")
	writeSkillFile(t, master, "demo", "SKILL.md", "master skill", time.Unix(100, 0))
	writeSkillFile(t, codex, "demo", "SKILL.md", "codex skill", time.Unix(200, 0))

	if err := a.ResolveSkillImportConflict("demo", skillTargetCodex, false); err != nil {
		t.Fatal(err)
	}
	if got := readSkillContent(t, master, "demo"); got != "master skill" {
		t.Fatalf("master content = %q", got)
	}
}

func TestDeleteSkillPreservesSyncedTargets(t *testing.T) {
	a, master, home := newSkillTestApp(t)
	claude := filepath.Join(home, ".claude", "skills")
	codex := filepath.Join(home, ".codex", "skills")
	writeSkillFile(t, master, "demo", "SKILL.md", "master skill", time.Unix(100, 0))
	writeSkillFile(t, claude, "demo", "SKILL.md", "claude skill", time.Unix(200, 0))
	writeSkillFile(t, codex, "demo", "SKILL.md", "codex skill", time.Unix(300, 0))

	if err := a.DeleteSkill("demo"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(master, "demo")); !os.IsNotExist(err) {
		t.Fatalf("master skill should be removed, stat err = %v", err)
	}
	if got := readSkillContent(t, claude, "demo"); got != "claude skill" {
		t.Fatalf("claude content = %q", got)
	}
	if got := readSkillContent(t, codex, "demo"); got != "codex skill" {
		t.Fatalf("codex content = %q", got)
	}
}

func TestDeleteSkillEverywhereRemovesMasterAndSyncedTargets(t *testing.T) {
	a, master, home := newSkillTestApp(t)
	claude := filepath.Join(home, ".claude", "skills")
	codex := filepath.Join(home, ".codex", "skills")
	writeSkillFile(t, master, "demo", "SKILL.md", "master skill", time.Unix(100, 0))
	writeSkillFile(t, claude, "demo", "SKILL.md", "claude skill", time.Unix(200, 0))
	writeSkillFile(t, codex, "demo", "SKILL.md", "codex skill", time.Unix(300, 0))

	if err := a.DeleteSkillEverywhere("demo"); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{master, claude, codex} {
		if _, err := os.Stat(filepath.Join(dir, "demo")); !os.IsNotExist(err) {
			t.Fatalf("%s/demo should be removed, stat err = %v", dir, err)
		}
	}
}
