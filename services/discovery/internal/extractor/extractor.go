// Package extractor implements the paper analysis pipeline. It sends the
// paper text to an LLM with a structured prompt, parses the JSON response
// into nodes and edges, and returns an ExtractionResult.
package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"discovery/internal/llm"
	"discovery/internal/model"
)

// Service orchestrates the extraction pipeline.
type Service struct{}

// NewService creates a new extractor Service.
func NewService() *Service {
	return &Service{}
}

// Extract sends the paper text to the LLM and returns structured nodes/edges.
func (s *Service) Extract(ctx context.Context, client *llm.Client, paper *model.Paper) (*model.ExtractionResult, error) {
	prompt := buildPrompt(paper)

	slog.Info("extraction_started",
		slog.String("paper_id", paper.ID),
		slog.String("title", paper.Title),
		slog.Int("text_length", len(paper.RawText)),
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
	)

	result, err := parseResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}

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

func buildPrompt(paper *model.Paper) string {
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
	sb.WriteString(paper.RawText)
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
