// Package stackdriver contains the configuration models for the Google Cloud
// Monitoring datasource plugin (plugin id: stackdriver).
package stackdriver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo). The plugin is presented in the UI
// as "Google Cloud Monitoring"; the id remains "stackdriver" for backward
// compatibility with datasources provisioned under the product's original name.
const PluginID = "stackdriver"

// AuthType is the authentication type selected in the configuration editor
// ("Authentication type"). Stored in jsonData.authenticationType.
//
// Values mirror the backend constants at
// pkg/cloudmonitoring/cloudmonitoring.go:52-55 verbatim.
type AuthType string

const (
	// AuthTypeJWT authenticates as a Google service account using a JWT credential
	// (either inline privateKey or a privateKeyPath on the Grafana server).
	// grafana-google-sdk-react/src/types.ts:4 (GoogleAuthType.JWT).
	AuthTypeJWT AuthType = "jwt"
	// AuthTypeGCE authenticates using the default service account of the GCE VM
	// Grafana is running on. grafana-google-sdk-react/src/types.ts:5 (GoogleAuthType.GCE).
	AuthTypeGCE AuthType = "gce"
	// AuthTypeWorkloadIdentityFederation uses Google Cloud Workload Identity Federation.
	// Only exposed in the editor when running in Grafana Cloud
	// (src/utils.ts:15 isCloud + src/components/ConfigEditor/ConfigEditor.tsx:32-38),
	// but the backend accepts it anywhere.
	AuthTypeWorkloadIdentityFederation AuthType = "workloadIdentityFederation"
	// AuthTypeForwardOAuthIdentity forwards the caller's OAuth identity to Cloud Monitoring.
	// Alerting queries are not supported with this method
	// (pkg/cloudmonitoring/cloudmonitoring.go:412-418).
	AuthTypeForwardOAuthIdentity AuthType = "forwardOAuthIdentity"
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

// Config is the fully loaded configuration of a Google Cloud Monitoring
// datasource instance. The plugin stores nothing plugin-specific at the root
// level — the backend derives every Google URL from routes[] + universeDomain
// (pkg/cloudmonitoring/httpclient.go:79-83) and never reads settings.URL,
// settings.BasicAuth, etc. — so only the parsed jsonData fields and decrypted
// secure data live here.
//
// The jsonData fields mirror the plugin's `datasourceJSONData` struct at
// pkg/cloudmonitoring/cloudmonitoring.go:203-215 verbatim (fields, json tags),
// with two additions: the frontend-only `gceDefaultProject` cache
// (src/datasource.ts:186-191) and the excluded-per-schema `enableSecureSocksProxy`
// flag are both accepted here so provisioning payloads deserialize cleanly, but
// the backend datasourceJSONData ignores them.
type Config struct {
	// jsonData fields — the first 9 mirror pkg/cloudmonitoring/cloudmonitoring.go:203-215.
	AuthenticationType           AuthType `json:"authenticationType,omitempty"`
	DefaultProject               string   `json:"defaultProject,omitempty"`
	ClientEmail                  string   `json:"clientEmail,omitempty"`
	TokenURI                     string   `json:"tokenUri,omitempty"`
	PrivateKeyPath               string   `json:"privateKeyPath,omitempty"`
	UsingImpersonation           bool     `json:"usingImpersonation,omitempty"`
	ServiceAccountToImpersonate  string   `json:"serviceAccountToImpersonate,omitempty"`
	WorkloadIdentityPoolProvider string   `json:"workloadIdentityPoolProvider,omitempty"`
	WifServiceAccountEmail       string   `json:"wifServiceAccountEmail,omitempty"`
	OAuthPassthroughEnabled      bool     `json:"oauthPassThru,omitempty"`
	UniverseDomain               string   `json:"universeDomain,omitempty"`

	// Frontend-managed runtime cache; the backend never reads this key
	// (see the field docs on `Config` and src/datasource.ts:186-191). Kept in the
	// struct so provisioning payloads that inadvertently carry it don't fail to parse.
	GCEDefaultProject string `json:"gceDefaultProject,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (privateKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's newDatasourceInfo (pkg/cloudmonitoring/cloudmonitoring.go:222-276)
// verbatim for parsing (default authenticationType to 'jwt' when empty,
// json-unmarshal jsonData, copy decrypted secrets into DecryptedSecureJSONData).
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

	logger.Debug("loading google cloud monitoring datasource config")

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
		logger.Error("google cloud monitoring datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("google cloud monitoring datasource config loaded",
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
//   - AuthenticationType: 'jwt' — matches both the frontend
//     (@grafana/google-sdk AuthConfig.tsx:40-48 useEffect defaulting new
//     datasources to GoogleAuthType.JWT) and the backend
//     (pkg/cloudmonitoring/cloudmonitoring.go:229-231 which stamps
//     jwtAuthentication when jsonData.AuthenticationType is empty).
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
// backend expectations at pkg/cloudmonitoring/:
//   - authenticationType must be one of the four allowed values (the switch at
//     httpclient.go:54-74 silently falls through for unknown types, leaving the
//     HTTP client with no token provider — which is not what a caller expects).
//   - authenticationType=='jwt' requires defaultProject, clientEmail, tokenUri
//     (the backend token provider will fail at request time otherwise). PrivateKey
//     may be supplied inline OR resolved server-side via privateKeyPath
//     (grafana-google-sdk-go/pkg/utils/utils.go:62-89).
//   - authenticationType=='workloadIdentityFederation' requires
//     workloadIdentityPoolProvider (httpclient.go:87-89).
//   - authenticationType=='workloadIdentityFederation' or 'forwardOAuthIdentity'
//     requires defaultProject (cloudmonitoring.go:121-125: CheckHealth refuses to
//     run without a project — token-forwarding auth types can't infer one from
//     credentials).
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
		// GCE has no strictly required field; the metadata server provides credentials and
		// (when defaultProject is empty) the default project id.
	case AuthTypeWorkloadIdentityFederation:
		if c.WorkloadIdentityPoolProvider == "" {
			errs = append(errs, errors.New("workloadIdentityPoolProvider is required for 'workloadIdentityFederation' authentication"))
		}
		if c.DefaultProject == "" {
			errs = append(errs, errors.New("defaultProject is required for 'workloadIdentityFederation' authentication"))
		}
	case AuthTypeForwardOAuthIdentity:
		if c.DefaultProject == "" {
			errs = append(errs, errors.New("defaultProject is required for 'forwardOAuthIdentity' authentication"))
		}
	case "":
		errs = append(errs, errors.New("authenticationType is required (one of: jwt, gce, workloadIdentityFederation, forwardOAuthIdentity)"))
	default:
		errs = append(errs, fmt.Errorf("unknown authenticationType: %s", c.AuthenticationType))
	}

	return errors.Join(errs...)
}
