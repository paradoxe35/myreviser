package ai

import (
	"errors"
	"strings"
	"sync"
)

// ReasoningStyle is how a provider wants to be told to think less.
//
// There is no portable parameter, and getting it wrong is not silently ignored: OpenAI returns 400
// for reasoning_effort on a model that does not reason, and OpenRouter returns 400 if it sees both
// its own shape and OpenAI's. So the shape is chosen per provider and, crucially, is recoverable.
type ReasoningStyle int

const (
	ReasoningOpenAIEffort ReasoningStyle = iota
	ReasoningOpenRouter
)

// ReasoningAware is a provider with a reasoning parameter to send. Anthropic has none — its
// thinking is opt-in, and a correction does not want it.
type ReasoningAware interface {
	SetLowReasoning(low bool)
}

// DetectReasoningStyle picks OpenRouter out by host rather than by provider type, because a custom
// provider can point at it too and it is the one OpenAI-compatible gateway with its own shape.
func DetectReasoningStyle(baseURL string) ReasoningStyle {
	if strings.Contains(strings.ToLower(baseURL), "openrouter.ai") {
		return ReasoningOpenRouter
	}
	return ReasoningOpenAIEffort
}

// rejected remembers which endpoint/model pairs refused a reasoning parameter, so the wasted round
// trip is paid once rather than on every correction.
//
// Process-scoped on purpose: a model's capabilities change under the same name, and a stale "this
// does not work" persisted to disk would be far more annoying than one extra request per launch.
var rejected sync.Map

// withReasoningFallback runs send with the reasoning parameter, and once without it if the provider
// rejects the request.
//
// The retry keys on the status rather than the message, because every provider words it
// differently and a match that misses just surfaces a confusing 400. The rejection is remembered
// only when dropping the parameter actually helped: a 400 has many causes — an unknown model, a
// malformed body, an exhausted quota — and caching on the status alone would let any of them switch
// reasoning off for the rest of the session on a model that never objected.
func withReasoningFallback(endpoint, model string, wanted bool, send func(includeReasoning bool) (string, error)) (string, error) {
	key := endpoint + "::" + model
	if _, refused := rejected.Load(key); !wanted || refused {
		return send(false)
	}

	result, err := send(true)
	if err == nil || !refusedReasoning(err) {
		return result, err
	}

	// If this fails too, the parameter was never the problem and nothing is cached.
	result, retryErr := send(false)
	if retryErr != nil {
		return "", retryErr
	}

	rejected.Store(key, struct{}{})
	return result, nil
}

func refusedReasoning(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == 400 || apiErr.StatusCode == 422
}
