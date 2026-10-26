package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

// panicMessage runs fn and returns the recovered panic message ("" if it
// did not panic). NewClient panics only on misconfiguration.
func panicMessage(fn func()) (message string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if text, ok := recovered.(string); ok {
				message = text
			}
		}
	}()
	fn()
	return ""
}

func TestNewClientRequiresAPIKey(t *testing.T) {
	t.Setenv("CORVO_API_KEY", "")

	message := panicMessage(func() {
		corvonfse.NewClient(corvonfse.Options{})
	})
	if !strings.Contains(message, "apiKey") {
		t.Fatalf("expected panic mentioning apiKey, got %q", message)
	}
}

func TestNewClientRejectsPublicHTTP(t *testing.T) {
	message := panicMessage(func() {
		corvonfse.NewClient(corvonfse.Options{
			APIKey:  apiKey,
			BaseURL: "http://api.exemplo.com",
		})
	})
	if !strings.Contains(message, "AllowInsecureHTTP") {
		t.Fatalf("expected panic mentioning AllowInsecureHTTP, got %q", message)
	}
}

func TestNewClientRejectsInvalidBaseURL(t *testing.T) {
	message := panicMessage(func() {
		corvonfse.NewClient(corvonfse.Options{APIKey: apiKey, BaseURL: "not-a-url"})
	})
	if !strings.Contains(message, "invalid baseUrl") {
		t.Fatalf("expected panic mentioning invalid baseUrl, got %q", message)
	}
}

func TestNewClientAcceptsPrivateHTTP(t *testing.T) {
	for _, baseURL := range []string{
		"http://localhost:4000",
		"http://127.0.0.1:3000",
		"http://10.1.2.3",
		"http://192.168.0.10",
		"http://172.16.0.9",
		"http://169.254.1.1",
	} {
		baseURL := baseURL
		message := panicMessage(func() {
			corvonfse.NewClient(corvonfse.Options{APIKey: apiKey, BaseURL: baseURL})
		})
		if message != "" {
			t.Errorf("expected %s to be accepted, got panic %q", baseURL, message)
		}
	}
}

func TestNewClientAllowInsecureHTTP(t *testing.T) {
	message := panicMessage(func() {
		corvonfse.NewClient(corvonfse.Options{
			APIKey:            apiKey,
			BaseURL:           "http://api.exemplo.com",
			AllowInsecureHTTP: true,
		})
	})
	if message != "" {
		t.Fatalf("expected http:// public host to be accepted with AllowInsecureHTTP, got %q", message)
	}
}

func TestNewClientUsesEnvironmentFallbacks(t *testing.T) {
	var call *capture
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := record(r)
		call = &captured
		writeJSON(w, http.StatusOK, `{"status":"ok"}`)
	}))
	t.Cleanup(server.Close)

	t.Setenv("CORVO_API_KEY", apiKey)
	t.Setenv("CORVO_BASE_URL", server.URL)

	client := corvonfse.NewClient(corvonfse.Options{})
	if _, err := client.Health.Live(context.Background()); err != nil {
		t.Fatalf("health.live via env-configured client: %v", err)
	}
	if call == nil {
		t.Fatal("expected a request to reach the env-configured server")
	}
	if call.header.Get("X-API-Key") != apiKey {
		t.Errorf("expected CORVO_API_KEY to be sent, got %q", call.header.Get("X-API-Key"))
	}
}

func TestNewClientDefaultsToPublicBaseURL(t *testing.T) {
	var requested string
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			requested = r.URL.String()
			return cannedResponse(http.StatusOK, `{"status":"ok"}`), nil
		}),
	}
	client := corvonfse.NewClient(corvonfse.Options{
		APIKey:     apiKey,
		HTTPClient: httpClient,
	})
	if _, err := client.Health.Ready(context.Background()); err != nil {
		t.Fatalf("health.ready: %v", err)
	}
	if requested != "https://api.corvo.tech/v1/health/ready" {
		t.Errorf("expected default base URL request, got %q", requested)
	}
}

func TestHealthLiveReturnsRawObject(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health/live" {
			t.Errorf("expected path /v1/health/live, got %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, `{"status":"ok","timestamp":"2026-10-13T00:00:00Z","uptime":12.5}`)
	}))

	data, err := client.Health.Live(context.Background())
	if err != nil {
		t.Fatalf("health.live: %v", err)
	}
	payload, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", data)
	}
	if payload["status"] != "ok" {
		t.Errorf("expected status ok, got %v", payload["status"])
	}
}

func TestVersionAndWebhookEvents(t *testing.T) {
	if corvonfse.Version != "0.2.0" {
		t.Errorf("expected Version 0.2.0, got %q", corvonfse.Version)
	}
	if len(corvonfse.WebhookEvents) != 9 {
		t.Fatalf("expected 9 webhook events, got %d", len(corvonfse.WebhookEvents))
	}
	if corvonfse.WebhookEvents[8] != "nfse.invoice.fiscal_discrepancy" {
		t.Errorf("expected last event nfse.invoice.fiscal_discrepancy, got %q", corvonfse.WebhookEvents[8])
	}
}
