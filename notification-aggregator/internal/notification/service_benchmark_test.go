package notification

import (
	"context"
	"io" // 追加: io.Discard を使うため
	"log/slog"
	"testing"
)

// 同一パッケージ内なので小文字の Notification 型や非公開メソッドにアクセス可能

type internalMockProvider struct {
	name string
	data []Notification
}

func (m *internalMockProvider) Fetch(ctx context.Context) ([]Notification, error) {
	return m.data, nil
}
func (m *internalMockProvider) Name() string { return m.name }

func BenchmarkAggregateAllInternal(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	dummyData := make([]Notification, 100)
	for i := range dummyData {
		dummyData[i] = Notification{ID: "bench", Title: "Optimizing"}
	}

	providers := []Provider{
		&internalMockProvider{name: "p1", data: dummyData},
		&internalMockProvider{name: "p2", data: dummyData},
		&internalMockProvider{name: "p3", data: dummyData},
	}

	svc := NewService(logger, nil, providers...)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.aggregateAllInternal(ctx, logger)
	}
}
