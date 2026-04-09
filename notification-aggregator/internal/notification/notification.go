package notification

import (
	"context"
	"time"
)

// Notification は集約される通知の基本構造です
type Notification struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Provider は外部の通知ソース（Slack, GitHub, DB等）を抽象化します
type Provider interface {
	Fetch(ctx context.Context) ([]Notification, error)
	Name() string
}
