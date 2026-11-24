package test

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

// single returns a client whose server records the one request it receives
// into *call and answers status with body.
func single(t *testing.T, call *capture, status int, body string) *corvonfse.Client {
	t.Helper()
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*call = record(r)
		if body == "" {
			w.WriteHeader(status)
			return
		}
		writeJSON(w, status, body)
	}))
	return client
}

func mustJSON(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("body is not a JSON object: %q (%v)", raw, err)
	}
	return out
}

func TestUserAgentCarriesTheVersion(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {}}`)
	_, _ = client.FiscalPortal.Status(context.Background())
	if call.header.Get("User-Agent") != "corvo-nfse-go/"+corvonfse.Version {
		t.Errorf("unexpected User-Agent %q", call.header.Get("User-Agent"))
	}
}

func TestInvoicesCreateScheduleAnswersWithASchedule(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusAccepted, `{"data": {
		"id": "sch_1", "company_id": "co_1", "kind": "recurring", "frequency": "MONTHLY",
		"anchor_on": "2026-12-01", "ends_on": null, "next_run_on": "2026-12-01", "last_run_on": null,
		"status": "active", "pause_reason": null, "attempts": 0, "created_at": "2026-11-24T10:00:00Z",
		"invoice": {"amount": 10}
	}}`)

	result, err := client.Invoices.Create(context.Background(), corvonfse.InvoiceCreateParams{
		CompanyID:      "co_1",
		Amount:         10,
		CompetenceDate: "2026-12-01",
		PurchaseOrder:  "PO-1",
		Notes:          "obs",
		Options: &corvonfse.InvoiceCreateOptions{
			Type:       corvonfse.InvoiceTypeRecurring,
			Recurrence: &corvonfse.InvoiceRecurrence{Frequency: corvonfse.FrequencyMonthly, EndsOn: "2027-12-01"},
		},
	}, "key-1")
	if err != nil {
		t.Fatalf("invoices.create: %v", err)
	}
	if !result.IsSchedule() || result.Invoice != nil {
		t.Fatalf("expected a schedule, got %+v", result)
	}
	if result.Schedule.ID != "sch_1" || result.Schedule.AnchorOn != "2026-12-01" || result.Schedule.Kind != "recurring" ||
		result.Schedule.Frequency != "MONTHLY" || result.Schedule.Status != corvonfse.ScheduleActive || result.Schedule.NextRunOn != "2026-12-01" {
		t.Errorf("schedule not decoded: %+v", result.Schedule)
	}
	if call.header.Get("Idempotency-Key") != "key-1" {
		t.Errorf("expected Idempotency-Key key-1, got %q", call.header.Get("Idempotency-Key"))
	}
	body := mustJSON(t, call.body)
	options := body["options"].(map[string]any)
	recurrence := options["recurrence"].(map[string]any)
	if options["type"] != "RECURRING" || recurrence["frequency"] != "MONTHLY" || recurrence["endsOn"] != "2027-12-01" {
		t.Errorf("unexpected options body: %v", options)
	}
	if body["purchaseOrder"] != "PO-1" || body["notes"] != "obs" {
		t.Errorf("expected purchaseOrder and notes in the body, got %v", body)
	}
}

func TestInvoicesCreateImmediateAnswersWithAnInvoice(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusAccepted, `{"data": {"id": "inv_9", "status": "queued", "stage": "pending",
		"documentsStatus": "pending", "documentsReadyAt": null, "emailStatus": null, "lastEmailStatus": null,
		"substitutesInvoiceId": null, "portalFlag": "taker_rejected",
		"serviceSnapshot": {"code": "010101", "description": "x", "municipalityCode": "3550308", "purchaseOrder": "PO-1", "notes": "n"}}}`)

	result, err := client.Invoices.Create(context.Background(), corvonfse.InvoiceCreateParams{
		CompanyID: "co_1",
		Customer: &corvonfse.InvoiceCustomerInput{
			Name: "Ana",
			Foreign: &corvonfse.ForeignCustomer{
				NoTaxIDReason: "exempt", Country: "PT", PostalCode: "1000", City: "Lisboa",
				Region: "LI", Street: "Rua", Number: "1", District: "Centro",
			},
		},
	}, "")
	if err != nil {
		t.Fatalf("invoices.create: %v", err)
	}
	if result.IsSchedule() || result.Invoice == nil || result.Invoice.ID != "inv_9" {
		t.Fatalf("expected an invoice, got %+v", result)
	}
	if result.Invoice.PortalFlag != "taker_rejected" || result.Invoice.ServiceSnapshot.PurchaseOrder != "PO-1" ||
		result.Invoice.ServiceSnapshot.Notes != "n" || result.Invoice.Stage != corvonfse.InvoiceStagePending {
		t.Errorf("invoice fields not decoded: %+v", result.Invoice)
	}
	if call.header.Get("Idempotency-Key") != "" {
		t.Errorf("expected no Idempotency-Key, got %q", call.header.Get("Idempotency-Key"))
	}
	customer := mustJSON(t, call.body)["customer"].(map[string]any)
	if _, has := customer["taxId"]; has {
		t.Errorf("a foreign customer must not send taxId, got %v", customer)
	}
	foreign := customer["foreign"].(map[string]any)
	if foreign["noTaxIdReason"] != "exempt" || foreign["country"] != "PT" {
		t.Errorf("unexpected foreign body: %v", foreign)
	}
}

func TestInvoicesListFiltersByCustomerTaxID(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": [], "pagination": {"limit": 50, "hasMore": false, "nextCursor": null}}`)
	_, _, err := client.Invoices.List(context.Background(), &corvonfse.InvoiceListParams{CustomerTaxID: "12345678901"})
	if err != nil {
		t.Fatalf("invoices.list: %v", err)
	}
	if call.query != "customerTaxId=12345678901" {
		t.Errorf("unexpected query %q", call.query)
	}
}

