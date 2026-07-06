// Package prometheusdatasource contains the configuration models for the
// Prometheus datasource plugin (id: prometheus).
package prometheusdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "prometheus"

// HTTPMethod is the HTTP verb the Prometheus client uses for /api/v1/query and
// /api/v1/query_range requests. Stored in jsonData.httpMethod. Mirrors the
// values validated at pkg/promlib/models/settings.go:92-95.
type HTTPMethod string

const (
	// HTTPMethodPOST is the default (getOptionsWithDefaults, PromSettings.tsx:74-82;
	// ApplyDefaults, pkg/promlib/models/settings.go:82-87).
	HTTPMethodPOST HTTPMethod = http.MethodPost
	// HTTPMethodGET is used for Prometheus < 2.1 or restricted networks.
	HTTPMethodGET HTTPMethod = http.MethodGet
)

// PromApplication mirrors the frontend PromApplication enum
// (packages/grafana-prometheus/src/types.ts:28-33).
type PromApplication string

const (
	PromApplicationPrometheus PromApplication = "Prometheus"
	PromApplicationCortex     PromApplication = "Cortex"
	PromApplicationMimir      PromApplication = "Mimir"
	PromApplicationThanos     PromApplication = "Thanos"
)

// PrometheusCacheLevel mirrors the frontend PrometheusCacheLevel enum
// (packages/grafana-prometheus/src/types.ts:21-26).
type PrometheusCacheLevel string

const (
	PrometheusCacheLevelLow    PrometheusCacheLevel = "Low"
	PrometheusCacheLevelMedium PrometheusCacheLevel = "Medium"
	PrometheusCacheLevelHigh   PrometheusCacheLevel = "High"
	PrometheusCacheLevelNone   PrometheusCacheLevel = "None"
)

// QueryEditorMode mirrors the frontend QueryEditorMode enum used for the
// "Default editor" select (packages/grafana-prometheus/src/querybuilder/shared/types.ts).
type QueryEditorMode string

