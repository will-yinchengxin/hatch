package zshrc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteVarsCreatesMissingZshrc(t *testing.T) {
	p := filepath.Join(t.TempDir(), ".zshrc")
	if err := WriteVars(p, []Var{{Name: "FOO", Value: "bar"}}); err != nil {
		t.Fatalf("missing ~/.zshrc should be created: %v", err)
	}
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "export FOO=bar") {
		t.Fatalf("block not written: %q", data)
	}
}
