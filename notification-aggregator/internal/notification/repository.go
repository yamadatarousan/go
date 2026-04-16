package notification

import (
	"context"
	"database/sql"
	"notification-sdk"
)

type SqliteRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &SqliteRepository{db: db}
}

// SaveAll は通知を DB に保存します。重複は ID で判断して無視します（Dedupe）。
func (r *SqliteRepository) SaveAll(ctx context.Context, notes []sdk.Notification) error {
	// 段階 3: 重複排除のため INSERT OR IGNORE を使用
	query := `
			INSERT OR IGNORE INTO notifications (id, source, title, content, created_at)
			VALUES (?, ?, ?, ?, ?)`

	for _, n := range notes {
		_, err := r.db.ExecContext(ctx, query, n.ID, n.Source, n.Title, n.Content, n.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// FetchCached は DB から過去の通知を取得します（キャッシュ利用）。
func (r *SqliteRepository) FetchCached(ctx context.Context) ([]sdk.Notification, error) {
	query := `SELECT id, source, title, content, created_at FROM notifications ORDER BY created_at DESC LIMIT 50`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []sdk.Notification
	for rows.Next() {
		var n sdk.Notification
		if err := rows.Scan(&n.ID, &n.Source, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, n)
	}
	return results, nil
}
