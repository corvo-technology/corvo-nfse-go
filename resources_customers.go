package corvonfse

import (
	"context"
	"net/http"
	"net/url"
)

// CNPJ looks up the registration data of a CNPJ.
type CNPJ struct {
	t *Transport
}

// Lookup returns the registration data of a CNPJ (digits or formatted). It
// fails with CNPJ_INVALID or CNPJ_LOOKUP_UNAVAILABLE.
// GET /cnpj/:cnpj — unwrap "data".
func (c *CNPJ) Lookup(ctx context.Context, cnpj string) (CnpjProfile, *CorvoError) {
	return requestJSON[CnpjProfile](ctx, c.t, requestSpec{
		method: http.MethodGet,
		path:   "/cnpj/" + url.PathEscape(cnpj),
	})
}

// Customers manages the customers saved from accepted invoices.
type Customers struct {
	t *Transport
}

func customerSearchQuery(params *CustomerSearchParams) url.Values {
	query := url.Values{}
	if params == nil {
		return query
	}
	queryAdd(query, "search", params.Search)
	queryAddInt(query, "limit", params.Limit)
	queryAdd(query, "cursor", params.Cursor)
	return query
}

// Search returns one page of saved customers, most recently used first.
// GET /customers — unwrap "none".
func (c *Customers) Search(ctx context.Context, params *CustomerSearchParams) ([]SavedCustomer, Pagination, *CorvoError) {
	return requestPage[SavedCustomer](ctx, c.t, "/customers", customerSearchQuery(params))
}

// Iterate walks every page of saved customers matching params, calling fn per
// item until it returns false or all pages are consumed.
func (c *Customers) Iterate(ctx context.Context, params *CustomerSearchParams, fn func(SavedCustomer) bool) *CorvoError {
	return iteratePages[SavedCustomer](ctx, c.t, "/customers", customerSearchQuery(params), fn)
}

// Get reads a saved customer by CPF or CNPJ (digits).
// GET /customers/:taxId — unwrap "data".
func (c *Customers) Get(ctx context.Context, taxID string) (SavedCustomer, *CorvoError) {
	return requestJSON[SavedCustomer](ctx, c.t, requestSpec{
		method: http.MethodGet,
		path:   "/customers/" + url.PathEscape(taxID),
	})
}

// Update corrects the name or the e-mail of a saved customer.
// PATCH /customers/:taxId — unwrap "data".
func (c *Customers) Update(ctx context.Context, taxID string, params CustomerUpdateParams) (SavedCustomer, *CorvoError) {
	return requestJSON[SavedCustomer](ctx, c.t, requestSpec{
		method:   http.MethodPatch,
		path:     "/customers/" + url.PathEscape(taxID),
		jsonBody: params,
	})
}

// Remove forgets a saved customer; issued invoices are not changed.
// DELETE /customers/:taxId — unwrap "data" (null).
func (c *Customers) Remove(ctx context.Context, taxID string) *CorvoError {
	_, err := requestJSON[any](ctx, c.t, requestSpec{
		method: http.MethodDelete,
		path:   "/customers/" + url.PathEscape(taxID),
	})
	return err
}

// ServiceFavorites manages the favorite services of a company.
type ServiceFavorites struct {
	t *Transport
}

// List returns the favorite services of a company (not paginated).
// GET /companies/:companyId/service-favorites — unwrap "data".
func (s *ServiceFavorites) List(ctx context.Context, companyID string) ([]ServiceFavorite, *CorvoError) {
	return requestJSON[[]ServiceFavorite](ctx, s.t, requestSpec{
		method: http.MethodGet,
		path:   "/companies/" + url.PathEscape(companyID) + "/service-favorites",
	})
}

// Create saves a favorite service.
// POST /companies/:companyId/service-favorites — unwrap "data".
func (s *ServiceFavorites) Create(ctx context.Context, companyID string, params ServiceFavoriteCreateParams) (ServiceFavorite, *CorvoError) {
	return requestJSON[ServiceFavorite](ctx, s.t, requestSpec{
		method:   http.MethodPost,
		path:     "/companies/" + url.PathEscape(companyID) + "/service-favorites",
		jsonBody: params,
	})
}

// Remove deletes a favorite service and reports whether the API acknowledged
// it. DELETE /companies/:companyId/service-favorites/:favoriteId — unwrap "data".
func (s *ServiceFavorites) Remove(ctx context.Context, companyID string, favoriteID string) (bool, *CorvoError) {
	result, err := requestJSON[struct {
		Deleted bool `json:"deleted"`
	}](ctx, s.t, requestSpec{
		method: http.MethodDelete,
		path:   "/companies/" + url.PathEscape(companyID) + "/service-favorites/" + url.PathEscape(favoriteID),
	})
	return result.Deleted, err
}
