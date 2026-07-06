// Package sentrydatasource contains the configuration models for the
// Sentry datasource plugin (plugin id: grafana-sentry-datasource).
package sentrydatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream repo).
const PluginID = "grafana-sentry-datasource"

// DefaultSentryURL is the Sentry base URL applied when jsonData.url is
// empty. It mirrors src/constants.ts:105 (DEFAULT_SENTRY_URL) and
// pkg/util/util.go (DefaultSentryURL) — both the frontend editor
// (src/editors/SentryConfigEditor.tsx:19) and the backend
// (pkg/plugin/settings.go:37-40, pkg/sentry/sentry.go:23-30) default an
// empty URL to this value for backwards compatibility.
const DefaultSentryURL = "https://sentry.io"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAuthToken is the Sentry API Bearer token, sent as
	// "Authorization: Bearer <token>" on every outgoing request
	// (pkg/sentry/client.go:37-40).
	SecureJsonDataKeyAuthToken SecureJsonDataKey = "authToken"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin. The Sentry
// datasource declares exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAuthToken,
}

// Config is the fully loaded configuration of a Sentry datasource
// instance.
//
// Fields mirror the plugin's own backend SentryConfig struct
// (pkg/plugin/settings.go:11-16) verbatim — same field names, same json
// tags, same types — except that the upstream `authToken` unexported
// field is replaced here by DecryptedSecureJSONData, which is populated
// from backend.DataSourceInstanceSettings.DecryptedSecureJSONData in
// LoadConfig.
//
// Root-level datasource fields (settings.URL, BasicAuth, etc.) are NOT
// carried on Config because the Sentry plugin never reads them:
// pkg/plugin/plugin.go builds its client from jsonData.url + orgSlug +
// the decrypted secure authToken, and pkg/plugin/settings.go only
// unmarshals settings.JSONData.
type Config struct {
	URL           string `json:"url"`
	OrgSlug       string `json:"orgSlug"`
	TLSSkipVerify bool   `json:"tlsSkipVerify"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (authToken).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`

	// Note: jsonData.enableSecureSocksProxy is written by the editor's
	// AdditionalSettings component (src/components/config-editor/AdditionalSettings.tsx:29-40)
	// and consumed transparently by the SDK's s.HTTPClientOptions(ctx)
	// call in pkg/plugin/plugin.go:51. The Sentry plugin's own Go code
	// never inspects it by name, and the upstream SentryConfig struct
	// (pkg/plugin/settings.go:11-16) does not carry it either. Following
	// AGENTS.md and upstream, it is not modeled on this Config; json
	// unmarshal silently ignores it on parse.
}

// LoadConfig parses a datasource instance's settings into a Config,
// mirroring pkg/plugin/settings.go:31-52 (GetSettings) verbatim: unmarshal
// jsonData (empty JSONData is a parse error, matching upstream), copy
// decrypted secrets by known key, apply the empty-URL default, then
// validate.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext
// so log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Callers that need each phase individually can invoke
// ApplyDefaults and Validate directly on a Config they assemble
// themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading sentry datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream GetSettings (pkg/plugin/settings.go:33-36) calls
	// json.Unmarshal on settings.JSONData unconditionally and returns
	// ErrorUnmarshalingSettings when the bytes are empty or malformed.
	// We mirror that behavior — empty JSONData is a parse error.
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
		logger.Error("sentry datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("sentry datasource config loaded",
		"url", cfg.URL,
		"orgSlug", cfg.OrgSlug,
		"tlsSkipVerify", cfg.TLSSkipVerify,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own code applies. Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// Curated defaults:
//   - URL: DefaultSentryURL ("https://sentry.io") — mirrors
//     src/constants.ts:105, src/editors/SentryConfigEditor.tsx:19,
//     pkg/plugin/settings.go:37-40, and pkg/sentry/sentry.go:23-30.
//     Both the editor and the backend treat an empty URL as
//     https://sentry.io for backwards compatibility.
//
// OrgSlug and the authToken secret have no defaults — the plugin errors
// out when either is empty.
func (c *Config) ApplyDefaults() {
	if c.URL == "" {
		c.URL = DefaultSentryURL
	}
}

// Validate checks the runtime contract that the plugin requires
// (pkg/plugin/settings.go:18-29, Validate; :41-50, GetSettings). Errors
// are joined so callers see every problem at once.
//
// Contracts enforced:
//   - URL must be non-empty. Callers should invoke ApplyDefaults first;
//     LoadConfig always does. The upstream Validate returns
//     ErrorInvalidSentryConfig ("invalid sentry configuration") when URL
//     is empty.
//   - OrgSlug must be non-empty. Upstream returns
//     ErrorInvalidOrganizationSlug ("invalid or empty organization slug").
//   - authToken (secureJsonData.authToken) must be non-empty. Upstream
//     returns ErrorInvalidAuthToken ("empty or invalid auth token found").
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("sentry URL (jsonData.url) is required"))
	}
	if c.OrgSlug == "" {
		errs = append(errs, errors.New("sentry organization slug (jsonData.orgSlug) is required"))
	}
	if c.DecryptedSecureJSONData[SecureJsonDataKeyAuthToken] == "" {
		errs = append(errs, errors.New("sentry auth token (secureJsonData.authToken) is required"))
	}

	return errors.Join(errs...)
}
