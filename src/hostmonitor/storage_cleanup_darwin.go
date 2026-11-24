package hostmonitor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

const (
	cleanupMinBytes   = 50 << 20
	cleanupScanTTL    = 15 * time.Minute
	cleanupPreviewTTL = 2 * time.Minute
	cleanupTreeLimit  = 50000
	cleanupDepthLimit = 32
)

type cleanupRule struct {
	name, relative string
	userFile       bool
}

// These exact paths contain regenerable download/build artifacts. Broader
// cache directories and custom tool locations only appear in the inventory.
var cleanupRules = [...]cleanupRule{
	{name: "npm 下载缓存", relative: ".npm/_cacache"},
	{name: "pip HTTP 缓存", relative: "Library/Caches/pip/http-v2"},
	{name: "pip 旧版 HTTP 缓存", relative: "Library/Caches/pip/http"},
	{name: "pip Wheel 缓存", relative: "Library/Caches/pip/wheels"},
	{name: "Go 构建缓存", relative: "Library/Caches/go-build"},
}

type StorageCleanupItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Bytes    uint64 `json:"bytes"`
	UserFile bool   `json:"userFile,omitempty"`
}

type StorageCleanupPreview struct {
	Token string               `json:"token"`
	Items []StorageCleanupItem `json:"items"`
}

type StorageCleanupResult struct {
	StorageCleanupItem
	Status    string `json:"status"`
	TrashPath string `json:"trashPath,omitempty"`
	Message   string `json:"message,omitempty"`
}

type cleanupTarget struct {
	item                StorageCleanupItem
	rule                cleanupRule
	home, parent, entry unix.Stat_t
	tree                [sha256.Size]byte
}

type storageCleanupSession struct {
	scanID         string
	expires        time.Time
	targets        map[string]cleanupTarget
	previewToken   string
	previewExpires time.Time
	selected       []cleanupTarget
}

// StorageAnalyzer retains at most one scan and one confirmation in memory.
type StorageAnalyzer struct {
	mu      sync.Mutex
	session storageCleanupSession
}

func storageToken() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func (a *StorageAnalyzer) Preview(scanID string, ids []string) (StorageCleanupPreview, error) {
	if !a.mu.TryLock() {
		return StorageCleanupPreview{}, errors.New("磁盘分析或清理正在进行")
	}
	defer a.mu.Unlock()
	return a.session.preview(scanID, ids)
}

func (s *storageCleanupSession) preview(scanID string, ids []string) (StorageCleanupPreview, error) {
	s.previewToken, s.selected = "", nil
	if scanID == "" || scanID != s.scanID || !time.Now().Before(s.expires) {
		return StorageCleanupPreview{}, errors.New("扫描结果已失效，请重新分析")
	}
	if len(ids) == 0 || len(ids) > len(s.targets) {
		return StorageCleanupPreview{}, errors.New("请选择本次扫描中的可清理项目")
	}
	var selected []cleanupTarget
	items := make([]StorageCleanupItem, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		target, ok := s.targets[id]
		if !ok || seen[id] {
			return StorageCleanupPreview{}, errors.New("选择包含无效或重复的项目，请重新选择")
		}
		selected = append(selected, target)
		items = append(items, target.item)
		seen[id] = true
	}
	token, err := storageToken()
	if err != nil {
		return StorageCleanupPreview{}, err
	}
	s.previewToken, s.selected = token, selected
	s.previewExpires = time.Now().Add(cleanupPreviewTTL)
	return StorageCleanupPreview{Token: token, Items: items}, nil
}

func (a *StorageAnalyzer) Apply(token string) ([]StorageCleanupResult, error) {
	if !a.mu.TryLock() {
		return nil, errors.New("磁盘分析或清理正在进行")
	}
	defer a.mu.Unlock()
	home, err := storageAccountHome()
	if err != nil {
		return nil, err
	}
	return a.session.apply(home, token, trashStorageCache)
}

