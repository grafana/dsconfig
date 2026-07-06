// Package falconlogscaledatasource contains the configuration models for the
// Falcon LogScale datasource plugin (id: grafana-falconlogscale-datasource).
package falconlogscaledatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id field.
const PluginID = "grafana-falconlogscale-datasource"

// DataSourceMode is the discriminator between the LogScale product (default)
// and CrowdStrike NGSIEM (which only supports OAuth2 client credentials and
// automatically pins the default repository to `search-all`). Stored in
// jsonData.mode. Mirrors DataSourceMode in src/types.ts:5-8.
type DataSourceMode string

const (
	// DataSourceModeLogScale is the default mode.
	DataSourceModeLogScale DataSourceMode = "LogScale"
	// DataSourceModeNGSIEM is the CrowdStrike NGSIEM mode.
	DataSourceModeNGSIEM DataSourceMode = "NGSIEM"
)

// NGSIEMRepos lists the repository / view names allowed when the datasource is
// in NGSIEM mode. Mirrors src/types.ts:52. Kept as a Go slice so callers can
// programmatically validate a defaultRepository value against it.
var NGSIEMRepos = []string{"search-all", "investigate_view", "third-party"}

// AuthMethod identifies the selected authentication method as exposed by the
// config editor's Auth component. It is a virtual discriminator — there is no
// upstream jsonData field named `authMethod`. See virtual_authMethod in
// dsconfig.json for the multi-field storage effects it drives.
type AuthMethod string

const (
	// AuthMethodToken selects LogScale personal-token auth (secureJsonData.accessToken).
	AuthMethodToken AuthMethod = "custom-token"
	// AuthMethodOAuth2ClientCredentials selects OAuth2 client-credentials auth
	// (jsonData.oauth2 + jsonData.oauth2ClientId + secureJsonData.oauth2ClientSecret).
	AuthMethodOAuth2ClientCredentials AuthMethod = "custom-oauth-client-secret"
	// AuthMethodBasicAuth selects HTTP Basic auth (root.basicAuth + root.basicAuthUser +
	// secureJsonData.basicAuthPassword).
	AuthMethodBasicAuth AuthMethod = "BasicAuth"
	// AuthMethodOAuthForward selects "Forward OAuth Identity" (jsonData.oauthPassThru=true).
	AuthMethodOAuthForward AuthMethod = "OAuthForward"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAccessToken is the LogScale personal token, set when
	// jsonData.authenticateWithToken is true.
	SecureJsonDataKeyAccessToken SecureJsonDataKey = "accessToken"
	// SecureJsonDataKeyOAuth2ClientSecret is the OAuth2 client secret, set when
	// jsonData.oauth2 is true.
	SecureJsonDataKeyOAuth2ClientSecret SecureJsonDataKey = "oauth2ClientSecret"
	// SecureJsonDataKeyBasicAuthPassword is the HTTP Basic-auth password, set
	// when root.basicAuth is true.
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessToken,
	SecureJsonDataKeyOAuth2ClientSecret,
	SecureJsonDataKeyBasicAuthPassword,
}

// DataLinkConfig mirrors the frontend DataLinkConfig type
// (src/components/DataLinks/types.ts:1-7) that the config editor writes into
// jsonData.dataLinks and the frontend result transformer consumes at
// src/logs.ts. The Falcon LogScale backend never reads this field.
type DataLinkConfig struct {
	Field         string `json:"field"`
	Label         string `json:"label"`
	MatcherRegex  string `json:"matcherRegex"`
	URL           string `json:"url"`
	DatasourceUID string `json:"datasourceUid,omitempty"`
}

