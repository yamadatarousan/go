package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Task 実行するタスクの定義
type Task struct {
	Name string
	Args []string
}

func main() {
	tasks := []Task{
		{Name: "Standard Test", Args: []string{"test", "./..."}},
		{Name: "Race Detector", Args: []string{"test", "-race", "./..."}},
		// Fuzzテストはデフォルトで無限に回るため、10秒間で切り上げる設定にしています
		{Name: "Fuzz Testing", Args: []string{"test", "-fuzz=Fuzz", "-run=^$", "-fuzztime=10s"}},
		{Name: "Vulnerability Check", Args: []string{"run", "golang.org/x/vuln/cmd/govulncheck@latest", "./..."}},
	}

	fmt.Println("🚀 開発前フルチェックを開始します...")
	start := time.Now()

	for _, t := range tasks {
		fmt.Printf("\n--- 🏃 Running: %s ---\n", t.Name)

		// goコマンドを叩く（govulncheckは go run ... で実行する構成にしています）
		cmd := exec.Command("go", t.Args...)

		// 出力をそのまま現在のターミナルに流す
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("\n❌ 【失敗】 %s でエラーが発生しました。\n", t.Name)
			os.Exit(1)
		}
		fmt.Printf("✅ 【完了】 %s\n", t.Name)
	}
	fmt.Printf("\n✨ すべての検査をパスしました！ (所要時間: %s)\n", time.Since(start).Round(time.Second))
}
