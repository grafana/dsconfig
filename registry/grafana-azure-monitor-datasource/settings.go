// Package azuremonitordatasource contains the configuration models for the
// Azure Monitor datasource plugin (id: grafana-azure-monitor-datasource).
package azuremonitordatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's `id` field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "grafana-azure-monitor-datasource"

// AuthType is the discriminator of jsonData.azureCredentials.authType.
// Mirrors the constants in grafana-azure-sdk-go/v2/azcredentials/credentials.go
// (which the backend's `azcredentials.FromDatasourceData` case-switches on).
type AuthType string

const (
	AuthTypeCurrentUser       AuthType = "currentuser"
	AuthTypeManagedIdentity   AuthType = "msi"
	AuthTypeWorkloadIdentity  AuthType = "workloadidentity"
	AuthTypeClientSecret      AuthType = "clientsecret"
	AuthTypeClientCertificate AuthType = "clientcertificate"
	AuthTypeClientSecretOBO   AuthType = "clientsecret-obo"
	AuthTypeAdPassword        AuthType = "ad-password"
)

// CertificateFormat is the value of
// jsonData.azureCredentials.certificateFormat when authType is
// `clientcertificate`. Mirrors
// grafana-azure-sdk-react/src/credentials/AzureCredentials.ts:58-61.
type CertificateFormat string

const (
	CertificateFormatPEM CertificateFormat = "pem"
	CertificateFormatPFX CertificateFormat = "pfx"
)

// LegacyCloudName is the value the pre-`azureCredentials` config editor
// wrote to jsonData.cloudName. Mapped to the modern `AzureCloud` values by
// resolveLegacyCloudName in the frontend (src/clouds.ts) and by
// pkg/azuremonitor/azmoncredentials/builder.go:126-142 on the backend.
type LegacyCloudName string

