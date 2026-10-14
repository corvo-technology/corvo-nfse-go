package test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

const companyJSON = `{
	"id": "09947127-2702-4f6d-9da8-cb463608d812",
	"tenantId": "6247dfd9-0adc-4bf6-aac9-ee9dc1e7ca8b",
	"legalName": "Empresa Exemplo Tecnologia LTDA",
	"tradeName": "Empresa Exemplo TESTE",
	"taxId": "41834687000148",
	"municipalReg": "123456",
	"municipalityCode": "3170206",
	"taxRegime": "Simples Nacional",
	"simpleNational": true,
	"specialRegime": null,
	"addressStreet": "Rua Exemplo de Testes",
	"addressNumber": "100",
	"addressComplement": null,
	"addressDistrict": "Centro",
	"addressCity": "Uberlândia",
	"addressState": "MG",
	"addressZip": "38400000",
	"fiscalSchemaVersion": "v2_02",
	"environment": "test",
	"status": "ready",
	"createdAt": "2026-10-12T13:04:38.985Z",
	"updatedAt": "2026-10-12T13:04:40.230Z",
	"deletedAt": null,
	"setupGaps": []
}`

func TestCompaniesCreatePassesCamelCaseBodyThrough(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusCreated, `{"data":`+companyJSON+`}`)
	}))

	company, err := client.Companies.Create(context.Background(), corvonfse.CompanyCreateParams{
		LegalName:    "Empresa Exemplo Tecnologia LTDA",
		TaxID:        "41.834.687/0001-48",
		TaxRegime:    corvonfse.TaxRegimeSimplesNacional,
		MunicipalReg: "123456",
		Address: &corvonfse.CompanyAddressInput{
			Street:   "Rua Exemplo",
			Number:   "100",
			District: "Centro",
			City:     "Uberlândia",
			State:    "MG",
			Zip:      "38400-000",
		},
	})
	if err != nil {
		t.Fatalf("companies.create: %v", err)
	}
	if call.method != http.MethodPost || call.path != "/v1/companies" {
		t.Errorf("expected POST /v1/companies, got %s %s", call.method, call.path)
	}
	for _, fragment := range []string{
		`"legalName":"Empresa Exemplo Tecnologia LTDA"`,
		`"taxId":"41.834.687/0001-48"`,
		`"taxRegime":"Simples Nacional"`,
		`"municipalReg":"123456"`,
		`"city":"Uberlândia"`,
	} {
		if !strings.Contains(call.body, fragment) {
			t.Errorf("expected body to contain %s, got %s", fragment, call.body)
		}
	}
	if company.ID != "09947127-2702-4f6d-9da8-cb463608d812" {
		t.Errorf("expected unwrapped company, got %+v", company)
	}
	if company.TaxRegime != corvonfse.TaxRegimeSimplesNacional {
		t.Errorf("expected tax regime from wire, got %q", company.TaxRegime)
	}
	if !company.SimpleNational {
		t.Error("expected simpleNational true")
	}
	if company.DeletedAt != "" || company.SpecialRegime != "" {
		t.Errorf("expected null wire fields to decode to zero value, got deletedAt %q specialRegime %q",
			company.DeletedAt, company.SpecialRegime)
	}
}

