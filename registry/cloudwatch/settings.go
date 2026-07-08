// Package cloudwatchdatasource contains the configuration models for the
// Amazon CloudWatch datasource plugin (plugin id `cloudwatch`).
//
// The CloudWatch plugin composes its configuration from two sources:
//
//   - The AWS auth surface shared across every Grafana AWS datasource, provided
//     by `@grafana/aws-sdk` `ConnectionConfig` on the frontend and by
//     `awsds.AWSDatasourceSettings` on the backend (github.com/grafana/grafana-aws-sdk).
//     CloudWatch opts into the ConnectionConfig proxy subsection by passing
//     `showHttpProxySettings`, so the proxy fields are part of its editor
//     surface (gated additionally on the `awsPerDatasourceHTTPProxyEnabled`
//     runtime toggle).
//   - A CloudWatch-specific block: `customMetricsNamespaces` (custom metric
//     namespaces), `logsTimeout` (Cloudwatch Logs polling budget; a Go
//     `time.Duration` with a lenient string-or-number custom Unmarshal),
//     `logGroups` / `defaultLogGroups` (default log group selection, with a
//     legacy string-array shape), and `tracingDatasourceUid` (link to a
//     grafana-x-ray-datasource instance).
//
// The Config below flattens the upstream `awsds.AWSDatasourceSettings`-embedded
// struct into a single Go value, matching what dsconfig entries otherwise do
// (see `registry/grafana-github-datasource`, `registry/grafana-athena-datasource`).
// It intentionally mirrors what the plugin's upstream `CloudWatchSettings`
// reads plus the additional editor-visible frontend fields the plugin's own
// backend does not consume (logGroups / defaultLogGroups / tracingDatasourceUid).
package cloudwatchdatasource

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id field.
const PluginID = "cloudwatch"

// DefaultLogsTimeout is the default poll budget the backend uses when
// jsonData.logsTimeout is empty (pkg/cloudwatch/models/settings.go:42-44).
const DefaultLogsTimeout = 30 * time.Minute

// AWSAuthType is the AWS authentication provider selected in the config
// editor's "Authentication Provider" Select. Stored in jsonData.authType.
//
// The upstream `awsds.AuthType` (pkg/awsds/settings.go:13-91) is a Go int
// enum with a custom MarshalJSON / UnmarshalJSON that maps to/from the string
// forms below and treats legacy values (`sharedCreds`, `arn`) as their modern
// equivalents. We keep the type as a plain string alias here because the
// dsconfig registry only cares about the on-wire representation.
type AWSAuthType string

const (
	// AWSAuthTypeDefault uses the AWS SDK default credential chain (env vars,
	// shared config, EC2/ECS/EKS metadata, ...). Editor label: "AWS SDK Default".
	AWSAuthTypeDefault AWSAuthType = "default"
	// AWSAuthTypeKeys uses the accessKey/secretKey pair from secureJsonData
	// (optionally with sessionToken). Editor label: "Access & secret key".
	AWSAuthTypeKeys AWSAuthType = "keys"
	// AWSAuthTypeCredentials reads a named profile from ~/.aws/credentials.
	// Editor label: "Credentials file".
	AWSAuthTypeCredentials AWSAuthType = "credentials"
	// AWSAuthTypeEC2IAMRole uses the IAM role attached to the current EC2
	// instance / ECS task / EKS pod. Editor label: "Workspace IAM Role".
	AWSAuthTypeEC2IAMRole AWSAuthType = "ec2_iam_role"
	// AWSAuthTypeGrafanaAssumeRole delegates to Grafana Cloud's temporary
	// credentials broker. Editor label: "Grafana Assume Role". Feature-gated
	// on `awsDatasourcesTempCredentials` in the plugin editor.
	AWSAuthTypeGrafanaAssumeRole AWSAuthType = "grafana_assume_role"
	// AWSAuthTypeARN is a legacy stored value that the backend
	// (awsds.AuthType.UnmarshalJSON) maps to Default. Kept for round-trip
	// fidelity with datasources provisioned before the value was renamed.
	// The editor shows a deprecation warning banner when this value is loaded.
	AWSAuthTypeARN AWSAuthType = "arn"
)

