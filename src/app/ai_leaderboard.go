package app

import (
	"context"

	"hatch/ai/leaderboard"
	"hatch/core/fsutil"
)

func (a *App) modelLeaderboards() *leaderboard.Service {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.leaderboards == nil {
		a.leaderboards = leaderboard.New(fsutil.AppDataPath("model-leaderboards.json"))
	}
	return a.leaderboards
}

// GetModelLeaderboards returns the on-disk cache immediately; it never touches the network.
func (a *App) GetModelLeaderboards() leaderboard.Snapshot {
	return a.modelLeaderboards().Snapshot()
}

// RefreshModelLeaderboards fetches stale sources (all of them when force) in parallel.
// Per-source failures are reported in Board.Error and keep the previous data.
func (a *App) RefreshModelLeaderboards(force bool) leaderboard.Snapshot {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.modelLeaderboards().Refresh(ctx, force)
}
