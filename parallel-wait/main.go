package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	results := make(chan string, 2)

	// 2つの処理を並列実行
	for i := 1; i <= 2; i++ {
		wg.Add(1) // 「これから1人作業に入るよ」と登録
		go func(id int) {
			defer wg.Done() // 作業が終わったら「1人抜けたよ」と報告
			time.Sleep(time.Duration(id) * time.Second)
			results <- fmt.Sprintf("API-%d の結果", id)
		}(i)
	}

	// 別ゴルーチンで「全員終わったらチャネルを閉じる」という監視役を置く
	go func() {
		wg.Wait()      // 全員の Done() が終わるまでここでブロック
		close(results) // 全員終わったのでチャネルを閉める
	}()

	// main側で「届いた順」に逐次処理する
	fmt.Println("全結果を待機中...")
	for res := range results { // closeされるまで、届いた結果を一つずつ取り出す
		fmt.Println("✅ 処理開始:", res)
	}
	fmt.Println("すべての逐次処理が完了しました。")
}
