package memory

import (
	"github.com/user/my-note-app/internal/domain"
	"sync"
)

// InMemoryNoteRepo は NoteRepository をオンメモリで実装します。
type InMemoryNoteRepo struct {
	mu     sync.Mutex
	notes  []domain.Note
	nextID int
}

// NewInMemoryNoteRepo はInMemoryNoteRepoを初期化します。
func NewInMemoryNoteRepo() *InMemoryNoteRepo {
	return &InMemoryNoteRepo{notes: make([]domain.Note, 0), nextID: 1}
}

// Save はメモをメモリに保存します。
func (r *InMemoryNoteRepo) Save(note domain.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	note.ID = r.nextID
	r.nextID++
	r.notes = append(r.notes, note)
	return nil
}

// FindAll はメモリ上のすべてのメモを取得します。
func (r *InMemoryNoteRepo) FindAll() ([]domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.notes, nil
}
