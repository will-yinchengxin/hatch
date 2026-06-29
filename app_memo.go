package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hatch/internal/config"
	"hatch/internal/memo"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const defaultMemoFileRetainDays = 90

type MemoDashboard struct {
	Today      int `json:"today"`
	Overdue    int `json:"overdue"`
	Reminders  int `json:"reminders"`
	ReviewDue  int `json:"reviewDue"`
	TotalMemos int `json:"totalMemos"`
	TotalVocab int `json:"totalVocab"`
}

type MemoSettingsView struct {
	FilesDir    string `json:"filesDir"`
	RetainDays  int    `json:"retainDays"`
	AutoCleanup bool   `json:"autoCleanup"`
}

type MemoAttachmentPreview struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Label   string `json:"label"`
	DataURL string `json:"dataUrl,omitempty"`
	Path    string `json:"path,omitempty"`
}

type VocabLookupResult struct {
	Word         string             `json:"word"`
	Phonetic     string             `json:"phonetic,omitempty"`
	AudioURL     string             `json:"audioUrl,omitempty"`
	PartOfSpeech string             `json:"partOfSpeech,omitempty"`
	Definitions  []string           `json:"definitions,omitempty"`
	Collocations []memo.Collocation `json:"collocations,omitempty"`
	Source       string             `json:"source,omitempty"`
}

func (a *App) memoStorePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "hatch", "memos.json")
	}
	return filepath.Join(dir, "Hatch", "memos.json")
}

func (a *App) memoFilesDir() string {
	a.mu.RLock()
	dir := ""
	if a.cfg != nil {
		dir = strings.TrimSpace(a.cfg.Memo.FilesDir)
	}
	a.mu.RUnlock()
	if dir != "" {
		return expandHome(dir)
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "hatch", "memo-files")
	}
	return filepath.Join(cfgDir, "Hatch", "memo-files")
}

func (a *App) ensureMemoStore() (*memo.Store, error) {
	a.memoMu.Lock()
	defer a.memoMu.Unlock()
	if a.memoStore != nil {
		return a.memoStore, nil
	}
	st, err := memo.New(a.memoStorePath())
	if err != nil {
		return nil, err
	}
	a.memoStore = st
	return st, nil
}

func (a *App) GetMemoDashboard() MemoDashboard {
	st, err := a.ensureMemoStore()
	if err != nil {
		return MemoDashboard{}
	}
	now := time.Now()
	today := now.Format("2006-01-02")
	var out MemoDashboard
	for _, item := range st.AllMemos() {
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
	for _, entry := range st.AllVocab() {
		out.TotalVocab++
		if entry.NextReviewAt == 0 || entry.NextReviewAt <= now.Unix() {
			out.ReviewDue++
		}
	}
	return out
}

func (a *App) ListMemos() ([]memo.Item, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return nil, err
	}
	items := st.AllMemos()
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Date != items[j].Date {
			return items[i].Date > items[j].Date
		}
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	return items, nil
}

func (a *App) SaveMemo(item memo.Item) (memo.Item, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.Item{}, err
	}
	now := time.Now().Unix()
	if strings.TrimSpace(item.ID) == "" {
		item.ID = fmt.Sprintf("memo-%d", time.Now().UnixNano())
		item.CreatedAt = now
	}
	if item.CreatedAt == 0 {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		item.Title = "未命名待办"
	}
	if item.Date == "" {
		item.Date = time.Now().Format("2006-01-02")
	}
	if item.Status == "" {
		item.Status = "todo"
	}
	if item.ReminderMode == "" {
		item.ReminderMode = "both"
	}
	if item.ReminderAt > now {
		item.ReminderSentAt = 0
	}
	saved, err := st.SaveMemo(item)
	if err != nil {
		return memo.Item{}, err
	}
	a.scheduleMemoReminder(saved)
	return saved, nil
}

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
	for _, att := range item.Attachments {
		if att.ManagedPath != "" && isPathInside(att.ManagedPath, a.memoFilesDir()) {
			_ = os.Remove(att.ManagedPath)
		}
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
	if done {
		item.Status = "done"
	} else {
		item.Status = "todo"
	}
	item.UpdatedAt = time.Now().Unix()
	saved, err := st.SaveMemo(item)
	if err != nil {
		return memo.Item{}, err
	}
	a.scheduleMemoReminder(saved)
	return saved, nil
}

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
	if item.ReminderMode == "sound" || item.ReminderMode == "both" || item.ReminderMode == "" {
		playMacSound()
	}
	if item.ReminderMode == "banner" || item.ReminderMode == "both" || item.ReminderMode == "" {
		sendMacNotification("Hatch 备忘录", item.Title)
	}
	item.ReminderSentAt = time.Now().Unix()
	item.UpdatedAt = item.ReminderSentAt
	_, _ = st.SaveMemo(item)
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "memo:reminder", item)
	}
	a.removeMemoReminder(id)
}

func playMacSound() {
	_ = exec.Command("osascript", "-e", "beep").Start()
}

