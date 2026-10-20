package test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

// invoiceWithStatus returns invoiceJSON with its status and the extra JSON
// fields (a leading-comma-free fragment such as `"cancelledAt": "..."`)
// replaced/added.
func invoiceWithStatus(status string, extra string) string {
	body := strings.Replace(invoiceJSON, `"status": "authorized"`, `"status": "`+status+`"`, 1)
	if extra != "" {
		body = strings.Replace(body, `"stage": "documents_ready"`, `"stage": "documents_ready", `+extra, 1)
	}
	return body
}

func TestInvoicesCancelSendsReasonAndReturnsCancelledInvoice(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data": `+invoiceWithStatus("cancelled", `
			"cancellationReason": "issue_error",
			"cancellationDescription": "Valor do servico informado errado",
			"cancelledAt": "2026-10-13T10:00:00.000Z",
			"cancellationRequestedAt": "2026-10-13T09:59:59.000Z"`)+`}`)
	}))

	invoice, err := client.Invoices.Cancel(context.Background(), "inv_1", corvonfse.InvoiceCancelParams{
		Reason:      corvonfse.CancellationReasonIssueError,
		Description: "Valor do servico informado errado",
	})
	if err != nil {
		t.Fatalf("invoices.cancel: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/invoices/inv_1/cancel" {
		t.Errorf("expected POST /v1/invoices/inv_1/cancel, got %s %s", call.method, call.path)
	}
	if call.header.Get("X-API-Key") != apiKey {
		t.Errorf("expected X-API-Key header, got %q", call.header.Get("X-API-Key"))
	}
	if call.header.Get("Content-Type") != "application/json" {
		t.Errorf("expected JSON content type, got %q", call.header.Get("Content-Type"))
	}
	expectedBody := `{"reason":"issue_error","description":"Valor do servico informado errado"}`
	if call.body != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, call.body)
	}
	if invoice.Status != corvonfse.InvoiceStatusCancelled {
		t.Errorf("expected status cancelled, got %q", invoice.Status)
	}
	if invoice.CancellationReason != corvonfse.CancellationReasonIssueError {
		t.Errorf("expected cancellation reason issue_error, got %q", invoice.CancellationReason)
	}
	if invoice.CancelledAt != "2026-10-13T10:00:00.000Z" {
		t.Errorf("expected cancelledAt set, got %q", invoice.CancelledAt)
	}
}

func TestInvoicesCancelOmitsEmptyDescription(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data": `+invoiceWithStatus("cancelled", `"cancellationReason": "other"`)+`}`)
	}))

	_, err := client.Invoices.Cancel(context.Background(), "inv_1", corvonfse.InvoiceCancelParams{
		Reason: corvonfse.CancellationReasonOther,
	})
	if err != nil {
		t.Fatalf("invoices.cancel: %v", err)
	}
	if call.body != `{"reason":"other"}` {
		t.Errorf("expected description omitted, got %s", call.body)
	}
}

func TestInvoicesCancelAcceptedReturnsCancellingInvoice(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusAccepted, `{"data": `+invoiceWithStatus("cancelling", `
			"cancellationReason": "service_not_provided",
			"cancellationDescription": "Servico nao foi prestado ao cliente",
			"cancelledAt": null,
			"cancellationRequestedAt": "2026-10-13T09:59:59.000Z"`)+`}`)
	}))

	invoice, err := client.Invoices.Cancel(context.Background(), "inv 1", corvonfse.InvoiceCancelParams{
		Reason:      corvonfse.CancellationReasonServiceNotProvided,
		Description: "Servico nao foi prestado ao cliente",
	})
	if err != nil {
		t.Fatalf("invoices.cancel: %v", err)
	}
	if call.path != "/v1/invoices/inv 1/cancel" {
		t.Errorf("expected escaped id in path, got %s", call.path)
	}
	if invoice.Status != corvonfse.InvoiceStatusCancelling {
		t.Errorf("expected status cancelling, got %q", invoice.Status)
	}
	if invoice.CancelledAt != "" {
		t.Errorf("expected null cancelledAt to map to empty, got %q", invoice.CancelledAt)
	}
	if invoice.CancellationRequestedAt != "2026-10-13T09:59:59.000Z" {
		t.Errorf("expected cancellationRequestedAt set, got %q", invoice.CancellationRequestedAt)
	}
}

func TestInvoicesCancelNotCancellableSurfacesCorvoError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusConflict, `{
			"statusCode": 409,
			"code": "ERROR_INVOICE_NOT_CANCELLABLE",
			"message": "Invoice cannot be cancelled in status retry_pending",
			"requestId": "req-409"
		}`)
	}))

	_, err := client.Invoices.Cancel(context.Background(), "inv_1", corvonfse.InvoiceCancelParams{
		Reason:      corvonfse.CancellationReasonOther,
		Description: "Cancelamento solicitado pelo cliente",
	})
	failure := requireError(t, err)
	if failure.Status != http.StatusConflict || failure.Code != "ERROR_INVOICE_NOT_CANCELLABLE" {
		t.Errorf("expected 409 ERROR_INVOICE_NOT_CANCELLABLE, got %+v", failure)
	}
	if failure.RequestID != "req-409" {
		t.Errorf("expected requestId req-409, got %q", failure.RequestID)
	}
}

