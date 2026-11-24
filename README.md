# corvo-nfse-go

Go SDK for the **[Corvo](https://corvo.tech) NFS-e API**: register companies, calculate taxes, issue Brazilian service invoices (NFS-e), follow them by webhook and download the PDF and XML.

- **Go 1.22+**, `net/http` only, **no dependencies**.
- Every call returns `(T, *CorvoError)`: API, network and timeout failures are values, never panics. The error carries the server's `Code`, the HTTP `Status`, the `RequestID` and the field-level `Details`.
- Cursor-paginated lists with an `Iterate` helper that walks every page.
- Idempotency keys for issuing an invoice and sending an email.
- HMAC-SHA256 webhook verification in constant time.

Full API reference: <https://corvo.tech/en/docs>. The machine-readable contract (OpenAPI 0.3.0) is at <https://corvo.tech/openapi.json>; a copy is vendored in `spec/openapi.json` and the tests check that the SDK implements every one of its 92 operations.

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
	TaxRegime:          corvonfse.TaxRegimeSimplesNacional, // MEI or Simples Nacional: the only regimes Corvo serves
	SimplesRate:        "6.54",                             // required for Simples Nacional: effective DAS rate (%), ignored for MEI
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
result, err := client.Invoices.Create(ctx, corvonfse.InvoiceCreateParams{
	CompanyTaxID: "11222333000181", // or CompanyID
	// Customer is optional: leave it nil for an invoice without one.
	Customer:       &corvonfse.InvoiceCustomerInput{TaxID: "52998224725", Name: "Cliente", Email: "cliente@example.com"},
	Service:        corvonfse.InvoiceServiceInput{Description: "Consultoria"}, // code and city default to the company's
	Amount:         150.50,
	CompetenceDate: "2026-10-16",
}, "order-42") // idempotency key: the same key never issues twice
invoice := result.Invoice // queued; follow it with Invoices.Get or webhooks
```

`Options.Type` `InvoiceTypeSchedule` or `InvoiceTypeRecurring` does not issue an invoice: the API creates a **schedule** and `result.Schedule` is set (`result.IsSchedule()`). The note exists only when the schedule runs (09:00 America/Sao_Paulo, Monday to Friday). `CompetenceDate` is the first run and must be from tomorrow on.

```go
result, err := client.Invoices.Create(ctx, corvonfse.InvoiceCreateParams{
	CompanyTaxID:   "11222333000181",
	Service:        corvonfse.InvoiceServiceInput{Description: "Mensalidade"},
	Amount:         300,
	CompetenceDate: "2026-12-01",
	Options: &corvonfse.InvoiceCreateOptions{
		Type:       corvonfse.InvoiceTypeRecurring,
		Recurrence: &corvonfse.InvoiceRecurrence{Frequency: corvonfse.FrequencyMonthly},
	},
}, "")
runs, _, err := client.Schedules.ListRuns(ctx, result.Schedule.ID, nil)
```

To declare a deduction from the ISS base, set `Deduction` (greater than 0 and less than `Amount` minus `Discount`). The invoice is still issued for the full `Amount`, and you are responsible for the declared value.

### Issue up to 50 invoices at once

`CreateBatch` takes 1 to 50 items (they can belong to different companies). Items are checked one by one: valid ones are saved together, invalid ones come back `rejected` and are not saved. Set `IdempotencyKey` on each item, so repeating the batch after a timeout returns the saved ones with `Replayed`.

```go
summary, err := client.Invoices.CreateBatch(ctx, []corvonfse.InvoiceBatchItem{
	{
		InvoiceCreateParams: corvonfse.InvoiceCreateParams{
			CompanyTaxID:   "11222333000181",
			Service:        corvonfse.InvoiceServiceInput{Description: "Consultoria"},
			Amount:         150.50,
			CompetenceDate: "2026-10-16",
		},
		IdempotencyKey: "sheet-row-2",
	},
})
for _, r := range summary.Results {
	if r.Status == "rejected" {
		fmt.Println(r.Index, r.Error.Code, r.Error.Message)
	}
}
```

The route allows 10 batches per minute per tenant (`429` with `Retry-After`).

### Cancel an invoice

```go
invoice, err := client.Invoices.Cancel(ctx, invoiceID, corvonfse.InvoiceCancelParams{
	Reason:      corvonfse.CancellationIssueError, // or CancellationServiceNotProvided, CancellationOther
	Description: "Wrong amount on the invoice",
})
```

- `queued`: cancelled at once, `Status` is `cancelled` with `CancelledAt` set.
- `authorized`: `Status` is `cancelling` while the cancellation is processed at the Sefin. The issuance credit is not refunded.
- `processing`: stays `processing` with `CancellationRequestedAt` set and is cancelled once authorized.

`Description` (15 to 255 characters) is required for an authorized or processing invoice and optional for a queued one. Read the invoice again (or listen to the `nfse.invoice.cancelled` and `nfse.invoice.cancellation_rejected` webhooks) to see the outcome: `cancelled`, or back to `authorized` with `LastErrorCode` when the cancellation was rejected.

### Look up a CNPJ, customers and favorite services

```go
profile, err := client.CNPJ.Lookup(ctx, "11222333000181") // legal name, address, MEI/Simples

customers, page, err := client.Customers.Search(ctx, &corvonfse.CustomerSearchParams{Search: "Ana"})
_, err = client.Customers.Update(ctx, "52998224725", corvonfse.CustomerUpdateParams{
	Email: corvonfse.Null[string](), // an explicit null clears the e-mail
})

fav, err := client.ServiceFavorites.Create(ctx, companyID, corvonfse.ServiceFavoriteCreateParams{
	Label: "Consultoria", ServiceCode: "010101", Description: "Consultoria em TI",
})
```

### Send the monthly package to the accountant

```go
_, err := client.AccountantPackage.Update(ctx, companyID, corvonfse.AccountantPackageUpdateParams{
	AccountantEmail: corvonfse.Some("contador@example.com"),
})
sent, err := client.AccountantPackage.Send(ctx, companyID, "2026-10") // sent.Status
zip, err := client.InvoicePackage.Download(ctx, companyID, "2026-10", nil) // zip.Buffer
```

### Import invoices from a spreadsheet

```go
started, err := client.Imports.Start(ctx, csvBytes, "notas.csv", companyID)
// poll until job.Status == corvonfse.ImportReady, fix rows, then emit
job, err := client.Imports.Get(ctx, started.ID)
job, err = client.Imports.Emit(ctx, started.ID, &corvonfse.ImportEmitParams{})
```

### Substitute an invoice, resolve manual review

```go
newInvoice, err := client.Invoices.Substitute(ctx, invoiceID, corvonfse.InvoiceSubstituteParams{
	Description: "Valor corrigido", // why
}, "sub-1")
_, err = client.Invoices.Review(ctx, corvonfse.InvoiceReviewParams{
	Action: corvonfse.ReviewRetry, InvoiceIDs: []string{id},
})
```

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
| `client.DPSSequence` | `Get`, `Set` (continue the DPS numbering) |
| `client.FiscalProfiles` | `List`, `Upsert`, `Approve` |
| `client.ServiceTaxProfiles` | `List`, `Create`, `Update`, `Remove`, `Approve` |
| `client.AccountantPackage` | `Get`, `Update`, `Send`, `ListSends` |
| `client.InvoicePackage` | `Summary`, `Download` (month ZIP) |
| `client.CNPJ` | `Lookup` |
| `client.Customers` | `Search`, `Iterate`, `Get`, `Update`, `Remove` |
| `client.ServiceFavorites` | `List`, `Create`, `Remove` |
| `client.Schedules` | `List`, `Iterate`, `Get`, `Update`, `Pause`, `Resume`, `Cancel`, `ListRuns` |
| `client.Imports` | `Start`, `List`, `Get`, `Emit`, `ListRows`, `IterateRows`, `UpdateRow`, `Discard`, `DownloadTemplate`, `DownloadErrors` |
| `client.Integrations` | `ListProviders` |
| `client.IntegrationConnections` | `List`, `Create`, `Get`, `Validate`, `UpdateCredentials`, `Activate`, `Pause`, `Disconnect`, `ListRules`, `ReplaceRules`, `ListEvents`, `ListOperations` |
| `client.IntegrationOperations` | `CompleteCustomer`, `Retry` |
| `client.FiscalPortal` | `Status` |
| `client.Certificates` | `Upload`, `List`, `Remove` |
| `client.Invoices` | `Create`, `CreateBatch`, `Calculate`, `Get`, `List`, `Iterate`, `Events`, `PDF`, `XML`, `DownloadURL`, `SendEmail`, `Substitute`, `Review`, `Overview`, `Cancel` |
| `client.Webhooks` | `List`, `Create`, `Remove`, `Test` + `Deliveries.List/Iterate/Stats/Retry` |
| `client.Credits` | `Balance`, `Plans`, `History`, `IterateHistory`, `PreviewPurchase` |
| `client.Billing` | `Plan`, `Usage` |
| `client.Plans` | `List`, `Get`, `Quote` (public price catalog) |
| `client.Health` | `Live`, `Ready` |

Changing plans and the billing portal exist only in the dashboard, so they are not in the SDK. Every method takes a `context.Context` first.

## Plans and credits

Plan slugs are `free`, `solo`, `equipe`, `empresa` and `escala` (plus `self-hosted` on a self-hosted instance). Treat the slug as an open string and read the allowances from the plan. `Credits.Balance` returns `Balance` (plan notes left plus loose notes), `PlanRemaining` and `AvulsoBalance`; there is no overage billing. `Billing.Plan` answers `NO_PLAN_ASSIGNED` (404) for a workspace without a plan.

A `corvo_test_` key only sees the test environment: it cannot reach live webhooks or integration connections, so use a `corvo_live_` key to manage those. Webhook URLs that resolve to private networks are rejected.

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
go test ./...    # unit tests with httptest + contract coverage of spec/openapi.json
golangci-lint run
```

`TestContractCoverage` loads `spec/openapi.json` and fails when an operation of the contract is not implemented; see `spec/README.md` to update it. The live test in `test/contract/` runs only when `CORVO_API_KEY` and `CORVO_BASE_URL` point to a reachable instance.

## License

[MIT](LICENSE) © Corvo Technology
