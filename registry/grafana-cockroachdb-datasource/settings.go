// Package cockroachdbdatasource contains the configuration models for the
// CockroachDB datasource plugin (id: grafana-cockroachdb-datasource).
package cockroachdbdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-cockroachdb-datasource"

// Connection-pool and query-timeout defaults, mirroring the plugin's own
// LoadSettings (pkg/plugin/settings.go:18-25,254-272) verbatim.
const (
	DefaultMaxOpenConns          = 5
	DefaultMaxIdleConns          = 2
	DefaultConnMaxLifetime int32 = 300 // 5 minutes
	DefaultQueryTimeout    int32 = 30  // 30 seconds
	// QueryTimeout is clamped to [MinQueryTimeout, MaxQueryTimeout] after
	// defaulting (pkg/plugin/settings.go:268-272).
	MinQueryTimeout int32 = 5
	MaxQueryTimeout int32 = 600
)

// AuthType is the authentication method selected in the configuration editor,
// stored in jsonData.authType as the full label string
// (CockroachAuthenticationType, src/types.ts:38-42).
type AuthType string

const (
	AuthTypeSQL      AuthType = "SQL Authentication"
	AuthTypeKerberos AuthType = "Kerberos Authentication"
	AuthTypeTLS      AuthType = "TLS/SSL Authentication"
)

// TLSMode is the CockroachDB TLS negotiation mode (jsonData.sslmode),
// CockroachTLSModes (src/types.ts:7-12). Only used for TLS/SSL Authentication.
type TLSMode string

const (
	TLSModeDisable    TLSMode = "disable"
	TLSModeRequire    TLSMode = "require"
	TLSModeVerifyCA   TLSMode = "verify-ca"
	TLSModeVerifyFull TLSMode = "verify-full"
)

// TLSMethod is how TLS certificates are supplied (jsonData.tlsConfigurationMethod),
// CockroachTLSMethods (src/types.ts:14-17).
type TLSMethod string

