package app

import (
	"hatch/ai/prompts"
	"hatch/core/config"
)

// ─── AI · 连接（OpenAI 兼容 provider）与提示词 ───────────────────────────────

// AIProvidersResult is the provider list plus the active one.
type AIProvidersResult struct {
	Providers        []config.AIProvider `json:"providers"`
	ActiveProviderID string              `json:"activeProviderId"`
}

// SaveAIConfig persists OpenAI-compatible model settings.
func (a *App) SaveAIConfig(ai config.AIConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.AI = ai
	return a.cfg.Save()
}

func (a *App) GetAIProviders() AIProvidersResult {
	a.mu.RLock()
	defer a.mu.RUnlock()
	providers := make([]config.AIProvider, len(a.cfg.AIProviders))
	copy(providers, a.cfg.AIProviders)
	return AIProvidersResult{Providers: providers, ActiveProviderID: a.cfg.ActiveProviderID}
}

// SaveAIProviders replaces the provider list and mirrors the active provider
// into the legacy cfg.AI fields.
func (a *App) SaveAIProviders(providers []config.AIProvider, activeID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.AIProviders = providers
	a.cfg.ActiveProviderID = activeID
	for _, p := range providers {
		if p.ID == activeID {
			a.cfg.AI.BaseURL = p.BaseURL
			a.cfg.AI.APIKey = p.APIKey
			a.cfg.AI.Model = p.Model
			break
		}
	}
	return a.cfg.Save()
}

// TestAIConfig sends a tiny request to the given OpenAI-compatible API.
func (a *App) TestAIConfig(ai config.AIConfig) (string, error) {
	ai.Enabled = true
	return a.callAIOp(prompts.OpTestAIConfig, ai, prompts.TestAIConfig, "ping")
}

// GetPromptTemplates returns the user's prompt overrides (op → system prompt).
func (a *App) GetPromptTemplates() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]string, len(a.cfg.PromptTemplates))
	for k, v := range a.cfg.PromptTemplates {
		out[k] = v
	}
	return out
}

// GetDefaultPrompts returns the built-in system prompts so the UI can show
// exactly what an override replaces.
func (a *App) GetDefaultPrompts() map[string]string {
	return prompts.Defaults()
}

// SavePromptTemplates replaces the prompt overrides map.
func (a *App) SavePromptTemplates(tpl map[string]string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.PromptTemplates = tpl
	return a.cfg.Save()
}
