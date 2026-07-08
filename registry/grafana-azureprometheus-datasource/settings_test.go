package azureprometheusdatasource

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
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		example        string
		settings       backend.DataSourceInstanceSettings
		wantErr        error
		wantURL        string
		wantHTTPMethod HTTPMethod
		wantAuthType   AuthType
		wantSecureKeys SecureJsonDataConfig
	}{
		{
			name:           "default example loads",
			example:        "",
			wantURL:        "https://<workspace>.prometheus.monitor.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
		},
		{
			name:           "client secret",
			example:        "clientSecret",
			wantURL:        "https://mimir.prometheus.monitor.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantAuthType:   AuthTypeClientSecret,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAzureClientSecret},
		},
		{
			name:           "managed identity",
			example:        "managedIdentity",
			wantURL:        "https://mimir.prometheus.monitor.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantAuthType:   AuthTypeManagedIdentity,
		},
		{
			name:           "workload identity",
			example:        "workloadIdentity",
			wantURL:        "https://mimir.prometheus.monitor.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantAuthType:   AuthTypeWorkloadIdentity,
		},
		{
			name:           "current user with fallback",
			example:        "currentUser",
			wantURL:        "https://mimir.prometheus.monitor.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantAuthType:   AuthTypeCurrentUser,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAzureClientSecret},
		},
		{
			name:           "custom endpoint resource ID",
			example:        "customEndpointResourceID",
			wantURL:        "https://mimir.prometheus.monitor.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantAuthType:   AuthTypeClientSecret,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAzureClientSecret},
		},
		{
			name:           "legacy client secret",
			example:        "legacyClientSecret",
			wantURL:        "https://mimir.prometheus.monitor.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantAuthType:   AuthTypeClientSecret,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyClientSecret},
		},
		{
			name:           "migrated from prometheus",
			example:        "migratedFromPrometheus",
			wantURL:        "https://mimir.prometheus.monitor.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantAuthType:   AuthTypeClientSecret,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAzureClientSecret},
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"httpMethod":"POST"}`),
			},
			wantErr: errors.New("Prometheus server URL"),
		},
		{
			name: "invalid httpMethod errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{"httpMethod":"PUT"}`),
			},
			wantErr: errors.New(`invalid httpMethod "PUT"`),
		},
		{
			name: "lowercase httpMethod normalises to uppercase",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{"httpMethod":"get"}`),
			},
			wantURL:        "https://prom.azure.com",
			wantHTTPMethod: HTTPMethodGET,
		},
		{
			name: "empty httpMethod defaults to POST",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{}`),
			},
			wantURL:        "https://prom.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
		},
		{
			name: "clientsecret without secret errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{"azureCredentials":{"authType":"clientsecret","tenantId":"t","clientId":"c"}}`),
			},
			wantErr: errors.New("authType \"clientsecret\" requires secureJsonData.azureClientSecret"),
		},
		{
			name: "clientsecret with legacy secret accepted",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://prom.azure.com",
				JSONData:                []byte(`{"azureCredentials":{"authType":"clientsecret","tenantId":"t","clientId":"c"}}`),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "old"},
			},
			wantURL:        "https://prom.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantAuthType:   AuthTypeClientSecret,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyClientSecret},
		},
		{
			name: "azureCredentials without authType errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{"azureCredentials":{"azureCloud":"AzureCloud"}}`),
			},
			wantErr: errors.New("jsonData.azureCredentials.authType is required"),
		},
		{
			name: "unknown authType errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{"azureCredentials":{"authType":"nonsense"}}`),
			},
			wantErr: errors.New(`unknown jsonData.azureCredentials.authType "nonsense"`),
		},
		{
			name: "no azureCredentials is allowed (backend leaves client unauth'd)",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{"httpMethod":"POST"}`),
			},
			wantURL:        "https://prom.azure.com",
			wantHTTPMethod: HTTPMethodPOST,
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{"timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "negative seriesLimit errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prom.azure.com",
				JSONData: []byte(`{"seriesLimit":-1}`),
			},
			wantErr: errors.New("seriesLimit must be non-negative"),
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

			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.wantHTTPMethod != "" && cfg.HTTPMethod != tt.wantHTTPMethod {
				t.Errorf("HTTPMethod = %q, want %q", cfg.HTTPMethod, tt.wantHTTPMethod)
			}
			if tt.wantAuthType != "" && cfg.EffectiveAuthType() != tt.wantAuthType {
				t.Errorf("EffectiveAuthType = %q, want %q", cfg.EffectiveAuthType(), tt.wantAuthType)
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
		want HTTPMethod
	}{
		{"empty httpMethod defaults to POST", Config{}, HTTPMethodPOST},
		{"existing POST preserved", Config{HTTPMethod: HTTPMethodPOST}, HTTPMethodPOST},
		{"GET preserved", Config{HTTPMethod: HTTPMethodGET}, HTTPMethodGET},
		{"lowercase get normalises to GET", Config{HTTPMethod: "get"}, HTTPMethodGET},
		{"whitespace stripped and uppercased", Config{HTTPMethod: "  post  "}, HTTPMethodPOST},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.HTTPMethod != tt.want {
				t.Errorf("HTTPMethod = %q, want %q", got.HTTPMethod, tt.want)
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
			name: "minimal happy path (no azure credentials)",
			cfg: Config{
				URL:        "https://prom.azure.com",
				HTTPMethod: HTTPMethodPOST,
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{HTTPMethod: HTTPMethodPOST},
			wantErr: "Prometheus server URL (root.url) is required",
		},
		{
			name: "invalid method",
			cfg: Config{
				URL:        "https://prom.azure.com",
				HTTPMethod: "PUT",
			},
			wantErr: `invalid httpMethod "PUT"`,
		},
		{
			name: "clientsecret without secret",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{"authType":"clientsecret","tenantId":"t","clientId":"c"}`),
			},
			wantErr: "authType \"clientsecret\" requires",
		},
		{
			name: "clientsecret with azureClientSecret",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{"authType":"clientsecret","tenantId":"t","clientId":"c"}`),
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAzureClientSecret: "secret",
				},
			},
		},
		{
			name: "clientsecret with legacy clientSecret fallback",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{"authType":"clientsecret","tenantId":"t","clientId":"c"}`),
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyClientSecret: "legacy",
				},
			},
		},
		{
			name: "msi requires no secret",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{"authType":"msi"}`),
			},
		},
		{
			name: "workloadidentity requires no secret",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{"authType":"workloadidentity"}`),
			},
		},
		{
			name: "currentuser requires no secret",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{"authType":"currentuser"}`),
			},
		},
		{
			name: "azureCredentials without authType",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{"azureCloud":"AzureCloud"}`),
			},
			wantErr: "jsonData.azureCredentials.authType is required",
		},
		{
			name: "unknown authType",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{"authType":"nonsense"}`),
			},
			wantErr: `unknown jsonData.azureCredentials.authType "nonsense"`,
		},
		{
			name: "invalid azureCredentials JSON",
			cfg: Config{
				URL:              "https://prom.azure.com",
				HTTPMethod:       HTTPMethodPOST,
				AzureCredentials: json.RawMessage(`{`),
			},
			wantErr: "jsonData.azureCredentials is not a valid credential object",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:        "https://prom.azure.com",
				HTTPMethod: HTTPMethodPOST,
				Timeout:    -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "negative seriesLimit",
			cfg: Config{
				URL:         "https://prom.azure.com",
				HTTPMethod:  HTTPMethodPOST,
				SeriesLimit: -5,
			},
			wantErr: "seriesLimit must be non-negative",
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

func TestEffectiveAuthType(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want AuthType
	}{
		{"empty credentials returns empty", ``, ""},
		{"clientsecret", `{"authType":"clientsecret"}`, AuthTypeClientSecret},
		{"msi", `{"authType":"msi"}`, AuthTypeManagedIdentity},
		{"invalid JSON returns empty", `{`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{}
			if tt.raw != "" {
				cfg.AzureCredentials = json.RawMessage(tt.raw)
			}
			got := cfg.EffectiveAuthType()
			if got != tt.want {
				t.Errorf("EffectiveAuthType = %q, want %q", got, tt.want)
			}
		})
	}
}
