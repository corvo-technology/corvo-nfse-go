package corvonfse

import (
	"context"
	"net/http"
	"net/url"
)

// Invoices issues and serves fiscal invoices (NFS-e).
type Invoices struct {
	t *Transport
}

func invoiceListQuery(params *InvoiceListParams) url.Values {
	query := url.Values{}
	if params == nil {
		return query
	}
	queryAdd(query, "companyId", params.CompanyID)
	queryAdd(query, "status", string(params.Status))
	queryAdd(query, "from", params.From)
	queryAdd(query, "to", params.To)
	queryAdd(query, "externalId", params.ExternalID)
	queryAddInt(query, "limit", params.Limit)
	queryAdd(query, "cursor", params.Cursor)
	return query
}

// Create issues an invoice. Pass a non-empty idempotencyKey to safely
// retry: the same key returns the first created invoice.
//
// POST /invoices — unwrap "data+props". Returns 202 while the invoice is
// queued.
func (i *Invoices) Create(ctx context.Context, params InvoiceCreateParams, idempotencyKey string) (Invoice, *CorvoError) {
	return requestJSON[Invoice](ctx, i.t, requestSpec{
		method:         http.MethodPost,
		path:           "/invoices",
		jsonBody:       params,
		idempotencyKey: idempotencyKey,
		unwrap:         UnwrapDataProps,
	})
}

// Get reads an invoice by ID. GET /invoices/:id — unwrap "data+props".
func (i *Invoices) Get(ctx context.Context, id string) (Invoice, *CorvoError) {
	return requestJSON[Invoice](ctx, i.t, requestSpec{
		method: http.MethodGet,
		path:   "/invoices/" + url.PathEscape(id),
		unwrap: UnwrapDataProps,
	})
}

// List returns one page of invoices plus its pagination.
// GET /invoices — unwrap "none", items wrapped in { props }.
func (i *Invoices) List(ctx context.Context, params *InvoiceListParams) ([]Invoice, Pagination, *CorvoError) {
	return requestPage[Invoice](ctx, i.t, "/invoices", invoiceListQuery(params), true)
}

// Iterate walks every page of invoices matching params, calling fn per
// item until it returns false or all pages are consumed. A page failure —
// including the first — is returned as *CorvoError.
func (i *Invoices) Iterate(ctx context.Context, params *InvoiceListParams, fn func(Invoice) bool) *CorvoError {
	return iteratePages[Invoice](ctx, i.t, "/invoices", invoiceListQuery(params), true, fn)
}

// Events returns one page of status transitions of an invoice.
// GET /invoices/:id/events — unwrap "none", items wrapped in { props }.
func (i *Invoices) Events(ctx context.Context, id string, params *ListParams) ([]InvoiceEvent, Pagination, *CorvoError) {
	return requestPage[InvoiceEvent](ctx, i.t,
		"/invoices/"+url.PathEscape(id)+"/events", listQuery(params), true)
}

// PDF downloads the PDF representation of an authorized invoice.
// GET /invoices/:id/pdf — raw binary.
func (i *Invoices) PDF(ctx context.Context, id string) (DownloadedFile, *CorvoError) {
	return requestFile(ctx, i.t, http.MethodGet, "/invoices/"+url.PathEscape(id)+"/pdf")
}

// XML downloads the XML (DPS/procNFe) representation of an invoice.
// GET /invoices/:id/xml — raw binary.
func (i *Invoices) XML(ctx context.Context, id string) (DownloadedFile, *CorvoError) {
	return requestFile(ctx, i.t, http.MethodGet, "/invoices/"+url.PathEscape(id)+"/xml")
}

// DownloadURL mints a short-lived signed URL for the invoice document in
// the given format (FormatPDF or FormatXML).
// POST /invoices/:id/download-url — unwrap "data".
func (i *Invoices) DownloadURL(ctx context.Context, id string, format DownloadFormat) (InvoiceDownloadUrl, *CorvoError) {
	return requestJSON[InvoiceDownloadUrl](ctx, i.t, requestSpec{
		method:   http.MethodPost,
		path:     "/invoices/" + url.PathEscape(id) + "/download-url",
		jsonBody: struct {
			Format DownloadFormat `json:"format"`
		}{Format: format},
	})
}

// SendEmail requests (re)delivery of the invoice e-mail. Pass a non-empty
// idempotencyKey to safely retry. POST /invoices/:id/email — unwrap "data".
func (i *Invoices) SendEmail(ctx context.Context, id string, idempotencyKey string) (InvoiceEmailResult, *CorvoError) {
	return requestJSON[InvoiceEmailResult](ctx, i.t, requestSpec{
		method:         http.MethodPost,
		path:           "/invoices/" + url.PathEscape(id) + "/email",
		idempotencyKey: idempotencyKey,
	})
}
