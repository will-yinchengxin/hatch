package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"hatch/ai/hooks"
	"hatch/ai/prompts"
	"hatch/core/config"
)

// ─── AI · Hooks ──────────────────────────────────────────────────────────────
//
// The registry lives in config.Hooks; internal/ai/hooks renders and syncs it
// into ~/.claude/settings.json and ~/.codex/hooks.json.

// GetHookEventSupport returns the event catalog for the UI dropdown and toggle gating.
func (a *App) GetHookEventSupport() []hooks.EventInfo {
	return hooks.Events()
}

// GetHookTemplates returns the built-in starter hooks gallery.
func (a *App) GetHookTemplates() []config.HookDef {
	return hooks.Templates()
}

// ListHooks returns the hook registry, newest first. On first call it seeds the
// default "hello" test hook and enables it for Claude Code (unless a
// hand-written hook would conflict), so the capability works out of the box.
func (a *App) ListHooks() ([]config.HookDef, error) {
	a.mu.Lock()
	if !a.cfg.HooksSeeded {
		def := hooks.DefaultHello()
		if conflict, err := hooks.ClaudeEntryExists(def); err == nil && !conflict {
			if werr := hooks.WriteClaude(def, true); werr == nil {
				def.EnabledInClaude = true
			}
		}
		a.cfg.Hooks = append(a.cfg.Hooks, def)
		a.cfg.HooksSeeded = true
		_ = a.cfg.Save()
	}
	out := hooks.Clone(a.cfg.Hooks)
	a.mu.Unlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt == out[j].UpdatedAt {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out, nil
}

// GetHooksStats returns the sidenav counts.
func (a *App) GetHooksStats() hooks.Stats {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return hooks.CountStats(a.cfg.Hooks)
}

// SaveHook creates or updates a hook (matched by ID), preserving sync flags.
// It does not write any target file; the frontend re-syncs enabled targets after save.
func (a *App) SaveHook(hook config.HookDef) error {
	hook, err := hooks.Normalize(hook)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, h := range a.cfg.Hooks {
		if h.ID == hook.ID {
			hook.EnabledInClaude = h.EnabledInClaude
			hook.EnabledInCodex = h.EnabledInCodex
			a.cfg.Hooks[i] = hook
			return a.cfg.Save()
		}
	}
	a.cfg.Hooks = append(a.cfg.Hooks, hook)
	return a.cfg.Save()
}

// DeleteHook removes a hook from the registry and from any target it was synced to.
func (a *App) DeleteHook(id string) error {
	if !hooks.ValidID(id) {
		return fmt.Errorf("invalid hook id")
	}
	a.mu.Lock()
	removed, found := hooks.Find(a.cfg.Hooks, id)
	if !found {
		a.mu.Unlock()
		return nil
	}
	keep := a.cfg.Hooks[:0]
	for _, h := range a.cfg.Hooks {
		if h.ID != id {
			keep = append(keep, h)
		}
	}
	a.cfg.Hooks = keep
	err := a.cfg.Save()
	remaining := hooks.Clone(a.cfg.Hooks)
	a.mu.Unlock()
	if err != nil {
		return err
	}

	if removed.EnabledInClaude {
		if err := hooks.WriteClaude(removed, false); err != nil {
			return err
		}
	}
	if removed.EnabledInCodex {
		return hooks.WriteCodex(remaining)
	}
	return nil
}

// SetHookEnabled syncs (or removes) one hook to/from target ("claude" | "codex").
// First-time enable refuses to clobber a matching hand-written hook; the
// config flag is only committed after the target file was written.
func (a *App) SetHookEnabled(id, target string, enabled bool) error {
	if !hooks.ValidID(id) {
		return fmt.Errorf("invalid hook id")
	}
	if target != hooks.TargetClaude && target != hooks.TargetCodex {
		return fmt.Errorf("unknown hook target: %s", target)
	}

	a.mu.RLock()
	hook, found := hooks.Find(a.cfg.Hooks, id)
	a.mu.RUnlock()
	if !found {
		return fmt.Errorf("hook not found: %s", id)
	}

	if enabled && !hooks.SupportedBy(hook.Event, target) {
		label := "Codex"
		if target == hooks.TargetClaude {
			label = "Claude Code"
		}
		return fmt.Errorf("事件「%s」不被 %s 支持", hook.Event, label)
	}

	switch target {
	case hooks.TargetClaude:
		// Re-syncing an already-enabled hook (edit/update) overwrites its own marked handler.
		if enabled && !hook.EnabledInClaude {
			exists, err := hooks.ClaudeEntryExists(hook)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("Claude Code 配置中已存在匹配的手写钩子，已跳过以避免覆盖；如需 Hatch 接管请先在 ~/.claude/settings.json 中删除该条目")
			}
		}
		if err := hooks.WriteClaude(hook, enabled); err != nil {
			return err
		}
	case hooks.TargetCodex:
		if enabled && !hook.EnabledInCodex {
			exists, err := hooks.CodexEntryExists(hook)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("Codex 配置中已存在匹配的手写钩子，已跳过以避免覆盖；如需 Hatch 接管请先在 ~/.codex/hooks.json 中删除该条目")
			}
		}
		// Codex's file is regenerated from the whole registry with the flag applied.
		a.mu.RLock()
		list := hooks.Clone(a.cfg.Hooks)
		a.mu.RUnlock()
		for i := range list {
			if list[i].ID == id {
				list[i].EnabledInCodex = enabled
			}
		}
		if err := hooks.WriteCodex(list); err != nil {
			return err
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.cfg.Hooks {
		if a.cfg.Hooks[i].ID != id {
			continue
		}
		if target == hooks.TargetClaude {
			a.cfg.Hooks[i].EnabledInClaude = enabled
		} else {
			a.cfg.Hooks[i].EnabledInCodex = enabled
		}
		a.cfg.Hooks[i].UpdatedAt = time.Now().Unix()
		break
	}
	return a.cfg.Save()
}

// SuggestHookFromDescription asks the configured AI to draft a command hook
// from a natural-language description. The draft is never persisted here.
func (a *App) SuggestHookFromDescription(description string) (config.HookDef, error) {
	ai, err := a.readyAI()
	if err != nil {
		return config.HookDef{}, err
	}
	userBytes, _ := json.Marshal(map[string]any{"description": description})
	raw, err := a.callAIOp(prompts.OpSuggestHookFromDescription, ai, prompts.HookDraft(), string(userBytes))
	if err != nil {
		return config.HookDef{}, err
	}
	return hooks.ParseDraft(raw)
}

// SuggestHookEdit asks the configured AI to refine an existing hook given an
// instruction. The improved draft keeps current.ID and is never persisted here.
func (a *App) SuggestHookEdit(current config.HookDef, instruction string) (config.HookDef, error) {
	ai, err := a.readyAI()
	if err != nil {
		return config.HookDef{}, err
	}
	userBytes, _ := json.Marshal(map[string]any{
		"current":     current,
		"instruction": instruction,
	})
	raw, err := a.callAIOp(prompts.OpSuggestHookEdit, ai, prompts.HookEdit(), string(userBytes))
	if err != nil {
		return config.HookDef{}, err
	}
	draft, err := hooks.ParseDraft(raw)
	if err != nil {
		return config.HookDef{}, err
	}
	if current.ID != "" {
		draft.ID = current.ID
	}
	return draft, nil
}
