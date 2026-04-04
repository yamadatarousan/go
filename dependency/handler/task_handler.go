package handler

import (
	"encoding/json"
	"github.com/example/todo/service"
	"net/http"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct{ Title string }
		json.NewDecoder(r.Body).Decode(&body)

		h.svc.AddTask(r.Context(), body.Title)
		w.WriteHeader(http.StatusCreated)
		return
	}

	if r.Method == http.MethodGet {
		tasks, _ := h.svc.GetTasks(r.Context())
		json.NewEncoder(w).Encode(tasks)
		return
	}
}
