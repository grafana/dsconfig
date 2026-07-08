package datadogdatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with optional root fields, jsonData, and secureJsonData) into
// the backend.DataSourceInstanceSettings shape LoadConfig expects.
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
		wantMode      PluginMode
		wantURL       string
		wantUser      string
		wantSize      int
		wantRateLimit float64
		wantLogLimits bool
		wantSecure    SecureJsonDataConfig
	}{
		{
			// The default example intentionally has empty apiKey/appKey
			// placeholders, so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (empty apiKey placeholder)",
			example: "",
			wantErr: errors.New("API key (secureJsonData.apiKey) is required"),
		},
		{
			name:       "default mode api + app key",
			example:    "directApiAppKey",
			wantMode:   PluginModeDefault,
			wantURL:    "https://api.datadoghq.com",
			wantSize:   100,
			wantSecure: SecureJsonDataConfig{SecureJsonDataKeyAPIKey, SecureJsonDataKeyAppKey},
		},
		{
			name:       "default mode EU region",
			example:    "directApiAppKeyEU",
			wantMode:   PluginModeDefault,
			wantURL:    "https://api.datadoghq.eu",
			wantSecure: SecureJsonDataConfig{SecureJsonDataKeyAPIKey, SecureJsonDataKeyAppKey},
		},
		{
			name:       "hosted metrics mode",
			example:    "hostedMetrics",
			wantMode:   PluginModeHostedMetrics,
			wantURL:    "https://dd-prod-10-prod-us-central-0.grafana.net/datadog",
			wantUser:   "123456",
			wantSecure: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:       "legacy hosted metrics without pluginMode resolves to hosted-metrics",
			example:    "legacyHostedMetricsNoPluginMode",
			wantMode:   PluginModeHostedMetrics,
			wantURL:    "https://dd-prod-10-prod-us-central-0.grafana.net/datadog",
			wantUser:   "123456",
			wantSecure: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:        "legacy api_key/app_key in jsonData migrate to secrets",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"api_key":"legacy-api-key","app_key":"legacy-app-key"}`),
			},
			wantMode:   PluginModeDefault,
			wantURL:    "https://api.datadoghq.com",
			wantSize:   100,
			wantSecure: SecureJsonDataConfig{SecureJsonDataKeyAPIKey, SecureJsonDataKeyAppKey},
		},
		{
			name:        "modern secrets win over legacy api_key/app_key",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"api_key":"legacy-api-key","app_key":"legacy-app-key"}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "modern-api-key", "appKey": "modern-app-key"},
			},
			wantMode:   PluginModeDefault,
			wantSecure: SecureJsonDataConfig{SecureJsonDataKeyAPIKey, SecureJsonDataKeyAppKey},
		},
		{
			name:        "quoted booleans parse leniently",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"logApiRateLimits":"true","disableDataLinks":"true"}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "k", "appKey": "k"},
			},
			wantMode:      PluginModeDefault,
			wantLogLimits: true,
		},
		{
			name:        "rateLimitMetrics coerces 0 to 100 when enabled",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"rateLimitEnabled":true}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "k", "appKey": "k"},
			},
			wantMode:      PluginModeDefault,
			wantRateLimit: 100,
		},
		{
			name:        "empty jsonData applies url and size defaults",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "k", "appKey": "k"},
			},
			wantMode: PluginModeDefault,
			wantURL:  "https://api.datadoghq.com",
			wantSize: 100,
		},
		{
			name:        "default mode missing appKey errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"pluginMode":"default"}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "k"},
			},
			wantErr: errors.New("App key (secureJsonData.appKey) is required"),
		},
		{
			name:        "hosted metrics missing password errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				BasicAuthEnabled: true,
				BasicAuthUser:    "123456",
				JSONData:         []byte(`{"pluginMode":"hosted-metrics","url":"https://dd.grafana.net/datadog"}`),
			},
			wantErr: errors.New("basic auth password (secureJsonData.basicAuthPassword)"),
		},
		{
			name:        "hosted metrics with default url errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				BasicAuthEnabled:        true,
				BasicAuthUser:           "123456",
				JSONData:                []byte(`{"pluginMode":"hosted-metrics"}`),
				DecryptedSecureJSONData: map[string]string{"basicAuthPassword": "p"},
			},
			wantErr: errors.New("non-default hosted metrics url"),
		},
		{
			name:        "invalid jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
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

			if tt.wantMode != "" && cfg.PluginMode != tt.wantMode {
				t.Errorf("PluginMode = %q, want %q", cfg.PluginMode, tt.wantMode)
			}
			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.wantUser != "" && cfg.BasicAuthUser != tt.wantUser {
				t.Errorf("BasicAuthUser = %q, want %q", cfg.BasicAuthUser, tt.wantUser)
			}
			if tt.wantSize != 0 && cfg.Size != tt.wantSize {
				t.Errorf("Size = %d, want %d", cfg.Size, tt.wantSize)
			}
			if tt.wantRateLimit != 0 && cfg.RateLimitMetrics != tt.wantRateLimit {
				t.Errorf("RateLimitMetrics = %v, want %v", cfg.RateLimitMetrics, tt.wantRateLimit)
			}
			if tt.wantLogLimits && !bool(cfg.LogAPIRateLimits) {
				t.Errorf("LogAPIRateLimits = %v, want true", bool(cfg.LogAPIRateLimits))
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

func TestLoadConfigMigratesLegacyKeyValues(t *testing.T) {
	settings := backend.DataSourceInstanceSettings{
		JSONData: []byte(`{"api_key":"legacy-api-key","app_key":"legacy-app-key"}`),
	}
	cfg, err := LoadConfig(t.Context(), settings)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got := cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey]; got != "legacy-api-key" {
		t.Errorf("apiKey = %q, want %q", got, "legacy-api-key")
	}
	if got := cfg.DecryptedSecureJSONData[SecureJsonDataKeyAppKey]; got != "legacy-app-key" {
		t.Errorf("appKey = %q, want %q", got, "legacy-app-key")
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want Config
	}{
		{
			name: "empty config gets default mode + url + size",
			in:   Config{},
			want: Config{PluginMode: PluginModeDefault, URL: DefaultDatadogAPIURL, Size: DefaultDatadogAPIResponseSize},
		},
		{
			name: "basicAuth without pluginMode resolves to hosted-metrics",
			in:   Config{BasicAuthEnabled: true},
			want: Config{BasicAuthEnabled: true, PluginMode: PluginModeHostedMetrics, URL: DefaultDatadogAPIURL, Size: DefaultDatadogAPIResponseSize},
		},
		{
			name: "explicit pluginMode is preserved",
			in:   Config{PluginMode: PluginModeHostedMetrics, URL: "https://dd.grafana.net/datadog", Size: 50},
			want: Config{PluginMode: PluginModeHostedMetrics, URL: "https://dd.grafana.net/datadog", Size: 50},
		},
		{
			name: "rateLimitMetrics defaults to 100 when enabled",
			in:   Config{RateLimitEnabled: true},
			want: Config{PluginMode: PluginModeDefault, URL: DefaultDatadogAPIURL, Size: DefaultDatadogAPIResponseSize, RateLimitEnabled: true, RateLimitMetrics: 100},
		},
		{
			name: "explicit rateLimitMetrics is preserved",
			in:   Config{RateLimitEnabled: true, RateLimitMetrics: 25},
			want: Config{PluginMode: PluginModeDefault, URL: DefaultDatadogAPIURL, Size: DefaultDatadogAPIResponseSize, RateLimitEnabled: true, RateLimitMetrics: 25},
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
			name: "default mode happy path",
			cfg: Config{
				PluginMode:              PluginModeDefault,
				URL:                     DefaultDatadogAPIURL,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "k", SecureJsonDataKeyAppKey: "k"},
			},
		},
		{
			name: "default mode missing apiKey",
			cfg: Config{
				PluginMode:              PluginModeDefault,
				URL:                     DefaultDatadogAPIURL,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAppKey: "k"},
			},
			wantErr: "API key (secureJsonData.apiKey) is required",
		},
		{
			name: "default mode missing appKey",
			cfg: Config{
				PluginMode:              PluginModeDefault,
				URL:                     DefaultDatadogAPIURL,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "k"},
			},
			wantErr: "App key (secureJsonData.appKey) is required",
		},
		{
			name: "default mode missing url",
			cfg: Config{
				PluginMode:              PluginModeDefault,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "k", SecureJsonDataKeyAppKey: "k"},
			},
			wantErr: "url (jsonData.url) is required",
		},
		{
			name: "hosted metrics happy path",
			cfg: Config{
				PluginMode:              PluginModeHostedMetrics,
				URL:                     "https://dd.grafana.net/datadog",
				BasicAuthUser:           "123456",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyBasicAuthPassword: "p"},
			},
		},
		{
			name: "hosted metrics with default url errors",
			cfg: Config{
				PluginMode:              PluginModeHostedMetrics,
				URL:                     DefaultDatadogAPIURL,
				BasicAuthUser:           "123456",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyBasicAuthPassword: "p"},
			},
			wantErr: "non-default hosted metrics url",
		},
		{
			name: "hosted metrics missing username",
			cfg: Config{
				PluginMode:              PluginModeHostedMetrics,
				URL:                     "https://dd.grafana.net/datadog",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyBasicAuthPassword: "p"},
			},
			wantErr: "basic auth username (root.basicAuthUser)",
		},
		{
			name: "hosted metrics missing password",
			cfg: Config{
				PluginMode:    PluginModeHostedMetrics,
				URL:           "https://dd.grafana.net/datadog",
				BasicAuthUser: "123456",
			},
			wantErr: "basic auth password (secureJsonData.basicAuthPassword)",
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
