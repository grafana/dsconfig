package csvdatasource

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
	if b, ok := value["basicAuth"].(bool); ok {
		settings.BasicAuthEnabled = b
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
		useSettings    bool
		wantErr        error
		wantURL        string
		wantStorage    StorageMode
		wantQueryParam string
		wantBasicAuth  bool
		wantBasicUser  string
		wantTLSAuth    bool
		wantTLSCA      bool
		wantOAuth      bool
		wantTimeout    float64
		wantCookies    []string
		wantSecureKeys SecureJsonDataConfig
	}{
		{
			name:        "default example loads",
			example:     "",
			wantURL:     "http://localhost:8080",
			wantStorage: StorageModeHTTP,
		},
		{
			name:           "http no auth with query params",
			example:        "httpNoAuth",
			wantURL:        "http://csv.example.com/data.csv",
			wantStorage:    StorageModeHTTP,
			wantQueryParam: "limit=100",
		},
		{
			name:           "http basic auth",
			example:        "httpBasicAuth",
			wantURL:        "https://csv.example.com/reports.csv",
			wantStorage:    StorageModeHTTP,
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:        "http oauth forward",
			example:     "httpOAuthForward",
			wantURL:     "https://csv.example.com/reports.csv",
			wantStorage: StorageModeHTTP,
			wantOAuth:   true,
		},
		{
			name:           "http tls mutual auth",
			example:        "httpTLSMutualAuth",
			wantURL:        "https://csv.example.com/reports.csv",
			wantStorage:    StorageModeHTTP,
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "http tls self-signed CA",
			example:        "httpTLSSelfSignedCA",
			wantURL:        "https://csv.internal.corp/reports.csv",
			wantStorage:    StorageModeHTTP,
			wantTLSCA:      true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:           "http advanced timeout + cookies + queryParams",
			example:        "httpAdvanced",
			wantURL:        "https://csv.example.com/reports.csv",
			wantStorage:    StorageModeHTTP,
			wantTimeout:    30,
			wantCookies:    []string{"session_id"},
			wantQueryParam: "format=csv&limit=1000",
		},
		{
			name:        "local filesystem storage",
			example:     "localFile",
			wantURL:     "/var/lib/csv-data",
			wantStorage: StorageModeLocal,
		},
		{
			name:        "legacy datasource with empty storage defaults to http",
			example:     "legacyEmptyStorage",
			wantURL:     "http://csv.legacy.example.com/data.csv",
			wantStorage: StorageModeHTTP,
		},
		{
			name:        "http storage missing URL errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"storage":"http"}`),
			},
			wantErr: errors.New("CSV URL (root.url) is required when storage is \"http\""),
		},
		{
			name:        "local storage missing URL errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"storage":"local"}`),
			},
			wantErr: errors.New("CSV file path (root.url) is required when storage is \"local\""),
		},
		{
			name:        "invalid storage value errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://csv",
				JSONData: []byte(`{"storage":"ftp"}`),
			},
			wantErr: errors.New(`invalid storage "ftp"`),
		},
		{
			name:        "basicAuth without user errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://csv",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{"storage":"http"}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name:        "tlsAuth without serverName errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://csv",
				JSONData: []byte(`{"storage":"http","tlsAuth":true}`),
				DecryptedSecureJSONData: map[string]string{
					"tlsClientCert": "pem",
					"tlsClientKey":  "pem",
				},
			},
			wantErr: errors.New("serverName (jsonData) is required"),
		},
		{
			name:        "tlsAuth without client cert errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://csv",
				JSONData: []byte(`{"storage":"http","tlsAuth":true,"serverName":"csv"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name:        "tlsAuthWithCACert without CA cert errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://csv",
				JSONData: []byte(`{"storage":"http","tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name:        "invalid jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://csv",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name:        "negative timeout errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://csv",
				JSONData: []byte(`{"storage":"http","timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name:        "empty jsonData applies http default and requires URL",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://csv",
				JSONData: []byte(`{}`),
			},
			wantURL:     "http://csv",
			wantStorage: StorageModeHTTP,
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
			if tt.wantStorage != "" && cfg.Storage != tt.wantStorage {
				t.Errorf("Storage = %q, want %q", cfg.Storage, tt.wantStorage)
			}
			if tt.wantQueryParam != "" && cfg.QueryParams != tt.wantQueryParam {
				t.Errorf("QueryParams = %q, want %q", cfg.QueryParams, tt.wantQueryParam)
			}
			if cfg.BasicAuth != tt.wantBasicAuth {
				t.Errorf("BasicAuth = %v, want %v", cfg.BasicAuth, tt.wantBasicAuth)
			}
			if tt.wantBasicUser != "" && cfg.BasicAuthUser != tt.wantBasicUser {
				t.Errorf("BasicAuthUser = %q, want %q", cfg.BasicAuthUser, tt.wantBasicUser)
			}
			if cfg.TLSAuth != tt.wantTLSAuth {
				t.Errorf("TLSAuth = %v, want %v", cfg.TLSAuth, tt.wantTLSAuth)
			}
			if cfg.TLSAuthWithCACert != tt.wantTLSCA {
				t.Errorf("TLSAuthWithCACert = %v, want %v", cfg.TLSAuthWithCACert, tt.wantTLSCA)
			}
			if cfg.OauthPassThru != tt.wantOAuth {
				t.Errorf("OauthPassThru = %v, want %v", cfg.OauthPassThru, tt.wantOAuth)
			}
			if tt.wantTimeout != 0 && cfg.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %v, want %v", cfg.Timeout, tt.wantTimeout)
			}
			if tt.wantCookies != nil && !reflect.DeepEqual(cfg.KeepCookies, tt.wantCookies) {
				t.Errorf("KeepCookies = %v, want %v", cfg.KeepCookies, tt.wantCookies)
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
		name string
		in   Config
		want Config
	}{
		{
			name: "empty storage defaults to http",
			in:   Config{URL: "http://csv"},
			want: Config{URL: "http://csv", Storage: DefaultStorageMode},
		},
		{
			name: "explicit http storage is preserved",
			in:   Config{URL: "http://csv", Storage: StorageModeHTTP},
			want: Config{URL: "http://csv", Storage: StorageModeHTTP},
		},
		{
			name: "explicit local storage is preserved",
			in:   Config{URL: "/var/lib/csv", Storage: StorageModeLocal},
			want: Config{URL: "/var/lib/csv", Storage: StorageModeLocal},
		},
		{
			name: "unknown storage is preserved (Validate rejects it)",
			in:   Config{URL: "x", Storage: StorageMode("ftp")},
			want: Config{URL: "x", Storage: StorageMode("ftp")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplyDefaults: got %#v, want %#v", got, tt.want)
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
			name: "minimal http happy path",
			cfg: Config{
				URL:     "http://csv",
				Storage: StorageModeHTTP,
			},
		},
		{
			name: "minimal local happy path",
			cfg: Config{
				URL:     "/var/lib/csv",
				Storage: StorageModeLocal,
			},
		},
		{
			name:    "missing URL for http",
			cfg:     Config{Storage: StorageModeHTTP},
			wantErr: "CSV URL (root.url) is required when storage is \"http\"",
		},
		{
			name:    "missing URL for local",
			cfg:     Config{Storage: StorageModeLocal},
			wantErr: "CSV file path (root.url) is required when storage is \"local\"",
		},
		{
			name:    "missing URL and empty storage errors as http",
			cfg:     Config{},
			wantErr: "CSV URL (root.url) is required when storage is \"http\"",
		},
		{
			name:    "invalid storage value",
			cfg:     Config{URL: "x", Storage: StorageMode("ftp")},
			wantErr: "invalid storage",
		},
		{
			name: "zero-value storage is accepted (defaults applied by ApplyDefaults)",
			cfg: Config{
				URL: "http://csv",
			},
			wantErr: "", // "" storage passes the switch; URL check treats empty as http.
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:       "http://csv",
				Storage:   StorageModeHTTP,
				BasicAuth: true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:     "https://csv",
				Storage: StorageModeHTTP,
				TLSAuth: true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://csv",
				Storage:    StorageModeHTTP,
				TLSAuth:    true,
				ServerName: "csv",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://csv",
				Storage:           StorageModeHTTP,
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "tlsAuthWithCACert with CA ok",
			cfg: Config{
				URL:               "https://csv",
				Storage:           StorageModeHTTP,
				TLSAuthWithCACert: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSCACert: "pem",
				},
			},
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:     "http://csv",
				Storage: StorageModeHTTP,
				Timeout: -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "zero timeout ok",
			cfg: Config{
				URL:     "http://csv",
				Storage: StorageModeHTTP,
				Timeout: 0,
			},
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
