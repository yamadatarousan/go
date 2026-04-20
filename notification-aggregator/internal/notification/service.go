package notification

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"notification-aggregator/internal/contextutil"
	"notification-sdk"
)

// 1. Repository が備えるべき機能を定義
type Repository interface {
	SaveAll(ctx context.Context, items []sdk.Notification) error
	FetchCached(ctx context.Context) ([]sdk.Notification, error)
}

type Service struct {
	repo      Repository
	providers []sdk.Provider
	logger    *slog.Logger
}

func NewService(logger *slog.Logger, repo Repository, providers ...sdk.Provider) *Service {
	return &Service{
		repo:      repo,
		providers: providers,
		logger:    logger,
	}
}

func (s *Service) AggregateAndSave(ctx context.Context) ([]sdk.Notification, error) {
	reqID := contextutil.GetRequestID(ctx)

	// Stage 4: このメソッド内でのログに Request ID を固定で付与する
	l := s.logger.With("request_id", reqID)
	l.InfoContext(ctx, "aggregation and save transaction started")

	/// 1. 並列取得 (Stage 2)
	// aggregateAllInternal にも ID を引き継ぐため、一時的に logger を差し替えるのではなく
	// 内部処理で l (子ロガー) を使用するよう配慮します。
	fetched, err := s.aggregateAllInternal(ctx, l)

	// まず、リクエスト自体が死んでいるかを確認する
	if ctx.Err() != nil {
		l.ErrorContext(ctx, "request context finished", "error", ctx.Err())
		return nil, ctx.Err()
	}

	if err != nil {
		// もしデータが取れていれば、エラーがあっても無視して保存に進む
		if len(fetched) == 0 {
			l.ErrorContext(ctx, "all providers failed", "error", err)
			return nil, fmt.Errorf("all providers failed: %w", err)
		}

		// 全滅じゃなければ、エラーをログに出しつつ処理を続行する、といった判断ができる
		l.WarnContext(ctx, "partial failure during aggregation", "error", err)
	}

	// ここを明示的に初期化しておく（nilを防ぐ）
	if fetched == nil {
		fetched = []sdk.Notification{}
	}

	// 2. DB保存 (Stage 3) と計測 (Stage 4）
	// 取得できた分（fetched）だけで保存処理を進める
	saveStart := time.Now()
	if len(fetched) > 0 {
		if err := s.repo.SaveAll(ctx, fetched); err != nil {
			l.ErrorContext(ctx, "failed to save to db", "error", err)
			return nil, fmt.Errorf("failed to save: %w", err)
		}
	}
	l.InfoContext(ctx, "db save finished", "duration_ms", time.Since(saveStart).Milliseconds())

	// 【診断ログを追加】
	l.Info("debug_check",
		"fetched_is_nil", fetched == nil,
		"fetched_len", len(fetched),
		"fetched_ptr", fmt.Sprintf("%p", fetched))

	// 3. 【修正】DBから再取得せず、手元の fetched を直接返す
	//    これで「DBから取れない」というバグ要因を一つ消せます
	return fetched, nil
}

func (s *Service) aggregateAllInternal(ctx context.Context, l *slog.Logger) ([]sdk.Notification, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// 通知結果とエラー、それぞれを回収するチャネルを用意
	// バッファサイズを provider 数にしておくことで、送信側のブロックを防ぐ
	numProviders := len(s.providers)
	resultChan := make(chan []sdk.Notification, numProviders)
	errChan := make(chan error, numProviders)

	var wg sync.WaitGroup

	for _, p := range s.providers {
		wg.Add(1)
		go func(p sdk.Provider) {
			defer wg.Done()

			// 各 Provider のログにも ID が付与される
			l.InfoContext(ctx, "fetching from provider", "provider", p.Name())

			ns, err := p.Fetch(ctx)
			if err != nil {
				l.ErrorContext(ctx, "provider fetch error", "provider", p.Name(), "error", err)
				// %w でラップしてチャネルへ
				errChan <- fmt.Errorf("provider %s: %w", p.Name(), err)
				return
			}
			resultChan <- ns
		}(p)
	}

	// 監視役：全 Worker が終わったらチャネルを閉じる
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	var all []sdk.Notification
	for ns := range resultChan {
		all = append(all, ns...)
	}

	// エラーを回収
	var allErrors []error
	for err := range errChan {
		allErrors = append(allErrors, err)
	}

	// errors.Join で複数のエラーを一つにまとめる
	// allErrors が空（len=0）なら、nil が返るので安全
	return all, errors.Join(allErrors...)
}
