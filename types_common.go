package corvonfse

// Pagination is the cursor pagination block returned by every list
// endpoint: { limit, hasMore, nextCursor }.
//
// NextCursor is an opaque base64url cursor and is empty when HasMore is
// false. Limit is 1–1000, default 50.
type Pagination struct {
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
	NextCursor string `json:"nextCursor"`
}

// ListParams are the common cursor parameters accepted by every list
// endpoint (companies, certificates, webhooks, credits, invoice events).
// A nil *ListParams is treated as the zero value.
type ListParams struct {
	// Limit is the page size (1–1000, default 50). Zero uses the default.
	Limit int
	// Cursor is an opaque cursor from a previous page's Pagination.NextCursor.
	Cursor string
}

// Environment distinguishes the test and live fiscal environments.
type Environment string

const (
	EnvironmentTest Environment = "test"
	EnvironmentLive Environment = "live"
)
