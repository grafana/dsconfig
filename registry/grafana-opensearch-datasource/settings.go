// Package opensearchdatasource contains the configuration models for the
// OpenSearch datasource plugin (id: grafana-opensearch-datasource).
package opensearchdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "grafana-opensearch-datasource"

// defaultTimeField mirrors the frontend `coerceOptions` fallback in
// `configuration/utils.ts:20` ("@timestamp" when the field is empty).
const defaultTimeField = "@timestamp"

// defaultMaxConcurrentShardRequestsOpenSearch mirrors both the frontend fallback
// (`defaultMaxConcurrentShardRequests` in `OpenSearchDetails.tsx:322-327`) and the
// backend fallback (`getMultiSearchQueryParameters` in `client.go:428`) for
// OpenSearch or Elasticsearch >=7.0.0.
const defaultMaxConcurrentShardRequestsOpenSearch = int64(5)

// defaultMaxConcurrentShardRequestsESLegacy mirrors the same fallback for
// Elasticsearch <7.0.0 (`OpenSearchDetails.tsx:326`, `client.go:411`).
const defaultMaxConcurrentShardRequestsESLegacy = int64(256)

// Flavor identifies the OpenSearch / Elasticsearch flavor of the target
// cluster. Mirrors `pkg/opensearch/client/models.go:13-18`.
type Flavor string

const (
	FlavorOpenSearch    Flavor = "opensearch"
	FlavorElasticsearch Flavor = "elasticsearch"
)

// Interval is the time-based index pattern selector. Empty means "No pattern".
// Mirrors `OpenSearchDetails.tsx:10-17`.
type Interval string

