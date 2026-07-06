// Package mysqldatasource contains the configuration models for the MySQL
// datasource plugin (id: mysql).
package mysqldatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "mysql"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the MySQL user password.
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// SecureJsonDataKeyTLSClientCert is the client certificate PEM (used when tlsAuth is enabled).
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the client private key PEM (used when tlsAuth is enabled).
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
	// SecureJsonDataKeyTLSCACert is the CA certificate PEM (used when tlsAuthWithCACert is enabled).
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
	SecureJsonDataKeyTLSCACert,
}

// Config is the fully loaded configuration of a MySQL datasource instance.
// The MySQL backend reads a mix of root-level fields (URL, User, Database)
// via backend.DataSourceInstanceSettings AND jsonData fields — see
// pkg/mysql/mysql.go:63-72 for the DataSourceInfo assembly. Callers reach
// everything directly as cfg.URL, cfg.User, cfg.Database, cfg.TLSAuth, etc.;
// enumerate configured secrets by iterating DecryptedSecureJSONData.
//
// The jsonData fields mirror the plugin's pkg/mysql/sqleng/sql_engine.go
// JsonData struct plus the MySQL-specific AllowCleartextPasswords addition
// and the connection-pool fields shared with the other sqleng datasources.
type Config struct {
	// Root-level fields (json:"-" on the struct because they don't live in jsonData).
	// The MySQL backend reads these directly off backend.DataSourceInstanceSettings.
	URL      string `json:"-"`
	User     string `json:"-"`
	Database string `json:"-"`

	// jsonData fields, matching pkg/mysql/sqleng/sql_engine.go:40-61 for the shared shape
	// plus AllowCleartextPasswords for MySQL-specific behaviour.
	JSONDatabase            string `json:"database,omitempty"`
	TLSAuth                 bool   `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert       bool   `json:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify           bool   `json:"tlsSkipVerify,omitempty"`
	AllowCleartextPasswords bool   `json:"allowCleartextPasswords,omitempty"`
	Timezone                string `json:"timezone,omitempty"`
	TimeInterval            string `json:"timeInterval,omitempty"`
	MaxOpenConns            int    `json:"maxOpenConns,omitempty"`
	MaxIdleConns            int    `json:"maxIdleConns,omitempty"`
	MaxIdleConnsAuto        bool   `json:"maxIdleConnsAuto,omitempty"`
	ConnMaxLifetime         int    `json:"connMaxLifetime,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (password, tlsClientCert, tlsClientKey, tlsCACert).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// EffectiveDatabase returns the effective database name for a MySQL connection,
// mirroring the backend fallback at pkg/mysql/mysql.go:58-61: prefer
// jsonData.database when set, otherwise root.database.
func (c Config) EffectiveDatabase() string {
	if c.JSONDatabase != "" {
		return c.JSONDatabase
	}
	return c.Database
}

// LoadConfig parses a datasource instance's settings into a Config. Root
// fields (URL, User, Database) are copied from backend.DataSourceInstanceSettings
// directly (matching pkg/mysql/mysql.go:65-72), and jsonData is unmarshaled
// verbatim from settings.JSONData.
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

	logger.Debug("loading mysql datasource config")

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
		logger.Error("mysql datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("mysql datasource config loaded",
		"hasURL", cfg.URL != "",
		"hasUser", cfg.User != "",
		"database", cfg.EffectiveDatabase(),
		"tlsAuth", cfg.TLSAuth,
	)
	return cfg, nil
}

// ApplyDefaults fills in a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// Curated defaults:
//   - MaxIdleConnsAuto: true — matches useMigrateDatabaseFields (packages/grafana-sql)
//     which sets this on first render for datasources that have never had connection
//     pool settings written.
//
// The other connection-pool fields (MaxOpenConns, MaxIdleConns, ConnMaxLifetime) are
// deliberately not defaulted here — the plugin's backend at pkg/mysql/mysql.go:45-51
// pulls Grafana-wide defaults from cfg.SQL().DefaultMax*, which we cannot resolve
// without a Grafana context. Downstream tooling should keep the zero value if
// unspecified and let the backend apply its runtime defaults.
func (c *Config) ApplyDefaults() {
	// If no connection pool fields are set at all, mark maxIdleConnsAuto as true
	// (matches the on-first-render behaviour in packages/grafana-sql
	// useMigrateDatabaseFields.ts).
	if c.MaxOpenConns == 0 && c.MaxIdleConns == 0 && !c.MaxIdleConnsAuto {
		c.MaxIdleConnsAuto = true
	}
}

// Validate checks the runtime contract that the plugin requires. The MySQL
// backend does not itself hard-fail on missing URL/User — it just builds a
// connection string with empty values and lets MySQL reject the login. The
// editor marks host and username as required (`ConfigurationEditor.tsx:77,104`),
// so we encode that here.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("host URL (root.url) is required"))
	}
	if c.User == "" {
		errs = append(errs, errors.New("username (root.user) is required"))
	}

	// TLS client-auth requires both cert and key. TLS CA verification requires the CA cert.
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
