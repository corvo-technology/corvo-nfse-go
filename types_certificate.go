package corvonfse

// Certificate is an A1 (PKCS#12) digital certificate registered on a
// company, used to sign fiscal documents.
type Certificate struct {
	ID           string `json:"id"`
	CompanyID    string `json:"companyId"`
	SerialNumber string `json:"serialNumber"`
	Issuer       string `json:"issuer"`
	Subject      string `json:"subject"`
	Fingerprint  string `json:"fingerprint"`
	ValidFrom    string `json:"validFrom"`
	ValidUntil   string `json:"validUntil"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
}
