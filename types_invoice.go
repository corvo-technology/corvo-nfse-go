package corvonfse

// InvoiceStatus is the lifecycle status of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusQueued                 InvoiceStatus = "queued"
	InvoiceStatusProcessing             InvoiceStatus = "processing"
	InvoiceStatusAuthorized             InvoiceStatus = "authorized"
	InvoiceStatusRejected               InvoiceStatus = "rejected"
	InvoiceStatusRetryPending           InvoiceStatus = "retry_pending"
	InvoiceStatusReconciliationRequired InvoiceStatus = "reconciliation_required"
	InvoiceStatusManualReview           InvoiceStatus = "manual_review"
	InvoiceStatusFailed                 InvoiceStatus = "failed"
	InvoiceStatusCancelling             InvoiceStatus = "cancelling"
	InvoiceStatusCancelled              InvoiceStatus = "cancelled"
)

// InvoiceCancellationReason is why an invoice is being cancelled.
type InvoiceCancellationReason string

const (
	CancellationIssueError         InvoiceCancellationReason = "issue_error"
	CancellationServiceNotProvided InvoiceCancellationReason = "service_not_provided"
	CancellationOther              InvoiceCancellationReason = "other"
)

// InvoiceCancelParams are the fields accepted by Invoices.Cancel. Description
// (15 to 255 characters) is required to cancel an authorized or processing
// invoice and optional for a queued one.
type InvoiceCancelParams struct {
	Reason      InvoiceCancellationReason `json:"reason"`
	Description string                    `json:"description,omitempty"`
}

// InvoiceStage is the coarse emission stage of an invoice.
type InvoiceStage string

const (
	InvoiceStagePending        InvoiceStage = "pending"
	InvoiceStageAuthorized     InvoiceStage = "authorized"
	InvoiceStageDocumentsReady InvoiceStage = "documents_ready"
	InvoiceStageCompleted      InvoiceStage = "completed"
	InvoiceStageFailed         InvoiceStage = "failed"
	InvoiceStageCancelled      InvoiceStage = "cancelled"
)

// DownloadFormat selects the binary format for Invoices.DownloadURL.
type DownloadFormat string

const (
	FormatPDF DownloadFormat = "pdf"
	FormatXML DownloadFormat = "xml"
)

// InvoiceCustomerInput is the customer (tomador) of an invoice. Send TaxID
// (CPF or CNPJ) for a Brazilian customer, or Foreign for a customer outside
// Brazil (it cannot be combined with TaxID or Address, and Name is then
// required). Omit the customer for an invoice without one. Name may be left
// empty when the customer was already saved: it is filled from the CPF/CNPJ.
type InvoiceCustomerInput struct {
	TaxID   string                  `json:"taxId,omitempty"`
	Name    string                  `json:"name,omitempty"`
	Email   string                  `json:"email,omitempty"`
	Address *InvoiceCustomerAddress `json:"address,omitempty"`
	Foreign *ForeignCustomer        `json:"foreign,omitempty"`
}

// InvoiceCustomerAddress is the optional customer address. All fields are
// optional; empty fields are omitted from the request.
type InvoiceCustomerAddress struct {
	Street           string `json:"street,omitempty"`
	City             string `json:"city,omitempty"`
	State            string `json:"state,omitempty"`
	Zip              string `json:"zip,omitempty"`
	Country          string `json:"country,omitempty"`
	MunicipalityCode string `json:"municipalityCode,omitempty"`
}

// ForeignCustomer is a customer outside Brazil. Identify it with TaxID or with
// NoTaxIDReason ("exempt" or "not_required").
type ForeignCustomer struct {
	TaxID         string `json:"taxId,omitempty"`
	NoTaxIDReason string `json:"noTaxIdReason,omitempty"`
	Country       string `json:"country"`
	PostalCode    string `json:"postalCode"`
	City          string `json:"city"`
	Region        string `json:"region"`
	Street        string `json:"street"`
	Number        string `json:"number"`
	District      string `json:"district"`
	Complement    string `json:"complement,omitempty"`
}

// InvoiceServiceInput describes the service. All fields are optional: Code
// (the national service code, 6 digits) defaults to the company's and then to
// 100901, MunicipalityCode (IBGE, 7 digits) to the company's city, and
// Description to the name of the service code.
type InvoiceServiceInput struct {
	Code             string `json:"code,omitempty"`
	MunicipalityCode string `json:"municipalityCode,omitempty"`
	Description      string `json:"description,omitempty"`
	// ServiceProfileID picks a service tax profile of the company.
	ServiceProfileID string `json:"serviceProfileId,omitempty"`
	// Fiscal overrides the tax classification of the service.
	Fiscal *ServiceFiscal `json:"fiscal,omitempty"`
}

