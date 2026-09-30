package llm

import "testing"

// TestClampMaxTokens pins the send-path guardrail: a stored value outside
// the range must degrade to a capped request instead of an unsendable
// one. The bounds mirror core's internal/llmprovider package;
// keep the two in sync.
func TestClampMaxTokens(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  int
	}{
		{name: "zero falls back to default", input: 0, want: DefaultMaxTokens},
		{name: "above limit clamps to limit", input: 1000000, want: MaxTokensLimit},
		{name: "in-range value passes through", input: 8192, want: 8192},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := clampMaxTokens(tc.input); got != tc.want {
				t.Errorf("clampMaxTokens(%d) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
