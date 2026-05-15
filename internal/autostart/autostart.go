package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const label = "com.hatch.app"

// Enabled reports whether Hatch's user LaunchAgent plist exists.
func Enabled() bool {
	_, err := os.Stat(plistPath())
	return err == nil
}

// SetEnabled installs or removes a user LaunchAgent. The plist is picked up on
// the next login; launchctl is intentionally avoided so this works without
// elevated permissions.
func SetEnabled(enabled bool) error {
	p := plistPath()
	if !enabled {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	target, err := launchTarget()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(plist(target)), 0o644)
}

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist")
}

func launchTarget() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, _ = filepath.EvalSymlinks(exe)
	if idx := strings.Index(exe, ".app/Contents/MacOS/"); idx > 0 {
		return exe[:idx+len(".app")], nil
	}
	return exe, nil
}

func plist(target string) string {
	var programArgs string
	if strings.HasSuffix(target, ".app") {
		programArgs = fmt.Sprintf(`<string>/usr/bin/open</string>
        <string>%s</string>`, xmlEscape(target))
	} else {
		programArgs = fmt.Sprintf(`<string>%s</string>`, xmlEscape(target))
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        %s
    </array>
    <key>RunAtLoad</key>
    <true/>
  </dict>
</plist>
`, label, programArgs)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
