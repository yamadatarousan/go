package provider

import (
	"notification-sdk"
)

// 設定ファイルを読み込むための構造体
type ProviderConfig struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type Config struct {
	Providers []ProviderConfig `json:"providers"`
}

// NewProvider は設定に基づいて適切な Provider インスタンスを生成します (Factory パターン)
func NewProvider(cfg ProviderConfig) sdk.Provider {
	switch cfg.Type {
	case "mock":
		// ※ MockProvider 側に Name フィールドを追加して持たせると、ログ出力時に便利です
		return &MockProvider{}
	case "slow":
		return &SlowProvider{}
	default:
		return nil //未知のタイプ
	}
}
