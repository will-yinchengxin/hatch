package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"hatch/internal/config"
)

func TestValidateYAML(t *testing.T) {
	if err := validateYAML(""); err != nil {
		t.Fatalf("empty should be allowed, got %v", err)
	}
	if err := validateYAML("mixed-port: 7890\nmode: rule\n"); err != nil {
		t.Fatalf("valid yaml rejected: %v", err)
	}
	if err := validateYAML("foo:\n  - bar\n   baz: 1\n"); err == nil {
		t.Fatalf("expected error for malformed yaml")
	}
}

func TestSafeProfilePath(t *testing.T) {
	cases := map[string]bool{
		"abc.yaml":      true,
		"rHWynz2.js":    true,
		"../etc/passwd": false,
		"a/b.yaml":      false,
		"a\\b.yaml":     false,
		"..":            false,
	}
	for in, want := range cases {
		if got := safeProfilePath(in); got != want {
			t.Fatalf("safeProfilePath(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestDeleteVPNBackup(t *testing.T) {
	backupDir := t.TempDir()
	a := &App{cfg: &config.Config{VPNBackupDir: backupDir}}

	good := filepath.Join(backupDir, "clash-verge-20240101-000000.zip")
	if err := os.WriteFile(good, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := a.DeleteVPNBackup(good); err != nil {
		t.Fatalf("delete valid backup failed: %v", err)
	}
	if _, err := os.Stat(good); !os.IsNotExist(err) {
		t.Fatalf("backup was not removed")
	}

	// Foreign name inside the dir must be refused.
	foreign := filepath.Join(backupDir, "important.zip")
	if err := os.WriteFile(foreign, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := a.DeleteVPNBackup(foreign); err == nil {
		t.Fatalf("expected rejection of non clash-verge file")
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Fatalf("foreign file should remain: %v", err)
	}

	// Path outside the backup dir must be refused.
	outside := filepath.Join(t.TempDir(), "clash-verge-x.zip")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := a.DeleteVPNBackup(outside); err == nil {
		t.Fatalf("expected rejection of path outside backup dir")
	}
}

func TestSaveVPNConfigFileValidatesYAML(t *testing.T) {
	clashDir := t.TempDir()
	a := &App{cfg: &config.Config{VPNAppDir: clashDir}}

	// Invalid YAML must be rejected and nothing written.
	if err := a.SaveVPNConfigFile("verge.yaml", "a:\n  - b\n   c: 1\n"); err == nil {
		t.Fatalf("expected YAML validation error")
	}
	if _, err := os.Stat(filepath.Join(clashDir, "verge.yaml")); !os.IsNotExist(err) {
		t.Fatalf("invalid content should not have been written")
	}

	// Valid YAML is written.
	if err := a.SaveVPNConfigFile("verge.yaml", "enable_tun_mode: true\n"); err != nil {
		t.Fatalf("valid save failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(clashDir, "verge.yaml"))
	if err != nil || string(data) != "enable_tun_mode: true\n" {
		t.Fatalf("unexpected written content: %q err=%v", string(data), err)
	}
}

func TestImportVPNBackupRejectsZipSlip(t *testing.T) {
	clashDir := t.TempDir()
	a := &App{cfg: &config.Config{VPNAppDir: clashDir}}

	zipPath := filepath.Join(t.TempDir(), "evil.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zf)
	// Malicious entry escaping the target dir.
	w, _ := zw.Create("../escape.txt")
	_, _ = w.Write([]byte("pwned"))
	// Benign entry that should land inside.
	w2, _ := zw.Create("verge.yaml")
	_, _ = w2.Write([]byte("mode: rule\n"))
	zw.Close()
	zf.Close()

	if _, err := a.ImportVPNBackup(zipPath); err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(clashDir), "escape.txt")); err == nil {
		t.Fatalf("zip-slip escaped the target directory")
	}
	if _, err := os.Stat(filepath.Join(clashDir, "verge.yaml")); err != nil {
		t.Fatalf("benign file not restored: %v", err)
	}
}
