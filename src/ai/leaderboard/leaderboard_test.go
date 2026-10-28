package leaderboard

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/parquet-go/parquet-go"
)

func ptr[T any](v T) *T { return &v }

func arenaFixture(t *testing.T, rows []arenaRow) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := parquet.Write(&buf, rows); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func epochFixture(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(content))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestArenaBoardsKeepSourceRankAndSplitVariants(t *testing.T) {
	body := arenaFixture(t, []arenaRow{
		{Model: "Claude Opus 5.5 (High)", Org: "anthropic", Rating: ptr(1548.0), Lower: ptr(1535.2), Upper: ptr(1560.8), Votes: ptr(2314.0), Rank: ptr[int64](2), Category: "coding", Date: "2026-10-22"},
		{Model: "Gemini 4 Argon (High)", Org: "google", Rating: ptr(1551.6), Rank: ptr[int64](1), Category: "coding", Date: "2026-10-22"},
		{Model: "GLM-5.3", Org: "zai", Rating: ptr(1500.0), Rank: ptr[int64](3), Category: "coding", Date: "2026-10-22"},
		{Model: "stale row", Org: "openai", Rating: ptr(9999.0), Rank: ptr[int64](1), Category: "coding", Date: "2025-01-24"},
		{Model: "no score", Org: "openai", Rank: ptr[int64](4), Category: "coding", Date: "2026-10-22"},
		{Model: "other category", Rating: ptr(1.0), Rank: ptr[int64](1), Category: "overall", Date: "2026-10-22"},
	})
	got, err := parseArena(body, map[string]string{"coding": "coding", "missing": "chinese"})
	if err != nil {
		t.Fatal(err)
	}
	coding := got["coding"]
	if coding.Published != "2026-10-22" || coding.Total != 3 || len(coding.Entries) != 3 {
		t.Fatalf("unexpected board: %+v", coding)
	}
	first, second := coding.Entries[0], coding.Entries[1]
	if first.Model != "Gemini 4 Argon" || first.Variant != "High" || first.Brand != "google" || first.Org != "Google" {
		t.Errorf("first entry: %+v", first)
	}
	if second.Rank != 2 || second.Lower != 1535.2 || second.Upper != 1560.8 || second.Votes != 2314 || second.Org != "Anthropic" {
		t.Errorf("second entry: %+v", second)
	}
	if coding.Entries[2].Org != "Z.ai" {
		t.Errorf("zai org not normalized: %+v", coding.Entries[2])
	}
	if got["missing"].Error == "" {
		t.Error("missing category should report an error")
	}
}

// agentRow mirrors the Agent Arena file, which has no rating columns at all.
type agentRow struct {
	Model    string  `parquet:"model_name"`
	Org      string  `parquet:"organization"`
	Score    float64 `parquet:"score"`
	ScoreLo  float64 `parquet:"score_ci_lower"`
	ScoreHi  float64 `parquet:"score_ci_upper"`
	Sessions float64 `parquet:"session_count"`
	Rank     int64   `parquet:"rank"`
	Category string  `parquet:"category"`
	Date     string  `parquet:"leaderboard_publish_date"`
}

