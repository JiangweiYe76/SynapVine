package db

import (
	"context"
	"database/sql"
	"fmt"
)

// mysqlMigration is a single idempotent schema change.
type mysqlMigration struct {
	name string
	stmt string
}

// mysqlMigrations are applied in order on startup.
var mysqlMigrations = []mysqlMigration{
	{
		name: "create_papers",
		stmt: `CREATE TABLE IF NOT EXISTS papers (
			id          VARCHAR(36)  NOT NULL,
			title       VARCHAR(500) NOT NULL,
			authors     TEXT,
			source_url  VARCHAR(1000),
			raw_text    LONGTEXT     NOT NULL,
			status      VARCHAR(50)  NOT NULL DEFAULT 'uploaded',
			created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_papers_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
	{
		name: "create_review_queue",
		stmt: `CREATE TABLE IF NOT EXISTS review_queue (
			id              VARCHAR(36) NOT NULL,
			paper_id        VARCHAR(36) NOT NULL,
			extracted_nodes JSON        NOT NULL,
			extracted_edges JSON        NOT NULL,
			status          VARCHAR(50) NOT NULL DEFAULT 'pending',
			reviewer_id     VARCHAR(36),
			review_notes    TEXT,
			created_at      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
			reviewed_at     TIMESTAMP   NULL,
			PRIMARY KEY (id),
			KEY idx_review_paper (paper_id),
			KEY idx_review_status (status),
			CONSTRAINT fk_review_paper FOREIGN KEY (paper_id) REFERENCES papers(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
}

// MigrateMySQL applies all known MySQL migrations.
func MigrateMySQL(ctx context.Context, conn *sql.DB) error {
	for _, m := range mysqlMigrations {
		if _, err := conn.ExecContext(ctx, m.stmt); err != nil {
			return fmt.Errorf("migration %q failed: %w", m.name, err)
		}
	}
	return nil
}
