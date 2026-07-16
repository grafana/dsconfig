// Package clickhousedatasource contains the configuration models for the
// ClickHouse datasource plugin (id: grafana-clickhouse-datasource).
package clickhousedatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream repo).
const PluginID = "grafana-clickhouse-datasource"

// Protocol is the wire protocol between Grafana and ClickHouse
// (jsonData.protocol). Mirrors the frontend enum in
// src/types/config.ts:163-166.
type Protocol string

const (
	// ProtocolNative uses the ClickHouse native TCP protocol (default ports
	// 9000 / 9440 for insecure / secure).
	ProtocolNative Protocol = "native"
	// ProtocolHTTP uses the HTTP protocol (default ports 8123 / 8443).
	ProtocolHTTP Protocol = "http"
)

// ConfigMode is the datasource UI-layout switch (jsonData.configMode).
// Mirrors the frontend union in src/types/config.ts:13.
type ConfigMode string

const (
	// ConfigModeClassic exposes every database/table to the query builder.
	ConfigModeClassic ConfigMode = "classic"
	// ConfigModeSingleTable pins the datasource to one table.
	ConfigModeSingleTable ConfigMode = "single-table"
)

// SignalType is the single-table datasource's data kind (jsonData.signalType).
// Mirrors src/types/config.ts:5.
type SignalType string

const (
	// SignalTypeLogs identifies a logs table.
	SignalTypeLogs SignalType = "logs"
	// SignalTypeTraces identifies a traces table.
	SignalTypeTraces SignalType = "traces"
)

// TraceDurationUnit is the unit of the traces.durationColumn
// (jsonData.traces.durationUnit). Mirrors TimeUnit in src/types/queryBuilder.ts
// (the four values the schema/editor accept).
type TraceDurationUnit string

const (
	TraceDurationUnitNanoseconds  TraceDurationUnit = "nanoseconds"
	TraceDurationUnitMicroseconds TraceDurationUnit = "microseconds"
	TraceDurationUnitMilliseconds TraceDurationUnit = "milliseconds"
	TraceDurationUnitSeconds      TraceDurationUnit = "seconds"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
// The plugin also writes dynamic keys of the form
// "secureHttpHeaders.<Header Name>" for header values marked secure — those
// are not enumerated here (see the schema's instructions for the pattern).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the ClickHouse user password
	// (pkg/plugin/settings.go:272-275).
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// SecureJsonDataKeyTLSCACert is the CA certificate PEM used to verify the
	// ClickHouse server's TLS cert (pkg/plugin/settings.go:276-279). Consumed
	// only when jsonData.tlsAuthWithCACert is true.
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
	// SecureJsonDataKeyTLSClientCert is the client certificate PEM used for
	// mTLS (pkg/plugin/settings.go:280-283). Consumed only when
	// jsonData.tlsAuth is true.
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the client key PEM used for mTLS
	// (pkg/plugin/settings.go:284-287). Consumed only when jsonData.tlsAuth
	// is true.
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
)

// SecureJsonDataConfig lists the static secret key names stored in
// secureJsonData. Dynamic secureHttpHeaders.<Name> keys are not part of this
// list — they are populated at runtime from jsonData.httpHeaders and looked up
// with a prefix scan in the backend (pkg/plugin/settings.go:319-344).
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the static secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// HttpHeader is one entry of the Custom HTTP Headers editor. Mirrors
// CHHttpHeader in src/types/config.ts:90-94.
type HttpHeader struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Secure bool   `json:"secure"`
}

// CustomSetting is one entry of the Custom Settings editor. Mirrors
// CustomSetting in pkg/plugin/settings.go:62-65.
type CustomSetting struct {
	Setting string `json:"setting"`
	Value   string `json:"value"`
}

// AliasTableEntry is one entry of the Column Alias Tables editor. Mirrors
// AliasTableEntry in src/types/config.ts:156-161.
type AliasTableEntry struct {
	TargetDatabase string `json:"targetDatabase"`
	TargetTable    string `json:"targetTable"`
	AliasDatabase  string `json:"aliasDatabase"`
	AliasTable     string `json:"aliasTable"`
}

