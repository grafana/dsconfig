// Package mockdatasource contains the configuration models for the
// Grafana Mock datasource plugin (id: grafana-mock-datasource).
package mockdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:3 in the upstream repo). The Mock plugin declares no
// aliasIDs.
const PluginID = "grafana-mock-datasource"

// CustomHealthCheckStatus is the numeric health status returned by the
// custom health check override. Mirrors backend.HealthStatus, which is
// what the backend maps it to (pkg/client/handler_checkhealth.go:32).
type CustomHealthCheckStatus int

const (
	// CustomHealthCheckStatusUnknown is the default status when the custom
	// health check option is enabled but `status` is not set (or set to 0).
	CustomHealthCheckStatusUnknown CustomHealthCheckStatus = 0
	// CustomHealthCheckStatusOK reports the datasource is healthy.
	CustomHealthCheckStatusOK CustomHealthCheckStatus = 1
	// CustomHealthCheckStatusError reports the datasource is unhealthy.
	CustomHealthCheckStatusError CustomHealthCheckStatus = 2
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the Basic-auth password, set
	// when root.basicAuth is true.
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyTLSCACert is the custom CA PEM, set when
	// jsonData.tlsAuthWithCACert is true.
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
	// SecureJsonDataKeyTLSClientCert is the mTLS client certificate PEM,
	// set when jsonData.tlsAuth is true.
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the mTLS client key PEM, set when
	// jsonData.tlsAuth is true.
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
)

// SecureJsonDataConfig lists the secret key names stored in
// secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the Mock plugin's editor may
// write. The Mock plugin itself defines no plugin-specific secrets
// (src/types/config.types.ts:8: `mockSecureConfigKeys = [] as const`);
// every key listed here is written by @grafana/plugin-ui's `Auth`
// component when the user selects Basic auth or enables TLS. The
// CustomHeaders component may additionally write indexed
// `httpHeaderValue<N>` secrets — those keys are dynamic and are not
// represented here.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// CustomHealthCheckConfig is the nested `customHealthCheck` object under
// jsonData. Mirrors pkg/models/settings.go:16-21 verbatim — same fields,
// same json tags — with the Status field re-typed as
// CustomHealthCheckStatus for callers.
type CustomHealthCheckConfig struct {
	// Status is the numeric health status returned by the override
	// (0=UNKNOWN, 1=OK, 2=ERROR).
	Status CustomHealthCheckStatus `json:"status"`
	// Message is the custom health check message. Blank falls back to
	// "health check message not specified" server-side.
	Message string `json:"message"`
	// Details is the opaque details string passed through to jsonDetails
	// on the CheckHealth response. Expected to be JSON but the backend
	// does not validate it.
	Details string `json:"details"`
	// SkipBackend, when true, causes the frontend to short-circuit
	// testDatasource() and synthesise the health response locally
	// (src/datasource.ts:32-58). Parsed by the backend but never read
	// there.
	SkipBackend bool `json:"skipBackend"`
}

// Config is the fully loaded configuration of a Grafana Mock datasource
// instance. It mirrors the plugin's own upstream
// pkg/models/settings.go:11-14 (CustomHealthCheckEnabled +
// CustomHealthCheck) and extends it with the standard HTTP-settings
// jsonData fields written by @grafana/plugin-ui's `Auth`
// (convertLegacyAuthProps).
//
// The Mock backend consumes settings in two places:
//   - pkg/client/client.go:21 — setting.HTTPClientOptions(ctx) builds the
//     HTTP client from the root fields (URL/basicAuth/basicAuthUser) and
//     the TLS/OAuth jsonData fields (tlsAuth, tlsAuthWithCACert,
//     tlsSkipVerify, serverName, oauthPassThru).
//   - pkg/client/handler_checkhealth.go:20-40 — reads
//     Config.CustomHealthCheckEnabled and Config.CustomHealthCheck to
//     shape the CheckHealth response.
type Config struct {
	// Root-level fields (json:"-" on the struct because they don't live
	// in jsonData). Populated by the editor and consumed by the SDK's
	// HTTPClientOptions() call. The Mock code itself never touches them
	// by name.
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`

	// Plugin-owned jsonData fields — verbatim from
	// pkg/models/settings.go:11-14.
	CustomHealthCheckEnabled bool                    `json:"customHealthCheckEnabled"`
	CustomHealthCheck        CustomHealthCheckConfig `json:"customHealthCheck"`

	// Additional jsonData fields written by @grafana/plugin-ui's `Auth`
	// (convertLegacyAuthProps) and read by the SDK's HTTPClientOptions.
	// Not declared by the plugin's own upstream Config, but the editor
	// persists them and the SDK reads them.
	TLSAuth           bool   `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool   `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify     bool   `json:"tlsSkipVerify,omitempty"`
	ServerName        string `json:"serverName,omitempty"`
	OauthPassThru     bool   `json:"oauthPassThru,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser) are copied from
// backend.DataSourceInstanceSettings directly; jsonData is unmarshaled
// verbatim from settings.JSONData; decrypted secrets are copied by known
// key name into DecryptedSecureJSONData.
//
// The unmarshal step mirrors the plugin's own LoadSettings
// (pkg/models/settings.go:23-27): a straight json.Unmarshal into the
// Config struct, no legacy fallbacks or lenient parsing. Any drift from
// upstream should be treated as a bug in this entry.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext
// so log lines carry the request/plugin context that Grafana injects.
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

	logger.Debug("loading mock datasource config")

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
		logger.Error("mock datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("mock datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"tlsAuth", cfg.TLSAuth,
		"customHealthCheck", cfg.CustomHealthCheckEnabled,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the
// same defaults the editor writes for a fresh datasource. Never
// blanket-apply every schema default — that would clobber intentional
// zero values.
//
// The Mock editor writes no defaults into jsonData on load. The
// CustomHealthCheck fields all default to their Go zero values
// (status=UNKNOWN=0, empty strings, skipBackend=false), which matches
// what the editor renders when jsonData.customHealthCheck is undefined
// (`src/editors/MockConfigEditor.tsx:40` uses `jsonData.customHealthCheck
// || { status: 0 }`). This method intentionally does nothing so we don't
// clobber intentional zero values on a stored datasource; the
// TestApplyDefaults test guards this.
func (c *Config) ApplyDefaults() {
}

// Validate checks the runtime contract that the plugin requires.
//
// The Mock plugin's backend never dials root.URL and never checks the
// health of the URL, so URL is not required by the backend contract.
// The editor and the plugin-ui ConnectionSettings component both mark
// URL as required — that visual requirement is captured in the schema
// via a UI hint rather than here. This validator only enforces the
// contract the runtime code actually needs:
//
//   - Basic auth requires a username (@grafana/plugin-ui BasicAuth.tsx
//     forces the pair when the method is selected).
//   - mTLS requires serverName + client cert + client key.
//   - Self-signed CA verification requires the CA PEM.
//   - CustomHealthCheck.Status must be one of {UNKNOWN, OK, ERROR}
//     because backend.HealthStatus(other) is a silent no-op — surfacing
//     it at load time avoids a confusing health check response.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

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

	if c.CustomHealthCheckEnabled {
		switch c.CustomHealthCheck.Status {
		case CustomHealthCheckStatusUnknown, CustomHealthCheckStatusOK, CustomHealthCheckStatusError:
		default:
			errs = append(errs, fmt.Errorf(
				"customHealthCheck.status must be 0 (UNKNOWN), 1 (OK), or 2 (ERROR); got %d",
				c.CustomHealthCheck.Status))
		}
	}

	return errors.Join(errs...)
}
