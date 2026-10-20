package memo

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hatch/core/fsutil"
)

// DefaultFileRetainDays is used when config.Memo.RetainDays is 0.
const DefaultFileRetainDays = 90

// Dashboard is the memo hero stats.
type Dashboard struct {
	Today      int `json:"today"`
	Overdue    int `json:"overdue"`
	Reminders  int `json:"reminders"`
	ReviewDue  int `json:"reviewDue"`
	TotalMemos int `json:"totalMemos"`
	TotalVocab int `json:"totalVocab"`
}

// Settings is the effective attachment storage policy shown on the 存储 page.
type Settings struct {
	FilesDir    string `json:"filesDir"`
	RetainDays  int    `json:"retainDays"`
	AutoCleanup bool   `json:"autoCleanup"`
}

// AttachmentPreview carries an inline data URL for managed images.
type AttachmentPreview struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Label   string `json:"label"`
	DataURL string `json:"dataUrl,omitempty"`
	Path    string `json:"path,omitempty"`
}

// DefaultStorePath is ~/Library/Application Support/Hatch/memos.json.
func DefaultStorePath() string {
	return fsutil.AppDataPath("memos.json")
}

// DefaultFilesDir is ~/Library/Application Support/Hatch/memo-files.
func DefaultFilesDir() string {
	return fsutil.AppDataPath("memo-files")
}

// BuildDashboard counts today's/overdue/upcoming memos and due vocab reviews.
func (s *Store) BuildDashboard(now time.Time) Dashboard {
	today := now.Format("2006-01-02")
	var out Dashboard
	for _, item := range s.AllMemos() {
		out.TotalMemos++
		if item.Date == today && item.Status != "done" {
			out.Today++
		}
		if item.Status != "done" && item.ReminderAt > 0 {
			if item.ReminderAt < now.Unix() {
				out.Overdue++
			} else {
				out.Reminders++
			}
		}
	}
	for _, entry := range s.AllVocab() {
		out.TotalVocab++
		if entry.NextReviewAt == 0 || entry.NextReviewAt <= now.Unix() {
			out.ReviewDue++
		}
	}
	return out
}

// SortedMemos returns memos by date desc, then most recently updated.
func (s *Store) SortedMemos() []Item {
	items := s.AllMemos()
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Date != items[j].Date {
			return items[i].Date > items[j].Date
		}
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	return items
}

// SortedVocab returns vocab entries alphabetically (case-insensitive).
func (s *Store) SortedVocab() []VocabEntry {
	entries := s.AllVocab()
	sort.SliceStable(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Word) < strings.ToLower(entries[j].Word)
	})
	return entries
}

// PrepareItem fills ids, timestamps and defaults before a memo is saved, and
// re-arms the reminder when it was moved into the future.
func PrepareItem(item Item, now time.Time) Item {
	ts := now.Unix()
	if strings.TrimSpace(item.ID) == "" {
		item.ID = fmt.Sprintf("memo-%d", now.UnixNano())
		item.CreatedAt = ts
	}
	if item.CreatedAt == 0 {
		item.CreatedAt = ts
	}
	item.UpdatedAt = ts
	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		item.Title = "未命名待办"
	}
	if item.Date == "" {
		item.Date = now.Format("2006-01-02")
	}
	if item.Status == "" {
		item.Status = "todo"
	}
	if item.ReminderMode == "" {
		item.ReminderMode = "both"
	}
	if item.ReminderAt > ts {
		item.ReminderSentAt = 0
	}
	return item
}

// PrepareVocab fills ids/timestamps and requires a word.
func PrepareVocab(entry VocabEntry, now time.Time) (VocabEntry, error) {
	ts := now.Unix()
	if strings.TrimSpace(entry.ID) == "" {
		entry.ID = fmt.Sprintf("word-%d", now.UnixNano())
		entry.CreatedAt = ts
	}
	if entry.CreatedAt == 0 {
		entry.CreatedAt = ts
	}
	entry.Word = strings.TrimSpace(entry.Word)
	if entry.Word == "" {
		return VocabEntry{}, fmt.Errorf("英文单词不能为空")
	}
	entry.UpdatedAt = ts
	if entry.NextReviewAt == 0 {
		entry.NextReviewAt = ts
	}
	return entry, nil
}

// ApplyReview schedules the next review: hard → 1d and reset level, known →
// 3+2·level days, mastered → 14+3·level days (level +2).
func ApplyReview(entry VocabEntry, result string, now time.Time) VocabEntry {
	switch result {
	case "hard":
		entry.ReviewLevel = 0
		entry.NextReviewAt = now.AddDate(0, 0, 1).Unix()
	case "known":
		entry.ReviewLevel++
		entry.NextReviewAt = now.AddDate(0, 0, 3+entry.ReviewLevel*2).Unix()
	case "mastered":
		entry.ReviewLevel += 2
		entry.NextReviewAt = now.AddDate(0, 0, 14+entry.ReviewLevel*3).Unix()
	default:
		entry.NextReviewAt = now.AddDate(0, 0, 1).Unix()
	}
	entry.LastReviewedAt = now.Unix()
	entry.UpdatedAt = now.Unix()
	return entry
}

