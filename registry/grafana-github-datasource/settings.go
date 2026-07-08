// Package githubdatasource contains the configuration models for the
// GitHub datasource plugin (grafana-github-datasource).
package githubdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-github-datasource"

// AuthType is the authentication type selected in the configuration editor
// ("Authentication Type"). Stored in jsonData.selectedAuthType.
type AuthType string

const (
	// AuthTypePAT authenticates with a GitHub fine grained personal access token.
	AuthTypePAT AuthType = "personal-access-token"
	// AuthTypeGithubApp authenticates as a GitHub App installation.
	AuthTypeGithubApp AuthType = "github-app"
)

// LicenseType is the GitHub license type selected in the configuration editor
// ("GitHub License Type"). Stored in jsonData.githubPlan.
type LicenseType string

const (
	// LicenseTypeBasic is the "Free, Pro & Team" GitHub plan.
	LicenseTypeBasic LicenseType = "github-basic"
	// LicenseTypeEnterpriseCloud is the "Enterprise Cloud" GitHub plan.
	LicenseTypeEnterpriseCloud LicenseType = "github-enterprise-cloud"
	// LicenseTypeEnterpriseServer is the "Enterprise Server" (on-prem) GitHub plan.
	LicenseTypeEnterpriseServer LicenseType = "github-enterprise-server"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAccessToken is set when the user authenticates with a
	// GitHub personal access token.
	SecureJsonDataKeyAccessToken SecureJsonDataKey = "accessToken"
	// SecureJsonDataKeyPrivateKey is set when the user authenticates as a
	// GitHub App installation (PEM-encoded RSA private key).
	SecureJsonDataKeyPrivateKey SecureJsonDataKey = "privateKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessToken,
	SecureJsonDataKeyPrivateKey,
}

// Config is the fully loaded configuration of a GitHub datasource instance.
// The plugin stores nothing plugin-specific at the root level (url, basicAuth,
// etc. are unused), so only the parsed jsonData fields and decrypted secure data (DecryptedSecureJSONData)
// live here. Callers reach everything directly as cfg.SelectedAuthType,
// cfg.GitHubURL, etc. Enumerate configured secrets by iterating DecryptedSecureJSONData.
type Config struct {
	// jsonData fields, matching the plugin's pkg/models/settings.go Settings
	// shape verbatim, including json tags. AppId/InstallationId are kept as
	// strings (normalized from the legacy string-or-number form by the custom
	// UnmarshalJSON below), with AppIdInt64/InstallationIdInt64 carrying the
	// parsed numeric values used by the backend.
	GithubPlan          LicenseType `json:"githubPlan,omitempty"`
	GitHubURL           string      `json:"githubUrl,omitempty"`
	SelectedAuthType    AuthType    `json:"selectedAuthType,omitempty"`
	AppId               string      `json:"appId,omitempty"`
	AppIdInt64          int64       `json:"-"`
	InstallationId      string      `json:"installationId,omitempty"`
	InstallationIdInt64 int64       `json:"-"`
	CachingEnabled      bool        `json:"cachingEnabled,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (accessToken, privateKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// UnmarshalJSON decodes jsonData into Config while tolerating appId and
// installationId being stored either as JSON strings (e.g. "1111") or, for
// legacy configs, as JSON numbers (e.g. 1111). Both are normalized to their
// string form. Mirrors the upstream Settings.UnmarshalJSON in
// pkg/models/settings.go.
func (c *Config) UnmarshalJSON(data []byte) error {
	type alias Config
	aux := struct {
		AppId          json.RawMessage `json:"appId,omitempty"`
		InstallationId json.RawMessage `json:"installationId,omitempty"`
		*alias
	}{alias: (*alias)(c)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.AppId = rawMessageToString(aux.AppId)
	c.InstallationId = rawMessageToString(aux.InstallationId)
	return nil
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/models/settings.go), including the legacy
// fallback: instances created before selectedAuthType was introduced store
// only an accessToken, so the auth type defaults to personal-access-token
// when an accessToken is present and no auth type is set.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse (unmarshal jsonData +
// legacy fallbacks + conditional int64 parsing under github-app auth), then
// (*Config).ApplyDefaults for curated editor-parity defaults, then
// (Config).Validate to enforce the plugin's runtime contract. This is the
// intended shape for upstream to sync to. Callers that need each phase
// individually can invoke ApplyDefaults and Validate directly on the returned
// Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading github datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Only parse the numeric IDs when the user has selected github-app auth,
	// mirroring the upstream LoadSettings behavior (pkg/models/settings.go).
	if cfg.SelectedAuthType == AuthTypeGithubApp {
		var err error
		if cfg.AppIdInt64, err = stringToInt64(cfg.AppId, "app id"); err != nil {
			logger.Error("failed to parse app id", "err", err)
			return cfg, err
		}
		if cfg.InstallationIdInt64, err = stringToInt64(cfg.InstallationId, "installation id"); err != nil {
			logger.Error("failed to parse installation id", "err", err)
			return cfg, err
		}
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	if cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken] != "" && cfg.SelectedAuthType == "" {
		logger.Info("no selectedAuthType set but accessToken present; defaulting to personal-access-token (legacy config)")
		cfg.SelectedAuthType = AuthTypePAT
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("github datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("github datasource config loaded",
		"authType", cfg.SelectedAuthType,
		"githubPlan", cfg.GithubPlan,
		"hasGithubUrl", cfg.GitHubURL != "",
	)
	return cfg, nil
}

func rawMessageToString(r json.RawMessage) string {
	return strings.Trim(string(r), `"`)
}

func stringToInt64(v string, m string) (int64, error) {
	out, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing %s", m)
	}
	return out, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the config editor writes into jsonData for a fresh datasource
// (mirroring the "" example in SettingsExamples). It is intentionally NOT
// called by LoadConfig — LoadConfig mirrors the upstream LoadSettings verbatim
// — so callers opt in when they want editor-parity defaults for programmatic
// or partial settings.
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - SelectedAuthType → AuthTypePAT
//   - GithubPlan       → LicenseTypeBasic
func (c *Config) ApplyDefaults() {
	if c.SelectedAuthType == "" {
		c.SelectedAuthType = AuthTypePAT
	}
	if c.GithubPlan == "" {
		c.GithubPlan = LicenseTypeBasic
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract: an auth method is selected and its required inputs are present,
// and an Enterprise Server plan has a githubUrl. It does not mutate the
// Config. Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	switch c.SelectedAuthType {
	case AuthTypePAT:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken] == "" {
			errs = append(errs, errors.New("access token is required for personal-access-token auth"))
		}
	case AuthTypeGithubApp:
		if c.AppIdInt64 <= 0 {
			errs = append(errs, errors.New("appId is required for github-app auth"))
		}
		if c.InstallationIdInt64 <= 0 {
			errs = append(errs, errors.New("installationId is required for github-app auth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPrivateKey] == "" {
			errs = append(errs, errors.New("privateKey is required for github-app auth"))
		}
	case "":
		errs = append(errs, errors.New("selectedAuthType is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown selectedAuthType %q", c.SelectedAuthType))
	}

	if c.GithubPlan == LicenseTypeEnterpriseServer && c.GitHubURL == "" {
		errs = append(errs, errors.New("githubUrl is required for Enterprise Server plan"))
	}

	return errors.Join(errs...)
}
