// Package databricksdatasource contains the configuration models for the
// Databricks datasource plugin (id: grafana-databricks-datasource).
package databricksdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's `id` field
// (src/plugin.json:5 in the upstream repo).
const PluginID = "grafana-databricks-datasource"

// AuthType is the authentication method selected in the configuration editor
// ("Authentication Type"). Stored in jsonData.authType. Mirrors the constants
// in pkg/authentication/constants.go.
type AuthType string

const (
	// AuthTypeUnknown is the pre-migration empty value; the backend treats it
	// as a Personal Access Token (pkg/models/settings.go:59).
	AuthTypeUnknown AuthType = ""
	// AuthTypePat authenticates with a Databricks Personal Access Token.
	AuthTypePat AuthType = "Pat"
	// AuthTypeOauthM2M authenticates with a Databricks service-principal
	// (machine-to-machine) OAuth client id + secret.
	AuthTypeOauthM2M AuthType = "OauthM2M"
	// AuthTypeOauthPT forwards the signed-in user's OAuth token (passthrough).
	AuthTypeOauthPT AuthType = "OauthPT"
	// AuthTypeOauthOBO authenticates via Azure On-Behalf-Of.
	AuthTypeOauthOBO AuthType = "OauthOBO"
	// AuthTypeAzureM2M authenticates with an Azure Entra ID service principal.
	AuthTypeAzureM2M AuthType = "AzureM2M"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the Personal Access Token (Pat / legacy
	// Unknown auth).
	SecureJsonDataKeyToken SecureJsonDataKey = "token"
	// SecureJsonDataKeyClientSecret is the OAuth M2M / Azure Entra ID M2M
	// client secret.
	SecureJsonDataKeyClientSecret SecureJsonDataKey = "clientSecret"
	// SecureJsonDataKeyAzureClientSecret is the Azure On-Behalf-Of client
	// secret, written by @grafana/azure-sdk.
	SecureJsonDataKeyAzureClientSecret SecureJsonDataKey = "azureClientSecret"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
	SecureJsonDataKeyClientSecret,
	SecureJsonDataKeyAzureClientSecret,
}

// Sentinel validation errors. The messages mirror the upstream constants in
// pkg/models/constants.go verbatim so callers can match on them.
var (
	ErrMissingHost         = errors.New("missing host")
	ErrMissingHTTPPath     = errors.New("missing http path")
	ErrMissingToken        = errors.New("missing token")
	ErrMissingClientID     = errors.New("missing clientId")
	ErrMissingClientSecret = errors.New("missing clientSecret")
	ErrMissingTenantID     = errors.New("missing tenantId")
	// ErrInvalidOAuth mirrors the backend's On-Behalf-Of guard
	// (pkg/models/settings.go:141-143, pkg/models/constants.go:26).
	ErrInvalidOAuth = errors.New("you must enable Forward OAuth Identity")
)

// Config is the fully loaded configuration of a Databricks datasource
// instance. It mirrors the jsonData fields the backend `Settings` struct
// declares (pkg/models/settings.go:21-42) plus the decrypted secure values.
//
// Root-level datasource settings (URL, User, BasicAuth*) are intentionally
// omitted — the Databricks backend authenticates entirely through jsonData +
// secureJsonData and never reads them (it builds its own connection from
// jsonData.host / jsonData.httpPath, pkg/database/connect.go:59-61).
//
// Note on Azure M2M keys: the editor writes jsonData.clientId / jsonData.tenantId
// (src/ConfigEditor.tsx:222,274,290) while the backend struct tags them
// clientID / tenantID (pkg/models/settings.go:32,34). Go's case-insensitive
// JSON unmarshal bridges the two, so this Config uses the editor's storage keys
// (clientId / tenantId) and still parses both forms.
type Config struct {
	// Connection.
	Host     string `json:"host,omitempty"`
	HTTPPath string `json:"httpPath,omitempty"`

	// Authentication.
	AuthType         AuthType        `json:"authType,omitempty"`
	ClientID         string          `json:"clientId,omitempty"`
	TenantID         string          `json:"tenantId,omitempty"`
	AzureCloud       string          `json:"azureCloud,omitempty"`
	AzureCredentials json.RawMessage `json:"azureCredentials,omitempty"`
	OAuthPassThru    bool            `json:"oauthPassThru,omitempty"`

	// Additional settings (numeric knobs are stored as strings by the editor
	// and backend, matching pkg/models/settings.go).
	Retries            string `json:"retries,omitempty"`
	Pause              string `json:"pause,omitempty"`
	Timeout            string `json:"timeout,omitempty"`
	MaxRows            string `json:"rows,omitempty"`
	RetryTimeout       string `json:"retryTimeout,omitempty"`
	Debug              bool   `json:"debug,omitempty"`
	EnableUnitySupport bool   `json:"enableUnitySupport,omitempty"`
	DefaultQueryFormat int    `json:"defaultQueryFormat,omitempty"`

	// Backend-only: force-enabled by LoadSettings (pkg/models/settings.go:161-168).
	CloudFetch bool `json:"cloudFetch,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (token, clientSecret, azureClientSecret).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It runs
// the full three-phase flow: parse (unmarshal jsonData + copy decrypted
// secrets), then (*Config).ApplyDefaults for curated editor/backend-parity
// defaults, then (Config).Validate to enforce the plugin's runtime contract.
//
// It mirrors the plugin's own LoadSettings (pkg/models/settings.go:45-171):
// the same secret keys per auth method, the same host/httpPath/token/clientId/
// tenantId/clientSecret requirements, the same AzureCloud default, the same
// CloudFetch force-enable, and the same On-Behalf-Of oauthPassThru guard.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context Grafana injects.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading databricks datasource config")

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
		logger.Error("databricks datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("databricks datasource config loaded",
		"authType", cfg.AuthType,
		"hasHost", cfg.Host != "",
		"unitySupport", cfg.EnableUnitySupport,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults a fresh Databricks datasource ends up with. Curated list (only
// these fields are touched):
//   - AuthType == ""            → AuthTypePat (editor useEffect, src/ConfigEditor.tsx:64-74)
//   - CloudFetch                → true (backend force-enable, pkg/models/settings.go:161-168)
//   - AzureCloud (AzureM2M only) → "AzureCloud" (pkg/models/settings.go:116-118)
//
// It is exported so callers that assemble a Config directly still get
// editor/backend parity.
func (c *Config) ApplyDefaults() {
	if c.AuthType == AuthTypeUnknown {
		c.AuthType = AuthTypePat
	}
	// The backend unconditionally enables CloudFetch on every load unless the
	// `disableCloudFetch` Grafana feature toggle is set, so true is the
	// effective default.
	c.CloudFetch = true
	// Azure Entra ID M2M defaults the cloud to AzureCloud when unset.
	if c.AuthType == AuthTypeAzureM2M && c.AzureCloud == "" {
		c.AzureCloud = "AzureCloud"
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract: host and httpPath are present, and the selected auth method has
// its required inputs. It mirrors the checks LoadSettings performs when
// validate is true (pkg/models/settings.go:50-144, invoked from
// pkg/driver/driver.go). Errors are joined so callers see every problem at
// once. It does not mutate the Config.
func (c Config) Validate() error {
	var errs []error

	if c.Host == "" {
		errs = append(errs, ErrMissingHost)
	}
	if c.HTTPPath == "" {
		errs = append(errs, ErrMissingHTTPPath)
	}

	switch c.AuthType {
	case AuthTypePat, AuthTypeUnknown:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, ErrMissingToken)
		}
	case AuthTypeOauthPT:
		// OAuth Passthrough uses the forwarded user token; no stored
		// credentials are required.
	case AuthTypeOauthM2M:
		if c.ClientID == "" {
			errs = append(errs, ErrMissingClientID)
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] == "" {
			errs = append(errs, ErrMissingClientSecret)
		}
	case AuthTypeAzureM2M:
		if c.TenantID == "" {
			errs = append(errs, ErrMissingTenantID)
		}
		if c.ClientID == "" {
			errs = append(errs, ErrMissingClientID)
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] == "" {
			errs = append(errs, ErrMissingClientSecret)
		}
	case AuthTypeOauthOBO:
		if !c.OAuthPassThru {
			errs = append(errs, ErrInvalidOAuth)
		}
		if len(c.AzureCredentials) == 0 {
			errs = append(errs, errors.New("azureCredentials is required for Azure On-Behalf-Of auth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAzureClientSecret] == "" {
			errs = append(errs, errors.New("azureClientSecret is required for Azure On-Behalf-Of auth"))
		}
	default:
		errs = append(errs, fmt.Errorf("unknown authType %q", c.AuthType))
	}

	return errors.Join(errs...)
}
