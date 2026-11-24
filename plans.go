package corvonfse

import (
	"context"
	"net/http"
	"net/url"
)

// Plans is the public price catalog. List and Get need no API key.
type Plans struct {
	t *Transport
}

// List returns one page of plan prices, from the smallest to the largest
// allowance. params is optional (Cycle "annual" returns discounted prices).
// GET /plans — unwrap "none".
func (p *Plans) List(ctx context.Context, params *PlansParams) ([]PlanPrice, Pagination, *CorvoError) {
	var lp *ListParams
	cycle := ""
	if params != nil {
		lp = &params.ListParams
		cycle = params.Cycle
	}
	query := listQuery(lp)
	queryAdd(query, "cycle", cycle)
	return requestPage[PlanPrice](ctx, p.t, "/plans", query)
}

// Get returns the price of one plan by slug (for example solo). cycle is
// "monthly" (default, when empty) or "annual". GET /plans/:slug — unwrap "data".
func (p *Plans) Get(ctx context.Context, slug string, cycle string) (PlanPrice, *CorvoError) {
	query := url.Values{}
	queryAdd(query, "cycle", cycle)
	return requestJSON[PlanPrice](ctx, p.t, requestSpec{
		method: http.MethodGet,
		path:   "/plans/" + url.PathEscape(slug),
		query:  query,
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
