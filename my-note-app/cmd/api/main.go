package main

import (
	"fmt"
	httpAdapter "github.com/user/my-note-app/internal/adapter/http"
	"github.com/user/my-note-app/internal/infrastructure/memory"
	"github.com/user/my-note-app/internal/usecase"
	"log"
	"net/http"
)

func main() {
	// 1. インフラ層（リポジトリ）の初期化
	repo := memory.NewInMemoryNoteRepo()

	// 2. ユースケース層の初期化（リポジトリを注入）
	u := usecase.NewNoteUsecase(repo)

	// 3. アダプター層（CLIハンドラー）の初期化（ユースケースを注入）
	handler := httpAdapter.NewNoteHTTPHandler(u)

	// 4. ルーティングの設定とサーバー起動
	http.Handle("/notes", handler)

	port := ":8080"
	fmt.Printf("HTTPサーバーを起動しました (http://localhost%s/notes) ...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
