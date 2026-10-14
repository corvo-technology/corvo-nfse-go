// Package test is the black-box test suite for the Corvo NFS-e Go SDK:
// every test drives the public corvonfse API against an httptest server
// (layer 1), the HMAC known vector (layer 2) and, when the environment
// provides credentials, a live instance (test/contract).
package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

// apiKey is a syntactically valid test key (corvo_test_ + 40 hex-like
// characters); it is never sent to a live API.
const apiKey = "corvo_test_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// newTestClient starts an httptest server for handler and returns a
// Client wired to it. The server is closed when the test finishes.
func newTestClient(t *testing.T, handler http.Handler) (*corvonfse.Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := corvonfse.NewClient(corvonfse.Options{
		APIKey:  apiKey,
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})
	return client, server
}

// capture is a snapshot of one outgoing request.
type capture struct {
	method string
	path   string
	query  string
	header http.Header
	body   string
}

// record snapshots an incoming request, draining its body.
func record(r *http.Request) capture {
	var body strings.Builder
	if r.Body != nil {
		_, _ = io.Copy(&body, r.Body)
	}
	return capture{
		method: r.Method,
		path:   r.URL.Path,
		query:  r.URL.RawQuery,
		header: r.Header.Clone(),
		body:   body.String(),
	}
}

// writeJSON writes a JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// requireError fails the test when err is nil and returns it otherwise,
// so callers can assert on CorvoError fields.
func requireError(t *testing.T, err *corvonfse.CorvoError) *corvonfse.CorvoError {
	t.Helper()
	if err == nil {
		t.Fatal("expected *CorvoError, got nil")
	}
	return err
}

// roundTripFunc adapts a function to http.RoundTripper so tests can
// intercept requests without a server (e.g. to observe the default base
// URL).
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// cannedResponse builds a *http.Response for roundTripFunc handlers.
func cannedResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
