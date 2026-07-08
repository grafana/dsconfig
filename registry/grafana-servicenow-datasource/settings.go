// Package servicenowdatasource contains the configuration models for the
// ServiceNow datasource plugin (plugin id: grafana-servicenow-datasource).
package servicenowdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "grafana-servicenow-datasource"

// DefaultQueryTimeoutSeconds is the per-query timeout the plugin falls back to
// when queryTimeoutSeconds is unset or less than 1
// (pkg/models/settings.go:73-77).
const DefaultQueryTimeoutSeconds = 30

// AuthMethod is the authentication method selected in the configuration editor
// ("Authentication Type" radio). Stored in jsonData.authMethod. Mirrors the
// plugin's pkg/models/auth_method.go:4-9 AuthMethod alias and constants.
type AuthMethod string

const (
	// AuthMethodBasicAuth authenticates with HTTP Basic auth ("Basic auth" in
	// the editor). This is the default.
	AuthMethodBasicAuth AuthMethod = "basicAuth"
	// AuthMethodServiceNowOAuth authenticates with the ServiceNow OAuth2
	// resource-owner password grant ("ServiceNow OAuth" in the editor).
	AuthMethodServiceNowOAuth AuthMethod = "serviceNowOAuth"
)

// GetAuthMethod resolves the effective authentication method, mirroring the
// plugin's pkg/models/auth_method.go:12-22 GetAuthMethod: an explicit
// serviceNowOAuth wins first; then the legacy oauthEnabled boolean selects OAuth
// (so oauthEnabled=true overrides even an explicit authMethod="basicAuth" — an
// upstream quirk); everything else (empty or unrecognized) resolves to
// basicAuth.
func GetAuthMethod(authMethod AuthMethod, oauthEnabled bool) AuthMethod {
	if authMethod == AuthMethodServiceNowOAuth {
		return AuthMethodServiceNowOAuth
	}
	if oauthEnabled {
		return AuthMethodServiceNowOAuth
	}
	return AuthMethodBasicAuth
}

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the ServiceNow account password. It
	// is the standard Grafana Basic-auth secret key and is used both as the HTTP
	// Basic password and as the OAuth2 password-grant password
	// (pkg/models/settings.go:92,99).
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyOAuthClientSecret is the OAuth application's client secret
	// (serviceNowOAuth; pkg/models/settings.go:101).
	SecureJsonDataKeyOAuthClientSecret SecureJsonDataKey = "oauthClientSecret"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads (LoadSettings copies
// them from DecryptedSecureJSONData, pkg/models/settings.go:92,99,101).
//
// Note: the CustomHeadersSettings editor also writes dynamic httpHeaderValue<N>
// secrets for custom HTTP headers; those indexed keys are not represented here
// because they are dynamic (see README).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyOAuthClientSecret,
}

