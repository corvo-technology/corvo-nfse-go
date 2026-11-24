package corvonfse

import (
	"context"
	"encoding/json"
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
	queryAdd(query, "customerTaxId", params.CustomerTaxID)
	queryAddInt(query, "limit", params.Limit)
	queryAdd(query, "cursor", params.Cursor)
	return query
}

// Create issues an invoice, or schedules one. Pass a non-empty idempotencyKey
// to safely retry: the same key returns the first result.
//
// With Options.Type IMMEDIATE (the default) the result has Invoice set (status
// "queued"). With SCHEDULE or RECURRING no invoice exists yet: the result has
// Schedule set, and the note is issued when the schedule runs (follow it with
// Schedules.ListRuns).
//
// POST /invoices — unwrap "data". Returns 202.
func (i *Invoices) Create(ctx context.Context, params InvoiceCreateParams, idempotencyKey string) (InvoiceCreateResult, *CorvoError) {
	var result InvoiceCreateResult
	raw, cerr := requestJSON[json.RawMessage](ctx, i.t, requestSpec{
		method:         http.MethodPost,
		path:           "/invoices",
		jsonBody:       params,
		idempotencyKey: idempotencyKey,
	})
	if cerr != nil {
		return result, cerr
	}
	if len(raw) == 0 {
		return result, nil
	}
	// A schedule answers in snake_case with anchor_on; an invoice never has it.
	var probe struct {
		AnchorOn *string `json:"anchor_on"`
	}
	_ = json.Unmarshal(raw, &probe)
	if probe.AnchorOn != nil {
		var schedule Schedule
		if err := json.Unmarshal(raw, &schedule); err != nil {
			return result, invalidJSON(http.StatusAccepted, "")
		}
		result.Schedule = &schedule
		return result, nil
	}
	var invoice Invoice
	if err := json.Unmarshal(raw, &invoice); err != nil {
		return result, invalidJSON(http.StatusAccepted, "")
	}
	result.Invoice = &invoice
	return result, nil
}

// CreateBatch issues 1 to 50 invoices in one call, for one company or several.
// Items are checked one by one: valid ones are saved together, invalid ones
// come back rejected and are not saved. Match Results[i].Index to the item you
// sent, fix only the rejected ones and send them again. Set IdempotencyKey on
// each item; the Idempotency-Key header is not used by this route.
//
// POST /invoices/batches — unwrap "data". Returns 202.
func (i *Invoices) CreateBatch(ctx context.Context, items []InvoiceBatchItem) (InvoiceBatchSummary, *CorvoError) {
	return requestJSON[InvoiceBatchSummary](ctx, i.t, requestSpec{
		method: http.MethodPost,
		path:   "/invoices/batches",
		jsonBody: struct {
			Invoices []InvoiceBatchItem `json:"invoices"`
		}{Invoices: items},
	})
}

// Calculate estimates the taxes of an invoice without issuing, storing or
// charging anything. A fiscal gap is not an error: the result comes back with
// Status "blocked", Blocking and PendingFields set.
// POST /invoices/calculate — unwrap "data".
func (i *Invoices) Calculate(ctx context.Context, params InvoiceCreateParams) (InvoiceCalculation, *CorvoError) {
	return requestJSON[InvoiceCalculation](ctx, i.t, requestSpec{
		method:   http.MethodPost,
		path:     "/invoices/calculate",
		jsonBody: params,
	})
}

// Get reads an invoice by ID. GET /invoices/:id — unwrap "data".
func (i *Invoices) Get(ctx context.Context, id string) (Invoice, *CorvoError) {
	return requestJSON[Invoice](ctx, i.t, requestSpec{
		method: http.MethodGet,
		path:   "/invoices/" + url.PathEscape(id),
	})
}

// List returns one page of invoices plus its pagination.
// GET /invoices — unwrap "none".
func (i *Invoices) List(ctx context.Context, params *InvoiceListParams) ([]Invoice, Pagination, *CorvoError) {
	return requestPage[Invoice](ctx, i.t, "/invoices", invoiceListQuery(params))
}

