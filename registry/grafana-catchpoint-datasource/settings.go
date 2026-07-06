// Package catchpointdatasource contains the configuration models for the
// Catchpoint datasource plugin (grafana-catchpoint-datasource) from the
// grafana/plugins monorepo. It has no hand-written config editor or per-plugin
// backend settings model; both are provided by the shared
// github.com/grafana/plugins/sdk/pluginspec SDK and specialized by the plugin's
// src/spec.ts.
package catchpointdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-catchpoint-datasource"

// ServiceID is the single service id declared by the Catchpoint spec.
const ServiceID = "catchpoint"

// AuthMethodID identifies the authentication method selected for a service,
// stored in jsonData.services.<serviceId>.auth.id (the $defs.authMethods key).
type AuthMethodID string

const (
	// AuthMethodBearerToken is the only auth method the Catchpoint spec exposes
	// (spec.ts $defs.authMethods.bearer_token, type "bearer").
	AuthMethodBearerToken AuthMethodID = "bearer_token"
)

// SecureJsonDataKey is a strictly-typed secureJsonData map key (write-only; read
// existing config via secureJsonFields). Secrets use flat dotted keys of the
// form "<serviceId>.<secret>".
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the Catchpoint REST API v2 Key (bearer) for the
	// "catchpoint" service. The SDK reads it as DecryptedSecureJSONData["catchpoint.token"].
	SecureJsonDataKeyToken SecureJsonDataKey = "catchpoint.token"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
}

// Config is the fully loaded configuration of a Catchpoint datasource instance: a
// spec-specific projection of the SDK's generic service-keyed jsonData model
// (sdk/pluginspec/pluginclient/config.go) into concrete nested structs so the
// dsconfig jsonData⇔struct conformance guard holds. Nothing is stored at root.
type Config struct {
	Services ServicesConfig `json:"services"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (catchpoint.token).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ServicesConfig holds the per-service configuration keyed by service id.
type ServicesConfig struct {
	Catchpoint ServiceConfig `json:"catchpoint"`
}

// ServiceConfig is the configuration for the "catchpoint" service.
type ServiceConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the auth block stored at jsonData.services.catchpoint.auth.
type AuthConfig struct {
	// Id is the selected auth method id ($defs.authMethods key); "bearer_token".
	Id AuthMethodID `json:"id,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a fully-defaulted,
// validated Config: parse jsonData + copy the decrypted secret (catchpoint.token),
// ApplyDefaults (auth.id → bearer_token), then Validate.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading catchpoint datasource config")

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
		logger.Error("catchpoint datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("catchpoint datasource config loaded", "authMethod", cfg.Services.Catchpoint.Auth.Id)
	return cfg, nil
}

// ApplyDefaults fills the zero-valued auth discriminator with the framework's
// default (auth.id → bearer_token, the server's first auth method).
func (c *Config) ApplyDefaults() {
	if c.Services.Catchpoint.Auth.Id == "" {
		c.Services.Catchpoint.Auth.Id = AuthMethodBearerToken
	}
}

// Validate enforces the health-check contract: a known auth method with its
// required inputs. Bearer auth needs the key.
func (c Config) Validate() error {
	var errs []error

	switch c.Services.Catchpoint.Auth.Id {
	case AuthMethodBearerToken:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, errors.New(`API v2 key (secureJsonData "catchpoint.token") is required`))
		}
	case "":
		errs = append(errs, errors.New("auth method is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown auth method %q", c.Services.Catchpoint.Auth.Id))
	}

	return errors.Join(errs...)
}
