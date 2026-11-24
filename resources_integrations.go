package corvonfse

import (
	"context"
	"net/http"
	"net/url"
)

// Integrations lists the payment platforms Corvo can connect to.
type Integrations struct {
	t *Transport
}

// ListProviders returns the supported platforms (not paginated).
// GET /integrations/providers — unwrap "data".
func (i *Integrations) ListProviders(ctx context.Context) ([]IntegrationProvider, *CorvoError) {
	return requestJSON[[]IntegrationProvider](ctx, i.t, requestSpec{
		method: http.MethodGet,
		path:   "/integrations/providers",
	})
}

// IntegrationConnections manages connections to payment platforms. A
// connection follows the environment of the key that created it: a test key
// does not see live connections.
type IntegrationConnections struct {
	t *Transport
}

func connectionPath(id string) string {
	return "/integrations/connections/" + url.PathEscape(id)
}

func (c *IntegrationConnections) action(ctx context.Context, id string, method string, suffix string) (IntegrationConnection, *CorvoError) {
	return requestJSON[IntegrationConnection](ctx, c.t, requestSpec{
		method: method,
		path:   connectionPath(id) + suffix,
	})
}

// List returns the connections (not paginated).
// GET /integrations/connections — unwrap "data".
func (c *IntegrationConnections) List(ctx context.Context) ([]IntegrationConnection, *CorvoError) {
	return requestJSON[[]IntegrationConnection](ctx, c.t, requestSpec{
		method: http.MethodGet,
		path:   "/integrations/connections",
	})
}

// Create creates a connection; the result carries the secrets, shown once.
// POST /integrations/connections — unwrap "data".
func (c *IntegrationConnections) Create(ctx context.Context, params IntegrationConnectionCreateParams) (IntegrationConnectionCreated, *CorvoError) {
	return requestJSON[IntegrationConnectionCreated](ctx, c.t, requestSpec{
		method:   http.MethodPost,
		path:     "/integrations/connections",
		jsonBody: params,
	})
}

// Get reads a connection. GET /integrations/connections/:id — unwrap "data".
func (c *IntegrationConnections) Get(ctx context.Context, id string) (IntegrationConnection, *CorvoError) {
	return c.action(ctx, id, http.MethodGet, "")
}

// Validate checks the credentials and identifies the external account.
// POST /integrations/connections/:id/validate — unwrap "data".
func (c *IntegrationConnections) Validate(ctx context.Context, id string) (IntegrationConnection, *CorvoError) {
	return c.action(ctx, id, http.MethodPost, "/validate")
}

// UpdateCredentials replaces the credentials (for example a rotated API key).
// PUT /integrations/connections/:id/credentials — unwrap "data".
func (c *IntegrationConnections) UpdateCredentials(ctx context.Context, id string, credentials map[string]string) (IntegrationConnection, *CorvoError) {
	return requestJSON[IntegrationConnection](ctx, c.t, requestSpec{
		method: http.MethodPut,
		path:   connectionPath(id) + "/credentials",
		jsonBody: struct {
			Credentials map[string]string `json:"credentials"`
		}{Credentials: credentials},
	})
}

// Activate starts issuing invoices automatically.
// POST /integrations/connections/:id/activate — unwrap "data".
func (c *IntegrationConnections) Activate(ctx context.Context, id string) (IntegrationConnection, *CorvoError) {
	return c.action(ctx, id, http.MethodPost, "/activate")
}

// Pause stops issuing; events are still recorded.
// POST /integrations/connections/:id/pause — unwrap "data".
func (c *IntegrationConnections) Pause(ctx context.Context, id string) (IntegrationConnection, *CorvoError) {
	return c.action(ctx, id, http.MethodPost, "/pause")
}

// Disconnect disconnects for good.
// POST /integrations/connections/:id/disconnect — unwrap "data".
func (c *IntegrationConnections) Disconnect(ctx context.Context, id string) (IntegrationConnection, *CorvoError) {
	return c.action(ctx, id, http.MethodPost, "/disconnect")
}

// ListRules returns the rules of a connection.
// GET /integrations/connections/:id/rules — unwrap "data".
func (c *IntegrationConnections) ListRules(ctx context.Context, id string) ([]IntegrationRule, *CorvoError) {
	return requestJSON[[]IntegrationRule](ctx, c.t, requestSpec{
		method: http.MethodGet,
		path:   connectionPath(id) + "/rules",
	})
}

// ReplaceRules replaces all the rules of a connection.
// PUT /integrations/connections/:id/rules — unwrap "data".
func (c *IntegrationConnections) ReplaceRules(ctx context.Context, id string, rules []IntegrationRuleInput) ([]IntegrationRule, *CorvoError) {
	return requestJSON[[]IntegrationRule](ctx, c.t, requestSpec{
		method: http.MethodPut,
		path:   connectionPath(id) + "/rules",
		jsonBody: struct {
			Rules []IntegrationRuleInput `json:"rules"`
		}{Rules: rules},
	})
}

// ListEvents returns the latest events received (limit 1 to 100, default 50;
// 0 uses the default). GET /integrations/connections/:id/events — unwrap "data".
func (c *IntegrationConnections) ListEvents(ctx context.Context, id string, limit int) ([]IntegrationEvent, *CorvoError) {
	query := url.Values{}
	queryAddInt(query, "limit", limit)
	return requestJSON[[]IntegrationEvent](ctx, c.t, requestSpec{
		method: http.MethodGet,
		path:   connectionPath(id) + "/events",
		query:  query,
	})
}

// ListOperations returns the latest operations (sales) and their invoice
// status. limit 0 uses the default.
// GET /integrations/connections/:id/operations — unwrap "data".
func (c *IntegrationConnections) ListOperations(ctx context.Context, id string, limit int) ([]IntegrationOperation, *CorvoError) {
	query := url.Values{}
	queryAddInt(query, "limit", limit)
	return requestJSON[[]IntegrationOperation](ctx, c.t, requestSpec{
		method: http.MethodGet,
		path:   connectionPath(id) + "/operations",
		query:  query,
	})
}

// IntegrationOperations acts on the operations (sales) of a connection.
type IntegrationOperations struct {
	t *Transport
}

// CompleteCustomer fills in the customer of an operation waiting for it.
// PATCH /integrations/operations/:id/customer — unwrap "data".
func (o *IntegrationOperations) CompleteCustomer(ctx context.Context, id string, params IntegrationCustomerParams) (IntegrationOperation, *CorvoError) {
	return requestJSON[IntegrationOperation](ctx, o.t, requestSpec{
		method:   http.MethodPatch,
		path:     "/integrations/operations/" + url.PathEscape(id) + "/customer",
		jsonBody: params,
	})
}

// Retry tries to issue the invoice of an operation again.
// POST /integrations/operations/:id/retry — unwrap "data".
func (o *IntegrationOperations) Retry(ctx context.Context, id string) (IntegrationOperationRetry, *CorvoError) {
	return requestJSON[IntegrationOperationRetry](ctx, o.t, requestSpec{
		method: http.MethodPost,
		path:   "/integrations/operations/" + url.PathEscape(id) + "/retry",
	})
}
