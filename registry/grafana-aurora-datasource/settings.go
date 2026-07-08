// Package auroradatasource contains the configuration models for the Amazon
// Aurora datasource plugin (grafana-aurora-datasource).
//
// The Aurora plugin composes its configuration from two sources:
//
//   - The AWS auth surface shared across every Grafana AWS datasource, provided
//     by `@grafana/aws-sdk` `ConnectionConfig` on the frontend and by
//     `awsds.AWSDatasourceSettings` on the backend (github.com/grafana/grafana-aws-sdk).
//     Aurora does NOT pass `showHttpProxySettings`, so the ConnectionConfig
//     proxy subsection is not part of its editor surface.
//   - An Aurora-specific block: an `engine` discriminator (Aurora Postgres or
//     Aurora MySQL — MySQL is a beta path per upstream README), the required
//     `dbUser` / `dbHost` / `dbPort` (used both for query traffic and — by
//     default — for the RDS IAM auth-token endpoint), plus optional
//     `dbHostAuth` / `dbPortAuth` overrides for setups where the auth token
//     must be signed with a different host/port than the SQL traffic.
//
// The Config below flattens the upstream `awsds.AWSDatasourceSettings`-embedded
// struct into a single Go value, matching what other dsconfig entries do
// (see `registry/grafana-athena-datasource`, `registry/grafana-redshift-datasource`).
// Aurora authenticates to the database with an RDS IAM auth token generated at
// connect time (pkg/plugin/connect.go:34-72), so there is no password secret.
package auroradatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id field.
const PluginID = "grafana-aurora-datasource"

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

// AuroraEngine is the Aurora database engine selected in the config editor's
// "Engine" Select. Stored in jsonData.engine and mirrors the plugin's own
// `SupportedEngine` type (`pkg/plugin/consts.go:5-11`).
type AuroraEngine string

const (
	// AuroraEnginePostgres is Aurora (PostgreSQL Compatible). Default engine
	// (ConfigEditor.tsx:47) and the fallback used when the stored engine is
	// empty or unrecognized (pkg/plugin/connect.go:83-85, 135-138).
	AuroraEnginePostgres AuroraEngine = "aurora-postgres"
	// AuroraEngineMySQL is Aurora (MySQL Compatible). Beta path per the
	// upstream README.
	AuroraEngineMySQL AuroraEngine = "aurora-mysql"
)

