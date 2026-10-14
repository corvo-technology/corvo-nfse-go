package corvonfse

import (
	"context"
	"net/http"
	"net/url"
)

// Companies manages fiscal companies (emitentes).
type Companies struct {
	t *Transport
}

func listQuery(q *ListParams) url.Values {
	query := url.Values{}
	if q == nil {
		return query
	}
	queryAddInt(query, "limit", q.Limit)
	queryAdd(query, "cursor", q.Cursor)
	return query
}

// Create registers a company. POST /companies — unwrap "data".
func (c *Companies) Create(ctx context.Context, params CompanyCreateParams) (Company, *CorvoError) {
	return requestJSON[Company](ctx, c.t, requestSpec{
		method:   http.MethodPost,
		path:     "/companies",
		jsonBody: params,
	})
}

// Get reads a company by ID. GET /companies/:id — unwrap "data".
func (c *Companies) Get(ctx context.Context, id string) (Company, *CorvoError) {
	return requestJSON[Company](ctx, c.t, requestSpec{
		method: http.MethodGet,
		path:   "/companies/" + url.PathEscape(id),
	})
}

// Update patches the given fields of a company. Unset fields are omitted
// and left unchanged server-side. PATCH /companies/:id — unwrap "data".
func (c *Companies) Update(ctx context.Context, id string, params CompanyUpdateParams) (Company, *CorvoError) {
	return requestJSON[Company](ctx, c.t, requestSpec{
		method:   http.MethodPatch,
		path:     "/companies/" + url.PathEscape(id),
		jsonBody: params,
	})
}

// Remove deletes a company. DELETE /companies/:id — 204 on success.
func (c *Companies) Remove(ctx context.Context, id string) *CorvoError {
	_, err := requestJSON[any](ctx, c.t, requestSpec{
		method: http.MethodDelete,
		path:   "/companies/" + url.PathEscape(id),
	})
	return err
}

// List returns one page of companies plus its pagination.
// GET /companies — unwrap "none".
func (c *Companies) List(ctx context.Context, params *ListParams) ([]Company, Pagination, *CorvoError) {
	return requestPage[Company](ctx, c.t, "/companies", listQuery(params), false)
}

// Iterate walks every page of companies, calling fn per item until it
// returns false or all pages are consumed. A page failure — including the
// first — is returned as *CorvoError.
func (c *Companies) Iterate(ctx context.Context, params *ListParams, fn func(Company) bool) *CorvoError {
	return iteratePages[Company](ctx, c.t, "/companies", listQuery(params), false, fn)
}
