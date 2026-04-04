package persistence

import (
	"context"
	"database/sql"
	"github.com/example/todo/domain"
)

type sqlTaskRepository struct {
	db *sql.DB
}

func NewSQLTaskRepository(db *sql.DB) domain.TaskRepository {
	return &sqlTaskRepository{db: db}
}

func (r *sqlTaskRepository) Create(ctx context.Context, t *domain.Task) error {
	return r.db.QueryRowContext(ctx, "INSERT INTO tasks(title) VALUES($1) RETURNING id", t.Title).Scan(&t.ID)
}

func (r *sqlTaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	rows, _ := r.db.QueryContext(ctx, "SELECT id, title FROM Tasks")
	defer rows.Close()
	var list []*domain.Task
	for rows.Next() {
		var t domain.Task
		rows.Scan(&t.ID, &t.Title)
		list = append(list, &t)
	}
	return list, nil
}
