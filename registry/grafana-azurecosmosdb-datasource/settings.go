// Package azurecosmosdbdatasource contains the configuration models for the
// Azure Cosmos DB datasource plugin (plugin id: grafana-azurecosmosdb-datasource).
package azurecosmosdbdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:6 in the upstream repo).
const PluginID = "grafana-azurecosmosdb-datasource"

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAccountKey is the Azure Cosmos DB account master
	// key (primary or secondary). Wrapped by azcosmos.NewKeyCredential
	// at pkg/cosmos/client.go:24 and used to sign every Cosmos DB REST
	// request.
	SecureJsonDataKeyAccountKey SecureJsonDataKey = "accountKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys used by the plugin. The Azure
// Cosmos DB datasource declares exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccountKey,
}

// Config is the fully loaded configuration of an Azure Cosmos DB
// datasource instance.
//
// The jsonData portion mirrors the plugin's own backend Settings
// (pkg/plugin/settings.go:10-13) verbatim for the fields that are actually
// read from jsonData — namely AccountEndpoint. The upstream Settings also
// declares AccountKey with a json:"accountKey,omitempty" tag, but the
// backend never unmarshals it from jsonData: pkg/plugin/settings.go:36-39
// populates AccountKey from DecryptedSecureJSONData["accountKey"]. Here
// the secret is stored exclusively in DecryptedSecureJSONData to keep the
// json-tagged struct in sync with the dsconfig schema's jsonData fields.
// See Upstream findings in the entry README.
//
// Root-level datasource fields (settings.URL, BasicAuth, User, etc.) are
// NOT carried on Config because the Cosmos DB plugin never reads them:
// pkg/plugin/settings.go:25-42 only touches settings.JSONData and
// settings.DecryptedSecureJSONData, and pkg/cosmos/client.go:23-52 builds
// its client solely from accountEndpoint + accountKey.
type Config struct {
	AccountEndpoint string `json:"accountEndpoint,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accountKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config,
// mirroring pkg/plugin/settings.go:25-42 (LoadSettings) verbatim:
// unmarshal jsonData into a lenient map, pluck out accountEndpoint, copy
// decrypted secrets by known key, then validate.
//
// Note on parse leniency: the upstream LoadSettings unmarshals into a
// map[string]any and picks the `accountEndpoint` value only if the type
// assertion to string succeeds — a non-string value is silently dropped
// rather than being a hard error. Here we mirror that with a direct
// json.Unmarshal into Config, which achieves the same effect for the
// happy path (matching string values) while surfacing malformed jsonData
// as a parse error, matching the upstream ErrorMessageInvalidJSON case.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext
// so log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults ->
// Validate. Callers that need each phase individually can invoke
// ApplyDefaults and Validate directly on a Config they assemble
// themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading azure cosmos db datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings (pkg/plugin/settings.go:26-29) calls
	// json.Unmarshal on settings.JSONData and wraps errors as
	// ErrorMessageInvalidJSON. We mirror that behavior — malformed
	// JSONData is a parse error. Empty JSONData is tolerated: upstream
	// unmarshals into an empty map and simply finds no accountEndpoint,
	// producing an empty AccountEndpoint that Validate then rejects.
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
		logger.Error("azure cosmos db datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("azure cosmos db datasource config loaded",
		"accountEndpoint", cfg.AccountEndpoint,
		"hasAccountKey", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccountKey] != "",
	)
	return cfg, nil
}

// ApplyDefaults is a no-op for the Azure Cosmos DB datasource: neither
// the config editor nor the backend applies any editor-parity defaults —
// both accountEndpoint and accountKey are strictly required and have no
// sensible zero value. Kept exported so callers can compose the three
// phases (parse -> ApplyDefaults -> Validate) uniformly across registry
// entries.
func (c *Config) ApplyDefaults() {
	// no defaults
}

// Validate checks the runtime contract that the plugin requires
// (pkg/plugin/settings.go:15-23, Settings.isValid). Errors are joined so
// callers see every problem at once.
//
// Contracts enforced:
//   - AccountEndpoint must be non-empty. Upstream returns
//     ErrorMessageEmptyAccountEndpoint ("account endpoint is empty").
//   - The accountKey secret (secureJsonData.accountKey) must be
//     non-empty. Upstream returns ErrorMessageEmptyAccountKey
//     ("account key is empty").
//
// Both errors are wrapped as backend.DownstreamError upstream; that
// classification is a plugin runtime concern, so this Validate returns
// plain errors.
func (c Config) Validate() error {
	var errs []error

	if c.AccountEndpoint == "" {
		errs = append(errs, errors.New("account endpoint (jsonData.accountEndpoint) is empty"))
	}
	if c.DecryptedSecureJSONData[SecureJsonDataKeyAccountKey] == "" {
		errs = append(errs, errors.New("account key (secureJsonData.accountKey) is empty"))
	}

	return errors.Join(errs...)
}
