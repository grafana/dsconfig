package wavefrontdatasource

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
		wantURL        string
		wantTimeout    int64
		wantSecureKeys SecureJsonDataConfig
		wantToken      string
	}{
		{
			// The default schema example has an empty token placeholder, so
			// LoadConfig's Validate step is expected to reject it.
			name:    "default example fails validation (empty token)",
			example: "",
			wantErr: errors.New("token"),
		},
		{
			name:           "api token against hosted cluster",
			example:        "apiToken",
			wantURL:        "https://try.wavefront.com",
			wantTimeout:    DefaultRequestTimeout,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
			wantToken:      "<your-wavefront-api-token>",
		},
		{
			name:           "self-managed cluster",
			example:        "selfManagedCluster",
			wantURL:        "https://mycluster.wavefront.com",
			wantTimeout:    DefaultRequestTimeout,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
			wantToken:      "<your-wavefront-api-token>",
		},
		{
			name:           "custom request timeout",
			example:        "customTimeout",
			wantURL:        "https://try.wavefront.com",
			wantTimeout:    60,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
			wantToken:      "<your-wavefront-api-token>",
		},
		{
			// Empty JSONData is a parse error upstream — pkg/models/settings.go:26-28
			// json.Unmarshal(nil, &settings) fails with "unexpected end of JSON input".
			name:        "empty settings error (empty JSONData)",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
			wantErr:     errors.New("parse jsonData"),
		},
		{
			name:        "invalid jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`invalid json`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name:        "missing url errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"requestTimeout":40}`),
				DecryptedSecureJSONData: map[string]string{"token": "tok"},
			},
			wantErr: errors.New("URL"),
		},
		{
			name:        "missing token errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"url":"https://foo.com"}`),
			},
			wantErr: errors.New("token"),
		},
		{
			name:        "absent request timeout defaults to 30",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://foo.com"}`),
				DecryptedSecureJSONData: map[string]string{"token": "tok"},
			},
			wantURL:     "https://foo.com",
			wantTimeout: DefaultRequestTimeout,
			wantToken:   "tok",
		},
		{
			name:        "null request timeout defaults to 30",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://foo.com","requestTimeout":null}`),
				DecryptedSecureJSONData: map[string]string{"token": "tok"},
			},
			wantURL:     "https://foo.com",
			wantTimeout: DefaultRequestTimeout,
			wantToken:   "tok",
		},
		{
			name:        "explicit request timeout is honored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://foo.com","requestTimeout":40}`),
				DecryptedSecureJSONData: map[string]string{"token": "tok"},
			},
			wantURL:     "https://foo.com",
			wantTimeout: 40,
			wantToken:   "tok",
		},
		{
			name:        "non-positive request timeout defaults to 30",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://foo.com","requestTimeout":0}`),
				DecryptedSecureJSONData: map[string]string{"token": "tok"},
			},
			wantURL:     "https://foo.com",
			wantTimeout: DefaultRequestTimeout,
			wantToken:   "tok",
		},
		{
			// enableSecureSocksProxy is intentionally omitted from both the
			// dsconfig schema and the Go Config struct; json unmarshal silently
			// ignores unknown fields.
			name:        "unknown enableSecureSocksProxy field is ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://foo.com","enableSecureSocksProxy":true}`),
				DecryptedSecureJSONData: map[string]string{"token": "tok"},
			},
			wantURL:     "https://foo.com",
			wantTimeout: DefaultRequestTimeout,
			wantToken:   "tok",
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
			if tt.wantTimeout != 0 && cfg.RequestTimeout != tt.wantTimeout {
				t.Errorf("RequestTimeout = %d, want %d", cfg.RequestTimeout, tt.wantTimeout)
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
			if tt.wantToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyToken] != tt.wantToken {
				t.Errorf("DecryptedSecureJSONData[token] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyToken], tt.wantToken)
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
			name: "zero timeout defaults to 30",
			in:   Config{},
			want: Config{RequestTimeout: DefaultRequestTimeout},
		},
		{
			name: "negative timeout defaults to 30",
			in:   Config{RequestTimeout: -5},
			want: Config{RequestTimeout: DefaultRequestTimeout},
		},
		{
			name: "explicit timeout is preserved",
			in:   Config{RequestTimeout: 45},
			want: Config{RequestTimeout: 45},
		},
		{
			name: "url is never defaulted",
			in:   Config{RequestTimeout: 10},
			want: Config{RequestTimeout: 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.RequestTimeout != tt.want.RequestTimeout {
				t.Errorf("RequestTimeout = %d, want %d", got.RequestTimeout, tt.want.RequestTimeout)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
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
				URL:                     "https://try.wavefront.com",
				RequestTimeout:          30,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "tok"},
			},
		},
		{
			name:    "empty url errors",
			cfg:     Config{DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "tok"}},
			wantErr: "URL",
		},
		{
			name:    "empty token errors",
			cfg:     Config{URL: "https://try.wavefront.com"},
			wantErr: "token",
		},
		{
			name:    "everything empty joins all errors",
			cfg:     Config{},
			wantErr: "URL",
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
