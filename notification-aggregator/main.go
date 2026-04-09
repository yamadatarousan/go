package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"notification-aggregator/internal/handler"
	"notification-aggregator/internal/notification"
	"notification-aggregator/internal/provider"
)

func main() {
	// 1. ログの設定 (段階5の準備を兼ねて slog を使用)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. 依存関係の構築 (段階1: 単一ソース)
	// 下位レイヤー (Provider) から順に作成して注入する
	mockP := &provider.MockProvider{}
	svc := notification.NewService(mockP)
	h := handler.NewNotificationHandler(svc)

	// 3. ルーティングの設定
	mux := http.NewServeMux()
	mux.Handle("/notifications", h)

	// 4. サーバーの起動
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	logger.Info("server starting", "addr", srv.Addr)

	// 補足: 上級編らしく Graceful Shutdown の準備
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", "error", err)
			os.Exit(1)
		}
	}()

	// 終了信号を待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Info("server shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
