// Package llm is Hatch's OpenAI-compatible chat client plus the JSONL trace
// store behind the AI call log. It knows nothing about Hatch features; callers
// pass in the system/user prompts and decide what to do with the reply.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hatch/core/config"
)

// Trace records one AI API call for inspection in the frontend.
type Trace struct {
	ID           int64  `json:"id"`
	At           string `json:"at"`
	Op           string `json:"op"`
	Model        string `json:"model"`
	URL          string `json:"url"`
	SystemPrompt string `json:"systemPrompt"`
	UserPrompt   string `json:"userPrompt"`
	RawResponse  string `json:"rawResponse"`
	Error        string `json:"error"`
	DurationMs   int64  `json:"durationMs"`
}

// Validate ensures the AI config has the minimum required fields.
func Validate(ai config.AIConfig) error {
	if !ai.Enabled {
		return fmt.Errorf("AI 配置未启用")
	}
	if strings.TrimSpace(ai.BaseURL) == "" {
		return fmt.Errorf("请先配置 API Base URL")
	}
	if strings.TrimSpace(ai.APIKey) == "" {
		return fmt.Errorf("请先配置 API Key")
	}
	if strings.TrimSpace(ai.Model) == "" {
		return fmt.Errorf("请先配置模型名")
	}
	return nil
}

// Chat sends one system+user exchange to <baseUrl>/chat/completions and returns
// the first choice's content. The returned Trace (without ID) is always filled,
// including on failure, so callers can record it.
func Chat(ctx context.Context, op string, ai config.AIConfig, systemPrompt, userPrompt string) (string, Trace, error) {
	apiURL := strings.TrimRight(strings.TrimSpace(ai.BaseURL), "/")
	if !strings.HasSuffix(apiURL, "/chat/completions") {
		apiURL += "/chat/completions"
	}
	trace := Trace{
		Op:           op,
		Model:        strings.TrimSpace(ai.Model),
		URL:          apiURL,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
	if err := Validate(ai); err != nil {
		return "", trace, err
	}

	payload, err := json.Marshal(map[string]any{
		"model": strings.TrimSpace(ai.Model),
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
	})
	if err != nil {
		return "", trace, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return "", trace, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(ai.APIKey))

	start := time.Now()
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	trace.DurationMs = time.Since(start).Milliseconds()
	trace.At = time.Now().Format("15:04:05")
	if err != nil {
		trace.Error = err.Error()
		return "", trace, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	trace.RawResponse = string(data)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		trace.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return "", trace, fmt.Errorf("AI 请求失败: HTTP %d %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if len(data) == 0 {
		trace.Error = fmt.Sprintf("空响应 (HTTP %d)", resp.StatusCode)
		return "", trace, fmt.Errorf("AI 返回空响应 (HTTP %d)", resp.StatusCode)
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		preview := string(data)
		if len(preview) > 300 {
			preview = preview[:300] + "…"
		}
		trace.Error = fmt.Sprintf("解析失败: %v", err)
		return "", trace, fmt.Errorf("解析 AI 响应失败: %w — 收到: %s", err, preview)
	}
	if len(decoded.Choices) == 0 {
		trace.Error = "choices 为空"
		return "", trace, fmt.Errorf("AI 返回为空")
	}
	return strings.TrimSpace(decoded.Choices[0].Message.Content), trace, nil
}

// ExtractJSONObject strips Markdown fences/prose and returns the outermost
// {...} span of s (or s itself when no object is found).
func ExtractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

// StripFence removes one leading code fence from the given list (e.g. "```sh",
// "```") and a trailing "```", returning the trimmed body.
func StripFence(raw string, fences ...string) string {
	raw = strings.TrimSpace(raw)
	for _, fence := range fences {
		if strings.HasPrefix(raw, fence) {
			raw = strings.TrimPrefix(raw, fence)
			raw = strings.TrimLeft(raw, " \t\r\n")
			break
		}
	}
	raw = strings.TrimSuffix(strings.TrimSpace(raw), "```")
	return strings.TrimSpace(raw)
}
