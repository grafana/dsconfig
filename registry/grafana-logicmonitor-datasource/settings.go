// Package logicmonitordatasource contains the configuration models for the
// LogicMonitor Devices datasource plugin (grafana-logicmonitor-datasource) from
// the grafana/plugins monorepo. It has no hand-written config editor or
// per-plugin backend settings model; both are provided by the shared
// github.com/grafana/plugins/sdk/pluginspec SDK and specialized by the plugin's
// src/spec.ts.
package logicmonitordatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-logicmonitor-datasource"

// ServiceID is the single service id declared by the LogicMonitor spec.
const ServiceID = "logicmonitor"

// AuthMethodID identifies the authentication method selected for a service,
// stored in jsonData.services.<serviceId>.auth.id (the $defs.authMethods key).
type AuthMethodID string

const (
	// AuthMethodAuthBearer is the only auth method the LogicMonitor spec exposes
	// (spec.ts $defs.authMethods.auth_bearer, type "bearer").
	AuthMethodAuthBearer AuthMethodID = "auth_bearer"
)

// SecureJsonDataKey is a strictly-typed secureJsonData map key (write-only; read
// existing config via secureJsonFields). Secrets use flat dotted keys of the
// form "<serviceId>.<secret>".
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the LogicMonitor REST API v3 bearer token for the
	// "logicmonitor" service. The SDK reads it as DecryptedSecureJSONData["logicmonitor.token"].
	SecureJsonDataKeyToken SecureJsonDataKey = "logicmonitor.token"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
}

// Config is the fully loaded configuration of a LogicMonitor datasource instance:
// a spec-specific projection of the SDK's generic service-keyed jsonData model
// (sdk/pluginspec/pluginclient/config.go) into concrete nested structs so the
// dsconfig jsonData⇔struct conformance guard holds. Nothing is stored at root.
type Config struct {
	Services  ServicesConfig  `json:"services"`
	Variables VariablesConfig `json:"variables"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (logicmonitor.token).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ServicesConfig holds the per-service configuration keyed by service id.
type ServicesConfig struct {
	LogicMonitor ServiceConfig `json:"logicmonitor"`
}

// ServiceConfig is the configuration for the "logicmonitor" service.
type ServiceConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the auth block stored at jsonData.services.logicmonitor.auth.
type AuthConfig struct {
	// Id is the selected auth method id ($defs.authMethods key); "auth_bearer".
	Id AuthMethodID `json:"id,omitempty"`
}

// VariablesConfig holds the connection variables shared across services.
type VariablesConfig struct {
	// AccountName is the LogicMonitor account subdomain; the base URL is
	// https://{account_name}.logicmonitor.com/santaba/rest.
	AccountName string `json:"account_name,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a fully-defaulted,
// validated Config: parse jsonData + copy the decrypted secret (logicmonitor.token),
// ApplyDefaults (auth.id → auth_bearer), then Validate.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading logicmonitor datasource config")

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
		logger.Error("logicmonitor datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("logicmonitor datasource config loaded", "authMethod", cfg.Services.LogicMonitor.Auth.Id)
	return cfg, nil
}

// ApplyDefaults fills the zero-valued auth discriminator with the framework's
// default (auth.id → auth_bearer, the server's first auth method).
func (c *Config) ApplyDefaults() {
	if c.Services.LogicMonitor.Auth.Id == "" {
		c.Services.LogicMonitor.Auth.Id = AuthMethodAuthBearer
	}
}

// Validate enforces the health-check contract: a known auth method with its
// required inputs (bearer token), and the account_name used to build the base URL.
func (c Config) Validate() error {
	var errs []error

	switch c.Services.LogicMonitor.Auth.Id {
	case AuthMethodAuthBearer:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, errors.New(`API v3 key (secureJsonData "logicmonitor.token") is required`))
		}
	case "":
		errs = append(errs, errors.New("auth method is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown auth method %q", c.Services.LogicMonitor.Auth.Id))
	}

	if c.Variables.AccountName == "" {
		errs = append(errs, errors.New("account_name is required to build the LogicMonitor API URL"))
	}

	return errors.Join(errs...)
}
