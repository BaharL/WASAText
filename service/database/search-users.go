package database

import (
	"context"
	"fmt"
)

// SearchUserResult is a minimal user view used for search results.
type SearchUserResult struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

// SearchUsers returns all users whose name contains the given `search` substring.
// The search is case-sensitive according to SQLite collation (per ora va benissimo così).
func (db *appdbimpl) SearchUsers(ctx context.Context, search string) ([]SearchUserResult, error) {
	rows, err := db.c.QueryContext(ctx, `
		SELECT identifier, name
		FROM users
		WHERE name LIKE '%' || ? || '%'
		ORDER BY name ASC
	`, search)
	if err != nil {
		return nil, fmt.Errorf("cannot search users: %w", err)
	}
	defer rows.Close()

	var results []SearchUserResult

	for rows.Next() {
		var u SearchUserResult
		if err := rows.Scan(&u.Identifier, &u.Name); err != nil {
			return nil, fmt.Errorf("cannot scan user row: %w", err)
		}
		results = append(results, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}
