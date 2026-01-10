package slogpretty

import (
	"context"
	"encoding/json"
	"io"
	stdlog "log"
	"log/slog"
	"strings"
	"time"

	"github.com/fatih/color"
)

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type PrettyHandler struct {
	opts PrettyHandlerOptions
	slog.Handler
	l     *stdlog.Logger
	attrs []slog.Attr
}

func (opts PrettyHandlerOptions) NewPrettyHandler(out io.Writer) *PrettyHandler {
	if opts.SlogOpts == nil {
		opts.SlogOpts = &slog.HandlerOptions{}
	}

	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdlog.New(out, "", 0),
		opts:    opts,
	}

	return h
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	n := *h
	n.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	n.Handler = h.Handler.WithAttrs(attrs)
	return &n
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	n := *h
	n.Handler = h.Handler.WithGroup(name)
	return &n
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	// message
	msg := r.Message
	if msg == "" {
		msg = "-"
	}

	// time
	t := r.Time
	if t.IsZero() {
		t = time.Now()
	}
	ts := color.HiBlackString(t.Format("2006-01-02 15:04:05"))

	// collect attrs (including WithAttrs)
	attrs := make(map[string]any, len(h.attrs)+8)
	for _, a := range h.attrs {
		attrs[a.Key] = valueToAny(a.Value)
	}

	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = valueToAny(a.Value)
		return true
	})

	// print fields as compact json at the end
	fields := ""
	if len(attrs) > 0 {
		b, _ := json.Marshal(attrs)
		fields = " " + color.HiBlackString(strings.TrimSpace(string(b)))
	}

	h.l.Println(ts, level, msg+fields)
	return nil
}

func valueToAny(v slog.Value) any {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time()
	case slog.KindGroup:
		m := map[string]any{}
		for _, a := range v.Group() {
			m[a.Key] = valueToAny(a.Value)
		}
		return m
	case slog.KindAny:
		return v.Any()
	default:
		return v.String()
	}
}
