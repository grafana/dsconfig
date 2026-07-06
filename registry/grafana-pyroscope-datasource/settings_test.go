package pyroscopedatasource

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
		wantErr        error
		wantURL        string
		wantBasicAuth  bool
		wantBasicUser  string
		wantTLSAuth    bool
		wantTLSCA      bool
		wantOAuth      bool
		wantMinStep    string
		wantSecureKeys SecureJsonDataConfig
	}{
		{
			name:    "default example loads",
			example: "",
			wantURL: "http://localhost:4040",
		},
		{
			name:        "no auth with 15s minStep",
			example:     "noAuth",
			wantURL:     "http://pyroscope.example.com:4040",
			wantMinStep: "15s",
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://pyroscope.example.com",
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:      "oauth forward",
			example:   "oauthForward",
			wantURL:   "https://pyroscope.example.com",
			wantOAuth: true,
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://pyroscope.example.com",
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://pyroscope.internal.corp",
			wantTLSCA:      true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:           "basic auth with 1m minStep",
			example:        "withMinStep",
			wantURL:        "https://pyroscope.example.com",
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantMinStep:    "1m",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name: "legacy plugin id (phlare) settings still load",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://phlare.example.com:4040",
				Type:     LegacyPluginID,
				JSONData: []byte(`{"minStep":"30s"}`),
			},
			wantURL:     "http://phlare.example.com:4040",
			wantMinStep: "30s",
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"minStep":"15s"}`),
			},
			wantErr: errors.New("Pyroscope URL"),
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://pyroscope",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://pyroscope",
				JSONData: []byte(`{"tlsAuth":true}`),
				DecryptedSecureJSONData: map[string]string{
					"tlsClientCert": "pem",
					"tlsClientKey":  "pem",
				},
			},
			wantErr: errors.New("serverName (jsonData) is required"),
		},
		{
			name: "tlsAuth without client cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://pyroscope",
				JSONData: []byte(`{"tlsAuth":true,"serverName":"pyroscope"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://pyroscope",
				JSONData: []byte(`{"tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://pyroscope",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://pyroscope",
				JSONData: []byte(`{"timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "invalid minStep errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://pyroscope",
				JSONData: []byte(`{"minStep":"5x"}`),
			},
			wantErr: errors.New(`minStep "5x"`),
		},
		{
			name: "empty minStep is accepted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://pyroscope",
				JSONData: []byte(`{}`),
			},
			wantURL: "http://pyroscope",
		},
		{
			name: "ms suffix accepted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://pyroscope",
				JSONData: []byte(`{"minStep":"500ms"}`),
			},
			wantURL:     "http://pyroscope",
			wantMinStep: "500ms",
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
			if tt.wantMinStep != "" && cfg.MinStep != tt.wantMinStep {
				t.Errorf("MinStep = %q, want %q", cfg.MinStep, tt.wantMinStep)
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
	// ApplyDefaults is intentionally a no-op for the Pyroscope entry — the
	// editor never persists a default value into jsonData on load (the "15s"
	// placeholder on Minimal step is a UI hint; the 15s fallback in query.go
	// is applied per query, never baked into stored settings). This test
	// guards that no field is silently defaulted.
	in := Config{
		URL:           "http://pyroscope",
		BasicAuth:     true,
		BasicAuthUser: "grafana",
	}
	got := in
	got.ApplyDefaults()
	if !reflect.DeepEqual(in, got) {
		t.Errorf("ApplyDefaults mutated Config: %#v -> %#v", in, got)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "minimal happy path",
			cfg: Config{
				URL: "http://pyroscope",
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{},
			wantErr: "Pyroscope URL (root.url) is required",
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:       "http://pyroscope",
				BasicAuth: true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:     "https://pyroscope",
				TLSAuth: true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://pyroscope",
				TLSAuth:    true,
				ServerName: "pyroscope",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://pyroscope",
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:     "http://pyroscope",
				Timeout: -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "minStep 15s ok",
			cfg: Config{
				URL:     "http://pyroscope",
				MinStep: "15s",
			},
		},
		{
			name: "minStep 500ms ok",
			cfg: Config{
				URL:     "http://pyroscope",
				MinStep: "500ms",
			},
		},
		{
			name: "minStep 1M (month) ok",
			cfg: Config{
				URL:     "http://pyroscope",
				MinStep: "1M",
			},
		},
		{
			name: "minStep with bogus unit errors",
			cfg: Config{
				URL:     "http://pyroscope",
				MinStep: "10x",
			},
			wantErr: `minStep "10x"`,
		},
		{
			name: "minStep with lowercase m (minutes) accepted",
			cfg: Config{
				URL:     "http://pyroscope",
				MinStep: "1m",
			},
		},
		{
			name: "minStep empty accepted",
			cfg: Config{
				URL: "http://pyroscope",
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
