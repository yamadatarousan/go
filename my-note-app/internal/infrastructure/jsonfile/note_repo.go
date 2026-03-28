package jsonfile

import (
	"encoding/json"
	"github.com/user/my-note-app/internal/domain"
	"os"
	"sync"
)

type JSONNoteRepo struct {
	mu       sync.Mutex
	filePath string
}

func NewJSONNoteRepo(path string) *JSONNoteRepo {
	return &JSONNoteRepo{filePath: path}
}

func (r *JSONNoteRepo) Save(note domain.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	notes, _ := r.FindAll()
	note.ID = len(notes) + 1
	notes = append(notes, note)

	data, _ := json.Marshal(notes)
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *JSONNoteRepo) FindAll() ([]domain.Note, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return []domain.Note{}, nil
	}
	var notes []domain.Note
	json.Unmarshal(data, &notes)
	return notes, nil
}
