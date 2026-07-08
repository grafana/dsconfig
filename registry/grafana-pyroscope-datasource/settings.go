// Package pyroscopedatasource contains the configuration models for the
// Grafana Pyroscope datasource plugin (id: grafana-pyroscope-datasource,
// aliasID: phlare).
package pyroscopedatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/gtime"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo). The plugin also declares a legacy
// aliasID at src/plugin.json:6.
const PluginID = "grafana-pyroscope-datasource"

// LegacyPluginID is the pre-rename plugin id ("phlare"). Kept as a
// documented constant so callers migrating provisioning payloads can resolve
// both names to the same schema.
const LegacyPluginID = "phlare"

// MinStepPattern mirrors the frontend validation regex at
// src/ConfigEditor.tsx:65 — a positive integer followed by one of the allowed
// duration unit specifiers. Kept exported so callers assembling a Config
// directly can pre-validate a duration string with the same rule the editor
// applies.
var MinStepPattern = regexp.MustCompile(`^\d+(ms|[Mwdhmsy])$`)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the Basic-auth password, set when
	// root.basicAuth is true.
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyTLSCACert is the custom CA PEM, set when
	// jsonData.tlsAuthWithCACert is true.
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
	// SecureJsonDataKeyTLSClientCert is the mTLS client certificate PEM, set
	// when jsonData.tlsAuth is true.
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the mTLS client key PEM, set when
	// jsonData.tlsAuth is true.
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
//
// Note: @grafana/plugin-ui's CustomHeaders component also writes indexed
// httpHeaderValue<N> secrets when the user configures custom HTTP headers.
// Those keys are dynamic (see README) and are not represented here. The
// Pyroscope datasource plugin itself defines no plugin-specific secrets
// beyond this shared HTTP-settings set.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// Config is the fully loaded configuration of a Grafana Pyroscope datasource
// instance.
//
// The Pyroscope backend's only server-side consumption of settings is:
//   - pkg/grafana-pyroscope-datasource/instance.go:51-68 (NewPyroscopeDatasource) —
//     reads settings.URL directly and hands backend.DataSourceInstanceSettings
//     to the SDK's HTTPClientOptions to build the HTTP client (that call is
//     what pulls in root basicAuth / TLS fields / custom headers / cookies).
//   - pkg/grafana-pyroscope-datasource/query.go:74-89, 173-187 — unmarshals
//     settings.JSONData into an ad-hoc `dsJsonModel struct { MinStep string
//     \`json:"minStep"\` }` inside query handling to derive the effective
//     query step (`max(query.Interval, MinStep)`, defaulting to 15s on parse
//     failure).
//
// The Pyroscope plugin ships no pkg/models/settings.go, so the jsonData shape
// on this struct is the intended settings model: it mirrors what the editor
// writes and what a Grafana-side caller needs to know about a Pyroscope
// datasource instance.
type Config struct {
	// Root-level fields (json:"-" on the struct because they don't live in
	// jsonData). URL is read by the Pyroscope backend directly
	// (pkg/grafana-pyroscope-datasource/instance.go:66). BasicAuth /
	// BasicAuthUser / WithCredentials are populated by the editor and consumed
	// by the SDK's HTTPClientOptions() call (instance.go:52) — the Pyroscope
	// code itself never touches them by name.
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`

	// jsonData fields — the subset the editor writes and/or the SDK reads.
	// Custom HTTP header pairs (jsonData.httpHeaderName<N> /
	// secureJsonData.httpHeaderValue<N>) are not modeled here because they are
	// dynamically indexed.
	TLSAuth           bool     `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool     `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify     bool     `json:"tlsSkipVerify,omitempty"`
	ServerName        string   `json:"serverName,omitempty"`
	Timeout           float64  `json:"timeout,omitempty"`
	KeepCookies       []string `json:"keepCookies,omitempty"`
	OauthPassThru     bool     `json:"oauthPassThru,omitempty"`
	MinStep           string   `json:"minStep,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser) are copied from
// backend.DataSourceInstanceSettings directly; jsonData is unmarshaled
// verbatim from settings.JSONData; decrypted secrets are copied by known key
// name into DecryptedSecureJSONData.
//
// The Pyroscope plugin has no upstream `LoadSettings` equivalent to mirror
// (see the package doc on Config). LoadConfig therefore represents the
// intended, flat shape a Grafana-side caller needs to interact with a
// Pyroscope datasource instance.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Callers that need each phase individually can invoke
// ApplyDefaults and Validate directly on a Config they assemble themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading pyroscope datasource config")

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
		logger.Error("pyroscope datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("pyroscope datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"tlsAuth", cfg.TLSAuth,
		"minStep", cfg.MinStep,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// The Pyroscope editor writes no defaults into jsonData on load. The visual
// defaults shown in the editor — the "15s" placeholder on Minimal step
// (`ConfigEditor.tsx:71`), the "Timeout in seconds" placeholder on Timeout
// (`AdvancedHttpSettings.tsx:74`) — are render-time UI hints, never
// persisted. The query-time 15s fallback for an empty/unparseable MinStep
// (`pkg/grafana-pyroscope-datasource/query.go:82-89`) is applied per query,
// not baked into stored settings. This method intentionally does nothing so
// we don't clobber intentional zero values on a stored datasource; the
// TestApplyDefaults test guards this.
func (c *Config) ApplyDefaults() {
}

// Validate checks the runtime contract that the plugin requires.
//
// The Pyroscope backend hard-fails without a URL at request time (the
// profiling client's ProfileTypes / SelectSeries / etc calls all issue
// requests against `settings.URL`). We surface the essentials at load time
// so provisioning tooling can reject misconfigurations upfront:
//
//   - URL is required.
//   - Basic auth requires a username (@grafana/plugin-ui BasicAuth.tsx forces
//     the pair when the method is selected).
//   - mTLS requires serverName + client cert + client key.
//   - Self-signed CA verification requires the CA PEM.
//   - `timeout` must be non-negative.
//   - `minStep`, when set, must match the same regex the editor enforces
//     (`^\d+(ms|[Mwdhmsy])$`); the query path silently falls back to 15s on a
//     parse failure, so this is the only place a malformed value would be
//     surfaced to a provisioning caller.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Pyroscope URL (root.url) is required"))
	}

	if c.BasicAuth {
		if c.BasicAuthUser == "" {
			errs = append(errs, errors.New("basicAuthUser (root) is required when basicAuth is true"))
		}
	}

	if c.TLSAuth {
		if c.ServerName == "" {
			errs = append(errs, errors.New("serverName (jsonData) is required when tlsAuth is true"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert] == "" {
			errs = append(errs, errors.New("tlsClientCert (secureJsonData) is required when tlsAuth is true"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey] == "" {
			errs = append(errs, errors.New("tlsClientKey (secureJsonData) is required when tlsAuth is true"))
		}
	}
	if c.TLSAuthWithCACert {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert] == "" {
			errs = append(errs, errors.New("tlsCACert (secureJsonData) is required when tlsAuthWithCACert is true"))
		}
	}

	if c.Timeout < 0 {
		errs = append(errs, fmt.Errorf("timeout must be non-negative, got %v", c.Timeout))
	}

	if c.MinStep != "" {
		if !MinStepPattern.MatchString(c.MinStep) {
			errs = append(errs, fmt.Errorf(
				"minStep %q: must match %s (e.g. \"15s\", \"1m\", \"500ms\")",
				c.MinStep, MinStepPattern,
			))
		} else if _, err := gtime.ParseDuration(c.MinStep); err != nil {
			// Should be unreachable given the regex above, but the query path
			// itself parses via gtime.ParseDuration and treats a failure as a
			// silent fallback to 15s. Surface the parse error explicitly for
			// provisioning callers so an unexpected input is never silently
			// swallowed.
			errs = append(errs, fmt.Errorf("minStep %q: %w", c.MinStep, err))
		}
	}

	return errors.Join(errs...)
}
