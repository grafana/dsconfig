// Package appdynamicsdatasource contains the configuration models for the
// AppDynamics datasource plugin (id: dlopes7-appdynamics-datasource).
package appdynamicsdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream plugin).
const PluginID = "dlopes7-appdynamics-datasource"

// AuthMethod identifies the resolved Controller (Metrics) API authentication
// method. There is no stored discriminator field; the backend derives the
// method at runtime from which credentials are present. Mirrors the untyped
// iota constants in pkg/appd/auth/auth_provider.go:22-29.
type AuthMethod string

const (
	// AuthMethodBasic authenticates the Controller API with a username +
	// password (root.basicAuthUser + secureJsonData.basicAuthPassword).
	AuthMethodBasic AuthMethod = "basic-auth"
	// AuthMethodAPIClient authenticates the Controller API with an OAuth2
	// client-credentials grant built from jsonData.clientName,
	// jsonData.clientDomain and secureJsonData.clientSecret.
	AuthMethodAPIClient AuthMethod = "api-client"
	// AuthMethodUnknown is returned when neither credential set is complete.
	AuthMethodUnknown AuthMethod = "unknown"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the Controller basic-auth password
	// (pkg/models/settings.go:50). Used only when no clientSecret is set.
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyClientSecret is the API Client (OAuth2) client secret
	// (pkg/models/settings.go:46-47). Takes precedence over basic auth.
	SecureJsonDataKeyClientSecret SecureJsonDataKey = "clientSecret"
	// SecureJsonDataKeyAnalyticsAPIKey is the Analytics (Events) API key, sent
	// as the X-Events-API-Key header (pkg/models/settings.go:53-55;
	// pkg/appd/analytics/client.go:53).
	SecureJsonDataKeyAnalyticsAPIKey SecureJsonDataKey = "analyticsAPIKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads when loading settings
// (pkg/models/settings.go:46-55).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyClientSecret,
	SecureJsonDataKeyAnalyticsAPIKey,
}

// Config is the fully loaded configuration of an AppDynamics datasource
// instance.
//
// The jsonData fields mirror the storage-relevant fields of the upstream
// backend Settings (pkg/models/settings.go:14-33) verbatim — same field names,
// same json tags (TLSSkipVerify, ClientName, ClientDomain, AnalyticsURL,
// AccountName). The upstream Settings also carries runtime-only fields
// (MetricsAuthorization, ProxyOptions, Inputs) that are not configuration
// storage; they are intentionally omitted here.
//
// Root-level fields (MetricsURL, BasicAuthUsername) and the decrypted secrets
// (ClientSecret, BasicAuthPassword, AnalyticsAPIKey) are carried with json:"-"
// tags so they don't collide with jsonData unmarshaling. LoadConfig populates
// them exactly like the upstream LoadSettings does, including the auth gating:
// BasicAuthUsername/BasicAuthPassword are only set when no clientSecret is
// present (pkg/models/settings.go:44-55).
type Config struct {
	// jsonData fields (mirror pkg/models/settings.go:15-27 json tags).
	TLSSkipVerify bool   `json:"tlsSkipVerify"`
	ClientName    string `json:"clientName"`
	ClientDomain  string `json:"clientDomain"`
	AnalyticsURL  string `json:"analyticsURL"`
	AccountName   string `json:"globalAccountName"`

	// Root datasource field the backend reads: config.URL -> MetricsURL
	// (pkg/models/settings.go:44).
	MetricsURL string `json:"-"`

	// Controller (Metrics) API credentials, populated by LoadConfig from the
	// decrypted secrets and root basic-auth username with the same gating the
	// backend applies (clientSecret wins; basic-auth fields only when
	// clientSecret is empty — pkg/models/settings.go:46-51).
	ClientSecret      string `json:"-"`
	BasicAuthUsername string `json:"-"`
	BasicAuthPassword string `json:"-"`

	// Analytics (Events) API key (pkg/models/settings.go:53-55).
	AnalyticsAPIKey string `json:"-"`

	// DecryptedSecureJSONData holds every decrypted secure value present by key
	// (basicAuthPassword, clientSecret, analyticsAPIKey), regardless of the
	// auth gating above. Use it to enumerate which secrets are configured.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the upstream LoadSettings (pkg/models/settings.go:36-63): unmarshal jsonData,
// lift the Controller URL off root config.URL, copy the decrypted secrets, and
// apply the auth gating where a present clientSecret selects API Client auth
// and otherwise the root basic-auth username + basicAuthPassword secret select
// basic auth.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Callers that assemble a Config themselves can invoke ApplyDefaults
// and Validate individually.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading appdynamics datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Controller URL comes from the root instance settings, not jsonData
	// (pkg/models/settings.go:44).
	cfg.MetricsURL = settings.URL

	// Copy decrypted secrets by known key name so callers can enumerate them.
	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	// Auth gating, mirroring pkg/models/settings.go:46-51: a non-empty
	// clientSecret selects API Client auth and suppresses the basic-auth
	// fields; otherwise the root basic-auth username + basicAuthPassword secret
	// are used.
	if secret := cfg.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret]; secret != "" {
		cfg.ClientSecret = secret
	} else {
		cfg.BasicAuthUsername = settings.BasicAuthUser
		cfg.BasicAuthPassword = cfg.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword]
	}

	// Analytics API key (pkg/models/settings.go:53-55).
	if key := cfg.DecryptedSecureJSONData[SecureJsonDataKeyAnalyticsAPIKey]; key != "" {
		cfg.AnalyticsAPIKey = key
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("appdynamics datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("appdynamics datasource config loaded",
		"authMethod", cfg.AuthMethod(),
		"hasURL", cfg.MetricsURL != "",
		"analyticsConfigured", cfg.IsAnalyticsConfigured(),
	)
	return cfg, nil
}

