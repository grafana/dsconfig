// Package astradbdatasource contains the configuration models for the
// AstraDB datasource plugin (grafana-astradb-datasource).
package astradbdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id field.
const PluginID = "grafana-astradb-datasource"

// AuthKind is the authentication mode selected in the configuration editor
// ("Authentication" radio). Stored in jsonData.authKind as a JSON number.
// Mirrors AuthType (uint8) + AuthTypeToken/AuthTypeCredentials in
// pkg/models/settings.go:11-16, and Connection.TOKEN/CREDENTIALS in
// src/components/ConfigEditor.tsx:11-14.
type AuthKind uint8

const (
	// AuthKindToken authenticates against DataStax Astra Cloud with an
	// application token over TLS. Numeric value 0 (default).
	AuthKindToken AuthKind = 0
	// AuthKindCredentials authenticates against a self-hosted Stargate
	// deployment with basic-auth username/password against the REST auth
	// endpoint. Numeric value 1.
	AuthKindCredentials AuthKind = 1
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyToken is the DataStax Astra application token used in
	// AuthKindToken mode. Consumed by NewStaticTokenProvider at
	// pkg/plugin/handlers_querydata.go:108.
	SecureJsonDataKeyToken SecureJsonDataKey = "token"
	// SecureJsonDataKeyPassword is the basic-auth password used in
	// AuthKindCredentials mode. Consumed by NewTableBasedTokenProvider at
	// pkg/plugin/handlers_querydata.go:117,125.
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyToken,
	SecureJsonDataKeyPassword,
}

// Config is the fully loaded configuration of an AstraDB datasource instance.
// The plugin stores nothing plugin-specific at the root level (url, basicAuth,
// user, etc. are unused — user lives in jsonData), so only the parsed jsonData
// fields and decrypted secure data live here. Callers reach everything
// directly as cfg.AuthKind, cfg.URI, etc.; enumerate configured secrets by
// iterating DecryptedSecureJSONData.
//
// The jsonData fields intentionally deviate from pkg/models/settings.go's
// Settings struct in one place: upstream's Settings.Token (json:"token") and
// Settings.Password (json:"password") are populated only via mapstructure
// decoding of DecryptedSecureJSONData (see pkg/models/settings.go:40-54).
// They are never actually stored in jsonData. This entry treats those secrets
// as first-class secureJsonData values and does NOT surface them as jsonData
// fields on Config — DecryptedSecureJSONData[token]/[password] carries them.
// Treat that upstream aliasing as a bug that the plugin's own LoadSettings
// should sync to this shape.
type Config struct {
	// jsonData fields, matching pkg/models/settings.go:18-27 minus the
	// misleading Token/Password aliases (see the type comment above).
	AuthKind     AuthKind `json:"authKind,omitempty"`
	URI          string   `json:"uri,omitempty"`
	GRPCEndpoint string   `json:"grpcEndpoint,omitempty"`
	AuthEndpoint string   `json:"authEndpoint,omitempty"`
	UserName     string   `json:"user,omitempty"`
	Secure       bool     `json:"secure,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (token, password).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It runs
// the full three-phase flow: parse -> ApplyDefaults -> Validate. Callers that
// need each phase individually can invoke ApplyDefaults and Validate directly
// on the returned Config.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// This is the intended shape for the plugin's own upstream LoadSettings
// (pkg/models/settings.go:29-56) to sync to. Two deliberate deviations from
// upstream:
//  1. Secure values are copied into DecryptedSecureJSONData by key rather than
//     mapstructure-decoded into Settings.Token / Settings.Password. The
//     effective mapping is identical (decryptedSecureJSONData["token"] ->
//     the secret we look up as SecureJsonDataKeyToken).
//  2. Both secrets are loaded unconditionally instead of gated on AuthKind, so
//     switching auth modes at runtime does not lose the "other" secret from
//     the loaded Config. Callers still authenticate using only the secret
//     that matches AuthKind.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading astradb datasource config")

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
		logger.Error("astradb datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("astradb datasource config loaded",
		"authKind", cfg.AuthKind,
		"hasURI", cfg.URI != "",
		"hasGRPCEndpoint", cfg.GRPCEndpoint != "",
		"secure", cfg.Secure,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the editor writes for a fresh datasource (mirroring the "" example
// in SettingsExamples).
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - AuthKind → AuthKindToken. This is a no-op in numeric terms
//     (AuthKindToken == 0 already), but it makes intent explicit and matches
//     the editor's `jsonData.authKind || Connection.TOKEN` fallback in
//     src/components/ConfigEditor.tsx:71.
//
// Everything else — URI, GRPCEndpoint, AuthEndpoint, UserName, Secure — is
// intentionally left as its zero value so ApplyDefaults never overwrites an
// intentional empty string or false.
func (c *Config) ApplyDefaults() {
	// AuthKindToken is the numeric zero value, so this branch is effectively
	// a no-op. It exists so any future non-zero default (or a signed AuthKind
	// type) still lands correctly, and as documentation of intent.
	if c.AuthKind == 0 {
		c.AuthKind = AuthKindToken
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime
// contract, mirroring the CheckHealth checks in
// pkg/plugin/handlers_checkhealth.go:12-40:
//
//   - AuthKindToken requires a non-empty URI and a token secret.
//   - AuthKindCredentials requires non-empty grpcEndpoint, authEndpoint,
//     user, and a password secret.
//   - Any other AuthKind value is rejected.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	switch c.AuthKind {
	case AuthKindToken:
		if c.URI == "" {
			errs = append(errs, errors.New("uri is required for token auth (authKind=0)"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyToken] == "" {
			errs = append(errs, errors.New("token is required for token auth (authKind=0)"))
		}
	case AuthKindCredentials:
		if c.GRPCEndpoint == "" {
			errs = append(errs, errors.New("grpcEndpoint is required for credentials auth (authKind=1)"))
		}
		if c.AuthEndpoint == "" {
			errs = append(errs, errors.New("authEndpoint is required for credentials auth (authKind=1)"))
		}
		if c.UserName == "" {
			errs = append(errs, errors.New("user is required for credentials auth (authKind=1)"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
			errs = append(errs, errors.New("password is required for credentials auth (authKind=1)"))
		}
	default:
		errs = append(errs, fmt.Errorf("unknown authKind %d (expected 0 for Token or 1 for Credentials)", c.AuthKind))
	}

	return errors.Join(errs...)
}
