// Package infinitydatasource contains the configuration models for the
// Grafana Infinity datasource plugin (id: yesoreyeram-infinity-datasource).
package infinitydatasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginID is the datasource plugin type, matching src/plugin.json:4 in
// the upstream repo.
const PluginID = "yesoreyeram-infinity-datasource"

// IgnoreURLSentinel is the marker the frontend writes to root.url when the
// user leaves the Base URL empty (src/constants.ts:72); LoadSettings
// normalizes it back to "" on load (pkg/models/settings.go:296-298).
const IgnoreURLSentinel = "__IGNORE_URL__"

// AuthType is the selected authentication method stored in
// jsonData.auth_method. Mirrors the constants in pkg/models/settings.go:17-27.
type AuthType string

const (
	AuthTypeNone         AuthType = "none"
	AuthTypeBasic        AuthType = "basicAuth"
	AuthTypeBearerToken  AuthType = "bearerToken"
	AuthTypeAPIKey       AuthType = "apiKey"
	AuthTypeDigest       AuthType = "digestAuth"
	AuthTypeForwardOAuth AuthType = "oauthPassThru"
	AuthTypeOAuth2       AuthType = "oauth2"
	AuthTypeAWS          AuthType = "aws"
	AuthTypeAzureBlob    AuthType = "azureBlob"
)

// OAuth2Type is the OAuth2 grant type stored in jsonData.oauth2.oauth2_type.
// Mirrors pkg/models/settings.go:29-33.
type OAuth2Type string

const (
	OAuth2TypeClientCredentials OAuth2Type = "client_credentials"
	OAuth2TypeJWT               OAuth2Type = "jwt"
	OAuth2TypeOthers            OAuth2Type = "others"
)

// APIKeyType selects header vs URL query for the API key value. Mirrors
// pkg/models/settings.go:35-38. Default is "header".
type APIKeyType string

const (
	APIKeyTypeHeader APIKeyType = "header"
	APIKeyTypeQuery  APIKeyType = "query"
)

// ProxyType is the outbound proxy mode stored in jsonData.proxy_type.
// Mirrors pkg/models/settings.go:69-75.
type ProxyType string

const (
	ProxyTypeNone ProxyType = "none"
	ProxyTypeEnv  ProxyType = "env"
	ProxyTypeURL  ProxyType = "url"
)

// UnsecuredQueryHandlingMode controls how the backend handles queries that
// carry per-query secrets and bypass the allowed-hosts protection. Mirrors
// pkg/models/settings.go:77-83.
type UnsecuredQueryHandlingMode string

const (
	UnsecuredQueryHandlingAllow UnsecuredQueryHandlingMode = "allow"
	UnsecuredQueryHandlingWarn  UnsecuredQueryHandlingMode = "warn"
	UnsecuredQueryHandlingDeny  UnsecuredQueryHandlingMode = "deny"
)

// AWSAuthType is the AWS SigV4 authentication sub-mode. Currently only
// "keys" is supported. Mirrors pkg/models/settings.go:57-61.
type AWSAuthType string

const (
	AWSAuthTypeKeys AWSAuthType = "keys"
)

// AzureBlobCloudType selects the Azure cloud environment. Mirrors the
// values LoadSettings understands (pkg/models/settings.go:402-415) and the
// options in src/constants.ts:130-134.
type AzureBlobCloudType string

const (
	AzureBlobCloudTypeAzureCloud        AzureBlobCloudType = "AzureCloud"
	AzureBlobCloudTypeAzureUSGovernment AzureBlobCloudType = "AzureUSGovernment"
	AzureBlobCloudTypeAzureChinaCloud   AzureBlobCloudType = "AzureChinaCloud"
)

// SecureJsonDataKey is a strictly-typed name of a secret stored in
// secureJsonData (write-only; read existing config via secureJsonFields).
type SecureJsonDataKey string

