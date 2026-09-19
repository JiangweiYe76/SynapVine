// Package extractor implements the paper analysis pipeline. It sends the
// paper text to an LLM with a structured prompt, parses the JSON response
// into nodes and edges, and returns an ExtractionResult.
package extractor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"discovery/internal/llm"
	"discovery/internal/model"
)

// ErrInputUnusable reports that the paper text carries too little
// information to be analyzed. Callers should treat this as a terminal
// condition (retrying the same text cannot succeed) rather than a
// transient extraction failure.
var ErrInputUnusable = errors.New("paper text is not usable for extraction")

// maxInputChars bounds the paper text embedded in the prompt. Provider
// context windows are finite while the provider's max_tokens only caps
// the *output*, so an unbounded PDF transcript would either be rejected
// outright (context length exceeded) or crowd out the output budget.
// Tuned conservatively so a 4096-token output still fits inside a 16k
// context window.
const maxInputChars = 24000

// minimumTextLength is the smallest amount of text worth sending to the
// LLM. Below this the model has nothing to ground its output on and
// fabricates concepts instead of extracting them.
const minimumTextLength = 200

// knownPlaceholders are raw_text values written by upstream services when
// real text extraction failed (e.g. scanned/image-only PDFs). They carry
// no analysable content.
var knownPlaceholders = []string{
	"(pdf uploaded",           // services/console fallback for failed PDF extraction
	"text extraction pending", // same fallback, matched on its distinctive fragment
}

// validateInput rejects paper text that cannot yield a meaningful
// extraction. It runs before any LLM call so unusable papers neither
// consume tokens nor produce fabricated nodes in the review queue.
func validateInput(rawText string) error {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return fmt.Errorf("%w: text is empty", ErrInputUnusable)
	}

	lower := strings.ToLower(trimmed)
	for _, placeholder := range knownPlaceholders {
		if strings.Contains(lower, placeholder) {
			return fmt.Errorf("%w: placeholder text detected", ErrInputUnusable)
		}
	}

	if n := len([]rune(trimmed)); n < minimumTextLength {
		return fmt.Errorf("%w: %d characters, minimum is %d", ErrInputUnusable, n, minimumTextLength)
	}
	return nil
}

// truncateText caps s at limit characters, keeping the head and tail and
// dropping the middle. Research papers concentrate their contributions in
// the opening sections and the conclusion, so head+tail preserves more
// signal than a plain prefix cut. It returns the possibly shortened text
// and the number of characters omitted (0 when no truncation happened).
func truncateText(s string, limit int) (string, int) {
	runes := []rune(s)
	if len(runes) <= limit {
		return s, 0
	}

	headLen := limit * 2 / 3
	tailLen := limit - headLen
	omitted := len(runes) - limit

	var sb strings.Builder
	sb.Grow(limit + 64)
	sb.WriteString(string(runes[:headLen]))
	fmt.Fprintf(&sb, "\n\n[... %d characters omitted ...]\n\n", omitted)
	sb.WriteString(string(runes[len(runes)-tailLen:]))
	return sb.String(), omitted
}

// Service orchestrates the extraction pipeline.
type Service struct{}

// NewService creates a new extractor Service.
func NewService() *Service {
	return &Service{}
}

