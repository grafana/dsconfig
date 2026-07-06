package tempodatasource

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
		name                 string
		example              string
		settings             backend.DataSourceInstanceSettings
		wantErr              error
		wantURL              string
		wantBasicAuth        bool
		wantBasicUser        string
		wantTLSAuth          bool
		wantTLSCA            bool
		wantOAuth            bool
		wantStreamingSearch  bool
		wantStreamingMetrics bool
		wantNodeGraph        bool
		wantSpanBarType      SpanBarType
		wantTimeRangeForTags int64
		wantTagLimit         int64
		wantServiceMapUID    string
		wantTracesToLogsUID  string
		wantSecureKeys       SecureJsonDataConfig
	}{
		{
			name:                 "default example loads with timeRangeForTags default applied",
			example:              "",
			wantURL:              "http://localhost:3200",
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name:                 "no auth with node graph",
			example:              "noAuth",
			wantURL:              "http://tempo.example.com:3200",
			wantNodeGraph:        true,
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name:                 "basic auth",
			example:              "basicAuth",
			wantURL:              "https://tempo.example.com",
			wantBasicAuth:        true,
			wantBasicUser:        "grafana",
			wantSecureKeys:       SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name:                 "oauth forward",
			example:              "oauthForward",
			wantURL:              "https://tempo.example.com",
			wantOAuth:            true,
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name:                 "tls mutual auth",
			example:              "tlsMutualAuth",
			wantURL:              "https://tempo.example.com",
			wantTLSAuth:          true,
			wantSecureKeys:       SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name:                 "tls self-signed CA",
			example:              "tlsSelfSignedCA",
			wantURL:              "https://tempo.internal.corp",
			wantTLSCA:            true,
			wantSecureKeys:       SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name:                 "streaming enabled search + metrics",
			example:              "streaming",
			wantURL:              "https://tempo.example.com",
			wantStreamingSearch:  true,
			wantStreamingMetrics: true,
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name:                 "full observability wiring",
			example:              "fullObservability",
			wantURL:              "https://tempo.example.com",
			wantNodeGraph:        true,
			wantSpanBarType:      SpanBarTypeDuration,
			wantServiceMapUID:    "prometheus",
			wantTracesToLogsUID:  "loki",
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name:                 "traceQL search + traceID query",
			example:              "traceQLSearchAndTraceID",
			wantURL:              "https://tempo.example.com",
			wantTagLimit:         10000,
			wantTimeRangeForTags: TimeRangeForTagsLast3Hours,
		},
		{
			name:                 "legacy tracesToLogs (v1)",
			example:              "legacyTracesToLogs",
			wantURL:              "https://tempo.example.com",
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{}`),
			},
			wantErr: errors.New("Tempo URL"),
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://tempo",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://tempo",
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
				URL:      "https://tempo",
				JSONData: []byte(`{"tlsAuth":true,"serverName":"tempo"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://tempo",
				JSONData: []byte(`{"tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{"timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "unknown spanBar.type errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{"spanBar":{"type":"WeirdValue"}}`),
			},
			wantErr: errors.New(`spanBar.type "WeirdValue"`),
		},
		{
			name: "spanBar Tag without tag key errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{"spanBar":{"type":"Tag"}}`),
			},
			wantErr: errors.New(`spanBar.tag is required when spanBar.type is "Tag"`),
		},
		{
			name: "invalid timeRangeForTags errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{"timeRangeForTags":999}`),
			},
			wantErr: errors.New("timeRangeForTags 999"),
		},
		{
			name: "tagLimit as numeric string parses",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{"tagLimit":"5000"}`),
			},
			wantURL:              "http://tempo",
			wantTagLimit:         5000,
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name: "tagLimit as number parses",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{"tagLimit":5000}`),
			},
			wantURL:              "http://tempo",
			wantTagLimit:         5000,
			wantTimeRangeForTags: TimeRangeForTagsLast30Minutes,
		},
		{
			name: "tagLimit malformed string errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{"tagLimit":"abc"}`),
			},
			wantErr: errors.New(`tagLimit`),
		},
		{
			name: "negative tagLimit errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://tempo",
				JSONData: []byte(`{"tagLimit":-1}`),
			},
			wantErr: errors.New("tagLimit must be non-negative"),
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
			if cfg.StreamingEnabled.Search != tt.wantStreamingSearch {
				t.Errorf("StreamingEnabled.Search = %v, want %v", cfg.StreamingEnabled.Search, tt.wantStreamingSearch)
			}
			if cfg.StreamingEnabled.Metrics != tt.wantStreamingMetrics {
				t.Errorf("StreamingEnabled.Metrics = %v, want %v", cfg.StreamingEnabled.Metrics, tt.wantStreamingMetrics)
			}
			if cfg.NodeGraph.Enabled != tt.wantNodeGraph {
				t.Errorf("NodeGraph.Enabled = %v, want %v", cfg.NodeGraph.Enabled, tt.wantNodeGraph)
			}
			if tt.wantSpanBarType != "" && cfg.SpanBar.Type != tt.wantSpanBarType {
				t.Errorf("SpanBar.Type = %q, want %q", cfg.SpanBar.Type, tt.wantSpanBarType)
			}
			if tt.wantTimeRangeForTags != 0 && cfg.TimeRangeForTags != tt.wantTimeRangeForTags {
				t.Errorf("TimeRangeForTags = %d, want %d", cfg.TimeRangeForTags, tt.wantTimeRangeForTags)
			}
			if tt.wantTagLimit != 0 && cfg.TagLimit != tt.wantTagLimit {
				t.Errorf("TagLimit = %d, want %d", cfg.TagLimit, tt.wantTagLimit)
			}
			if tt.wantServiceMapUID != "" && cfg.ServiceMap.DatasourceUID != tt.wantServiceMapUID {
				t.Errorf("ServiceMap.DatasourceUID = %q, want %q", cfg.ServiceMap.DatasourceUID, tt.wantServiceMapUID)
			}
			if tt.wantTracesToLogsUID != "" && cfg.TracesToLogsV2.DatasourceUID != tt.wantTracesToLogsUID {
				t.Errorf("TracesToLogsV2.DatasourceUID = %q, want %q", cfg.TracesToLogsV2.DatasourceUID, tt.wantTracesToLogsUID)
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
	t.Run("timeRangeForTags default is applied", func(t *testing.T) {
		var c Config
		c.ApplyDefaults()
		if c.TimeRangeForTags != TimeRangeForTagsLast30Minutes {
			t.Errorf("TimeRangeForTags = %d, want %d", c.TimeRangeForTags, TimeRangeForTagsLast30Minutes)
		}
	})
	t.Run("existing timeRangeForTags is preserved", func(t *testing.T) {
		c := Config{TimeRangeForTags: TimeRangeForTagsLast7Days}
		c.ApplyDefaults()
		if c.TimeRangeForTags != TimeRangeForTagsLast7Days {
			t.Errorf("TimeRangeForTags = %d, want %d", c.TimeRangeForTags, TimeRangeForTagsLast7Days)
		}
	})
	t.Run("no other field is defaulted", func(t *testing.T) {
		in := Config{
			URL:              "http://tempo",
			BasicAuth:        true,
			BasicAuthUser:    "grafana",
			TimeRangeForTags: TimeRangeForTagsLast30Minutes, // seed to isolate other fields
		}
		got := in
		got.ApplyDefaults()
		if !reflect.DeepEqual(in, got) {
			t.Errorf("ApplyDefaults mutated Config beyond timeRangeForTags: %#v -> %#v", in, got)
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
				URL:              "http://tempo",
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{TimeRangeForTags: TimeRangeForTagsLast30Minutes},
			wantErr: "Tempo URL (root.url) is required",
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:              "http://tempo",
				BasicAuth:        true,
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:              "https://tempo",
				TLSAuth:          true,
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://tempo",
				TLSAuth:    true,
				ServerName: "tempo",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://tempo",
				TLSAuthWithCACert: true,
				TimeRangeForTags:  TimeRangeForTagsLast30Minutes,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:              "http://tempo",
				Timeout:          -1,
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "spanBar Duration ok",
			cfg: Config{
				URL:              "http://tempo",
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
				SpanBar:          SpanBarConfig{Type: SpanBarTypeDuration},
			},
		},
		{
			name: "spanBar Tag with tag key ok",
			cfg: Config{
				URL:              "http://tempo",
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
				SpanBar:          SpanBarConfig{Type: SpanBarTypeTag, Tag: "environment"},
			},
		},
		{
			name: "spanBar Tag without tag key errors",
			cfg: Config{
				URL:              "http://tempo",
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
				SpanBar:          SpanBarConfig{Type: SpanBarTypeTag},
			},
			wantErr: `spanBar.tag is required when spanBar.type is "Tag"`,
		},
		{
			name: "spanBar unknown type errors",
			cfg: Config{
				URL:              "http://tempo",
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
				SpanBar:          SpanBarConfig{Type: "Whatever"},
			},
			wantErr: `spanBar.type "Whatever"`,
		},
		{
			name: "timeRangeForTags 3h ok",
			cfg: Config{
				URL:              "http://tempo",
				TimeRangeForTags: TimeRangeForTagsLast3Hours,
			},
		},
		{
			name: "timeRangeForTags 999 errors",
			cfg: Config{
				URL:              "http://tempo",
				TimeRangeForTags: 999,
			},
			wantErr: "timeRangeForTags 999",
		},
		{
			name: "negative tagLimit errors",
			cfg: Config{
				URL:              "http://tempo",
				TimeRangeForTags: TimeRangeForTagsLast30Minutes,
				TagLimit:         -5,
			},
			wantErr: "tagLimit must be non-negative",
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
