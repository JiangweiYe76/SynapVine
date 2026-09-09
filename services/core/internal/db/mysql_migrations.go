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
	// skipIf optionally reports whether this migration should be
	// skipped for the current database (e.g. the column already
	// exists because the table was created fresh with it included).
	// The migration runner re-runs every entry on startup, so any
	// non-idempotent statement (ALTER TABLE ADD COLUMN) must be
	// guarded here.
	skipIf func(ctx context.Context, conn *sql.DB) (bool, error)
}

// columnExists reports whether the given column is present on a table
// in the current database.
func columnExists(ctx context.Context, conn *sql.DB, table, column string) (bool, error) {
	var n int
	err := conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		table, column).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
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
				arxiv_id    VARCHAR(32)  NULL,
				raw_text    LONGTEXT     NOT NULL,
				pdf_data    LONGBLOB,
				status      VARCHAR(50)  NOT NULL DEFAULT 'uploaded',
				created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				KEY idx_papers_status (status),
				UNIQUE KEY uq_papers_arxiv_id (arxiv_id)
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
	{
		name: "create_llm_providers",
		stmt: `CREATE TABLE IF NOT EXISTS llm_providers (
			id          VARCHAR(36)  NOT NULL,
			name        VARCHAR(100) NOT NULL,
			base_url    VARCHAR(500) NOT NULL,
			api_key     VARCHAR(500) NOT NULL,
			model       VARCHAR(100) NOT NULL,
			max_tokens  INT          NOT NULL DEFAULT 4096,
			temperature DOUBLE       NOT NULL DEFAULT 0.7,
			is_default  BOOLEAN      NOT NULL DEFAULT FALSE,
			is_enabled  BOOLEAN      NOT NULL DEFAULT TRUE,
			created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uq_llm_providers_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
	{
		name: "create_embedding_providers",
		stmt: `CREATE TABLE IF NOT EXISTS embedding_providers (
			id          VARCHAR(36)  NOT NULL,
			name        VARCHAR(100) NOT NULL,
			base_url    VARCHAR(500) NOT NULL,
			api_key     VARCHAR(500) NOT NULL,
			model       VARCHAR(100) NOT NULL,
			dimensions  INT          NOT NULL DEFAULT 1536,
			is_default  BOOLEAN      NOT NULL DEFAULT FALSE,
			is_enabled  BOOLEAN      NOT NULL DEFAULT TRUE,
			created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uq_embedding_providers_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
	{
		// API keys are encrypted at rest (AES-256-GCM, base64 envelope).
		// Widen the column to fit nonce + ciphertext + tag + encoding.
		name: "widen_llm_providers_api_key",
		stmt: `ALTER TABLE llm_providers MODIFY COLUMN api_key VARCHAR(1024) NOT NULL`,
	},
	{
		name: "widen_embedding_providers_api_key",
		stmt: `ALTER TABLE embedding_providers MODIFY COLUMN api_key VARCHAR(1024) NOT NULL`,
	},
	{
		// ArXiv ingestion dedup: papers fetched by the discovery
		// scheduler carry their ArXiv identifier. NULL is allowed
		// (manually uploaded papers), duplicates are rejected by the
		// unique key.
		name: "add_papers_arxiv_id",
		stmt: `ALTER TABLE papers ADD COLUMN arxiv_id VARCHAR(32) NULL, ADD UNIQUE KEY uq_papers_arxiv_id (arxiv_id)`,
		skipIf: func(ctx context.Context, conn *sql.DB) (bool, error) {
			return columnExists(ctx, conn, "papers", "arxiv_id")
		},
	},
	{
		// LLM token usage accounting: one row per analysis call,
		// submitted by the discovery service after a successful
		// extraction. Consumed by the console usage overview.
		name: "create_llm_usage",
		stmt: `CREATE TABLE IF NOT EXISTS llm_usage (
			id                VARCHAR(36)  NOT NULL,
			paper_id          VARCHAR(36)  NOT NULL,
			provider_id       VARCHAR(36)  NOT NULL,
			model             VARCHAR(100) NOT NULL,
			prompt_tokens     INT          NOT NULL DEFAULT 0,
			completion_tokens INT          NOT NULL DEFAULT 0,
			total_tokens      INT          NOT NULL DEFAULT 0,
			created_at        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_llm_usage_provider_time (provider_id, created_at),
			KEY idx_llm_usage_paper (paper_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
}

// MigrateMySQL applies all known MySQL migrations.
func MigrateMySQL(ctx context.Context, conn *sql.DB) error {
	for _, m := range mysqlMigrations {
		if m.skipIf != nil {
			skip, err := m.skipIf(ctx, conn)
			if err != nil {
				return fmt.Errorf("migration %q precondition failed: %w", m.name, err)
			}
			if skip {
				continue
			}
		}
		if _, err := conn.ExecContext(ctx, m.stmt); err != nil {
			return fmt.Errorf("migration %q failed: %w", m.name, err)
		}
	}
	return nil
}
