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
