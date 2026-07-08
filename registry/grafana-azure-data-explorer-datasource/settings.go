// Package azuredataexplorerdatasource contains the configuration models
// for the Azure Data Explorer datasource plugin
// (id: grafana-azure-data-explorer-datasource).
package azuredataexplorerdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's `id`
// field (src/plugin.json:4 in the upstream repo).
const PluginID = "grafana-azure-data-explorer-datasource"

// AuthType is the discriminator of jsonData.azureCredentials.authType.
// Mirrors the subset of `AzureAuthType` (grafana-azure-sdk-react
// src/credentials/AzureCredentials.ts) that ADX's `AzureCredentialsForm`
// actually exposes.
type AuthType string

const (
	AuthTypeCurrentUser      AuthType = "currentuser"
	AuthTypeManagedIdentity  AuthType = "msi"
	AuthTypeWorkloadIdentity AuthType = "workloadidentity"
	AuthTypeClientSecret     AuthType = "clientsecret"
	AuthTypeClientSecretOBO  AuthType = "clientsecret-obo"
)

// DataConsistency is the value of jsonData.dataConsistency.
// Mirrors the options in `src/components/ConfigEditor/QueryConfig.tsx:16-19`.
type DataConsistency string

const (
	DataConsistencyStrong DataConsistency = "strongconsistency"
	DataConsistencyWeak   DataConsistency = "weakconsistency"
)

// EditorMode is the value of jsonData.defaultEditorMode.
// Mirrors `src/types/index.ts:50-54`.
type EditorMode string

const (
	EditorModeVisual EditorMode = "visual"
	EditorModeRaw    EditorMode = "raw"
	EditorModeOpenAI EditorMode = "openai"
)

// LegacyCloudName is the value the pre-`azureCredentials` config editor
// wrote to jsonData.azureCloud. Mapped to the modern `AzureCloud` values
// by `pkg/azuredx/adxauth/adxcredentials/builder.go:107-121`
// `resolveLegacyCloudName`.
type LegacyCloudName string

const (
	LegacyCloudNamePublic       LegacyCloudName = "azuremonitor"
	LegacyCloudNameChina        LegacyCloudName = "chinaazuremonitor"
	LegacyCloudNameUSGovernment LegacyCloudName = "govazuremonitor"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData.
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAzureClientSecret is the modern client secret
	// written by `@grafana/azure-sdk`'s `AzureCredentialsForm`.
	SecureJsonDataKeyAzureClientSecret SecureJsonDataKey = "azureClientSecret"
	// SecureJsonDataKeyClientSecret is the legacy top-level client secret,
	// preserved for backward compatibility.
	SecureJsonDataKeyClientSecret SecureJsonDataKey = "clientSecret"
	// SecureJsonDataKeyOpenAIAPIKey is the OpenAI API key consumed by the
	// `askOpenAI` resource endpoint. No editor UI; provisioning-only.
	SecureJsonDataKeyOpenAIAPIKey SecureJsonDataKey = "OpenAIAPIKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAzureClientSecret,
	SecureJsonDataKeyClientSecret,
	SecureJsonDataKeyOpenAIAPIKey,
}