// isKnown reports whether v is one of the AWS auth values the backend
// recognizes (including the legacy `arn`).
func (v AWSAuthType) isKnown() bool {
	switch v {
	case AWSAuthTypeDefault, AWSAuthTypeKeys, AWSAuthTypeCredentials,
		AWSAuthTypeEC2IAMRole, AWSAuthTypeGrafanaAssumeRole, AWSAuthTypeARN:
		return true
	}
	return false
}

// AWSProxyType is the HTTP proxy strategy the AWS SDK client uses. Stored in
// jsonData.proxyType.
type AWSProxyType string

const (
	// AWSProxyTypeNone disables proxy usage entirely.
	AWSProxyTypeNone AWSProxyType = "none"
	// AWSProxyTypeEnv (default) reads HTTP_PROXY / HTTPS_PROXY from the
	// process environment (ConnectionConfig.tsx:302).
	AWSProxyTypeEnv AWSProxyType = "env"
	// AWSProxyTypeURL uses the URL specified in jsonData.proxyUrl (with
	// optional username/password credentials).
	AWSProxyTypeURL AWSProxyType = "url"
)

// isKnown reports whether v is one of the recognised proxy strategies.
func (v AWSProxyType) isKnown() bool {
	switch v {
	case AWSProxyTypeNone, AWSProxyTypeEnv, AWSProxyTypeURL:
		return true
	}
	return false
}

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAccessKey is the AWS access key ID, set when
	// jsonData.authType == "keys".
	SecureJsonDataKeyAccessKey SecureJsonDataKey = "accessKey"
	// SecureJsonDataKeySecretKey is the AWS secret access key, set when
	// jsonData.authType == "keys".
	SecureJsonDataKeySecretKey SecureJsonDataKey = "secretKey"
	// SecureJsonDataKeySessionToken is an optional AWS STS session token
	// paired with keys auth. Backend-only: no editor UI writes it, but the
	// backend (awsds/settings.go:137) still reads it from decrypted secure
	// data.
	SecureJsonDataKeySessionToken SecureJsonDataKey = "sessionToken"
	// SecureJsonDataKeyProxyPassword is the optional password for a URL-mode
	// HTTP proxy. Editor-visible only when both `showHttpProxySettings` and
	// the `awsPerDatasourceHTTPProxyEnabled` runtime toggle are on; backend
	// reads it via `awsds/settings.go:138`.
	SecureJsonDataKeyProxyPassword SecureJsonDataKey = "proxyPassword"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessKey,
	SecureJsonDataKeySecretKey,
	SecureJsonDataKeySessionToken,
	SecureJsonDataKeyProxyPassword,
}

// Duration mirrors the upstream `models.Duration` type
// (pkg/cloudwatch/models/settings.go:13-15,52-77): a wrapper around
// `time.Duration` with a custom `UnmarshalJSON` that accepts both a duration
// string (e.g. "30m", "1.5s", "2000ms") and a raw nanosecond number, and
// treats the empty string as "leave zero" so ApplyDefaults can fill in the
// 30-minute default. Empty JSON input (missing / null) also leaves the value
// zero.
type Duration struct {
	time.Duration
}

// UnmarshalJSON is the CloudWatch-plugin-verbatim implementation: string
// values are passed to time.ParseDuration; float64 values are treated as
// raw nanoseconds; the empty string leaves the value zero so ApplyDefaults
// can substitute DefaultLogsTimeout. Invalid values return a downstream error.
func (d *Duration) UnmarshalJSON(b []byte) error {
	// Handle JSON null explicitly so it doesn't fall through into the
	// default-case error below.
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		return nil
	}

	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case float64:
		*d = Duration{time.Duration(v)}
	case string:
		if v == "" {
			return nil
		}
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return backend.DownstreamError(err)
		}
		*d = Duration{parsed}
	default:
		return backend.DownstreamError(fmt.Errorf("invalid duration: %#v", raw))
	}

	return nil
}

