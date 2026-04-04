package service

import (
	"context"
	"github.com/example/todo/domain"
)

type TaskService struct {
	repo domain.TaskRepository
}

func NewTaskService(repo domain.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) AddTask(ctx context.Context, title string) error {
	task := &domain.Task{Title: title}
	return s.repo.Create(ctx, task)
}

func (s *TaskService) GetTasks(ctx context.Context) ([]*domain.Task, error) {
	return s.repo.List(ctx)
}
