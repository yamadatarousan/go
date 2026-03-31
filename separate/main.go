package main

import (
	"my-todo-app/handler"
	"my-todo-app/service"
	"net/http"
)

func main() {
	todoService := service.NewTodoService()
	todohandler := handler.NewTodoHandler(todoService)

	http.Handle("/todo", todohandler)

	println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
