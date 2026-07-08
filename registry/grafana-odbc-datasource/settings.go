// Package odbcdatasource contains the configuration models for the Sqlyze
// (ODBC) datasource plugin (grafana-odbc-datasource).
package odbcdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream plugin).
const PluginID = "grafana-odbc-datasource"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
//
// For this plugin, secret keys are DYNAMIC: each equals the Name of a driver
// setting whose Secure flag is enabled (pkg/models/settings.go:33-41). There
// is no fixed secret key. The constant below is the conventional password key
// ('pwd') from the plugin README's driver-settings table, declared as the
// representative secret so the schema has a concrete secureJsonData key.
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPwd is the conventional password driver-setting key
	// ('pwd'). Representative only — any secure setting Name becomes a secret key.
	SecureJsonDataKeyPwd SecureJsonDataKey = "pwd"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the representative secret keys used by the plugin.
// Runtime secrets are dynamic (see SecureJsonDataKey); LoadConfig resolves
// whatever secure setting names are present regardless of this list.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPwd,
}

// Setting is a single driver setting concatenated into the ODBC connection
// string. Mirrors the upstream Setting struct (pkg/models/settings.go:15-19).
// The upstream struct carries no json tags and relies on Go's case-insensitive
// matching; the tags here use the lowercase keys the config editor actually
// writes (src/ConfigEditor.tsx:78,93,100).
type Setting struct {
	Name   string `json:"name"`
	Value  string `json:"value,omitempty"`
	Secure bool   `json:"secure,omitempty"`
}

// Config is the fully loaded configuration of a Sqlyze (ODBC) datasource
// instance. The backend (pkg/driver.go:27-36) reads nothing at the root level,
// so only parsed jsonData fields and the decrypted secure data live here.
//
// The jsonData fields mirror the upstream Settings struct
// (pkg/models/settings.go:21-26) verbatim, including the backend-only DSN
// field. Storage keys use the lowercase form the editor writes (Go's
// case-insensitive unmarshal keeps the capitalized upstream fixtures working).
type Config struct {
	Driver   string     `json:"driver,omitempty"`
	DSN      string     `json:"DSN,omitempty"`
	Timeout  string     `json:"timeout,omitempty"`
	Settings []*Setting `json:"settings,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values, keyed by the
	// secure setting's Name (dynamic).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/models/settings.go:28-54): unmarshal jsonData,
// then resolve every secure driver setting's value from the decrypted secure
// data keyed by the setting's Name (failing with "Missing <name>" if absent),
// then default the timeout and validate.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse (unmarshal jsonData +
// resolve dynamic secrets) -> ApplyDefaults (timeout) -> Validate (driver +
// timeout contract). Callers that need each phase individually can invoke
// ApplyDefaults and Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading odbc datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Resolve secure driver settings from the decrypted secure data, keyed by
	// the setting's Name (mirrors pkg/models/settings.go:33-41). Secret keys
	// are dynamic — there is no fixed secret key.
	for _, s := range cfg.Settings {
		if s == nil || !s.Secure {
			continue
		}
		val, ok := settings.DecryptedSecureJSONData[s.Name]
		if !ok {
			logger.Error("missing secure driver setting", "name", s.Name)
			return cfg, fmt.Errorf("Missing %s", s.Name)
		}
		s.Value = val
		cfg.DecryptedSecureJSONData[SecureJsonDataKey(s.Name)] = val
	}

	logger.Debug("resolved secure driver settings", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("odbc datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("odbc datasource config loaded",
		"hasDriver", cfg.Driver != "",
		"hasDSN", cfg.DSN != "",
		"settingsCount", len(cfg.Settings),
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin writes for a fresh datasource. Never blanket-apply every
// schema default — that would clobber intentional zero values.
//
// Curated defaults:
//   - Timeout: "10" when empty — mirrors LoadSettings (pkg/models/settings.go:50-52).
func (c *Config) ApplyDefaults() {
	if c.Timeout == "" {
		c.Timeout = "10"
	}
}

// Validate checks the runtime contract the plugin enforces at settings-load
// time. LoadSettings unconditionally rejects an empty driver via
// CheckDriverFileExists (pkg/models/settings.go:43-44,68-71), and Connect
// parses the timeout with strconv.Atoi (pkg/database/connect.go:18), so a
// non-integer timeout makes every connection fail.
//
// Filesystem checks that CheckDriverFileExists performs on path-style drivers
// (existence, executable bit) are intentionally omitted here — they depend on
// the runtime host and cannot be evaluated at config-load time.
//
// Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.Driver == "" {
		errs = append(errs, errors.New("driver is required"))
	}

	if c.Timeout != "" {
		if _, err := strconv.Atoi(c.Timeout); err != nil {
			errs = append(errs, fmt.Errorf("timeout must be an integer number of seconds, got %q", c.Timeout))
		}
	}

	return errors.Join(errs...)
}
