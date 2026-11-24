package corvonfse

// BillingPlanResponse is the current plan, subscription and usage. The
// nested plan/subscription/usage objects use snake_case field names —
// that is the wire format returned by the API and it is passed through
// unchanged.
type BillingPlanResponse struct {
	Plan         BillingPlan         `json:"plan"`
	Subscription BillingSubscription `json:"subscription"`
	Usage        BillingUsage        `json:"usage"`
}

// BillingPlan is the plan descriptor. Fields use snake_case wire names.
type BillingPlan struct {
	// Slug is an open string: free, self-hosted, solo, equipe, empresa or escala
	// today. Do not hard-code the list.
	Slug               string `json:"slug"`
	Name               string `json:"name"`
	MonthlyAllowance   int    `json:"monthly_allowance"`
	DailyLimit         int    `json:"daily_limit"`
	MaxActiveCompanies int    `json:"max_active_companies"`
	MonthlyPriceCents  int    `json:"monthly_price_cents"`
}

// BillingSubscription is the subscription state. Fields use snake_case
// wire names.
type BillingSubscription struct {
	Status            string `json:"status"`
	PaidUntil         string `json:"paid_until"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
	ScheduledPlan     string `json:"scheduled_plan"`
	Blocked           bool   `json:"blocked"`
	// StripePortalAvailable is true when the customer portal can be opened.
	StripePortalAvailable bool `json:"stripe_portal_available"`
	// StripeSubscription is true when the subscription is billed by Stripe.
	StripeSubscription bool `json:"stripe_subscription"`
	// PixSubscription is the Pix subscription, or nil when there is none.
	PixSubscription *BillingPixSubscription `json:"pix_subscription"`
}

// BillingPixSubscription is the Pix subscription of the workspace.
type BillingPixSubscription struct {
	BrCode         string `json:"brCode"`
	PaymentID      string `json:"paymentId"`
	PaymentLinkURL string `json:"paymentLinkUrl"`
	PlanSlug       string `json:"planSlug"`
	// Status is "active" or "pending".
	Status string `json:"status"`
}

// BillingUsage is the usage block of the plan response. Fields use
// snake_case wire names.
type BillingUsage struct {
	// AvulsoBalance is the loose notes bought separately.
	AvulsoBalance      int    `json:"avulso_balance"`
	HardLimit          int    `json:"hard_limit"`
	CreditsUsed        int    `json:"credits_used"`
	CreditsReserved    int    `json:"credits_reserved"`
	CreditsRemaining   int    `json:"credits_remaining"`
	MonthlyAllowance   int    `json:"monthly_allowance"`
	DailyUsage         int    `json:"daily_usage"`
	DailyLimit         int    `json:"daily_limit"`
	ActiveCompanies    int    `json:"active_companies"`
	MaxActiveCompanies int    `json:"max_active_companies"`
	PeriodStart        string `json:"period_start"`
	PeriodEnd          string `json:"period_end"`
}

// QuotaStatus is the camelCase quota view returned by Billing.Usage.
type QuotaStatus struct {
	CanEmit bool `json:"canEmit"`
	// Reason is set when CanEmit is false: no_credits, daily_limit_exceeded,
	// company_limit_exceeded or plan_blocked.
	Reason             string `json:"reason"`
	CreditsRemaining   int    `json:"creditsRemaining"`
	CreditsUsed        int    `json:"creditsUsed"`
	CreditsReserved    int    `json:"creditsReserved"`
	MonthlyAllowance   int    `json:"monthlyAllowance"`
	DailyUsage         int    `json:"dailyUsage"`
	DailyLimit         int    `json:"dailyLimit"`
	ActiveCompanies    int    `json:"activeCompanies"`
	MaxActiveCompanies int    `json:"maxActiveCompanies"`
	PeriodStart        string `json:"periodStart"`
	PeriodEnd          string `json:"periodEnd"`
}
