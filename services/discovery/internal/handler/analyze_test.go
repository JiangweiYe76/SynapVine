package handler

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"discovery/internal/extractor"
	"discovery/internal/pipeline"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// TestAnalyzeErrorResponse_StatusCodes pins the mapping from pipeline
// failure to HTTP status. Unusable paper text is a property of the stored
// data, not a fault in the service, so it must surface as 422 rather than
// 500 — otherwise monitoring reads a retryable server error and callers
// retry a paper that will never become usable.
func TestAnalyzeErrorResponse_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{
			name:       "paper text too short is unprocessable",
			err:        fmt.Errorf("%w: %w", pipeline.ErrExtraction, fmt.Errorf("%w: 40 characters, minimum is 200", extractor.ErrInputUnusable)),
			wantStatus: 422,
			wantError:  "paper_text_unusable",
		},
		{
			name:       "placeholder text is unprocessable",
			err:        fmt.Errorf("%w: %w", pipeline.ErrExtraction, fmt.Errorf("%w: placeholder text detected", extractor.ErrInputUnusable)),
			wantStatus: 422,
			wantError:  "paper_text_unusable",
		},
		{
			name:       "llm failure is a bad gateway",
			err:        fmt.Errorf("%w: %w", pipeline.ErrExtraction, errors.New("llm request failed")),
			wantStatus: 502,
			wantError:  "extraction_failed",
		},
		{
			name:       "paper fetch failure is a bad gateway",
			err:        fmt.Errorf("%w: %w", pipeline.ErrPaperFetch, errors.New("core unavailable")),
			wantStatus: 502,
			wantError:  "paper_fetch_failed",
		},
		{
			name:       "provider fetch failure is a bad gateway",
			err:        fmt.Errorf("%w: %w", pipeline.ErrProviderFetch, errors.New("no default provider")),
			wantStatus: 502,
			wantError:  "provider_fetch_failed",
		},
		{
			name:       "review submit failure is a bad gateway",
			err:        fmt.Errorf("%w: %w", pipeline.ErrReviewSubmit, errors.New("core rejected")),
			wantStatus: 502,
			wantError:  "review_submit_failed",
		},
		{
			name:       "unknown failure is internal",
			err:        errors.New("something else"),
			wantStatus: 500,
			wantError:  "internal_error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			c := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(c)

			if err := analyzeErrorResponse(c, tc.err); err != nil {
				t.Fatalf("analyzeErrorResponse: %v", err)
			}
			if c.Response().StatusCode() != tc.wantStatus {
				t.Errorf("status = %d, want %d", c.Response().StatusCode(), tc.wantStatus)
			}
			if !strings.Contains(string(c.Response().Body()), tc.wantError) {
				t.Errorf("body %q does not contain %q", c.Response().Body(), tc.wantError)
			}
		})
	}
}
