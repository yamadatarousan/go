package notification

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Service struct {
	repo      *Repository
	providers []Provider
	logger    *slog.Logger
}

func NewService(logger *slog.Logger, repo *Repository, providers ...Provider) *Service {
	return &Service{
		repo:      repo,
		providers: providers,
		logger:    logger,
	}
}

func (s *Service) AggregateAll(ctx context.Context) []Notification {
	// 全体のタイムアウトを設定（例: 2秒）
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// 各 Provider の結果を収集するチャネル
	resultChan := make(chan []Notification, len(s.providers))
	var wg sync.WaitGroup

	for _, p := range s.providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()

			// 各 Provider ごとの処理
			s.logger.InfoContext(ctx, "fetching from provider", "name", p.Name())

			// Provider 内でタイムアウトが起きた際も ctx.Done() で検知できる
			ns, err := p.Fetch(ctx)
			if err != nil {
				s.logger.ErrorContext(ctx, "fetch error", "name", p.Name(), "error", err)
				return
			}
			resultChan <- ns
		}(p)
	}

	// すべての Provider が終わったらチャネルを閉じる
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var allNotifications []Notification
	// 結果を収集（タイムアウトしたものはここに含まれない）
	for ns := range resultChan {
		allNotifications = append(allNotifications, ns...)
	}

	return allNotifications
}

func (s *Service) AggregateAndSave(ctx context.Context) ([]Notification, error) {
	// 1. 段階 2 の並列取得を実行
	fetched := s.AggregateAll(ctx)

	// 2. 段階 3: DB に保存（重複排除は Repository 側で実施）
	if err := s.repo.SaveAll(ctx, fetched); err != nil {
		s.logger.ErrorContext(ctx, "failed to save notifications", "error", err)
		// 保存失敗しても、メモリ上のデータは返せるので続行
	}

	// 3. 最新の状態を DB から取得して返す
	return s.repo.FetchCached(ctx)
}