const (
	// TLSMethodFilePath reads certs from paths on the Grafana server (sslRootCertFile / sslCertFile / sslKeyFile).
	TLSMethodFilePath TLSMethod = "file-path"
	// TLSMethodFileContent reads inline PEM content from secureJsonData (tlsCACert / tlsClientCert / tlsClientKey).
	// This is the editor default (ConfigEditor.tsx:177-181) — note it differs from PostgreSQL's 'file-path' default.
	TLSMethodFileContent TLSMethod = "file-content"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in secureJsonData
// (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the user password (read at pkg/plugin/settings.go:251).
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// SecureJsonDataKeyTLSCACert / TLSClientCert / TLSClientKey are the inline PEM
	// values used when tlsConfigurationMethod is 'file-content' (pkg/plugin/tlsmanager.go:40-47).
	SecureJsonDataKeyTLSCACert     SecureJsonDataKey = "tlsCACert"
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	SecureJsonDataKeyTLSClientKey  SecureJsonDataKey = "tlsClientKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// Config is the fully loaded configuration of a CockroachDB datasource instance.
//
// Unlike most datasources, CockroachDB stores every connection field in
// jsonData — url, user and database are read from config.JSONData
// (pkg/plugin/settings.go:28-41,247), NOT from the root-level
// backend.DataSourceInstanceSettings. Therefore no root fields are carried on
// Config. Secrets live in DecryptedSecureJSONData; enumerate configured secrets
// by iterating it.
//
// The json-tagged fields below are exactly the jsonData storage keys and are the
// single source of truth checked against dsconfig.json by the conformance suite.
type Config struct {
	URL                    string    `json:"url,omitempty"`
	Database               string    `json:"database,omitempty"`
	User                   string    `json:"user,omitempty"`
	AuthType               AuthType  `json:"authType,omitempty"`
	SSLMode                TLSMode   `json:"sslmode,omitempty"`
	TLSConfigurationMethod TLSMethod `json:"tlsConfigurationMethod,omitempty"`
	SSLRootCertFile        string    `json:"sslRootCertFile,omitempty"`
	SSLCertFile            string    `json:"sslCertFile,omitempty"`
	SSLKeyFile             string    `json:"sslKeyFile,omitempty"`
	ConfigFilePath         string    `json:"configFilePath,omitempty"`
	CredentialCache        string    `json:"credentialCache,omitempty"`
	KerberosServerName     string    `json:"kerberosServerName,omitempty"`
	MaxOpenConns           int       `json:"maxOpenConns,omitempty"`
	MaxIdleConns           int       `json:"maxIdleConns,omitempty"`
	// MaxIdleConnsAuto is frontend-only: the editor uses it to sync maxIdleConns
	// to maxOpenConns (ConnectionLimits.tsx:70-92). The backend never reads it,
	// but it is a real stored jsonData field so it is modeled here for parity.
	MaxIdleConnsAuto bool  `json:"maxIdleConnsAuto,omitempty"`
	ConnMaxLifetime  int32 `json:"connMaxLifetime,omitempty"`
	QueryTimeout     int32 `json:"queryTimeout,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (password, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// Password returns the decrypted user password from secureJsonData. The backend
// reads it at pkg/plugin/settings.go:251 (settings.Password).
func (c Config) Password() string {
	return c.DecryptedSecureJSONData[SecureJsonDataKeyPassword]
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/plugin/settings.go:245-274): unmarshal jsonData
// (url/user/database/authType/... all live in jsonData), copy the decrypted
// secrets, then apply defaults and validate. Runs parse → ApplyDefaults →
// Validate. Because this is the intended shape for upstream LoadSettings to sync
// to, any parse divergence from upstream is a bug in this entry.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading cockroachdb datasource config")

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
		logger.Error("cockroachdb datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("cockroachdb datasource config loaded",
		"hasURL", cfg.URL != "",
		"hasUser", cfg.User != "",
		"database", cfg.Database,
		"authType", cfg.AuthType,
	)
	return cfg, nil
}

// ApplyDefaults fills in zero-valued fields with the same defaults the plugin
// applies. Two groups:
//
//   - Connection-pool + query-timeout defaults, mirroring LoadSettings
//     (pkg/plugin/settings.go:254-272) verbatim. NOTE: like the backend, a stored
//     0 is replaced with the default here, so 0 does NOT survive as
//     "unlimited"/"reuse forever" despite the editor tooltips (see README).
//   - Editor-parity discriminator defaults for sslmode ('require',
//     ConfigEditor.tsx:331) and tlsConfigurationMethod ('file-content',
//     ConfigEditor.tsx:177-181). The backend does NOT apply these; the editor
//     does. Kept curated (only these zero-valued fields are touched).
func (c *Config) ApplyDefaults() {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = DefaultMaxOpenConns
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = DefaultMaxIdleConns
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = DefaultConnMaxLifetime
	}
	if c.QueryTimeout == 0 {
		c.QueryTimeout = DefaultQueryTimeout
	}
	if c.QueryTimeout < MinQueryTimeout {
		c.QueryTimeout = MinQueryTimeout
	} else if c.QueryTimeout > MaxQueryTimeout {
		c.QueryTimeout = MaxQueryTimeout
	}

	if c.SSLMode == "" {
		c.SSLMode = TLSModeRequire
	}
	if c.TLSConfigurationMethod == "" {
		c.TLSConfigurationMethod = TLSMethodFileContent
	}
}

// Validate checks the runtime contract the plugin requires. It mirrors the
// backend's isValid (pkg/plugin/settings.go:54-68) — url/user/database always
// required, password required unless authType is Kerberos — and adds the
// per-auth-method input checks the backend performs at connect time
// (generateTLSConfig / IsValidFilePathTLS / IsValidFileContentTLS,
// pkg/plugin/{settings,tlsmanager}.go). Errors are joined so callers see every
// problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("host URL (jsonData.url) is required"))
	}
	if c.User == "" {
		errs = append(errs, errors.New("username (jsonData.user) is required"))
	}
	if c.Database == "" {
		errs = append(errs, errors.New("database name (jsonData.database) is required"))
	}

	// Password is required for every auth method except Kerberos
	// (pkg/plugin/settings.go:64).
	if c.AuthType != AuthTypeKerberos && c.Password() == "" {
		errs = append(errs, errors.New("password (secureJsonData.password) is required unless authType is 'Kerberos Authentication'"))
	}

	switch c.AuthType {
	case "", AuthTypeSQL, AuthTypeKerberos, AuthTypeTLS:
		// ok (empty is tolerated: the backend treats a non-Kerberos authType as SQL auth)
	default:
		errs = append(errs, fmt.Errorf("unknown authType: %s (want %q, %q, or %q)", c.AuthType, AuthTypeSQL, AuthTypeKerberos, AuthTypeTLS))
	}

	switch c.SSLMode {
	case "", TLSModeDisable, TLSModeRequire, TLSModeVerifyCA, TLSModeVerifyFull:
		// ok
	default:
		errs = append(errs, fmt.Errorf("unknown sslmode: %s (want disable, require, verify-ca, or verify-full)", c.SSLMode))
	}

	switch c.TLSConfigurationMethod {
	case "", TLSMethodFilePath, TLSMethodFileContent:
		// ok
	default:
		errs = append(errs, fmt.Errorf("unknown tlsConfigurationMethod: %s (want file-path or file-content)", c.TLSConfigurationMethod))
	}

	// Kerberos requires a credential cache path (editor marks it required;
	// backend feeds it into the krb5 connection string).
	if c.AuthType == AuthTypeKerberos && c.CredentialCache == "" {
		errs = append(errs, errors.New("credentialCache (jsonData.credentialCache) is required for Kerberos Authentication"))
	}

	// TLS/SSL Authentication cert requirements, mirroring generateTLSConfig's
	// IsValidFilePathTLS / IsValidFileContentTLS (pkg/plugin/tlsmanager.go:25-50).
	// Only enforced when sslmode is not 'disable' to match editor visibility.
	if c.AuthType == AuthTypeTLS && c.SSLMode != TLSModeDisable {
		switch c.TLSConfigurationMethod {
		case TLSMethodFilePath:
			if c.SSLRootCertFile == "" || c.SSLCertFile == "" || c.SSLKeyFile == "" {
				errs = append(errs, errors.New("file-path TLS requires sslRootCertFile, sslCertFile, and sslKeyFile (jsonData)"))
			}
		case TLSMethodFileContent:
			if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert] == "" ||
				c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert] == "" ||
				c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey] == "" {
				errs = append(errs, errors.New("file-content TLS requires tlsCACert, tlsClientCert, and tlsClientKey (secureJsonData)"))
			}
		}
	}

	if c.MaxOpenConns < 0 {
		errs = append(errs, fmt.Errorf("maxOpenConns must be non-negative, got %d", c.MaxOpenConns))
	}
	if c.MaxIdleConns < 0 {
		errs = append(errs, fmt.Errorf("maxIdleConns must be non-negative, got %d", c.MaxIdleConns))
	}
	if c.ConnMaxLifetime < 0 {
		errs = append(errs, fmt.Errorf("connMaxLifetime must be non-negative, got %d", c.ConnMaxLifetime))
	}

	return errors.Join(errs...)
}
