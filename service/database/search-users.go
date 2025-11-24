package database

import (
	"context"
	"fmt"
)

// UserSearchResult is a minimal user view used for search results.
type UserSearchResult struct {
	ID   int64
	Name string
}

// SearchUsers returns all users whose name contains the given `search` substring.
// The search is case-sensitive according to SQLite collation (per ora va benissimo così).
func (db *appdbimpl) SearchUsers(ctx context.Context, search string) ([]UserSearchResult, error) {
	rows, err := db.c.QueryContext(ctx, `
		SELECT id, name
		FROM users
		WHERE name LIKE '%' || ? || '%'
		ORDER BY name ASC
	`, search)
	if err != nil {
		return nil, fmt.Errorf("cannot search users: %w", err)
	}
	defer rows.Close()

	var results []UserSearchResult

	for rows.Next() {
		var u UserSearchResult
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, fmt.Errorf("cannot scan user row: %w", err)
		}
		results = append(results, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}
