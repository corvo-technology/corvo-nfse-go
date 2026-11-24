package corvonfse

// CnpjProfile is the registration data of a CNPJ (Receita Federal).
// MEI and Simples are nil when the registry does not say.
type CnpjProfile struct {
	Active           bool          `json:"active"`
	Address          PostalAddress `json:"address"`
	LegalName        string        `json:"legalName"`
	MEI              *bool         `json:"mei"`
	MunicipalityCode string        `json:"municipalityCode"`
	Simples          *bool         `json:"simples"`
	Situation        string        `json:"situation"`
	TradeName        string        `json:"tradeName"`
}

// PostalAddress is a full Brazilian address.
type PostalAddress struct {
	City       string `json:"city"`
	Complement string `json:"complement"`
	District   string `json:"district"`
	Number     string `json:"number"`
	State      string `json:"state"`
	Street     string `json:"street"`
	Zip        string `json:"zip"`
}

// SavedCustomer is a customer saved from accepted invoices.
type SavedCustomer struct {
	CreatedAt  string `json:"createdAt"`
	Email      string `json:"email"`
	LastUsedAt string `json:"lastUsedAt"`
	Name       string `json:"name"`
	TaxID      string `json:"taxId"`
	UpdatedAt  string `json:"updatedAt"`
}

// CustomerSearchParams are the filters accepted by Customers.Search and
// Customers.Iterate. A nil pointer is treated as the zero value.
type CustomerSearchParams struct {
	// Search matches part of the name or the first digits of the CPF/CNPJ
	// (punctuation is ignored). Up to 100 characters.
	Search string
	Limit  int
	// Cursor is the opaque cursor of the previous page (here, the offset of
	// the next page): pass it back untouched.
	Cursor string
}

// CustomerUpdateParams are the fields accepted by Customers.Update. A nil
// field is left out; Null[string]() clears the e-mail.
type CustomerUpdateParams struct {
	Name  string            `json:"name,omitempty"`
	Email *Nullable[string] `json:"email,omitempty"`
}

// ServiceFavorite is a favorite service saved for a company.
type ServiceFavorite struct {
	AmountCents      *int   `json:"amountCents"`
	CompanyID        string `json:"companyId"`
	CreatedAt        string `json:"createdAt"`
	Description      string `json:"description"`
	ID               string `json:"id"`
	Label            string `json:"label"`
	MunicipalityCode string `json:"municipalityCode"`
	ServiceCode      string `json:"serviceCode"`
	TenantID         string `json:"tenantId"`
}

// ServiceFavoriteCreateParams are the fields accepted by
// ServiceFavorites.Create.
type ServiceFavoriteCreateParams struct {
	Label       string `json:"label"`
	ServiceCode string `json:"serviceCode"`
	Description string `json:"description"`
	// AmountCents is the default amount, in centavos (nil for none).
	AmountCents      *int   `json:"amountCents,omitempty"`
	MunicipalityCode string `json:"municipalityCode,omitempty"`
}

// AccountantSettings is the accountant e-mail of the monthly package.
// Email is empty when the monthly send is off.
type AccountantSettings struct {
	Email         string `json:"email"`
	LastSentMonth string `json:"lastSentMonth"`
}

// AccountantSettingsUpdate is the answer of AccountantPackage.Update.
// TestEmailSent reports that a test e-mail went to the new address.
type AccountantSettingsUpdate struct {
	Email         string `json:"email"`
	LastSentMonth string `json:"lastSentMonth"`
	TestEmailSent bool   `json:"testEmailSent"`
}

// AccountantPackageUpdateParams are the fields accepted by
// AccountantPackage.Update. Null[string]() turns the monthly send off.
type AccountantPackageUpdateParams struct {
	AccountantEmail *Nullable[string] `json:"accountantEmail,omitempty"`
}

// AccountantSendResult is the answer of AccountantPackage.Send. Status is
// sent, skipped-test, skipped-no-email, skipped-empty, skipped-too-large or
// skipped-already-sent.
type AccountantSendResult struct {
	Status string `json:"status"`
}

// AccountantSend is one attempt to send the package to the accountant.
// Status is sent, skipped-empty, skipped-too-large or failed; Trigger is
// scheduled, manual or mcp-or-api.
type AccountantSend struct {
	CreatedAt    string `json:"createdAt"`
	ID           string `json:"id"`
	InvoiceCount int    `json:"invoiceCount"`
	Month        string `json:"month"`
	Recipient    string `json:"recipient"`
	Status       string `json:"status"`
	TotalCents   int    `json:"totalCents"`
	Trigger      string `json:"trigger"`
}

// PackageSummary is the count and total of the month package.
type PackageSummary struct {
	CompanyName  string  `json:"companyName"`
	CompanyTaxID string  `json:"companyTaxId"`
	Count        int     `json:"count"`
	Limit        int     `json:"limit"`
	Month        string  `json:"month"`
	Total        float64 `json:"total"`
}

// DPSSequenceInfo is the DPS numbering of a company.
type DPSSequenceInfo struct {
	LastNumber int    `json:"lastNumber"`
	Series     string `json:"series"`
}

// DPSSequenceSetParams are the fields accepted by DPSSequence.Set. LastNumber
// is the last DPS number issued elsewhere (nil keeps the current one). A number
// below the current one fails with DPS_NUMBER_BELOW_CURRENT.
type DPSSequenceSetParams struct {
	LastNumber *int `json:"lastNumber,omitempty"`
}

// PortalStatus tells whether the national NFS-e portal is answering for the
// environment of the key.
type PortalStatus struct {
	Available   bool        `json:"available"`
	Environment Environment `json:"environment"`
	RetryAt     string      `json:"retryAt"`
	Since       string      `json:"since"`
}
