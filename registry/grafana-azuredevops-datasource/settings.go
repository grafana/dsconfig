// Package azuredevopsdatasource contains the configuration models for the
// Azure DevOps datasource plugin (plugin id: grafana-azuredevops-datasource).
package azuredevopsdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream plugin) and the backend PluginID constant
// (pkg/plugin/constants.go:6).
const PluginID = "grafana-azuredevops-datasource"

// DefaultProjectsLimit is the fallback applied to jsonData.projectsLimit when it
// is unset or < 1. It mirrors both the editor's initial value
// (src/editors/AzDoConfigEditor.tsx:28) and the backend coercion
// (pkg/plugin/settings.go:46-48).
const DefaultProjectsLimit = 100

// AuthType is the authentication type stored in jsonData.authType. The plugin
// currently supports exactly one method; the frontend type pins it to the
// literal "patToken" (src/types.ts:6) and the editor stamps that value on every
// jsonData write (src/editors/AzDoConfigEditor.tsx:38). The backend declares the
// field but does not branch on it ("Not in use yet", pkg/plugin/settings.go:11).
type AuthType string

const (
	// AuthTypePAT authenticates with an Azure DevOps personal access token. It
	// is the only supported (and default) authentication type.
	AuthTypePAT AuthType = "patToken"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPATToken is the Azure DevOps personal access token. It is
	// copied from DecryptedSecureJSONData["patToken"] (pkg/plugin/settings.go:39-41)
	// and used as the HTTP Basic password by azuredevops.NewPatConnection /
	// CreateBasicAuthHeaderValue (pkg/plugin/plugin.go:74,77).
	SecureJsonDataKeyPATToken SecureJsonDataKey = "patToken"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads. The Azure DevOps
// datasource declares exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPATToken,
}

// Config is the fully loaded configuration of an Azure DevOps datasource
// instance.
//
// The json-tagged fields mirror the jsonData portion of the plugin's backend
// AzDoConfig struct (pkg/plugin/settings.go:10-17) verbatim: authType, url,
// projectsLimit, username. Note that url lives in jsonData (not at the
// datasource root).
//
// The PAT lives in secureJsonData and is modeled in DecryptedSecureJSONData
// rather than as the upstream struct's `PATToken string json:"-"` field. Root-
// level datasource fields (settings.URL, BasicAuth, etc.) are NOT carried
// because the plugin never reads them: pkg/plugin/settings.go:29-51 builds the
// config from jsonData + the decrypted patToken only.
//
// jsonData.enableSecureSocksProxy (the upstream struct's ProxyEnabled,
// pkg/plugin/settings.go:15) is intentionally omitted (AGENTS.md exclusion);
// json unmarshal silently ignores it on parse.
type Config struct {
	URL           string   `json:"url"`
	AuthType      AuthType `json:"authType"`
	ProjectsLimit int      `json:"projectsLimit"`
	Username      string   `json:"username"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (patToken).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config, mirroring
// the plugin's GetSettings (pkg/plugin/settings.go:29-51): unmarshal jsonData
// (empty JSONData is a parse error, matching upstream json.Unmarshal), copy the
// decrypted patToken, then normalize (ApplyDefaults) and validate.
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

	logger.Debug("loading azure devops datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream GetSettings (pkg/plugin/settings.go:31) calls json.Unmarshal on
	// s.JSONData unconditionally and returns the error when the bytes are empty
	// or malformed. Mirror that behavior — Grafana always sends at least "{}",
	// so this only rejects a truly-empty payload.
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
		logger.Error("azure devops datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("azure devops datasource config loaded",
		"url", cfg.URL,
		"authType", cfg.AuthType,
		"projectsLimit", cfg.ProjectsLimit,
		"hasUsername", cfg.Username != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same defaults
// the editor writes and the backend applies. Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// Curated defaults:
//   - AuthType: AuthTypePAT ("patToken") when empty — mirrors the editor stamp
//     (src/editors/AzDoConfigEditor.tsx:38) and the single supported method.
//   - ProjectsLimit: DefaultProjectsLimit (100) when < 1 — mirrors the editor's
//     initial value (src/editors/AzDoConfigEditor.tsx:28) and the backend
//     coercion (pkg/plugin/settings.go:46-48).
//
// The patToken secret has no default — the plugin errors out when it is empty.
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AuthTypePAT
	}
	if c.ProjectsLimit < 1 {
		c.ProjectsLimit = DefaultProjectsLimit
	}
}

// Validate checks the runtime contract the plugin enforces before building the
// client (pkg/plugin/settings.go:19-27,35-45): a non-empty url
// (ErrorInvalidURL, "invalid URL") and a non-empty patToken
// (ErrorInvalidPATToken, "invalid PAT"). Errors are joined so callers see every
// problem at once (upstream returns only the first).
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("url (jsonData.url) is required"))
	}
	if c.DecryptedSecureJSONData[SecureJsonDataKeyPATToken] == "" {
		errs = append(errs, errors.New("personal access token (secureJsonData.patToken) is required"))
	}

	return errors.Join(errs...)
}