type storageTrashMover func(path string, fd int) (string, error)

var errStorageOutcomeUncertain = errors.New("无法核实移动结果，请检查原路径和废纸篓")

func (s *storageCleanupSession) apply(home, token string, trashMove storageTrashMover) ([]StorageCleanupResult, error) {
	now := time.Now()
	if token == "" || token != s.previewToken || !now.Before(s.expires) || !now.Before(s.previewExpires) {
		return nil, errors.New("清理确认已失效，请重新分析和选择")
	}
	selected := s.selected
	s.previewToken, s.selected = "", nil
	results := make([]StorageCleanupResult, 0, len(selected))
	for _, target := range selected {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		path, err := moveCleanupTarget(ctx, home, target, trashMove)
		cancel()
		result := StorageCleanupResult{StorageCleanupItem: target.item, Status: "moved", TrashPath: path}
		if err != nil {
			result.Status = "skipped"
			result.Message = err.Error()
			if path != "" || errors.Is(err, errStorageOutcomeUncertain) {
				result.Status = "review"
			}
		}
		if result.Status != "skipped" {
			delete(s.targets, target.item.ID)
		}
		results = append(results, result)
	}
	return results, nil
}

const cleanupOpenFlags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW
const protectedStorageFlags = unix.UF_IMMUTABLE | unix.UF_APPEND | unix.UF_DATAVAULT |
	unix.SF_IMMUTABLE | unix.SF_APPEND | unix.SF_NOUNLINK | unix.SF_RESTRICTED |
	unix.SF_DATALESS | unix.SF_FIRMLINK

func safeCleanupStat(stat unix.Stat_t, device int32) bool {
	return stat.Dev == device && stat.Uid == uint32(os.Getuid()) &&
		stat.Mode&0022 == 0 && stat.Flags&protectedStorageFlags == 0
}

func sameStorageNode(a, b unix.Stat_t) bool {
	return a.Dev == b.Dev && a.Ino == b.Ino && a.Btim == b.Btim
}

func sameStorageMetadata(a, b unix.Stat_t) bool {
	return sameStorageNode(a, b) && a.Mode == b.Mode && a.Uid == b.Uid && a.Gid == b.Gid &&
		a.Flags == b.Flags && a.Nlink == b.Nlink && a.Size == b.Size && a.Blocks == b.Blocks &&
		a.Mtim == b.Mtim && a.Ctim == b.Ctim
}

func openCleanupHome(path string) (*os.File, unix.Stat_t, error) {
	var stat unix.Stat_t
	fd, err := unix.Open(path, (cleanupOpenFlags&^unix.O_NOFOLLOW)|unix.O_NOFOLLOW_ANY, 0)
	if err != nil {
		return nil, stat, err
	}
	var fs unix.Statfs_t
	if err := unix.Fstat(fd, &stat); err != nil || !safeCleanupStat(stat, stat.Dev) ||
		unix.Fstatfs(fd, &fs) != nil || fs.Flags&unix.MNT_LOCAL == 0 {
		unix.Close(fd)
		return nil, stat, errors.New("用户目录的权限或卷不符合安全要求")
	}
	return os.NewFile(uintptr(fd), path), stat, nil
}

func openCleanupDirAt(parent *os.File, name string, device int32) (*os.File, unix.Stat_t, error) {
	return openCleanupEntryAt(parent, name, device, false)
}

