package clickhousedatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with jsonData and secureJsonData) into the
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
	return backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: secure,
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name                 string
		example              string
		settings             backend.DataSourceInstanceSettings
		wantErr              string
		wantHost             string
		wantPort             int64
		wantProtocol         Protocol
		wantSecure           bool
		wantConfigMode       ConfigMode
		wantSignalType       SignalType
		wantSecureKeys       []SecureJsonDataKey
		wantPassword         string
		wantSecureHTTPHeader map[string]string
		wantDialTimeout      string
		wantQueryTimeout     string
		wantLogsDefaultTable string
		wantTraceDefTable    string
	}{
		{
			name:    "default example fails validation (empty host/port)",
			example: "",
			wantErr: "host (jsonData.host) is required",
		},
		{
			name:                 "native insecure example",
			example:              "nativeInsecure",
			wantHost:             "localhost",
			wantPort:             9000,
			wantProtocol:         ProtocolNative,
			wantConfigMode:       ConfigModeClassic,
			wantSecureKeys:       []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantPassword:         "changeme",
			wantDialTimeout:      "10",
			wantQueryTimeout:     "60",
			wantLogsDefaultTable: "otel_logs",
			wantTraceDefTable:    "otel_traces",
		},
		{
			name:           "native TLS example",
			example:        "nativeSecure",
			wantHost:       "my-cluster.clickhouse.cloud",
			wantPort:       9440,
			wantProtocol:   ProtocolNative,
			wantSecure:     true,
			wantConfigMode: ConfigModeClassic,
			wantSecureKeys: []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantPassword:   "changeme",
		},
		{
			name:                 "http with headers, incl. secureHttpHeaders",
			example:              "httpSecureWithHeaders",
			wantHost:             "clickhouse.internal",
			wantPort:             8443,
			wantProtocol:         ProtocolHTTP,
			wantSecure:           true,
			wantConfigMode:       ConfigModeClassic,
			wantSecureKeys:       []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantPassword:         "changeme",
			wantSecureHTTPHeader: map[string]string{"X-Api-Key": "abcd1234"},
		},
		{
			name:           "mTLS client cert example",
			example:        "tlsClientAuth",
			wantHost:       "clickhouse.internal",
			wantPort:       9440,
			wantProtocol:   ProtocolNative,
			wantSecure:     true,
			wantConfigMode: ConfigModeClassic,
			wantSecureKeys: []SecureJsonDataKey{
				SecureJsonDataKeyPassword,
				SecureJsonDataKeyTLSClientCert,
				SecureJsonDataKeyTLSClientKey,
			},
		},
		{
			name:           "TLS with CA cert example",
			example:        "tlsWithCACert",
			wantHost:       "clickhouse.internal",
			wantPort:       9440,
			wantProtocol:   ProtocolNative,
			wantSecure:     true,
			wantConfigMode: ConfigModeClassic,
			wantSecureKeys: []SecureJsonDataKey{
				SecureJsonDataKeyPassword,
				SecureJsonDataKeyTLSCACert,
			},
		},
		{
			name:                 "otel single-table logs example",
			example:              "otelLogsSingleTable",
			wantHost:             "clickhouse.internal",
			wantPort:             9440,
			wantProtocol:         ProtocolNative,
			wantSecure:           true,
			wantConfigMode:       ConfigModeSingleTable,
			wantSignalType:       SignalTypeLogs,
			wantSecureKeys:       []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantLogsDefaultTable: "otel_logs",
			wantTraceDefTable:    "otel_traces",
		},
		{
			name:              "otel single-table traces example",
			example:           "otelTracesSingleTable",
			wantHost:          "clickhouse.internal",
			wantPort:          9440,
			wantProtocol:      ProtocolNative,
			wantSecure:        true,
			wantConfigMode:    ConfigModeSingleTable,
			wantSignalType:    SignalTypeTraces,
			wantSecureKeys:    []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantTraceDefTable: "otel_traces",
		},
		{
			// Legacy example: `server` and numeric `timeout` are migrated to
			// host and dialTimeout by the custom UnmarshalJSON.
			name:            "legacy v3 server / timeout normalizes",
			example:         "legacyV3ServerField",
			wantHost:        "legacy-clickhouse.internal",
			wantPort:        9000,
			wantProtocol:    ProtocolNative,
			wantConfigMode:  ConfigModeClassic,
			wantSecureKeys:  []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantDialTimeout: "10",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "port as string is accepted",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"host":"h","port":"9000","protocol":"native"}`),
				DecryptedSecureJSONData: map[string]string{"password": "p"},
			},
			wantHost:     "h",
			wantPort:     9000,
			wantProtocol: ProtocolNative,
		},
		{
			name: "unknown protocol errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"host":"h","port":9000,"protocol":"grpc"}`),
			},
			wantErr: `unknown protocol "grpc"`,
		},
		{
			name: "single-table without signalType errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"host":"h","port":9000,"protocol":"native","configMode":"single-table"}`),
			},
			wantErr: "signalType is required when configMode is single-table",
		},
		{
			name: "tlsAuth without client cert errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"host":"h","port":9000,"protocol":"native","tlsAuth":true}`),
			},
			wantErr: "tlsClientCert (secureJsonData) is required when tlsAuth is true",
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"host":"h","port":9000,"protocol":"native","tlsAuthWithCACert":true}`),
			},
			wantErr: "tlsCACert (secureJsonData) is required when tlsAuthWithCACert is true",
		},
		{
			name:     "empty settings default to native + classic and fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  "host (jsonData.host) is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" || (tt.settings.JSONData == nil && tt.settings.DecryptedSecureJSONData == nil && tt.wantErr == "") {
				if _, ok := SettingsExamples().Examples[tt.example]; ok {
					settings = settingsFromExample(t, tt.example)
				}
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("LoadConfig: expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("LoadConfig: error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}

			if tt.wantHost != "" && cfg.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", cfg.Host, tt.wantHost)
			}
			if tt.wantPort != 0 && cfg.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", cfg.Port, tt.wantPort)
			}
			if tt.wantProtocol != "" && cfg.Protocol != tt.wantProtocol {
				t.Errorf("Protocol = %q, want %q", cfg.Protocol, tt.wantProtocol)
			}
			if cfg.Secure != tt.wantSecure {
				t.Errorf("Secure = %v, want %v", cfg.Secure, tt.wantSecure)
			}
			if tt.wantConfigMode != "" && cfg.ConfigMode != tt.wantConfigMode {
				t.Errorf("ConfigMode = %q, want %q", cfg.ConfigMode, tt.wantConfigMode)
			}
			if tt.wantSignalType != "" && cfg.SignalType != tt.wantSignalType {
				t.Errorf("SignalType = %q, want %q", cfg.SignalType, tt.wantSignalType)
			}
			if tt.wantSecureKeys != nil {
				gotKeys := []SecureJsonDataKey{}
				for k := range cfg.DecryptedSecureJSONData {
					gotKeys = append(gotKeys, k)
				}
				sort.Slice(gotKeys, func(i, j int) bool { return gotKeys[i] < gotKeys[j] })
				want := append([]SecureJsonDataKey(nil), tt.wantSecureKeys...)
				sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
				if !reflect.DeepEqual(gotKeys, want) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, want)
				}
			}
			if tt.wantPassword != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword] != tt.wantPassword {
				t.Errorf("password = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword], tt.wantPassword)
			}
			if tt.wantSecureHTTPHeader != nil {
				if !reflect.DeepEqual(cfg.SecureHttpHeaders, tt.wantSecureHTTPHeader) {
					t.Errorf("SecureHttpHeaders = %v, want %v", cfg.SecureHttpHeaders, tt.wantSecureHTTPHeader)
				}
			}
			if tt.wantDialTimeout != "" && cfg.DialTimeout != tt.wantDialTimeout {
				t.Errorf("DialTimeout = %q, want %q", cfg.DialTimeout, tt.wantDialTimeout)
			}
			if tt.wantQueryTimeout != "" && cfg.QueryTimeout != tt.wantQueryTimeout {
				t.Errorf("QueryTimeout = %q, want %q", cfg.QueryTimeout, tt.wantQueryTimeout)
			}
			if tt.wantLogsDefaultTable != "" {
				if cfg.Logs == nil || cfg.Logs.DefaultTable != tt.wantLogsDefaultTable {
					t.Errorf("Logs.DefaultTable = %+v, want %q", cfg.Logs, tt.wantLogsDefaultTable)
				}
			}
			if tt.wantTraceDefTable != "" {
				if cfg.Traces == nil || cfg.Traces.DefaultTable != tt.wantTraceDefTable {
					t.Errorf("Traces.DefaultTable = %+v, want %q", cfg.Traces, tt.wantTraceDefTable)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	empty := Config{}
	empty.ApplyDefaults()

	if empty.Protocol != ProtocolNative {
		t.Errorf("Protocol = %q, want native", empty.Protocol)
	}
	if empty.ConfigMode != ConfigModeClassic {
		t.Errorf("ConfigMode = %q, want classic", empty.ConfigMode)
	}
	if !empty.EnableMapKeysDiscovery {
		t.Errorf("EnableMapKeysDiscovery = false, want true")
	}
	if empty.DialTimeout != "10" || empty.QueryTimeout != "60" ||
		empty.ConnMaxLifetime != "5" || empty.MaxIdleConns != "25" || empty.MaxOpenConns != "50" {
		t.Errorf("timeouts not defaulted correctly: %+v", empty)
	}
	if !empty.EnableSchemaCache || empty.SchemaCacheTTLSeconds != 60 {
		t.Errorf("schema cache defaults wrong: enabled=%v ttl=%d", empty.EnableSchemaCache, empty.SchemaCacheTTLSeconds)
	}
	if empty.Logs == nil || empty.Logs.DefaultTable != "otel_logs" {
		t.Errorf("Logs default table wrong: %+v", empty.Logs)
	}
	if empty.Traces == nil || empty.Traces.DefaultTable != "otel_traces" ||
		empty.Traces.DurationUnit != TraceDurationUnitNanoseconds {
		t.Errorf("Traces defaults wrong: %+v", empty.Traces)
	}

	// A user-set custom value must be preserved.
	custom := Config{
		Protocol:              ProtocolHTTP,
		ConfigMode:            ConfigModeSingleTable,
		SchemaCacheTTLSeconds: 5,
		DialTimeout:           "30",
	}
	custom.ApplyDefaults()
	if custom.Protocol != ProtocolHTTP {
		t.Errorf("Protocol overridden: got %q", custom.Protocol)
	}
	if custom.ConfigMode != ConfigModeSingleTable {
		t.Errorf("ConfigMode overridden: got %q", custom.ConfigMode)
	}
	if custom.SchemaCacheTTLSeconds != 5 {
		t.Errorf("SchemaCacheTTLSeconds overridden: got %d", custom.SchemaCacheTTLSeconds)
	}
	if custom.DialTimeout != "30" {
		t.Errorf("DialTimeout overridden: got %q", custom.DialTimeout)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "empty errors on host and port",
			cfg:     Config{Protocol: ProtocolNative},
			wantErr: "host",
		},
		{
			name:    "missing port only",
			cfg:     Config{Host: "h", Protocol: ProtocolNative},
			wantErr: "port",
		},
		{
			name:    "missing protocol",
			cfg:     Config{Host: "h", Port: 9000},
			wantErr: "protocol (jsonData.protocol) is required",
		},
		{
			name:    "unknown protocol",
			cfg:     Config{Host: "h", Port: 9000, Protocol: "grpc"},
			wantErr: `unknown protocol "grpc"`,
		},
		{
			name: "native + username",
			cfg:  Config{Host: "h", Port: 9000, Protocol: ProtocolNative},
		},
		{
			name: "tlsAuth requires both cert and key",
			cfg: Config{
				Host: "h", Port: 9000, Protocol: ProtocolNative,
				TLSAuth: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "cert",
				},
			},
			wantErr: "tlsClientKey (secureJsonData) is required when tlsAuth is true",
		},
		{
			name: "tlsAuthWithCACert without CA",
			cfg: Config{
				Host: "h", Port: 9000, Protocol: ProtocolNative,
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert",
		},
		{
			name: "single-table without signal type",
			cfg: Config{
				Host: "h", Port: 9000, Protocol: ProtocolNative,
				ConfigMode: ConfigModeSingleTable,
			},
			wantErr: "signalType is required when configMode is single-table",
		},
		{
			name: "single-table logs is ok",
			cfg: Config{
				Host: "h", Port: 9000, Protocol: ProtocolNative,
				ConfigMode: ConfigModeSingleTable, SignalType: SignalTypeLogs,
			},
		},
		{
			name: "single-table unknown signal errors",
			cfg: Config{
				Host: "h", Port: 9000, Protocol: ProtocolNative,
				ConfigMode: ConfigModeSingleTable, SignalType: "metrics",
			},
			wantErr: `unknown signalType "metrics"`,
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

func TestSecureJsonDataKeysAreUnique(t *testing.T) {
	seen := map[SecureJsonDataKey]bool{}
	for _, k := range SecureJsonDataKeys {
		if seen[k] {
			t.Errorf("duplicate secure key %q", k)
		}
		seen[k] = true
	}
}

// TestConfigJSONRoundtrip ensures that the custom UnmarshalJSON produces a
// Config that can be reserialized back into an equivalent (post-defaults)
// shape. Uses a small hand-written object so the test is independent of the
// example set.
func TestConfigJSONRoundtrip(t *testing.T) {
	payload := []byte(`{
		"host":"clickhouse",
		"port":9000,
		"protocol":"native",
		"username":"default",
		"logs":{"defaultTable":"logs_v2","otelEnabled":true},
		"traces":{"defaultTable":"spans","durationUnit":"milliseconds"}
	}`)

	var cfg Config
	if err := json.Unmarshal(payload, &cfg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	cfg.ApplyDefaults()

	if cfg.Logs == nil || cfg.Logs.DefaultTable != "logs_v2" || !cfg.Logs.OtelEnabled {
		t.Errorf("Logs: %+v", cfg.Logs)
	}
	if cfg.Traces == nil || cfg.Traces.DefaultTable != "spans" || cfg.Traces.DurationUnit != TraceDurationUnitMilliseconds {
		t.Errorf("Traces: %+v", cfg.Traces)
	}
	// Sanity check that omitempty didn't collapse the nested objects at the
	// top-level, by re-marshaling and confirming the OTel keys are present.
	out, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(out), `"logs":{`) || !strings.Contains(string(out), `"traces":{`) {
		t.Errorf("marshalled output missing logs/traces: %s", out)
	}
}

// Guardrail: make sure the plugin ID we advertise matches the upstream file.
func TestPluginIDMatches(t *testing.T) {
	if !strings.EqualFold(PluginID, "grafana-clickhouse-datasource") {
		t.Fatalf("PluginID = %q, want grafana-clickhouse-datasource", PluginID)
	}
}

// Sanity: json marshal round-trip on a fully-configured Config should not
// error, and roundtripped LoadConfig should not surface any parse errors on
// the byte output.
func TestRoundtripWithLoadConfig(t *testing.T) {
	cfg := Config{
		Host:                   "h",
		Port:                   9000,
		Protocol:               ProtocolNative,
		Username:               "default",
		ConfigMode:             ConfigModeClassic,
		EnableMapKeysDiscovery: true,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	settings := backend.DataSourceInstanceSettings{
		JSONData:                data,
		DecryptedSecureJSONData: map[string]string{"password": "p"},
	}
	got, err := LoadConfig(t.Context(), settings)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.Host != cfg.Host || got.Port != cfg.Port || got.Protocol != cfg.Protocol {
		t.Errorf("round-trip mismatch: %+v vs %+v", got, cfg)
	}
	if errors.Is(err, errors.New("")) {
		// unreachable, kept only for the errors import
	}
}
