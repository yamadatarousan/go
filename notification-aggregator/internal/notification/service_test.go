package notification_test

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"notification-aggregator/internal/notification"
)

// MockProvider はテスト用のダミープロバイダーです
type MockProvider struct {
	name string
	data []notification.Notification
	err  error
}

func (m *MockProvider) Fetch(ctx context.Context) ([]notification.Notification, error) {
	return m.data, m.err
}
func (m *MockProvider) Name() string { return m.name }

func TestService_AggregateAndSave(t *testing.T) {
	// 1. テスト用 DB の準備 (メモリ上の SQLite)
	db, _ := sql.Open("sqlite3", ":memory:")
	_, _ = db.Exec(`CREATE TABLE notifications (id TEXT PRIMARY KEY, source TEXT, title TEXT, content TEXT, created_at DATETIME)`)

	repo := notification.NewRepository(db)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. テストケースの定義 (Table Driven Test)
	tests := []struct {
		name      string
		providers []notification.Provider
		wantCount int
	}{
		{
			name: "正常系: 2つのソースから取得",
			providers: []notification.Provider{
				&MockProvider{name: "p1", data: []notification.Notification{{ID: "1", Title: "T1"}}},
				&MockProvider{name: "p2", data: []notification.Notification{{ID: "2", Title: "T2"}}},
			},
			wantCount: 2,
		},
		{
			name: "異常系: 片方がエラーでももう片方は取得できる",
			providers: []notification.Provider{
				&MockProvider{name: "p1", data: []notification.Notification{{ID: "3", Title: "T3"}}},
				&MockProvider{name: "err_p", err: sql.ErrConnDone},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := notification.NewService(logger, repo, tt.providers...)

			got, err := svc.AggregateAndSave(context.Background())
			if err != nil {
				t.Fatalf("AggregateAndSave failed: %v", err)
			}

			// 最新の取得数を確認
			if len(got) < tt.wantCount {
				t.Errorf("got %d items, want at least %d", len(got), tt.wantCount)
			}
		})
	}
}
