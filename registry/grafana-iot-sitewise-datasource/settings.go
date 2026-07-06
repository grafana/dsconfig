// Package iotsitewisedatasource contains the configuration models for the
// AWS IoT SiteWise datasource plugin (grafana-iot-sitewise-datasource).
//
// The plugin composes its configuration from two sources:
//
//   - The AWS auth surface shared across every Grafana AWS datasource,
//     rendered by `@grafana/aws-sdk` `ConnectionConfig` on the frontend and
//     backed by `awsds.AWSDatasourceSettings` on the backend
//     (github.com/grafana/grafana-aws-sdk).
//   - An Edge Kernel block used when `defaultRegion` is the sentinel string
//     `"Edge"`: `endpoint` (gateway URL), `cert` (PEM SSL certificate), and
//     the `edgeAuthMode` + `edgeAuthUser` + `edgeAuthPass` credentials for
//     the on-prem authentication proxy.
//
// The Config below flattens the upstream `awsds.AWSDatasourceSettings`-embedded
// struct into a single Go value, matching what dsconfig entries otherwise do
// (see `registry/grafana-athena-datasource`). The json tags model what the
// config editor actually writes on the wire.
package iotsitewisedatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id field.
const PluginID = "grafana-iot-sitewise-datasource"

// AWSAuthType is the AWS authentication provider selected in the config
// editor's "Authentication Provider" Select. Stored in jsonData.authType.
//
// The upstream `awsds.AuthType` (pkg/awsds/settings.go) is a Go int enum with
// a custom MarshalJSON / UnmarshalJSON that maps to/from the string forms
// below and treats legacy values (`sharedCreds`, `arn`) as their modern
// equivalents. We keep the type as a plain string alias here because the
// dsconfig registry only cares about the on-wire representation.
type AWSAuthType string

const (
	// AWSAuthTypeDefault uses the AWS SDK default credential chain (env vars,
	// shared config, EC2/ECS metadata, ...). Editor label: "AWS SDK Default".
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

// EdgeAuthMode is the Edge Kernel authentication proxy mode selected by the
// "Authentication Mode" Select when `defaultRegion` is the sentinel string
// `"Edge"`. Stored in jsonData.edgeAuthMode. Constants mirror the plugin's
// own at pkg/models/setting.go:12-14.
type EdgeAuthMode string

const (
	// EdgeAuthModeDefault ("Standard") delegates to the AWS auth provider
	// configured above; no separate Edge username/password is required.
	EdgeAuthModeDefault EdgeAuthMode = "default"
	// EdgeAuthModeLinux authenticates against the local Linux PAM proxy
	// running on the SiteWise Edge gateway.
	EdgeAuthModeLinux EdgeAuthMode = "linux"
	// EdgeAuthModeLDAP authenticates against an LDAP server via the Edge
	// gateway's auth proxy.
	EdgeAuthModeLDAP EdgeAuthMode = "ldap"
)

// isKnown reports whether v is one of the Edge auth modes the backend
// recognizes.
func (v EdgeAuthMode) isKnown() bool {
	switch v {
	case EdgeAuthModeDefault, EdgeAuthModeLinux, EdgeAuthModeLDAP:
		return true
	}
	return false
}

// EdgeRegion is the sentinel value stored in jsonData.defaultRegion that
// switches the plugin into Edge Kernel mode. It is not a real AWS region.
// Mirrors pkg/models/setting.go:11.
const EdgeRegion = "Edge"

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
	// backend (awsds/settings.go and pkg/models/setting.go:46) still reads it
	// from decrypted secure data.
	SecureJsonDataKeySessionToken SecureJsonDataKey = "sessionToken"
	// SecureJsonDataKeyEdgeAuthPass is the password sent to the Edge Kernel
	// local authentication proxy, paired with EdgeAuthUser. Required when
	// edgeAuthMode != "default".
	SecureJsonDataKeyEdgeAuthPass SecureJsonDataKey = "edgeAuthPass"
	// SecureJsonDataKeyCert is the SSL certificate used to authenticate the
	// Edge Kernel connection (PEM-encoded, must begin with
	// "-----BEGIN CERTIFICATE-----"). Required when defaultRegion == "Edge".
	SecureJsonDataKeyCert SecureJsonDataKey = "cert"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessKey,
	SecureJsonDataKeySecretKey,
	SecureJsonDataKeySessionToken,
	SecureJsonDataKeyEdgeAuthPass,
	SecureJsonDataKeyCert,
}

