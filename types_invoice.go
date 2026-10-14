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
)

// InvoiceStage is the coarse emission stage of an invoice.
type InvoiceStage string

const (
	InvoiceStagePending          InvoiceStage = "pending"
	InvoiceStageAuthorized       InvoiceStage = "authorized"
	InvoiceStageDocumentsReady   InvoiceStage = "documents_ready"
	InvoiceStageCompleted        InvoiceStage = "completed"
	InvoiceStageFailed           InvoiceStage = "failed"
)

// DownloadFormat selects the binary format for Invoices.DownloadURL.
type DownloadFormat string

const (
	FormatPDF DownloadFormat = "pdf"
	FormatXML DownloadFormat = "xml"
)

// InvoiceCustomerInput is the customer (tomador) of an invoice.
type InvoiceCustomerInput struct {
	Name    string                    `json:"name"`
	TaxID   string                    `json:"taxId"`
	Email   string                    `json:"email,omitempty"`
	Address *InvoiceCustomerAddress   `json:"address,omitempty"`
}

// InvoiceCustomerAddress is the optional customer address. All fields are
// optional; empty fields are omitted from the request.
type InvoiceCustomerAddress struct {
	Street  string `json:"street,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Zip     string `json:"zip,omitempty"`
	Country string `json:"country,omitempty"`
}

// InvoiceServiceInput describes the serviced item (serviço) — Code is the
// municipal service code (six digits).
type InvoiceServiceInput struct {
	Code            string `json:"code"`
	MunicipalityCode string `json:"municipalityCode,omitempty"`
	Description     string `json:"description,omitempty"`
}

// InvoiceCreateOptions are the optional emission knobs accepted by
// Invoices.Create.
type InvoiceCreateOptions struct {
	Email *InvoiceEmailOption `json:"email,omitempty"`
}

// InvoiceEmailOption controls whether the NFS-e e-mail is sent.
type InvoiceEmailOption struct {
	Send *bool `json:"send,omitempty"`
}

// InvoiceCreateParams are the fields accepted by Invoices.Create.
type InvoiceCreateParams struct {
	CompanyID      string                 `json:"companyId,omitempty"`
	CompanyTaxID   string                 `json:"companyTaxId,omitempty"`
	ExternalID     string                 `json:"externalId,omitempty"`
	CorrelationID  string                 `json:"correlationId,omitempty"`
	Customer       InvoiceCustomerInput   `json:"customer"`
	Service        InvoiceServiceInput    `json:"service"`
	Amount         float64                `json:"amount"`
	CompetenceDate string                 `json:"competenceDate"`
	Email          string                 `json:"email,omitempty"`
	Options        *InvoiceCreateOptions  `json:"options,omitempty"`
}

// InvoiceListParams are the filters accepted by Invoices.List and
// Invoices.Iterate. A nil *InvoiceListParams is treated as the zero value.
type InvoiceListParams struct {
	CompanyID  string
	Status     InvoiceStatus
	From       string
	To         string
	ExternalID string
	Limit      int
	Cursor     string
}

// Invoice is a fiscal invoice. Nullable server fields map to their zero
// value when null. TaxSnapshot holds provider-specific tax details as
// returned by the API, unchanged.
type Invoice struct {
	ID                 string         `json:"id"`
	CompanyID          string         `json:"companyId"`
	TenantID           string         `json:"tenantId"`
	Environment        Environment    `json:"environment"`
	DocumentKeyVersion int            `json:"documentKeyVersion"`
	EmailSnapshot      InvoiceEmailSnapshot `json:"emailSnapshot"`
	ExternalID         string         `json:"externalId"`
	IdempotencyKey     string         `json:"idempotencyKey"`
	Status             InvoiceStatus  `json:"status"`
	CustomerSnapshot   InvoiceCustomerSnapshot `json:"customerSnapshot"`
	ServiceSnapshot    InvoiceServiceSnapshot  `json:"serviceSnapshot"`
	TaxSnapshot        map[string]any `json:"taxSnapshot"`
	Amount             float64        `json:"amount"`
	Currency           string         `json:"currency"`
	CompetenceDate     string         `json:"competenceDate"`
	DPSID              string         `json:"dpsId"`
	DPSSeries          string         `json:"dpsSeries"`
	DPSNumber          int            `json:"dpsNumber"`
	AccessKey          string         `json:"accessKey"`
	NFSENumber         int            `json:"nfseNumber"`
	FiscalProvider     string         `json:"fiscalProvider"`
	FiscalSchemaVersion string        `json:"fiscalSchemaVersion"`
	IssuedAt           string         `json:"issuedAt"`
	AuthorizedAt       string         `json:"authorizedAt"`
	LastErrorCode      string         `json:"lastErrorCode"`
	LastErrorMessage   string         `json:"lastErrorMessage"`
	LastErrorProvider  string         `json:"lastErrorProvider"`
	DocumentsReadyAt   string         `json:"documentsReadyAt"`
	LastEmailStatus    string         `json:"lastEmailStatus"`
	LastEmailSentAt    string         `json:"lastEmailSentAt"`
	CreatedAt          string         `json:"createdAt"`
	UpdatedAt          string         `json:"updatedAt"`
	DocumentsStatus    string         `json:"documentsStatus"`
	EmailStatus        string         `json:"emailStatus"`
	Stage              InvoiceStage   `json:"stage"`
}

// InvoiceEmailSnapshot records the e-mail options in force for the invoice.
type InvoiceEmailSnapshot struct {
	Send bool `json:"send"`
}

// InvoiceCustomerSnapshot is the customer data frozen at emission time.
type InvoiceCustomerSnapshot struct {
	Name    string                    `json:"name"`
	Email   string                    `json:"email"`
	TaxID   string                    `json:"taxId"`
	Address *InvoiceCustomerAddress   `json:"address"`
}

// InvoiceServiceSnapshot is the serviced item frozen at emission time.
type InvoiceServiceSnapshot struct {
	Code            string `json:"code"`
	Description     string `json:"description"`
	MunicipalityCode string `json:"municipalityCode"`
}

// InvoiceEvent is a status transition recorded for an invoice.
type InvoiceEvent struct {
	ID         string        `json:"id"`
	InvoiceID  string        `json:"invoiceId"`
	FromStatus InvoiceStatus `json:"fromStatus"`
	ToStatus   InvoiceStatus `json:"toStatus"`
	Reason     string        `json:"reason"`
	CreatedAt  string        `json:"createdAt"`
}

// InvoiceDownloadUrl is a short-lived signed URL for an invoice document.
type InvoiceDownloadUrl struct {
	DownloadURL string `json:"downloadUrl"`
	ExpiresAt   string `json:"expiresAt"`
	FileName    string `json:"fileName"`
	MimeType    string `json:"mimeType"`
}

// InvoiceEmailResult acknowledges an e-mail send request.
type InvoiceEmailResult struct {
	EmailID string `json:"emailId"`
	Status  string `json:"status"`
}