func TestCompaniesListReturnsPageAndPagination(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{
			"data": [`+companyJSON+`],
			"pagination": {"limit": 1, "hasMore": true, "nextCursor": "abc123"}
		}`)
	}))

	items, pagination, err := client.Companies.List(context.Background(), &corvonfse.ListParams{Limit: 1})
	if err != nil {
		t.Fatalf("companies.list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].AddressCity != "Uberlândia" {
		t.Errorf("expected company fields decoded, got %+v", items[0])
	}
	if pagination.Limit != 1 || !pagination.HasMore || pagination.NextCursor != "abc123" {
		t.Errorf("expected pagination {1 true abc123}, got %+v", pagination)
	}
}

func TestCompaniesUpdatePatchesGivenFields(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data":`+companyJSON+`}`)
	}))

	_, err := client.Companies.Update(context.Background(), "cmp_1", corvonfse.CompanyUpdateParams{
		MunicipalReg: "654321",
		TaxRegime:    corvonfse.TaxRegimeLucroPresumido,
	})
	if err != nil {
		t.Fatalf("companies.update: %v", err)
	}
	if call.method != http.MethodPatch || call.path != "/v1/companies/cmp_1" {
		t.Errorf("expected PATCH /v1/companies/cmp_1, got %s %s", call.method, call.path)
	}
	if !strings.Contains(call.body, `"municipalReg":"654321"`) {
		t.Errorf("expected municipalReg in body, got %s", call.body)
	}
	if strings.Contains(call.body, `"legalName"`) {
		t.Errorf("expected unset fields to be omitted, got %s", call.body)
	}
}

func TestCompaniesIterateWalksCursorPages(t *testing.T) {
	var calls []capture
	page := 0
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, record(r))
		page++
		switch page {
		case 1:
			writeJSON(w, http.StatusOK, `{
				"data": [{"id": "cmp_1"}, {"id": "cmp_2"}],
				"pagination": {"limit": 2, "hasMore": true, "nextCursor": "cursor-1"}
			}`)
		case 2:
			writeJSON(w, http.StatusOK, `{
				"data": [{"id": "cmp_3"}],
				"pagination": {"limit": 2, "hasMore": false, "nextCursor": null}
			}`)
		default:
			t.Errorf("unexpected extra page request: %s?%s", r.URL.Path, r.URL.RawQuery)
			writeJSON(w, http.StatusOK, `{"data":[],"pagination":{"limit":2,"hasMore":false,"nextCursor":null}}`)
		}
	}))

	var seen []string
	err := client.Companies.Iterate(context.Background(), &corvonfse.ListParams{Limit: 2}, func(company corvonfse.Company) bool {
		seen = append(seen, company.ID)
		return true
	})
	if err != nil {
		t.Fatalf("companies.iterate: %v", err)
	}
	if strings.Join(seen, ",") != "cmp_1,cmp_2,cmp_3" {
		t.Errorf("expected all items across pages, got %v", seen)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(calls))
	}
	if calls[0].query != "limit=2" {
		t.Errorf("expected first page query limit=2, got %q", calls[0].query)
	}
	if calls[1].query != "cursor=cursor-1&limit=2" {
		t.Errorf("expected second page to continue cursor, got %q", calls[1].query)
	}
}

func TestCompaniesIterateStopsEarly(t *testing.T) {
	var calls []capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, record(r))
		writeJSON(w, http.StatusOK, `{
			"data": [{"id": "cmp_1"}, {"id": "cmp_2"}],
			"pagination": {"limit": 2, "hasMore": true, "nextCursor": "cursor-1"}
		}`)
	}))

	err := client.Companies.Iterate(context.Background(), nil, func(company corvonfse.Company) bool {
		return company.ID != "cmp_1" // stop on the first item
	})
	if err != nil {
		t.Fatalf("expected nil error when stopping early, got %v", err)
	}
	if len(calls) != 1 {
		t.Errorf("expected a single request when stopping early, got %d", len(calls))
	}
}

func TestCompaniesIterateFirstPageFailureIsCorvoError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"code":"ERROR_INTERNAL","message":"boom"}`)
	}))

	var seen int
	err := client.Companies.Iterate(context.Background(), nil, func(company corvonfse.Company) bool {
		seen++
		return true
	})
	failure := requireError(t, err)
	if failure.Code != "ERROR_INTERNAL" || failure.Status != http.StatusInternalServerError {
		t.Errorf("expected first-page failure surfaced as CorvoError, got %+v", failure)
	}
	if seen != 0 {
		t.Errorf("expected no items before the failure, got %d", seen)
	}
}