// Iterate walks every page of invoices matching params, calling fn per
// item until it returns false or all pages are consumed. A page failure —
// including the first — is returned as *CorvoError.
func (i *Invoices) Iterate(ctx context.Context, params *InvoiceListParams, fn func(Invoice) bool) *CorvoError {
	return iteratePages[Invoice](ctx, i.t, "/invoices", invoiceListQuery(params), fn)
}

// Events returns one page of status transitions of an invoice.
// GET /invoices/:id/events — unwrap "none".
func (i *Invoices) Events(ctx context.Context, id string, params *ListParams) ([]InvoiceEvent, Pagination, *CorvoError) {
	return requestPage[InvoiceEvent](ctx, i.t,
		"/invoices/"+url.PathEscape(id)+"/events", listQuery(params))
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
		method: http.MethodPost,
		path:   "/invoices/" + url.PathEscape(id) + "/download-url",
		jsonBody: struct {
			Format DownloadFormat `json:"format"`
		}{Format: format},
	})
}

// SendEmail requests (re)delivery of the invoice e-mail. params is optional:
// set Email to send to another address than the customer's. Pass a non-empty
// idempotencyKey to safely retry. POST /invoices/:id/email — unwrap "data".
func (i *Invoices) SendEmail(ctx context.Context, id string, params *InvoiceEmailParams, idempotencyKey string) (InvoiceEmailResult, *CorvoError) {
	spec := requestSpec{
		method:         http.MethodPost,
		path:           "/invoices/" + url.PathEscape(id) + "/email",
		idempotencyKey: idempotencyKey,
	}
	if params != nil {
		spec.jsonBody = params
	}
	return requestJSON[InvoiceEmailResult](ctx, i.t, spec)
}

// Substitute replaces an authorized invoice with a corrected one: only the
// description (why), service.description, purchaseOrder and notes can change.
// The result is the new invoice, queued. Pass a non-empty idempotencyKey to
// safely retry. POST /invoices/:id/substitute — unwrap "data". Returns 202.
func (i *Invoices) Substitute(ctx context.Context, id string, params InvoiceSubstituteParams, idempotencyKey string) (Invoice, *CorvoError) {
	return requestJSON[Invoice](ctx, i.t, requestSpec{
		method:         http.MethodPost,
		path:           "/invoices/" + url.PathEscape(id) + "/substitute",
		jsonBody:       params,
		idempotencyKey: idempotencyKey,
	})
}

// Review resolves invoices in manual_review (up to 100): retry sends them
// again with the same DPS, discard rejects them once the portal confirms the
// DPS was never registered. Invoices not in manual review are skipped.
// POST /invoices/review — unwrap "data". Returns 202.
func (i *Invoices) Review(ctx context.Context, params InvoiceReviewParams) (InvoiceReviewResult, *CorvoError) {
	return requestJSON[InvoiceReviewResult](ctx, i.t, requestSpec{
		method:   http.MethodPost,
		path:     "/invoices/review",
		jsonBody: params,
	})
}

// Overview returns the invoice counters, the daily activity and the recent
// invoices of the window (OverviewPeriod7d by default, or OverviewPeriod30d).
// GET /invoices/overview — unwrap "data".
func (i *Invoices) Overview(ctx context.Context, period InvoiceOverviewPeriod) (InvoiceOverview, *CorvoError) {
	query := url.Values{}
	queryAdd(query, "period", string(period))
	return requestJSON[InvoiceOverview](ctx, i.t, requestSpec{
		method: http.MethodGet,
		path:   "/invoices/overview",
		query:  query,
	})
}

// Cancel cancels an invoice. A queued invoice is cancelled at once (Status
// "cancelled"); an authorized one moves to "cancelling" while the cancellation
// is processed, and a processing one stays "processing" with
// CancellationRequestedAt set. Call Get later: "cancelled" when done, or back
// to "authorized" with LastErrorCode when the cancellation was rejected.
//
// POST /invoices/:id/cancel — unwrap "data". Returns 200 or 202.
func (i *Invoices) Cancel(ctx context.Context, id string, params InvoiceCancelParams) (Invoice, *CorvoError) {
	return requestJSON[Invoice](ctx, i.t, requestSpec{
		method:   http.MethodPost,
		path:     "/invoices/" + url.PathEscape(id) + "/cancel",
		jsonBody: params,
	})
}
