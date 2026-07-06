package parca

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
			name:    "default example loads",
			example: "",
			wantURL: "http://localhost:7070",
		},
		{
			name:    "no auth",
			example: "noAuth",
			wantURL: "http://parca.example.com:7070",
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://parca.example.com",
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:      "oauth forward",
			example:   "oauthForward",
			wantURL:   "https://parca.example.com",
			wantOAuth: true,
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://parca.example.com",
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://parca.internal.corp",
			wantTLSCA:      true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:           "advanced http timeout + cookies",
			example:        "advancedHttp",
			wantURL:        "https://parca.example.com",
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantTimeout:    30,
			wantCookies:    []string{"session_id"},
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:        "missing URL errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{}`),
			},
			wantErr: errors.New("Parca URL"),
		},
		{
			name:        "basicAuth without user errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://parca",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name:        "tlsAuth without serverName errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://parca",
				JSONData: []byte(`{"tlsAuth":true}`),
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
				URL:      "https://parca",
				JSONData: []byte(`{"tlsAuth":true,"serverName":"parca"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name:        "tlsAuthWithCACert without CA cert errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://parca",
				JSONData: []byte(`{"tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name:        "invalid jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://parca",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name:        "negative timeout errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://parca",
				JSONData: []byte(`{"timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name:        "empty jsonData is accepted",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://parca",
				JSONData: []byte(`{}`),
			},
			wantURL: "http://parca",
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
	// ApplyDefaults is intentionally a no-op for the Parca entry — the
	// editor never persists a default value into jsonData on load (the
	// "Timeout in seconds" placeholder on Timeout is a UI hint). This test
	// guards that no field is silently defaulted.
	in := Config{
		URL:           "http://parca",
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
				URL: "http://parca",
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{},
			wantErr: "Parca URL (root.url) is required",
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:       "http://parca",
				BasicAuth: true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:     "https://parca",
				TLSAuth: true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://parca",
				TLSAuth:    true,
				ServerName: "parca",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://parca",
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "tlsAuthWithCACert with CA ok",
			cfg: Config{
				URL:               "https://parca",
				TLSAuthWithCACert: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSCACert: "pem",
				},
			},
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:     "http://parca",
				Timeout: -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "zero timeout ok",
			cfg: Config{
				URL:     "http://parca",
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
