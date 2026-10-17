package corvonfse

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"regexp"
	"strings"
)

// signatureRE matches the X-Corvo-Signature header format: sha256=<hex>.
var signatureRE = regexp.MustCompile(`^sha256=([0-9a-f]{64})$`)

// VerifySignature reports whether signatureHeader is a valid HMAC-SHA256
// signature of rawBody under secret (the secret returned at webhook
// creation).
//
// The header is the value of X-Corvo-Signature: "sha256=<hex>". The
// comparison is constant-time (crypto/subtle), so verification itself does
// not leak timing information. Malformed headers, an empty secret or a
// mismatch all return false — VerifySignature never panics.
//
//	signed := r.Header.Get("X-Corvo-Signature")
//	if !corvonfse.VerifySignature(body, signed, secret) {
//		// reject
//	}
func VerifySignature(rawBody []byte, signatureHeader string, secret string) bool {
	if secret == "" || signatureHeader == "" {
		return false
	}
	match := signatureRE.FindStringSubmatch(strings.TrimSpace(signatureHeader))
	if match == nil {
		return false
	}
	received, err := hex.DecodeString(match[1])
	if err != nil || len(received) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(rawBody) // hash.Hash.Write never fails
	expected := mac.Sum(nil)
	return subtle.ConstantTimeCompare(received, expected) == 1
}
