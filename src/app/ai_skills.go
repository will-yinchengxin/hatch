package app

import (
	"encoding/json"
	"fmt"

	"hatch/ai/llm"
	"hatch/ai/prompts"
	"hatch/ai/skills"
)

// ─── AI · Skills ─────────────────────────────────────────────────────────────

// skillStore opens the configured (or default) SKILL.md master store.
func (a *App) skillStore() (skills.Store, error) {
	a.mu.RLock()
	dir := a.cfg.SkillsDir
	a.mu.RUnlock()
	return skills.Open(dir)
}

// GetSkills returns all skills in the master store with their sync status.
func (a *App) GetSkills() ([]skills.Entry, error) {
	st, err := a.skillStore()
	if err != nil {
		return nil, err
	}
	return st.List()
}

// GetSkillsStats returns quick counts for the hero stats.
func (a *App) GetSkillsStats() skills.Stats {
	list, err := a.GetSkills()
	if err != nil {
		return skills.Stats{}
	}
	return skills.CountStats(list)
}

// GetSkillContent returns the SKILL.md content for one skill.
func (a *App) GetSkillContent(name string) (string, error) {
	st, err := a.skillStore()
	if err != nil {
		return "", err
	}
	return st.Content(name)
}

// SaveSkillContent creates or updates SKILL.md in the master store.
func (a *App) SaveSkillContent(name, content string) error {
	st, err := a.skillStore()
	if err != nil {
		return err
	}
	return st.Save(name, content)
}

// DeleteSkill removes a skill from the master store only; synced copies in
// ~/.claude/skills and ~/.codex/skills are left alone.
func (a *App) DeleteSkill(name string) error {
	st, err := a.skillStore()
	if err != nil {
		return err
	}
	return st.Delete(name)
}

// DeleteSkillEverywhere removes a skill from the master store and both sync targets.
func (a *App) DeleteSkillEverywhere(name string) error {
	st, err := a.skillStore()
	if err != nil {
		return err
	}
	return st.DeleteEverywhere(name)
}

// SetSkillSync copies (enabled) or removes one skill in target ("claude" | "codex").
func (a *App) SetSkillSync(name, target string, enabled bool) error {
	st, err := a.skillStore()
	if err != nil {
		return err
	}
	return st.SetSync(name, target, enabled)
}

// ImportSkillsFrom copies new skills from source ("claude" | "codex") into the
// master store; same-name skills with different content come back as conflicts.
func (a *App) ImportSkillsFrom(source string) (skills.ImportResult, error) {
	st, err := a.skillStore()
	if err != nil {
		return skills.ImportResult{}, err
	}
	return st.ImportFrom(source)
}

// ResolveSkillImportConflict applies (useSource) or dismisses one import conflict.
func (a *App) ResolveSkillImportConflict(name, source string, useSource bool) error {
	st, err := a.skillStore()
	if err != nil {
		return err
	}
	return st.ResolveConflict(name, source, useSource)
}

// CreateSkillTemplate returns a starter SKILL.md template for a new skill.
func (a *App) CreateSkillTemplate(name string) string {
	return skills.Template(name)
}

// SuggestSkillContent asks the configured AI to optimize a SKILL.md and returns
// the rewritten markdown (fences stripped) for user confirmation.
func (a *App) SuggestSkillContent(name, currentContent, instruction string) (string, error) {
	ai, err := a.readyAI()
	if err != nil {
		return "", err
	}
	userBytes, _ := json.Marshal(map[string]any{
		"name":           name,
		"currentContent": currentContent,
		"instruction":    instruction,
	})
	raw, err := a.callAIOp(prompts.OpSuggestSkillContent, ai, prompts.SkillContent, string(userBytes))
	if err != nil {
		return "", err
	}
	raw = llm.StripFence(raw, "```markdown", "```md", "```")
	if raw == "" {
		return "", fmt.Errorf("AI 返回了空内容（原始响应可在 AI 调用日志中查看）")
	}
	return raw, nil
}

// SuggestSkillFromDescription drafts a brand-new SKILL.md from a one-line description.
func (a *App) SuggestSkillFromDescription(name, description string) (string, error) {
	return a.SuggestSkillContent(name, "", "为这个新 skill 起草一份完整的 SKILL.md。description: "+description)
}
