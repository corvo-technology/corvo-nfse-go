package corvonfse

import (
	"context"
	"net/http"
	"net/url"
)

// AccountantPackage configures and sends the monthly package (XML, PDF and
// summary of the month) to the accountant of a company.
type AccountantPackage struct {
	t *Transport
}

func accountantPath(companyID string) string {
	return "/companies/" + url.PathEscape(companyID) + "/accountant-package"
}

// Get returns the accountant e-mail of the company.
// GET /companies/:companyId/accountant-package — unwrap "data".
func (a *AccountantPackage) Get(ctx context.Context, companyID string) (AccountantSettings, *CorvoError) {
	return requestJSON[AccountantSettings](ctx, a.t, requestSpec{
		method: http.MethodGet,
		path:   accountantPath(companyID),
	})
}

// Update sets the accountant e-mail; Null[string]() turns the monthly send
// off. PUT /companies/:companyId/accountant-package — unwrap "data".
func (a *AccountantPackage) Update(ctx context.Context, companyID string, params AccountantPackageUpdateParams) (AccountantSettingsUpdate, *CorvoError) {
	return requestJSON[AccountantSettingsUpdate](ctx, a.t, requestSpec{
		method:   http.MethodPut,
		path:     accountantPath(companyID),
		jsonBody: params,
	})
}

// Send sends the package of a month (YYYY-MM) to the accountant now.
// POST /companies/:companyId/accountant-package/send — unwrap "data".
func (a *AccountantPackage) Send(ctx context.Context, companyID string, month string) (AccountantSendResult, *CorvoError) {
	return requestJSON[AccountantSendResult](ctx, a.t, requestSpec{
		method: http.MethodPost,
		path:   accountantPath(companyID) + "/send",
		jsonBody: struct {
			Month string `json:"month"`
		}{Month: month},
	})
}

// ListSends returns the last 12 attempts to send the package (not paginated).
// GET /companies/:companyId/accountant-package/sends — unwrap "data".
func (a *AccountantPackage) ListSends(ctx context.Context, companyID string) ([]AccountantSend, *CorvoError) {
	return requestJSON[[]AccountantSend](ctx, a.t, requestSpec{
		method: http.MethodGet,
		path:   accountantPath(companyID) + "/sends",
	})
}

// InvoicePackage builds the month package (ZIP with xml/, pdf/ and
// resumo.csv) of a company, at most 500 notes (PACKAGE_TOO_LARGE above that).
type InvoicePackage struct {
	t *Transport
}

// Summary returns the count and total of the month (YYYY-MM) package without
// building it. GET /companies/:companyId/invoice-package/summary — unwrap "data".
func (p *InvoicePackage) Summary(ctx context.Context, companyID string, month string) (PackageSummary, *CorvoError) {
	query := url.Values{}
	query.Set("month", month)
	return requestJSON[PackageSummary](ctx, p.t, requestSpec{
		method: http.MethodGet,
		path:   "/companies/" + url.PathEscape(companyID) + "/invoice-package/summary",
		query:  query,
	})
}

// Download returns the ZIP of the month (YYYY-MM). include limits the parts to
// "xml" and/or "pdf" (nil includes both). 10 packages per hour per workspace.
// GET /companies/:companyId/invoice-package — raw binary.
func (p *InvoicePackage) Download(ctx context.Context, companyID string, month string, include []string) (DownloadedFile, *CorvoError) {
	query := url.Values{}
	query.Set("month", month)
	if len(include) > 0 {
		queryAdd(query, "include", joinComma(include))
	}
	return requestFileQuery(ctx, p.t, http.MethodGet,
		"/companies/"+url.PathEscape(companyID)+"/invoice-package", query)
}

// DPSSequence reads and continues the DPS numbering of a company, for example
// to continue from another system.
type DPSSequence struct {
	t *Transport
}

// Get returns the series and the last DPS number.
// GET /companies/:id/dps-sequence — unwrap "data".
func (d *DPSSequence) Get(ctx context.Context, companyID string) (DPSSequenceInfo, *CorvoError) {
	return requestJSON[DPSSequenceInfo](ctx, d.t, requestSpec{
		method: http.MethodGet,
		path:   "/companies/" + url.PathEscape(companyID) + "/dps-sequence",
	})
}

// Set continues the numbering from the last number issued elsewhere. A number
// below the current one fails with DPS_NUMBER_BELOW_CURRENT.
// PUT /companies/:id/dps-sequence — unwrap "data".
func (d *DPSSequence) Set(ctx context.Context, companyID string, params DPSSequenceSetParams) (DPSSequenceInfo, *CorvoError) {
	return requestJSON[DPSSequenceInfo](ctx, d.t, requestSpec{
		method:   http.MethodPut,
		path:     "/companies/" + url.PathEscape(companyID) + "/dps-sequence",
		jsonBody: params,
	})
}

// FiscalPortal reports the national NFS-e portal availability.
type FiscalPortal struct {
	t *Transport
}

