// Package tempodatasource contains the configuration models for the Tempo
// datasource plugin (id: tempo).
package tempodatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "tempo"

// SpanBarType is the discriminator for the extra span-bar label rendered next
// to each span in the trace view. Mirrors the three literal values written by
// the editor Select (@grafana/o11y-ds-frontend SpanBarSettings.tsx:23-25 —
// constants `NONE = "None"`, `DURATION = "Duration"`, `TAG = "Tag"`).
type SpanBarType string

const (
	// SpanBarTypeNone hides the extra span-bar label.
	SpanBarTypeNone SpanBarType = "None"
	// SpanBarTypeDuration renders the span duration (default; shown as the
	// editor placeholder even when unset).
	SpanBarTypeDuration SpanBarType = "Duration"
	// SpanBarTypeTag renders the value of the tag named by SpanBar.Tag.
	SpanBarTypeTag SpanBarType = "Tag"
)

// TraceqlSearchScope is the TraceQL scope name used on a static search filter.
// Mirrors the frontend enum at src/dataquery.ts:105-113.
type TraceqlSearchScope string

const (
	TraceqlSearchScopeEvent           TraceqlSearchScope = "event"
	TraceqlSearchScopeInstrumentation TraceqlSearchScope = "instrumentation"
	TraceqlSearchScopeIntrinsic       TraceqlSearchScope = "intrinsic"
	TraceqlSearchScopeLink            TraceqlSearchScope = "link"
	TraceqlSearchScopeResource        TraceqlSearchScope = "resource"
	TraceqlSearchScopeSpan            TraceqlSearchScope = "span"
	TraceqlSearchScopeUnscoped        TraceqlSearchScope = "unscoped"
)

