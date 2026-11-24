# Public API changelog

What changed in the public API since **2026-10-26** (version `0.2.0` of the SDKs), so each SDK repository knows what to update. The contract itself is [openapi.json](./openapi.json); the route list is in [README.md](./README.md).

Built from `git log --since=2026-10-26` and the diff of the controllers and DTOs against commit `0543600` (the last one of 2026-10-26).

## 0.3.0 (2026-11-24)

### Breaking changes

Update these first. Each one changes something an SDK built for 0.2.0 sends or reads.

1. **Plan slugs were renamed.** `developer`, `growth`, `scale` and `volume` no longer exist. The slugs are now `free`, `solo`, `equipe`, `empresa` and `escala` (plus `self-hosted` on a self-hosted instance). Where it shows: `plan.slug` in `GET /billing/plan` and `slug` of each item in `GET /credits/plans`. Do not hard-code the list: treat the slug as an open string and read the allowances from the plan (`monthlyCreditAllowance`: `solo` 100, `equipe` 250, `empresa` 600, `escala` 1200; extra notes cost R$ 0,99).
2. **`POST /invoices` with `options.type` `SCHEDULE` no longer answers with an invoice.** `SCHEDULE` and the new `RECURRING` create a *schedule* and answer `202` with it (see `Schedule`, `snake_case` fields). The note only exists when the schedule runs (09:00 America/Sao_Paulo, Monday to Friday; a weekend date runs on Monday). `competenceDate` is now the date of the first run and must be from tomorrow on. Handle both bodies in `invoices_create`: `data` is an `Invoice` for `IMMEDIATE` and a `Schedule` otherwise. Follow the runs with `schedules_listRuns`.
3. **`GET /credits` changed its fields.** `overageUnits` and `overageAmountCents` are gone (there is no overage billing). `balance` is now the plan notes left plus the loose notes bought separately (plan notes of a lapsed plan count 0). New: `planRemaining`, `avulsoBalance`. `hardLimit` stays.
4. **`GET /billing/plan` changed its fields.** `usage.overage_units` and `usage.overage_amount_cents` are gone; `usage.avulso_balance` is new. `subscription` gained `stripe_portal_available`, `stripe_subscription` and `pix_subscription` (an object or `null`). `plan.monthly_price_cents` follows the new price table. The 404 of a workspace without a plan now has `code: "NO_PLAN_ASSIGNED"`.
5. **`POST /credits/purchase/preview` is now `GET /credits/purchase/preview?units=N`.** The old route only answered `410 Gone`. The new one prices a one-off purchase: `{ units, unitPriceCents, totalCents }` (R$ 0,99 each, at least 3 notes, whole numbers). Buying is done in the dashboard.
6. **A test key can no longer reach live webhooks.** Webhooks created with an API key are tied to that key's environment (`environment` is `test` or `live`) and only receive its events; `GET /webhooks`, `GET /webhooks/deliveries` and the delivery stats show only the key's environment. A client that managed live webhooks with a `corvo_test_` key must use a `corvo_live_` key. (Integration connections and operations follow the same rule from their first release.)
7. **Webhook destinations are checked more strictly.** URLs that resolve to private, loopback, link-local, CGNAT, benchmarking, multicast or NAT64/6to4 addresses are refused when the webhook is registered and when it is delivered.

### Corrections of the documented shapes

The old Swagger (served only in development) did not match the real answers. If an SDK model was copied from it, check these against `openapi.json`:

- **Invoice.** The API returns the stored fields plus the derived `documentsStatus` (`pending` or `ready`), `emailStatus` and `stage`. It has `documentsReadyAt`, `lastEmailStatus` and `lastEmailSentAt`. The fields `documentsReady` and `lastEmail` shown by the old Swagger do not exist in the invoice (`documentsReady` exists only inside webhook payloads). Dates are ISO 8601; unknown values are `null`, not absent.
- **Lists** answer `{ data, pagination: { limit, hasMore, nextCursor } }` (camelCase). The cursor is opaque. In `customers_search` and the schedule lists it is the offset of the next page, still to be passed back untouched.
- **Errors** are `{ statusCode, code, message, details?, requestId }`; a few feature codes have no `ERROR_` prefix (`CNPJ_INVALID`, `CNPJ_LOOKUP_UNAVAILABLE`, `CHECKOUT_INVALID`, `PACKAGE_TOO_LARGE`, `NO_PLAN_ASSIGNED`, `DPS_NUMBER_BELOW_CURRENT`, ...). `code` is not an enum: new codes appear.
- **Casing.** `GET /billing/plan`, `/schedules` and `/imports` answer in `snake_case`; everything else in `camelCase`.
- **Invoice statuses** (10, unchanged): `queued`, `processing`, `authorized`, `rejected`, `retry_pending`, `reconciliation_required`, `manual_review`, `failed`, `cancelling`, `cancelled`. New transitions: `authorized` can go straight to `cancelled` (the Sefin cancelled it because a substitute was authorized, or the cancellation was made on the national portal); `manual_review` can be resolved to `reconciliation_required`, `authorized` or `rejected` (see `invoices_review`).

