package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"hatch/ai/llm"
	"hatch/ai/prompts"
	"hatch/core/config"
	"hatch/core/fsutil"
	"hatch/workbench/zshrc"
)

// ─── 工作台 · AI 辅助 ────────────────────────────────────────────────────────
//
// Every suggestion is returned for the user to confirm in the UI; nothing here
// persists metadata or rewrites files.

// AIScriptSuggestion is the AI-filled script metadata draft.
type AIScriptSuggestion struct {
	Name      string                 `json:"name"`
	Desc      string                 `json:"desc"`
	Type      config.ScriptType      `json:"type"`
	Commands  []string               `json:"commands"`
	PidFile   string                 `json:"pidFile"`
	NeedsArgs bool                   `json:"needsArgs"`
	Args      []config.ArgDef        `json:"args"`
	Tags      []string               `json:"tags"`
	Presets   []config.CommandPreset `json:"presets"`
}

// ZshrcOrganizationSuggestion is a proposed full ~/.zshrc rewrite.
type ZshrcOrganizationSuggestion struct {
	Summary []string `json:"summary"`
	Content string   `json:"content"`
}

// SuggestScriptMeta reads a script and asks the configured model for metadata.
func (a *App) SuggestScriptMeta(meta config.ScriptMeta) (AIScriptSuggestion, error) {
	ai, err := a.readyAI()
	if err != nil {
		return AIScriptSuggestion{}, err
	}
	tf, err := fsutil.ReadTextFile(meta.Path, 12000)
	if err != nil {
		return AIScriptSuggestion{}, err
	}
	if tf.Binary {
		return AIScriptSuggestion{}, fmt.Errorf("二进制文件无法让 AI 分析，请在「配置」里手动填写")
	}
	content := tf.Content
	userBytes, _ := json.Marshal(map[string]any{
		"path":     meta.Path,
		"name":     meta.Name,
		"desc":     meta.Desc,
		"type":     meta.Type,
		"language": meta.Language,
		"commands": meta.Commands,
		"pidFile":  meta.PidFile,
		"tags":     meta.Tags,
		"content":  content,
	})
	raw, err := a.callAIOp(prompts.OpSuggestScriptMeta, ai, prompts.ScriptMeta, string(userBytes))
	if err != nil {
		return AIScriptSuggestion{}, err
	}
	var out AIScriptSuggestion
	if err := json.Unmarshal([]byte(llm.ExtractJSONObject(raw)), &out); err != nil {
		return AIScriptSuggestion{}, fmt.Errorf("AI 返回无法解析: %w", err)
	}
	if out.Type != config.TypeService {
		out.Type = config.TypeOneshot
		out.Commands = nil
		out.PidFile = ""
	}
	return out, nil
}

// SuggestScriptContent asks the configured AI to write or improve a script.
func (a *App) SuggestScriptContent(path, currentContent, instruction string) (string, error) {
	ai, err := a.readyAI()
	if err != nil {
		return "", err
	}
	userBytes, _ := json.Marshal(map[string]any{
		"path":           path,
		"currentContent": currentContent,
		"instruction":    instruction,
	})
	raw, err := a.callAIOp(prompts.OpSuggestScriptContent, ai, prompts.ScriptContent, string(userBytes))
	if err != nil {
		return "", err
	}
	raw = llm.StripFence(raw, "```zsh", "```shell", "```sh", "```bash", "```python", "```ruby", "```node", "```")
	if raw == "" {
		return "", fmt.Errorf("AI 返回了空内容（原始响应可在 AI 调用日志中查看）")
	}
	return raw, nil
}

// SuggestScriptContentStream delivers SuggestScriptContent via "ai:stream"
// events ({id, chunk, done, error}). It currently emits the whole result as one
// chunk, keeping the streaming API stable for a future SSE implementation.
func (a *App) SuggestScriptContentStream(id, path, currentContent, instruction string) {
	go func() {
		text, err := a.SuggestScriptContent(path, currentContent, instruction)
		payload := map[string]any{"id": id, "chunk": text, "done": true, "error": ""}
		if err != nil {
			payload["chunk"], payload["error"] = "", err.Error()
		}
		a.emit("ai:stream", payload)
	}()
}

// SuggestZshrcAnnotation asks the configured model for a short note for one zshrc entry.
func (a *App) SuggestZshrcAnnotation(entry zshrc.Entry) (string, error) {
	ai, err := a.readyAI()
	if err != nil {
		return "", err
	}
	userBytes, _ := json.Marshal(entry)
	raw, err := a.callAIOp(prompts.OpSuggestZshrcAnnotation, ai, prompts.ZshrcAnnotation, string(userBytes))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}

// SuggestZshrcOrganization asks AI to reorganize ~/.zshrc and returns a preview;
// ApplyZshrcOrganization writes it after user approval.
func (a *App) SuggestZshrcOrganization(mode string) (ZshrcOrganizationSuggestion, error) {
	ai, err := a.readyAI()
	if err != nil {
		return ZshrcOrganizationSuggestion{}, err
	}
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return ZshrcOrganizationSuggestion{}, err
	}
	content, err := readZshrc()
	if err != nil {
		return ZshrcOrganizationSuggestion{}, err
	}
	if len(content) > 24000 {
		content = content[:24000]
	}
	if strings.TrimSpace(mode) == "" {
		mode = "按功能模块优先分组；同类 export/PATH/alias/function/source 放在相近区域。"
	}
	userBytes, _ := json.Marshal(map[string]any{
		"mode":    mode,
		"path":    p,
		"entries": zshrc.ParseContent(content, nil),
		"content": content,
	})
	raw, err := a.callAIOp(prompts.OpSuggestZshrcOrganization, ai, prompts.ZshrcOrganization, string(userBytes))
	if err != nil {
		return ZshrcOrganizationSuggestion{}, err
	}
	var out ZshrcOrganizationSuggestion
	if err := json.Unmarshal([]byte(llm.ExtractJSONObject(raw)), &out); err != nil {
		return ZshrcOrganizationSuggestion{}, fmt.Errorf("AI 返回无法解析: %w", err)
	}
	if strings.TrimSpace(out.Content) == "" {
		return ZshrcOrganizationSuggestion{}, fmt.Errorf("AI 未返回整理后的 zshrc 内容")
	}
	return out, nil
}
