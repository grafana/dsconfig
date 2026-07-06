// Package googlesheetsdatasource contains the configuration models for the
// Google Sheets datasource plugin (grafana-googlesheets-datasource).
package googlesheetsdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:6 in the upstream repo).
const PluginID = "grafana-googlesheets-datasource"

// AuthType is the authentication type selected in the configuration editor
// ("Authentication type"). Stored in jsonData.authenticationType.
type AuthType string

const (
	// AuthTypeAPIKey authenticates using a Google Cloud API key. Requires spreadsheets
	// to be publicly shared. src/types.ts:11 (GoogleSheetsAuth.API).
	AuthTypeAPIKey AuthType = "key"
	// AuthTypeJWT authenticates as a Google service account using a JWT credential
	// (either inline privateKey or a privateKeyPath on the Grafana server).
	// grafana-google-sdk-react/src/types.ts:4 (GoogleAuthType.JWT).
	AuthTypeJWT AuthType = "jwt"
	// AuthTypeGCE authenticates using the default service account of the GCE VM
	// Grafana is running on. grafana-google-sdk-react/src/types.ts:5 (GoogleAuthType.GCE).
	AuthTypeGCE AuthType = "gce"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAPIKey is set when the user authenticates with an API key.
	SecureJsonDataKeyAPIKey SecureJsonDataKey = "apiKey"
	// SecureJsonDataKeyPrivateKey is set when the user authenticates with a Google
	// JWT File and supplies the private key inline (as opposed to via privateKeyPath).
	SecureJsonDataKeyPrivateKey SecureJsonDataKey = "privateKey"
	// SecureJsonDataKeyJWT is the legacy blob of the full JWT service-account JSON.
	// Preserved for backward compatibility; no runtime code path depends on it
	// (pkg/models/settings.go:17,45).
	SecureJsonDataKeyJWT SecureJsonDataKey = "jwt"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAPIKey,
	SecureJsonDataKeyPrivateKey,
	SecureJsonDataKeyJWT,
}

// Config is the fully loaded configuration of a Google Sheets datasource instance.
// The plugin stores nothing plugin-specific at the root level (url, basicAuth,
// etc. are unused), so only the parsed jsonData fields and decrypted secure data
// live here. Callers reach everything directly as cfg.AuthenticationType, etc.;
// enumerate configured secrets by iterating DecryptedSecureJSONData.
//
// The jsonData fields mirror the plugin's pkg/models/settings.go DatasourceSettings
// verbatim (fields, json tags), minus the unrelated InstanceSettings back-reference
// which belongs to the SDK, not the config.
type Config struct {
	// jsonData fields, matching pkg/models/settings.go:12-26 exactly.
	AuthType           AuthType `json:"authType,omitempty"`
	AuthenticationType AuthType `json:"authenticationType,omitempty"`
	DefaultProject     string   `json:"defaultProject,omitempty"`
	ClientEmail        string   `json:"clientEmail,omitempty"`
	TokenURI           string   `json:"tokenUri,omitempty"`
	PrivateKeyPath     string   `json:"privateKeyPath,omitempty"`
	DefaultSheetID     string   `json:"defaultSheetID,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (apiKey, privateKey, jwt).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/models/settings.go:29-53), including the
// legacy fallback: instances created before authenticationType was introduced
// stored the auth type in authType, so we copy authType → authenticationType
// when authType is set (matching pkg/models/settings.go:49-51 verbatim).
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse (unmarshal jsonData +
// legacy fallback + copy decrypted secrets), then (*Config).ApplyDefaults for
// curated editor-parity defaults, then (Config).Validate to enforce the
// plugin's runtime contract. Callers that need each phase individually can
// invoke ApplyDefaults and Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading google sheets datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Legacy migration: authType -> authenticationType.
	// Mirrors pkg/models/settings.go:49-51 verbatim.
	if cfg.AuthType != "" {
		if cfg.AuthenticationType == "" {
			logger.Info("no authenticationType set but legacy authType present; migrating (legacy config)",
				"authType", cfg.AuthType,
			)
		}
		cfg.AuthenticationType = cfg.AuthType
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("google sheets datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("google sheets datasource config loaded",
		"authenticationType", cfg.AuthenticationType,
		"hasDefaultSheetID", cfg.DefaultSheetID != "",
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued discriminators with the
// same defaults the editor writes for a fresh datasource. Never blanket-apply
// every schema default — that would clobber intentional zero values.
//
// Curated defaults:
//   - AuthenticationType: 'jwt' — matches AuthConfig.tsx:40-48 which defaults
//     new datasources to GoogleAuthType.JWT.
func (c *Config) ApplyDefaults() {
	if c.AuthenticationType == "" {
		c.AuthenticationType = AuthTypeJWT
	}
}

// Validate checks the runtime contract that the plugin requires. Encodes the
// backend expectations at pkg/googlesheets/googleclient.go:
//   - authenticationType must be set (line 133-135, "missing AuthenticationType setting")
//   - authenticationType=='key' requires apiKey (line 138-142, "missing API Key")
//   - authenticationType=='jwt' requires defaultProject, clientEmail, tokenUri, and a
//     private key (line 240-245, validateDataSourceSettings; the private key may come
//     from secureJsonData.privateKey OR by resolving privateKeyPath server-side).
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	switch c.AuthenticationType {
	case AuthTypeAPIKey:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey] == "" {
			errs = append(errs, errors.New("apiKey is required for 'key' authentication"))
		}
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
		// The private key can be supplied inline as secureJsonData.privateKey OR
		// resolved server-side via privateKeyPath — exactly one must be present.
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPrivateKey] == "" && c.PrivateKeyPath == "" {
			errs = append(errs, errors.New("either secureJsonData.privateKey or jsonData.privateKeyPath is required for 'jwt' authentication"))
		}
	case AuthTypeGCE:
		// GCE has no required secret; the backend fetches credentials from the metadata server.
	case "":
		errs = append(errs, errors.New("authenticationType is required (one of: key, jwt, gce)"))
	default:
		errs = append(errs, fmt.Errorf("unknown authenticationType: %s", c.AuthenticationType))
	}

	return errors.Join(errs...)
}
