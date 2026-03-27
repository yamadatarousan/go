package repository

type Todo struct {
	ID   int
	Task string
}

// Repository インターフェース
// これを満たせば、メモリだろうがファイルだろうがOK
type Repository interface {
	Save(todo Todo) error
	FindAll() ([]Todo, error)
}
