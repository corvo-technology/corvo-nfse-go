package corvonfse

// CreditBalance is the tenant credit balance. Balance is the plan notes left
// plus the loose (avulso) notes bought separately; plan notes of a lapsed plan
// count 0. There is no overage billing.
type CreditBalance struct {
	// Balance is PlanRemaining plus AvulsoBalance.
	Balance int `json:"balance"`
	// PlanRemaining is the notes left in the plan allowance.
	PlanRemaining int `json:"planRemaining"`
	// AvulsoBalance is the loose notes bought separately.
	AvulsoBalance int `json:"avulsoBalance"`
	HardLimit     int `json:"hardLimit"`
}

// CreditPlan is a plan of the credits catalog (GET /credits/plans). Slug is an
// open string (free, self-hosted, solo, equipe, empresa, escala today): do not
// hard-code the list, read the allowances from the plan.
type CreditPlan struct {
	ID                     string  `json:"id"`
	Slug                   string  `json:"slug"`
	Name                   string  `json:"name"`
	Description            string  `json:"description"`
	MonthlyCreditAllowance int     `json:"monthlyCreditAllowance"`
	DailyLimit             int     `json:"dailyLimit"` // 0 when the plan has no daily limit
	MaxActiveCompanies     int     `json:"maxActiveCompanies"`
	CreditUnitPriceCents   int     `json:"creditUnitPriceCents"`
	ThroughputRPS          float64 `json:"throughputRps"`
	IncludesWebhooks       bool    `json:"includesWebhooks"`
	IncludesXMLDownload    bool    `json:"includesXmlDownload"`
	IncludesPDFDownload    bool    `json:"includesPdfDownload"`
	IncludesAPIKeys        bool    `json:"includesApiKeys"`
	SupportTier            string  `json:"supportTier"`
	Active                 bool    `json:"active"`
	CreatedAt              string  `json:"createdAt"`
	UpdatedAt              string  `json:"updatedAt"`
}

// CreditTransaction is one entry of the credit ledger. Amount is in credits
// (not BRL); fields ending in Cents are in centavos.
type CreditTransaction struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	// Type is purchase, subscription_charge, topup, overage, invoice_authorized,
	// invoice_rejected, expiration, message_sent, refund, adjustment or bonus.
	Type                  string         `json:"type"`
	Amount                int            `json:"amount"`
	Description           string         `json:"description"`
	Source                string         `json:"source"` // subscription, avulso or bonus
	AmountBRLCents        int            `json:"amountBrlCents"`
	UnitPriceCentsApplied int            `json:"unitPriceCentsApplied"`
	PlanIDSnapshot        string         `json:"planIdSnapshot"`
	ReservedInvoiceID     string         `json:"reservedInvoiceId"`
	PeriodID              string         `json:"periodId"`
	IsReservation         bool           `json:"isReservation"`
	RelatedTransactionID  string         `json:"relatedTransactionId"`
	Metadata              map[string]any `json:"metadata"`
	CreatedAt             string         `json:"createdAt"`
}

// PurchasePreview prices a one-off purchase of loose notes
// (GET /credits/purchase/preview): R$ 0,99 each, at least 3 notes. Buying is
// done in the dashboard.
type PurchasePreview struct {
	Units          int `json:"units"`
	UnitPriceCents int `json:"unitPriceCents"`
	TotalCents     int `json:"totalCents"`
}

// PlanQuote is a public catalog entry (GET /plans): prices in centavos,
// monthly cycle only.
type PlanQuote struct {
	Slug                  string `json:"slug"`
	Name                  string `json:"name"`
	MonthlyAllowance      int    `json:"monthlyAllowance"`
	BasePriceCents        int    `json:"basePriceCents"`
	PerCreditOverageCents int    `json:"perCreditOverageCents"`
	MonthlyTotalCents     int    `json:"monthlyTotalCents"`
	Currency              string `json:"currency"`
}

// PlanRecommendation is the plan recommended for a monthly volume
// (GET /plans/quote). Note the snake_case wire names.
type PlanRecommendation struct {
	Volume              int    `json:"volume"`
	RecommendedPlan     string `json:"recommended_plan"`
	PlanName            string `json:"plan_name"`
	MonthlyAllowance    int    `json:"monthly_allowance"`
	OverageCredits      int    `json:"overage_credits"`
	OverageCostCents    int    `json:"overage_cost_cents"`
	BasePriceCents      int    `json:"base_price_cents"`
	TotalEstimatedCents int    `json:"total_estimated_cents"`
	Currency            string `json:"currency"`
	BillingCycle        string `json:"billing_cycle"`
}
