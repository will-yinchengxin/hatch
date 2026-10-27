package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"hatch/core/config"
)

// Hatch probes a server the way a client would: start it (or POST to it),
// run initialize → notifications/initialized → tools/list, then stop. This
// is a minimal JSON-RPC client for stdio and streamable HTTP; legacy SSE and
// WebSocket servers are reported as unsupported.

const probeProtocolVersion = "2025-06-18"

// ProbeTimeout bounds one probe; first runs of npx / uvx / docker may download.
var ProbeTimeout = 60 * time.Second

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type initResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type toolsResult struct {
	Tools      []config.MCPTool `json:"tools"`
	NextCursor string           `json:"nextCursor"`
}

// transport is one request/response channel to a server.
type transport interface {
	call(ctx context.Context, id int, method string, params any) (json.RawMessage, error)
	notify(ctx context.Context, method string) error
	close()
}

type authError struct{ msg string }

func (e authError) Error() string { return e.msg }

// Probe connects to e as an MCP client.
func Probe(e Entry) config.MCPProbe {
	start := time.Now()
	res := config.MCPProbe{CheckedAt: start.Unix()}
	ctx, cancel := context.WithTimeout(context.Background(), ProbeTimeout)
	defer cancel()

	env := loginEnv()
	spec := resolveSpec(e, env)
	var tr transport
	var stderr *tailBuffer
	switch spec.Transport {
	case TransportStdio:
		st, err := startStdio(ctx, spec, env, e.File)
		if err != nil {
			res.Status, res.Summary = "failed", "启动失败: "+err.Error()
			return res
		}
		tr, stderr = st, st.stderr
	case TransportHTTP:
		tr = &httpTransport{url: spec.URL, headers: spec.Headers}
	default:
		res.Status, res.Summary = "unsupported", "暂不支持探测 "+strings.ToUpper(spec.Transport)+" 传输，可在客户端中用 /mcp 查看"
		return res
	}
	defer tr.close()

	err := func() error {
		raw, err := tr.call(ctx, 1, "initialize", map[string]any{
			"protocolVersion": probeProtocolVersion,
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "hatch", "version": "1"},
		})
		if err != nil {
			return err
		}
		var init initResult
		_ = json.Unmarshal(raw, &init)
		res.ServerName, res.ServerVersion, res.ProtocolVersion = init.ServerInfo.Name, init.ServerInfo.Version, init.ProtocolVersion
		if h, ok := tr.(*httpTransport); ok && init.ProtocolVersion != "" {
			h.protocol = init.ProtocolVersion
		}
		if err := tr.notify(ctx, "notifications/initialized"); err != nil {
			return err
		}
		cursor := ""
		for page := 0; page < 10; page++ {
			params := map[string]any{}
			if cursor != "" {
				params["cursor"] = cursor
			}
			raw, err := tr.call(ctx, 2+page, "tools/list", params)
			if err != nil {
				return fmt.Errorf("tools/list: %w", err)
			}
			var tl toolsResult
			if err := json.Unmarshal(raw, &tl); err != nil {
				return fmt.Errorf("tools/list 返回无法解析: %w", err)
			}
			for _, t := range tl.Tools {
				t.Description = truncate(firstLine(t.Description), 160)
				res.Tools = append(res.Tools, t)
			}
			if cursor = tl.NextCursor; cursor == "" {
				break
			}
		}
		return nil
	}()
	res.LatencyMs = time.Since(start).Milliseconds()
	if stderr != nil {
		res.StderrTail = stderr.String()
	}
	var ae authError
	switch {
	case errors.As(err, &ae):
		res.Status, res.Summary = "auth_required", "需要授权："+ae.msg+"。在终端运行 claude mcp login / codex mcp login 完成 OAuth，或填写认证 Header"
	case err != nil && ctx.Err() != nil:
		res.Status, res.Summary = "timeout", fmt.Sprintf("%d 秒内未完成握手（首次运行 npx/uvx/docker 可能仍在下载，可稍后重试）", int(ProbeTimeout.Seconds()))
	case err != nil:
		res.Status, res.Summary = "failed", err.Error()
	default:
		res.Status = "ok"
		res.Summary = fmt.Sprintf("已连接 · %d 个工具 · %dms", len(res.Tools), res.LatencyMs)
	}
	return res
}

// ── environment resolution ─────────────────────────────────────────────────

var (
	loginEnvOnce sync.Mutex
	loginEnvAt   time.Time
	loginEnvMap  map[string]string
)

