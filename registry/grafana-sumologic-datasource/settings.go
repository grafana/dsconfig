// Package sumologicdatasource contains the configuration models for the
// Grafana Sumo Logic datasource plugin (id: grafana-sumologic-datasource).
package sumologicdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:3 in the upstream plugin).
const PluginID = "grafana-sumologic-datasource"

// Default values applied to empty/zero jsonData fields on load. They mirror the
// upstream backend constants (pkg/models/settings.go:17-19) and the frontend
// constants (src/constants.ts:53,55).
const (
	// DefaultApiURL is the API URL applied when jsonData.apiUrl is empty
	// (pkg/models/settings.go:17,43-45).
	DefaultApiURL = "https://api.sumologic.com/api/"
	// DefaultTimeout is the request timeout (seconds) applied when
	// jsonData.timeout is zero (pkg/models/settings.go:18,46-48).
	DefaultTimeout = 30
	// DefaultInterval is the log-polling interval (milliseconds) applied when
	// jsonData.interval is zero (pkg/models/settings.go:19,49-51).
	DefaultInterval = 1000
)

// AuthenticationMethod is the authentication method discriminator stored in
// jsonData.authMethod. Mirrors the upstream type (pkg/models/settings.go:11-15).
type AuthenticationMethod string

const (
	// AuthenticationMethodAccessKey is the only method the plugin supports:
	// HTTP basic auth with an access ID + access key (pkg/sumo/client.go:51-55).
	AuthenticationMethodAccessKey AuthenticationMethod = "accessKey"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAccessKey is the Sumo Logic Access Key, used as the HTTP
	// basic-auth password (pkg/models/settings.go:39; pkg/sumo/client.go:54).
	SecureJsonDataKeyAccessKey SecureJsonDataKey = "accessKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads when loading settings
// (pkg/models/settings.go:39).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessKey,
}

// Config is the fully loaded configuration of a Sumo Logic datasource instance.
//
// The jsonData fields mirror the upstream backend Settings struct
// (pkg/models/settings.go:21-28) — same fields, same json tags — except the
// decrypted access key, which the upstream keeps as Settings.AccessKey
// (json:"-", populated from DecryptedSecureJSONData) and which this entry holds
// in the DecryptedSecureJSONData map instead. The plugin reads no root-level
// datasource fields (it builds its own HTTP basic auth from jsonData.accessId +
// secureJsonData.accessKey, pkg/sumo/client.go:51-55), so none are carried here.
type Config struct {
	// jsonData fields (pkg/models/settings.go:22-27).
	AuthenticationMethod AuthenticationMethod `json:"authMethod"`
	ApiURL               string               `json:"apiUrl"`
	AccessID             string               `json:"accessId"`
	Timeout              int                  `json:"timeout"`
	Interval             int                  `json:"interval"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the upstream LoadSettings (pkg/models/settings.go:30-54): treat a nil/empty
// jsonData as {}, unmarshal it into Config, copy the decrypted access key, and
// apply the same defaults (authMethod, apiUrl, timeout, interval).
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Upstream applies the defaults inside LoadSettings and validates
// separately (via Client.Validate in the health check, pkg/sumo/client.go:24-26
// and pkg/plugin/handlers_checkhealth.go:14); this entry folds both into
// LoadConfig. Callers that assemble a Config themselves can invoke ApplyDefaults
// and Validate individually.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading sumologic datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Copy decrypted secrets by known key name (pkg/models/settings.go:39).
	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("sumologic datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("sumologic datasource config loaded",
		"authMethod", cfg.AuthenticationMethod,
		"apiUrl", cfg.ApiURL,
		"hasAccessId", cfg.AccessID != "",
	)
	return cfg, nil
}

// ApplyDefaults fills the zero-valued fields with the same defaults the upstream
// LoadSettings applies (pkg/models/settings.go:40-51). Never blanket-apply every
// schema default — this list is curated to match the plugin's own defaulting.
//
// Curated defaults:
//   - AuthenticationMethod -> AuthenticationMethodAccessKey (:40-42)
//   - ApiURL              -> DefaultApiURL (:43-45)
//   - Timeout             -> DefaultTimeout (:46-48)
//   - Interval            -> DefaultInterval (:49-51)
func (c *Config) ApplyDefaults() {
	if c.AuthenticationMethod == "" {
		c.AuthenticationMethod = AuthenticationMethodAccessKey
	}
	if c.ApiURL == "" {
		c.ApiURL = DefaultApiURL
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	if c.Interval == 0 {
		c.Interval = DefaultInterval
	}
}

// Validate checks the runtime contract the plugin enforces in Settings.Validate
// (pkg/models/settings.go:56-72), which the health check runs via
// Client.Validate. The error strings match the upstream verbatim. Errors are
// joined so callers see every problem at once.
//
// Contract:
//   - apiUrl is required (:58-60).
//   - authMethod is required (:61-63).
//   - for accessKey auth: accessId (:65-67) and accessKey (:68-70) are required.
func (c Config) Validate() error {
	var errs []error

	if c.ApiURL == "" {
		errs = append(errs, errors.New("invalid API URL"))
	}
	if c.AuthenticationMethod == "" {
		errs = append(errs, errors.New("invalid authentication method"))
	}
	if c.AuthenticationMethod == AuthenticationMethodAccessKey {
		if c.AccessID == "" {
			errs = append(errs, errors.New("invalid access id"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey] == "" {
			errs = append(errs, errors.New("invalid access key"))
		}
	}

	return errors.Join(errs...)
}
