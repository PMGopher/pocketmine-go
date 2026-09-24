package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"pocketmine-go/pocketmine/log"
)

// slogHandler routes gophertunnel's log/slog output (connection and packet decoding errors) into
// the server's logger. It has no PocketMine-MP counterpart: RakLib and the protocol library log
// through the server logger directly. Without it gophertunnel discards these messages.
type slogHandler struct {
	logger log.Logger
	attrs  []slog.Attr
}

func newSlogLogger(logger log.Logger) *slog.Logger {
	return slog.New(&slogHandler{logger: logger})
}

func (h *slogHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Message)
	write := func(a slog.Attr) bool {
		fmt.Fprintf(&b, " %s=%v", a.Key, a.Value)
		return true
	}
	for _, a := range h.attrs {
		write(a)
	}
	r.Attrs(write)

	msg := b.String()
	switch {
	case r.Level >= slog.LevelError:
		h.logger.Error(msg)
	case r.Level >= slog.LevelWarn:
		h.logger.Warning(msg)
	case r.Level >= slog.LevelInfo:
		h.logger.Info(msg)
	default:
		h.logger.Debug(msg)
	}
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{logger: h.logger, attrs: append(append([]slog.Attr(nil), h.attrs...), attrs...)}
}

func (h *slogHandler) WithGroup(string) slog.Handler { return h }
