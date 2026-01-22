package api

import "strings"

func isUniqueUsernameError(err error) bool {
    if err == nil {
        return false
    }

    msg := strings.ToLower(err.Error())

    // SQLite
    if strings.Contains(msg, "unique constraint failed") {
        return true
    }

    // Postgres
    if strings.Contains(msg, "duplicate key") &&
        strings.Contains(msg, "unique") {
        return true
    }

    // MySQL
    if strings.Contains(msg, "duplicate entry") {
        return true
    }

    return false
}
