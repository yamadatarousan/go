package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 実際のAPIを呼び出す関数
func fetchAPI(ctx context.Context, name, url string, resultChan chan<- string) {
	// 1. Contextを紐付けたHTTPリクエストを作成
	// ※このctxがキャンセルされると、実行中の client.Do(req) がエラーですぐに中断されます
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		resultChan <- fmt.Sprintf("[%s] リクエスト作成エラー: %v", name, err)
		return
	}

	start := time.Now()
	client := &http.Client{}

	// 2. HTTPリクエストを実行
	resp, err := client.Do(req)
	if err != nil {
		// main側で cancel() が呼ばれて通信が中断された場合、ここに到達します
		if ctx.Err() == context.Canceled {
			fmt.Printf("--- ⚠ [%s] の通信はキャンセルされました（HTTPリクエスト中断）\n", name)
			return
		}
		// その他のエラー
		resultChan <- fmt.Sprintf("[%s] 通信エラー: %v", name, err)
		return
	}
	defer resp.Body.Close() // 確実なリソース解放

	// 3. レスポンスボディの読み込み
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resultChan <- fmt.Sprintf("[%s] 読み込みエラー: %v", name, err)
		return
	}

	duration := time.Since(start)

	// 結果を見やすくするため、レスポンスの先頭50文字だけ切り出す
	preview := string(body)
	if len(preview) > 50 {
		preview = preview[:50] + "…"
	}

	// 4. チャネルに結果を送信
	resultChan <- fmt.Sprintf("[%s] 成功 (処理時間: %v)\nレスポンス: %s", name, duration, preview)
}

func main() {
	// キャンセル可能なコンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// バッファサイズ2のチャネルを作成（ゴルーチンリーク防止）
	resultChan := make(chan string, 2)

	// 今回テスト利用する2つの公開API
	url1 := "https://jsonplaceholder.typicode.com/todos/1" // JSONPlaceholder
	url2 := "https://api.github.com/users/octocat"         // GitHub API

	fmt.Println("2つのAPIへのリクエストを同時に開始します...")

	// それぞれ別のゴルーチンでAPI呼び出しを開始
	go fetchAPI(ctx, "JSONPlaceholder", url1, resultChan)
	go fetchAPI(ctx, "GITHUB API", url2, resultChan)

	// 最初にチャネルに届いた結果を受信（ここでブロックして待機）
	// ここに2つのチャネルの結果が入ることはあるがmain.goは上から順の逐次実行のため最初に返ってきた結果しか実行されない
	fastestResult := <-resultChan
	fmt.Printf("\n✅ 最初に返ってきた結果を採用:\n%s\n\n", fastestResult)

	// --- 最も重要なポイント ---
	// 最初の結果を受け取った直後にキャンセルを実行し、
	// 遅い側のゴルーチンのHTTP通信を強制的に中断させる
	cancel()

	// ※ 遅い側のHTTP通信がキャンセルされ、ログが出力されるのを確認するための待機
	time.Sleep(1 * time.Second)
	fmt.Println("すべての処理が終了しました。")
}
