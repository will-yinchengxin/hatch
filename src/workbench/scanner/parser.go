package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"hatch/core/config"
)

// parseHatchAnnotations tries to read @hatch.* comment block from the file.
// Returns a partially-filled ScriptMeta and true if any annotation was found.
func parseHatchAnnotations(path string) (config.ScriptMeta, bool) {
	f, err := os.Open(path)
	if err != nil {
		return config.ScriptMeta{}, false
	}
	defer f.Close()

	var meta config.ScriptMeta
	found := false

	scanner := bufio.NewScanner(f)
	// Only scan first 50 lines to avoid reading large binaries
	lineCount := 0
	for scanner.Scan() && lineCount < 50 {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Strip comment prefixes: #, //, --
		for _, prefix := range []string{"#", "//", "--"} {
			if strings.HasPrefix(line, prefix) {
				line = strings.TrimSpace(line[len(prefix):])
				break
			}
		}

		if !strings.HasPrefix(line, "@hatch.") {
			continue
		}
		found = true

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip inline comments after value
		if idx := strings.Index(val, "#"); idx > 0 {
			val = strings.TrimSpace(val[:idx])
		}

		switch key {
		case "@hatch.name":
			meta.Name = val
		case "@hatch.desc":
			meta.Desc = val
		case "@hatch.type":
			switch val {
			case "service":
				meta.Type = config.TypeService
			default:
				meta.Type = config.TypeOneshot
			}
		case "@hatch.commands":
			for _, cmd := range strings.Split(val, ",") {
				if c := strings.TrimSpace(cmd); c != "" {
					meta.Commands = append(meta.Commands, c)
				}
			}
		case "@hatch.pidfile":
			meta.PidFile = val
		case "@hatch.tags":
			for _, tag := range strings.Split(val, ",") {
				if t := strings.TrimSpace(tag); t != "" {
					meta.Tags = append(meta.Tags, t)
				}
			}
		case "@hatch.schedule":
			meta.Schedule = val
		case "@hatch.requires":
			for _, r := range strings.Split(val, ",") {
				if c := strings.TrimSpace(r); c != "" {
					meta.Requires = append(meta.Requires, c)
				}
			}
		case "@hatch.health":
			meta.HealthURL = val
		case "@hatch.arg":
			if arg, ok := parseArgAnnotation(val); ok {
				meta.Args = append(meta.Args, arg)
			}
		case "@hatch.preset":
			name, args, _ := strings.Cut(val, "|")
			name, args = strings.TrimSpace(name), strings.TrimSpace(args)
			if name != "" || args != "" {
				if name == "" {
					name = args
				}
				meta.Presets = append(meta.Presets, config.CommandPreset{Name: name, Args: args})
			}
		}
	}

	return meta, found
}

