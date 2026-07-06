// Package dynamodbdatasource contains the configuration models for the
// DynamoDB datasource plugin (`grafana-dynamodb-datasource`).
//
// DynamoDB's config editor is minimal: it composes only the shared
// `@grafana/aws-sdk` `ConnectionConfig` with `hideAssumeRoleArn` set (so the
// Assume Role subsection is hidden entirely) plus a `<DataSourceDescription>`
// intro block. On top of the AWS surface it stores three backend-only
// driver settings — `timeout`, `retries`, `pause` — a V2 migration marker
// (`isV2`), and two legacy V1-shape fields (`region`, `accessId`) that the
// backend's `LoadSettings` (`pkg/models/settings.go:38-85`) folds into the
// modern shape on load.
//
// The Config below flattens the upstream `models.Settings`
// (`pkg/models/settings.go:26-35`, which itself embeds
// `awsds.AWSDatasourceSettings` from `pkg/awsds/settings.go:94-117`) into a
// single Go value, matching what other AWS registry entries do (see
// `registry/grafana-athena-datasource`, `registry/grafana-x-ray-datasource`).
// This avoids pulling `grafana-aws-sdk` into the shared registry `go.mod`.
package dynamodbdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id
// field.
const PluginID = "grafana-dynamodb-datasource"

// AWSAuthType is the AWS authentication provider selected in the config
// editor's "Authentication Provider" Select. Stored in jsonData.authType.
//
// The upstream `awsds.AuthType` (pkg/awsds/settings.go:13-91) is a Go int
// enum with a custom MarshalJSON / UnmarshalJSON that maps to/from the
// string forms below and treats legacy values (`sharedCreds`, `arn`) as
// their modern equivalents. We keep the type as a plain string alias here
// because the dsconfig registry only cares about the on-wire
// representation.
//
// Note that unlike Athena/X-Ray, DynamoDB does NOT support Grafana Assume
// Role: `grafana-dynamodb-datasource` is absent from
// ConnectionConfig.tsx's `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` allow-list
// (grafana-aws-sdk-react `v0.10.2`), so the "Grafana Assume Role" option
// is filtered out of the Select at render time.
type AWSAuthType string

const (
	// AWSAuthTypeDefault uses the AWS SDK default credential chain (env
	// vars, shared config, EC2/ECS/EKS metadata, ...). Editor label:
	// "AWS SDK Default".
	AWSAuthTypeDefault AWSAuthType = "default"
	// AWSAuthTypeKeys uses the accessKey/secretKey pair from
	// secureJsonData (optionally with sessionToken). Editor label:
	// "Access & secret key". Also the value the backend force-writes
	// during V1 migration (pkg/models/settings.go:45).
	AWSAuthTypeKeys AWSAuthType = "keys"
	// AWSAuthTypeCredentials reads a named profile from
	// ~/.aws/credentials. Editor label: "Credentials file".
	AWSAuthTypeCredentials AWSAuthType = "credentials"
	// AWSAuthTypeEC2IAMRole uses the IAM role attached to the current
	// EC2 instance / ECS task / EKS pod. Editor label: "Workspace IAM
	// Role".
	AWSAuthTypeEC2IAMRole AWSAuthType = "ec2_iam_role"
	// AWSAuthTypeARN is a legacy stored value that the backend
	// (awsds.AuthType.UnmarshalJSON, awsds/settings.go:87-88) maps to
	// Default. Kept for round-trip fidelity with datasources
	// provisioned before the value was renamed.
	AWSAuthTypeARN AWSAuthType = "arn"
)

// isKnown reports whether v is one of the AWS auth values the backend
// recognizes for DynamoDB (including the legacy `arn`, excluding
// `grafana_assume_role` which the plugin does not support).
func (v AWSAuthType) isKnown() bool {
	switch v {
	case AWSAuthTypeDefault, AWSAuthTypeKeys, AWSAuthTypeCredentials,
		AWSAuthTypeEC2IAMRole, AWSAuthTypeARN:
		return true
	}
	return false
}

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAccessKey is the AWS access key ID, set when
	// jsonData.authType == "keys". Note: under V1 storage this key
	// actually held the SECRET key (the naming was fixed at V2 — see
	// pkg/models/settings.go:46 and src/utils.ts:4-22).
	SecureJsonDataKeyAccessKey SecureJsonDataKey = "accessKey"
	// SecureJsonDataKeySecretKey is the AWS secret access key, set when
	// jsonData.authType == "keys". Only used in V2 storage; V1 stored
	// the secret in secureJsonData.accessKey.
	SecureJsonDataKeySecretKey SecureJsonDataKey = "secretKey"
	// SecureJsonDataKeySessionToken is an optional AWS STS session
	// token paired with keys auth. Backend-only: no editor UI writes
	// it, but the backend (pkg/models/settings.go:50,54) still reads
	// it from decrypted secure data.
	SecureJsonDataKeySessionToken SecureJsonDataKey = "sessionToken"
)

