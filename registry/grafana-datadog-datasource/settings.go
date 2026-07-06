// Package datadogdatasource contains the configuration models for the Grafana
// Datadog datasource plugin (id: grafana-datadog-datasource).
package datadogdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream plugin).
const PluginID = "grafana-datadog-datasource"

// DefaultDatadogAPIURL is the Datadog API URL applied when jsonData.url is
// empty. Mirrors pkg/models/constants.go:4 (DefaultDatadogAPIURL) and the
// backend default set in defaultJSONSettings (pkg/models/settings.go:107-112).
const DefaultDatadogAPIURL = "https://api.datadoghq.com"

// DefaultDatadogAPIResponseSize is the default jsonData.size, applied when it
// is zero. Mirrors pkg/models/constants.go:7 (DefaultDatadogAPIResponseSize).
const DefaultDatadogAPIResponseSize = 100

// PluginMode is the connection/authentication mode selected in the config
// editor's "Mode" radio. Stored under jsonData.pluginMode. Mirrors the
// upstream PluginMode type (pkg/models/settings.go:12-17).
type PluginMode string

const (
	// PluginModeDefault connects directly to the Datadog API and
	// authenticates with an API key + Application key (secureJsonData).
	PluginModeDefault PluginMode = "default"
	// PluginModeHostedMetrics connects through a Grafana Cloud Datadog proxy
	// and authenticates with datasource root basic auth.
	PluginModeHostedMetrics PluginMode = "hosted-metrics"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAPIKey is the Datadog API key, sent as the DD-API-KEY
	// header (pkg/datadog/client/client_v1.go:98). Required in default mode.
	SecureJsonDataKeyAPIKey SecureJsonDataKey = "apiKey"
	// SecureJsonDataKeyAppKey is the Datadog Application key, sent as the
	// DD-APPLICATION-KEY header (pkg/datadog/client/client_v1.go:99). Required
	// in default mode.
	SecureJsonDataKeyAppKey SecureJsonDataKey = "appKey"
	// SecureJsonDataKeyBasicAuthPassword is the Grafana Cloud API key used as
	// the basic-auth password in hosted-metrics mode
	// (pkg/models/settings.go:70; pkg/datadog/client/client_v1.go:86).
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads when loading
// settings (pkg/models/settings.go:57-58,70).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAPIKey,
	SecureJsonDataKeyAppKey,
	SecureJsonDataKeyBasicAuthPassword,
}

// boolMaybeQuoted is a bool that also unmarshals from the JSON strings "true"
// / "false" (in addition to the JSON booleans). It mirrors the upstream
// boolMaybeQuoted type (pkg/models/settings.go:120-133) verbatim so LoadConfig
// tolerates the same legacy encodings the plugin does. Its Go kind is bool, so
// the conformance suite still maps the corresponding jsonData fields to the
// "boolean" value type.
type boolMaybeQuoted bool

// UnmarshalJSON accepts true/false as either JSON booleans or quoted JSON
// strings. Anything other than "true" (quotes stripped) decodes to false,
// matching pkg/models/settings.go:122-126.
func (b *boolMaybeQuoted) UnmarshalJSON(data []byte) error {
	data = maybeRemoveQuotes(data)
	*b = boolMaybeQuoted(string(data) == "true")
	return nil
}

func maybeRemoveQuotes(data []byte) []byte {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		return data[1 : len(data)-1]
	}
	return data
}

