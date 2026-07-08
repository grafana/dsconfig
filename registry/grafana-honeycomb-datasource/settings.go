// Package honeycombdatasource contains the configuration models for the
// Honeycomb datasource plugin (plugin id: grafana-honeycomb-datasource).
package honeycombdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream plugin).
const PluginID = "grafana-honeycomb-datasource"

// DefaultHoneycombAPIURL is the Honeycomb API base URL applied when
// jsonData.hostname is empty. It mirrors src/types.ts:129
// (defaultConfigOptions.hostname) on the frontend and the backend default
// seeded in LoadSettings (pkg/models/settings.go:28).
const DefaultHoneycombAPIURL = "https://api.honeycomb.io"

// DefaultRetentionLimitDays is the default jsonData.retentionLimit (in days)
// applied when it is zero. It mirrors the backend default seeded in
// LoadSettings (pkg/models/settings.go:29) and the editor's empty-input
// fallback of 7 (src/Views/ConfigEditor.tsx:46).
const DefaultRetentionLimitDays = 7

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAPIKey is the Honeycomb Team API key, sent as the
	// "X-Honeycomb-Team" header on every outgoing request
	// (pkg/httpclient/client.go:39-42).
	SecureJsonDataKeyAPIKey SecureJsonDataKey = "apiKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads when loading
// settings (pkg/models/settings.go:27). The Honeycomb datasource declares
// exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAPIKey,
}

// Config is the fully loaded configuration of a Honeycomb datasource
// instance.
//
// The jsonData fields mirror the plugin's upstream backend Settings struct
// (pkg/models/settings.go:13-19) verbatim — same fields, same json tags,
// same types — except that the upstream APIKey field (json:"-", copied from
// the decrypted secrets) is replaced here by DecryptedSecureJSONData, which
// holds the decrypted secure values by key.
//
// Root-level datasource fields (settings.URL, BasicAuth, etc.) are NOT
// carried on Config because the plugin never reads them: pkg/main.go:40-66
// builds its HTTP client from jsonData.hostname + the decrypted apiKey and
// pkg/models/settings.go only unmarshals settings.JSONData.
type Config struct {
	// jsonData fields, matching pkg/models/settings.go:13-19.
	Env            string `json:"environment"`
	Hostname       string `json:"hostname"`
	RetentionLimit int    `json:"retentionLimit"`
	Team           string `json:"team"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (apiKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. The parse
// phase mirrors the plugin's own LoadSettings (pkg/models/settings.go:23-39)
// verbatim: seed the hostname and retentionLimit defaults, unmarshal jsonData
// over them (empty/malformed JSONData is a parse error, matching upstream's
// unconditional json.Unmarshal), and copy the decrypted apiKey secret.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate, returning a fully-defaulted, validated Config. Callers that
// assemble a Config themselves can invoke ApplyDefaults and Validate
// individually.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading honeycomb datasource config")

	// Seed defaults before unmarshal, mirroring LoadSettings
	// (pkg/models/settings.go:26-30): absent hostname/retentionLimit keys
	// keep these values.
	cfg := Config{
		Hostname:                DefaultHoneycombAPIURL,
		RetentionLimit:          DefaultRetentionLimitDays,
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings calls json.Unmarshal on s.JSONData
	// unconditionally (pkg/models/settings.go:32) and returns the error when
	// the bytes are empty or malformed. We mirror that behavior.
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		logger.Error("failed to parse jsonData", "err", err)
		return cfg, fmt.Errorf("parse jsonData: %w", err)
	}

	// Copy decrypted secrets by known key name (pkg/models/settings.go:27).
	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("honeycomb datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("honeycomb datasource config loaded",
		"hostname", cfg.Hostname,
		"team", cfg.Team,
		"hasEnvironment", cfg.Env != "",
		"retentionLimit", cfg.RetentionLimit,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own code applies (pkg/models/settings.go:28-29;
// src/types.ts:129; src/Views/ConfigEditor.tsx:46). Never blanket-apply
// every schema default — that would clobber intentional zero values.
//
// Curated defaults:
//   - Hostname: DefaultHoneycombAPIURL ("https://api.honeycomb.io") when empty.
//   - RetentionLimit: DefaultRetentionLimitDays (7) when zero.
//
// Team and the apiKey secret have no defaults — the plugin errors out when
// either is empty.
func (c *Config) ApplyDefaults() {
	if c.Hostname == "" {
		c.Hostname = DefaultHoneycombAPIURL
	}
	if c.RetentionLimit == 0 {
		c.RetentionLimit = DefaultRetentionLimitDays
	}
}

// Validate checks the runtime contract the plugin enforces in its health
// check (pkg/models/settings.go:45-71, Settings.Validate, called from
// pkg/plugin/healthcheck.go:13). Errors are joined so callers see every
// problem at once (upstream short-circuits on the first error; joining is
// the registry convention).
//
// Contracts enforced:
//   - hostname must be non-empty (:48-51), parse as a request URI (:52-56),
//     and use the https scheme (:57-60). Callers should invoke ApplyDefaults
//     first; LoadConfig always does.
//   - apiKey (secureJsonData.apiKey) must be non-empty (:61-64).
//   - team (jsonData.team) must be non-empty (:65-68).
func (c Config) Validate() error {
	var errs []error

	if strings.TrimSpace(c.Hostname) == "" {
		errs = append(errs, errors.New("hostname (jsonData.hostname) is required"))
	} else if u, err := url.ParseRequestURI(c.Hostname); err != nil {
		errs = append(errs, fmt.Errorf("invalid hostname URL (jsonData.hostname): %w", err))
	} else if u.Scheme != "https" {
		errs = append(errs, errors.New("hostname URL (jsonData.hostname) must use the https scheme"))
	}

	if strings.TrimSpace(c.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey]) == "" {
		errs = append(errs, errors.New("API key (secureJsonData.apiKey) is required"))
	}

	if strings.TrimSpace(c.Team) == "" {
		errs = append(errs, errors.New("team name (jsonData.team) is required"))
	}

	return errors.Join(errs...)
}
