package lokidatasource

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
		name             string
		example          string // schema.go SettingsExamples key ("" default example is loaded when non-empty)
		settings         backend.DataSourceInstanceSettings
		wantErr          error
		wantURL          string
		wantBasicAuth    bool
		wantBasicUser    string
		wantTLSAuth      bool
		wantTLSCA        bool
		wantOAuth        bool
		wantMaxLines     string
		wantDerivedCount int
		wantManageAlerts bool
		wantSecureKeys   SecureJsonDataConfig
	}{
		{
			// The default example fills in URL, so it validates cleanly.
			name:    "default example loads",
			example: "",
			wantURL: "http://localhost:3100",
		},
		{
			name:             "no auth",
			example:          "noAuth",
			wantURL:          "http://loki.example.com:3100",
			wantMaxLines:     "5000",
			wantManageAlerts: true,
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://loki.example.com",
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:      "oauth forward",
			example:   "oauthForward",
			wantURL:   "https://loki.example.com",
			wantOAuth: true,
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://loki.example.com",
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://loki.internal.corp",
			wantTLSCA:      true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:             "basic auth with derived fields",
			example:          "withDerivedFields",
			wantURL:          "https://loki.example.com",
			wantBasicAuth:    true,
			wantBasicUser:    "grafana",
			wantMaxLines:     "1500",
			wantDerivedCount: 2,
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"maxLines":"1000"}`),
			},
			wantErr: errors.New("Loki URL"),
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://localhost:3100",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://loki.example.com",
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
				URL:      "https://loki.example.com",
				JSONData: []byte(`{"tlsAuth":true,"serverName":"loki"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://loki.example.com",
				JSONData: []byte(`{"tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:3100",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:3100",
				JSONData: []byte(`{"timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "invalid derivedField matcherType errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:3100",
				JSONData: []byte(`{"derivedFields":[{"name":"x","matcherRegex":".*","matcherType":"exact"}]}`),
			},
			wantErr: errors.New(`derivedFields[0].matcherType "exact"`),
		},
		{
			name: "derivedField missing name errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:3100",
				JSONData: []byte(`{"derivedFields":[{"matcherRegex":".*"}]}`),
			},
			wantErr: errors.New("derivedFields[0].name is required"),
		},
		{
			name: "derivedField empty matcherType is accepted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:3100",
				JSONData: []byte(`{"derivedFields":[{"name":"x","matcherRegex":".*"}]}`),
			},
			wantURL:          "http://localhost:3100",
			wantDerivedCount: 1,
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
			if tt.wantMaxLines != "" && cfg.MaxLines != tt.wantMaxLines {
				t.Errorf("MaxLines = %q, want %q", cfg.MaxLines, tt.wantMaxLines)
			}
			if len(cfg.DerivedFields) != tt.wantDerivedCount {
				t.Errorf("len(DerivedFields) = %d, want %d", len(cfg.DerivedFields), tt.wantDerivedCount)
			}
			if cfg.ManageAlerts != tt.wantManageAlerts {
				t.Errorf("ManageAlerts = %v, want %v", cfg.ManageAlerts, tt.wantManageAlerts)
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
	// ApplyDefaults is intentionally a no-op for the Loki entry — the editor
	// never persists a default value into jsonData on load. This test guards
	// that no field is silently defaulted (regression protection if a future
	// author adds defaults without updating the schema).
	in := Config{
		URL:           "http://localhost:3100",
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
				URL: "http://localhost:3100",
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{},
			wantErr: "Loki URL (root.url) is required",
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:       "http://localhost:3100",
				BasicAuth: true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:     "https://loki",
				TLSAuth: true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://loki",
				TLSAuth:    true,
				ServerName: "loki",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://loki",
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:     "http://localhost:3100",
				Timeout: -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "derivedField happy path (regex)",
			cfg: Config{
				URL: "http://localhost:3100",
				DerivedFields: []DerivedFieldConfig{
					{Name: "TraceID", MatcherRegex: `trace=(\w+)`, MatcherType: DerivedFieldMatcherRegex, URL: "${__value.raw}"},
				},
			},
		},
		{
			name: "derivedField happy path (label)",
			cfg: Config{
				URL: "http://localhost:3100",
				DerivedFields: []DerivedFieldConfig{
					{Name: "svc", MatcherRegex: "service", MatcherType: DerivedFieldMatcherLabel},
				},
			},
		},
		{
			name: "derivedField missing name",
			cfg: Config{
				URL: "http://localhost:3100",
				DerivedFields: []DerivedFieldConfig{
					{MatcherRegex: ".*"},
				},
			},
			wantErr: "derivedFields[0].name is required",
		},
		{
			name: "derivedField missing regex",
			cfg: Config{
				URL: "http://localhost:3100",
				DerivedFields: []DerivedFieldConfig{
					{Name: "x"},
				},
			},
			wantErr: "derivedFields[0].matcherRegex is required",
		},
		{
			name: "derivedField invalid matcherType",
			cfg: Config{
				URL: "http://localhost:3100",
				DerivedFields: []DerivedFieldConfig{
					{Name: "x", MatcherRegex: ".*", MatcherType: "exact"},
				},
			},
			wantErr: `derivedFields[0].matcherType "exact"`,
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
