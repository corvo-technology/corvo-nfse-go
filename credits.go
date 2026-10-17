package corvonfse

import (
	"context"
	"net/http"
)

// Credits reports the tenant credit balance, purchasable plans and the
// credit ledger.
type Credits struct {
	t *Transport
}

// Balance returns the current credit balance. GET /credits — unwrap "data".
func (c *Credits) Balance(ctx context.Context) (CreditBalance, *CorvoError) {
	return requestJSON[CreditBalance](ctx, c.t, requestSpec{
		method: http.MethodGet,
		path:   "/credits",
	})
}

// Plans returns one page of credit plans.
// GET /credits/plans — unwrap "none".
func (c *Credits) Plans(ctx context.Context, params *ListParams) ([]CreditPlan, Pagination, *CorvoError) {
	return requestPage[CreditPlan](ctx, c.t, "/credits/plans", listQuery(params))
}

// History returns one page of credit transactions, newest first.
// GET /credits/history — unwrap "none".
func (c *Credits) History(ctx context.Context, params *ListParams) ([]CreditTransaction, Pagination, *CorvoError) {
	return requestPage[CreditTransaction](ctx, c.t, "/credits/history", listQuery(params))
}

// IterateHistory walks every page of the credit ledger, calling fn per
// entry until it returns false or all pages are consumed. A page failure —
// including the first — is returned as *CorvoError.
func (c *Credits) IterateHistory(ctx context.Context, params *ListParams, fn func(CreditTransaction) bool) *CorvoError {
	return iteratePages[CreditTransaction](ctx, c.t, "/credits/history", listQuery(params), fn)
}
