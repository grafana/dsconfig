// Package parca contains the configuration models for the Grafana Parca
// datasource plugin (id: parca).
package parca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo). Parca declares no aliasIDs.
const PluginID = "parca"

// DeprecationDate mirrors the frontend constant at src/ConfigEditor.tsx:17.
// The Parca plugin is scheduled for deprecation on this date and will no
// longer receive updates after that. Exposed as a documented Go constant so
// callers (provisioning tools, migration scripts) can surface the same
// notice the editor renders in its warning banner.
const DeprecationDate = "2nd of January 2027"

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
// Parca datasource plugin itself defines no plugin-specific secrets beyond
// this shared HTTP-settings set.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// Config is the fully loaded configuration of a Grafana Parca datasource
// instance.
//
// The Parca backend's only server-side consumption of settings is:
//   - pkg/parca/plugin.go:65 — settings.HTTPClientOptions(ctx) to build the
//     HTTP client (that call is what pulls in root basicAuth / TLS fields /
//     custom headers / cookies).
//   - pkg/parca/plugin.go:77 — settings.URL is passed directly to
//     `queryv1alpha1connect.NewQueryServiceClient(httpClient, settings.URL,
//     connect.WithGRPCWeb())` as the base URL of the Connect/gRPC-web
//     profiling client.
//
// The Parca plugin ships **no** pkg/models/settings.go and **no** upstream
// LoadSettings, and its ParcaDataSourceOptions (src/types.ts:21) is a blank
// interface. The jsonData shape on this struct therefore mirrors what the
// editor writes and what a Grafana-side caller needs to know about a Parca
// datasource instance; every jsonData field here is written by
// @grafana/plugin-ui's Auth / AdvancedHttpSettings and consumed by the SDK's
// HTTPClientOptions, not by Parca-owned code.
type Config struct {
	// Root-level fields (json:"-" on the struct because they don't live in
	// jsonData). URL is read by the Parca backend directly
	// (pkg/parca/plugin.go:77). BasicAuth / BasicAuthUser / WithCredentials
	// are populated by the editor and consumed by the SDK's
	// HTTPClientOptions() call (pkg/parca/plugin.go:65) — the Parca code
	// itself never touches them by name.
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`

	// jsonData fields — the subset the editor writes and/or the SDK reads.
	// Custom HTTP header pairs (jsonData.httpHeaderName<N> /
	// secureJsonData.httpHeaderValue<N>) are not modeled here because they
	// are dynamically indexed.
	TLSAuth           bool     `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool     `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify     bool     `json:"tlsSkipVerify,omitempty"`
	ServerName        string   `json:"serverName,omitempty"`
	Timeout           float64  `json:"timeout,omitempty"`
	KeepCookies       []string `json:"keepCookies,omitempty"`
	OauthPassThru     bool     `json:"oauthPassThru,omitempty"`

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
// The Parca plugin has no upstream `LoadSettings` equivalent to mirror (see
// the package doc on Config). LoadConfig therefore represents the intended,
// flat shape a Grafana-side caller needs to interact with a Parca datasource
// instance.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext
// so log lines carry the request/plugin context that Grafana injects.
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

	logger.Debug("loading parca datasource config")

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
		logger.Error("parca datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("parca datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"tlsAuth", cfg.TLSAuth,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply
// every schema default — that would clobber intentional zero values.
//
// The Parca editor writes no defaults into jsonData on load. There are no
// visible placeholders that translate to persisted defaults; the
// "Timeout in seconds" placeholder on Timeout
// (`@grafana/plugin-ui AdvancedHttpSettings.tsx:74`) is a render-time UI
// hint, never persisted. This method intentionally does nothing so we don't
// clobber intentional zero values on a stored datasource; the
// TestApplyDefaults test guards this.
func (c *Config) ApplyDefaults() {
}

// Validate checks the runtime contract that the plugin requires.
//
// The Parca backend hard-fails without a URL at request time (the profiling
// client's ProfileTypes / Query / Values calls all issue requests against
// `settings.URL`). We surface the essentials at load time so provisioning
// tooling can reject misconfigurations upfront:
//
//   - URL is required.
//   - Basic auth requires a username (@grafana/plugin-ui BasicAuth.tsx
//     forces the pair when the method is selected).
//   - mTLS requires serverName + client cert + client key.
//   - Self-signed CA verification requires the CA PEM.
//   - `timeout` must be non-negative.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Parca URL (root.url) is required"))
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
