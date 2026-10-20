package mcp

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

// GitClone clones url into dst (2 minute timeout).
func GitClone(url, dst string) error {
	out, err := runGit("git", "clone", url, dst)
	if err != nil {
		return fmt.Errorf("git clone failed: %s", out)
	}
	return nil
}

// GitPull runs `git pull` inside dir (2 minute timeout).
func GitPull(dir string) error {
	out, err := runGit("git", "-C", dir, "pull")
	if err != nil {
		return fmt.Errorf("git pull 失败: %s", out)
	}
	return nil
}

func runGit(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := ExecCommand(ctx, name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return strings.TrimSpace(buf.String()), err
}
