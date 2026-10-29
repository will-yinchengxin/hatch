package skills

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestImportFromSkipsNonSkillDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	src := filepath.Join(home, ".claude", "skills")
	write := func(rel, body string) {
		p := filepath.Join(src, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("real/SKILL.md", "---\nname: real\n---\n")
	write(".trash/old/SKILL.md", "---\nname: old\n---\n")
	write(".system/sys/SKILL.md", "---\nname: sys\n---\n")
	write("synced/bucket/data.json", "{}")
	write(".DS_Store", "")

	st := Store{Dir: filepath.Join(home, "master")}
	if err := os.MkdirAll(st.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := st.ImportFrom(TargetClaude)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Imported) != 1 || r.Imported[0] != "real" {
		t.Fatalf("imported = %v, want [real]", r.Imported)
	}
	if len(r.Skipped) != 0 || len(r.Conflicts) != 0 {
		t.Fatalf("skipped = %v, conflicts = %v, want none", r.Skipped, r.Conflicts)
	}
	entries, _ := os.ReadDir(st.Dir)
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if len(names) != 1 || names[0] != "real" {
		t.Fatalf("master dir = %v, want [real]", names)
	}
}
