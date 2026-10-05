package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Reasoning and the final answer share DeepSeek's output token budget.
const deepSeekMaxTokens = 16384

// The current SDK has no top-level thinking field. Add the documented extension
// only to DeepSeek requests, retaining SDK response and error handling.
type deepSeekTransport struct {
	base http.RoundTripper
}

func (t deepSeekTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if err := req.Body.Close(); err != nil {
		return nil, err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	payload["thinking"] = json.RawMessage(`{"type":"enabled"}`)
	body, err = json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	clone := req.Clone(req.Context())
	clone.Body = io.NopCloser(bytes.NewReader(body))
	clone.ContentLength = int64(len(body))
	clone.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	return t.base.RoundTrip(clone)
}
