package saphanadatasource

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with jsonData and secureJsonData; SAP HANA has no root-level
// fields) into the backend.DataSourceInstanceSettings shape LoadConfig expects.
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
		example           string // schema.go SettingsExamples key (non-empty only)
		useDefaultExample bool   // load the "" (default) example
		settings          backend.DataSourceInstanceSettings
		wantErr           string
		wantServer        string
		wantPort          int64
		wantInstance      string
		wantDatabaseName  string
		wantUsername      string
		wantTimeout       string
		wantTLSDisabled   bool
		wantSkipVerify    bool
		wantTLSAuth       bool
		wantTLSWithCA     bool
		wantPassword      string
		wantSecureKeys    SecureJsonDataConfig
	}{
		{
			// The default schema example has empty server/username/password
			// placeholders, so LoadConfig's Validate step rejects it.
			name:              "default example fails validation (empty placeholders)",
			useDefaultExample: true,
			wantErr:           "invalid server name",
		},
		{
			name:           "basic auth over host+port",
			example:        "basicAuthPort",
			wantServer:     "hana.example.com",
			wantPort:       443,
			wantUsername:   "GRAFANA_READER",
			wantTimeout:    DefaultTimeout,
			wantPassword:   examplePassword,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
		},
		{
			name:             "basic auth over tenant instance (no port)",
			example:          "basicAuthInstance",
			wantServer:       "hana.example.com",
			wantInstance:     "00",
			wantDatabaseName: "HXE",
			wantUsername:     "GRAFANA_READER",
			wantPassword:     examplePassword,
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyPassword},
		},
		{
			name:           "tls client-certificate auth",
			example:        "tlsClientAuth",
			wantServer:     "hana.example.com",
			wantPort:       443,
			wantTLSAuth:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "basic auth with custom CA cert",
			example:        "tlsWithCACert",
			wantServer:     "hana.example.com",
			wantPort:       443,
			wantUsername:   "GRAFANA_READER",
			wantTLSWithCA:  true,
			wantPassword:   examplePassword,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword, SecureJsonDataKeyTLSCACert},
		},
		{
			name:           "basic auth skipping TLS verify",
			example:        "tlsSkipVerify",
			wantServer:     "hana.example.com",
			wantPort:       443,
			wantUsername:   "GRAFANA_READER",
			wantSkipVerify: true,
			wantPassword:   examplePassword,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
		},
		{
			name:            "basic auth over plaintext (TLS disabled)",
			example:         "tlsDisabled",
			wantServer:      "hana.example.com",
			wantPort:        30015,
			wantUsername:    "GRAFANA_READER",
			wantTLSDisabled: true,
			wantPassword:    examplePassword,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyPassword},
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "missing server errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"port":443,"username":"u"}`),
				DecryptedSecureJSONData: map[string]string{"password": "p"},
			},
			wantErr: "invalid server name",
		},
		{
			name: "port 0 without instance+database errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"server":"hana","username":"u"}`),
				DecryptedSecureJSONData: map[string]string{"password": "p"},
			},
			wantErr: "invalid port or instance",
		},
		{
			name: "basic auth missing username errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"server":"hana","port":443}`),
				DecryptedSecureJSONData: map[string]string{"password": "p"},
			},
			wantErr: "username is either empty",
		},
		{
			name: "basic auth missing password errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"server":"hana","port":443,"username":"u"}`),
			},
			wantErr: "password is either empty",
		},
		{
			name: "tls client auth missing cert errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"server":"hana","port":443,"tlsAuth":true}`),
			},
			wantErr: "tlsClientCert",
		},
		{
			name: "tls client auth missing key errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"server":"hana","port":443,"tlsAuth":true}`),
				DecryptedSecureJSONData: map[string]string{"tlsClientCert": "cert"},
			},
			wantErr: "tlsClientKey",
		},
		{
			name: "instance connection succeeds without a port",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"server":"hana","instance":"00","databaseName":"HXE","username":"u"}`),
				DecryptedSecureJSONData: map[string]string{"password": "p"},
			},
			wantServer:       "hana",
			wantInstance:     "00",
			wantDatabaseName: "HXE",
			wantUsername:     "u",
			wantTimeout:      DefaultTimeout,
		},
		{
			// After ApplyDefaults, empty settings still fail Validate (no server,
			// no port/instance, no username/password).
			name:     "empty settings fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  "invalid server name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var settings backend.DataSourceInstanceSettings
			switch {
			case tt.useDefaultExample:
				settings = settingsFromExample(t, "")
			case tt.example != "":
				settings = settingsFromExample(t, tt.example)
			default:
				settings = tt.settings
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

			if tt.wantServer != "" && cfg.Server != tt.wantServer {
				t.Errorf("Server = %q, want %q", cfg.Server, tt.wantServer)
			}
			if tt.wantPort != 0 && cfg.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", cfg.Port, tt.wantPort)
			}
			if tt.wantInstance != "" && cfg.Instance != tt.wantInstance {
				t.Errorf("Instance = %q, want %q", cfg.Instance, tt.wantInstance)
			}
			if tt.wantDatabaseName != "" && cfg.DatabaseName != tt.wantDatabaseName {
				t.Errorf("DatabaseName = %q, want %q", cfg.DatabaseName, tt.wantDatabaseName)
			}
			if tt.wantUsername != "" && cfg.Username != tt.wantUsername {
				t.Errorf("Username = %q, want %q", cfg.Username, tt.wantUsername)
			}
			if tt.wantTimeout != "" && cfg.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %q, want %q", cfg.Timeout, tt.wantTimeout)
			}
			if cfg.TlsDisabled != tt.wantTLSDisabled {
				t.Errorf("TlsDisabled = %v, want %v", cfg.TlsDisabled, tt.wantTLSDisabled)
			}
			if cfg.InsecureSkipVerify != tt.wantSkipVerify {
				t.Errorf("InsecureSkipVerify = %v, want %v", cfg.InsecureSkipVerify, tt.wantSkipVerify)
			}
			if cfg.TlsClientAuth != tt.wantTLSAuth {
				t.Errorf("TlsClientAuth = %v, want %v", cfg.TlsClientAuth, tt.wantTLSAuth)
			}
			if cfg.TlsAuthWithCACert != tt.wantTLSWithCA {
				t.Errorf("TlsAuthWithCACert = %v, want %v", cfg.TlsAuthWithCACert, tt.wantTLSWithCA)
			}
			if tt.wantPassword != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword] != tt.wantPassword {
				t.Errorf("password = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword], tt.wantPassword)
			}
			if tt.wantSecureKeys != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if v, ok := cfg.DecryptedSecureJSONData[key]; ok && v != "" {
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
		name        string
		in          Config
		wantTimeout string
	}{
		{
			name:        "empty timeout defaults to 30",
			in:          Config{},
			wantTimeout: DefaultTimeout,
		},
		{
			name:        "existing timeout is preserved",
			in:          Config{Timeout: "60"},
			wantTimeout: "60",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %q, want %q", got.Timeout, tt.wantTimeout)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	pw := map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "p"}
	clientKeyPair := map[SecureJsonDataKey]string{
		SecureJsonDataKeyTLSClientCert: "cert",
		SecureJsonDataKeyTLSClientKey:  "key",
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "basic auth over port happy path",
			cfg:  Config{Server: "hana", Port: 443, Username: "u", DecryptedSecureJSONData: pw},
		},
		{
			name: "basic auth over instance+database happy path",
			cfg:  Config{Server: "hana", Instance: "00", DatabaseName: "HXE", Username: "u", DecryptedSecureJSONData: pw},
		},
		{
			name:    "missing server",
			cfg:     Config{Port: 443, Username: "u", DecryptedSecureJSONData: pw},
			wantErr: "invalid server name",
		},
		{
			name:    "port 0 with only instance (no database) errors",
			cfg:     Config{Server: "hana", Instance: "00", Username: "u", DecryptedSecureJSONData: pw},
			wantErr: "invalid port or instance",
		},
		{
			name:    "port 0 with only database (no instance) errors",
			cfg:     Config{Server: "hana", DatabaseName: "HXE", Username: "u", DecryptedSecureJSONData: pw},
			wantErr: "invalid port or instance",
		},
		{
			name:    "basic auth missing username",
			cfg:     Config{Server: "hana", Port: 443, DecryptedSecureJSONData: pw},
			wantErr: "username is either empty",
		},
		{
			name:    "basic auth missing password",
			cfg:     Config{Server: "hana", Port: 443, Username: "u", DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
			wantErr: "password is either empty",
		},
		{
			name: "tls client auth happy path (no username/password)",
			cfg:  Config{Server: "hana", Port: 443, TlsClientAuth: true, DecryptedSecureJSONData: clientKeyPair},
		},
		{
			name:    "tls client auth missing cert",
			cfg:     Config{Server: "hana", Port: 443, TlsClientAuth: true, DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyTLSClientKey: "key"}},
			wantErr: "tlsClientCert",
		},
		{
			name:    "tls client auth missing key",
			cfg:     Config{Server: "hana", Port: 443, TlsClientAuth: true, DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyTLSClientCert: "cert"}},
			wantErr: "tlsClientKey",
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

// TestSettingsExamples guards the SettingsExamples contract: the default example
// keyed "" must exist, and every example must carry a jsonData object plus a
// non-empty secureJsonData object whose keys are all known secret keys.
func TestSettingsExamples(t *testing.T) {
	examples := SettingsExamples().Examples

	if _, ok := examples[""]; !ok {
		t.Fatalf("missing default example keyed \"\"")
	}

	known := map[string]bool{}
	for _, k := range SecureJsonDataKeys {
		known[string(k)] = true
	}

	for key, ex := range examples {
		value, ok := ex.Value.(map[string]any)
		if !ok {
			t.Fatalf("example %q value is not an object", key)
		}
		if _, ok := value["jsonData"].(map[string]any); !ok {
			t.Errorf("example %q missing jsonData object", key)
		}
		secure, ok := value["secureJsonData"].(map[string]any)
		if !ok || len(secure) == 0 {
			t.Errorf("example %q missing non-empty secureJsonData object", key)
			continue
		}
		for k := range secure {
			if !known[k] {
				t.Errorf("example %q references unknown secret key %q", key, k)
			}
		}
	}
}
