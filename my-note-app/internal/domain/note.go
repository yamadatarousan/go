package domain

type Note struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

// NoteRepository はデータの保存・取得を行うインターフェース（ポート）
// 具体的な実装（DB、メモリなど）はインフラ層で行います。
type NoteRepository interface {
	Save(note Note) error
	FindAll() ([]Note, error)
}
