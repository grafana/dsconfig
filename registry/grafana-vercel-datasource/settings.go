// Package verceldatasource contains the configuration models for the Vercel
// datasource plugin (grafana-vercel-datasource) from the grafana/plugins
// monorepo. It has no hand-written config editor or per-plugin backend settings
// model; both are provided by the shared
// github.com/grafana/plugins/sdk/pluginspec SDK and specialized by the plugin's
// src/spec.ts.
package verceldatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-vercel-datasource"

// ServiceID is the single service id declared by the Vercel spec.
const ServiceID = "vercel"

// AuthMethodID identifies the authentication method selected for a service,
// stored in jsonData.services.<serviceId>.auth.id (the $defs.authMethods key).
type AuthMethodID string

const (
	// AuthMethodVercelAPIKey is the only auth method the Vercel spec exposes
	// (spec.ts $defs.authMethods.vercelApiKey, type "bearer").
	AuthMethodVercelAPIKey AuthMethodID = "vercelApiKey"
)

// SecureJsonDataKey is a strictly-typed secureJsonData map key (write-only; read
// existing config via secureJsonFields). Secrets use flat dotted keys of the
// form "<serviceId>.<secret>".
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the Vercel Access Token (bearer) for the
	// "vercel" service. The SDK reads it as DecryptedSecureJSONData["vercel.token"].
	SecureJsonDataKeyToken SecureJsonDataKey = "vercel.token"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
}

// Config is the fully loaded configuration of a Vercel datasource instance.
//
// This projects the SDK's generic service-keyed jsonData model
// (sdk/pluginspec/pluginclient/config.go) into concrete, spec-specific nested
// structs so the dsconfig single-source-of-truth conformance guard (schema
// jsonData fields ⇔ struct json tags) holds. The plugin stores nothing at the
// datasource root level.
type Config struct {
	Services  ServicesConfig  `json:"services"`
	Variables VariablesConfig `json:"variables"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (vercel.token).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ServicesConfig holds the per-service configuration keyed by service id.
type ServicesConfig struct {
	Vercel ServiceConfig `json:"vercel"`
}

// ServiceConfig is the configuration for the "vercel" service.
type ServiceConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the auth block stored at jsonData.services.vercel.auth.
type AuthConfig struct {
	// Id is the selected auth method id ($defs.authMethods key); "vercelApiKey".
	Id AuthMethodID `json:"id,omitempty"`
}

// VariablesConfig holds the connection variables shared across services.
type VariablesConfig struct {
	// TeamID is the optional Vercel team id (only needed for team-scoped tokens).
	TeamID string `json:"team_id,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a fully-defaulted,
// validated Config: parse jsonData + copy the decrypted secret the plugin reads
// (vercel.token), ApplyDefaults (auth.id → vercelApiKey), then Validate.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading vercel datasource config")

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
		logger.Error("vercel datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("vercel datasource config loaded", "authMethod", cfg.Services.Vercel.Auth.Id)
	return cfg, nil
}

// ApplyDefaults fills the zero-valued auth discriminator with the framework's
// default (auth.id → vercelApiKey, the server's first auth method).
func (c *Config) ApplyDefaults() {
	if c.Services.Vercel.Auth.Id == "" {
		c.Services.Vercel.Auth.Id = AuthMethodVercelAPIKey
	}
}

// Validate enforces the health-check contract: a known auth method with its
// required inputs. Bearer auth needs the access token; team_id is optional.
func (c Config) Validate() error {
	var errs []error

	switch c.Services.Vercel.Auth.Id {
	case AuthMethodVercelAPIKey:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, errors.New(`access token (secureJsonData "vercel.token") is required`))
		}
	case "":
		errs = append(errs, errors.New("auth method is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown auth method %q", c.Services.Vercel.Auth.Id))
	}

	return errors.Join(errs...)
}
