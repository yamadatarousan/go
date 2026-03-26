package main

import "fmt"

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func main() {
	var todos []Todo

	newTodo := Todo{ID: 1, Title: "Goの学習", Done: false}
	todos = append(todos, newTodo)

	todos = append(todos, Todo{ID: 2, Title: "買い物に行く", Done: false})

	for _, t := range todos {
		fmt.Printf("ID: %d, タイトル: %s, 完了: %v\n", t.ID, t.Title, t.Done)
	}
}
