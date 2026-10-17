package corvonfse

import (
	"context"
	"net/http"
	"net/url"
)

// Plans is the public price catalog (monthly cycle only).
type Plans struct {
	t *Transport
}

// List returns one page of plan prices, from the smallest to the largest
// allowance. GET /plans — unwrap "none".
func (p *Plans) List(ctx context.Context, params *ListParams) ([]PlanQuote, Pagination, *CorvoError) {
	return requestPage[PlanQuote](ctx, p.t, "/plans", listQuery(params))
}

// Get returns the price of one plan by slug (free, developer, growth, scale or
// volume). GET /plans/:slug — unwrap "data".
func (p *Plans) Get(ctx context.Context, slug string) (PlanQuote, *CorvoError) {
	return requestJSON[PlanQuote](ctx, p.t, requestSpec{
		method: http.MethodGet,
		path:   "/plans/" + url.PathEscape(slug),
	})
}

// Quote recommends a plan for a monthly invoice volume. The result is nil
// (with no error) when the volume is missing or not positive.
// GET /plans/quote — unwrap "data".
func (p *Plans) Quote(ctx context.Context, volume int) (*PlanRecommendation, *CorvoError) {
	query := url.Values{}
	queryAddInt(query, "volume", volume)
	return requestJSON[*PlanRecommendation](ctx, p.t, requestSpec{
		method: http.MethodGet,
		path:   "/plans/quote",
		query:  query,
	})
}
