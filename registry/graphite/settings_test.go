package graphitedatasource

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
		wantVersion      GraphiteVersion
		wantType         GraphiteType
		wantRollup       bool
		wantHasImportCfg bool
		wantSecureKeys   SecureJsonDataConfig
	}{
		{
			// The default example fills in URL and graphiteVersion, so it validates cleanly.
			name:        "default example loads",
			example:     "",
			wantURL:     "http://localhost:8080",
			wantVersion: DefaultGraphiteVersion,
		},
		{
			name:        "no auth",
			example:     "noAuth",
			wantURL:     "http://graphite.example.com:8080",
			wantVersion: GraphiteVersion11,
			wantType:    GraphiteTypeDefault,
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://graphite.example.com",
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantVersion:    GraphiteVersion11,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:        "oauth forward",
			example:     "oauthForward",
			wantURL:     "https://graphite.example.com",
			wantOAuth:   true,
			wantVersion: GraphiteVersion11,
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://graphite.example.com",
			wantTLSAuth:    true,
			wantVersion:    GraphiteVersion11,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://graphite.internal.corp",
			wantTLSCA:      true,
			wantVersion:    GraphiteVersion11,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:        "metrictank with rollup",
			example:     "metrictank",
			wantURL:     "https://metrictank.example.com",
			wantVersion: GraphiteVersion11,
			wantType:    GraphiteTypeMetrictank,
			wantRollup:  true,
		},
		{
			name:             "with label mappings",
			example:          "withLabelMappings",
			wantURL:          "http://localhost:8080",
			wantVersion:      GraphiteVersion11,
			wantHasImportCfg: true,
		},
		{
			name:        "legacy direct access",
			example:     "legacyDirectAccess",
			wantURL:     "http://graphite.example.com:8080",
			wantVersion: GraphiteVersion10,
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"graphiteVersion":"1.1"}`),
			},
			wantErr: errors.New("Graphite URL"),
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://localhost:8080",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{"graphiteVersion":"1.1"}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://graphite.example.com",
				JSONData: []byte(`{"graphiteVersion":"1.1","tlsAuth":true}`),
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
				URL:      "https://graphite.example.com",
				JSONData: []byte(`{"graphiteVersion":"1.1","tlsAuth":true,"serverName":"graphite"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://graphite.example.com",
				JSONData: []byte(`{"graphiteVersion":"1.1","tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8080",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8080",
				JSONData: []byte(`{"graphiteVersion":"1.1","timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "invalid graphiteVersion errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8080",
				JSONData: []byte(`{"graphiteVersion":"2.0"}`),
			},
			wantErr: errors.New(`invalid graphiteVersion "2.0"`),
		},
		{
			name: "invalid graphiteType errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8080",
				JSONData: []byte(`{"graphiteVersion":"1.1","graphiteType":"prometheus"}`),
			},
			wantErr: errors.New(`invalid graphiteType "prometheus"`),
		},
		{
			name: "empty graphiteVersion is defaulted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8080",
				JSONData: []byte(`{}`),
			},
			wantURL:     "http://localhost:8080",
			wantVersion: DefaultGraphiteVersion,
		},
		{
			name: "empty graphiteType is accepted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8080",
				JSONData: []byte(`{"graphiteVersion":"1.1"}`),
			},
			wantURL:     "http://localhost:8080",
			wantVersion: GraphiteVersion11,
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
			if tt.wantVersion != "" && cfg.GraphiteVersion != tt.wantVersion {
				t.Errorf("GraphiteVersion = %q, want %q", cfg.GraphiteVersion, tt.wantVersion)
			}
			if tt.wantType != "" && cfg.GraphiteType != tt.wantType {
				t.Errorf("GraphiteType = %q, want %q", cfg.GraphiteType, tt.wantType)
			}
			if cfg.RollupIndicatorEnabled != tt.wantRollup {
				t.Errorf("RollupIndicatorEnabled = %v, want %v", cfg.RollupIndicatorEnabled, tt.wantRollup)
			}
			if tt.wantHasImportCfg {
				if len(cfg.ImportConfiguration.Loki.Mappings) == 0 {
					t.Errorf("ImportConfiguration.Loki.Mappings = empty, want non-empty")
				}
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
			name: "empty graphiteVersion is defaulted to 1.1",
			in:   Config{URL: "http://localhost:8080"},
			want: Config{URL: "http://localhost:8080", GraphiteVersion: DefaultGraphiteVersion},
		},
		{
			name: "explicit graphiteVersion is preserved",
			in:   Config{URL: "http://localhost:8080", GraphiteVersion: GraphiteVersion09},
			want: Config{URL: "http://localhost:8080", GraphiteVersion: GraphiteVersion09},
		},
		{
			name: "graphiteType is NOT defaulted",
			in:   Config{URL: "http://localhost:8080"},
			want: Config{URL: "http://localhost:8080", GraphiteVersion: DefaultGraphiteVersion, GraphiteType: ""},
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
			name: "minimal happy path",
			cfg: Config{
				URL:             "http://localhost:8080",
				GraphiteVersion: GraphiteVersion11,
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{GraphiteVersion: GraphiteVersion11},
			wantErr: "Graphite URL (root.url) is required",
		},
		{
			name: "invalid graphiteVersion",
			cfg: Config{
				URL:             "http://localhost:8080",
				GraphiteVersion: "0.5",
			},
			wantErr: `invalid graphiteVersion "0.5"`,
		},
		{
			name: "empty graphiteVersion is accepted (defaults come from ApplyDefaults)",
			cfg: Config{
				URL: "http://localhost:8080",
			},
		},
		{
			name: "invalid graphiteType",
			cfg: Config{
				URL:             "http://localhost:8080",
				GraphiteVersion: GraphiteVersion11,
				GraphiteType:    "foo",
			},
			wantErr: `invalid graphiteType "foo"`,
		},
		{
			name: "empty graphiteType is accepted",
			cfg: Config{
				URL:             "http://localhost:8080",
				GraphiteVersion: GraphiteVersion11,
				GraphiteType:    "",
			},
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:             "http://localhost:8080",
				GraphiteVersion: GraphiteVersion11,
				BasicAuth:       true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:             "https://graphite",
				GraphiteVersion: GraphiteVersion11,
				TLSAuth:         true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:             "https://graphite",
				GraphiteVersion: GraphiteVersion11,
				TLSAuth:         true,
				ServerName:      "graphite",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://graphite",
				GraphiteVersion:   GraphiteVersion11,
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:             "http://localhost:8080",
				GraphiteVersion: GraphiteVersion11,
				Timeout:         -1,
			},
			wantErr: "timeout must be non-negative",
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
