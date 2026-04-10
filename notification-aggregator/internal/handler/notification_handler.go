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
	notes, err := h.svc.AggregateAndSave(r.Context())
	if err != nil {
		// 後で Stage 4 (slog) で詳細に出力しますが、一旦簡易エラーハンドリング
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}