// ServiceFiscal is the optional tax classification of a service.
type ServiceFiscal struct {
	ClassCode          string `json:"classCode,omitempty"`
	CST                string `json:"cst,omitempty"`
	NBSCode            string `json:"nbsCode,omitempty"`
	OperationIndicator string `json:"operationIndicator,omitempty"`
}

// InvoiceType selects immediate, scheduled or recurring issuance.
type InvoiceType string

const (
	InvoiceTypeImmediate InvoiceType = "IMMEDIATE"
	// InvoiceTypeSchedule creates a Schedule that issues the invoice once, on
	// CompetenceDate (from tomorrow, up to 15 days ahead).
	InvoiceTypeSchedule InvoiceType = "SCHEDULE"
	// InvoiceTypeRecurring creates a Schedule that issues on Recurrence
	// (CompetenceDate is the first run, up to 31 days ahead).
	InvoiceTypeRecurring InvoiceType = "RECURRING"
)

// RecurrenceFrequency is how often a recurring schedule runs.
type RecurrenceFrequency string

const (
	FrequencyWeekly   RecurrenceFrequency = "WEEKLY"
	FrequencyBiweekly RecurrenceFrequency = "BIWEEKLY"
	FrequencyMonthly  RecurrenceFrequency = "MONTHLY"
)

// InvoiceRecurrence is required for InvoiceTypeRecurring and refused
// otherwise. EndsOn (YYYY-MM-DD) is optional.
type InvoiceRecurrence struct {
	Frequency RecurrenceFrequency `json:"frequency"`
	EndsOn    string              `json:"endsOn,omitempty"`
}

// InvoiceCreateOptions are the optional emission knobs accepted by
// Invoices.Create. Type SCHEDULE or RECURRING does not issue an invoice: the
// API answers with a Schedule (see InvoiceCreateResult).
type InvoiceCreateOptions struct {
	Email      *InvoiceEmailOption `json:"email,omitempty"`
	Type       InvoiceType         `json:"type,omitempty"`
	Recurrence *InvoiceRecurrence  `json:"recurrence,omitempty"`
}

// InvoiceEmailOption controls whether the NFS-e e-mail is sent.
type InvoiceEmailOption struct {
	Send *bool `json:"send,omitempty"`
}

// InvoiceCreateParams are the fields accepted by Invoices.Create and
// Invoices.Calculate. Identify the company by CompanyID or CompanyTaxID.
type InvoiceCreateParams struct {
	CompanyID    string `json:"companyId,omitempty"`
	CompanyTaxID string `json:"companyTaxId,omitempty"`
	ExternalID   string `json:"externalId,omitempty"`
	// Customer is optional: leave it nil for an invoice without a customer.
	Customer *InvoiceCustomerInput `json:"customer,omitempty"`
	Service  InvoiceServiceInput   `json:"service"`
	// Amount in BRL, positive, up to 2 decimal places.
	Amount float64 `json:"amount"`
	// CompetenceDate is YYYY-MM-DD.
	CompetenceDate string `json:"competenceDate"`
	// Discount is the unconditional discount in BRL, from 0 up to Amount, at
	// most 2 decimal places.
	Discount *float64 `json:"discount,omitempty"`
	// Deduction is the ISS base deduction declared by the issuer, in BRL:
	// greater than 0 and less than Amount minus Discount, at most 2 decimal
	// places (0 is the same as omitting it). The invoice is still issued for
	// the full Amount; the issuer is responsible for the declared value.
	Deduction *float64 `json:"deduction,omitempty"`
	// Email requests the e-mail notification (default false);
	// Options.Email.Send wins when set.
	Email   *bool                 `json:"email,omitempty"`
	Options *InvoiceCreateOptions `json:"options,omitempty"`
	// PurchaseOrder is the purchase order number (up to 60 characters, DPS xPed).
	PurchaseOrder string `json:"purchaseOrder,omitempty"`
	// Notes is additional information (up to 2000 characters, DPS xInfComp).
	Notes string `json:"notes,omitempty"`
}

// InvoiceCreateResult is the answer of Invoices.Create: an Invoice for an
// IMMEDIATE emission, or a Schedule for SCHEDULE and RECURRING. Exactly one of
// the two is set.
type InvoiceCreateResult struct {
	Invoice  *Invoice
	Schedule *Schedule
}

