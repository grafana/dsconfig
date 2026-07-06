package splunkdatasource

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
		name           string
		example        string
		settings       backend.DataSourceInstanceSettings
		wantErr        error
		wantURL        string
		wantAuthType   AuthType
		wantPreview    bool
		wantTimeout    int64
		wantPattern    string
		wantTimeField  string
		wantSecureKeys SecureJsonDataConfig
	}{
		{
			// The default example is Basic auth but leaves basicAuthUser and the
			// password empty as placeholders, so validation fails.
			name:    "default example fails validation (empty basic auth placeholders)",
			example: "",
			wantErr: errors.New("required"),
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://splunk.example.com:8089",
			wantAuthType:   AuthTypeBasicAuth,
			wantTimeout:    DefaultTimeoutInSeconds,
			wantTimeField:  DefaultTimeField,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:           "alternative token auth",
			example:        "alternativeToken",
			wantURL:        "https://splunk.example.com:8089",
			wantAuthType:   AuthTypeAlternativeToken,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAuthToken},
		},
		{
			name:         "oauth forward",
			example:      "oauthForward",
			wantURL:      "https://splunk.example.com:8089",
			wantAuthType: AuthTypeOAuthForward,
		},
		{
			name:         "basic auth with tls",
			example:      "basicAuthWithTLS",
			wantURL:      "https://splunk.example.com:8089",
			wantAuthType: AuthTypeBasicAuth,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyBasicAuthPassword,
				SecureJsonDataKeyTLSCACert,
				SecureJsonDataKeyTLSClientCert,
				SecureJsonDataKeyTLSClientKey,
			},
		},
		{
			name:           "token auth with advanced options",
			example:        "tokenWithAdvancedOptions",
			wantURL:        "https://splunk.example.com:8089",
			wantAuthType:   AuthTypeAlternativeToken,
			wantPreview:    true,
			wantTimeout:    60,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAuthToken},
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"custom-splunk"}`),
				DecryptedSecureJSONData: map[string]string{"authToken": "tok"},
			},
			wantErr: errors.New("URL (root.url) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://splunk.example.com:8089",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "empty authType treated as basic auth",
			settings: backend.DataSourceInstanceSettings{
				URL:           "https://splunk.example.com:8089",
				BasicAuthUser: "admin",
				JSONData:      []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "pw",
				},
			},
			wantURL:       "https://splunk.example.com:8089",
			wantAuthType:  AuthTypeBasicAuth,
			wantTimeout:   DefaultTimeoutInSeconds,
			wantTimeField: DefaultTimeField,
		},
		{
			name: "streamMode migrates to previewMode",
			settings: backend.DataSourceInstanceSettings{
				URL:           "https://splunk.example.com:8089",
				BasicAuthUser: "admin",
				JSONData:      []byte(`{"streamMode":true}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "pw",
				},
			},
			wantAuthType: AuthTypeBasicAuth,
			wantPreview:  true,
		},
		{
			name: "internal field pattern defaults when filtration enabled",
			settings: backend.DataSourceInstanceSettings{
				URL:           "https://splunk.example.com:8089",
				BasicAuthUser: "admin",
				JSONData:      []byte(`{"internalFieldsFiltration":true}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "pw",
				},
			},
			wantAuthType: AuthTypeBasicAuth,
			wantPattern:  DefaultInternalFieldPattern,
		},
		{
			name: "internal field pattern cleared when filtration disabled",
			settings: backend.DataSourceInstanceSettings{
				URL:           "https://splunk.example.com:8089",
				BasicAuthUser: "admin",
				JSONData:      []byte(`{"internalFieldsFiltration":false,"internalFieldPattern":"custom"}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "pw",
				},
			},
			wantAuthType: AuthTypeBasicAuth,
			wantPattern:  "",
		},
		{
			name: "timeout below one is bumped to default",
			settings: backend.DataSourceInstanceSettings{
				URL:           "https://splunk.example.com:8089",
				BasicAuthUser: "admin",
				JSONData:      []byte(`{"timeoutInSeconds":0}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "pw",
				},
			},
			wantAuthType: AuthTypeBasicAuth,
			wantTimeout:  DefaultTimeoutInSeconds,
		},
		{
			name: "alternative token without token errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://splunk.example.com:8089",
				JSONData: []byte(`{"authType":"custom-splunk"}`),
			},
			wantErr: errors.New("authToken (secureJsonData) is required"),
		},
		{
			name: "unknown authType errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://splunk.example.com:8089",
				JSONData: []byte(`{"authType":"bogus"}`),
			},
			wantErr: errors.New(`unknown authType "bogus"`),
		},
		{
			name: "tls client auth without cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://splunk.example.com:8089",
				JSONData: []byte(`{"authType":"custom-splunk","tlsAuth":true}`),
				DecryptedSecureJSONData: map[string]string{
					"authToken": "tok",
				},
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.settings.JSONData == nil && tt.settings.URL == "" {
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

			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.wantAuthType != "" && cfg.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", cfg.AuthType, tt.wantAuthType)
			}
			if cfg.PreviewMode != tt.wantPreview {
				t.Errorf("PreviewMode = %v, want %v", cfg.PreviewMode, tt.wantPreview)
			}
			if tt.wantTimeout != 0 && cfg.TimeoutInSeconds != tt.wantTimeout {
				t.Errorf("TimeoutInSeconds = %d, want %d", cfg.TimeoutInSeconds, tt.wantTimeout)
			}
			if tt.wantTimeField != "" && cfg.TimeField != tt.wantTimeField {
				t.Errorf("TimeField = %q, want %q", cfg.TimeField, tt.wantTimeField)
			}
			if tt.name == "internal field pattern cleared when filtration disabled" && cfg.InternalFieldPattern != "" {
				t.Errorf("InternalFieldPattern = %q, want empty", cfg.InternalFieldPattern)
			}
			if tt.wantPattern != "" && cfg.InternalFieldPattern != tt.wantPattern {
				t.Errorf("InternalFieldPattern = %q, want %q", cfg.InternalFieldPattern, tt.wantPattern)
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
		name       string
		in         Config
		wantAuth   AuthType
		wantField  FieldSearchType
		wantVarLvl VariableSearchLevel
	}{
		{
			name:       "empty config gets editor defaults",
			in:         Config{},
			wantAuth:   AuthTypeBasicAuth,
			wantField:  FieldSearchTypeQuick,
			wantVarLvl: VariableSearchLevelFast,
		},
		{
			name:       "existing auth type preserved",
			in:         Config{AuthType: AuthTypeAlternativeToken},
			wantAuth:   AuthTypeAlternativeToken,
			wantField:  FieldSearchTypeQuick,
			wantVarLvl: VariableSearchLevelFast,
		},
		{
			name:       "existing search modes preserved",
			in:         Config{FieldSearchType: FieldSearchTypeFull, VariableSearchLevel: VariableSearchLevelVerbose},
			wantAuth:   AuthTypeBasicAuth,
			wantField:  FieldSearchTypeFull,
			wantVarLvl: VariableSearchLevelVerbose,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuth {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuth)
			}
			if got.FieldSearchType != tt.wantField {
				t.Errorf("FieldSearchType = %q, want %q", got.FieldSearchType, tt.wantField)
			}
			if got.VariableSearchLevel != tt.wantVarLvl {
				t.Errorf("VariableSearchLevel = %q, want %q", got.VariableSearchLevel, tt.wantVarLvl)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "basic auth happy path",
			cfg: Config{
				URL:           "https://splunk.example.com:8089",
				AuthType:      AuthTypeBasicAuth,
				BasicAuthUser: "admin",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBasicAuthPassword: "pw",
				},
			},
		},
		{
			name: "empty authType treated as basic auth happy path",
			cfg: Config{
				URL:           "https://splunk.example.com:8089",
				BasicAuthUser: "admin",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBasicAuthPassword: "pw",
				},
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{AuthType: AuthTypeOAuthForward},
			wantErr: "URL (root.url) is required",
		},
		{
			name: "basic auth missing user",
			cfg: Config{
				URL:      "https://splunk.example.com:8089",
				AuthType: AuthTypeBasicAuth,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBasicAuthPassword: "pw",
				},
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "basic auth missing password",
			cfg: Config{
				URL:           "https://splunk.example.com:8089",
				AuthType:      AuthTypeBasicAuth,
				BasicAuthUser: "admin",
			},
			wantErr: "basicAuthPassword (secureJsonData) is required",
		},
		{
			name: "alternative token happy path",
			cfg: Config{
				URL:      "https://splunk.example.com:8089",
				AuthType: AuthTypeAlternativeToken,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAuthToken: "tok",
				},
			},
		},
		{
			name: "alternative token missing token",
			cfg: Config{
				URL:      "https://splunk.example.com:8089",
				AuthType: AuthTypeAlternativeToken,
			},
			wantErr: "authToken (secureJsonData) is required",
		},
		{
			name: "oauth forward needs no secret",
			cfg: Config{
				URL:      "https://splunk.example.com:8089",
				AuthType: AuthTypeOAuthForward,
			},
		},
		{
			name: "unknown auth type",
			cfg: Config{
				URL:      "https://splunk.example.com:8089",
				AuthType: "bogus",
			},
			wantErr: `unknown authType "bogus"`,
		},
		{
			name: "tls mutual auth happy path",
			cfg: Config{
				URL:           "https://splunk.example.com:8089",
				AuthType:      AuthTypeBasicAuth,
				BasicAuthUser: "admin",
				TLSAuth:       true,
				ServerName:    "splunk.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBasicAuthPassword: "pw",
					SecureJsonDataKeyTLSClientCert:     "cert",
					SecureJsonDataKeyTLSClientKey:      "key",
				},
			},
		},
		{
			name: "tls mutual auth missing key",
			cfg: Config{
				URL:           "https://splunk.example.com:8089",
				AuthType:      AuthTypeBasicAuth,
				BasicAuthUser: "admin",
				TLSAuth:       true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBasicAuthPassword: "pw",
					SecureJsonDataKeyTLSClientCert:     "cert",
				},
			},
			wantErr: "tlsClientKey (secureJsonData) is required",
		},
		{
			name: "custom CA missing cert",
			cfg: Config{
				URL:               "https://splunk.example.com:8089",
				AuthType:          AuthTypeOAuthForward,
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
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
