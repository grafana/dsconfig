// Package graphitedatasource contains the configuration models for the
// Graphite datasource plugin (id: graphite).
package graphitedatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "graphite"

// GraphiteVersion is the schema version of the target Graphite server. Mirrors
// the values enumerated at src/versions.ts:3.
type GraphiteVersion string

const (
	// GraphiteVersion09 corresponds to the "0.9.x" option in the editor.
	GraphiteVersion09 GraphiteVersion = "0.9"
	// GraphiteVersion10 corresponds to the "1.0.x" option in the editor.
	GraphiteVersion10 GraphiteVersion = "1.0"
	// GraphiteVersion11 is the default written by ConfigEditor.tsx:43-45.
	GraphiteVersion11 GraphiteVersion = "1.1"
	// DefaultGraphiteVersion mirrors DEFAULT_GRAPHITE_VERSION from
	// src/versions.ts:5 — the fallback the editor writes on mount if the
	// datasource has no version set.
	DefaultGraphiteVersion GraphiteVersion = GraphiteVersion11
)

// GraphiteType is the backend flavour, from src/types.ts:30-33.
type GraphiteType string

const (
	// GraphiteTypeDefault targets a stock Graphite / graphite-web server.
	GraphiteTypeDefault GraphiteType = "default"
	// GraphiteTypeMetrictank targets a Metrictank server; setting this
	// reveals the Rollup indicator switch in the editor
	// (ConfigEditor.tsx:95).
	GraphiteTypeMetrictank GraphiteType = "metrictank"
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
// Note: @grafana/ui's CustomHeadersSettings component also writes indexed
// httpHeaderValue<N> secrets when the user configures custom HTTP headers.
// Those keys are not represented here because they are dynamic (see README).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// GraphiteMetricLokiMatcher is a single node matcher inside a Graphite→Loki
// mapping. When LabelName is set, the segment is a label capture; otherwise
// Value is matched literally. Mirrors src/types.ts:68-71.
type GraphiteMetricLokiMatcher struct {
	Value     string `json:"value"`
	LabelName string `json:"labelName,omitempty"`
}

// GraphiteLokiMapping is a single mapping — an ordered list of matchers
// (src/types.ts:64-66).
type GraphiteLokiMapping struct {
	Matchers []GraphiteMetricLokiMatcher `json:"matchers"`
}

// GraphiteToLokiQueryImportConfiguration is the Loki-specific portion of the
// import configuration (src/types.ts:60-62).
type GraphiteToLokiQueryImportConfiguration struct {
	Mappings []GraphiteLokiMapping `json:"mappings"`
}

// GraphiteQueryImportConfiguration is the cross-datasource migration hint used
// by Explore's datasource-switch flow (src/types.ts:56-58). Neither Graphite
// nor Loki reads this at query time.
type GraphiteQueryImportConfiguration struct {
	Loki GraphiteToLokiQueryImportConfiguration `json:"loki"`
}

// Config is the fully loaded configuration of a Graphite datasource instance.
//
// The Graphite plugin has no upstream `pkg/models/settings.go` — the backend
// reads only `settings.URL` and `settings.ID` (pkg/graphite/graphite.go:51-52)
// and delegates HTTP client construction to `settings.HTTPClientOptions(ctx)`
// (pkg/graphite/graphite.go:38). This Config therefore represents the shape
// the editor writes and the SDK reads, not a plugin-owned settings model.
//
// Root-level fields the editor writes (URL, BasicAuth, BasicAuthUser,
// WithCredentials) are carried with `json:"-"` tags so they don't collide
// with jsonData unmarshaling. Access is intentionally omitted because the
// Graphite editor never exposes it (see settings.ts RootConfig for the note).
type Config struct {
	// Root-level fields (json:"-" — not part of jsonData). URL is required and
	// checked by admission_handler.go:51 before any other setting is read.
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`

	// jsonData fields — the subset the editor writes and/or the SDK reads.
	// Custom HTTP header pairs (jsonData.httpHeaderName<N> /
	// secureJsonData.httpHeaderValue<N>) are not modeled here because they are
	// dynamically indexed.
	TLSAuth                bool                             `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert      bool                             `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify          bool                             `json:"tlsSkipVerify,omitempty"`
	ServerName             string                           `json:"serverName,omitempty"`
	Timeout                float64                          `json:"timeout,omitempty"`
	KeepCookies            []string                         `json:"keepCookies,omitempty"`
	OauthPassThru          bool                             `json:"oauthPassThru,omitempty"`
	GraphiteVersion        GraphiteVersion                  `json:"graphiteVersion,omitempty"`
	GraphiteType           GraphiteType                     `json:"graphiteType,omitempty"`
	RollupIndicatorEnabled bool                             `json:"rollupIndicatorEnabled,omitempty"`
	ImportConfiguration    GraphiteQueryImportConfiguration `json:"importConfiguration,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser, WithCredentials) are copied from
// backend.DataSourceInstanceSettings directly; jsonData is unmarshaled
// verbatim from settings.JSONData; decrypted secrets are copied by known key
// name into DecryptedSecureJSONData.
//
// The Graphite plugin has no upstream `LoadSettings` equivalent to mirror
// (see the doc on Config). LoadConfig therefore represents the intended, flat
// shape a Grafana-side caller needs to interact with a Graphite datasource
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

	logger.Debug("loading graphite datasource config")

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
		logger.Error("graphite datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("graphite datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"tlsAuth", cfg.TLSAuth,
		"graphiteVersion", cfg.GraphiteVersion,
		"graphiteType", cfg.GraphiteType,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// Curated defaults:
//   - GraphiteVersion: '1.1' — the editor's componentDidMount always writes
//     `DEFAULT_GRAPHITE_VERSION` on load if the field is empty
//     (ConfigEditor.tsx:43-45, versions.ts:5). Any datasource that has been
//     opened in the editor at least once carries a non-empty value; we mirror
//     that behaviour so provisioning payloads that omit graphiteVersion still
//     match editor parity.
//
// GraphiteType is intentionally NOT defaulted: the editor's Select renders no
// selection until the user picks one, so an untouched datasource may carry an
// empty graphiteType. Applying a default here would fabricate a choice the
// user never made.
func (c *Config) ApplyDefaults() {
	if c.GraphiteVersion == "" {
		c.GraphiteVersion = DefaultGraphiteVersion
	}
}

// Validate checks the runtime contract that the plugin requires. The Graphite
// backend's admission handler (pkg/graphite/admission_handler.go:45-53)
// rejects any request with an empty URL or an unsupported apiVersion; the
// health check subsequently issues a `constantLine(100)` render against the
// URL. This method encodes the same URL requirement, plus the enum
// constraints on GraphiteVersion and GraphiteType, and the TLS field pairs
// required by @grafana/ui's TLSAuthSettings + the SDK. Errors are joined so
// callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Graphite URL (root.url) is required"))
	}

	switch c.GraphiteVersion {
	case "", GraphiteVersion09, GraphiteVersion10, GraphiteVersion11:
		// OK. Empty is accepted here because callers may call Validate before
		// ApplyDefaults; LoadConfig always applies defaults first.
	default:
		errs = append(errs, fmt.Errorf("invalid graphiteVersion %q: must be %q, %q, or %q",
			c.GraphiteVersion, GraphiteVersion09, GraphiteVersion10, GraphiteVersion11))
	}

	switch c.GraphiteType {
	case "", GraphiteTypeDefault, GraphiteTypeMetrictank:
		// OK. Empty is the untouched-editor state.
	default:
		errs = append(errs, fmt.Errorf("invalid graphiteType %q: must be %q or %q",
			c.GraphiteType, GraphiteTypeDefault, GraphiteTypeMetrictank))
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

	return errors.Join(errs...)
}
