// Package newrelicdatasource contains the configuration models for the
// New Relic datasource plugin (plugin id: grafana-newrelic-datasource).
package newrelicdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:4 in the upstream plugin).
const PluginID = "grafana-newrelic-datasource"

// DefaultTimeoutInSeconds is the HTTP timeout applied when
// jsonData.timeoutInSeconds is absent or less than 1. Mirrors the editor
// default (src/components/ConfigEditor.tsx:57) and the backend coercion
// (pkg/models/settings.go:38-40).
const DefaultTimeoutInSeconds int64 = 300

// Region is the New Relic data-center region selected in the config editor's
// "Region" Select. Stored under jsonData.region. Mirrors the frontend union
// NewRelicSupportedRegion (src/types.ts:4) and the regions option list
// (src/types.ts:187-190); the backend keeps it as a plain string
// (pkg/models/settings.go:13) and passes it to the New Relic client's
// ConfigRegion only when non-empty (pkg/datasource/newrelic_client.go:47-49).
type Region string

const (
	// RegionUS is the New Relic US data center.
	RegionUS Region = "US"
	// RegionEU is the New Relic EU data center.
	RegionEU Region = "EU"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPersonalAPIKey is the New Relic user/personal API key,
	// applied via ConfigPersonalAPIKey (pkg/datasource/newrelic_client.go:43).
	SecureJsonDataKeyPersonalAPIKey SecureJsonDataKey = "personalApiKey"
	// SecureJsonDataKeyAccountID is the New Relic account ID (stored as a
	// string, parsed to a number) used as the NerdGraph $accountId NRQL
	// variable (pkg/datasource/insights/insights_client.go:24-46).
	SecureJsonDataKeyAccountID SecureJsonDataKey = "accountId"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads when loading
// settings (pkg/models/settings.go:34,36).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPersonalAPIKey,
	SecureJsonDataKeyAccountID,
}

// Config is the fully loaded configuration of a New Relic datasource instance.
//
// The jsonData fields mirror the plugin's upstream backend Settings struct
// (pkg/models/settings.go:12-24) verbatim — same field names, same json tags,
// same types — for the jsonData keys (region, timeoutInSeconds, and the three
// internal base-URL overrides). AccountID is the numeric account ID the
// upstream Settings also carries (json:"-"), parsed from the decrypted
// secureJsonData.accountId string; it is the value the NerdGraph queries use
// (pkg/datasource/insights/insights_client.go:24-46).
//
// The raw decrypted secrets (personalApiKey, accountId) live in
// DecryptedSecureJSONData. Root-level datasource fields (settings.URL, basic
// auth, etc.) are NOT carried because the New Relic backend never reads them:
// LoadSettings (pkg/models/settings.go:27-48) reads only config.JSONData and
// config.DecryptedSecureJSONData, and GetNewRelicClient
// (pkg/datasource/newrelic_client.go:30-59) builds the client purely from the
// parsed Settings.
type Config struct {
	// jsonData fields (mirror pkg/models/settings.go:13-23 json tags verbatim).
	Region           Region `json:"region"`
	TimeoutInSeconds int64  `json:"timeoutInSeconds"`
	RestBaseUrl      string `json:"restBaseURL"`
	InfraBaseUrl     string `json:"infrastructureBaseURL"`
	NerdGraphBaseURL string `json:"nerdGraphBaseURL"`

	// AccountID is the numeric New Relic account ID parsed from the decrypted
	// secureJsonData.accountId string (pkg/models/settings.go:36,42-46). Kept
	// json:"-" so it is not treated as a jsonData field.
	AccountID int `json:"-"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (personalApiKey, accountId).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the upstream LoadSettings (pkg/models/settings.go:27-48): unmarshal jsonData
// (empty JSONData is a parse error, matching upstream), copy decrypted secrets
// by known key, and parse the account ID string into the numeric AccountID via
// strconv.Atoi (only set when parsing succeeds). The timeout default is applied
// in ApplyDefaults and the personalApiKey/AccountID contract is enforced in
// Validate, together reproducing what NewInstance does when it runs
// LoadSettings + CheckSettings (pkg/datasource/datasource.go:44-54).
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

	logger.Debug("loading newrelic datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings (pkg/models/settings.go:30-32) unmarshals
	// config.JSONData unconditionally and returns an error when the bytes are
	// empty or malformed. Mirror that behavior verbatim.
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		logger.Error("failed to parse jsonData", "err", err)
		return cfg, fmt.Errorf("parse jsonData: %w", err)
	}

	// Copy decrypted secrets by known key name (pkg/models/settings.go:34,36).
	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	// Parse the account ID string into the numeric AccountID, mirroring
	// LoadSettings (pkg/models/settings.go:42-46): only set it when Atoi
	// succeeds, leaving it zero (which Validate rejects) otherwise.
	if id, err := strconv.Atoi(cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccountID]); err == nil {
		cfg.AccountID = id
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("newrelic datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("newrelic datasource config loaded",
		"region", cfg.Region,
		"timeoutInSeconds", cfg.TimeoutInSeconds,
		"hasAccountID", cfg.AccountID != 0,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own code applies. Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// Curated default:
//   - TimeoutInSeconds: DefaultTimeoutInSeconds (300) when < 1, mirroring both
//     the editor default (src/components/ConfigEditor.tsx:57) and the backend
//     coercion (pkg/models/settings.go:38-40).
//
// Region has no default: the editor writes none (placeholder "default") and an
// empty region falls back to the New Relic client default (US) at
// pkg/datasource/newrelic_client.go:47-49.
func (c *Config) ApplyDefaults() {
	if c.TimeoutInSeconds < 1 {
		c.TimeoutInSeconds = DefaultTimeoutInSeconds
	}
}

// Validate checks the runtime contract the plugin enforces before it will
// create a datasource instance (pkg/datasource/handler_checkhealth.go:138-148,
// CheckSettings, invoked by NewInstance at pkg/datasource/datasource.go:50-54).
// Errors are joined so callers see every problem at once.
//
// Contracts enforced:
//   - personalApiKey (secureJsonData) must be non-empty after trimming
//     whitespace — upstream isEmpty check
//     (handler_checkhealth.go:99-101,139-141), "Enter a personal API key."
//   - AccountID must be non-zero — upstream (handler_checkhealth.go:143-145),
//     "Enter an account ID. This must be a valid, positive number." AccountID is
//     the value parsed from secureJsonData.accountId during LoadConfig; callers
//     assembling a Config directly must set it.
func (c Config) Validate() error {
	var errs []error

	if strings.TrimSpace(c.DecryptedSecureJSONData[SecureJsonDataKeyPersonalAPIKey]) == "" {
		errs = append(errs, errors.New("personal API key (secureJsonData.personalApiKey) is required"))
	}
	if c.AccountID == 0 {
		errs = append(errs, errors.New("account ID (secureJsonData.accountId) is required and must be a valid, non-zero number"))
	}

	return errors.Join(errs...)
}
