package notification

import (
	"fmt"
	"notification-sdk"
	"testing"
)

func BenchmarkDedupe(b *testing.B) {
	// 1000件のテストデータを用意
	data := make([]sdk.Notification, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = sdk.Notification{
			ID:    fmt.Sprintf("id-%d", i%100),
			Title: "Benchmark Test",
		}
	}

	b.Run("Slice-Loop", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DedupeSlice(data)
		}
	})

	b.Run("Map-Chack", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DedupeMap(data)
		}
	})
}

func FuzzDedupeMap(f *testing.F) {
	// 1. シード（初期のテストケース）を追加
	f.Add(5) // データ件数のヒント

	f.Fuzz(func(t *testing.T, n int) {
		// n が負数だったり大きすぎたりしないよう調整
		count := n % 100
		if count < 0 {
			count = -count
		}

		// ランダムなIDを持つデータを作成
		data := make([]sdk.Notification, count)
		for i := 0; i < count; i++ {
			data[i] = sdk.Notification{
				ID: fmt.Sprintf("id-%d", i), // あえて重複しやすいIDを生成
			}
		}

		// 実行
		got := DedupeMap(data)

		// 検証: 結果の中に重複したIDが存在しないことを確認
		seen := make(map[string]bool)
		for _, v := range got {
			if seen[v.ID] {
				t.Errorf("重複が検出されました: ID=%s", v.ID)
			}
			seen[v.ID] = true
		}
	})
}
