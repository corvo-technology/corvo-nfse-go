package corvonfse

// TaxRegime is one of the Brazilian corporate tax regimes accepted when
// creating a company.
type TaxRegime string

const (
	TaxRegimeMEI             TaxRegime = "MEI"
	TaxRegimeSimplesNacional TaxRegime = "Simples Nacional"
	TaxRegimeLucroPresumido  TaxRegime = "Lucro Presumido"
	TaxRegimeLucroReal       TaxRegime = "Lucro Real"
)

// CompanyAddressInput is the address supplied when creating or updating a
// company. All fields are optional. State is the two-letter UF and Zip the CEP
// (8 digits, with or without punctuation).
type CompanyAddressInput struct {
	Street     string `json:"street,omitempty"`
	Number     string `json:"number,omitempty"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	Zip        string `json:"zip,omitempty"`
}

// CompanyCreateParams are the fields accepted by Companies.Create.
type CompanyCreateParams struct {
	LegalName string `json:"legalName"`
	// TaxID is the CNPJ.
	TaxID     string    `json:"taxId"`
	TaxRegime TaxRegime `json:"taxRegime"`
	// MunicipalityCode is the IBGE code of the company's city (7 digits).
	MunicipalityCode string `json:"municipalityCode"`
	TradeName        string `json:"tradeName,omitempty"`
	// MunicipalReg is stored, but only sent on the invoice when
	// SendMunicipalReg is true.
	MunicipalReg string `json:"municipalReg,omitempty"`
	// SendMunicipalReg defaults to false: turn it on only if the city
	// registered the company in the national registry.
	SendMunicipalReg *bool `json:"sendMunicipalReg,omitempty"`
	// DefaultServiceCode is the national service code (6 digits) used when an
	// invoice omits service.code.
	DefaultServiceCode string               `json:"defaultServiceCode,omitempty"`
	SimpleNational     *bool                `json:"simpleNational,omitempty"`
	SpecialRegime      string               `json:"specialRegime,omitempty"`
	Address            *CompanyAddressInput `json:"address,omitempty"`
}

// CompanyUpdateParams are the fields accepted by Companies.Update. Fields
// left at their zero value are omitted from the request, leaving the
// server-side value unchanged.
type CompanyUpdateParams struct {
	LegalName          string               `json:"legalName,omitempty"`
	TradeName          string               `json:"tradeName,omitempty"`
	MunicipalReg       string               `json:"municipalReg,omitempty"`
	SendMunicipalReg   *bool                `json:"sendMunicipalReg,omitempty"`
	MunicipalityCode   string               `json:"municipalityCode,omitempty"`
	DefaultServiceCode string               `json:"defaultServiceCode,omitempty"`
	TaxRegime          TaxRegime            `json:"taxRegime,omitempty"`
	SimpleNational     *bool                `json:"simpleNational,omitempty"`
	SpecialRegime      string               `json:"specialRegime,omitempty"`
	Address            *CompanyAddressInput `json:"address,omitempty"`
}

// Company is a fiscal company (emitente) registered with the API. Optional
// fields are absent from the JSON when unset and map to their zero value.
// SetupGaps lists what is still pending (municipality_code, certificate).
type Company struct {
	ID                  string    `json:"id"`
	TenantID            string    `json:"tenantId"`
	LegalName           string    `json:"legalName"`
	TradeName           string    `json:"tradeName"`
	TaxID               string    `json:"taxId"`
	MunicipalReg        string    `json:"municipalReg"`
	SendMunicipalReg    bool      `json:"sendMunicipalReg"`
	DefaultServiceCode  string    `json:"defaultServiceCode"`
	MunicipalityCode    string    `json:"municipalityCode"`
	TaxRegime           TaxRegime `json:"taxRegime"`
	SimpleNational      bool      `json:"simpleNational"`
	SpecialRegime       string    `json:"specialRegime"`
	AddressStreet       string    `json:"addressStreet"`
	AddressNumber       string    `json:"addressNumber"`
	AddressComplement   string    `json:"addressComplement"`
	AddressDistrict     string    `json:"addressDistrict"`
	AddressCity         string    `json:"addressCity"`
	AddressState        string    `json:"addressState"`
	AddressZip          string    `json:"addressZip"`
	FiscalSchemaVersion string    `json:"fiscalSchemaVersion"`
	Environment         string    `json:"environment"`
	Status              string    `json:"status"`
	CreatedAt           string    `json:"createdAt"`
	UpdatedAt           string    `json:"updatedAt"`
	DeletedAt           string    `json:"deletedAt"`
	SetupGaps           []string  `json:"setupGaps"`
}
