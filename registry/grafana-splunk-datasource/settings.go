// Package splunkdatasource contains the configuration models for the Splunk
// datasource plugin (id: grafana-splunk-datasource).
package splunkdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json's id field.
const PluginID = "grafana-splunk-datasource"

// Backend defaulting constants mirrored from the plugin's LoadSettings
// (pkg/models/settings.go) and query_limits.go.
const (
	// DefaultInternalFieldPattern is applied when internal field filtration is
	// enabled but no pattern is provided (pkg/models/settings.go:108-110).
	DefaultInternalFieldPattern = "^_.+"
	// DefaultTimeField is applied when tsField is empty (pkg/models/settings.go:112-114).
	DefaultTimeField = "_time"
	// DefaultTimeoutInSeconds is applied when timeoutInSeconds is below 1
	// (pkg/models/settings.go:121-123).
	DefaultTimeoutInSeconds int64 = 30
	// DefaultResultLimit is the safety limit the backend resolves a 0 (unlimited)
	// maxResultCount to (pkg/models/query_limits.go:10,48-51). It is a runtime
	// resolution only and is not written back into jsonData.
	DefaultResultLimit = 10000
)

// AuthType is the authentication method selected in the config editor's Auth
// component, stored verbatim in jsonData.authType. An empty value is treated as
// AuthTypeBasicAuth by the backend (pkg/models/settings.go:95).
type AuthType string

const (
	// AuthTypeBasicAuth authenticates with root.basicAuthUser +
	// secureJsonData.basicAuthPassword (the default method).
	AuthTypeBasicAuth AuthType = "BasicAuth"
	// AuthTypeAlternativeToken authenticates with a Splunk auth token in
	// secureJsonData.authToken, sent as "Authorization: Bearer". Labelled
	// "Alternative authentication" in the editor.
	AuthTypeAlternativeToken AuthType = "custom-splunk"
	// AuthTypeOAuthForward forwards the signed-in user's OAuth identity
	// (jsonData.oauthPassThru=true). Only offered when the
	// splunkEnableOAuthForwarding feature toggle is on.
	AuthTypeOAuthForward AuthType = "OAuthForward"
)

// FieldSearchType is the fields search mode (jsonData.fieldSearchType).
type FieldSearchType string

const (
	// FieldSearchTypeQuick is the "quick" fields search mode (editor default).
	FieldSearchTypeQuick FieldSearchType = "quick"
	// FieldSearchTypeFull is the "full" fields search mode.
	FieldSearchTypeFull FieldSearchType = "full"
)

// VariableSearchLevel is the variables search mode (jsonData.variableSearchLevel).
type VariableSearchLevel string

const (
	// VariableSearchLevelFast is the "fast" variables search mode (editor default).
	VariableSearchLevelFast VariableSearchLevel = "fast"
	// VariableSearchLevelSmart is the "smart" variables search mode.
	VariableSearchLevelSmart VariableSearchLevel = "smart"
	// VariableSearchLevelVerbose is the "verbose" variables search mode.
	VariableSearchLevelVerbose VariableSearchLevel = "verbose"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyBasicAuthPassword is the Basic-auth password, set when the
	// BasicAuth method is selected.
	SecureJsonDataKeyBasicAuthPassword SecureJsonDataKey = "basicAuthPassword"
	// SecureJsonDataKeyAuthToken is the Splunk auth token, set when the
	// custom-splunk ("Alternative authentication") method is selected.
	SecureJsonDataKeyAuthToken SecureJsonDataKey = "authToken"
	// SecureJsonDataKeyTLSCACert is the custom CA certificate, set when
	// jsonData.tlsAuthWithCACert is true.
	SecureJsonDataKeyTLSCACert SecureJsonDataKey = "tlsCACert"
	// SecureJsonDataKeyTLSClientCert is the TLS client certificate, set when
	// jsonData.tlsAuth is true.
	SecureJsonDataKeyTLSClientCert SecureJsonDataKey = "tlsClientCert"
	// SecureJsonDataKeyTLSClientKey is the TLS client key, set when
	// jsonData.tlsAuth is true.
	SecureJsonDataKeyTLSClientKey SecureJsonDataKey = "tlsClientKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys modeled by this entry.
//
// Note: the backend additionally reads a legacy secureJsonData.APIKey
// (pkg/models/settings.go:91) that no code path consumes and the editor never
// writes; it is intentionally excluded here (see the entry README).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyAuthToken,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
}

