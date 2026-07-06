// Package oracledatasource contains the configuration models for the Oracle
// Database datasource plugin (id: grafana-oracle-datasource).
package oracledatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream repo).
const PluginID = "grafana-oracle-datasource"

// ConnectionType is the connection method selected in the configuration editor
// ("Connection methods"). It is an editor-local selector derived from
// jsonData.useTNSNamesBasedConnection; it is not stored under this name.
type ConnectionType string

const (
	// ConnectionTypeTCP is "Host with TCP Port" (host + port + database).
	ConnectionTypeTCP ConnectionType = "tcp"
	// ConnectionTypeTNS is "TNSNames Entry" (a tnsnames.ora entry).
	ConnectionTypeTNS ConnectionType = "tns"
)

// AuthType is the authentication method selected in the configuration editor.
// It is an editor-local selector derived from jsonData.useKerberosAuthentication;
// it is not stored under this name.
type AuthType string

const (
	// AuthTypeBasic is Basic Authentication (username + password).
	AuthTypeBasic AuthType = "basic"
	// AuthTypeKerberos is Kerberos Authentication (no username/password).
	AuthTypeKerberos AuthType = "kerberos"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the Oracle user password (basic auth).
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
}

// Backend defaults mirror pkg/models/settings.go:35-39 and the defaulting logic
// in ConnectionOptions (pkg/models/settings.go:51-75).
const (
	// DefaultConnectionPoolSize is applied when connectionPoolSize is 0.
	DefaultConnectionPoolSize = 50
	// DefaultDataProxyTimeout is applied (seconds) when dataProxyTimeout is 0.
	DefaultDataProxyTimeout = 120
	// DefaultRowLimit is applied when rowLimit is <= 0.
	DefaultRowLimit int64 = 1000000
	// DefaultTimezone is applied when timezone_name is empty.
	DefaultTimezone = "UTC"
)