const (
	LegacyCloudNamePublic       LegacyCloudName = "azuremonitor"
	LegacyCloudNameChina        LegacyCloudName = "chinaazuremonitor"
	LegacyCloudNameUSGovernment LegacyCloudName = "govazuremonitor"
	LegacyCloudNameCustomized   LegacyCloudName = "customizedazuremonitor"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData. Every constant here maps to a key one of Azure Monitor's
// authType paths reads.
type SecureJsonDataKey string

const (
	// Modern secure keys, written by @grafana/azure-sdk's AzureCredentialsForm.
	SecureJsonDataKeyAzureClientSecret   SecureJsonDataKey = "azureClientSecret"
	SecureJsonDataKeyClientCertificate   SecureJsonDataKey = "clientCertificate"
	SecureJsonDataKeyPrivateKey          SecureJsonDataKey = "privateKey"
	SecureJsonDataKeyCertificatePassword SecureJsonDataKey = "certificatePassword"
	// Entra password (authType='ad-password'). Backend-only today.
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// Legacy top-level client secret preserved for backward compatibility.
	SecureJsonDataKeyClientSecret SecureJsonDataKey = "clientSecret"
	// Deprecated App Insights API key preserved on migrated datasources.
	SecureJsonDataKeyAppInsightsAPIKey SecureJsonDataKey = "appInsightsApiKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAzureClientSecret,
	SecureJsonDataKeyClientCertificate,
	SecureJsonDataKeyPrivateKey,
	SecureJsonDataKeyCertificatePassword,
	SecureJsonDataKeyPassword,
	SecureJsonDataKeyClientSecret,
	SecureJsonDataKeyAppInsightsAPIKey,
}

// Config is the fully loaded configuration of an Azure Monitor datasource
// instance. It mirrors the union of every jsonData field the backend
// touches (`pkg/azuremonitor/types/types.go:28-37`,
// `azmoncredentials/builder.go`, `loganalytics/azure-log-analytics-datasource.go:328,415`,
// `routes.go:93-106`) plus the deprecated fields the frontend still writes
// through the type in `src/types/types.ts:30-53`.
//
// Root-level datasource settings (URL, User, BasicAuth*) are intentionally
// omitted — the Azure Monitor backend never reads them.
type Config struct {
	// Modern discriminated-union credentials. Opaque here — parsed
	// downstream by grafana-azure-sdk-go's azcredentials.FromDatasourceData.
	AzureCredentials json.RawMessage `json:"azureCredentials,omitempty"`

	// Config-editor-visible jsonData fields.
	SubscriptionID   string   `json:"subscriptionId,omitempty"`
	BasicLogsEnabled bool     `json:"basicLogsEnabled,omitempty"`
	Timeout          float64  `json:"timeout,omitempty"`
	KeepCookies      []string `json:"keepCookies,omitempty"`

	// Written by @grafana/azure-sdk when the selected auth type warrants it.
	OAuthPassThru       bool `json:"oauthPassThru,omitempty"`
	DisableGrafanaCache bool `json:"disableGrafanaCache,omitempty"`

	// Backend-only.
	CustomizedRoutes map[string]any `json:"customizedRoutes,omitempty"`

	// Deprecated Application Insights / Log Analytics fields.
	AppInsightsAppID             string          `json:"appInsightsAppId,omitempty"`
	LogAnalyticsDefaultWorkspace string          `json:"logAnalyticsDefaultWorkspace,omitempty"`
	AzureLogAnalyticsSameAs      json.RawMessage `json:"azureLogAnalyticsSameAs,omitempty"`
	LogAnalyticsTenantID         string          `json:"logAnalyticsTenantId,omitempty"`
	LogAnalyticsClientID         string          `json:"logAnalyticsClientId,omitempty"`
	LogAnalyticsSubscriptionID   string          `json:"logAnalyticsSubscriptionId,omitempty"`

	// Legacy top-level credentials (pre-migration).
	AzureAuthType AuthType `json:"azureAuthType,omitempty"`
	CloudName     string   `json:"cloudName,omitempty"`
	TenantID      string   `json:"tenantId,omitempty"`
	ClientID      string   `json:"clientId,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// EffectiveAuthType returns the authType that would actually be used by the
// backend. Priority: `jsonData.azureCredentials.authType` if the
// azureCredentials object exists; otherwise the legacy top-level
// `jsonData.azureAuthType`; otherwise an empty string (the backend then
// falls back to Grafana's default cloud credentials).
func (c Config) EffectiveAuthType() AuthType {
	if len(c.AzureCredentials) > 0 {
		var probe struct {
			AuthType AuthType `json:"authType"`
		}
		if err := json.Unmarshal(c.AzureCredentials, &probe); err == nil && probe.AuthType != "" {
			return probe.AuthType
		}
	}
	return c.AzureAuthType
}

// LoadConfig runs the full parse -> ApplyDefaults -> Validate flow and
// returns a fully-defaulted, validated Config. Mirrors the parse path in
// pkg/azuremonitor/azuremonitor.go:147-158 (unmarshal jsonData twice, once
// into a generic map and once into AzureMonitorSettings) followed by
// pkg/azuremonitor/azmoncredentials/builder.go:13-31 (modern-then-legacy
// credential parse).
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading azure monitor datasource config")

	cfg := Config{
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
		logger.Error("azure monitor datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("azure monitor datasource config loaded",
		"authType", cfg.EffectiveAuthType(),
		"hasSubscription", cfg.SubscriptionID != "",
		"hasCustomizedRoutes", len(cfg.CustomizedRoutes) > 0,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with
// editor-parity defaults. The curated list is intentionally tiny: the
// editor's own default credential shape is Grafana-instance-dependent
// (managedIdentityEnabled / workloadIdentityEnabled — see
// `src/credentials.ts:66-74`), so we do NOT default `AzureCredentials`
// here. `BasicLogsEnabled` is defaulted only to satisfy the editor's
// `options.basicLogsEnabled ?? false` render behavior at
// `src/components/ConfigEditor/BasicLogsToggle.tsx:56`.
func (c *Config) ApplyDefaults() {
	// No boolean-valued default: false is already the Go zero.
	// Keeping ApplyDefaults exported so callers can still invoke it for
	// symmetry with LoadConfig; future defaults belong here.
}

// Validate checks the runtime contract. Mirrors the hard-fail checks the
// backend performs during instance creation and query execution:
//   - `pkg/azuremonitor/azmoncredentials/builder.go:97-101` — legacy
//     `azureAuthType=clientsecret` requires `secureJsonData.clientSecret`.
//   - `grafana-azure-sdk-go/v2/azcredentials/builder.go` `case
//     clientcertificate` — hard-fails if the required secret for the chosen
//     `certificateFormat` is missing.
//   - `pkg/azuremonitor/routes.go:100-102` — `AzureCustomizedCloud` requires
//     `jsonData.customizedRoutes`.
//   - `pkg/azuremonitor/loganalytics/azure-log-analytics-datasource.go:415-432`
//     — `azureLogAnalyticsSameAs=false` disables Log Analytics with a
//     specific error message.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	// Legacy `azureAuthType` path: only clientsecret is fully supported by
	// the legacy fallback, and it requires clientSecret.
	if len(c.AzureCredentials) == 0 && c.AzureAuthType == AuthTypeClientSecret {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] == "" {
			errs = append(errs, errors.New(
				"legacy clientsecret authentication requires secureJsonData.clientSecret "+
					"(pkg/azuremonitor/azmoncredentials/builder.go:97-101)",
			))
		}
	}

	// Modern discriminated-union path: enforce the per-authType secret
	// requirements the shared azcredentials builder enforces.
	if len(c.AzureCredentials) > 0 {
		var probe struct {
			AuthType          AuthType          `json:"authType"`
			CertificateFormat CertificateFormat `json:"certificateFormat"`
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
			case AuthTypeClientCertificate:
				if c.DecryptedSecureJSONData[SecureJsonDataKeyClientCertificate] == "" {
					errs = append(errs, errors.New(
						"authType \"clientcertificate\" requires secureJsonData.clientCertificate",
					))
				}
				switch probe.CertificateFormat {
				case CertificateFormatPEM:
					if c.DecryptedSecureJSONData[SecureJsonDataKeyPrivateKey] == "" {
						errs = append(errs, errors.New(
							"certificateFormat \"pem\" requires secureJsonData.privateKey",
						))
					}
				case CertificateFormatPFX:
					if c.DecryptedSecureJSONData[SecureJsonDataKeyCertificatePassword] == "" {
						errs = append(errs, errors.New(
							"certificateFormat \"pfx\" requires secureJsonData.certificatePassword",
						))
					}
				case "":
					errs = append(errs, errors.New(
						"authType \"clientcertificate\" requires certificateFormat (pem or pfx)",
					))
				default:
					errs = append(errs, fmt.Errorf(
						"unknown certificateFormat %q (want pem or pfx)", probe.CertificateFormat,
					))
				}
			case AuthTypeAdPassword:
				if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
					errs = append(errs, errors.New(
						"authType \"ad-password\" requires secureJsonData.password",
					))
				}
			case AuthTypeManagedIdentity, AuthTypeWorkloadIdentity, AuthTypeCurrentUser:
				// No secret required.
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

	// Customized cloud: `cloudName == 'customizedazuremonitor'` requires
	// `customizedRoutes` on the backend (routes.go:100-102).
	if c.CloudName == string(LegacyCloudNameCustomized) && len(c.CustomizedRoutes) == 0 {
		errs = append(errs, errors.New(
			"cloudName \"customizedazuremonitor\" requires jsonData.customizedRoutes",
		))
	}

	// `azureLogAnalyticsSameAs=false` disables Log Analytics with a hard
	// error at query time. We surface it here so provisioning catches it.
	if len(c.AzureLogAnalyticsSameAs) > 0 {
		if same, err := parseAzureLogAnalyticsSameAs(c.AzureLogAnalyticsSameAs); err != nil {
			errs = append(errs, fmt.Errorf(
				"jsonData.azureLogAnalyticsSameAs: %w (must be bool or bool-parsable string)", err,
			))
		} else if !same {
			errs = append(errs, errors.New(
				"jsonData.azureLogAnalyticsSameAs=false disables Log Analytics; "+
					"update credentials to share Azure Monitor authentication "+
					"(loganalytics/azure-log-analytics-datasource.go:415-432)",
			))
		}
	}

	return errors.Join(errs...)
}

// parseAzureLogAnalyticsSameAs mirrors the backend parse at
// azure-log-analytics-datasource.go:415-427: accept a JSON bool, or a JSON
// string that strconv.ParseBool can parse.
func parseAzureLogAnalyticsSameAs(raw json.RawMessage) (bool, error) {
	var asBool bool
	if err := json.Unmarshal(raw, &asBool); err == nil {
		return asBool, nil
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err != nil {
		return false, fmt.Errorf("unknown value: %s", string(raw))
	}
	parsed, err := strconv.ParseBool(asString)
	if err != nil {
		return false, fmt.Errorf("unknown value %q", asString)
	}
	return parsed, nil
}
