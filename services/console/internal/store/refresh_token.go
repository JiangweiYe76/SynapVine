package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RefreshTokenStore persists opaque refresh tokens. We store the token
// id (a UUID) as the lookup key. The full token value is only ever
// shown to the client once; the row contents themselves are not
// sensitive because the JWT access token is still required to use them.
type RefreshTokenStore struct {
	db *sql.DB
}

// NewRefreshTokenStore returns a RefreshTokenStore backed by the given
// *sql.DB.
func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{db: db}
}

// Create inserts a new refresh token row.
func (s *RefreshTokenStore) Create(ctx context.Context, id, userID string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (id, user_id, expires_at, created_at)
		 VALUES (?, ?, ?, ?)`,
		id, userID, expiresAt, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// Lookup returns the user_id and expiry for the given refresh token id.
// Returns ErrNotFound when the token does not exist or has been deleted
// (we hard-delete on use / logout, not soft-delete).
func (s *RefreshTokenStore) Lookup(ctx context.Context, id string) (string, time.Time, error) {
	var userID string
	var expiresAt time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM refresh_tokens WHERE id = ?`, id,
	).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", time.Time{}, ErrNotFound
		}
		return "", time.Time{}, fmt.Errorf("lookup refresh token: %w", err)
	}
	return userID, expiresAt, nil
}

// Delete removes a single refresh token. Used for logout.
func (s *RefreshTokenStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

// DeleteAllForUser removes every refresh token belonging to the given
// user. Used to force a full sign-out across every device.
func (s *RefreshTokenStore) DeleteAllForUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete refresh tokens for user: %w", err)
	}
	return nil
}
