// Package influxdbdatasource contains the configuration models for the
// InfluxDB datasource plugin (id: influxdb).
package influxdbdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "influxdb"

// InfluxVersion is the query-language discriminator stored under
// jsonData.version. Values mirror the frontend InfluxVersion enum
// (src/types.ts:5-9) and the backend constants
// (pkg/influxdb/settings.go:3-7).
type InfluxVersion string

const (
	// InfluxVersionInfluxQL selects the InfluxQL query path
	// (pkg/influxdb/influxdb.go:100-101).
	InfluxVersionInfluxQL InfluxVersion = "InfluxQL"
	// InfluxVersionFlux selects the Flux query path
	// (pkg/influxdb/influxdb.go:98-99).
	InfluxVersionFlux InfluxVersion = "Flux"
	// InfluxVersionSQL selects the FlightSQL query path
	// (pkg/influxdb/influxdb.go:102-103).
	InfluxVersionSQL InfluxVersion = "SQL"
	// DefaultInfluxVersion mirrors the backend fallback at
	// pkg/influxdb/influxdb.go:53-56.
	DefaultInfluxVersion = InfluxVersionInfluxQL
)

// InfluxHTTPMode is the HTTP verb used by the InfluxQL query path
// (jsonData.httpMode). Options come from InfluxInfluxQLConfig.tsx:25-28.
type InfluxHTTPMode string

const (
	// InfluxHTTPModeGET is the default (pkg/influxdb/influxdb.go:43-46).
	InfluxHTTPModeGET InfluxHTTPMode = http.MethodGet
	// InfluxHTTPModePOST is used when the InfluxQL query body is too large
	// for a URL query string.
	InfluxHTTPModePOST InfluxHTTPMode = http.MethodPost
	// DefaultInfluxHTTPMode mirrors the backend fallback at
	// pkg/influxdb/influxdb.go:43-46.
	DefaultInfluxHTTPMode = InfluxHTTPModeGET
)

// InfluxProduct mirrors the product-detection value written by the v2 editor
// (versions.ts:24-155). Not consumed by the backend.
type InfluxProduct string

const (
	InfluxProductCloudDedicated  InfluxProduct = "InfluxDB Cloud Dedicated"
	InfluxProductCloudServerless InfluxProduct = "InfluxDB Cloud Serverless"
	InfluxProductClustered       InfluxProduct = "InfluxDB Clustered"
	InfluxProductEnterprise1x    InfluxProduct = "InfluxDB Enterprise 1.x"
	InfluxProductEnterprise3x    InfluxProduct = "InfluxDB Enterprise 3.x"
	InfluxProductCloudTSM        InfluxProduct = "InfluxDB Cloud (TSM)"
	InfluxProductCloud1          InfluxProduct = "InfluxDB Cloud 1"
	InfluxProductOSS1x           InfluxProduct = "InfluxDB OSS 1.x"
	InfluxProductOSS2x           InfluxProduct = "InfluxDB OSS 2.x"
	InfluxProductOSS3x           InfluxProduct = "InfluxDB OSS 3.x"
)