// Config is the fully loaded configuration of a ServiceNow datasource instance.
//
// The json-tagged fields mirror the jsonData portion the plugin's LoadSettings
// unmarshals from config.JSONData (pkg/models/settings.go:41-51): authMethod,
// oauthClientID, oauthEnabled, useSysTables, queryTimeoutSeconds.
//
// Root-level datasource fields the backend reads are carried with json:"-" so
// they never collide with jsonData: URL (config.URL, pkg/models/settings.go:82)
// and BasicAuthUser (config.BasicAuthUser, pkg/models/settings.go:91,98). The
// account password and OAuth client secret live in secureJsonData and are
// modeled in DecryptedSecureJSONData rather than as struct fields.
//
// The standard root basicAuth (enabled) boolean is intentionally not carried:
// the plugin ignores its stored value and derives BasicAuthEnabled from
// authMethod (pkg/models/settings.go:87-96). jsonData.enableSecureSocksProxy is
// intentionally omitted (AGENTS.md exclusion); json unmarshal silently ignores
// it on parse.
type Config struct {
	// Root-level fields read by the backend (json:"-" — not jsonData).
	URL           string `json:"-"`
	BasicAuthUser string `json:"-"`

	// jsonData fields.
	AuthMethod          AuthMethod `json:"authMethod,omitempty"`
	OAuthClientID       string     `json:"oauthClientID,omitempty"`
	OAuthEnabled        bool       `json:"oauthEnabled,omitempty"` // legacy; predates AuthMethod
	UseSysTables        bool       `json:"useSysTables,omitempty"`
	QueryTimeoutSeconds int        `json:"queryTimeoutSeconds,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, oauthClientSecret).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config, mirroring
// the plugin's LoadSettings (pkg/models/settings.go:40-105): copy the root URL
// and basicAuthUser, unmarshal jsonData, copy the decrypted secrets by known
// key, resolve the auth method and default the query timeout, then validate.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults -> Validate.
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on a Config they assemble themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading servicenow datasource config")

	cfg := Config{
		URL:                     settings.URL,
		BasicAuthUser:           settings.BasicAuthUser,
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// The plugin's LoadSettings unmarshals config.JSONData unconditionally
	// (pkg/models/settings.go:52); Grafana always sends at least "{}". We guard
	// against truly-empty bytes so an empty payload defaults cleanly and fails at
	// Validate (missing URL) rather than at parse.
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
		logger.Error("servicenow datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("servicenow datasource config loaded",
		"authMethod", cfg.AuthMethod,
		"hasURL", cfg.URL != "",
		"hasUser", cfg.BasicAuthUser != "",
		"useSysTables", cfg.UseSysTables,
		"queryTimeoutSeconds", cfg.QueryTimeoutSeconds,
	)
	return cfg, nil
}

// ApplyDefaults resolves the effective auth method (mirroring GetAuthMethod,
// including the legacy oauthEnabled fallback) and defaults the query timeout —
// the same normalization the plugin's LoadSettings performs on every load
// (pkg/models/settings.go:73-86, pkg/models/auth_method.go:12-22). Never
// blanket-apply every schema default — that would clobber intentional zero
// values.
//
// Curated defaults:
//   - AuthMethod: GetAuthMethod(AuthMethod, OAuthEnabled) — empty resolves to
//     basicAuth; a legacy oauthEnabled=true resolves to serviceNowOAuth (even
//     over an explicit authMethod="basicAuth", an upstream quirk). An
//     unrecognized authMethod also resolves to basicAuth, matching upstream
//     GetAuthMethod.
//   - QueryTimeoutSeconds: any value < 1 becomes DefaultQueryTimeoutSeconds (30).
func (c *Config) ApplyDefaults() {
	c.AuthMethod = GetAuthMethod(c.AuthMethod, c.OAuthEnabled)
	if c.QueryTimeoutSeconds < 1 {
		c.QueryTimeoutSeconds = DefaultQueryTimeoutSeconds
	}
}

// Validate checks the runtime contract required for a working ServiceNow
// datasource. It mirrors the plugin's IsValid (pkg/models/settings.go:108-134)
// for the URL and Basic-auth requirements and additionally enforces the OAuth2
// password-grant contract the client actually needs
// (pkg/httputil/auth.go:23-36): serviceNowOAuth requires the account
// username/password AND the OAuth client id/secret. Errors are joined so callers
// see every problem at once.
//
// This is intentionally stricter than upstream IsValid, which is inconsistent:
// without custom headers IsValid only checks username+password (never
// oauthClientID) and with custom headers it only checks oauthClientID. See the
// README discrepancies section.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("invalid server name: URL (root.url) is required"))
	} else if _, err := url.Parse(c.URL); err != nil {
		errs = append(errs, fmt.Errorf("invalid server name: %w", err))
	}

	switch c.AuthMethod {
	case AuthMethodBasicAuth:
		if c.BasicAuthUser == "" {
			errs = append(errs, errors.New("invalid username: basicAuthUser (root.basicAuthUser) is required for basicAuth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword] == "" {
			errs = append(errs, errors.New("invalid password: basicAuthPassword (secureJsonData) is required for basicAuth"))
		}
	case AuthMethodServiceNowOAuth:
		// ServiceNow OAuth is an OAuth2 resource-owner password grant: it needs
		// the account username/password AND the OAuth client credentials.
		if c.BasicAuthUser == "" {
			errs = append(errs, errors.New("invalid username: basicAuthUser (root.basicAuthUser) is required for serviceNowOAuth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword] == "" {
			errs = append(errs, errors.New("invalid password: basicAuthPassword (secureJsonData) is required for serviceNowOAuth"))
		}
		if c.OAuthClientID == "" {
			errs = append(errs, errors.New("invalid oauth configuration: oauthClientID (jsonData) is required for serviceNowOAuth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyOAuthClientSecret] == "" {
			errs = append(errs, errors.New("invalid oauth configuration: oauthClientSecret (secureJsonData) is required for serviceNowOAuth"))
		}
	case "":
		errs = append(errs, errors.New("authentication method not set"))
	default:
		errs = append(errs, fmt.Errorf("invalid authentication method: %s", c.AuthMethod))
	}

	return errors.Join(errs...)
}
