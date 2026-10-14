package corvonfse

// Detail is a single field-level validation error returned by the API
// inside the `details` array of an error body.
type Detail struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CorvoError is the error type returned by every SDK operation.
//
// It is never nil on success: callers check `err != nil`. API, network and
// timeout failures are returned as *CorvoError — the SDK never panics for
// them (NewClient panics only on misconfiguration: missing API key or a
// public http:// base URL).
//
// Status is the HTTP status (0 for network/timeout). Code is stable
// (ERROR_VALIDATION, ERROR_COMPANY_NOT_FOUND, ...), falling back to
// "http_<status>" for unrecognised failures and to "network_error" or
// "timeout" for transport failures. RequestID matches the X-Request-Id
// header. RetryAfter is populated on 429 responses (seconds).
type CorvoError struct {
	Status     int      `json:"statusCode"`
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Details    []Detail `json:"details,omitempty"`
	RequestID  string   `json:"requestId"`
	RetryAfter int      `json:"retryAfter,omitempty"`
}

// Error implements the error interface as "<code>: <message>".
func (e *CorvoError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}
