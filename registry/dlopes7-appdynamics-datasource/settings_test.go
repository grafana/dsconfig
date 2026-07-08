package appdynamicsdatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with optional root fields url/basicAuth/basicAuthUser,
// jsonData, and secureJsonData) into the backend.DataSourceInstanceSettings
// shape LoadConfig expects.
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
		name          string
		example       string
		settings      backend.DataSourceInstanceSettings
		useSettings   bool
		wantErr       error
		wantAuth      AuthMethod
		wantURL       string
		wantUser      string
		wantAnalytics bool
		wantSecure    SecureJsonDataConfig
	}{
		{
			// The default example has no url and an empty clientSecret
			// placeholder, so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (no url, no auth)",
			example: "",
			wantErr: errors.New("controller URL (root.url) is required"),
		},
		{
			name:       "api client",
			example:    "apiClient",
			wantAuth:   AuthMethodAPIClient,
			wantURL:    "https://controller.example.com",
			wantSecure: SecureJsonDataConfig{SecureJsonDataKeyClientSecret},
		},
		{
			name:       "basic auth",
			example:    "basicAuth",
			wantAuth:   AuthMethodBasic,
			wantURL:    "https://controller.example.com",
			wantUser:   "<your-username>",
			wantSecure: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:          "api client with analytics",
			example:       "apiClientWithAnalytics",
			wantAuth:      AuthMethodAPIClient,
			wantURL:       "https://controller.example.com",
			wantAnalytics: true,
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyClientSecret, SecureJsonDataKeyAnalyticsAPIKey},
		},
		{
			name:          "basic auth with analytics",
			example:       "basicAuthWithAnalytics",
			wantAuth:      AuthMethodBasic,
			wantURL:       "https://controller.example.com",
			wantUser:      "<your-username>",
			wantAnalytics: true,
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword, SecureJsonDataKeyAnalyticsAPIKey},
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
			name:        "api client missing domain errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://controller.example.com",
				JSONData:                []byte(`{"clientName":"my-client"}`),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "s"},
			},
			wantErr: errors.New("jsonData.clientDomain is required"),
		},
		{
			name:        "basic auth missing password errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:           "https://controller.example.com",
				BasicAuthUser: "admin@customer1",
				JSONData:      []byte(`{}`),
			},
			wantErr: errors.New("secureJsonData.basicAuthPassword is required"),
		},
		{
			name:        "missing url errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"clientName":"my-client","clientDomain":"customer1"}`),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "s"},
			},
			wantErr: errors.New("controller URL (root.url) is required"),
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

			if tt.wantAuth != "" && cfg.AuthMethod() != tt.wantAuth {
				t.Errorf("AuthMethod() = %q, want %q", cfg.AuthMethod(), tt.wantAuth)
			}
			if tt.wantURL != "" && cfg.MetricsURL != tt.wantURL {
				t.Errorf("MetricsURL = %q, want %q", cfg.MetricsURL, tt.wantURL)
			}
			if tt.wantUser != "" && cfg.BasicAuthUsername != tt.wantUser {
				t.Errorf("BasicAuthUsername = %q, want %q", cfg.BasicAuthUsername, tt.wantUser)
			}
			if cfg.IsAnalyticsConfigured() != tt.wantAnalytics {
				t.Errorf("IsAnalyticsConfigured() = %v, want %v", cfg.IsAnalyticsConfigured(), tt.wantAnalytics)
			}
			if tt.wantSecure != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if _, ok := cfg.DecryptedSecureJSONData[key]; ok {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantSecure) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantSecure)
				}
			}
		})
	}
}

// TestLoadConfigClientSecretOverridesBasicAuth verifies the auth gating that
// mirrors pkg/models/settings.go:46-51: when a clientSecret is present, the
// basic-auth fields are suppressed even if a basicAuthPassword secret and root
// basicAuthUser are also supplied, so API Client auth wins.
func TestLoadConfigClientSecretOverridesBasicAuth(t *testing.T) {
	settings := backend.DataSourceInstanceSettings{
		URL:           "https://controller.example.com",
		BasicAuthUser: "admin@customer1",
		JSONData:      []byte(`{"clientName":"my-client","clientDomain":"customer1"}`),
		DecryptedSecureJSONData: map[string]string{
			"clientSecret":      "the-secret",
			"basicAuthPassword": "the-password",
		},
	}
	cfg, err := LoadConfig(t.Context(), settings)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got := cfg.AuthMethod(); got != AuthMethodAPIClient {
		t.Errorf("AuthMethod() = %q, want %q", got, AuthMethodAPIClient)
	}
	if cfg.BasicAuthUsername != "" {
		t.Errorf("BasicAuthUsername = %q, want empty (suppressed by clientSecret)", cfg.BasicAuthUsername)
	}
	if cfg.BasicAuthPassword != "" {
		t.Errorf("BasicAuthPassword = %q, want empty (suppressed by clientSecret)", cfg.BasicAuthPassword)
	}
	if cfg.ClientSecret != "the-secret" {
		t.Errorf("ClientSecret = %q, want %q", cfg.ClientSecret, "the-secret")
	}
	// Both secrets remain enumerable in the decrypted map.
	if _, ok := cfg.DecryptedSecureJSONData[SecureJsonDataKeyBasicAuthPassword]; !ok {
		t.Errorf("expected basicAuthPassword to remain in DecryptedSecureJSONData")
	}
}

