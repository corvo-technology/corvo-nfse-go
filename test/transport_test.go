package test

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

func TestTransportBuildsURLMethodQueryAndHeaders(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data":[],"pagination":{"limit":10,"hasMore":false,"nextCursor":null}}`)
	}))

	_, pagination, err := client.Companies.List(context.Background(), &corvonfse.ListParams{Limit: 10})
	if err != nil {
		t.Fatalf("companies.list: %v", err)
	}
	if call.method != http.MethodGet {
		t.Errorf("expected GET, got %s", call.method)
	}
	if call.path != "/v1/companies" {
		t.Errorf("expected path /v1/companies, got %s", call.path)
	}
	if call.query != "limit=10" {
		t.Errorf("expected query limit=10 (empty cursor skipped), got %q", call.query)
	}
	if call.header.Get("X-API-Key") != apiKey {
		t.Errorf("expected X-API-Key header, got %q", call.header.Get("X-API-Key"))
	}
	if call.header.Get("Accept") != "application/json" {
		t.Errorf("expected Accept application/json, got %q", call.header.Get("Accept"))
	}
	if pagination.Limit != 10 || pagination.HasMore {
		t.Errorf("expected limit 10 and no more pages, got %+v", pagination)
	}
}

func TestTransportSendsJSONBodyAndIdempotencyKey(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		writeJSON(w, http.StatusAccepted, `{"data":{"id":"inv_1"}}`)
	}))

	send := true
	invoice, err := client.Invoices.Create(context.Background(), corvonfse.InvoiceCreateParams{
		Customer: &corvonfse.InvoiceCustomerInput{
			Name:  "Cliente Fictício de Testes",
			TaxID: "52998224725",
		},
		Service: corvonfse.InvoiceServiceInput{
			Code:        "010101",
			Description: "Software development",
		},
		Amount:         150.5,
		CompetenceDate: "2026-10-12T00:00:00.000Z",
		Options: &corvonfse.InvoiceCreateOptions{
			Email: &corvonfse.InvoiceEmailOption{Send: &send},
		},
	}, "pedido_123")
	if err != nil {
		t.Fatalf("invoices.create: %v", err)
	}
	if invoice.ID != "inv_1" {
		t.Errorf("expected unwrapped id inv_1, got %q", invoice.ID)
	}
	if call.header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", call.header.Get("Content-Type"))
	}
	if call.header.Get("Idempotency-Key") != "pedido_123" {
		t.Errorf("expected Idempotency-Key pedido_123, got %q", call.header.Get("Idempotency-Key"))
	}
	for _, fragment := range []string{
		`"amount":150.5`,
		`"taxId":"52998224725"`,
		`"send":true`,
		`"competenceDate":"2026-10-12T00:00:00.000Z"`,
	} {
		if !strings.Contains(call.body, fragment) {
			t.Errorf("expected body to contain %s, got %s", fragment, call.body)
		}
	}
}

func TestTransportUnwrapsData(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"data":{"id":"cmp_1","legalName":"Empresa Exemplo LTDA"}}`)
	}))

	company, err := client.Companies.Get(context.Background(), "cmp_1")
	if err != nil {
		t.Fatalf("companies.get: %v", err)
	}
	if company.ID != "cmp_1" || company.LegalName != "Empresa Exemplo LTDA" {
		t.Errorf("expected unwrapped data, got %+v", company)
	}
}

func TestTransportUnwrapsDataProps(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"data":{"id":"inv_1","nfseNumber":1001,"status":"authorized"}}`)
	}))

	invoice, err := client.Invoices.Get(context.Background(), "inv_1")
	if err != nil {
		t.Fatalf("invoices.get: %v", err)
	}
	if invoice.ID != "inv_1" || invoice.NFSENumber != 1001 || invoice.Status != corvonfse.InvoiceStatusAuthorized {
		t.Errorf("expected unwrapped invoice, got %+v", invoice)
	}
}

