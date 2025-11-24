package database

// Package database is the middleware between the app and the persistent storage.
// All data (de)serialization to/from the SQL database is handled here.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB used by the rest of the app.
type AppDatabase interface {
	// LoginOrCreateUser: if a user with the given name exists, it returns its identifier.
	// Otherwise, it creates a new user and returns the new identifier.
	LoginOrCreateUser(ctx context.Context, name string) (string, error)

	// SetUserName updates the user's name given their identifier.
	SetUserName(ctx context.Context, identifier string, newName string) error

	// Search users by name (prefix / like)
	SearchUsers(ctx context.Context, search string) ([]SearchUserResult, error)

	// Ping checks that the DB connection is still alive.
	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// SearchUserResult rappresenta un singolo utente trovato dalla ricerca.
type SearchUserResult struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}

	// Keep the example table from the original template (optional, but harmless).
	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='example_table';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE example_table (id INTEGER NOT NULL PRIMARY KEY, name TEXT);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating example_table structure: %w", err)
		}
	}

	// Create the "users" table if it does not exist.
	usersStmt := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		identifier TEXT UNIQUE NOT NULL
	);`
	if _, err := db.Exec(usersStmt); err != nil {
		return nil, fmt.Errorf("error creating users table: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

// SearchUsers cerca gli utenti per nome usando una ricerca tipo prefisso.
func (db *appdbimpl) SearchUsers(ctx context.Context, search string) ([]SearchUserResult, error) {
	rows, err := db.c.QueryContext(ctx, `
		SELECT identifier, name
		FROM users
		WHERE name LIKE ? || '%'
		ORDER BY name ASC
	`, search)
	if err != nil {
		return nil, fmt.Errorf("error searching users: %w", err)
	}
	defer rows.Close()

	var results []SearchUserResult

	for rows.Next() {
		var r SearchUserResult
		if err := rows.Scan(&r.Identifier, &r.Name); err != nil {
			return nil, fmt.Errorf("error scanning search result row: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error while searching users: %w", err)
	}

	return results, nil
}
