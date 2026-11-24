package corvonfse

import "encoding/json"

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

// Nullable is an optional JSON field that can also be sent as an explicit
// null. Use a nil *Nullable[T] to leave the field out of the request, Some(v)
// to set it and Null[T]() to clear it on the server.
type Nullable[T any] struct {
	Value  T
	IsNull bool
}

// Some builds a Nullable that sends v.
func Some[T any](v T) *Nullable[T] {
	return &Nullable[T]{Value: v}
}

// Null builds a Nullable that sends an explicit JSON null.
func Null[T any]() *Nullable[T] {
	return &Nullable[T]{IsNull: true}
}

// MarshalJSON encodes the null or the value.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}
