package test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

func TestInvoicesCalculateReturnsTheEstimate(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data":{
			"preEmission": true, "status": "ready",
			"taxes": {
				"iss": {"base": "15000.00", "rate": "3", "amount": "450.00"},
				"ibsTotal": {"base": "14550.00", "amount": "14.55"}
			},
			"netReceivableAmount": "15000.00", "blocking": null
		}}`)
	}))

	result, err := client.Invoices.Calculate(context.Background(), corvonfse.InvoiceCreateParams{
		CompanyTaxID:   "11222333000181",
		Amount:         15000,
		CompetenceDate: "2026-10-16",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/invoices/calculate" {
		t.Errorf("unexpected request %s %s", call.method, call.path)
	}
	if result.Status != "ready" || result.Taxes.ISS == nil || result.Taxes.ISS.Amount != "450.00" {
		t.Errorf("unexpected estimate: %+v", result)
	}
	if result.Taxes.IBSTotal == nil || result.Taxes.IBSTotal.Rate != "" {
		t.Errorf("ibsTotal must have no rate: %+v", result.Taxes.IBSTotal)
	}
	if strings.Contains(call.body, `"customer"`) {
		t.Errorf("a nil customer must be omitted, body: %s", call.body)
	}
}

func TestInvoicesCalculateBlockedIsDataNotError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"data":{"status":"blocked",
			"blocking":{"code":"TAX_RATE_NOT_FOUND","message":"x","step":"LegacyTaxCalculator"},
			"pendingFields":[{"field":"iss.rate","reason":"missing"}]}}`)
	}))
	result, err := client.Invoices.Calculate(context.Background(), corvonfse.InvoiceCreateParams{Amount: 10})
	if err != nil {
		t.Fatalf("a blocked estimate must not be an error: %v", err)
	}
	if result.Blocking == nil || result.Blocking.Code != "TAX_RATE_NOT_FOUND" || len(result.PendingFields) != 1 {
		t.Errorf("unexpected blocked estimate: %+v", result)
	}
}

func TestPlansCatalog(t *testing.T) {
	var paths []string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		switch {
		case strings.HasPrefix(r.URL.Path, "/v1/plans/quote"):
			writeJSON(w, http.StatusOK, `{"data":{"volume":20000,"recommended_plan":"volume","total_estimated_cents":179990}}`)
		case r.URL.Path == "/v1/plans/growth":
			writeJSON(w, http.StatusOK, `{"data":{"slug":"growth","monthlyAllowance":750,"monthlyTotalCents":8990}}`)
		default:
			writeJSON(w, http.StatusOK, `{"data":[{"slug":"developer","monthlyAllowance":200}],"pagination":{"limit":5,"hasMore":false,"nextCursor":null}}`)
		}
	}))
	ctx := context.Background()

	items, page, err := client.Plans.List(ctx, &corvonfse.ListParams{Limit: 5})
	if err != nil || len(items) != 1 || items[0].Slug != "developer" || page.HasMore {
		t.Fatalf("list: %+v %+v %v", items, page, err)
	}
	plan, err := client.Plans.Get(ctx, "growth")
	if err != nil || plan.MonthlyTotalCents != 8990 {
		t.Fatalf("get: %+v %v", plan, err)
	}
	quote, err := client.Plans.Quote(ctx, 20000)
	if err != nil || quote == nil || quote.RecommendedPlan != "volume" || quote.TotalEstimatedCents != 179990 {
		t.Fatalf("quote: %+v %v", quote, err)
	}
	want := []string{"/v1/plans?limit=5", "/v1/plans/growth", "/v1/plans/quote?volume=20000"}
	for i, p := range want {
		if paths[i] != p {
			t.Errorf("request %d: got %s want %s", i, paths[i], p)
		}
	}
}

func TestPlansQuoteNullWhenVolumeIsInvalid(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"data":null}`)
	}))
	quote, err := client.Plans.Quote(context.Background(), 0)
	if err != nil || quote != nil {
		t.Errorf("expected (nil, nil), got %+v %v", quote, err)
	}
}
