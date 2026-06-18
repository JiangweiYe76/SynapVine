// Package store provides database-backed persistence for users, refresh
// tokens, and audit events. The console service used to keep users in
// an in-memory map seeded with a hard-coded admin; this package is the
// replacement that closes that gap.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"console/internal/model"
)

// ErrNotFound is returned by store methods when the requested row does
// not exist. Callers translate this into a 404.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is returned when a uniqueness constraint is violated
// (e.g. username already taken).
var ErrDuplicate = errors.New("duplicate")

// UserStore persists User rows in MySQL.
type UserStore struct {
	db *sql.DB
}

// NewUserStore returns a UserStore backed by the given *sql.DB.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// Create inserts a new user. Returns ErrDuplicate when the username is
// already taken.
func (s *UserStore) Create(ctx context.Context, u *model.User) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, username, password, role, token_ver, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.Password, u.Role, u.TokenVer, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// GetByUsername fetches a user by username, including the password hash.
// Returns ErrNotFound when no row matches.
func (s *UserStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, username, password, role, token_ver, created_at, updated_at
		 FROM users WHERE username = ?`, username,
	)
	return scanUser(row)
}

// GetByID fetches a user by primary key. Returns ErrNotFound when no
// row matches.
func (s *UserStore) GetByID(ctx context.Context, id string) (*model.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, username, password, role, token_ver, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	)
	return scanUser(row)
}

// Count returns the total number of users in the table. Used by the
// seed tool to decide whether to bootstrap a first admin.
func (s *UserStore) Count(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

// BumpTokenVer increments token_ver on the given user. The auth handler
// calls this on logout and password change to invalidate every
// outstanding JWT issued before the bump.
func (s *UserStore) BumpTokenVer(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET token_ver = token_ver + 1, updated_at = ? WHERE id = ?`,
		time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("bump token_ver: %w", err)
	}
	return nil
}

// scanUser is a shared row scanner for SELECT statements that project
// the full user column set.
func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.TokenVer, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

// isDuplicateKey reports whether err is a MySQL 1062 duplicate-key error.
// We avoid pulling in the driver's error type so the rest of the code
// stays decoupled from the driver.
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	// MySQL error 1062: Duplicate entry ... for key ...
	return containsAny(err.Error(), "Error 1062", "Duplicate entry")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) == 0 {
			continue
		}
		if indexOf(s, sub) >= 0 {
			return true
		}
	}
	return false
}

// indexOf is a tiny strings.Index replacement that avoids importing
// "strings" for a single call.
func indexOf(s, sub string) int {
	n, m := len(s), len(sub)
	if m == 0 {
		return 0
	}
	for i := 0; i+m <= n; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}
