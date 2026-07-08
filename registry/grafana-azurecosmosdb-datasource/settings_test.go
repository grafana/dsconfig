package azurecosmosdbdatasource

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
		wantEndpoint   string
		wantSecureKeys SecureJsonDataConfig
		wantAccountKey string
	}{
		{
			// The default schema example has an empty accountEndpoint and
			// empty accountKey placeholder, so LoadConfig's Validate step
			// is expected to reject both.
			name:    "default example fails validation (empty endpoint and key)",
			example: "",
			wantErr: errors.New("account endpoint"),
		},
		{
			name:           "account key auth",
			example:        "accountKey",
			wantEndpoint:   "https://my-account.documents.azure.com:443/",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccountKey},
			wantAccountKey: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX==",
		},
		{
			// Empty JSONData is tolerated by upstream: the map[string]any
			// unmarshal path becomes a no-op, leaving AccountEndpoint
			// empty and letting Validate produce the endpoint-empty
			// error.
			name:        "empty settings error (empty JSONData)",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
			wantErr:     errors.New("account endpoint"),
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
			name:        "missing endpoint errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{"accountKey": "key"},
			},
			wantErr: errors.New("account endpoint"),
		},
		{
			name:        "missing account key errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"accountEndpoint":"https://a.documents.azure.com:443/"}`),
			},
			wantErr: errors.New("account key"),
		},
		{
			name:        "both missing joins both errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{}`),
			},
			wantErr: errors.New("account endpoint"),
		},
		{
			name:        "happy path from inline settings",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"accountEndpoint":"https://a.documents.azure.com:443/"}`),
				DecryptedSecureJSONData: map[string]string{"accountKey": "key"},
			},
			wantEndpoint:   "https://a.documents.azure.com:443/",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccountKey},
			wantAccountKey: "key",
		},
		{
			// enableSecureSocksProxy is intentionally omitted from both
			// the dsconfig schema and the Go Config struct; json
			// unmarshal silently ignores unknown fields.
			name:        "unknown enableSecureSocksProxy field is ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"accountEndpoint":"https://a.documents.azure.com:443/","enableSecureSocksProxy":true}`),
				DecryptedSecureJSONData: map[string]string{"accountKey": "key"},
			},
			wantEndpoint:   "https://a.documents.azure.com:443/",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccountKey},
			wantAccountKey: "key",
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

			if tt.wantEndpoint != "" && cfg.AccountEndpoint != tt.wantEndpoint {
				t.Errorf("AccountEndpoint = %q, want %q", cfg.AccountEndpoint, tt.wantEndpoint)
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
			if tt.wantAccountKey != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccountKey] != tt.wantAccountKey {
				t.Errorf("DecryptedSecureJSONData[accountKey] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccountKey], tt.wantAccountKey)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	// ApplyDefaults is a no-op — the plugin applies no editor-parity
	// defaults. This test guards against a future change silently
	// clobbering values.
	tests := []struct {
		name string
		in   Config
		want Config
	}{
		{
			name: "empty config unchanged",
			in:   Config{},
			want: Config{},
		},
		{
			name: "populated config unchanged",
			in: Config{
				AccountEndpoint:         "https://a.documents.azure.com:443/",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccountKey: "key"},
			},
			want: Config{
				AccountEndpoint:         "https://a.documents.azure.com:443/",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccountKey: "key"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AccountEndpoint != tt.want.AccountEndpoint {
				t.Errorf("AccountEndpoint = %q, want %q", got.AccountEndpoint, tt.want.AccountEndpoint)
			}
			if !reflect.DeepEqual(got.DecryptedSecureJSONData, tt.want.DecryptedSecureJSONData) {
				t.Errorf("DecryptedSecureJSONData = %v, want %v", got.DecryptedSecureJSONData, tt.want.DecryptedSecureJSONData)
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
			name: "happy path",
			cfg: Config{
				AccountEndpoint:         "https://a.documents.azure.com:443/",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccountKey: "key"},
			},
		},
		{
			name:    "empty endpoint errors",
			cfg:     Config{DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccountKey: "key"}},
			wantErr: "account endpoint",
		},
		{
			name:    "empty account key errors",
			cfg:     Config{AccountEndpoint: "https://a.documents.azure.com:443/"},
			wantErr: "account key",
		},
		{
			name:    "both empty joins errors",
			cfg:     Config{},
			wantErr: "account endpoint",
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