// IsSchedule reports whether the API created a schedule instead of an invoice.
func (r InvoiceCreateResult) IsSchedule() bool {
	return r.Schedule != nil
}

// InvoiceListParams are the filters accepted by Invoices.List and
// Invoices.Iterate. A nil *InvoiceListParams is treated as the zero value.
type InvoiceListParams struct {
	CompanyID  string
	Status     InvoiceStatus
	From       string
	To         string
	ExternalID string
	// CustomerTaxID filters by the customer CPF or CNPJ (digits).
	CustomerTaxID string
	Limit         int
	Cursor        string
}

// Invoice is a fiscal invoice. Nullable server fields map to their zero
// value when null. TaxSnapshot holds provider-specific tax details as
// returned by the API, unchanged.
type Invoice struct {
	ID                 string                  `json:"id"`
	CompanyID          string                  `json:"companyId"`
	TenantID           string                  `json:"tenantId"`
	Environment        Environment             `json:"environment"`
	DocumentKeyVersion int                     `json:"documentKeyVersion"`
	EmailSnapshot      InvoiceEmailSnapshot    `json:"emailSnapshot"`
	ExternalID         string                  `json:"externalId"`
	IdempotencyKey     string                  `json:"idempotencyKey"`
	Status             InvoiceStatus           `json:"status"`
	CustomerSnapshot   InvoiceCustomerSnapshot `json:"customerSnapshot"`
	ServiceSnapshot    InvoiceServiceSnapshot  `json:"serviceSnapshot"`
	TaxSnapshot        map[string]any          `json:"taxSnapshot"`
	Amount             float64                 `json:"amount"`
	// Discount and Deduction are set only when declared on the invoice.
	Discount            *float64 `json:"discount,omitempty"`
	Deduction           *float64 `json:"deduction,omitempty"`
	Currency            string   `json:"currency"`
	CompetenceDate      string   `json:"competenceDate"`
	DPSID               string   `json:"dpsId"`
	DPSSeries           string   `json:"dpsSeries"`
	DPSNumber           int64    `json:"dpsNumber"`
	AccessKey           string   `json:"accessKey"`
	NFSENumber          int64    `json:"nfseNumber"`
	FiscalProvider      string   `json:"fiscalProvider"`
	FiscalSchemaVersion string   `json:"fiscalSchemaVersion"`
	IssuedAt            string   `json:"issuedAt"`
	AuthorizedAt        string   `json:"authorizedAt"`
	LastErrorCode       string   `json:"lastErrorCode"`
	LastErrorMessage    string   `json:"lastErrorMessage"`
	LastErrorProvider   string   `json:"lastErrorProvider"`
	DocumentsReadyAt    string   `json:"documentsReadyAt"`
	// SubstitutesInvoiceID and SubstitutionDescription are set on an invoice
	// that substitutes another one.
	SubstitutesInvoiceID    string `json:"substitutesInvoiceId"`
	SubstitutionDescription string `json:"substitutionDescription"`
	// PortalFlag is "taker_rejected" or "blocked" (from the national portal).
	PortalFlag      string       `json:"portalFlag"`
	LastEmailStatus string       `json:"lastEmailStatus"`
	LastEmailSentAt string       `json:"lastEmailSentAt"`
	CreatedAt       string       `json:"createdAt"`
	UpdatedAt       string       `json:"updatedAt"`
	DocumentsStatus string       `json:"documentsStatus"`
	EmailStatus     string       `json:"emailStatus"`
	Stage           InvoiceStage `json:"stage"`
	// ScheduledFor is when a scheduled invoice is released (ISO); empty for an
	// immediate one.
	ScheduledFor string `json:"scheduledFor"`
	// CancellationReason and CancellationDescription echo the cancellation
	// request; CancelledAt is when the invoice reached "cancelled" and
	// CancellationRequestedAt when a cancellation was requested while it was
	// still processing. All are empty until a cancellation is requested.
	CancellationReason      InvoiceCancellationReason `json:"cancellationReason"`
	CancellationDescription string                    `json:"cancellationDescription"`
	CancelledAt             string                    `json:"cancelledAt"`
	CancellationRequestedAt string                    `json:"cancellationRequestedAt"`
}

