// Package elasticsearchdatasource contains the configuration models for the
// Elasticsearch datasource plugin (id: elasticsearch).
package elasticsearchdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "elasticsearch"

// defaultMaxConcurrentShardRequests mirrors both the frontend fallback
// (`defaultMaxConcurrentShardRequests` in `ElasticDetails.tsx:234-236`) and the
// backend fallback (`defaultMaxConcurrentShardRequests` in `elasticsearch.go:69`).
const defaultMaxConcurrentShardRequests = int64(5)

// defaultTimeField mirrors the frontend `coerceOptions` fallback in
// `configuration/utils.ts:21` ("@timestamp" when the field is empty).
const defaultTimeField = "@timestamp"

// Interval is the time-based index pattern selector. Empty means "No pattern".
// Mirrors `src/types.ts:60`.
type Interval string

const (
	IntervalHourly  Interval = "Hourly"
	IntervalDaily   Interval = "Daily"
	IntervalWeekly  Interval = "Weekly"
	IntervalMonthly Interval = "Monthly"
	IntervalYearly  Interval = "Yearly"
)

// QueryType is the default query mode selector. Mirrors `src/types.ts:84`.
type QueryType string

const (
	QueryTypeMetrics     QueryType = "metrics"
	QueryTypeLogs        QueryType = "logs"
	QueryTypeRawData     QueryType = "raw_data"
	QueryTypeRawDocument QueryType = "raw_document"
)

// DefaultQueryMode is the fallback written by the editor's `coerceOptions`
// (configuration/utils.ts:26 -> ElasticDetails.tsx:237-239).
const DefaultQueryMode = QueryTypeMetrics

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the Basic-auth password, set when
	// root.basicAuth is true.
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyAPIKey is the Elasticsearch API key, set when
	// jsonData.apiKeyAuth is true. The backend sends it as
	// `Authorization: ApiKey <value>` (pkg/elasticsearch/elasticsearch.go:125-131).
	SecureJsonDataKeyAPIKey SecureJsonDataKey = "apiKey"
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
// Note: @grafana/plugin-ui's CustomHeaders component writes indexed
// httpHeaderValue<N> secrets when custom HTTP headers are configured;
// @grafana/aws-sdk's SIGV4ConnectionConfig writes sigV4AccessKey and
// sigV4SecretKey when the SigV4 auth method is selected. Neither is
// represented here because they are contributed by external components.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyAPIKey,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// DataLinkConfig mirrors the plugin's `DataLinkConfig`
// (`src/types.ts:135-140`). When DatasourceUID is set the URL string is
// treated as a query for that internal data source; otherwise URL is an
// external link template.
type DataLinkConfig struct {
	Field           string `json:"field"`
	URL             string `json:"url,omitempty"`
	URLDisplayLabel string `json:"urlDisplayLabel,omitempty"`
	DatasourceUID   string `json:"datasourceUid,omitempty"`
}

