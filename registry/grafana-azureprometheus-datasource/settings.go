// Package azureprometheusdatasource contains the configuration models for
// the Azure Monitor Managed Service for Prometheus datasource plugin
// (id: grafana-azureprometheus-datasource).
package azureprometheusdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's `id` field
// (src/plugin.json:6 in the upstream repo).
const PluginID = "grafana-azureprometheus-datasource"

// AuthType is the discriminator of jsonData.azureCredentials.authType.
// Mirrors the constants in grafana-azure-sdk-go/v2/azcredentials/credentials.go
// (which the backend's `azcredentials.FromDatasourceData` case-switches on).
type AuthType string

const (
	AuthTypeCurrentUser       AuthType = "currentuser"
	AuthTypeManagedIdentity   AuthType = "msi"
	AuthTypeWorkloadIdentity  AuthType = "workloadidentity"
	AuthTypeClientSecret      AuthType = "clientsecret"
	AuthTypeClientSecretOBO   AuthType = "clientsecret-obo"
	AuthTypeClientCertificate AuthType = "clientcertificate"
	AuthTypeAdPassword        AuthType = "ad-password"
)

// HTTPMethod is the HTTP verb the Prometheus client uses. Stored in
// jsonData.httpMethod. Mirrors the values validated at
// grafana-prometheus-datasource/pkg/promlib/models/settings.go:92-95.
type HTTPMethod string

const (
	HTTPMethodPOST HTTPMethod = http.MethodPost
	HTTPMethodGET  HTTPMethod = http.MethodGet
)

// PromApplication mirrors the frontend PromApplication enum
// (@grafana/prometheus types.ts).
type PromApplication string

const (
	PromApplicationPrometheus PromApplication = "Prometheus"
	PromApplicationCortex     PromApplication = "Cortex"
	PromApplicationMimir      PromApplication = "Mimir"
	PromApplicationThanos     PromApplication = "Thanos"
)

// PrometheusCacheLevel mirrors the frontend PrometheusCacheLevel enum.
type PrometheusCacheLevel string

const (
	PrometheusCacheLevelLow    PrometheusCacheLevel = "Low"
	PrometheusCacheLevelMedium PrometheusCacheLevel = "Medium"
	PrometheusCacheLevelHigh   PrometheusCacheLevel = "High"
	PrometheusCacheLevelNone   PrometheusCacheLevel = "None"
)

// QueryEditorMode mirrors the frontend QueryEditorMode enum.
type QueryEditorMode string

const (
	QueryEditorModeBuilder QueryEditorMode = "builder"
	QueryEditorModeCode    QueryEditorMode = "code"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// Modern app-registration client secret, written by @grafana/azure-sdk.
	SecureJsonDataKeyAzureClientSecret SecureJsonDataKey = "azureClientSecret"
	// Legacy client secret preserved for backward compatibility with
	// pre-migration datasources. Read as a fallback by the backend.
	SecureJsonDataKeyClientSecret SecureJsonDataKey = "clientSecret"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAzureClientSecret,
	SecureJsonDataKeyClientSecret,
}

// ExemplarTraceIdDestination mirrors the frontend
// `ExemplarTraceIdDestination` type and the backend
// `ExemplarTraceIDDestination` (pkg/promlib/models/settings.go).
type ExemplarTraceIdDestination struct {
	Name            string `json:"name"`
	URL             string `json:"url,omitempty"`
	URLDisplayLabel string `json:"urlDisplayLabel,omitempty"`
	DatasourceUID   string `json:"datasourceUid,omitempty"`
}

