// Package snowflakedatasource contains the configuration models for the
// Snowflake datasource plugin (grafana-snowflake-datasource).
package snowflakedatasource

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching plugin.json's id field
// (src/plugin.json:5 in the upstream plugin).
const PluginID = "grafana-snowflake-datasource"

// AuthType is the authentication method selected in the configuration editor
// ("Authentication Type"). Stored in jsonData.authType. Mirrors the backend
// constants in pkg/constants.go:12-16.
type AuthType string

const (
	// AuthTypeUnknown is an empty/missing authType; the backend treats it as
	// password (pkg/settings.go:76).
	AuthTypeUnknown AuthType = ""
	// AuthTypePassword authenticates with a Snowflake password.
	AuthTypePassword AuthType = "password"
	// AuthTypeKeyPair authenticates with an RSA key pair (JWT).
	AuthTypeKeyPair AuthType = "keypair"
	// AuthTypeOauth authenticates by forwarding the user's OAuth identity.
	AuthTypeOauth AuthType = "oauth"
	// AuthTypePAT authenticates with a Snowflake programmatic access token.
	AuthTypePAT AuthType = "pat"
)

// Error values mirroring the plugin's own error messages (pkg/constants.go:23-42)
// so LoadConfig fails the same way as the upstream LoadSettings.
var (
	// ErrInvalidPassword is returned when password auth has no password.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrInvalidPrivateKey is returned when key-pair auth has no/invalid private key.
	ErrInvalidPrivateKey = errors.New("invalid private key")
	// ErrInvalidPAT is returned when PAT auth has no token.
	ErrInvalidPAT = errors.New("invalid programmatic access token")
	// ErrInvalidOAuth is returned when oauth auth does not enable oauthPassThru.
	ErrInvalidOAuth = errors.New("you must enable Forward OAuth Identity")
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	// SecureJsonDataKeyPassword is the Snowflake password (password auth).
	SecureJsonDataKeyPassword SecureJsonDataKey = "password"
	// SecureJsonDataKeyPrivateKey is the RSA private key PEM (key-pair auth).
	SecureJsonDataKeyPrivateKey SecureJsonDataKey = "privateKey"
	// SecureJsonDataKeyPrivateKeyPassphrase is the passphrase for an encrypted
	// private key (key-pair auth).
	SecureJsonDataKeyPrivateKeyPassphrase SecureJsonDataKey = "privateKeyPassphrase"
	// SecureJsonDataKeyPATToken is the programmatic access token (PAT auth).
	SecureJsonDataKeyPATToken SecureJsonDataKey = "patToken"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the fixed secret keys the plugin reads. Session
// parameters marked secure add further dynamic keys named after the setting;
// those are user-defined and not enumerated here.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyPassword,
	SecureJsonDataKeyPrivateKey,
	SecureJsonDataKeyPrivateKeyPassphrase,
	SecureJsonDataKeyPATToken,
}

// Setting is one session parameter element of jsonData.settings. Mirrors the
// upstream Setting (pkg/settings.go:17-21) and the frontend Setting
// (src/types.ts:41-45). The upstream backend struct leaves Secure without a
// json tag (relying on case-insensitive unmarshal of the frontend's "secure");
// this model uses json:"secure" to round-trip the stored shape faithfully.
type Setting struct {
	Name   string `json:"name"`
	Value  string `json:"value,omitempty"`
	Secure bool   `json:"secure,omitempty"`
}

