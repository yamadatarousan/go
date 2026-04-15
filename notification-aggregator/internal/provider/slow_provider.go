package provider

import (
	"context"
	"notification-sdk"
	"time"
)

type SlowProvider struct{}

func (s *SlowProvider) Name() string { return "slow_service" }

func (s *SlowProvider) Fetch(ctx context.Context) ([]sdk.Notification, error) {
	select {
	case <-time.After(5 * time.Second): // 意図的な遅延
		return []sdk.Notification{{ID: "slow-1", Title: "Slow News"}}, nil
	case <-ctx.Done(): // Service 側のタイムアウトでここが反応する
		return nil, ctx.Err()
	}
}
