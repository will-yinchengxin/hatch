package app

import "hatch/core/config"

// ─── AI · 文档链接 ───────────────────────────────────────────────────────────

// GetDocLinks returns the configured quick-access doc links.
func (a *App) GetDocLinks() []config.DocLink {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]config.DocLink, len(a.cfg.DocLinks))
	copy(out, a.cfg.DocLinks)
	return out
}

// SaveDocLinks persists the doc links list.
func (a *App) SaveDocLinks(links []config.DocLink) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.DocLinks = links
	return a.cfg.Save()
}
