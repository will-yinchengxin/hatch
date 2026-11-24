package hostmonitor

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

const (
	storageScanTimeout    = 75 * time.Second
	storageEntryTimeout   = 8 * time.Second
	storageEntryLimit     = 256
	storageDisplayLimit   = 12
	developerStorageGroup = "开发工具数据"
)

type StorageReport struct {
	ID          string               `json:"id"`
	Groups      []StorageGroup       `json:"groups"`
	Candidates  []StorageCleanupItem `json:"candidates"`
	CleanupNote string               `json:"cleanupNote,omitempty"`
}

type StorageGroup struct {
	Name    string `json:"name"`
	path    string
	Entries []StorageUsage `json:"entries"`
	Partial bool           `json:"partial"`
	More    int            `json:"more"`
}

type StorageUsage struct {
	Path      string `json:"path"`
	Bytes     uint64 `json:"bytes"`
	Partial   bool   `json:"partial"`
	CleanupID string `json:"cleanupId,omitempty"`
	source    *unix.Stat_t
}

func storageAccountHome() (string, error) {
	uid := os.Getuid()
	if uid == 0 || uid != os.Geteuid() {
		return "", errors.New("仅支持以普通用户运行的本机清理")
	}
	account, err := user.LookupId(strconv.Itoa(uid))
	if err != nil || !filepath.IsAbs(account.HomeDir) || filepath.Clean(account.HomeDir) != account.HomeDir ||
		filepath.Dir(account.HomeDir) != "/Users" || account.HomeDir == "/Users/Shared" {
		return "", errors.New("无法核实当前账户的本机用户目录")
	}
	return account.HomeDir, nil
}

func (a *StorageAnalyzer) Scan() (StorageReport, error) {
	if !a.mu.TryLock() {
		return StorageReport{}, errors.New("磁盘分析或清理正在进行")
	}
	defer a.mu.Unlock()
	a.session = storageCleanupSession{}

	id, err := storageToken()
	if err != nil {
		return StorageReport{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), storageScanTimeout)
	defer cancel()
	report := StorageReport{ID: id, Candidates: []StorageCleanupItem{}}
	targets := make(map[string]cleanupTarget)
	home, homeErr := storageAccountHome()
	if homeErr != nil {
		report.CleanupNote = homeErr.Error()
	} else if !nativeStorageTrashAvailable {
		report.CleanupNote = "当前构建缺少 macOS 废纸篓支持。"
	} else {
		// Evaluate fixed targets first so wide, read-only inventory cannot
		// prevent the known cache locations from being inspected.
		cacheCtx, cancelCache := context.WithTimeout(ctx, 16*time.Second)
		report.Candidates, report.CleanupNote = scanCleanupCandidates(cacheCtx, home, targets)
		cancelCache()
	}
	report.Groups = scanStorageGroups(ctx, home)
	if home != "" && nativeStorageTrashAvailable {
		linkCtx, cancelLink := context.WithTimeout(context.Background(), time.Second)
		attachStorageActions(linkCtx, home, &report, targets)
		cancelLink()
	}
	expires := time.Now().Add(cleanupScanTTL)
	a.session = storageCleanupSession{scanID: id, targets: targets, expires: expires}
	return report, nil
}

func attachStorageActions(ctx context.Context, homePath string, report *StorageReport, targets map[string]cleanupTarget) {
	cacheIDs := make(map[string]string, len(report.Candidates))
	for _, item := range report.Candidates {
		cacheIDs[item.Path] = item.ID
	}
	home, stat, err := openCleanupHome(homePath)
	allowDownloads := err == nil
	if allowDownloads {
		allowDownloads = checkStorageTrash(home, stat.Dev) == nil
		home.Close()
	}
	for i := range report.Groups {
		group := &report.Groups[i]
		for j := range group.Entries {
			entry := &group.Entries[j]
			if id := cacheIDs[entry.Path]; id != "" && !entry.Partial {
				entry.CleanupID = id
			}
			if !allowDownloads || group.path != filepath.Join(homePath, "Downloads") {
				continue
			}
			rule, ok := downloadCleanupRule(homePath, *entry)
			if !ok {
				continue
			}
			target, err := inspectCleanupTarget(ctx, homePath, rule)
			if err != nil || !sameStorageMetadata(*entry.source, target.entry) {
				continue
			}
			id, err := storageToken()
			if err != nil {
				continue
			}
			entry.CleanupID = id
			target.item.ID, target.item.Bytes = id, entry.Bytes
			targets[id] = target
		}
		if group.Name == developerStorageGroup {
			for _, item := range report.Candidates {
				if idx := slices.IndexFunc(group.Entries, func(entry StorageUsage) bool { return entry.Path == item.Path }); idx >= 0 {
					group.Entries[idx] = StorageUsage{Path: item.Path, Bytes: item.Bytes, CleanupID: item.ID}
				} else {
					// A parent such as ~/.npm can contain personal settings.
					// Surface the exact checked leaf as its own row.
					group.Entries = append(group.Entries, StorageUsage{Path: item.Path, Bytes: item.Bytes, CleanupID: item.ID})
				}
			}
			sort.SliceStable(group.Entries, func(a, b int) bool { return group.Entries[a].Bytes > group.Entries[b].Bytes })
		}
	}
}