// parseArgAnnotation parses "@hatch.arg: name | --flag | default | required".
// Only name is mandatory; flag empty means a positional argument. Default is
// shown as a hint only: the UI omits empty args so the script applies it.
func parseArgAnnotation(val string) (config.ArgDef, bool) {
	parts := strings.Split(val, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if parts[0] == "" {
		return config.ArgDef{}, false
	}
	arg := config.ArgDef{Name: parts[0]}
	if len(parts) > 1 {
		arg.Flag = parts[1]
	}
	if len(parts) > 2 {
		arg.Default = parts[2]
	}
	if len(parts) > 3 {
		arg.Required = strings.EqualFold(parts[3], "required")
	}
	return arg, true
}

// detectServicePattern scans the first 200 lines for start/stop/restart/reload
// patterns common in init-style and daemon control scripts (shell & python).
// Returns (true, commands) when at least start+stop are found.
func detectServicePattern(path string) (bool, []string) {
	f, err := os.Open(path)
	if err != nil {
		return false, nil
	}
	defer f.Close()

	var sb strings.Builder
	sc := bufio.NewScanner(f)
	lineCount := 0
	for sc.Scan() && lineCount < 200 {
		lineCount++
		sb.WriteString(strings.ToLower(sc.Text()))
		sb.WriteByte('\n')
	}
	text := sb.String()

	known := []string{"start", "stop", "restart", "reload", "status"}
	var found []string
	for _, cmd := range known {
		if matchesServiceCmd(text, cmd) {
			found = append(found, cmd)
		}
	}

	if containsStr(found, "start") && containsStr(found, "stop") {
		return true, found
	}
	return false, nil
}

// matchesServiceCmd returns true when cmd appears as a standalone service verb
// in text (case statement arm, quoted comparison, usage hint, etc.).
func matchesServiceCmd(text, cmd string) bool {
	for _, pattern := range []string{
		cmd + ")",       // case arm:   start)
		`"` + cmd + `"`, // quoted:     "start"
		`'` + cmd + `'`, // quoted:     'start'
		"|" + cmd + "|", // usage hint: {start|stop}
		"|" + cmd + "}",
		"{" + cmd + "|",
	} {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// headerComment returns the first descriptive comment near the top of a
// script (after the shebang, skipping @hatch lines, separators and encoding
// pragmas), or a one-line Python docstring. Stops at the first code line.
func headerComment(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for n := 0; sc.Scan() && n < 20; n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#!") {
			continue
		}
		text, isComment := "", false
		for _, prefix := range []string{"#", "//", "--"} {
			if strings.HasPrefix(line, prefix) {
				text, isComment = strings.TrimSpace(strings.TrimLeft(line, prefix)), true
				break
			}
		}
		if !isComment {
			for _, q := range []string{`"""`, "'''"} {
				if strings.HasPrefix(line, q) {
					text, isComment = strings.TrimSpace(strings.Trim(line, q)), true
				}
			}
		}
		if !isComment {
			return ""
		}
		lower := strings.ToLower(text)
		if text == "" || strings.HasPrefix(text, "@") || strings.Trim(text, "-=*#~ ") == "" ||
			strings.Contains(lower, "coding") || strings.HasPrefix(lower, "shellcheck") || strings.HasPrefix(lower, "vim:") {
			continue
		}
		if r := []rune(text); len(r) > 120 {
			text = string(r[:120]) + "…"
		}
		return text
	}
	return ""
}

// detectLanguage infers the scripting language from shebang or file extension.
func detectLanguage(path string) string {
	ext := strings.ToLower(path)
	switch {
	case strings.HasSuffix(ext, ".py"):
		return "python"
	case strings.HasSuffix(ext, ".sh") || strings.HasSuffix(ext, ".bash"):
		return "shell"
	case strings.HasSuffix(ext, ".rb"):
		return "ruby"
	case strings.HasSuffix(ext, ".js"):
		return "node"
	}

	// Read shebang
	f, err := os.Open(path)
	if err != nil {
		return "binary"
	}
	defer f.Close()

	buf := make([]byte, 128)
	n, _ := f.Read(buf)
	first := string(buf[:n])
	if !strings.HasPrefix(first, "#!") {
		return "binary"
	}
	// e.g. "#!/usr/bin/env python3" → fields: ["/usr/bin/env", "python3"]
	fields := strings.Fields(first[2:])
	if len(fields) == 0 {
		return "binary"
	}
	// When the interpreter is `env`, the real interpreter is the next word.
	interp := strings.ToLower(filepath.Base(fields[0]))
	if interp == "env" && len(fields) > 1 {
		interp = strings.ToLower(filepath.Base(fields[1]))
	}
	switch {
	case strings.HasPrefix(interp, "python"):
		return "python"
	case interp == "bash" || interp == "sh" || interp == "zsh" || interp == "dash":
		return "shell"
	case strings.HasPrefix(interp, "ruby"):
		return "ruby"
	case interp == "node" || strings.HasPrefix(interp, "node"):
		return "node"
	default:
		return "shell"
	}
}
