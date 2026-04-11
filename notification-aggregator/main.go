package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3" // ドライバのインポートを忘れずに
	"notification-aggregator/internal/handler"
	"notification-aggregator/internal/notification"
	"notification-aggregator/internal/provider"
)

func main() {
	// 1. ログの設定 (段階5の準備を兼ねて slog を使用)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. DBの初期化と接続プールの設定 (ここが Stage 3 のキモ)
	db, err := sql.Open("sqlite3", "./notifications.db")
	if err != nil {
		logger.Error("failed to open db", "error", err)
		os.Exit(1)
	}

	// [FIX] 上級編の運用論点に基づいた設定
	db.SetMaxOpenConns(20)                  // 最大同時接続数
	db.SetMaxIdleConns(20)                  // アイドル状態の接続保持数
	db.SetConnMaxIdleTime(2 * time.Minute)  // アイドル接続の寿命
	db.SetConnMaxLifetime(30 * time.Minute) // 接続自体の最大寿命

	// テーブルの準備
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		source TEXT,
		title TEXT,
		content TEXT,
		created_at DATETIME
	)`)

	// 2. 依存関係の構築
	// 段階 2: 複数のソースを登録
	// 下位レイヤー (Provider) から順に作成して注入する
	repo := notification.NewRepository(db)
	p1 := &provider.MockProvider{}
	p2 := &provider.SlowProvider{} // 遅いソース
	svc := notification.NewService(logger, repo, p1, p2)
	h := handler.NewNotificationHandler(svc)

	// 4. ルーティングとミドルウェアの適用 (Stage 4)
	mux := http.NewServeMux()
	// ハンドラーを RequestIDMiddleware でラップして登録
	mux.Handle("/notifications", handler.RequestIDMiddleware(h))

	// 4. サーバーの起動
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	logger.Info("server starting", "addr", srv.Addr)

	// 補足: 上級編らしく Graceful Shutdown の準備
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", "error", err)
			os.Exit(1)
		}
	}()

	// 終了信号を待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("server shutting down...")

	// 終了処理のためのコンテキスト (5秒以内にクリーンアップ)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server exited gracefully")
}