func downloadCleanupRule(home string, entry StorageUsage) (cleanupRule, bool) {
	name := filepath.Base(entry.Path)
	if entry.Partial || entry.source == nil || filepath.Dir(entry.Path) != filepath.Join(home, "Downloads") ||
		strings.HasPrefix(name, ".") || !utf8.ValidString(name) ||
		strings.ContainsFunc(name, func(r rune) bool { return unicode.IsControl(r) || unicode.Is(unicode.Cf, r) }) {
		return cleanupRule{}, false
	}
	return cleanupRule{name: name, relative: filepath.Join("Downloads", name), userFile: true}, true
}

func scanCleanupCandidates(ctx context.Context, homePath string, targets map[string]cleanupTarget) ([]StorageCleanupItem, string) {
	items := []StorageCleanupItem{}
	home, stat, err := openCleanupHome(homePath)
	if err != nil {
		return items, "无法安全访问用户目录，清理不可用。"
	}
	err = checkStorageTrash(home, stat.Dev)
	home.Close()
	if err != nil {
		return items, "无法核实同卷废纸篓，清理不可用。"
	}
	incomplete := false
	for _, rule := range cleanupRules {
		path := filepath.Join(homePath, rule.relative)
		// Check before invoking du, so even these fixed paths cannot opt an
		// external mount or a symlinked cache into the cleanup workflow.
		before, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil || !before.IsDir() {
			incomplete = true
			continue
		}
		target, err := inspectCleanupTarget(ctx, homePath, rule)
		if err != nil {
			incomplete = true
			continue
		}
		usage := measureStorage(ctx, path)
		var after unix.Stat_t
		if usage.Partial || unix.Lstat(path, &after) != nil || !sameStorageMetadata(target.entry, after) {
			incomplete = true
			continue
		}
		if usage.Bytes < cleanupMinBytes {
			continue
		}
		id, err := storageToken()
		if err != nil {
			incomplete = true
			continue
		}
		target.item.ID, target.item.Bytes = id, usage.Bytes
		targets[id] = target
		items = append(items, target.item)
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Bytes > items[j].Bytes })
	if incomplete {
		return items, "部分已知缓存未通过完整检查，已从清理清单排除。"
	}
	return items, ""
}

