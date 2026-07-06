package zipkindatasource

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
		name                  string
		example               string
		settings              backend.DataSourceInstanceSettings
		wantErr               error
		wantURL               string
		wantBasicAuth         bool
		wantBasicUser         string
		wantTLSAuth           bool
		wantTLSCA             bool
		wantOAuth             bool
		wantNodeGraph         bool
		wantSpanBarType       SpanBarType
		wantTracesToLogsUID   string
		wantTracesToMetricsID string
		wantSecureKeys        SecureJsonDataConfig
	}{
		{
			name:    "default example loads",
			example: "",
			wantURL: "http://localhost:9411",
		},
		{
			name:          "no auth with node graph",
			example:       "noAuth",
			wantURL:       "http://zipkin.example.com:9411",
			wantNodeGraph: true,
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://zipkin.example.com",
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:      "oauth forward",
			example:   "oauthForward",
			wantURL:   "https://zipkin.example.com",
			wantOAuth: true,
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://zipkin.example.com",
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://zipkin.internal.corp",
			wantTLSCA:      true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:                  "full observability wiring",
			example:               "fullObservability",
			wantURL:               "https://zipkin.example.com",
			wantNodeGraph:         true,
			wantSpanBarType:       SpanBarTypeDuration,
			wantTracesToLogsUID:   "loki",
			wantTracesToMetricsID: "prometheus",
		},
		{
			name:    "legacy tracesToLogs (v1)",
			example: "legacyTracesToLogs",
			wantURL: "https://zipkin.example.com",
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{}`),
			},
			wantErr: errors.New("Zipkin URL"),
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://zipkin",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://zipkin",
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
				URL:      "https://zipkin",
				JSONData: []byte(`{"tlsAuth":true,"serverName":"zipkin"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://zipkin",
				JSONData: []byte(`{"tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://zipkin",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://zipkin",
				JSONData: []byte(`{"timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "unknown spanBar.type errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://zipkin",
				JSONData: []byte(`{"spanBar":{"type":"WeirdValue"}}`),
			},
			wantErr: errors.New(`spanBar.type "WeirdValue"`),
		},
		{
			name: "spanBar Tag without tag key errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://zipkin",
				JSONData: []byte(`{"spanBar":{"type":"Tag"}}`),
			},
			wantErr: errors.New(`spanBar.tag is required when spanBar.type is "Tag"`),
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
			if cfg.NodeGraph.Enabled != tt.wantNodeGraph {
				t.Errorf("NodeGraph.Enabled = %v, want %v", cfg.NodeGraph.Enabled, tt.wantNodeGraph)
			}
			if tt.wantSpanBarType != "" && cfg.SpanBar.Type != tt.wantSpanBarType {
				t.Errorf("SpanBar.Type = %q, want %q", cfg.SpanBar.Type, tt.wantSpanBarType)
			}
			if tt.wantTracesToLogsUID != "" && cfg.TracesToLogsV2.DatasourceUID != tt.wantTracesToLogsUID {
				t.Errorf("TracesToLogsV2.DatasourceUID = %q, want %q", cfg.TracesToLogsV2.DatasourceUID, tt.wantTracesToLogsUID)
			}
			if tt.wantTracesToMetricsID != "" && cfg.TracesToMetrics.DatasourceUID != tt.wantTracesToMetricsID {
				t.Errorf("TracesToMetrics.DatasourceUID = %q, want %q", cfg.TracesToMetrics.DatasourceUID, tt.wantTracesToMetricsID)
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
	t.Run("no defaults are written", func(t *testing.T) {
		in := Config{
			URL:           "http://zipkin",
			BasicAuth:     true,
			BasicAuthUser: "grafana",
		}
		got := in
		got.ApplyDefaults()
		if !reflect.DeepEqual(in, got) {
			t.Errorf("ApplyDefaults mutated Config: %#v -> %#v", in, got)
		}
	})

	t.Run("empty Config stays empty", func(t *testing.T) {
		var in Config
		got := in
		got.ApplyDefaults()
		if !reflect.DeepEqual(in, got) {
			t.Errorf("ApplyDefaults mutated empty Config: %#v -> %#v", in, got)
		}
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "minimal happy path",
			cfg: Config{
				URL: "http://zipkin",
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{},
			wantErr: "Zipkin URL (root.url) is required",
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:       "http://zipkin",
				BasicAuth: true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:     "https://zipkin",
				TLSAuth: true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://zipkin",
				TLSAuth:    true,
				ServerName: "zipkin",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://zipkin",
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:     "http://zipkin",
				Timeout: -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "spanBar Duration ok",
			cfg: Config{
				URL:     "http://zipkin",
				SpanBar: SpanBarConfig{Type: SpanBarTypeDuration},
			},
		},
		{
			name: "spanBar Tag with tag key ok",
			cfg: Config{
				URL:     "http://zipkin",
				SpanBar: SpanBarConfig{Type: SpanBarTypeTag, Tag: "environment"},
			},
		},
		{
			name: "spanBar Tag without tag key errors",
			cfg: Config{
				URL:     "http://zipkin",
				SpanBar: SpanBarConfig{Type: SpanBarTypeTag},
			},
			wantErr: `spanBar.tag is required when spanBar.type is "Tag"`,
		},
		{
			name: "spanBar unknown type errors",
			cfg: Config{
				URL:     "http://zipkin",
				SpanBar: SpanBarConfig{Type: "Whatever"},
			},
			wantErr: `spanBar.type "Whatever"`,
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