// MarshalJSON serialises the duration using Go's standard duration string
// form (e.g. "30m0s"). The upstream `models.Duration` does not define a
// MarshalJSON; we add one so callers that round-trip a Config back to JSON
// see a stable, string-form value.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// LogGroup is one entry in jsonData.logGroups. Mirrors the frontend `LogGroup`
// type (src/dataquery.ts:326-343 of grafana-cloudwatch-datasource).
type LogGroup struct {
	ARN          string `json:"arn"`
	Name         string `json:"name"`
	AccountID    string `json:"accountId,omitempty"`
	AccountLabel string `json:"accountLabel,omitempty"`
}

// Config is the fully loaded configuration of a CloudWatch datasource
// instance. It flattens the upstream `CloudWatchSettings`
// (`pkg/cloudwatch/models/settings.go:16-24`, which embeds
// `awsds.AWSDatasourceSettings` from `pkg/awsds/settings.go:94-117`) and adds
// the editor-visible frontend fields the plugin's backend does not itself
// read (`logGroups`, `defaultLogGroups`, `tracingDatasourceUid`).
//
// The AWS-shared fields carry the same tags the ConnectionConfig writes on
// the frontend (camelCase). The CloudWatch-specific fields are spelled
// exactly as the upstream backend struct spells them.
//
// Root-level datasource fields (url, basicAuth, ...) are not carried because
// the plugin's own backend does not read them. `awsds.Load` does read the
// root-level `Database` field as a legacy profile fallback (awsds/settings.go:132),
// but that fallback only takes effect when jsonData.profile is empty and
// applies to how the parsed struct is populated at runtime — it is not a
// stored config decision this schema needs to model.
type Config struct {
	// ---- AWS SDK ConnectionConfig fields ----

	// AuthType is the AWS credentials chain to use.
	AuthType AWSAuthType `json:"authType,omitempty"`
	// Profile is the ~/.aws/credentials profile name for `authType == "credentials"`.
	Profile string `json:"profile,omitempty"`
	// AssumeRoleARN is the STS role ARN the selected provider should assume.
	// Backend json tag is `assumeRoleARN` (uppercase RN); we use the
	// camelCase form the frontend actually writes. Both spellings decode via
	// Go's case-insensitive Unmarshal.
	AssumeRoleARN string `json:"assumeRoleArn,omitempty"`
	// ExternalID is the STS external ID for cross-account assume-role.
	ExternalID string `json:"externalId,omitempty"`

	// ProxyType is the HTTP proxy strategy for the AWS SDK client.
	ProxyType AWSProxyType `json:"proxyType,omitempty"`
	// ProxyURL is the proxy URL used when ProxyType is "url".
	ProxyURL string `json:"proxyUrl,omitempty"`
	// ProxyUsername is the optional proxy username used when ProxyType is "url".
	ProxyUsername string `json:"proxyUsername,omitempty"`

	// Endpoint overrides the default AWS service endpoint.
	Endpoint string `json:"endpoint,omitempty"`
	// DefaultRegion is the AWS region to run queries in (backend also copies
	// this into `Region` when Region is empty; see awsds/settings.go:127-129).
	DefaultRegion string `json:"defaultRegion,omitempty"`

	// ---- CloudWatch-specific ----

	// CustomMetricsNamespaces is a comma-separated list of namespaces the
	// query editor uses to populate its metric selectors. Backend field name
	// is `Namespace` (pkg/cloudwatch/models/settings.go:18).
	CustomMetricsNamespaces string `json:"customMetricsNamespaces,omitempty"`
	// LogsTimeout is the Cloudwatch Logs polling budget. See Duration for
	// the accepted on-wire shapes; ApplyDefaults fills the 30-minute default
	// when the value is zero.
	LogsTimeout Duration `json:"logsTimeout,omitempty"`

	// LogGroups is the list of default log groups (new object shape) used
	// as query defaults. Frontend-only: the CloudWatch backend does not read
	// jsonData.logGroups from CloudWatchSettings, but the logs query builder
	// consumes it per query.
	LogGroups []LogGroup `json:"logGroups,omitempty"`
	// DefaultLogGroups is the deprecated string-name form of LogGroups. The
	// editor migrates this to LogGroups on first open; kept for round-trip
	// fidelity with older provisioned configs.
	DefaultLogGroups []string `json:"defaultLogGroups,omitempty"`
	// TracingDatasourceUID is the grafana-x-ray-datasource UID used to
	// create trace links from log entries containing an @xrayTraceId field.
	// Frontend-only: not read by the CloudWatch backend.
	TracingDatasourceUID string `json:"tracingDatasourceUid,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessKey, secretKey, sessionToken, proxyPassword). Written by
	// LoadConfig; never marshaled.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It
// mirrors the plugin's backend Load flow: unmarshal jsonData (with the
// upstream custom Duration UnmarshalJSON handling the string-or-number
// logsTimeout), copy the decrypted secrets used by the plugin, apply the
// editor's parity defaults (curated), then Validate the runtime contract.
// The three phases are documented individually below.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig is the intended shape for the plugin's own upstream
// `LoadCloudWatchSettings` to sync to. Callers that need each phase
// individually can invoke ApplyDefaults and Validate directly on the returned
// Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading cloudwatch datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	// Mirror LoadCloudWatchSettings' check (`len > 1`) so a "{}" body is
	// tolerated exactly the way the plugin tolerates it.
	if len(settings.JSONData) > 1 {
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
		logger.Error("cloudwatch datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("cloudwatch datasource config loaded",
		"authType", cfg.AuthType,
		"defaultRegion", cfg.DefaultRegion,
		"logsTimeout", cfg.LogsTimeout.Duration,
		"hasCustomMetricsNamespaces", cfg.CustomMetricsNamespaces != "",
		"defaultLogGroupsCount", len(cfg.LogGroups)+len(cfg.DefaultLogGroups),
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the config editor / backend Load writes for a fresh datasource
// (mirroring the "" example in SettingsExamples).
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - AuthType → AWSAuthTypeDefault (matches the reference AWS pack default
//     and the backend `awsds.AuthTypeDefault` iota-zero).
//   - ProxyType → AWSProxyTypeEnv (matches ConnectionConfig.tsx:300 default
//     and the reference AWS pack default).
//   - LogsTimeout → DefaultLogsTimeout (30m), mirroring
//     pkg/cloudwatch/models/settings.go:42-44.
//
// Selectors that must be picked at runtime (DefaultRegion, LogGroups,
// CustomMetricsNamespaces, TracingDatasourceUID) intentionally have no
// default.
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AWSAuthTypeDefault
	}
	if c.ProxyType == "" {
		c.ProxyType = AWSProxyTypeEnv
	}
	if c.LogsTimeout.Duration == 0 {
		c.LogsTimeout = Duration{DefaultLogsTimeout}
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract: a known auth method and proxy type are selected, their required
// inputs are present, and the AWS region the backend actually needs to build
// service clients (`defaultRegion`) is non-empty. Errors are joined so
// callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if !c.AuthType.isKnown() {
		errs = append(errs, fmt.Errorf("unknown authType %q", c.AuthType))
	}

	switch c.AuthType {
	case AWSAuthTypeKeys:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey] == "" {
			errs = append(errs, errors.New("accessKey is required for keys auth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeySecretKey] == "" {
			errs = append(errs, errors.New("secretKey is required for keys auth"))
		}
	}

	if !c.ProxyType.isKnown() {
		errs = append(errs, fmt.Errorf("unknown proxyType %q", c.ProxyType))
	}
	if c.ProxyType == AWSProxyTypeURL && c.ProxyURL == "" {
		errs = append(errs, errors.New("proxyUrl is required when proxyType is 'url'"))
	}

	if c.DefaultRegion == "" {
		errs = append(errs, errors.New("defaultRegion is required"))
	}

	return errors.Join(errs...)
}
