// Package lokidatasource contains the configuration models for the Loki
// datasource plugin (id: loki).
package lokidatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "loki"

// DerivedFieldMatcherType is the discriminator for a derived field's
// extraction strategy — either a regex against the log line or a label name.
// Mirrors the two string values written by the editor Select at
// src/configuration/DerivedField.tsx:99-100.
type DerivedFieldMatcherType string

const (
	// DerivedFieldMatcherRegex extracts the field by applying a regular
	// expression to the log message. Default in src/configuration/DerivedField.tsx:61.
	DerivedFieldMatcherRegex DerivedFieldMatcherType = "regex"
	// DerivedFieldMatcherLabel extracts the field by reading a label of the
	// log stream (src/configuration/DerivedField.tsx:100).
	DerivedFieldMatcherLabel DerivedFieldMatcherType = "label"
)

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
// Those keys are not represented here because they are dynamic (see README).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// DerivedFieldConfig mirrors the frontend DerivedFieldConfig type
// (src/types.ts:56-64) that the config editor writes into
// jsonData.derivedFields and the frontend result transformer consumes at
// src/datasource.ts:397. The Loki backend never reads this field.
type DerivedFieldConfig struct {
	Name            string                  `json:"name"`
	MatcherRegex    string                  `json:"matcherRegex"`
	MatcherType     DerivedFieldMatcherType `json:"matcherType,omitempty"`
	URL             string                  `json:"url,omitempty"`
	URLDisplayLabel string                  `json:"urlDisplayLabel,omitempty"`
	DatasourceUID   string                  `json:"datasourceUid,omitempty"`
	TargetBlank     bool                    `json:"targetBlank,omitempty"`
}

// Config is the fully loaded configuration of a Loki datasource instance.
//
// The Loki backend's only server-side consumption of settings is
// pkg/loki/loki.go:48-72, which reads settings.URL directly and hands
// backend.DataSourceInstanceSettings to the SDK's HTTPClientOptions to build
// the HTTP client (that call is what pulls in root basicAuth / TLS fields /
// custom headers / cookies). No jsonData field is unmarshaled server-side.
//
// The jsonData fields on this struct therefore represent the shape the
// frontend writes and the SDK reads, not a plugin-owned upstream settings
// model — the Loki plugin does not ship a pkg/models/settings.go. `URL`,
// `BasicAuth`, and `BasicAuthUser` are carried as root fields so callers get
// a single flat Config that mirrors what a client of the Loki datasource
// would need to authenticate.
type Config struct {
	// Root-level fields (json:"-" on the struct because they don't live in jsonData).
	// URL is read by the Loki backend directly (pkg/loki/loki.go:66). BasicAuth /
	// BasicAuthUser / WithCredentials are populated by the editor and consumed by
	// the SDK's HTTPClientOptions() call (pkg/loki/loki.go:51) — the Loki code
	// itself never touches them by name.
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`

	// jsonData fields — the subset the editor writes and/or the SDK reads.
	// Custom HTTP header pairs (jsonData.httpHeaderName<N> /
	// secureJsonData.httpHeaderValue<N>) are not modeled here because they are
	// dynamically indexed.
	TLSAuth           bool                 `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool                 `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify     bool                 `json:"tlsSkipVerify,omitempty"`
	ServerName        string               `json:"serverName,omitempty"`
	Timeout           float64              `json:"timeout,omitempty"`
	KeepCookies       []string             `json:"keepCookies,omitempty"`
	OauthPassThru     bool                 `json:"oauthPassThru,omitempty"`
	ManageAlerts      bool                 `json:"manageAlerts,omitempty"`
	MaxLines          string               `json:"maxLines,omitempty"`
	DerivedFields     []DerivedFieldConfig `json:"derivedFields,omitempty"`

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
// The Loki plugin has no upstream `LoadSettings` equivalent to mirror (see the
// package doc on Config). LoadConfig therefore represents the intended, flat
// shape a Grafana-side caller needs to interact with a Loki datasource
// instance.
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

	logger.Debug("loading loki datasource config")

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
		logger.Error("loki datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("loki datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"tlsAuth", cfg.TLSAuth,
		"derivedFields", len(cfg.DerivedFields),
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// Loki's editor writes no defaults into jsonData on load (unlike Prometheus,
// which defaults httpMethod to POST). The visual defaults that appear in the
// editor — the "1000" placeholder on Maximum lines (`QuerySettings.tsx:42`),
// the "regex" default matcherType for a newly-added derived field
// (`DerivedFields.tsx:98`), the `config.defaultDatasourceManageAlertsUiToggle`
// fallback for the Manage alerts switch (`AlertingSettings.tsx:29`) — are all
// rendered from `??` fallbacks at render time and never persisted. This method
// intentionally does nothing so we don't clobber intentional zero values on a
// stored datasource.
func (c *Config) ApplyDefaults() {
}

// Validate checks the runtime contract that the plugin requires. The Loki
// backend hard-fails without a URL at request time
// (pkg/loki/loki.go:66 followed by fmt.Sprintf("/loki/api/v1/%s", url) at
// loki.go:119). It also encodes the TLS field pairs required by
// @grafana/plugin-ui and the SDK, and sanity-checks numeric ranges. Errors are
// joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Loki URL (root.url) is required"))
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

	for i, df := range c.DerivedFields {
		if df.Name == "" {
			errs = append(errs, fmt.Errorf("derivedFields[%d].name is required", i))
		}
		if df.MatcherRegex == "" {
			errs = append(errs, fmt.Errorf("derivedFields[%d].matcherRegex is required", i))
		}
		switch df.MatcherType {
		case "", DerivedFieldMatcherRegex, DerivedFieldMatcherLabel:
			// OK. Empty is accepted because the frontend treats it as "regex"
			// (DerivedField.tsx:61).
		default:
			errs = append(errs, fmt.Errorf("derivedFields[%d].matcherType %q: must be %q or %q",
				i, df.MatcherType, DerivedFieldMatcherRegex, DerivedFieldMatcherLabel))
		}
	}

	return errors.Join(errs...)
}
