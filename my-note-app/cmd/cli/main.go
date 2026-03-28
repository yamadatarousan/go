package main

import (
	cliAdapter "github.com/user/my-note-app/internal/adapter/cli"
	"github.com/user/my-note-app/internal/infrastructure/jsonfile"
	"github.com/user/my-note-app/internal/usecase"
	"os"
)

func main() {
	// 1. インフラ層（リポジトリ）の初期化
	// repo := memory.NewInMemoryNoteRepo()
	repo := jsonfile.NewJSONNoteRepo("notes.json")

	// 2. ユースケース層の初期化（リポジトリを注入）
	u := usecase.NewNoteUsecase(repo)

	// 3. アダプター層（CLIハンドラー）の初期化（ユースケースを注入）
	cliApp := cliAdapter.NewNoteCLI(u)

	// 4. コマンドライン引数を渡して実行
	// os.Args[1:] はプログラム名を除く引数のリスト
	cliApp.Run(os.Args[1:])
}
