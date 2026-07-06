// Package jiradatasource contains the configuration models for the Jira
// datasource plugin (plugin id: grafana-jira-datasource).
package jiradatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream plugin).
const PluginID = "grafana-jira-datasource"

// AuthMethod is the authentication method selected in the configuration editor
// ("Authentication method" radio). Stored in jsonData.authMethod. Mirrors the
// plugin's pkg/models/auth_method.go:3-8 AuthMethod alias and constants.
type AuthMethod string

const (
	// AuthMethodBasicAuth authenticates with an Atlassian account email + API
	// token (HTTP Basic), or a lone API/personal-access token as a Bearer token
	// when no user is set ("Basic Auth" in the editor). It is the default.
	AuthMethodBasicAuth AuthMethod = "basicAuth"
	// AuthMethodOAuth2 authenticates with the OAuth 2.0 client-credentials grant
	// against Atlassian ("OAuth 2.0 — Service Account" in the editor). Jira Cloud
	// only.
	AuthMethodOAuth2 AuthMethod = "oauth2"
)

// ResolveAuthMethod mirrors the plugin's GetAuthMethod
// (pkg/models/auth_method.go:11-15): only the exact string "oauth2" selects
// OAuth 2.0; every other value (including empty and unknown strings) resolves to
// basicAuth.
func ResolveAuthMethod(v string) AuthMethod {
	if AuthMethod(v) == AuthMethodOAuth2 {
		return AuthMethodOAuth2
	}
	return AuthMethodBasicAuth
}

// Hosting is the provider selected in the configuration editor ("Provider"
// radio). Stored in jsonData.hosting. Mirrors the frontend Provider enum
// (src/types.ts:57-60). Upstream types it as a plain string
// (pkg/models/settings.go:17); a named type is used here for clarity.
type Hosting string

const (
	// HostingCloud is Atlassian-hosted Jira Cloud (REST API v3). It is the
	// default.
	HostingCloud Hosting = "cloud"
	// HostingServer is self-hosted Jira Data Center / Jira Server (REST API v2).
	HostingServer Hosting = "server"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the Jira API token / personal access token
	// (Basic Auth). Sent as the HTTP Basic password, or as a Bearer token when
	// no user email is set.
	SecureJsonDataKeyToken SecureJsonDataKey = "token"
	// SecureJsonDataKeyOAuthClientSecret is the OAuth 2.0 client secret for the
	// Jira service account (OAuth 2.0 auth).
	SecureJsonDataKeyOAuthClientSecret SecureJsonDataKey = "oauthClientSecret"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads (LoadSettings copies
// them from DecryptedSecureJSONData by auth method, pkg/models/settings.go:52-72).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
	SecureJsonDataKeyOAuthClientSecret,
}

// Config is the fully loaded configuration of a Jira datasource instance.
//
// The json-tagged fields mirror the jsonData portion of the plugin's backend
// Settings struct (pkg/models/settings.go:14-21) verbatim: url, user, hosting,
// scopedToken, cloudId, authMethod, oauthClientID. Note that url and user live
// in jsonData (not at the datasource root). The plugin's Settings struct also
// declares Token and OAuthClientSecret (plain fields populated from
// secureSettings) plus HttpClientOptions; the two secrets are modeled here in
// DecryptedSecureJSONData, and HttpClientOptions is not configuration storage,
// so neither is carried as a struct field (see the README discrepancies section).
//
// Root-level datasource fields (settings.URL, BasicAuth, etc.) are NOT carried
// on Config because the plugin never reads them: pkg/plugin.go:171-228 builds
// the client from jsonData + decrypted secrets only.
//
// jsonData.enableSecureSocksProxy is intentionally omitted (AGENTS.md
// exclusion); json unmarshal silently ignores it on parse.
type Config struct {
	URL           string     `json:"url"`
	User          string     `json:"user"`
	Hosting       Hosting    `json:"hosting"`
	ScopedToken   bool       `json:"scopedToken"`
	CloudId       string     `json:"cloudId"`
	AuthMethod    AuthMethod `json:"authMethod"`
	OAuthClientID string     `json:"oauthClientID"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (token, oauthClientSecret).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config, mirroring
// the plugin's LoadSettings (pkg/models/settings.go:31-82): unmarshal jsonData
// (empty JSONData is a parse error, matching upstream json.Unmarshal), copy the
// decrypted secrets by known key, then normalize (ApplyDefaults) and validate.
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

	logger.Debug("loading jira datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings (pkg/models/settings.go:33) calls json.Unmarshal on
	// config.JSONData unconditionally and returns the error when the bytes are
	// empty or malformed. We mirror that behavior — empty JSONData is a parse
	// error.
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
		logger.Error("jira datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("jira datasource config loaded",
		"authMethod", cfg.AuthMethod,
		"hosting", cfg.Hosting,
		"hasUser", cfg.User != "",
		"scopedToken", cfg.ScopedToken,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of fields with the same defaults/normalizations
// the plugin's own LoadSettings applies on every load. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// Curated normalizations (mirroring pkg/models/settings.go:37,43-45 and
// pkg/models/auth_method.go:11-15):
//   - AuthMethod: GetAuthMethod resolves an empty or unknown value to basicAuth;
//     only "oauth2" stays oauth2.
//   - URL: prepend "https://" when the URL is non-empty and has no http(s)://
//     scheme. Guarded on non-empty so an unset URL stays empty for Validate to
//     reject (upstream checks URL-empty before this normalization).
func (c *Config) ApplyDefaults() {
	c.AuthMethod = ResolveAuthMethod(string(c.AuthMethod))
	if c.URL != "" && !strings.HasPrefix(c.URL, "https://") && !strings.HasPrefix(c.URL, "http://") {
		c.URL = "https://" + c.URL
	}
}

// Validate checks the runtime contract the plugin enforces before building the
// client (pkg/models/settings.go:39-72 and the getEndpoint contract in
// pkg/plugin.go:214-228). Errors are joined so callers see every problem at once
// (upstream returns only the first).
//
// Contracts enforced:
//   - url must be non-empty ("URL is missing").
//   - oauth2: oauthClientID, secureJsonData.oauthClientSecret and cloudId must
//     all be non-empty.
//   - basicAuth (and, matching upstream's switch default, any non-oauth2 value):
//     secureJsonData.token must be non-empty; and when scopedToken is on, cloudId
//     must be non-empty (scoped tokens route through the Atlassian gateway).
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("URL is missing"))
	}

	switch c.AuthMethod {
	case AuthMethodOAuth2:
		if c.OAuthClientID == "" {
			errs = append(errs, errors.New("OAuth client ID is missing"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyOAuthClientSecret] == "" {
			errs = append(errs, errors.New("OAuth client secret is missing"))
		}
		if c.CloudId == "" {
			errs = append(errs, errors.New("Cloud ID is required for OAuth 2.0 authentication"))
		}
	default:
		// basicAuth (and any non-oauth2 value, matching the upstream switch
		// default in pkg/models/settings.go:65-72).
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, errors.New("token is missing"))
		}
		if c.ScopedToken && c.CloudId == "" {
			errs = append(errs, errors.New("cloud ID is required for scoped token"))
		}
	}

	return errors.Join(errs...)
}
