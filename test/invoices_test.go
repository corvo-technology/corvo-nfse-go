package test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

const invoiceJSON = `{
	"id": "f0add80b-5026-4a11-8f00-5fa03be98b5b",
	"companyId": "09947127-2702-4f6d-9da8-cb463608d812",
	"tenantId": "6247dfd9-0adc-4bf6-aac9-ee9dc1e7ca8b",
	"environment": "test",
	"documentKeyVersion": 2,
	"emailSnapshot": {"send": false},
	"externalId": "order-42",
	"idempotencyKey": null,
	"status": "authorized",
	"customerSnapshot": {
		"name": "Cliente Fictício de Testes",
		"email": "cliente@example.com",
		"taxId": "52998224725",
		"address": {"city": "Uberlândia", "state": "MG"}
	},
	"serviceSnapshot": {"code": "010101", "description": "Software development", "municipalityCode": "3170206"},
	"taxSnapshot": {},
	"amount": 150.5,
	"currency": "BRL",
	"competenceDate": "2026-10-12T00:00:00.000Z",
	"dpsId": "DPS317020624183468700014800001000000000000001",
	"dpsSeries": "1",
	"dpsNumber": 1,
	"accessKey": "02910000000000000001",
	"nfseNumber": 1001,
	"fiscalProvider": "provider",
	"fiscalSchemaVersion": "v2_02",
	"issuedAt": "2026-10-12T13:04:40.000Z",
	"authorizedAt": "2026-10-12T13:04:41.000Z",
	"lastErrorCode": null,
	"lastErrorMessage": null,
	"lastErrorProvider": null,
	"documentsReadyAt": "2026-10-12T13:04:42.000Z",
	"lastEmailStatus": null,
	"lastEmailSentAt": null,
	"createdAt": "2026-10-12T13:04:38.985Z",
	"updatedAt": "2026-10-12T13:04:42.000Z",
	"documentsStatus": "ready",
	"emailStatus": null,
	"stage": "documents_ready"
}`

func TestInvoicesListPassesFiltersAndUnwrapsPropsItems(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{
			"data": [{"props": `+invoiceJSON+`}],
			"pagination": {"limit": 1, "hasMore": false, "nextCursor": null}
		}`)
	}))

	items, pagination, err := client.Invoices.List(context.Background(), &corvonfse.InvoiceListParams{
		CompanyID:  "cmp_1",
		Status:     corvonfse.InvoiceStatusAuthorized,
		From:       "2026-10-01",
		To:         "2026-10-31",
		ExternalID: "order-42",
		Limit:      1,
	})
	if err != nil {
		t.Fatalf("invoices.list: %v", err)
	}
	if call.path != "/v1/invoices" {
		t.Errorf("expected path /v1/invoices, got %s", call.path)
	}
	query := call.query
	for _, fragment := range []string{
		"companyId=cmp_1",
		"status=authorized",
		"from=2026-10-01",
		"to=2026-10-31",
		"externalId=order-42",
		"limit=1",
	} {
		if !strings.Contains(query, fragment) {
			t.Errorf("expected query to contain %s, got %q", fragment, query)
		}
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(items))
	}
	if items[0].NFSENumber != 1001 || items[0].Status != corvonfse.InvoiceStatusAuthorized {
		t.Errorf("expected props item decoded, got %+v", items[0])
	}
	if items[0].Stage != corvonfse.InvoiceStageDocumentsReady {
		t.Errorf("expected stage documents_ready, got %q", items[0].Stage)
	}
	if pagination.Limit != 1 || pagination.HasMore {
		t.Errorf("expected single non-exhausted page, got %+v", pagination)
	}
}

func TestInvoicesEventsUseCursorParams(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{
			"data": [{"props": {"id": "evt_1", "invoiceId": "inv_1", "toStatus": "authorized", "reason": "ok"}}],
			"pagination": {"limit": 10, "hasMore": false, "nextCursor": null}
		}`)
	}))

	events, _, err := client.Invoices.Events(context.Background(), "inv_1", &corvonfse.ListParams{
		Limit:  10,
		Cursor: "prev-cursor",
	})
	if err != nil {
		t.Fatalf("invoices.events: %v", err)
	}
	if call.path != "/v1/invoices/inv_1/events" {
		t.Errorf("expected events path, got %s", call.path)
	}
	if call.query != "cursor=prev-cursor&limit=10" {
		t.Errorf("expected cursor and limit query, got %q", call.query)
	}
	if len(events) != 1 || events[0].ToStatus != corvonfse.InvoiceStatusAuthorized {
		t.Errorf("expected one event, got %+v", events)
	}
}

