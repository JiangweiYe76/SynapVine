package repository

import (
	"context"
	"testing"
	"time"

	"core/internal/model"
	"core/internal/testutil"
)

// TestLLMUsageInsertAndSummary covers usage recording and the
// aggregated summary (totals, per-provider breakdown, per-day series).
func TestLLMUsageInsertAndSummary(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewLLMUsageRepository(conn)

	if _, err := conn.Exec("DELETE FROM llm_usage"); err != nil {
		t.Fatalf("clean llm_usage: %v", err)
	}

	events := []model.LLMUsageRecordRequest{
		{PaperID: "p1", ProviderID: "prov-a", Model: "gpt-a", PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
		{PaperID: "p2", ProviderID: "prov-a", Model: "gpt-a", PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		{PaperID: "p3", ProviderID: "prov-b", Model: "gpt-b", PromptTokens: 200, CompletionTokens: 100, TotalTokens: 300},
	}
	for _, e := range events {
		if err := repo.Insert(context.Background(), &e); err != nil {
			t.Fatalf("insert usage: %v", err)
		}
	}

	summary, err := repo.Summary(context.Background(), UsageFilter{})
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	if summary.Totals.Calls != 3 || summary.Totals.TotalTokens != 465 ||
		summary.Totals.PromptTokens != 310 || summary.Totals.CompletionTokens != 155 {
		t.Errorf("totals = %+v, want 3 calls / 310 / 155 / 465", summary.Totals)
	}

	if len(summary.ByProvider) != 2 {
		t.Fatalf("by_provider = %d entries, want 2", len(summary.ByProvider))
	}
	// Ordered by total tokens descending: prov-b (300) first.
	if summary.ByProvider[0].ProviderID != "prov-b" || summary.ByProvider[0].TotalTokens != 300 {
		t.Errorf("by_provider[0] = %+v, want prov-b / 300", summary.ByProvider[0])
	}
	if summary.ByProvider[1].ProviderID != "prov-a" || summary.ByProvider[1].Calls != 2 || summary.ByProvider[1].TotalTokens != 165 {
		t.Errorf("by_provider[1] = %+v, want prov-a / 2 calls / 165", summary.ByProvider[1])
	}

	if len(summary.ByDay) != 1 {
		t.Fatalf("by_day = %d entries, want 1", len(summary.ByDay))
	}
	wantDate := time.Now().Format("2006-01-02")
	if summary.ByDay[0].Date != wantDate {
		t.Errorf("by_day date = %q, want %q", summary.ByDay[0].Date, wantDate)
	}
	if summary.ByDay[0].Calls != 3 || summary.ByDay[0].TotalTokens != 465 {
		t.Errorf("by_day = %+v, want 3 calls / 465 tokens", summary.ByDay[0])
	}
}

// TestLLMUsageSummaryDateFilter verifies the from/to window bounds.
func TestLLMUsageSummaryDateFilter(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewLLMUsageRepository(conn)

	if _, err := conn.Exec("DELETE FROM llm_usage"); err != nil {
		t.Fatalf("clean llm_usage: %v", err)
	}

	e := model.LLMUsageRecordRequest{PaperID: "p1", ProviderID: "prov-a", Model: "gpt-a", TotalTokens: 10}
	if err := repo.Insert(context.Background(), &e); err != nil {
		t.Fatalf("insert usage: %v", err)
	}

	// A window ending today (exclusive tomorrow) must include the row.
	today := time.Now().Truncate(24 * time.Hour)
	from := today
	to := today.AddDate(0, 0, 1)
	summary, err := repo.Summary(context.Background(), UsageFilter{From: &from, To: &to})
	if err != nil {
		t.Fatalf("summary with window: %v", err)
	}
	if summary.Totals.Calls != 1 {
		t.Errorf("calls in window = %d, want 1", summary.Totals.Calls)
	}

	// A window ending yesterday must exclude it.
	yesterdayTo := today
	summary, err = repo.Summary(context.Background(), UsageFilter{To: &yesterdayTo})
	if err != nil {
		t.Fatalf("summary with past window: %v", err)
	}
	if summary.Totals.Calls != 0 {
		t.Errorf("calls in past window = %d, want 0", summary.Totals.Calls)
	}
}

// TestLLMUsageInsertValidation verifies required-field enforcement.
func TestLLMUsageInsertValidation(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewLLMUsageRepository(conn)

	err := repo.Insert(context.Background(), &model.LLMUsageRecordRequest{Model: "gpt-a"})
	if err == nil {
		t.Fatal("insert without paper_id/provider_id should fail")
	}
}
