package logging

import (
	"context"
	"log/slog"
	"notification-aggregator/internal/contextutil"
)

// ContextHandler は、Context 内の値をログ属性に自動追加するラッパーです
type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) handle(ctx context.Context, r slog.Record) error {
	// 既存の contextutil を使って ID を取得
	if id := contextutil.GetRequestID(ctx); id != "unknown" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}
