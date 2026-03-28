package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ------------------------------------------------------------------
// 1. サービス層のインターフェース定義
// ------------------------------------------------------------------
// なぜ定義するのか：
// ハンドラが「具体的な実装」を知らなくて済むようにするためです。
// これにより、テスト時に本物の代わりに「モック」を渡せるようになります。
type GreetingService interface {
	GenerateGreeting(name string) (string, error)
}

// ------------------------------------------------------------------
// 2. サービス層の実装 (本番用)
// ------------------------------------------------------------------
// 構造体が空なのは、現時点でDB接続などの「状態」を保持する必要がないからです。
// メソッドを定義するための「土台」として存在しています。
type greetingServiceImpl struct{}

// NewGreetingService は本番用のサービスを生成します。
func NewGreetingService() GreetingService {
	return &greetingServiceImpl{}
}

// GenerateGreeting は実際のビジネスロジックです。
func (s *greetingServiceImpl) GenerateGreeting(name string) (string, error) {
	if name == "" {
		return "", errors.New("name is required")
	}
	return fmt.Sprintf("Hello, %s!", name), nil
}

// ------------------------------------------------------------------
// 3. ハンドラ層 (HTTPリクエスト処理)
// ------------------------------------------------------------------
type GreetingHandler struct {
	// ポイント：ここを greetingServiceImpl(具象) ではなく
	// GreetingService(インターフェース) 型にする。
	// これが「依存性の注入 (DI)」を受け入れるための窓口になります。
	service GreetingService
}

// NewGreetingHandler はハンドラを生成します。
// 外部から「GreetingServiceを満たす何か」を受け取ってセットします。
func NewGreetingHandler(service GreetingService) *GreetingHandler {
	return &GreetingHandler{service: service}
}

// ServeHTTP は http.Handler インターフェースを実装するためのメソッドです。
func (h *GreetingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. リクエストの解析
	name := r.URL.Query().Get("name")

	// 2. サービス層への委譲
	// h.service が本物かモックかは、このハンドラは知りません。
	greeting, err := h.service.GenerateGreeting(name)
	if err != nil {
		// ビジネスロジックでエラーが出た場合のHTTPレスポンス処理
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. レスポンスの構築
	res := map[string]string{"message": greeting}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// ------------------------------------------------------------------
// 4. メイン関数 (依存関係の組み立て)
// ------------------------------------------------------------------
func main() {
	// 依存関係を下から順に組み立てる
	svc := NewGreetingService()
	handler := NewGreetingHandler(svc)

	http.Handle("/greet", handler)
	fmt.Println("Server starting at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
