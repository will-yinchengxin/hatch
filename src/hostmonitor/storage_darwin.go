package hostmonitor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type StorageEntry struct {
	Name    string `json:"name"`
	Bytes   uint64 `json:"bytes"`
	Partial bool   `json:"partial"`
}

func parseDiskUsage(output string) (uint64, error) {
	fields := strings.Fields(output)
	if len(fields) < 2 {
		return 0, fmt.Errorf("磁盘分析结果不完整")
	}
	blocks, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil || blocks > ^uint64(0)/1024 {
		return 0, fmt.Errorf("磁盘分析结果无效")
	}
	return blocks * 1024, nil
}

func AnalyzeHostStorage() ([]StorageEntry, error) {
	return analyzeStorageRoot("/", func(path string) ([]byte, error) {
		return command(30*time.Second, "/usr/bin/du", "-skx", path)
	})
}

func analyzeStorageRoot(root string, usage func(string) ([]byte, error)) ([]StorageEntry, error) {
	children, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	entries := make([]StorageEntry, 0, len(children))
	for _, child := range children {
		// Symlinks are aliases on macOS; following them would count data twice.
		if child.Type()&os.ModeSymlink != 0 {
			continue
		}
		name := child.Name()
		data, runErr := usage(filepath.Join(root, name))
		bytes, parseErr := parseDiskUsage(string(data))
		entries = append(entries, StorageEntry{Name: name, Bytes: bytes, Partial: runErr != nil || parseErr != nil})
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].Bytes > entries[right].Bytes })
	return entries, nil
}
