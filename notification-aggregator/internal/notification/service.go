package notification

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Service struct {
	providers []Provider
	logger    *slog.Logger
}

func NewService(logger *slog.Logger, providers ...Provider) *Service {
	return &Service{
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