func TestInvoicesSendEmailToAnotherAddress(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusAccepted, `{"data": {"emailId": "m1", "status": "queued"}}`)
	_, err := client.Invoices.SendEmail(context.Background(), "inv_1", &corvonfse.InvoiceEmailParams{Email: "outro@example.com"}, "k")
	if err != nil {
		t.Fatalf("invoices.sendEmail: %v", err)
	}
	if mustJSON(t, call.body)["email"] != "outro@example.com" || call.header.Get("Idempotency-Key") != "k" {
		t.Errorf("unexpected email request: body=%q key=%q", call.body, call.header.Get("Idempotency-Key"))
	}
}

func TestInvoicesSubstitute(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusAccepted, `{"data": {"id": "inv_2", "status": "queued", "substitutesInvoiceId": "inv_1", "substitutionDescription": "valor errado"}}`)
	invoice, err := client.Invoices.Substitute(context.Background(), "inv_1", corvonfse.InvoiceSubstituteParams{
		Description:   "valor errado",
		Service:       &corvonfse.InvoiceSubstituteService{Description: "novo"},
		PurchaseOrder: "PO-2",
	}, "sub-1")
	if err != nil {
		t.Fatalf("invoices.substitute: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/invoices/inv_1/substitute" || call.header.Get("Idempotency-Key") != "sub-1" {
		t.Errorf("unexpected request: %s %s key=%q", call.method, call.path, call.header.Get("Idempotency-Key"))
	}
	if invoice.SubstitutesInvoiceID != "inv_1" || invoice.SubstitutionDescription != "valor errado" {
		t.Errorf("substitution fields not decoded: %+v", invoice)
	}
	body := mustJSON(t, call.body)
	if body["description"] != "valor errado" || body["purchaseOrder"] != "PO-2" || body["service"].(map[string]any)["description"] != "novo" {
		t.Errorf("unexpected body: %v", body)
	}
}

func TestInvoicesReviewAndOverview(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusAccepted, `{"data": {"results": [
		{"id": "a", "outcome": "accepted"}, {"id": "b", "outcome": "skipped", "reason": "not_in_manual_review"}]}}`)
	result, err := client.Invoices.Review(context.Background(), corvonfse.InvoiceReviewParams{
		Action: corvonfse.ReviewRetry, InvoiceIDs: []string{"a", "b"},
	})
	if err != nil {
		t.Fatalf("invoices.review: %v", err)
	}
	if call.path != "/v1/invoices/review" || len(result.Results) != 2 || result.Results[1].Reason != "not_in_manual_review" {
		t.Errorf("unexpected review: path=%s result=%+v", call.path, result)
	}
	if mustJSON(t, call.body)["action"] != "retry" {
		t.Errorf("unexpected body %q", call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"period": "30d", "from": "2026-10-26", "to": "2026-11-24", "timezone": "America/Sao_Paulo",
		"counts": {"authorized": 5, "processing": 1, "rejected": 0}, "dailyActivity": [{"date": "2026-11-24", "count": 2}], "recentInvoices": [{"id": "x"}]}}`)
	overview, err := client.Invoices.Overview(context.Background(), corvonfse.OverviewPeriod30d)
	if err != nil {
		t.Fatalf("invoices.overview: %v", err)
	}
	if call.path != "/v1/invoices/overview" || call.query != "period=30d" {
		t.Errorf("unexpected request: %s ?%s", call.path, call.query)
	}
	if overview.Counts.Authorized != 5 || overview.DailyActivity[0].Count != 2 || overview.RecentInvoices[0].ID != "x" {
		t.Errorf("overview not decoded: %+v", overview)
	}
}

func TestCreditsPreviewPurchase(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {"units": 10, "unitPriceCents": 99, "totalCents": 990}}`)
	preview, err := client.Credits.PreviewPurchase(context.Background(), 10)
	if err != nil {
		t.Fatalf("credits.previewPurchase: %v", err)
	}
	if call.method != http.MethodGet || call.path != "/v1/credits/purchase/preview" || call.query != "units=10" {
		t.Errorf("unexpected request: %s %s ?%s", call.method, call.path, call.query)
	}
	if preview.Units != 10 || preview.UnitPriceCents != 99 || preview.TotalCents != 990 {
		t.Errorf("preview not decoded: %+v", preview)
	}
}

func TestBillingPlanNotFoundHasNoPlanAssignedCode(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusNotFound, `{"statusCode": 404, "code": "NO_PLAN_ASSIGNED", "message": "no plan", "requestId": "r1"}`)
	_, err := client.Billing.Plan(context.Background())
	err = requireError(t, err)
	if err.Status != 404 || err.Code != "NO_PLAN_ASSIGNED" || err.RequestID != "r1" {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestErrorDetailsAsAnObjectDoNotBreakParsing(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusUnprocessableEntity, `{"statusCode": 422, "code": "PACKAGE_TOO_LARGE", "message": "grande", "details": {"count": 501, "limit": 500}}`)
	_, err := client.InvoicePackage.Summary(context.Background(), "co_1", "2026-10")
	err = requireError(t, err)
	if err.Code != "PACKAGE_TOO_LARGE" || err.Message != "grande" || err.Details != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestRateLimitErrorKeepsRetryAfter(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "12")
		writeJSON(w, http.StatusTooManyRequests, `{"statusCode": 429, "code": "ERROR_RATE_LIMITED", "message": "slow down"}`)
	}))
	_, _, err := client.Schedules.List(context.Background(), nil)
	err = requireError(t, err)
	if err.Status != 429 || err.RetryAfter != 12 {
		t.Errorf("expected 429 with RetryAfter 12, got %+v", err)
	}
}

