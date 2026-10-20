// Package diagnostics implements the workbench health checks: per-script
// dependency/HTTP checks and $PATH hygiene.
package diagnostics

import (
	"net/http"
	"os"
	"strings"
	"time"

	"hatch/core/config"
	"hatch/core/macos"
)

// HealthCheckResult reports dependency and HTTP health for a script.
type HealthCheckResult struct {
	Script         string   `json:"script"`
	MissingDeps    []string `json:"missingDeps"`
	HealthURL      string   `json:"healthUrl"`
	HealthOK       bool     `json:"healthOk"`
	HealthStatus   int      `json:"healthStatus"` // HTTP status code; 0 if no URL or unreachable
	HealthLatency  int64    `json:"healthLatencyMs"`
	HealthError    string   `json:"healthError"`
	HasAnnotations bool     `json:"hasAnnotations"`
}

// PathEntry describes one directory on $PATH.
type PathEntry struct {
	Path        string `json:"path"`
	Exists      bool   `json:"exists"`
	Duplicate   bool   `json:"duplicate"`   // appears more than once
	ShadowsName string `json:"shadowsName"` // if a binary in this dir shadows one earlier on PATH
}

// PathHealth is the result of CheckPath.
type PathHealth struct {
	Entries        []PathEntry `json:"entries"`
	MissingCount   int         `json:"missingCount"`
	DuplicateCount int         `json:"duplicateCount"`
}

// CheckScript verifies meta's @hatch.requires deps and @hatch.health URL (2s timeout).
func CheckScript(path string, meta config.ScriptMeta) HealthCheckResult {
	r := HealthCheckResult{Script: path}
	r.HasAnnotations = len(meta.Requires) > 0 || meta.HealthURL != ""

	for _, dep := range meta.Requires {
		if macos.FindCLI(dep) == "" {
			r.MissingDeps = append(r.MissingDeps, dep)
		}
	}

	r.HealthURL = meta.HealthURL
	if meta.HealthURL != "" {
		client := &http.Client{Timeout: 2 * time.Second}
		start := time.Now()
		resp, err := client.Get(meta.HealthURL)
		r.HealthLatency = time.Since(start).Milliseconds()
		if err != nil {
			r.HealthError = err.Error()
		} else {
			defer resp.Body.Close()
			r.HealthStatus = resp.StatusCode
			r.HealthOK = resp.StatusCode >= 200 && resp.StatusCode < 400
		}
	}
	return r
}

// CheckPath inspects $PATH from a login zsh and reports missing, duplicate and
// shadowing directories.
func CheckPath() PathHealth {
	rawPath, err := macos.LoginShellPATH()
	if err != nil {
		return PathHealth{}
	}
	seen := map[string]int{}
	binToFirstDir := map[string]string{}

	h := PathHealth{}
	for _, p := range strings.Split(rawPath, ":") {
		if p == "" {
			continue
		}
		entry := PathEntry{Path: p}
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			entry.Exists = true
		} else {
			h.MissingCount++
		}
		seen[p]++
		if seen[p] > 1 {
			entry.Duplicate = true
			h.DuplicateCount++
		}
		// Detect shadowing: list binaries in this dir
		if entry.Exists {
			if files, err := os.ReadDir(p); err == nil {
				for _, f := range files {
					if f.IsDir() {
						continue
					}
					name := f.Name()
					if prevDir, ok := binToFirstDir[name]; ok && prevDir != p && entry.ShadowsName == "" {
						entry.ShadowsName = name + " (also in " + prevDir + ")"
					} else if !ok {
						binToFirstDir[name] = p
					}
				}
			}
		}
		h.Entries = append(h.Entries, entry)
	}
	return h
}
