// Package helloworlddatasource contains the configuration models for the
// Grafana Hello World datasource plugin (id: grafana-helloworld-datasource).
//
// Hello World is a minimal sample/template datasource: it has NO configuration
// surface. The frontend config editor renders static text and persists
// nothing, and the backend never reads instance settings. This package
// therefore models an empty configuration — Config carries no jsonData or root
// fields. The lone secureJsonData key (apiKey) is a placeholder required by the
// dsconfig validator (which rejects an empty fields array) and by the shared
// conformance suite (which requires at least one secureJsonData key); the
// plugin never reads it. See the entry README for the full rationale.
package helloworlddatasource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream repo). It also matches the backend's own
// PluginId constant (pkg/main.go:13). Hello World declares no aliasIDs.
const PluginID = "grafana-helloworld-datasource"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAPIKey is a PLACEHOLDER secret. The Hello World plugin
	// defines no secrets (src/module.tsx:14 declares `SecureConfig = {}`) and
	// never reads this value. It exists only so the schema can declare at
	// least one field / secureJsonData key (see the package doc and README).
	SecureJsonDataKeyAPIKey SecureJsonDataKey = "apiKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys modeled by this entry. The list
// contains a single placeholder (apiKey); the Hello World plugin itself
// defines and reads no secrets.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAPIKey,
}

// Config is the fully loaded configuration of a Hello World datasource
// instance.
//
// It is intentionally empty apart from DecryptedSecureJSONData: the plugin has
// no jsonData fields (src/module.tsx:13 declares `Config = {} &
// DataSourceJsonData`) and reads no root-level datasource fields (pkg/main.go's
// instance factory ignores backend.DataSourceInstanceSettings entirely). There
// is no upstream pkg/models/settings.go or LoadSettings to mirror.
//
// DecryptedSecureJSONData is populated from
// backend.DataSourceInstanceSettings.DecryptedSecureJSONData for the modeled
// (placeholder) secret keys, but the plugin never consumes it.
type Config struct {
	// DecryptedSecureJSONData holds the decrypted secure values by key. Only
	// the placeholder apiKey is modeled; the plugin reads none of it.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config and returns
// a fully-defaulted, validated Config. For Hello World there is nothing to
// parse (no jsonData fields) and nothing to validate (no required inputs); the
// method still runs the standard three-phase flow — parse -> ApplyDefaults ->
// Validate — for uniformity with other registry entries and so callers can
// rely on the same contract.
//
// Malformed settings.JSONData is still reported as a parse error (mirroring the
// robustness of sibling entries), even though the Hello World backend itself
// never unmarshals settings.JSONData.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context Grafana injects.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading hello world datasource config")

	cfg := Config{
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
		logger.Error("hello world datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("hello world datasource config loaded")
	return cfg, nil
}

// ApplyDefaults fills in editor-parity defaults for zero-valued fields. The
// Hello World editor persists no fields and writes no defaults, so this is
// intentionally a no-op. Kept exported so callers can compose the standard
// three-phase flow uniformly; the TestApplyDefaults test guards that it
// mutates nothing.
func (c *Config) ApplyDefaults() {
}

// Validate checks the runtime contract the plugin requires. Hello World
// requires nothing — CheckHealth always returns OK and QueryData returns a
// static frame regardless of settings (pkg/main.go:17-38) — so Validate always
// returns nil. Kept exported for uniformity with other entries and so callers
// can validate a Config they assembled themselves.
func (c Config) Validate() error {
	return nil
}
