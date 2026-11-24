package corvonfse

// IntegrationProvider is a payment platform Corvo can connect to.
type IntegrationProvider struct {
	CredentialFields []IntegrationCredentialField `json:"credentialFields"`
	ID               string                       `json:"id"`
	Name             string                       `json:"name"`
	SupportedFlows   []string                     `json:"supportedFlows"`
	Version          string                       `json:"version"`
	// WebhookInstructions has the "en" and "pt" texts.
	WebhookInstructions map[string]string `json:"webhookInstructions"`
}

// IntegrationCredentialField describes one credential a provider needs.
type IntegrationCredentialField struct {
	Generate     bool   `json:"generate"`
	InWebhookURL bool   `json:"inWebhookUrl"`
	Key          string `json:"key"`
	Label        string `json:"label"`
	RevealOnce   bool   `json:"revealOnce"`
	Secret       bool   `json:"secret"`
}

// IntegrationConnection is a connection to a payment platform. State is
// draft, active, paused, reauth_required or disconnected. A connection follows
// the environment of the key that created it.
type IntegrationConnection struct {
	ActiveFrom        string      `json:"activeFrom"`
	CompanyID         string      `json:"companyId"`
	CreatedAt         string      `json:"createdAt"`
	Environment       Environment `json:"environment"`
	ExternalAccountID string      `json:"externalAccountId"`
	ID                string      `json:"id"`
	Provider          string      `json:"provider"`
	State             string      `json:"state"`
	WebhookURL        string      `json:"webhookUrl"`
}

// IntegrationConnectionCreated is the answer of IntegrationConnections.Create.
// Secrets are shown only once.
type IntegrationConnectionCreated struct {
	IntegrationConnection
	Secrets map[string]string `json:"secrets"`
}

// IntegrationConnectionCreateParams are the fields accepted by
// IntegrationConnections.Create.
type IntegrationConnectionCreateParams struct {
	CompanyID   string            `json:"companyId"`
	Provider    string            `json:"provider"`
	Credentials map[string]string `json:"credentials"`
	ActiveFrom  string            `json:"activeFrom,omitempty"`
	Environment Environment       `json:"environment,omitempty"`
}

// IntegrationEvent is an event received for a connection. State is received,
// processing, retry_pending, processed, ignored, needs_attention or dead.
type IntegrationEvent struct {
	Attempts     int    `json:"attempts"`
	CreatedAt    string `json:"createdAt"`
	DeliveryID   string `json:"deliveryId"`
	ExternalType string `json:"externalType"`
	ID           string `json:"id"`
	LastError    string `json:"lastError"`
	OperationID  string `json:"operationId"`
	State        string `json:"state"`
}

// IntegrationOperationCustomer is the customer of an operation.
type IntegrationOperationCustomer struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	TaxID       string `json:"taxId"`
	TaxIDMasked bool   `json:"taxIdMasked"`
}

// IntegrationOperation is a sale and its invoice status. CommercialState is
// pending, paid, refunded, disputed or canceled; FiscalStatus is none,
// blocked, needs_attention or submitted. Customer is nil when unknown.
type IntegrationOperation struct {
	AmountCents     int                           `json:"amountCents"`
	BlockReason     string                        `json:"blockReason"`
	CommercialState string                        `json:"commercialState"`
	CreatedAt       string                        `json:"createdAt"`
	Currency        string                        `json:"currency"`
	Customer        *IntegrationOperationCustomer `json:"customer"`
	FiscalStatus    string                        `json:"fiscalStatus"`
	ID              string                        `json:"id"`
	InvoiceID       string                        `json:"invoiceId"`
	OperationKey    string                        `json:"operationKey"`
	PaidAt          string                        `json:"paidAt"`
	UpdatedAt       string                        `json:"updatedAt"`
}

// IntegrationOperationRetry is the answer of IntegrationOperations.Retry.
// State is processed, ignored or needs_attention.
type IntegrationOperationRetry struct {
	LastError   string `json:"lastError"`
	OperationID string `json:"operationId"`
	State       string `json:"state"`
}

// IntegrationRule says how a product of the platform becomes an invoice.
type IntegrationRule struct {
	AllowAnonymousCustomer bool           `json:"allowAnonymousCustomer"`
	CancelOnRefund         bool           `json:"cancelOnRefund"`
	ExternalProductID      string         `json:"externalProductId"`
	ID                     string         `json:"id"`
	SendEmail              bool           `json:"sendEmail"`
	Service                map[string]any `json:"service"`
}

// IntegrationRuleInput is one rule of IntegrationConnections.ReplaceRules.
type IntegrationRuleInput struct {
	SendEmail              bool                `json:"sendEmail"`
	Service                InvoiceServiceInput `json:"service"`
	AllowAnonymousCustomer *bool               `json:"allowAnonymousCustomer,omitempty"`
	CancelOnRefund         *bool               `json:"cancelOnRefund,omitempty"`
	// ExternalProductID matches a product of the platform; nil matches any.
	ExternalProductID *string `json:"externalProductId,omitempty"`
}

// IntegrationCustomerParams are the fields accepted by
// IntegrationOperations.CompleteCustomer.
type IntegrationCustomerParams struct {
	TaxID string `json:"taxId"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}
