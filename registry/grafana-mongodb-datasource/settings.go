// Package mongodbdatasource contains the configuration models for the MongoDB
// datasource plugin (id: grafana-mongodb-datasource).
package mongodbdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream plugin).
const PluginID = "grafana-mongodb-datasource"

// AuthType is the authentication method selected in the configuration editor,
// stored in jsonData.authType. Mirrors the values the backend accepts at
// pkg/models/settings.go:66 and pkg/datasource/client.go:139-161.
type AuthType string

const (
	// AuthTypeNoAuth connects without authentication.
	AuthTypeNoAuth AuthType = "NoAuth"
	// AuthTypeBasicAuth authenticates with a username/password (editor label "Credentials").
	AuthTypeBasicAuth AuthType = "BasicAuth"
	// AuthTypeKerberos authenticates with Kerberos (requires authMechanism=GSSAPI in the connection string).
	AuthTypeKerberos AuthType = "custom-Kerberos"
)

// IsValid reports whether the auth type is one of the values the backend
// accepts (pkg/models/settings.go:66). Empty and unknown values are invalid;
// the backend (and ApplyDefaults) coerce them to BasicAuth.
func (a AuthType) IsValid() bool {
	switch a {
	case AuthTypeNoAuth, AuthTypeBasicAuth, AuthTypeKerberos:
		return true
	default:
		return false
	}
}

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the Credentials (basic auth) password.
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyKerberosPassword is the Kerberos client principal password.
	SecureJsonDataKeyKerberosPassword SecureJsonDataKey = "kerberosPassword"
	// SecureJsonDataKeyTLSCertificateKeyFilePassword decrypts an encrypted TLS client key.
	SecureJsonDataKeyTLSCertificateKeyFilePassword SecureJsonDataKey = "tlsCertificateKeyFilePassword"
	// SecureJsonDataKeyTLSCACert is the CA certificate PEM (used when tlsAuthWithCACert is enabled).
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
	// SecureJsonDataKeyTLSClientCert is the client certificate PEM (used when tlsAuth is enabled).
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the client private key PEM (used when tlsAuth is enabled).
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
	// SecureJsonDataKeyPassword is the legacy (pre-v1.9.0) basic auth password.
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads (pkg/models/settings.go
// LoadSettings + mapstructure.Decode of DecryptedSecureJSONData).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyKerberosPassword,
	SecureJsonDataKeyTLSCertificateKeyFilePassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
	SecureJsonDataKeyPassword,
}

