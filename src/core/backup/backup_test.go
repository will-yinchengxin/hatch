package backup

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestImportSkipsEntriesEscapingSkillsDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := t.TempDir()
	skills := filepath.Join(dir, "skills")
	zipPath := filepath.Join(dir, "b.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range map[string]string{
		"skills/ok/SKILL.md": "ok",
		"skills/../../evil":  "evil",
	} {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte(body))
	}
	zw.Close()
	f.Close()

	n, err := Import(zipPath, skills)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("restored %d files, want 1", n)
	}
	if _, err := os.Stat(filepath.Join(dir, "..", "evil")); err == nil {
		t.Fatal("zip entry escaped the skills dir")
	}
	if _, err := os.Stat(filepath.Join(skills, "ok", "SKILL.md")); err != nil {
		t.Fatal("valid skill entry not restored")
	}
}

func TestExportImportRoundTripKeepsSkillsLayout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	skills := filepath.Join(dir, "skills")
	if err := os.MkdirAll(filepath.Join(skills, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "demo", "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Export(dir, skills)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(res.Path)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if f.Name != "skills/demo/SKILL.md" && f.Name != "config.json" && f.Name != "scripts.json" {
			t.Fatalf("unexpected entry %q", f.Name)
		}
	}
	zr.Close()

	restored := filepath.Join(dir, "restored")
	if _, err := Import(res.Path, restored); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(restored, "demo", "SKILL.md")); err != nil {
		t.Fatal("skill not restored to <skillsDir>/demo/SKILL.md")
	}
}

func TestImportAcceptsLegacyNestedSkills(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "old.zip")
	f, _ := os.Create(zipPath)
	zw := zip.NewWriter(f)
	for _, name := range []string{"skills/skills/a/SKILL.md", "skills/skills/b/SKILL.md"} {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte("x"))
	}
	zw.Close()
	f.Close()

	skills := filepath.Join(dir, "skills")
	if _, err := Import(zipPath, skills); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"a", "b"} {
		if _, err := os.Stat(filepath.Join(skills, n, "SKILL.md")); err != nil {
			t.Fatalf("legacy entry %s not restored at the top level", n)
		}
	}
	if _, err := os.Stat(filepath.Join(skills, "skills")); err == nil {
		t.Fatal("legacy zip created a nested skills/skills dir")
	}
}
