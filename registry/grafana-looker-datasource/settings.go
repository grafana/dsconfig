// Package lookerdatasource contains the configuration models for the Grafana
// Looker datasource plugin (id: grafana-looker-datasource).
package lookerdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:3 in the upstream plugin).
const PluginID = "grafana-looker-datasource"

// AuthType is the authentication type stored in jsonData.auth_type. The plugin
// exposes a single value, "client_secret" (Looker API3 credentials). Mirrors
// the upstream AuthType (pkg/models/config.go:22-26).
type AuthType string

const (
	// AuthTypeClientSecret authenticates with Looker API3 credentials
	// (client id + client secret). It is the only auth type the plugin
	// supports and the default the backend applies when auth_type is empty
	// (pkg/models/config.go:25,48-50).
	AuthTypeClientSecret AuthType = "client_secret"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyClientSecret is the Looker API3 client secret. The
	// backend reads it from DecryptedSecureJSONData["client_secret"]
	// (pkg/models/config.go:69).
	SecureJsonDataKeyClientSecret SecureJsonDataKey = "client_secret"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads when loading settings
// (pkg/models/config.go:69).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyClientSecret,
}

// Config is the fully loaded configuration of a Grafana Looker datasource
// instance.
//
// The jsonData fields mirror the upstream backend Config (pkg/models/config.go:14-20)
// verbatim for the storage keys — same json tags. The upstream Config also
// carries ClientSecret (json:"-", the decrypted secret) and HttpClientOptions
// (json:"-", the computed SDK http options); this registry model represents the
// decrypted secret via DecryptedSecureJSONData instead and omits the http
// options, which the plugin computes but never applies to the Looker SDK client
// (pkg/models/config.go:70-71; pkg/looker/client.go:22-38).
//
// The plugin stores nothing plugin-specific at the root level (it reads
// jsonData.base_url, not settings.URL, and never reads named root fields), so no
// root fields live here.
type Config struct {
	// jsonData fields, matching pkg/models/config.go:15-17.
	BaseURL  string   `json:"base_url"`
	AuthType AuthType `json:"auth_type"`
	ClientId string   `json:"client_id"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (client_secret).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the upstream LoadConfig (pkg/models/config.go:58-74): default an empty
// jsonData payload to "{}", unmarshal jsonData, and copy the decrypted
// client_secret from DecryptedSecureJSONData.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Note this composes what the upstream splits: the plugin's own
// LoadConfig calls ApplyDefaults but runs Validate separately, from the health
// check (pkg/handler_healthcheck.go:13). Callers that assemble a Config
// themselves can invoke ApplyDefaults and Validate individually.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading looker datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Default an empty payload to "{}" before unmarshal, mirroring
	// pkg/models/config.go:60-63.
	jsonData := settings.JSONData
	if len(jsonData) == 0 {
		jsonData = []byte(`{}`)
	}
	if err := json.Unmarshal(jsonData, &cfg); err != nil {
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
		logger.Error("looker datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("looker datasource config loaded",
		"authType", cfg.AuthType,
		"hasBaseURL", cfg.BaseURL != "",
	)
	return cfg, nil
}

// ApplyDefaults mirrors the upstream (*Config).ApplyDefaults
// (pkg/models/config.go:47-56): default auth_type to client_secret when empty,
// trim surrounding whitespace from base_url/client_id/client_secret, and strip
// a single trailing slash from base_url. It is kept exported so callers that
// assemble a Config directly can get the same editor/backend parity LoadConfig
// applies.
//
// Curated list (only these fields are touched):
//   - AuthType                    -> AuthTypeClientSecret when empty
//   - BaseURL                     -> trimmed, trailing "/" stripped
//   - ClientId                    -> trimmed
//   - DecryptedSecureJSONData[client_secret] -> trimmed (when present)
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AuthTypeClientSecret
	}
	c.BaseURL = strings.TrimSpace(c.BaseURL)
	c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")
	c.ClientId = strings.TrimSpace(c.ClientId)
	if secret, ok := c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret]; ok {
		c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] = strings.TrimSpace(secret)
	}
}

// Validate checks the runtime contract the plugin enforces in
// (*Config).Validate (pkg/models/config.go:28-45), which the health check runs
// (pkg/handler_healthcheck.go:13): base_url is always required, and for
// client_secret auth both client_id and client_secret are required. Errors are
// joined so callers see every problem at once. The messages match the upstream
// verbatim.
func (c Config) Validate() error {
	var errs []error

	if c.BaseURL == "" {
		errs = append(errs, errors.New("invalid/empty Looker base url"))
	}
	if c.AuthType == AuthTypeClientSecret {
		if c.ClientId == "" {
			errs = append(errs, errors.New("invalid/empty Looker client id"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] == "" {
			errs = append(errs, errors.New("invalid/empty Looker client secret"))
		}
	}

	return errors.Join(errs...)
}