// LogsConfig is the nested jsonData.logs configuration. Mirrors CHLogsConfig
// in src/types/config.ts:101-116 verbatim (json tags match the storage keys).
type LogsConfig struct {
	DefaultDatabase      string   `json:"defaultDatabase,omitempty"`
	DefaultTable         string   `json:"defaultTable,omitempty"`
	OtelEnabled          bool     `json:"otelEnabled,omitempty"`
	OtelVersion          string   `json:"otelVersion,omitempty"`
	FilterTimeColumn     string   `json:"filterTimeColumn,omitempty"`
	TimeColumn           string   `json:"timeColumn,omitempty"`
	LevelColumn          string   `json:"levelColumn,omitempty"`
	MessageColumn        string   `json:"messageColumn,omitempty"`
	SelectContextColumns bool     `json:"selectContextColumns,omitempty"`
	ContextColumns       []string `json:"contextColumns,omitempty"`
	ShowLogLinks         bool     `json:"showLogLinks,omitempty"`
}

// TracesConfig is the nested jsonData.traces configuration. Mirrors
// CHTracesConfig in src/types/config.ts:118-154 verbatim.
type TracesConfig struct {
	DefaultDatabase                     string            `json:"defaultDatabase,omitempty"`
	DefaultTable                        string            `json:"defaultTable,omitempty"`
	OtelEnabled                         bool              `json:"otelEnabled,omitempty"`
	OtelVersion                         string            `json:"otelVersion,omitempty"`
	TraceIdColumn                       string            `json:"traceIdColumn,omitempty"`
	SpanIdColumn                        string            `json:"spanIdColumn,omitempty"`
	OperationNameColumn                 string            `json:"operationNameColumn,omitempty"`
	ParentSpanIdColumn                  string            `json:"parentSpanIdColumn,omitempty"`
	ServiceNameColumn                   string            `json:"serviceNameColumn,omitempty"`
	DurationColumn                      string            `json:"durationColumn,omitempty"`
	DurationUnit                        TraceDurationUnit `json:"durationUnit,omitempty"`
	StartTimeColumn                     string            `json:"startTimeColumn,omitempty"`
	TagsColumn                          string            `json:"tagsColumn,omitempty"`
	ServiceTagsColumn                   string            `json:"serviceTagsColumn,omitempty"`
	KindColumn                          string            `json:"kindColumn,omitempty"`
	StatusCodeColumn                    string            `json:"statusCodeColumn,omitempty"`
	StatusMessageColumn                 string            `json:"statusMessageColumn,omitempty"`
	StateColumn                         string            `json:"stateColumn,omitempty"`
	InstrumentationLibraryNameColumn    string            `json:"instrumentationLibraryNameColumn,omitempty"`
	InstrumentationLibraryVersionColumn string            `json:"instrumentationLibraryVersionColumn,omitempty"`
	FlattenNested                       bool              `json:"flattenNested,omitempty"`
	TraceEventsColumnPrefix             string            `json:"traceEventsColumnPrefix,omitempty"`
	TraceLinksColumnPrefix              string            `json:"traceLinksColumnPrefix,omitempty"`
	ShowTraceLinks                      bool              `json:"showTraceLinks,omitempty"`
	TraceTimestampTableSuffix           string            `json:"traceTimestampTableSuffix,omitempty"`
}

