package scanner

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
		}
	}

	return meta, found
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

// probeHelp runs `script --help` with a 2-second timeout and returns stdout+stderr.
func probeHelp(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--help")
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
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
