// Package adobeanalyticsdatasource contains the configuration models for the
// Adobe Analytics datasource plugin (grafana-adobeanalytics-datasource) from the
// grafana/plugins monorepo. It has no hand-written config editor or per-plugin
// backend settings model; both are provided by the shared
// github.com/grafana/plugins/sdk/pluginspec SDK and specialized by the plugin's
// src/spec.ts.
package adobeanalyticsdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-adobeanalytics-datasource"

// ServiceID is the single service id declared by the Adobe Analytics spec.
const ServiceID = "adobe_analytics"

// AuthMethodID identifies the authentication method selected for a service,
// stored in jsonData.services.<serviceId>.auth.id (the $defs.authMethods key).
type AuthMethodID string

const (
	// AuthMethodOAuth2M2M is the only auth method the Adobe Analytics spec exposes
	// (spec.ts $defs.authMethods.oauth2_m2m, type "oauth2_client_credentials").
	AuthMethodOAuth2M2M AuthMethodID = "oauth2_m2m"
)

// SecureJsonDataKey is a strictly-typed secureJsonData map key (write-only; read
// existing config via secureJsonFields). Secrets use flat dotted keys of the
// form "<serviceId>.<secret>".
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyClientSecret is the OAuth2 client secret for the
	// "adobe_analytics" service. The SDK reads it as
	// DecryptedSecureJSONData["adobe_analytics.clientSecret"].
	SecureJsonDataKeyClientSecret SecureJsonDataKey = "adobe_analytics.clientSecret"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyClientSecret,
}

// Config is the fully loaded configuration of an Adobe Analytics datasource
// instance: a spec-specific projection of the SDK's generic service-keyed
// jsonData model (sdk/pluginspec/pluginclient/config.go) into concrete nested
// structs so the dsconfig jsonData⇔struct conformance guard holds. Nothing is
// stored at root.
type Config struct {
	Services  ServicesConfig  `json:"services"`
	Variables VariablesConfig `json:"variables"`

	// DecryptedSecureJSONData holds the decrypted secure values by key (adobe_analytics.clientSecret).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// ServicesConfig holds the per-service configuration keyed by service id.
type ServicesConfig struct {
	AdobeAnalytics ServiceConfig `json:"adobe_analytics"`
}

// ServiceConfig is the configuration for the "adobe_analytics" service.
type ServiceConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the auth block stored at jsonData.services.adobe_analytics.auth.
type AuthConfig struct {
	// Id is the selected auth method id ($defs.authMethods key); "oauth2_m2m".
	Id AuthMethodID `json:"id,omitempty"`
	// ClientId is the OAuth2 client id.
	ClientId string `json:"clientId,omitempty"`
}

// VariablesConfig holds the connection variables shared across services.
type VariablesConfig struct {
	// GlobalCompanyID builds the base URL https://analytics.adobe.io/api/{global_company_id}.
	GlobalCompanyID string `json:"global_company_id,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a fully-defaulted,
// validated Config: parse jsonData + copy the decrypted secret
// (adobe_analytics.clientSecret), ApplyDefaults (auth.id → oauth2_m2m), then Validate.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading adobe analytics datasource config")

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
		logger.Error("adobe analytics datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("adobe analytics datasource config loaded", "authMethod", cfg.Services.AdobeAnalytics.Auth.Id)
	return cfg, nil
}

// ApplyDefaults fills the zero-valued auth discriminator with the framework's
// default (auth.id → oauth2_m2m, the server's first auth method).
func (c *Config) ApplyDefaults() {
	if c.Services.AdobeAnalytics.Auth.Id == "" {
		c.Services.AdobeAnalytics.Auth.Id = AuthMethodOAuth2M2M
	}
}

// Validate enforces the health-check contract: a known auth method with its
// required inputs (OAuth2 client id + client secret), and the global_company_id
// used to build the base URL.
func (c Config) Validate() error {
	var errs []error

	switch c.Services.AdobeAnalytics.Auth.Id {
	case AuthMethodOAuth2M2M:
		if c.Services.AdobeAnalytics.Auth.ClientId == "" {
			errs = append(errs, errors.New("clientId is required for OAuth2 client-credentials auth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] == "" {
			errs = append(errs, errors.New(`client secret (secureJsonData "adobe_analytics.clientSecret") is required`))
		}
	case "":
		errs = append(errs, errors.New("auth method is required"))
	default:
		errs = append(errs, fmt.Errorf("unknown auth method %q", c.Services.AdobeAnalytics.Auth.Id))
	}

	if c.Variables.GlobalCompanyID == "" {
		errs = append(errs, errors.New("global_company_id is required to build the Adobe Analytics API URL"))
	}

	return errors.Join(errs...)
}