func TestInvoicesIterateKeepsFiltersAcrossPages(t *testing.T) {
	var calls []capture
	page := 0
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, record(r))
		page++
		if page == 1 {
			writeJSON(w, http.StatusOK, `{
				"data": [{"props": {"id": "inv_1"}}],
				"pagination": {"limit": 1, "hasMore": true, "nextCursor": "c1"}
			}`)
			return
		}
		writeJSON(w, http.StatusOK, `{
			"data": [{"props": {"id": "inv_2"}}],
			"pagination": {"limit": 1, "hasMore": false, "nextCursor": null}
		}`)
	}))

	var seen []string
	err := client.Invoices.Iterate(context.Background(), &corvonfse.InvoiceListParams{
		Status: corvonfse.InvoiceStatusAuthorized,
		Limit:  1,
	}, func(invoice corvonfse.Invoice) bool {
		seen = append(seen, invoice.ID)
		return true
	})
	if err != nil {
		t.Fatalf("invoices.iterate: %v", err)
	}
	if strings.Join(seen, ",") != "inv_1,inv_2" {
		t.Errorf("expected both invoices, got %v", seen)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(calls))
	}
	if !strings.Contains(calls[1].query, "status=authorized") || !strings.Contains(calls[1].query, "cursor=c1") {
		t.Errorf("expected filters to survive paging, got %q", calls[1].query)
	}
}

func TestInvoicesGetUsesPropsEnvelope(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data": {"props": `+invoiceJSON+`}}`)
	}))

	invoice, err := client.Invoices.Get(context.Background(), "f0add80b-5026-4a11-8f00-5fa03be98b5b")
	if err != nil {
		t.Fatalf("invoices.get: %v", err)
	}
	expectedPath := "/v1/invoices/f0add80b-5026-4a11-8f00-5fa03be98b5b"
	if call.path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, call.path)
	}
	if invoice.Amount != 150.5 || invoice.DocumentKeyVersion != 2 {
		t.Errorf("expected invoice fields decoded, got %+v", invoice)
	}
	if invoice.CustomerSnapshot.Name != "Cliente Fictício de Testes" {
		t.Errorf("expected customer snapshot decoded, got %+v", invoice.CustomerSnapshot)
	}
}

func TestInvoicesCreatePropagatesValidationDetails(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusBadRequest, `{
			"statusCode": 400,
			"code": "ERROR_VALIDATION",
			"message": "Request validation failed",
			"details": [{"field": "service.code", "code": "ERROR_INVALID_SERVICE_CODE", "message": "Invalid code"}],
			"requestId": "req-1"
		}`)
	}))

	_, err := client.Invoices.Create(context.Background(), corvonfse.InvoiceCreateParams{
		Customer: corvonfse.InvoiceCustomerInput{Name: "Cliente", TaxID: "52998224725"},
		Service:  corvonfse.InvoiceServiceInput{Code: "000000"},
		Amount:   10,
	}, "")
	failure := requireError(t, err)
	if failure.Code != "ERROR_VALIDATION" {
		t.Errorf("expected ERROR_VALIDATION, got %q", failure.Code)
	}
	if len(failure.Details) != 1 || failure.Details[0].Field != "service.code" {
		t.Errorf("expected service.code detail, got %+v", failure.Details)
	}
}

func TestInvoicesDownloadURL(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data": {
			"downloadUrl": "https://files.corvo.tech/nfse-1001.pdf",
			"expiresAt": "2026-10-13T13:04:42.000Z",
			"fileName": "nfse-1001.pdf",
			"mimeType": "application/pdf"
		}}`)
	}))

	signed, err := client.Invoices.DownloadURL(context.Background(), "inv_1", corvonfse.FormatPDF)
	if err != nil {
		t.Fatalf("invoices.downloadUrl: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/invoices/inv_1/download-url" {
		t.Errorf("expected POST /v1/invoices/inv_1/download-url, got %s %s", call.method, call.path)
	}
	if call.body != `{"format":"pdf"}` {
		t.Errorf("expected format body, got %s", call.body)
	}
	if signed.DownloadURL == "" || signed.FileName != "nfse-1001.pdf" {
		t.Errorf("expected unwrapped download url, got %+v", signed)
	}
}

func TestInvoicesSendEmailWithIdempotencyKey(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusAccepted, `{"data": {"emailId": "mail_1", "status": "queued"}}`)
	}))

	result, err := client.Invoices.SendEmail(context.Background(), "inv_1", "mail-1")
	if err != nil {
		t.Fatalf("invoices.sendEmail: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/invoices/inv_1/email" {
		t.Errorf("expected POST /v1/invoices/inv_1/email, got %s %s", call.method, call.path)
	}
	if call.header.Get("Idempotency-Key") != "mail-1" {
		t.Errorf("expected Idempotency-Key mail-1, got %q", call.header.Get("Idempotency-Key"))
	}
	if result.EmailID != "mail_1" || result.Status != "queued" {
		t.Errorf("expected email result, got %+v", result)
	}
}
