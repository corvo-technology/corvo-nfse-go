package corvonfse

// FiscalProfile is a validity period of the fiscal profile of a company.
// ReviewStatus is unreviewed, approved or rejected; SpecialTaxRegime is "0" to
// "6"; TaxRegime is MEI, SIMPLES_NACIONAL, LUCRO_PRESUMIDO, LUCRO_REAL,
// OUTRO_REGIME or UNKNOWN.
type FiscalProfile struct {
	CompanyID           string `json:"companyId"`
	CreatedAt           string `json:"createdAt"`
	IBSCBSRegularFrom   string `json:"ibsCbsRegularFrom"`
	IBSCBSRegularOption bool   `json:"ibsCbsRegularOption"`
	ID                  string `json:"id"`
	ReviewedAt          string `json:"reviewedAt"`
	ReviewedBy          string `json:"reviewedBy"`
	ReviewStatus        string `json:"reviewStatus"`
	SimpleNational      bool   `json:"simpleNational"`
	SimplesApuracao     string `json:"simplesApuracao"`
	SimplesIssRate      string `json:"simplesIssRate"`
	Source              string `json:"source"`
	SpecialTaxRegime    string `json:"specialTaxRegime"`
	TaxRegime           string `json:"taxRegime"`
	TenantID            string `json:"tenantId"`
	ValidFrom           string `json:"validFrom"`
	ValidUntil          string `json:"validUntil"`
}

// FiscalProfileUpsertParams are the fields accepted by
// FiscalProfiles.Upsert.
type FiscalProfileUpsertParams struct {
	// TaxRegime is required: MEI, SIMPLES_NACIONAL, LUCRO_PRESUMIDO,
	// LUCRO_REAL, OUTRO_REGIME or UNKNOWN.
	TaxRegime           string `json:"taxRegime"`
	Confirm             *bool  `json:"confirm,omitempty"`
	IBSCBSRegularFrom   string `json:"ibsCbsRegularFrom,omitempty"`
	IBSCBSRegularOption *bool  `json:"ibsCbsRegularOption,omitempty"`
	SimpleNational      *bool  `json:"simpleNational,omitempty"`
	// SimplesApuracao is "SN" when set.
	SimplesApuracao  string `json:"simplesApuracao,omitempty"`
	SpecialTaxRegime string `json:"specialTaxRegime,omitempty"`
	ValidFrom        string `json:"validFrom,omitempty"`
}

// WithholdingRates are the withholding rates (decimal strings) of a service
// tax profile.
type WithholdingRates struct {
	COFINS string `json:"cofins,omitempty"`
	CSLL   string `json:"csll,omitempty"`
	INSS   string `json:"inss,omitempty"`
	IRRF   string `json:"irrf,omitempty"`
	PIS    string `json:"pis,omitempty"`
}

// ServiceTaxProfile is the tax classification of a service in the catalog of a
// company.
type ServiceTaxProfile struct {
	CatalogVersion            string           `json:"catalogVersion"`
	ClassCode                 string           `json:"classCode"`
	ClassificationSource      string           `json:"classificationSource"`
	CompanyID                 string           `json:"companyId"`
	CreatedAt                 string           `json:"createdAt"`
	CST                       string           `json:"cst"`
	EffectiveFrom             string           `json:"effectiveFrom"`
	EffectiveUntil            string           `json:"effectiveUntil"`
	ID                        string           `json:"id"`
	IncidenceMunicipalityCode string           `json:"incidenceMunicipalityCode"`
	IssRate                   string           `json:"issRate"`
	IssTaxation               string           `json:"issTaxation"` // NORMAL, FIXED or EXEMPT
	IssWithholdingDefault     bool             `json:"issWithholdingDefault"`
	MunicipalServiceCode      string           `json:"municipalServiceCode"`
	Name                      string           `json:"name"`
	NBSCode                   string           `json:"nbsCode"`
	OperationIndicator        string           `json:"operationIndicator"`
	OperationNature           string           `json:"operationNature"` // SERVICE, GOODS, MIXED or PARTNER_SALON
	RegimeApplicability       []string         `json:"regimeApplicability"`
	ReviewStatus              string           `json:"reviewStatus"`
	ServiceCode               string           `json:"serviceCode"`
	TenantID                  string           `json:"tenantId"`
	WithholdingPolicy         string           `json:"withholdingPolicy"` // UNKNOWN, NONE_CONFIRMED or CONFIGURED
	WithholdingRates          WithholdingRates `json:"withholdingRates"`
}

