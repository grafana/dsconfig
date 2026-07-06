// Package gitlabdatasource contains the configuration models for the
// GitLab datasource plugin (plugin id: grafana-gitlab-datasource).
package gitlabdatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream repo).
const PluginID = "grafana-gitlab-datasource"

// DefaultURL is the GitLab API base URL applied when the root url is empty. It
// mirrors DefaultURL (src/types.ts:63) and baseGitLabURL
// (pkg/models/settings.go:23) — both the frontend editor
// (src/views/ConfigEditor.tsx:16,93) and the backend
// (pkg/models/settings.go:35-37) treat an empty URL as this value. go-gitlab
// appends "api/v4/" when the URL does not already end with it (go-gitlab
// gitlab.go:564-578), so a self-hosted URL may be given with or without the
// "/api/v4" suffix.
const DefaultURL = "https://gitlab.com/api/v4"

// DefaultPageLimit is the maximum number of API pages a query fetches when
// jsonData.pageLimit is 0. Mirrors basePageLimit (pkg/models/settings.go:25),
// applied at pkg/models/settings.go:39-41.
const DefaultPageLimit = 5

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyAccessToken is the GitLab personal (or project/group)
	// access token. It is copied from DecryptedSecureJSONData["accessToken"]
	// (pkg/models/settings.go:47) and sent by go-gitlab as the "PRIVATE-TOKEN"
	// request header (pkg/gitlab/datasource.go:176; go-gitlab gitlab.go:855-858),
	// NOT as an "Authorization: Bearer" header.
	SecureJsonDataKeyAccessToken SecureJsonDataKey = "accessToken"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads. The GitLab
// datasource declares exactly one secret.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyAccessToken,
}

// Config is the fully loaded configuration of a GitLab datasource instance.
//
// The json-tagged field (pageLimit) mirrors the jsonData portion the plugin's
// LoadSettings unmarshals from config.JSONData (pkg/models/settings.go:19,30).
//
// URL is a root-level datasource field read by the backend, carried with
// json:"-" so it never collides with jsonData. The plugin's LoadSettings
// unmarshals config.JSONData into a Settings whose URL is tagged json:"url"
// (pkg/models/settings.go:17,30) but then immediately overwrites it with
// config.URL (:34), so jsonData.url is dead and the effective URL is purely the
// datasource root url — we model it as a root field accordingly.
//
// The access token lives in secureJsonData and is modeled in
// DecryptedSecureJSONData rather than as a struct field. The upstream
// Settings.SdkClientOptions (httpclient.Options, pkg/models/settings.go:20) is
// runtime transport state, not configuration, and is not carried here.
// jsonData.enableSecureSocksProxy is intentionally omitted (AGENTS.md
// exclusion); json unmarshal silently ignores it on parse.
type Config struct {
	// Root-level field read by the backend (json:"-" — not jsonData).
	URL string `json:"-"`

	// jsonData field.
	PageLimit int `json:"pageLimit,omitempty"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (accessToken).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config, mirroring
// the plugin's LoadSettings (pkg/models/settings.go:28-63): unmarshal jsonData,
// take the URL from the datasource root (config.URL), copy the decrypted
// accessToken, default the URL and page limit, then validate that a token is
// present.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse -> ApplyDefaults -> Validate.
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on a Config they assemble themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading gitlab datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream LoadSettings (pkg/models/settings.go:30) unmarshals
	// config.JSONData unconditionally and returns an error when the bytes are
	// empty or malformed. Mirror that behavior verbatim — Grafana always sends
	// at least "{}", so this only rejects a truly-empty payload.
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		logger.Error("failed to parse jsonData", "err", err)
		return cfg, fmt.Errorf("parse jsonData: %w", err)
	}

	// The instance URL is a root datasource field: LoadSettings overwrites any
	// jsonData.url with config.URL (pkg/models/settings.go:34).
	cfg.URL = settings.URL

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[string(key)]; ok {
			cfg.DecryptedSecureJSONData[key] = val
		}
	}

	logger.Debug("loaded secure keys", "count", len(cfg.DecryptedSecureJSONData))

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("gitlab datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("gitlab datasource config loaded",
		"url", cfg.URL,
		"pageLimit", cfg.PageLimit,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own LoadSettings applies on every load
// (pkg/models/settings.go:35-41). Never blanket-apply every schema default —
// that would clobber intentional zero values.
//
// Curated defaults:
//   - URL: DefaultURL ("https://gitlab.com/api/v4") when empty — mirrors
//     src/types.ts:63 and pkg/models/settings.go:23,35-37.
//   - PageLimit: DefaultPageLimit (5) when 0 — mirrors basePageLimit and
//     pkg/models/settings.go:25,39-41.
//
// The accessToken secret has no default — the plugin errors out when it is empty.
func (c *Config) ApplyDefaults() {
	if c.URL == "" {
		c.URL = DefaultURL
	}
	if c.PageLimit == 0 {
		c.PageLimit = DefaultPageLimit
	}
}

// Validate checks the runtime contract the plugin requires. It mirrors the one
// hard requirement LoadSettings enforces (pkg/models/settings.go:48-50): a
// non-empty access token, returned upstream as ErrorEmptyAccessToken ("access
// token can not be blank"). Errors are joined so callers see every problem at
// once.
//
// The URL is intentionally NOT required here: the backend defaults an empty URL
// to https://gitlab.com/api/v4 (pkg/models/settings.go:35-37), so callers
// should invoke ApplyDefaults first (LoadConfig always does). The config editor
// nonetheless marks the URL field required (see README discrepancies).
func (c Config) Validate() error {
	var errs []error

	if c.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken] == "" {
		errs = append(errs, errors.New("access token (secureJsonData.accessToken) can not be blank"))
	}

	return errors.Join(errs...)
}
