// Package solarwindsdatasource contains the configuration models for the
// SolarWinds datasource plugin (grafana-solarwinds-datasource) from the
// grafana/plugins monorepo. It has no hand-written config editor or per-plugin
// backend settings model; both are provided by the shared
// github.com/grafana/plugins/sdk/pluginspec SDK and specialized by the plugin's
// src/spec.ts.
package solarwindsdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-solarwinds-datasource"

// ServiceID is the single service id declared by the SolarWinds spec.
const ServiceID = "solarwinds"

// AuthMethodID identifies the authentication method selected for a service,
// stored in jsonData.services.<serviceId>.auth.id (the $defs.authMethods key).
type AuthMethodID string

const (
	// AuthMethodBasic is the only auth method the SolarWinds spec exposes
	// (spec.ts $defs.authMethods.basic_auth, type "basic", showTLSOptions: true).
	AuthMethodBasic AuthMethodID = "basic_auth"
)

// SecureJsonDataKey is a strictly-typed secureJsonData map key (write-only; read
// existing config via secureJsonFields). Secrets use flat dotted keys of the
// form "<serviceId>.<secret>".
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the basic-auth password for the "solarwinds" service.
	SecureJsonDataKeyPassword SecureJsonDataKey = "solarwinds.password"
	// SecureJsonDataKeyTLSSelfSignedCert is the self-signed CA certificate (when enabled).
	SecureJsonDataKeyTLSSelfSignedCert SecureJsonDataKey = "solarwinds.tls.selfSignedCert"
	// SecureJsonDataKeyTLSClientCert is the TLS client certificate (when client auth enabled).
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "solarwinds.tls.clientCert"
	// SecureJsonDataKeyTLSClientKey is the TLS client key (when client auth enabled).
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "solarwinds.tls.clientKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
	SecureJsonDataKeyTLSSelfSignedCert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// Config is the fully loaded configuration of a SolarWinds datasource instance: a
// spec-specific projection of the SDK's generic service-keyed jsonData model
// (sdk/pluginspec/pluginclient/config.go) into concrete nested structs so the
// dsconfig jsonData⇔struct conformance guard holds. Nothing is stored at root.
type Config struct {
	Services  ServicesConfig  `json:"services"`
	Variables VariablesConfig `json:"variables"`

	// DecryptedSecureJSONData holds the decrypted secure values by key.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ServicesConfig holds the per-service configuration keyed by service id.
type ServicesConfig struct {
	Solarwinds ServiceConfig `json:"solarwinds"`
}

// ServiceConfig is the configuration for the "solarwinds" service.
type ServiceConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the auth block stored at jsonData.services.solarwinds.auth.
type AuthConfig struct {
	// Id is the selected auth method id ($defs.authMethods key); "basic_auth".
	Id AuthMethodID `json:"id,omitempty"`
	// UserName is the basic-auth username.
	UserName string `json:"username,omitempty"`
	// TLS holds the optional TLS settings shown when the auth method has showTLSOptions.
	TLS TLSConfig `json:"tls"`
}

// TLSConfig holds the TLS settings stored at jsonData.services.solarwinds.auth.tls.
type TLSConfig struct {
	SelfSignedCert   SelfSignedCertConfig `json:"selfSignedCert"`
	ClientAuth       ClientAuthConfig     `json:"clientAuth"`
	SkipVerification bool                 `json:"skipVerification,omitempty"`
}

// SelfSignedCertConfig toggles the self-signed CA certificate.
type SelfSignedCertConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// ClientAuthConfig holds TLS client-authentication settings.
type ClientAuthConfig struct {
	Enabled    bool   `json:"enabled,omitempty"`
	ServerName string `json:"serverName,omitempty"`
}

// VariablesConfig holds the connection variables shared across services.
type VariablesConfig struct {
	// URL is the SolarWinds instance URL; base URL is {url}:17774/SolarWinds/InformationService/v3/Json.
	URL string `json:"url,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a fully-defaulted,
// validated Config: parse jsonData + copy the decrypted secrets the plugin reads,
// ApplyDefaults (auth.id → basic_auth), then Validate.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading solarwinds datasource config")

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
		logger.Error("solarwinds datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("solarwinds datasource config loaded", "authMethod", cfg.Services.Solarwinds.Auth.Id)
	return cfg, nil
}

// ApplyDefaults fills the zero-valued auth discriminator with the framework's
// default (auth.id → basic_auth, the server's first auth method).
func (c *Config) ApplyDefaults() {
	if c.Services.Solarwinds.Auth.Id == "" {
		c.Services.Solarwinds.Auth.Id = AuthMethodBasic
	}
}

// Validate enforces the health-check contract: a known auth method with its
// required inputs (basic auth username + password), the url used to build the
// base URL, and — when a TLS option is enabled — its accompanying certificate(s).
func (c Config) Validate() error {
	var errs []error

	auth := c.Services.Solarwinds.Auth
	switch auth.Id {
	case AuthMethodBasic:
		if auth.UserName == "" {
			errs = append(errs, errors.New("username is required for basic auth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
			errs = append(errs, errors.New(`password (secureJsonData "solarwinds.password") is required for basic auth`))
		}
	case "":
		errs = append(errs, errors.New("auth method is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown auth method %q", auth.Id))
	}

	if c.Variables.URL == "" {
		errs = append(errs, errors.New("url is required to build the SolarWinds API URL"))
	}

	if auth.TLS.SelfSignedCert.Enabled && c.DecryptedSecureJSONData[SecureJsonDataKeyTLSSelfSignedCert] == "" {
		errs = append(errs, errors.New(`self-signed certificate (secureJsonData "solarwinds.tls.selfSignedCert") is required when it is enabled`))
	}
	if auth.TLS.ClientAuth.Enabled {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert] == "" {
			errs = append(errs, errors.New(`client certificate (secureJsonData "solarwinds.tls.clientCert") is required when TLS client auth is enabled`))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey] == "" {
			errs = append(errs, errors.New(`client key (secureJsonData "solarwinds.tls.clientKey") is required when TLS client auth is enabled`))
		}
	}

	return errors.Join(errs...)
}