func openCleanupEntryAt(parent *os.File, name string, device int32, userFile bool) (*os.File, unix.Stat_t, error) {
	var stat unix.Stat_t
	flags := cleanupOpenFlags
	if userFile {
		// Check the type before opening and use O_NONBLOCK in case a
		// concurrently replaced entry is a FIFO.
		flags = (flags &^ unix.O_DIRECTORY) | unix.O_NONBLOCK
		if err := unix.Fstatat(int(parent.Fd()), name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return nil, stat, err
		}
		if !safeCleanupFile(stat, device) {
			return nil, stat, errors.New("文件类型、所有者或权限不符合安全要求")
		}
	}
	fd, err := unix.Openat(int(parent.Fd()), name, flags, 0)
	if err != nil {
		return nil, stat, err
	}
	if err := unix.Fstat(fd, &stat); err != nil || !safeCleanupStat(stat, device) || (userFile && !safeCleanupFile(stat, device)) {
		unix.Close(fd)
		return nil, stat, errors.New("项目的类型、所有者、权限或卷不符合安全要求")
	}
	return os.NewFile(uintptr(fd), name), stat, nil
}

func safeCleanupFile(stat unix.Stat_t, device int32) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFREG && stat.Nlink == 1 && safeCleanupStat(stat, device)
}

func openCleanupParent(home *os.File, relative string, homeStat unix.Stat_t) (*os.File, unix.Stat_t, error) {
	var parent *os.File
	stat := homeStat
	for _, name := range strings.Split(filepath.Dir(relative), string(filepath.Separator)) {
		from := home
		if parent != nil {
			from = parent
		}
		next, nextStat, err := openCleanupDirAt(from, name, homeStat.Dev)
		if parent != nil {
			parent.Close()
		}
		if err != nil {
			return nil, stat, err
		}
		parent, stat = next, nextStat
	}
	return parent, stat, nil
}

func inspectCleanupTarget(ctx context.Context, homePath string, rule cleanupRule) (cleanupTarget, error) {
	target := cleanupTarget{rule: rule, item: StorageCleanupItem{Name: rule.name, Path: filepath.Join(homePath, rule.relative), UserFile: rule.userFile}}
	home, stat, err := openCleanupHome(homePath)
	if err != nil {
		return target, fmt.Errorf("检查用户目录：%w", err)
	}
	defer home.Close()
	target.home = stat
	parent, stat, err := openCleanupParent(home, rule.relative, target.home)
	if err != nil {
		return target, fmt.Errorf("检查缓存上级目录：%w", err)
	}
	defer parent.Close()
	target.parent = stat
	target.entry, target.tree, err = inspectCleanupEntry(ctx, parent, filepath.Base(rule.relative), stat.Dev, rule.userFile)
	return target, err
}

func inspectCleanupEntry(ctx context.Context, parent *os.File, name string, dev int32, userFile bool) (unix.Stat_t, [sha256.Size]byte, error) {
	var digest [sha256.Size]byte
	dir, stat, err := openCleanupEntryAt(parent, name, dev, userFile)
	if err != nil {
		return stat, digest, err
	}
	defer dir.Close()
	if userFile {
		return stat, digest, ctx.Err()
	}
	hasher := sha256.New()
	remaining := cleanupTreeLimit
	if err := hashCleanupDirectory(ctx, dir, stat, hasher, &remaining, 0); err != nil {
		return stat, digest, err
	}
	copy(digest[:], hasher.Sum(nil))
	return stat, digest, nil
}

