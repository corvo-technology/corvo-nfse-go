package test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

// callers maps every operationId of the contract (spec/openapi.json) to the SDK
// call that implements it. TestContractCoverage fails when the contract has an
// operation that is not here (a new API route), when an entry is not in the
// contract, and when a call does not hit the method and path of its operation.
//
// To update the contract: replace spec/openapi.json (see spec/README.md), run
// the test and add the new operations below.
const (
	sid  = "id-1"
	line = 7
)

var callers = map[string]func(ctx context.Context, c *corvonfse.Client){
	"accountantPackage_get": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.AccountantPackage.Get(ctx, sid)
	},
	"accountantPackage_listSends": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.AccountantPackage.ListSends(ctx, sid)
	},
	"accountantPackage_send": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.AccountantPackage.Send(ctx, sid, "2026-10")
	},
	"accountantPackage_update": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.AccountantPackage.Update(ctx, sid, corvonfse.AccountantPackageUpdateParams{})
	},
	"billing_getPlan": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Billing.Plan(ctx)
	},
	"billing_getUsage": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Billing.Usage(ctx)
	},
	"certificates_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Certificates.List(ctx, sid, nil)
	},
	"certificates_revoke": func(ctx context.Context, c *corvonfse.Client) {
		_ = c.Certificates.Remove(ctx, sid, sid)
	},
	"certificates_upload": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Certificates.Upload(ctx, sid, []byte("p12"), "pw", "")
	},
	"cnpj_lookup": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.CNPJ.Lookup(ctx, sid)
	},
	"companies_create": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Companies.Create(ctx, corvonfse.CompanyCreateParams{})
	},
	"companies_delete": func(ctx context.Context, c *corvonfse.Client) {
		_ = c.Companies.Remove(ctx, sid)
	},
	"companies_get": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Companies.Get(ctx, sid)
	},
	"companies_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Companies.List(ctx, nil)
	},
	"companies_update": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Companies.Update(ctx, sid, corvonfse.CompanyUpdateParams{})
	},
	"credits_getBalance": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Credits.Balance(ctx)
	},
	"credits_listHistory": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Credits.History(ctx, nil)
	},
	"credits_listPlans": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Credits.Plans(ctx, nil)
	},
	"credits_previewPurchase": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Credits.PreviewPurchase(ctx, 3)
	},
	"customers_delete": func(ctx context.Context, c *corvonfse.Client) {
		_ = c.Customers.Remove(ctx, sid)
	},
	"customers_get": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Customers.Get(ctx, sid)
	},
	"customers_search": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Customers.Search(ctx, nil)
	},
	"customers_update": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Customers.Update(ctx, sid, corvonfse.CustomerUpdateParams{})
	},
	"dpsSequence_get": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.DPSSequence.Get(ctx, sid)
	},
	"dpsSequence_set": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.DPSSequence.Set(ctx, sid, corvonfse.DPSSequenceSetParams{})
	},
	"fiscalPortal_status": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.FiscalPortal.Status(ctx)
	},
	"fiscalProfiles_approve": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.FiscalProfiles.Approve(ctx, sid, sid)
	},
	"fiscalProfiles_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.FiscalProfiles.List(ctx, sid)
	},
	"fiscalProfiles_upsert": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.FiscalProfiles.Upsert(ctx, sid, corvonfse.FiscalProfileUpsertParams{TaxRegime: "MEI"})
	},
	"health_live": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Health.Live(ctx)
	},
	"imports_discard": func(ctx context.Context, c *corvonfse.Client) {
		_ = c.Imports.Discard(ctx, sid)
	},
	"imports_downloadErrors": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Imports.DownloadErrors(ctx, sid)
	},
	"imports_downloadTemplate": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Imports.DownloadTemplate(ctx)
	},
	"imports_emit": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Imports.Emit(ctx, sid, nil)
	},
	"imports_get": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Imports.Get(ctx, sid)
	},
	"imports_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Imports.List(ctx, 0)
	},
	"imports_listRows": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Imports.ListRows(ctx, sid, nil)
	},
	"imports_start": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Imports.Start(ctx, []byte("a;b"), "", "")
	},
	"imports_updateRow": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Imports.UpdateRow(ctx, sid, line, nil)
	},
	"integrationConnections_activate": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.Activate(ctx, sid)
	},
	"integrationConnections_create": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.Create(ctx, corvonfse.IntegrationConnectionCreateParams{})
	},
	"integrationConnections_disconnect": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.Disconnect(ctx, sid)
	},
	"integrationConnections_get": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.Get(ctx, sid)
	},
	"integrationConnections_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.List(ctx)
	},
	"integrationConnections_listEvents": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.ListEvents(ctx, sid, 0)
	},
	"integrationConnections_listOperations": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.ListOperations(ctx, sid, 0)
	},
	"integrationConnections_listRules": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.ListRules(ctx, sid)
	},
	"integrationConnections_pause": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.Pause(ctx, sid)
	},
	"integrationConnections_replaceRules": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.ReplaceRules(ctx, sid, nil)
	},
	"integrationConnections_updateCredentials": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.UpdateCredentials(ctx, sid, nil)
	},
	"integrationConnections_validate": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationConnections.Validate(ctx, sid)
	},
	"integrationOperations_completeCustomer": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationOperations.CompleteCustomer(ctx, sid, corvonfse.IntegrationCustomerParams{})
	},
	"integrationOperations_retry": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.IntegrationOperations.Retry(ctx, sid)
	},
	"integrations_listProviders": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Integrations.ListProviders(ctx)
	},
	"invoicePackage_download": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.InvoicePackage.Download(ctx, sid, "2026-10", nil)
	},
	"invoicePackage_summary": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.InvoicePackage.Summary(ctx, sid, "2026-10")
	},
	"invoices_calculate": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.Calculate(ctx, corvonfse.InvoiceCreateParams{})
	},
	"invoices_cancel": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.Cancel(ctx, sid, corvonfse.InvoiceCancelParams{})
	},
	"invoices_create": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.Create(ctx, corvonfse.InvoiceCreateParams{}, "")
	},
	"invoices_createBatch": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.CreateBatch(ctx, nil)
	},
	"invoices_createDownloadUrl": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.DownloadURL(ctx, sid, corvonfse.FormatPDF)
	},
	"invoices_downloadPdf": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.PDF(ctx, sid)
	},
	"invoices_downloadXml": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.XML(ctx, sid)
	},
	"invoices_get": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.Get(ctx, sid)
	},
	"invoices_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Invoices.List(ctx, nil)
	},
	"invoices_listEvents": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Invoices.Events(ctx, sid, nil)
	},
	"invoices_overview": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.Overview(ctx, "")
	},
	"invoices_review": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.Review(ctx, corvonfse.InvoiceReviewParams{})
	},
	"invoices_sendEmail": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.SendEmail(ctx, sid, nil, "")
	},
	"invoices_substitute": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Invoices.Substitute(ctx, sid, corvonfse.InvoiceSubstituteParams{}, "")
	},
	"schedules_cancel": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Schedules.Cancel(ctx, sid)
	},
	"schedules_get": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Schedules.Get(ctx, sid)
	},
	"schedules_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Schedules.List(ctx, nil)
	},
	"schedules_listRuns": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Schedules.ListRuns(ctx, sid, nil)
	},
	"schedules_pause": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Schedules.Pause(ctx, sid)
	},
	"schedules_resume": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Schedules.Resume(ctx, sid)
	},
	"schedules_update": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Schedules.Update(ctx, sid, corvonfse.ScheduleUpdateParams{})
	},
	"serviceFavorites_create": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.ServiceFavorites.Create(ctx, sid, corvonfse.ServiceFavoriteCreateParams{})
	},
	"serviceFavorites_delete": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.ServiceFavorites.Remove(ctx, sid, sid)
	},
	"serviceFavorites_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.ServiceFavorites.List(ctx, sid)
	},
	"serviceTaxProfiles_approve": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.ServiceTaxProfiles.Approve(ctx, sid, sid)
	},
	"serviceTaxProfiles_create": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.ServiceTaxProfiles.Create(ctx, sid, corvonfse.ServiceTaxProfileCreateParams{})
	},
	"serviceTaxProfiles_delete": func(ctx context.Context, c *corvonfse.Client) {
		_ = c.ServiceTaxProfiles.Remove(ctx, sid, sid)
	},
	"serviceTaxProfiles_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.ServiceTaxProfiles.List(ctx, sid, "")
	},
	"serviceTaxProfiles_update": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.ServiceTaxProfiles.Update(ctx, sid, sid, corvonfse.ServiceTaxProfileUpdateParams{})
	},
	"webhookDeliveries_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Webhooks.Deliveries.List(ctx, nil)
	},
	"webhookDeliveries_retry": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Webhooks.Deliveries.Retry(ctx, sid)
	},
	"webhookDeliveries_stats": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Webhooks.Deliveries.Stats(ctx, 0, "")
	},
	"webhooks_create": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Webhooks.Create(ctx, corvonfse.WebhookCreateParams{})
	},
	"webhooks_delete": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Webhooks.Remove(ctx, sid)
	},
	"webhooks_list": func(ctx context.Context, c *corvonfse.Client) {
		_, _, _ = c.Webhooks.List(ctx, nil)
	},
	"webhooks_test": func(ctx context.Context, c *corvonfse.Client) {
		_, _ = c.Webhooks.Test(ctx, sid)
	},
}

