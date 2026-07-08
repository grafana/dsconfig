// Package csvdatasource contains the configuration models for the Grafana
// CSV datasource plugin (id: marcusolsson-csv-datasource).
package csvdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream repo). The CSV plugin declares no
// aliasIDs.
const PluginID = "marcusolsson-csv-datasource"

// StorageMode selects the backend storage source used to read CSV data.
// The value is stored as a JSON string under jsonData.storage. Options are
// defined verbatim by the RadioButtonGroup at src/ConfigEditor.tsx:66-69.
type StorageMode string

const (
	// StorageModeHTTP fetches CSV data over HTTP against root.url plus the
	// query editor's per-query Path (pkg/http_storage.go:73-131). This is
	// the default; both the frontend (src/utils.ts:4-10) and the backend
	// (pkg/settings.go:22-24) normalize an empty jsonData.storage to this
	// value for backwards compatibility.
	StorageModeHTTP StorageMode = "http"
	// StorageModeLocal reads CSV data from the local filesystem at
	// root.url (or root.url + query.Path) via pkg/local_storage.go:33-45.
	// Local storage is admin-gated: the plugin process must be started
	// with GF_PLUGIN_ALLOW_LOCAL_MODE=true or the backend returns
	// "local mode has been disabled by your administrator"
	// (pkg/datasource.go:44,158-160).
	StorageModeLocal StorageMode = "local"
	// DefaultStorageMode mirrors src/utils.ts:9 (defaultOptions.storage =
	// 'http') and pkg/settings.go:24 (if empty, "http"). Applied by
	// ApplyDefaults so a datasource that never touched the storage
	// selector still round-trips with StorageMode set.
	DefaultStorageMode StorageMode = StorageModeHTTP
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the Basic-auth password, set
	// when root.basicAuth is true (@grafana/plugin-ui BasicAuth.js:74).
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyTLSCACert is the custom CA PEM, set when
	// jsonData.tlsAuthWithCACert is true (@grafana/plugin-ui
	// SelfSignedCertificate.js:45-51).
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
	// SecureJsonDataKeyTLSClientCert is the mTLS client certificate PEM,
	// set when jsonData.tlsAuth is true (@grafana/plugin-ui
	// TLSClientAuth.js:70-81).
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the mTLS client key PEM, set when
	// jsonData.tlsAuth is true (@grafana/plugin-ui TLSClientAuth.js:97-108).
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
//
// Note: @grafana/plugin-ui's CustomHeaders component also writes indexed
// httpHeaderValue<N> secrets when the user configures custom HTTP headers.
// Those keys are dynamic (see README) and are not represented here. The
// CSV datasource plugin itself defines no plugin-specific secrets beyond
// this shared HTTP-settings set.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// Config is the fully loaded configuration of a Grafana CSV datasource
// instance.
//
// The Storage and QueryParams fields mirror the upstream backend
// PluginSettings struct verbatim (pkg/settings.go:10-13 — same field
// names, same json tags, same types). Every other jsonData field is
// written by @grafana/plugin-ui's Auth / AdvancedHttpSettings components
// and consumed by the SDK's HTTPClientOptions call in pkg/datasource.go:35,
// not by CSV-plugin-owned code.
//
// Root-level fields (URL, BasicAuth, BasicAuthUser, WithCredentials) are
// carried with json:"-" tags so they don't collide with jsonData
// unmarshaling. URL is DUAL-PURPOSE: it is the HTTP base URL when
// Storage="http" (pkg/http_storage.go:79) and a filesystem base path when
// Storage="local" (pkg/local_storage.go:33-45). BasicAuth /
// BasicAuthUser / WithCredentials are consumed by
// settings.HTTPClientOptions(ctx) via the SDK; the CSV plugin's own code
// never inspects them by name.
type Config struct {
	// Root-level fields (json:"-" — not part of jsonData).
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`

	// CSV plugin-owned jsonData fields. Mirror pkg/settings.go:10-13
	// verbatim (same field names, same json tags, same string types).
	Storage     StorageMode `json:"storage"`
	QueryParams string      `json:"queryParams"`

	// Auth / TLS / HTTP jsonData fields written by @grafana/plugin-ui's
	// Auth / AdvancedHttpSettings and read by the SDK's HTTPClientOptions.
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
// verbatim from settings.JSONData (mirroring the upstream PluginSettings
// parsing at pkg/settings.go:15-27); decrypted secrets are copied by known
// key name into DecryptedSecureJSONData.
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

	logger.Debug("loading csv datasource config")

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
		logger.Error("csv datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("csv datasource config loaded",
		"storage", cfg.Storage,
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"tlsAuth", cfg.TLSAuth,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the plugin's own code applies. Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// Curated defaults:
//   - Storage: "http" — mirrors src/utils.ts:4-10 (getOptionsWithDefaults)
//     and pkg/settings.go:22-24 (LoadPluginSettings). Both the editor and
//     the backend treat an empty storage as "http" for backwards
//     compatibility.
//
// The other fields (Timeout, KeepCookies, TLS toggles, etc.) have no
// persisted defaults — the editor never writes them until the user makes
// a change, and the SDK's HTTPClientOptions handles their zero values.
func (c *Config) ApplyDefaults() {
	if c.Storage == "" {
		c.Storage = DefaultStorageMode
	}
}

// Validate checks the runtime contract that the plugin requires. Errors
// are joined so callers see every problem at once.
//
// Contracts enforced:
//   - Storage must be one of {"", "http", "local"}. An empty value is
//     accepted because callers may call Validate before ApplyDefaults;
//     LoadConfig always applies defaults first.
//   - When Storage == "http", the root URL is required — pkg/datasource.go:
//     121-125 CheckHealth rejects an empty URL with "URL is required for
//     HTTP storage".
//   - When Storage == "local", the root URL is required — it is the base
//     filesystem path pkg/local_storage.go:33,49 opens. An empty path
//     would make os.Open("") fail with a downstream error at every
//     query. NOTE: this method does NOT check that the local mode is
//     admin-enabled (GF_PLUGIN_ALLOW_LOCAL_MODE=true) — that flag lives
//     on the plugin process at runtime, not in the datasource settings.
//   - Basic auth requires a username (@grafana/plugin-ui BasicAuth.js
//     forces the pair when the method is selected).
//   - mTLS requires ServerName + client cert + client key.
//   - Self-signed CA verification requires the CA PEM.
//   - Timeout must be non-negative (@grafana/plugin-ui
//     AdvancedHttpSettings.js:69 already sets `min={0}` on the input).
func (c Config) Validate() error {
	var errs []error

	switch c.Storage {
	case "", StorageModeHTTP, StorageModeLocal:
		// OK. "" is the pre-ApplyDefaults state.
	default:
		errs = append(errs, fmt.Errorf("invalid storage %q: must be \"http\" or \"local\"", c.Storage))
	}

	// URL is required in every valid storage mode — HTTP uses it as the
	// endpoint; local uses it as the filesystem base path.
	if c.URL == "" {
		switch c.Storage {
		case StorageModeLocal:
			errs = append(errs, errors.New("CSV file path (root.url) is required when storage is \"local\""))
		default:
			// Cover both http and the empty (pre-defaults) state
			// because the empty state defaults to http.
			errs = append(errs, errors.New("CSV URL (root.url) is required when storage is \"http\""))
		}
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
