package test

import (
	"strings"
	"testing"

	corvonfse "github.com/corvo-technology/corvo-nfse-go"
)

// The known HMAC-SHA256 vector shared by every Corvo SDK.
const (
	signedBody   = `{"type":"nfse.invoice.authorized","data":{"id":"inv_1"}}`
	signedSecret = "whsec_test_123"
	signedHeader = "sha256=8f9bd040d8f717b19618ab3395ddfa4b8c176b8dc863195227b88dd85426bbbc"
)

func TestVerifySignatureAcceptsKnownVector(t *testing.T) {
	if !corvonfse.VerifySignature([]byte(signedBody), signedHeader, signedSecret) {
		t.Error("expected the known HMAC-SHA256 vector to verify")
	}
}

func TestVerifySignatureRejectsTampering(t *testing.T) {
	tampered := strings.Replace(signedBody, "inv_1", "inv_2", 1)
	if corvonfse.VerifySignature([]byte(tampered), signedHeader, signedSecret) {
		t.Error("expected a tampered body to be rejected")
	}
	if corvonfse.VerifySignature([]byte(signedBody), signedHeader, "whsec_other") {
		t.Error("expected a wrong secret to be rejected")
	}
	// Flip one hex character of the signature.
	broken := signedHeader[:len(signedHeader)-1] + "0"
	if broken == signedHeader {
		broken = signedHeader[:len(signedHeader)-1] + "1"
	}
	if corvonfse.VerifySignature([]byte(signedBody), broken, signedSecret) {
		t.Error("expected a corrupted signature to be rejected")
	}
}

func TestVerifySignatureRejectsMalformedInput(t *testing.T) {
	for name, header := range map[string]string{
		"missing prefix":     strings.TrimPrefix(signedHeader, "sha256="),
		"bad hex":            "sha256=zzzz",
		"empty header":       "",
		"uppercase hex":      "sha256=" + strings.ToUpper(signedHeader[len("sha256="):]),
		"wrong length":       "sha256=abc",
		"other algorithm":    "sha512=8f9bd040d8f717b19618ab3395ddfa4b8c176b8dc863195227b88dd85426bbbc",
		"trailing garbage":   signedHeader + "!",
	} {
		if corvonfse.VerifySignature([]byte(signedBody), header, signedSecret) {
			t.Errorf("%s: expected rejection of header %q", name, header)
		}
	}
	if corvonfse.VerifySignature([]byte(signedBody), signedHeader, "") {
		t.Error("expected an empty secret to be rejected")
	}
	if corvonfse.VerifySignature(nil, signedHeader, signedSecret) {
		t.Error("expected a nil body to be rejected against a non-empty signature")
	}
}

func TestVerifySignatureTrimsHeaderWhitespace(t *testing.T) {
	if !corvonfse.VerifySignature([]byte(signedBody), "  "+signedHeader+"\t", signedSecret) {
		t.Error("expected surrounding whitespace to be trimmed before matching")
	}
}
