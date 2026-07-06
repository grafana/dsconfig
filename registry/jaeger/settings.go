// Package jaegerdatasource contains the configuration models for the Jaeger
// datasource plugin (id: jaeger).
package jaegerdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "jaeger"

// SpanBarType is the discriminator for the extra span-bar label rendered next
// to each span in the trace view. Mirrors the three literal values written by
// the editor Select (@grafana/o11y-ds-frontend SpanBarSettings.tsx —
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

// TraceToLogsTag is one key/value tag mapping used by the trace-to-logs
// editor (`TagMappingInput`). `value` is optional.
type TraceToLogsTag struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// TraceToLogsTagPair is the same shape but with a required `value`, used by
// trace-to-metrics (TraceToMetricsSettings.tsx).
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

// NodeGraphConfig toggles the node-graph view above the trace view
// (NodeGraphSettings.tsx).
type NodeGraphConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// SpanBarConfig picks the extra label rendered next to service/operation on
// each span row (SpanBarSettings.tsx).
type SpanBarConfig struct {
	Type SpanBarType `json:"type,omitempty"`
	Tag  string      `json:"tag,omitempty"`
}

// TraceIdTimeParamsConfig is the plugin-local shape written by
// `src/configuration/TraceIdTimeParams.tsx:11-45` and read by
// `pkg/jaeger/client.go:242-266` (via `types.SettingsJSONData`).
type TraceIdTimeParamsConfig struct {
	// Enabled toggles whether `start` and `end` query parameters are
	// appended to `GET /api/traces/{traceID}`.
	Enabled bool `json:"enabled,omitempty"`
}

// Config is the fully loaded configuration of a Jaeger datasource instance.
//
// The Jaeger backend's server-side settings reads are minimal:
//   - pkg/jaeger/jaeger.go:27-55 (NewDatasource) — reads settings.URL and
//     unmarshals settings.JSONData into a nested `{ traceIdTimeParams: {
//     enabled: bool } }` shape; everything else is delegated to
//     settings.HTTPClientOptions.
//   - pkg/jaeger/client.go:242-266 (Trace) — decodes settings.JSONData again
//     as types.SettingsJSONData to gate the trace-by-ID time parameters.
//
// The plugin ships no pkg/models/settings.go and no typed LoadSettings — the
// jsonData shape on this struct is the intended settings model: it mirrors
// what the editor writes and what a Grafana-side caller needs to know about a
// Jaeger datasource instance.
type Config struct {
	// Root-level fields (json:"-" so they are not part of jsonData). URL is
	// read by NewDatasource (pkg/jaeger/jaeger.go:38-40) and by every client
	// method (pkg/jaeger/client.go:29). BasicAuth / BasicAuthUser /
	// WithCredentials are consumed by settings.HTTPClientOptions()
	// (pkg/jaeger/jaeger.go:28).
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

	TracesToLogsV2 TraceToLogsV2Config `json:"tracesToLogsV2,omitempty"`
	TracesToLogs   TraceToLogsV1Config `json:"tracesToLogs,omitempty"`

	TracesToMetrics TraceToMetricsConfig `json:"tracesToMetrics,omitempty"`

	NodeGraph NodeGraphConfig `json:"nodeGraph,omitempty"`
	SpanBar   SpanBarConfig   `json:"spanBar,omitempty"`

	// TraceIdTimeParams is the only jsonData field the backend actually
	// reads (pkg/jaeger/client.go:242-266).
	TraceIdTimeParams TraceIdTimeParamsConfig `json:"traceIdTimeParams,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, BasicAuth, BasicAuthUser) are copied from
// backend.DataSourceInstanceSettings directly; jsonData is unmarshaled from
// settings.JSONData with the default decoder; decrypted secrets are copied by
// known key name into DecryptedSecureJSONData.
//
// The Jaeger plugin has no upstream `LoadSettings` equivalent to mirror —
// pkg/jaeger/jaeger.go:27-55 (NewDatasource) is the only server-side read of
// settings and it just uses settings.URL + settings.HTTPClientOptions plus a
// nested unmarshal of jsonData.traceIdTimeParams. LoadConfig therefore
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

	logger.Debug("loading jaeger datasource config")

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
		logger.Error("jaeger datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("jaeger datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"traceIdTimeParams", cfg.TraceIdTimeParams.Enabled,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// Jaeger's editor writes no persisted defaults into jsonData: the "Duration"
// placeholder on `spanBar.type` (SpanBarSettings.tsx) is a render-time
// fallback and never persisted. There is no defaulting to apply here; the
// method is kept for API symmetry with the other registry entries and for
// callers that assemble a Config directly.
func (c *Config) ApplyDefaults() {
	// Intentionally empty. The Jaeger editor persists no defaults.
	_ = c
}

// Validate checks the runtime contract that the plugin requires. Errors are
// joined so callers see every problem at once.
//
// The Jaeger backend does minimal pre-validation — it just fails at request
// time when URL is empty (pkg/jaeger/jaeger.go:38-40) or when a request runs.
// We surface the essentials at load time so provisioning tooling can reject
// misconfigurations upfront:
//
//   - URL is required (pkg/jaeger/jaeger.go:38-40, pkg/jaeger/client.go:29).
//   - Basic auth requires a username (@grafana/plugin-ui BasicAuth.tsx forces
//     both fields when the method is selected).
//   - mTLS requires serverName + client cert + client key.
//   - Self-signed CA verification requires the CA PEM.
//   - `timeout` must be non-negative.
//   - `spanBar.type` must be one of the three allowed literals when set.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Jaeger URL (root.url) is required"))
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

	return errors.Join(errs...)
}
