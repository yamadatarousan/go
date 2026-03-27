package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"todo-app/repository"
)

type TodoService struct {
	repo repository.Repository
}

func (s *TodoService) ListAll() {
	todos, _ := s.repo.FindAll()
	fmt.Println("\n--- TODO LIST ---")
	if len(todos) == 0 {
		fmt.Println("(タスクはありません)")
	}
	for i, t := range todos {
		fmt.Printf("%d: %s\n", i+1, t.Task)
	}
	fmt.Println("-----------------")
}

func (s *TodoService) Add(task string) {
	t := repository.Todo{Task: task}
	s.repo.Save(t)
	fmt.Println("✓ タスクを追加しました")
}

func main() {
	// ★ ここを切り替えるだけで、保存先がメモリからファイルに変わる
	// repo := repository.NewMemoryRepo()
	repo := repository.NewFileRepo("todos.json")
	service := TodoService{repo: repo}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n1:一覧  2:追加  3:終了 > ")
		scanner.Scan()
		input := scanner.Text()

		switch input {
		case "1":
			service.ListAll()
		case "2":
			fmt.Print("タスクを入力してください: ")
			scanner.Scan()
			task := scanner.Text()
			if strings.TrimSpace(task) != "" {
				service.Add(task)
			}
		case "3":
			fmt.Println("バイバイ！")
			return
		default:
			fmt.Println("1, 2, 3 のどれかを入力してね")
		}
	}
}
