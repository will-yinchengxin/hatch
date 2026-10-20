package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hatch/core/config"
	"hatch/core/strutil"
)

// InferFromDir guesses a stdio MCPServer definition from a local checkout:
// package.json bin/start/main first, then the first executable file.
func InferFromDir(dir string) config.MCPServer {
	name := filepath.Base(dir)
	id := strutil.SanitizeID(strings.TrimSuffix(strings.TrimSuffix(name, "-mcp-server"), "_mcp_server"))
	server := config.MCPServer{
		ID:         id,
		Name:       humanizeName(name),
		SourceType: "local",
		LocalPath:  dir,
		Transport:  TransportStdio,
		Cwd:        dir,
		UpdatedAt:  time.Now().Unix(),
	}
	if desc := readReadmeSummary(dir); desc != "" {
		server.Description = desc
	}
	if pkg, ok := readPackageJSON(dir); ok {
		if pkg.Name != "" {
			server.PackageName = pkg.Name
		}
		if server.ID == "" {
			server.ID = strutil.SanitizeID(pkg.Name)
		}
		if server.Name == "" {
			server.Name = pkg.Name
		}
		if binPath := firstPackageBin(pkg); binPath != "" {
			server.Command = "node"
			server.Args = []string{filepath.Join(dir, binPath)}
			return server
		}
		if start := strings.TrimSpace(pkg.Scripts["start"]); strings.HasPrefix(start, "node ") {
			server.Command = "node"
			server.Args = strings.Fields(strings.TrimPrefix(start, "node "))
			for i, arg := range server.Args {
				if !filepath.IsAbs(arg) && strings.HasSuffix(arg, ".js") {
					server.Args[i] = filepath.Join(dir, arg)
				}
			}
			return server
		}
		if pkg.Main != "" {
			server.Command = "node"
			server.Args = []string{filepath.Join(dir, pkg.Main)}
			return server
		}
		server.Command = "npm"
		server.Args = []string{"run", "start"}
		return server
	}
	if exe := firstExecutableFile(dir); exe != "" {
		server.Command = exe
		if strings.Contains(filepath.Base(exe), "prometheus") {
			server.Args = []string{"--prometheus.url", "http://127.0.0.1:9090"}
		}
		return server
	}
	return server
}

type packageJSON struct {
	Name    string            `json:"name"`
	Main    string            `json:"main"`
	Bin     json.RawMessage   `json:"bin"`
	Scripts map[string]string `json:"scripts"`
}

func readPackageJSON(dir string) (packageJSON, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return packageJSON{}, false
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return packageJSON{}, false
	}
	if pkg.Scripts == nil {
		pkg.Scripts = map[string]string{}
	}
	return pkg, true
}

func firstPackageBin(pkg packageJSON) string {
	if len(pkg.Bin) == 0 || string(pkg.Bin) == "null" {
		return ""
	}
	var one string
	if json.Unmarshal(pkg.Bin, &one) == nil {
		return one
	}
	var many map[string]string
	if json.Unmarshal(pkg.Bin, &many) != nil {
		return ""
	}
	keys := make([]string, 0, len(many))
	for k := range many {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return many[keys[0]]
}

func readReadmeSummary(dir string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, "README*"))
	sort.Strings(matches)
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(strings.Trim(line, "#> "))
			if line == "" || strings.HasPrefix(line, "[!") || strings.HasPrefix(line, "<") {
				continue
			}
			if len(line) > 180 {
				line = line[:180] + "..."
			}
			return line
		}
	}
	return ""
}

func firstExecutableFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0o111 != 0 {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

// Normalize trims fields, validates transport-specific requirements and
// initializes nil maps so the server is safe to persist.
func Normalize(server config.MCPServer) (config.MCPServer, error) {
	server.ID = strutil.SanitizeID(server.ID)
	if !ValidID(server.ID) {
		return config.MCPServer{}, fmt.Errorf("invalid MCP id")
	}
	server.Name = strutil.OrDefault(strings.TrimSpace(server.Name), server.ID)
	server.SourceType = strutil.OrDefault(strings.TrimSpace(server.SourceType), "local")
	server.Transport = strings.TrimSpace(server.Transport)
	if server.Transport == "" {
		server.Transport = TransportStdio
	}
	if server.Transport != TransportStdio && server.Transport != TransportHTTP {
		return config.MCPServer{}, fmt.Errorf("unsupported MCP transport: %s", server.Transport)
	}
	server.Command = strings.TrimSpace(server.Command)
	server.Cwd = strings.TrimSpace(server.Cwd)
	server.URL = strings.TrimSpace(server.URL)
	server.SourceURL = strings.TrimSpace(server.SourceURL)
	server.LocalPath = strings.TrimSpace(server.LocalPath)
	server.PackageName = strings.TrimSpace(server.PackageName)
	server.BearerTokenEnvVar = strings.TrimSpace(server.BearerTokenEnvVar)
	if server.Transport == TransportStdio && server.Command == "" {
		return config.MCPServer{}, fmt.Errorf("stdio MCP command is required")
	}
	if server.Transport == TransportHTTP && server.URL == "" {
		return config.MCPServer{}, fmt.Errorf("HTTP MCP URL is required")
	}
	server.Args = strutil.CleanSlice(server.Args)
	server.Tags = strutil.CleanSlice(server.Tags)
	if server.Env == nil {
		server.Env = map[string]string{}
	}
	if server.HTTPHeaders == nil {
		server.HTTPHeaders = map[string]string{}
	}
	if server.LastTestResults == nil {
		server.LastTestResults = map[string]config.MCPTestResult{}
	}
	server.UpdatedAt = time.Now().Unix()
	return server, nil
}

func humanizeName(name string) string {
	name = strings.TrimSuffix(strings.TrimSuffix(name, "-mcp-server"), "_mcp_server")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	parts := strings.Fields(name)
	for i, p := range parts {
		if p == "mcp" {
			parts[i] = "MCP"
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// Clone deep-copies a server list (slices, maps and test results).
func Clone(in []config.MCPServer) []config.MCPServer {
	out := make([]config.MCPServer, len(in))
	for i, s := range in {
		out[i] = s
		out[i].Args = append([]string(nil), s.Args...)
		out[i].Tags = append([]string(nil), s.Tags...)
		out[i].Env = strutil.CloneMap(s.Env)
		out[i].HTTPHeaders = strutil.CloneMap(s.HTTPHeaders)
		if s.LastTestResults != nil {
			out[i].LastTestResults = make(map[string]config.MCPTestResult, len(s.LastTestResults))
			for k, v := range s.LastTestResults {
				out[i].LastTestResults[k] = v
			}
		}
	}
	return out
}