func TestTransportNullEnvelopeYieldsZeroValue(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"data":null}`)
	}))

	company, err := client.Companies.Get(context.Background(), "cmp_1")
	if err != nil {
		t.Fatalf("expected nil error for null envelope, got %v", err)
	}
	if company.ID != "" {
		t.Errorf("expected zero value for null envelope, got %+v", company)
	}
}

func TestTransportMapsRESTError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusBadRequest, `{
			"statusCode": 400,
			"code": "ERROR_VALIDATION",
			"message": "Request validation failed",
			"details": [
				{"field": "customer.taxId", "code": "ERROR_INVALID_TAX_ID", "message": "Invalid tax ID"}
			],
			"requestId": "9f1fc995-4ee4-4a92-b285-cf8dff227a97"
		}`)
	}))

	_, err := client.Companies.Get(context.Background(), "cmp_1")
	failure := requireError(t, err)
	if failure.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", failure.Status)
	}
	if failure.Code != "ERROR_VALIDATION" {
		t.Errorf("expected code ERROR_VALIDATION, got %q", failure.Code)
	}
	if failure.Message != "Request validation failed" {
		t.Errorf("expected validation message, got %q", failure.Message)
	}
	if failure.RequestID != "9f1fc995-4ee4-4a92-b285-cf8dff227a97" {
		t.Errorf("expected requestId from body, got %q", failure.RequestID)
	}
	if len(failure.Details) != 1 || failure.Details[0].Field != "customer.taxId" {
		t.Errorf("expected one detail for customer.taxId, got %+v", failure.Details)
	}
}

func TestTransportErrorMessageArrayAndHeaderFallback(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "hdr-9")
		writeJSON(w, http.StatusUnprocessableEntity, `{"message":["campo inválido"]}`)
	}))

	_, err := client.Companies.Get(context.Background(), "cmp_1")
	failure := requireError(t, err)
	if failure.Code != "http_422" {
		t.Errorf("expected fallback code http_422, got %q", failure.Code)
	}
	if failure.Message != "campo inválido" {
		t.Errorf("expected joined array message, got %q", failure.Message)
	}
	if failure.RequestID != "hdr-9" {
		t.Errorf("expected requestId from header, got %q", failure.RequestID)
	}
}

func TestTransportRetryAfterFromHeader(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "45")
		writeJSON(w, http.StatusTooManyRequests, `{"code":"ERROR_RATE_LIMITED","message":"Rate limited"}`)
	}))

	_, _, err := client.Companies.List(context.Background(), nil)
	failure := requireError(t, err)
	if failure.Status != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", failure.Status)
	}
	if failure.RetryAfter != 45 {
		t.Errorf("expected retryAfter 45 from header, got %d", failure.RetryAfter)
	}
}

func TestTransportRetryAfterFromBodyWins(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "45")
		writeJSON(w, http.StatusTooManyRequests, `{"code":"ERROR_RATE_LIMITED","message":"Rate limited","retryAfter":30}`)
	}))

	_, _, err := client.Companies.List(context.Background(), nil)
	failure := requireError(t, err)
	if failure.RetryAfter != 30 {
		t.Errorf("expected retryAfter 30 from body, got %d", failure.RetryAfter)
	}
}

func TestTransportNoContentIsNilError(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.Companies.Remove(context.Background(), "cmp_1"); err != nil {
		t.Fatalf("expected nil error on 204, got %v", err)
	}
	if call.method != http.MethodDelete || call.path != "/v1/companies/cmp_1" {
		t.Errorf("expected DELETE /v1/companies/cmp_1, got %s %s", call.method, call.path)
	}
}

func TestTransportNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	baseURL := server.URL
	server.Close() // nothing listens on baseURL anymore

	client := corvonfse.NewClient(corvonfse.Options{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Timeout: 2 * time.Second,
	})
	_, _, err := client.Companies.List(context.Background(), nil)
	failure := requireError(t, err)
	if failure.Code != "network_error" {
		t.Errorf("expected code network_error, got %q", failure.Code)
	}
	if failure.Status != 0 {
		t.Errorf("expected status 0 for network failure, got %d", failure.Status)
	}
	if !strings.Contains(failure.Message, "Network request failed") {
		t.Errorf("expected network failure message, got %q", failure.Message)
	}
}

func TestTransportTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			writeJSON(w, http.StatusOK, `{"data":{}}`)
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(server.Close)

	client := corvonfse.NewClient(corvonfse.Options{
		APIKey:  apiKey,
		BaseURL: server.URL,
		Timeout: 50 * time.Millisecond,
	})
	_, _, err := client.Companies.List(context.Background(), nil)
	failure := requireError(t, err)
	if failure.Code != "timeout" {
		t.Errorf("expected code timeout, got %q (%s)", failure.Code, failure.Message)
	}
	if failure.Status != 0 {
		t.Errorf("expected status 0 for timeout, got %d", failure.Status)
	}
}

func TestTransportInvalidJSON(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "<html>erro</html>")
	}))

	_, err := client.Companies.Get(context.Background(), "cmp_1")
	failure := requireError(t, err)
	if failure.Code != "invalid_json" {
		t.Errorf("expected code invalid_json, got %q", failure.Code)
	}
	if failure.Status != http.StatusOK {
		t.Errorf("expected status 200, got %d", failure.Status)
	}
}

func TestCertificatesUploadSendsMultipart(t *testing.T) {
	var call capture
	var rawContentType string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawContentType = r.Header.Get("Content-Type")
		call = record(r)
		writeJSON(w, http.StatusOK, `{"data":{"id":"cert_1","serialNumber":"01:02:03"}}`)
	}))

	certificate, err := client.Certificates.Upload(context.Background(), "cmp_1",
		[]byte("PKCS12-BYTES"), "Teste123!", "certificado.p12")
	if err != nil {
		t.Fatalf("certificates.upload: %v", err)
	}
	if certificate.ID != "cert_1" {
		t.Errorf("expected unwrapped certificate, got %+v", certificate)
	}
	if call.path != "/v1/companies/cmp_1/certificates" {
		t.Errorf("expected multipart upload path, got %s", call.path)
	}
	mediaType, params, parseErr := mime.ParseMediaType(rawContentType)
	if parseErr != nil || mediaType != "multipart/form-data" {
		t.Fatalf("expected multipart/form-data content type, got %q (%v)", rawContentType, parseErr)
	}
	if params["boundary"] == "" {
		t.Fatal("expected a multipart boundary in the content type")
	}

	reader := multipart.NewReader(strings.NewReader(call.body), params["boundary"])
	filePart, partErr := reader.NextPart()
	if partErr != nil {
		t.Fatalf("expected file part: %v", partErr)
	}
	if filePart.FormName() != "file" || filePart.FileName() != "certificado.p12" {
		t.Errorf("expected file part named file/certificado.p12, got %q/%q", filePart.FormName(), filePart.FileName())
	}
	if contentType := filePart.Header.Get("Content-Type"); contentType != "application/x-pkcs12" {
		t.Errorf("expected application/x-pkcs12 part, got %q", contentType)
	}
	fileBytes, _ := io.ReadAll(filePart)
	if string(fileBytes) != "PKCS12-BYTES" {
		t.Errorf("expected PKCS12 bytes, got %q", fileBytes)
	}
	passwordPart, partErr := reader.NextPart()
	if partErr != nil {
		t.Fatalf("expected password part: %v", partErr)
	}
	if passwordPart.FormName() != "password" {
		t.Errorf("expected password field, got %q", passwordPart.FormName())
	}
	password, _ := io.ReadAll(passwordPart)
	if string(password) != "Teste123!" {
		t.Errorf("expected password Teste123!, got %q", password)
	}
}

func TestTransportDownloadFile(t *testing.T) {
	var call capture
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = record(r)
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="nfse-1001.pdf"`)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "%PDF-1.7 fake")
	}))

	document, err := client.Invoices.PDF(context.Background(), "inv_1")
	if err != nil {
		t.Fatalf("invoices.pdf: %v", err)
	}
	if call.path != "/v1/invoices/inv_1/pdf" {
		t.Errorf("expected path /v1/invoices/inv_1/pdf, got %s", call.path)
	}
	if string(document.Buffer) != "%PDF-1.7 fake" {
		t.Errorf("expected raw PDF bytes, got %q", document.Buffer)
	}
	if document.MimeType != "application/pdf" {
		t.Errorf("expected mime type application/pdf, got %q", document.MimeType)
	}
	if document.FileName != "nfse-1001.pdf" {
		t.Errorf("expected file name nfse-1001.pdf, got %q", document.FileName)
	}
}

func TestTransportNormalizesBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/companies" {
			t.Errorf("expected path /v1/companies with trailing-slash base, got %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, `{"data":[],"pagination":{"limit":50,"hasMore":false,"nextCursor":null}}`)
	}))
	t.Cleanup(server.Close)

	client := corvonfse.NewClient(corvonfse.Options{
		APIKey:  apiKey,
		BaseURL: server.URL + "/",
	})
	if _, _, err := client.Companies.List(context.Background(), nil); err != nil {
		t.Fatalf("companies.list with trailing slash base: %v", err)
	}
}
