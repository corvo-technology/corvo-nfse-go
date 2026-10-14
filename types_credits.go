package corvonfse

// CreditBalance is the tenant credit balance.
type CreditBalance struct {
	TenantID           string  `json:"tenantId"`
	Balance            float64 `json:"balance"`
	OverageUnits       int     `json:"overageUnits"`
	OverageAmountCents int     `json:"overageAmountCents"`
	HardLimit          int     `json:"hardLimit"`
}

// CreditPlan is a purchasable credit plan.
type CreditPlan struct {
	ID                   string   `json:"id"`
	Slug                 string   `json:"slug"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	DailyLimit           int      `json:"dailyLimit"`
	MaxActiveCompanies   int      `json:"maxActiveCompanies"`
	CreditUnitPriceCents int      `json:"creditUnitPriceCents"`
	ThroughputRPS        int      `json:"throughputRps"`
	IncludesWebhooks     bool     `json:"includesWebhooks"`
	IncludesXMLDownload  bool     `json:"includesXmlDownload"`
	IncludesPDFDownload  bool     `json:"includesPdfDownload"`
	IncludesAPIKeys      bool     `json:"includesApiKeys"`
	SupportTier          string   `json:"supportTier"`
	Active               bool     `json:"active"`
	CreatedAt            string   `json:"createdAt"`
	UpdatedAt            string   `json:"updatedAt"`
}

// CreditTransaction is one entry of the credit ledger.
type CreditTransaction struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Amount        float64     `json:"amount"`
	CreditDelta   float64     `json:"creditDelta"`
	BalanceAfter  float64     `json:"balanceAfter"`
	Environment   Environment `json:"environment"`
	Source        string      `json:"source"`
	ReferenceType string      `json:"referenceType"`
	ReferenceID   string      `json:"referenceId"`
	CreatedAt     string      `json:"createdAt"`
}
