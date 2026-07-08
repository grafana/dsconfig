package jenkinsdatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full
// instance settings object with jsonData and secureJsonData) into the
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
		wantURL        string
		wantUsername   string
		wantSecureKeys SecureJsonDataConfig
		wantPassword   string
	}{
		{
			// The default schema example has an empty URL placeholder,
			// so LoadConfig's Validate step is expected to reject it.
			name:    "default example fails validation (empty URL)",
			example: "",
			wantErr: errors.New("jenkins URL"),
		},
		{
			name:           "anonymous access",
			example:        "anonymous",
			wantURL:        "https://jenkins.example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPassword:   "",
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://jenkins.example.com",
			wantUsername:   "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPassword:   "hunter2",
		},
		{
			name:           "api token",
			example:        "apiToken",
			wantURL:        "https://jenkins.example.com",
			wantUsername:   "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPassword:   "11a1b2c3d4e5f60718293a4b5c6d7e8f90",
		},
		{
			name:           "legacy username only",
			example:        "legacyUsernameOnly",
			wantURL:        "https://jenkins.example.com",
			wantUsername:   "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPassword:   "",
		},
		{
			// Empty JSONData is a parse error upstream —
			// pkg/plugin/settings.go unconditionally
			// json.Unmarshal(nil, cfg) which fails.
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
			name:        "missing url errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"username":"grafana"}`),
				DecryptedSecureJSONData: map[string]string{"password": "hunter2"},
			},
			wantErr: errors.New("jenkins URL"),
		},
		{
			name:        "url only (anonymous)",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"url":"https://jenkins.example.com"}`),
			},
			wantURL:      "https://jenkins.example.com",
			wantUsername: "",
		},
		{
			name:        "username and password parse",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://jenkins.example.com","username":"grafana"}`),
				DecryptedSecureJSONData: map[string]string{"password": "hunter2"},
			},
			wantURL:      "https://jenkins.example.com",
			wantUsername: "grafana",
			wantPassword: "hunter2",
		},
		{
			// enableSecureSocksProxy is intentionally omitted from both
			// the dsconfig schema and the Go Config struct; json
			// unmarshal silently ignores unknown fields.
			name:        "unknown enableSecureSocksProxy field is ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"url":"https://jenkins.example.com","enableSecureSocksProxy":true}`),
			},
			wantURL: "https://jenkins.example.com",
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

			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if cfg.Username != tt.wantUsername {
				t.Errorf("Username = %q, want %q", cfg.Username, tt.wantUsername)
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
			if cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword] != tt.wantPassword {
				t.Errorf("DecryptedSecureJSONData[password] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword], tt.wantPassword)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want Config
	}{
		{
			name: "no-op on empty config",
			in:   Config{},
			want: Config{},
		},
		{
			name: "populated config is untouched",
			in:   Config{URL: "https://jenkins.example.com", Username: "grafana"},
			want: Config{URL: "https://jenkins.example.com", Username: "grafana"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
			if got.Username != tt.want.Username {
				t.Errorf("Username = %q, want %q", got.Username, tt.want.Username)
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
			name: "happy path (basic auth)",
			cfg: Config{
				URL:      "https://jenkins.example.com",
				Username: "grafana",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyPassword: "hunter2",
				},
			},
		},
		{
			name: "anonymous is valid",
			cfg: Config{
				URL: "https://jenkins.example.com",
			},
		},
		{
			name: "password without username is valid at load time",
			cfg: Config{
				URL: "https://jenkins.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyPassword: "hunter2",
				},
			},
		},
		{
			name:    "empty url errors",
			cfg:     Config{Username: "grafana"},
			wantErr: "jenkins URL",
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