// InvoiceBatchItem is one item of Invoices.CreateBatch: the fields of
// InvoiceCreateParams plus its own idempotency key.
type InvoiceBatchItem struct {
	InvoiceCreateParams
	// IdempotencyKey (up to 255 characters) makes the item safe to retry:
	// repeating it with the same content returns the saved invoice with
	// Replayed set. The Idempotency-Key header is not used by the batch route.
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// InvoiceBatchError is why a batch item was rejected. Rejected items are not
// saved. Details holds the field-level issues, when there are any.
type InvoiceBatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// InvoiceBatchResult is the outcome of one batch item.
type InvoiceBatchResult struct {
	// Index is the position of the item in the request, starting at 0.
	Index int `json:"index"`
	// Status is "accepted" or "rejected".
	Status string `json:"status"`
	// Replayed is true when the item repeated an idempotency key already saved
	// with the same content.
	Replayed bool `json:"replayed"`
	// Invoice is set when Status is "accepted".
	Invoice *Invoice `json:"invoice,omitempty"`
	// Error is set when Status is "rejected".
	Error *InvoiceBatchError `json:"error,omitempty"`
}

// InvoiceBatchSummary is the result of Invoices.CreateBatch.
type InvoiceBatchSummary struct {
	Total    int `json:"total"`
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
	// Results has one entry per item, in the order sent.
	Results []InvoiceBatchResult `json:"results"`
}

// InvoiceEmailSnapshot records the e-mail options in force for the invoice.
type InvoiceEmailSnapshot struct {
	Send bool `json:"send"`
}

// InvoiceCustomerSnapshot is the customer data frozen at emission time.
type InvoiceCustomerSnapshot struct {
	Name    string                  `json:"name"`
	Email   string                  `json:"email"`
	TaxID   string                  `json:"taxId"`
	Address *InvoiceCustomerAddress `json:"address"`
	Foreign map[string]string       `json:"foreign"`
}

// InvoiceServiceSnapshot is the serviced item frozen at emission time.
type InvoiceServiceSnapshot struct {
	Code             string            `json:"code"`
	Description      string            `json:"description"`
	MunicipalityCode string            `json:"municipalityCode"`
	Fiscal           map[string]string `json:"fiscal"`
	Notes            string            `json:"notes"`
	PurchaseOrder    string            `json:"purchaseOrder"`
	ServiceProfileID string            `json:"serviceProfileId"`
}

// InvoiceEvent is a status transition recorded for an invoice.
type InvoiceEvent struct {
	ID         string        `json:"id"`
	InvoiceID  string        `json:"invoiceId"`
	FromStatus InvoiceStatus `json:"fromStatus"`
	ToStatus   InvoiceStatus `json:"toStatus"`
	Reason     string        `json:"reason"`
	// Source is "portal" for an event reported by the national portal, empty
	// otherwise; PortalEvent and PortalNsu then identify it.
	Source      string `json:"source"`
	PortalEvent string `json:"portalEvent"`
	PortalNsu   string `json:"portalNsu"`
	CreatedAt   string `json:"createdAt"`
}

// InvoiceDownloadUrl is a short-lived signed URL for an invoice document.
type InvoiceDownloadUrl struct {
	DownloadURL string `json:"downloadUrl"`
	ExpiresAt   string `json:"expiresAt"`
	FileName    string `json:"fileName"`
	MimeType    string `json:"mimeType"`
}

// InvoiceEmailParams are the optional fields of Invoices.SendEmail. Email
// sends the invoice to another address than the customer's.
type InvoiceEmailParams struct {
	Email string `json:"email,omitempty"`
}

// InvoiceSubstituteParams are the fields accepted by Invoices.Substitute: only
// the reason, the service description, the purchase order and the notes can
// change in a substitute.
type InvoiceSubstituteParams struct {
	// Description is why the invoice is being substituted (required).
	Description   string                    `json:"description"`
	Service       *InvoiceSubstituteService `json:"service,omitempty"`
	PurchaseOrder string                    `json:"purchaseOrder,omitempty"`
	Notes         string                    `json:"notes,omitempty"`
}

// InvoiceSubstituteService is the part of the service a substitute may change.
type InvoiceSubstituteService struct {
	Description string `json:"description,omitempty"`
}

// InvoiceReviewAction is what Invoices.Review does with invoices in manual
// review.
type InvoiceReviewAction string

const (
	ReviewRetry   InvoiceReviewAction = "retry"
	ReviewDiscard InvoiceReviewAction = "discard"
)

// InvoiceReviewParams are the fields accepted by Invoices.Review (up to 100
// invoice ids).
type InvoiceReviewParams struct {
	Action     InvoiceReviewAction `json:"action"`
	InvoiceIDs []string            `json:"invoiceIds"`
}

// InvoiceReviewItem is the outcome for one invoice of Invoices.Review.
// Outcome is "accepted" or "skipped"; Reason is set when skipped (not_found,
// not_in_manual_review or changed).
type InvoiceReviewItem struct {
	ID      string `json:"id"`
	Outcome string `json:"outcome"`
	Reason  string `json:"reason"`
}

// InvoiceReviewResult is the answer of Invoices.Review.
type InvoiceReviewResult struct {
	Results []InvoiceReviewItem `json:"results"`
}

// InvoiceOverviewPeriod is the window of Invoices.Overview.
type InvoiceOverviewPeriod string

const (
	OverviewPeriod7d  InvoiceOverviewPeriod = "7d"
	OverviewPeriod30d InvoiceOverviewPeriod = "30d"
)

// OverviewCounts are the invoice counters of the overview.
type OverviewCounts struct {
	Authorized int `json:"authorized"`
	Processing int `json:"processing"`
	Rejected   int `json:"rejected"`
}

// OverviewDay is the number of invoices authorized on one day.
type OverviewDay struct {
	Count int    `json:"count"`
	Date  string `json:"date"`
}

// InvoiceOverview is the answer of Invoices.Overview.
type InvoiceOverview struct {
	Counts         OverviewCounts        `json:"counts"`
	DailyActivity  []OverviewDay         `json:"dailyActivity"`
	From           string                `json:"from"`
	Period         InvoiceOverviewPeriod `json:"period"`
	RecentInvoices []Invoice             `json:"recentInvoices"`
	Timezone       string                `json:"timezone"`
	To             string                `json:"to"`
}

// InvoiceEmailResult acknowledges an e-mail send request.
type InvoiceEmailResult struct {
	EmailID string `json:"emailId"`
	Status  string `json:"status"`
}

// CalculatedTaxAmount is one tax of a pre-emission estimate. Rate is empty on
// IBSTotal, which sums the state and municipal IBS.
type CalculatedTaxAmount struct {
	Base   string `json:"base"`
	Rate   string `json:"rate,omitempty"`
	Amount string `json:"amount"`
}

// CalculatedWithholding is a withheld tax of a pre-emission estimate.
type CalculatedWithholding struct {
	Tax    string `json:"tax"`
	Base   string `json:"base"`
	Rate   string `json:"rate"`
	Amount string `json:"amount"`
}

// CalculatedTaxes groups the taxes of a pre-emission estimate.
type CalculatedTaxes struct {
	ISS      *CalculatedTaxAmount `json:"iss,omitempty"`
	CBS      *CalculatedTaxAmount `json:"cbs,omitempty"`
	IBSState *CalculatedTaxAmount `json:"ibsState,omitempty"`
	IBSCity  *CalculatedTaxAmount `json:"ibsCity,omitempty"`
	IBSTotal *CalculatedTaxAmount `json:"ibsTotal,omitempty"`
}

// CalculationBlocking explains why an estimate is blocked.
type CalculationBlocking struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Step    string `json:"step"`
}

// CalculationPendingField is a field missing for the estimate.
type CalculationPendingField struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// InvoiceCalculation is the pre-emission estimate returned by
// Invoices.Calculate. Status is "ready" or "blocked": a fiscal gap is data, not
// an error. Money values are decimal strings.
type InvoiceCalculation struct {
	PreEmission         bool                      `json:"preEmission"`
	Status              string                    `json:"status"`
	Disclaimer          string                    `json:"disclaimer"`
	CompetenceDate      string                    `json:"competenceDate"`
	TaxRegime           string                    `json:"taxRegime"`
	Classification      map[string]any            `json:"classification"`
	Applicability       map[string]any            `json:"applicability"`
	Taxes               CalculatedTaxes           `json:"taxes"`
	Withholdings        []CalculatedWithholding   `json:"withholdings"`
	NetReceivableAmount string                    `json:"netReceivableAmount"`
	CollectionWaived    bool                      `json:"collectionWaived"`
	Calculation         map[string]any            `json:"calculation"`
	Warnings            []map[string]any          `json:"warnings"`
	PendingFields       []CalculationPendingField `json:"pendingFields"`
	Blocking            *CalculationBlocking      `json:"blocking"`
	Trace               []map[string]any          `json:"trace"`
}
