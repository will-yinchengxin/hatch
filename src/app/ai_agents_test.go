package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveCodexConfigRejectsInvalidTOML(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	a := &App{}
	if err := a.SaveCodexConfig("model = ["); err == nil {
		t.Fatal("expected invalid TOML to be rejected")
	}
}