### New operations

| Area | Operations |
| --- | --- |
| Invoices | `invoices_substitute` (`POST /invoices/{id}/substitute`), `invoices_review` (resolve `manual_review`), `invoicePackage_download` / `invoicePackage_summary` (monthly ZIP for the accountant), `fiscalPortal_status` |
| Companies | `dpsSequence_get` / `dpsSequence_set` (continue the DPS numbering from another system; error `DPS_NUMBER_BELOW_CURRENT`), `accountantPackage_get` / `_update` / `_send` / `_listSends` |
| Customers | `cnpj_lookup` (`GET /cnpj/{cnpj}`), `customers_search`, `customers_get`, `customers_update`, `customers_delete` |
| Service favorites | `serviceFavorites_list` / `_create` / `_delete` |
| Schedules | `schedules_list` / `_get` / `_update` / `_pause` / `_resume` / `_cancel` / `_listRuns` |
| Imports | `imports_start` / `_list` / `_get` / `_emit` / `_listRows` / `_updateRow` / `_discard` / `_downloadTemplate` / `_downloadErrors` |
| Integrations | `integrations_listProviders`, `integrationConnections_*` (create, list, get, validate, updateCredentials, activate, pause, disconnect, listRules, replaceRules, listEvents, listOperations), `integrationOperations_completeCustomer` / `_retry` |
| Credits | `credits_previewPurchase` (see breaking change 5) |
| Health | `health_live` (existed, now documented) |

The fiscal profile and the service tax profile routes (`fiscalProfiles_*`, `serviceTaxProfiles_*`) already existed in 0.2.0; they are documented now.

### New and changed fields

Requests (`POST /invoices`, `POST /invoices/calculate`, batch items, schedules):

- `options.type` accepts `RECURRING`, with `options.recurrence: { frequency: WEEKLY | BIWEEKLY | MONTHLY, endsOn? }` (required for `RECURRING`, refused otherwise).
- `purchaseOrder` (up to 60 characters, DPS `xPed`) and `notes` (up to 2000, DPS `xInfComp`) on the invoice. They come back in `serviceSnapshot.purchaseOrder` and `serviceSnapshot.notes`.
- `customer.foreign` for a customer outside Brazil (`taxId` or `noTaxIdReason`, `country`, `postalCode`, `city`, `region`, `street`, `number`, `district`, `complement?`). It cannot be combined with `customer.taxId` or `customer.address`; `customer.name` is then required.
- `customer.name` may be omitted when the customer was already saved: it is filled from the CPF/CNPJ.
- `GET /invoices` accepts `customerTaxId` (CPF or CNPJ, digits).
- `POST /invoices/{id}/email` accepts `{ email }` to send to another address than the customer's.
- `POST /invoices/{id}/substitute` replaces an authorized invoice: only `description` (why), `service.description`, `purchaseOrder` and `notes` can change.
- Company: `municipalityAgreement` (what the Receita says about the municipality's adhesion to the national system).

Responses:

- Invoice: `substitutesInvoiceId`, `substitutionDescription` (on a substitute), `portalFlag` (`taker_rejected` or `blocked`, from the national portal).
- Invoice events: `source` (`portal` or `null`), `portalEvent`, `portalNsu`.
- Webhook event type `nfse.invoice.taker_rejected` (the customer rejected the invoice on the portal). The full list is `WebhookEventType` in the document: 12 `nfse.invoice.*` types.
- Error codes added: `ERROR_INVOICE_NOT_SUBSTITUTABLE`, `ERROR_SCHEDULE_NOT_FOUND`, `ERROR_SCHEDULE_LIMIT`, `ERROR_SCHEDULE_STATE`, `ERROR_IMPORT_*`. `lastErrorCode` of an invoice can be a Sefin rule (`E0014`, ...) or a Corvo code such as `DPS_NUMBER_ALREADY_USED`.

### Behavior changes that are not breaking

- A competence date in the future is refused for an immediate emission (`422` with `ERROR_FUTURE_DATE` in `details[].code`; Sefin rule `E0015`); use `SCHEDULE` to issue later. `competenceDate` may be at most 15 days ahead (31 for `RECURRING`).
- Test-environment DPS numbering starts at a high number so it does not collide with notes issued before (`E0014`). Use `dpsSequence_set` to continue from the last number issued elsewhere.
- A MEI or Simples Nacional issuer can only change the description, the purchase order and the notes in a substitute (`E0061`).
- Removing a member from the workspace revokes the API keys that member created.

## 0.2.0 (2026-10-26)

First contract the SDKs were written for. It was never published as a document: it had to be read from the controllers.
