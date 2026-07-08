package splunkmonitoringdatasource

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
		wantRealm       string
		wantMetricsURL  string
		wantSignalflow  string
		wantSecureKeys  SecureJsonDataConfig
		wantAccessToken string
	}{
		{
			// The default schema example has an empty realm and an empty
			// accessToken placeholder, so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (empty realm and accessToken)",
			example: "",
			wantErr: errors.New("access token"),
		},
		{
			name:            "realm based",
			example:         "realm",
			wantRealm:       "us1",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "<your-access-token>",
		},
		{
			name:            "realm plus custom urls",
			example:         "customUrls",
			wantRealm:       "us1",
			wantMetricsURL:  "https://api.custom.example.com",
			wantSignalflow:  "https://stream.custom.example.com",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "<your-access-token>",
		},
		{
			name:            "custom urls without realm",
			example:         "customUrlsWithoutRealm",
			wantMetricsURL:  "https://api.custom.example.com",
			wantSignalflow:  "https://stream.custom.example.com",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "<your-access-token>",
		},
		{
			// Empty JSONData is a parse error upstream: pkg/models/settings.go:23
			// json.Unmarshal(nil, &settings) fails with "unexpected end of JSON
			// input".
			name:        "empty settings error (empty JSONData)",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
			wantErr:     errors.New("parse jsonData"),
		},
		{
			name:        "malformed jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name:        "missing access token errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"realmName":"us1"}`),
			},
			wantErr: errors.New("access token"),
		},
		{
			name:        "empty realm without custom urls errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantErr: errors.New("realm (jsonData.realmName) is required"),
		},
		{
			name:        "empty realm with only metrics url errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url_metrics_metadata":"https://api.custom.example.com"}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantErr: errors.New("realm (jsonData.realmName) is required"),
		},
		{
			name:        "empty realm with both custom urls is valid",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url_metrics_metadata":"https://api.custom.example.com","url_signalflow":"https://stream.custom.example.com"}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantMetricsURL:  "https://api.custom.example.com",
			wantSignalflow:  "https://stream.custom.example.com",
			wantAccessToken: "tok",
		},
		{
			name:        "realm set with token parses",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"realmName":"eu0"}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantRealm:       "eu0",
			wantAccessToken: "tok",
		},
		{
			// enableSecureSocksProxy is intentionally omitted from both the
			// dsconfig schema and the Go Config struct; json unmarshal silently
			// ignores unknown fields.
			name:        "unknown enableSecureSocksProxy field is ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"realmName":"us1","enableSecureSocksProxy":true}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantRealm:       "us1",
			wantAccessToken: "tok",
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

			if tt.wantRealm != "" && cfg.Realm != tt.wantRealm {
				t.Errorf("Realm = %q, want %q", cfg.Realm, tt.wantRealm)
			}
			if tt.wantMetricsURL != "" && cfg.URLMetricsMetaData != tt.wantMetricsURL {
				t.Errorf("URLMetricsMetaData = %q, want %q", cfg.URLMetricsMetaData, tt.wantMetricsURL)
			}
			if tt.wantSignalflow != "" && cfg.URLSignalFlow != tt.wantSignalflow {
				t.Errorf("URLSignalFlow = %q, want %q", cfg.URLSignalFlow, tt.wantSignalflow)
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
			if tt.wantAccessToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken] != tt.wantAccessToken {
				t.Errorf("DecryptedSecureJSONData[accessToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken], tt.wantAccessToken)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	// The plugin has no editor-parity defaults, so ApplyDefaults is a no-op:
	// it must never mutate a Config.
	tests := []struct {
		name string
		in   Config
	}{
		{name: "empty config unchanged", in: Config{}},
		{name: "realm preserved", in: Config{Realm: "us1"}},
		{
			name: "custom urls preserved",
			in: Config{
				URLMetricsMetaData: "https://api.custom.example.com",
				URLSignalFlow:      "https://stream.custom.example.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.Realm != tt.in.Realm {
				t.Errorf("Realm = %q, want %q (ApplyDefaults must not mutate)", got.Realm, tt.in.Realm)
			}
			if got.URLMetricsMetaData != tt.in.URLMetricsMetaData {
				t.Errorf("URLMetricsMetaData = %q, want %q (ApplyDefaults must not mutate)", got.URLMetricsMetaData, tt.in.URLMetricsMetaData)
			}
			if got.URLSignalFlow != tt.in.URLSignalFlow {
				t.Errorf("URLSignalFlow = %q, want %q (ApplyDefaults must not mutate)", got.URLSignalFlow, tt.in.URLSignalFlow)
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
			name: "realm and token",
			cfg: Config{
				Realm:                   "us1",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
		},
		{
			name: "both custom urls without realm",
			cfg: Config{
				URLMetricsMetaData:      "https://api.custom.example.com",
				URLSignalFlow:           "https://stream.custom.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
		},
		{
			name: "realm and custom urls",
			cfg: Config{
				Realm:                   "us1",
				URLMetricsMetaData:      "https://api.custom.example.com",
				URLSignalFlow:           "https://stream.custom.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
		},
		{
			name:    "missing token errors",
			cfg:     Config{Realm: "us1"},
			wantErr: "access token",
		},
		{
			name: "empty realm without urls errors",
			cfg: Config{
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
			wantErr: "realm",
		},
		{
			name: "empty realm with only metrics url errors",
			cfg: Config{
				URLMetricsMetaData:      "https://api.custom.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
			wantErr: "realm",
		},
		{
			name: "empty realm with only signalflow url errors",
			cfg: Config{
				URLSignalFlow:           "https://stream.custom.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
			wantErr: "realm",
		},
		{
			name:    "everything empty joins errors (token first)",
			cfg:     Config{},
			wantErr: "access token",
		},
		{
			name: "whitespace-only realm treated as empty",
			cfg: Config{
				Realm:                   "   ",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
			wantErr: "realm",
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
