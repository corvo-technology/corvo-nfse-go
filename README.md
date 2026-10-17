# corvo-nfse-go

Go SDK for the **[Corvo](https://corvo.tech) NFS-e API**: register companies, calculate taxes, issue Brazilian service invoices (NFS-e), follow them by webhook and download the PDF and XML.

- **Go 1.22+**, `net/http` only, **no dependencies**.
- Every call returns `(T, *CorvoError)`: API, network and timeout failures are values, never panics. The error carries the server's `Code`, the HTTP `Status`, the `RequestID` and the field-level `Details`.
- Cursor-paginated lists with an `Iterate` helper that walks every page.
- Idempotency keys for issuing an invoice and sending an email.
- HMAC-SHA256 webhook verification in constant time.

Full API reference: <https://corvo.tech/en/docs>.

## Install

```sh
go get github.com/corvo-technology/corvo-nfse-go
```

The package name is `corvonfse`.

## Quickstart

Start with a `corvo_test_...` key: invoices issued with it have no fiscal value.

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
	client := corvonfse.NewClient(corvonfse.Options{APIKey: "corvo_test_..."})
	ctx := context.Background()

	companies, page, err := client.Companies.List(ctx, &corvonfse.ListParams{Limit: 10})
	if err != nil {
		log.Fatalf("%s (HTTP %d, request %s)", err.Code, err.Status, err.RequestID)
	}
	fmt.Println(len(companies), page.HasMore)
}
```

`NewClient` is the only call that panics: it fails fast on a missing API key, an invalid base URL, or a plain `http://` URL on a public host.

### Register a company

```go
company, err := client.Companies.Create(ctx, corvonfse.CompanyCreateParams{
	LegalName:          "Empresa Exemplo Ltda",
	TaxID:              "11222333000181",
	TaxRegime:          corvonfse.TaxRegimeSimplesNacional,
	MunicipalityCode:   "3550308", // IBGE code, 7 digits
	DefaultServiceCode: "010101",  // optional: used when an invoice omits service.code
})
```

In `live`, also upload the A1 certificate. The company is ready when `company.Status` is `"ready"` and `company.SetupGaps` is empty.

### Estimate the taxes before issuing

Nothing is issued, stored or charged.

```go
estimate, err := client.Invoices.Calculate(ctx, corvonfse.InvoiceCreateParams{
	CompanyTaxID:   "11222333000181",
	Service:        corvonfse.InvoiceServiceInput{Description: "Consultoria"},
	Amount:         1500,
	CompetenceDate: "2026-10-16",
})
// estimate.Status is "ready" or "blocked": a fiscal gap is data, not an error.
```

### Issue an invoice

Issuing is asynchronous: a successful call means the request was accepted, not that the invoice is authorized. Follow the status with `Invoices.Get` or webhooks.

```go
invoice, err := client.Invoices.Create(ctx, corvonfse.InvoiceCreateParams{
	CompanyTaxID: "11222333000181", // or CompanyID
	// Customer is optional: leave it nil for an invoice without one.
	Customer:       &corvonfse.InvoiceCustomerInput{TaxID: "52998224725", Name: "Cliente", Email: "cliente@example.com"},
	Service:        corvonfse.InvoiceServiceInput{Description: "Consultoria"}, // code and city default to the company's
	Amount:         150.50,
	CompetenceDate: "2026-10-16",
}, "order-42") // idempotency key: the same key never issues twice
```

Set `Options: &corvonfse.InvoiceCreateOptions{Type: corvonfse.InvoiceTypeSchedule}` with a future `CompetenceDate` to schedule the issuance.

### Lists and iteration

```go
items, page, err := client.Companies.List(ctx, &corvonfse.ListParams{Limit: 100})

err = client.Invoices.Iterate(ctx, &corvonfse.InvoiceListParams{
	Status: corvonfse.InvoiceStatusAuthorized,
}, func(inv corvonfse.Invoice) bool {
	fmt.Println(inv.ID)
	return true // return false to stop early
})
```

### Download the PDF or XML

The invoice must be `authorized` and have its documents ready.

```go
pdf, err := client.Invoices.PDF(ctx, invoice.ID)
if err == nil {
	_ = os.WriteFile("nfse.pdf", pdf.Buffer, 0o644) // also pdf.MimeType and pdf.FileName
}
```

### Verify webhook signatures

```go
body, _ := io.ReadAll(r.Body) // the raw bytes: do not parse and re-serialize first
if !corvonfse.VerifySignature(body, r.Header.Get("X-Corvo-Signature"), secret) {
	http.Error(w, "invalid signature", http.StatusBadRequest)
	return
}
// Deduplicate by event id: the same event can be delivered more than once.
```

`VerifySignature` never panics: a malformed header, an empty secret or a mismatch return `false`.

## Resources

| Field | Operations |
| --- | --- |
| `client.Companies` | `Create`, `Get`, `Update`, `Remove`, `List`, `Iterate` |
| `client.Certificates` | `Upload`, `List`, `Remove` |
| `client.Invoices` | `Create`, `Calculate`, `Get`, `List`, `Iterate`, `Events`, `PDF`, `XML`, `DownloadURL`, `SendEmail` |
| `client.Webhooks` | `List`, `Create`, `Remove`, `Test` + `Deliveries.List/Iterate/Stats/Retry` |
| `client.Credits` | `Balance`, `Plans`, `History`, `IterateHistory` |
| `client.Billing` | `Plan`, `Usage` |
| `client.Plans` | `List`, `Get`, `Quote` (public price catalog) |
| `client.Health` | `Live`, `Ready` |

Changing plans and the billing portal exist only in the dashboard, so they are not in the SDK. Every method takes a `context.Context` first.

## Configuration

| Option | Env fallback | Default |
| --- | --- | --- |
| `APIKey` | `CORVO_API_KEY` | required |
| `BaseURL` | `CORVO_BASE_URL` | `https://api.corvo.tech` |
| `Prefix` | none | `v1` |
| `Timeout` | none | 30s per request |
| `AllowInsecureHTTP` | none | `false` (loopback and private hosts are always allowed) |
| `HTTPClient` | none | pooled `*http.Client` |

## Behavior

- **Auth.** Every request sends `X-API-Key`. Keys are `corvo_test_...` or `corvo_live_...` and belong to one account and environment; the key decides the environment.
- **Responses.** Single resources and list items come back as the API's `data`; lists add `pagination`. Optional fields that are absent map to their zero value.
- **Errors.** Network failures, timeouts, non-2xx responses and malformed JSON all return a `*CorvoError`. `Code` is the API's `ERROR_*` code, falling back to `http_<status>`; transport failures use `network_error`, `timeout` and `invalid_json`. On `429`, `RetryAfter` holds the seconds to wait. `Iterate` reports a failed page the same way.

## Development

```sh
go build ./...
go vet ./...
go test ./...    # unit tests with httptest; the contract test skips without env
golangci-lint run
```

The contract test in `test/contract/` runs only when `CORVO_API_KEY` and `CORVO_BASE_URL` point to a reachable instance.

## License

[MIT](LICENSE) © Corvo Technology
