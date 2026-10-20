package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"hatch/core/config"
	"hatch/memo"
)

func TestMemoCRUDAndReview(t *testing.T) {
	dir := t.TempDir()
	a := &App{
		cfg:        &config.Config{Memo: config.MemoSettings{FilesDir: filepath.Join(dir, "files"), RetainDays: 90, AutoCleanup: true}},
		memoTimers: map[string]*time.Timer{},
	}
	st, err := memo.New(filepath.Join(dir, "memos.json"))
	if err != nil {
		t.Fatal(err)
	}
	a.memoStore = st

	item, err := a.SaveMemo(memo.Item{Title: "Write report", Date: "2026-06-29"})
	if err != nil {
		t.Fatal(err)
	}
	if item.ID == "" || item.Status != "todo" {
		t.Fatalf("unexpected memo: %#v", item)
	}
	item, err = a.MarkMemoDone(item.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != "done" {
		t.Fatalf("expected done, got %q", item.Status)
	}

	word, err := a.SaveVocabEntry(memo.VocabEntry{Word: "resilient", Chinese: "有韧性的"})
	if err != nil {
		t.Fatal(err)
	}
	reviewed, err := a.ReviewVocabEntry(word.ID, "known")
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.LastReviewedAt == 0 || reviewed.NextReviewAt <= time.Now().Unix() {
		t.Fatalf("expected future review date: %#v", reviewed)
	}
}

func TestCleanupMemoFilesMarksManagedAttachment(t *testing.T) {
	dir := t.TempDir()
	filesDir := filepath.Join(dir, "files")
	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldFile := filepath.Join(filesDir, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := &App{
		cfg:        &config.Config{Memo: config.MemoSettings{FilesDir: filesDir, RetainDays: 1, AutoCleanup: true}},
		memoTimers: map[string]*time.Timer{},
	}
	st, err := memo.New(filepath.Join(dir, "memos.json"))
	if err != nil {
		t.Fatal(err)
	}
	a.memoStore = st
	_, err = st.SaveMemo(memo.Item{
		ID:        "m1",
		Title:     "with attachment",
		Date:      "2026-06-29",
		Status:    "todo",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		Attachments: []memo.Attachment{{
			ID:          "a1",
			Kind:        "file",
			Label:       "old.txt",
			ManagedPath: oldFile,
			CreatedAt:   time.Now().AddDate(0, 0, -2).Unix(),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	removed, err := a.CleanupMemoFiles()
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed=%d, want 1", removed)
	}
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, stat err=%v", err)
	}
	item, ok := st.Memo("m1")
	if !ok || item.Attachments[0].CleanedAt == 0 {
		t.Fatalf("expected attachment marked cleaned: %#v", item)
	}
}
