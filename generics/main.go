package main

import (
	"fmt"
	"strings"
)

// ToInClause は、スライスをSQLの IN 句のような形式 (1, 2, 3) や ('a', 'b') に変換します。
// int, string, floatなど「比較可能」な型なら何でも共通化できます。
func ToInClause[T any](values []T) string {
	var elements []string
	for _, v := range values {
		// %v を使って文字列化するロジックはどの型でも共通
		elements = append(elements, fmt.Sprintf("'%v'", v))
	}
	return fmt.Sprintf("(%s)", strings.Join(elements, ", "))
}

// Filter は、特定条件に合致する要素だけを抽出する汎用的な処理です。
func Filter[T any](list []T, keep func(T) bool) []T {
	var result []T
	for _, item := range list {
		if keep(item) {
			result = append(result, item)
		}
	}
	return result
}

func main() {
	// 検索条件の組み立て
	ids := []int{101, 202, 303}
	tags := []string{"golang", "generics", "clean-code"}

	fmt.Println("IDs IN:", ToInClause(ids))   // ('101', '202', '303')
	fmt.Println("Tags IN:", ToInClause(tags)) // ('golang', 'generics', 'clean-code')

	longTags := Filter(tags, func(s string) bool { return len(s) > 8 })
	fmt.Println("Long Tags:", longTags)
}
