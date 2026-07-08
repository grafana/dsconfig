// Package zendeskdatasource contains the configuration models for the Zendesk
// datasource plugin (grafana-zendesk-datasource) from the grafana/plugins
// monorepo. It has no hand-written config editor or per-plugin backend settings
// model; both are provided by the shared
// github.com/grafana/plugins/sdk/pluginspec SDK and specialized by the plugin's
// src/spec.ts.
package zendeskdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-zendesk-datasource"

// ServiceID is the single service id declared by the Zendesk spec
// (spec.ts services: ['zendesk']; the service name shown in the editor is "Tickets").
const ServiceID = "zendesk"

// AuthMethodID identifies the authentication method selected for a service,
// stored in jsonData.services.<serviceId>.auth.id. It is the $defs.authMethods
// key (not the auth type), mirroring how the pluginspec SDK stores it.
type AuthMethodID string

const (
	// AuthMethodBasic is the only auth method the Zendesk spec exposes
	// (spec.ts $defs.authMethods.basic_auth, type "basic").
	AuthMethodBasic AuthMethodID = "basic_auth"
)

// SecureJsonDataKey is a strictly-typed secureJsonData map key (write-only; read
// existing config via secureJsonFields). Secrets use flat dotted keys of the
// form "<serviceId>.<secret>".
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the Zendesk API token used for basic auth on
	// the "zendesk" service. The SDK reads it as
	// DecryptedSecureJSONData["zendesk.password"]
	// (sdk/pluginspec/pluginclient/pluginclient.go).
	SecureJsonDataKeyPassword SecureJsonDataKey = "zendesk.password"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
}

// Config is the fully loaded configuration of a Zendesk datasource instance.
//
// These plugins share the SDK's service-keyed jsonData model
// (sdk/pluginspec/pluginclient/config.go: JsonData{Services map[string]ServiceConfig,
// Variables map[string]string, ...}). This entry projects that generic map-based
// model into concrete, spec-specific nested structs so the dsconfig
// single-source-of-truth conformance guard (schema jsonData fields ⇔ struct json
// tags) holds; the JSON on the wire is identical to what the SDK parses.
//
// The plugin stores nothing at the datasource root level (the base URL is derived
// from the subdomain variable), so only the parsed jsonData fields and decrypted
// secrets (DecryptedSecureJSONData) live here.
type Config struct {
	Services  ServicesConfig  `json:"services"`
	Variables VariablesConfig `json:"variables"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (zendesk.password).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ServicesConfig holds the per-service configuration keyed by service id.
type ServicesConfig struct {
	Zendesk ServiceConfig `json:"zendesk"`
}

// ServiceConfig is the configuration for the "zendesk" (Tickets) service.
type ServiceConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the auth block stored at jsonData.services.zendesk.auth.
type AuthConfig struct {
	// Id is the selected auth method id ($defs.authMethods key); "basic_auth" for Zendesk.
	Id AuthMethodID `json:"id,omitempty"`
	// UserName is the basic-auth username (the Zendesk login email; editor label "Email").
	UserName string `json:"username,omitempty"`
}

// VariablesConfig holds the connection variables shared across services.
type VariablesConfig struct {
	// Subdomain builds the server URL https://{subdomain}.zendesk.com/api/v2/.
	Subdomain string `json:"subdomain,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a Config and returns a
// fully-defaulted, validated Config. It runs the full three-phase flow:
//
//  1. Parse — unmarshal jsonData into Config and copy the decrypted secrets the
//     plugin reads (zendesk.password). This mirrors how the shared backend SDK
//     parses settings in sdk/pluginspec/pluginclient/pluginclient.go.
//  2. ApplyDefaults — fill zero-valued discriminators with the framework's
//     defaults (auth.id → basic_auth).
//  3. Validate — enforce the health-check contract (auth method + inputs, subdomain).
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so log
// lines carry the request/plugin context Grafana injects. Callers that need each
// phase individually can invoke ApplyDefaults and Validate directly.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading zendesk datasource config")

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
		logger.Error("zendesk datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("zendesk datasource config loaded",
		"authMethod", cfg.Services.Zendesk.Auth.Id,
		"hasSubdomain", cfg.Variables.Subdomain != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued discriminators with the same
// defaults the backend SDK applies for a fresh datasource
// (sdk/pluginspec/pluginclient/pluginclient.go defaults an empty auth.id to the
// server's first auth method). Only these fields are touched, and only when
// zero-valued:
//   - Services.Zendesk.Auth.Id → AuthMethodBasic
func (c *Config) ApplyDefaults() {
	if c.Services.Zendesk.Auth.Id == "" {
		c.Services.Zendesk.Auth.Id = AuthMethodBasic
	}
}

// Validate checks that a loaded Config satisfies the runtime contract required
// for a working Zendesk datasource: a known auth method with its required inputs,
// and the subdomain used to build the API base URL.
//
// The shared backend SDK (pluginclient.New) is lenient — it parses and
// builds service clients without hard-failing on missing credentials — but a
// health check (and every query) fails at request time via applyAuth when the
// basic-auth username/password are empty
// (sdk/pluginspec/pluginclient/serviceclient.go), and the server URL is unusable
// without a subdomain. Validate encodes that health-check contract. Errors are
// joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	switch c.Services.Zendesk.Auth.Id {
	case AuthMethodBasic:
		if c.Services.Zendesk.Auth.UserName == "" {
			errs = append(errs, errors.New("username (email) is required for basic auth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
			errs = append(errs, errors.New(`API token (secureJsonData "zendesk.password") is required for basic auth`))
		}
	case "":
		errs = append(errs, errors.New("auth method is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown auth method %q", c.Services.Zendesk.Auth.Id))
	}

	if c.Variables.Subdomain == "" {
		errs = append(errs, errors.New("subdomain is required to build the Zendesk API URL"))
	}

	return errors.Join(errs...)
}
