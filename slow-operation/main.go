package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func slowOperation(ctx context.Context) error {
	fmt.Println("  [処理開始] DBまたは外部APIへリクエストを送信...")

	// 5秒かかる処理をシミュレート
	// select文を使って、処理の完了かコンテキストのキャンセルのどちらか早い方を待ち受けます
	select {
	case <-time.After(5 * time.Second):
		// タイムアウトせずに処理が完了した場合
		fmt.Println("  [処理完了] データを受信しました")
		return nil
	case <-ctx.Done():
		// タイムアウト（または親コンテキストからのキャンセル）が発生した場合
		// %w を使って ctx.Err() をラップし、どこでどんなエラーが起きたかの文脈を残す
		return fmt.Errorf("slowOperation aborted during DB/API call: %w", ctx.Err())
	}
}

func main() {
	fmt.Println("=== メイン処理開始 ===")

	// 2秒でタイムアウトするコンテキストを作成
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	err := slowOperation(ctx)

	if err != nil {
		fmt.Printf("エラー発生: %v\n", err)

		// エラーがタイムアウトによるものか（元のエラーが context.DeadlineExceeded か）を判定
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("-> 診断: 設定された時間内に処理が完了しなかったため、安全に打ち切りました。")
		}
		return
	}

	fmt.Println("=== メイン処理正常終了 ===")
}
