package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hatch/ai/llm"
	"hatch/ai/prompts"
	"hatch/core/macos"
	"hatch/memo"
)

// ─── ✓ 备忘录 · 单词本与复盘 ─────────────────────────────────────────────────

// ListVocabEntries returns vocab entries alphabetically.
func (a *App) ListVocabEntries() ([]memo.VocabEntry, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return nil, err
	}
	return st.SortedVocab(), nil
}

func (a *App) SaveVocabEntry(entry memo.VocabEntry) (memo.VocabEntry, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.VocabEntry{}, err
	}
	entry, err = memo.PrepareVocab(entry, time.Now())
	if err != nil {
		return memo.VocabEntry{}, err
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

// ReviewVocabEntry records a review result ("hard" | "known" | "mastered")
// and schedules the next review.
func (a *App) ReviewVocabEntry(id, result string) (memo.VocabEntry, error) {
	st, err := a.ensureMemoStore()
	if err != nil {
		return memo.VocabEntry{}, err
	}
	for _, entry := range st.AllVocab() {
		if entry.ID == id {
			return st.SaveVocab(memo.ApplyReview(entry, result, time.Now()))
		}
	}
	return memo.VocabEntry{}, fmt.Errorf("word not found: %s", id)
}

// LookupVocabWord queries dictionaryapi.dev for phonetic, audio and definitions.
func (a *App) LookupVocabWord(word string) (memo.LookupResult, error) {
	return memo.LookupWord(word)
}

// SpeakVocabWord pronounces a word with macOS `say` (fallback when the
// dictionary audio URL is missing or fails to play).
func (a *App) SpeakVocabWord(word string) error {
	word = strings.TrimSpace(word)
	if word == "" {
		return fmt.Errorf("英文单词不能为空")
	}
	if runes := []rune(word); len(runes) > 120 {
		word = string(runes[:120])
	}
	return macos.Say(word)
}

// SuggestVocabEntry asks AI to fill Chinese meaning, phonetic, collocations
// and an example; the draft is returned for confirmation, not saved.
func (a *App) SuggestVocabEntry(entry memo.VocabEntry) (memo.VocabEntry, error) {
	ai, err := a.readyAI()
	if err != nil {
		return memo.VocabEntry{}, err
	}
	data, _ := json.Marshal(entry)
	text, err := a.callAIOp(prompts.OpSuggestVocabEntry, ai, prompts.VocabEntry, string(data))
	if err != nil {
		return memo.VocabEntry{}, err
	}
	var out memo.VocabEntry
	if err := json.Unmarshal([]byte(llm.ExtractJSONObject(text)), &out); err != nil {
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
