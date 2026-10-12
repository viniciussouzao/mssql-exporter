//go:build windows

package main

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"

	"golang.org/x/sys/windows/svc/eventlog"
)

// eventLogID is the event ID used for every entry; the source is
// registered with EventCreate.exe, which accepts IDs 1-1000.
const eventLogID = 1

// eventLogHandler is a slog.Handler that formats records like the default
// text handler and writes them to the Windows Event Log at the matching
// level (Error, Warning or Information).
type eventLogHandler struct {
	log   *eventlog.Log
	inner slog.Handler
	mu    *sync.Mutex
	buf   *bytes.Buffer
}

func newEventLogHandler(l *eventlog.Log, level slog.Leveler) *eventLogHandler {
	buf := &bytes.Buffer{}
	return &eventLogHandler{
		log: l,
		inner: slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: level,
			// The Event Log records its own timestamp.
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if len(groups) == 0 && a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}),
		mu:  &sync.Mutex{},
		buf: buf,
	}
}

func (h *eventLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *eventLogHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.buf.Reset()
	if err := h.inner.Handle(ctx, r); err != nil {
		return err
	}
	msg := strings.TrimSpace(h.buf.String())

	switch {
	case r.Level >= slog.LevelError:
		return h.log.Error(eventLogID, msg)
	case r.Level >= slog.LevelWarn:
		return h.log.Warning(eventLogID, msg)
	default:
		return h.log.Info(eventLogID, msg)
	}
}

// WithAttrs and WithGroup share the buffer and mutex: the derived inner
// handlers keep writing to the same buffer.
func (h *eventLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &eventLogHandler{log: h.log, inner: h.inner.WithAttrs(attrs), mu: h.mu, buf: h.buf}
}

func (h *eventLogHandler) WithGroup(name string) slog.Handler {
	return &eventLogHandler{log: h.log, inner: h.inner.WithGroup(name), mu: h.mu, buf: h.buf}
}
