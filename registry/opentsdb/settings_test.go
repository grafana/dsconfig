package opentsdbdatasource

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
		name            string
		example         string // schema.go SettingsExamples key ("" default example is loaded when non-empty)
		settings        backend.DataSourceInstanceSettings
		wantErr         error
		wantURL         string
		wantBasicAuth   bool
		wantBasicUser   string
		wantTLSAuth     bool
		wantTLSCA       bool
		wantOAuth       bool
		wantTSDBVersion OpenTsdbVersion
		wantTSDBResolut OpenTsdbResolution
		wantLookupLimit int32
		wantSecureKeys  SecureJsonDataConfig
	}{
		{
			// The default example fills in URL + all three OpenTSDB fields, so it validates cleanly.
			name:            "default example loads",
			example:         "",
			wantURL:         "http://localhost:4242",
			wantTSDBVersion: DefaultOpenTsdbVersion,
			wantTSDBResolut: DefaultOpenTsdbResolution,
			wantLookupLimit: DefaultLookupLimit,
		},
		{
			name:            "no auth (v2.4)",
			example:         "noAuth",
			wantURL:         "http://opentsdb.example.com:4242",
			wantTSDBVersion: OpenTsdbVersion24,
			wantTSDBResolut: OpenTsdbResolutionSecond,
			wantLookupLimit: 1000,
		},
		{
			name:            "basic auth",
			example:         "basicAuth",
			wantURL:         "https://opentsdb.example.com",
			wantBasicAuth:   true,
			wantBasicUser:   "grafana",
			wantTSDBVersion: OpenTsdbVersionLTE21,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:            "oauth forward",
			example:         "oauthForward",
			wantURL:         "https://opentsdb.example.com",
			wantOAuth:       true,
			wantTSDBVersion: OpenTsdbVersionLTE21,
		},
		{
			name:            "tls mutual auth",
			example:         "tlsMutualAuth",
			wantURL:         "https://opentsdb.example.com",
			wantTLSAuth:     true,
			wantTSDBVersion: OpenTsdbVersionLTE21,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:            "tls self-signed CA",
			example:         "tlsSelfSignedCA",
			wantURL:         "https://opentsdb.internal.corp",
			wantTLSCA:       true,
			wantTSDBVersion: OpenTsdbVersionLTE21,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:            "millisecond resolution",
			example:         "millisecondResolution",
			wantURL:         "http://opentsdb.example.com:4242",
			wantTSDBVersion: OpenTsdbVersion23,
			wantTSDBResolut: OpenTsdbResolutionMillisecond,
			wantLookupLimit: 1000,
		},
		{
			name:            "large lookup limit",
			example:         "largeLookupLimit",
			wantURL:         "http://opentsdb.example.com:4242",
			wantTSDBVersion: OpenTsdbVersion24,
			wantLookupLimit: 10000,
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"tsdbVersion":1,"tsdbResolution":1,"lookupLimit":1000}`),
			},
			wantErr: errors.New("OpenTSDB URL"),
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://localhost:4242",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{"tsdbVersion":1}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://opentsdb.example.com",
				JSONData: []byte(`{"tsdbVersion":1,"tlsAuth":true}`),
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
				URL:      "https://opentsdb.example.com",
				JSONData: []byte(`{"tsdbVersion":1,"tlsAuth":true,"serverName":"opentsdb"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://opentsdb.example.com",
				JSONData: []byte(`{"tsdbVersion":1,"tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:4242",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:4242",
				JSONData: []byte(`{"tsdbVersion":1,"timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "invalid tsdbVersion errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:4242",
				JSONData: []byte(`{"tsdbVersion":5}`),
			},
			wantErr: errors.New("invalid tsdbVersion"),
		},
		{
			name: "invalid tsdbResolution errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:4242",
				JSONData: []byte(`{"tsdbVersion":1,"tsdbResolution":3}`),
			},
			wantErr: errors.New("invalid tsdbResolution"),
		},
		{
			name: "negative lookupLimit errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:4242",
				JSONData: []byte(`{"tsdbVersion":1,"lookupLimit":-100}`),
			},
			wantErr: errors.New("lookupLimit must be non-negative"),
		},
		{
			name: "empty jsonData is fully defaulted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:4242",
				JSONData: []byte(`{}`),
			},
			wantURL:         "http://localhost:4242",
			wantTSDBVersion: DefaultOpenTsdbVersion,
			wantTSDBResolut: DefaultOpenTsdbResolution,
			wantLookupLimit: DefaultLookupLimit,
		},
		{
			name: "zero tsdbVersion is defaulted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:4242",
				JSONData: []byte(`{"tsdbVersion":0,"tsdbResolution":2,"lookupLimit":500}`),
			},
			wantURL:         "http://localhost:4242",
			wantTSDBVersion: DefaultOpenTsdbVersion,
			wantTSDBResolut: OpenTsdbResolutionMillisecond,
			wantLookupLimit: 500,
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
			if tt.wantTSDBVersion != 0 && cfg.TSDBVersion != tt.wantTSDBVersion {
				t.Errorf("TSDBVersion = %v, want %v", cfg.TSDBVersion, tt.wantTSDBVersion)
			}
			if tt.wantTSDBResolut != 0 && cfg.TSDBResolution != tt.wantTSDBResolut {
				t.Errorf("TSDBResolution = %v, want %v", cfg.TSDBResolution, tt.wantTSDBResolut)
			}
			if tt.wantLookupLimit != 0 && cfg.LookupLimit != tt.wantLookupLimit {
				t.Errorf("LookupLimit = %v, want %v", cfg.LookupLimit, tt.wantLookupLimit)
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
			name: "empty tsdbVersion / resolution / lookupLimit are all defaulted",
			in:   Config{URL: "http://localhost:4242"},
			want: Config{
				URL:            "http://localhost:4242",
				TSDBVersion:    DefaultOpenTsdbVersion,
				TSDBResolution: DefaultOpenTsdbResolution,
				LookupLimit:    DefaultLookupLimit,
			},
		},
		{
			name: "explicit tsdbVersion is preserved",
			in:   Config{URL: "http://localhost:4242", TSDBVersion: OpenTsdbVersion24},
			want: Config{
				URL:            "http://localhost:4242",
				TSDBVersion:    OpenTsdbVersion24,
				TSDBResolution: DefaultOpenTsdbResolution,
				LookupLimit:    DefaultLookupLimit,
			},
		},
		{
			name: "explicit millisecond resolution is preserved",
			in:   Config{URL: "http://localhost:4242", TSDBResolution: OpenTsdbResolutionMillisecond},
			want: Config{
				URL:            "http://localhost:4242",
				TSDBVersion:    DefaultOpenTsdbVersion,
				TSDBResolution: OpenTsdbResolutionMillisecond,
				LookupLimit:    DefaultLookupLimit,
			},
		},
		{
			name: "explicit lookupLimit is preserved",
			in:   Config{URL: "http://localhost:4242", LookupLimit: 25000},
			want: Config{
				URL:            "http://localhost:4242",
				TSDBVersion:    DefaultOpenTsdbVersion,
				TSDBResolution: DefaultOpenTsdbResolution,
				LookupLimit:    25000,
			},
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
				URL:            "http://localhost:4242",
				TSDBVersion:    OpenTsdbVersionLTE21,
				TSDBResolution: OpenTsdbResolutionSecond,
				LookupLimit:    1000,
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{TSDBVersion: OpenTsdbVersionLTE21},
			wantErr: "OpenTSDB URL (root.url) is required",
		},
		{
			name: "invalid tsdbVersion",
			cfg: Config{
				URL:         "http://localhost:4242",
				TSDBVersion: OpenTsdbVersion(9),
			},
			wantErr: "invalid tsdbVersion",
		},
		{
			name: "zero tsdbVersion is accepted (defaults come from ApplyDefaults)",
			cfg: Config{
				URL: "http://localhost:4242",
			},
		},
		{
			name: "invalid tsdbResolution",
			cfg: Config{
				URL:            "http://localhost:4242",
				TSDBVersion:    OpenTsdbVersionLTE21,
				TSDBResolution: OpenTsdbResolution(3),
			},
			wantErr: "invalid tsdbResolution",
		},
		{
			name: "zero tsdbResolution is accepted",
			cfg: Config{
				URL:         "http://localhost:4242",
				TSDBVersion: OpenTsdbVersionLTE21,
			},
		},
		{
			name: "negative lookupLimit errors",
			cfg: Config{
				URL:         "http://localhost:4242",
				TSDBVersion: OpenTsdbVersionLTE21,
				LookupLimit: -5,
			},
			wantErr: "lookupLimit must be non-negative",
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:         "http://localhost:4242",
				TSDBVersion: OpenTsdbVersionLTE21,
				BasicAuth:   true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:         "https://opentsdb",
				TSDBVersion: OpenTsdbVersionLTE21,
				TLSAuth:     true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:         "https://opentsdb",
				TSDBVersion: OpenTsdbVersionLTE21,
				TLSAuth:     true,
				ServerName:  "opentsdb",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://opentsdb",
				TSDBVersion:       OpenTsdbVersionLTE21,
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:         "http://localhost:4242",
				TSDBVersion: OpenTsdbVersionLTE21,
				Timeout:     -1,
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
