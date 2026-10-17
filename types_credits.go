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
	ID                     string `json:"id"`
	Slug                   string `json:"slug"`
	Name                   string `json:"name"`
	Description            string `json:"description"`
	MonthlyCreditAllowance int    `json:"monthlyCreditAllowance"`
	DailyLimit             int    `json:"dailyLimit"` // 0 when the plan has no daily limit
	MaxActiveCompanies     int    `json:"maxActiveCompanies"`
	CreditUnitPriceCents   int    `json:"creditUnitPriceCents"`
	ThroughputRPS          int    `json:"throughputRps"`
	IncludesWebhooks       bool   `json:"includesWebhooks"`
	IncludesXMLDownload    bool   `json:"includesXmlDownload"`
	IncludesPDFDownload    bool   `json:"includesPdfDownload"`
	IncludesAPIKeys        bool   `json:"includesApiKeys"`
	SupportTier            string `json:"supportTier"`
	Active                 bool   `json:"active"`
	CreatedAt              string `json:"createdAt"`
	UpdatedAt              string `json:"updatedAt"`
}

// CreditTransaction is one entry of the credit ledger. Amount is in credits
// (not BRL); fields ending in Cents are in centavos.
type CreditTransaction struct {
	ID                    string  `json:"id"`
	TenantID              string  `json:"tenantId"`
	Type                  string  `json:"type"`
	Amount                float64 `json:"amount"`
	Description           string  `json:"description"`
	Source                string  `json:"source"` // subscription, avulso or bonus
	AmountBRLCents        int     `json:"amountBrlCents"`
	UnitPriceCentsApplied int     `json:"unitPriceCentsApplied"`
	PlanIDSnapshot        string  `json:"planIdSnapshot"`
	ReservedInvoiceID     string  `json:"reservedInvoiceId"`
	PeriodID              string  `json:"periodId"`
	IsReservation         bool    `json:"isReservation"`
	RelatedTransactionID  string  `json:"relatedTransactionId"`
	CreatedAt             string  `json:"createdAt"`
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
