// Package atlassianstatuspagedatasource contains the configuration models for
// the Atlassian Statuspage datasource plugin (grafana-atlassianstatuspage-datasource)
// from the grafana/plugins monorepo. It has no hand-written config editor or
// per-plugin backend settings model; both are provided by the shared
// github.com/grafana/plugins/sdk/pluginspec SDK and specialized by the plugin's
// src/spec.ts.
//
// This datasource queries the public Atlassian Statuspage API and has no
// authentication, so it stores no secureJsonData secrets.
package atlassianstatuspagedatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field.
const PluginID = "grafana-atlassianstatuspage-datasource"

// ServiceID is the single service id declared by the Atlassian Statuspage spec.
const ServiceID = "atlassianstatuspage"

// SecureJsonDataConfig lists the secret key names stored in secureJsonData. This
// datasource has no authentication, so the list is empty.
type SecureJsonDataConfig []string

// SecureJsonDataKeys are the secret keys the plugin reads (none).
var SecureJsonDataKeys = SecureJsonDataConfig{}

// Config is the fully loaded configuration of an Atlassian Statuspage datasource
// instance: a spec-specific projection of the SDK's generic jsonData model
// (sdk/pluginspec/pluginclient/config.go) into concrete nested structs so the
// dsconfig jsonData⇔struct conformance guard holds. Nothing is stored at root and
// there are no secrets.
type Config struct {
	Variables VariablesConfig `json:"variables"`
}

// VariablesConfig holds the connection variables shared across services.
type VariablesConfig struct {
	// URL is the Statuspage URL; the base URL is built as {url}/api/v2.
	URL string `json:"url,omitempty"`
}

// LoadConfig parses a datasource instance's settings into a fully-defaulted,
// validated Config: parse jsonData, ApplyDefaults (no-op — no discriminators),
// then Validate (url required). There are no secrets to copy.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading atlassian statuspage datasource config")

	cfg := Config{}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("atlassian statuspage datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("atlassian statuspage datasource config loaded", "hasURL", cfg.Variables.URL != "")
	return cfg, nil
}

// ApplyDefaults is a no-op: this datasource has no auth discriminator or other
// zero-valued fields to default. It is kept exported for API parity with the
// other registry entries.
func (c *Config) ApplyDefaults() {}

// Validate enforces the runtime contract: the url used to build the base URL.
func (c Config) Validate() error {
	if c.Variables.URL == "" {
		return errors.New("url is required to build the Atlassian Statuspage API URL")
	}
	return nil
}
