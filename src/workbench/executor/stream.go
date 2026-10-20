package executor

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	EventLog    = "executor:log"    // {id, line, stream}  stream: stdout|stderr
	EventDone   = "executor:done"   // {id, exitCode}
	EventStatus = "executor:status" // {id, status}  status: running|stopped|error
)

type LogEvent struct {
	ID     string `json:"id"`
	Line   string `json:"line"`
	Stream string `json:"stream"` // "stdout" | "stderr"
}

type DoneEvent struct {
	ID       string `json:"id"`
	ExitCode int    `json:"exitCode"`
}

type StatusEvent struct {
	ID     string `json:"id"`
	Status string `json:"status"` // "running" | "stopped" | "error"
}

// Emitter wraps a Wails context and provides typed event helpers.
type Emitter struct {
	ctx context.Context
}

func NewEmitter(ctx context.Context) *Emitter {
	return &Emitter{ctx: ctx}
}

func (e *Emitter) Log(id, line, stream string) {
	runtime.EventsEmit(e.ctx, EventLog, LogEvent{ID: id, Line: line, Stream: stream})
}

func (e *Emitter) Done(id string, exitCode int) {
	runtime.EventsEmit(e.ctx, EventDone, DoneEvent{ID: id, ExitCode: exitCode})
}

func (e *Emitter) Status(id, status string) {
	runtime.EventsEmit(e.ctx, EventStatus, StatusEvent{ID: id, Status: status})
}
