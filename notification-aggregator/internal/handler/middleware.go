package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"notification-aggregator/internal/contextutil"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// RequestIDMiddleware は各リクエストに一意の ID を付与します。
// これにより、複数のログを一つのリクエストとして相関させることができます。
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// クライアントから ID が送られてきている場合はそれを使用、なければ生成
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.New().String()
		}

		// contextutil を使ってセット
		ctx := contextutil.WithRequestID(r.Context(), id)

		// レスポンスヘッダーにも付与（調査時に便利）
		w.Header().Set("X-Request-Id", id)
		// 次のハンドラーに新しい Context を持ったリクエストを渡す
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			reqID := contextutil.GetRequestID(r.Context())

			// JSONログを出力
			logger.Info("http_request",
				slog.Group("http",
					slog.Int("status", sw.status),
					slog.Duration("duration", time.Since(start)),
					slog.String("request_id", reqID),
					slog.String("path", r.URL.Path),
				),
			)
		})
	}
}
