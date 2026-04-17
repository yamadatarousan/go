package notification_test

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"testing"
	"time"

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
		wantCount int   // 正常系なら期待件数、異常系なら 0
	}{
		{
			name: "正常系: 複数Providerから合計3件取得して保存",
			providers: []sdk.Provider{
				&MockProvider{name: "p1", data: []sdk.Notification{{ID: "1"}, {ID: "2"}}},
				&MockProvider{name: "p2", data: []sdk.Notification{{ID: "3"}}},
			},
			saveErr:   nil,
			wantCount: 3,
		},
		{
			name: "異常系: DB保存失敗時は 0 件を返す",
			providers: []sdk.Provider{
				&MockProvider{name: "p1", data: []sdk.Notification{{ID: "3", Title: "T3"}}},
			},
			saveErr:   context.DeadlineExceeded, // ここでRepositoryの挙動を「仕込む」
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RepositoryをMockで作る
			repo := &MockRepository{saveErr: tt.saveErr}

			svc := notification.NewService(logger, repo, tt.providers...)

			got, err := svc.AggregateAndSave(context.Background())

			// 検証1: エラーの有無が期待通りか
			if (err != nil) != (tt.saveErr != nil) {
				t.Fatalf("[%s] エラーの有無が一致しません: 期待=%v, 実際=%v", tt.name, tt.saveErr, err)
			}

			// 検証2: 戻り値の件数が期待通りか (異常系なら 0, 正常系なら providersの件数)
			if len(got) != tt.wantCount {
				t.Errorf("[%s] 戻り値の件数が違います: 期待=%d, 実際=%d", tt.name, tt.wantCount, len(got))
			}

			// 3. 【スパイ】Repositoryにデータが正しく渡されたかの検証
			if tt.saveErr == nil {
				// 正常系なら Provider の合計件数が保存されているはず
				if len(repo.saved) != tt.wantCount {
					t.Errorf("[%s] 保存された件数が不正です: 期待=%d, 実際=%d", tt.name, tt.wantCount, len(repo.saved))
				}
			}
		})
	}
}

// 1. このテスト専用の「激遅 Provider」を定義
type slowMockProvider struct{}

// Name メソッドを実装（インターフェースの要件）
func (p *slowMockProvider) Name() string {
	return "slow-mock"
}

// Fetch メソッドを実装（インターフェースの要件）
func (p *slowMockProvider) Fetch(ctx context.Context) ([]sdk.Notification, error) {
	select {
	case <-time.After(5 * time.Second): // 5秒待機
		return []sdk.Notification{{ID: "slow", Title: "Late"}}, nil
	case <-ctx.Done(): // 親の Context がキャンセルされたら即座に終了
		return nil, ctx.Err()
	}
}

func TestAggregateAndSave_GoroutineLeak(t *testing.T) {
	// 1. 実行前の goroutine 数を記録
	initialGoroutines := runtime.NumGoroutine()

	// 2. 5秒かかる「激遅」MockProvider を作成
	slowProvider := &slowMockProvider{}

	// 3. 依存関係のセットアップ
	repo := &MockRepository{}
	logger := slog.Default()

	service := notification.NewService(logger, repo, slowProvider)

	// 4. 実行（内部で2秒のタイムアウトが発生し、すぐ戻ってくる）
	ctx := context.Background()
	_, _ = service.AggregateAndSave(ctx)

	// 5. Worker goroutine が後片付け（チャネルへの送信・終了）を終えるのを少し待つ
	time.Sleep(200 * time.Millisecond)

	// 6. 実行後の goroutine 数を記録
	finalGoroutines := runtime.NumGoroutine()

	// 7. リーク判定
	if finalGoroutines > initialGoroutines {
		t.Errorf("Goroutine leak detected! Initial: %d, Final: %d", initialGoroutines, finalGoroutines)
	} else {
		t.Logf("No leak: Initial: %d, Final: %d", initialGoroutines, finalGoroutines)
	}
}
