package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RegistryBase is the official MCP Registry (API frozen at v0.1). Its search
// is a noisy name-substring match, so Hatch uses it mainly to resolve exact
// names into structured install metadata.
var RegistryBase = "https://registry.modelcontextprotocol.io/v0.1"

var registryClient = &http.Client{Timeout: 15 * time.Second}

type registryEnv struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsRequired  bool   `json:"isRequired"`
	IsSecret    bool   `json:"isSecret"`
	Default     string `json:"default"`
	Value       string `json:"value"`
}

type registryArg struct {
	Type       string `json:"type"` // positional | named
	Name       string `json:"name"`
	Value      string `json:"value"`
	Default    string `json:"default"`
	ValueHint  string `json:"valueHint"`
	IsRequired bool   `json:"isRequired"`
}

type registryServer struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	WebsiteURL  string `json:"websiteUrl"`
	Repository  struct {
		URL string `json:"url"`
	} `json:"repository"`
	Packages []struct {
		RegistryType     string                `json:"registryType"`
		Identifier       string                `json:"identifier"`
		Version          string                `json:"version"`
		RuntimeHint      string                `json:"runtimeHint"`
		Transport        struct{ Type string } `json:"transport"`
		RuntimeArguments []registryArg         `json:"runtimeArguments"`
		PackageArguments []registryArg         `json:"packageArguments"`
		Env              []registryEnv         `json:"environmentVariables"`
	} `json:"packages"`
	Remotes []struct {
		Type    string        `json:"type"`
		URL     string        `json:"url"`
		Headers []registryEnv `json:"headers"`
	} `json:"remotes"`
}

// RegistryHit is one search result.
type RegistryHit struct {
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Kinds       string `json:"kinds,omitempty"` // "npm, remote"
}

func registryGet(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, RegistryBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := registryClient.Do(req)
	if err != nil {
		return fmt.Errorf("无法访问 MCP Registry: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Registry 中没有这个服务")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MCP Registry 返回 %d", resp.StatusCode)
	}
	return json.Unmarshal(body, out)
}

// SearchRegistry lists latest versions whose name contains query.
func SearchRegistry(ctx context.Context, query string) ([]RegistryHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("请输入关键词")
	}
	var res struct {
		Servers []struct {
			Server registryServer `json:"server"`
		} `json:"servers"`
	}
	if err := registryGet(ctx, "/servers?version=latest&limit=50&search="+url.QueryEscape(query), &res); err != nil {
		return nil, err
	}
	out := []RegistryHit{}
	for _, s := range res.Servers {
		var kinds []string
		for _, p := range s.Server.Packages {
			kinds = append(kinds, p.RegistryType)
		}
		if len(s.Server.Remotes) > 0 {
			kinds = append(kinds, "remote")
		}
		out = append(out, RegistryHit{Name: s.Server.Name, Title: s.Server.Title, Description: s.Server.Description, Version: s.Server.Version, Kinds: strings.Join(kinds, ", ")})
	}
	return out, nil
}

// LookupRegistry resolves an exact registry name into one draft per install
// option (npm / pypi / oci package, remote endpoint).
func LookupRegistry(ctx context.Context, name string) ([]Draft, error) {
	name = strings.TrimSpace(name)
	var res struct {
		Server registryServer `json:"server"`
	}
	if err := registryGet(ctx, "/servers/"+url.PathEscape(name)+"/versions/latest", &res); err != nil {
		return nil, err
	}
	drafts := DraftsFromRegistry(res.Server)
	if len(drafts) == 0 {
		return nil, fmt.Errorf("%s 没有 Hatch 可用的安装方式（仅支持 npm / pypi / oci 包和远程端点）", name)
	}
	return drafts, nil
}

// DraftsFromRegistry maps server.json to drafts.
func DraftsFromRegistry(s registryServer) []Draft {
	base := Draft{
		Name:        SanitizeName(s.Name),
		Title:       s.Title,
		Description: s.Description,
		Origin:      "registry:" + s.Name,
		Homepage:    strings.TrimSpace(s.WebsiteURL),
	}
	if base.Homepage == "" {
		base.Homepage = s.Repository.URL
	}
	var out []Draft
	for _, r := range s.Remotes {
		d := base
		transport := TransportHTTP
		if r.Type == "sse" {
			transport = TransportSSE
		}
		d.Spec = Spec{Transport: transport, URL: r.URL, Headers: map[string]string{}}
		for _, h := range r.Headers {
			d.Spec.Headers[h.Name] = firstNonEmpty(h.Value, h.Default)
			d.Inputs = append(d.Inputs, InputDef{Kind: "header", Key: h.Name, Description: h.Description, Required: h.IsRequired, Secret: h.IsSecret, Default: h.Default})
		}
		d.Variant = "远程 · " + r.Type
		if strings.Contains(r.URL, "{") {
			d.Notes = append(d.Notes, "URL 中含 {变量}，请替换为实际值")
		}
		d.Spec = d.Spec.Normalize()
		out = append(out, d)
	}
	for _, p := range s.Packages {
		if p.Transport.Type != "" && p.Transport.Type != "stdio" {
			continue
		}
		d := base
		var cmd string
		var args []string
		switch p.RegistryType {
		case "npm":
			cmd, args = "npx", []string{"-y", p.Identifier}
		case "pypi":
			cmd, args = "uvx", []string{p.Identifier}
		case "oci":
			image := p.Identifier
			if p.Version != "" && !strings.Contains(image[strings.LastIndex(image, "/")+1:], ":") {
				image += ":" + p.Version
			}
			cmd, args = "docker", []string{"run", "-i", "--rm"}
			for _, e := range p.Env {
				args = append(args, "-e", e.Name)
			}
			args = append(args, image)
		default:
			continue
		}
		if p.RuntimeHint != "" && p.RegistryType != "oci" {
			cmd = p.RuntimeHint
		}
		for _, a := range p.PackageArguments {
			v := firstNonEmpty(a.Value, a.Default)
			if v == "" && !a.IsRequired {
				continue
			}
			if v == "" {
				v = "<" + firstNonEmpty(a.ValueHint, a.Name, "value") + ">"
				d.Notes = append(d.Notes, "请替换参数 "+v)
			}
			if a.Type == "named" {
				args = append(args, a.Name, v)
			} else {
				args = append(args, v)
			}
		}
		d.Spec = Spec{Transport: TransportStdio, Command: cmd, Args: args, Env: map[string]string{}}
		for _, e := range p.Env {
			d.Spec.Env[e.Name] = firstNonEmpty(e.Value, e.Default)
			d.Inputs = append(d.Inputs, InputDef{Kind: "env", Key: e.Name, Description: e.Description, Required: e.IsRequired, Secret: e.IsSecret, Default: e.Default})
		}
		d.Variant = p.RegistryType + " · stdio"
		d.Spec = d.Spec.Normalize()
		out = append(out, d)
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