func TestCNPJLookup(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {"active": true, "legalName": "ACME LTDA", "tradeName": "Acme", "situation": "ATIVA",
		"mei": true, "simples": null, "municipalityCode": "3550308",
		"address": {"street": "Rua A", "number": "1", "district": "Centro", "city": "Sao Paulo", "state": "SP", "zip": "01000000"}}}`)
	profile, err := client.CNPJ.Lookup(context.Background(), "12345678000199")
	if err != nil {
		t.Fatalf("cnpj.lookup: %v", err)
	}
	if call.path != "/v1/cnpj/12345678000199" {
		t.Errorf("unexpected path %s", call.path)
	}
	if !profile.Active || profile.LegalName != "ACME LTDA" || profile.MEI == nil || !*profile.MEI || profile.Simples != nil || profile.Address.City != "Sao Paulo" {
		t.Errorf("profile not decoded: %+v", profile)
	}
}

func TestCustomers(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": [{"taxId": "12345678901", "name": "Ana", "email": null,
		"createdAt": "c", "updatedAt": "u", "lastUsedAt": "l"}], "pagination": {"limit": 20, "hasMore": true, "nextCursor": "20"}}`)
	customers, pagination, err := client.Customers.Search(context.Background(), &corvonfse.CustomerSearchParams{Search: "Ana", Limit: 20, Cursor: "0"})
	if err != nil {
		t.Fatalf("customers.search: %v", err)
	}
	if call.path != "/v1/customers" || call.query != "cursor=0&limit=20&search=Ana" {
		t.Errorf("unexpected request: %s ?%s", call.path, call.query)
	}
	if len(customers) != 1 || customers[0].Name != "Ana" || customers[0].Email != "" || !pagination.HasMore || pagination.NextCursor != "20" {
		t.Errorf("unexpected page: %+v %+v", customers, pagination)
	}

	// PATCH with an explicit null clears the e-mail.
	client = single(t, &call, http.StatusOK, `{"data": {"taxId": "12345678901", "name": "Ana Maria"}}`)
	customer, err := client.Customers.Update(context.Background(), "12345678901", corvonfse.CustomerUpdateParams{
		Name: "Ana Maria", Email: corvonfse.Null[string](),
	})
	if err != nil {
		t.Fatalf("customers.update: %v", err)
	}
	if call.method != http.MethodPatch || call.path != "/v1/customers/12345678901" || customer.Name != "Ana Maria" {
		t.Errorf("unexpected update: %s %s %+v", call.method, call.path, customer)
	}
	body := mustJSON(t, call.body)
	if email, has := body["email"]; !has || email != nil {
		t.Errorf("expected email:null in the body, got %q", call.body)
	}

	// Leaving the e-mail out sends only the name.
	_, _ = client.Customers.Update(context.Background(), "1", corvonfse.CustomerUpdateParams{Name: "X"})
	if call.body != `{"name":"X"}` {
		t.Errorf("expected only the name, got %q", call.body)
	}
	_, _ = client.Customers.Update(context.Background(), "1", corvonfse.CustomerUpdateParams{Email: corvonfse.Some("a@b.co")})
	if call.body != `{"email":"a@b.co"}` {
		t.Errorf("expected the e-mail value, got %q", call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": null}`)
	if err := client.Customers.Remove(context.Background(), "12345678901"); err != nil {
		t.Fatalf("customers.remove: %v", err)
	}
	if call.method != http.MethodDelete || call.path != "/v1/customers/12345678901" {
		t.Errorf("unexpected delete: %s %s", call.method, call.path)
	}
}

func TestServiceFavorites(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusCreated, `{"data": {"id": "f1", "companyId": "co_1", "tenantId": "t", "label": "Consultoria",
		"serviceCode": "010101", "description": "d", "amountCents": 15000, "municipalityCode": null, "createdAt": "c"}}`)
	amount := 15000
	favorite, err := client.ServiceFavorites.Create(context.Background(), "co_1", corvonfse.ServiceFavoriteCreateParams{
		Label: "Consultoria", ServiceCode: "010101", Description: "d", AmountCents: &amount,
	})
	if err != nil {
		t.Fatalf("serviceFavorites.create: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/companies/co_1/service-favorites" {
		t.Errorf("unexpected request: %s %s", call.method, call.path)
	}
	if favorite.ID != "f1" || favorite.AmountCents == nil || *favorite.AmountCents != 15000 || favorite.MunicipalityCode != "" {
		t.Errorf("favorite not decoded: %+v", favorite)
	}
	if mustJSON(t, call.body)["amountCents"] != float64(15000) {
		t.Errorf("unexpected body %q", call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": [{"id": "f1"}, {"id": "f2"}]}`)
	list, err := client.ServiceFavorites.List(context.Background(), "co_1")
	if err != nil || len(list) != 2 || list[1].ID != "f2" {
		t.Fatalf("serviceFavorites.list: %v %+v", err, list)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"deleted": true}}`)
	deleted, err := client.ServiceFavorites.Remove(context.Background(), "co_1", "f1")
	if err != nil || !deleted || call.method != http.MethodDelete || call.path != "/v1/companies/co_1/service-favorites/f1" {
		t.Errorf("serviceFavorites.remove: %v %v %s %s", err, deleted, call.method, call.path)
	}
}

func TestAccountantPackage(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {"email": "contador@example.com", "lastSentMonth": "2026-10", "testEmailSent": true}}`)
	updated, err := client.AccountantPackage.Update(context.Background(), "co_1", corvonfse.AccountantPackageUpdateParams{
		AccountantEmail: corvonfse.Some("contador@example.com"),
	})
	if err != nil {
		t.Fatalf("accountantPackage.update: %v", err)
	}
	if call.method != http.MethodPut || call.path != "/v1/companies/co_1/accountant-package" || call.body != `{"accountantEmail":"contador@example.com"}` {
		t.Errorf("unexpected request: %s %s %q", call.method, call.path, call.body)
	}
	if !updated.TestEmailSent || updated.Email != "contador@example.com" {
		t.Errorf("update not decoded: %+v", updated)
	}

	// null turns the monthly send off.
	_, _ = client.AccountantPackage.Update(context.Background(), "co_1", corvonfse.AccountantPackageUpdateParams{AccountantEmail: corvonfse.Null[string]()})
	if call.body != `{"accountantEmail":null}` {
		t.Errorf("expected accountantEmail:null, got %q", call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"status": "skipped-already-sent"}}`)
	sent, err := client.AccountantPackage.Send(context.Background(), "co_1", "2026-10")
	if err != nil || sent.Status != "skipped-already-sent" || call.path != "/v1/companies/co_1/accountant-package/send" || call.body != `{"month":"2026-10"}` {
		t.Errorf("accountantPackage.send: %v %+v %s %q", err, sent, call.path, call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": [{"id": "s1", "month": "2026-10", "recipient": "c@x.co", "status": "sent", "trigger": "manual", "invoiceCount": 3, "totalCents": 30000, "createdAt": "c"}]}`)
	sends, err := client.AccountantPackage.ListSends(context.Background(), "co_1")
	if err != nil || len(sends) != 1 || sends[0].InvoiceCount != 3 || sends[0].Trigger != "manual" || call.path != "/v1/companies/co_1/accountant-package/sends" {
		t.Errorf("accountantPackage.listSends: %v %+v %s", err, sends, call.path)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"email": null, "lastSentMonth": null}}`)
	settings, err := client.AccountantPackage.Get(context.Background(), "co_1")
	if err != nil || settings.Email != "" {
		t.Errorf("accountantPackage.get: %v %+v", err, settings)
	}
}

func TestInvoicePackage(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {"companyName": "ACME", "companyTaxId": "123", "month": "2026-10", "count": 4, "limit": 500, "total": 1234.5}}`)
	summary, err := client.InvoicePackage.Summary(context.Background(), "co_1", "2026-10")
	if err != nil {
		t.Fatalf("invoicePackage.summary: %v", err)
	}
	if call.path != "/v1/companies/co_1/invoice-package/summary" || call.query != "month=2026-10" || summary.Count != 4 || summary.Total != 1234.5 {
		t.Errorf("unexpected summary: %s ?%s %+v", call.path, call.query, summary)
	}

	zip := []byte("PK\x03\x04binary\x00zip")
	server, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="notas-2026-10.zip"`)
		_, _ = w.Write(zip)
	}))
	file, err := server.InvoicePackage.Download(context.Background(), "co_1", "2026-10", []string{"xml", "pdf"})
	if err != nil {
		t.Fatalf("invoicePackage.download: %v", err)
	}
	if call.path != "/v1/companies/co_1/invoice-package" || call.query != "include=xml%2Cpdf&month=2026-10" {
		t.Errorf("unexpected request: %s ?%s", call.path, call.query)
	}
	if string(file.Buffer) != string(zip) || file.MimeType != "application/zip" || file.FileName != "notas-2026-10.zip" {
		t.Errorf("unexpected file: %+v", file)
	}
}

func TestDPSSequenceAndFiscalPortal(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {"series": "900", "lastNumber": 42}}`)
	last := 100
	seq, err := client.DPSSequence.Set(context.Background(), "co_1", corvonfse.DPSSequenceSetParams{LastNumber: &last})
	if err != nil {
		t.Fatalf("dpsSequence.set: %v", err)
	}
	if call.method != http.MethodPut || call.path != "/v1/companies/co_1/dps-sequence" || call.body != `{"lastNumber":100}` || seq.LastNumber != 42 || seq.Series != "900" {
		t.Errorf("unexpected set: %s %s %q %+v", call.method, call.path, call.body, seq)
	}

	client = single(t, &call, http.StatusBadRequest, `{"statusCode": 400, "code": "DPS_NUMBER_BELOW_CURRENT", "message": "menor"}`)
	_, err = client.DPSSequence.Set(context.Background(), "co_1", corvonfse.DPSSequenceSetParams{LastNumber: &last})
	if e := requireError(t, err); e.Code != "DPS_NUMBER_BELOW_CURRENT" {
		t.Errorf("unexpected error %+v", e)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"available": false, "environment": "live", "since": "2026-11-24T10:00:00Z", "retryAt": null}}`)
	status, err := client.FiscalPortal.Status(context.Background())
	if err != nil || status.Available || status.Environment != corvonfse.EnvironmentLive || status.Since == "" || call.path != "/v1/fiscal-portal/status" {
		t.Errorf("fiscalPortal.status: %v %+v %s", err, status, call.path)
	}
}

func TestSchedules(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": [{"id": "s1", "kind": "once", "status": "paused", "pause_reason": "failures", "frequency": null}],
		"pagination": {"limit": 10, "hasMore": false, "nextCursor": null}}`)
	list, _, err := client.Schedules.List(context.Background(), &corvonfse.ScheduleListParams{
		Status: corvonfse.SchedulePaused, CompanyID: "co_1", Limit: 10,
	})
	if err != nil {
		t.Fatalf("schedules.list: %v", err)
	}
	if call.query != "companyId=co_1&limit=10&status=paused" || len(list) != 1 || list[0].PauseReason != "failures" || list[0].Frequency != "" {
		t.Errorf("unexpected list: ?%s %+v", call.query, list)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"id": "s1", "status": "active", "kind": "recurring"}}`)
	_, err = client.Schedules.Update(context.Background(), "s1", corvonfse.ScheduleUpdateParams{
		Frequency: corvonfse.FrequencyWeekly, EndsOn: corvonfse.Null[string](), Invoice: map[string]any{"amount": 20},
	})
	if err != nil {
		t.Fatalf("schedules.update: %v", err)
	}
	body := mustJSON(t, call.body)
	if call.method != http.MethodPatch || call.path != "/v1/schedules/s1" || body["frequency"] != "WEEKLY" || body["endsOn"] != nil {
		t.Errorf("unexpected update: %s %s %q", call.method, call.path, call.body)
	}
	if _, has := body["endsOn"]; !has {
		t.Errorf("expected endsOn:null, got %q", call.body)
	}

	for name, run := range map[string]func() (corvonfse.Schedule, *corvonfse.CorvoError){
		"/v1/schedules/s1/pause": func() (corvonfse.Schedule, *corvonfse.CorvoError) {
			return client.Schedules.Pause(context.Background(), "s1")
		},
		"/v1/schedules/s1/resume": func() (corvonfse.Schedule, *corvonfse.CorvoError) {
			return client.Schedules.Resume(context.Background(), "s1")
		},
		"/v1/schedules/s1/cancel": func() (corvonfse.Schedule, *corvonfse.CorvoError) {
			return client.Schedules.Cancel(context.Background(), "s1")
		},
	} {
		if _, err := run(); err != nil || call.method != http.MethodPost || call.path != name {
			t.Errorf("expected POST %s, got %s %s (%v)", name, call.method, call.path, err)
		}
	}

	client = single(t, &call, http.StatusOK, `{"data": [{"id": "r1", "outcome": "failed", "error_code": "E0014", "attempt": 2, "invoice_id": null}],
		"pagination": {"limit": 50, "hasMore": false, "nextCursor": null}}`)
	runs, _, err := client.Schedules.ListRuns(context.Background(), "s1", nil)
	if err != nil || len(runs) != 1 || runs[0].ErrorCode != "E0014" || runs[0].Attempt != 2 || call.path != "/v1/schedules/s1/runs" {
		t.Errorf("schedules.listRuns: %v %+v %s", err, runs, call.path)
	}
}

func TestImportsStartSendsMultipart(t *testing.T) {
	var call capture
	var fields = map[string]string{}
	var fileType, fileName, fileBody string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		mediaType, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if mediaType != "multipart/form-data" {
			t.Errorf("expected multipart, got %q", mediaType)
		}
		reader := multipart.NewReader(strings.NewReader(call.body), params["boundary"])
		for {
			part, err := reader.NextPart()
			if err != nil {
				break
			}
			data, _ := io.ReadAll(part)
			if part.FormName() == "file" {
				fileType, fileName, fileBody = part.Header.Get("Content-Type"), part.FileName(), string(data)
			} else {
				fields[part.FormName()] = string(data)
			}
		}
		writeJSON(w, http.StatusAccepted, `{"data": {"id": "imp_1", "status": "processing", "duplicate_of": null}}`)
	}))

	started, err := client.Imports.Start(context.Background(), []byte("cnpj;valor\n1;2"), "notas.xlsx", "co_1")
	if err != nil {
		t.Fatalf("imports.start: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/imports" {
		t.Errorf("unexpected request: %s %s", call.method, call.path)
	}
	if started.ID != "imp_1" || started.Status != corvonfse.ImportProcessing || started.DuplicateOf != "" {
		t.Errorf("unexpected result: %+v", started)
	}
	if fileName != "notas.xlsx" || fileBody != "cnpj;valor\n1;2" || !strings.Contains(fileType, "spreadsheetml") || fields["companyId"] != "co_1" {
		t.Errorf("unexpected multipart: name=%q type=%q body=%q fields=%v", fileName, fileType, fileBody, fields)
	}

	// A csv with no company: default name, text/csv, no companyId field.
	fields = map[string]string{}
	if _, err := client.Imports.Start(context.Background(), []byte("a"), "", ""); err != nil {
		t.Fatalf("imports.start csv: %v", err)
	}
	if fileName != "import.csv" || fileType != "text/csv" || len(fields) != 0 {
		t.Errorf("unexpected csv multipart: name=%q type=%q fields=%v", fileName, fileType, fields)
	}
}

func TestImports(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {"id": "imp_1", "status": "ready", "company_id": null, "file_name": "a.csv",
		"rows_expected": 3, "unknown_columns": ["x"], "created_at": "c", "emit_started_at": null, "emit_finished_at": null, "error_code": null,
		"counters": {"total": 3, "valid": 2, "invalid": 1, "authorized_amount": 10.5, "awaiting_credits": 0}}}`)
	job, err := client.Imports.Get(context.Background(), "imp_1")
	if err != nil {
		t.Fatalf("imports.get: %v", err)
	}
	if job.Status != corvonfse.ImportReady || job.RowsExpected == nil || *job.RowsExpected != 3 || job.Counters.Valid != 2 || job.Counters.AuthorizedAmount != 10.5 || job.UnknownColumns[0] != "x" || job.CompanyID != "" {
		t.Errorf("job not decoded: %+v", job)
	}

	yes := true
	_, err = client.Imports.Emit(context.Background(), "imp_1", &corvonfse.ImportEmitParams{SendEmail: &yes})
	if call.method != http.MethodPost || call.path != "/v1/imports/imp_1/emit" || call.body != `{"sendEmail":true}` || err != nil {
		t.Errorf("imports.emit: %v %s %s %q", err, call.method, call.path, call.body)
	}
	_, _ = client.Imports.Emit(context.Background(), "imp_1", nil)
	if call.body != `{}` {
		t.Errorf("expected an empty object body, got %q", call.body)
	}

	_, _ = client.Imports.List(context.Background(), 5)
	if call.path != "/v1/imports" || call.query != "limit=5" {
		t.Errorf("imports.list: %s ?%s", call.path, call.query)
	}

	_, _ = client.Imports.UpdateRow(context.Background(), "imp_1", 4, map[string]any{"amount": 99})
	if call.method != http.MethodPatch || call.path != "/v1/imports/imp_1/rows/4" || call.body != `{"item":{"amount":99}}` {
		t.Errorf("imports.updateRow: %s %s %q", call.method, call.path, call.body)
	}

	client = single(t, &call, http.StatusNoContent, "")
	if err := client.Imports.Discard(context.Background(), "imp_1"); err != nil || call.method != http.MethodDelete || call.path != "/v1/imports/imp_1" {
		t.Errorf("imports.discard: %v %s %s", err, call.method, call.path)
	}
}