// Config is the fully loaded configuration of an Azure Prometheus
// datasource instance. It merges:
//
//   - the upstream backend `PromOptions` (pkg/promlib/models/settings.go)
//     mirrored verbatim — same fields, same json tags — because Azure
//     Prometheus reuses `promlib` for all query/resource execution;
//   - the Azure-specific jsonData fields from `AzurePromDataSourceOptions`
//     (`src/configuration/AzureCredentialsConfig.ts:67-70`):
//     `azureCredentials`, `azureEndpointResourceId`,
//     `prometheus-type-migration`;
//   - `DecryptedSecureJSONData` for the write-only secrets;
//   - the root-level `URL` field the SDK feeds into `promlib` via
//     `backend.DataSourceInstanceSettings`.
//
// Root basic-auth / withCredentials fields are intentionally omitted — the
// plugin's editor sets `visibleMethods=[azureAuthId]` and clears
// `basicAuth`/`withCredentials`/`oauthPassThru` on every save
// (`src/configuration/DataSourceHttpSettingsOverhaul.tsx:121-131`), and the
// plugin's own Go code (`pkg/datasource.go`, `pkg/azureauth/azure.go`) never
// reads them by name.
type Config struct {
	// Root-level. `json:"-"` because it lives on the SDK instance, not
	// jsonData. The Prometheus backend reads it via `promlib` (which
	// resolves the base URL from `settings.URL`) at query time.
	URL string `json:"-"`

	// --- Azure-specific jsonData fields ---

	// AzureCredentials is the opaque discriminated-union object written by
	// `@grafana/azure-sdk`. Parsed downstream by
	// `github.com/grafana/grafana-azure-sdk-go/v2/azcredentials.FromDatasourceData`
	// (invoked at `pkg/azureauth/azure.go:23`).
	AzureCredentials json.RawMessage `json:"azureCredentials,omitempty"`

	// AzureEndpointResourceID optionally overrides the OAuth scope audience
	// the plugin builds from the Azure cloud's `prometheusResourceId`
	// property. Editor-hidden; provisioning-only.
	AzureEndpointResourceID string `json:"azureEndpointResourceId,omitempty"`

	// PrometheusTypeMigration is `true` for datasources migrated from
	// vanilla Prometheus. Frontend-only sentinel used by
	// `DataSourceHttpSettingsOverhaul.tsx:101-117` to render a migration
	// banner. Storage key contains a hyphen.
	PrometheusTypeMigration bool `json:"prometheus-type-migration,omitempty"`

	// --- Prometheus jsonData (mirrors pkg/promlib/models/settings.go PromOptions verbatim) ---

	OAuthPassThru                       bool                         `json:"oauthPassThru,omitempty"`
	ManageAlerts                        bool                         `json:"manageAlerts,omitempty"`
	AllowAsRecordingRulesTarget         bool                         `json:"allowAsRecordingRulesTarget,omitempty"`
	Timeout                             float64                      `json:"timeout,omitempty"`
	KeepCookies                         []string                     `json:"keepCookies,omitempty"`
	TimeInterval                        string                       `json:"timeInterval,omitempty"`
	QueryTimeout                        string                       `json:"queryTimeout,omitempty"`
	DefaultEditor                       QueryEditorMode              `json:"defaultEditor,omitempty"`
	DisableMetricsLookup                bool                         `json:"disableMetricsLookup,omitempty"`
	PrometheusType                      PromApplication              `json:"prometheusType,omitempty"`
	PrometheusVersion                   string                       `json:"prometheusVersion,omitempty"`
	CacheLevel                          PrometheusCacheLevel         `json:"cacheLevel,omitempty"`
	IncrementalQuerying                 bool                         `json:"incrementalQuerying,omitempty"`
	IncrementalQueryOverlapWindow       string                       `json:"incrementalQueryOverlapWindow,omitempty"`
	DisableRecordingRules               bool                         `json:"disableRecordingRules,omitempty"`
	CustomQueryParameters               string                       `json:"customQueryParameters,omitempty"`
	HTTPMethod                          HTTPMethod                   `json:"httpMethod,omitempty"`
	SeriesLimit                         int64                        `json:"seriesLimit,omitempty"`
	SeriesEndpoint                      bool                         `json:"seriesEndpoint,omitempty"`
	ExemplarTraceIdDestinations         []ExemplarTraceIdDestination `json:"exemplarTraceIdDestinations,omitempty"`
	MaxSamplesProcessedWarningThreshold float64                      `json:"maxSamplesProcessedWarningThreshold,omitempty"`
	MaxSamplesProcessedErrorThreshold   float64                      `json:"maxSamplesProcessedErrorThreshold,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// EffectiveAuthType returns the authType the backend will parse out of
// `jsonData.azureCredentials`. Returns an empty string when no
// `azureCredentials` object is set (in which case
// `azcredentials.FromDatasourceData` returns `nil, nil` and the plugin
// leaves the HTTP client unauthenticated — see `pkg/azureauth/azure.go:23-29`).
func (c Config) EffectiveAuthType() AuthType {
	if len(c.AzureCredentials) == 0 {
		return ""
	}
	var probe struct {
		AuthType AuthType `json:"authType"`
	}
	if err := json.Unmarshal(c.AzureCredentials, &probe); err != nil {
		return ""
	}
	return probe.AuthType
}

// LoadConfig runs the full parse -> ApplyDefaults -> Validate flow and
// returns a fully-defaulted, validated Config. Parse mirrors both:
//
//   - `pkg/promlib/models/settings.go` `ParsePromOptions` for the
//     Prometheus knobs (same field set, same json tags), and
//   - `pkg/azureauth/azure.go:19-27` `ConfigureAzureAuthentication` for the
//     Azure credential extraction (delegated to
//     `azcredentials.FromDatasourceData` — kept opaque here as a
//     `json.RawMessage`).
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading azure prometheus datasource config")

	cfg := Config{
		URL:                     settings.URL,
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("azure prometheus datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("azure prometheus datasource config loaded",
		"hasURL", cfg.URL != "",
		"authType", cfg.EffectiveAuthType(),
		"httpMethod", cfg.HTTPMethod,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor and backend write. Currently only `HTTPMethod` is
// defaulted (uppercased then defaulted to POST), matching the backend
// `PromOptions.ApplyDefaults` (pkg/promlib/models/settings.go:82-87).
//
// Other editor "visual defaults" (cacheLevel=Low, defaultEditor=builder,
// seriesLimit=40000, incrementalQueryOverlapWindow=10m) are never written
// to storage on load and are therefore not applied here.
func (c *Config) ApplyDefaults() {
	c.HTTPMethod = HTTPMethod(strings.ToUpper(strings.TrimSpace(string(c.HTTPMethod))))
	if c.HTTPMethod == "" {
		c.HTTPMethod = HTTPMethodPOST
	}
}

// Validate checks the runtime contract that the plugin requires. Errors are
// joined so callers see every problem at once.
//
//   - URL is required (Prometheus backend hard-fails on empty
//     `settings.URL`, `pkg/promlib/admission_handler.go:51`).
//   - `HTTPMethod` must be empty/POST/GET
//     (`pkg/promlib/models/settings.go:92-95`).
//   - When `jsonData.azureCredentials` is set, `authType` must be one of the
//     values `grafana-azure-sdk-go/v2/azcredentials.FromDatasourceData`
//     recognises, and `clientsecret` requires
//     `secureJsonData.azureClientSecret` or the legacy
//     `secureJsonData.clientSecret` (the shared azcredentials builder hard-
//     fails otherwise).
//   - Numeric fields (`Timeout`, `SeriesLimit`) must be non-negative.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Prometheus server URL (root.url) is required"))
	}

	switch m := HTTPMethod(strings.ToUpper(string(c.HTTPMethod))); m {
	case "", HTTPMethodPOST, HTTPMethodGET:
		// OK.
	default:
		errs = append(errs, fmt.Errorf("invalid httpMethod %q: must be GET or POST", c.HTTPMethod))
	}

	if len(c.AzureCredentials) > 0 {
		var probe struct {
			AuthType AuthType `json:"authType"`
		}
		if err := json.Unmarshal(c.AzureCredentials, &probe); err != nil {
			errs = append(errs, fmt.Errorf(
				"jsonData.azureCredentials is not a valid credential object: %w", err,
			))
		} else {
			switch probe.AuthType {
			case AuthTypeClientSecret, AuthTypeClientSecretOBO:
				if c.DecryptedSecureJSONData[SecureJsonDataKeyAzureClientSecret] == "" &&
					c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] == "" {
					errs = append(errs, fmt.Errorf(
						"authType %q requires secureJsonData.azureClientSecret (or the legacy secureJsonData.clientSecret)",
						probe.AuthType,
					))
				}
			case AuthTypeManagedIdentity, AuthTypeWorkloadIdentity, AuthTypeCurrentUser:
				// No secret required at the credential level. Current User
				// with `serviceCredentialsEnabled=true` and a
				// `clientsecret` fallback would also need a secret, but
				// the fallback lives inside `AzureCredentials` and we
				// keep this validator focused on the top-level authType.
			case AuthTypeClientCertificate:
				// The Azure-SDK backend accepts clientcertificate but this
				// plugin's editor never selects it; allow it to pass so
				// provisioning stays flexible.
			case AuthTypeAdPassword:
				// Same as clientcertificate — backend accepts, editor
				// doesn't offer.
			case "":
				errs = append(errs, errors.New(
					"jsonData.azureCredentials.authType is required when azureCredentials is set",
				))
			default:
				errs = append(errs, fmt.Errorf(
					"unknown jsonData.azureCredentials.authType %q", probe.AuthType,
				))
			}
		}
	}

	if c.Timeout < 0 {
		errs = append(errs, fmt.Errorf("timeout must be non-negative, got %v", c.Timeout))
	}
	if c.SeriesLimit < 0 {
		errs = append(errs, fmt.Errorf("seriesLimit must be non-negative, got %d", c.SeriesLimit))
	}

	return errors.Join(errs...)
}
