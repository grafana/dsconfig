// Package mssqldatasource contains the configuration models for the
// Microsoft SQL Server datasource plugin (id: mssql).
package mssqldatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "mssql"

// AuthType is the authentication type selected in the configuration editor.
// Stored in jsonData.authenticationType. Mirrors src/types.ts:16-24.
type AuthType string

const (
	AuthTypeSQL                     AuthType = "SQL Server Authentication"
	AuthTypeWindows                 AuthType = "Windows Authentication"
	AuthTypeAzureAD                 AuthType = "Azure AD Authentication"
	AuthTypeKerberosRaw             AuthType = "Windows AD: Username + password"
	AuthTypeKerberosKeytab          AuthType = "Windows AD: Keytab"
	AuthTypeKerberosCache           AuthType = "Windows AD: Credential cache"
	AuthTypeKerberosCacheLookupFile AuthType = "Windows AD: Credential cache file"
)

// EncryptOption is the SSL/TLS negotiation mode stored in jsonData.encrypt.
// Mirrors src/types.ts:26-30.
type EncryptOption string

const (
	EncryptDisable EncryptOption = "disable"
	EncryptFalse   EncryptOption = "false"
	EncryptTrue    EncryptOption = "true"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in secureJsonData.
type SecureJsonDataKey string

const (
	SecureJsonDataKeyPassword          SecureJsonDataKey = "password"
	SecureJsonDataKeyAzureClientSecret SecureJsonDataKey = "azureClientSecret"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
	SecureJsonDataKeyAzureClientSecret,
}

// Config is the fully loaded configuration of a Microsoft SQL Server
// datasource instance. The MSSQL backend reads a mix of root-level fields
// (URL, User, Database) and jsonData fields — see pkg/mssql/mssql.go:29-64.
type Config struct {
	// Root-level fields the backend reads from backend.DataSourceInstanceSettings.
	URL      string `json:"-"`
	User     string `json:"-"`
	Database string `json:"-"`

	// jsonData fields.
	JSONDatabase              string          `json:"database,omitempty"`
	AuthenticationType        AuthType        `json:"authenticationType,omitempty"`
	Encrypt                   EncryptOption   `json:"encrypt,omitempty"`
	TLSSkipVerify             bool            `json:"tlsSkipVerify,omitempty"`
	SSLRootCertFile           string          `json:"sslRootCertFile,omitempty"`
	ServerName                string          `json:"serverName,omitempty"`
	KeytabFilePath            string          `json:"keytabFilePath,omitempty"`
	CredentialCache           string          `json:"credentialCache,omitempty"`
	CredentialCacheLookupFile string          `json:"credentialCacheLookupFile,omitempty"`
	ConfigFilePath            string          `json:"configFilePath,omitempty"`
	UDPConnectionLimit        int             `json:"UDPConnectionLimit,omitempty"`
	EnableDNSLookupKDC        string          `json:"enableDNSLookupKDC,omitempty"`
	AzureCredentials          json.RawMessage `json:"azureCredentials,omitempty"` // opaque — see @grafana/azure-sdk
	TimeInterval              string          `json:"timeInterval,omitempty"`
	ConnectionTimeout         int             `json:"connectionTimeout,omitempty"`
	MaxOpenConns              int             `json:"maxOpenConns,omitempty"`
	MaxIdleConns              int             `json:"maxIdleConns,omitempty"`
	MaxIdleConnsAuto          bool            `json:"maxIdleConnsAuto,omitempty"`
	ConnMaxLifetime           int             `json:"connMaxLifetime,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (password, azureClientSecret).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// EffectiveDatabase returns the effective database name, mirroring the backend
// fallback at pkg/mssql/mssql.go:49-52.
func (c Config) EffectiveDatabase() string {
	if c.JSONDatabase != "" {
		return c.JSONDatabase
	}
	return c.Database
}

// LoadConfig runs the full parse -> ApplyDefaults -> Validate flow.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading mssql datasource config")

	cfg := Config{
		URL:                     settings.URL,
		User:                    settings.User,
		Database:                settings.Database,
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
		logger.Error("mssql datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("mssql datasource config loaded",
		"hasURL", cfg.URL != "",
		"authType", cfg.AuthenticationType,
		"encrypt", cfg.Encrypt,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with editor-parity defaults.
// Curated:
//   - AuthenticationType: 'SQL Server Authentication' (matches ConfigurationEditor.tsx:340
//     `jsonData.authenticationType || MSSQLAuthenticationType.sqlAuth`).
//   - Encrypt: 'false' (matches ConfigurationEditor.tsx:231 `jsonData.encrypt || MSSQLEncryptOptions.false`
//     and backend default at pkg/mssql/mssql.go:33).
//   - UDPConnectionLimit: 1 (matches Kerberos.tsx:184 default and pkg/mssql/kerberos/kerberos.go:37).
func (c *Config) ApplyDefaults() {
	if c.AuthenticationType == "" {
		c.AuthenticationType = AuthTypeSQL
	}
	if c.Encrypt == "" {
		c.Encrypt = EncryptFalse
	}
	if c.UDPConnectionLimit == 0 {
		c.UDPConnectionLimit = 1
	}
}

// Validate checks the runtime contract. Encodes the required-when rules from
// ConfigurationEditor.tsx and Kerberos.tsx. Errors are joined.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("host (root.url) is required"))
	}
	if c.EffectiveDatabase() == "" {
		errs = append(errs, errors.New("database (jsonData.database) is required"))
	}

	switch c.AuthenticationType {
	case AuthTypeSQL, AuthTypeKerberosRaw:
		if c.User == "" {
			errs = append(errs, fmt.Errorf("username (root.user) is required for %q", c.AuthenticationType))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
			errs = append(errs, fmt.Errorf("password (secureJsonData.password) is required for %q", c.AuthenticationType))
		}
	case AuthTypeKerberosKeytab:
		if c.User == "" {
			errs = append(errs, errors.New("username (root.user) is required for 'Windows AD: Keytab' auth"))
		}
		if c.KeytabFilePath == "" {
			errs = append(errs, errors.New("keytabFilePath (jsonData.keytabFilePath) is required for 'Windows AD: Keytab' auth"))
		}
	case AuthTypeKerberosCache:
		if c.CredentialCache == "" {
			errs = append(errs, errors.New("credentialCache (jsonData.credentialCache) is required for 'Windows AD: Credential cache' auth"))
		}
	case AuthTypeKerberosCacheLookupFile:
		if c.User == "" {
			errs = append(errs, errors.New("username (root.user) is required for 'Windows AD: Credential cache file' auth"))
		}
		if c.CredentialCacheLookupFile == "" {
			errs = append(errs, errors.New("credentialCacheLookupFile (jsonData.credentialCacheLookupFile) is required for 'Windows AD: Credential cache file' auth"))
		}
	case AuthTypeWindows, AuthTypeAzureAD:
		// Windows Authentication is SSO; Azure AD credentials live in a nested object we don't validate deeply here.
	case "":
		errs = append(errs, errors.New("authenticationType (jsonData.authenticationType) is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown authenticationType: %q", c.AuthenticationType))
	}

	switch c.Encrypt {
	case "", EncryptDisable, EncryptFalse, EncryptTrue:
		// ok
	default:
		errs = append(errs, fmt.Errorf("unknown encrypt option: %q (want disable, false, or true)", c.Encrypt))
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
	if c.ConnectionTimeout < 0 {
		errs = append(errs, fmt.Errorf("connectionTimeout must be non-negative, got %d", c.ConnectionTimeout))
	}
	if c.UDPConnectionLimit < 0 {
		errs = append(errs, fmt.Errorf("UDPConnectionLimit must be non-negative, got %d", c.UDPConnectionLimit))
	}

	return errors.Join(errs...)
}
