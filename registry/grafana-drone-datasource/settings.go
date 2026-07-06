// Package dronedatasource contains the configuration models for the Drone
// datasource plugin (grafana-drone-datasource) from the grafana/plugins
// monorepo. It has no hand-written config editor or per-plugin backend settings
// model; both are provided by the shared
// github.com/grafana/plugins/sdk/pluginspec SDK and specialized by the plugin's
// src/spec.ts.
package dronedatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-drone-datasource"

// ServiceID is the single service id declared by the Drone spec.
const ServiceID = "drone"

// AuthMethodID identifies the authentication method selected for a service,
// stored in jsonData.services.<serviceId>.auth.id (the $defs.authMethods key).
type AuthMethodID string

const (
	// AuthMethodAuthBearer is the only auth method the Drone spec exposes
	// (spec.ts $defs.authMethods.auth_bearer, type "bearer").
	AuthMethodAuthBearer AuthMethodID = "auth_bearer"
)

// SecureJsonDataKey is a strictly-typed secureJsonData map key (write-only; read
// existing config via secureJsonFields). Secrets use flat dotted keys of the
// form "<serviceId>.<secret>".
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the Drone API token (bearer) for the "drone"
	// service. The SDK reads it as DecryptedSecureJSONData["drone.token"].
	SecureJsonDataKeyToken SecureJsonDataKey = "drone.token"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
}

// Config is the fully loaded configuration of a Drone datasource instance: a
// spec-specific projection of the SDK's generic service-keyed jsonData model
// (sdk/pluginspec/pluginclient/config.go) into concrete nested structs so the
// dsconfig jsonData⇔struct conformance guard holds. Nothing is stored at root.
type Config struct {
	Services  ServicesConfig  `json:"services"`
	Variables VariablesConfig `json:"variables"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (drone.token).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ServicesConfig holds the per-service configuration keyed by service id.
type ServicesConfig struct {
	Drone ServiceConfig `json:"drone"`
}

// ServiceConfig is the configuration for the "drone" service.
type ServiceConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the auth block stored at jsonData.services.drone.auth.
type AuthConfig struct {
	// Id is the selected auth method id ($defs.authMethods key); "auth_bearer".
	Id AuthMethodID `json:"id,omitempty"`
}

// VariablesConfig holds the connection variables shared across services.
type VariablesConfig struct {
	// URL is the Drone server URL; the base URL is built as {url}/api.
	URL string `json:"url,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a fully-defaulted,
// validated Config: parse jsonData + copy the decrypted secret (drone.token),
// ApplyDefaults (auth.id → auth_bearer), then Validate.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading drone datasource config")

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
		logger.Error("drone datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("drone datasource config loaded", "authMethod", cfg.Services.Drone.Auth.Id)
	return cfg, nil
}

// ApplyDefaults fills the zero-valued auth discriminator with the framework's
// default (auth.id → auth_bearer, the server's first auth method).
func (c *Config) ApplyDefaults() {
	if c.Services.Drone.Auth.Id == "" {
		c.Services.Drone.Auth.Id = AuthMethodAuthBearer
	}
}

// Validate enforces the health-check contract: a known auth method with its
// required inputs (bearer token), and the url used to build the API base URL.
func (c Config) Validate() error {
	var errs []error

	switch c.Services.Drone.Auth.Id {
	case AuthMethodAuthBearer:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, errors.New(`API token (secureJsonData "drone.token") is required`))
		}
	case "":
		errs = append(errs, errors.New("auth method is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown auth method %q", c.Services.Drone.Auth.Id))
	}

	if c.Variables.URL == "" {
		errs = append(errs, errors.New("url is required to build the Drone API URL"))
	}

	return errors.Join(errs...)
}
