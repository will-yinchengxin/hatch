package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"hatch/internal/config"
)

// ignoredExts are file extensions we never treat as scripts.
var ignoredExts = map[string]bool{
	".md": true, ".txt": true, ".log": true, ".json": true,
	".yaml": true, ".yml": true, ".toml": true, ".ini": true,
	".png": true, ".jpg": true, ".gif": true, ".svg": true,
	".woff2": true, ".woff": true, ".css": true, ".html": true,
}

// ScanDir walks dir and returns a ScriptMeta for every executable file found.
// overrides is applied last so users can correct auto-detected metadata.
func ScanDir(dir string, overrides map[string]config.ScriptMeta) ([]config.ScriptMeta, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var scripts []config.ScriptMeta
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ignoredExts[ext] {
			continue
		}

		fullPath := filepath.Join(dir, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if !isExecutable(info) && !isScript(name, fullPath) {
			continue
		}

		meta := buildMeta(fullPath, name, info)

		if ov, ok := overrides[fullPath]; ok {
			meta = mergeOverride(meta, ov)
		}

		scripts = append(scripts, meta)
	}

	return scripts, nil
}

// buildMeta constructs ScriptMeta using the three-tier fallback:
//  1. @hatch.* annotations in the file header
//  2. --help probe output (sets Desc if still empty)
//  3. sensible defaults
func buildMeta(path, name string, info os.FileInfo) config.ScriptMeta {
	meta, hasAnnotations := parseHatchAnnotations(path)
	meta.Path = path
	meta.UpdatedAt = info.ModTime().Unix()
	meta.Language = detectLanguage(path)

	if meta.Name == "" {
		meta.Name = strings.TrimSuffix(name, filepath.Ext(name))
	}
	if meta.Type == "" {
		meta.Type = config.TypeOneshot
	}

	// Tier 2a: auto-detect service pattern when no annotations declared a type.
	if !hasAnnotations && meta.Type == config.TypeOneshot {
		if isService, cmds := detectServicePattern(path); isService {
			meta.Type = config.TypeService
			if len(meta.Commands) == 0 {
				meta.Commands = cmds
			}
		}
	}

	if !hasAnnotations && meta.Desc == "" {
		if help := probeHelp(path); help != "" {
			for _, line := range strings.Split(help, "\n") {
				if line = strings.TrimSpace(line); line != "" {
					if len(line) > 120 {
						line = line[:120] + "…"
					}
					meta.Desc = line
					break
				}
			}
		}
	}

	if meta.Type == config.TypeService && meta.PidFile == "" {
		meta.PidFile = "/tmp/hatch/" + name + ".pid"
	}

	return meta
}

// isExecutable returns true when the file has any execute bit set.
func isExecutable(info os.FileInfo) bool {
	return info.Mode()&0o111 != 0
}

// isScript returns true for known script extensions even if not +x.
func isScript(name, path string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".sh", ".bash", ".zsh", ".py", ".rb", ".js":
		return true
	}
	// Check shebang
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 4)
	n, _ := f.Read(buf)
	return n >= 2 && buf[0] == '#' && buf[1] == '!'
}

// mergeOverride applies non-zero fields from ov onto base.
func mergeOverride(base, ov config.ScriptMeta) config.ScriptMeta {
	if ov.Name != "" {
		base.Name = ov.Name
	}
	if ov.Desc != "" {
		base.Desc = ov.Desc
	}
	if ov.Type != "" {
		base.Type = ov.Type
	}
	if len(ov.Commands) > 0 {
		base.Commands = ov.Commands
	}
	if ov.PidFile != "" {
		base.PidFile = ov.PidFile
	}
	if ov.Args != nil {
		base.Args = ov.Args
	}
	if len(ov.Presets) > 0 {
		base.Presets = ov.Presets
	}
	if len(ov.Tags) > 0 {
		base.Tags = ov.Tags
	}
	// Preserve runtime fields from base
	base.LastRunAt = ov.LastRunAt
	return base
}
