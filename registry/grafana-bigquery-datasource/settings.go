// Package bigquerydatasource contains the configuration models for the
// Google BigQuery datasource plugin (grafana-bigquery-datasource).
package bigquerydatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:3 in the upstream repo).
const PluginID = "grafana-bigquery-datasource"

// AuthType is the authentication type selected in the configuration editor
// ("Authentication type"). Stored in jsonData.authenticationType.
type AuthType string

const (
	// AuthTypeJWT authenticates as a Google service account using a JWT credential
	// (either inline privateKey or a privateKeyPath on the Grafana server).
	// grafana-google-sdk-react/src/types.ts:4 (GoogleAuthType.JWT).
	AuthTypeJWT AuthType = "jwt"
	// AuthTypeGCE authenticates using the default service account of the GCE VM
	// Grafana is running on. grafana-google-sdk-react/src/types.ts:5 (GoogleAuthType.GCE).
	AuthTypeGCE AuthType = "gce"
	// AuthTypeForwardOAuthIdentity forwards the caller's OAuth identity to BigQuery.
	// grafana-google-sdk-react/src/types.ts:7 (GoogleAuthType.ForwardOAuthIdentity).
	AuthTypeForwardOAuthIdentity AuthType = "forwardOAuthIdentity"
	// AuthTypeWorkloadIdentityFederation uses Google Cloud Workload Identity Federation.
	// Only exposed in the editor when running in Grafana Cloud (src/types.ts:47,
	// src/utils.ts:323 isCloud), but the backend accepts it anywhere.
	AuthTypeWorkloadIdentityFederation AuthType = "workloadIdentityFederation"
)

// QueryPriority is the desired default BigQuery job priority. Currently stored
// in jsonData but not consumed by the plugin at runtime (see the flatRateProject /
// queryPriority notes in the README).
type QueryPriority string

const (
	// QueryPriorityInteractive is the default INTERACTIVE priority.
	QueryPriorityInteractive QueryPriority = "INTERACTIVE"
	// QueryPriorityBatch is BATCH priority — cheaper but subject to Google's batch slot pool.
	QueryPriorityBatch QueryPriority = "BATCH"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPrivateKey is set when the user authenticates with a Google JWT
	// File and supplies the private key inline (as opposed to via privateKeyPath).
	SecureJsonDataKeyPrivateKey SecureJsonDataKey = "privateKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPrivateKey,
}

