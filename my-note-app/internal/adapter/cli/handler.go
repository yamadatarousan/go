package cli

import (
	"fmt"
	"github.com/user/my-note-app/internal/usecase"
)

// NoteCLI はコマンドライン引数を処理します。
type NoteCLI struct {
	usecase *usecase.NoteUsecase
}

// NewNoteCLI はNoteCLIを初期化します。
func NewNoteCLI(u *usecase.NoteUsecase) *NoteCLI {
	return &NoteCLI{usecase: u}
}

// Run は引数を解析し、対応するユースケースを呼び出します。
func (cli *NoteCLI) Run(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cli [list | add <content>]")
		return
	}

	command := args[0]
	switch command {
	case "list":
		notes, err := cli.usecase.GetNotes()
		if err != nil {
			fmt.Println("Error:", err)
		}
		for _, n := range notes {
			fmt.Printf("[%d] %s\n", n.ID, n.Content)
		}
	case "add":
		if len(args) < 2 {
			fmt.Println("Error: メモの内容を指定してください")
			return
		}
		content := args[1]
		_, err := cli.usecase.CreateNote(content)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("メモを追加しました。")
	default:
		fmt.Println("不明なコマンドです。")
	}

}