// ServiceTaxProfileCreateParams are the fields accepted by
// ServiceTaxProfiles.Create. Name, ServiceCode and EffectiveFrom (YYYY-MM-DD)
// are required.
type ServiceTaxProfileCreateParams struct {
	Name                      string            `json:"name"`
	ServiceCode               string            `json:"serviceCode"`
	EffectiveFrom             string            `json:"effectiveFrom"`
	EffectiveUntil            string            `json:"effectiveUntil,omitempty"`
	CatalogVersion            string            `json:"catalogVersion,omitempty"`
	ClassCode                 string            `json:"classCode,omitempty"`
	ClassificationSource      string            `json:"classificationSource,omitempty"`
	CST                       string            `json:"cst,omitempty"`
	IncidenceMunicipalityCode string            `json:"incidenceMunicipalityCode,omitempty"`
	IssRate                   string            `json:"issRate,omitempty"`
	IssTaxation               string            `json:"issTaxation,omitempty"`
	IssWithholdingDefault     *bool             `json:"issWithholdingDefault,omitempty"`
	MunicipalServiceCode      string            `json:"municipalServiceCode,omitempty"`
	NBSCode                   string            `json:"nbsCode,omitempty"`
	OperationIndicator        string            `json:"operationIndicator,omitempty"`
	OperationNature           string            `json:"operationNature,omitempty"`
	RegimeApplicability       []string          `json:"regimeApplicability,omitempty"`
	WithholdingPolicy         string            `json:"withholdingPolicy,omitempty"`
	WithholdingRates          map[string]string `json:"withholdingRates,omitempty"`
}

// ServiceTaxProfileUpdateParams are the fields accepted by
// ServiceTaxProfiles.Update. A nil field is left out; the *Nullable fields can
// be cleared with Null[string]().
type ServiceTaxProfileUpdateParams struct {
	Name                      string            `json:"name,omitempty"`
	ServiceCode               string            `json:"serviceCode,omitempty"`
	EffectiveFrom             string            `json:"effectiveFrom,omitempty"`
	EffectiveUntil            *Nullable[string] `json:"effectiveUntil,omitempty"`
	CatalogVersion            *Nullable[string] `json:"catalogVersion,omitempty"`
	ClassCode                 *Nullable[string] `json:"classCode,omitempty"`
	ClassificationSource      string            `json:"classificationSource,omitempty"`
	CST                       *Nullable[string] `json:"cst,omitempty"`
	IncidenceMunicipalityCode *Nullable[string] `json:"incidenceMunicipalityCode,omitempty"`
	IssRate                   *Nullable[string] `json:"issRate,omitempty"`
	IssTaxation               *Nullable[string] `json:"issTaxation,omitempty"`
	IssWithholdingDefault     *bool             `json:"issWithholdingDefault,omitempty"`
	MunicipalServiceCode      *Nullable[string] `json:"municipalServiceCode,omitempty"`
	NBSCode                   *Nullable[string] `json:"nbsCode,omitempty"`
	OperationIndicator        *Nullable[string] `json:"operationIndicator,omitempty"`
	OperationNature           string            `json:"operationNature,omitempty"`
	RegimeApplicability       []string          `json:"regimeApplicability,omitempty"`
	WithholdingPolicy         string            `json:"withholdingPolicy,omitempty"`
	WithholdingRates          map[string]string `json:"withholdingRates,omitempty"`
}
