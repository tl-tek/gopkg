// copied and modified from https://github.com/samber/slog-formatter/blob/4ce6c72294ee982ed459d49a2faac326a257181e/handler.go

package logger

import (
	"context"
	"log/slog"
)

type Option struct {
	// optional: applies formatters
	Formatters []Formatter
	// optional: fetch attributes from context
	AttrFromContext []func(ctx context.Context) []slog.Attr
}

func (o Option) NewHandler() func(slog.Handler) slog.Handler {
	return func(handler slog.Handler) slog.Handler {
		return &Handler{
			groups:          []string{},
			handler:         handler,
			formatters:      o.Formatters,
			attrFromContext: o.AttrFromContext,
		}
	}
}

var _ slog.Handler = (*Handler)(nil)

type Handler struct {
	groups          []string
	handler         slog.Handler
	formatters      []Formatter
	attrFromContext []func(ctx context.Context) []slog.Attr
}

// Enabled implements slog.Handler.
func (h *Handler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.handler.Enabled(ctx, l)
}

// Handle implements slog.Handler.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	r2 := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(attr slog.Attr) bool {
		r2.AddAttrs(h.transformAttr(h.groups, attr))
		return true
	})

	if len(h.attrFromContext) > 0 {
		for _, f := range h.attrFromContext {
			r2.AddAttrs(f(ctx)...)
		}
	}

	return h.handler.Handle(ctx, r2)
}

// WithAttrs implements slog.Handler.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	attrs = h.transformAttrs(h.groups, attrs)

	return &Handler{
		groups:          h.groups,
		handler:         h.handler.WithAttrs(attrs),
		formatters:      h.formatters,
		attrFromContext: h.attrFromContext,
	}
}

// WithGroup implements slog.Handler.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		groups:          append(h.groups, name),
		handler:         h.handler.WithGroup(name),
		formatters:      h.formatters,
		attrFromContext: h.attrFromContext,
	}
}

func (h *Handler) transformAttrs(groups []string, attrs []slog.Attr) []slog.Attr {
	for i := range attrs {
		attrs[i] = h.transformAttr(groups, attrs[i])
	}
	return attrs
}

func (h *Handler) transformAttr(groups []string, attr slog.Attr) slog.Attr {
	for attr.Value.Kind() == slog.KindLogValuer {
		attr.Value = attr.Value.LogValuer().LogValue()
	}

	for _, formatter := range h.formatters {
		if v, ok := formatter(groups, attr); ok {
			attr.Value = v
		}
	}

	return attr
}
