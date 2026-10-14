# corvo-nfse-go

Idiomatic Go SDK for the **Corvo NFS-e API** — issue, track and serve Brazilian
fiscal invoices (NFS-e) with cursor-paginated lists, webhook verification and
typed errors.

- **Go 1.22+**, `net/http` only — **zero runtime dependencies**
- 28 API operations + 2 health checks, mirroring the
  [`corvo-nfse-node`](https://github.com/corvo-technology/corvo-nfse-node) contract
- Result-style `(T, *CorvoError)` returns — never panics for API, network or
  timeout failures
- `X-API-Key` auth, `https://api.corvo.tech` default, `v1` prefix, self-hosted
  friendly
- HMAC-SHA256 webhook verification, constant-time

```text
github.com/corvo-technology/corvo-nfse-go   (import alias: corvonfse)
```

## Install

```sh
go get github.com/corvo-technology/corvo-nfse-go
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

func main() {
	// APIKey falls back to CORVO_API_KEY; BaseURL to CORVO_BASE_URL
	// (default https://api.corvo.tech).
	client := corvonfse.NewClient(corvonfse.Options{
		APIKey: "corvo_live_...",
	})

	invoice, err := client.Invoices.Create(context.Background(), corvonfse.InvoiceCreateParams{
		CompanyID: "09947127-2702-4f6d-9da8-cb463608d812",
		Customer: corvonfse.InvoiceCustomerInput{
			Name:  "Cliente Exemplo",
			TaxID: "52998224725",
		},
		Service: corvonfse.InvoiceServiceInput{
			Code:        "010101",
			Description: "Software development",
		},
		Amount:         150.50,
		CompetenceDate: "2026-10-12T00:00:00.000Z",
	}, "") // pass an idempotency key to safely retry
	if err != nil {
		log.Fatalf("create: %v", err) // ERROR_* codes, status, details, requestId
	}
	fmt.Println(invoice.ID, invoice.Status)
}
```

### Resources

| Field | Operations |
| --- | --- |
| `client.Companies` | `Create`, `Get`, `Update`, `Remove`, `List`, `Iterate` |
| `client.Certificates` | `Upload`, `List`, `Remove` |
| `client.Invoices` | `Create`, `Get`, `List`, `Iterate`, `Events`, `PDF`, `XML`, `DownloadURL`, `SendEmail` |
| `client.Webhooks` | `List`, `Create`, `Remove`, `Test` + `Deliveries.List/Iterate/Stats/Retry` |
| `client.Credits` | `Balance`, `Plans`, `History`, `IterateHistory` |
| `client.Billing` | `Plan`, `Usage` |
| `client.Health` | `Live`, `Ready` |

Every method takes a `context.Context` first. Every API call returns
`(T, *CorvoError)` — check `err != nil`, then read `Status`, `Code`,
`Message`, `Details`, `RequestID` and `RetryAfter`.

### Lists and iteration

`List` returns the page items plus the pagination block; `Iterate` walks every
page for you and surfaces a first-page failure as `*CorvoError`:

```go
items, page, err := client.Companies.List(ctx, &corvonfse.ListParams{Limit: 100})
if err != nil {
	// handle
}
fmt.Println(len(items), page.HasMore, page.NextCursor)

err = client.Invoices.Iterate(ctx, &corvonfse.InvoiceListParams{
	Status: corvonfse.InvoiceStatusAuthorized,
}, func(inv corvonfse.Invoice) bool {
	fmt.Println(inv.ID)
	return true // return false to stop early
})
```

### Binary downloads

```go
pdf, err := client.Invoices.PDF(ctx, invoice.ID)
if err == nil {
	_ = os.WriteFile("nfse.pdf", pdf.Buffer, 0o644)
	fmt.Println(pdf.MimeType, pdf.FileName) // from the response headers
}
```

### Webhook verification

```go
body, _ := io.ReadAll(r.Body)
if !corvonfse.VerifySignature(body, r.Header.Get("X-Corbo-Signature"), secret) {
	http.Error(w, "invalid signature", http.StatusBadRequest)
	return
}
```

`VerifySignature` computes HMAC-SHA256 over the raw body and compares in
constant time (`crypto/subtle`). Malformed headers, an empty secret or a
mismatch return `false` — it never panics.

### Configuration

| Option | Env fallback | Default |
| --- | --- | --- |
| `APIKey` | `CORVO_API_KEY` | required (panic if missing) |
| `BaseURL` | `CORVO_BASE_URL` | `https://api.corvo.tech` |
| `Prefix` | — | `v1` |
| `Timeout` | — | 30s per request |
| `AllowInsecureHTTP` | — | `false` (loopback/private hosts always allowed) |
| `HTTPClient` | — | pooled `*http.Client` (SDK default) |

`NewClient` panics only on misconfiguration — a missing API key, an invalid
base URL, or plain `http://` against a public host — since those are programmer
errors detected at startup. Runtime failures are always returned values.

The CLI, the official OpenAPI spec and Postman collection ship with the
canonical [Node SDK](https://github.com/corvo-technology/corvo-nfse-node) and
work with any of the SDKs; this package is the Go library.

## The contract (in-depth)

Nine SDKs, one behavior. This SDK implements the same analysis as every other Corvo SDK, mirrored from the canonical Node SDK.

**Auth.** Every request carries `X-API-Key` (never `Authorization: Bearer`). Keys are scoped per tenant; the base URL defaults to `https://api.corvo.tech` and is overridable for self-hosted deployments. All routes live under the `v1` prefix.

**Response envelope.** Endpoints return one of three shapes, and the SDK unwraps them idiomatically: `{ "data": ... }` (single resource), `{ "data": { "props": ... } }` (resource behind a `props` wrapper), or no envelope (health checks, raw downloads). List endpoints return `{ "data": [...], "pagination": { "limit", "hasMore", "nextCursor" } }`. A `204 No Content` maps to an empty result.

**Errors never throw.** Network failures, timeouts, non-2xx responses, and malformed JSON all resolve to a typed error carrying `status`, `code`, `message`, `details`, `requestId`, and `retryAfter` — returned alongside (or as) the result, never raised. `code` falls back to `http_<status>` when the body doesn't carry one; transport-level failures use `network_error` / `timeout` / `invalid_json`. The only operation that raises is the `iterate` helper, and only on the first page's failure.

**Pagination.** List endpoints are cursor-paginated. The SDK exposes both a `list` (single page) and an `iterate`/stream helper that walks `nextCursor` until `hasMore` is false.

**Idempotency.** Invoice create and send-email accept an idempotency key, sent as the `Idempotency-Key` header so retries are safe.

**Webhooks.** Deliveries are signed with HMAC-SHA256 over the raw body. The SDK exposes a timing-safe `verifySignature(rawBody, header, secret)`.

**Surface.** Exactly 28 operations across companies, certificates, invoices, webhooks (incl. nested deliveries), credits, and billing — plus `health/live` and `health/ready`.

## Development

```sh
go build ./...   # compile
go vet ./...     # static analysis
go test ./...    # unit + HMAC vector; contract test skips without env
golangci-lint run
```

Tests live in `test/` and mock the API with `httptest` (no network). The
contract test in `test/contract/` runs only when `CORVO_API_KEY` and
`CORVO_BASE_URL` are set — that is how CI exercises a live instance.

## License

[MIT](LICENSE) © Corvo Technology 2026
