package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// テスト用の「モック」サービス
// ------------------------------------------------------------------
// 本物の greetingServiceImpl の代わりにこれを使います。
type MockService struct {
	// テストケースごとに「どんな結果を返したいか」を柔軟に変えられるよう、
	// 関数自体を変数として持っておきます。
	MockFunc func(name string) (string, error)
}

// Interface (GreetingService) を満たすためのメソッド実装。
func (m *MockService) GenerateGreeting(name string) (string, error) {
	// 内部に保持しているテスト用の関数を実行するだけ
	return m.MockFunc(name)
}

// ------------------------------------------------------------------
// ハンドラのテスト本体
// ------------------------------------------------------------------
func TestGreetingHandler(t *testing.T) {
	// テーブル駆動テスト: 入力と期待値を整理
	tests := []struct {
		name       string // テストケース名
		query      string // リクエストに含める名前
		mockRes    string // モックに返してほしい挨拶
		mockErr    error  // モックに返してほしいエラー
		wantStatus int    // 期待するHTTPステータス
		wantBody   string // 期待するレスポンスボディ
	}{
		{
			name:       "正常系：モックが成功を返す場合",
			query:      "Alice",
			mockRes:    "Mock Hello Alice", // 本物のロジックとは違う値を返させてみる
			mockErr:    nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"Mock Hello Alice"}`,
		},
		{
			name:       "異常系：モックがエラーを返す場合",
			query:      "BadGuy",
			mockRes:    "",
			mockErr:    errors.New("service error occurred"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "service error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 【モックの準備】
			// このテストケース限定の「嘘の挙動」を定義する。
			mock := &MockService{
				MockFunc: func(name string) (string, error) {
					return tt.mockRes, tt.mockErr
				},
			}

			// 2. 【依存性の注入 (DI)】
			// 本物のサービスの代わりに、今作った「モック」をハンドラに渡す。
			handler := NewGreetingHandler(mock)

			// 3. 【疑似リクエストの発行】
			// 実際にネットワークを介さず、メモリ上でリクエストをシミュレートする。
			req := httptest.NewRequest(http.MethodGet, "/greet?name="+tt.query, nil)
			rec := httptest.NewRecorder() // レスポンスを受け止める器

			// 4. 【実行】
			// ハンドラの ServeHTTP を直接叩く
			handler.ServeHTTP(rec, req)

			// 5. 【検証】
			// HTTPステータスコードが正しいか？
			if rec.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tt.wantStatus)
			}

			// ボディの内容が正しいか？
			// (http.Error は末尾に改行を付けるので TrimSpace で除去)
			actualBody := strings.TrimSpace(rec.Body.String())
			if actualBody != tt.wantBody {
				t.Errorf("body: got %q, want %q", actualBody, tt.wantBody)
			}
		})

	}

}
