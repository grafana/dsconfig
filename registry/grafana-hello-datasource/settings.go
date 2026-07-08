// Package hellodatasource contains the configuration models for the Hello
// datasource plugin (grafana-hello-datasource) from the grafana/plugins monorepo.
// It has no hand-written config editor or per-plugin backend settings model; both
// are provided by the shared github.com/grafana/plugins/sdk/pluginspec SDK and
// specialized by the plugin's src/spec.ts. Hello is an experimental plugin used
// for testing the framework; neither of its services requires authentication, so
// it stores no secureJsonData secrets.
package hellodatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-hello-datasource"

// AuthMethodID identifies the authentication method selected for a service,
// stored in jsonData.services.<serviceId>.auth.id (the $defs.authMethods key).
type AuthMethodID string

const (
	// AuthMethodNone is the only auth method both Hello services expose
	// (spec.ts $defs.authMethods.none, type "none").
	AuthMethodNone AuthMethodID = "none"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData. Hello
// requires no authentication, so the list is empty.
type SecureJsonDataConfig []string

// SecureJsonDataKeys are the secret keys the plugin reads (none).
var SecureJsonDataKeys = SecureJsonDataConfig{}

// Config is the fully loaded configuration of a Hello datasource instance: a
// spec-specific projection of the SDK's generic service-keyed jsonData model
// (sdk/pluginspec/pluginclient/config.go) into concrete nested structs so the
// dsconfig jsonData⇔struct conformance guard holds. Nothing is stored at root and
// there are no secrets.
type Config struct {
	Services ServicesConfig `json:"services"`
}

// ServicesConfig holds the per-service configuration keyed by service id.
type ServicesConfig struct {
	HTTPBin     ServiceConfig `json:"httpbin"`
	PostmanEcho ServiceConfig `json:"postman_echo"`
}

// ServiceConfig is the configuration for a Hello service.
type ServiceConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the auth block stored at jsonData.services.<id>.auth.
type AuthConfig struct {
	// Id is the selected auth method id ($defs.authMethods key); "none".
	Id AuthMethodID `json:"id,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a fully-defaulted,
// validated Config: parse jsonData, ApplyDefaults (both auth.id → none), then
// Validate. There are no secrets to copy.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading hello datasource config")

	cfg := Config{}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("hello datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("hello datasource config loaded")
	return cfg, nil
}

// ApplyDefaults fills the zero-valued auth discriminators with the framework's
// default (each service's auth.id → none).
func (c *Config) ApplyDefaults() {
	if c.Services.HTTPBin.Auth.Id == "" {
		c.Services.HTTPBin.Auth.Id = AuthMethodNone
	}
	if c.Services.PostmanEcho.Auth.Id == "" {
		c.Services.PostmanEcho.Auth.Id = AuthMethodNone
	}
}

// Validate checks that each service's auth method is the expected 'none'. There
// are no credentials to require.
func (c Config) Validate() error {
	var errs []error
	for name, id := range map[string]AuthMethodID{
		"httpbin":      c.Services.HTTPBin.Auth.Id,
		"postman_echo": c.Services.PostmanEcho.Auth.Id,
	} {
		switch id {
		case AuthMethodNone, "":
			// no auth required
		default:
			errs = append(errs, fmt.Errorf("service %q: unknown auth method %q", name, id))
		}
	}
	return errors.Join(errs...)
}