func TestImportsListRowsPaginates(t *testing.T) {
	page := 0
	var queries []string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		page++
		if page == 1 {
			writeJSON(w, http.StatusOK, `{"data": [{"line": 2, "status": "invalid", "error_code": "ERROR_IMPORT_ROW", "error_message": "x", "invoice_id": null, "data": {"a": 1}}],
				"pagination": {"limit": 1, "hasMore": true, "nextCursor": "1"}}`)
			return
		}
		writeJSON(w, http.StatusOK, `{"data": [{"line": 3, "status": "invalid"}], "pagination": {"limit": 1, "hasMore": false, "nextCursor": null}}`)
	}))
	var lines []int
	err := client.Imports.IterateRows(context.Background(), "imp_1", &corvonfse.ImportRowListParams{Status: corvonfse.ImportRowInvalid, Limit: 1},
		func(row corvonfse.ImportRow) bool {
			lines = append(lines, row.Line)
			return true
		})
	if err != nil {
		t.Fatalf("imports.iterateRows: %v", err)
	}
	if len(lines) != 2 || lines[0] != 2 || lines[1] != 3 {
		t.Errorf("expected lines 2 and 3, got %v", lines)
	}
	if len(queries) != 2 || queries[0] != "limit=1&status=invalid" || queries[1] != "cursor=1&limit=1&status=invalid" {
		t.Errorf("unexpected queries: %v", queries)
	}
}

