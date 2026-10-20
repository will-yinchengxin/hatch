// Package workspace holds the data model and snapshot storage behind AI
// workspace profiles. Profiles are applied manually; before applying, Hatch
// snapshots every agent config file and synced skill dir so a failed or
// unwanted apply can be rolled back byte-for-byte.
package workspace

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hatch/core/config"
	"hatch/core/fsutil"
)

// MaxSnapshots is how many snapshots are kept on disk.
const MaxSnapshots = 20

// Preview is the dry-run result of applying a profile.
type Preview struct {
	Profile  config.AIWorkspaceProfile `json:"profile"`
	Changes  []string                  `json:"changes"`
	Errors   []string                  `json:"errors"`
	Warnings []string                  `json:"warnings"`
	CanApply bool                      `json:"canApply"`
}

// SnapshotInfo is the list-view metadata of one snapshot.
type SnapshotInfo struct {
	ID          string `json:"id"`
	ProfileID   string `json:"profileId"`
	ProfileName string `json:"profileName"`
	CreatedAt   int64  `json:"createdAt"`
}

// Health is the AI environment health report.
type Health struct {
	Overall        string        `json:"overall"`
	CurrentProfile string        `json:"currentProfile"`
	ProfileState   string        `json:"profileState"`
	Checks         []HealthCheck `json:"checks"`
}

// HealthCheck is one row in the health report.
type HealthCheck struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Status     string `json:"status"` // ok | warn | error
	Detail     string `json:"detail"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Add appends c and escalates Overall (ok → warn → error).
func (h *Health) Add(c HealthCheck) {
	if c.Status == "error" {
		h.Overall = "error"
	} else if c.Status == "warn" && h.Overall == "ok" {
		h.Overall = "warn"
	}
	h.Checks = append(h.Checks, c)
}

// AddJSONCheck validates a JSON config file (missing counts as ok).
func (h *Health) AddJSONCheck(path, id, label string) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		h.Checks = append(h.Checks, HealthCheck{ID: id, Label: label, Status: "ok", Detail: "尚未创建"})
		return
	}
	if err != nil || !json.Valid(data) {
		h.Overall = "error"
		h.Checks = append(h.Checks, HealthCheck{ID: id, Label: label, Status: "error", Detail: "JSON 无效", Suggestion: "修复原始配置后再同步"})
		return
	}
	h.Checks = append(h.Checks, HealthCheck{ID: id, Label: label, Status: "ok", Detail: "JSON 有效"})
}

// FileSnapshot captures one file's bytes and mode (or its absence).
type FileSnapshot struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Mode   uint32 `json:"mode,omitempty"`
	Data   string `json:"data,omitempty"`
}

// DirSnapshot captures every file under a directory (or its absence).
type DirSnapshot struct {
	Path    string         `json:"path"`
	Exists  bool           `json:"exists"`
	Entries []FileSnapshot `json:"entries,omitempty"`
}

// Snapshot is the full pre-apply state persisted as <id>.json.
type Snapshot struct {
	SnapshotInfo
	Files       []FileSnapshot                  `json:"files"`
	SkillDirs   []DirSnapshot                   `json:"skillDirs"`
	MCPStates   []config.AgentResourceSelection `json:"mcpStates"`
	HookStates  []config.AgentResourceSelection `json:"hookStates"`
	LastProfile string                          `json:"lastProfile"`
}

// ValidID accepts "ws-…" profile ids and "snapshot-…" snapshot ids.
func ValidID(id string) bool {
	if !(strings.HasPrefix(id, "ws-") || strings.HasPrefix(id, "snapshot-")) || len(id) > 80 {
		return false
	}
	for _, r := range id {
		if !(r == '-' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

// NormalizeSelections trims ids, drops blanks/duplicates and no-op selections.
func NormalizeSelections(in []config.AgentResourceSelection) []config.AgentResourceSelection {
	seen := map[string]bool{}
	var out []config.AgentResourceSelection
	for _, s := range in {
		s.ID = strings.TrimSpace(s.ID)
		if s.ID != "" && !seen[s.ID] && (s.Claude || s.Codex) {
			seen[s.ID] = true
			out = append(out, s)
		}
	}
	return out
}

// SelectionFor returns the selection for id (all-false when absent).
func SelectionFor(in []config.AgentResourceSelection, id string) config.AgentResourceSelection {
	for _, s := range in {
		if s.ID == id {
			return s
		}
	}
	return config.AgentResourceSelection{ID: id}
}

// FirstExecutableToken returns the leading command of a simple shell command,
// or "" when the command uses pipes/substitution and cannot be checked.
func FirstExecutableToken(command string) string {
	command = strings.TrimSpace(command)
	if command == "" || strings.ContainsAny(command, "|;&$`()") {
		return ""
	}
	return strings.Fields(command)[0]
}

// SnapshotFile captures path.
func SnapshotFile(path string) (FileSnapshot, error) {
	f := FileSnapshot{Path: path}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return f, nil
	}
	if err != nil {
		return f, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return f, err
	}
	f.Exists = true
	f.Mode = uint32(info.Mode())
	f.Data = base64.StdEncoding.EncodeToString(data)
	return f, nil
}

// RestoreFile writes back (or removes) a captured file.
func RestoreFile(f FileSnapshot) error {
	if !f.Exists {
		if err := os.Remove(f.Path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(f.Data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(f.Path, data, os.FileMode(f.Mode))
}

// SnapshotDir captures every file under path.
func SnapshotDir(path string) (DirSnapshot, error) {
	d := DirSnapshot{Path: path}
	if !fsutil.DirExists(path) {
		return d, nil
	}
	d.Exists = true
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := SnapshotFile(p)
		if err != nil {
			return err
		}
		d.Entries = append(d.Entries, f)
		return nil
	})
	return d, err
}

// RestoreDir replaces path with the captured contents (or removes it).
func RestoreDir(d DirSnapshot) error {
	if err := os.RemoveAll(d.Path); err != nil {
		return err
	}
	if !d.Exists {
		return nil
	}
	if err := os.MkdirAll(d.Path, 0o755); err != nil {
		return err
	}
	for _, f := range d.Entries {
		if err := RestoreFile(f); err != nil {
			return err
		}
	}
	return nil
}

// SnapshotDirPath is ~/Library/Application Support/Hatch/ai-workspace-snapshots.
func SnapshotDirPath() string {
	return fsutil.AppDataPath("ai-workspace-snapshots")
}

// WriteSnapshot persists s as <dir>/<id>.json.
func WriteSnapshot(s Snapshot) error {
	dir := SnapshotDirPath()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, s.ID+".json"), data, 0o600)
}

// ReadSnapshot loads one snapshot by id.
func ReadSnapshot(id string) (Snapshot, error) {
	var snap Snapshot
	data, err := os.ReadFile(filepath.Join(SnapshotDirPath(), id+".json"))
	if err != nil {
		return snap, err
	}
	err = json.Unmarshal(data, &snap)
	return snap, err
}

// ListSnapshots returns snapshot metadata, newest first.
func ListSnapshots() []SnapshotInfo {
	dir := SnapshotDirPath()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil
	}
	entries, _ := os.ReadDir(dir)
	var out []SnapshotInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var s Snapshot
		if json.Unmarshal(data, &s) == nil && s.ID != "" {
			out = append(out, s.SnapshotInfo)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out
}

// PruneSnapshots keeps the newest MaxSnapshots snapshots.
func PruneSnapshots() {
	list := ListSnapshots()
	for i := MaxSnapshots; i < len(list); i++ {
		_ = os.Remove(filepath.Join(SnapshotDirPath(), list[i].ID+".json"))
	}
}