// Extract sends the paper text to the LLM and returns structured nodes/edges.
// It returns ErrInputUnusable when the paper text cannot ground an
// extraction, before spending any tokens.
func (s *Service) Extract(ctx context.Context, client *llm.Client, paper *model.Paper) (*model.ExtractionResult, error) {
	if err := validateInput(paper.RawText); err != nil {
		slog.Warn("extraction_input_rejected",
			slog.String("paper_id", paper.ID),
			slog.String("title", paper.Title),
			slog.Int("text_length", len(paper.RawText)),
			slog.Any("error", err),
		)
		return nil, err
	}

	text, omitted := truncateText(paper.RawText, maxInputChars)
	prompt := buildPrompt(paper, text)

	slog.Info("extraction_started",
		slog.String("paper_id", paper.ID),
		slog.String("title", paper.Title),
		slog.Int("text_length", len(paper.RawText)),
		slog.Int("omitted_chars", omitted),
	)

	resp, err := client.Complete(ctx, llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		JSONMode: true,
	})
	if err != nil {
		return nil, fmt.Errorf("llm call failed: %w", err)
	}

	slog.Info("extraction_llm_responded",
		slog.String("paper_id", paper.ID),
		slog.Int("tokens_used", resp.TokensUsed),
		slog.Int("content_length", len(resp.Content)),
		slog.String("finish_reason", resp.FinishReason),
	)

	// A "length" finish reason means the model ran out of output budget,
	// so the JSON body is truncated and unmarshalling would report a
	// confusing syntax error. Surface the real cause instead.
	if resp.FinishReason == "length" {
		return nil, fmt.Errorf("llm output truncated: hit max_tokens limit (%d)", client.MaxTokens())
	}

	result, err := parseResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}

	// Carry token consumption for usage accounting in core.
	result.PromptTokens = resp.PromptTokens
	result.CompletionTokens = resp.CompletionTokens
	result.TotalTokens = resp.TokensUsed

	slog.Info("extraction_completed",
		slog.String("paper_id", paper.ID),
		slog.Int("nodes", len(result.Nodes)),
		slog.Int("edges", len(result.Edges)),
	)

	return result, nil
}

const systemPrompt = `You are an AI research paper analyst. Your task is to extract AI concepts (nodes) and their relationships (edges) from research papers.

Rules:
- Nodes represent AI concepts, techniques, models, architectures, or algorithms mentioned in the paper
- Edges represent relationships between these concepts (e.g., "improves", "extends", "uses", "replaces", "inspired by")
- Each node must have a concise name (1-3 words), a brief description, and a relevance score (0-10)
- Each edge must reference source and target node names, a relation description, and a weight (0-1)
- Focus on the paper's main contributions and key concepts, not peripheral mentions
- Be precise: avoid vague or overly generic relationships
- Output valid JSON only`

func buildPrompt(paper *model.Paper, text string) string {
	var sb strings.Builder
	sb.WriteString("Analyze the following research paper and extract AI concepts and their relationships.\n\n")
	sb.WriteString("Title: ")
	sb.WriteString(paper.Title)
	sb.WriteString("\n")
	if paper.Authors != "" {
		sb.WriteString("Authors: ")
		sb.WriteString(paper.Authors)
		sb.WriteString("\n")
	}
	sb.WriteString("\n--- Paper Text ---\n")
	sb.WriteString(text)
	sb.WriteString("\n--- End of Paper ---\n\n")
	sb.WriteString(`Return a JSON object with exactly this structure:
{
  "nodes": [
    {"name": "ConceptName", "description": "Brief description", "relevance": 8.5}
  ],
  "edges": [
    {"source": "ConceptA", "target": "ConceptB", "relation": "improves upon", "weight": 0.8}
  ]
}`)
	return sb.String()
}

// parseResponse parses the LLM's JSON response into an ExtractionResult.
func parseResponse(content string) (*model.ExtractionResult, error) {
	// Some LLMs wrap JSON in markdown code fences.
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		start, end := 0, len(lines)
		for i, line := range lines {
			if strings.HasPrefix(line, "```") && i == 0 {
				start = 1
			}
			if strings.HasPrefix(line, "```") && i > 0 {
				end = i
			}
		}
		if start < end {
			content = strings.Join(lines[start:end], "\n")
		}
	}

	var result model.ExtractionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("unmarshal extraction result: %w (raw: %.200s)", err, content)
	}

	// Basic validation.
	if len(result.Nodes) == 0 {
		return nil, fmt.Errorf("llm returned no nodes")
	}
	for i, node := range result.Nodes {
		if node.Name == "" {
			return nil, fmt.Errorf("node %d has empty name", i)
		}
	}
	for i, edge := range result.Edges {
		if edge.Source == "" || edge.Target == "" {
			return nil, fmt.Errorf("edge %d has empty source or target", i)
		}
	}

	return &result, nil
}
