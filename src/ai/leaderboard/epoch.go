package leaderboard

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

const (
	epochECI   = "epoch_capabilities_index/eci_scores.csv"
	epochNames = "epoch_capabilities_index/processed_data_for_eci.csv" // model_version → display name
	maxCSV     = 4 << 20
)

// efforts are the reasoning-effort suffixes Epoch appends to model versions
// ("claude-opus-5-5_max"). "unknown" is dropped rather than shown.
var efforts = map[string]string{"minimal": "Minimal", "low": "Low", "medium": "Medium", "high": "High", "xhigh": "xHigh", "max": "Max", "none": "None", "unknown": ""}

func parseEpoch(body []byte, picks map[string]string) (map[string]boardData, error) {
	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, errors.New("Epoch AI 返回了无效的压缩包")
	}
	files := map[string]*zip.File{}
	for _, f := range archive.File {
		files[f.Name] = f
	}
	names := map[string]string{}
	if rows, err := readCSV(files[epochNames]); err == nil {
		for _, r := range rows {
			if v, m := r["model_version"], r["Model"]; v != "" && m != "" {
				names[v] = m
			}
		}
	}
	out := map[string]boardData{}
	for id, pick := range picks {
		name := pick
		if pick == "eci" {
			name = epochECI
		}
		rows, err := readCSV(files[name])
		if err != nil {
			out[id] = boardData{Error: err.Error()}
			continue
		}
		var board boardData
		if pick == "eci" {
			board = epochIndexBoard(rows)
		} else {
			board = epochBenchmarkBoard(rows, names)
		}
		// Epoch has no per-board publish date; the archive entry's timestamp is when the data was exported.
		if modified := files[name].Modified; !modified.IsZero() {
			board.Published = modified.Format("2006-01-02")
		}
		out[id] = board
	}
	return out, nil
}

func epochIndexBoard(rows []map[string]string) boardData {
	var entries []Entry
	for _, r := range rows {
		score, ok := number(r["eci"])
		model := firstNonEmpty(r["Display name"], r["Model"])
		if !ok || model == "" {
			continue
		}
		e := Entry{Model: model, Score: score, Released: date(r["date"])}
		e.Lower, _ = number(r["eci_ci_low"])
		e.Upper, _ = number(r["eci_ci_high"])
		e.Brand, e.Org = brandOf(r["Organization"], model)
		entries = append(entries, e)
	}
	return ranked(entries)
}

func epochBenchmarkBoard(rows []map[string]string, names map[string]string) boardData {
	var entries []Entry
	for _, r := range rows {
		score, ok := number(r["mean_score"])
		version := r["Model version"]
		if !ok || version == "" || score < 0 || score > 1 {
			continue
		}
		model, variant := version, ""
		if i := strings.LastIndex(version, "_"); i > 0 {
			if label, known := efforts[strings.ToLower(version[i+1:])]; known {
				model, variant = version[:i], label
			}
		}
		if display := names[version]; display != "" {
			model = display
		} else if display := names[model]; display != "" {
			model = display
		}
		e := Entry{Model: model, Variant: variant, Score: score * 100, Released: date(r["Release date"])}
		e.Brand, e.Org = brandOf(r["Organization"], model)
		entries = append(entries, e)
	}
	return ranked(entries)
}

// ranked orders by score with competition ranking (ties share a rank).
func ranked(entries []Entry) boardData {
	if len(entries) == 0 {
		return boardData{Error: "来源暂无该榜单数据"}
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Score > entries[j].Score })
	for i := range entries {
		entries[i].Rank = i + 1
		if i > 0 && entries[i].Score == entries[i-1].Score {
			entries[i].Rank = entries[i-1].Rank
		}
	}
	return boardData{Entries: head(entries), Total: len(entries)}
}

func readCSV(f *zip.File) ([]map[string]string, error) {
	if f == nil {
		return nil, errors.New("Epoch AI 压缩包中缺少该榜单文件")
	}
	if f.UncompressedSize64 > maxCSV {
		return nil, fmt.Errorf("%s 超过解压上限", f.Name)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	reader := csv.NewReader(io.LimitReader(rc, maxCSV))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("%s 表头无效", f.Name)
	}
	for i := range header {
		header[i] = strings.TrimSpace(strings.TrimPrefix(header[i], "\ufeff"))
	}
	var rows []map[string]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s 格式无效", f.Name)
		}
		row := make(map[string]string, len(header))
		for i, key := range header {
			if i < len(record) {
				row[key] = strings.TrimSpace(record[i])
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func number(s string) (float64, bool) {
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil && v == v && v < 1e300 && v > -1e300
}

// date keeps a YYYY-MM-DD release date and drops anything else rather than guessing.
func date(s string) string {
	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		return s[:10]
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
