package domain

import "context"

type Task struct {
	ID    int64
	Title string
}

// TaskRepository はデータ永続化の境界を定義する
type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	List(ctx context.Context) ([]*Task, error)
}


