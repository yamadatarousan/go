package memory

import (
	"context"
	"github.com/example/todo/domain"
	"sync"
)

type taskRepository struct {
	mu     sync.RWMutex
	tasks  map[int64]*domain.Task
	nextID int64
}

func NewTaskRepository() domain.TaskRepository {
	return &taskRepository{tasks: make(map[int64]*domain.Task)}
}

func (r *taskRepository) Create(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	task.ID = r.nextID
	r.tasks[task.ID] = task
	return nil
}

func (r *taskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*domain.Task
	for _, t := range r.tasks {
		list = append(list, t)
	}
	return list, nil
}
