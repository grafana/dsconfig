// Package postgresqldatasource contains the configuration models for the
// PostgreSQL datasource plugin (id: grafana-postgresql-datasource).
package postgresqldatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
// The plugin.json also declares an aliasIDs entry for the legacy short name 'postgres'.
const PluginID = "grafana-postgresql-datasource"

// TLSMode is the PostgreSQL TLS negotiation mode (jsonData.sslmode).
type TLSMode string

const (
	TLSModeDisable    TLSMode = "disable"
	TLSModeRequire    TLSMode = "require"
	TLSModeVerifyCA   TLSMode = "verify-ca"
	TLSModeVerifyFull TLSMode = "verify-full"
)

// TLSMethod is how TLS certificates are supplied (jsonData.tlsConfigurationMethod).
type TLSMethod string

const (
	// TLSMethodFilePath reads certs from paths on the Grafana server (sslRootCertFile / sslCertFile / sslKeyFile).
	TLSMethodFilePath TLSMethod = "file-path"
	// TLSMethodFileContent reads inline PEM content from secureJsonData (tlsCACert / tlsClientCert / tlsClientKey).
	TLSMethodFileContent TLSMethod = "file-content"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in secureJsonData.
type SecureJsonDataKey string

const (
	SecureJsonDataKeyPassword      SecureJsonDataKey = "password"
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

// Config is the fully loaded configuration of a PostgreSQL datasource instance.
// The PostgreSQL backend reads a mix of root-level fields (URL, User, Database)
// and jsonData fields — see pkg/postgresql/postgres.go:87-115. Callers reach
// everything directly as cfg.URL, cfg.User, etc.; enumerate configured secrets
// by iterating DecryptedSecureJSONData.
type Config struct {
	// Root-level fields the backend reads from backend.DataSourceInstanceSettings.
	URL      string `json:"-"`
	User     string `json:"-"`
	Database string `json:"-"`

	// jsonData fields.
	JSONDatabase           string    `json:"database,omitempty"`
	SSLMode                TLSMode   `json:"sslmode,omitempty"`
	TLSConfigurationMethod TLSMethod `json:"tlsConfigurationMethod,omitempty"`
	SSLRootCertFile        string    `json:"sslRootCertFile,omitempty"`
	SSLCertFile            string    `json:"sslCertFile,omitempty"`
	SSLKeyFile             string    `json:"sslKeyFile,omitempty"`
	PostgresVersion        int       `json:"postgresVersion,omitempty"`
	Timescaledb            bool      `json:"timescaledb,omitempty"`
	TimeInterval           string    `json:"timeInterval,omitempty"`
	MaxOpenConns           int       `json:"maxOpenConns,omitempty"`
	ConnMaxLifetime        int       `json:"connMaxLifetime,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// EffectiveDatabase returns the effective database name, mirroring the backend
// fallback at pkg/postgresql/postgres.go:101-104: prefer jsonData.database when
// set, otherwise root.database.
func (c Config) EffectiveDatabase() string {
	if c.JSONDatabase != "" {
		return c.JSONDatabase
	}
	return c.Database
}

// LoadConfig parses a datasource instance's settings into a Config. Runs
// parse → ApplyDefaults → Validate.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading postgresql datasource config")

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
		logger.Error("postgresql datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("postgresql datasource config loaded",
		"hasURL", cfg.URL != "",
		"hasUser", cfg.User != "",
		"database", cfg.EffectiveDatabase(),
		"sslmode", cfg.SSLMode,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Curated:
//   - SSLMode: 'require' (matches ConfigurationEditor.tsx:197)
//   - TLSConfigurationMethod: 'file-path' (matches ConfigurationEditor.tsx:234
//     and backend default at postgres.go:92)
func (c *Config) ApplyDefaults() {
	if c.SSLMode == "" {
		c.SSLMode = TLSModeRequire
	}
	if c.TLSConfigurationMethod == "" {
		c.TLSConfigurationMethod = TLSMethodFilePath
	}
}

// Validate checks the runtime contract. The editor marks Host URL, Username,
// and Database name as required (ConfigurationEditor.tsx:129,140,156). Errors
// are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("host URL (root.url) is required"))
	}
	if c.User == "" {
		errs = append(errs, errors.New("username (root.user) is required"))
	}
	if c.EffectiveDatabase() == "" {
		errs = append(errs, errors.New("database name (jsonData.database) is required"))
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

	// When file-content is selected under a non-disable sslmode, we need at least a client cert+key
	// pair if the user wants mutual TLS; verify-ca / verify-full additionally need the CA cert.
	if c.SSLMode != "" && c.SSLMode != TLSModeDisable && c.TLSConfigurationMethod == TLSMethodFileContent {
		if c.SSLMode == TLSModeVerifyCA || c.SSLMode == TLSModeVerifyFull {
			if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert] == "" {
				errs = append(errs, errors.New("tlsCACert (secureJsonData) is required when tlsConfigurationMethod is 'file-content' and sslmode is 'verify-ca' or 'verify-full'"))
			}
		}
	}

	if c.MaxOpenConns < 0 {
		errs = append(errs, fmt.Errorf("maxOpenConns must be non-negative, got %d", c.MaxOpenConns))
	}
	if c.ConnMaxLifetime < 0 {
		errs = append(errs, fmt.Errorf("connMaxLifetime must be non-negative, got %d", c.ConnMaxLifetime))
	}

	return errors.Join(errs...)
}
