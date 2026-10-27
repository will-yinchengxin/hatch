package hostmonitor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestParseDiskUsage(t *testing.T) {
	bytes, err := parseDiskUsage("12345\t/Users/example/Documents\n")
	if err != nil || bytes != 12345*1024 {
		t.Fatalf("bytes=%d err=%v", bytes, err)
	}
	if _, err := parseDiskUsage("not-a-number\t/path"); err == nil {
		t.Fatal("invalid output must fail")
	}
}

func TestAnalyzeStorageRoot(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Users", ".hidden", "System"} {
		if err := os.Mkdir(filepath.Join(root, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("Users", filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	entries, err := analyzeStorageRoot(root, func(path string) ([]byte, error) {
		if filepath.Base(path) == "System" {
			return []byte("2\t" + path), errors.New("permission denied")
		}
		return []byte("1\t" + path), nil
	})
	if err != nil || len(entries) != 3 {
		t.Fatalf("entries=%v err=%v", entries, err)
	}
	if entries[0].Name != "System" || !entries[0].Partial || entries[0].Bytes != 2048 {
		t.Fatalf("partial result lost: %+v", entries[0])
	}
}