// DefaultMaxSeries mirrors the backend fallback at
// pkg/influxdb/influxdb.go:48-51.
const DefaultMaxSeries int32 = 1000

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the Flux/SQL bearer token. Read by the
	// backend at pkg/influxdb/influxdb.go:81.
	SecureJsonDataKeyToken SecureJsonDataKey = "token"
	// SecureJsonDataKeyPassword is the legacy v1 InfluxQL database
	// password paired with root.user (InfluxInfluxQLConfig.tsx:98-107).
	// Not consumed by the current backend.
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// SecureJsonDataKeyBasicAuthPassword is the HTTP Basic password paired
	// with root.basicAuthUser (v2 editor's Basic auth radio,
	// AuthSettings.tsx:172-178).
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

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
//
// Note: @grafana/ui's CustomHeadersSettings component (used by the v2
// editor's AdvancedHttpSettings.tsx:102) also writes indexed
// httpHeaderValue<N> secrets when the user configures custom HTTP headers.
// Those keys are not represented here because they are dynamic (see README).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
	SecureJsonDataKeyPassword,
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// Config is the fully loaded configuration of an InfluxDB datasource
// instance. It combines the InfluxDB-specific jsonData fields that the
// backend unmarshals directly (pkg/influxdb/models/datasource_info.go:11-32)
// with the HTTP/TLS jsonData fields the SDK's HTTPClientOptions consumes.
//
// Root-level fields (URL, BasicAuth, BasicAuthUser, User, Database,
// WithCredentials) are carried with json:"-" tags so they don't collide
// with jsonData unmarshaling. The backend reads settings.URL directly
// (pkg/influxdb/influxdb.go:72) and falls back to settings.Database when
// jsonData.dbName is empty (pkg/influxdb/influxdb.go:58-61); BasicAuth and
// User are consumed by the SDK transport layer.
type Config struct {
	// Root-level fields (json:"-" — not part of jsonData). URL is required.
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	User            string `json:"-"`
	Database        string `json:"-"`
	WithCredentials bool   `json:"-"`

	// HTTP transport jsonData fields the SDK reads via HTTPClientOptions
	// (TLS, cookies, timeout, OAuth forward). Custom HTTP header pairs
	// (jsonData.httpHeaderName<N> / secureJsonData.httpHeaderValue<N>) are
	// not modeled here because they are dynamically indexed.
	TLSAuth           bool     `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool     `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify     bool     `json:"tlsSkipVerify,omitempty"`
	ServerName        string   `json:"serverName,omitempty"`
	Timeout           float64  `json:"timeout,omitempty"`
	KeepCookies       []string `json:"keepCookies,omitempty"`
	OauthPassThru     bool     `json:"oauthPassThru,omitempty"`

	// InfluxDB-specific jsonData fields — mirror
	// pkg/influxdb/models/datasource_info.go:17-27 verbatim (same field
	// names, same json tags).
	DbName        string         `json:"dbName,omitempty"`
	Version       InfluxVersion  `json:"version,omitempty"`
	HTTPMode      InfluxHTTPMode `json:"httpMode,omitempty"`
	TimeInterval  string         `json:"timeInterval,omitempty"`
	ShowTagTime   string         `json:"showTagTime,omitempty"`
	DefaultBucket string         `json:"defaultBucket,omitempty"`
	Organization  string         `json:"organization,omitempty"`
	MaxSeries     int32          `json:"maxSeries,omitempty"`
	InsecureGrpc  bool           `json:"insecureGrpc,omitempty"`

	// v2 editor extras — not read by the backend.
	Product     InfluxProduct `json:"product,omitempty"`
	PdcInjected bool          `json:"pdcInjected,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (token, password, basicAuthPassword, tlsCACert, tlsClientCert,
	// tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser, User, Database, WithCredentials)
// are copied from backend.DataSourceInstanceSettings directly; jsonData is
// unmarshaled verbatim from settings.JSONData mirroring
// pkg/influxdb/influxdb.go:37-51 (which decodes into DatasourceInfo);
// decrypted secrets are copied by known key name into
// DecryptedSecureJSONData.
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

	logger.Debug("loading influxdb datasource config")

	cfg := Config{
		URL:                     settings.URL,
		BasicAuth:               settings.BasicAuthEnabled,
		BasicAuthUser:           settings.BasicAuthUser,
		User:                    settings.User,
		Database:                settings.Database,
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
		logger.Error("influxdb datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("influxdb datasource config loaded",
		"hasURL", cfg.URL != "",
		"version", cfg.Version,
		"httpMode", cfg.HTTPMode,
		"maxSeries", cfg.MaxSeries,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the backend applies at pkg/influxdb/influxdb.go:43-56 when
// instantiating a datasource. Never blanket-apply every schema default —
// that would clobber intentional zero values.
//
// Curated defaults (verbatim from the backend):
//   - Version: "InfluxQL" — pkg/influxdb/influxdb.go:53-56 (`version == "" -> "InfluxQL"`)
//   - HTTPMode: "GET" — pkg/influxdb/influxdb.go:43-46 (`httpMode == "" -> "GET"`)
//   - MaxSeries: 1000 — pkg/influxdb/influxdb.go:48-51 (`maxSeries == 0 -> 1000`)
//
// The backend's dbName fallback to settings.Database
// (pkg/influxdb/influxdb.go:58-61) is applied here too: an empty DbName is
// filled from the parsed root Database. This keeps callers that build a
// Config directly in parity with the backend's DatasourceInfo shape.
func (c *Config) ApplyDefaults() {
	if c.Version == "" {
		c.Version = DefaultInfluxVersion
	}
	if c.HTTPMode == "" {
		c.HTTPMode = DefaultInfluxHTTPMode
	}
	if c.MaxSeries == 0 {
		c.MaxSeries = DefaultMaxSeries
	}
	if c.DbName == "" && c.Database != "" {
		c.DbName = c.Database
	}
}

// Validate checks the runtime contract that the plugin requires. The
// InfluxDB backend reads settings.URL directly (pkg/influxdb/influxdb.go:72)
// and rejects a URL with an empty host or scheme
// (pkg/influxdb/influxql/influxql.go:149-151). This method encodes the URL
// requirement, the enum constraints on Version and HTTPMode, the per-version
// backend contract (InfluxQL/SQL need dbName; Flux needs organization +
// defaultBucket + token; SQL needs token), the non-negative bound on numeric
// fields, and the TLS + Basic-auth field pairs required by the SDK. Errors
// are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("InfluxDB URL (root.url) is required"))
	}

	switch c.Version {
	case "", InfluxVersionInfluxQL, InfluxVersionFlux, InfluxVersionSQL:
		// OK. Empty is accepted because callers may call Validate before
		// ApplyDefaults; LoadConfig always applies defaults first.
	default:
		errs = append(errs, fmt.Errorf("invalid version %q: must be one of InfluxQL, Flux, SQL", c.Version))
	}

	switch c.HTTPMode {
	case "", InfluxHTTPModeGET, InfluxHTTPModePOST:
		// OK.
	default:
		errs = append(errs, fmt.Errorf("invalid httpMode %q: must be GET or POST", c.HTTPMode))
	}

	// Per-version contract.
	switch c.Version {
	case InfluxVersionInfluxQL:
		if c.DbName == "" && c.Database == "" {
			errs = append(errs, errors.New("dbName (jsonData) is required when version is InfluxQL"))
		}
	case InfluxVersionFlux:
		if c.Organization == "" {
			errs = append(errs, errors.New("organization (jsonData) is required when version is Flux"))
		}
		if c.DefaultBucket == "" {
			errs = append(errs, errors.New("defaultBucket (jsonData) is required when version is Flux"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, errors.New("token (secureJsonData) is required when version is Flux"))
		}
	case InfluxVersionSQL:
		if c.DbName == "" {
			errs = append(errs, errors.New("dbName (jsonData) is required when version is SQL"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, errors.New("token (secureJsonData) is required when version is SQL"))
		}
	}

	if c.MaxSeries < 0 {
		errs = append(errs, fmt.Errorf("maxSeries must be non-negative, got %d", c.MaxSeries))
	}
	if c.Timeout < 0 {
		errs = append(errs, fmt.Errorf("timeout must be non-negative, got %v", c.Timeout))
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

	return errors.Join(errs...)
}
