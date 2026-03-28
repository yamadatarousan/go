package usecase

import (
	"fmt"
	"github.com/user/my-note-app/internal/domain"
)

// NoteUsecase はメモに関するビジネスロジックを実行します。
type NoteUsecase struct {
	repo domain.NoteRepository
}

// NewNoteUsecase はNoteUsecaseを初期化します。
func NewNoteUsecase(repo domain.NoteRepository) *NoteUsecase {
	return &NoteUsecase{repo: repo}
}

// CreateNote は新しいメモを作成するビジネスロジックです。
func (u *NoteUsecase) CreateNote(content string) (domain.Note, error) {
	if content == "" {
		return domain.Note{}, fmt.Errorf("メモの内容は空にできません")
	}
	note := domain.Note{Content: content}
	err := u.repo.Save(note)
	return note, err
}

// GetNotes はメモの一覧を取得するビジネスロジックです。
func (u *NoteUsecase) GetNotes() ([]domain.Note, error) {
	return u.repo.FindAll()
}
