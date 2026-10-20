package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"hatch/core/config"
	"hatch/core/fsutil"
	"hatch/core/macos"
	"hatch/memo"
)

// ─── ✓ 备忘录 · 待办、提醒与附件 ─────────────────────────────────────────────
//
// Reminders are in-process timers: they only fire while Hatch is running and
// emit "memo:reminder" plus a best-effort macOS banner/sound.

func (a *App) ensureMemoStore() (*memo.Store, error) {
	a.memoMu.Lock()
	defer a.memoMu.Unlock()
	if a.memoStore != nil {
		return a.memoStore, nil
	}
	st, err := memo.New(memo.DefaultStorePath())
	if err != nil {
		return nil, err
	}
	a.memoStore = st
	return st, nil
}

// memoFilesDir is the managed attachment directory (config or default).
func (a *App) memoFilesDir() string {
	a.mu.RLock()
	dir := ""
	if a.cfg != nil {
		dir = strings.TrimSpace(a.cfg.Memo.FilesDir)
	}
	a.mu.RUnlock()
	if dir != "" {
		return fsutil.ExpandHome(dir)
	}
	return memo.DefaultFilesDir()
}

func (a *App) GetMemoDashboard() memo.Dashboard {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.Dashboard{}
	}
	return st.BuildDashboard(time.Now())
}

// ListMemos returns memos by date desc, then most recently updated.
func (a *App) ListMemos() ([]memo.Item, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return nil, err
	}
	return st.SortedMemos(), nil
}

// SaveMemo fills defaults, saves and (re)schedules the reminder.
func (a *App) SaveMemo(item memo.Item) (memo.Item, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.Item{}, err
	}
	saved, err := st.SaveMemo(memo.PrepareItem(item, time.Now()))
	if err != nil {
		return memo.Item{}, err
	}
	a.scheduleMemoReminder(saved)
	return saved, nil
}

// DeleteMemo removes a memo, its reminder and its managed attachment files.
func (a *App) DeleteMemo(id string) error {
	st, err := a.ensureMemoStore()
	if err != nil {
		return err
	}
	item, ok, err := st.DeleteMemo(id)
	if err != nil || !ok {
		return err
	}
	a.removeMemoReminder(id)
	filesDir := a.memoFilesDir()
	for _, att := range item.Attachments {
		memo.RemoveManagedFile(att, filesDir)
	}
	return nil
}

func (a *App) MarkMemoDone(id string, done bool) (memo.Item, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.Item{}, err
	}
	item, ok := st.Memo(id)
	if !ok {
		return memo.Item{}, fmt.Errorf("memo not found: %s", id)
	}
	item.Status = "todo"
	if done {
		item.Status = "done"
	}
	item.UpdatedAt = time.Now().Unix()
	saved, err := st.SaveMemo(item)
	if err != nil {
		return memo.Item{}, err
	}
	a.scheduleMemoReminder(saved)
	return saved, nil
}

// ─── Reminders ───────────────────────────────────────────────────────────────

// ScheduleMemoReminders arms timers for every pending future reminder.
func (a *App) ScheduleMemoReminders() {
	st, err := a.ensureMemoStore()
	if err != nil {
		return
	}
	for _, item := range st.AllMemos() {
		a.scheduleMemoReminder(item)
	}
}

func (a *App) scheduleMemoReminder(item memo.Item) {
	a.removeMemoReminder(item.ID)
	if item.ID == "" || item.Status == "done" || item.ReminderAt <= time.Now().Unix() || item.ReminderSentAt > 0 {
		return
	}
	delay := time.Until(time.Unix(item.ReminderAt, 0))
	if delay <= 0 {
		return
	}
	timer := time.AfterFunc(delay, func() {
		a.fireMemoReminder(item.ID)
	})
	a.memoMu.Lock()
	a.memoTimers[item.ID] = timer
	a.memoMu.Unlock()
}

func (a *App) removeMemoReminder(id string) {
	a.memoMu.Lock()
	defer a.memoMu.Unlock()
	if timer, ok := a.memoTimers[id]; ok {
		timer.Stop()
		delete(a.memoTimers, id)
	}
}

