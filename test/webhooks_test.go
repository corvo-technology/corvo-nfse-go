package test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

const webhookJSON = `{
	"id": "46ee05ce-9dd6-47b7-8261-28061ec627c8",
	"url": "https://webhook.site/eb9651dd",
	"events": ["nfse.invoice.authorized", "nfse.invoice.rejected"],
	"active": true,
	"createdAt": "2026-10-11T00:35:38.843Z",
	"updatedAt": "2026-10-11T00:35:38.843Z"
}`

const deliveryJSON = `{
	"id": "84018385-f25d-463d-af23-cea43d0671d2",
	"tenantId": "6247dfd9-0adc-4bf6-aac9-ee9dc1e7ca8b",
	"webhookConfigId": "46ee05ce-9dd6-47b7-8261-28061ec627c8",
	"eventType": "nfse.invoice.email_sent",
	"environment": "test",
	"url": "https://webhook.site/eb9651dd",
	"payload": {"id": "evt_1", "type": "nfse.invoice.email_sent"},
	"statusCode": 200,
	"responseBody": null,
	"errorMessage": null,
	"status": "sent",
	"durationMs": 1535,
	"retryCount": 0,
	"nextRetryAt": null,
	"lastError": null,
	"createdAt": "2026-10-12T13:04:51.575Z"
}`

func TestWebhooksList(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/webhooks" {
			t.Errorf("expected path /v1/webhooks, got %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, `{
			"data": [`+webhookJSON+`],
			"pagination": {"limit": 1, "hasMore": false, "nextCursor": null}
		}`)
	}))

	hooks, pagination, err := client.Webhooks.List(context.Background(), &corvonfse.ListParams{Limit: 1})
	if err != nil {
		t.Fatalf("webhooks.list: %v", err)
	}
	if len(hooks) != 1 || hooks[0].URL != "https://webhook.site/eb9651dd" {
		t.Errorf("expected one webhook, got %+v", hooks)
	}
	if len(hooks[0].Events) != 2 {
		t.Errorf("expected two events, got %+v", hooks[0].Events)
	}
	if pagination.HasMore {
		t.Errorf("expected no more pages, got %+v", pagination)
	}
}

func TestWebhooksCreateReturnsSecretOnce(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusCreated, `{"data": {`+
			`"id": "hook_1", "url": "https://webhook.site/eb9651dd", `+
			`"events": ["nfse.invoice.authorized"], "active": true, `+
			`"createdAt": "2026-10-11T00:35:38.843Z", "updatedAt": "2026-10-11T00:35:38.843Z", `+
			`"secret": "whsec_abc"}}`)
	}))

	hook, err := client.Webhooks.Create(context.Background(), corvonfse.WebhookCreateParams{
		URL:    "https://webhook.site/eb9651dd",
		Events: []corvonfse.WebhookEvent{"nfse.invoice.authorized"},
	})
	if err != nil {
		t.Fatalf("webhooks.create: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/webhooks" {
		t.Errorf("expected POST /v1/webhooks, got %s %s", call.method, call.path)
	}
	if !strings.Contains(call.body, `"url":"https://webhook.site/eb9651dd"`) ||
		!strings.Contains(call.body, `"events":["nfse.invoice.authorized"]`) {
		t.Errorf("expected url and events body, got %s", call.body)
	}
	if hook.Secret != "whsec_abc" {
		t.Errorf("expected secret revealed on create, got %q", hook.Secret)
	}
}

func TestWebhooksRemoveAndTestFlags(t *testing.T) {
	var calls []capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, record(r))
		if strings.HasSuffix(r.URL.Path, "/test") {
			writeJSON(w, http.StatusOK, `{"data": {"sent": true}}`)
			return
		}
		writeJSON(w, http.StatusOK, `{"data": {"deleted": true}}`)
	}))

	deleted, err := client.Webhooks.Remove(context.Background(), "hook_1")
	if err != nil {
		t.Fatalf("webhooks.remove: %v", err)
	}
	if !deleted {
		t.Error("expected deleted true")
	}
	sent, err := client.Webhooks.Test(context.Background(), "hook_1")
	if err != nil {
		t.Fatalf("webhooks.test: %v", err)
	}
	if !sent {
		t.Error("expected sent true")
	}
	if calls[0].method != http.MethodDelete || calls[0].path != "/v1/webhooks/hook_1" {
		t.Errorf("expected DELETE /v1/webhooks/hook_1, got %s %s", calls[0].method, calls[0].path)
	}
	if calls[1].method != http.MethodPost || calls[1].path != "/v1/webhooks/hook_1/test" {
		t.Errorf("expected POST /v1/webhooks/hook_1/test, got %s %s", calls[1].method, calls[1].path)
	}
}

func TestDeliveriesListFiltersWindow(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{
			"data": [`+deliveryJSON+`],
			"pagination": {"limit": 10, "hasMore": false, "nextCursor": null}
		}`)
	}))

	deliveries, _, err := client.Webhooks.Deliveries.List(context.Background(), &corvonfse.DeliveryListParams{
		Hours:       24,
		Environment: corvonfse.EnvironmentTest,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("deliveries.list: %v", err)
	}
	if call.path != "/v1/webhooks/deliveries" {
		t.Errorf("expected deliveries path, got %s", call.path)
	}
	for _, fragment := range []string{"hours=24", "environment=test", "limit=10"} {
		if !strings.Contains(call.query, fragment) {
			t.Errorf("expected query %q to contain %s", call.query, fragment)
		}
	}
	if len(deliveries) != 1 || deliveries[0].StatusCode != 200 || deliveries[0].DurationMS != 1535 {
		t.Errorf("expected one delivery, got %+v", deliveries)
	}
	if deliveries[0].Payload["type"] != "nfse.invoice.email_sent" {
		t.Errorf("expected payload passed through, got %+v", deliveries[0].Payload)
	}
}