func TestImportsDownloads(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="modelo.csv"`)
		_, _ = w.Write([]byte("a;b\n"))
	}))
	template, err := client.Imports.DownloadTemplate(context.Background())
	if err != nil || call.path != "/v1/imports/template" || string(template.Buffer) != "a;b\n" || template.FileName != "modelo.csv" || !strings.HasPrefix(template.MimeType, "text/csv") {
		t.Errorf("imports.downloadTemplate: %v %s %+v", err, call.path, template)
	}
	errorsFile, err := client.Imports.DownloadErrors(context.Background(), "imp_1")
	if err != nil || call.path != "/v1/imports/imp_1/errors.csv" || len(errorsFile.Buffer) == 0 {
		t.Errorf("imports.downloadErrors: %v %s %+v", err, call.path, errorsFile)
	}
}

func TestFiscalAndServiceTaxProfiles(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {"id": "fp_1", "taxRegime": "MEI", "reviewStatus": "unreviewed", "validFrom": "2026-01-01", "validUntil": null, "simplesApuracao": null}}`)
	confirm := true
	profile, err := client.FiscalProfiles.Upsert(context.Background(), "co_1", corvonfse.FiscalProfileUpsertParams{TaxRegime: "MEI", Confirm: &confirm})
	if err != nil {
		t.Fatalf("fiscalProfiles.upsert: %v", err)
	}
	if call.method != http.MethodPut || call.path != "/v1/companies/co_1/fiscal-profile" || profile.ID != "fp_1" || profile.ValidUntil != "" ||
		mustJSON(t, call.body)["taxRegime"] != "MEI" {
		t.Errorf("unexpected upsert: %s %s %q %+v", call.method, call.path, call.body, profile)
	}
	_, _ = client.FiscalProfiles.Approve(context.Background(), "co_1", "fp_1")
	if call.method != http.MethodPost || call.path != "/v1/companies/co_1/fiscal-profile/fp_1/approve" {
		t.Errorf("fiscalProfiles.approve: %s %s", call.method, call.path)
	}

	client = single(t, &call, http.StatusOK, `{"data": [{"id": "sp_1", "name": "Consultoria", "withholdingRates": {"irrf": "1.5"}, "regimeApplicability": ["MEI"], "issWithholdingDefault": true}]}`)
	profiles, err := client.ServiceTaxProfiles.List(context.Background(), "co_1", "2026-11-24")
	if err != nil {
		t.Fatalf("serviceTaxProfiles.list: %v", err)
	}
	if call.query != "date=2026-11-24" || len(profiles) != 1 || profiles[0].WithholdingRates.IRRF != "1.5" || !profiles[0].IssWithholdingDefault {
		t.Errorf("unexpected list: ?%s %+v", call.query, profiles)
	}

	_, _ = client.ServiceTaxProfiles.Update(context.Background(), "co_1", "sp_1", corvonfse.ServiceTaxProfileUpdateParams{
		Name: "Novo", IssRate: corvonfse.Null[string](), EffectiveUntil: corvonfse.Some("2027-01-01"),
	})
	body := mustJSON(t, call.body)
	if call.method != http.MethodPatch || call.path != "/v1/companies/co_1/service-tax-profiles/sp_1" || body["name"] != "Novo" || body["effectiveUntil"] != "2027-01-01" {
		t.Errorf("unexpected update: %s %s %q", call.method, call.path, call.body)
	}
	if v, has := body["issRate"]; !has || v != nil {
		t.Errorf("expected issRate:null, got %q", call.body)
	}

	_, _ = client.ServiceTaxProfiles.Create(context.Background(), "co_1", corvonfse.ServiceTaxProfileCreateParams{Name: "N", ServiceCode: "010101", EffectiveFrom: "2026-01-01"})
	if call.method != http.MethodPost || call.path != "/v1/companies/co_1/service-tax-profiles" {
		t.Errorf("serviceTaxProfiles.create: %s %s", call.method, call.path)
	}
	_ = client.ServiceTaxProfiles.Remove(context.Background(), "co_1", "sp_1")
	if call.method != http.MethodDelete || call.path != "/v1/companies/co_1/service-tax-profiles/sp_1" {
		t.Errorf("serviceTaxProfiles.remove: %s %s", call.method, call.path)
	}
	_, _ = client.ServiceTaxProfiles.Approve(context.Background(), "co_1", "sp_1")
	if call.method != http.MethodPost || call.path != "/v1/companies/co_1/service-tax-profiles/sp_1/approve" {
		t.Errorf("serviceTaxProfiles.approve: %s %s", call.method, call.path)
	}
}

