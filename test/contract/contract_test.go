package contract

import (
	"context"
	"os"
	"testing"
	"time"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

// TestLiveInstance runs the shared two-check contract against a live
// instance: health.live responds ok and companies.list honours limit.
// It skips unless CORVO_API_KEY and CORVO_BASE_URL are both set.
func TestLiveInstance(t *testing.T) {
	if os.Getenv("CORVO_API_KEY") == "" || os.Getenv("CORVO_BASE_URL") == "" {
		t.Skip("contract test skipped: set CORVO_API_KEY and CORVO_BASE_URL to run")
	}

	client := corvonfse.NewClient(corvonfse.Options{Timeout: 15 * time.Second})
	ctx := context.Background()

	live, err := client.Health.Live(ctx)
	if err != nil {
		t.Fatalf("health.live: %v", err)
	}
	payload, ok := live.(map[string]any)
	if !ok {
		t.Fatalf("health.live: expected object, got %T", live)
	}
	if payload["status"] != "ok" {
		t.Errorf("health.live: expected status ok, got %v", payload["status"])
	}

	_, pagination, err := client.Companies.List(ctx, &corvonfse.ListParams{Limit: 1})
	if err != nil {
		t.Fatalf("companies.list: %v", err)
	}
	if pagination.Limit != 1 {
		t.Errorf("companies.list: expected pagination.limit 1, got %d", pagination.Limit)
	}
}
