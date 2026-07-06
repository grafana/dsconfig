package opensearchdatasource

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
	if s, ok := value["user"].(string); ok {
		settings.User = s
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
		wantDB         string
		wantFlavor     Flavor
		wantVersion    string
		wantTimeField  string
		wantMaxShards  int64
		wantAuthSigV4  bool
		wantAuthOAuth  bool
		wantBasicAuth  bool
		wantBasicUser  string
		wantTLSAuth    bool
		wantTLSCA      bool
		wantServerless bool
		wantSecureKeys SecureJsonDataConfig
		wantDataLinks  int
	}{
		{
			name:          "default example loads",
			example:       "",
			wantURL:       "http://localhost:9200",
			wantDB:        "es-index-name",
			wantFlavor:    FlavorOpenSearch,
			wantVersion:   "1.0.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequestsOpenSearch,
		},
		{
			name:          "no auth",
			example:       "noAuth",
			wantURL:       "http://opensearch.example.com:9200",
			wantDB:        "[logstash-]YYYY.MM.DD",
			wantFlavor:    FlavorOpenSearch,
			wantVersion:   "2.11.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequestsOpenSearch,
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantURL:        "https://opensearch.example.com:9200",
			wantDB:         "grafana-logs",
			wantFlavor:     FlavorOpenSearch,
			wantVersion:    "2.11.0",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequestsOpenSearch,
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:          "sigv4 managed",
			example:       "sigV4Managed",
			wantURL:       "https://vpc-example.us-east-1.es.amazonaws.com",
			wantDB:        "grafana-logs",
			wantFlavor:    FlavorOpenSearch,
			wantVersion:   "2.11.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequestsOpenSearch,
			wantAuthSigV4: true,
		},
		{
			name:           "sigv4 serverless",
			example:        "sigV4Serverless",
			wantURL:        "https://<collection>.us-east-1.aoss.amazonaws.com",
			wantDB:         "grafana-logs",
			wantFlavor:     FlavorOpenSearch,
			wantVersion:    "1.0.0",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequestsOpenSearch,
			wantAuthSigV4:  true,
			wantServerless: true,
		},
		{
			name:          "oauth forward",
			example:       "oauthForward",
			wantURL:       "https://opensearch.example.com:9200",
			wantDB:        "grafana-logs",
			wantFlavor:    FlavorOpenSearch,
			wantVersion:   "2.11.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequestsOpenSearch,
			wantAuthOAuth: true,
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://opensearch.example.com:9200",
			wantDB:         "grafana-logs",
			wantFlavor:     FlavorOpenSearch,
			wantVersion:    "2.11.0",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequestsOpenSearch,
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://opensearch.internal.corp:9200",
			wantDB:         "grafana-logs",
			wantFlavor:     FlavorOpenSearch,
			wantVersion:    "2.11.0",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequestsOpenSearch,
			wantTLSCA:      true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:          "elasticsearch legacy",
			example:       "elasticsearchLegacy",
			wantURL:       "https://es.internal.corp:9200",
			wantDB:        "logstash-*",
			wantFlavor:    FlavorElasticsearch,
			wantVersion:   "6.8.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequestsESLegacy,
		},
		{
			name:           "logs with data links",
			example:        "logsWithDataLinks",
			wantURL:        "https://opensearch.example.com:9200",
			wantDB:         "[app-logs-]YYYY.MM.DD",
			wantFlavor:     FlavorOpenSearch,
			wantVersion:    "2.11.0",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequestsOpenSearch,
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
			wantDataLinks:  2,
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp"}`),
			},
			wantErr: errors.New("OpenSearch URL"),
		},
		{
			name: "missing flavor errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"version":"1.0.0","database":"i","timeField":"@timestamp"}`),
			},
			wantErr: errors.New("flavor (jsonData.flavor) is required"),
		},
		{
			name: "invalid flavor errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"flavor":"bogus","version":"1.0.0","database":"i","timeField":"@timestamp"}`),
			},
			wantErr: errors.New(`invalid flavor "bogus"`),
		},
		{
			name: "missing version errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"flavor":"opensearch","database":"i","timeField":"@timestamp"}`),
			},
			wantErr: errors.New("version (jsonData.version) is required"),
		},
		{
			name: "timeField defaults to @timestamp when omitted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i"}`),
			},
			wantURL:       "http://localhost:9200",
			wantDB:        "i",
			wantFlavor:    FlavorOpenSearch,
			wantVersion:   "1.0.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequestsOpenSearch,
		},
		{
			name: "invalid interval errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp","interval":"bogus"}`),
			},
			wantErr: errors.New(`invalid interval "bogus"`),
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
				JSONData: []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp","maxConcurrentShardRequests":"7"}`),
			},
			wantURL:       "http://localhost:9200",
			wantDB:        "i",
			wantFlavor:    FlavorOpenSearch,
			wantVersion:   "1.0.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: 7,
		},
		{
			name: "maxConcurrentShardRequests defaults to 256 for Elasticsearch <7",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"flavor":"elasticsearch","version":"6.8.0","database":"i","timeField":"@timestamp"}`),
			},
			wantURL:       "http://localhost:9200",
			wantDB:        "i",
			wantFlavor:    FlavorElasticsearch,
			wantVersion:   "6.8.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequestsESLegacy,
		},
		{
			name: "maxConcurrentShardRequests defaults to 5 for Elasticsearch >=7",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"flavor":"elasticsearch","version":"7.10.0","database":"i","timeField":"@timestamp"}`),
			},
			wantURL:       "http://localhost:9200",
			wantDB:        "i",
			wantFlavor:    FlavorElasticsearch,
			wantVersion:   "7.10.0",
			wantTimeField: defaultTimeField,
			wantMaxShards: defaultMaxConcurrentShardRequestsOpenSearch,
		},
		{
			name: "serverless overrides flavor and version",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://opensearch.serverless.aws",
				JSONData: []byte(`{"serverless":true,"flavor":"elasticsearch","database":"i","timeField":"@timestamp"}`),
			},
			wantURL:        "https://opensearch.serverless.aws",
			wantDB:         "i",
			wantFlavor:     FlavorOpenSearch,
			wantVersion:    "1.0.0",
			wantTimeField:  defaultTimeField,
			wantMaxShards:  defaultMaxConcurrentShardRequestsOpenSearch,
			wantServerless: true,
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://localhost:9200",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp"}`),
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
				JSONData:         []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp"}`),
			},
			wantErr: errors.New("basicAuthPassword (secureJsonData) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://opensearch.example.com",
				JSONData: []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp","tlsAuth":true}`),
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
				URL:      "https://opensearch.example.com",
				JSONData: []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp","tlsAuth":true,"serverName":"es"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://opensearch.example.com",
				JSONData: []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp","tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:9200",
				JSONData: []byte(`{"flavor":"opensearch","version":"1.0.0","database":"i","timeField":"@timestamp","timeout":-5}`),
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
			if tt.wantDB != "" && cfg.JSONDatabase != tt.wantDB {
				t.Errorf("JSONDatabase = %q, want %q", cfg.JSONDatabase, tt.wantDB)
			}
			if tt.wantFlavor != "" && cfg.Flavor != tt.wantFlavor {
				t.Errorf("Flavor = %q, want %q", cfg.Flavor, tt.wantFlavor)
			}
			if tt.wantVersion != "" && cfg.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", cfg.Version, tt.wantVersion)
			}
			if tt.wantTimeField != "" && cfg.TimeField != tt.wantTimeField {
				t.Errorf("TimeField = %q, want %q", cfg.TimeField, tt.wantTimeField)
			}
			if tt.wantMaxShards != 0 && cfg.MaxConcurrentShardRequests != tt.wantMaxShards {
				t.Errorf("MaxConcurrentShardRequests = %d, want %d", cfg.MaxConcurrentShardRequests, tt.wantMaxShards)
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
			if cfg.Serverless != tt.wantServerless {
				t.Errorf("Serverless = %v, want %v", cfg.Serverless, tt.wantServerless)
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
			name: "empty config gets timeField and OpenSearch shards default",
			in:   Config{},
			want: Config{
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequestsOpenSearch,
			},
		},
		{
			name: "existing timeField preserved",
			in:   Config{TimeField: "@ts"},
			want: Config{
				TimeField:                  "@ts",
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequestsOpenSearch,
			},
		},
		{
			name: "elasticsearch legacy gets 256 shards default",
			in:   Config{Flavor: FlavorElasticsearch, Version: "6.8.0"},
			want: Config{
				Flavor:                     FlavorElasticsearch,
				Version:                    "6.8.0",
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequestsESLegacy,
			},
		},
		{
			name: "elasticsearch >=7 gets 5 shards default",
			in:   Config{Flavor: FlavorElasticsearch, Version: "7.10.0"},
			want: Config{
				Flavor:                     FlavorElasticsearch,
				Version:                    "7.10.0",
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequestsOpenSearch,
			},
		},
		{
			name: "serverless forces flavor + version + shards",
			in:   Config{Serverless: true, Flavor: FlavorElasticsearch, Version: "6.8.0"},
			want: Config{
				Serverless:                 true,
				Flavor:                     FlavorOpenSearch,
				Version:                    "6.8.0",
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequestsOpenSearch,
			},
		},
		{
			name: "existing positive shards preserved",
			in:   Config{MaxConcurrentShardRequests: 12},
			want: Config{
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: 12,
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
			if got.Flavor != tt.want.Flavor {
				t.Errorf("Flavor = %q, want %q", got.Flavor, tt.want.Flavor)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %q, want %q", got.Version, tt.want.Version)
			}
			if got.MaxConcurrentShardRequests != tt.want.MaxConcurrentShardRequests {
				t.Errorf("MaxConcurrentShardRequests = %d, want %d", got.MaxConcurrentShardRequests, tt.want.MaxConcurrentShardRequests)
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
				Flavor:                     FlavorOpenSearch,
				Version:                    "1.0.0",
				JSONDatabase:               "i",
				TimeField:                  defaultTimeField,
				MaxConcurrentShardRequests: defaultMaxConcurrentShardRequestsOpenSearch,
			},
		},
		{
			name: "missing URL",
			cfg: Config{
				Flavor:    FlavorOpenSearch,
				Version:   "1.0.0",
				TimeField: defaultTimeField,
			},
			wantErr: "OpenSearch URL (root.url) is required",
		},
		{
			name: "missing flavor",
			cfg: Config{
				URL:       "http://localhost:9200",
				Version:   "1.0.0",
				TimeField: defaultTimeField,
			},
			wantErr: "flavor (jsonData.flavor) is required",
		},
		{
			name: "missing version",
			cfg: Config{
				URL:       "http://localhost:9200",
				Flavor:    FlavorOpenSearch,
				TimeField: defaultTimeField,
			},
			wantErr: "version (jsonData.version) is required",
		},
		{
			name: "missing timeField",
			cfg: Config{
				URL:     "http://localhost:9200",
				Flavor:  FlavorOpenSearch,
				Version: "1.0.0",
			},
			wantErr: "timeField (jsonData.timeField) is required",
		},
		{
			name: "invalid interval",
			cfg: Config{
				URL:             "http://localhost:9200",
				Flavor:          FlavorOpenSearch,
				Version:         "1.0.0",
				TimeField:       defaultTimeField,
				IntervalPattern: "Everly",
			},
			wantErr: `invalid interval "Everly"`,
		},
		{
			name: "basicAuth needs user + password",
			cfg: Config{
				URL:       "http://localhost:9200",
				Flavor:    FlavorOpenSearch,
				Version:   "1.0.0",
				TimeField: defaultTimeField,
				BasicAuth: true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:       "https://opensearch",
				Flavor:    FlavorOpenSearch,
				Version:   "1.0.0",
				TimeField: defaultTimeField,
				TLSAuth:   true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://opensearch",
				Flavor:     FlavorOpenSearch,
				Version:    "1.0.0",
				TimeField:  defaultTimeField,
				TLSAuth:    true,
				ServerName: "opensearch.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://opensearch",
				Flavor:            FlavorOpenSearch,
				Version:           "1.0.0",
				TimeField:         defaultTimeField,
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:       "http://localhost:9200",
				Flavor:    FlavorOpenSearch,
				Version:   "1.0.0",
				TimeField: defaultTimeField,
				Timeout:   -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "negative maxConcurrentShardRequests",
			cfg: Config{
				URL:                        "http://localhost:9200",
				Flavor:                     FlavorOpenSearch,
				Version:                    "1.0.0",
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
