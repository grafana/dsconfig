package servicenowdatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with root fields, jsonData, and secureJsonData) into the
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
	settings := backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: secure,
	}
	if s, ok := value["url"].(string); ok {
		settings.URL = s
	}
	if s, ok := value["basicAuthUser"].(string); ok {
		settings.BasicAuthUser = s
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name             string
		example          string // schema.go SettingsExamples key ("" = use inline settings)
		settings         backend.DataSourceInstanceSettings
		wantErr          error
		wantAuthMethod   AuthMethod
		wantURL          string
		wantUser         string
		wantClientID     string
		wantSecureKeys   SecureJsonDataConfig
		wantQueryTimeout int
		wantUseSysTables bool
	}{
		{
			// The default example intentionally leaves basicAuthUser and
			// basicAuthPassword empty, so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (empty basic credentials)",
			example: "",
			wantErr: errors.New("invalid username: basicAuthUser"),
		},
		{
			name:             "basic auth",
			example:          "basicAuth",
			wantAuthMethod:   AuthMethodBasicAuth,
			wantURL:          "https://acme.service-now.com",
			wantUser:         "grafana_reader",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
			wantQueryTimeout: 30,
			wantUseSysTables: true,
		},
		{
			name:             "servicenow oauth",
			example:          "serviceNowOAuth",
			wantAuthMethod:   AuthMethodServiceNowOAuth,
			wantURL:          "https://acme.service-now.com",
			wantUser:         "grafana_reader",
			wantClientID:     "<your-oauth-client-id>",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword, SecureJsonDataKeyOAuthClientSecret},
			wantQueryTimeout: 30,
		},
		{
			name:             "legacy oauthEnabled resolves to oauth",
			example:          "legacyOAuthEnabled",
			wantAuthMethod:   AuthMethodServiceNowOAuth,
			wantURL:          "https://acme.service-now.com",
			wantUser:         "grafana_reader",
			wantClientID:     "<your-oauth-client-id>",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword, SecureJsonDataKeyOAuthClientSecret},
			wantQueryTimeout: 30,
		},
		{
			name: "empty authMethod defaults to basic auth",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				JSONData:                []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{"basicAuthPassword": "pw"},
			},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantURL:          "https://acme.service-now.com",
			wantUser:         "grafana_reader",
			wantQueryTimeout: 30,
		},
		{
			name: "unknown authMethod is coerced to basic auth (mirrors GetAuthMethod)",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				JSONData:                []byte(`{"authMethod":"bogus"}`),
				DecryptedSecureJSONData: map[string]string{"basicAuthPassword": "pw"},
			},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantQueryTimeout: 30,
		},
		{
			name: "queryTimeoutSeconds is preserved when >= 1",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				JSONData:                []byte(`{"authMethod":"basicAuth","queryTimeoutSeconds":120}`),
				DecryptedSecureJSONData: map[string]string{"basicAuthPassword": "pw"},
			},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantQueryTimeout: 120,
		},
		{
			name: "queryTimeoutSeconds < 1 defaults to 30",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				JSONData:                []byte(`{"authMethod":"basicAuth","queryTimeoutSeconds":0}`),
				DecryptedSecureJSONData: map[string]string{"basicAuthPassword": "pw"},
			},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantQueryTimeout: 30,
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				BasicAuthUser:           "grafana_reader",
				JSONData:                []byte(`{"authMethod":"basicAuth"}`),
				DecryptedSecureJSONData: map[string]string{"basicAuthPassword": "pw"},
			},
			wantErr: errors.New("URL (root.url) is required"),
		},
		{
			name: "basic auth without password errors",
			settings: backend.DataSourceInstanceSettings{
				URL:           "https://acme.service-now.com",
				BasicAuthUser: "grafana_reader",
				JSONData:      []byte(`{"authMethod":"basicAuth"}`),
			},
			wantErr: errors.New("invalid password: basicAuthPassword"),
		},
		{
			name: "basic auth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://acme.service-now.com",
				JSONData:                []byte(`{"authMethod":"basicAuth"}`),
				DecryptedSecureJSONData: map[string]string{"basicAuthPassword": "pw"},
			},
			wantErr: errors.New("invalid username: basicAuthUser"),
		},
		{
			name: "oauth without clientID errors",
			settings: backend.DataSourceInstanceSettings{
				URL:           "https://acme.service-now.com",
				BasicAuthUser: "grafana_reader",
				JSONData:      []byte(`{"authMethod":"serviceNowOAuth"}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "pw",
					"oauthClientSecret": "sec",
				},
			},
			wantErr: errors.New("oauthClientID (jsonData) is required"),
		},
		{
			name: "oauth without clientSecret errors",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				JSONData:                []byte(`{"authMethod":"serviceNowOAuth","oauthClientID":"cid"}`),
				DecryptedSecureJSONData: map[string]string{"basicAuthPassword": "pw"},
			},
			wantErr: errors.New("oauthClientSecret (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://acme.service-now.com",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			// Empty settings default to basicAuth, which requires URL + credentials.
			name:     "empty settings fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("URL (root.url) is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" || (tt.settings.JSONData == nil && tt.settings.URL == "" && tt.wantErr == nil) {
				if _, ok := SettingsExamples().Examples[tt.example]; ok {
					settings = settingsFromExample(t, tt.example)
				}
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("LoadConfig: expected error %q, got nil", tt.wantErr)
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
			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.wantUser != "" && cfg.BasicAuthUser != tt.wantUser {
				t.Errorf("BasicAuthUser = %q, want %q", cfg.BasicAuthUser, tt.wantUser)
			}
			if tt.wantClientID != "" && cfg.OAuthClientID != tt.wantClientID {
				t.Errorf("OAuthClientID = %q, want %q", cfg.OAuthClientID, tt.wantClientID)
			}
			if tt.wantQueryTimeout != 0 && cfg.QueryTimeoutSeconds != tt.wantQueryTimeout {
				t.Errorf("QueryTimeoutSeconds = %d, want %d", cfg.QueryTimeoutSeconds, tt.wantQueryTimeout)
			}
			if cfg.UseSysTables != tt.wantUseSysTables {
				t.Errorf("UseSysTables = %v, want %v", cfg.UseSysTables, tt.wantUseSysTables)
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
		name             string
		in               Config
		wantAuthMethod   AuthMethod
		wantQueryTimeout int
	}{
		{
			name:             "empty config gets basicAuth + 30s timeout",
			in:               Config{},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantQueryTimeout: 30,
		},
		{
			name:             "legacy oauthEnabled resolves to serviceNowOAuth",
			in:               Config{OAuthEnabled: true},
			wantAuthMethod:   AuthMethodServiceNowOAuth,
			wantQueryTimeout: 30,
		},
		{
			name:             "explicit serviceNowOAuth is preserved",
			in:               Config{AuthMethod: AuthMethodServiceNowOAuth},
			wantAuthMethod:   AuthMethodServiceNowOAuth,
			wantQueryTimeout: 30,
		},
		{
			name:             "explicit basicAuth is preserved",
			in:               Config{AuthMethod: AuthMethodBasicAuth},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantQueryTimeout: 30,
		},
		{
			name:             "unknown authMethod coerces to basicAuth",
			in:               Config{AuthMethod: "bogus"},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantQueryTimeout: 30,
		},
		{
			// Upstream GetAuthMethod (pkg/models/auth_method.go:12-22) only
			// short-circuits on an explicit serviceNowOAuth; the legacy
			// oauthEnabled flag then wins over an explicit basicAuth. Faithfully
			// mirrored here (see README discrepancies).
			name:             "legacy oauthEnabled overrides explicit basicAuth (upstream quirk)",
			in:               Config{AuthMethod: AuthMethodBasicAuth, OAuthEnabled: true},
			wantAuthMethod:   AuthMethodServiceNowOAuth,
			wantQueryTimeout: 30,
		},
		{
			name:             "valid queryTimeout is preserved",
			in:               Config{QueryTimeoutSeconds: 90},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantQueryTimeout: 90,
		},
		{
			name:             "negative queryTimeout defaults to 30",
			in:               Config{QueryTimeoutSeconds: -5},
			wantAuthMethod:   AuthMethodBasicAuth,
			wantQueryTimeout: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthMethod != tt.wantAuthMethod {
				t.Errorf("AuthMethod = %q, want %q", got.AuthMethod, tt.wantAuthMethod)
			}
			if got.QueryTimeoutSeconds != tt.wantQueryTimeout {
				t.Errorf("QueryTimeoutSeconds = %d, want %d", got.QueryTimeoutSeconds, tt.wantQueryTimeout)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	basicPassword := map[SecureJsonDataKey]string{SecureJsonDataKeyBasicAuthPassword: "pw"}
	oauthSecrets := map[SecureJsonDataKey]string{
		SecureJsonDataKeyBasicAuthPassword: "pw",
		SecureJsonDataKeyOAuthClientSecret: "sec",
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "basicAuth happy path",
			cfg: Config{
				AuthMethod:              AuthMethodBasicAuth,
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				DecryptedSecureJSONData: basicPassword,
			},
		},
		{
			name: "basicAuth missing user",
			cfg: Config{
				AuthMethod:              AuthMethodBasicAuth,
				URL:                     "https://acme.service-now.com",
				DecryptedSecureJSONData: basicPassword,
			},
			wantErr: "invalid username: basicAuthUser",
		},
		{
			name: "basicAuth missing password",
			cfg: Config{
				AuthMethod:    AuthMethodBasicAuth,
				URL:           "https://acme.service-now.com",
				BasicAuthUser: "grafana_reader",
			},
			wantErr: "invalid password: basicAuthPassword",
		},
		{
			name: "serviceNowOAuth happy path",
			cfg: Config{
				AuthMethod:              AuthMethodServiceNowOAuth,
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				OAuthClientID:           "cid",
				DecryptedSecureJSONData: oauthSecrets,
			},
		},
		{
			name: "serviceNowOAuth missing clientID",
			cfg: Config{
				AuthMethod:              AuthMethodServiceNowOAuth,
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				DecryptedSecureJSONData: oauthSecrets,
			},
			wantErr: "oauthClientID (jsonData) is required",
		},
		{
			name: "serviceNowOAuth missing clientSecret",
			cfg: Config{
				AuthMethod:              AuthMethodServiceNowOAuth,
				URL:                     "https://acme.service-now.com",
				BasicAuthUser:           "grafana_reader",
				OAuthClientID:           "cid",
				DecryptedSecureJSONData: basicPassword,
			},
			wantErr: "oauthClientSecret (secureJsonData) is required",
		},
		{
			name: "serviceNowOAuth missing username (password grant needs it)",
			cfg: Config{
				AuthMethod:              AuthMethodServiceNowOAuth,
				URL:                     "https://acme.service-now.com",
				OAuthClientID:           "cid",
				DecryptedSecureJSONData: oauthSecrets,
			},
			wantErr: "invalid username: basicAuthUser",
		},
		{
			name: "missing URL",
			cfg: Config{
				AuthMethod:              AuthMethodBasicAuth,
				BasicAuthUser:           "grafana_reader",
				DecryptedSecureJSONData: basicPassword,
			},
			wantErr: "invalid server name: URL (root.url) is required",
		},
		{
			name: "empty auth method errors",
			cfg: Config{
				URL:           "https://acme.service-now.com",
				BasicAuthUser: "grafana_reader",
			},
			wantErr: "authentication method not set",
		},
		{
			name: "unknown auth method errors",
			cfg: Config{
				AuthMethod:    "bogus",
				URL:           "https://acme.service-now.com",
				BasicAuthUser: "grafana_reader",
			},
			wantErr: "invalid authentication method: bogus",
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

// TestSettingsExamples guards the AGENTS.md example requirements: the default
// example keyed by "" exists, and every example carries a jsonData object and a
// non-empty secureJsonData object whose keys are all known secret keys.
func TestSettingsExamples(t *testing.T) {
	examples := SettingsExamples().Examples
	if _, ok := examples[""]; !ok {
		t.Fatalf(`missing default example keyed by ""`)
	}

	known := map[string]bool{}
	for _, k := range SecureJsonDataKeys {
		known[string(k)] = true
	}

	for key, ex := range examples {
		value, ok := ex.Value.(map[string]any)
		if !ok {
			t.Fatalf("example %q value is not an object", key)
		}
		if _, ok := value["jsonData"].(map[string]any); !ok {
			t.Errorf("example %q missing jsonData object", key)
		}
		secure, ok := value["secureJsonData"].(map[string]any)
		if !ok || len(secure) == 0 {
			t.Errorf("example %q missing non-empty secureJsonData object", key)
			continue
		}
		for secretKey := range secure {
			if !known[secretKey] {
				t.Errorf("example %q references unknown secret key %q", key, secretKey)
			}
		}
	}
}
