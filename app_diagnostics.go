package main

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"hatch/internal/config"
)

// ─── Script health check ─────────────────────────────────────────────────────

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

// CheckScriptHealth verifies @hatch.requires deps and @hatch.health URL for one script.
// Backed by 2s HTTP timeout to avoid blocking the UI.
func (a *App) CheckScriptHealth(path string) HealthCheckResult {
	r := HealthCheckResult{Script: path}

	// Find meta — try override first, then store
	var meta config.ScriptMeta
	a.mu.RLock()
	if ov, ok := a.cfg.OverrideMeta[path]; ok {
		meta = ov
	}
	a.mu.RUnlock()
	if meta.Path == "" && a.store != nil {
		for _, m := range a.store.All() {
			if m.Path == path {
				meta = m
				break
			}
		}
	}

	r.HasAnnotations = len(meta.Requires) > 0 || meta.HealthURL != ""

	for _, dep := range meta.Requires {
		if findCLI(dep) == "" {
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

// CheckCommandAvailable checks if a command exists on PATH, returning a quick yes/no.
func (a *App) CheckCommandAvailable(name string) bool {
	return findCLI(name) != ""
}

// ─── PATH health ─────────────────────────────────────────────────────────────

// PathEntry describes one directory on $PATH.
type PathEntry struct {
	Path        string `json:"path"`
	Exists      bool   `json:"exists"`
	Duplicate   bool   `json:"duplicate"`   // appears more than once
	ShadowsName string `json:"shadowsName"` // if a binary in this dir shadows one earlier on PATH
}

// PathHealth is the result of CheckPathHealth.
type PathHealth struct {
	Entries        []PathEntry `json:"entries"`
	MissingCount   int         `json:"missingCount"`
	DuplicateCount int         `json:"duplicateCount"`
}

// CheckPathHealth inspects $PATH from a login zsh and reports missing/duplicate dirs.
func (a *App) CheckPathHealth() PathHealth {
	out, err := exec.Command("/bin/zsh", "-l", "-c", "echo $PATH").Output()
	if err != nil {
		return PathHealth{}
	}
	rawPath := strings.TrimSpace(string(out))
	parts := strings.Split(rawPath, ":")

	seen := map[string]int{}
	binToFirstDir := map[string]string{}

	h := PathHealth{}
	for _, p := range parts {
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
