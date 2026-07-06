// Package splunkmonitoringdatasource contains the configuration models for the
// Splunk Infrastructure Monitoring (SignalFx) datasource plugin
// (id: grafana-splunk-monitoring-datasource).
package splunkmonitoringdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream plugin).
const PluginID = "grafana-splunk-monitoring-datasource"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAccessToken is the Splunk Observability API access
	// token, sent as the "X-SF-TOKEN" header on every outgoing request
	// (pkg/client/rest.go:225). Required (pkg/models/settings.go:27-30).
	SecureJsonDataKeyAccessToken SecureJsonDataKey = "accessToken"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin. The Splunk
// Infrastructure Monitoring datasource declares exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessToken,
}

// Config is the fully loaded configuration of a Splunk Infrastructure
// Monitoring datasource instance.
//
// The jsonData fields mirror the plugin's own backend Settings struct
// (pkg/models/settings.go:13-19) verbatim — same json tags (realmName,
// url_metrics_metadata, url_signalflow) — except that:
//   - the upstream AccessToken field (populated from the decrypted secret) is
//     replaced here by DecryptedSecureJSONData; and
//   - the upstream HttpClientOptions field (json:"-", SDK proxy plumbing loaded
//     via s.HTTPClientOptions) is not carried, matching the API-token cousin
//     entries (github, sentry, datadog).
//
// Root-level datasource fields (settings.URL, BasicAuth, etc.) are NOT carried
// because the plugin never reads them: the backend derives its endpoints from
// the realm + custom URL overrides (pkg/client/rest.go:339-353) and its client
// from the decrypted accessToken (pkg/client/client.go:61-76).
type Config struct {
	Realm              string `json:"realmName,omitempty"`
	URLMetricsMetaData string `json:"url_metrics_metadata,omitempty"`
	URLSignalFlow      string `json:"url_signalflow,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessToken).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`

	// Note: jsonData.enableSecureSocksProxy is written by the editor's Secure
	// Socks Proxy switch (src/components/ConfigEditor.tsx:132-137) and consumed
	// transparently by the SDK proxy plumbing via settings.HttpClientOptions
	// (pkg/client/rest.go:307-314). The plugin's own backend never inspects it
	// by name and the upstream Settings struct (pkg/models/settings.go:13-19)
	// does not carry it. Following AGENTS.md and upstream, it is not modeled on
	// this Config; json unmarshal silently ignores it on parse.
}

// LoadConfig parses a datasource instance's settings into a Config, mirroring
// the plugin's LoadSettings (pkg/models/settings.go:21-42) for the parse phase:
// json.Unmarshal the jsonData (empty or malformed JSONData is a parse error,
// exactly as upstream), then copy the decrypted secrets by known key.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate, returning a fully-defaulted, validated Config. Because this is the
// intended shape for the plugin's own upstream LoadSettings to sync to, any
// parse divergence from upstream should be treated as a bug in this entry.
// Callers that assemble a Config themselves can invoke ApplyDefaults and
// Validate individually.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading splunk infrastructure monitoring datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings (pkg/models/settings.go:22-26) calls json.Unmarshal
	// on settings.JSONData unconditionally and returns the error when the bytes
	// are empty or malformed. We mirror that behavior — empty JSONData is a
	// parse error.
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		logger.Error("failed to parse jsonData", "err", err)
		return cfg, fmt.Errorf("parse jsonData: %w", err)
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("splunk infrastructure monitoring datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("splunk infrastructure monitoring datasource config loaded",
		"realm", cfg.Realm,
		"hasMetricsMetadataURL", cfg.URLMetricsMetaData != "",
		"hasSignalflowURL", cfg.URLSignalFlow != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own code applies. Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// The Splunk Infrastructure Monitoring plugin has no editor-parity defaults:
// the config editor writes no default values (the "us1" realm placeholder is
// only a placeholder, not a default), and the backend supplies no defaults for
// the realm or the custom URLs — an empty realm with no custom URLs simply
// yields the broken host https://api..signalfx.com. ApplyDefaults is therefore
// intentionally a no-op, kept exported for API parity with the other registry
// entries and for callers that assemble a Config directly.
func (c *Config) ApplyDefaults() {}

// Validate checks the runtime contract the plugin enforces. Errors are joined
// so callers see every problem at once.
//
// Contracts enforced:
//   - The access token (secureJsonData.accessToken) must be non-empty. Upstream
//     LoadSettings returns "invalid/empty access token"
//     (pkg/models/settings.go:27-30) and NewSignalFxClient returns "required
//     access token is missing" (pkg/client/client.go:62-63).
//   - The realm (jsonData.realmName) must be non-empty UNLESS both custom URLs
//     (url_metrics_metadata and url_signalflow) are set. This mirrors the
//     effective health/runtime contract: GetBaseURL (pkg/client/rest.go:339-353)
//     derives the metrics-metadata and SignalFlow base URLs from the realm, and
//     an empty realm without both overrides produces the broken host
//     https://api..signalfx.com, so CheckHealth (pkg/client/rest.go:82-88) and
//     queries fail. Upstream LoadSettings does not validate the realm; like the
//     datadog entry's Validate (which mirrors the health check rather than
//     LoadSettings), this encodes the contract a working datasource requires.
func (c Config) Validate() error {
	var errs []error

	if c.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken] == "" {
		errs = append(errs, errors.New("access token (secureJsonData.accessToken) is required"))
	}

	if strings.TrimSpace(c.Realm) == "" &&
		(strings.TrimSpace(c.URLMetricsMetaData) == "" || strings.TrimSpace(c.URLSignalFlow) == "") {
		errs = append(errs, errors.New("realm (jsonData.realmName) is required unless both custom URLs (jsonData.url_metrics_metadata and jsonData.url_signalflow) are set"))
	}

	return errors.Join(errs...)
}
