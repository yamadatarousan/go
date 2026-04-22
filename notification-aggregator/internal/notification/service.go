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

func (s *Service) AggregateAndSave(ctx context.Context) ([]sdk.Notification, []string, error) {
	reqID := contextutil.GetRequestID(ctx)
	l := s.logger.With("request_id", reqID)
	l.InfoContext(ctx, "aggregation and save transaction started")

	// 警告リストを初期化 (nullにならないようにする)
	warnings := []string{}

	// 1. 並列取得 (ここは多少のエラーは許容する)
	fetched, err := s.aggregateAllInternal(ctx, l)

	// ここでエラーがあっても、取得できたものがあれば処理を続ける
	if err != nil {
		// もしデータが取れていれば、エラーがあっても無視して保存に進む
		if len(fetched) == 0 {
			l.ErrorContext(ctx, "all providers failed", "error", err)
			return nil, []string{err.Error()}, fmt.Errorf("all providers failed: %w", err)
		}
		// 部分成功なら警告に追加して続行
		warnings = append(warnings, err.Error())
		l.WarnContext(ctx, "partial failure during aggregation", "error", err)
	}

	// ここを明示的に初期化しておく（nilを防ぐ）
	if fetched == nil {
		fetched = []sdk.Notification{}
	}

	// 2. DB保存 (Stage 3) と計測 (Stage 4）
	// 取得できた分（fetched）だけで保存処理を進める
	if len(fetched) > 0 {
		if err := s.repo.SaveAll(ctx, fetched); err != nil {
			l.ErrorContext(ctx, "failed to save to db", "error", err)
			return nil, warnings, fmt.Errorf("failed to save: %w", err)
		}
		l.InfoContext(ctx, "db save finished", "count", len(fetched))
	}

	return fetched, warnings, nil
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
