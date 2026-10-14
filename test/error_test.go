package test

import (
	"encoding/json"
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

func TestCorvoErrorMessageFormat(t *testing.T) {
	failure := &corvonfse.CorvoError{
		Status:  400,
		Code:    "ERROR_VALIDATION",
		Message: "Request validation failed",
	}
	if failure.Error() != "ERROR_VALIDATION: Request validation failed" {
		t.Errorf("expected '<code>: <message>', got %q", failure.Error())
	}
}

func TestCorvoErrorParsesWireFixture(t *testing.T) {
	const fixture = `{
		"statusCode": 400,
		"code": "ERROR_VALIDATION",
		"message": "Request validation failed",
		"details": [
			{"field": "customer.taxId", "code": "ERROR_INVALID_TAX_ID", "message": "Invalid tax ID"}
		],
		"requestId": "9f1fc995-4ee4-4a92-b285-cf8dff227a97",
		"retryAfter": 30
	}`

	var failure corvonfse.CorvoError
	if err := json.Unmarshal([]byte(fixture), &failure); err != nil {
		t.Fatalf("expected fixture to parse via json tags: %v", err)
	}
	if failure.Status != 400 || failure.Code != "ERROR_VALIDATION" {
		t.Errorf("expected status/code from fixture, got %+v", failure)
	}
	if !strings.HasPrefix(failure.Error(), "ERROR_VALIDATION: ") {
		t.Errorf("expected '<code>: <message>' format, got %q", failure.Error())
	}
	if len(failure.Details) != 1 {
		t.Fatalf("expected one detail, got %+v", failure.Details)
	}
	detail := failure.Details[0]
	if detail.Field != "customer.taxId" || detail.Code != "ERROR_INVALID_TAX_ID" {
		t.Errorf("expected detail fields decoded, got %+v", detail)
	}
	if failure.RequestID != "9f1fc995-4ee4-4a92-b285-cf8dff227a97" {
		t.Errorf("expected requestId decoded, got %q", failure.RequestID)
	}
	if failure.RetryAfter != 30 {
		t.Errorf("expected retryAfter 30, got %d", failure.RetryAfter)
	}
}
