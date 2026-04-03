package main

import (
	"regexp"
	"strings"
)

var (
	// 英数字とハイフン以外を除去するための正規表現
	invalidChars = regexp.MustCompile(`[^a-z0-9-]`)
)

func NormalizeID(id string) string {
	// 1. 前後の空白をトリム
	s := strings.TrimSpace(id)
	// 2. 小文字に変換
	s = strings.ToLower(s)
	// 3. 許可されていない文字を除去
	s = invalidChars.ReplaceAllString(s, "")

	return s
}


