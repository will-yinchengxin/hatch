package memo

import "testing"

func TestParseDictionaryLookup(t *testing.T) {
	got := parseDictionaryLookup("resilient", []dictionaryEntry{{
		Word:     "resilient",
		Phonetic: "/rɪˈzɪliənt/",
		Phonetics: []struct {
			Text  string `json:"text"`
			Audio string `json:"audio"`
		}{{Audio: "https://example.test/resilient.mp3"}},
		Meanings: []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions  []struct {
				Definition string `json:"definition"`
				Example    string `json:"example"`
			} `json:"definitions"`
		}{{
			PartOfSpeech: "adjective",
			Definitions: []struct {
				Definition string `json:"definition"`
				Example    string `json:"example"`
			}{{Definition: "able to recover quickly", Example: "The system is resilient."}},
		}},
	}})
	if got.Word != "resilient" || got.Phonetic == "" || got.AudioURL == "" || got.PartOfSpeech != "adjective" {
		t.Fatalf("unexpected lookup result: %#v", got)
	}
	if len(got.Definitions) != 1 || len(got.Collocations) != 1 {
		t.Fatalf("expected definition and example colocation: %#v", got)
	}
}