func TestArenaAgentSchemaUsesIPSScore(t *testing.T) {
	var buf bytes.Buffer
	err := parquet.Write(&buf, []agentRow{
		{Model: "Claude Opus 5.5 (Max)", Org: "anthropic", Score: 0.183, ScoreLo: 0.17, ScoreHi: 0.2, Sessions: 5000, Rank: 1, Category: "overall", Date: "2026-10-23"},
		{Model: "Below baseline", Org: "openai", Score: -0.02, Rank: 2, Category: "overall", Date: "2026-10-23"},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := buf.Bytes()
	got, err := parseArena(body, map[string]string{"agent": "overall"})
	if err != nil {
		t.Fatal(err)
	}
	e := got["agent"].Entries[0]
	if e.Score != 0.183 || e.Lower != 0.17 || e.Upper != 0.2 || e.Votes != 5000 || e.Variant != "Max" {
		t.Errorf("agent entry: %+v", e)
	}
	if neg := got["agent"].Entries[1]; neg.Score != -0.02 {
		t.Errorf("negative IPS score must survive: %+v", neg)
	}
}

func TestArenaRejectsGarbage(t *testing.T) {
	if _, err := parseArena([]byte("not parquet"), map[string]string{"x": "overall"}); err == nil {
		t.Fatal("expected error")
	}
}

const epochNamesCSV = "model_id,model_version,Model\nm1,gpt-6-astra_max,GPT-6 Astra\nm2,kimi-k3,Kimi K3\n"

func TestEpochBoardsMapNamesEffortsAndTies(t *testing.T) {
	body := epochFixture(t, map[string]string{
		epochNames: epochNamesCSV,
		epochECI: "\ufeffModel,Display name,eci,eci_ci_low,eci_ci_high,date,Organization\n" +
			"Claude Opus 5.5,Claude Opus 5.5,168.83,165.25,173.95,2026-09-22,Anthropic\n" +
			"Gemini 4 Argon,Gemini 4 Argon,166.45,163,170,2026-09-30,\"Google DeepMind,Google\"\n" +
			"No Score,No Score,,,,2026-01-01,Anthropic\n",
		"gpqa_diamond.csv": "Model version,mean_score,Release date,Organization\n" +
			"gpt-6-astra_max,0.9577,2026-09-03,OpenAI\n" +
			"kimi-k3_high,0.9,2026-07-01,Moonshot\n" +
			"mystery-model_unknown,0.9,,\n" +
			"broken,1.7,2026-01-01,OpenAI\n",
	})
	got, err := parseEpoch(body, map[string]string{"eci": "eci", "gpqa": "gpqa_diamond.csv", "gone": "gone.csv"})
	if err != nil {
		t.Fatal(err)
	}
	eci := got["eci"]
	if eci.Total != 2 || eci.Entries[0].Model != "Claude Opus 5.5" || eci.Entries[0].Released != "2026-09-22" || eci.Entries[0].Lower != 165.25 {
		t.Errorf("eci: %+v", eci)
	}
	if eci.Entries[1].Org != "Google" || eci.Entries[1].Brand != "google" {
		t.Errorf("co-developer org: %+v", eci.Entries[1])
	}
	gpqa := got["gpqa"].Entries
	if len(gpqa) != 3 {
		t.Fatalf("gpqa entries: %+v", gpqa)
	}
	if gpqa[0].Model != "GPT-6 Astra" || gpqa[0].Variant != "Max" || gpqa[0].Score < 95.76 || gpqa[0].Score > 95.78 {
		t.Errorf("display name / percent: %+v", gpqa[0])
	}
	if gpqa[1].Model != "Kimi K3" || gpqa[1].Variant != "High" || gpqa[1].Brand != "moonshot" {
		t.Errorf("name mapped via base version: %+v", gpqa[1])
	}
	if gpqa[1].Rank != 2 || gpqa[2].Rank != 2 || gpqa[2].Model != "mystery-model" || gpqa[2].Variant != "" {
		t.Errorf("ties / unknown effort: %+v %+v", gpqa[1], gpqa[2])
	}
	if got["gone"].Error == "" {
		t.Error("missing csv should report an error")
	}
}

// fakeSources serves every feed from one test server and counts requests.
type fakeSources struct {
	calls atomic.Int32
	fail  atomic.Bool
	delay time.Duration
	arena []byte
	epoch []byte
}

func newService(t *testing.T, src *fakeSources) (*Service, string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		src.calls.Add(1)
		time.Sleep(src.delay)
		if src.fail.Load() {
			http.Error(w, "nope", http.StatusNotFound)
			return
		}
		if r.URL.Path == "/epoch" {
			w.Write(src.epoch)
			return
		}
		w.Write(src.arena)
	}))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "boards.json")
	s := New(path)
	for id := range s.urls {
		s.urls[id] = server.URL + "/" + id
	}
	return s, path
}

func defaultSources(t *testing.T) *fakeSources {
	var rows []arenaRow
	for _, category := range []string{"overall", "coding", "chinese", "image_to_webdev"} {
		rows = append(rows, arenaRow{Model: "Model A", Org: "openai", Rating: ptr(1400.0), Score: ptr(0.1), Rank: ptr[int64](1), Category: category, Date: "2026-10-22"})
	}
	return &fakeSources{
		arena: arenaFixture(t, rows),
		epoch: epochFixture(t, map[string]string{
			epochECI:           "Model,eci,date,Organization\nX,150,2026-01-01,OpenAI\n",
			"gpqa_diamond.csv": "Model version,mean_score,Release date,Organization\nx_high,0.5,2026-01-01,OpenAI\n",
		}),
	}
}

