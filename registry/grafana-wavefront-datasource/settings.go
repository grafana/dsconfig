// Package wavefrontdatasource contains the configuration models for the
// Wavefront (VMware Aria Operations for Applications) datasource plugin
// (plugin id: grafana-wavefront-datasource).
package wavefrontdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream plugin).
const PluginID = "grafana-wavefront-datasource"

// DefaultRequestTimeout is the per-request timeout (seconds) applied when
// jsonData.requestTimeout is unset or non-positive. Mirrors
// pkg/models/constant.go:4 (defaultRequestTimeout), which the plugin seeds on
// its Settings struct before unmarshal (pkg/models/settings.go:23-24) and
// re-applies in getHTTPClient for any value <= 0 (pkg/datasource/client.go:20-22).
const DefaultRequestTimeout int64 = 30

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the Wavefront API token, sent as
	// "Authorization: Bearer <token>" on every outgoing request
	// (pkg/datasource/datasource.go:45-47; pkg/wavefront/client.go:39-44).
	SecureJsonDataKeyToken SecureJsonDataKey = "token"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads when loading
// settings (pkg/models/settings.go:29-31). The Wavefront datasource declares
// exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
}

// Config is the fully loaded configuration of a Wavefront datasource instance.
//
// The jsonData fields mirror the plugin's upstream backend Settings struct
// (pkg/models/settings.go:14-19) verbatim for the stored keys — same json
// tags, same int64 timeout type. The upstream `Token` field (populated from
// the decrypted secret) is replaced here by DecryptedSecureJSONData, and the
// upstream `ProxyOptions *proxy.Options` field is intentionally omitted: it is
// an SDK-derived runtime value built from config.HTTPClientOptions(ctx)
// (pkg/models/settings.go:39-43), not a stored setting.
//
// Root-level datasource fields (settings.URL, BasicAuth, etc.) are NOT carried
// because the Wavefront plugin never reads them: pkg/models/settings.go
// unmarshals only config.JSONData, and pkg/datasource/datasource.go builds the
// client from jsonData.url + the decrypted token. jsonData.enableSecureSocksProxy
// is also not modeled — upstream's Settings struct does not carry it, the plugin
// never inspects it by name, and json unmarshal silently ignores it on parse.
type Config struct {
	URL            string `json:"url"`
	RequestTimeout int64  `json:"requestTimeout"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (token).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config, mirroring
// pkg/models/settings.go:22-45 (LoadSettings): unconditionally unmarshal
// jsonData (empty or malformed JSONData is a parse error, matching upstream's
// json.Unmarshal at :26-28), copy the decrypted token by known key, then apply
// defaults and validate.
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

	logger.Debug("loading wavefront datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings (pkg/models/settings.go:26-28) calls
	// json.Unmarshal on config.JSONData unconditionally and returns an
	// unmarshal error when the bytes are empty or malformed. Mirror that —
	// empty JSONData is a parse error.
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		logger.Error("failed to parse jsonData", "err", err)
		return cfg, fmt.Errorf("parse jsonData: %w", err)
	}

	// Copy the decrypted secret by known key name
	// (pkg/models/settings.go:29-31).
	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("wavefront datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("wavefront datasource config loaded",
		"url", cfg.URL,
		"requestTimeout", cfg.RequestTimeout,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own code applies. Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// Curated defaults:
//   - RequestTimeout: DefaultRequestTimeout (30) when <= 0. Mirrors the
//     upstream seed (pkg/models/settings.go:23-24) plus the getHTTPClient
//     coercion of any value <= 0 to 30 (pkg/datasource/client.go:20-22). A
//     missing key, an explicit null, and an explicit 0 all resolve to 30.
//
// URL has no default — the backend errors on an empty url rather than
// supplying one (the editor's https://try.wavefront.com pre-fill is a
// frontend convenience only), so it is enforced by Validate instead.
func (c *Config) ApplyDefaults() {
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = DefaultRequestTimeout
	}
}

// Validate checks the runtime contract that the plugin requires
// (pkg/models/settings.go:32-37, LoadSettings; mirrored in the health check
// pkg/datasource/handler_healthcheck.go:100-111, CheckSettings). Errors are
// joined so callers see every problem at once.
//
// Contracts enforced:
//   - URL must be non-empty. Upstream returns errors.New("invalid url").
//   - token (secureJsonData.token) must be non-empty. Upstream returns
//     errors.New("invalid credentials").
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("wavefront API URL (jsonData.url) is required"))
	}
	if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
		errs = append(errs, errors.New("wavefront token (secureJsonData.token) is required"))
	}

	return errors.Join(errs...)
}
