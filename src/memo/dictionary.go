package memo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LookupResult is a dictionaryapi.dev lookup flattened for the vocab form.
type LookupResult struct {
	Word         string        `json:"word"`
	Phonetic     string        `json:"phonetic,omitempty"`
	AudioURL     string        `json:"audioUrl,omitempty"`
	PartOfSpeech string        `json:"partOfSpeech,omitempty"`
	Definitions  []string      `json:"definitions,omitempty"`
	Collocations []Collocation `json:"collocations,omitempty"`
	Source       string        `json:"source,omitempty"`
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

// LookupWord queries the free dictionaryapi.dev English dictionary.
func LookupWord(word string) (LookupResult, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return LookupResult{}, fmt.Errorf("英文单词不能为空")
	}
	endpoint := "https://api.dictionaryapi.dev/api/v2/entries/en/" + url.PathEscape(word)
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return LookupResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return LookupResult{}, fmt.Errorf("词典查询失败: %s", resp.Status)
	}
	var payload []dictionaryEntry
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return LookupResult{}, err
	}
	return parseDictionaryLookup(word, payload), nil
}

// parseDictionaryLookup keeps the first entry's phonetic/audio, its first part
// of speech and up to three definitions (examples become collocations).
func parseDictionaryLookup(word string, payload []dictionaryEntry) LookupResult {
	out := LookupResult{Word: word, Source: "dictionaryapi.dev"}
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
				out.Collocations = append(out.Collocations, Collocation{
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
