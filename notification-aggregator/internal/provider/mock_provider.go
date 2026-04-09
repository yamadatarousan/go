package provider

import (
	"context"
	"notification-aggregator/internal/notification"
	"time"
)

type MockProvider struct{}

func (m *MockProvider) Name() string { return "mock_service" }

func (m *MockProvider) Fetch(ctx context.Context) ([]notification.Notification, error) {
	// 実際にはここで http.Get などを行いますが、まずは段階 1 の動作確認用に固定値を返します
	return []notification.Notification{
		{ID: "1", Source: "mock", Title: "Hello Go", Content: "Level up!", CreatedAt: time.Now()},
	}, nil
}
