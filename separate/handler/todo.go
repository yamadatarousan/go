package handler

import (
	"encoding/json"
	"my-todo-app/service" // 自分のプロジェクトパスに合わせてください
	"net/http"
)

type TodoHandler struct {
	Service *service.TodoService
}

func NewTodoHandler(s *service.TodoService) *TodoHandler {
	return &TodoHandler{Service: s}
}

func (h *TodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 1. Serviceを呼び出してデータを取得
	todos := h.Service.ListTodos()

	// 2. JSONに変換してレスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}
