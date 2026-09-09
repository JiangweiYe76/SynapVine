package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"core/internal/model"

	"github.com/google/uuid"
)

// LLMUsageRepository persists LLM token usage events in MySQL.
type LLMUsageRepository struct {
	db *sql.DB
}

// NewLLMUsageRepository returns a LLMUsageRepository backed by the
// given *sql.DB.
func NewLLMUsageRepository(db *sql.DB) *LLMUsageRepository {
	return &LLMUsageRepository{db: db}
}

// Insert records a single usage event.
func (r *LLMUsageRepository) Insert(ctx context.Context, u *model.LLMUsageRecordRequest) error {
	if u.PaperID == "" || u.ProviderID == "" {
		return fmt.Errorf("paper_id and provider_id are required")
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO llm_usage (id, paper_id, provider_id, model, prompt_tokens, completion_tokens, total_tokens)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), u.PaperID, u.ProviderID, u.Model,
		u.PromptTokens, u.CompletionTokens, u.TotalTokens,
	)
	if err != nil {
		return fmt.Errorf("insert llm usage: %w", err)
	}
	return nil
}

// UsageFilter bounds the summary query. Zero values mean unbounded.
type UsageFilter struct {
	From *time.Time
	To   *time.Time
}

// Summary returns aggregated usage (grand totals, per-provider
// breakdown, per-day series) for the given filter.
func (r *LLMUsageRepository) Summary(ctx context.Context, f UsageFilter) (*model.LLMUsageSummaryResponse, error) {
	where := "1=1"
	var args []any
	if f.From != nil {
		where += " AND created_at >= ?"
		args = append(args, *f.From)
	}
	if f.To != nil {
		where += " AND created_at < ?"
		args = append(args, *f.To)
	}

	summary := &model.LLMUsageSummaryResponse{
		ByProvider: []model.LLMUsageProviderSummary{},
		ByDay:      []model.LLMUsageDaySummary{},
	}

	// Grand totals.
	totals := model.LLMUsageProviderSummary{}
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(COUNT(*),0), COALESCE(SUM(prompt_tokens),0),
		       COALESCE(SUM(completion_tokens),0), COALESCE(SUM(total_tokens),0)
		FROM llm_usage WHERE `+where, args...).Scan(
		&totals.Calls, &totals.PromptTokens, &totals.CompletionTokens, &totals.TotalTokens,
	)
	if err != nil {
		return nil, fmt.Errorf("llm usage totals: %w", err)
	}
	summary.Totals = totals

	// Per-provider breakdown.
	provQuery := `
		SELECT provider_id, model, COUNT(*), COALESCE(SUM(prompt_tokens),0),
		       COALESCE(SUM(completion_tokens),0), COALESCE(SUM(total_tokens),0)
		FROM llm_usage WHERE ` + where + `
		GROUP BY provider_id, model
		ORDER BY SUM(total_tokens) DESC`
	rows, err := r.db.QueryContext(ctx, provQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("llm usage by provider: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p model.LLMUsageProviderSummary
		if err := rows.Scan(&p.ProviderID, &p.Model, &p.Calls, &p.PromptTokens, &p.CompletionTokens, &p.TotalTokens); err != nil {
			return nil, fmt.Errorf("scan llm usage provider row: %w", err)
		}
		summary.ByProvider = append(summary.ByProvider, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate llm usage provider rows: %w", err)
	}

	// Per-day series. DATE_FORMAT keeps the wire value a plain
	// YYYY-MM-DD string, consistent with the API contract.
	dayQuery := `
		SELECT DATE_FORMAT(created_at, '%Y-%m-%d'), COUNT(*), COALESCE(SUM(total_tokens),0)
		FROM llm_usage WHERE ` + where + `
		GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')
		ORDER BY DATE_FORMAT(created_at, '%Y-%m-%d') ASC`
	dayRows, err := r.db.QueryContext(ctx, dayQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("llm usage by day: %w", err)
	}
	defer dayRows.Close()
	for dayRows.Next() {
		var d model.LLMUsageDaySummary
		if err := dayRows.Scan(&d.Date, &d.Calls, &d.TotalTokens); err != nil {
			return nil, fmt.Errorf("scan llm usage day row: %w", err)
		}
		summary.ByDay = append(summary.ByDay, d)
	}
	if err := dayRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate llm usage day rows: %w", err)
	}

	return summary, nil
}