// Config is the fully loaded configuration of an Oracle datasource instance.
// The Oracle backend reads exactly one root-level field (URL, from
// backend.DataSourceInstanceSettings.URL — pkg/models/settings.go:79) plus the
// decrypted "password" secret; everything else comes from jsonData. Callers reach
// values directly as cfg.User, cfg.Database, cfg.UseTNSNamesBasedConnection, etc.;
// the password is in DecryptedSecureJSONData.
//
// The jsonData fields mirror the plugin's pkg/models/settings.go
// DBConnectionOptions json tags verbatim (minus the computed/non-jsonData fields
// and the excluded Secure Socks Proxy toggle).
type Config struct {
	// Root-level field the backend reads from backend.DataSourceInstanceSettings.
	// json:"-" because it does not live in jsonData.
	URL string `json:"-"`

	// jsonData fields (json tags match pkg/models/settings.go:22-32).
	TZName                     string `json:"timezone_name,omitempty"`
	DSTEnabled                 bool   `json:"use_dst,omitempty"`
	Database                   string `json:"database,omitempty"`
	User                       string `json:"user,omitempty"`
	UseTNSNamesBasedConnection bool   `json:"useTNSNamesBasedConnection,omitempty"`
	TNSNamesEntry              string `json:"tnsNamesEntry,omitempty"`
	UseKerberosAuthentication  bool   `json:"useKerberosAuthentication,omitempty"`
	ConnectionPoolSize         int    `json:"connectionPoolSize,omitempty"`
	DataProxyTimeout           int    `json:"dataProxyTimeout,omitempty"`
	PrefetchRowsCount          int    `json:"prefetchRowsCount,omitempty"`
	RowLimit                   int64  `json:"rowLimit,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (password).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ConnectionType returns the editor-local connection-method selector value
// derived from UseTNSNamesBasedConnection (ConfigEditor.tsx:183).
func (c Config) ConnectionType() ConnectionType {
	if c.UseTNSNamesBasedConnection {
		return ConnectionTypeTNS
	}
	return ConnectionTypeTCP
}

// AuthType returns the editor-local authentication selector value derived from
// UseKerberosAuthentication (ConfigEditor.tsx:204).
func (c Config) AuthType() AuthType {
	if c.UseKerberosAuthentication {
		return AuthTypeKerberos
	}
	return AuthTypeBasic
}

// EffectiveTNSNamesEntry returns the effective tnsnames.ora entry, mirroring the
// backend legacy fallback at pkg/models/settings.go:82-84: when a TNSNames
// connection has no jsonData.tnsNamesEntry but a root URL is set (v3.3.0 →
// v3.3.2 upgrade), the root URL is used as the entry.
func (c Config) EffectiveTNSNamesEntry() string {
	if c.UseTNSNamesBasedConnection && c.TNSNamesEntry == "" && c.URL != "" {
		return c.URL
	}
	return c.TNSNamesEntry
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's ConnectionOptions loader (pkg/models/settings.go:41-88): the root
// URL is copied from backend.DataSourceInstanceSettings.URL, jsonData is
// unmarshaled verbatim, the decrypted password is copied in, and the legacy
// TNSNames-in-URL fallback is applied.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse (unmarshal jsonData + copy
// secrets + legacy TNSNames fallback), then (*Config).ApplyDefaults for the
// backend-parity defaults ConnectionOptions applies inline, then
// (Config).Validate to enforce the runtime contract. Callers that need each
// phase individually can invoke ApplyDefaults and Validate directly on the
// returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading oracle datasource config")

	cfg := Config{
		URL:                     settings.URL,
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

	// Legacy TNSNames fallback: v3.3.0 stored the entry in the root URL; v3.3.2
	// moved it to jsonData.tnsNamesEntry (pkg/models/settings.go:82-84).
	if cfg.UseTNSNamesBasedConnection && cfg.TNSNamesEntry == "" && cfg.URL != "" {
		logger.Info("migrating legacy TNSNames entry from root url to jsonData.tnsNamesEntry")
		cfg.TNSNamesEntry = cfg.URL
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("oracle datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("oracle datasource config loaded",
		"connectionType", cfg.ConnectionType(),
		"authType", cfg.AuthType(),
		"hasURL", cfg.URL != "",
		"hasTNSNamesEntry", cfg.EffectiveTNSNamesEntry() != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same defaults
// the backend applies inline in ConnectionOptions (pkg/models/settings.go:51-75)
// for a fresh datasource. Never blanket-apply every schema default — that would
// clobber intentional zero values.
//
// Curated defaults:
//   - TZName             → "UTC" when empty (pkg/models/settings.go:51-53)
//   - ConnectionPoolSize → 50 when 0    (pkg/models/settings.go:63-65)
//   - DataProxyTimeout   → 120 when 0   (pkg/models/settings.go:68-70)
//   - RowLimit           → 1000000 when <= 0 (pkg/models/settings.go:73-75)
//
// PrefetchRowsCount is intentionally not defaulted — the backend only appends it
// to the driver connection string when > 0 (pkg/models/settings.go:130-156), so
// its zero value is meaningful. The environment-variable overrides for pool size
// and timeout (GF_PLUGINS_ORACLE_DATASOURCE_POOLSIZE / GF_DATAPROXY_TIMEOUT) are
// deliberately not resolved here — they only apply at runtime when the stored
// value is 0.
func (c *Config) ApplyDefaults() {
	if c.TZName == "" {
		c.TZName = DefaultTimezone
	}
	if c.ConnectionPoolSize == 0 {
		c.ConnectionPoolSize = DefaultConnectionPoolSize
	}
	if c.DataProxyTimeout == 0 {
		c.DataProxyTimeout = DefaultDataProxyTimeout
	}
	if c.RowLimit <= 0 {
		c.RowLimit = DefaultRowLimit
	}
}

// Validate checks the runtime contract the plugin requires to build a working
// connection string (pkg/models/settings.go:90-157) together with the config
// editor's required markers (ConfigEditor.tsx:230,250,274,330,356):
//
//   - Host with TCP Port (useTNSNamesBasedConnection=false): root url and
//     jsonData.database are required (used as <url>/<database>).
//   - TNSNames (useTNSNamesBasedConnection=true): a tnsNamesEntry is required
//     (the legacy root-url fallback counts, via EffectiveTNSNamesEntry).
//   - Basic auth (useKerberosAuthentication=false): jsonData.user and
//     secureJsonData.password are required.
//   - Kerberos auth (useKerberosAuthentication=true): no username/password.
//
// Note: the backend accepts Kerberos with either connection method, even though
// the editor only exposes Kerberos for TNSNames connections — so Validate does
// not reject Kerberos + Host with TCP Port. Errors are joined so callers see
// every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.UseTNSNamesBasedConnection {
		if c.EffectiveTNSNamesEntry() == "" {
			errs = append(errs, errors.New("tnsNamesEntry (jsonData.tnsNamesEntry) is required for TNSNames connections"))
		}
	} else {
		if c.URL == "" {
			errs = append(errs, errors.New("host (root.url) is required for Host with TCP Port connections"))
		}
		if c.Database == "" {
			errs = append(errs, errors.New("database name (jsonData.database) is required for Host with TCP Port connections"))
		}
	}

	if !c.UseKerberosAuthentication {
		if c.User == "" {
			errs = append(errs, errors.New("user (jsonData.user) is required for basic authentication"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
			errs = append(errs, errors.New("password (secureJsonData.password) is required for basic authentication"))
		}
	}

	if c.ConnectionPoolSize < 0 {
		errs = append(errs, fmt.Errorf("connectionPoolSize must be non-negative, got %d", c.ConnectionPoolSize))
	}
	if c.DataProxyTimeout < 0 {
		errs = append(errs, fmt.Errorf("dataProxyTimeout must be non-negative, got %d", c.DataProxyTimeout))
	}
	if c.PrefetchRowsCount < 0 {
		errs = append(errs, fmt.Errorf("prefetchRowsCount must be non-negative, got %d", c.PrefetchRowsCount))
	}
	if c.RowLimit < 0 {
		errs = append(errs, fmt.Errorf("rowLimit must be non-negative, got %d", c.RowLimit))
	}

	return errors.Join(errs...)
}
