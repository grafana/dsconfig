// Package salesforcedatasource contains the configuration models for the
// Salesforce datasource plugin (plugin id: grafana-salesforce-datasource).
package salesforcedatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream repo).
const PluginID = "grafana-salesforce-datasource"

// Salesforce OAuth host constants, mirroring the plugin's own
// pkg/models/settings.go:13-16 (TokenURLProd / TokenURLSandbox) and the
// frontend src/constants.ts:9-12 (Production / SandBox).
const (
	// TokenURLProd is the production Salesforce login/OAuth host.
	TokenURLProd = "https://login.salesforce.com"
	// TokenURLSandbox is the sandbox Salesforce login/OAuth host.
	TokenURLSandbox = "https://test.salesforce.com"
)

// AuthType is the authentication method selected in the configuration editor
// ("Authentication" radio). Stored in jsonData.authType. Mirrors the plugin's
// pkg/models/settings.go:18-23 AuthType alias and constants.
type AuthType string

const (
	// AuthTypeUser authenticates with the OAuth2 username-password grant
	// ("Credentials" in the editor).
	AuthTypeUser AuthType = "user"
	// AuthTypeJWT authenticates with the OAuth2 JWT bearer grant ("JWT" in the
	// editor).
	AuthTypeJWT AuthType = "jwt"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the Salesforce password (user auth).
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// SecureJsonDataKeySecurityToken is the Salesforce security token, appended
	// to the password before the token request (user auth; used but not
	// validated).
	SecureJsonDataKeySecurityToken SecureJsonDataKey = "securityToken"
	// SecureJsonDataKeyClientID is the connected app consumer key (user auth
	// client_id; JWT issuer).
	SecureJsonDataKeyClientID SecureJsonDataKey = "clientID"
	// SecureJsonDataKeyClientSecret is the connected app consumer secret (user
	// auth client_secret).
	SecureJsonDataKeyClientSecret SecureJsonDataKey = "clientSecret"
	// SecureJsonDataKeyCert is the connected app digital-signature certificate
	// (jwt auth).
	SecureJsonDataKeyCert SecureJsonDataKey = "cert"
	// SecureJsonDataKeyPrivateKey is the connected app RSA private key that
	// signs the JWT assertion (jwt auth).
	SecureJsonDataKeyPrivateKey SecureJsonDataKey = "privateKey"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the secret keys the plugin reads (GetSettings copies
// them from DecryptedSecureJSONData by auth method, pkg/models/settings.go:48-72).
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
	SecureJsonDataKeySecurityToken,
	SecureJsonDataKeyClientID,
	SecureJsonDataKeyClientSecret,
	SecureJsonDataKeyCert,
	SecureJsonDataKeyPrivateKey,
}

// Config is the fully loaded configuration of a Salesforce datasource instance.
//
// The json-tagged fields mirror the jsonData portion of the plugin's backend
// Settings struct (pkg/models/settings.go:25-39) verbatim: authType, user,
// sandbox, tokenUrl. The plugin's Settings struct additionally declares the six
// secrets (Password, SecurityToken, ClientID, ClientSecret, Cert, PrivateKey)
// with json tags, but those values are stored in secureJsonData and copied from
// DecryptedSecureJSONData in GetSettings — so they are modeled here in
// DecryptedSecureJSONData, not as jsonData struct fields (see the README
// discrepancies section).
//
// Root-level datasource fields (settings.URL, BasicAuth, etc.) are NOT carried
// on Config because the plugin never reads them: pkg/plugin/datasource.go:14-19
// builds the client from jsonData + decrypted secrets only.
//
// jsonData.enableSecureSocksProxy is intentionally omitted (AGENTS.md
// exclusion); json unmarshal silently ignores it on parse.
type Config struct {
	AuthType AuthType `json:"authType"`
	User     string   `json:"user"`
	Sandbox  bool     `json:"sandbox"`
	TokenURL string   `json:"tokenUrl"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (password, securityToken, clientID, clientSecret, cert, privateKey).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config, mirroring
// the plugin's GetSettings (pkg/models/settings.go:41-80): unmarshal jsonData
// (empty JSONData is a parse error, matching upstream json.Unmarshal), copy the
// decrypted secrets by known key, normalize the auth type and token URL, then
// validate.
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

	logger.Debug("loading salesforce datasource config")

	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}

	// Upstream GetSettings (pkg/models/settings.go:43-45) calls json.Unmarshal
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
		logger.Error("salesforce datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("salesforce datasource config loaded",
		"authType", cfg.AuthType,
		"hasUser", cfg.User != "",
		"tokenUrl", cfg.TokenURL,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own GetSettings applies on every load. Never
// blanket-apply every schema default — that would clobber intentional zero
// values.
//
// Curated defaults (mirroring pkg/models/settings.go:46-47,82-100):
//   - AuthType: normalizeAuthType defaults an empty authType to "user".
//     Note: the plugin's normalizeAuthType also has a jwt auto-detection branch
//     (cert && privateKey && !password), but it inspects the jsonData-level
//     struct fields, which are empty for secureJsonData-stored secrets and are
//     copied only after normalization runs — so an empty authType always
//     resolves to "user" in practice. This mirrors that effective behavior and
//     deliberately does not consult DecryptedSecureJSONData (see the README).
//   - TokenURL: normalizeTokenUrl defaults an empty tokenUrl to
//     TokenURLSandbox when Sandbox is true, otherwise TokenURLProd.
func (c *Config) ApplyDefaults() {
	if strings.TrimSpace(string(c.AuthType)) == "" {
		c.AuthType = AuthTypeUser
	}
	if c.TokenURL == "" {
		if c.Sandbox {
			c.TokenURL = TokenURLSandbox
		} else {
			c.TokenURL = TokenURLProd
		}
	}
}

// Validate checks the runtime contract the plugin enforces before fetching a
// token (pkg/plugin/client.go:77 calls Settings.Validate;
// pkg/models/settings.go:102-124). Errors are joined so callers see every
// problem at once (upstream returns only the first).
//
// Contracts enforced, keyed by auth method (with the secrets read from
// DecryptedSecureJSONData rather than jsonData struct fields):
//   - jwt: cert and privateKey must be non-empty ("invalid or empty
//     certificate" / "invalid or empty private key"). clientID and user are
//     also needed at connect time but upstream does not validate them.
//   - user (and any non-jwt authType, matching upstream's branch): user,
//     password, clientID and clientSecret must be non-empty ("invalid or empty
//     username" / "password" / "client id" / "client secret"). securityToken is
//     used but not validated.
func (c Config) Validate() error {
	var errs []error

	if c.AuthType == AuthTypeJWT {
		if c.DecryptedSecureJSONData[SecureJsonDataKeyCert] == "" {
			errs = append(errs, errors.New("invalid or empty certificate"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPrivateKey] == "" {
			errs = append(errs, errors.New("invalid or empty private key"))
		}
		return errors.Join(errs...)
	}

	// user auth (and, matching upstream, any non-jwt authType).
	if c.User == "" {
		errs = append(errs, errors.New("invalid or empty username"))
	}
	if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
		errs = append(errs, errors.New("invalid or empty password"))
	}
	if c.DecryptedSecureJSONData[SecureJsonDataKeyClientID] == "" {
		errs = append(errs, errors.New("invalid or empty client id"))
	}
	if c.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] == "" {
		errs = append(errs, errors.New("invalid or empty client secret"))
	}

	return errors.Join(errs...)
}
