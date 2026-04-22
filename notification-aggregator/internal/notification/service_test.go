package notification_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"testing"
	"time"

	"notification-aggregator/internal/contextutil"
	"notification-aggregator/internal/logging"
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
	// テストのためにあえて少し待機する（例えば 100ms）
	// この間にキャンセルが来れば、それを即座に捕まえる
	select {
	case <-time.After(100 * time.Millisecond):
		return m.data, m.err
	case <-ctx.Done():
		// キャンセルが来たら、処理を中断して理由を返す
		return nil, ctx.Err()
	}
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

			got, _, err := svc.AggregateAndSave(context.Background())

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
	_, _, _ = service.AggregateAndSave(ctx)

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

func TestAggregateAndSave_PartialFailure(t *testing.T) {
	// 1. 成功する Provider (データを入れる)
	successProvider := &MockProvider{
		name: "Slack",
		data: []sdk.Notification{{ID: "msg-1", Title: "Hello"}},
		err:  nil,
	}

	// 2. 失敗する Provider (エラーを入れる)
	failProvider := &MockProvider{
		name: "Github",
		data: nil,
		err:  errors.New("connection reset by peer"),
	}

	repo := &MockRepository{}
	logger := slog.Default()

	// 既存の構造体を渡すだけ
	service := notification.NewService(logger, repo, successProvider, failProvider)

	// 3. 実行
	ctx := context.Background()
	notifications, _, err := service.AggregateAndSave(ctx)

	// 4. 検証
	// 成功分は取れているか？
	if len(notifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(notifications))
	}

	// エラーメッセージの確認
	if err != nil {
		fmt.Printf("\n--- Joined Error Message ---\n%v\n----------------------------\n", err)
	}
}

func TestAggregateAndSave_Cancellation(t *testing.T) {
	// 1. キャンセルできる Context を作成
	ctx, cancel := context.WithCancel(context.Background())

	// 実行した瞬間にキャンセルする
	cancel()

	// サービスを実行
	repo := &MockRepository{}
	provider := &MockProvider{name: "Slow", data: nil, err: nil}
	service := notification.NewService(slog.Default(), repo, provider)

	_, _, err := service.AggregateAndSave(ctx)

	// キャンセルが伝わっていれば、エラーが返ってくるはず
	if err == nil {
		t.Error("expected error due to cancellation, but got nil")
	}

	// エラーが context.Canceled か確認
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestAggregateAndSave_TDT(t *testing.T) {
	// テストケースの「表（テーブル）」を定義する
	tests := []struct {
		name         string
		setupContext func() (context.Context, context.CancelFunc)
		providers    []sdk.Provider
		wantCount    int
		wantErr      error
	}{
		{
			name: "Success: 全Providerが成功",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			providers: []sdk.Provider{
				&MockProvider{name: "P1", data: []sdk.Notification{{ID: "1"}}, err: nil},
				&MockProvider{name: "P2", data: []sdk.Notification{{ID: "2"}}, err: nil},
			},
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name: "Partial Failure: 一部失敗しても続行する",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			providers: []sdk.Provider{
				&MockProvider{name: "Success", data: []sdk.Notification{{ID: "1"}}, err: nil},
				&MockProvider{name: "Fail", data: nil, err: errors.New("network error")}, // 部分失敗
			},
			wantCount: 1,   // 成功した1件だけ取れるはず
			wantErr:   nil, // 致命的エラーではないので戻り値エラーは nil
		},
		{
			name: "Fatal: キャンセルされたら即座にエラーを返す",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // 実行した瞬間にキャンセル状態にする
				return ctx, cancel
			},
			providers: []sdk.Provider{
				// 少し待つ（キャンセルを検知する）Provider
				&MockProvider{name: "Slow", data: nil, err: nil},
			},
			wantCount: 0,
			wantErr:   context.Canceled, // キャンセルエラーが返ってくるはず
		},
	}

	// テーブルの要素をループで回して、サブテストとして実行
	for _, tt := range tests {
		// t.Run で個別のテストとして実行する
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupContext()
			defer cancel()

			// 1. まず普通に JSONHandler を作る
			baseHandler := slog.NewJSONHandler(os.Stdout, nil)
			// 2. それを自作の ContextHandler で包む
			logger := slog.New(&logging.ContextHandler{Handler: baseHandler})

			// テスト用の Request ID を仕込む
			testID := "test-req-123"
			ctx = contextutil.WithRequestID(ctx, testID)

			repo := &MockRepository{}
			service := notification.NewService(logger, repo, tt.providers...)

			// 実行
			got, _, err := service.AggregateAndSave(ctx)

			// 1. エラーの検証
			if tt.wantErr != nil {
				// 期待するエラーがある場合（errors.Is で中身を判定）
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				// エラーを期待していないのにエラーが出た場合
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}

			// 2. 件数の検証
			if len(got) != tt.wantCount {
				t.Errorf("expected %d notifications, got %d", tt.wantCount, len(got))
			}
		})
	}
}