func TestDeliveriesStatsAndRetry(t *testing.T) {
	var calls []capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, record(r))
		if strings.HasSuffix(r.URL.Path, "/stats") {
			writeJSON(w, http.StatusOK, `{"data": {"total": 5, "sent": 4, "failed": 1, "pending": 0, "delayed": 0}}`)
			return
		}
		writeJSON(w, http.StatusOK, `{"data": {"retried": true}}`)
	}))

	stats, err := client.Webhooks.Deliveries.Stats(context.Background(), 6, corvonfse.EnvironmentLive)
	if err != nil {
		t.Fatalf("deliveries.stats: %v", err)
	}
	if stats.Total != 5 || stats.Failed != 1 {
		t.Errorf("expected stats decoded, got %+v", stats)
	}
	if calls[0].path != "/v1/webhooks/deliveries/stats" ||
		calls[0].query != "environment=live&hours=6" {
		t.Errorf("expected stats query, got %s?%s", calls[0].path, calls[0].query)
	}

	retried, err := client.Webhooks.Deliveries.Retry(context.Background(), "del_1")
	if err != nil {
		t.Fatalf("deliveries.retry: %v", err)
	}
	if !retried {
		t.Error("expected retried true")
	}
	if calls[1].path != "/v1/webhooks/deliveries/del_1/retry" {
		t.Errorf("expected retry path, got %s", calls[1].path)
	}
}

func TestDeliveriesIterate(t *testing.T) {
	page := 0
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			writeJSON(w, http.StatusOK, `{
				"data": [`+deliveryJSON+`],
				"pagination": {"limit": 1, "hasMore": true, "nextCursor": "d1"}
			}`)
			return
		}
		if r.URL.Query().Get("cursor") != "d1" {
			t.Errorf("expected cursor d1 on second page, got %q", r.URL.RawQuery)
		}
		writeJSON(w, http.StatusOK, `{
			"data": [],
			"pagination": {"limit": 1, "hasMore": false, "nextCursor": null}
		}`)
	}))

	var seen int
	err := client.Webhooks.Deliveries.Iterate(context.Background(), nil, func(delivery corvonfse.WebhookDelivery) bool {
		seen++
		return true
	})
	if err != nil {
		t.Fatalf("deliveries.iterate: %v", err)
	}
	if seen != 1 || page != 2 {
		t.Errorf("expected 1 item over 2 pages, got seen=%d pages=%d", seen, page)
	}
}
