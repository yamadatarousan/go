package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Item struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
}

const fileName = "data.json"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使い方:")
		fmt.Println("  add [タスク内容] : タスクを追加して保存")
		fmt.Println("  list            : 保存されたタスクを一覧表示")
		return
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("エラー: タスク内容を入力してください。")
			return
		}
		addTask(os.Args[2])
	case "list":
		listTasks()
	default:
		fmt.Printf("未知のコマンド: %s\n", subcommand)
	}
}

func loadData() []Item {
	var items []Item
	file, err := os.ReadFile(fileName)
	if err != nil {
		return []Item{}
	}
	json.Unmarshal(file, &items)
	return items
}

func addTask(task string) {
	items := loadData()

	newItem := Item{
		ID:   len(items) + 1,
		Task: task,
	}
	items = append(items, newItem)

	data, _ := json.MarshalIndent(items, "", "  ")
	err := os.WriteFile(fileName, data, 0644)
	if err != nil {
		fmt.Println("保存に失敗しました:", err)
		return
	}
	fmt.Printf("追加完了: %s\n", task)
}

func listTasks() {
	items := loadData()
	if len(items) == 0 {
		fmt.Println("タスクはありません。")
		return
	}

	fmt.Println("--- タスク一覧 ---")
	for _, item := range items {
		fmt.Printf("[%d] %s\n", item.ID, item.Task)
	}
}
