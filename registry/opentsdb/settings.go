// Package opentsdbdatasource contains the configuration models for the
// OpenTSDB datasource plugin (id: opentsdb).
package opentsdbdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "opentsdb"

// OpenTsdbVersion is the OpenTSDB "Version" selector — a numeric enum stored
// as a JSON number under jsonData.tsdbVersion. Options come from
// src/components/OpenTsdbDetails.tsx:8-13.
type OpenTsdbVersion float32

const (
	// OpenTsdbVersionLTE21 corresponds to the "<=2.1" option (value 1) and is
	// the default the frontend falls back to (src/datasource.ts:60).
	OpenTsdbVersionLTE21 OpenTsdbVersion = 1
	// OpenTsdbVersion22 corresponds to the "==2.2" option (value 2).
	OpenTsdbVersion22 OpenTsdbVersion = 2
	// OpenTsdbVersion23 corresponds to the "==2.3" option (value 3) and adds
	// showQuery=true on outgoing query bodies (src/datasource.ts:187-189).
	OpenTsdbVersion23 OpenTsdbVersion = 3
	// OpenTsdbVersion24 corresponds to the "==2.4" option (value 4) and
	// switches the response parser to the array shape
	// (pkg/opentsdb/utils.go:138,247-254).
	OpenTsdbVersion24 OpenTsdbVersion = 4
	// DefaultOpenTsdbVersion mirrors src/datasource.ts:60's `|| 1` fallback:
	// the frontend treats an empty/zero tsdbVersion as OpenTsdbVersionLTE21.
	DefaultOpenTsdbVersion OpenTsdbVersion = OpenTsdbVersionLTE21
)

// OpenTsdbResolution is the OpenTSDB "Resolution" selector — a numeric enum
// stored as a JSON number under jsonData.tsdbResolution. Options come from
// src/components/OpenTsdbDetails.tsx:15-18.
type OpenTsdbResolution int32

const (
	// OpenTsdbResolutionSecond corresponds to the "second" option (value 1)
	// and is the default the frontend falls back to (src/datasource.ts:61).
	OpenTsdbResolutionSecond OpenTsdbResolution = 1
	// OpenTsdbResolutionMillisecond corresponds to the "millisecond" option
	// (value 2); when selected the frontend adds msResolution=true to
	// outgoing query bodies (src/datasource.ts:178-180).
	OpenTsdbResolutionMillisecond OpenTsdbResolution = 2
	// DefaultOpenTsdbResolution mirrors src/datasource.ts:61's `|| 1` fallback.
	DefaultOpenTsdbResolution OpenTsdbResolution = OpenTsdbResolutionSecond
)

// DefaultLookupLimit mirrors src/datasource.ts:62's `|| 1000` fallback:
// the row cap the frontend uses when the editor never wrote a value.
const DefaultLookupLimit int32 = 1000

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

// Config is the fully loaded configuration of an OpenTSDB datasource instance.
//
// The three OpenTSDB-specific jsonData fields (TSDBVersion, TSDBResolution,
// LookupLimit) mirror the upstream backend JSONData struct verbatim
// (pkg/opentsdb/opentsdb.go:65-69 — same field names, same json tags, same
// numeric types: float32 / int32 / int32). The backend also reads
// settings.URL directly (pkg/opentsdb/opentsdb.go:47) and delegates
// everything else — TLS, cookies, timeouts, basic auth, OAuth forward — to
// settings.HTTPClientOptions(ctx).
//
// Root-level fields (URL, BasicAuth, BasicAuthUser, WithCredentials) are
// carried with json:"-" tags so they don't collide with jsonData
// unmarshaling. They are needed on Config so Validate() can enforce the
// paired requirements (basicAuthUser when basicAuth is true, etc.). Access is
// intentionally omitted because the OpenTSDB editor never exposes it (see
// settings.ts RootConfig for the note).
type Config struct {
	// Root-level fields (json:"-" — not part of jsonData). URL is required
	// and read by NewDatasource at pkg/opentsdb/opentsdb.go:47.
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`

	// jsonData fields the SDK reads via HTTPClientOptions (TLS, cookies,
	// timeout, OAuth forward). Custom HTTP header pairs
	// (jsonData.httpHeaderName<N> / secureJsonData.httpHeaderValue<N>) are
	// not modeled here because they are dynamically indexed.
	TLSAuth           bool     `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool     `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify     bool     `json:"tlsSkipVerify,omitempty"`
	ServerName        string   `json:"serverName,omitempty"`
	Timeout           float64  `json:"timeout,omitempty"`
	KeepCookies       []string `json:"keepCookies,omitempty"`
	OauthPassThru     bool     `json:"oauthPassThru,omitempty"`

	// OpenTSDB-specific jsonData fields mirroring
	// pkg/opentsdb/opentsdb.go:65-69 verbatim (no omitempty, matching the
	// upstream JSONData struct so a round-trip yields the same bytes).
	TSDBVersion    OpenTsdbVersion    `json:"tsdbVersion"`
	TSDBResolution OpenTsdbResolution `json:"tsdbResolution"`
	LookupLimit    int32              `json:"lookupLimit"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser, WithCredentials) are copied from
