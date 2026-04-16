package notification_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"notification-aggregator/internal/notification"
	"notification-sdk"
)

// MockProvider はテスト用のダミープロバイダーです
type MockProvider struct {
	name string
	data []sdk.Notification
	err  error
}

func (m *MockProvider) Fetch(ctx context.Context) ([]sdk.Notification, error) {
	return m.data, m.err
}
func (m *MockProvider) Name() string { return m.name }

// Repository Mock: DBを使わずプロパティで制御する
type MockRepository struct {
	saveErr error
	saved   []sdk.Notification
}

func (m *MockRepository) SaveAll(ctx context.Context, items []sdk.Notification) error {
	m.saved = items  // 渡された引数をプロパティに保存（記録）
	return m.saveErr // プロパティにセットした値を返す（制御）
}

func (m *MockRepository) FetchCached(ctx context.Context) ([]sdk.Notification, error) {
	return m.saved, nil
}

// --- テスト本体 ---

func TestService_AggregateAndSave(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	tests := []struct {
		name      string
		providers []sdk.Provider
		saveErr   error // Repositoryを失敗させたい時にセットする
		wantCount int
	}{
		{
			name: "正常系: 2つのソースから取得",
			providers: []sdk.Provider{
				&MockProvider{name: "p1", data: []sdk.Notification{{ID: "1", Title: "T1"}}},
				&MockProvider{name: "p2", data: []sdk.Notification{{ID: "2", Title: "T2"}}},
			},
			wantCount: 2,
		},
		{
			name: "異常系: DB保存が失敗しても取得データは返る",
			providers: []sdk.Provider{
				&MockProvider{name: "p1", data: []sdk.Notification{{ID: "3", Title: "T3"}}},
			},
			saveErr:   context.DeadlineExceeded, // ここでRepositoryの挙動を「仕込む」
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RepositoryをMockで作る
			repo := &MockRepository{saveErr: tt.saveErr}

			svc := notification.NewService(logger, repo, tt.providers...)

			got, err := svc.AggregateAndSave(context.Background())
			if err != nil {
				t.Fatalf("AggregateAndSave failed: %v", err)
			}

			if len(got) != tt.wantCount {
				t.Errorf("got %d items, want %d", len(got), tt.wantCount)
			}
		})
	}
}