// DataLinkConfig mirrors the plugin's DataLinkConfig (pkg/models/settings.go:14-23)
// stored in jsonData.dataLinks. Only the JSON-serialized fields are carried here
// (the compiled *regexp.Regexp fields are runtime-only).
type DataLinkConfig struct {
	Field         string `json:"field"`
	Label         string `json:"label"`
	MatcherRegex  string `json:"matcherRegex"`
	URL           string `json:"url"`
	DatasourceUID string `json:"datasourceUid,omitempty"`
}

// Config is the fully loaded configuration of a Splunk datasource instance.
//
// The plugin's upstream Settings struct (pkg/models/settings.go:26-48) is a
// subset of what the config editor writes. This Config carries every jsonData
// field the editor persists (including frontend-only ones) so provisioning
// callers can round-trip the full editor state.
//
// Root-level fields (URL, BasicAuthUser) are tagged json:"-" and populated by
// LoadConfig from backend.DataSourceInstanceSettings. URL is read directly by
// the plugin backend (pkg/models/settings.go:90); BasicAuthUser is consumed by
// the SDK HTTP client for Basic auth.
type Config struct {
	// Root-level fields (not stored in jsonData).
	URL           string `json:"-"`
	BasicAuthUser string `json:"-"`

	// jsonData fields — the union of what the editor writes and the backend reads.
	AuthType      AuthType `json:"authType,omitempty"`
	OAuthPassThru bool     `json:"oauthPassThru,omitempty"`

	// TLS (rendered by @grafana/plugin-ui's Auth component; consumed by the SDK).
	TLSAuthWithCACert bool   `json:"tlsAuthWithCACert,omitempty"`
	TLSAuth           bool   `json:"tlsAuth,omitempty"`
	ServerName        string `json:"serverName,omitempty"`
	TLSSkipVerify     bool   `json:"tlsSkipVerify,omitempty"`

	// Advanced HTTP settings (@grafana/plugin-ui AdvancedHttpSettings).
	KeepCookies []string `json:"keepCookies,omitempty"`
	Timeout     float64  `json:"timeout,omitempty"`

	// Advanced options.
	MaxResultCount int  `json:"maxResultCount,omitempty"`
	PreviewMode    bool `json:"previewMode,omitempty"`
	// AsyncMode is stored under json:"pollSearchResult" upstream.
	PollSearchResult bool `json:"pollSearchResult,omitempty"`
	// MinPollInterval/MaxPollInterval are declared number on the frontend type but
	// persisted as strings by the editor; stored as strings for round-trip safety.
	MinPollInterval string `json:"minPollInterval,omitempty"`
	MaxPollInterval string `json:"maxPollInterval,omitempty"`
	AutoCancel      string `json:"autoCancel,omitempty"`
	// TimeoutInSeconds is the plugin-specific request timeout (distinct from Timeout).
	TimeoutInSeconds         int64               `json:"timeoutInSeconds,omitempty"`
	StatusBuckets            string              `json:"statusBuckets,omitempty"`
	InternalFieldsFiltration bool                `json:"internalFieldsFiltration,omitempty"`
	InternalFieldPattern     string              `json:"internalFieldPattern,omitempty"`
	TimeField                string              `json:"tsField,omitempty"`
	FieldSearchType          FieldSearchType     `json:"fieldSearchType,omitempty"`
	VariableSearchLevel      VariableSearchLevel `json:"variableSearchLevel,omitempty"`
	DefaultEarliestTime      string              `json:"defaultEarliestTime,omitempty"`
	// StreamMode is a legacy flag migrated into PreviewMode.
	StreamMode bool             `json:"streamMode,omitempty"`
	DataLinks  []DataLinkConfig `json:"dataLinks,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (basicAuthPassword, authToken, tlsCACert, tlsClientCert, tlsClientKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. Root fields
// (URL, BasicAuthUser) are copied from backend.DataSourceInstanceSettings;
// jsonData is unmarshaled from settings.JSONData; decrypted secrets are copied by
// known key name into DecryptedSecureJSONData.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults -> Validate.
// The parse phase mirrors the plugin's LoadSettings (pkg/models/settings.go):
// streamMode is migrated to previewMode, internalFieldPattern defaults to
// "^_.+" (then cleared when filtration is off), tsField defaults to "_time",
// and timeoutInSeconds below 1 is bumped to 30. Callers that need each phase
// individually can invoke ApplyDefaults and Validate on a Config they assemble.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading splunk datasource config")

	cfg := Config{
		URL:                     settings.URL,
		BasicAuthUser:           settings.BasicAuthUser,
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
			logger.Error("failed to parse jsonData", "err", err)
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	// Mirror the backend LoadSettings parse-time defaulting
	// (pkg/models/settings.go:102-123), preserving upstream order.
	if cfg.StreamMode {
		cfg.PreviewMode = true
	}
	if cfg.InternalFieldPattern == "" {
		cfg.InternalFieldPattern = DefaultInternalFieldPattern
	}
	if cfg.TimeField == "" {
		cfg.TimeField = DefaultTimeField
	}
	if !cfg.InternalFieldsFiltration {
		cfg.InternalFieldPattern = ""
	}
	if cfg.TimeoutInSeconds < 1 {
		cfg.TimeoutInSeconds = DefaultTimeoutInSeconds
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("splunk datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("splunk datasource config loaded",
		"hasURL", cfg.URL != "",
		"authType", cfg.AuthType,
		"tlsAuth", cfg.TLSAuth,
		"tlsAuthWithCACert", cfg.TLSAuthWithCACert,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued discriminator fields with the
// same effective defaults the config editor uses for a fresh datasource. It is
// intentionally minimal — never blanket-apply every schema default, which would
// clobber intentional zero values.
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - AuthType            -> AuthTypeBasicAuth      (editor default; empty means BasicAuth)
//   - FieldSearchType     -> FieldSearchTypeQuick   (editor Select default 'quick')
//   - VariableSearchLevel -> VariableSearchLevelFast (editor Select default 'fast')
func (c *Config) ApplyDefaults() {
	if c.AuthType == "" {
		c.AuthType = AuthTypeBasicAuth
	}
	if c.FieldSearchType == "" {
		c.FieldSearchType = FieldSearchTypeQuick
	}
	if c.VariableSearchLevel == "" {
		c.VariableSearchLevel = VariableSearchLevelFast
	}
}

// Validate checks the runtime contract the plugin requires. The backend needs a
// URL to build its Splunk REST endpoints (pkg/splunk/client.go:67-68), and each
// authentication method has its own required inputs. TLS is independent of the
// auth method: mutual TLS needs a client cert + key, and custom-CA verification
// needs the CA cert. Errors are joined so callers see every problem at once.
func (c Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("URL (root.url) is required"))
	}

	switch c.AuthType {
	case AuthTypeBasicAuth, "":
		// Empty authType is treated as BasicAuth by the backend
		// (pkg/models/settings.go:95).
		if c.BasicAuthUser == "" {
			errs = append(errs, errors.New("basicAuthUser (root) is required for BasicAuth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword] == "" {
			errs = append(errs, errors.New("basicAuthPassword (secureJsonData) is required for BasicAuth"))
		}
	case AuthTypeAlternativeToken:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAuthToken] == "" {
			errs = append(errs, errors.New("authToken (secureJsonData) is required for Alternative authentication"))
		}
	case AuthTypeOAuthForward:
		// No secret required; Grafana core forwards the user's OAuth token.
	default:
		errs = append(errs, fmt.Errorf("unknown authType %q", c.AuthType))
	}

	if c.TLSAuth {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert] == "" {
			errs = append(errs, errors.New("tlsClientCert (secureJsonData) is required when tlsAuth is true"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey] == "" {
			errs = append(errs, errors.New("tlsClientKey (secureJsonData) is required when tlsAuth is true"))
		}
	}

	if c.TLSAuthWithCACert {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert] == "" {
			errs = append(errs, errors.New("tlsCACert (secureJsonData) is required when tlsAuthWithCACert is true"))
		}
	}

	if c.TimeoutInSeconds < 0 {
		errs = append(errs, fmt.Errorf("timeoutInSeconds must be non-negative, got %d", c.TimeoutInSeconds))
	}

	return errors.Join(errs...)
}