// AuthMethod resolves which Controller (Metrics) API authentication method a
// loaded Config uses, mirroring NewMetricsProvider's precedence
// (pkg/appd/auth/auth_provider.go:55-89): basic auth when a username +
// password are both present; API Client when clientSecret + Controller URL +
// clientName + clientDomain are all present; otherwise unknown. Because
// LoadConfig applies the same gating the backend does, a present clientSecret
// leaves the basic-auth fields empty, so API Client wins as documented.
func (c Config) AuthMethod() AuthMethod {
	if c.BasicAuthPassword != "" && c.BasicAuthUsername != "" {
		return AuthMethodBasic
	}
	if c.ClientSecret != "" && c.MetricsURL != "" && c.ClientName != "" && c.ClientDomain != "" {
		return AuthMethodAPIClient
	}
	return AuthMethodUnknown
}

// IsAnalyticsConfigured reports whether the optional Analytics (Events) API is
// fully configured, mirroring HealthDiagnostics.IsAnalyticsConfigured
// (pkg/appd/health_diagnostics.go:70-84): all of analyticsURL,
// globalAccountName and the analyticsAPIKey secret must be non-empty.
func (c Config) IsAnalyticsConfigured() bool {
	return strings.TrimSpace(c.AnalyticsURL) != "" &&
		strings.TrimSpace(c.AccountName) != "" &&
		strings.TrimSpace(c.AnalyticsAPIKey) != ""
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the config editor writes into a fresh datasource. The AppDynamics
// config editor writes NO defaults into storage (there is no auth-type
// discriminator, and every field starts empty/false), so this is intentionally
// a no-op. It is kept exported for parity with the other registry entries and
// so callers can rely on the same parse -> ApplyDefaults -> Validate contract.
func (c *Config) ApplyDefaults() {
	// No editor-parity defaults: the AppDynamics config editor persists no
	// default values (ConfigEditor.tsx). Fields default to their Go zero values
	// (empty strings, false), matching a brand-new datasource instance.
}

// Validate checks the runtime contract the plugin enforces in its health check
// (pkg/appd/health_diagnostics.go:87-135, CheckSettings), evaluated against the
// same auth-gated state the backend sees. Errors are joined so callers see
// every problem at once.
//
// Contracts enforced:
//   - The Controller URL (root.url -> MetricsURL) is required (:88-90).
//   - At least one Controller auth method must be present: either API Client
//     (clientName/clientDomain/clientSecret) or basic auth
//     (basicAuthUser/basicAuthPassword) (:92-95).
//   - If any API Client field is set, all three are required (:98-114).
//   - If any basic-auth field is set, both are required (:116-128).
//
// The optional Analytics API is NOT validated here: an incomplete Analytics
// config disables Analytics silently upstream (IsAnalyticsConfigured), it is
// not an error.
func (c Config) Validate() error {
	var errs []error

	if strings.TrimSpace(c.MetricsURL) == "" {
		errs = append(errs, errors.New("controller URL (root.url) is required"))
	}

	clientAny := c.ClientSecret != "" || c.ClientName != "" || c.ClientDomain != ""
	basicAny := c.BasicAuthUsername != "" || c.BasicAuthPassword != ""

	if !clientAny && !basicAny {
		errs = append(errs, errors.New("no authentication configured: provide API Client credentials (jsonData.clientName, jsonData.clientDomain, secureJsonData.clientSecret) or basic auth (root.basicAuthUser + secureJsonData.basicAuthPassword)"))
	}

	if clientAny {
		if c.ClientSecret == "" {
			errs = append(errs, errors.New("secureJsonData.clientSecret is required for API Client auth"))
		}
		if c.ClientName == "" {
			errs = append(errs, errors.New("jsonData.clientName is required for API Client auth"))
		}
		if c.ClientDomain == "" {
			errs = append(errs, errors.New("jsonData.clientDomain is required for API Client auth"))
		}
	}

	if basicAny {
		if c.BasicAuthUsername == "" {
			errs = append(errs, errors.New("root.basicAuthUser is required for basic auth"))
		}
		if c.BasicAuthPassword == "" {
			errs = append(errs, errors.New("secureJsonData.basicAuthPassword is required for basic auth"))
		}
	}

	return errors.Join(errs...)
}