// Config is the fully loaded configuration of a Snowflake datasource instance.
// The plugin authenticates entirely from jsonData + secureJsonData
// (pkg/settings.go:56-192) and never reads root-level datasource fields, so
// only the parsed jsonData fields and decrypted secure data live here. Callers
// reach everything directly as cfg.Account, cfg.AuthType, etc., and enumerate
// configured secrets by iterating DecryptedSecureJSONData.
//
// The jsonData fields mirror the upstream Settings (pkg/settings.go:24-45), json
// tags verbatim, plus the four frontend-only fields (defaultQuery,
// defaultVariableQuery, defaultInterpolation, timeInterval) that live in
// jsonData but are read only by the frontend (src/datasource.ts:31-36).
type Config struct {
	Account              string    `json:"account"`
	Username             string    `json:"username"`
	Region               string    `json:"region"`
	Role                 string    `json:"role"`
	AuthType             AuthType  `json:"authType"`
	Warehouse            string    `json:"warehouse"`
	Database             string    `json:"database"`
	Schema               string    `json:"schema"`
	DefaultQuery         string    `json:"defaultQuery"`
	DefaultVariableQuery string    `json:"defaultVariableQuery"`
	DefaultInterpolation string    `json:"defaultInterpolation"`
	TimeInterval         string    `json:"timeInterval"`
	LoginTimeout         int64     `json:"loginTimeout"`
	RequestTimeout       int64     `json:"requestTimeout"`
	Settings             []Setting `json:"settings"`
	OAuthPassThrough     bool      `json:"oauthPassThru"`
	RowLimit             int64     `json:"rowLimit"`

	// DecryptedSecureJSONData holds the decrypted secure values by key
	// (password, privateKey, privateKeyPassphrase, patToken).
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/settings.go:56-192): the auth type defaults to
// password, a missing/empty authType is treated as password, and each auth
// method requires its own credential.
//
// ctx is used to derive a contextual logger via backend.Logger.FromContext so
// log lines carry the request/plugin context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse (unmarshal jsonData + copy
// decrypted secrets), then (*Config).ApplyDefaults for curated editor-parity
// defaults, then (Config).Validate to enforce the plugin's runtime contract.
// Callers that need each phase individually can invoke ApplyDefaults and
// Validate directly on the returned Config.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading snowflake datasource config")

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
		logger.Error("snowflake datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("snowflake datasource config loaded",
		"authType", cfg.AuthType,
		"hasAccount", cfg.Account != "",
		"hasUsername", cfg.Username != "",
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same defaults
// the editor and backend apply for a fresh datasource. It is intentionally
// small: only the auth discriminator is defaulted.
//
// Curated list (only these fields are touched, and only when zero-valued):
//   - AuthType → AuthTypePassword (mirrors pkg/settings.go:60-61 and
//     src/editors/ConfigEditor.tsx:39; an empty authType is treated as password).
func (c *Config) ApplyDefaults() {
	if c.AuthType == AuthTypeUnknown {
		c.AuthType = AuthTypePassword
	}
}

// Validate checks that a loaded Config satisfies the plugin's runtime contract,
// mirroring LoadSettings (pkg/settings.go:76-151): the selected auth method has
// its required credential. It does not require account/username — the upstream
// LoadSettings does not either (those surface as connection errors later). It
// does not mutate the Config. Errors are joined so callers see every problem at
// once.
func (c Config) Validate() error {
	var errs []error

	switch c.AuthType {
	case AuthTypePassword, AuthTypeUnknown:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPassword] == "" {
			errs = append(errs, ErrInvalidPassword)
		}
	case AuthTypeKeyPair:
		key := c.DecryptedSecureJSONData[SecureJsonDataKeyPrivateKey]
		if key == "" || !isValidPrivateKeyPEM(key) {
			errs = append(errs, ErrInvalidPrivateKey)
		}
	case AuthTypePAT:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyPATToken] == "" {
			errs = append(errs, ErrInvalidPAT)
		}
	case AuthTypeOauth:
		if !c.OAuthPassThrough {
			errs = append(errs, ErrInvalidOAuth)
		}
	default:
		errs = append(errs, fmt.Errorf("unknown authType %q", c.AuthType))
	}

	return errors.Join(errs...)
}

// isValidPrivateKeyPEM reports whether s is a PKCS#8 PEM private key block,
// mirroring the upstream check at pkg/settings.go:96-97 (a valid block whose
// type is "PRIVATE KEY" or "ENCRYPTED PRIVATE KEY"). It does not attempt to
// parse or decrypt the key (that requires a passphrase and extra dependencies);
// it validates the PEM envelope only.
func isValidPrivateKeyPEM(s string) bool {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return false
	}
	return block.Type == "PRIVATE KEY" || block.Type == "ENCRYPTED PRIVATE KEY"
}
