package main

import (
	"os"
	"path/filepath"
	"testing"

	"hatch/internal/config"
)

func TestWorkspaceSelectionNormalizationAndIDs(t *testing.T) {
	got := normalizeSelections([]config.AgentResourceSelection{
		{ID: " skill ", Claude: true},
		{ID: "skill", Codex: true},
		{ID: "", Claude: true},
		{ID: "off"},
	})
	if len(got) != 1 || got[0].ID != "skill" || !got[0].Claude || got[0].Codex {
		t.Fatalf("unexpected normalized selections: %#v", got)
	}
	for _, id := range []string{"ws-abc123", "snapshot-abc123"} {
		if !validWorkspaceID(id) {
			t.Fatalf("expected valid id: %s", id)
		}
	}
	for _, id := range []string{"ws-../bad", "snapshot-a/b", "other-123"} {
		if validWorkspaceID(id) {
			t.Fatalf("expected invalid id: %s", id)
		}
	}
}

func TestWorkspaceFileSnapshotRestoresOriginalBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	snap, err := snapshotFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("after"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := restoreFile(snap); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "before" {
		t.Fatalf("restored content = %q", got)
	}
}