const (
	QueryEditorModeBuilder QueryEditorMode = "builder"
	QueryEditorModeCode    QueryEditorMode = "code"
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

// Config is the fully loaded configuration of a Prometheus datasource
// instance. It mirrors the plugin's upstream PromOptions
// (pkg/promlib/models/settings.go:27-53) verbatim — same json tags — plus
// DecryptedSecureJSONData for the write-only secrets and a URL root-field
// that the backend reads directly from backend.DataSourceInstanceSettings.
//
// The base DataSourceJsonData fields (authType, defaultRegion, profile,
// alertmanagerUid, disableGrafanaCache) exist on the upstream PromOptions via
// an embedded struct but are not written by the Prometheus editor and are not
// consumed by the Prometheus plugin's own code. They are intentionally omitted
// from Config so the schema/struct parity check stays tight.
type Config struct {
	// Root-level fields (json:"-" on the struct because they don't live in jsonData).
	// URL is required and is read directly by the Prometheus backend
	// (pkg/promlib/admission_handler.go:51, pkg/promlib/querydata/request.go:61).
	// BasicAuth / BasicAuthUser / WithCredentials are populated by the editor and
	// consumed by the SDK's settings.HTTPClientOptions() inside CreateTransportOptions
	// (pkg/promlib/client/transport.go:18) — the Prometheus code itself never touches
	// them by name.
	URL             string `json:"-"`
	BasicAuth       bool   `json:"-"`
	BasicAuthUser   string `json:"-"`
	WithCredentials bool   `json:"-"`

	// jsonData fields — subset that the editor writes and/or the backend reads.
	// Custom HTTP header pairs (jsonData.httpHeaderName<N> / secureJsonData.httpHeaderValue<N>)
	// are not modeled here because they are dynamically indexed.
	TLSAuth                             bool                         `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert                   bool                         `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify                       bool                         `json:"tlsSkipVerify,omitempty"`
	ServerName                          string                       `json:"serverName,omitempty"`
	Timeout                             float64                      `json:"timeout,omitempty"`
	KeepCookies                         []string                     `json:"keepCookies,omitempty"`
	OauthPassThru                       bool                         `json:"oauthPassThru,omitempty"`
	ManageAlerts                        bool                         `json:"manageAlerts,omitempty"`
	AllowAsRecordingRulesTarget         bool                         `json:"allowAsRecordingRulesTarget,omitempty"`
	TimeInterval                        string                       `json:"timeInterval,omitempty"`
	QueryTimeout                        string                       `json:"queryTimeout,omitempty"`
	DefaultEditor                       QueryEditorMode              `json:"defaultEditor,omitempty"`
	DisableMetricsLookup                bool                         `json:"disableMetricsLookup,omitempty"`
	PrometheusType                      PromApplication              `json:"prometheusType,omitempty"`
	PrometheusVersion                   string                       `json:"prometheusVersion,omitempty"`
	CacheLevel                          PrometheusCacheLevel         `json:"cacheLevel,omitempty"`
	IncrementalQuerying                 bool                         `json:"incrementalQuerying,omitempty"`
	IncrementalQueryOverlapWindow       string                       `json:"incrementalQueryOverlapWindow,omitempty"`
	DisableRecordingRules               bool                         `json:"disableRecordingRules,omitempty"`
	CustomQueryParameters               string                       `json:"customQueryParameters,omitempty"`
	HTTPMethod                          HTTPMethod                   `json:"httpMethod,omitempty"`
	SeriesLimit                         int64                        `json:"seriesLimit,omitempty"`
	SeriesEndpoint                      bool                         `json:"seriesEndpoint,omitempty"`
	ExemplarTraceIdDestinations         []ExemplarTraceIdDestination `json:"exemplarTraceIdDestinations,omitempty"`
	MaxSamplesProcessedWarningThreshold float64                      `json:"maxSamplesProcessedWarningThreshold,omitempty"`
	MaxSamplesProcessedErrorThreshold   float64                      `json:"maxSamplesProcessedErrorThreshold,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ExemplarTraceIdDestination mirrors the frontend `ExemplarTraceIdDestination`
// type (packages/grafana-prometheus/src/types.ts:57-62) and the backend
// `ExemplarTraceIDDestination` (pkg/promlib/models/settings.go:56-61).
type ExemplarTraceIdDestination struct {
	Name            string `json:"name"`
	URL             string `json:"url,omitempty"`
	URLDisplayLabel string `json:"urlDisplayLabel,omitempty"`
	DatasourceUID   string `json:"datasourceUid,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root fields
// (URL, BasicAuth, BasicAuthUser, WithCredentials) are copied from
// backend.DataSourceInstanceSettings directly; jsonData is unmarshaled verbatim
// from settings.JSONData. Mirrors ParsePromOptions
// (pkg/promlib/models/settings.go:65-79) with the added root-field capture that
// the Prometheus editor and SDK actually rely on.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults -> Validate.
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading prometheus datasource config")

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
		logger.Error("prometheus datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("prometheus datasource config loaded",
		"hasURL", cfg.URL != "",
		"basicAuth", cfg.BasicAuth,
		"tlsAuth", cfg.TLSAuth,
		"httpMethod", cfg.HTTPMethod,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// Curated defaults:
//   - HTTPMethod: POST — mirrors both getOptionsWithDefaults (PromSettings.tsx:74-82)
//     and the backend's ApplyDefaults (pkg/promlib/models/settings.go:82-87).
//     The HTTP method is uppercased first, matching the backend.
//
// Other apparent "defaults" (cacheLevel Low, defaultEditor Builder, seriesLimit
// 40000, incrementalQueryOverlapWindow 10m) live only in the editor's visual
// fallbacks (`Select value={... ?? Low}`, placeholder text) and are NOT
// written to storage on load, so they are not applied here.
func (c *Config) ApplyDefaults() {
	c.HTTPMethod = HTTPMethod(strings.ToUpper(strings.TrimSpace(string(c.HTTPMethod))))
	if c.HTTPMethod == "" {
		c.HTTPMethod = HTTPMethodPOST
	}
}

// Validate checks the runtime contract that the plugin requires. The
// Prometheus backend hard-fails without a URL (pkg/promlib/admission_handler.go:51)
// and rejects any HTTPMethod other than POST/GET/empty
// (pkg/promlib/models/settings.go:92-95). It also encodes the TLS field pairs
// required by @grafana/plugin-ui and the SDK. Errors are joined so callers see
// every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("Prometheus server URL (root.url) is required"))
	}

	switch m := HTTPMethod(strings.ToUpper(string(c.HTTPMethod))); m {
	case "", HTTPMethodPOST, HTTPMethodGET:
		// OK.
	default:
		errs = append(errs, fmt.Errorf("invalid httpMethod %q: must be GET or POST", c.HTTPMethod))
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
	if c.SeriesLimit < 0 {
		errs = append(errs, fmt.Errorf("seriesLimit must be non-negative, got %d", c.SeriesLimit))
	}

	return errors.Join(errs...)
}
