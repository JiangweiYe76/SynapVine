package db

import (
	"context"
	"database/sql"
	"fmt"
)

// migration is a single idempotent schema change. Each migration must be
// safe to re-run: use CREATE TABLE IF NOT EXISTS / CREATE INDEX IF NOT
// EXISTS or guard with information_schema.
type migration struct {
	name string
	stmt string
}

// migrations are applied in order. Adding a new entry is the only way
// to evolve the schema; never edit an existing statement in place.
var migrations = []migration{
	{
		name: "create_users",
		stmt: `CREATE TABLE IF NOT EXISTS users (
			id           VARCHAR(36)  NOT NULL,
			username     VARCHAR(255) NOT NULL,
			password     VARCHAR(255) NOT NULL,
			role         ENUM('admin','editor','viewer') NOT NULL DEFAULT 'viewer',
			token_ver    INT          NOT NULL DEFAULT 0,
			created_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uq_users_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
	{
		name: "create_refresh_tokens",
		stmt: `CREATE TABLE IF NOT EXISTS refresh_tokens (
			id          VARCHAR(36)  NOT NULL,
			user_id     VARCHAR(36)  NOT NULL,
			expires_at  TIMESTAMP    NOT NULL,
			created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_refresh_user (user_id),
			KEY idx_refresh_expires (expires_at),
			CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
	{
		name: "create_audit_events",
		stmt: `CREATE TABLE IF NOT EXISTS audit_events (
			id         BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id    VARCHAR(36),
			username   VARCHAR(255),
			action     VARCHAR(100) NOT NULL,
			resource   VARCHAR(255),
			ip         VARCHAR(45),
			created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			KEY idx_audit_user (user_id),
			KEY idx_audit_action (action),
			KEY idx_audit_created (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
}

// Migrate applies all known migrations to the connected database. MySQL
// has no advisory lock primitive that's portable across versions, so we
// rely on each statement being idempotent (IF NOT EXISTS). The schema is
// small and the server is single-instance, so this is sufficient.
func Migrate(ctx context.Context, conn *sql.DB) error {
	for _, m := range migrations {
		if _, err := conn.ExecContext(ctx, m.stmt); err != nil {
			return fmt.Errorf("migration %q failed: %w", m.name, err)
		}
	}
	return nil
}
