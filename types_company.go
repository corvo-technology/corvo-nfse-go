package corvonfse

// TaxRegime is one of the Brazilian corporate tax regimes accepted when
// creating a company.
type TaxRegime string

const (
	TaxRegimeMEI              TaxRegime = "MEI"
	TaxRegimeSimplesNacional  TaxRegime = "Simples Nacional"
	TaxRegimeLucroPresumido   TaxRegime = "Lucro Presumido"
	TaxRegimeLucroReal        TaxRegime = "Lucro Real"
)

// CompanyAddressInput is the address supplied when creating or updating a
// company. Zip is the CEP (e.g. "38400-000").
type CompanyAddressInput struct {
	Street    string `json:"street"`
	Number    string `json:"number,omitempty"`
	Complement string `json:"complement,omitempty"`
	District  string `json:"district"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
}

// CompanyCreateParams are the fields accepted by Companies.Create.
type CompanyCreateParams struct {
	LegalName       string                `json:"legalName"`
	TradeName       string                `json:"tradeName,omitempty"`
	TaxID           string                `json:"taxId"`
	MunicipalReg    string                `json:"municipalReg,omitempty"`
	MunicipalityCode string               `json:"municipalityCode,omitempty"`
	TaxRegime       TaxRegime             `json:"taxRegime"`
	SimpleNational  *bool                 `json:"simpleNational,omitempty"`
	SpecialRegime   string                `json:"specialRegime,omitempty"`
	Address         *CompanyAddressInput  `json:"address,omitempty"`
}

// CompanyUpdateParams are the fields accepted by Companies.Update. Fields
// left at their zero value are omitted from the request, leaving the
// server-side value unchanged.
type CompanyUpdateParams struct {
	LegalName       string               `json:"legalName,omitempty"`
	TradeName       string               `json:"tradeName,omitempty"`
	MunicipalReg    string               `json:"municipalReg,omitempty"`
	MunicipalityCode string              `json:"municipalityCode,omitempty"`
	TaxRegime       TaxRegime            `json:"taxRegime,omitempty"`
	SimpleNational  *bool                `json:"simpleNational,omitempty"`
	SpecialRegime   string               `json:"specialRegime,omitempty"`
	Address         *CompanyAddressInput `json:"address,omitempty"`
}

// Company is a fiscal company (emitente) registered with the API.
// Nullable server fields are mapped to their zero value when null.
type Company struct {
	ID                 string   `json:"id"`
	TenantID           string   `json:"tenantId"`
	LegalName          string   `json:"legalName"`
	TradeName          string   `json:"tradeName"`
	TaxID              string   `json:"taxId"`
	MunicipalReg       string   `json:"municipalReg"`
	MunicipalityCode   string   `json:"municipalityCode"`
	TaxRegime          TaxRegime `json:"taxRegime"`
	SimpleNational     bool     `json:"simpleNational"`
	SpecialRegime      string   `json:"specialRegime"`
	AddressStreet      string   `json:"addressStreet"`
	AddressNumber      string   `json:"addressNumber"`
	AddressComplement  string   `json:"addressComplement"`
	AddressDistrict    string   `json:"addressDistrict"`
	AddressCity        string   `json:"addressCity"`
	AddressState       string   `json:"addressState"`
	AddressZip         string   `json:"addressZip"`
	FiscalSchemaVersion string  `json:"fiscalSchemaVersion"`
	Environment        string   `json:"environment"`
	Status             string   `json:"status"`
	CreatedAt          string   `json:"createdAt"`
	UpdatedAt          string   `json:"updatedAt"`
	DeletedAt          string   `json:"deletedAt"`
	SetupGaps          []string `json:"setupGaps"`
}