type specOperation struct {
	id       string
	method   string
	path     string
	required []string // required query parameters
}

func loadSpec(t *testing.T) []specOperation {
	t.Helper()
	raw, err := os.ReadFile("../spec/openapi.json")
	if err != nil {
		t.Fatalf("read spec/openapi.json: %v", err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
			Parameters  []struct {
				Name     string `json:"name"`
				In       string `json:"in"`
				Required bool   `json:"required"`
				Ref      string `json:"$ref"`
			} `json:"parameters"`
		} `json:"paths"`
		Components struct {
			Parameters map[string]struct {
				Name     string `json:"name"`
				In       string `json:"in"`
				Required bool   `json:"required"`
			} `json:"parameters"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse spec/openapi.json: %v", err)
	}
	var ops []specOperation
	for path, methods := range doc.Paths {
		for method, op := range methods {
			if op.OperationID == "" {
				continue
			}
			so := specOperation{id: op.OperationID, method: strings.ToUpper(method), path: path}
			for _, p := range op.Parameters {
				name, in, required := p.Name, p.In, p.Required
				if p.Ref != "" {
					ref := doc.Components.Parameters[p.Ref[strings.LastIndex(p.Ref, "/")+1:]]
					name, in, required = ref.Name, ref.In, ref.Required
				}
				if in == "query" && required {
					so.required = append(so.required, name)
				}
			}
			ops = append(ops, so)
		}
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].id < ops[j].id })
	return ops
}

var pathParam = regexp.MustCompile(`\{([^}]+)\}`)

// expectedPath turns a path template into the path the SDK must call (with the
// /v1 prefix), using sid for every parameter and line for the row line.
func expectedPath(template string) string {
	return "/v1" + pathParam.ReplaceAllStringFunc(template, func(m string) string {
		if m == "{line}" {
			return "7"
		}
		return sid
	})
}

func TestContractCoverage(t *testing.T) {
	ops := loadSpec(t)
	if len(ops) != 92 {
		t.Fatalf("expected 92 operations in spec/openapi.json (contract 0.3.0), got %d: update the SDK and this count together", len(ops))
	}

	inSpec := map[string]bool{}
	for _, op := range ops {
		inSpec[op.id] = true
	}
	for id := range callers {
		if !inSpec[id] {
			t.Errorf("callers has %q, which is not in the contract: remove it", id)
		}
	}

	for _, op := range ops {
		op := op
		t.Run(op.id, func(t *testing.T) {
			call, ok := callers[op.id]
			if !ok {
				t.Fatalf("operation %s (%s %s) is not implemented by the SDK: add it to the SDK and to callers", op.id, op.method, op.path)
			}
			var got []capture
			client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, record(r))
				writeJSON(w, http.StatusOK, `{"data": null}`)
			}))
			call(context.Background(), client)
			if len(got) != 1 {
				t.Fatalf("expected exactly one request, got %d", len(got))
			}
			if got[0].method != op.method || got[0].path != expectedPath(op.path) {
				t.Errorf("expected %s %s, got %s %s", op.method, expectedPath(op.path), got[0].method, got[0].path)
			}
			if got[0].header.Get("X-API-Key") != apiKey {
				t.Errorf("expected the x-api-key header on %s", op.id)
			}
			for _, name := range op.required {
				if !strings.Contains("&"+got[0].query, "&"+name+"=") {
					t.Errorf("expected required query parameter %q, got %q", name, got[0].query)
				}
			}
		})
	}
}
