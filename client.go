package corvonfse

import (
	"net/http"
	"os"
	"time"
)

// Version is the SDK version.
const Version = "0.3.0"

// Options configures NewClient.
//
// Zero fields fall back to the environment (CORVO_API_KEY, CORVO_BASE_URL)
// and then to the public defaults: APIKey is required (no default), BaseURL
// defaults to https://api.corvo.tech and Prefix to "v1".
type Options struct {
	// APIKey is sent as the X-API-Key header. Falls back to CORVO_API_KEY.
	APIKey string
	// BaseURL of the API. Falls back to CORVO_BASE_URL, then to
	// https://api.corvo.tech.
	BaseURL string
	// Prefix of the API routes. Defaults to "v1".
	Prefix string
	// Timeout for each request. Defaults to 30s; negative disables it.
	// Ignored when HTTPClient is provided.
	Timeout time.Duration
	// AllowInsecureHTTP permits http:// base URLs on public hosts.
	// Loopback and private hosts are always allowed.
	AllowInsecureHTTP bool
	// HTTPClient overrides the underlying *http.Client (connection pooling,
	// proxies, instrumentation). Its own Timeout applies instead of Timeout.
	HTTPClient *http.Client
}

// Client is the Corvo NFS-e API client. Create it with NewClient and use
// its resource fields (Companies, Invoices, ...) to call the API.
//
// Every operation returns (T, *CorvoError): on failure err != nil and T is
// the zero value. The SDK never panics for API, network or timeout
// failures — only NewClient panics, and only on misconfiguration.
type Client struct {
	AccountantPackage      *AccountantPackage
	Billing                *Billing
	Certificates           *Certificates
	CNPJ                   *CNPJ
	Companies              *Companies
	Credits                *Credits
	Customers              *Customers
	DPSSequence            *DPSSequence
	FiscalPortal           *FiscalPortal
	FiscalProfiles         *FiscalProfiles
	Health                 *Health
	Imports                *Imports
	IntegrationConnections *IntegrationConnections
	IntegrationOperations  *IntegrationOperations
	Integrations           *Integrations
	InvoicePackage         *InvoicePackage
	Invoices               *Invoices
	Plans                  *Plans
	Schedules              *Schedules
	ServiceFavorites       *ServiceFavorites
	ServiceTaxProfiles     *ServiceTaxProfiles
	Webhooks               *Webhooks
}

// NewClient builds a Client from opts, falling back to the CORVO_API_KEY
// and CORVO_BASE_URL environment variables.
//
// It panics on misconfiguration — a missing API key, an invalid base URL,
// or a plain http:// URL pointing at a public host — since these are
// programmer errors detected at startup, not runtime API failures.
func NewClient(opts Options) *Client {
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("CORVO_API_KEY")
	}
	if apiKey == "" {
		panic("Corvo: apiKey is required — pass Options.APIKey or set CORVO_API_KEY")
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("CORVO_BASE_URL")
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	prefix := opts.Prefix
	if prefix == "" {
		prefix = defaultPrefix
	}

	transport := newTransport(
		apiKey,
		baseURL,
		prefix,
		opts.Timeout,
		opts.AllowInsecureHTTP,
		opts.HTTPClient,
	)

	return &Client{
		AccountantPackage:      &AccountantPackage{t: transport},
		Billing:                &Billing{t: transport},
		Certificates:           &Certificates{t: transport},
		CNPJ:                   &CNPJ{t: transport},
		Companies:              &Companies{t: transport},
		Credits:                &Credits{t: transport},
		Customers:              &Customers{t: transport},
		DPSSequence:            &DPSSequence{t: transport},
		FiscalPortal:           &FiscalPortal{t: transport},
		FiscalProfiles:         &FiscalProfiles{t: transport},
		Health:                 &Health{t: transport},
		Imports:                &Imports{t: transport},
		IntegrationConnections: &IntegrationConnections{t: transport},
		IntegrationOperations:  &IntegrationOperations{t: transport},
		Integrations:           &Integrations{t: transport},
		InvoicePackage:         &InvoicePackage{t: transport},
		Invoices:               &Invoices{t: transport},
		Plans:                  &Plans{t: transport},
		Schedules:              &Schedules{t: transport},
		ServiceFavorites:       &ServiceFavorites{t: transport},
		ServiceTaxProfiles:     &ServiceTaxProfiles{t: transport},
		Webhooks:               newWebhooks(transport),
	}
}
