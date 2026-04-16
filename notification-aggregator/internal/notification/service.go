package notification

import (
	"context"
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
	start := time.Now()
	reqID := contextutil.GetRequestID(ctx)

	// Stage 4: このメソッド内でのログに Request ID を固定で付与する
	l := s.logger.With("request_id", reqID)
	l.InfoContext(ctx, "aggregation and save transaction started")

	/// 1. 並列取得 (Stage 2)
	// aggregateAllInternal にも ID を引き継ぐため、一時的に logger を差し替えるのではなく
	// 内部処理で l (子ロガー) を使用するよう配慮します。
	fetched := s.aggregateAllInternal(ctx, l)

	// 2. DB保存 (Stage 3) と計測 (Stage 4）
	saveStart := time.Now()
	if err := s.repo.SaveAll(ctx, fetched); err != nil {
		l.ErrorContext(ctx, "failed to save notifications", "error", err)
	}
	l.InfoContext(ctx, "db save finished", "duration_ms", time.Since(saveStart).Milliseconds())

	// 3. 最新リストの取得
	res, err := s.repo.FetchCached(ctx)

	l.InfoContext(ctx, "aggregation and save completed",
		"total_duration_ms", time.Since(start).Milliseconds(),
		"items_fetched", len(fetched),
		"items_total", len(res),
	)

	return res, err
}

func (s *Service) aggregateAllInternal(ctx context.Context, l *slog.Logger) []sdk.Notification {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resultChan := make(chan []sdk.Notification, len(s.providers))
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
				return
			}
			resultChan <- ns
		}(p)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var all []sdk.Notification
	for ns := range resultChan {
		all = append(all, ns...)
	}
	return all
}