// TimeRangeForTagsSeconds enumerates the five allowed values for the
// `timeRangeForTags` selector (TagsTimeRangeSettings.tsx:15-21). The default
// is 30 minutes.
const (
	TimeRangeForTagsLast30Minutes = 1800
	TimeRangeForTagsLast3Hours    = 10800
	TimeRangeForTagsLast24Hours   = 86400
	TimeRangeForTagsLast3Days     = 259200
	TimeRangeForTagsLast7Days     = 604800
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

// StreamingEnabledConfig holds the two streaming toggles (StreamingSection.tsx:14-17,
// 44-77).
type StreamingEnabledConfig struct {
	// Search enables streaming for TraceQL search queries. Min Tempo: 2.2.0.
	Search bool `json:"search,omitempty"`
	// Metrics enables streaming for TraceQL metrics queries. Min Tempo: 2.7.0.
	Metrics bool `json:"metrics,omitempty"`
}

// TraceToLogsV1Config is the legacy v1 shape retained for round-trip parity
// only. `getTraceToLogsOptions`
// (packages/grafana-o11y-ds-frontend/src/TraceToLogs/TraceToLogsSettings.tsx:55-73)
// migrates it to TraceToLogsV2Config on read; the editor clears the underlying
// key whenever `tracesToLogsV2` is written.
type TraceToLogsV1Config struct {
	DatasourceUID      string           `json:"datasourceUid,omitempty"`
	Tags               []string         `json:"tags,omitempty"`
	MappedTags         []TraceToLogsTag `json:"mappedTags,omitempty"`
	MapTagNamesEnabled bool             `json:"mapTagNamesEnabled,omitempty"`
	SpanStartTimeShift string           `json:"spanStartTimeShift,omitempty"`
	SpanEndTimeShift   string           `json:"spanEndTimeShift,omitempty"`
	FilterByTraceID    bool             `json:"filterByTraceID,omitempty"`
	FilterBySpanID     bool             `json:"filterBySpanID,omitempty"`
	LokiSearch         bool             `json:"lokiSearch,omitempty"`
}

// TraceToLogsTag is one key/value tag mapping used by the trace-to-logs and
// trace-to-profiles editors (`TagMappingInput`). `value` is optional.
type TraceToLogsTag struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// TraceToLogsTagPair is the same shape but with a required `value`, used by
// trace-to-metrics (TraceToMetricsSettings.tsx:19).
type TraceToLogsTagPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TraceToLogsV2Config is the v2 shape written by TraceToLogsSection.
type TraceToLogsV2Config struct {
	DatasourceUID      string           `json:"datasourceUid,omitempty"`
	Tags               []TraceToLogsTag `json:"tags,omitempty"`
	SpanStartTimeShift string           `json:"spanStartTimeShift,omitempty"`
	SpanEndTimeShift   string           `json:"spanEndTimeShift,omitempty"`
	FilterByTraceID    bool             `json:"filterByTraceID,omitempty"`
	FilterBySpanID     bool             `json:"filterBySpanID,omitempty"`
	Query              string           `json:"query,omitempty"`
	CustomQuery        bool             `json:"customQuery"`
}

// TraceToMetricsQuery is one named Prometheus query offered on the
// trace→metrics link.
type TraceToMetricsQuery struct {
	Name  string `json:"name,omitempty"`
	Query string `json:"query,omitempty"`
}

// TraceToMetricsConfig is the shape written by TraceToMetricsSection.
type TraceToMetricsConfig struct {
	DatasourceUID      string                `json:"datasourceUid,omitempty"`
	Tags               []TraceToLogsTagPair  `json:"tags,omitempty"`
	Queries            []TraceToMetricsQuery `json:"queries,omitempty"`
	SpanStartTimeShift string                `json:"spanStartTimeShift,omitempty"`
	SpanEndTimeShift   string                `json:"spanEndTimeShift,omitempty"`
}

// TraceToProfilesConfig is the shape written by TraceToProfilesSection.
type TraceToProfilesConfig struct {
	DatasourceUID string           `json:"datasourceUid,omitempty"`
	Tags          []TraceToLogsTag `json:"tags,omitempty"`
	ProfileTypeID string           `json:"profileTypeId,omitempty"`
	Query         string           `json:"query,omitempty"`
	CustomQuery   bool             `json:"customQuery"`
}

// ServiceMapConfig points at the Prometheus datasource that stores the
// service-graph metrics (ServiceGraphSettings.tsx:32-37).
type ServiceMapConfig struct {
	DatasourceUID string `json:"datasourceUid,omitempty"`
}

// NodeGraphConfig toggles the node-graph view above the trace view
// (NodeGraphSettings.tsx:22-46).
type NodeGraphConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// SpanBarConfig picks the extra label rendered next to service/operation on
// each span row (SpanBarSettings.tsx).
type SpanBarConfig struct {
	Type SpanBarType `json:"type,omitempty"`
	Tag  string      `json:"tag,omitempty"`
}

// TraceqlFilter is a single static TraceQL search filter (dataquery.ts:115-142).
type TraceqlFilter struct {
	// ID uniquely identifies the filter within the editor; not used in query generation.
	ID            string             `json:"id"`
	IsCustomValue bool               `json:"isCustomValue,omitempty"`
	Operator      string             `json:"operator,omitempty"`
	Scope         TraceqlSearchScope `json:"scope,omitempty"`
	Tag           string             `json:"tag,omitempty"`
	// Value is either a single string or a []string depending on the operator.
	// Modeled as `any` so the raw shape round-trips through LoadConfig.
	Value     any    `json:"value,omitempty"`
	ValueType string `json:"valueType,omitempty"`
}

// SearchConfig is the shape written by TraceQLSearchSettings.tsx:28-46.
type SearchConfig struct {
	// Hide removes the TraceQL search tab from the query editor.
	Hide bool `json:"hide,omitempty"`
	// Filters are pre-configured static filters exposed in the search UI.
	Filters []TraceqlFilter `json:"filters,omitempty"`
}

// TraceQueryConfig is the shape written by QuerySettings.tsx (TraceID query
// time-range shifts).
type TraceQueryConfig struct {
	TimeShiftEnabled   bool   `json:"timeShiftEnabled,omitempty"`
	SpanStartTimeShift string `json:"spanStartTimeShift,omitempty"`
	SpanEndTimeShift   string `json:"spanEndTimeShift,omitempty"`
}

// Config is the fully loaded configuration of a Tempo datasource instance.
//
// The Tempo backend's server-side settings reads are minimal:
//   - pkg/tempo/tempo.go:52-90 (NewDatasource) — reads settings.URL and delegates the
//     rest to settings.HTTPClientOptions (SDK) and newGrpcClient.
//   - pkg/tempo/tempo.go:150-208 (CheckHealth) — decodes JSONData as
//     map[string]interface{} to probe streamingEnabled.search.
//   - pkg/tempo/grpc.go:85-137 — reads settings.URL, settings.BasicAuthEnabled,
//     and settings.ProxyClient to configure the gRPC streaming client.
//
// The Tempo plugin ships no pkg/models/settings.go, so the jsonData shape on
// this struct is the intended settings model: it mirrors what the editor
// writes and what a Grafana-side caller needs to know about a Tempo
// datasource instance.
type Config struct {
	// Root-level fields (json:"-" so they are not part of jsonData). URL is
	// read by both the HTTP client (pkg/tempo/tempo.go:78) and the gRPC client
	// (pkg/tempo/grpc.go:86); BasicAuth / BasicAuthUser / WithCredentials are
	// consumed by settings.HTTPClientOptions() at pkg/tempo/tempo.go:54 and
	// BasicAuth also gates gRPC per-RPC credentials (pkg/tempo/grpc.go:178-184).
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

	StreamingEnabled StreamingEnabledConfig `json:"streamingEnabled,omitempty"`

	TracesToLogsV2 TraceToLogsV2Config `json:"tracesToLogsV2,omitempty"`
	TracesToLogs   TraceToLogsV1Config `json:"tracesToLogs,omitempty"`

	TracesToMetrics  TraceToMetricsConfig  `json:"tracesToMetrics,omitempty"`
	TracesToProfiles TraceToProfilesConfig `json:"tracesToProfiles,omitempty"`

	ServiceMap ServiceMapConfig `json:"serviceMap,omitempty"`
	NodeGraph  NodeGraphConfig  `json:"nodeGraph,omitempty"`
	SpanBar    SpanBarConfig    `json:"spanBar,omitempty"`

	Search SearchConfig `json:"search,omitempty"`

	TraceQuery TraceQueryConfig `json:"traceQuery,omitempty"`

	// TimeRangeForTags is one of the five allowed second-counts (default 1800).
	TimeRangeForTags int64 `json:"timeRangeForTags,omitempty"`

	// TagLimit — TagLimitSettings.tsx writes v.currentTarget.value from an
	// Input type="number", so the persisted value is technically a string in
	// storage. We accept both by using json.Number-style tolerant parsing in
	// UnmarshalJSON below.
	TagLimit int64 `json:"tagLimit,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// UnmarshalJSON is lenient about `tagLimit`: the editor's `<Input type="number">`
// binds `v.currentTarget.value` (TagLimitSettings.tsx:32-34), so provisioning
// payloads and legacy datasources may store the value as either a number or a
// numeric string. Everything else is parsed by the default decoder.
func (c *Config) UnmarshalJSON(data []byte) error {
	type alias Config
	tmp := struct {
		*alias
		TagLimit json.RawMessage `json:"tagLimit,omitempty"`
	}{alias: (*alias)(c)}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if len(tmp.TagLimit) == 0 || string(tmp.TagLimit) == "null" {
		return nil
	}
	// Try number first, then string (which the editor writes).
	var n int64
	if err := json.Unmarshal(tmp.TagLimit, &n); err == nil {
		c.TagLimit = n
		return nil
	}
	var s string
	if err := json.Unmarshal(tmp.TagLimit, &s); err != nil {
		return fmt.Errorf("tagLimit: expected number or numeric string, got %s", string(tmp.TagLimit))
	}
	if s == "" {
		return nil
	}
	// Trim through fmt.Sscanf so we accept "5000" and "5000 " alike.
	var v int64
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		return fmt.Errorf("tagLimit: %q is not a valid integer", s)
	}
	c.TagLimit = v
	return nil
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser) are copied from
// backend.DataSourceInstanceSettings directly; jsonData is unmarshaled from
// settings.JSONData via Config.UnmarshalJSON; decrypted secrets are copied by
// known key name into DecryptedSecureJSONData.
//
// The Tempo plugin has no upstream `LoadSettings` equivalent to mirror — the
// Go code only reads settings.URL and decodes JSONData as a bare map in
// CheckHealth (see the package doc on Config). LoadConfig therefore
// represents the intended, flat shape a Grafana-side caller needs.
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

	logger.Debug("loading tempo datasource config")

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
		logger.Error("tempo datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("tempo datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"streamingSearch", cfg.StreamingEnabled.Search,
		"streamingMetrics", cfg.StreamingEnabled.Metrics,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// Tempo's editor writes exactly one hard default into jsonData: `timeRangeForTags`
// falls back to DEFAULT_TIME_RANGE_FOR_TAGS (1800) at read time
// (TagsTimeRangeSettings.tsx:30-36). Everything else — the "5000" tagLimit
// placeholder (TagLimitSettings.tsx:31), the "Duration" spanBar placeholder
// (SpanBarSettings.tsx:44), the default TraceQL filters seeded on first load
// (datasource.ts:147-159) — is either a render-time `??` fallback or seeded by
// the frontend datasource constructor, never persisted. We apply the
// `timeRangeForTags` default so callers assembling a Config directly get the
// same effective value the editor would use for tags queries.
func (c *Config) ApplyDefaults() {
	if c.TimeRangeForTags == 0 {
		c.TimeRangeForTags = TimeRangeForTagsLast30Minutes
	}
}

// Validate checks the runtime contract that the plugin requires. Errors are
// joined so callers see every problem at once.
//
// The Tempo backend does not pre-validate settings — it just fails at request
// time when URL parsing fails (pkg/tempo/tempo.go:211-216, grpc.go:86-89) or
// when a request runs. We surface the essentials at load time so provisioning
// tooling can reject misconfigurations upfront:
//
//   - URL is required (pkg/tempo/tempo.go:78, grpc.go:86).
//   - Basic auth requires a username (@grafana/plugin-ui BasicAuth.tsx:24-27
//     forces both fields when the method is selected).
//   - mTLS requires serverName + client cert + client key.
//   - Self-signed CA verification requires the CA PEM.
//   - `timeout` must be non-negative.
//   - `spanBar.type` must be one of the three allowed literals when set.
//   - `timeRangeForTags` must be one of the five allowed second-counts.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Tempo URL (root.url) is required"))
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

	switch c.SpanBar.Type {
	case "", SpanBarTypeNone, SpanBarTypeDuration, SpanBarTypeTag:
		// OK; empty is accepted because the editor renders it as the
		// "Duration" placeholder without persisting.
	default:
		errs = append(errs, fmt.Errorf(
			"spanBar.type %q: must be %q, %q, or %q",
			c.SpanBar.Type, SpanBarTypeNone, SpanBarTypeDuration, SpanBarTypeTag,
		))
	}
	if c.SpanBar.Type == SpanBarTypeTag && c.SpanBar.Tag == "" {
		errs = append(errs, errors.New(`spanBar.tag is required when spanBar.type is "Tag"`))
	}

	switch c.TimeRangeForTags {
	case 0,
		TimeRangeForTagsLast30Minutes,
		TimeRangeForTagsLast3Hours,
		TimeRangeForTagsLast24Hours,
		TimeRangeForTagsLast3Days,
		TimeRangeForTagsLast7Days:
		// OK. 0 slips through here because Validate runs after ApplyDefaults
		// which sets the default; the explicit 0 case exists for callers that
		// bypass ApplyDefaults.
	default:
		errs = append(errs, fmt.Errorf(
			"timeRangeForTags %d: must be one of %d, %d, %d, %d, %d (seconds)",
			c.TimeRangeForTags,
			TimeRangeForTagsLast30Minutes,
			TimeRangeForTagsLast3Hours,
			TimeRangeForTagsLast24Hours,
			TimeRangeForTagsLast3Days,
			TimeRangeForTagsLast7Days,
		))
	}

	if c.TagLimit < 0 {
		errs = append(errs, fmt.Errorf("tagLimit must be non-negative, got %d", c.TagLimit))
	}

	return errors.Join(errs...)
}
