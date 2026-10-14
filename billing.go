package corvonfse

import (
	"context"
	"net/http"
)

// Billing reports the subscription plan and the emission quota.
type Billing struct {
	t *Transport
}

// Plan returns the current plan, subscription and usage.
// GET /billing/plan — unwrap "data".
func (b *Billing) Plan(ctx context.Context) (BillingPlanResponse, *CorvoError) {
	return requestJSON[BillingPlanResponse](ctx, b.t, requestSpec{
		method: http.MethodGet,
		path:   "/billing/plan",
	})
}

// Usage returns the emission quota (canEmit, credits, period).
// GET /billing/usage — unwrap "data".
func (b *Billing) Usage(ctx context.Context) (QuotaStatus, *CorvoError) {
	return requestJSON[QuotaStatus](ctx, b.t, requestSpec{
		method: http.MethodGet,
		path:   "/billing/usage",
	})
}
