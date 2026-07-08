package prometheusdatasource

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
		example          string // schema.go SettingsExamples key ("" = use inline settings)
		settings         backend.DataSourceInstanceSettings
		wantErr          error
		wantURL          string
		wantHTTPMethod   HTTPMethod
		wantBasicAuth    bool
		wantBasicUser    string
		wantTLSAuth      bool
		wantTLSCA        bool
		wantSecureKeys   SecureJsonDataConfig
		wantSamplesWarn  float64
		wantSamplesError float64
	}{
		{
			// The default example fills in URL, so it validates cleanly.
			name:           "default example loads",
			example:        "",
			wantURL:        "http://localhost:9090",
			wantHTTPMethod: HTTPMethodPOST,
		},
		{
			name:           "no auth",
			example:        "noAuth",
			wantURL:        "http://prometheus.example.com:9090",
			wantHTTPMethod: HTTPMethodPOST,
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://prometheus.example.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:           "oauth forward",
			example:        "oauthForward",
			wantURL:        "https://prometheus.example.com",
			wantHTTPMethod: HTTPMethodPOST,
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://prometheus.example.com",
			wantHTTPMethod: HTTPMethodPOST,
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://prometheus.internal.corp",
			wantHTTPMethod: HTTPMethodPOST,
			wantTLSCA:      true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:           "GET method",
			example:        "getHTTPMethod",
			wantURL:        "http://legacy-prom.example.com:9090",
			wantHTTPMethod: HTTPMethodGET,
		},
		{
			name:           "mimir with exemplars",
			example:        "mimirWithExemplars",
			wantURL:        "https://mimir.example.com/prometheus",
			wantHTTPMethod: HTTPMethodPOST,
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
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
				URL:      "http://localhost:9090",
				JSONData: []byte(`{"httpMethod":"PUT"}`),
			},
			wantErr: errors.New(`invalid httpMethod "PUT"`),
		},
		{
			name: "lowercase httpMethod normalises to uppercase",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9090",
				JSONData: []byte(`{"httpMethod":"get"}`),
			},
			wantURL:        "http://localhost:9090",
			wantHTTPMethod: HTTPMethodGET,
		},
		{
			name: "empty httpMethod defaults to POST",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9090",
				JSONData: []byte(`{}`),
			},
			wantURL:        "http://localhost:9090",
			wantHTTPMethod: HTTPMethodPOST,
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://localhost:9090",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prometheus.example.com",
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
				URL:      "https://prometheus.example.com",
				JSONData: []byte(`{"tlsAuth":true,"serverName":"prom"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://prometheus.example.com",
				JSONData: []byte(`{"tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9090",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9090",
				JSONData: []byte(`{"timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "negative seriesLimit errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9090",
				JSONData: []byte(`{"seriesLimit":-1}`),
			},
			wantErr: errors.New("seriesLimit must be non-negative"),
		},
		{
			name: "backend-only thresholds are parsed but do not require editor UI",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9090",
				JSONData: []byte(`{"maxSamplesProcessedWarningThreshold":100000000,"maxSamplesProcessedErrorThreshold":200000000}`),
			},
			wantURL:          "http://localhost:9090",
			wantHTTPMethod:   HTTPMethodPOST,
			wantSamplesWarn:  100000000,
			wantSamplesError: 200000000,
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
			if tt.wantSamplesWarn != 0 && cfg.MaxSamplesProcessedWarningThreshold != tt.wantSamplesWarn {
				t.Errorf("MaxSamplesProcessedWarningThreshold = %v, want %v", cfg.MaxSamplesProcessedWarningThreshold, tt.wantSamplesWarn)
			}
			if tt.wantSamplesError != 0 && cfg.MaxSamplesProcessedErrorThreshold != tt.wantSamplesError {
				t.Errorf("MaxSamplesProcessedErrorThreshold = %v, want %v", cfg.MaxSamplesProcessedErrorThreshold, tt.wantSamplesError)
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
			name: "empty httpMethod defaults to POST",
			in:   Config{},
			want: Config{HTTPMethod: HTTPMethodPOST},
		},
		{
			name: "existing POST preserved",
			in:   Config{HTTPMethod: HTTPMethodPOST},
			want: Config{HTTPMethod: HTTPMethodPOST},
		},
		{
			name: "GET preserved",
			in:   Config{HTTPMethod: HTTPMethodGET},
			want: Config{HTTPMethod: HTTPMethodGET},
		},
		{
			name: "lowercase get normalises to GET",
			in:   Config{HTTPMethod: "get"},
			want: Config{HTTPMethod: HTTPMethodGET},
		},
		{
			name: "whitespace stripped and uppercased",
			in:   Config{HTTPMethod: "  post  "},
			want: Config{HTTPMethod: HTTPMethodPOST},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.HTTPMethod != tt.want.HTTPMethod {
				t.Errorf("HTTPMethod = %q, want %q", got.HTTPMethod, tt.want.HTTPMethod)
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
			name: "minimal happy path",
			cfg: Config{
				URL:        "http://localhost:9090",
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
				URL:        "http://localhost:9090",
				HTTPMethod: "PUT",
			},
			wantErr: `invalid httpMethod "PUT"`,
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:        "http://localhost:9090",
				HTTPMethod: HTTPMethodPOST,
				BasicAuth:  true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:        "https://prom",
				HTTPMethod: HTTPMethodPOST,
				TLSAuth:    true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://prom",
				HTTPMethod: HTTPMethodPOST,
				TLSAuth:    true,
				ServerName: "prom",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://prom",
				HTTPMethod:        HTTPMethodPOST,
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:        "http://localhost:9090",
				HTTPMethod: HTTPMethodPOST,
				Timeout:    -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "negative seriesLimit",
			cfg: Config{
				URL:         "http://localhost:9090",
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
