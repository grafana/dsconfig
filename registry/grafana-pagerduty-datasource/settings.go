// Package pagerdutydatasource contains the configuration models for the
// PagerDuty datasource plugin (plugin id: grafana-pagerduty-datasource).
package pagerdutydatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json:5 in the
// upstream repo.
const PluginID = "grafana-pagerduty-datasource"

// AuthSchemeID identifies the selected OpenAPI security scheme, stored in
// jsonData.auth.id. PagerDuty's customization declares exactly one scheme
// (pkg/customization.json:6-11) and disables no-auth (supportsNoAuth:false at
// pkg/customization.json:5), so the only valid value is "api_key".
type AuthSchemeID string

const (
	// AuthSchemeIDAPIKey is the only security scheme PagerDuty exposes: the
	// "api_key" apiKey scheme (pkg/spec.json:3699-3706, pkg/customization.json:7).
	AuthSchemeIDAPIKey AuthSchemeID = "api_key"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
//
// The generic OpenAPI datasource framework namespaces every secret by the
// security scheme id: auth.<schemeId>.apiKey (pkg/openapids/httpclient.go:43).
// With only the "api_key" scheme, PagerDuty stores exactly one secret under
// the literal dotted key below.
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAPIKey is the PagerDuty REST API key, read at
	// pkg/openapids/httpclient.go:43 as
	// DecryptedSecureJSONData["auth.api_key.apiKey"].
	SecureJsonDataKeyAPIKey SecureJsonDataKey = "auth.api_key.apiKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin. PagerDuty stores
// exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAPIKey,
}

// AuthConfig is the nested jsonData.auth object. Only the scheme id is stored
// for the api_key scheme; the per-scheme credential sub-objects
// (jsonData.auth.<scheme>.{username,clientId}) are only written for the basic
// and oauth2 schemes (pkg/openapids/options.go:32-58,
// src/openapids/components/config-editor/Auth/Auth.tsx:131-145), which
// PagerDuty does not use.
type AuthConfig struct {
	// ID is the selected security scheme id (jsonData.auth.id). The config
	// editor sets it automatically to the only method on mount
	// (src/openapids/components/config-editor/Auth/Auth.tsx:190-199); the
	// backend reads it to pick the scheme (pkg/openapids/httpclient.go:34-35).
	ID AuthSchemeID `json:"id"`
}

// Config is the fully loaded configuration of a PagerDuty datasource instance.
//
// The PagerDuty backend is the generic OpenAPI datasource driver: it reads
// jsonData.auth.id (pkg/openapids/options.go:49) and the decrypted secret
// keyed auth.<schemeId>.apiKey (pkg/openapids/httpclient.go:43). It stores
// nothing at the datasource root level, so only the nested auth object and the
// decrypted secure data live here.
//
// jsonData.servers (pkg/openapids/options.go:22) and
// jsonData.enableSecureSocksProxy are intentionally not modeled: PagerDuty's
// spec has a single server and no variables, so the editor never writes
// servers (Connection renders nothing,
// src/openapids/components/config-editor/Connection.tsx:43-45) and the base URL
// is always https://api.pagerduty.com (pkg/spec.json:164-169); the Secure
// Socks Proxy field is excluded per AGENTS.md. json.Unmarshal silently ignores
// both on parse.
type Config struct {
	Auth AuthConfig `json:"auth"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (auth.api_key.apiKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the parse behavior of the generic OpenAPI driver's
// loadOptionsFromPluginSettings (pkg/openapids/options.go:36-70): an absent
// jsonData is treated as an empty object (not a parse error), jsonData is
// unmarshaled to recover auth.id, and the decrypted secrets are copied by known
// key.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> (*Config).ApplyDefaults
// -> (Config).Validate. The generic driver itself never validates the loaded
// options, so this is the intended shape for the plugin's own loader to sync
// to. Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on a Config they assemble themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading pagerduty datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Mirror pkg/openapids/options.go:38-43: a nil/empty jsonData is treated as
	// an empty object rather than a parse error.
	raw := settings.JSONData
	if len(raw) == 0 {
		raw = []byte(`{}`)
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
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
		logger.Error("pagerduty datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("pagerduty datasource config loaded", "authScheme", cfg.Auth.ID)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// editor-parity defaults a fresh datasource starts with. The config editor
// auto-selects the only auth method on mount and writes jsonData.auth.id
// (src/openapids/components/config-editor/Auth/Auth.tsx:190-199); this mirrors
// that by defaulting an empty scheme id to "api_key". It is kept exported so
// callers that assemble a Config directly can still get editor-parity.
//
// Curated list (only this field is touched, and only when zero-valued):
//   - Auth.ID → AuthSchemeIDAPIKey
func (c *Config) ApplyDefaults() {
	if c.Auth.ID == "" {
		c.Auth.ID = AuthSchemeIDAPIKey
	}
}

// Validate checks the runtime contract for a working PagerDuty datasource: the
// api_key scheme must be selected (no-auth is unsupported,
// pkg/customization.json:5) and the REST API key secret must be present.
//
// The generic OpenAPI driver does not itself reject a missing scheme id or key
// — with an empty or unknown scheme it simply builds an unauthenticated client
// (pkg/openapids/httpclient.go:38-40) whose requests PagerDuty answers with
// HTTP 401 at health-check time (pkg/openapids/plugin.go:80-85). This method
// encodes that working contract so provisioning and preview callers fail fast.
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	switch c.Auth.ID {
	case AuthSchemeIDAPIKey:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey] == "" {
			errs = append(errs, errors.New("PagerDuty API key (secureJsonData['auth.api_key.apiKey']) is required"))
		}
	case "":
		errs = append(errs, errors.New("authentication scheme (jsonData.auth.id) is required; PagerDuty supports only 'api_key'"))
	default:
		errs = append(errs, fmt.Errorf("unknown authentication scheme %q (jsonData.auth.id); PagerDuty supports only 'api_key'", c.Auth.ID))
	}

	return errors.Join(errs...)
}