func (a *App) fireMemoReminder(id string) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return
	}
	item, ok := st.Memo(id)
	if !ok || item.Status == "done" || item.ReminderSentAt > 0 {
		return
	}
	mode := item.ReminderMode
	if mode == "sound" || mode == "both" || mode == "" {
		macos.Beep()
	}
	if mode == "banner" || mode == "both" || mode == "" {
		macos.Notify("Hatch 备忘录", item.Title)
	}
	item.ReminderSentAt = time.Now().Unix()
	item.UpdatedAt = item.ReminderSentAt
	_, _ = st.SaveMemo(item)
	a.emit("memo:reminder", item)
	a.removeMemoReminder(id)
}

// ─── Attachments & storage policy ────────────────────────────────────────────

// GetMemoSettings returns the effective attachment storage policy. Auto
// cleanup defaults to on until the user saves settings explicitly.
func (a *App) GetMemoSettings() memo.Settings {
	a.mu.RLock()
	retain := memo.DefaultFileRetainDays
	auto := true
	if a.cfg != nil {
		if a.cfg.Memo.RetainDays > 0 {
			retain = a.cfg.Memo.RetainDays
		}
		if a.cfg.Memo.RetainDays > 0 || strings.TrimSpace(a.cfg.Memo.FilesDir) != "" {
			auto = a.cfg.Memo.AutoCleanup
		}
	}
	a.mu.RUnlock()
	return memo.Settings{FilesDir: a.memoFilesDir(), RetainDays: retain, AutoCleanup: auto}
}

func (a *App) SaveMemoSettings(filesDir string, retainDays int, autoCleanup bool) error {
	if retainDays <= 0 {
		retainDays = memo.DefaultFileRetainDays
	}
	filesDir = strings.TrimSpace(filesDir)
	if filesDir == "" {
		filesDir = a.memoFilesDir()
	}
	filesDir = fsutil.ExpandHome(filesDir)
	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.Memo = config.MemoSettings{FilesDir: filesDir, RetainDays: retainDays, AutoCleanup: autoCleanup}
	err := a.cfg.Save()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	if autoCleanup {
		_, _ = a.CleanupMemoFiles()
	}
	return nil
}

// ImportMemoAttachment copies a local file into the managed directory (or
// wraps a URL as a link) and returns the attachment for the memo form.
func (a *App) ImportMemoAttachment(sourcePath, label, kind string) (memo.Attachment, error) {
	return memo.ImportAttachment(sourcePath, label, kind, a.memoFilesDir())
}

func (a *App) DeleteMemoAttachment(memoID, attachmentID string) (memo.Item, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.Item{}, err
	}
	item, ok := st.Memo(memoID)
	if !ok {
		return memo.Item{}, fmt.Errorf("memo not found: %s", memoID)
	}
	filesDir := a.memoFilesDir()
	next := make([]memo.Attachment, 0, len(item.Attachments))
	for _, att := range item.Attachments {
		if att.ID == attachmentID {
			memo.RemoveManagedFile(att, filesDir)
			continue
		}
		next = append(next, att)
	}
	item.Attachments = next
	item.UpdatedAt = time.Now().Unix()
	return st.SaveMemo(item)
}

func (a *App) OpenMemoAttachment(att memo.Attachment) error {
	target, err := memo.OpenTarget(att)
	if err != nil {
		return err
	}
	return macos.Open(target)
}

func (a *App) GetMemoAttachmentPreview(att memo.Attachment) (memo.AttachmentPreview, error) {
	return memo.Preview(att, a.memoFilesDir())
}

// CleanupMemoFiles force-deletes expired managed attachment files.
func (a *App) CleanupMemoFiles() (int, error) {
	return a.cleanupMemoFiles(true)
}

func (a *App) applyMemoRetention() {
	if a.GetMemoSettings().AutoCleanup {
		_, _ = a.cleanupMemoFiles(false)
	}
}

func (a *App) cleanupMemoFiles(force bool) (int, error) {
	settings := a.GetMemoSettings()
	if !force && !settings.AutoCleanup {
		return 0, nil
	}
	st, err := a.ensureMemoStore()
	if err != nil {
		return 0, err
	}
	return st.CleanupAttachments(settings.FilesDir, settings.RetainDays), nil
}