func TestIntegrations(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": [{"id": "stripe", "name": "Stripe", "version": "1", "supportedFlows": ["payment"],
		"credentialFields": [{"key": "apiKey", "label": "API key", "secret": true, "revealOnce": true}],
		"webhookInstructions": {"en": "set", "pt": "configure"}}]}`)
	providers, err := client.Integrations.ListProviders(context.Background())
	if err != nil || len(providers) != 1 || !providers[0].CredentialFields[0].Secret || providers[0].WebhookInstructions["pt"] != "configure" || call.path != "/v1/integrations/providers" {
		t.Errorf("integrations.listProviders: %v %+v %s", err, providers, call.path)
	}

	client = single(t, &call, http.StatusCreated, `{"data": {"id": "c1", "provider": "stripe", "state": "draft", "environment": "test",
		"companyId": "co_1", "webhookUrl": "https://api.corvo.tech/hook", "activeFrom": null, "externalAccountId": null, "createdAt": "c",
		"secrets": {"webhookSecret": "whsec"}}}`)
	created, err := client.IntegrationConnections.Create(context.Background(), corvonfse.IntegrationConnectionCreateParams{
		CompanyID: "co_1", Provider: "stripe", Credentials: map[string]string{"apiKey": "sk"}, Environment: corvonfse.EnvironmentTest,
	})
	if err != nil {
		t.Fatalf("integrationConnections.create: %v", err)
	}
	if created.ID != "c1" || created.State != "draft" || created.Secrets["webhookSecret"] != "whsec" || created.Environment != corvonfse.EnvironmentTest {
		t.Errorf("connection not decoded: %+v", created)
	}
	if mustJSON(t, call.body)["environment"] != "test" {
		t.Errorf("unexpected body %q", call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": [{"id": "r1", "sendEmail": true, "service": {"code": "010101"}, "externalProductId": null}]}`)
	yes := true
	product := "prod_1"
	rules, err := client.IntegrationConnections.ReplaceRules(context.Background(), "c1", []corvonfse.IntegrationRuleInput{
		{SendEmail: true, Service: corvonfse.InvoiceServiceInput{Code: "010101"}, CancelOnRefund: &yes, ExternalProductID: &product},
	})
	if err != nil || len(rules) != 1 || !rules[0].SendEmail || call.method != http.MethodPut || call.path != "/v1/integrations/connections/c1/rules" {
		t.Errorf("integrationConnections.replaceRules: %v %+v %s %s", err, rules, call.method, call.path)
	}
	rule := mustJSON(t, call.body)["rules"].([]any)[0].(map[string]any)
	if rule["sendEmail"] != true || rule["externalProductId"] != "prod_1" || rule["cancelOnRefund"] != true {
		t.Errorf("unexpected rules body: %q", call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"id": "c1", "state": "active"}}`)
	for suffix, run := range map[string]func() (corvonfse.IntegrationConnection, *corvonfse.CorvoError){
		"/activate": func() (corvonfse.IntegrationConnection, *corvonfse.CorvoError) {
			return client.IntegrationConnections.Activate(context.Background(), "c1")
		},
		"/pause": func() (corvonfse.IntegrationConnection, *corvonfse.CorvoError) {
			return client.IntegrationConnections.Pause(context.Background(), "c1")
		},
		"/disconnect": func() (corvonfse.IntegrationConnection, *corvonfse.CorvoError) {
			return client.IntegrationConnections.Disconnect(context.Background(), "c1")
		},
		"/validate": func() (corvonfse.IntegrationConnection, *corvonfse.CorvoError) {
			return client.IntegrationConnections.Validate(context.Background(), "c1")
		},
	} {
		if _, err := run(); err != nil || call.method != http.MethodPost || call.path != "/v1/integrations/connections/c1"+suffix {
			t.Errorf("expected POST .../c1%s, got %s %s (%v)", suffix, call.method, call.path, err)
		}
	}

	_, _ = client.IntegrationConnections.UpdateCredentials(context.Background(), "c1", map[string]string{"apiKey": "new"})
	if call.method != http.MethodPut || call.body != `{"credentials":{"apiKey":"new"}}` {
		t.Errorf("updateCredentials: %s %q", call.method, call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": [{"id": "op1", "amountCents": 5000, "fiscalStatus": "needs_attention", "commercialState": "paid",
		"customer": {"name": null, "email": null, "taxId": null, "taxIdMasked": false}, "invoiceId": null}]}`)
	ops, err := client.IntegrationConnections.ListOperations(context.Background(), "c1", 20)
	if err != nil || call.query != "limit=20" || len(ops) != 1 || ops[0].AmountCents != 5000 || ops[0].Customer == nil || ops[0].FiscalStatus != "needs_attention" {
		t.Errorf("integrationConnections.listOperations: %v ?%s %+v", err, call.query, ops)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"id": "op1", "fiscalStatus": "submitted"}}`)
	op, err := client.IntegrationOperations.CompleteCustomer(context.Background(), "op1", corvonfse.IntegrationCustomerParams{TaxID: "12345678901", Name: "Ana"})
	if err != nil || op.FiscalStatus != "submitted" || call.method != http.MethodPatch || call.path != "/v1/integrations/operations/op1/customer" || mustJSON(t, call.body)["taxId"] != "12345678901" {
		t.Errorf("integrationOperations.completeCustomer: %v %+v %s %s %q", err, op, call.method, call.path, call.body)
	}

	client = single(t, &call, http.StatusOK, `{"data": {"state": "needs_attention", "lastError": "boom", "operationId": "op1"}}`)
	retried, err := client.IntegrationOperations.Retry(context.Background(), "op1")
	if err != nil || retried.State != "needs_attention" || retried.LastError != "boom" || call.path != "/v1/integrations/operations/op1/retry" {
		t.Errorf("integrationOperations.retry: %v %+v %s", err, retried, call.path)
	}
}

func TestCompanyAndWebhookNewFields(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": {"id": "co_1", "status": "ready", "setupGaps": [],
		"municipalityAgreement": {"status": "active", "agreementType": "convenio", "checkedAt": "2026-11-01T00:00:00Z"},
		"warnings": [{"code": "TAX_REGIME_REGISTRY_MISMATCH", "message": "x"}], "simplesRate": "6.54"}}`)
	company, err := client.Companies.Get(context.Background(), "co_1")
	if err != nil {
		t.Fatalf("companies.get: %v", err)
	}
	if company.MunicipalityAgreement == nil || company.MunicipalityAgreement.Status != "active" || len(company.Warnings) != 1 || company.SimplesRate != "6.54" {
		t.Errorf("company fields not decoded: %+v", company)
	}

	client = single(t, &call, http.StatusOK, `{"data": [{"id": "w1", "url": "https://example.com/h", "active": true, "environment": "test",
		"events": ["nfse.invoice.taker_rejected"], "createdAt": "c", "updatedAt": "u"}], "pagination": {"limit": 50, "hasMore": false, "nextCursor": null}}`)
	hooks, _, err := client.Webhooks.List(context.Background(), nil)
	if err != nil || len(hooks) != 1 || hooks[0].Environment != corvonfse.EnvironmentTest || hooks[0].Events[0] != "nfse.invoice.taker_rejected" {
		t.Errorf("webhooks.list: %v %+v", err, hooks)
	}
}

func TestInvoiceEventPortalFields(t *testing.T) {
	var call capture
	client := single(t, &call, http.StatusOK, `{"data": [{"id": "e1", "invoiceId": "i1", "fromStatus": "authorized", "toStatus": "cancelled",
		"reason": null, "source": "portal", "portalEvent": "101101", "portalNsu": "99", "createdAt": "c"}],
		"pagination": {"limit": 50, "hasMore": false, "nextCursor": null}}`)
	events, _, err := client.Invoices.Events(context.Background(), "i1", nil)
	if err != nil || len(events) != 1 || events[0].Source != "portal" || events[0].PortalEvent != "101101" || events[0].PortalNsu != "99" {
		t.Errorf("invoices.events: %v %+v", err, events)
	}
}
