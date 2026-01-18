package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"io"
	"os"
	"path/filepath"
)

// CreateGroup creates a new group conversation and adds the owner plus
// the given members (identified by their identifiers).
func (db *appdbimpl) CreateGroup(
	ctx context.Context,
	ownerIdentifier string,
	name string,
	memberIdentifiers []string,
) (int64, error) {

	ownerID, err := db.getUserIDByIdentifier(ctx, ownerIdentifier)
	if err != nil {
		return 0, err
	}

	tx, err := db.c.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// 1) Insert into conversations.
	res, err := tx.ExecContext(ctx, `
		INSERT INTO conversations (type, name)
		VALUES ('group', ?)
	`, name)
	if err != nil {
		return 0, fmt.Errorf("cannot insert conversation: %w", err)
	}

	chatID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("cannot get conversation id: %w", err)
	}

	// 2) Add owner as member.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO conversation_members (conversation_id, user_id)
		VALUES (?, ?)
	`, chatID, ownerID); err != nil {
		return 0, fmt.Errorf("cannot add owner as member: %w", err)
	}

	// 3) Add additional members (if any).
	for _, ident := range memberIdentifiers {
		ident = strings.TrimSpace(ident)
		if ident == "" {
			continue
		}
		userID, err := db.getUserIDByIdentifier(ctx, ident)
		if err != nil {
			// You may choose to skip invalid users instead of failing.
			return 0, fmt.Errorf("cannot resolve member %q: %w", ident, err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO conversation_members (conversation_id, user_id)
			VALUES (?, ?)
		`, chatID, userID); err != nil {
			return 0, fmt.Errorf("cannot add member %q: %w", ident, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("cannot commit group creation: %w", err)
	}

	return chatID, nil
}

// AddMembersToGroup adds one or more members to an existing group conversation.
func (db *appdbimpl) AddMembersToGroup(
	ctx context.Context,
	requesterIdentifier string,
	chatID int64,
	memberIdentifiers []string,
) error {

	requesterID, err := db.getUserIDByIdentifier(ctx, requesterIdentifier)
	if err != nil {
		return err
	}

	// Ensure requester is a member of the group.
	var exists int
	err = db.c.QueryRowContext(ctx, `
		SELECT 1
		FROM conversation_members cm
		JOIN conversations c ON c.id = cm.conversation_id
		WHERE cm.conversation_id = ? AND cm.user_id = ? AND c.type = 'group'
	`, chatID, requesterID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("requester not member of group or group not found: %w", err)
		}
		return fmt.Errorf("cannot check requester membership: %w", err)
	}

	for _, ident := range memberIdentifiers {
		ident = strings.TrimSpace(ident)
		if ident == "" {
			continue
		}
		userID, err := db.getUserIDByIdentifier(ctx, ident)
		if err != nil {
			return fmt.Errorf("cannot resolve member %q: %w", ident, err)
		}
		if _, err := db.c.ExecContext(ctx, `
			INSERT OR IGNORE INTO conversation_members (conversation_id, user_id)
			VALUES (?, ?)
		`, chatID, userID); err != nil {
			return fmt.Errorf("cannot add member %q: %w", ident, err)
		}
	}

	return nil
}

// LeaveGroup removes the requester from the group.
func (db *appdbimpl) LeaveGroup(
	ctx context.Context,
	requesterIdentifier string,
	chatID int64,
) error {

	requesterID, err := db.getUserIDByIdentifier(ctx, requesterIdentifier)
	if err != nil {
		return err
	}

	res, err := db.c.ExecContext(ctx, `
		DELETE FROM conversation_members
		WHERE conversation_id = ? AND user_id = ?
	`, chatID, requesterID)
	if err != nil {
		return fmt.Errorf("cannot leave group: %w", err)
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}

	return err
}

// SetGroupName updates the name of the group (if requester is a member).
func (db *appdbimpl) SetGroupName(
	ctx context.Context,
	requesterIdentifier string,
	chatID int64,
	newName string,
) error {

	requesterID, err := db.getUserIDByIdentifier(ctx, requesterIdentifier)
	if err != nil {
		return err
	}

	// Ensure requester is a member.
	var exists int
	err = db.c.QueryRowContext(ctx, `
		SELECT 1
		FROM conversation_members
		WHERE conversation_id = ? AND user_id = ?
	`, chatID, requesterID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("requester not member of group: %w", err)
		}
		return fmt.Errorf("cannot check requester membership: %w", err)
	}

	_, err = db.c.ExecContext(ctx, `
		UPDATE conversations
		SET name = ?
		WHERE id = ? AND type = 'group'
	`, newName, chatID)
	if err != nil {
		return fmt.Errorf("cannot update group name: %w", err)
	}

	return nil
}

// SetGroupPhoto sets or updates the photo URL of the group.
func (rt *_router) setGroupPhoto(
  w http.ResponseWriter,
  r *http.Request,
  ps httprouter.Params,
  ctx reqcontext.RequestContext,
) {
  if ctx.UserIdentifier == "" {
    writeErrorJSON(w, http.StatusUnauthorized, "Invalid or missing token")
    return
  }

  chatID, err := strconv.ParseInt(ps.ByName("chatId"), 10, 64)
  if err != nil || chatID <= 0 {
    writeErrorJSON(w, http.StatusBadRequest, "Invalid chatId")
    return
  }

  // Max 10MB
  if err := r.ParseMultipartForm(10 << 20); err != nil {
    writeErrorJSON(w, http.StatusBadRequest, "Invalid multipart form")
    return
  }

  file, header, err := r.FormFile("file")
  if err != nil {
    writeErrorJSON(w, http.StatusBadRequest, "Missing file field")
    return
  }
  defer file.Close()

  contentType := header.Header.Get("Content-Type")
  if contentType == "" {
    buf := make([]byte, 512)
    n, _ := file.Read(buf)
    contentType = http.DetectContentType(buf[:n])
    if _, err := file.Seek(0, io.SeekStart); err != nil {
      writeErrorJSON(w, http.StatusInternalServerError, "cannot reset file reader")
      return
    }
  }

  if !isValidMediaType(contentType) {
    writeErrorJSON(w, http.StatusBadRequest, "unsupported media type")
    return
  }

  ext := getExtensionFromContentType(contentType)
  if ext == "" {
    ext = strings.ToLower(filepath.Ext(header.Filename))
  }
  if ext == "" {
    ext = ".jpg"
  }

  // uploads/groups
  if err := os.MkdirAll("./uploads/groups", 0o755); err != nil {
    writeErrorJSON(w, http.StatusInternalServerError, "cannot create groups directory")
    return
  }

  filename := fmt.Sprintf("group-%d%s", chatID, ext)
  dstPath := filepath.Join("./uploads/groups", filename)

  dst, err := os.Create(dstPath)
  if err != nil {
    writeErrorJSON(w, http.StatusInternalServerError, "cannot create destination file")
    return
  }
  defer dst.Close()

  if _, err := io.Copy(dst, file); err != nil {
    writeErrorJSON(w, http.StatusInternalServerError, "cannot save file")
    return
  }

  // ServeFiles è /v1/uploads/...
  photoURL := "/v1/uploads/groups/" + filename

  if err := rt.db.SetGroupPhoto(r.Context(), ctx.UserIdentifier, chatID, photoURL); err != nil {
    ctx.Logger.WithError(err).Error("cannot persist group photo url")
    writeErrorJSON(w, http.StatusInternalServerError, "Internal server error")
    return
  }

  writeJSON(w, http.StatusOK, map[string]string{"photoUrl": photoURL})
}