func scanStorageGroups(ctx context.Context, home string) []StorageGroup {
	groups := []StorageGroup{
		{Name: "根目录", path: "/"},
		{Name: "已安装应用", path: "/Applications"},
	}
	if home != "" {
		groups = append(groups, StorageGroup{Name: "用户目录", path: home})
		for _, location := range []struct{ name, relative string }{
			{"下载", "Downloads"},
			{"用户资源库", "Library"},
			{"应用缓存", "Library/Caches"},
			{"应用容器", "Library/Containers"},
			{"共享容器", "Library/Group Containers"},
			{"应用支持", "Library/Application Support"},
		} {
			groups = append(groups, StorageGroup{Name: location.name, path: filepath.Join(home, location.relative)})
		}
	}
	longest := 0
	for i := range groups {
		listStorageGroup(&groups[i])
		longest = max(longest, len(groups[i].Entries))
	}
	if home != "" {
		dev := StorageGroup{Name: developerStorageGroup, Entries: []StorageUsage{}}
		for _, relative := range []string{
			".npm", ".cache", ".cargo", ".pnpm-store", ".gradle", ".m2", ".docker", "go/pkg",
			"Library/Caches/pip", "Library/Caches/uv", "Library/Caches/go-build", "Library/pnpm",
			"Library/Developer/Xcode/DerivedData", "Library/Developer/CoreSimulator", "Library/Developer/Xcode/iOS DeviceSupport",
		} {
			path := filepath.Join(home, relative)
			info, err := os.Lstat(path)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil || !info.IsDir() || !sameStorageVolume(home, path) {
				dev.Partial = true
				continue
			}
			dev.Entries = append(dev.Entries, StorageUsage{Path: path})
		}
		groups = append(groups, dev)
		longest = max(longest, len(dev.Entries))
	}

	jobs := make(chan *StorageUsage)
	var workers sync.WaitGroup
	for i := 0; i < storageWorkers; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for entry := range jobs {
				measured := measureStorage(ctx, entry.Path)
				entry.Bytes, entry.Partial = measured.Bytes, measured.Partial
			}
		}()
	}
	// Interleave groups to give each a chance within the shared time budget.
	for i := 0; i < longest; i++ {
		for j := range groups {
			if i < len(groups[j].Entries) {
				jobs <- &groups[j].Entries[i]
			}
		}
	}
	close(jobs)
	workers.Wait()

	for i := range groups {
		group := &groups[i]
		visible := group.Entries[:0]
		for _, entry := range group.Entries {
			group.Partial = group.Partial || entry.Partial
			if entry.Bytes >= cleanupMinBytes || (group.path == "/" && entry.Bytes > 0) {
				visible = append(visible, entry)
			}
		}
		sort.SliceStable(visible, func(a, b int) bool { return visible[a].Bytes > visible[b].Bytes })
		if len(visible) > storageDisplayLimit {
			group.More = len(visible) - storageDisplayLimit
			visible = visible[:storageDisplayLimit]
		}
		group.Entries = visible
	}
	return groups
}

func listStorageGroup(group *StorageGroup) {
	group.Entries = []StorageUsage{}
	fd, err := unix.Open(group.path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW_ANY, 0)
	if err != nil {
		group.Partial = !errors.Is(err, os.ErrNotExist)
		return
	}
	dir := os.NewFile(uintptr(fd), group.path)
	defer dir.Close()
	var base, parent unix.Stat_t
	var volume unix.Statfs_t
	if unix.Fstat(fd, &base) != nil || unix.Fstatfs(fd, &volume) != nil || volume.Flags&unix.MNT_LOCAL == 0 ||
		(group.path != "/Applications" && (unix.Stat(filepath.Dir(group.path), &parent) != nil || base.Dev != parent.Dev)) {
		group.Partial = true
		return
	}
	entries, err := dir.ReadDir(storageEntryLimit + 1)
	group.Partial = err != nil && !errors.Is(err, io.EOF)
	if len(entries) > storageEntryLimit {
		group.Partial = true
		entries = entries[:storageEntryLimit]
	}
	for _, entry := range entries {
		var child unix.Stat_t
		if unix.Fstatat(fd, entry.Name(), &child, unix.AT_SYMLINK_NOFOLLOW) != nil {
			group.Partial = true
			continue
		}
		if child.Mode&unix.S_IFMT != unix.S_IFDIR && child.Mode&unix.S_IFMT != unix.S_IFREG {
			continue
		}
		// APFS presents these local Data-volume directories at the root.
		path := filepath.Join(group.path, entry.Name())
		if child.Dev != base.Dev && path != "/Applications" && path != "/Library" &&
			path != "/Users" && path != "/private" && path != "/usr/local" {
			continue
		}
		usage := StorageUsage{Path: path}
		if child.Mode&unix.S_IFMT == unix.S_IFREG {
			usage.source = &child
		}
		group.Entries = append(group.Entries, usage)
	}
}

func sameStorageVolume(root, path string) bool {
	var base, entry unix.Stat_t
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW_ANY, 0)
	if err != nil {
		return false
	}
	defer unix.Close(fd)
	return unix.Stat(root, &base) == nil && unix.Fstat(fd, &entry) == nil && base.Dev == entry.Dev
}

func measureStorage(ctx context.Context, path string) StorageUsage {
	ctx, cancel := context.WithTimeout(ctx, storageEntryTimeout)
	defer cancel()
	if ctx.Err() != nil {
		return StorageUsage{Path: path, Partial: true}
	}
	output, err := exec.CommandContext(ctx, "/usr/bin/du", "-skx", path).Output()
	bytes, parseErr := parseDiskUsage(string(output))
	return StorageUsage{Path: path, Bytes: bytes, Partial: err != nil || parseErr != nil}
}
