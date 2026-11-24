package hostmonitor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func cleanupFixture(t *testing.T) (string, cleanupRule) {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rule := cleanupRules[0]
	for _, dir := range []string{rule.relative, ".Trash"} {
		if err := os.MkdirAll(filepath.Join(home, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(home, rule.relative, "cache"), []byte("test cache"), 0600); err != nil {
		t.Fatal(err)
	}
	return home, rule
}

func fixtureTrashMover(home string) storageTrashMover {
	return func(path string, _ int) (string, error) {
		destination := filepath.Join(home, ".Trash", filepath.Base(path))
		if _, err := os.Lstat(destination); err == nil {
			destination += "-fixture"
		}
		if err := unix.RenameatxNp(unix.AT_FDCWD, path, unix.AT_FDCWD, destination, unix.RENAME_EXCL); err != nil {
			return "", err
		}
		return destination, nil
	}
}

func TestStorageCleanupMovesOnlyPreviewedCache(t *testing.T) {
	home, rule := cleanupFixture(t)
	ctx := context.Background()
	target, err := inspectCleanupTarget(ctx, home, rule)
	if err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(home, rule.relative)
	collision := filepath.Join(home, ".Trash", filepath.Base(cache))
	if err := os.WriteFile(collision, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}
	s := storageCleanupSession{scanID: "scan", targets: map[string]cleanupTarget{"allowed": target}, expires: time.Now().Add(time.Minute)}
	if _, err := s.preview("scan", []string{cache}); err == nil {
		t.Fatal("a caller-supplied path was accepted")
	}
	if _, err := s.preview("scan", []string{"allowed", "allowed"}); err == nil {
		t.Fatal("a duplicate candidate was accepted")
	}
	preview, err := s.preview("scan", []string{"allowed"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := s.apply(home, preview.Token, fixtureTrashMover(home))
	if err != nil || len(result) != 1 || result[0].Status != "moved" {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Fatalf("source still present: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(result[0].TrashPath, "cache")); err != nil || string(data) != "test cache" {
		t.Fatalf("cache not in temporary Trash: %q %v", data, err)
	}
	if data, err := os.ReadFile(collision); err != nil || string(data) != "existing" {
		t.Fatalf("existing Trash item changed: %q %v", data, err)
	}
	if _, err := s.apply(home, preview.Token, fixtureTrashMover(home)); err == nil {
		t.Fatal("confirmation token was reusable")
	}
}

func TestStorageCleanupSkipsChangedCacheAndUnsafePaths(t *testing.T) {
	t.Run("nested file changed", func(t *testing.T) {
		home, rule := cleanupFixture(t)
		ctx := context.Background()
		target, err := inspectCleanupTarget(ctx, home, rule)
		if err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(home, rule.relative, "cache")
		if err := os.WriteFile(file, []byte("new cache content"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := moveCleanupTarget(ctx, home, target, fixtureTrashMover(home)); err == nil {
			t.Fatal("changed contents passed revalidation")
		}
		if _, err := os.Stat(file); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("symlink in cache", func(t *testing.T) {
		home, rule := cleanupFixture(t)
		if err := os.Symlink("/System", filepath.Join(home, rule.relative, "shortcut")); err != nil {
			t.Fatal(err)
		}
		if _, err := inspectCleanupTarget(context.Background(), home, rule); err == nil {
			t.Fatal("directory containing a symlink was approved")
		}
	})
	t.Run("hard link to data outside the cache", func(t *testing.T) {
		home, rule := cleanupFixture(t)
		if err := os.Link(filepath.Join(home, rule.relative, "cache"), filepath.Join(home, "kept")); err != nil {
			t.Fatal(err)
		}
		if _, err := inspectCleanupTarget(context.Background(), home, rule); err == nil {
			t.Fatal("directory containing a hard link was approved")
		}
	})
	t.Run("parent replaced by symlink", func(t *testing.T) {
		home, rule := cleanupFixture(t)
		ctx := context.Background()
		target, err := inspectCleanupTarget(ctx, home, rule)
		if err != nil {
			t.Fatal(err)
		}
		parent := filepath.Dir(filepath.Join(home, rule.relative))
		if err := os.Rename(parent, parent+"-moved"); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(parent+"-moved", parent); err != nil {
			t.Fatal(err)
		}
		if _, err := moveCleanupTarget(ctx, home, target, fixtureTrashMover(home)); err == nil {
			t.Fatal("parent symlink was followed")
		}
	})
	t.Run("trash replaced by symlink", func(t *testing.T) {
		home, rule := cleanupFixture(t)
		ctx := context.Background()
		target, err := inspectCleanupTarget(ctx, home, rule)
		if err != nil {
			t.Fatal(err)
		}
		trash := filepath.Join(home, ".Trash")
		if err := os.Rename(trash, trash+"-moved"); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(trash+"-moved", trash); err != nil {
			t.Fatal(err)
		}
		if _, err := moveCleanupTarget(ctx, home, target, fixtureTrashMover(home)); err == nil {
			t.Fatal("Trash symlink was followed")
		}
	})
}

func TestStorageCleanupRejectsExpiredConfirmation(t *testing.T) {
	home, rule := cleanupFixture(t)
	target, err := inspectCleanupTarget(context.Background(), home, rule)
	if err != nil {
		t.Fatal(err)
	}
	s := storageCleanupSession{scanID: "scan", targets: map[string]cleanupTarget{"allowed": target}, expires: time.Now().Add(time.Minute)}
	preview, err := s.preview("scan", []string{"allowed"})
	if err != nil {
		t.Fatal(err)
	}
	s.previewExpires = time.Now().Add(-time.Second)
	if _, err := s.apply(home, preview.Token, fixtureTrashMover(home)); err == nil {
		t.Fatal("expired confirmation was accepted")
	}
	if _, err := os.Stat(target.item.Path); err != nil {
		t.Fatal(err)
	}
}

func TestNativeStorageTrashRejectsReplacedPath(t *testing.T) {
	if !nativeStorageTrashAvailable {
		t.Skip("native Trash requires CGO")
	}
	home, rule := cleanupFixture(t)
	path := filepath.Join(home, rule.relative)
	original, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer original.Close()
	if err := os.Rename(path, path+"-renamed"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := trashStorageCache(path, int(original.Fd())); err == nil {
		t.Fatal("native Trash accepted a replaced path")
	}
	for _, directory := range []string{path, path + "-renamed"} {
		if _, err := os.Stat(directory); err != nil {
			t.Fatal(err)
		}
	}
}

func TestStorageCleanupKeepsUnselectedCandidates(t *testing.T) {
	home, first := cleanupFixture(t)
	second := cleanupRules[1]
	if err := os.MkdirAll(filepath.Join(home, second.relative), 0700); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	firstTarget, err := inspectCleanupTarget(ctx, home, first)
	if err != nil {
		t.Fatal(err)
	}
	secondTarget, err := inspectCleanupTarget(ctx, home, second)
	if err != nil {
		t.Fatal(err)
	}
	firstTarget.item.ID, secondTarget.item.ID = "first", "second"
	session := storageCleanupSession{
		scanID: "scan", expires: time.Now().Add(time.Minute),
		targets: map[string]cleanupTarget{"first": firstTarget, "second": secondTarget},
	}
	preview, err := session.preview("scan", []string{"first"})
	if err != nil {
		t.Fatal(err)
	}
	if results, err := session.apply(home, preview.Token, fixtureTrashMover(home)); err != nil || results[0].Status != "moved" {
		t.Fatalf("first cleanup: %+v, %v", results, err)
	}
	if _, err := session.preview("scan", []string{"first"}); err == nil {
		t.Fatal("moved candidate is still selectable")
	}
	preview, err = session.preview("scan", []string{"second"})
	if err != nil {
		t.Fatalf("remaining candidate requires another scan: %v", err)
	}
	if results, err := session.apply(home, preview.Token, fixtureTrashMover(home)); err != nil || results[0].Status != "moved" {
		t.Fatalf("second cleanup: %+v, %v", results, err)
	}
}

func TestStorageInventoryAuthorizesOnlyCheckedDownloadsAndExactCaches(t *testing.T) {
	home, cacheRule := cleanupFixture(t)
	downloads := filepath.Join(home, "Downloads")
	if err := os.Mkdir(downloads, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"archive.zip", "changed.zip", ".hidden.zip"} {
		if err := os.WriteFile(filepath.Join(downloads, name), []byte("test data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(downloads, "folder"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(filepath.Join(downloads, ".hidden.zip"), filepath.Join(downloads, "hardlink.zip")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(downloads, "archive.zip"), filepath.Join(downloads, "link.zip")); err != nil {
		t.Fatal(err)
	}
	group := StorageGroup{Name: "下载", path: downloads}
	listStorageGroup(&group)
	if err := os.WriteFile(filepath.Join(downloads, "changed.zip"), []byte("new contents"), 0600); err != nil {
		t.Fatal(err)
	}
	cache := StorageCleanupItem{
		ID: "cache", Name: cacheRule.name, Path: filepath.Join(home, cacheRule.relative), Bytes: cleanupMinBytes,
	}
	report := StorageReport{
		Candidates: []StorageCleanupItem{cache},
		Groups:     []StorageGroup{group, {Name: developerStorageGroup, Entries: []StorageUsage{{Path: filepath.Dir(cache.Path)}}}},
	}
	targets := map[string]cleanupTarget{}
	attachStorageActions(context.Background(), home, &report, targets)
	if len(targets) != 1 {
		t.Fatalf("unexpected download targets: %+v", targets)
	}
	var id string
	for _, entry := range report.Groups[0].Entries {
		if entry.Path == filepath.Join(downloads, "archive.zip") {
			id = entry.CleanupID
		} else if entry.CleanupID != "" {
			t.Fatalf("unsafe entry has cleanup action: %s", entry.Path)
		}
	}
	if id == "" || !targets[id].item.UserFile {
		t.Fatal("checked download lacks cleanup action")
	}
	for _, entry := range report.Groups[1].Entries {
		if entry.Path == cache.Path && entry.CleanupID == cache.ID {
			continue
		}
		if entry.Path != filepath.Dir(cache.Path) || entry.CleanupID != "" {
			t.Fatalf("incorrect cache action: %+v", entry)
		}
	}
	s := storageCleanupSession{scanID: "scan", targets: targets, expires: time.Now().Add(time.Minute)}
	preview, err := s.preview("scan", []string{id})
	if err != nil {
		t.Fatal(err)
	}
	results, err := s.apply(home, preview.Token, fixtureTrashMover(home))
	if err != nil || len(results) != 1 || results[0].Status != "moved" {
		t.Fatalf("download cleanup: %+v, %v", results, err)
	}
	if _, err := os.Stat(filepath.Join(downloads, "changed.zip")); err != nil {
		t.Fatal(err)
	}
}
