package extractor

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateInputRejectsUnusable verifies the guards that reject text
// too short or carrying a no-content placeholder.
func TestValidateInputRejectsUnusable(t *testing.T) {
	cases := []struct {
		name string
		text string
		want error
	}{
		{"empty", "", ErrInputUnusable},
		{"whitespace", "   \n\t  ", ErrInputUnusable},
		{"too short", "short", ErrInputUnusable},
		{
			"placeholder case-insensitive",
			"(PDF uploaded — text extraction pending)",
			ErrInputUnusable,
		},
		{
			"placeholder mid text",
			"Some real intro... \n(PDF uploaded — text extraction pending)",
			ErrInputUnusable,
		},
		{
			"valid long text",
			strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInput(tc.text)
			if !errors.Is(err, tc.want) {
				t.Errorf("validateInput(%q) err = %v, want %v", tc.text, err, tc.want)
			}
		})
	}
}

// TestTruncateTextHeadTail verifies truncation keeps the head and tail,
// omits the middle, and is identity for short input.
func TestTruncateTextHeadTail(t *testing.T) {
	// Generous input well over the cap.
	big := strings.Repeat("abcdefgh", 5000) // 40000 chars
	truncated, omitted := truncateText(big, 24000)
	if omitted <= 0 {
		t.Fatalf("omitted = %d, want > 0", omitted)
	}
	if len([]rune(truncated)) > 24000+64 {
		t.Errorf("truncated length = %d, exceeds cap (+marker)", len([]rune(truncated)))
	}
	// Head preserved verbatim.
	if !strings.HasPrefix(truncated, "abcdefgh") {
		t.Error("truncated text lost its head")
	}
	// Tail preserved verbatim.
	if !strings.HasSuffix(truncated, "abcdefgh") {
		t.Error("truncated text lost its tail")
	}
	// Short input passes through unchanged.
	original := "small"
	got, o := truncateText(original, 24000)
	if got != original || o != 0 {
		t.Errorf("truncateText(short) = (%q, %d), want (%q, 0)", got, o, original)
	}
}
