package repository

type MemoryRepo struct {
	todos []Todo
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{todos: []Todo{}}
}

func (m *MemoryRepo) Save(todo Todo) error {
	m.todos = append(m.todos, todo)
	return nil
}

func (m *MemoryRepo) FindAll() ([]Todo, error) {
	return m.todos, nil
}
