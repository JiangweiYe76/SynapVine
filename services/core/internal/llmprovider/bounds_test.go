package llmprovider

import "testing"

// TestClampMaxTokens pins the send-path guardrail: a stored value outside
// the range must degrade to a capped request instead of an unsendable one.
func TestClampMaxTokens(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  int
	}{
		{name: "zero falls back to default", input: 0, want: DefaultMaxTokens},
		{name: "negative falls back to default", input: -100, want: DefaultMaxTokens},
		{name: "ordinary value passes through", input: 8192, want: 8192},
		{name: "default passes through", input: DefaultMaxTokens, want: DefaultMaxTokens},
		{name: "limit passes through", input: MaxTokensLimit, want: MaxTokensLimit},
		{name: "above limit clamps to limit", input: 1000000, want: MaxTokensLimit},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClampMaxTokens(tc.input); got != tc.want {
				t.Errorf("ClampMaxTokens(%d) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

// TestBoundsOrder keeps the budget usable: the default must sit strictly
// below the ceiling, otherwise every defaulted provider would be
// rejected or clamped on write.
func TestBoundsOrder(t *testing.T) {
	if DefaultMaxTokens <= 0 || DefaultMaxTokens >= MaxTokensLimit {
		t.Errorf("bounds out of order: default=%d limit=%d", DefaultMaxTokens, MaxTokensLimit)
	}
}