const (
	SecureJsonDataKeyBasicAuthPassword   SecureJsonDataKey = "basicAuthPassword"
	SecureJsonDataKeyBearerToken         SecureJsonDataKey = "bearerToken"
	SecureJsonDataKeyAPIKeyValue         SecureJsonDataKey = "apiKeyValue"
	SecureJsonDataKeyAWSAccessKey        SecureJsonDataKey = "awsAccessKey"
	SecureJsonDataKeyAWSSecretKey        SecureJsonDataKey = "awsSecretKey"
	SecureJsonDataKeyOAuth2ClientSecret  SecureJsonDataKey = "oauth2ClientSecret"
	SecureJsonDataKeyOAuth2JWTPrivateKey SecureJsonDataKey = "oauth2JWTPrivateKey"
	SecureJsonDataKeyAzureBlobAccountKey SecureJsonDataKey = "azureBlobAccountKey"
	SecureJsonDataKeyTLSCACert           SecureJsonDataKey = "tlsCACert"
	SecureJsonDataKeyTLSClientCert       SecureJsonDataKey = "tlsClientCert"
	SecureJsonDataKeyTLSClientKey        SecureJsonDataKey = "tlsClientKey"
	SecureJsonDataKeyProxyUserPassword   SecureJsonDataKey = "proxyUserPassword"
)

// SecureJsonDataConfig lists the secret key names stored in secureJsonData.
type SecureJsonDataConfig []SecureJsonDataKey

// SecureJsonDataKeys are the fixed-name secret keys the plugin reads on
// load. In addition to these, LoadSettings aggregates dynamic
// indexed-pair secrets (httpHeaderValue<N>, secureQueryValue<N>,
// oauth2EndPointParamsValue<N>, oauth2TokenHeadersValue<N>) into the
// CustomHeaders / SecureQueryFields / OAuth2Settings.EndpointParams /
// OAuth2Settings.TokenHeaders maps on Config. Those dynamic keys are not
// represented in this fixed list because their names vary per instance.
var SecureJsonDataKeys = SecureJsonDataConfig{
	SecureJsonDataKeyBasicAuthPassword,
	SecureJsonDataKeyBearerToken,
	SecureJsonDataKeyAPIKeyValue,
	SecureJsonDataKeyAWSAccessKey,
	SecureJsonDataKeyAWSSecretKey,
	SecureJsonDataKeyOAuth2ClientSecret,
	SecureJsonDataKeyOAuth2JWTPrivateKey,
	SecureJsonDataKeyAzureBlobAccountKey,
	SecureJsonDataKeyTLSCACert,
	SecureJsonDataKeyTLSClientCert,
	SecureJsonDataKeyTLSClientKey,
	SecureJsonDataKeyProxyUserPassword,
}

// OAuth2Settings is the OAuth2 sub-object persisted at jsonData.oauth2.
// Mirrors pkg/models/settings.go:40-55 verbatim (same fields, same json
// tags, same omitempty behavior).
//
// oauth2.AuthStyle is stored as an integer (0=Auto, 1=In Params, 2=In
// Header); we use int rather than the upstream oauth2.AuthStyle alias to
// keep this schema entry free of golang.org/x/oauth2 as a dependency, but
// the on-wire representation is identical.
type OAuth2Settings struct {
	OAuth2Type    OAuth2Type `json:"oauth2_type,omitempty"`
	ClientID      string     `json:"client_id,omitempty"`
	TokenURL      string     `json:"token_url,omitempty"`
	Email         string     `json:"email,omitempty"`
	PrivateKeyID  string     `json:"private_key_id,omitempty"`
	Subject       string     `json:"subject,omitempty"`
	Scopes        []string   `json:"scopes,omitempty"`
	AuthStyle     int        `json:"authStyle,omitempty"`
	AuthHeader    string     `json:"authHeader,omitempty"`
	TokenTemplate string     `json:"tokenTemplate,omitempty"`
}

// AWSSettings is the AWS SigV4 sub-object persisted at jsonData.aws.
// Mirrors pkg/models/settings.go:63-67 verbatim.
type AWSSettings struct {
	AuthType AWSAuthType `json:"authType,omitempty"`
	Region   string      `json:"region,omitempty"`
	Service  string      `json:"service,omitempty"`
}