// ImportAttachment turns a URL into a link attachment, or copies a local file
// into <filesDir>/<YYYY-MM>/ and returns the managed attachment.
func ImportAttachment(sourcePath, label, kind, filesDir string) (Attachment, error) {
	sourcePath = strings.TrimSpace(sourcePath)
	label = strings.TrimSpace(label)
	kind = strings.TrimSpace(kind)
	if strings.HasPrefix(sourcePath, "http://") || strings.HasPrefix(sourcePath, "https://") {
		if label == "" {
			label = sourcePath
		}
		return Attachment{
			ID:        fmt.Sprintf("att-%d", time.Now().UnixNano()),
			Kind:      "link",
			Label:     label,
			URL:       sourcePath,
			CreatedAt: time.Now().Unix(),
		}, nil
	}
	sourcePath = fsutil.ExpandHome(sourcePath)
	info, err := os.Stat(sourcePath)
	if err != nil {
		return Attachment{}, err
	}
	if info.IsDir() {
		return Attachment{}, fmt.Errorf("附件必须是文件或 URL")
	}
	monthDir := filepath.Join(filesDir, time.Now().Format("2006-01"))
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		return Attachment{}, err
	}
	id := fmt.Sprintf("att-%d", time.Now().UnixNano())
	target := filepath.Join(monthDir, id+filepath.Ext(sourcePath))
	mimeType, err := fsutil.CopyFileWithMime(sourcePath, target)
	if err != nil {
		return Attachment{}, err
	}
	if label == "" {
		label = filepath.Base(sourcePath)
	}
	if kind == "" || kind == "file" {
		if strings.HasPrefix(mimeType, "image/") {
			kind = "image"
		} else {
			kind = "file"
		}
	}
	return Attachment{
		ID:          id,
		Kind:        kind,
		Label:       label,
		SourcePath:  sourcePath,
		ManagedPath: target,
		Mime:        mimeType,
		Size:        info.Size(),
		CreatedAt:   time.Now().Unix(),
	}, nil
}

// RemoveManagedFile deletes att's managed copy when it lives inside filesDir.
func RemoveManagedFile(att Attachment, filesDir string) {
	if att.ManagedPath != "" && fsutil.IsPathInside(att.ManagedPath, filesDir) {
		_ = os.Remove(att.ManagedPath)
	}
}

// OpenTarget returns where an attachment should be opened from.
func OpenTarget(att Attachment) (string, error) {
	switch {
	case att.URL != "":
		return att.URL, nil
	case att.ManagedPath != "":
		return att.ManagedPath, nil
	case att.SourcePath != "":
		return att.SourcePath, nil
	default:
		return "", fmt.Errorf("附件没有可打开的位置")
	}
}

// Preview inlines managed images as a data URL; other kinds return metadata only.
func Preview(att Attachment, filesDir string) (AttachmentPreview, error) {
	out := AttachmentPreview{ID: att.ID, Kind: att.Kind, Label: att.Label, Path: att.ManagedPath}
	if att.Kind != "image" || att.ManagedPath == "" || !fsutil.IsPathInside(att.ManagedPath, filesDir) {
		return out, nil
	}
	data, err := os.ReadFile(att.ManagedPath)
	if err != nil {
		return out, err
	}
	mimeType := att.Mime
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	out.DataURL = "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
	return out, nil
}

// CleanupAttachments deletes managed attachment files older than retainDays
// and marks them CleanedAt; memo text records are kept. Returns files removed.
func (s *Store) CleanupAttachments(filesDir string, retainDays int) int {
	cutoff := time.Now().AddDate(0, 0, -retainDays)
	removed := 0
	for _, item := range s.AllMemos() {
		changed := false
		for i := range item.Attachments {
			att := &item.Attachments[i]
			if att.CleanedAt > 0 || att.ManagedPath == "" || !fsutil.IsPathInside(att.ManagedPath, filesDir) {
				continue
			}
			if att.CreatedAt > 0 && time.Unix(att.CreatedAt, 0).After(cutoff) {
				continue
			}
			if os.Remove(att.ManagedPath) == nil {
				removed++
			}
			att.CleanedAt = time.Now().Unix()
			changed = true
		}
		if changed {
			item.UpdatedAt = time.Now().Unix()
			_, _ = s.SaveMemo(item)
		}
	}
	return removed
}
