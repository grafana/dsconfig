// Package dynatracedatasource contains the configuration models for the
// Grafana Dynatrace datasource plugin (id: grafana-dynatrace-datasource).
package dynatracedatasource

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream plugin).
const PluginID = "grafana-dynatrace-datasource"

// DefaultHTTPClientTimeout is the HTTP client timeout (seconds) applied when
// jsonData.httpClientTimeout is unset or <= 0. Mirrors the upstream default in
// LoadSettings (pkg/models/settings.go:50-53) and the editor's fallback of 30
// (src/components/config/ConfigEditor.tsx:156).
const DefaultHTTPClientTimeout = 30

// APIType is the connection type selected in the config editor's "Dynatrace API
// Type" radio. Stored under jsonData.apiType. Mirrors the upstream string
// constants (pkg/models/settings.go:16-20).
type APIType string

const (
	// APITypeSaaS builds https://<environmentId>.live.dynatrace.com/api/...
	APITypeSaaS APIType = "saas"
	// APITypeManaged builds https://<domain>/e/<environmentId>/api/...
	APITypeManaged APIType = "managed"
	// APITypeURL treats environmentId as the full base URL, appending /api/...
	APITypeURL APIType = "url"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAPIToken is the classic Dynatrace API token, sent as the
	// "Authorization: Api-Token <token>" header for the classic API endpoints
	// (pkg/dynatrace/client/rest.go:171-172).
	SecureJsonDataKeyAPIToken SecureJsonDataKey = "apiToken"
	// SecureJsonDataKeyPlatformToken is the Dynatrace platform token, sent as
	// the "Authorization: Bearer <token>" header for the Grail platform API
	// (pkg/dynatrace/client/rest.go:173-174).
	SecureJsonDataKeyPlatformToken SecureJsonDataKey = "platformToken"
	// SecureJsonDataKeyTLSCACert is the PEM CA certificate used when
	// tlsAuthWithCACert is enabled (pkg/dynatrace/client/rest.go:162-164;
	// validated at pkg/models/settings.go:68-76).
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads when loading settings
// (pkg/models/settings.go:46-48).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAPIToken,
	SecureJsonDataKeyPlatformToken,
	SecureJsonDataKeyTLSCACert,
}

