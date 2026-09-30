// Package llmprovider defines the bounds that an LLM provider's
// max_tokens setting must respect.
//
// The setting is an output budget: the largest number of tokens a single
// request may generate. It is not the model's context window. Every
// request must fit input plus output inside the context window, so an
// output budget that exceeds the window makes each request unsendable —
// the provider rejects it before generating anything, and no amount of
// retrying helps.
//
// Enforcing the bounds here, next to the setting they govern, keeps the
// ceiling from drifting between the write path (which validates) and the
// send path (which clamps).
package llmprovider

// DefaultMaxTokens is the output budget applied when a provider does not
// specify one.
//
// Extracting the node/edge graph from a full research paper needs room:
// a 15-page paper yields roughly 20 nodes and 25 edges, and models write
// JSON far more verbosely than the prompt's "brief description" suggests,
// measuring around 7k completion tokens. A 4096 budget truncates that
// response mid-JSON and the extraction fails on every paper of that size.
// 16384 leaves several times that headroom.
const DefaultMaxTokens = 16384

// MaxTokensLimit is the largest output budget a provider may be
// configured with.
//
// The value bounds a single response, so it must stay well under any
// plausible context window: the discovery service sends up to
// maxInputChars (24000) of paper text, and input plus output has to fit
// alongside the window. 32768 keeps the total inside the 32k-class
// windows that are the smallest in common use, while leaving ample room
// above DefaultMaxTokens.
const MaxTokensLimit = 32768

// ClampMaxTokens forces a stored max_tokens value inside [1, MaxTokensLimit],
// falling back to DefaultMaxTokens for non-positive input. The send path
// applies this so a row outside the range can never produce an
// unsendable request; the write path rejects such values instead so bad
// configuration is surfaced rather than silently stored.
func ClampMaxTokens(n int) int {
	if n <= 0 {
		return DefaultMaxTokens
	}
	if n > MaxTokensLimit {
		return MaxTokensLimit
	}
	return n
}
