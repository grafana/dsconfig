// Package jenkinsdatasource contains the configuration models for the
// Jenkins datasource plugin (plugin id: grafana-jenkins-datasource).
package jenkinsdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id
// field (src/plugin.json:5 in the upstream repo).
const PluginID = "grafana-jenkins-datasource"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the Jenkins user password (or API
	// token). It is passed straight into httpclient.BasicAuthOptions
	// (pkg/plugin/datasource.go:66-71), which in turn sets the
	// "Authorization: Basic ..." header on every outgoing request
	// (pkg/jenkins/client.go:498-500). Only used when
	// jsonData.username != "".
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
)

// SecureJsonDataConfig lists the secret key names stored in
// secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin. The
// Jenkins datasource declares exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
}

// Config is the fully loaded configuration of a Jenkins datasource
// instance.
//
// Fields mirror the plugin's own backend Settings struct
// (pkg/plugin/settings.go:10-14) verbatim — same field names, same json
// tags — except that the upstream unexported `Password` field is
// replaced here by DecryptedSecureJSONData, which is populated from
// backend.DataSourceInstanceSettings.DecryptedSecureJSONData in
// LoadConfig.
//
// Root-level datasource fields (settings.URL, BasicAuth, etc.) are NOT
// carried on Config because the Jenkins plugin never reads them:
// pkg/plugin/datasource.go builds its client from jsonData.url +
// jsonData.username + the decrypted secure password, and
// pkg/plugin/settings.go only unmarshals settings.JSONData.
type Config struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (password).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config,
// mirroring pkg/plugin/settings.go:16-28 (LoadSettings) verbatim:
// unmarshal jsonData (empty JSONData is a parse error, matching
// upstream), copy decrypted secrets by known key, then validate.
//
// Upstream's LoadSettings returns a DownstreamError when the URL is
// empty; here we surface the same policy as a Validate error so
// callers see every problem at once.
//
// ctx is used to derive a contextual logger via
// backend.Logger.FromContext so log lines carry the request/plugin
// context that Grafana injects.
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

	logger.Debug("loading jenkins datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings (pkg/plugin/settings.go:18-21) calls
	// json.Unmarshal on settings.JSONData unconditionally and returns
	// "could not unmarshal plugin settings json" when the bytes are
	// empty or malformed. Mirror that behavior — empty JSONData is a
	// parse error.
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
		logger.Error("jenkins datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("jenkins datasource config loaded",
		"url", cfg.URL,
		"hasUsername", cfg.Username != "",
		"hasPassword", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword] != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own code applies. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// The Jenkins plugin has no upstream-side defaults: pkg/plugin/settings.go
// applies no ApplyDefaults-style logic (URL empty is a hard error, and
// username / password are allowed to be empty for anonymous access), so
// ApplyDefaults is intentionally a no-op today. It is kept as a stable
// extension point and to match the shape callers expect from every
// entry (LoadConfig -> ApplyDefaults -> Validate).
func (c *Config) ApplyDefaults() {}

// Validate checks the runtime contract that the plugin requires
// (pkg/plugin/settings.go:23-25). Errors are joined so callers see
// every problem at once.
//
// Contracts enforced:
//   - URL must be non-empty. Upstream returns
//     DownstreamError("URL is missing") when it is empty.
//
// Username and password are intentionally NOT required: the backend
// wires BasicAuth only when username != "" (pkg/plugin/datasource.go:66-71),
// which means an empty username is a supported "anonymous access"
// configuration, and password is only consulted when a username is
// present.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("jenkins URL (jsonData.url) is required"))
	}

	return errors.Join(errs...)
}
