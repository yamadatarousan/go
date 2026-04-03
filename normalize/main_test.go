package main

import (
	"testing"
)

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"基本形", "USER-123", "user-123"},
		{"空白あり", "  user-456  ", "user-456"},
		{"不正な記号混入", "user_789!@#", "user789"},
		{"全角（そのまま除去）", "ｕｓｅｒ", ""},
		{"空文字", "", ""},
		{"ハイフンのみ", "---", "---"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeID(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func FuzzNormalizeID(f *testing.F) {
	f.Add("Valid-ID-123")
	f.Add("   ")
	f.Add("!@#$%^&*()")
	f.Add("漢字とEmoji🚀")

	f.Fuzz(func(t *testing.T, input string) {
		// 修正点: () を付けて引数を渡す
		result := NormalizeID(input)

		// 1. クラッシュ（パニック）しないことを検証
		// 2. 冪等性を検証
		if NormalizeID(result) != result {
			t.Errorf("Idempotency failed for input %q: first result %q, second result %q", input, result, NormalizeID(result))
		}
	})
}