func TestApplyDefaults(t *testing.T) {
	// The AppDynamics editor writes no defaults, so ApplyDefaults is a no-op:
	// a config is unchanged by it.
	tests := []struct {
		name string
		in   Config
	}{
		{name: "empty config unchanged", in: Config{}},
		{
			name: "populated config unchanged",
			in: Config{
				TLSSkipVerify: true,
				ClientName:    "my-client",
				ClientDomain:  "customer1",
				MetricsURL:    "https://controller.example.com",
				ClientSecret:  "s",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if !reflect.DeepEqual(got, tt.in) {
				t.Errorf("ApplyDefaults mutated config: got %#v, want %#v (no-op)", got, tt.in)
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
			name: "api client happy path",
			cfg: Config{
				MetricsURL:   "https://controller.example.com",
				ClientName:   "my-client",
				ClientDomain: "customer1",
				ClientSecret: "s",
			},
		},
		{
			name: "basic auth happy path",
			cfg: Config{
				MetricsURL:        "https://controller.example.com",
				BasicAuthUsername: "admin@customer1",
				BasicAuthPassword: "p",
			},
		},
		{
			name:    "missing url",
			cfg:     Config{ClientName: "my-client", ClientDomain: "customer1", ClientSecret: "s"},
			wantErr: "controller URL (root.url) is required",
		},
		{
			name:    "no auth configured",
			cfg:     Config{MetricsURL: "https://controller.example.com"},
			wantErr: "no authentication configured",
		},
		{
			name: "api client missing clientName",
			cfg: Config{
				MetricsURL:   "https://controller.example.com",
				ClientDomain: "customer1",
				ClientSecret: "s",
			},
			wantErr: "jsonData.clientName is required for API Client auth",
		},
		{
			name: "api client missing clientDomain",
			cfg: Config{
				MetricsURL:   "https://controller.example.com",
				ClientName:   "my-client",
				ClientSecret: "s",
			},
			wantErr: "jsonData.clientDomain is required for API Client auth",
		},
		{
			name: "api client missing clientSecret",
			cfg: Config{
				MetricsURL:   "https://controller.example.com",
				ClientName:   "my-client",
				ClientDomain: "customer1",
			},
			wantErr: "secureJsonData.clientSecret is required for API Client auth",
		},
		{
			name: "basic auth missing username",
			cfg: Config{
				MetricsURL:        "https://controller.example.com",
				BasicAuthPassword: "p",
			},
			wantErr: "root.basicAuthUser is required for basic auth",
		},
		{
			name: "basic auth missing password",
			cfg: Config{
				MetricsURL:        "https://controller.example.com",
				BasicAuthUsername: "admin@customer1",
			},
			wantErr: "secureJsonData.basicAuthPassword is required for basic auth",
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

func TestIsAnalyticsConfigured(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{
			name: "all three set",
			cfg:  Config{AnalyticsURL: "https://analytics.api.appdynamics.com", AccountName: "acct", AnalyticsAPIKey: "k"},
			want: true,
		},
		{name: "missing key", cfg: Config{AnalyticsURL: "https://analytics.api.appdynamics.com", AccountName: "acct"}},
		{name: "missing account", cfg: Config{AnalyticsURL: "https://analytics.api.appdynamics.com", AnalyticsAPIKey: "k"}},
		{name: "missing url", cfg: Config{AccountName: "acct", AnalyticsAPIKey: "k"}},
		{name: "none set", cfg: Config{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsAnalyticsConfigured(); got != tt.want {
				t.Errorf("IsAnalyticsConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSettingsExamples guards the example set: the default "" example exists,
// and every example carries a jsonData object plus a non-empty secureJsonData
// using only known secret keys.
func TestSettingsExamples(t *testing.T) {
	examples := SettingsExamples().Examples
	if _, ok := examples[""]; !ok {
		t.Fatalf(`missing default "" example`)
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
			t.Errorf("example %q has no jsonData object", key)
		}
		secure, ok := value["secureJsonData"].(map[string]any)
		if !ok || len(secure) == 0 {
			t.Errorf("example %q has no secureJsonData", key)
			continue
		}
		for secretKey := range secure {
			if !known[secretKey] {
				t.Errorf("example %q references unknown secret key %q", key, secretKey)
			}
		}
	}
}