// Config is the fully loaded configuration of a Google BigQuery datasource instance.
// The plugin stores nothing plugin-specific at the root level (url, basicAuth, etc.
// are unused), so only the parsed jsonData fields and decrypted secure data live here.
// Callers reach everything directly as cfg.AuthenticationType, etc.; enumerate
// configured secrets by iterating DecryptedSecureJSONData.
//
// The jsonData fields mirror the plugin's pkg/bigquery/types/types.go BigQuerySettings
// verbatim (fields, json tags), minus the SDK back-references (DatasourceId, Updated)
// and the transient decrypted PrivateKey — those belong to the SDK / decrypted-secret
// staging, not the config.
type Config struct {
	// jsonData fields, matching pkg/bigquery/types/types.go:9-32 exactly.
	AuthenticationType           AuthType      `json:"authenticationType,omitempty"`
	DefaultProject               string        `json:"defaultProject,omitempty"`
	ClientEmail                  string        `json:"clientEmail,omitempty"`
	TokenURI                     string        `json:"tokenUri,omitempty"`
	PrivateKeyPath               string        `json:"privateKeyPath,omitempty"`
	UsingImpersonation           bool          `json:"usingImpersonation,omitempty"`
	ServiceAccountToImpersonate  string        `json:"serviceAccountToImpersonate,omitempty"`
	WorkloadIdentityPoolProvider string        `json:"workloadIdentityPoolProvider,omitempty"`
	WifServiceAccountEmail       string        `json:"wifServiceAccountEmail,omitempty"`
	OAuthPassthroughEnabled      bool          `json:"oauthPassThru,omitempty"`
	ProcessingLocation           string        `json:"processingLocation,omitempty"`
	ServiceEndpoint              string        `json:"serviceEndpoint,omitempty"`
	MaxBytesBilled               int64         `json:"MaxBytesBilled,omitempty"`
	FlatRateProject              string        `json:"flatRateProject,omitempty"`
	QueryPriority                QueryPriority `json:"queryPriority,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (privateKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's loadSettings (pkg/bigquery/settings.go:22-39), including the
// GetPrivateKey behaviour (inline privateKey or file at privateKeyPath).
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse (unmarshal jsonData + copy
// decrypted secrets), then (*Config).ApplyDefaults for curated editor-parity
// defaults, then (Config).Validate to enforce the plugin's runtime contract.
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading bigquery datasource config")

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
		logger.Error("bigquery datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("bigquery datasource config loaded",
		"authenticationType", cfg.AuthenticationType,
		"defaultProject", cfg.DefaultProject,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued discriminators with the
// same defaults the editor writes for a fresh datasource. Never blanket-apply
// every schema default — that would clobber intentional zero values.
//
// Curated defaults:
//   - AuthenticationType: 'jwt' — matches @grafana/google-sdk's AuthConfig
//     defaulting new datasources to GoogleAuthType.JWT.
//   - OAuthPassthroughEnabled: derived from AuthenticationType. AuthConfig.tsx:73-74
//     sets it to true for forwardOAuthIdentity / workloadIdentityFederation and false
//     otherwise as a side-effect of the auth-type radio, so mirror that here for
//     provisioning payloads that omit the field.
func (c *Config) ApplyDefaults() {
	if c.AuthenticationType == "" {
		c.AuthenticationType = AuthTypeJWT
	}

	switch c.AuthenticationType {
	case AuthTypeForwardOAuthIdentity, AuthTypeWorkloadIdentityFederation:
		c.OAuthPassthroughEnabled = true
	default:
		c.OAuthPassthroughEnabled = false
	}
}

// Validate checks the runtime contract that the plugin requires. Encodes the
// backend expectations at pkg/bigquery/http_client.go:
//   - authenticationType must be one of the four allowed values (line 51-89 switch;
//     unknown auth types fall through to the default 'jwt' branch and trip the
//     credential validator, which is not what a caller expects).
//   - authenticationType=='jwt' requires defaultProject, clientEmail, privateKey,
//     tokenUri (line 117-121, validateDataSourceSettings). PrivateKey may be
//     supplied inline OR resolved server-side via privateKeyPath (utils.go:62-80).
//   - authenticationType=='workloadIdentityFederation' requires
//     workloadIdentityPoolProvider (line 95).
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	switch c.AuthenticationType {
	case AuthTypeJWT:
		if c.DefaultProject == "" {
			errs = append(errs, errors.New("defaultProject is required for 'jwt' authentication"))
		}
		if c.ClientEmail == "" {
			errs = append(errs, errors.New("clientEmail is required for 'jwt' authentication"))
		}
		if c.TokenURI == "" {
			errs = append(errs, errors.New("tokenUri is required for 'jwt' authentication"))
		}
		// The private key can be supplied inline OR resolved server-side via privateKeyPath.
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPrivateKey] == "" && c.PrivateKeyPath == "" {
			errs = append(errs, errors.New("either secureJsonData.privateKey or jsonData.privateKeyPath is required for 'jwt' authentication"))
		}
	case AuthTypeGCE:
		// GCE has no strictly required field; defaultProject is convention but the
		// metadata server provides one when absent.
	case AuthTypeForwardOAuthIdentity:
		// Forward OAuth Identity requires no credentials in the datasource — the caller's
		// OAuth token is forwarded.
	case AuthTypeWorkloadIdentityFederation:
		if c.WorkloadIdentityPoolProvider == "" {
			errs = append(errs, errors.New("workloadIdentityPoolProvider is required for 'workloadIdentityFederation' authentication"))
		}
	case "":
		errs = append(errs, errors.New("authenticationType is required (one of: jwt, gce, forwardOAuthIdentity, workloadIdentityFederation)"))
	default:
		errs = append(errs, fmt.Errorf("unknown authenticationType: %s", c.AuthenticationType))
	}

	if c.MaxBytesBilled < 0 {
		errs = append(errs, fmt.Errorf("MaxBytesBilled must be non-negative, got %d", c.MaxBytesBilled))
	}

	if c.QueryPriority != "" && c.QueryPriority != QueryPriorityInteractive && c.QueryPriority != QueryPriorityBatch {
		errs = append(errs, fmt.Errorf("unknown queryPriority: %s (want INTERACTIVE or BATCH)", c.QueryPriority))
	}

	return errors.Join(errs...)
}
