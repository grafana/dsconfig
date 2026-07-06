// Package saphanadatasource contains the configuration models for the SAP HANA®
// datasource plugin (id: grafana-saphana-datasource).
package saphanadatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream repo).
const PluginID = "grafana-saphana-datasource"

// DefaultTimeout is the connection timeout (in seconds, as a string) the backend
// applies when jsonData.timeout is empty (pkg/models/settings.go:72-74).
const DefaultTimeout = "30"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the SAP HANA user password (basic auth).
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// SecureJsonDataKeyTLSClientCert is the X.509 client certificate PEM (used when tlsAuth is enabled).
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the X.509 client private key PEM (used when tlsAuth is enabled).
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

// Validation errors. The first four mirror pkg/models/errors.go verbatim; the last
// two encode the X.509 client-auth contract the driver requires at connect time
// (pkg/plugin/driver.go:51,77 — tls.X509KeyPair fails on an empty cert/key).
var (
	// ErrInvalidServerName mirrors ErrorMessageInvalidServerName (errors.go:7).
	ErrInvalidServerName = errors.New("invalid server name. Either empty or not set")
	// ErrInvalidPort mirrors ErrorMessageInvalidPort (errors.go:8).
	ErrInvalidPort = errors.New("invalid port or instance. add a port or tenant instance + tenant database name")
	// ErrInvalidUsername mirrors ErrorMessageInvalidUserName (errors.go:9).
	ErrInvalidUsername = errors.New("username is either empty or not set and TLS Client authentication is not enabled")
	// ErrInvalidPassword mirrors ErrorMessageInvalidPassword (errors.go:10).
	ErrInvalidPassword = errors.New("password is either empty or not set and TLS Client authentication is not enabled")
	// ErrMissingTLSClientCert is returned when tlsAuth is enabled without a client certificate.
	ErrMissingTLSClientCert = errors.New("tlsClientCert (secureJsonData) is required when TLS Client Auth (tlsAuth) is enabled")
	// ErrMissingTLSClientKey is returned when tlsAuth is enabled without a client key.
	ErrMissingTLSClientKey = errors.New("tlsClientKey (secureJsonData) is required when TLS Client Auth (tlsAuth) is enabled")
)

// Config is the fully loaded configuration of a SAP HANA datasource instance.
// The SAP HANA backend reads NO root-level fields — LoadSettings
// (pkg/models/settings.go:54-83) unmarshals only jsonData and reads secrets from
// DecryptedSecureJSONData, and the connection is built purely from those
// (pkg/plugin/driver.go:62-107). Callers reach values directly as cfg.Server,
// cfg.Port, cfg.TlsClientAuth, etc.; secrets are in DecryptedSecureJSONData.
//
// The jsonData fields mirror the plugin's pkg/models/settings.go Settings struct
// verbatim (same field names, same json tags) minus the four secret fields
// (Password, TlsCACert, TlsClientCert, TlsClientKey — moved into
// DecryptedSecureJSONData) and the excluded Secure Socks Proxy ProxyOptions.
//
// Note (upstream quirk): the backend tags Password `json:"-,omitempty"`, which
// names the field "-" rather than skipping it; it is only ever populated from
// DecryptedSecureJSONData, so the effect matches a plain skip. This entry keeps
// the password out of the struct entirely and reads it from the secure map.
type Config struct {
	// jsonData fields (json tags match pkg/models/settings.go:20-34).
	Server             string `json:"server,omitempty"`
	Instance           string `json:"instance,omitempty"`
	DatabaseName       string `json:"databaseName,omitempty"`
	Username           string `json:"username,omitempty"`
	DefaultSchema      string `json:"defaultSchema,omitempty"`
	Timeout            string `json:"timeout,omitempty"`
	Port               int64  `json:"port,omitempty"`
	TlsDisabled        bool   `json:"tlsDisabled,omitempty"`
	InsecureSkipVerify bool   `json:"tlsSkipVerify,omitempty"`
	TlsClientAuth      bool   `json:"tlsAuth,omitempty"`
	TlsAuthWithCACert  bool   `json:"tlsAuthWithCACert,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (password, tlsClientCert, tlsClientKey, tlsCACert).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/models/settings.go:54-83): jsonData is
// unmarshaled verbatim, the decrypted secrets are copied in, the timeout default
// is applied, and the IsValid runtime contract is enforced.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse (unmarshal jsonData + copy
// secrets), then (*Config).ApplyDefaults for the timeout default the backend
// applies inline, then (Config).Validate to enforce IsValid's contract. This is
// the intended shape for upstream to sync to. Callers that need each phase
// individually can invoke ApplyDefaults and Validate directly on the returned
// Config.
//
// Upstream quirk mirrored in behaviour, not in control-flow: when no password is
// supplied and tlsAuth is false, LoadSettings short-circuits and returns
// IsValid() before applying the timeout default or copying the TLS certs
// (settings.go:58-61). Because that path always yields a validation error, the
// three-phase flow here returns an equivalent error; the only difference is that
// ApplyDefaults still runs first, which is inconsequential for a rejected config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading saphana datasource config")

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
		logger.Error("saphana datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("saphana datasource config loaded",
		"hasServer", cfg.Server != "",
		"hasPort", cfg.Port != 0,
		"instanceConnection", cfg.Instance != "" && cfg.DatabaseName != "",
		"tlsDisabled", cfg.TlsDisabled,
		"tlsClientAuth", cfg.TlsClientAuth,
	)
	return cfg, nil
}

// ApplyDefaults fills the curated set of zero-valued fields with the same
// defaults the backend applies inline in LoadSettings for a fresh datasource
// (pkg/models/settings.go:72-74). Never blanket-apply every schema default —
// that would clobber intentional zero values.
//
// Curated defaults:
//   - Timeout → "30" when empty (pkg/models/settings.go:72-74)
//
// TlsDisabled is intentionally not defaulted: its zero value (false) already
// means "TLS enabled", which is the editor default (ConfigEditor.tsx:173).
func (c *Config) ApplyDefaults() {
	if c.Timeout == "" {
		c.Timeout = DefaultTimeout
	}
}

// Validate checks the runtime contract the plugin requires, mirroring
// Settings.IsValid (pkg/models/settings.go:37-51) and adding the X.509
// client-auth cert/key requirement the driver enforces at connect time
// (pkg/plugin/driver.go:51,77):
//
//   - server is always required.
//   - a port OR both a tenant instance and a tenant database name is required
//     (the backend rejects port==0 with an empty instance or database name).
//   - basic auth (tlsAuth=false): username and password are both required.
//   - TLS client auth (tlsAuth=true): a client certificate and key are required
//     (username/password are not).
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.Server == "" {
		errs = append(errs, ErrInvalidServerName)
	}
	if c.Port == 0 && (c.Instance == "" || c.DatabaseName == "") {
		errs = append(errs, ErrInvalidPort)
	}

	if c.TlsClientAuth {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert] == "" {
			errs = append(errs, ErrMissingTLSClientCert)
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey] == "" {
			errs = append(errs, ErrMissingTLSClientKey)
		}
	} else {
		if c.Username == "" {
			errs = append(errs, ErrInvalidUsername)
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
			errs = append(errs, ErrInvalidPassword)
		}
	}

	return errors.Join(errs...)
}
