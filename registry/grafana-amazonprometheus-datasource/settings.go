// Package amazonprometheusdatasource contains the configuration models for
// the Amazon Managed Service for Prometheus datasource plugin
// (id: grafana-amazonprometheus-datasource).
package amazonprometheusdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's `id` field
// (src/plugin.json:6 in the upstream repo).
const PluginID = "grafana-amazonprometheus-datasource"

// SigV4AuthType is the discriminator stored in jsonData.sigV4AuthType.
// Mirrors the values in `@grafana/aws-sdk` (`AwsAuthType`, src/types.ts) and
// `grafana-aws-sdk` (`pkg/awsds/settings.go`). `arn` is deprecated and
// preserved for round-trip fidelity with datasources provisioned before it
// was renamed to `default`.
type SigV4AuthType string

const (
	SigV4AuthTypeDefault           SigV4AuthType = "default"
	SigV4AuthTypeKeys              SigV4AuthType = "keys"
	SigV4AuthTypeCredentials       SigV4AuthType = "credentials"
	SigV4AuthTypeEC2IAMRole        SigV4AuthType = "ec2_iam_role"
	SigV4AuthTypeGrafanaAssumeRole SigV4AuthType = "grafana_assume_role"
	// SigV4AuthTypeARN is a legacy value that the backend maps to Default;
	// kept for round-trip fidelity with pre-rename provisioned configs.
	SigV4AuthTypeARN SigV4AuthType = "arn"
)

// isKnown reports whether v is one of the SigV4 auth values the backend
// recognizes (including the legacy `arn`).
func (v SigV4AuthType) isKnown() bool {
	switch v {
	case SigV4AuthTypeDefault, SigV4AuthTypeKeys, SigV4AuthTypeCredentials,
		SigV4AuthTypeEC2IAMRole, SigV4AuthTypeGrafanaAssumeRole, SigV4AuthTypeARN:
		return true
	}
	return false
}

// HTTPMethod is the HTTP verb the Prometheus client uses. Stored in
// jsonData.httpMethod. Mirrors the values validated by the promlib backend
// (grafana-prometheus-datasource/pkg/promlib).
type HTTPMethod string

const (
	HTTPMethodPOST HTTPMethod = http.MethodPost
	HTTPMethodGET  HTTPMethod = http.MethodGet
)

// PromApplication mirrors the frontend PromApplication enum
// (@grafana/prometheus types.ts).
type PromApplication string

const (
	PromApplicationPrometheus PromApplication = "Prometheus"
	PromApplicationCortex     PromApplication = "Cortex"
	PromApplicationMimir      PromApplication = "Mimir"
	PromApplicationThanos     PromApplication = "Thanos"
)

// PrometheusCacheLevel mirrors the frontend PrometheusCacheLevel enum.
type PrometheusCacheLevel string

const (
	PrometheusCacheLevelLow    PrometheusCacheLevel = "Low"
	PrometheusCacheLevelMedium PrometheusCacheLevel = "Medium"
	PrometheusCacheLevelHigh   PrometheusCacheLevel = "High"
	PrometheusCacheLevelNone   PrometheusCacheLevel = "None"
)

// QueryEditorMode mirrors the frontend QueryEditorMode enum.
type QueryEditorMode string

const (
	QueryEditorModeBuilder QueryEditorMode = "builder"
	QueryEditorModeCode    QueryEditorMode = "code"
)

// DefaultSigV4Service is the AWS service namespace the plugin defaults
// `jsonData.sigv4Service` to when the field is empty or missing. Mirrors
// the fallback in `pkg/datasource.go:126` (`extendClientOpts`).
const DefaultSigV4Service = "aps"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeySigV4AccessKey is the AWS access key ID, set when
	// jsonData.sigV4AuthType == "keys".
	SecureJsonDataKeySigV4AccessKey SecureJsonDataKey = "sigV4AccessKey"
	// SecureJsonDataKeySigV4SecretKey is the AWS secret access key, set
	// when jsonData.sigV4AuthType == "keys".
	SecureJsonDataKeySigV4SecretKey SecureJsonDataKey = "sigV4SecretKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeySigV4AccessKey,
	SecureJsonDataKeySigV4SecretKey,
}