// Config is the fully loaded configuration of a ClickHouse datasource instance.
// The ClickHouse backend does not read any root-level datasource fields (URL,
// User, Database, BasicAuth*, …); every setting lives in jsonData or
// secureJsonData (pkg/plugin/settings.go). Callers reach every value directly
// as cfg.Host, cfg.Port, cfg.Protocol, cfg.Logs.DefaultTable, etc. Enumerate
// configured secrets by iterating DecryptedSecureJSONData.
//
// The jsonData fields mirror src/types/config.ts CHConfig plus the two
// backend-only cache knobs from pkg/plugin/settings.go. The
// enableSecureSocksProxy field is intentionally omitted per registry policy.
type Config struct {
	Host     string   `json:"host,omitempty"`
	Port     int64    `json:"port,omitempty"`
	Protocol Protocol `json:"protocol,omitempty"`
	Secure   bool     `json:"secure,omitempty"`
	Path     string   `json:"path,omitempty"`

	TLSSkipVerify     bool `json:"tlsSkipVerify,omitempty"`
	TLSAuth           bool `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert bool `json:"tlsAuthWithCACert,omitempty"`

	Username string `json:"username,omitempty"`

	DefaultDatabase string `json:"defaultDatabase,omitempty"`
	DefaultTable    string `json:"defaultTable,omitempty"`

	// Timeouts are stored as JSON strings. The upstream LoadSettings also
	// tolerates JSON numbers for dialTimeout / queryTimeout as a legacy shape
	// (pkg/plugin/settings.go:159-186), which our LoadConfig mirrors via
	// UnmarshalJSON below.
	DialTimeout     string `json:"dialTimeout,omitempty"`
	QueryTimeout    string `json:"queryTimeout,omitempty"`
	ConnMaxLifetime string `json:"connMaxLifetime,omitempty"`
	MaxIdleConns    string `json:"maxIdleConns,omitempty"`
	MaxOpenConns    string `json:"maxOpenConns,omitempty"`

	ValidateSql            bool `json:"validateSql,omitempty"`
	EnableMapKeysDiscovery bool `json:"enableMapKeysDiscovery,omitempty"`

	Logs   *LogsConfig   `json:"logs,omitempty"`
	Traces *TracesConfig `json:"traces,omitempty"`

	AliasTables []AliasTableEntry `json:"aliasTables,omitempty"`

	HttpHeaders           []HttpHeader `json:"httpHeaders,omitempty"`
	ForwardGrafanaHeaders bool         `json:"forwardGrafanaHeaders,omitempty"`

	CustomSettings []CustomSetting `json:"customSettings,omitempty"`

	EnableRowLimit              bool  `json:"enableRowLimit,omitempty"`
	RowLimit                    int64 `json:"rowLimit,omitempty"`
	HideTableNameInAdhocFilters bool  `json:"hideTableNameInAdhocFilters,omitempty"`

	// Frontend-only fields (stored in jsonData, never read by the backend).
	ConfigMode ConfigMode `json:"configMode,omitempty"`
	SignalType SignalType `json:"signalType,omitempty"`
	Version    string     `json:"version,omitempty"`

	// Backend-only fields (no editor UI). Both are defaulted by the backend
	// when unset (pkg/plugin/settings.go:225-252).
	EnableSchemaCache     bool `json:"enableSchemaCache,omitempty"`
	SchemaCacheTTLSeconds int  `json:"schemaCacheTTLSeconds,omitempty"`

	// SecureHttpHeaders carries the decrypted values of secure custom HTTP
	// headers, keyed by the raw header name (no `secureHttpHeaders.` prefix).
	// Populated at load time from any secureJsonData keys that begin with
	// `secureHttpHeaders.` (pkg/plugin/settings.go:319-344).
	SecureHttpHeaders map[string]string `json:"-"`

	// DecryptedSecureJSONData holds the decrypted static secret values by key
	// (password, tlsCACert, tlsClientCert, tlsClientKey). Values under the
	// dynamic secureHttpHeaders.<Name> keys are stripped into
	// SecureHttpHeaders above during LoadConfig.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// secureHTTPHeaderKeyPrefix is the prefix used in secureJsonData for secure
// custom HTTP headers, mirroring pkg/plugin/settings.go:67.
const secureHTTPHeaderKeyPrefix = "secureHttpHeaders."

// UnmarshalJSON decodes jsonData into Config while tolerating the legacy
// storage shapes upstream still accepts:
//   - `server` (v3) as an alternative name for `host`
//   - `timeout` (v3) as an alternative name for `dialTimeout`
//   - `port`, `dialTimeout`, and `queryTimeout` supplied as JSON numbers
//     instead of the current string / number form
//
// Mirrors pkg/plugin/settings.go:87-186 verbatim.
func (c *Config) UnmarshalJSON(data []byte) error {
	type alias Config
	aux := struct {
		Server       *string         `json:"server,omitempty"`
		Timeout      json.RawMessage `json:"timeout,omitempty"`
		Port         json.RawMessage `json:"port,omitempty"`
		DialTimeout  json.RawMessage `json:"dialTimeout,omitempty"`
		QueryTimeout json.RawMessage `json:"queryTimeout,omitempty"`
		*alias
	}{alias: (*alias)(c)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Legacy: `server` migrates into Host if `host` is empty.
	if c.Host == "" && aux.Server != nil {
		c.Host = strings.TrimSpace(*aux.Server)
	}

	// Port: JSON number or JSON string.
	if len(aux.Port) > 0 {
		port, err := parseInt64Raw(aux.Port)
		if err != nil {
			return fmt.Errorf("could not parse port value: %w", err)
		}
		c.Port = port
	}

	// DialTimeout: JSON string or JSON number (as int64 seconds).
	c.DialTimeout = ""
	if len(aux.DialTimeout) > 0 {
		c.DialTimeout = rawTimeoutAsString(aux.DialTimeout)
	}
	if c.DialTimeout == "" && len(aux.Timeout) > 0 {
		c.DialTimeout = rawTimeoutAsString(aux.Timeout)
	}

	// QueryTimeout: JSON string or JSON number.
	c.QueryTimeout = ""
	if len(aux.QueryTimeout) > 0 {
		c.QueryTimeout = rawTimeoutAsString(aux.QueryTimeout)
	}

	return nil
}

// parseInt64Raw accepts a json.RawMessage that is either a JSON number or a
// quoted JSON string and returns its int64 value.
func parseInt64Raw(raw json.RawMessage) (int64, error) {
	s := strings.TrimSpace(string(raw))
	if s == "null" || s == "" {
		return 0, nil
	}
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		s = strings.Trim(s, `"`)
	}
	return strconv.ParseInt(s, 10, 64)
}

