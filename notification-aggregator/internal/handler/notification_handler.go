package handler

import (
	"encoding/json"
	"net/http"
	"notification-aggregator/internal/notification"
)

type NotificationHandler struct {
	svc *notification.Service
}

func NewNotificationHandler(svc *notification.Service) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ここで Service の AggregateAll を呼び出す
	// r.Context() を渡すことで、ブラウザがリクエストをキャンセルした際に
	// 背後の goroutine（Provider へのリクエスト）も連動して止まるようになります。
	notes := h.svc.AggregateAll(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notes); err != nil {
		// ログ出力などのエラーハンドリング（Stage 4で詳細化）
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
