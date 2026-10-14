package corvonfse

import (
	"context"
	"net/http"
	"net/url"
)

// Webhooks manages webhook endpoints and their deliveries. Create
// returns the signing secret exactly once — verify deliveries with
// VerifySignature.
type Webhooks struct {
	// Deliveries inspects and retries webhook deliveries.
	Deliveries *Deliveries

	t *Transport
}

// Deliveries manages the delivery log of all webhooks.
type Deliveries struct {
	t *Transport
}

func newWebhooks(t *Transport) *Webhooks {
	return &Webhooks{
		Deliveries: &Deliveries{t: t},
		t:          t,
	}
}

// flagResults are the { flag } responses of small POST/DELETE endpoints.
type (
	webhookRemoveResult struct {
		Deleted bool `json:"deleted"`
	}
	webhookTestResult struct {
		Sent bool `json:"sent"`
	}
	deliveryRetryResult struct {
		Retried bool `json:"retried"`
	}
)

// List returns one page of webhook configs.
// GET /webhooks — unwrap "none".
func (w *Webhooks) List(ctx context.Context, params *ListParams) ([]Webhook, Pagination, *CorvoError) {
	return requestPage[Webhook](ctx, w.t, "/webhooks", listQuery(params), false)
}

// Create registers a webhook and returns its signing secret — the only
// time it is revealed. POST /webhooks — unwrap "data".
func (w *Webhooks) Create(ctx context.Context, params WebhookCreateParams) (WebhookWithSecret, *CorvoError) {
	return requestJSON[WebhookWithSecret](ctx, w.t, requestSpec{
		method:   http.MethodPost,
		path:     "/webhooks",
		jsonBody: params,
	})
}

// Remove deletes a webhook config. DELETE /webhooks/:id — unwrap "data",
// reporting whether the API acknowledged the deletion.
func (w *Webhooks) Remove(ctx context.Context, id string) (bool, *CorvoError) {
	result, err := requestJSON[webhookRemoveResult](ctx, w.t, requestSpec{
		method: http.MethodDelete,
		path:   "/webhooks/" + url.PathEscape(id),
	})
	return result.Deleted, err
}

// Test sends a webhook.test event to the endpoint.
// POST /webhooks/:id/test — unwrap "data".
func (w *Webhooks) Test(ctx context.Context, id string) (bool, *CorvoError) {
	result, err := requestJSON[webhookTestResult](ctx, w.t, requestSpec{
		method: http.MethodPost,
		path:   "/webhooks/" + url.PathEscape(id) + "/test",
	})
	return result.Sent, err
}

func deliveryListQuery(params *DeliveryListParams) url.Values {
	query := url.Values{}
	if params == nil {
		return query
	}
	queryAddInt(query, "hours", params.Hours)
	queryAdd(query, "environment", string(params.Environment))
	queryAddInt(query, "limit", params.Limit)
	queryAdd(query, "cursor", params.Cursor)
	return query
}

// List returns one page of deliveries, newest first.
// GET /webhooks/deliveries — unwrap "none".
func (d *Deliveries) List(ctx context.Context, params *DeliveryListParams) ([]WebhookDelivery, Pagination, *CorvoError) {
	return requestPage[WebhookDelivery](ctx, d.t, "/webhooks/deliveries", deliveryListQuery(params), false)
}

// Iterate walks every page of deliveries matching params, calling fn per
// item until it returns false or all pages are consumed. A page failure —
// including the first — is returned as *CorvoError.
func (d *Deliveries) Iterate(ctx context.Context, params *DeliveryListParams, fn func(WebhookDelivery) bool) *CorvoError {
	return iteratePages[WebhookDelivery](ctx, d.t, "/webhooks/deliveries", deliveryListQuery(params), false, fn)
}

// Stats returns the delivery counters for the last hours in the given
// environment (0 hours and "" environment use the API defaults).
// GET /webhooks/deliveries/stats — unwrap "data".
func (d *Deliveries) Stats(ctx context.Context, hours int, environment Environment) (WebhookDeliveryStats, *CorvoError) {
	query := url.Values{}
	queryAddInt(query, "hours", hours)
	queryAdd(query, "environment", string(environment))
	return requestJSON[WebhookDeliveryStats](ctx, d.t, requestSpec{
		method: http.MethodGet,
		path:   "/webhooks/deliveries/stats",
		query:  query,
	})
}

// Retry re-enqueues a failed delivery.
// POST /webhooks/deliveries/:deliveryId/retry — unwrap "data".
func (d *Deliveries) Retry(ctx context.Context, deliveryID string) (bool, *CorvoError) {
	result, err := requestJSON[deliveryRetryResult](ctx, d.t, requestSpec{
		method: http.MethodPost,
		path:   "/webhooks/deliveries/" + url.PathEscape(deliveryID) + "/retry",
	})
	return result.Retried, err
}