// rawTimeoutAsString reduces a timeout stored either as a JSON string ("10")
// or as a JSON number (10) to its string form. Empty / null yields "".
func rawTimeoutAsString(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "null" || s == "" {
		return ""
	}
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return strings.Trim(s, `"`)
	}
	// JSON number: parse as float first (tolerates 10, 10.0), then format.
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return strconv.FormatInt(int64(f), 10)
	}
	return s
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/plugin/settings.go:80-317) parse phase — v3
// legacy fallbacks (`server`, `timeout`), string-or-number tolerance for port
// and timeouts, secure key copies, and the secureHttpHeaders.<Name> prefix
// scan — then runs (*Config).ApplyDefaults for editor-parity defaults and
// (Config).Validate to enforce the runtime contract.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading clickhouse datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
		SecureHttpHeaders:       map[string]string{},
	}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Copy static decrypted secrets by their known keys.
	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}
	// Strip dynamic secureHttpHeaders.<Name> keys into SecureHttpHeaders,
	// mirroring loadHttpHeaders in pkg/plugin/settings.go:319-344.
	for k, v := range settings.DecryptedSecureJSONData {
		if !strings.HasPrefix(k, secureHTTPHeaderKeyPrefix) {
			continue
		}
		if v == "" {
			continue
		}
		name := strings.TrimSpace(k[len(secureHTTPHeaderKeyPrefix):])
		if name == "" {
			continue
		}
		cfg.SecureHttpHeaders[name] = v
	}

	logger.Debug("loaded secure keys",
		"static", len(cfg.DecryptedSecureJSONData),
		"httpHeaders", len(cfg.SecureHttpHeaders),
	)

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("clickhouse datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("clickhouse datasource config loaded",
		"host", cfg.Host,
		"port", cfg.Port,
		"protocol", cfg.Protocol,
		"secure", cfg.Secure,
		"configMode", cfg.ConfigMode,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the config editor and backend LoadSettings write for a fresh
// datasource. Never blanket-apply every schema default — that would clobber
// intentional zero values.
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - Protocol               → ProtocolNative
//   - ConfigMode             → ConfigModeClassic
//   - EnableMapKeysDiscovery → true (matches editor default)
//   - DialTimeout            → "10" (matches pkg/plugin/settings.go:255-257)
//   - QueryTimeout           → "60"
//   - ConnMaxLifetime        → "5"
//   - MaxIdleConns           → "25"
//   - MaxOpenConns           → "50"
//   - EnableSchemaCache      → true (backend default)
//   - SchemaCacheTTLSeconds  → 60 (backend default when <=0)
//   - Logs.DefaultTable      → "otel_logs" (matches useConfigDefaults hook)
//   - Traces.DefaultTable    → "otel_traces"
//   - Traces.DurationUnit    → TraceDurationUnitNanoseconds
func (c *Config) ApplyDefaults() {
	if c.Protocol == "" {
		c.Protocol = ProtocolNative
	}
	if c.ConfigMode == "" {
		c.ConfigMode = ConfigModeClassic
	}
	if !c.EnableMapKeysDiscovery {
		// The editor default is true; mirror it here for editor parity. Users
		// who want to disable it must set false explicitly.
		c.EnableMapKeysDiscovery = true
	}
	if strings.TrimSpace(c.DialTimeout) == "" {
		c.DialTimeout = "10"
	}
	if strings.TrimSpace(c.QueryTimeout) == "" {
		c.QueryTimeout = "60"
	}
	if strings.TrimSpace(c.ConnMaxLifetime) == "" {
		c.ConnMaxLifetime = "5"
	}
	if strings.TrimSpace(c.MaxIdleConns) == "" {
		c.MaxIdleConns = "25"
	}
	if strings.TrimSpace(c.MaxOpenConns) == "" {
		c.MaxOpenConns = "50"
	}
	if !c.EnableSchemaCache {
		c.EnableSchemaCache = true
	}
	if c.SchemaCacheTTLSeconds <= 0 {
		c.SchemaCacheTTLSeconds = 60
	}
	if c.Logs == nil {
		c.Logs = &LogsConfig{}
	}
	if c.Logs.DefaultTable == "" {
		c.Logs.DefaultTable = "otel_logs"
	}
	if c.Traces == nil {
		c.Traces = &TracesConfig{}
	}
	if c.Traces.DefaultTable == "" {
		c.Traces.DefaultTable = "otel_traces"
	}
	if c.Traces.DurationUnit == "" {
		c.Traces.DurationUnit = TraceDurationUnitNanoseconds
	}
}

// Validate checks the runtime contract that the plugin requires. The
// ClickHouse backend hard-fails when Host or Port is missing
// (pkg/plugin/settings.go:69-76 → isValid()) and when Protocol is an unknown
// value. TLS toggles that require certificates are also enforced here to
// match how the driver actually behaves.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.Host == "" {
		errs = append(errs, errors.New("host (jsonData.host) is required"))
	}
	if c.Port == 0 {
		errs = append(errs, errors.New("port (jsonData.port) is required"))
	}
	switch c.Protocol {
	case ProtocolNative, ProtocolHTTP:
		// ok
	case "":
		errs = append(errs, errors.New("protocol (jsonData.protocol) is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown protocol %q (jsonData.protocol)", c.Protocol))
	}

	if c.TLSAuth {
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

	if c.ConfigMode == ConfigModeSingleTable {
		switch c.SignalType {
		case SignalTypeLogs, SignalTypeTraces:
			// ok
		case "":
			errs = append(errs, errors.New("signalType is required when configMode is single-table"))
		default:
			errs = append(errs, fmt.Errorf("unknown signalType %q", c.SignalType))
		}
	}

	return errors.Join(errs...)
}