const (
	IntervalHourly  Interval = "Hourly"
	IntervalDaily   Interval = "Daily"
	IntervalWeekly  Interval = "Weekly"
	IntervalMonthly Interval = "Monthly"
	IntervalYearly  Interval = "Yearly"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the Basic-auth password, set
	// when root.basicAuth is true (`BasicAuthSettings.mjs:23-31`).
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyTLSCACert is the custom CA PEM, set when
	// jsonData.tlsAuthWithCACert is true (`TLSAuthSettings.mjs:62-75`).
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
	// SecureJsonDataKeyTLSClientCert is the mTLS client certificate PEM,
	// set when jsonData.tlsAuth is true (`TLSAuthSettings.mjs:88-101`).
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the mTLS client key PEM, set when
	// jsonData.tlsAuth is true (`TLSAuthSettings.mjs:102-115`).
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin's editor.
//
// Note: @grafana/ui's CustomHeadersSettings component writes indexed
// httpHeaderValue<N> secrets when custom HTTP headers are configured;
// @grafana/aws-sdk's SIGV4ConnectionConfig writes sigV4AccessKey and
// sigV4SecretKey when the SigV4 auth method is selected. The backend also
// reads a legacy secureJsonData.password when root.user is set and
// root.basicAuth is false (client.go:294-298). None of those categories is
// represented here because they are either contributed by external components
// or backend-only paths without an editor field.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// DataLinkConfig mirrors the plugin's `DataLinkConfig`
// (`src/types.ts:101-106`). When DatasourceUID is set the URL string is
// treated as a query for that internal data source; otherwise URL is an
// external link template. Title is the (optional) label shown next to the
// link. Note the OpenSearch shape uses `Title` (not `URLDisplayLabel`).
type DataLinkConfig struct {
	Field         string `json:"field"`
	URL           string `json:"url,omitempty"`
	Title         string `json:"title,omitempty"`
	DatasourceUID string `json:"datasourceUid,omitempty"`
}

// Config is the fully loaded configuration of an OpenSearch datasource
// instance.
//
// The OpenSearch plugin has no upstream `pkg/models/settings.go` — its
// `NewOpenSearchDatasource` (`pkg/opensearch/opensearch.go:28-39`) delegates
// to `client.NewDatasourceHttpClient` which unmarshals a small subset
// (serverless, oauthPassThru) into an anonymous struct, and its per-query
// `client.NewClient` (`pkg/opensearch/client/client.go:96-153`) reads the
// remaining jsonData keys through `simplejson`. This Config therefore
// represents the shape the editor writes and the backend consumes, not a
// plugin-owned settings model.
//
// Root-level fields the editor writes and the SDK's
// `backend.DataSourceInstanceSettings` exposes (URL, BasicAuth, BasicAuthUser,
// User, Database) are carried with `json:"-"` tags so they don't collide with
// jsonData unmarshaling. URL is read by the plugin's own code (opensearch.go:99,
// client.go:245,479); BasicAuth / BasicAuthUser / User are honored at
// client.go:288-298,520-530; Database is a legacy root-level echo written by
// the editor's intervalHandler (OpenSearchDetails.tsx:269) but never read by
// the backend. The editor also writes `root.access` (proxy/direct) and
// `root.withCredentials`, but the SDK does not surface them on
// DataSourceInstanceSettings so they are not modeled on Config; they are still
// declared in `dsconfig.json` because provisioning payloads can set them.
type Config struct {
	// Root-level fields (json:"-" — not part of jsonData).
	URL           string `json:"-"`
	BasicAuth     bool   `json:"-"`
	BasicAuthUser string `json:"-"`
	User          string `json:"-"`
	Database      string `json:"-"`

	// jsonData — auth toggles (independent booleans, no discriminator).
	OauthPassThru bool `json:"oauthPassThru,omitempty"`
	SigV4Auth     bool `json:"sigV4Auth,omitempty"`

	// jsonData — TLS (written by @grafana/ui HttpProxySettings + TLSAuthSettings).
	TLSAuth           bool   `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool   `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify     bool   `json:"tlsSkipVerify,omitempty"`
	ServerName        string `json:"serverName,omitempty"`

	// jsonData — HTTP (written by @grafana/ui DataSourceHttpSettings, only when
	// access='proxy').
	Timeout     float64  `json:"timeout,omitempty"`
	KeepCookies []string `json:"keepCookies,omitempty"`

	// jsonData — OpenSearch-specific storage (OpenSearchDetails.tsx).
	JSONDatabase               string   `json:"database,omitempty"`
	IntervalPattern            Interval `json:"interval,omitempty"`
	TimeField                  string   `json:"timeField,omitempty"`
	Serverless                 bool     `json:"serverless,omitempty"`
	Flavor                     Flavor   `json:"flavor,omitempty"`
	Version                    string   `json:"version,omitempty"`
	VersionLabel               string   `json:"versionLabel,omitempty"`
	MaxConcurrentShardRequests int64    `json:"maxConcurrentShardRequests,omitempty"`
	TimeInterval               string   `json:"timeInterval,omitempty"`
	PPLEnabled                 bool     `json:"pplEnabled,omitempty"`

	// jsonData — Logs sub-section (LogsConfig.tsx).
	LogMessageField string `json:"logMessageField,omitempty"`
	LogLevelField   string `json:"logLevelField,omitempty"`

	// jsonData — Data links (DataLinks.tsx).
	DataLinks []DataLinkConfig `json:"dataLinks,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// UnmarshalJSON tolerates `maxConcurrentShardRequests` being stored either as a
// JSON number (float64 on the wire) or as a JSON string. Mirrors the backend's
// use of `simplejson.Get("maxConcurrentShardRequests").MustInt(...)` at
// `client.go:411,428` which accepts both wire representations transparently.
func (c *Config) UnmarshalJSON(data []byte) error {
	type alias Config
	aux := struct {
		MaxConcurrentShardRequests json.RawMessage `json:"maxConcurrentShardRequests,omitempty"`
		*alias
	}{alias: (*alias)(c)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.MaxConcurrentShardRequests = 0
	if len(aux.MaxConcurrentShardRequests) > 0 {
		var n float64
		if err := json.Unmarshal(aux.MaxConcurrentShardRequests, &n); err == nil {
			c.MaxConcurrentShardRequests = int64(n)
		} else {
			var s string
			if err := json.Unmarshal(aux.MaxConcurrentShardRequests, &s); err == nil {
				var parsed int64
				if _, perr := fmt.Sscanf(s, "%d", &parsed); perr == nil {
					c.MaxConcurrentShardRequests = parsed
				}
				// Silently fall through to 0 when parsing fails — LoadConfig's
				// ApplyDefaults will then apply the flavor/version-aware default.
			}
		}
	}
	return nil
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, Access, BasicAuth, BasicAuthUser, WithCredentials, User,
// Database) are copied from backend.DataSourceInstanceSettings directly;
// jsonData is unmarshaled from settings.JSONData; decrypted secrets are
// copied by known key name into DecryptedSecureJSONData.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. It mirrors the sequence the plugin's own health check
// (`opensearch.go:45-174`) and its client factory
// (`client.go:96-153`) use, including the serverless overrides (flavor forced
// to "opensearch", version forced to "1.0.0", maxConcurrentShardRequests forced
// to 5, pplEnabled forced to true) applied by the editor's serverless toggle
// at `OpenSearchDetails.tsx:66-80`.
//
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on a Config they assemble themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading opensearch datasource config")

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
		logger.Error("opensearch datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("opensearch datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"sigV4Auth", cfg.SigV4Auth,
		"tlsAuth", cfg.TLSAuth,
		"flavor", cfg.Flavor,
		"version", cfg.Version,
		"serverless", cfg.Serverless,
		"database", cfg.JSONDatabase,
		"timeField", cfg.TimeField,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource (`coerceOptions` in
// `src/configuration/utils.ts:6-30`) and the backend applies at load
// (`client.go:411,428`). Never blanket-apply every schema default — that
// would clobber intentional zero values.
//
// Curated defaults:
//   - TimeField: "@timestamp" — the frontend's `coerceOptions` writes it when
//     empty (utils.ts:20). The backend also hard-fails on an empty timeField
//     (opensearch.go:70-75), so this is not merely cosmetic.
//   - MaxConcurrentShardRequests: 5 (or 256 for Elasticsearch <7.0.0) — both
//     the frontend fallback (`defaultMaxConcurrentShardRequests`) and the
//     backend's non-positive coercion at `client.go:411-433`.
//   - PPLEnabled: true — the frontend `coerceOptions` fallback (utils.ts:27)
//     and the switch's default rendering (OpenSearchDetails.tsx:218).
//   - Serverless overrides: when Serverless is true, force Flavor="opensearch",
//     Version="1.0.0", MaxConcurrentShardRequests=5, PPLEnabled=true.
//     Mirrors `getServerlessSettings` at OpenSearchDetails.tsx:66-80.
//
// Other apparent "defaults" (logMessageField=”, logLevelField=”) live only
// in the editor's coercion; they are the JSON zero values so applying them
// here would be a no-op.
func (c *Config) ApplyDefaults() {
	if c.TimeField == "" {
		c.TimeField = defaultTimeField
	}
	if c.Serverless {
		c.Flavor = FlavorOpenSearch
		if c.Version == "" {
			c.Version = "1.0.0"
		}
		if c.MaxConcurrentShardRequests == 0 {
			c.MaxConcurrentShardRequests = defaultMaxConcurrentShardRequestsOpenSearch
		}
	}
	if c.MaxConcurrentShardRequests == 0 {
		c.MaxConcurrentShardRequests = defaultMaxConcurrentShardRequestsFor(c.Flavor, c.Version)
	}
	// PPLEnabled is frontend-only (`OpenSearchDetails.tsx:218` +
	// `client.go:557-560` chooses the PPL endpoint by flavor, not by this
	// toggle). The editor's `coerceOptions` defaults it to true on mount, but
	// we cannot mirror that in Go: `bool` cannot distinguish "not set" from
	// "explicitly false", and applying a default here would clobber a caller's
	// intentional `false`. Callers who care about the editor-parity default
	// should set PPLEnabled=true themselves.
}

// defaultMaxConcurrentShardRequestsFor returns 256 for Elasticsearch <7.0.0
// and 5 for OpenSearch or Elasticsearch >=7.0.0. Mirrors the frontend
// `defaultMaxConcurrentShardRequests` in OpenSearchDetails.tsx:322-327 and
// the backend's coercions in client.go:411,428.
//
// The upstream version-comparison uses semver's `lt`; here we accept the
// zero-value fallback of 5 whenever the version does not parse or is empty,
// matching the backend's `MustInt(5)` behavior.
func defaultMaxConcurrentShardRequestsFor(flavor Flavor, version string) int64 {
	if flavor == FlavorElasticsearch && isElasticsearchPre7(version) {
		return defaultMaxConcurrentShardRequestsESLegacy
	}
	return defaultMaxConcurrentShardRequestsOpenSearch
}

// isElasticsearchPre7 returns true when version is a valid semver string with
// major < 7. Anything else (empty, non-semver, "7.x", "opensearch-1.0.0")
// returns false so the caller falls back to the OpenSearch default.
func isElasticsearchPre7(version string) bool {
	if len(version) == 0 {
		return false
	}
	var major int
	if _, err := fmt.Sscanf(version, "%d.", &major); err != nil {
		return false
	}
	return major < 7
}

// Validate checks the runtime contract that the plugin requires. The backend's
// health check hard-fails on:
//   - missing or unknown flavor (opensearch.go:56-61)
//   - unparseable version (opensearch.go:63-68)
//   - empty timeField (opensearch.go:70-75)
//
// Beyond that, the plugin honors the SDK's HTTPClientOptions transport
// builder for auth/TLS wiring. This method encodes the URL requirement
// (implicit in every HTTP-based datasource), the flavor / version /
// timeField requirements, the enum constraints on Interval / Access /
// Flavor, and the per-auth-method contracts on secureJsonData. Errors are
// joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("OpenSearch URL (root.url) is required"))
	}

	switch c.Flavor {
	case FlavorOpenSearch, FlavorElasticsearch:
		// OK.
	case "":
		errs = append(errs, errors.New("flavor (jsonData.flavor) is required (health check hard-fails without it)"))
	default:
		errs = append(errs, fmt.Errorf("invalid flavor %q: must be opensearch or elasticsearch", c.Flavor))
	}

	if c.Version == "" {
		errs = append(errs, errors.New("version (jsonData.version) is required"))
	}

	if c.TimeField == "" {
		errs = append(errs, errors.New("timeField (jsonData.timeField) is required"))
	}

	switch c.IntervalPattern {
	case "", IntervalHourly, IntervalDaily, IntervalWeekly, IntervalMonthly, IntervalYearly:
		// OK. Empty means "No pattern".
	default:
		errs = append(errs, fmt.Errorf("invalid interval %q: must be one of Hourly, Daily, Weekly, Monthly, Yearly",
			c.IntervalPattern))
	}

	if c.BasicAuth {
		if c.BasicAuthUser == "" {
			errs = append(errs, errors.New("basicAuthUser (root) is required when basicAuth is true"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword] == "" {
			errs = append(errs, errors.New("basicAuthPassword (secureJsonData) is required when basicAuth is true"))
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
	if c.MaxConcurrentShardRequests < 0 {
		errs = append(errs, fmt.Errorf("maxConcurrentShardRequests must be non-negative, got %d", c.MaxConcurrentShardRequests))
	}

	return errors.Join(errs...)
}