// Walk directory descriptors without following links. Hash metadata only so
// a changed cache can be rejected without reading or uploading file contents.
func hashCleanupDirectory(ctx context.Context, dir *os.File, original unix.Stat_t, digest hash.Hash, remaining *int, depth int) error {
	if depth > cleanupDepthLimit {
		return errors.New("目录层级过多")
	}
	entries, err := dir.ReadDir(*remaining + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	*remaining -= len(entries)
	if *remaining < 0 {
		return errors.New("缓存文件数量超过安全检查上限")
	}
	slices.SortFunc(entries, func(a, b os.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return errors.New("缓存复核超时")
		}
		var stat unix.Stat_t
		if err := unix.Fstatat(int(dir.Fd()), entry.Name(), &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		if !safeCleanupStat(stat, original.Dev) {
			return errors.New("缓存含有权限异常、受保护或跨卷的项目")
		}
		fmt.Fprintf(digest, "%d %q %d %d %d %d %d %d %d %v %v %v\n", depth, entry.Name(), stat.Ino,
			stat.Mode, stat.Uid, stat.Gid, stat.Flags, stat.Size, stat.Blocks, stat.Btim, stat.Mtim, stat.Ctim)
		switch stat.Mode & unix.S_IFMT {
		case unix.S_IFDIR:
			child, opened, err := openCleanupDirAt(dir, entry.Name(), original.Dev)
			if err != nil {
				return err
			}
			if !sameStorageMetadata(stat, opened) {
				child.Close()
				return errors.New("缓存内容在检查时发生变化")
			}
			err = hashCleanupDirectory(ctx, child, opened, digest, remaining, depth+1)
			child.Close()
			if err != nil {
				return err
			}
		case unix.S_IFREG:
			if stat.Nlink != 1 {
				return errors.New("缓存含有硬链接")
			}
		default:
			return errors.New("缓存含有符号链接或特殊文件")
		}
	}
	var current unix.Stat_t
	if unix.Fstat(int(dir.Fd()), &current) != nil || !sameStorageMetadata(original, current) {
		return errors.New("缓存内容在检查时发生变化")
	}
	return nil
}

func checkStorageTrash(home *os.File, dev int32) error {
	// TCC can deny enumeration even when the system Trash API can move a
	// file there. Check the directory's identity and mode without opening it.
	var stat unix.Stat_t
	if err := unix.Fstatat(int(home.Fd()), ".Trash", &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return fmt.Errorf("废纸篓不可用：%w", err)
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFDIR || !safeCleanupStat(stat, dev) {
		return errors.New("废纸篓的权限或卷不符合安全要求")
	}
	return nil
}

func moveCleanupTarget(ctx context.Context, homePath string, target cleanupTarget, trashMove storageTrashMover) (string, error) {
	home, stat, err := openCleanupHome(homePath)
	if err != nil {
		return "", err
	}
	defer home.Close()
	if !sameStorageNode(stat, target.home) {
		return "", errors.New("用户目录已变化，请重新分析")
	}

	parent, parentStat, err := openCleanupParent(home, target.rule.relative, stat)
	if err != nil {
		return "", err
	}
	defer parent.Close()
	if !sameStorageNode(parentStat, target.parent) {
		return "", errors.New("项目的上级目录已变化，请重新分析")
	}
	name := filepath.Base(target.rule.relative)
	entry, digest, err := inspectCleanupEntry(ctx, parent, name, stat.Dev, target.rule.userFile)
	if err != nil {
		return "", err
	}
	if !sameStorageMetadata(entry, target.entry) || digest != target.tree || ctx.Err() != nil {
		return "", errors.New("项目内容已变化或复核超时，请重新分析")
	}

	// Reopen the fixed path and keep the checked directory descriptor alive
	// while the system moves that exact object by file reference.
	currentParent, currentParentStat, err := openCleanupParent(home, target.rule.relative, stat)
	if err != nil {
		return "", err
	}
	defer currentParent.Close()
	leaf, current, err := openCleanupEntryAt(currentParent, name, stat.Dev, target.rule.userFile)
	if err != nil {
		return "", err
	}
	defer leaf.Close()
	if !sameStorageNode(currentParentStat, parentStat) || !sameStorageMetadata(entry, current) || ctx.Err() != nil {
		return "", errors.New("清理路径在复核后发生变化")
	}
	if err := checkStorageTrash(home, stat.Dev); err != nil {
		return "", err
	}
	path, err := trashMove(filepath.Join(homePath, target.rule.relative), int(leaf.Fd()))
	if err != nil {
		return path, err
	}
	if filepath.Dir(path) != filepath.Join(homePath, ".Trash") ||
		unix.Lstat(path, &current) != nil || !sameStorageNode(entry, current) {
		return path, errStorageOutcomeUncertain
	}
	return path, nil
}
