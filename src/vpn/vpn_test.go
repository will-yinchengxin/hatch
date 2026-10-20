package vpn

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	a := Manager{BackupDir: backupDir}

	good := filepath.Join(backupDir, "clash-verge-20240101-000000.zip")
	if err := os.WriteFile(good, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := a.DeleteBackup(good); err != nil {
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
	if err := a.DeleteBackup(foreign); err == nil {
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
	if err := a.DeleteBackup(outside); err == nil {
		t.Fatalf("expected rejection of path outside backup dir")
	}
}

func TestSaveVPNConfigFileValidatesYAML(t *testing.T) {
	clashDir := t.TempDir()
	a := Manager{AppDir: clashDir}

	// Invalid YAML must be rejected and nothing written.
	if err := a.WriteConfigFile("verge.yaml", "a:\n  - b\n   c: 1\n"); err == nil {
		t.Fatalf("expected YAML validation error")
	}
	if _, err := os.Stat(filepath.Join(clashDir, "verge.yaml")); !os.IsNotExist(err) {
		t.Fatalf("invalid content should not have been written")
	}

	// Valid YAML is written.
	if err := a.WriteConfigFile("verge.yaml", "enable_tun_mode: true\n"); err != nil {
		t.Fatalf("valid save failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(clashDir, "verge.yaml"))
	if err != nil || string(data) != "enable_tun_mode: true\n" {
		t.Fatalf("unexpected written content: %q err=%v", string(data), err)
	}
}

func TestValidateVPNSettings(t *testing.T) {
	clashDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(clashDir, "verge.yaml"), []byte("enable_tun_mode: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(clashDir, "profiles.yaml"), []byte("items: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	backupDir := filepath.Join(t.TempDir(), "backups")
	a := Manager{}

	got := a.ValidatePaths(clashDir, backupDir)
	if !got.AppDirOK || !got.BackupDirOK || len(got.Errors) != 0 {
		t.Fatalf("unexpected validation: %+v", got)
	}

	got = a.ValidatePaths(filepath.Join(t.TempDir(), "missing"), backupDir)
	if got.AppDirOK || len(got.Errors) == 0 {
		t.Fatalf("missing app dir should fail: %+v", got)
	}
}

func TestGetVPNHealthReportMissingDir(t *testing.T) {
	a := Manager{AppDir: filepath.Join(t.TempDir(), "missing")}

	got := a.CheckHealth()
	if got.Overall != "error" {
		t.Fatalf("overall = %q, want error: %+v", got.Overall, got)
	}
	if len(got.Checks) == 0 || got.Checks[0].ID != "app-dir" {
		t.Fatalf("expected app-dir check, got %+v", got.Checks)
	}
}

func TestPreviewVPNBackup(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "backup.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zf)
	for name, content := range map[string]string{
		"verge.yaml":          "enable_tun_mode: true\n",
		"profiles.yaml":       "current: p1\nitems: []\n",
		"profiles/p1.yaml":    "mixed-port: 7890\n",
		"../should-skip.yaml": "x: 1\n",
	} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(content))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zf.Close(); err != nil {
		t.Fatal(err)
	}

	a := Manager{}
	got, err := a.PreviewBackup(zipPath)
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}
	if !got.Valid || got.NumFiles != 3 {
		t.Fatalf("unexpected preview: %+v", got)
	}
	if len(got.Warnings) == 0 {
		t.Fatalf("expected warning for suspicious path")
	}
}

func TestImportVPNBackupRejectsZipSlip(t *testing.T) {
	clashDir := t.TempDir()
	backupDir := t.TempDir()
	a := Manager{AppDir: clashDir, BackupDir: backupDir}
	if err := os.WriteFile(filepath.Join(clashDir, "profiles.yaml"), []byte("items: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

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

	msg, err := a.ImportBackup(zipPath)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if !strings.Contains(msg, "恢复前备份") {
		t.Fatalf("expected pre-restore backup message, got %q", msg)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(clashDir), "escape.txt")); err == nil {
		t.Fatalf("zip-slip escaped the target directory")
	}
	if _, err := os.Stat(filepath.Join(clashDir, "verge.yaml")); err != nil {
		t.Fatalf("benign file not restored: %v", err)
	}
	backups, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 || !strings.HasPrefix(backups[0].Name(), "clash-verge-pre-restore-") {
		t.Fatalf("expected one pre-restore backup, got %+v", backups)
	}
}