func TestRefreshFetchesEachFeedOnceAndCaches(t *testing.T) {
	src := defaultSources(t)
	s, path := newService(t, src)
	snap := s.Refresh(context.Background(), false)
	if got := int(src.calls.Load()); got != len(feeds) {
		t.Fatalf("expected one request per feed (%d), got %d", len(feeds), got)
	}
	if snap.Stale || snap.Refreshing || snap.SyncedAt == 0 {
		t.Errorf("snapshot flags: stale=%v refreshing=%v synced=%d", snap.Stale, snap.Refreshing, snap.SyncedAt)
	}
	for _, b := range snap.Boards {
		if b.Error != "" || len(b.Entries) == 0 {
			t.Errorf("%s: error=%q entries=%d", b.ID, b.Error, len(b.Entries))
		}
	}
	s.Refresh(context.Background(), false)
	if got := int(src.calls.Load()); got != len(feeds) {
		t.Errorf("fresh cache should not refetch, got %d calls", got)
	}
	restarted := New(path)
	if b := restarted.Snapshot().Boards[0]; len(b.Entries) == 0 || b.SyncedAt == 0 {
		t.Errorf("cache not restored after restart: %+v", b)
	}
	s.now = func() time.Time { return time.Now().Add(freshFor + time.Minute) }
	if !s.Snapshot().Stale {
		t.Error("cache older than freshFor should be stale")
	}
}

func TestFailedRefreshKeepsDataAndBacksOff(t *testing.T) {
	src := defaultSources(t)
	s, _ := newService(t, src)
	s.Refresh(context.Background(), false)
	src.fail.Store(true)
	before := src.calls.Load()
	snap := s.Refresh(context.Background(), true)
	if src.calls.Load()-before != int32(len(feeds)) {
		t.Fatalf("force refresh should hit every feed once (4xx is not retried)")
	}
	for _, b := range snap.Boards {
		if b.Error == "" || len(b.Entries) == 0 {
			t.Errorf("%s should keep old entries and report error, got error=%q entries=%d", b.ID, b.Error, len(b.Entries))
		}
	}
	before = src.calls.Load()
	s.Refresh(context.Background(), false)
	if src.calls.Load() != before {
		t.Error("failed feeds must not be retried within retryAfter")
	}
	// Old data still inside freshFor is not refetched; once stale, the backoff has long passed.
	s.now = func() time.Time { return time.Now().Add(freshFor + time.Second) }
	s.Refresh(context.Background(), false)
	if src.calls.Load() == before {
		t.Error("stale failed feeds should be retried after retryAfter")
	}
}

func TestConcurrentRefreshSharesOneFlight(t *testing.T) {
	src := defaultSources(t)
	src.delay = 50 * time.Millisecond
	s, _ := newService(t, src)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); s.Refresh(context.Background(), true) }()
	}
	wg.Wait()
	if got := int(src.calls.Load()); got != len(feeds) {
		t.Errorf("concurrent refreshes should share one flight, got %d calls", got)
	}
}

func TestOldCacheLayoutIsDiscarded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "boards.json")
	os.WriteFile(path, []byte(`{"arena-coding":{"entries":[{"rank":1,"model":"old"}],"lastSuccess":1}}`), 0o600)
	snap := New(path).Snapshot()
	if !snap.Stale || len(snap.Boards[0].Entries) != 0 {
		t.Errorf("legacy cache should be ignored: %+v", snap.Boards[0])
	}
}

func TestDefinitionsReferenceKnownFeeds(t *testing.T) {
	known := map[string]bool{}
	for _, f := range feeds {
		known[f.id] = true
	}
	seen := map[string]bool{}
	for _, d := range definitions {
		if !known[d.feed] || d.pick == "" || seen[d.ID] {
			t.Errorf("bad definition %s", d.ID)
		}
		seen[d.ID] = true
		if d.Source == "Arena" && strings.Contains(d.URL, "scoring=raw") {
			t.Errorf("%s links to the raw board but reads style-controlled data", d.ID)
		}
	}
}

// TestLiveSources hits the real endpoints: HATCH_LEADERBOARD_LIVE=1 go test -run Live -v
func TestLiveSources(t *testing.T) {
	if os.Getenv("HATCH_LEADERBOARD_LIVE") == "" {
		t.Skip("set HATCH_LEADERBOARD_LIVE=1 to query real sources")
	}
	s := New(filepath.Join(t.TempDir(), "boards.json"))
	snap := s.Refresh(context.Background(), true)
	for _, b := range snap.Boards {
		if b.Error != "" || len(b.Entries) < 5 {
			t.Errorf("%s: error=%q entries=%d", b.ID, b.Error, len(b.Entries))
			continue
		}
		top := b.Entries[0]
		t.Logf("%-22s total=%-4d published=%-10s #1 %s %s (%s) %.2f", b.ID, b.Total, b.Published, top.Model, top.Variant, top.Org, top.Score)
	}
}
