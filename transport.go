package corvonfse

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.corvo.tech"
	defaultPrefix  = "v1"
	defaultTimeout = 30 * time.Second
	maxErrorBody   = 1 << 20 // 1 MiB cap when reading error response bodies
)

// Unwrap selects how the response envelope is unwrapped before decoding:
// "data" returns body.data, "data+props" returns body.data.props, and
// "none" returns the whole body (used by list endpoints that return
// { data, pagination } together, and by the raw health endpoints).
type Unwrap string

const (
	UnwrapData      Unwrap = "data"
	UnwrapDataProps Unwrap = "data+props"
	UnwrapNone      Unwrap = "none"
)

// DownloadedFile is a binary response (invoice PDF/XML) with its metadata
// parsed from the response headers.
type DownloadedFile struct {
	Buffer   []byte
	MimeType string
	FileName string
}

// Transport performs HTTP requests against the Corvo NFS-e API. It is
// created by NewClient; use the resource fields on Client instead of
// touching it directly.
type Transport struct {
	apiKey string
	base   string // normalized base URL without trailing slashes
	prefix string // "" or "v1"
	client *http.Client
}

type requestSpec struct {
	method         string
	path           string
	query          url.Values
	jsonBody       any       // marshaled as JSON when body is nil
	body           io.Reader // raw body (multipart); takes precedence
	contentType    string    // Content-Type for the raw body
	idempotencyKey string
	unwrap         Unwrap
}

var filenameRE = regexp.MustCompile(`filename="([^"]+)"`)

// newTransport validates the options (panicking on misconfiguration, the
// documented failure mode of NewClient) and builds a pooled HTTP client.
func newTransport(apiKey, baseURL, prefix string, timeout time.Duration, allowInsecureHTTP bool, httpClient *http.Client) *Transport {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		panic(fmt.Sprintf("Corvo: invalid baseUrl: %s", baseURL))
	}
	if parsed.Scheme == "http" && !allowInsecureHTTP && !isPrivateHost(parsed.Hostname()) {
		panic(fmt.Sprintf(
			"Corvo: http:// is only allowed for loopback/private hosts (got %s). Use https:// or set AllowInsecureHTTP: true.",
			parsed.Hostname(),
		))
	}
	if httpClient == nil {
		if timeout == 0 {
			timeout = defaultTimeout
		}
		transport := http.DefaultTransport
		if base, ok := http.DefaultTransport.(*http.Transport); ok {
			transport = base.Clone()
		}
		httpClient = &http.Client{
			Timeout:   timeout,
			Transport: transport,
		}
	}
	return &Transport{
		apiKey: apiKey,
		base:   strings.TrimRight(baseURL, "/"),
		prefix: strings.Trim(prefix, "/"),
		client: httpClient,
	}
}

// isPrivateHost reports whether host is loopback or a private/link-local
// address, where plain http:// is allowed (mirrors the Node SDK).
func isPrivateHost(host string) bool {
	h := strings.ToLower(host)
	if h == "localhost" || h == "[::1]" || h == "::1" || strings.HasSuffix(h, ".localhost") {
		return true
	}
	ip := net.ParseIP(h)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		return true
	}
	v4 := ip.To4()
	return v4 != nil && v4[0] == 0 // 0.0.0.0/8
}

// queryAdd sets a string query parameter, skipping empty values (the Node
// SDK skips undefined values the same way).
func queryAdd(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

// queryAddInt sets an integer query parameter, skipping zero values.
func queryAddInt(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, strconv.Itoa(value))
	}
}

// do executes the request and maps failures to *CorvoError. On success the
// caller owns resp.Body and must close it.
func (t *Transport) do(ctx context.Context, spec requestSpec) (*http.Response, *CorvoError) {
	if ctx == nil {
		ctx = context.Background()
	}

	var body io.Reader
	switch {
	case spec.body != nil:
		body = spec.body
	case spec.jsonBody != nil:
		encoded, err := json.Marshal(spec.jsonBody)
		if err != nil {
			return nil, &CorvoError{
				Status:  0,
				Code:    "encode_error",
				Message: "Failed to encode request body: " + err.Error(),
			}
		}
		body = bytes.NewReader(encoded)
	}

	target := t.base
	if t.prefix != "" {
		target += "/" + t.prefix
	}
	target += spec.path
	if len(spec.query) > 0 {
		target += "?" + spec.query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, spec.method, target, body)
	if err != nil {
		return nil, &CorvoError{
			Status:  0,
			Code:    "network_error",
			Message: "Network request failed: " + err.Error(),
		}
	}
	req.Header.Set("X-API-Key", t.apiKey)
	req.Header.Set("Accept", "application/json")
	if spec.contentType != "" {
		req.Header.Set("Content-Type", spec.contentType)
	} else if spec.jsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if spec.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", spec.idempotencyKey)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, transportError(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		return nil, toError(resp)
	}
	return resp, nil
}

