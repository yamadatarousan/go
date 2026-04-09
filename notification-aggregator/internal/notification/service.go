package notification

import (
	"context"
	"fmt"
)

type Service struct {
	provider Provider
}

func NewService(p Provider) *Service {
	return &Service{provider: p}
}

func (s *Service) GetNotifications(ctx context.Context) ([]Notification, error) {
	// 段階 1 では単一のソースから取得します
	items, err := s.provider.Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("provider %s failure: %w", s.provider.Name(), err)
	}
	return items, nil
}
