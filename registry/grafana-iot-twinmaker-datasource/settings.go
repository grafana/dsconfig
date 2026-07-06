// Package iottwinmakerdatasource contains the configuration models for the
// AWS IoT TwinMaker datasource, which ships as a nested datasource inside
// the AWS IoT TwinMaker App plugin (`grafana-iot-twinmaker-app`). The
// datasource itself has plugin id `grafana-iot-twinmaker-datasource`.
//
// The datasource composes its configuration from:
//
//   - The AWS auth surface shared across every Grafana AWS datasource,
//     rendered by `@grafana/aws-sdk@0.8.3` `ConnectionConfig` on the frontend
//     and backed by `awsds.AWSDatasourceSettings` on the backend
//     (github.com/grafana/grafana-aws-sdk).
//   - Two TwinMaker-specific fields — `workspaceId` and `assumeRoleArnWriter`
//     — required to talk to the IoT TwinMaker API and, respectively, to
//     grant the Alarm Configuration Panel write access.
//
// The Config below flattens the upstream `awsds.AWSDatasourceSettings`
// embedded struct into a single Go value, matching the pattern used by
// sibling AWS entries (see `registry/grafana-iot-sitewise-datasource`,
// `registry/grafana-x-ray-datasource`, `registry/grafana-athena-datasource`).
package iottwinmakerdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching
// `src/datasource/plugin.json`'s id field. Note that the containing app
// plugin has a different id (`grafana-iot-twinmaker-app`); the app's
// includes[] list references this datasource id.
const PluginID = "grafana-iot-twinmaker-datasource"

// AWSAuthType is the AWS authentication provider selected in the config
// editor's "Authentication Provider" Select. Stored in jsonData.authType.
//
// The upstream `awsds.AuthType` (pkg/awsds/settings.go) is a Go int enum
// with a custom MarshalJSON / UnmarshalJSON that maps to/from the string
// forms below and treats legacy values (`sharedCreds`, `arn`) as their
// modern equivalents. We keep the type as a plain string alias here because
// the dsconfig registry only cares about the on-wire representation.
type AWSAuthType string

const (
	// AWSAuthTypeDefault uses the AWS SDK default credential chain (env
	// vars, shared config, EC2/ECS/EKS metadata, ...). Editor label:
	// "AWS SDK Default".
	AWSAuthTypeDefault AWSAuthType = "default"
	// AWSAuthTypeKeys uses the accessKey/secretKey pair from secureJsonData.
	// Editor label: "Access & secret key".
	AWSAuthTypeKeys AWSAuthType = "keys"
	// AWSAuthTypeCredentials reads a named profile from ~/.aws/credentials.
	// Editor label: "Credentials file".
	AWSAuthTypeCredentials AWSAuthType = "credentials"
	// AWSAuthTypeEC2IAMRole uses the IAM role attached to the current EC2
	// instance / ECS task / EKS pod. Editor label: "Workspace IAM Role".
	AWSAuthTypeEC2IAMRole AWSAuthType = "ec2_iam_role"
	// AWSAuthTypeGrafanaAssumeRole delegates to Grafana Cloud's temporary
	// credentials broker. Storage-valid but NOT editor-selectable for
	// TwinMaker: `@grafana/aws-sdk@0.8.3`'s ConnectionConfig restricts this
	// provider to a fixed allow-list (`DS_TYPES_THAT_SUPPORT_TEMP_CREDS`)
	// of cloudwatch/athena/amazonprometheus.
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
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin's backend.
// Notably, `sessionToken` — which is part of the shared AWS secure shape —
// is NOT read by TwinMaker's Load (pkg/models/settings.go:37-38 copies only
// accessKey/secretKey). See the README's discrepancies section.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessKey,
	SecureJsonDataKeySecretKey,
}

// Config is the fully loaded configuration of an IoT TwinMaker datasource
// instance. It flattens the upstream `TwinMakerDataSourceSetting`
// (`pkg/models/settings.go:12-18`, which embeds `awsds.AWSDatasourceSettings`).
//
// The AWS-shared fields (authType, profile, assumeRoleArn, ...) carry the
// same tags the ConnectionConfig writes on the frontend (camelCase). The
// TwinMaker-specific fields (workspaceId, assumeRoleArnWriter) use the
// frontend spellings from `src/datasource/types.ts:24-27`.
//
// Root-level datasource fields (url, basicAuth, ...) are not carried
// because the plugin does not read them.
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
	// Endpoint overrides the default AWS service endpoint.
	Endpoint string `json:"endpoint,omitempty"`
	// DefaultRegion is the AWS region to run queries in (backend also
	// copies this into `Region` when Region is empty; see settings.go:30-32
	// and finally falls back to "us-east-1" at settings.go:33-35).
	DefaultRegion string `json:"defaultRegion,omitempty"`

	// ---- IoT TwinMaker-specific ----

	// WorkspaceID identifies the TwinMaker workspace this datasource
	// queries. Required at runtime by CheckHealth
	// (pkg/plugin/datasource.go:172-177).
	WorkspaceID string `json:"workspaceId,omitempty"`
	// AssumeRoleARNWriter is the optional STS role ARN used when the Alarm
	// Configuration Panel writes property values (pkg/models/settings.go:63-66).
	AssumeRoleARNWriter string `json:"assumeRoleArnWriter,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessKey, secretKey). Written by LoadConfig; never marshaled.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It
// mirrors the plugin's backend Load flow: unmarshal jsonData, copy the
// decrypted secrets the plugin actually reads, apply the editor's parity
// defaults (curated), then Validate the runtime contract. The three
// phases are documented individually below.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext
// so log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig is the intended shape for the plugin's own upstream
// `TwinMakerDataSourceSetting.Load` to sync to. Callers that need each
// phase individually can invoke ApplyDefaults and Validate directly on the
// returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading iot twinmaker datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	// Upstream Load mirrors the same "only unmarshal when JSONData has more
	// than one byte" guard (pkg/models/settings.go:21).
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
		logger.Error("iot twinmaker datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("iot twinmaker datasource config loaded",
		"authType", cfg.AuthType,
		"defaultRegion", cfg.DefaultRegion,
		"workspaceId", cfg.WorkspaceID,
		"hasWriteRole", cfg.AssumeRoleARNWriter != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the config editor + backend Load write for a fresh datasource
// (mirroring the "" example in SettingsExamples).
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - AuthType      → AWSAuthTypeDefault (matches the reference AWS pack
//     default and the backend `awsds.AuthTypeDefault` iota-zero).
//   - DefaultRegion → "us-east-1" (mirrors the editor's on-mount write at
//     ConfigEditor.tsx:31-36 and the backend fallback at settings.go:33-35).
//
// Other fields (WorkspaceID, AssumeRoleARN, AssumeRoleARNWriter) have no
// defaults because they must be picked by the user for a working
// datasource.
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AWSAuthTypeDefault
	}
	if c.DefaultRegion == "" {
		c.DefaultRegion = "us-east-1"
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract: a known auth method is selected, its required inputs are
// present, the TwinMaker workspace id is set, and an assume-role ARN is
// set (required by CheckHealth at pkg/plugin/datasource.go:179-184 even
// though the ConnectionConfig editor labels the field "Optional"). Errors
// are joined so callers see every problem at once.
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

	if c.WorkspaceID == "" {
		errs = append(errs, errors.New("workspaceId is required"))
	}
	if c.AssumeRoleARN == "" {
		errs = append(errs, errors.New("assumeRoleArn is required"))
	}

	return errors.Join(errs...)
}
