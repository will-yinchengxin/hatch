// Package fsutil holds small filesystem helpers shared by every Hatch module:
// Hatch data-dir resolution, timestamped backups, recursive copies and path
// containment checks.
package fsutil

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AppDataDir returns ~/Library/Application Support/Hatch joined with elem.
// It does not create the directory.
func AppDataDir(elem ...string) (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{cfgDir, "Hatch"}, elem...)...), nil
}

// AppDataPath is AppDataDir without the error: when the user config dir cannot
// be resolved it falls back to $TMPDIR/hatch so callers always get a usable path.
func AppDataPath(elem ...string) string {
	if dir, err := AppDataDir(elem...); err == nil {
		return dir
	}
	return filepath.Join(append([]string{os.TempDir(), "hatch"}, elem...)...)
}

// EnsureDir creates dir (and parents) and returns it.
func EnsureDir(dir string) (string, error) {
	return dir, os.MkdirAll(dir, 0o755)
}

// Backup copies path to "<path>.hatch-backup-<YYYYMMDD-HHMMSS>" (with a numeric
// suffix on collision) and returns the backup path.
func Backup(path string) (string, error) {
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()

	stamp := time.Now().Format("20060102-150405")
	for i := 0; ; i++ {
		suffix := stamp
		if i > 0 {
			suffix = fmt.Sprintf("%s-%d", stamp, i)
		}
		dst := path + ".hatch-backup-" + suffix
		out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return "", err
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return dst, nil
	}
}

// BackupIfExists runs Backup only when path exists.
func BackupIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	_, err := Backup(path)
	return err
}

// WriteWithBackup backs up an existing file, ensures the parent dir exists and
// writes data. It is the standard way Hatch rewrites third-party config files.
func WriteWithBackup(path string, data []byte, perm os.FileMode) error {
	if err := BackupIfExists(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

// DirExists reports whether path is an existing directory.
func DirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// CopyDir copies src directory to dst recursively (dst is created if needed).
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

// CopyFileWithMime copies src to a new file dst (failing if dst exists) and
// returns the sniffed MIME type of src.
func CopyFileWithMime(src, dst string) (string, error) {
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

// ExpandHome expands a leading "~" or "~/" to the user's home directory.
func ExpandHome(path string) string {
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

// IsPathInside reports whether path is strictly inside root (after ~ expansion).
func IsPathInside(path, root string) bool {
	path = ExpandHome(path)
	root = ExpandHome(root)
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}
