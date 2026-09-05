package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func clearReasoningCache() {
	rejected.Range(func(key, _ any) bool {
		rejected.Delete(key)
		return true
	})
}

const okReply = `{"choices":[{"message":{"role":"assistant","content":"corrigé"}}]}`

// bodies records every request body the server saw, so a retry is visible rather than inferred.
func recordingServer(t *testing.T, replies ...func(w http.ResponseWriter)) (*httptest.Server, *[]OpenAIRequest) {
	t.Helper()
	var seen []OpenAIRequest
	attempt := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body OpenAIRequest
		_ = json.Unmarshal(raw, &body)
		seen = append(seen, body)

		reply := replies[len(replies)-1]
		if attempt < len(replies) {
			reply = replies[attempt]
		}
		attempt++
		reply(w)
	}))
	t.Cleanup(server.Close)

	return server, &seen
}

func ok(w http.ResponseWriter) { io.WriteString(w, okReply) }
func badRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, `{"error":{"message":"unsupported parameter"}}`)
}

func provider(t *testing.T, url string, low bool) *OpenAIProvider {
	t.Helper()
	clearReasoningCache()
	p := NewOpenAIProvider("sk-test", url, "gpt-test", 1.0)
	p.LowReasoning = low
	return p
}

func TestLowReasoningSendsEffort(t *testing.T) {
	server, seen := recordingServer(t, ok)

	if _, err := provider(t, server.URL, true).ReviseText(context.Background(), "text", "prompt"); err != nil {
		t.Fatalf("revise: %v", err)
	}

	if len(*seen) != 1 || (*seen)[0].ReasoningEffort != "low" {
		t.Fatalf("expected one request asking for low effort, got %+v", *seen)
	}
}

func TestReasoningIsNotSentWhenNotWanted(t *testing.T) {
	server, seen := recordingServer(t, ok)

	if _, err := provider(t, server.URL, false).ReviseText(context.Background(), "text", "prompt"); err != nil {
		t.Fatalf("revise: %v", err)
	}

	if len(*seen) != 1 || (*seen)[0].ReasoningEffort != "" || (*seen)[0].Reasoning != nil {
		t.Fatalf("expected one plain request, got %+v", *seen)
	}
}

func TestARejectedParameterIsRetriedWithoutIt(t *testing.T) {
	server, seen := recordingServer(t, badRequest, ok)

	result, err := provider(t, server.URL, true).ReviseText(context.Background(), "text", "prompt")
	if err != nil {
		t.Fatalf("revise: %v", err)
	}

	if result != "corrigé" {
		t.Fatalf("expected the retry's result, got %q", result)
	}
	if len(*seen) != 2 || (*seen)[0].ReasoningEffort != "low" || (*seen)[1].ReasoningEffort != "" {
		t.Fatalf("expected a request with the parameter then one without, got %+v", *seen)
	}
}

func TestARejectionIsRememberedForTheNextCall(t *testing.T) {
	server, seen := recordingServer(t, badRequest, ok)
	p := provider(t, server.URL, true)

	if _, err := p.ReviseText(context.Background(), "text", "prompt"); err != nil {
		t.Fatalf("first revise: %v", err)
	}
	if _, err := p.ReviseText(context.Background(), "text", "prompt"); err != nil {
		t.Fatalf("second revise: %v", err)
	}

	// Three requests, not four: the second call skipped the parameter it already knows is refused.
	if len(*seen) != 3 || (*seen)[2].ReasoningEffort != "" {
		t.Fatalf("expected the rejection to be remembered, got %+v", *seen)
	}
}

// A 400 has many causes. Caching on the status alone would switch reasoning off for the session on
// a model that never objected to it.
func TestAPersistentBadRequestIsReportedAndNotCached(t *testing.T) {
	server, seen := recordingServer(t, badRequest, badRequest)
	p := provider(t, server.URL, true)

	if _, err := p.ReviseText(context.Background(), "text", "prompt"); err == nil {
		t.Fatal("expected the error to surface")
	}
	if len(*seen) != 2 {
		t.Fatalf("expected one retry, got %d requests", len(*seen))
	}

	*seen = nil
	server2, seen2 := recordingServer(t, ok)
	p.BaseURL = server2.URL
	if _, err := p.ReviseText(context.Background(), "text", "prompt"); err != nil {
		t.Fatalf("revise: %v", err)
	}
	if (*seen2)[0].ReasoningEffort != "low" {
		t.Fatal("an unrelated 400 disabled reasoning for later calls")
	}
}

func TestAServerErrorIsNotRetried(t *testing.T) {
	server, seen := recordingServer(t, func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error":{"message":"down"}}`)
	})

	if _, err := provider(t, server.URL, true).ReviseText(context.Background(), "text", "prompt"); err == nil {
		t.Fatal("expected the error to surface")
	}
	if len(*seen) != 1 {
		t.Fatalf("a server error should not be retried, got %d requests", len(*seen))
	}
}

func TestOpenRouterGetsItsOwnShape(t *testing.T) {
	if style := DetectReasoningStyle("https://openrouter.ai/api/v1"); style != ReasoningOpenRouter {
		t.Fatalf("expected the OpenRouter shape, got %v", style)
	}
	if style := DetectReasoningStyle("https://api.openai.com/v1"); style != ReasoningOpenAIEffort {
		t.Fatalf("expected the OpenAI shape, got %v", style)
	}
}
