package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	path, err := defaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDefaultsModulesWhenMissing(t *testing.T) {
	writeConfigFile(t, `{"directories":[]}`)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Modules != DefaultModules() {
		t.Fatalf("modules = %+v, want all enabled", cfg.Modules)
	}
}

func TestLoadKeepsDefaultsForOmittedModuleKeys(t *testing.T) {
	writeConfigFile(t, `{"modules":{"vpn":false}}`)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := Modules{Workbench: true, AI: true, VPN: false, Memo: true, Host: true}
	if cfg.Modules != want {
		t.Fatalf("modules = %+v, want %+v", cfg.Modules, want)
	}
}

func TestSavePreservesModulesEditedOnDisk(t *testing.T) {
	path := writeConfigFile(t, `{"modules":{"workbench":true,"ai":true,"vpn":true,"memo":true}}`)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// User hides memo in an editor while Hatch is running...
	if err := os.WriteFile(path, []byte(`{"modules":{"memo":false}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	// ...then Hatch saves an unrelated change with its stale in-memory copy.
	cfg.AutoLaunch = true
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	var saved Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Modules.Memo || !saved.Modules.AI || !saved.AutoLaunch {
		t.Fatalf("hand edit was overwritten: %s", data)
	}
	if mods, err := LoadModules(); err != nil || mods.Memo {
		t.Fatalf("LoadModules = %+v, %v", mods, err)
	}
}

func TestModulesNormalizedNeverAllHidden(t *testing.T) {
	if got := (Modules{}).Normalized(); !got.Workbench {
		t.Fatalf("all-hidden modules should fall back to workbench: %+v", got)
	}
	in := Modules{VPN: true}
	if got := in.Normalized(); got != in {
		t.Fatalf("normalize changed a valid selection: %+v", got)
	}
}

func TestExportVarInPath(t *testing.T) {
	cases := []struct {
		e    ExportVar
		want bool
	}{
		{ExportVar{Name: "A", Value: "/usr/local/x"}, true},
		{ExportVar{Name: "A", Value: "~/bin"}, true},
		{ExportVar{Name: "A", Value: "secret"}, false},
		{ExportVar{Name: "A", Value: "/tmp/x", Kind: "var"}, false},
		{ExportVar{Name: "A", Value: "rel", Kind: "path"}, true},
	}
	for _, c := range cases {
		if got := c.e.InPath(); got != c.want {
			t.Errorf("%+v InPath = %v, want %v", c.e, got, c.want)
		}
	}
}
