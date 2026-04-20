package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3" // ドライバのインポートを忘れずに
	"notification-aggregator/internal/handler"
	"notification-aggregator/internal/logging"
	"notification-aggregator/internal/notification"
	"notification-aggregator/internal/provider"

	"notification-sdk"
)

func main() {
	// 1. ログの設定 (段階5の準備を兼ねて slog を使用)
	baseHandler := slog.NewJSONHandler(os.Stdout, nil)
	// ★ここを自作の ContextHandler で包む
	logger := slog.New(&logging.ContextHandler{Handler: baseHandler})
	slog.SetDefault(logger)

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
	// --- 1. 設定ファイルの読み込み ---
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		logger.Error("failed to read config file", "error", err)
		os.Exit(1)
	}

	// JSONをパースするための構造体定義（一時的なものでもOK）
	var config struct {
		Providers []provider.ProviderConfig `json:"providers"`
	}

	if err := json.Unmarshal(configFile, &config); err != nil {
		logger.Error("failed to unmarshal config", "error", err)
		os.Exit(1)
	}

	// --- 2. Factory を使ってプロバイダーを動的に生成 ---
	var providers []sdk.Provider
	for _, pCfg := range config.Providers {
		p := provider.NewProvider(pCfg)
		if p != nil {
			providers = append(providers, p)
			logger.Info("provider loaded", "name", pCfg.Name, "type", pCfg.Type)
		} else {
			logger.Warn("unknown provider type", "type", pCfg.Type)
		}
	}

	// 段階 2: 複数のソースを登録
	// 下位レイヤー (Provider) から順に作成して注入する
	repo := notification.NewRepository(db)
	svc := notification.NewService(logger, repo, providers...)
	h := handler.NewNotificationHandler(svc)

	// 4. ルーティングとミドルウェアの適用 (Stage 4)
	mux := http.NewServeMux()

	// 1. LoggingMiddleware(logger) を実行すると、
	// 「loggerを内蔵した、RequestIDMiddlewareと同じ型の関数」が返ってくる。
	loggingWithLogger := handler.LoggingMiddleware(logger)

	// 2. それを使ってハンドラーを包む
	hWithLog := loggingWithLogger(h)

	// 3. さらに RequestIDMiddleware で包む
	// finalHandler := handler.RequestIDMiddleware(hWithLog)

	// --- これを1行で書くと ---
	// finalHandler := handler.RequestIDMiddleware(handler.LoggingMiddleware(logger)(h))

	// ハンドラーを RequestIDMiddleware でラップして登録
	mux.Handle("/notifications", handler.RequestIDMiddleware(hWithLog))

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