func (a *App) GetMemoSettings() MemoSettingsView {
	a.mu.RLock()
	defer a.mu.RUnlock()
	retain := defaultMemoFileRetainDays
	auto := true
	if a.cfg != nil {
		if a.cfg.Memo.RetainDays > 0 {
			retain = a.cfg.Memo.RetainDays
		}
		if a.cfg.Memo.RetainDays > 0 || strings.TrimSpace(a.cfg.Memo.FilesDir) != "" {
			auto = a.cfg.Memo.AutoCleanup
		}
	}
	return MemoSettingsView{FilesDir: a.memoFilesDir(), RetainDays: retain, AutoCleanup: auto}
}

func (a *App) SaveMemoSettings(filesDir string, retainDays int, autoCleanup bool) error {
	if retainDays <= 0 {
		retainDays = defaultMemoFileRetainDays
	}
	filesDir = strings.TrimSpace(filesDir)
	if filesDir == "" {
		filesDir = a.memoFilesDir()
	}
	if err := os.MkdirAll(expandHome(filesDir), 0o755); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.Memo = config.MemoSettings{FilesDir: expandHome(filesDir), RetainDays: retainDays, AutoCleanup: autoCleanup}
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

func (a *App) ImportMemoAttachment(sourcePath, label, kind string) (memo.Attachment, error) {
	sourcePath = strings.TrimSpace(sourcePath)
	label = strings.TrimSpace(label)
	kind = strings.TrimSpace(kind)
	if strings.HasPrefix(sourcePath, "http://") || strings.HasPrefix(sourcePath, "https://") {
		if label == "" {
			label = sourcePath
		}
		return memo.Attachment{
			ID:        fmt.Sprintf("att-%d", time.Now().UnixNano()),
			Kind:      "link",
			Label:     label,
			URL:       sourcePath,
			CreatedAt: time.Now().Unix(),
		}, nil
	}
	sourcePath = expandHome(sourcePath)
	info, err := os.Stat(sourcePath)
	if err != nil {
		return memo.Attachment{}, err
	}
	if info.IsDir() {
		return memo.Attachment{}, fmt.Errorf("附件必须是文件或 URL")
	}
	filesDir := filepath.Join(a.memoFilesDir(), time.Now().Format("2006-01"))
	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		return memo.Attachment{}, err
	}
	id := fmt.Sprintf("att-%d", time.Now().UnixNano())
	ext := filepath.Ext(sourcePath)
	target := filepath.Join(filesDir, id+ext)
	mimeType, err := copyFileWithMime(sourcePath, target)
	if err != nil {
		return memo.Attachment{}, err
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
	return memo.Attachment{
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

func (a *App) DeleteMemoAttachment(memoID, attachmentID string) (memo.Item, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.Item{}, err
	}
	item, ok := st.Memo(memoID)
	if !ok {
		return memo.Item{}, fmt.Errorf("memo not found: %s", memoID)
	}
	next := make([]memo.Attachment, 0, len(item.Attachments))
	for _, att := range item.Attachments {
		if att.ID == attachmentID {
			if att.ManagedPath != "" && isPathInside(att.ManagedPath, a.memoFilesDir()) {
				_ = os.Remove(att.ManagedPath)
			}
			continue
		}
		next = append(next, att)
	}
	item.Attachments = next
	item.UpdatedAt = time.Now().Unix()
	return st.SaveMemo(item)
}

func (a *App) OpenMemoAttachment(att memo.Attachment) error {
	switch {
	case att.URL != "":
		return exec.Command("open", att.URL).Start()
	case att.ManagedPath != "":
		return exec.Command("open", att.ManagedPath).Start()
	case att.SourcePath != "":
		return exec.Command("open", att.SourcePath).Start()
	default:
		return fmt.Errorf("附件没有可打开的位置")
	}
}

func (a *App) GetMemoAttachmentPreview(att memo.Attachment) (MemoAttachmentPreview, error) {
	out := MemoAttachmentPreview{ID: att.ID, Kind: att.Kind, Label: att.Label, Path: att.ManagedPath}
	if att.Kind != "image" || att.ManagedPath == "" || !isPathInside(att.ManagedPath, a.memoFilesDir()) {
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
	cutoff := time.Now().AddDate(0, 0, -settings.RetainDays)
	st, err := a.ensureMemoStore()
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, item := range st.AllMemos() {
		changed := false
		for i := range item.Attachments {
			att := &item.Attachments[i]
			if att.CleanedAt > 0 || att.ManagedPath == "" || !isPathInside(att.ManagedPath, settings.FilesDir) {
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
			_, _ = st.SaveMemo(item)
		}
	}
	return removed, nil
}

func (a *App) ListVocabEntries() ([]memo.VocabEntry, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return nil, err
	}
	entries := st.AllVocab()
	sort.SliceStable(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Word) < strings.ToLower(entries[j].Word)
	})
	return entries, nil
}

func (a *App) SaveVocabEntry(entry memo.VocabEntry) (memo.VocabEntry, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.VocabEntry{}, err
	}
	now := time.Now().Unix()
	if strings.TrimSpace(entry.ID) == "" {
		entry.ID = fmt.Sprintf("word-%d", time.Now().UnixNano())
		entry.CreatedAt = now
	}
	if entry.CreatedAt == 0 {
		entry.CreatedAt = now
	}
	entry.Word = strings.TrimSpace(entry.Word)
	if entry.Word == "" {
		return memo.VocabEntry{}, fmt.Errorf("英文单词不能为空")
	}
	entry.UpdatedAt = now
	if entry.NextReviewAt == 0 {
		entry.NextReviewAt = now
	}
	return st.SaveVocab(entry)
}