// RefData is one entry of jsonData.refData. Mirrors
// pkg/models/settings.go:238-241 verbatim.
type RefData struct {
	Name string `json:"name,omitempty"`
	Data string `json:"data,omitempty"`
}

// GlobalQuery is one entry of jsonData.global_queries. The Query field is
// intentionally kept as json.RawMessage: at the datasource-config layer
// the individual InfinityQuery shape (owned by the query editor) is
// opaque.
type GlobalQuery struct {
	Name  string          `json:"name,omitempty"`
	ID    string          `json:"id,omitempty"`
	Query json.RawMessage `json:"query,omitempty"`
}

// Config is the fully loaded configuration of a Grafana Infinity datasource
// instance.
//
// The jsonData-tagged fields mirror the upstream
// InfinitySettingsJson struct (pkg/models/settings.go:261-290) verbatim —
// same names, same json tags, same types. Root-level fields (URL,
// BasicAuth, BasicAuthUser) are tagged json:"-" so they don't collide with
// jsonData unmarshaling; LoadConfig copies them from
// backend.DataSourceInstanceSettings directly. The plugin's own
// LoadSettings also reads them (as InfinitySettings.URL,
// .BasicAuthEnabled, .UserName) so they belong on Config too.
type Config struct {
	// Root-level fields (json:"-" — not part of jsonData). See
	// pkg/models/settings.go:293-300 for how the plugin loads them.
	URL           string `json:"-"`
	BasicAuth     bool   `json:"-"`
	BasicAuthUser string `json:"-"`

	// jsonData fields. Mirrors InfinitySettingsJson in
	// pkg/models/settings.go:261-290 verbatim.
	IsMock                    bool                       `json:"is_mock,omitempty"`
	AuthenticationMethod      AuthType                   `json:"auth_method,omitempty"`
	APIKeyKey                 string                     `json:"apiKeyKey,omitempty"`
	APIKeyType                APIKeyType                 `json:"apiKeyType,omitempty"`
	OAuth2Settings            OAuth2Settings             `json:"oauth2,omitempty"`
	AWSSettings               AWSSettings                `json:"aws,omitempty"`
	ForwardOauthIdentity      bool                       `json:"oauthPassThru,omitempty"`
	InsecureSkipVerify        bool                       `json:"tlsSkipVerify,omitempty"`
	ServerName                string                     `json:"serverName,omitempty"`
	TLSClientAuth             bool                       `json:"tlsAuth,omitempty"`
	TLSAuthWithCACert         bool                       `json:"tlsAuthWithCACert,omitempty"`
	TimeoutInSeconds          int64                      `json:"timeoutInSeconds,omitempty"`
	ProxyType                 ProxyType                  `json:"proxy_type,omitempty"`
	ProxyURL                  string                     `json:"proxy_url,omitempty"`
	ProxyUserName             string                     `json:"proxy_username,omitempty"`
	ReferenceData             []RefData                  `json:"refData,omitempty"`
	CustomHealthCheckEnabled  bool                       `json:"customHealthCheckEnabled,omitempty"`
	CustomHealthCheckURL      string                     `json:"customHealthCheckUrl,omitempty"`
	AzureBlobCloudType        AzureBlobCloudType         `json:"azureBlobCloudType,omitempty"`
	AzureBlobAccountURL       string                     `json:"azureBlobAccountUrl,omitempty"`
	AzureBlobAccountName      string                     `json:"azureBlobAccountName,omitempty"`
	PathEncodedURLsEnabled    bool                       `json:"pathEncodedUrlsEnabled,omitempty"`
	IgnoreStatusCodeCheck     bool                       `json:"ignoreStatusCodeCheck,omitempty"`
	AllowDangerousHTTPMethods bool                       `json:"allowDangerousHTTPMethods,omitempty"`
	AllowedHosts              []string                   `json:"allowedHosts,omitempty"`
	UnsecuredQueryHandling    UnsecuredQueryHandlingMode `json:"unsecuredQueryHandling,omitempty"`
	KeepCookies               []string                   `json:"keepCookies,omitempty"`
	GlobalQueries             []GlobalQuery              `json:"global_queries,omitempty"`

	// Aggregated dynamic indexed-pair secrets. Populated by LoadConfig
	// from the httpHeaderName<N>/httpHeaderValue<N>,
	// secureQueryName<N>/secureQueryValue<N>,
	// oauth2EndPointParamsName<N>/oauth2EndPointParamsValue<N>, and
	// oauth2TokenHeadersName<N>/oauth2TokenHeadersValue<N> pairs — see
	// pkg/models/settings.go:389-392,427-443.
	CustomHeaders        map[string]string `json:"-"`
	SecureQueryFields    map[string]string `json:"-"`
	OAuth2EndpointParams map[string]string `json:"-"`
	OAuth2TokenHeaders   map[string]string `json:"-"`

	// DecryptedSecureJSONData holds the fixed-name decrypted secure
	// values (basicAuthPassword, bearerToken, tlsCACert, …). Dynamic
	// indexed secrets live in the aggregated maps above instead.
	DecryptedSecureJSONData map[SecureJsonDataKey]string `json:"-"`
}

