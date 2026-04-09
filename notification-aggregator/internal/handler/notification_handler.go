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
	notes, err := h.svc.GetNotifications(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}