func (a *App) DeleteVocabEntry(id string) error {
	st, err := a.ensureMemoStore()
	if err != nil {
		return err
	}
	return st.DeleteVocab(id)
}

func (a *App) ReviewVocabEntry(id, result string) (memo.VocabEntry, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.VocabEntry{}, err
	}
	for _, entry := range st.AllVocab() {
		if entry.ID != id {
			continue
		}
		now := time.Now()
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
		return st.SaveVocab(entry)
	}
	return memo.VocabEntry{}, fmt.Errorf("word not found: %s", id)
}

func (a *App) LookupVocabWord(word string) (VocabLookupResult, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return VocabLookupResult{}, fmt.Errorf("英文单词不能为空")
	}
	endpoint := "https://api.dictionaryapi.dev/api/v2/entries/en/" + url.PathEscape(word)
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return VocabLookupResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return VocabLookupResult{}, fmt.Errorf("词典查询失败: %s", resp.Status)
	}
	var payload []dictionaryEntry
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return VocabLookupResult{}, err
	}
	return parseDictionaryLookup(word, payload), nil
}

func (a *App) SuggestVocabEntry(entry memo.VocabEntry) (memo.VocabEntry, error) {
	a.mu.RLock()
	ai := a.activeAIConfig()
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		return memo.VocabEntry{}, err
	}
	system := `你是 Hatch 的英文单词整理助手。根据用户提供的英文单词和已有信息，补全中文释义、音标、词性、常用语法组合和例句。
只返回 JSON，不要 markdown。格式：
{"word":"","chinese":"","phonetic":"","partOfSpeech":"","collocations":[{"pattern":"","meaning":"","example":""}],"example":""}`
	data, _ := json.Marshal(entry)
	text, err := a.callAIOp("SuggestVocabEntry", ai, system, string(data))
	if err != nil {
		return memo.VocabEntry{}, err
	}
	var out memo.VocabEntry
	if err := json.Unmarshal([]byte(extractJSON(text)), &out); err != nil {
		return memo.VocabEntry{}, err
	}
	if out.Word == "" {
		out.Word = entry.Word
	}
	if out.AudioURL == "" {
		out.AudioURL = entry.AudioURL
	}
	return out, nil
}

type dictionaryEntry struct {
	Word      string `json:"word"`
	Phonetic  string `json:"phonetic"`
	Phonetics []struct {
		Text  string `json:"text"`
		Audio string `json:"audio"`
	} `json:"phonetics"`
	Meanings []struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Definitions  []struct {
			Definition string `json:"definition"`
			Example    string `json:"example"`
		} `json:"definitions"`
	} `json:"meanings"`
}

func parseDictionaryLookup(word string, payload []dictionaryEntry) VocabLookupResult {
	out := VocabLookupResult{Word: word, Source: "dictionaryapi.dev"}
	if len(payload) == 0 {
		return out
	}
	first := payload[0]
	if first.Word != "" {
		out.Word = first.Word
	}
	out.Phonetic = first.Phonetic
	for _, p := range first.Phonetics {
		if out.Phonetic == "" && p.Text != "" {
			out.Phonetic = p.Text
		}
		if out.AudioURL == "" && p.Audio != "" {
			out.AudioURL = p.Audio
		}
	}
	for _, m := range first.Meanings {
		if out.PartOfSpeech == "" {
			out.PartOfSpeech = m.PartOfSpeech
		}
		for _, d := range m.Definitions {
			if d.Definition != "" {
				out.Definitions = append(out.Definitions, d.Definition)
			}
			if d.Example != "" {
				out.Collocations = append(out.Collocations, memo.Collocation{
					Pattern: m.PartOfSpeech,
					Example: d.Example,
				})
			}
			if len(out.Definitions) >= 3 {
				break
			}
		}
		if len(out.Definitions) >= 3 {
			break
		}
	}
	return out
}

func copyFileWithMime(src, dst string) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	var head bytes.Buffer
	tee := io.TeeReader(io.LimitReader(in, 512), &head)
	_, _ = io.Copy(io.Discard, tee)
	mimeType := http.DetectContentType(head.Bytes())
	if extType := mime.TypeByExtension(filepath.Ext(src)); extType != "" && mimeType == "application/octet-stream" {
		mimeType = extType
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return mimeType, err
}

func isPathInside(path, root string) bool {
	path = expandHome(path)
	root = expandHome(root)
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end >= start {
		return s[start : end+1]
	}
	return s
}
