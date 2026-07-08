package mqttdatasource

import (
	"encoding/json"
	"reflect"
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
		name              string
		example           string // schema.go SettingsExamples key ("" reserved for the default)
		useExample        bool
		settings          backend.DataSourceInstanceSettings
		wantErr           string // substring match; empty = expect no error
		wantURI           string
		wantClientID      string
		wantUsername      string
		wantTLSAuth       bool
		wantTLSSkipVerify bool
		wantTLSWithCA     bool
		wantSecureKeys    SecureJsonDataConfig
		wantPassword      string
		wantCACert        string
		wantClientCert    string
		wantClientKey     string
	}{
		{
			// The default schema example intentionally has an empty URI so
			// Validate is expected to reject it.
			name:       "default example fails validation (empty uri)",
			example:    "",
			useExample: true,
			wantErr:    "mqtt broker URI",
		},
		{
			name:           "anonymous TCP example",
			example:        "anonymousTCP",
			useExample:     true,
			wantURI:        "tcp://broker.example.com:1883",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPassword:   "",
		},
		{
			name:           "basic auth over TCP",
			example:        "basicAuthTCP",
			useExample:     true,
			wantURI:        "tcp://broker.example.com:1883",
			wantUsername:   "grafana",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPassword:   "s3cret",
		},
		{
			name:           "TLS client certificate authentication",
			example:        "tlsClientAuth",
			useExample:     true,
			wantURI:        "tls://broker.example.com:8883",
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
			wantClientCert: "-----BEGIN CERTIFICATE-----\nMIIB...clientcert...==\n-----END CERTIFICATE-----",
			wantClientKey:  "-----BEGIN RSA PRIVATE KEY-----\nMIIE...clientkey...==\n-----END RSA PRIVATE KEY-----",
		},
		{
			name:           "self-signed CA",
			example:        "selfSignedCA",
			useExample:     true,
			wantURI:        "tls://broker.internal.corp:8883",
			wantTLSWithCA:  true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
			wantCACert:     "-----BEGIN CERTIFICATE-----\nMIIC...cacert...==\n-----END CERTIFICATE-----",
		},
		{
			name:           "mutual TLS with CA and basic auth",
			example:        "mutualTLSWithCA",
			useExample:     true,
			wantURI:        "tls://broker.internal.corp:8883",
			wantClientID:   "grafana-mqtt-1",
			wantUsername:   "grafana",
			wantTLSAuth:    true,
			wantTLSWithCA:  true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword, SecureJsonDataKeyTLSCACert, SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
			wantPassword:   "s3cret",
			wantCACert:     "-----BEGIN CERTIFICATE-----\nMIIC...cacert...==\n-----END CERTIFICATE-----",
			wantClientCert: "-----BEGIN CERTIFICATE-----\nMIIB...clientcert...==\n-----END CERTIFICATE-----",
			wantClientKey:  "-----BEGIN RSA PRIVATE KEY-----\nMIIE...clientkey...==\n-----END RSA PRIVATE KEY-----",
		},
		{
			name:              "TLS skip verify",
			example:           "tlsSkipVerify",
			useExample:        true,
			wantURI:           "tls://broker.example.com:8883",
			wantTLSSkipVerify: true,
			wantSecureKeys:    SecureJsonDataConfig{SecureJsonDataKeyPassword},
		},
		{
			name:           "WebSocket transport",
			example:        "webSocket",
			useExample:     true,
			wantURI:        "wss://broker.example.com:443/mqtt",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
		},
		{
			// Empty JSONData is not a parse error for MQTT because we
			// mirror the upstream flow by skipping unmarshal on empty
			// bytes; Validate then rejects the resulting empty URI.
			name:     "empty settings default to empty URI and fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  "mqtt broker URI",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "missing uri errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"clientID":"x"}`),
			},
			wantErr: "mqtt broker URI",
		},
		{
			name: "clientCert without key errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"uri":"tls://broker:8883","tlsAuth":true}`),
				DecryptedSecureJSONData: map[string]string{"tlsClientCert": "cert"},
			},
			wantErr: "tlsClientKey is required",
		},
		{
			name: "clientKey without cert errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"uri":"tls://broker:8883","tlsAuth":true}`),
				DecryptedSecureJSONData: map[string]string{"tlsClientKey": "key"},
			},
			wantErr: "tlsClientCert is required",
		},
		{
			// The MQTT plugin doesn't use enableSecureSocksProxy in its own
			// Go code, but its editor writes the key when the Grafana-level
			// switch is enabled. json.Unmarshal silently ignores it.
			name: "unknown enableSecureSocksProxy field is ignored",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"uri":"tcp://broker:1883","enableSecureSocksProxy":true}`),
			},
			wantURI: "tcp://broker:1883",
		},
		{
			name: "all TLS toggles parse correctly",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"uri":"tls://h:8883","tlsAuth":true,"tlsAuthWithCACert":true,"tlsSkipVerify":true}`),
				DecryptedSecureJSONData: map[string]string{"tlsClientCert": "c", "tlsClientKey": "k", "tlsCACert": "ca"},
			},
			wantURI:           "tls://h:8883",
			wantTLSAuth:       true,
			wantTLSWithCA:     true,
			wantTLSSkipVerify: true,
			wantClientCert:    "c",
			wantClientKey:     "k",
			wantCACert:        "ca",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.useExample {
				settings = settingsFromExample(t, tt.example)
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("LoadConfig: expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("LoadConfig: error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}

			if tt.wantURI != "" && cfg.URI != tt.wantURI {
				t.Errorf("URI = %q, want %q", cfg.URI, tt.wantURI)
			}
			if tt.wantClientID != "" && cfg.ClientID != tt.wantClientID {
				t.Errorf("ClientID = %q, want %q", cfg.ClientID, tt.wantClientID)
			}
			if tt.wantUsername != "" && cfg.Username != tt.wantUsername {
				t.Errorf("Username = %q, want %q", cfg.Username, tt.wantUsername)
			}
			if cfg.TLSAuth != tt.wantTLSAuth {
				t.Errorf("TLSAuth = %v, want %v", cfg.TLSAuth, tt.wantTLSAuth)
			}
			if cfg.TLSAuthWithCACert != tt.wantTLSWithCA {
				t.Errorf("TLSAuthWithCACert = %v, want %v", cfg.TLSAuthWithCACert, tt.wantTLSWithCA)
			}
			if cfg.TLSSkipVerify != tt.wantTLSSkipVerify {
				t.Errorf("TLSSkipVerify = %v, want %v", cfg.TLSSkipVerify, tt.wantTLSSkipVerify)
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
			if tt.wantPassword != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword] != tt.wantPassword {
				t.Errorf("Secrets[password] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword], tt.wantPassword)
			}
			if tt.wantCACert != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert] != tt.wantCACert {
				t.Errorf("Secrets[tlsCACert] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert], tt.wantCACert)
			}
			if tt.wantClientCert != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert] != tt.wantClientCert {
				t.Errorf("Secrets[tlsClientCert] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientCert], tt.wantClientCert)
			}
			if tt.wantClientKey != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey] != tt.wantClientKey {
				t.Errorf("Secrets[tlsClientKey] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyTLSClientKey], tt.wantClientKey)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	// ApplyDefaults is currently a no-op — every schema default is the Go
	// zero value. Guard the contract so future edits don't quietly start
	// mutating fields.
	t.Run("no-op for empty config", func(t *testing.T) {
		got := Config{}
		got.ApplyDefaults()
		if !reflect.DeepEqual(got, Config{}) {
			t.Errorf("ApplyDefaults() mutated an empty config: %+v", got)
		}
	})
	t.Run("preserves populated config", func(t *testing.T) {
		in := Config{
			URI:               "tcp://broker:1883",
			ClientID:          "grafana",
			Username:          "u",
			TLSAuth:           true,
			TLSAuthWithCACert: true,
			TLSSkipVerify:     true,
		}
		got := in
		got.ApplyDefaults()
		if got.URI != in.URI || got.ClientID != in.ClientID || got.Username != in.Username ||
			got.TLSAuth != in.TLSAuth || got.TLSAuthWithCACert != in.TLSAuthWithCACert ||
			got.TLSSkipVerify != in.TLSSkipVerify {
			t.Errorf("ApplyDefaults() mutated a populated config: got %+v want %+v", got, in)
		}
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "minimal anonymous connection",
			cfg:  Config{URI: "tcp://broker:1883"},
		},
		{
			name: "TLS keypair present",
			cfg: Config{
				URI: "tls://broker:8883",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "cert",
					SecureJsonDataKeyTLSClientKey:  "key",
				},
			},
		},
		{
			name:    "empty URI errors",
			cfg:     Config{},
			wantErr: "mqtt broker URI",
		},
		{
			name: "client cert without key errors",
			cfg: Config{
				URI: "tls://broker:8883",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "cert",
				},
			},
			wantErr: "tlsClientKey is required",
		},
		{
			name: "client key without cert errors",
			cfg: Config{
				URI: "tls://broker:8883",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientKey: "key",
				},
			},
			wantErr: "tlsClientCert is required",
		},
		{
			name: "unused TLS toggles do not cause errors",
			cfg: Config{
				URI:               "tcp://broker:1883",
				TLSAuth:           true,
				TLSAuthWithCACert: true,
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