// Config is the fully loaded configuration of a MongoDB datasource instance.
//
// The json-tagged fields mirror the json-tagged fields of the plugin's backend
// Settings struct (pkg/models/settings.go:15-45) — the jsonData shape. The
// MongoDB backend also reads two ROOT-level datasource fields
// (config.BasicAuthEnabled, config.BasicAuthUser) which are carried here tagged
// json:"-". Decrypted secrets live in DecryptedSecureJSONData.
//
// The upstream Settings struct also declares json tags for tlsCACert,
// tlsClientCert, tlsClientKey and tlsCertificateKeyFilePassword, but the config
// editor and provisioning store those in secureJsonData and
// mapstructure.Decode(DecryptedSecureJSONData, &settings) (settings.go:154)
// overwrites them from secureJsonData — so they are modeled here as secrets
// (DecryptedSecureJSONData), not jsonData fields.
type Config struct {
	// jsonData fields (json tags == dsconfig.json jsonData field keys).
	Connection           string   `json:"connection,omitempty"`
	AuthType             AuthType `json:"authType,omitempty"`
	KerberosUser         string   `json:"kerberosUser,omitempty"`
	KeyTabFilePath       string   `json:"keyTabFilePath,omitempty"`
	GlobalCcacheFilePath string   `json:"globalCcacheFilePath,omitempty"`
	CcacheLookupFile     string   `json:"ccacheLookupFile,omitempty"`
	ServerName           string   `json:"serverName,omitempty"`
	TLSAuth              bool     `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert    bool     `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify        bool     `json:"tlsSkipVerify,omitempty"`
	ResponseRowsLimit    string   `json:"responseRowsLimit,omitempty"`
	ValidateSyntax       bool     `json:"validate,omitempty"`

	// Legacy jsonData fields.
	User                   string `json:"user,omitempty"`
	InsecureSkipValidation bool   `json:"skipTLSValidation,omitempty"`
	Credentials            bool   `json:"credentials,omitempty"`

	// Root-level datasource fields the backend reads (json:"-" so they never
	// collide with jsonData). Resolved by LoadConfig (legacy user + secrets can
	// force BasicAuthEnabled on and populate BasicAuthUser).
	BasicAuthEnabled bool   `json:"-"`
	BasicAuthUser    string `json:"-"`

	// DecryptedSecureJSONData holds the decrypted secure values by key.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// KerberosEnabled reports whether Kerberos authentication is actually active,
// mirroring pkg/models/settings.go:138: the connection string must contain
// authMechanism=GSSAPI and a Kerberos username must be set.
func (c Config) KerberosEnabled() bool {
	return strings.Contains(c.Connection, "authMechanism=GSSAPI") && c.KerberosUser != ""
}

// BasicAuthPassword returns the effective basic-auth password, preferring the
// modern secureJsonData.basicAuthPassword and falling back to the legacy
// secureJsonData.password (pkg/models/settings.go:105,122-136).
func (c Config) BasicAuthPassword() string {
	if pw := c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword]; pw != "" {
		return pw
	}
	return c.DecryptedSecureJSONData[SecureJsonDataKeyPassword]
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/models/settings.go:50-162): it copies the two
// root basic-auth fields off backend.DataSourceInstanceSettings, unmarshals
// jsonData, copies decrypted secrets, applies the legacy fallbacks (jsonData.user
// -> basicAuthUser, secure password enables basic auth, skipTLSValidation ->
// tlsSkipVerify), then runs ApplyDefaults and Validate.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults -> Validate.
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading mongodb datasource config")

	cfg := Config{
		BasicAuthEnabled:        settings.BasicAuthEnabled,
		BasicAuthUser:           settings.BasicAuthUser,
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

	// Legacy: jsonData.user is the pre-v1.9.0 username (settings.go:102-107).
	if cfg.User != "" {
		cfg.BasicAuthUser = cfg.User
		cfg.BasicAuthEnabled = true
	}
	// The modern root basicAuthUser overrides the legacy value (settings.go:111-121).
	if settings.BasicAuthUser != "" {
		cfg.BasicAuthUser = settings.BasicAuthUser
		cfg.BasicAuthEnabled = true
	}
	// A configured basic-auth password (modern or legacy) enables basic auth
	// (settings.go:105,122-136).
	if cfg.BasicAuthPassword() != "" {
		cfg.BasicAuthEnabled = true
	}
	// Legacy skipTLSValidation is copied into tlsSkipVerify (settings.go:143-148).
	if cfg.InsecureSkipValidation {
		cfg.TLSSkipVerify = true
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("mongodb datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("mongodb datasource config loaded",
		"authType", cfg.AuthType,
		"hasConnection", cfg.Connection != "",
		"basicAuthEnabled", cfg.BasicAuthEnabled,
		"kerberosEnabled", cfg.KerberosEnabled(),
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of fields with the same defaults the backend
// applies for a fresh datasource. Never blanket-apply every schema default —
// that would clobber intentional zero values.
//
// Curated defaults (mirroring pkg/models/settings.go):
//   - AuthType: empty or unrecognized -> BasicAuth (settings.go:66-71)
//   - ResponseRowsLimit: empty -> "10000" (settings.go:56-58)
func (c *Config) ApplyDefaults() {
	if !c.AuthType.IsValid() {
		c.AuthType = AuthTypeBasicAuth
	}
	if c.ResponseRowsLimit == "" {
		c.ResponseRowsLimit = "10000"
	}
}

// Validate checks the runtime contract the plugin requires. The MongoDB backend
// hard-fails when the connection string is empty (pkg/datasource/client.go:134-137)
// and when TLS is enabled without the corresponding certificate material
// (client.go:391-399 for the CA, client.go:407-515 for the client cert/key).
// It only logs warnings for empty basic-auth or Kerberos credentials
// (settings.go:76-83, client.go:144-157), so those are not enforced here.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.Connection == "" {
		errs = append(errs, errors.New("connection string (jsonData.connection) is required"))
	}

	switch c.AuthType {
	case AuthTypeNoAuth, AuthTypeBasicAuth, AuthTypeKerberos:
		// valid
	case "":
		errs = append(errs, errors.New("authType is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown authType %q", c.AuthType))
	}

	if c.TLSAuth {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert] == "" {
			errs = append(errs, errors.New("tlsClientCert (secureJsonData) is required when tlsAuth is true"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey] == "" {
			errs = append(errs, errors.New("tlsClientKey (secureJsonData) is required when tlsAuth is true"))
		}
	}
	if c.TLSAuthWithCACert {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert] == "" {
			errs = append(errs, errors.New("tlsCACert (secureJsonData) is required when tlsAuthWithCACert is true"))
		}
	}

	return errors.Join(errs...)
}
