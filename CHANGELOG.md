# Changelog

## 0.3.0

Follows the public API contract 0.3.0 (OpenAPI at <https://corvo.tech/openapi.json>, vendored in `spec/openapi.json`). The SDK now covers all 92 operations.

### Breaking

- `Invoices.Create` returns `InvoiceCreateResult` (`Invoice` or `Schedule`): `Options.Type` `SCHEDULE` and the new `RECURRING` answer with a schedule (snake_case fields), not an invoice.
- `Invoices.SendEmail` takes `*InvoiceEmailParams` (optional destination) before the idempotency key.
- `CreditBalance`: removed `OverageUnits`, `OverageAmountCents` and `TenantID`; added `PlanRemaining` and `AvulsoBalance`; `Balance` is an `int` and includes loose notes. `CreditTransaction.Amount` is an `int`.
- `BillingUsage`: removed `OverageUnits` and `OverageAmountCents`; added `AvulsoBalance`. `BillingSubscription` gained `StripePortalAvailable`, `StripeSubscription` and `PixSubscription`.
- Plan slugs are now `free`, `self-hosted`, `solo`, `equipe`, `empresa` and `escala` (they were `developer`, `growth`, `scale`, `volume`). Slugs are open strings.
- `ThroughputRPS` of `CreditPlan` is a `float64`; `WebhookDelivery.ResponseBody` is a `string`.
- `InvoiceCustomerInput.TaxID` is optional (a foreign customer has none).

### Added

- Resources: `CNPJ`, `Customers`, `ServiceFavorites`, `AccountantPackage`, `InvoicePackage`, `DPSSequence`, `FiscalPortal`, `Schedules`, `Imports` (multipart upload, template and errors CSV), `FiscalProfiles`, `ServiceTaxProfiles`, `Integrations`, `IntegrationConnections`, `IntegrationOperations`.
- `Invoices.Substitute`, `Invoices.Review`, `Invoices.Overview`, `Credits.PreviewPurchase` (`GET /credits/purchase/preview`, replaces the retired POST).
- `InvoiceListParams.CustomerTaxID`; `purchaseOrder` and `notes`; foreign customers; `options.recurrence`; webhook event `nfse.invoice.taker_rejected` (12 events); `Webhook.Environment`.
- New fields on `Invoice` (`SubstitutesInvoiceID`, `SubstitutionDescription`, `PortalFlag`), `InvoiceEvent` (`Source`, `PortalEvent`, `PortalNsu`) and `Company` (`MunicipalityAgreement`, `Warnings`).
- `Nullable[T]` (`Some`, `Null`) to send an explicit JSON `null` in updates.
- `User-Agent: corvo-nfse-go/<version>` on every request.
- `TestContractCoverage`: loads `spec/openapi.json` and fails when an operation is missing or calls the wrong method or path.

### Changed

- Error `details` that is an object (not an array) no longer breaks error parsing.
- CI: Go 1.22, 1.23 and stable with `-race`, gofmt check and golangci-lint.

## 0.2.0

Batches, cancel, deduction, Simples rates; MEI and Simples Nacional only.