// SecureJsonDataConfig lists the secret key names stored in
// secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessKey,
	SecureJsonDataKeySecretKey,
	SecureJsonDataKeySessionToken,
}

// Config is the fully loaded configuration of a DynamoDB datasource
// instance. It flattens the upstream `models.Settings`
// (`pkg/models/settings.go:26-35`) — which embeds
// `awsds.AWSDatasourceSettings` (`pkg/awsds/settings.go:94-117`) — into
// one struct.
//
// The AWS-shared fields (authType, profile, endpoint, defaultRegion)
// carry the same json tags the ConnectionConfig writes on the frontend
// (camelCase). The DynamoDB-specific fields (isV2, timeout, retries,
// pause) and legacy fields (region, accessId) come from
// `DynamoDBConfigOptions` (src/types.ts:19-26) and mirror the backend's
// own tags verbatim.
//
// Root-level datasource fields (url, basicAuth, ...) are not carried
// because the plugin's backend does not read them.
type Config struct {
	// ---- AWS SDK ConnectionConfig fields ----

	// AuthType is the AWS credentials chain to use. Editor value; also
	// the discriminator for accessKey/secretKey/profile.
	AuthType AWSAuthType `json:"authType,omitempty"`
	// Profile is the ~/.aws/credentials profile name for
	// `authType == "credentials"`.
	Profile string `json:"profile,omitempty"`
	// Endpoint overrides the default AWS service endpoint (visible in
	// the editor because DynamoDB does not pass `skipEndpoint`).
	Endpoint string `json:"endpoint,omitempty"`
	// DefaultRegion is the AWS region to run queries in (backend also
	// copies this into `Region` when Region is empty; see
	// awsds/settings.go:127-129).
	DefaultRegion string `json:"defaultRegion,omitempty"`

	// ---- DynamoDB-specific ----

	// IsV2 is the V2 migration marker. The editor's useEffect writes
	// this on fresh datasources (src/components/ConfigEditor.tsx:28-42);
	// the backend (pkg/models/settings.go:44) triggers V1 migration
	// when this is false.
	IsV2 bool `json:"isV2,omitempty"`
	// Timeout is the query timeout in seconds. String on the wire;
	// parsed with utils.ParseInt. Default "60" (pkg/models/settings.go:75-77).
	Timeout string `json:"timeout,omitempty"`
	// Retries is the retry count for the DynamoDB SQL driver. String
	// on the wire. Default "5" (pkg/models/settings.go:81-83).
	Retries string `json:"retries,omitempty"`
	// Pause is the pause (seconds) between retries. String on the
	// wire. Default "5" (pkg/models/settings.go:78-80).
	Pause string `json:"pause,omitempty"`

	// ---- Legacy V1 fields ----

	// LegacyRegion is the pre-V2 region storage. When IsV2 is false
	// and LegacyRegion is non-empty, the backend copies it into
	// DefaultRegion (pkg/models/settings.go:47-49).
	LegacyRegion string `json:"region,omitempty"`
	// LegacyAccessKey is the pre-V2 storage location for the AWS
	// Access Key ID (plain jsonData, misleadingly named `accessId` on
	// the wire — pkg/models/settings.go:29). When IsV2 is false the
	// backend uses this value as `AccessKey`
	// (pkg/models/settings.go:46,123-128).
	LegacyAccessKey string `json:"accessId,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessKey, secretKey, sessionToken). Written by LoadConfig;
	// never marshaled.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It
// mirrors the plugin's backend Load flow
// (`pkg/models/settings.go:38-85`): unmarshal jsonData, copy the
// decrypted secrets used by the plugin, apply V1→V2 migration when
// needed, apply the editor's parity defaults, then Validate the runtime
// contract. The three phases (parse → ApplyDefaults → Validate) are
// documented individually below.
//
// ctx is used to derive a contextual logger via
// backend.Logger.FromContext so log lines carry the request/plugin
// context that Grafana injects.
//
// LoadConfig is the intended shape for the plugin's own upstream
// `LoadSettings` to sync to. Callers that need each phase individually
// can invoke ApplyDefaults and Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading dynamodb datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Copy every decrypted secure key straight through; V1 migration
	// (below) may reinterpret some of them.
	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	// V1 → V2 migration mirrors pkg/models/settings.go:44-55.
	// When jsonData.isV2 is false or missing, the backend forces
	// authType to keys, treats jsonData.accessId as the AccessKey ID
	// and secureJsonData.accessKey as the SECRET key, and copies
	// jsonData.region into DefaultRegion when the latter is empty. We
	// don't mutate the stored jsonData (LoadConfig is a read-only load
	// step); the resulting Config just reflects the effective V2 shape
	// with legacy fields preserved for round-trip.
	if !cfg.IsV2 {
		if cfg.AuthType == "" {
			cfg.AuthType = AWSAuthTypeKeys
		}
		if cfg.LegacyRegion != "" && cfg.DefaultRegion == "" {
			cfg.DefaultRegion = cfg.LegacyRegion
		}
		// Under V1 storage, secureJsonData.accessKey holds the SECRET
		// key and jsonData.accessId (LegacyAccessKey) holds the ID.
		// Rebind them into the modern-shape secure map so Validate
		// and downstream consumers see V2 semantics.
		if legacySecret, ok := cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey]; ok {
			if _, hasSecret := cfg.DecryptedSecureJSONData[SecureJsonDataKeySecretKey]; !hasSecret {
				cfg.DecryptedSecureJSONData[SecureJsonDataKeySecretKey] = legacySecret
			}
			if cfg.LegacyAccessKey != "" {
				cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey] = cfg.LegacyAccessKey
			}
		} else if cfg.LegacyAccessKey != "" {
			cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey] = cfg.LegacyAccessKey
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("dynamodb datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("dynamodb datasource config loaded",
		"authType", cfg.AuthType,
		"defaultRegion", cfg.DefaultRegion,
		"isV2", cfg.IsV2,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the
// same defaults the editor and backend Load write for a fresh
// datasource.
//
// Curated list (only these fields are touched, and only when
// zero-valued):
//   - AuthType → AWSAuthTypeDefault (matches the reference AWS pack
//     default and the backend `awsds.AuthTypeDefault` iota-zero; note
//     that the DynamoDB editor's own useEffect
//     `src/components/ConfigEditor.tsx:28-42` writes `keys` on a fresh
//     datasource instead — we prefer the AWS-pack default for schema
//     consistency across registry entries).
//   - Timeout / Retries / Pause → "60" / "5" / "5" (mirrors
//     pkg/models/settings.go:75-83). These are only relevant when the
//     plugin's SQL driver assembles DriverSettings; setting them here
//     ensures a Config assembled outside LoadConfig still exposes the
//     same defaults the plugin uses at runtime.
//
// DefaultRegion intentionally has no default because it must be
// picked from the AWS account being connected to.
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AWSAuthTypeDefault
	}
	if c.Timeout == "" {
		c.Timeout = "60"
	}
	if c.Retries == "" {
		c.Retries = "5"
	}
	if c.Pause == "" {
		c.Pause = "5"
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract (mirrors pkg/models/settings.go:57-68): a known auth
// method is selected, its required inputs are present, and
// DefaultRegion is non-empty. Errors are joined so callers see every
// problem at once.
func (c Config) Validate() error {
	var errs []error

	if !c.AuthType.isKnown() {
		errs = append(errs, fmt.Errorf("unknown authType %q", c.AuthType))
	}

	switch c.AuthType {
	case AWSAuthTypeKeys:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey] == "" {
			errs = append(errs, errors.New("missing access key"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeySecretKey] == "" {
			errs = append(errs, errors.New("missing secret key"))
		}
	}

	if c.DefaultRegion == "" {
		errs = append(errs, errors.New("missing region"))
	}

	return errors.Join(errs...)
}