// transportError maps network failures and timeouts to *CorvoError with
// Status 0 and code "timeout" or "network_error".
func transportError(err error) *CorvoError {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return &CorvoError{
			Status:  0,
			Code:    "timeout",
			Message: "Request timed out",
		}
	}
	message := err.Error()
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Err != nil {
		message = urlErr.Err.Error()
	}
	return &CorvoError{
		Status:  0,
		Code:    "network_error",
		Message: "Network request failed: " + message,
	}
}

// toError builds the typed error from a non-2xx response, honouring the
// body shape { statusCode, code, message, details?, requestId } plus the
// X-Request-Id and Retry-After headers.
func toError(resp *http.Response) *CorvoError {
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))

	var parsed struct {
		Code       string          `json:"code"`
		Message    json.RawMessage `json:"message"`
		Error      string          `json:"error"`
		Details    []Detail        `json:"details"`
		RequestID  string          `json:"requestId"`
		RetryAfter *int            `json:"retryAfter"`
	}
	_ = json.Unmarshal(payload, &parsed)

	code := parsed.Code
	if code == "" {
		code = fmt.Sprintf("http_%d", resp.StatusCode)
	}

	message := decodeErrorMessage(parsed.Message)
	if message == "" {
		message = parsed.Error
	}
	if message == "" {
		message = fmt.Sprintf("Request failed with status %d", resp.StatusCode)
	}

	requestID := parsed.RequestID
	if requestID == "" {
		requestID = resp.Header.Get("X-Request-Id")
	}

	retryAfter := 0
	if parsed.RetryAfter != nil {
		retryAfter = *parsed.RetryAfter
	} else if header := resp.Header.Get("Retry-After"); header != "" {
		if seconds, err := strconv.Atoi(header); err == nil {
			retryAfter = seconds
		}
	}

	return &CorvoError{
		Status:     resp.StatusCode,
		Code:       code,
		Message:    message,
		Details:    parsed.Details,
		RequestID:  requestID,
		RetryAfter: retryAfter,
	}
}

// decodeErrorMessage accepts a string or an array of strings from the
// `message` field of an error body; anything else yields "".
func decodeErrorMessage(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}
	var asArray []string
	if err := json.Unmarshal(raw, &asArray); err == nil {
		return strings.Join(asArray, "; ")
	}
	return ""
}

func invalidJSON(status int, requestID string) *CorvoError {
	return &CorvoError{
		Status:    status,
		Code:      "invalid_json",
		Message:   "Invalid JSON in response",
		RequestID: requestID,
	}
}

// readSuccess reads and closes a 2xx response body. Empty bodies (204 or
// blank 200) yield a nil payload, which decodes to the zero value.
func readSuccess(resp *http.Response) ([]byte, *CorvoError) {
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, transportError(err)
	}
	if len(payload) == 0 {
		return nil, nil
	}
	return payload, nil
}

// unwrap extracts the payload from the response envelope. The second
// result reports a null/absent result (decoded as the zero value), the
// third a malformed JSON body. A valid JSON body that is not an object is
// treated as "no data" — mirroring the Node SDK, which reads .data off any
// value without failing.
func unwrap(payload []byte, mode Unwrap) (json.RawMessage, bool, error) {
	if mode == "" {
		mode = UnwrapData
	}
	if mode == UnwrapNone {
		return payload, false, nil
	}
	if !json.Valid(payload) {
		return nil, false, errors.New("invalid json")
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, true, nil // not an object: no `data` key
	}
	data, ok := envelope["data"]
	if !ok || isEmptyJSON(data) {
		return nil, true, nil
	}
	if mode == UnwrapDataProps {
		var propsEnvelope map[string]json.RawMessage
		if err := json.Unmarshal(data, &propsEnvelope); err != nil {
			return nil, true, nil // data is not an object: no `props` key
		}
		props, ok := propsEnvelope["props"]
		if !ok || isEmptyJSON(props) {
			return nil, true, nil
		}
		return props, false, nil
	}
	return data, false, nil
}

