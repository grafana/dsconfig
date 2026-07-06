// Package xraydatasource contains the configuration models for the AWS
// Application Signals (X-Ray) datasource plugin (`grafana-x-ray-datasource`).
//
// The X-Ray plugin's editor is minimal: it composes only the shared
// `@grafana/aws-sdk` `ConnectionConfig` (rendered without
// `showHttpProxySettings`, `hideAssumeRoleArn`, or `skipEndpoint`) plus the
// Secure Socks Proxy toggle. There are no plugin-specific jsonData fields.
// The plugin's backend `getDsSettings` (pkg/datasource/configuration.go:8-20)
// delegates to `awsds.AWSDatasourceSettings.Load` verbatim and then folds one
// legacy behavior on top: when `jsonData.profile` is empty, the top-level
// datasource `database` field is used as the profile name.
//
// The Config below flattens the upstream `awsds.AWSDatasourceSettings` into a
// single Go value (matching what other AWS registry entries do — see
// `registry/grafana-athena-datasource`, `registry/cloudwatch`) and adds a
// separate `Database` root field to carry the legacy profile fallback.
package xraydatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id field.
// Kept as `grafana-x-ray-datasource` for backward compatibility even though
// the plugin was renamed to "AWS Application Signals" in v2.16.0.
const PluginID = "grafana-x-ray-datasource"

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
	// on `awsDatasourcesTempCredentials` in the plugin editor;
	// grafana-x-ray-datasource is in ConnectionConfig's allow-list
	// (`DS_TYPES_THAT_SUPPORT_TEMP_CREDS`).
	AWSAuthTypeGrafanaAssumeRole AWSAuthType = "grafana_assume_role"
	// AWSAuthTypeARN is a legacy stored value that the backend
	// (awsds.AuthType.UnmarshalJSON) maps to Default. Kept for round-trip
	// fidelity with datasources provisioned before the value was renamed.
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
	// data. Also used for the Grafana Assume Role flow added in plugin
	// v2.17.0.
	SecureJsonDataKeySessionToken SecureJsonDataKey = "sessionToken"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessKey,
	SecureJsonDataKeySecretKey,
	SecureJsonDataKeySessionToken,
}

// Config is the fully loaded configuration of an X-Ray / AWS Application
// Signals datasource instance. It flattens the upstream
// `awsds.AWSDatasourceSettings` (`pkg/awsds/settings.go:94-117`) — X-Ray's
// backend reads settings directly through `awsds.Load` (see
// `pkg/datasource/configuration.go:8-20`) without wrapping it in a
// plugin-specific settings type.
//
// One root-level datasource field is carried: `Database`, used as a legacy
// fallback for `Profile` when the latter is empty
// (`pkg/datasource/configuration.go:16-18`). It is tagged `json:"-"` so it
// does not collide with the jsonData decode; callers must populate it from
// `backend.DataSourceInstanceSettings.Database` before invoking Validate,
// which LoadConfig does.
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
	// Endpoint overrides the default AWS service endpoint.
	Endpoint string `json:"endpoint,omitempty"`
	// DefaultRegion is the AWS region to run queries in (backend also copies
	// this into `Region` when Region is empty; see awsds/settings.go:127-129).
	DefaultRegion string `json:"defaultRegion,omitempty"`

	// ---- Root-level datasource field ----

	// Database is the top-level datasource `database` field. X-Ray uses it as
	// a legacy fallback for Profile (`pkg/datasource/configuration.go:16-18`)
	// when Profile is empty. Tagged `json:"-"` so it does not collide with a
	// `database` key inside jsonData; LoadConfig populates it from
	// `settings.Database`.
	Database string `json:"-"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessKey, secretKey, sessionToken). Written by LoadConfig; never
	// marshaled.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// EffectiveProfile returns the profile the plugin actually uses at runtime,
// mirroring the fallback in `pkg/datasource/configuration.go:16-18`: if
// jsonData.profile is empty, the root-level `database` field is used as the
// profile name.
func (c Config) EffectiveProfile() string {
	if c.Profile != "" {
		return c.Profile
	}
	return c.Database
}

// LoadConfig parses a datasource instance's settings into a Config. It
// mirrors the plugin's backend Load flow: unmarshal jsonData, copy the
// decrypted secrets used by the plugin, capture the root-level `Database`
// legacy-profile fallback, apply the editor's parity defaults (curated),
// then Validate the runtime contract. The three phases are documented
// individually below.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig is the intended shape for the plugin's own upstream
// `getDsSettings` to sync to. Callers that need each phase individually can
// invoke ApplyDefaults and Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading x-ray datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Capture the legacy root-level `database` value that getDsSettings uses
	// as the profile fallback.
	cfg.Database = settings.Database

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("x-ray datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("x-ray datasource config loaded",
		"authType", cfg.AuthType,
		"defaultRegion", cfg.DefaultRegion,
		"hasProfile", cfg.Profile != "",
		"hasLegacyProfileFallback", cfg.Profile == "" && cfg.Database != "",
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
//
// DefaultRegion intentionally has no default because it must be picked from
// the AWS account being connected to.
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AWSAuthTypeDefault
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract: a known auth method is selected, its required inputs are present,
// and the AWS region the backend actually needs to build service clients
// (`defaultRegion`) is non-empty. Errors are joined so callers see every
// problem at once.
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

	if c.DefaultRegion == "" {
		errs = append(errs, errors.New("defaultRegion is required"))
	}

	return errors.Join(errs...)
}
