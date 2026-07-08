// Package yugabytedatasource contains the configuration models for the
// Yugabyte datasource plugin (id: grafana-yugabyte-datasource).
package yugabytedatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-yugabyte-datasource"

// SecureJsonDataKey is a strictly-typed name of a secret stored in secureJsonData.
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the Yugabyte user password. Read at
	// pkg/settings.go:39 and interpolated into the pgx connection string.
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
}

// Config is the fully loaded configuration of a Yugabyte datasource instance.
// The plugin's own backend Settings struct (pkg/settings.go:11-16) is intentionally
// mirrored here, plus DecryptedSecureJSONData for the decrypted password. The
// Yugabyte backend reads root URL and User directly from
// backend.DataSourceInstanceSettings (pkg/settings.go:25,38) and jsonData.database
// via json.Unmarshal (pkg/settings.go:15,43).
type Config struct {
	// Root-level fields the backend reads from backend.DataSourceInstanceSettings.
	// Tagged json:"-" to avoid collisions with jsonData unmarshal.
	URL  string `json:"-"`
	User string `json:"-"`

	// Connection is derived from URL by SplitHostPort — see pkg/settings.go:25-34.
	Connection Connection `json:"-"`

	// Database is populated from jsonData.database via json.Unmarshal.
	// Mirrors pkg/settings.go:15 verbatim (Settings.Database).
	Database string `json:"database,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// Connection mirrors pkg/settings.go:18-22 verbatim — the parts of the URL the
// backend splits out before assembling the pgx connection string.
type Connection struct {
	URL  string
	Host string
	Port string
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/settings.go:24-49) verbatim: SplitHostPort on
// the root URL first (so an invalid URL fails fast), then unmarshal jsonData
// into the Database field, then copy decrypted secrets. Runs
// parse → ApplyDefaults → Validate.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading yugabyte datasource config")

	host, port, err := net.SplitHostPort(settings.URL)
	if err != nil {
		logger.Error("failed to split host:port from URL", "url", settings.URL, "err", err)
		return Config{}, fmt.Errorf("parse url: %w", err)
	}

	cfg := Config{
		URL:  settings.URL,
		User: settings.User,
		Connection: Connection{
			URL:  settings.URL,
			Host: host,
			Port: port,
		},
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return Config{}, fmt.Errorf("parse jsonData: %w", err)
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
		logger.Error("yugabyte datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("yugabyte datasource config loaded",
		"host", cfg.Connection.Host,
		"port", cfg.Connection.Port,
		"hasUser", cfg.User != "",
		"database", cfg.Database,
	)
	return cfg, nil
}

// ApplyDefaults is intentionally a no-op: the Yugabyte plugin has no
// discriminator or license field that needs an editor-parity default. It is
// kept exported so callers assembling a Config directly still get a uniform
// LoadConfig-style parse → ApplyDefaults → Validate flow.
func (c *Config) ApplyDefaults() {}

// Validate checks the runtime contract. The editor marks Host URL, Database,
// and Username as required (ConfigEditor.tsx:40,49,62); the password is
// optional in the editor but the backend interpolates it verbatim into the
// connection string. Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("host URL (root.url) is required"))
	} else if c.Connection.Host == "" || c.Connection.Port == "" {
		errs = append(errs, errors.New("host URL must be of the form 'host:port' (e.g. 'localhost:5433')"))
	}
	if c.User == "" {
		errs = append(errs, errors.New("username (root.user) is required"))
	}
	if c.Database == "" {
		errs = append(errs, errors.New("database name (jsonData.database) is required"))
	}

	return errors.Join(errs...)
}
