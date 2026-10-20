package corvonfse

// WebhookEvent is an event type a webhook can subscribe to. The accepted
// values are listed in WebhookEvents.
type WebhookEvent string

// WebhookEvents are the eleven event types accepted when creating a
// webhook. The webhook.test event is only sent by Webhooks.Test and cannot be
// subscribed.
var WebhookEvents = [...]WebhookEvent{
	"nfse.invoice.authorized",
	"nfse.invoice.rejected",
	"nfse.invoice.failed",
	"nfse.invoice.reconciliation_required",
	"nfse.invoice.manual_review",
	"nfse.invoice.documents_ready",
	"nfse.invoice.email_sent",
	"nfse.invoice.email_failed",
	"nfse.invoice.fiscal_discrepancy",
	"nfse.invoice.cancelled",
	"nfse.invoice.cancellation_rejected",
}

// WebhookCreateParams are the fields accepted by Webhooks.Create.
type WebhookCreateParams struct {
	URL    string         `json:"url"`
	Events []WebhookEvent `json:"events"`
}

// Webhook is a registered webhook endpoint.
type Webhook struct {
	ID        string         `json:"id"`
	URL       string         `json:"url"`
	Events    []WebhookEvent `json:"events"`
	Active    bool           `json:"active"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

// WebhookWithSecret is returned by Webhooks.Create only — the signing
// secret is shown once and cannot be retrieved again.
type WebhookWithSecret struct {
	ID        string         `json:"id"`
	URL       string         `json:"url"`
	Events    []WebhookEvent `json:"events"`
	Active    bool           `json:"active"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
	Secret    string         `json:"secret"`
}

// DeliveryListParams are the filters accepted by Deliveries.List and
// Deliveries.Iterate. A nil *DeliveryListParams is treated as the zero
// value.
type DeliveryListParams struct {
	// Hours limits the window to the last N hours. Zero means no window.
	Hours int
	// Environment filters by EnvironmentTest or EnvironmentLive.
	Environment Environment
	Limit       int
	Cursor      string
}

// WebhookDelivery is one delivery attempt of a webhook event.
type WebhookDelivery struct {
	ID              string         `json:"id"`
	TenantID        string         `json:"tenantId"`
	WebhookConfigID string         `json:"webhookConfigId"`
	EventType       string         `json:"eventType"`
	Environment     Environment    `json:"environment"`
	URL             string         `json:"url"`
	Payload         map[string]any `json:"payload"`
	StatusCode      int            `json:"statusCode"`
	ResponseBody    any            `json:"responseBody"`
	ErrorMessage    string         `json:"errorMessage"`
	Status          string         `json:"status"`
	DurationMS      int            `json:"durationMs"`
	RetryCount      int            `json:"retryCount"`
	NextRetryAt     string         `json:"nextRetryAt"`
	LastError       string         `json:"lastError"`
	CreatedAt       string         `json:"createdAt"`
}

// WebhookDeliveryStats are the delivery counters for a time window.
type WebhookDeliveryStats struct {
	Total   int `json:"total"`
	Sent    int `json:"sent"`
	Failed  int `json:"failed"`
	Pending int `json:"pending"`
	Delayed int `json:"delayed"`
}