// Config is the fully loaded configuration of an Elasticsearch datasource
// instance.
//
// The Elasticsearch plugin has no upstream `pkg/models/settings.go` — its
// `NewDatasource` (pkg/elasticsearch/elasticsearch.go:106-243) unmarshals
// jsonData as a raw `map[string]any` and only accesses a curated list of
// keys. This Config therefore represents the shape the editor writes and
// the backend consumes, not a plugin-owned settings model.
//
// Root-level fields the editor writes (URL, BasicAuth, BasicAuthUser,
// WithCredentials, Database) are carried with `json:"-"` tags so they don't
// collide with jsonData unmarshaling. URL and Database are read by the
// plugin's own Go code (elasticsearch.go:196,213 for URL;
// elasticsearch.go:169 for the legacy Database fallback); BasicAuth,
// BasicAuthUser, and WithCredentials are honored via the SDK's
// `HTTPClientOptions` transport builder.
type Config struct {
	// Root-level fields (json:"-" — not part of jsonData).
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`
	Database        string `json:"-"`

	// jsonData — auth discriminators (managed by the virtual authMethod selector).
	OauthPassThru bool `json:"oauthPassThru,omitempty"`
	APIKeyAuth    bool `json:"apiKeyAuth,omitempty"`
	SigV4Auth     bool `json:"sigV4Auth,omitempty"`

	// jsonData — TLS (written by @grafana/plugin-ui Auth/TLSSettings).
	TLSAuth           bool   `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool   `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify     bool   `json:"tlsSkipVerify,omitempty"`
	ServerName        string `json:"serverName,omitempty"`

	// jsonData — Advanced HTTP (written by @grafana/plugin-ui AdvancedHttpSettings).
	Timeout     float64  `json:"timeout,omitempty"`
	KeepCookies []string `json:"keepCookies,omitempty"`

	// jsonData — Elasticsearch details (ElasticDetails.tsx).
	Index                      string    `json:"index,omitempty"`
	IntervalPattern            Interval  `json:"interval,omitempty"`
	TimeField                  string    `json:"timeField,omitempty"`
	MaxConcurrentShardRequests int64     `json:"maxConcurrentShardRequests,omitempty"`
	TimeInterval               string    `json:"timeInterval,omitempty"`
	IncludeFrozen              bool      `json:"includeFrozen,omitempty"`
	DefaultQueryMode           QueryType `json:"defaultQueryMode,omitempty"`

	// jsonData — Logs sub-section (LogsConfig.tsx).
	LogMessageField string `json:"logMessageField,omitempty"`
	LogLevelField   string `json:"logLevelField,omitempty"`

	// jsonData — Data links (DataLinks.tsx).
	DataLinks []DataLinkConfig `json:"dataLinks,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, apiKey, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// UnmarshalJSON tolerates maxConcurrentShardRequests being stored either as a
// JSON number (float64 on the wire) or as a JSON string. Mirrors the switch
// on `jsonData["maxConcurrentShardRequests"]` in the upstream backend
// (pkg/elasticsearch/elasticsearch.go:172-185).
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
		// Try JSON number first.
		var n float64
		if err := json.Unmarshal(aux.MaxConcurrentShardRequests, &n); err == nil {
			c.MaxConcurrentShardRequests = int64(n)
		} else {
			// Fall back to JSON string containing a number.
			var s string
			if err := json.Unmarshal(aux.MaxConcurrentShardRequests, &s); err == nil {
				// Silently fall through to the "0" branch when parsing fails —
				// LoadConfig's ApplyDefaults will then set the default of 5,
				// which mirrors the backend's `default:` branch at
				// elasticsearch.go:184.
				var parsed int64
				if _, perr := fmt.Sscanf(s, "%d", &parsed); perr == nil {
					c.MaxConcurrentShardRequests = parsed
				}
			}
		}
	}
	return nil
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser, WithCredentials, Database) are
// copied from backend.DataSourceInstanceSettings directly; jsonData is
// unmarshaled from settings.JSONData; decrypted secrets are copied by known
// key name into DecryptedSecureJSONData.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. It mirrors the sequence the plugin's own `NewDatasource` uses
// (elasticsearch.go:106-243), including the legacy fallback that promotes
// settings.Database to jsonData.index when the latter is empty
// (elasticsearch.go:164-170) and the coercion of a missing / non-positive
// maxConcurrentShardRequests to 5.
//
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on a Config they assemble themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading elasticsearch datasource config")

	cfg := Config{
		URL:                     settings.URL,
		BasicAuth:               settings.BasicAuthEnabled,
		BasicAuthUser:           settings.BasicAuthUser,
		Database:                settings.Database,
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Legacy fallback: pre-jsonData datasources stored the index name in the
	// root-level `database` field (elasticsearch.go:164-170).
	if cfg.Index == "" && cfg.Database != "" {
		cfg.Index = cfg.Database
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("elasticsearch datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("elasticsearch datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"apiKeyAuth", cfg.APIKeyAuth,
		"sigV4Auth", cfg.SigV4Auth,
		"tlsAuth", cfg.TLSAuth,
		"index", cfg.Index,
		"timeField", cfg.TimeField,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource (`coerceOptions` in
// `src/configuration/utils.ts:14-29`) and the backend applies at load
// (`elasticsearch.go:172-189`). Never blanket-apply every schema default —
// that would clobber intentional zero values.
//
// Curated defaults:
//   - TimeField: "@timestamp" — the frontend's `coerceOptions` writes it when
//     empty (utils.ts:21). The backend also hard-fails on an empty timeField
//     (elasticsearch.go:145-147), so this is not merely cosmetic.
//   - MaxConcurrentShardRequests: 5 — both the frontend fallback
//     (`defaultMaxConcurrentShardRequests()`, ElasticDetails.tsx:234-236)
//     and the backend's non-positive coercion (elasticsearch.go:187-189).
//   - DefaultQueryMode: "metrics" — frontend `coerceOptions` fallback
//     (utils.ts:26 -> ElasticDetails.tsx:237-239).
//
// Other apparent "defaults" (includeFrozen=false, logMessageField=”,
// logLevelField=”) live only in the editor's coercion; they are the JSON
// zero values so applying them here would be a no-op.
func (c *Config) ApplyDefaults() {
	if c.TimeField == "" {
		c.TimeField = defaultTimeField
	}
	if c.MaxConcurrentShardRequests <= 0 {
		c.MaxConcurrentShardRequests = defaultMaxConcurrentShardRequests
	}
	if c.DefaultQueryMode == "" {
		c.DefaultQueryMode = DefaultQueryMode
	}
}

// Validate checks the runtime contract that the plugin requires. The
// backend's NewDatasource hard-fails on:
//   - empty timeField (elasticsearch.go:145-147: "elasticsearch time field name is required")
//   - a non-string timeField in the jsonData map (elasticsearch.go:140-143)
//
// Beyond that, the plugin honors the SDK's HTTPClientOptions transport
// builder for auth/TLS wiring. This method encodes the URL requirement
// (implicit in every HTTP-based datasource), the timeField requirement, the
// enum constraints on Interval / DefaultQueryMode, and the per-auth-method
// contracts on secureJsonData. Errors are joined so callers see every
// problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Elasticsearch URL (root.url) is required"))
	}

	// Editor-populated index: required by the editor UI (ElasticDetails.tsx:47);
	// the backend does not hard-fail on an empty index, it just runs an
	// index-less msearch which almost always returns nothing useful. We
	// mirror the editor's data contract, but callers assembling a Config
	// directly may skip it by pre-filling Index or Database.
	if c.Index == "" {
		errs = append(errs, errors.New("index (jsonData.index) is required"))
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

	switch c.DefaultQueryMode {
	case "", QueryTypeMetrics, QueryTypeLogs, QueryTypeRawData, QueryTypeRawDocument:
		// OK. Empty is accepted here because callers may call Validate before
		// ApplyDefaults; LoadConfig always applies defaults first.
	default:
		errs = append(errs, fmt.Errorf("invalid defaultQueryMode %q: must be metrics, logs, raw_data, or raw_document",
			c.DefaultQueryMode))
	}

	if c.BasicAuth {
		if c.BasicAuthUser == "" {
			errs = append(errs, errors.New("basicAuthUser (root) is required when basicAuth is true"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword] == "" {
			errs = append(errs, errors.New("basicAuthPassword (secureJsonData) is required when basicAuth is true"))
		}
	}

	if c.APIKeyAuth {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey] == "" {
			errs = append(errs, errors.New("apiKey (secureJsonData) is required when apiKeyAuth is true"))
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