// Config is the fully loaded configuration of a Dynatrace datasource instance.
//
// The jsonData fields mirror the json-tagged fields of the upstream backend
// Settings struct (pkg/models/settings.go:23-37) verbatim — same json tags,
// same types (APIType is a typed alias over the upstream plain string for
// editor-parity constants, still a Go string kind). The upstream secrets
// (APIToken, PlatformToken, TlsCACert; all json:"-") are replaced here by
// DecryptedSecureJSONData, populated from
// backend.DataSourceInstanceSettings.DecryptedSecureJSONData in LoadConfig.
//
// Root-level datasource fields (settings.URL, BasicAuth, etc.) are NOT carried
// on Config because the Dynatrace backend never reads them: pkg/models/settings.go
// unmarshals only config.JSONData and reads only decrypted secrets, and
// pkg/dynatrace/client/rest.go builds every URL from apiType/environmentId/domain.
//
// The upstream Settings also carries SdkProxyOptions (json:"-", loaded from the
// SDK HTTPClientOptions) and an unused Inputs []models.Input field — both are
// runtime/dead and intentionally omitted here (see README "Upstream findings").
type Config struct {
	APIType           APIType `json:"apiType"`
	EnvironmentID     string  `json:"environmentId"`
	Domain            string  `json:"domain"`
	SkipTLSVerify     bool    `json:"tlsSkipVerify"`
	TLSAuthWithCACert bool    `json:"tlsAuthWithCACert"`
	HTTPClientTimeout int     `json:"httpClientTimeout"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (apiToken, platformToken, tlsCACert).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the upstream LoadSettings (pkg/models/settings.go:40-62): unmarshal jsonData
// (empty JSONData is a parse error, matching upstream), copy the decrypted
// apiToken/platformToken/tlsCACert secrets by known key, then default the HTTP
// client timeout. Unlike upstream — which loads SDK proxy options and defers
// validation to the health check's CheckSettings — LoadConfig folds the health
// check's runtime contract into Validate.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Callers that assemble a Config themselves can invoke ApplyDefaults
// and Validate individually.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading dynatrace datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings (pkg/models/settings.go:42-44) calls
	// json.Unmarshal on config.JSONData unconditionally and returns an error
	// when the bytes are empty or malformed. Mirror that behavior.
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		logger.Error("failed to parse jsonData", "err", err)
		return cfg, fmt.Errorf("parse jsonData: %w", err)
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("dynatrace datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("dynatrace datasource config loaded",
		"apiType", cfg.APIType,
		"hasEnvironmentId", cfg.EnvironmentID != "",
		"hasApiToken", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIToken] != "",
		"hasPlatformToken", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPlatformToken] != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own code applies. Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// Curated defaults:
//   - APIType: APITypeSaaS when empty — editor parity
//     (src/components/config/ConfigEditor.tsx:68, `jsonData.apiType || 'saas'`);
//     the backend GetHostURL also treats an empty apiType as SaaS
//     (pkg/dynatrace/client/rest.go:44-52).
//   - HTTPClientTimeout: DefaultHTTPClientTimeout (30) when <= 0
//     (pkg/models/settings.go:50-53).
func (c *Config) ApplyDefaults() {
	if c.APIType == "" {
		c.APIType = APITypeSaaS
	}
	if c.HTTPClientTimeout <= 0 {
		c.HTTPClientTimeout = DefaultHTTPClientTimeout
	}
}

// Validate checks the runtime contract the plugin enforces before a datasource
// is usable. It folds the health check's CheckSettings
// (pkg/dynatrace/handler_healthcheck.go:141-159) together with the upstream
// Settings.Validate (pkg/models/settings.go:64-77). Errors are joined so
// callers see every problem at once.
//
// Contracts enforced:
//   - environmentId is required in every mode (NoEnvironmentIDError,
//     handler_healthcheck.go:142-144).
//   - domain is required when apiType == managed (NoUrlError,
//     handler_healthcheck.go:147-149).
//   - at least one of apiToken / platformToken is required (NoAPITokenError,
//     handler_healthcheck.go:151-153; settings.go:65-67).
//   - when tlsAuthWithCACert is true, tlsCACert must be present
//     (settings.go:68-70) and a parseable PEM certificate (settings.go:71-76).
func (c Config) Validate() error {
	var errs []error

	if c.EnvironmentID == "" {
		errs = append(errs, errors.New("environment ID (jsonData.environmentId) is required"))
	}

	if c.APIType == APITypeManaged && c.Domain == "" {
		errs = append(errs, errors.New("domain (jsonData.domain) is required for the managed API type"))
	}

	apiToken := c.DecryptedSecureJSONData[SecureJsonDataKeyAPIToken]
	platformToken := c.DecryptedSecureJSONData[SecureJsonDataKeyPlatformToken]
	if apiToken == "" && platformToken == "" {
		errs = append(errs, errors.New("an API token (secureJsonData.apiToken) or platform token (secureJsonData.platformToken) is required"))
	}

	if c.TLSAuthWithCACert {
		caCert := c.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert]
		switch {
		case caCert == "":
			errs = append(errs, errors.New("TLS CA certificate (secureJsonData.tlsCACert) is required when TLS auth with CA cert is enabled"))
		default:
			caPool := x509.NewCertPool()
			if ok := caPool.AppendCertsFromPEM([]byte(caCert)); !ok {
				errs = append(errs, errors.New("failed to parse TLS CA PEM certificate"))
			}
		}
	}

	return errors.Join(errs...)
}