// Config is the fully loaded configuration of an IoT SiteWise datasource
// instance. It flattens the upstream `AWSSiteWiseDataSourceSetting`
// (`pkg/models/setting.go:16-22`, which embeds `awsds.AWSDatasourceSettings`).
//
// The AWS-shared fields (authType, profile, assumeRoleArn, ...) carry the
// same tags the ConnectionConfig writes on the frontend (camelCase). The
// Edge-specific fields (edgeAuthMode, edgeAuthUser) also use the frontend
// spellings.
//
// Root-level datasource fields (url, basicAuth, ...) are not carried because
// the plugin does not read them.
type Config struct {
	// ---- AWS SDK ConnectionConfig fields ----

	// AuthType is the AWS credentials chain to use. Editor value; also the
	// discriminator for accessKey/secretKey/profile/assumeRoleArn.
	AuthType AWSAuthType `json:"authType,omitempty"`
	// Profile is the ~/.aws/credentials profile name for
	// `authType == "credentials"`.
	Profile string `json:"profile,omitempty"`
	// AssumeRoleARN is the STS role ARN the selected provider should assume.
	// Backend json tag is `assumeRoleARN` (uppercase RN); we accept the
	// camelCase form the frontend actually writes.
	AssumeRoleARN string `json:"assumeRoleArn,omitempty"`
	// ExternalID is the STS external ID for cross-account assume-role.
	ExternalID string `json:"externalId,omitempty"`
	// Endpoint overrides the default AWS service endpoint. Required by the
	// backend when DefaultRegion == "Edge" (pkg/models/setting.go:57-59).
	Endpoint string `json:"endpoint,omitempty"`
	// DefaultRegion is the AWS region to run queries in (backend also copies
	// this into `Region` when Region is empty; see setting.go:31-33). The
	// sentinel value "Edge" activates Edge Kernel mode.
	DefaultRegion string `json:"defaultRegion,omitempty"`

	// ---- IoT SiteWise-specific ----

	// EdgeAuthMode is the Edge Kernel authentication proxy mode. Backend
	// defaults empty to "default" when DefaultRegion == "Edge".
	EdgeAuthMode EdgeAuthMode `json:"edgeAuthMode,omitempty"`
	// EdgeAuthUser is the username sent to the Edge Kernel auth proxy.
	EdgeAuthUser string `json:"edgeAuthUser,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessKey, secretKey, sessionToken, edgeAuthPass, cert). Written by
	// LoadConfig; never marshaled.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It
// mirrors the plugin's backend Load flow: unmarshal jsonData, copy the
// decrypted secrets used by the plugin, apply the editor's parity defaults
// (curated), then Validate the runtime contract. The three phases are
// documented individually below.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig is the intended shape for the plugin's own upstream
// `AWSSiteWiseDataSourceSetting.Load` to sync to. Callers that need each
// phase individually can invoke ApplyDefaults and Validate directly on the
// returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading iot sitewise datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	// Upstream Load mirrors the same "only unmarshal when JSONData has more
	// than one byte" guard (setting.go:25).
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
		logger.Error("iot sitewise datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("iot sitewise datasource config loaded",
		"authType", cfg.AuthType,
		"defaultRegion", cfg.DefaultRegion,
		"isEdge", cfg.DefaultRegion == EdgeRegion,
		"edgeAuthMode", cfg.EdgeAuthMode,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the config editor + backend Load write for a fresh datasource
// (mirroring the "" example in SettingsExamples).
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - AuthType     → AWSAuthTypeDefault (matches the reference AWS pack
//     default and the backend `awsds.AuthTypeDefault` iota-zero).
//   - EdgeAuthMode → EdgeAuthModeDefault, but only when DefaultRegion ==
//     "Edge". This mirrors the plugin's own default in
//     pkg/models/setting.go:40-42.
//
// DefaultRegion intentionally has no default because it must be picked at
// runtime from the account being connected to (or set to "Edge" for an
// on-prem gateway).
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AWSAuthTypeDefault
	}
	if c.DefaultRegion == EdgeRegion && c.EdgeAuthMode == "" {
		c.EdgeAuthMode = EdgeAuthModeDefault
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract: a known auth method is selected, its required inputs are
// present, and — when running against an Edge Kernel gateway — the endpoint,
// SSL certificate, and (when the proxy mode is not "default") the Edge
// username + password are present. Mirrors pkg/models/setting.go:52-74 for
// the Edge branch and adds the AWS-side keys check. Errors are joined so
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

	if c.DefaultRegion == EdgeRegion {
		if !c.EdgeAuthMode.isKnown() {
			errs = append(errs, fmt.Errorf("unknown edgeAuthMode %q", c.EdgeAuthMode))
		}
		if c.Endpoint == "" {
			errs = append(errs, errors.New("edge region requires an explicit endpoint"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyCert] == "" {
			errs = append(errs, errors.New("edge region requires an SSL certificate"))
		}
		if c.EdgeAuthMode != EdgeAuthModeDefault {
			if c.EdgeAuthUser == "" {
				errs = append(errs, errors.New("missing edge auth user"))
			}
			if c.DecryptedSecureJSONData[SecureJsonDataKeyEdgeAuthPass] == "" {
				errs = append(errs, errors.New("missing edge auth password"))
			}
		}
	}

	return errors.Join(errs...)
}