// Status tells whether the national portal is answering for the environment of
// the key. GET /fiscal-portal/status — unwrap "data".
func (f *FiscalPortal) Status(ctx context.Context) (PortalStatus, *CorvoError) {
	return requestJSON[PortalStatus](ctx, f.t, requestSpec{
		method: http.MethodGet,
		path:   "/fiscal-portal/status",
	})
}

// FiscalProfiles manages the fiscal profile history of a company.
type FiscalProfiles struct {
	t *Transport
}

// List returns the fiscal profile history (not paginated).
// GET /companies/:id/fiscal-profile — unwrap "data".
func (f *FiscalProfiles) List(ctx context.Context, companyID string) ([]FiscalProfile, *CorvoError) {
	return requestJSON[[]FiscalProfile](ctx, f.t, requestSpec{
		method: http.MethodGet,
		path:   "/companies/" + url.PathEscape(companyID) + "/fiscal-profile",
	})
}

// Upsert creates a validity period (idempotent for an identical desired
// state). PUT /companies/:id/fiscal-profile — unwrap "data".
func (f *FiscalProfiles) Upsert(ctx context.Context, companyID string, params FiscalProfileUpsertParams) (FiscalProfile, *CorvoError) {
	return requestJSON[FiscalProfile](ctx, f.t, requestSpec{
		method:   http.MethodPut,
		path:     "/companies/" + url.PathEscape(companyID) + "/fiscal-profile",
		jsonBody: params,
	})
}

// Approve approves a fiscal profile of the company.
// POST /companies/:id/fiscal-profile/:profileId/approve — unwrap "data".
func (f *FiscalProfiles) Approve(ctx context.Context, companyID string, profileID string) (FiscalProfile, *CorvoError) {
	return requestJSON[FiscalProfile](ctx, f.t, requestSpec{
		method: http.MethodPost,
		path:   "/companies/" + url.PathEscape(companyID) + "/fiscal-profile/" + url.PathEscape(profileID) + "/approve",
	})
}

// ServiceTaxProfiles manages the service tax catalog of a company.
type ServiceTaxProfiles struct {
	t *Transport
}

func serviceTaxPath(companyID string) string {
	return "/companies/" + url.PathEscape(companyID) + "/service-tax-profiles"
}

// List returns the profiles effective on date (YYYY-MM-DD; empty is today).
// GET /companies/:id/service-tax-profiles — unwrap "data".
func (s *ServiceTaxProfiles) List(ctx context.Context, companyID string, date string) ([]ServiceTaxProfile, *CorvoError) {
	query := url.Values{}
	queryAdd(query, "date", date)
	return requestJSON[[]ServiceTaxProfile](ctx, s.t, requestSpec{
		method: http.MethodGet,
		path:   serviceTaxPath(companyID),
		query:  query,
	})
}

// Create adds a profile to the company catalog.
// POST /companies/:id/service-tax-profiles — unwrap "data".
func (s *ServiceTaxProfiles) Create(ctx context.Context, companyID string, params ServiceTaxProfileCreateParams) (ServiceTaxProfile, *CorvoError) {
	return requestJSON[ServiceTaxProfile](ctx, s.t, requestSpec{
		method:   http.MethodPost,
		path:     serviceTaxPath(companyID),
		jsonBody: params,
	})
}

// Update changes a profile of the company catalog.
// PATCH /companies/:id/service-tax-profiles/:serviceId — unwrap "data".
func (s *ServiceTaxProfiles) Update(ctx context.Context, companyID string, serviceID string, params ServiceTaxProfileUpdateParams) (ServiceTaxProfile, *CorvoError) {
	return requestJSON[ServiceTaxProfile](ctx, s.t, requestSpec{
		method:   http.MethodPatch,
		path:     serviceTaxPath(companyID) + "/" + url.PathEscape(serviceID),
		jsonBody: params,
	})
}

// Remove deletes an expired profile that was not approved yet.
// DELETE /companies/:id/service-tax-profiles/:serviceId — unwrap "data" (null).
func (s *ServiceTaxProfiles) Remove(ctx context.Context, companyID string, serviceID string) *CorvoError {
	_, err := requestJSON[any](ctx, s.t, requestSpec{
		method: http.MethodDelete,
		path:   serviceTaxPath(companyID) + "/" + url.PathEscape(serviceID),
	})
	return err
}

// Approve approves the saved classification of a service for the company.
// POST /companies/:id/service-tax-profiles/:serviceId/approve — unwrap "data".
func (s *ServiceTaxProfiles) Approve(ctx context.Context, companyID string, serviceID string) (ServiceTaxProfile, *CorvoError) {
	return requestJSON[ServiceTaxProfile](ctx, s.t, requestSpec{
		method: http.MethodPost,
		path:   serviceTaxPath(companyID) + "/" + url.PathEscape(serviceID) + "/approve",
	})
}