// isKnown reports whether v is one of the Aurora engine values the backend
// treats as first-class. Any other value falls back to AuroraEnginePostgres
// at connect time.
func (v AuroraEngine) isKnown() bool {
	switch v {
	case AuroraEnginePostgres, AuroraEngineMySQL:
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
	// plugin's parseSettings (pkg/plugin/driver.go:112) still reads it from
	// decrypted secure data.
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

// Config is the fully loaded configuration of an Aurora datasource instance.
// It flattens the upstream `AuroraConfigSettings`
// (`pkg/plugin/driver.go:90-101`, which embeds
// `awsds.AWSDatasourceSettings` from `pkg/awsds/settings.go:94-117`).
//
// The AWS-shared fields carry the same tags the ConnectionConfig writes on
// the frontend (camelCase). The Aurora-specific fields carry the exact tags
// the upstream backend struct declares (all camelCase — `engine`, `dbUser`,
// `dbName`, `dbHost`, `dbPort`, `dbHostAuth`, `dbPortAuth`) so a Config
// round-trips identically through the plugin's own parseSettings.
//
// Root-level datasource fields (url, basicAuth, ...) are not carried because
// the plugin does not read them — the RDS endpoint lives in jsonData.dbHost /
// jsonData.dbPort.
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
	// DefaultRegion is the AWS region used both for the AWS SDK calls that
	// resolve credentials and for the RDS `generate-db-auth-token` call
	// (pkg/plugin/connect.go:40, 68).
	DefaultRegion string `json:"defaultRegion,omitempty"`

	// ---- Aurora-specific ----

	// Engine picks the Aurora engine. Empty / unknown values fall back to
	// aurora-postgres at connect time (pkg/plugin/connect.go:83-85, 135-138).
	Engine AuroraEngine `json:"engine,omitempty"`
	// DBUser is the database principal the RDS IAM auth token impersonates.
	// Required by the editor and by the backend to build the DSN.
	DBUser string `json:"dbUser,omitempty"`
	// DBName is the database name used to build the SQL driver DSN. Not
	// used for the RDS auth token itself.
	DBName string `json:"dbName,omitempty"`
	// DBHost is the Aurora cluster endpoint (host portion only). Required.
	DBHost string `json:"dbHost,omitempty"`
	// DBPort is the Aurora cluster port. Required.
	DBPort int `json:"dbPort,omitempty"`
	// DBHostAuth is an optional host override used only for the RDS
	// `generate-db-auth-token` call (pkg/plugin/connect.go:59-62). Falls
	// back to DBHost when empty.
	DBHostAuth string `json:"dbHostAuth,omitempty"`
	// DBPortAuth is an optional port override used only for the RDS
	// `generate-db-auth-token` call (pkg/plugin/connect.go:63-66). Falls
	// back to DBPort when zero.
	DBPortAuth int `json:"dbPortAuth,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessKey, secretKey, sessionToken). Written by LoadConfig; never
	// marshaled.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It
// mirrors the plugin's own parseSettings flow (`pkg/plugin/driver.go:103-115`):
// unmarshal jsonData, copy the decrypted secrets used by the plugin, apply
// the editor's parity defaults (curated), then Validate the runtime
// contract. The three phases are documented individually below.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig is the intended shape for the plugin's own upstream
// `parseSettings` to sync to. Callers that need each phase individually can
// invoke ApplyDefaults and Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading aurora datasource config")

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
		logger.Error("aurora datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("aurora datasource config loaded",
		"authType", cfg.AuthType,
		"engine", cfg.Engine,
		"defaultRegion", cfg.DefaultRegion,
		"hasDBHost", cfg.DBHost != "",
		"hasDBUser", cfg.DBUser != "",
		"hasSplitAuthEndpoint", cfg.DBHostAuth != "" || cfg.DBPortAuth != 0,
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
//   - Engine   → AuroraEnginePostgres (matches ConfigEditor.tsx:47 default
//     and the backend fallback at pkg/plugin/connect.go:83-85).
//
// Aurora selectors (DBUser, DBHost, DBPort, DBName) intentionally have no
// default because they must be picked at runtime from the specific cluster.
// DBHostAuth / DBPortAuth intentionally have no default because their empty
// values are what tell the backend to fall back to DBHost / DBPort.
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AWSAuthTypeDefault
	}
	if c.Engine == "" {
		c.Engine = AuroraEnginePostgres
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract: a known AWS auth method is selected (with its required inputs),
// a known Aurora engine is selected (bad values would silently coerce to
// postgres at connect time — we surface the mismatch here instead), the
// AWS region is set (used both for credential resolution and for the RDS
// auth-token signature), and the three editor-required fields (dbUser,
// dbHost, dbPort) are populated. Errors are joined so callers see every
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

	if !c.Engine.isKnown() {
		errs = append(errs, fmt.Errorf("unknown engine %q (want aurora-postgres or aurora-mysql)", c.Engine))
	}

	if c.DefaultRegion == "" {
		errs = append(errs, errors.New("defaultRegion is required"))
	}
	if c.DBUser == "" {
		errs = append(errs, errors.New("dbUser is required"))
	}
	if c.DBHost == "" {
		errs = append(errs, errors.New("dbHost is required"))
	}
	if c.DBPort <= 0 {
		errs = append(errs, errors.New("dbPort is required"))
	}
	if c.DBPortAuth < 0 {
		errs = append(errs, fmt.Errorf("dbPortAuth must be non-negative, got %d", c.DBPortAuth))
	}

	return errors.Join(errs...)
}
