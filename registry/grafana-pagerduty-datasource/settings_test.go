package pagerdutydatasource

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
		name           string
		example        string // schema.go SettingsExamples key
		settings       backend.DataSourceInstanceSettings
		useSettings    bool
		wantErr        error
		wantScheme     AuthSchemeID
		wantSecureKeys SecureJsonDataConfig
		wantAPIKey     string
	}{
		{
			// The default schema example intentionally has an empty API key
			// placeholder, so LoadConfig's Validate step is expected to reject
			// it.
			name:    "default example fails validation (empty api key placeholder)",
			example: "",
			wantErr: errors.New("API key"),
		},
		{
			name:           "api key example",
			example:        "apiKey",
			wantScheme:     AuthSchemeIDAPIKey,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
			wantAPIKey:     "<your-pagerduty-api-token>",
		},
		{
			// Empty settings: nil JSONData is treated as {} (mirroring
			// pkg/openapids/options.go:38-43), ApplyDefaults sets auth.id to
			// api_key, then Validate rejects the missing key.
			name:        "empty settings default to api_key and fail validation",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
			wantErr:     errors.New("API key"),
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
			name:        "api key via inline settings",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"auth":{"id":"api_key"}}`),
				DecryptedSecureJSONData: map[string]string{"auth.api_key.apiKey": "tok"},
			},
			wantScheme:     AuthSchemeIDAPIKey,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
			wantAPIKey:     "tok",
		},
		{
			name:        "unknown scheme errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"auth":{"id":"basic"}}`),
				DecryptedSecureJSONData: map[string]string{"auth.api_key.apiKey": "tok"},
			},
			wantErr: errors.New("unknown authentication scheme"),
		},
		{
			// The generic framework fields jsonData.servers and
			// jsonData.enableSecureSocksProxy are intentionally not modeled;
			// json unmarshal silently ignores them.
			name:        "unmodeled servers and enableSecureSocksProxy ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"auth":{"id":"api_key"},"servers":{"url":"https://x"},"enableSecureSocksProxy":true}`),
				DecryptedSecureJSONData: map[string]string{"auth.api_key.apiKey": "tok"},
			},
			wantScheme:     AuthSchemeIDAPIKey,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
			wantAPIKey:     "tok",
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

			if tt.wantScheme != "" && cfg.Auth.ID != tt.wantScheme {
				t.Errorf("Auth.ID = %q, want %q", cfg.Auth.ID, tt.wantScheme)
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
			if tt.wantAPIKey != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey] != tt.wantAPIKey {
				t.Errorf("DecryptedSecureJSONData[auth.api_key.apiKey] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey], tt.wantAPIKey)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want AuthSchemeID
	}{
		{
			name: "empty scheme defaults to api_key",
			in:   Config{},
			want: AuthSchemeIDAPIKey,
		},
		{
			name: "existing scheme is preserved",
			in:   Config{Auth: AuthConfig{ID: "basic"}},
			want: "basic",
		},
		{
			name: "api_key scheme is preserved",
			in:   Config{Auth: AuthConfig{ID: AuthSchemeIDAPIKey}},
			want: AuthSchemeIDAPIKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.Auth.ID != tt.want {
				t.Errorf("Auth.ID = %q, want %q", got.Auth.ID, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "api_key with key",
			cfg: Config{
				Auth:                    AuthConfig{ID: AuthSchemeIDAPIKey},
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "tok"},
			},
		},
		{
			name:    "api_key without key errors",
			cfg:     Config{Auth: AuthConfig{ID: AuthSchemeIDAPIKey}},
			wantErr: "API key",
		},
		{
			name:    "empty scheme errors",
			cfg:     Config{},
			wantErr: "authentication scheme (jsonData.auth.id) is required",
		},
		{
			name:    "unknown scheme errors",
			cfg:     Config{Auth: AuthConfig{ID: "oauth2"}},
			wantErr: `unknown authentication scheme "oauth2"`,
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
