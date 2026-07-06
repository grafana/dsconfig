package elasticsearchdatasource

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
	if s, ok := value["database"].(string); ok {
		settings.Database = s
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		example        string // schema.go SettingsExamples key ("" = use inline settings)
		settings       backend.DataSourceInstanceSettings
		wantErr        error
		wantURL        string
		wantIndex      string
		wantTimeField  string
		wantMaxShards  int64
		wantAuthAPIKey bool
		wantAuthSigV4  bool
		wantAuthOAuth  bool
		wantBasicAuth  bool
		wantBasicUser  string
		wantTLSAuth    bool
		wantTLSCA      bool
		wantMode       QueryType
		wantSecureKeys SecureJsonDataConfig
		wantDataLinks  int
	}{
		{
			// The default example fills in URL + index + timeField, so it
			// validates cleanly.
			name:          "default example loads",
			example:       "",
			wantURL:       "http://localhost:9200",
			wantIndex:     "es-index-name",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequests,
			wantMode:      DefaultQueryMode,
		},
		{
			name:          "no auth",
			example:       "noAuth",
			wantURL:       "http://elasticsearch.example.com:9200",
			wantIndex:     "[logstash-]YYYY.MM.DD",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequests,
			wantMode:      DefaultQueryMode,
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://elasticsearch.example.com:9200",
			wantIndex:      "grafana-logs",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequests,
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantMode:       DefaultQueryMode,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:           "api key",
			example:        "apiKey",
			wantURL:        "https://elasticsearch.example.com:9200",
			wantIndex:      "grafana-logs",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequests,
			wantAuthAPIKey: true,
			wantMode:       QueryTypeLogs,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
		},
		{
			name:          "sigv4",
			example:       "sigV4",
			wantURL:       "https://vpc-example.us-east-1.es.amazonaws.com",
			wantIndex:     "grafana-logs",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequests,
			wantAuthSigV4: true,
			wantMode:      DefaultQueryMode,
		},
		{
			name:          "oauth forward",
			example:       "oauthForward",
			wantURL:       "https://elasticsearch.example.com:9200",
			wantIndex:     "grafana-logs",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequests,
			wantAuthOAuth: true,
			wantMode:      DefaultQueryMode,
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://elasticsearch.example.com:9200",
			wantIndex:      "grafana-logs",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequests,
			wantTLSAuth:    true,
			wantMode:       DefaultQueryMode,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://elasticsearch.internal.corp:9200",
			wantIndex:      "grafana-logs",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequests,
			wantTLSCA:      true,
			wantMode:       DefaultQueryMode,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:           "logs with data links",
			example:        "logsWithDataLinks",
			wantURL:        "https://elasticsearch.example.com:9200",
			wantIndex:      "[app-logs-]YYYY.MM.DD",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequests,
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantMode:       QueryTypeLogs,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
			wantDataLinks:  2,
		},
		{
			// Legacy datasources with only root.database still resolve to a
			// working index via the fallback in LoadConfig.
			name:          "legacy database fallback",
			example:       "legacyDatabaseFallback",
			wantURL:       "http://elasticsearch.example.com:9200",
			wantIndex:     "grafana-legacy",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequests,
			wantMode:      DefaultQueryMode,
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"index":"i","timeField":"@timestamp"}`),
			},
			wantErr: errors.New("Elasticsearch URL"),
		},
		{
			name: "missing index errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"timeField":"@timestamp"}`),
			},
			wantErr: errors.New("index (jsonData.index) is required"),
		},
		{
			// timeField defaults via ApplyDefaults when missing, so a missing
			// timeField after ApplyDefaults is impossible; Validate still
			// guards against a caller invoking Validate directly.
			name: "timeField defaults to @timestamp when omitted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"index":"i"}`),
			},
			wantURL:       "http://localhost:9200",
			wantIndex:     "i",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequests,
			wantMode:      DefaultQueryMode,
		},
		{
			name: "empty jsonData errors on missing index (URL only)",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{}`),
			},
			wantErr: errors.New("index (jsonData.index) is required"),
		},
		{
			name: "invalid interval errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","interval":"bogus"}`),
			},
			wantErr: errors.New(`invalid interval "bogus"`),
		},
		{
			name: "invalid defaultQueryMode errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","defaultQueryMode":"bogus"}`),
			},
			wantErr: errors.New(`invalid defaultQueryMode "bogus"`),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "maxConcurrentShardRequests as a JSON string is parsed",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","maxConcurrentShardRequests":"7"}`),
			},
			wantURL:       "http://localhost:9200",
			wantIndex:     "i",
			wantTimeField: defaultTimeField,
			wantMaxShards: 7,
			wantMode:      DefaultQueryMode,
		},
		{
			name: "maxConcurrentShardRequests as an unparseable string defaults to 5",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","maxConcurrentShardRequests":"not-a-number"}`),
			},
			wantURL:       "http://localhost:9200",
			wantIndex:     "i",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequests,
			wantMode:      DefaultQueryMode,
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://localhost:9200",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{"index":"i","timeField":"@timestamp"}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "p",
				},
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "basicAuth without password errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://localhost:9200",
				BasicAuthEnabled: true,
				BasicAuthUser:    "user",
				JSONData:         []byte(`{"index":"i","timeField":"@timestamp"}`),
			},
			wantErr: errors.New("basicAuthPassword (secureJsonData) is required"),
		},
		{
			name: "apiKeyAuth without apiKey errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","apiKeyAuth":true}`),
			},
			wantErr: errors.New("apiKey (secureJsonData) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://elasticsearch.example.com",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","tlsAuth":true}`),
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
				URL:      "https://elasticsearch.example.com",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","tlsAuth":true,"serverName":"es"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://elasticsearch.example.com",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"index":"i","timeField":"@timestamp","timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
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
			if tt.wantIndex != "" && cfg.Index != tt.wantIndex {
				t.Errorf("Index = %q, want %q", cfg.Index, tt.wantIndex)
			}
			if tt.wantTimeField != "" && cfg.TimeField != tt.wantTimeField {
				t.Errorf("TimeField = %q, want %q", cfg.TimeField, tt.wantTimeField)
			}
			if tt.wantMaxShards != 0 && cfg.MaxConcurrentShardRequests != tt.wantMaxShards {
				t.Errorf("MaxConcurrentShardRequests = %d, want %d", cfg.MaxConcurrentShardRequests, tt.wantMaxShards)
			}
			if cfg.APIKeyAuth != tt.wantAuthAPIKey {
				t.Errorf("APIKeyAuth = %v, want %v", cfg.APIKeyAuth, tt.wantAuthAPIKey)
			}
			if cfg.SigV4Auth != tt.wantAuthSigV4 {
				t.Errorf("SigV4Auth = %v, want %v", cfg.SigV4Auth, tt.wantAuthSigV4)
			}
			if cfg.OauthPassThru != tt.wantAuthOAuth {
				t.Errorf("OauthPassThru = %v, want %v", cfg.OauthPassThru, tt.wantAuthOAuth)
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
			if tt.wantMode != "" && cfg.DefaultQueryMode != tt.wantMode {
				t.Errorf("DefaultQueryMode = %q, want %q", cfg.DefaultQueryMode, tt.wantMode)
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
			if tt.wantDataLinks != 0 && len(cfg.DataLinks) != tt.wantDataLinks {
				t.Errorf("len(DataLinks) = %d, want %d", len(cfg.DataLinks), tt.wantDataLinks)
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
			name: "empty config gets timeField, shards, and query mode",
			in:   Config{},
			want: Config{
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequests,
				DefaultQueryMode:           DefaultQueryMode,
			},
		},
		{
			name: "existing timeField preserved",
			in:   Config{TimeField: "@ts"},
			want: Config{
				TimeField:                  "@ts",
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequests,
				DefaultQueryMode:           DefaultQueryMode,
			},
		},
		{
			name: "existing positive shards preserved",
			in:   Config{MaxConcurrentShardRequests: 12},
			want: Config{
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: 12,
				DefaultQueryMode:           DefaultQueryMode,
			},
		},
		{
			name: "non-positive shards coerced to 5",
			in:   Config{MaxConcurrentShardRequests: -3},
			want: Config{
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequests,
				DefaultQueryMode:           DefaultQueryMode,
			},
		},
		{
			name: "existing query mode preserved",
			in:   Config{DefaultQueryMode: QueryTypeLogs},
			want: Config{
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequests,
				DefaultQueryMode:           QueryTypeLogs,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.TimeField != tt.want.TimeField {
				t.Errorf("TimeField = %q, want %q", got.TimeField, tt.want.TimeField)
			}
			if got.MaxConcurrentShardRequests != tt.want.MaxConcurrentShardRequests {
				t.Errorf("MaxConcurrentShardRequests = %d, want %d", got.MaxConcurrentShardRequests, tt.want.MaxConcurrentShardRequests)
			}
			if got.DefaultQueryMode != tt.want.DefaultQueryMode {
				t.Errorf("DefaultQueryMode = %q, want %q", got.DefaultQueryMode, tt.want.DefaultQueryMode)
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
				URL:                        "http://localhost:9200",
				Index:                      "i",
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequests,
			},
		},
		{
			name: "missing URL",
			cfg: Config{
				Index:     "i",
				TimeField: defaultTimeField,
			},
			wantErr: "Elasticsearch URL (root.url) is required",
		},
		{
			name: "missing index",
			cfg: Config{
				URL:       "http://localhost:9200",
				TimeField: defaultTimeField,
			},
			wantErr: "index (jsonData.index) is required",
		},
		{
			name: "missing timeField",
			cfg: Config{
				URL:   "http://localhost:9200",
				Index: "i",
			},
			wantErr: "timeField (jsonData.timeField) is required",
		},
		{
			name: "invalid interval",
			cfg: Config{
				URL:             "http://localhost:9200",
				Index:           "i",
				TimeField:       defaultTimeField,
				IntervalPattern: "Everly",
			},
			wantErr: `invalid interval "Everly"`,
		},
		{
			name: "invalid defaultQueryMode",
			cfg: Config{
				URL:              "http://localhost:9200",
				Index:            "i",
				TimeField:        defaultTimeField,
				DefaultQueryMode: "bogus",
			},
			wantErr: `invalid defaultQueryMode "bogus"`,
		},
		{
			name: "basicAuth needs user + password",
			cfg: Config{
				URL:       "http://localhost:9200",
				Index:     "i",
				TimeField: defaultTimeField,
				BasicAuth: true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "apiKeyAuth needs apiKey",
			cfg: Config{
				URL:        "http://localhost:9200",
				Index:      "i",
				TimeField:  defaultTimeField,
				APIKeyAuth: true,
			},
			wantErr: "apiKey (secureJsonData) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:       "https://elasticsearch",
				Index:     "i",
				TimeField: defaultTimeField,
				TLSAuth:   true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://elasticsearch",
				Index:      "i",
				TimeField:  defaultTimeField,
				TLSAuth:    true,
				ServerName: "es",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://elasticsearch",
				Index:             "i",
				TimeField:         defaultTimeField,
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:       "http://localhost:9200",
				Index:     "i",
				TimeField: defaultTimeField,
				Timeout:   -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "negative maxConcurrentShardRequests",
			cfg: Config{
				URL:                        "http://localhost:9200",
				Index:                      "i",
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: -1,
			},
			wantErr: "maxConcurrentShardRequests must be non-negative",
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
