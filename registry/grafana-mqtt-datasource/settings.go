// Package mqttdatasource contains the configuration models for the MQTT
// datasource plugin (plugin id: grafana-mqtt-datasource).
package mqttdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream repo).
const PluginID = "grafana-mqtt-datasource"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the MQTT basic-auth password. Applied
	// via paho.SetPassword at pkg/mqtt/client.go:58-60 only when non-empty.
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// SecureJsonDataKeyTLSCACert is the PEM-encoded CA certificate used to
	// verify the MQTT server's TLS certificate. Loaded via
	// x509.NewCertPool().AppendCertsFromPEM at pkg/mqtt/client.go:75-79 when
	// non-empty.
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
	// SecureJsonDataKeyTLSClientCert is the PEM-encoded client certificate
	// used for mutual TLS. Loaded via tls.X509KeyPair(tlsClientCert,
	// tlsClientKey) at pkg/mqtt/client.go:66-73 when either the cert or the
	// key is non-empty.
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the PEM-encoded private key that
	// pairs with tlsClientCert. Loaded via the same tls.X509KeyPair call.
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
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

// Config is the fully loaded configuration of an MQTT datasource instance.
//
// The jsonData fields mirror the plugin's frontend MqttDataSourceOptions
// (src/types.ts:10-17), not the backend mqtt.Options (pkg/mqtt/client.go:26-35):
// the backend struct omits tlsAuth and tlsAuthWithCACert because it treats
// them as editor-visibility toggles, but the schema surfaces them because
// the editor writes them into jsonData. Follow-up in the plugin's own
// LoadSettings should sync to this shape.
//
// Root-level datasource fields (settings.URL, BasicAuthEnabled, User, etc.)
// are NOT carried on Config because the MQTT plugin never reads them:
// pkg/plugin/datasource.go:60-83 only unmarshals settings.JSONData and copies
// keys out of settings.DecryptedSecureJSONData.
type Config struct {
	URI               string `json:"uri"`
	ClientID          string `json:"clientID"`
	Username          string `json:"username"`
	TLSAuth           bool   `json:"tlsAuth"`
	TLSAuthWithCACert bool   `json:"tlsAuthWithCACert"`
	TLSSkipVerify     bool   `json:"tlsSkipVerify"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (password, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config,
// mirroring pkg/plugin/datasource.go:60-83 (getDatasourceSettings): unmarshal
// jsonData into the config struct, then copy each known secret from
// s.DecryptedSecureJSONData into DecryptedSecureJSONData.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext
// so log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Callers that need each phase individually can invoke
// ApplyDefaults and Validate directly on a Config they assemble
// themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading mqtt datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream getDatasourceSettings (pkg/plugin/datasource.go:63-65) calls
	// json.Unmarshal on settings.JSONData unconditionally. Mirror that:
	// only skip parsing when JSONData is entirely empty so a fresh
	// datasource with no jsonData yet is not a parse error, while any
	// present-but-malformed bytes still surface as an error.
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
		logger.Error("mqtt datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("mqtt datasource config loaded",
		"hasURI", cfg.URI != "",
		"hasClientID", cfg.ClientID != "",
		"tlsSkipVerify", cfg.TLSSkipVerify,
		"tlsAuth", cfg.TLSAuth,
		"tlsAuthWithCACert", cfg.TLSAuthWithCACert,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with editor-parity
// defaults. The MQTT plugin's editor writes explicit `false` for each of the
// three boolean toggles (src/types.ts:14-16 declares them as required
// booleans; the editor treats missing values as `|| false` at
// src/ConfigEditor.tsx:92,99,103). The zero value of a Go bool is already
// false, so this method is currently a no-op. It exists so any future
// non-zero default (e.g. a discriminator that lands as an int) still applies
// consistently, and to document the intent that ApplyDefaults is the sole
// entry point for populating editor-parity defaults.
func (c *Config) ApplyDefaults() {
	// Intentionally empty. All schema defaults land at Go zero values.
	_ = c
}

// Validate checks the runtime contract that the plugin requires. Mirrors
// the implicit checks in pkg/mqtt/client.go:
//
//   - URI must be non-empty: opts.AddBroker(o.URI) is called unconditionally
//     at pkg/mqtt/client.go:46 and connecting to an empty broker address
//     fails at Connect time.
//   - If either tlsClientCert or tlsClientKey is non-empty, both must be
//     present: tls.X509KeyPair (pkg/mqtt/client.go:66-73) parses them as a
//     PEM keypair and fails when only one side is provided.
//
// The plugin does not require basic-auth username/password (both are only
// applied when non-empty), does not require a CA cert (the CA cert pool is
// only populated when tlsCACert is non-empty), and does not read tlsAuth or
// tlsAuthWithCACert at all — so this validator ignores those toggles.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URI == "" {
		errs = append(errs, errors.New("mqtt broker URI (jsonData.uri) is required"))
	}

	clientCert := c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert]
	clientKey := c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey]
	switch {
	case clientCert != "" && clientKey == "":
		errs = append(errs, errors.New("secureJsonData.tlsClientKey is required when secureJsonData.tlsClientCert is set"))
	case clientKey != "" && clientCert == "":
		errs = append(errs, errors.New("secureJsonData.tlsClientCert is required when secureJsonData.tlsClientKey is set"))
	}

	return errors.Join(errs...)
}
