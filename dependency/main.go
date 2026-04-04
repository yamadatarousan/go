// main.go
package main

import (
	"github.com/example/todo/handler"
	"github.com/example/todo/infrastructure/memory" // memory 実装をインポート
	"github.com/example/todo/service"
	"net/http"
)

func main() {
	// 1. 永続化層の準備
	// database/sqlの実装ではなく、メモリ（Map）実装を生成
	repo := memory.NewTaskRepository()

	// 2. 依存関係の注入 (DI)
	// Service も Handler も、「repo が Map か DB か」を知らないまま受け取ります
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	// 3. サーバー起動
	http.Handle("/tasks", h)

	println("Server started at :8080 (Memory Mode)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
