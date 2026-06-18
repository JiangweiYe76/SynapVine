package store

import (
	"context"
	"database/sql"
	"fmt"
)

// AuditStore appends audit events. Writes are best-effort: callers
// should log the error but not fail the user-facing request when an
// audit insert fails.
type AuditStore struct {
	db *sql.DB
}

// NewAuditStore returns an AuditStore backed by the given *sql.DB.
func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

// AuditEvent is one row in audit_events.
type AuditEvent struct {
	UserID   string
	Username string
	Action   string
	Resource string
	IP       string
}

// Log inserts a new audit event.
func (s *AuditStore) Log(ctx context.Context, e AuditEvent) error {
	// user_id and username are nullable; pass nil when empty so the
	// column stores SQL NULL instead of an empty string.
	var userID, username any
	if e.UserID != "" {
		userID = e.UserID
	}
	if e.Username != "" {
		username = e.Username
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_events (user_id, username, action, resource, ip)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, username, e.Action, nullString(e.Resource), nullString(e.IP),
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