// Config is the fully loaded configuration of a Falcon LogScale datasource
// instance.
//
// The plugin's upstream Settings struct (pkg/plugin/settings.go:11-26) is a
// narrower slice of what the config editor writes. This Config carries every
// jsonData field the editor persists (including frontend-only ones — see
// BaseURL and the incremental-querying pair below) so provisioning callers can
// round-trip the full editor state.
//
// Root-level fields (URL, BasicAuth, BasicAuthUser) are tagged json:"-" and
// populated by LoadConfig from backend.DataSourceInstanceSettings, mirroring
// how the plugin's LoadSettings reads them (pkg/plugin/settings.go:38,59).
type Config struct {
	// Root-level fields (not stored in jsonData).
	// URL — the datasource URL; the backend requires it non-empty
	// (pkg/plugin/settings.go:41). BasicAuth / BasicAuthUser are populated by
	// @grafana/plugin-ui's Auth component and consumed by the plugin backend at
	// pkg/plugin/settings.go:59-60.
	URL           string `json:"-"`
	BasicAuth     bool   `json:"-"`
	BasicAuthUser string `json:"-"`

	// jsonData fields — the union of what the editor writes and the backend reads.
	//
	// BaseURL is frontend-only: the LogScale token authentication component
	// (src/components/ConfigEditor/ConfigEditor.tsx:155) snapshots root.url into
	// this field, but the backend unconditionally overwrites it from config.URL
	// before use (pkg/plugin/settings.go:39). The upstream struct declares
	// json:"baseURL" (uppercase URL) — see the entry README for the discrepancy;
	// the frontend writes lowercase-u `baseUrl`, which is what this schema
	// records.
	BaseURL                       string           `json:"baseUrl,omitempty"`
	AuthenticateWithToken         bool             `json:"authenticateWithToken,omitempty"`
	OAuthPassThru                 bool             `json:"oauthPassThru,omitempty"`
	OAuth2                        bool             `json:"oauth2,omitempty"`
	OAuth2ClientID                string           `json:"oauth2ClientId,omitempty"`
	Mode                          DataSourceMode   `json:"mode,omitempty"`
	DefaultRepository             string           `json:"defaultRepository,omitempty"`
	DataLinks                     []DataLinkConfig `json:"dataLinks,omitempty"`
	IncrementalQuerying           bool             `json:"incrementalQuerying,omitempty"`
	IncrementalQueryOverlapWindow string           `json:"incrementalQueryOverlapWindow,omitempty"`
	KeepCookies                   []string         `json:"keepCookies,omitempty"`
	Timeout                       float64          `json:"timeout,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessToken, oauth2ClientSecret, basicAuthPassword).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser) are copied from
// backend.DataSourceInstanceSettings directly; jsonData is unmarshaled from
// settings.JSONData; decrypted secrets are copied by known key name into
// DecryptedSecureJSONData.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Callers that need each phase individually can invoke ApplyDefaults
// and Validate directly on a Config they assemble themselves.
//
// The upstream LoadSettings (pkg/plugin/settings.go:32-62) is narrower — it
// only pulls three secrets and rewrites BaseURL / GraphqlEndpoint /
// RestEndpoint. This entry's LoadConfig is a superset: it parses every field
// the editor writes so provisioning tooling can round-trip a full instance.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading falcon logscale datasource config")

	cfg := Config{
		URL:                     settings.URL,
		BasicAuth:               settings.BasicAuthEnabled,
		BasicAuthUser:           settings.BasicAuthUser,
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
		logger.Error("falcon logscale datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("falcon logscale datasource config loaded",
		"hasURL", cfg.URL != "",
		"mode", cfg.Mode,
		"authenticateWithToken", cfg.AuthenticateWithToken,
		"oauth2", cfg.OAuth2,
		"basicAuth", cfg.BasicAuth,
		"oauthPassThru", cfg.OAuthPassThru,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// The editor defaults `mode` to LogScale on load (ConfigEditor.tsx:59). NGSIEM
// mode additionally auto-sets defaultRepository to `search-all`
// (ConfigEditor.tsx:127-137,193-200); we mirror that side-effect here so a
// provisioned NGSIEM datasource without a repository still ends up in the
// same state as one saved via the UI.
func (c *Config) ApplyDefaults() {
	if c.Mode == "" {
		c.Mode = DataSourceModeLogScale
	}
	if c.Mode == DataSourceModeNGSIEM && c.DefaultRepository == "" {
		c.DefaultRepository = "search-all"
	}
}

// Validate checks the runtime contract the plugin requires. The backend hard-
// fails without a URL (pkg/plugin/settings.go:41). Each selected authentication
// method has its own required inputs; the checks below enforce them so
// misconfigured instances fail loudly at load time rather than at first query.
// NGSIEM mode additionally requires OAuth2 client credentials — no other auth
// method is exposed by the editor when mode is NGSIEM (ConfigEditor.tsx:274-278).
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("URL (root.url) is required"))
	}

	switch c.Mode {
	case "", DataSourceModeLogScale, DataSourceModeNGSIEM:
		// OK; empty accepted because ApplyDefaults sets it to LogScale.
	default:
		errs = append(errs, fmt.Errorf("unknown mode %q (want %q or %q)", c.Mode, DataSourceModeLogScale, DataSourceModeNGSIEM))
	}

	// Exactly one authentication method should be selected. The editor's
	// clearAuthSettings() (ConfigEditor.tsx:61-68) enforces this on every auth-
	// selector change; we mirror it as a validation so provisioned datasources
	// that hand-set multiple flags fail loudly rather than silently choosing one
	// arm inside the client factory (pkg/plugin/plugin.go:52-70).
	selected := 0
	if c.AuthenticateWithToken {
		selected++
	}
	if c.OAuth2 {
		selected++
	}
	if c.OAuthPassThru {
		selected++
	}
	if c.BasicAuth {
		selected++
	}
	if selected > 1 {
		errs = append(errs, errors.New("only one authentication method may be enabled at a time (authenticateWithToken, oauth2, oauthPassThru, basicAuth)"))
	}

	if c.AuthenticateWithToken {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken] == "" {
			errs = append(errs, errors.New("accessToken (secureJsonData) is required when authenticateWithToken is true"))
		}
	}

	if c.OAuth2 {
		if c.OAuth2ClientID == "" {
			errs = append(errs, errors.New("oauth2ClientId (jsonData) is required when oauth2 is true"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyOAuth2ClientSecret] == "" {
			errs = append(errs, errors.New("oauth2ClientSecret (secureJsonData) is required when oauth2 is true"))
		}
	}

	if c.BasicAuth {
		if c.BasicAuthUser == "" {
			errs = append(errs, errors.New("basicAuthUser (root) is required when basicAuth is true"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword] == "" {
			errs = append(errs, errors.New("basicAuthPassword (secureJsonData) is required when basicAuth is true"))
		}
	}

	if c.Mode == DataSourceModeNGSIEM {
		if !c.OAuth2 {
			errs = append(errs, errors.New("NGSIEM mode requires OAuth2 client credentials (jsonData.oauth2=true)"))
		}
	}

	if c.Timeout < 0 {
		errs = append(errs, fmt.Errorf("timeout must be non-negative, got %v", c.Timeout))
	}

	for i, dl := range c.DataLinks {
		if dl.Field == "" {
			errs = append(errs, fmt.Errorf("dataLinks[%d].field is required", i))
		}
		if dl.MatcherRegex == "" {
			errs = append(errs, fmt.Errorf("dataLinks[%d].matcherRegex is required", i))
		}
	}

	return errors.Join(errs...)
}
