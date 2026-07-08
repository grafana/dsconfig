package jiradatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with jsonData and secureJsonData) into the
// backend.DataSourceInstanceSettings shape LoadConfig expects.
func settingsFromExample(t *testing.T, exampleKey string) backend.DataSourceInstanceSettings {
	t.Helper()
	ex, ok := SettingsExamples().Examples[exampleKey]
	if !ok {
		t.Fatalf("unknown example %q", exampleKey)
	}
	value, ok := ex.Value.(map[string]any)
	if !ok {
		t.Fatalf("example %q value is not an object", exampleKey)
	}
	jsonData, err := json.Marshal(value["jsonData"])
	if err != nil {
		t.Fatalf("marshal jsonData: %v", err)
	}
	secure := map[string]string{}
	if raw, ok := value["secureJsonData"].(map[string]any); ok {
		for k, v := range raw {
			s, _ := v.(string)
			secure[k] = s
		}
	}
	return backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: secure,
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name            string
		example         string // schema.go SettingsExamples key
		settings        backend.DataSourceInstanceSettings
		useSettings     bool
		wantErr         error
		wantAuthMethod  AuthMethod
		wantHosting     Hosting
		wantURL         string
		wantUser        string
		wantScopedToken bool
		wantCloudId     string
		wantSecureKeys  SecureJsonDataConfig
	}{
		{
			// The default schema example has no url and an empty token
			// placeholder, so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (no url, empty token)",
			example: "",
			wantErr: errors.New("URL is missing"),
		},
		{
			name:           "basic auth cloud",
			example:        "basicAuthCloud",
			wantAuthMethod: AuthMethodBasicAuth,
			wantHosting:    HostingCloud,
			wantURL:        "https://mycompany.atlassian.net",
			wantUser:       "user@example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			name:           "basic auth server",
			example:        "basicAuthServer",
			wantAuthMethod: AuthMethodBasicAuth,
			wantHosting:    HostingServer,
			wantURL:        "https://jira.example.com",
			wantUser:       "user@example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			name:           "bearer token server (no user)",
			example:        "bearerTokenServer",
			wantAuthMethod: AuthMethodBasicAuth,
			wantHosting:    HostingServer,
			wantURL:        "https://jira.example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			name:            "basic auth scoped token",
			example:         "basicAuthScopedToken",
			wantAuthMethod:  AuthMethodBasicAuth,
			wantHosting:     HostingCloud,
			wantURL:         "https://mycompany.atlassian.net",
			wantUser:        "user@example.com",
			wantScopedToken: true,
			wantCloudId:     "<your-jira-cloud-id>",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			name:           "oauth2",
			example:        "oauth2",
			wantAuthMethod: AuthMethodOAuth2,
			wantHosting:    HostingCloud,
			wantURL:        "https://mycompany.atlassian.net",
			wantCloudId:    "<your-jira-cloud-id>",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyOAuthClientSecret},
		},
		{
			name:           "legacy basic auth without authMethod",
			example:        "legacyBasicAuthNoAuthMethod",
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "https://mycompany.atlassian.net",
			wantUser:       "user@example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			// Empty JSONData is a parse error upstream —
			// pkg/models/settings.go:33 json.Unmarshal(nil, settings) fails.
			name:        "empty settings error (empty JSONData)",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
			wantErr:     errors.New("parse jsonData"),
		},
		{
			name:        "invalid jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name:        "basic auth missing token errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authMethod":"basicAuth","url":"https://mycompany.atlassian.net","user":"u"}`),
			},
			wantErr: errors.New("token is missing"),
		},
		{
			name:        "oauth2 missing client id errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authMethod":"oauth2","url":"https://mycompany.atlassian.net","cloudId":"c"}`),
				DecryptedSecureJSONData: map[string]string{"oauthClientSecret": "s"},
			},
			wantErr: errors.New("OAuth client ID is missing"),
		},
		{
			name:        "oauth2 missing cloud id errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authMethod":"oauth2","url":"https://mycompany.atlassian.net","oauthClientID":"id"}`),
				DecryptedSecureJSONData: map[string]string{"oauthClientSecret": "s"},
			},
			wantErr: errors.New("Cloud ID is required for OAuth 2.0 authentication"),
		},
		{
			name:        "scoped token missing cloud id errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authMethod":"basicAuth","url":"https://mycompany.atlassian.net","user":"u","scopedToken":true}`),
				DecryptedSecureJSONData: map[string]string{"token": "t"},
			},
			wantErr: errors.New("cloud ID is required for scoped token"),
		},
		{
			// A URL with no scheme is normalized to https:// on load
			// (pkg/models/settings.go:43-45).
			name:        "url without scheme is normalized to https",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authMethod":"basicAuth","url":"mycompany.atlassian.net","user":"u"}`),
				DecryptedSecureJSONData: map[string]string{"token": "t"},
			},
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "https://mycompany.atlassian.net",
		},
		{
			// An http:// URL keeps its scheme untouched.
			name:        "http url keeps its scheme",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authMethod":"basicAuth","url":"http://jira.example.com","user":"u"}`),
				DecryptedSecureJSONData: map[string]string{"token": "t"},
			},
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "http://jira.example.com",
		},
		{
			// An unknown authMethod resolves to basicAuth and therefore requires
			// a token (pkg/models/auth_method.go:11-15).
			name:        "unknown authMethod resolves to basic auth",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authMethod":"whatever","url":"https://mycompany.atlassian.net","user":"u"}`),
				DecryptedSecureJSONData: map[string]string{"token": "t"},
			},
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "https://mycompany.atlassian.net",
		},
		{
			// enableSecureSocksProxy is intentionally omitted from both the
			// dsconfig schema and the Go Config struct; json unmarshal silently
			// ignores unknown fields.
			name:        "unknown enableSecureSocksProxy field is ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authMethod":"basicAuth","url":"https://mycompany.atlassian.net","user":"u","enableSecureSocksProxy":true}`),
				DecryptedSecureJSONData: map[string]string{"token": "t"},
			},
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "https://mycompany.atlassian.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if !tt.useSettings {
				settings = settingsFromExample(t, tt.example)
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("LoadConfig: expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Fatalf("LoadConfig: error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}

			if tt.wantAuthMethod != "" && cfg.AuthMethod != tt.wantAuthMethod {
				t.Errorf("AuthMethod = %q, want %q", cfg.AuthMethod, tt.wantAuthMethod)
			}
			if tt.wantHosting != "" && cfg.Hosting != tt.wantHosting {
				t.Errorf("Hosting = %q, want %q", cfg.Hosting, tt.wantHosting)
			}
			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.wantUser != "" && cfg.User != tt.wantUser {
				t.Errorf("User = %q, want %q", cfg.User, tt.wantUser)
			}
			if cfg.ScopedToken != tt.wantScopedToken {
				t.Errorf("ScopedToken = %v, want %v", cfg.ScopedToken, tt.wantScopedToken)
			}
			if tt.wantCloudId != "" && cfg.CloudId != tt.wantCloudId {
				t.Errorf("CloudId = %q, want %q", cfg.CloudId, tt.wantCloudId)
			}
			if tt.wantSecureKeys != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if _, ok := cfg.DecryptedSecureJSONData[key]; ok {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantSecureKeys) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantSecureKeys)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name           string
		in             Config
		wantAuthMethod AuthMethod
		wantURL        string
	}{
		{
			name:           "empty config defaults to basic auth, url untouched",
			in:             Config{},
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "",
		},
		{
			name:           "oauth2 is preserved",
			in:             Config{AuthMethod: AuthMethodOAuth2},
			wantAuthMethod: AuthMethodOAuth2,
		},
		{
			name:           "unknown auth method resolves to basic auth",
			in:             Config{AuthMethod: "whatever"},
			wantAuthMethod: AuthMethodBasicAuth,
		},
		{
			name:           "url without scheme gets https prefix",
			in:             Config{URL: "mycompany.atlassian.net"},
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "https://mycompany.atlassian.net",
		},
		{
			name:           "http url keeps its scheme",
			in:             Config{URL: "http://jira.example.com"},
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "http://jira.example.com",
		},
		{
			name:           "https url keeps its scheme",
			in:             Config{URL: "https://mycompany.atlassian.net"},
			wantAuthMethod: AuthMethodBasicAuth,
			wantURL:        "https://mycompany.atlassian.net",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthMethod != tt.wantAuthMethod {
				t.Errorf("AuthMethod = %q, want %q", got.AuthMethod, tt.wantAuthMethod)
			}
			if got.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tt.wantURL)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	const url = "https://mycompany.atlassian.net"
	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "basic auth happy path",
			cfg: Config{
				AuthMethod:              AuthMethodBasicAuth,
				URL:                     url,
				User:                    "u",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "t"},
			},
		},
		{
			name: "basic auth bearer (no user) happy path",
			cfg: Config{
				AuthMethod:              AuthMethodBasicAuth,
				URL:                     url,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "t"},
			},
		},
		{
			name:    "basic auth missing token errors",
			cfg:     Config{AuthMethod: AuthMethodBasicAuth, URL: url},
			wantErr: "token is missing",
		},
		{
			name: "missing url errors",
			cfg: Config{
				AuthMethod:              AuthMethodBasicAuth,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "t"},
			},
			wantErr: "URL is missing",
		},
		{
			name: "scoped token requires cloud id",
			cfg: Config{
				AuthMethod:              AuthMethodBasicAuth,
				URL:                     url,
				ScopedToken:             true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "t"},
			},
			wantErr: "cloud ID is required for scoped token",
		},
		{
			name: "scoped token with cloud id ok",
			cfg: Config{
				AuthMethod:              AuthMethodBasicAuth,
				URL:                     url,
				ScopedToken:             true,
				CloudId:                 "c",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "t"},
			},
		},
		{
			name: "oauth2 happy path",
			cfg: Config{
				AuthMethod:              AuthMethodOAuth2,
				URL:                     url,
				OAuthClientID:           "id",
				CloudId:                 "c",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyOAuthClientSecret: "s"},
			},
		},
		{
			name:    "oauth2 missing everything joins errors",
			cfg:     Config{AuthMethod: AuthMethodOAuth2, URL: url},
			wantErr: "OAuth client ID is missing",
		},
		{
			name: "oauth2 missing cloud id errors",
			cfg: Config{
				AuthMethod:              AuthMethodOAuth2,
				URL:                     url,
				OAuthClientID:           "id",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyOAuthClientSecret: "s"},
			},
			wantErr: "Cloud ID is required for OAuth 2.0 authentication",
		},
		{
			// Any non-oauth2 authType validates as basic auth (matching the
			// upstream switch default).
			name:    "empty auth method validates as basic auth",
			cfg:     Config{URL: url},
			wantErr: "token is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate: unexpected error %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate: expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate: error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