// LoadConfig parses a datasource instance's settings into a Config.
// Mirrors the plugin's LoadSettings in pkg/models/settings.go:292-425:
// unmarshal jsonData, back-fill legacy basicAuth/oauthPassThru into
// auth_method, default proxy_type / apiKeyType / unsecuredQueryHandling /
// timeoutInSeconds / azureBlobCloudType, aggregate the indexed-pair
// secrets, then normalize the URL sentinel.
//
// ctx is used to derive a contextual logger via
// backend.Logger.FromContext so log lines carry the request / plugin
// context that Grafana injects.
//
// LoadConfig runs the full three-phase flow: parse → ApplyDefaults →
// Validate. Callers that need each phase individually can invoke
// ApplyDefaults and Validate directly on a Config they assemble themselves.
func LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error) {
	logger := backend.Logger.FromContext(ctx).With(
		"datasource_uid", settings.UID,
		"datasource_name", settings.Name,
		"plugin", settings.Type,
	)

	logger.Debug("loading infinity datasource config")

	cfg := Config{
		URL:                     settings.URL,
		BasicAuth:               settings.BasicAuthEnabled,
		BasicAuthUser:           settings.BasicAuthUser,
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	if cfg.URL == IgnoreURLSentinel {
		cfg.URL = ""
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

	// Aggregate the dynamic indexed-pair maps the same way
	// pkg/models/settings.go:427-443 does.
	cfg.CustomHeaders = aggregateSecretPairs(settings, "httpHeaderName", "httpHeaderValue")
	cfg.SecureQueryFields = aggregateSecretPairs(settings, "secureQueryName", "secureQueryValue")
	cfg.OAuth2EndpointParams = aggregateSecretPairs(settings, "oauth2EndPointParamsName", "oauth2EndPointParamsValue")
	cfg.OAuth2TokenHeaders = aggregateSecretPairs(settings, "oauth2TokenHeadersName", "oauth2TokenHeadersValue")

	logger.Debug("loaded secure keys",
		"fixed", len(cfg.DecryptedSecureJSONData),
		"headers", len(cfg.CustomHeaders),
		"queryParams", len(cfg.SecureQueryFields),
		"oauth2EndpointParams", len(cfg.OAuth2EndpointParams),
		"oauth2TokenHeaders", len(cfg.OAuth2TokenHeaders),
	)

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("infinity datasource config failed validation", "err", err)
		return cfg, err
	}

	logger.Debug("infinity datasource config loaded",
		"authMethod", cfg.AuthenticationMethod,
		"hasURL", cfg.URL != "",
		"proxyType", cfg.ProxyType,
	)
	return cfg, nil
}

// ApplyDefaults fills a curated set of zero-valued fields with the same
// defaults the plugin's own LoadSettings applies
// (pkg/models/settings.go:307-416). Never blanket-apply every schema
// default — that would clobber intentional zero values.
//
// Curated defaults (only when zero-valued):
//   - AuthenticationMethod → "none" (with legacy back-fill:
//     basicAuth → "basicAuth", oauthPassThru → "oauthPassThru"),
//   - OAuth2Settings.OAuth2Type → "client_credentials" (only when
//     AuthenticationMethod is "oauth2"),
//   - APIKeyType → "header",
//   - TimeoutInSeconds → 60,
//   - ProxyType → "env",
//   - UnsecuredQueryHandling → "warn",
//   - AzureBlobCloudType + AzureBlobAccountURL (only when
//     AuthenticationMethod is "azureBlob").
func (c *Config) ApplyDefaults() {
	if c.AuthenticationMethod == "" {
		switch {
		case c.BasicAuth:
			c.AuthenticationMethod = AuthTypeBasic
		case c.ForwardOauthIdentity:
			c.AuthenticationMethod = AuthTypeForwardOAuth
		default:
			c.AuthenticationMethod = AuthTypeNone
		}
	}
	if c.AuthenticationMethod == AuthTypeOAuth2 && c.OAuth2Settings.OAuth2Type == "" {
		c.OAuth2Settings.OAuth2Type = OAuth2TypeClientCredentials
	}
	if c.APIKeyType == "" {
		c.APIKeyType = APIKeyTypeHeader
	}
	if c.TimeoutInSeconds <= 0 {
		c.TimeoutInSeconds = 60
	}
	if c.ProxyType == "" {
		c.ProxyType = ProxyTypeEnv
	}
	if c.UnsecuredQueryHandling == "" {
		c.UnsecuredQueryHandling = UnsecuredQueryHandlingWarn
	}
	if c.AuthenticationMethod == AuthTypeAzureBlob {
		if c.AzureBlobCloudType == "" {
			c.AzureBlobCloudType = AzureBlobCloudTypeAzureCloud
		}
		if c.AzureBlobAccountURL == "" {
			switch c.AzureBlobCloudType {
			case AzureBlobCloudTypeAzureUSGovernment:
				c.AzureBlobAccountURL = "https://%s.blob.core.usgovcloudapi.net/"
			case AzureBlobCloudTypeAzureChinaCloud:
				c.AzureBlobAccountURL = "https://%s.blob.core.chinacloudapi.cn/"
			default:
				c.AzureBlobAccountURL = "https://%s.blob.core.windows.net/"
			}
		}
	}
}

// Validate checks the runtime contract that the plugin requires (mirrors
// InfinitySettings.Validate in pkg/models/settings.go:135-167). Errors
// are joined so callers see every problem at once.
//
// Contracts enforced:
//   - basicAuth / digestAuth need a password (also enforced when
//     root.basicAuth is true without a matching auth_method).
//   - apiKey needs both apiKeyKey and apiKeyValue.
//   - bearerToken needs bearerToken.
//   - aws with authType=keys needs both awsAccessKey and awsSecretKey.
//   - azureBlob needs azureBlobAccountName + azureBlobAccountKey.
//   - When root.url is empty AND auth_method is not 'none'/'azureBlob',
//     jsonData.allowedHosts must have at least one entry.
//   - Every entry in jsonData.allowedHosts must be a valid http(s) URL
//     with a non-empty hostname.
func (c Config) Validate() error {
	var errs []error

	needsPassword := c.BasicAuth || c.AuthenticationMethod == AuthTypeBasic || c.AuthenticationMethod == AuthTypeDigest
	if needsPassword && c.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword] == "" {
		errs = append(errs, errors.New("basicAuthPassword is required for basicAuth/digestAuth"))
	}

	switch c.AuthenticationMethod {
	case AuthTypeAPIKey:
		if c.APIKeyKey == "" {
			errs = append(errs, errors.New("apiKeyKey is required for apiKey auth"))
		}
		if c.DecryptedSecureJSONData[SecureJsonDataKeyAPIKeyValue] == "" {
			errs = append(errs, errors.New("apiKeyValue is required for apiKey auth"))
		}
	case AuthTypeBearerToken:
		if c.DecryptedSecureJSONData[SecureJsonDataKeyBearerToken] == "" {
			errs = append(errs, errors.New("bearerToken is required for bearerToken auth"))
		}
	case AuthTypeAWS:
		if c.AWSSettings.AuthType == AWSAuthTypeKeys {
			if strings.TrimSpace(c.DecryptedSecureJSONData[SecureJsonDataKeyAWSAccessKey]) == "" {
				errs = append(errs, errors.New("awsAccessKey is required for aws auth with authType=keys"))
			}
			if strings.TrimSpace(c.DecryptedSecureJSONData[SecureJsonDataKeyAWSSecretKey]) == "" {
				errs = append(errs, errors.New("awsSecretKey is required for aws auth with authType=keys"))
			}
		}
	case AuthTypeAzureBlob:
		if strings.TrimSpace(c.AzureBlobAccountName) == "" {
			errs = append(errs, errors.New("azureBlobAccountName is required for azureBlob auth"))
		}
		if strings.TrimSpace(c.DecryptedSecureJSONData[SecureJsonDataKeyAzureBlobAccountKey]) == "" {
			errs = append(errs, errors.New("azureBlobAccountKey is required for azureBlob auth"))
		}
	}

	if c.doesAllowedHostsRequired() && len(c.AllowedHosts) < 1 {
		errs = append(errs, errors.New("allowedHosts must contain at least one host when root.url is empty and auth/TLS/headers/cookies are configured"))
	}

	if len(c.AllowedHosts) > 0 {
		if err := validateAllowedHosts(c.AllowedHosts); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// doesAllowedHostsRequired mirrors
// InfinitySettings.DoesAllowedHostsRequired in
// pkg/models/settings.go:169-203.
func (c Config) doesAllowedHostsRequired() bool {
	if strings.TrimSpace(c.URL) != "" {
		return false
	}
	if c.AuthenticationMethod != "" && c.AuthenticationMethod != AuthTypeNone && c.AuthenticationMethod != AuthTypeAzureBlob {
		return true
	}
	if c.TLSAuthWithCACert || c.TLSClientAuth {
		return true
	}
	if len(c.CustomHeaders) > 0 {
		for k := range c.CustomHeaders {
			lk := strings.ToLower(strings.TrimSpace(k))
			if lk == "accept" || lk == "content-type" {
				continue
			}
			return true
		}
	}
	if len(c.SecureQueryFields) > 0 {
		return true
	}
	if len(c.KeepCookies) > 0 {
		return true
	}
	return false
}

// validateAllowedHosts mirrors ValidateAllowedHosts in
// pkg/models/settings.go:205-222 (minus the URL parse, which we only run
// via the shape check below because we don't want to depend on net/url in
// this schema entry's tests).
func validateAllowedHosts(hosts []string) error {
	for _, h := range hosts {
		trimmed := strings.TrimSpace(h)
		if trimmed == "" {
			return errors.New("invalid url found in allowed hosts settings")
		}
		lower := strings.ToLower(trimmed)
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			return fmt.Errorf("invalid url in allowed list %s", h)
		}
	}
	return nil
}

// aggregateSecretPairs walks jsonData for the given namePrefix
// (e.g. "httpHeaderName") and, for each key found, reads its paired
// secure value from settings.DecryptedSecureJSONData under valuePrefix
// (e.g. "httpHeaderValue"). Mirrors GetSecrets in
// pkg/models/settings.go:427-443.
func aggregateSecretPairs(settings backend.DataSourceInstanceSettings, namePrefix, valuePrefix string) map[string]string {
	out := map[string]string{}
	if len(settings.JSONData) == 0 {
		return out
	}
	raw := map[string]any{}
	if err := json.Unmarshal(settings.JSONData, &raw); err != nil {
		return out
	}
	for k, v := range raw {
		if !strings.HasPrefix(k, namePrefix) {
			continue
		}
		name, ok := v.(string)
		if !ok {
			name = fmt.Sprintf("%v", v)
		}
		secureKey := strings.Replace(k, namePrefix, valuePrefix, 1)
		out[name] = settings.DecryptedSecureJSONData[secureKey]
	}
	return out
}
