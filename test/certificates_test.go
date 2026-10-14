package test

import (
	"context"
	"net/http"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

const certificateJSON = `{
	"id": "cert_1",
	"companyId": "cmp_1",
	"serialNumber": "01:02:03:04",
	"issuer": "AC Certificadora",
	"subject": "Empresa Exemplo LTDA",
	"fingerprint": "AB:CD",
	"validFrom": "2026-01-01",
	"validUntil": "2027-01-01",
	"status": "active",
	"createdAt": "2026-10-12T13:04:38.985Z"
}`

func TestCertificatesList(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{
			"data": [`+certificateJSON+`],
			"pagination": {"limit": 10, "hasMore": false, "nextCursor": null}
		}`)
	}))

	items, pagination, err := client.Certificates.List(context.Background(), "cmp_1", &corvonfse.ListParams{Limit: 10})
	if err != nil {
		t.Fatalf("certificates.list: %v", err)
	}
	if call.method != http.MethodGet || call.path != "/v1/companies/cmp_1/certificates" {
		t.Errorf("expected GET /v1/companies/cmp_1/certificates, got %s %s", call.method, call.path)
	}
	if call.query != "limit=10" {
		t.Errorf("expected limit=10, got %q", call.query)
	}
	if len(items) != 1 || items[0].ID != "cert_1" {
		t.Errorf("expected one certificate, got %+v", items)
	}
	if pagination.HasMore {
		t.Errorf("expected no more pages, got %+v", pagination)
	}
}

func TestCertificatesRemove(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.Certificates.Remove(context.Background(), "cmp_1", "cert_1"); err != nil {
		t.Fatalf("certificates.remove: %v", err)
	}
	if call.method != http.MethodDelete || call.path != "/v1/companies/cmp_1/certificates/cert_1" {
		t.Errorf("expected DELETE certificate path, got %s %s", call.method, call.path)
	}
}
