package service

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type TodoService struct{}

func NewTodoService() *TodoService {
	return &TodoService{}
}

func (s *TodoService) ListTodos() []Todo {
	return []Todo{
		{ID: 1, Title: "Handlerを分ける"},
		{ID: 2, Title: "Serviceを分ける"},
	}
}
