package notification

import (
	"notification-sdk"
)

func DedupeSlice(notifications []sdk.Notification) []sdk.Notification {
	result := []sdk.Notification{}
	for _, n := range notifications {
		exists := false
		for _, r := range result {
			if n.ID == r.ID {
				exists = true
				break
			}
		}
		if !exists {
			result = append(result, n)
		}
	}
	return result
}

func DedupeMap(notifications []sdk.Notification) []sdk.Notification {
	result := make([]sdk.Notification, 0, len(notifications))
	seen := make(map[string]bool)
	for _, n := range notifications {
		if !seen[n.ID] {
			seen[n.ID] = true
			result = append(result, n)
		}
	}
	return result
}
