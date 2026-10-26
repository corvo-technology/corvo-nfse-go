package test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

func TestInvoicesCreateBatchPostsItemsWithoutIdempotencyHeader(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusAccepted, `{"data": {
			"total": 2, "accepted": 1, "rejected": 1,
			"results": [
				{"index": 0, "status": "accepted", "replayed": true, "invoice": `+invoiceJSON+`},
				{"index": 1, "status": "rejected", "error": {"code": "ERROR_VALIDATION", "message": "Request validation failed",
					"details": [{"field": "amount", "code": "ERROR_OUT_OF_RANGE", "message": "x"}]}}
			]
		}}`)
	}))

	deduction := 20.0
	summary, err := client.Invoices.CreateBatch(context.Background(), []corvonfse.InvoiceBatchItem{{
		InvoiceCreateParams: corvonfse.InvoiceCreateParams{
			CompanyTaxID:   "11222333000181",
			Service:        corvonfse.InvoiceServiceInput{Description: "Consultoria"},
			Amount:         150.5,
			Deduction:      &deduction,
			CompetenceDate: "2026-10-01",
		},
		IdempotencyKey: "sheet-row-2",
	}})
	if err != nil {
		t.Fatalf("invoices.createBatch: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/invoices/batches" {
		t.Errorf("expected POST /v1/invoices/batches, got %s %s", call.method, call.path)
	}
	if got := call.header.Get("Idempotency-Key"); got != "" {
		t.Errorf("the batch route must not send Idempotency-Key, got %q", got)
	}
	for _, fragment := range []string{`"invoices":[`, `"idempotencyKey":"sheet-row-2"`, `"deduction":20`, `"companyTaxId":"11222333000181"`} {
		if !strings.Contains(call.body, fragment) {
			t.Errorf("expected body to contain %s, got %s", fragment, call.body)
		}
	}
	if summary.Total != 2 || summary.Accepted != 1 || summary.Rejected != 1 {
		t.Errorf("unexpected summary %+v", summary)
	}
	if !summary.Results[0].Replayed || summary.Results[0].Invoice == nil || summary.Results[0].Invoice.ID == "" {
		t.Errorf("expected a replayed accepted invoice, got %+v", summary.Results[0])
	}
	if summary.Results[1].Error == nil || summary.Results[1].Error.Code != "ERROR_VALIDATION" {
		t.Errorf("expected a rejected item with ERROR_VALIDATION, got %+v", summary.Results[1])
	}
}

func TestInvoicesCancelPostsReasonAndDescription(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		body := strings.Replace(invoiceJSON, `"status": "authorized"`, `"status": "cancelling"`, 1)
		writeJSON(w, http.StatusAccepted, `{"data": `+body+`}`)
	}))

	invoice, err := client.Invoices.Cancel(context.Background(), "inv_1", corvonfse.InvoiceCancelParams{
		Reason:      corvonfse.CancellationIssueError,
		Description: "Wrong amount on the invoice",
	})
	if err != nil {
		t.Fatalf("invoices.cancel: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/invoices/inv_1/cancel" {
		t.Errorf("expected POST /v1/invoices/inv_1/cancel, got %s %s", call.method, call.path)
	}
	for _, fragment := range []string{`"reason":"issue_error"`, `"description":"Wrong amount on the invoice"`} {
		if !strings.Contains(call.body, fragment) {
			t.Errorf("expected body to contain %s, got %s", fragment, call.body)
		}
	}
	if invoice.Status != corvonfse.InvoiceStatusCancelling {
		t.Errorf("expected status cancelling, got %q", invoice.Status)
	}
}

func TestInvoicesCancelReturnsTheTypedError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusConflict, `{"statusCode": 409, "code": "ERROR_INVOICE_NOT_CANCELLABLE", "message": "not cancellable"}`)
	}))
	_, err := client.Invoices.Cancel(context.Background(), "inv_1", corvonfse.InvoiceCancelParams{Reason: corvonfse.CancellationOther})
	if err == nil || err.Code != "ERROR_INVOICE_NOT_CANCELLABLE" || err.Status != http.StatusConflict {
		t.Fatalf("expected ERROR_INVOICE_NOT_CANCELLABLE (409), got %+v", err)
	}
}