// ExemplarTraceIdDestination mirrors the frontend
// `ExemplarTraceIdDestination` type and the backend `promlib`
// `ExemplarTraceIDDestination`.
type ExemplarTraceIdDestination struct {
	Name            string `json:"name"`
	URL             string `json:"url,omitempty"`
	URLDisplayLabel string `json:"urlDisplayLabel,omitempty"`
	DatasourceUID   string `json:"datasourceUid,omitempty"`
}

// Config is the fully loaded configuration of an Amazon Prometheus
// datasource instance. It flattens:
//
//   - The Prometheus knobs from `@grafana/prometheus`'s `PromOptions`
//     (`packages/grafana-prometheus/src/types.ts` @ 13.1.6) which the
//     `promlib` backend parses at query time.
//   - The Amazon-specific jsonData fields from `DataSourceOptions`
//     (`src/configuration/DataSourceOptions.ts`): `sigV4Auth`,
//     `sigv4Service` (lowercase 'v'), `forwardGrafanaUserHeader`, and the
//     `prometheus-type-migration` sentinel.
//   - The SigV4 credential fields written by `@grafana/aws-sdk`'s
//     `SIGV4ConnectionConfig` (`SIGV4ConnectionConfig.tsx:20-27`):
//     `sigV4AuthType`, `sigV4Profile`, `sigV4AssumeRoleArn`,
//     `sigV4ExternalId`, `sigV4Region`, plus the `sigV4AccessKey` /
//     `sigV4SecretKey` secrets in `DecryptedSecureJSONData`.
//   - The root-level `URL` field the SDK feeds into `promlib`.
//
// Root basic-auth / withCredentials fields are intentionally omitted — the
// plugin's editor sets `visibleMethods=[sigV4Id]` and clears
// `basicAuth`/`withCredentials`/`oauthPassThru` on every save
// (`src/configuration/DataSourceHttpSettingsOverhaul.tsx:103-115`), and the
// plugin's own Go code (`pkg/datasource.go`) never reads them by name.
type Config struct {
	// Root-level. `json:"-"` because it lives on the SDK instance, not
	// jsonData. The Prometheus backend reads it via `promlib` (which
	// resolves the base URL from `settings.URL`) at query time.
	URL string `json:"-"`

	// --- SigV4 auth (mandatory) ---

	// SigV4Auth is the enabling flag. Forced to true on every editor mount.
	SigV4Auth bool `json:"sigV4Auth,omitempty"`
	// SigV4AuthType is the AWS credentials chain to use for signing.
	SigV4AuthType SigV4AuthType `json:"sigV4AuthType,omitempty"`
	// SigV4Profile is the ~/.aws/credentials profile name for
	// `sigV4AuthType == "credentials"`.
	SigV4Profile string `json:"sigV4Profile,omitempty"`
	// SigV4AssumeRoleArn is the optional STS role ARN the selected provider
	// should assume.
	SigV4AssumeRoleArn string `json:"sigV4AssumeRoleArn,omitempty"`
	// SigV4ExternalId is the optional STS external ID for cross-account
	// assume-role.
	SigV4ExternalId string `json:"sigV4ExternalId,omitempty"`
	// SigV4Region is the default AWS region to sign against.
	SigV4Region string `json:"sigV4Region,omitempty"`
	// SigV4Service is the AWS service namespace to sign requests against.
	// NOTE: lowercase 'v' in the json tag (`sigv4Service`) — this differs
	// from every other sigV4-prefixed field. Defaults to "aps" when empty.
	SigV4Service string `json:"sigv4Service,omitempty"`

	// --- Amazon-specific jsonData fields ---

	// ForwardGrafanaUserHeader forwards the logged-in Grafana user's
	// `X-Grafana-User` header to the workspace. Consumed by
	// `pkg/datasource.go:95-99`.
	ForwardGrafanaUserHeader bool `json:"forwardGrafanaUserHeader,omitempty"`

	// PrometheusTypeMigration is `true` for datasources migrated from
	// vanilla Prometheus. Frontend-only sentinel used by
	// `ConfigEditor.tsx:37-48` to render a migration banner. Storage key
	// contains a hyphen.
	PrometheusTypeMigration bool `json:"prometheus-type-migration,omitempty"`

	// --- Prometheus jsonData (mirrors @grafana/prometheus PromOptions verbatim) ---

	OAuthPassThru                       bool                         `json:"oauthPassThru,omitempty"`
	ManageAlerts                        bool                         `json:"manageAlerts,omitempty"`
	AllowAsRecordingRulesTarget         bool                         `json:"allowAsRecordingRulesTarget,omitempty"`
	Timeout                             float64                      `json:"timeout,omitempty"`
	KeepCookies                         []string                     `json:"keepCookies,omitempty"`
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
	// (sigV4AccessKey, sigV4SecretKey). Written by LoadConfig; never
	// marshaled.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig runs the full parse -> ApplyDefaults -> Validate flow and
// returns a fully-defaulted, validated Config. Parse mirrors the plugin's
// own settings-load path — the plugin has no custom `LoadSettings`; instead
// it delegates jsonData parsing to `promlib/utils.GetJsonData`
// (`pkg/datasource.go:30`) at HTTP-client construction time and reads a
// handful of individual fields (`forwardGrafanaUserHeader`, `sigv4Service`)
// via `maputil`. This entry consolidates that spread into a single typed
// struct.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading amazon prometheus datasource config")

	cfg := Config{
		URL:                     settings.URL,
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
		logger.Error("amazon prometheus datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("amazon prometheus datasource config loaded",
		"hasURL", cfg.URL != "",
		"sigV4AuthType", cfg.SigV4AuthType,
		"sigV4Region", cfg.SigV4Region,
		"httpMethod", cfg.HTTPMethod,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor and backend write:
//
//   - SigV4Auth → true (editor forces this on every mount,
//     `DataSourceHttpSettingsOverhaul.tsx:27-38`).
//   - SigV4Service → "aps" when empty (backend fallback,
//     `pkg/datasource.go:124-129`).
//   - HTTPMethod → uppercased then defaulted to POST (mirrors the promlib
//     backend `PromOptions.ApplyDefaults`).
//
// Other editor "visual defaults" (cacheLevel=Low, defaultEditor=builder,
// seriesLimit=40000, incrementalQueryOverlapWindow=10m) are never written
// to storage on load and are therefore not applied here.
func (c *Config) ApplyDefaults() {
	c.SigV4Auth = true
	if strings.TrimSpace(c.SigV4Service) == "" {
		c.SigV4Service = DefaultSigV4Service
	}
	c.HTTPMethod = HTTPMethod(strings.ToUpper(strings.TrimSpace(string(c.HTTPMethod))))
	if c.HTTPMethod == "" {
		c.HTTPMethod = HTTPMethodPOST
	}
}

// Validate checks the runtime contract that the plugin requires. Errors are
// joined so callers see every problem at once.
//
//   - URL is required (Prometheus backend hard-fails on empty
//     `settings.URL`).
//   - `HTTPMethod` must be empty/POST/GET (promlib validates).
//   - `SigV4AuthType` must be one of the recognised AWS auth types (or the
//     legacy `arn`), and `keys` requires both `sigV4AccessKey` and
//     `sigV4SecretKey` in `DecryptedSecureJSONData`.
//   - `SigV4Region` is required — the SigV4 signer must know which region
//     to sign against, and the workspace URL alone does not carry that
//     information reliably.
//   - Numeric fields (`Timeout`, `SeriesLimit`) must be non-negative.
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

	if !c.SigV4AuthType.isKnown() && c.SigV4AuthType != "" {
		errs = append(errs, fmt.Errorf("unknown sigV4AuthType %q", c.SigV4AuthType))
	}

	if c.SigV4AuthType == SigV4AuthTypeKeys {
		if c.DecryptedSecureJSONData[SecureJsonDataKeySigV4AccessKey] == "" {
			errs = append(errs, errors.New("secureJsonData.sigV4AccessKey is required when sigV4AuthType is 'keys'"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeySigV4SecretKey] == "" {
			errs = append(errs, errors.New("secureJsonData.sigV4SecretKey is required when sigV4AuthType is 'keys'"))
		}
	}

	if c.SigV4Region == "" {
		errs = append(errs, errors.New("jsonData.sigV4Region is required"))
	}

	if c.Timeout < 0 {
		errs = append(errs, fmt.Errorf("timeout must be non-negative, got %v", c.Timeout))
	}
	if c.SeriesLimit < 0 {
		errs = append(errs, fmt.Errorf("seriesLimit must be non-negative, got %d", c.SeriesLimit))
	}

	return errors.Join(errs...)
}
