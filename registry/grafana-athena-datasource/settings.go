// Package athenadatasource contains the configuration models for the Amazon
// Athena datasource plugin (grafana-athena-datasource).
//
// The Athena plugin composes its configuration from two sources:
//
//   - The AWS auth surface shared across every Grafana AWS datasource, provided
//     by `@grafana/aws-sdk` `ConnectionConfig` on the frontend and by
//     `awsds.AWSDatasourceSettings` on the backend (github.com/grafana/grafana-aws-sdk).
//   - A small Athena-specific block: `catalog`, `database`, `workgroup`,
//     `outputLocation` (frontend spelling — camelCase). The backend struct
//     tags the same keys as PascalCase (`Catalog`, `Database`, ...) but Go's
//     case-insensitive `encoding/json` matcher happily decodes either form.
//
// The Config below flattens the upstream `awsds.AWSDatasourceSettings`-embedded
// struct into a single Go value, matching what dsconfig entries otherwise do
// (see `registry/grafana-github-datasource`). The camelCase json tags model
// what the config editor actually writes on the wire; the PascalCase quirk is
// documented in the README.
package athenadatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id field.
const PluginID = "grafana-athena-datasource"

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
	// backend (awsds/settings.go:137 and pkg/athena/models/settings.go:47)
	// still reads it from decrypted secure data.
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

// Config is the fully loaded configuration of an Athena datasource instance.
// It flattens the upstream `AthenaDataSourceSettings`
// (`pkg/athena/models/settings.go:23-32`, which embeds
// `awsds.AWSDatasourceSettings` from `pkg/awsds/settings.go:94-117`).
//
// The AWS-shared fields (authType, profile, assumeRoleArn, ...) carry the same
// tags the ConnectionConfig writes on the frontend (camelCase). The
// Athena-specific fields (catalog, database, workgroup, outputLocation) are
// spelled camelCase here to match the frontend's writes; Go's
// case-insensitive `encoding/json` also accepts the PascalCase spelling
// (`Catalog`, `Database`, `WorkGroup`, `OutputLocation`) that the upstream
// backend struct uses, so legacy provisioned configs still load. See the
// README's "Upstream findings" section for the discrepancy.
//
// Root-level datasource fields (url, basicAuth, ...) are not carried because
// the plugin does not read them.
type Config struct {
	// ---- AWS SDK ConnectionConfig fields ----

	// AuthType is the AWS credentials chain to use. Editor value; also the
	// discriminator for accessKey/secretKey/profile/assumeRoleArn.
	AuthType AWSAuthType `json:"authType,omitempty"`
	// Profile is the ~/.aws/credentials profile name for `authType == "credentials"`.
	Profile string `json:"profile,omitempty"`
	// AssumeRoleARN is the STS role ARN the selected provider should assume.
	// Backend json tag is `assumeRoleARN` (uppercase RN); we accept the
	// camelCase form the frontend actually writes.
	AssumeRoleARN string `json:"assumeRoleArn,omitempty"`
	// ExternalID is the STS external ID for cross-account assume-role.
	ExternalID string `json:"externalId,omitempty"`
	// Endpoint overrides the default AWS service endpoint.
	Endpoint string `json:"endpoint,omitempty"`
	// DefaultRegion is the AWS region to run queries in (backend also copies
	// this into `Region` when Region is empty; see awsds/settings.go:127-129).
	DefaultRegion string `json:"defaultRegion,omitempty"`

	// ---- Athena-specific ----

	// Catalog is the Athena data catalog (labelled "Data source" in the editor).
	Catalog string `json:"catalog,omitempty"`
	// Database is the Athena database within the selected catalog.
	Database string `json:"database,omitempty"`
	// Workgroup is the Athena workgroup.
	Workgroup string `json:"workgroup,omitempty"`
	// OutputLocation is the S3 URI used for query results (falls back to the
	// workgroup's configured location when empty).
	OutputLocation string `json:"outputLocation,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessKey, secretKey, sessionToken). Written by LoadConfig; never
	// marshaled.
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
// `AthenaDataSourceSettings.Load` to sync to. Callers that need each phase
// individually can invoke ApplyDefaults and Validate directly on the returned
// Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading athena datasource config")

	cfg := Config{
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
		logger.Error("athena datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("athena datasource config loaded",
		"authType", cfg.AuthType,
		"defaultRegion", cfg.DefaultRegion,
		"hasCatalog", cfg.Catalog != "",
		"hasDatabase", cfg.Database != "",
		"hasWorkgroup", cfg.Workgroup != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the config editor writes for a fresh datasource (mirroring the
// "" example in SettingsExamples).
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - AuthType → AWSAuthTypeDefault (matches the reference AWS pack default
//     and the backend `awsds.AuthTypeDefault` iota-zero).
//
// Athena-specific selectors (Catalog, Database, Workgroup, OutputLocation)
// intentionally have no default because they must be picked at runtime from
// the account being connected to.
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AWSAuthTypeDefault
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract: a known auth method is selected, its required inputs are present,
// and the Athena selectors the backend actually needs to run a query
// (defaultRegion, catalog, database, workgroup) are non-empty. Errors are
// joined so callers see every problem at once.
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
	if c.Catalog == "" {
		errs = append(errs, errors.New("catalog is required"))
	}
	if c.Database == "" {
		errs = append(errs, errors.New("database is required"))
	}
	if c.Workgroup == "" {
		errs = append(errs, errors.New("workgroup is required"))
	}

	return errors.Join(errs...)
}