func TestInvoicesCancelInvalidDescriptionSurfacesCorvoError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusUnprocessableEntity, `{
			"statusCode": 422,
			"code": "ERROR_VALIDATION",
			"message": "description must have 15 to 255 characters",
			"details": [{"field": "description", "code": "ERROR_INVALID_LENGTH", "message": "too short"}],
			"requestId": "req-422"
		}`)
	}))

	_, err := client.Invoices.Cancel(context.Background(), "inv_1", corvonfse.InvoiceCancelParams{
		Reason:      corvonfse.CancellationReasonOther,
		Description: "short",
	})
	failure := requireError(t, err)
	if failure.Status != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", failure.Status)
	}
	if len(failure.Details) != 1 || failure.Details[0].Field != "description" {
		t.Errorf("expected description detail, got %+v", failure.Details)
	}
}

func TestInvoicesGetParsesCancelledInvoice(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"data": `+invoiceWithStatus("cancelled", `
			"cancellationReason": "issue_error",
			"cancellationDescription": "Valor do servico informado errado",
			"cancelledAt": "2026-10-13T10:00:00.000Z",
			"cancellationRequestedAt": "2026-10-13T09:59:59.000Z"`)+`}`)
	}))

	invoice, err := client.Invoices.Get(context.Background(), "inv_1")
	if err != nil {
		t.Fatalf("invoices.get: %v", err)
	}
	if invoice.Status != corvonfse.InvoiceStatusCancelled {
		t.Errorf("expected status cancelled, got %q", invoice.Status)
	}
	if invoice.CancellationReason != corvonfse.CancellationReasonIssueError {
		t.Errorf("expected reason issue_error, got %q", invoice.CancellationReason)
	}
	if invoice.CancellationDescription != "Valor do servico informado errado" {
		t.Errorf("expected description, got %q", invoice.CancellationDescription)
	}
	if invoice.CancelledAt != "2026-10-13T10:00:00.000Z" {
		t.Errorf("expected cancelledAt, got %q", invoice.CancelledAt)
	}
	if invoice.CancellationRequestedAt != "2026-10-13T09:59:59.000Z" {
		t.Errorf("expected cancellationRequestedAt, got %q", invoice.CancellationRequestedAt)
	}
	if invoice.NFSENumber != 1001 {
		t.Errorf("expected other fields still decoded, got %+v", invoice)
	}
}

func TestInvoicesGetParsesCancellingInvoice(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"data": `+invoiceWithStatus("cancelling", `
			"cancellationReason": "other",
			"cancellationDescription": "Cancelamento solicitado pelo cliente",
			"cancelledAt": null,
			"cancellationRequestedAt": "2026-10-13T09:59:59.000Z"`)+`}`)
	}))

	invoice, err := client.Invoices.Get(context.Background(), "inv_1")
	if err != nil {
		t.Fatalf("invoices.get: %v", err)
	}
	if invoice.Status != corvonfse.InvoiceStatusCancelling {
		t.Errorf("expected status cancelling, got %q", invoice.Status)
	}
	if invoice.CancellationReason != corvonfse.CancellationReasonOther {
		t.Errorf("expected reason other, got %q", invoice.CancellationReason)
	}
	if invoice.CancelledAt != "" {
		t.Errorf("expected empty cancelledAt, got %q", invoice.CancelledAt)
	}
	if invoice.CancellationRequestedAt == "" {
		t.Error("expected cancellationRequestedAt set")
	}
}

func TestInvoicesGetWithoutCancellationFieldsLeavesThemEmpty(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"data": `+invoiceWithStatus("authorized", `
			"cancellationReason": null,
			"cancellationDescription": null,
			"cancelledAt": null,
			"cancellationRequestedAt": null`)+`}`)
	}))

	invoice, err := client.Invoices.Get(context.Background(), "inv_1")
	if err != nil {
		t.Fatalf("invoices.get: %v", err)
	}
	if invoice.CancellationReason != "" || invoice.CancellationDescription != "" ||
		invoice.CancelledAt != "" || invoice.CancellationRequestedAt != "" {
		t.Errorf("expected null cancellation fields to be empty, got %+v", invoice)
	}
}

func TestCancellationReasonValues(t *testing.T) {
	cases := map[corvonfse.CancellationReason]string{
		corvonfse.CancellationReasonIssueError:         "issue_error",
		corvonfse.CancellationReasonServiceNotProvided: "service_not_provided",
		corvonfse.CancellationReasonOther:              "other",
	}
	for reason, wire := range cases {
		if string(reason) != wire {
			t.Errorf("expected %q, got %q", wire, reason)
		}
	}
	if corvonfse.InvoiceStatusCancelling != "cancelling" || corvonfse.InvoiceStatusCancelled != "cancelled" {
		t.Error("expected cancelling/cancelled status values")
	}
}
