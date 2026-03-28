package http

import (
	"encoding/json"
	"github.com/user/my-note-app/internal/usecase"
	"net/http"
)

// NoteHTTPHandler はHTTPリクエストを処理します。
type NoteHTTPHandler struct {
	usecase *usecase.NoteUsecase
}

// NewNoteHTTPHandler はNoteHTTPHandlerを初期化します。
func NewNoteHTTPHandler(u *usecase.NoteUsecase) *NoteHTTPHandler {
	return &NoteHTTPHandler{usecase: u}
}

// ServeHTTP はリクエストを受け取り、ユースケースを呼び出します。
func (h *NoteHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		notes, err := h.usecase.GetNotes()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		json.NewEncoder(w).Encode(notes)

	case http.MethodPost:
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "無効なリクエストです"})
		}

		note, err := h.usecase.CreateNote(req.Content)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(note)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