// Config is the fully loaded configuration of a Grafana Datadog datasource
// instance.
//
// The jsonData fields mirror the upstream internal jsonSettings parse struct
// (pkg/models/settings.go:94-105) for the current (non-legacy) storage keys —
// same json tags, same lenient boolMaybeQuoted booleans, same float64/int
// numeric types.
//
// Root-level fields (BasicAuthEnabled, BasicAuthUser) are carried with
// json:"-" tags so they don't collide with jsonData unmarshaling. The Datadog
// backend does read them: pkg/models/settings.go:68-72 copies
// config.BasicAuthEnabled / config.BasicAuthUser and uses BasicAuthEnabled as
// the legacy hosted-metrics signal (getPluginMode, :84-92). The root url is
// NOT carried — the backend reads jsonData.url, not settings.URL.
type Config struct {
	// Root-level datasource fields the backend reads (json:"-" — not
	// jsonData). Only meaningful in hosted-metrics mode.
	BasicAuthEnabled bool   `json:"-"`
	BasicAuthUser    string `json:"-"`

	// jsonData fields. Mirror pkg/models/settings.go:94-105 for current keys.
	PluginMode       PluginMode      `json:"pluginMode,omitempty"`
	URL              string          `json:"url"`
	LogAPIRateLimits boolMaybeQuoted `json:"logApiRateLimits"`
	RateLimitEnabled boolMaybeQuoted `json:"rateLimitEnabled"`
	RateLimitMetrics float64         `json:"rateLimitMetrics"`
	DisableDataLinks boolMaybeQuoted `json:"disableDataLinks"`
	Size             int             `json:"size"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (apiKey, appKey, basicAuthPassword).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// legacyJSONKeys captures the 1.x plaintext credential keys the editor and
// backend migrate into secureJsonData on load (pkg/models/settings.go:52-58,
// 103-104). They are parsed only inside LoadConfig and never persisted on
// Config, so they never appear as schema jsonData fields.
type legacyJSONKeys struct {
	APIKey *string `json:"api_key"`
	AppKey *string `json:"app_key"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the upstream LoadSettings (pkg/models/settings.go:38-82): seed the url/size
// defaults (defaultJSONSettings), unmarshal jsonData, migrate the legacy
// api_key/app_key jsonData keys into secureJsonData when the modern secrets are
// unset, copy decrypted secrets by known key, and lift the root basic-auth
// fields off the instance settings.
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

	logger.Debug("loading datadog datasource config")

	// Seed defaults before unmarshal, mirroring defaultJSONSettings
	// (pkg/models/settings.go:107-112): absent url/size keys keep these.
	cfg := Config{
		URL:                     DefaultDatadogAPIURL,
		Size:                    DefaultDatadogAPIResponseSize,
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	var legacy legacyJSONKeys
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
		// Best-effort legacy key parse; malformed jsonData already failed above.
		_ = json.Unmarshal(settings.JSONData, &legacy)
	}

	// Copy decrypted secrets by known key name.
	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	// Migrate 1.x plaintext api_key/app_key into secureJsonData when the modern
	// secret is unset (pkg/models/settings.go:52-54,114-118).
	migrateLegacySecret(legacy.APIKey, SecureJsonDataKeyAPIKey, cfg.DecryptedSecureJSONData)
	migrateLegacySecret(legacy.AppKey, SecureJsonDataKeyAppKey, cfg.DecryptedSecureJSONData)

	// Root datasource fields the backend reads (pkg/models/settings.go:68-72).
	cfg.BasicAuthEnabled = settings.BasicAuthEnabled
	cfg.BasicAuthUser = settings.BasicAuthUser

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("datadog datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("datadog datasource config loaded",
		"pluginMode", cfg.PluginMode,
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuthEnabled,
	)
	return cfg, nil
}

// migrateLegacySecret copies a legacy plaintext value into the secure map when
// the target key is not already set, mirroring migrateToSecureKey
// (pkg/models/settings.go:114-118).
func migrateLegacySecret(legacy *string, key SecureJsonDataKey, secure map[SecureJsonDataKey]string) {
	if legacy != nil && secure[key] == "" {
		secure[key] = *legacy
	}
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own code applies. Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// Curated defaults:
//   - PluginMode: resolved via the editor/backend getPluginMode rule
//     (src/components/ConfigEditor.tsx:527-535, pkg/models/settings.go:84-92) —
//     when unset, hosted-metrics if root BasicAuthEnabled is true, otherwise
//     default.
//   - URL: DefaultDatadogAPIURL when empty (pkg/models/settings.go:108-109).
//   - Size: DefaultDatadogAPIResponseSize when zero
//     (pkg/models/settings.go:110).
//   - RateLimitMetrics: 100 when RateLimitEnabled is true and it is 0
//     (pkg/models/settings.go:63-65).
func (c *Config) ApplyDefaults() {
	if c.PluginMode == "" {
		if c.BasicAuthEnabled {
			c.PluginMode = PluginModeHostedMetrics
		} else {
			c.PluginMode = PluginModeDefault
		}
	}
	if c.URL == "" {
		c.URL = DefaultDatadogAPIURL
	}
	if c.Size == 0 {
		c.Size = DefaultDatadogAPIResponseSize
	}
	if bool(c.RateLimitEnabled) && c.RateLimitMetrics == 0 {
		c.RateLimitMetrics = 100
	}
}

// Validate checks the runtime contract the plugin enforces in its health check
// (pkg/datadog/health_diagnostics.go:81-105, CheckSettings). Errors are joined
// so callers see every problem at once.
//
// Contracts enforced:
//   - Default mode (PluginMode != hosted-metrics): apiKey and appKey are both
//     required (:82-89).
//   - Hosted-metrics mode: url must not be the default Datadog API URL
//     (:91-93), and both the basic-auth username (root) and password
//     (secureJsonData.basicAuthPassword) are required (:94-99).
//   - url is required in every mode (:101-103). Callers should invoke
//     ApplyDefaults first (LoadConfig always does) so the default URL is in
//     place.
func (c Config) Validate() error {
	var errs []error

	if c.PluginMode == PluginModeHostedMetrics {
		if c.URL == DefaultDatadogAPIURL {
			errs = append(errs, errors.New("hosted-metrics mode requires a non-default hosted metrics url (jsonData.url)"))
		}
		if strings.TrimSpace(c.BasicAuthUser) == "" {
			errs = append(errs, errors.New("hosted-metrics mode requires a basic auth username (root.basicAuthUser)"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword] == "" {
			errs = append(errs, errors.New("hosted-metrics mode requires a basic auth password (secureJsonData.basicAuthPassword)"))
		}
	} else {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey] == "" {
			errs = append(errs, errors.New("API key (secureJsonData.apiKey) is required in default mode"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAppKey] == "" {
			errs = append(errs, errors.New("App key (secureJsonData.appKey) is required in default mode"))
		}
	}

	if strings.TrimSpace(c.URL) == "" {
		errs = append(errs, errors.New("url (jsonData.url) is required"))
	}

	return errors.Join(errs...)
}
