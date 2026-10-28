package leaderboard

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/parquet-go/parquet-go"
)

// arenaRow covers both schemas in lmarena-ai/leaderboard-dataset: Bradley–Terry
// subsets (rating*, vote_count) and Agent Arena (score*, session_count).
type arenaRow struct {
	Model    string   `parquet:"model_name,optional"`
	Org      string   `parquet:"organization,optional"`
	Rating   *float64 `parquet:"rating,optional"`
	Lower    *float64 `parquet:"rating_lower,optional"`
	Upper    *float64 `parquet:"rating_upper,optional"`
	Votes    *float64 `parquet:"vote_count,optional"`
	Score    *float64 `parquet:"score,optional"`
	ScoreLo  *float64 `parquet:"score_ci_lower,optional"`
	ScoreHi  *float64 `parquet:"score_ci_upper,optional"`
	Sessions *float64 `parquet:"session_count,optional"`
	Rank     *int64   `parquet:"rank,optional"`
	Category string   `parquet:"category,optional"`
	Date     string   `parquet:"leaderboard_publish_date,optional"`
}

func parseArena(body []byte, picks map[string]string) (out map[string]boardData, err error) {
	// A schema change upstream must fail this feed, never crash the app.
	defer func() {
		if recover() != nil {
			out, err = nil, errors.New("Arena 数据结构已变化，无法解析")
		}
	}()
	file, err := parquet.OpenFile(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, errors.New("Arena 返回了无效的 Parquet 文件")
	}
	// Columns absent from the file decode as zero, not nil, so pick the schema explicitly.
	_, bradleyTerry := file.Schema().Lookup("rating")
	rows, err := parquet.Read[arenaRow](bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("Arena 数据解析失败：%w", err)
	}
	byCategory := map[string][]arenaRow{}
	for _, r := range rows {
		byCategory[r.Category] = append(byCategory[r.Category], r)
	}
	out = map[string]boardData{}
	for id, category := range picks {
		out[id] = arenaBoard(byCategory[category], bradleyTerry)
	}
	return out, nil
}

func arenaBoard(rows []arenaRow, bradleyTerry bool) boardData {
	// "latest" files can still carry older rows for a category; keep the newest publish.
	published := ""
	for _, r := range rows {
		if r.Date > published {
			published = r.Date
		}
	}
	var entries []Entry
	for _, r := range rows {
		if r.Date != published || r.Model == "" {
			continue
		}
		e := Entry{Model: r.Model}
		switch {
		case bradleyTerry && finite(r.Rating):
			e.Score, e.Lower, e.Upper = *r.Rating, value(r.Lower), value(r.Upper)
			e.Votes = int(value(r.Votes))
		case !bradleyTerry && finite(r.Score):
			e.Score, e.Lower, e.Upper = *r.Score, value(r.ScoreLo), value(r.ScoreHi)
			e.Votes = int(value(r.Sessions))
		default:
			continue
		}
		if r.Rank != nil && *r.Rank > 0 {
			e.Rank = int(*r.Rank)
		}
		e.Model, e.Variant = splitVariant(e.Model)
		e.Brand, e.Org = brandOf(r.Org, e.Model)
		entries = append(entries, e)
	}
	if len(entries) == 0 {
		return boardData{Error: "来源暂无该榜单数据"}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Rank != b.Rank && a.Rank > 0 && b.Rank > 0 {
			return a.Rank < b.Rank
		}
		return a.Score > b.Score
	})
	// Rows without a source rank get their position; ranks the source gave stay untouched.
	for i := range entries {
		if entries[i].Rank == 0 {
			entries[i].Rank = i + 1
		}
	}
	return boardData{Entries: head(entries), Total: len(entries), Published: published}
}

func finite(p *float64) bool { return p != nil && !math.IsNaN(*p) && !math.IsInf(*p, 0) }

func value(p *float64) float64 {
	if finite(p) {
		return *p
	}
	return 0
}

func head(entries []Entry) []Entry {
	if len(entries) > keepEntries {
		return entries[:keepEntries]
	}
	return entries
}