func isEmptyJSON(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || string(trimmed) == "null"
}

// requestJSON performs a request and decodes the (unwrapped) JSON response
// into T. A 204/empty response or a null envelope yields the zero value.
func requestJSON[T any](ctx context.Context, t *Transport, spec requestSpec) (T, *CorvoError) {
	var out T

	resp, cerr := t.do(ctx, spec)
	if cerr != nil {
		return out, cerr
	}
	payload, cerr := readSuccess(resp)
	if cerr != nil {
		return out, cerr
	}
	if payload == nil {
		return out, nil
	}
	requestID := resp.Header.Get("X-Request-Id")
	decoded, isNull, err := unwrap(payload, spec.unwrap)
	if err != nil {
		return out, invalidJSON(resp.StatusCode, requestID)
	}
	if isNull {
		return out, nil
	}
	if err := json.Unmarshal(decoded, &out); err != nil {
		return out, invalidJSON(resp.StatusCode, requestID)
	}
	return out, nil
}

// requestFile performs a GET and returns the raw bytes plus the mime type
// and file name advertised in the response headers.
func requestFile(ctx context.Context, t *Transport, method, path string) (DownloadedFile, *CorvoError) {
	var out DownloadedFile

	resp, cerr := t.do(ctx, requestSpec{method: method, path: path})
	if cerr != nil {
		return out, cerr
	}
	payload, cerr := readSuccess(resp)
	if cerr != nil {
		return out, cerr
	}
	out.Buffer = payload
	out.MimeType = resp.Header.Get("Content-Type")
	if match := filenameRE.FindStringSubmatch(resp.Header.Get("Content-Disposition")); match != nil {
		out.FileName = match[1]
	}
	return out, nil
}

// rawPage is the list envelope: { data: T[], pagination }.
type rawPage struct {
	Data       []json.RawMessage `json:"data"`
	Pagination Pagination        `json:"pagination"`
}

// requestPage fetches one list page and decodes each item, unwrapping the
// props envelope ({ props: T }) when the endpoint serializes items that way.
func requestPage[T any](ctx context.Context, t *Transport, path string, query url.Values, propsItems bool) ([]T, Pagination, *CorvoError) {
	page, cerr := requestJSON[rawPage](ctx, t, requestSpec{
		method: http.MethodGet,
		path:   path,
		query:  query,
		unwrap: UnwrapNone,
	})
	if cerr != nil {
		return nil, Pagination{}, cerr
	}

	items := make([]T, 0, len(page.Data))
	for _, raw := range page.Data {
		if propsItems {
			var envelope map[string]json.RawMessage
			if err := json.Unmarshal(raw, &envelope); err != nil {
				return nil, Pagination{}, invalidJSON(http.StatusOK, "")
			}
			props, ok := envelope["props"]
			if !ok || isEmptyJSON(props) {
				return nil, Pagination{}, invalidJSON(http.StatusOK, "")
			}
			raw = props
		}
		var item T
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, Pagination{}, invalidJSON(http.StatusOK, "")
		}
		items = append(items, item)
	}
	return items, page.Pagination, nil
}

// iteratePages walks every page of a list endpoint, invoking fn per item
// until all pages are consumed or fn returns false. A page failure —
// including the first — is surfaced as *CorvoError: iterate is the one
// place the SDK reports failure outside the (T, *CorvoError) shape.
func iteratePages[T any](ctx context.Context, t *Transport, path string, query url.Values, propsItems bool, fn func(T) bool) *CorvoError {
	if query == nil {
		query = url.Values{}
	}
	for {
		items, pagination, cerr := requestPage[T](ctx, t, path, query, propsItems)
		if cerr != nil {
			return cerr
		}
		for _, item := range items {
			if !fn(item) {
				return nil
			}
		}
		if !pagination.HasMore || pagination.NextCursor == "" {
			return nil
		}
		query = cloneValues(query)
		query.Set("cursor", pagination.NextCursor)
	}
}

func cloneValues(in url.Values) url.Values {
	out := make(url.Values, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}
