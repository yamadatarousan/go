package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// 結果を格納する構造体
type Result struct {
	URL    string
	Status int
	Err    error
}

func main() {
	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://httpbin.org/delay/5", // 意図的に遅いURL
	}

	// 1. タイムアウト付きのContextを作成（3秒で打ち切り）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results := make(chan Result, len(urls))

	for _, url := range urls {
		// goroutineの起動
		go func(u string) {
			results <- fetch(ctx, u)
		}(url)
	}

	// 2. 結果の集約
	for i := 0; i < len(urls); i++ {
		select {
		case res := <-results:
			if res.Err != nil {
				fmt.Printf("[Error] %s: %v\n", res.URL, res.Err)
			} else {
				fmt.Printf("[Success] %s: %d\n", res.URL, res.Status)
			}
		case <-ctx.Done():
			// 全体のタイムアウトが発生した場合
			fmt.Println("Overall timeout reached!")
			return
		}
	}
}

func fetch(ctx context.Context, url string) Result {
	// リクエスト作成
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Result{URL: url, Err: err}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: url, Err: err}
	}
	defer resp.Body.Close()

	return Result{URL: url, Status: resp.StatusCode, Err: nil}
}