// loginEnv is the environment of a login zsh that sourced ~/.zshrc, which is
// what a terminal-launched client sees (the GUI app's own env is minimal).
func loginEnv() map[string]string {
	loginEnvOnce.Lock()
	defer loginEnvOnce.Unlock()
	if loginEnvMap != nil && time.Since(loginEnvAt) < time.Minute {
		return loginEnvMap
	}
	out := map[string]string{}
	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok {
			out[k] = v
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	raw, err := exec.CommandContext(ctx, "/bin/zsh", "-l", "-c", "source ~/.zshrc >/dev/null 2>&1; env -0").Output()
	if err == nil {
		for _, kv := range bytes.Split(raw, []byte{0}) {
			if k, v, ok := strings.Cut(string(kv), "="); ok && k != "" {
				out[k] = v
			}
		}
	}
	loginEnvMap, loginEnvAt = out, time.Now()
	return out
}

var envRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?\}`)

// expandRefs applies Claude's ${VAR} / ${VAR:-default} expansion.
func expandRefs(s string, env map[string]string) string {
	return envRef.ReplaceAllStringFunc(s, func(m string) string {
		sub := envRef.FindStringSubmatch(m)
		if v, ok := env[sub[1]]; ok && v != "" {
			return v
		}
		if sub[2] != "" {
			return sub[3]
		}
		return m
	})
}

// resolveSpec turns e into what the client would actually launch.
func resolveSpec(e Entry, env map[string]string) Spec {
	s := e.Spec.Normalize()
	home, _ := os.UserHomeDir()
	if e.Location.Client != ClientCodex {
		s.Command = expandRefs(s.Command, env)
		for i := range s.Args {
			s.Args[i] = expandRefs(s.Args[i], env)
		}
		s.URL = expandRefs(s.URL, env)
		s.Env = expandMap(s.Env, env)
		s.Headers = expandMap(s.Headers, env)
	}
	if strings.HasPrefix(s.Command, "~/") {
		s.Command = filepath.Join(home, s.Command[2:])
	}
	if s.Cwd != "" && !filepath.IsAbs(s.Cwd) {
		s.Cwd = filepath.Join(filepath.Dir(e.File), s.Cwd)
	}
	if e.Location.Client == ClientCodex {
		if s.Env == nil {
			s.Env = map[string]string{}
		}
		for _, k := range stringSlice(s.Extra["env_vars"]) {
			if v, ok := env[k]; ok {
				s.Env[k] = v
			}
		}
		if s.Headers == nil {
			s.Headers = map[string]string{}
		}
		if s.BearerTokenEnvVar != "" {
			s.Headers["Authorization"] = "Bearer " + env[s.BearerTokenEnvVar]
		}
		for h, k := range stringMap(s.Extra["env_http_headers"]) {
			s.Headers[h] = env[k]
		}
	}
	return s
}

func expandMap(m map[string]string, env map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = expandRefs(v, env)
	}
	return out
}

// ── stdio ───────────────────────────────────────────────────────────────────

type stdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	lines  chan []byte
	stderr *tailBuffer
	mu     sync.Mutex
}

func startStdio(ctx context.Context, s Spec, env map[string]string, file string) (*stdioTransport, error) {
	// Same as a terminal-launched client: login PATH + ~/.zshrc, then exec.
	args := append([]string{"-l", "-c", `source ~/.zshrc >/dev/null 2>&1; exec "$0" "$@"`, s.Command}, s.Args...)
	cmd := exec.CommandContext(ctx, "/bin/zsh", args...)
	merged := make([]string, 0, len(env)+len(s.Env))
	for k, v := range env {
		if _, override := s.Env[k]; !override {
			merged = append(merged, k+"="+v)
		}
	}
	for k, v := range s.Env {
		merged = append(merged, k+"="+v)
	}
	cmd.Env = merged
	if s.Cwd != "" {
		cmd.Dir = s.Cwd
	} else if file != "" {
		cmd.Dir = filepath.Dir(file)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	t := &stdioTransport{cmd: cmd, stdin: stdin, lines: make(chan []byte, 64), stderr: &tailBuffer{max: 4096}}
	cmd.Stderr = t.stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), 16<<20)
		for sc.Scan() {
			line := append([]byte(nil), sc.Bytes()...)
			t.lines <- line
		}
		close(t.lines)
	}()
	return t, nil
}

func (t *stdioTransport) write(msg any) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err = t.stdin.Write(append(b, '\n'))
	return err
}

func (t *stdioTransport) call(ctx context.Context, id int, method string, params any) (json.RawMessage, error) {
	if err := t.write(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
		return nil, fmt.Errorf("写入失败（进程已退出？）: %w", err)
	}
	want := fmt.Sprint(id)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case line, ok := <-t.lines:
			if !ok {
				return nil, fmt.Errorf("进程已退出%s", exitSuffix(t.cmd, t.stderr))
			}
			line = bytes.TrimSpace(line)
			if len(line) == 0 || line[0] != '{' {
				continue
			}
			var msg rpcMessage
			if json.Unmarshal(line, &msg) != nil {
				continue
			}
			if msg.Method != "" {
				if len(msg.ID) > 0 { // server → client request (roots/list, ping …)
					reply := map[string]any{"jsonrpc": "2.0", "id": msg.ID}
					if msg.Method == "ping" {
						reply["result"] = map[string]any{}
					} else {
						reply["error"] = map[string]any{"code": -32601, "message": "not supported by hatch probe"}
					}
					_ = t.write(reply)
				}
				continue
			}
			if string(msg.ID) != want {
				continue
			}
			if msg.Error != nil {
				return nil, fmt.Errorf("%s 返回错误 %d: %s", method, msg.Error.Code, msg.Error.Message)
			}
			return msg.Result, nil
		}
	}
}

func (t *stdioTransport) notify(_ context.Context, method string) error {
	return t.write(map[string]any{"jsonrpc": "2.0", "method": method})
}

func (t *stdioTransport) close() {
	_ = t.stdin.Close()
	if t.cmd.Process == nil {
		return
	}
	pgid := t.cmd.Process.Pid
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	done := make(chan struct{})
	go func() { _ = t.cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
}

func exitSuffix(cmd *exec.Cmd, stderr *tailBuffer) string {
	_ = cmd.Wait()
	tail := strings.TrimSpace(stderr.String())
	if tail == "" {
		return "，未输出任何内容"
	}
	lines := strings.Split(tail, "\n")
	return "：" + strings.TrimSpace(lines[len(lines)-1])
}

// ── streamable HTTP ─────────────────────────────────────────────────────────

type httpTransport struct {
	url      string
	headers  map[string]string
	session  string
	protocol string
}

var probeHTTP = &http.Client{Timeout: 30 * time.Second}

func (h *httpTransport) post(ctx context.Context, body any) (*http.Response, error) {
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range h.headers {
		if strings.TrimSpace(v) != "" && strings.TrimSpace(v) != "Bearer" {
			req.Header.Set(k, v)
		}
	}
	if h.session != "" {
		req.Header.Set("Mcp-Session-Id", h.session)
	}
	if h.protocol != "" {
		req.Header.Set("MCP-Protocol-Version", h.protocol)
	}
	resp, err := probeHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		return nil, authError{msg: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	if s := resp.Header.Get("Mcp-Session-Id"); s != "" {
		h.session = s
	}
	return resp, nil
}

func (h *httpTransport) call(ctx context.Context, id int, method string, params any) (json.RawMessage, error) {
	resp, err := h.post(ctx, map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	want := fmt.Sprint(id)
	handle := func(data []byte) (json.RawMessage, bool, error) {
		var msg rpcMessage
		if json.Unmarshal(data, &msg) != nil || string(msg.ID) != want || msg.Method != "" {
			return nil, false, nil
		}
		if msg.Error != nil {
			return nil, true, fmt.Errorf("%s 返回错误 %d: %s", method, msg.Error.Code, msg.Error.Message)
		}
		return msg.Result, true, nil
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		sc := bufio.NewScanner(resp.Body)
		sc.Buffer(make([]byte, 0, 64*1024), 16<<20)
		var data bytes.Buffer
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "data:") {
				data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
				continue
			}
			if line == "" && data.Len() > 0 {
				if r, ok, err := handle(data.Bytes()); ok {
					return r, err
				}
				data.Reset()
			}
		}
		if data.Len() > 0 {
			if r, ok, err := handle(data.Bytes()); ok {
				return r, err
			}
		}
		return nil, fmt.Errorf("%s: 事件流中没有响应", method)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if r, ok, err := handle(b); ok {
		return r, err
	}
	return nil, fmt.Errorf("%s: 无法解析响应", method)
}

func (h *httpTransport) notify(ctx context.Context, method string) error {
	resp, err := h.post(ctx, map[string]any{"jsonrpc": "2.0", "method": method})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (h *httpTransport) close() {
	if h.session == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, h.url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Mcp-Session-Id", h.session)
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}
	if resp, err := probeHTTP.Do(req); err == nil {
		resp.Body.Close()
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

type tailBuffer struct {
	mu  sync.Mutex
	buf []byte
	max int
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, p...)
	if len(t.buf) > t.max {
		t.buf = t.buf[len(t.buf)-t.max:]
	}
	return len(p), nil
}

func (t *tailBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.buf)
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