// backend.DataSourceInstanceSettings directly; jsonData is unmarshaled
// verbatim from settings.JSONData mirroring the upstream backend
// JSONData struct (pkg/opentsdb/opentsdb.go:39-42); decrypted secrets are
// copied by known key name into DecryptedSecureJSONData.
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

	logger.Debug("loading opentsdb datasource config")

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
		logger.Error("opentsdb datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("opentsdb datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"tlsAuth", cfg.TLSAuth,
		"tsdbVersion", cfg.TSDBVersion,
		"tsdbResolution", cfg.TSDBResolution,
		"lookupLimit", cfg.LookupLimit,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the frontend falls back to at src/datasource.ts:60-62. Never
// blanket-apply every schema default — that would clobber intentional zero
// values.
//
// Curated defaults:
//   - TSDBVersion: 1 (`<=2.1`) — src/datasource.ts:60 `|| 1`
//   - TSDBResolution: 1 (`second`) — src/datasource.ts:61 `|| 1`
//   - LookupLimit: 1000 — src/datasource.ts:62 `|| 1000`
//
// These mirror the runtime fallbacks in the plugin's OpenTsDatasource
// constructor: an untouched, provisioned datasource carries zeros in
// jsonData, and the frontend replaces them with the defaults above every
// time the datasource is instantiated. Applying the same defaults here gives
// LoadConfig callers editor-parity even before the editor has run.
func (c *Config) ApplyDefaults() {
	if c.TSDBVersion == 0 {
		c.TSDBVersion = DefaultOpenTsdbVersion
	}
	if c.TSDBResolution == 0 {
		c.TSDBResolution = DefaultOpenTsdbResolution
	}
	if c.LookupLimit == 0 {
		c.LookupLimit = DefaultLookupLimit
	}
}

// Validate checks the runtime contract that the plugin requires. The
// OpenTSDB backend reads settings.URL directly (pkg/opentsdb/opentsdb.go:47)
// and issues CheckHealth against {url}/api/suggest?q=cpu&type=metrics; an
// empty URL fails immediately. This method encodes the URL requirement, the
// enum constraints on TSDBVersion and TSDBResolution, the non-negative bound
// on LookupLimit and Timeout, and the TLS + Basic-auth field pairs required
// by @grafana/ui's TLSAuthSettings + BasicAuthSettings and the SDK. Errors
// are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("OpenTSDB URL (root.url) is required"))
	}

	switch c.TSDBVersion {
	case 0, OpenTsdbVersionLTE21, OpenTsdbVersion22, OpenTsdbVersion23, OpenTsdbVersion24:
		// OK. Zero is accepted because callers may call Validate before
		// ApplyDefaults; LoadConfig always applies defaults first.
	default:
		errs = append(errs, fmt.Errorf("invalid tsdbVersion %v: must be one of 1 (<=2.1), 2 (==2.2), 3 (==2.3), 4 (==2.4)", c.TSDBVersion))
	}

	switch c.TSDBResolution {
	case 0, OpenTsdbResolutionSecond, OpenTsdbResolutionMillisecond:
		// OK. Zero is the pre-defaults state.
	default:
		errs = append(errs, fmt.Errorf("invalid tsdbResolution %d: must be 1 (second) or 2 (millisecond)", c.TSDBResolution))
	}

	if c.LookupLimit < 0 {
		errs = append(errs, fmt.Errorf("lookupLimit must be non-negative, got %d", c.LookupLimit))
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
