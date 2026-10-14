package corvonfse

import (
	"context"
	"net/http"
)

// Health exposes the unauthenticated liveness and readiness probes. Their
// responses are raw JSON (no data envelope) and are returned unchanged as
// any.
type Health struct {
	t *Transport
}

// Live reports process liveness: { status, timestamp, uptime }.
// GET /health/live — unwrap "none".
func (h *Health) Live(ctx context.Context) (any, *CorvoError) {
	return requestJSON[any](ctx, h.t, requestSpec{
		method: http.MethodGet,
		path:   "/health/live",
		unwrap: UnwrapNone,
	})
}

// Ready reports dependency readiness (database, storage).
// GET /health/ready — unwrap "none".
func (h *Health) Ready(ctx context.Context) (any, *CorvoError) {
	return requestJSON[any](ctx, h.t, requestSpec{
		method: http.MethodGet,
		path:   "/health/ready",
		unwrap: UnwrapNone,
	})
}