// Config is the fully loaded configuration of an Azure Data Explorer
// datasource instance. It flattens the union of every jsonData field the
// backend touches (`pkg/azuredx/models/settings.go`,
// `pkg/azuredx/adxauth/adxcredentials/builder.go`) plus the editor-visible
// fields written by the frontend.
//
// Root-level datasource settings (URL, User, BasicAuth*) are intentionally
// omitted — the ADX backend never reads them.
type Config struct {
	// Modern discriminated-union credentials. Opaque here — parsed
	// downstream by grafana-azure-sdk-go's azcredentials.FromDatasourceData.
	AzureCredentials json.RawMessage `json:"azureCredentials,omitempty"`

	// Connection.
	ClusterURL string `json:"clusterUrl,omitempty"`

	// Editor-visible additional settings.
	Application        string            `json:"application,omitempty"`
	DefaultDatabase    string            `json:"defaultDatabase,omitempty"`
	QueryTimeoutRaw    string            `json:"queryTimeout,omitempty"`
	QueryTimeout       time.Duration     `json:"-"`
	DynamicCaching     bool              `json:"dynamicCaching,omitempty"`
	CacheMaxAge        string            `json:"cacheMaxAge,omitempty"`
	DataConsistency    DataConsistency   `json:"dataConsistency,omitempty"`
	DefaultEditorMode  EditorMode        `json:"defaultEditorMode,omitempty"`
	UseSchemaMapping   bool              `json:"useSchemaMapping,omitempty"`
	SchemaMappings     []json.RawMessage `json:"schemaMappings,omitempty"`
	EnableUserTracking bool              `json:"enableUserTracking,omitempty"`
	KeepCookies        []string          `json:"keepCookies,omitempty"`

	// Written by @grafana/azure-sdk for `clientsecret-obo`; required by
	// the backend for OBO instance creation.
	OAuthPassThru bool `json:"oauthPassThru,omitempty"`

	// Frontend-only, dead. Preserved so provisioned datasources still parse.
	MinimalCache float64 `json:"minimalCache,omitempty"`

	// Legacy top-level credentials (pre-migration).
	AzureCloud string `json:"azureCloud,omitempty"`
	OnBehalfOf bool   `json:"onBehalfOf,omitempty"`
	TenantID   string `json:"tenantId,omitempty"`
	ClientID   string `json:"clientId,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// EffectiveAuthType returns the authType that would actually be used by
// the backend. Priority: `jsonData.azureCredentials.authType` when the
// azureCredentials object exists; otherwise the legacy top-level tuple
// (tenantId+clientId+clientSecret with optional onBehalfOf) resolves to
// `clientsecret` or `clientsecret-obo`; otherwise an empty string.
func (c Config) EffectiveAuthType() AuthType {
	if len(c.AzureCredentials) > 0 {
		var probe struct {
			AuthType AuthType `json:"authType"`
		}
		if err := json.Unmarshal(c.AzureCredentials, &probe); err == nil && probe.AuthType != "" {
			return probe.AuthType
		}
	}
	if c.TenantID != "" && c.ClientID != "" && c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] != "" {
		if c.OnBehalfOf {
			return AuthTypeClientSecretOBO
		}
		return AuthTypeClientSecret
	}
	return ""
}

// LoadConfig runs the full parse -> ApplyDefaults -> Validate flow and
// returns a fully-defaulted, validated Config. Mirrors the parse path in
// `pkg/azuredx/models/settings.go:48-95` `DatasourceSettings.Load`
// (unmarshal jsonData, parse queryTimeout as a Go duration) followed by
// the credential-resolution logic in
// `pkg/azuredx/adxauth/adxcredentials/builder.go:12-38`
// (modern-then-legacy credential parse).
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading azure data explorer datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Parse the query timeout as a Go duration string. Mirrors
	// pkg/azuredx/models/settings.go:65-71 — empty defaults to 30s,
	// invalid values fail hard.
	if cfg.QueryTimeoutRaw == "" {
		cfg.QueryTimeout = 30 * time.Second
	} else {
		parsed, err := time.ParseDuration(cfg.QueryTimeoutRaw)
		if err != nil {
			logger.Error("failed to parse queryTimeout", "err", err, "raw", cfg.QueryTimeoutRaw)
			return cfg, fmt.Errorf("parse queryTimeout %q: %w", cfg.QueryTimeoutRaw, err)
		}
		cfg.QueryTimeout = parsed
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("azure data explorer datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("azure data explorer datasource config loaded",
		"authType", cfg.EffectiveAuthType(),
		"hasClusterURL", cfg.ClusterURL != "",
		"hasDefaultDatabase", cfg.DefaultDatabase != "",
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with
// editor-parity defaults. The curated list mirrors the writes the
// `QueryConfig` component performs in its `useEffect`
// (`src/components/ConfigEditor/QueryConfig.tsx:27-34`) — namely
// `dataConsistency` and `defaultEditorMode`. Credential defaults are
// deliberately NOT applied here because the editor's choice is
// Grafana-instance-dependent (`azure.managedIdentityEnabled` /
// `workloadIdentityEnabled` / `userIdentityEnabled` — see
// `src/components/ConfigEditor/AzureCredentialsConfig.ts:22-34`).
func (c *Config) ApplyDefaults() {
	if c.DataConsistency == "" {
		c.DataConsistency = DataConsistencyStrong
	}
	if c.DefaultEditorMode == "" {
		c.DefaultEditorMode = EditorModeVisual
	}
}

// Validate checks the runtime contract. Mirrors the hard-fail checks the
// backend performs during instance creation:
//   - `pkg/azuredx/adxauth/adxcredentials/builder.go:89-98`
//     (`ensureOnBehalfOfSupported`) — `clientsecret-obo` requires
//     `jsonData.oauthPassThru == true`.
//   - `grafana-azure-sdk-go/v2/azcredentials/builder.go` case `clientsecret`
//     — requires `secureJsonData.azureClientSecret` (or the legacy
//     `secureJsonData.clientSecret` via `getFromLegacy`).
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

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
				if probe.AuthType == AuthTypeClientSecretOBO && !c.OAuthPassThru {
					errs = append(errs, errors.New(
						"authType \"clientsecret-obo\" requires jsonData.oauthPassThru == true "+
							"(pkg/azuredx/adxauth/adxcredentials/builder.go:89-98)",
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
	} else if c.TenantID != "" || c.ClientID != "" {
		// Legacy top-level credentials path: `getFromLegacy` requires all
		// three of tenantId, clientId, clientSecret to be non-empty
		// (pkg/azuredx/adxauth/adxcredentials/builder.go:60-62). If any
		// is set but they're not all present, the backend silently drops
		// them; we surface it here.
		if c.TenantID == "" || c.ClientID == "" ||
			c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] == "" {
			errs = append(errs, errors.New(
				"legacy credentials require all of jsonData.tenantId, jsonData.clientId, and secureJsonData.clientSecret",
			))
		}
		if c.OnBehalfOf && !c.OAuthPassThru {
			errs = append(errs, errors.New(
				"legacy onBehalfOf=true requires jsonData.oauthPassThru == true "+
					"(pkg/azuredx/adxauth/adxcredentials/builder.go:89-98)",
			))
		}
	}

	// Backend `DatasourceSettings.Load` fails hard if the parsed
	// queryTimeout exceeds one hour (`formatTimeout`).
	if c.QueryTimeout > time.Hour {
		errs = append(errs, fmt.Errorf(
			"queryTimeout %s exceeds the one-hour maximum enforced by pkg/azuredx/models/settings.go formatTimeout",
			c.QueryTimeout,
		))
	}

	// Backend `NewConnectionProperties` writes `dataConsistency` verbatim
	// into the Kusto request. Reject unknown values early.
	if c.DataConsistency != "" &&
		c.DataConsistency != DataConsistencyStrong &&
		c.DataConsistency != DataConsistencyWeak {
		errs = append(errs, fmt.Errorf(
			"unknown jsonData.dataConsistency %q (want strongconsistency or weakconsistency)",
			c.DataConsistency,
		))
	}

	return errors.Join(errs...)
}
