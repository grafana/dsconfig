package oracledatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with an optional root url, jsonData, and secureJsonData) into
// the backend.DataSourceInstanceSettings shape LoadConfig expects.
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
	if u, ok := value["url"].(string); ok {
		settings.URL = u
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name              string
		example           string // schema.go SettingsExamples key (non-empty only)
		useDefaultExample bool   // load the "" (default) example
		settings          backend.DataSourceInstanceSettings
		wantErr           error
		wantConnType      ConnectionType
		wantAuthType      AuthType
		wantURL           string
		wantUser          string
		wantDatabase      string
		wantTNSEntry      string // expected cfg.TNSNamesEntry after load (post-migration)
		wantPassword      string
		wantSecureKeys    SecureJsonDataConfig
		wantPoolSize      int
		wantTimeout       int
		wantPrefetch      int
		wantRowLimit      int64
		wantTimezoneName  string
	}{
		{
			// The default schema example has empty url/user/database/password
			// placeholders, so LoadConfig's Validate step is expected to reject it.
			name:              "default example fails validation (empty placeholders)",
			useDefaultExample: true,
			wantErr:           errors.New("host (root.url) is required"),
		},
		{
			name:             "basic auth over host with tcp port",
			example:          "basicAuthTcp",
			wantConnType:     ConnectionTypeTCP,
			wantAuthType:     AuthTypeBasic,
			wantURL:          "oracle.example.com:1521",
			wantUser:         "grafana_reader",
			wantDatabase:     "ORCLPDB1",
			wantPassword:     "changeme",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPoolSize:     DefaultConnectionPoolSize,
			wantTimeout:      DefaultDataProxyTimeout,
			wantRowLimit:     DefaultRowLimit,
			wantTimezoneName: DefaultTimezone,
		},
		{
			name:         "basic auth over tnsnames entry",
			example:      "basicAuthTns",
			wantConnType: ConnectionTypeTNS,
			wantAuthType: AuthTypeBasic,
			wantUser:     "grafana_reader",
			wantTNSEntry: "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=oracle-db1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=FREEPDB1)(SERVER=DEDICATED)))",
			wantPassword: "changeme",
		},
		{
			name:         "kerberos auth over tnsnames entry",
			example:      "kerberosTns",
			wantConnType: ConnectionTypeTNS,
			wantAuthType: AuthTypeKerberos,
			wantTNSEntry: "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=oracle-db1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=FREEPDB1)(SERVER=DEDICATED)))",
		},
		{
			name:             "tuned additional settings",
			example:          "tunedSettings",
			wantConnType:     ConnectionTypeTCP,
			wantAuthType:     AuthTypeBasic,
			wantURL:          "oracle.example.com:1521",
			wantUser:         "grafana_reader",
			wantDatabase:     "ORCLPDB1",
			wantPassword:     "changeme",
			wantPoolSize:     100,
			wantTimeout:      200,
			wantPrefetch:     500,
			wantRowLimit:     500000,
			wantTimezoneName: "Europe/Berlin",
		},
		{
			name:         "legacy tnsnames entry migrated from root url",
			example:      "legacyTnsInUrl",
			wantConnType: ConnectionTypeTNS,
			wantAuthType: AuthTypeBasic,
			wantURL:      "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=oracle-db1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=FREEPDB1)(SERVER=DEDICATED)))",
			wantTNSEntry: "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=oracle-db1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=FREEPDB1)(SERVER=DEDICATED)))",
			wantUser:     "grafana_reader",
			wantPassword: "changeme",
		},
		{
			name: "kerberos with host+tcp is accepted (backend contract)",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1521",
				JSONData: []byte(`{"useKerberosAuthentication":true,"database":"ORCLPDB1"}`),
			},
			wantConnType: ConnectionTypeTCP,
			wantAuthType: AuthTypeKerberos,
			wantURL:      "localhost:1521",
			wantDatabase: "ORCLPDB1",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "tcp basic missing database errors",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "localhost:1521",
				JSONData:                []byte(`{"user":"u"}`),
				DecryptedSecureJSONData: map[string]string{"password": "p"},
			},
			wantErr: errors.New("database name (jsonData.database) is required"),
		},
		{
			name: "tns basic missing password errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"useTNSNamesBasedConnection":true,"tnsNamesEntry":"entry","user":"u"}`),
			},
			wantErr: errors.New("password (secureJsonData.password) is required"),
		},
		{
			// After ApplyDefaults, an empty config becomes tcp + basic, which
			// requires url/database/user/password — Validate rejects it.
			name:     "empty settings default to tcp+basic and fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("host (root.url) is required"),
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

			if tt.wantConnType != "" && cfg.ConnectionType() != tt.wantConnType {
				t.Errorf("ConnectionType() = %q, want %q", cfg.ConnectionType(), tt.wantConnType)
			}
			if tt.wantAuthType != "" && cfg.AuthType() != tt.wantAuthType {
				t.Errorf("AuthType() = %q, want %q", cfg.AuthType(), tt.wantAuthType)
			}
			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.wantUser != "" && cfg.User != tt.wantUser {
				t.Errorf("User = %q, want %q", cfg.User, tt.wantUser)
			}
			if tt.wantDatabase != "" && cfg.Database != tt.wantDatabase {
				t.Errorf("Database = %q, want %q", cfg.Database, tt.wantDatabase)
			}
			if tt.wantTNSEntry != "" && cfg.TNSNamesEntry != tt.wantTNSEntry {
				t.Errorf("TNSNamesEntry = %q, want %q", cfg.TNSNamesEntry, tt.wantTNSEntry)
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
			if tt.wantPoolSize != 0 && cfg.ConnectionPoolSize != tt.wantPoolSize {
				t.Errorf("ConnectionPoolSize = %d, want %d", cfg.ConnectionPoolSize, tt.wantPoolSize)
			}
			if tt.wantTimeout != 0 && cfg.DataProxyTimeout != tt.wantTimeout {
				t.Errorf("DataProxyTimeout = %d, want %d", cfg.DataProxyTimeout, tt.wantTimeout)
			}
			if tt.wantPrefetch != 0 && cfg.PrefetchRowsCount != tt.wantPrefetch {
				t.Errorf("PrefetchRowsCount = %d, want %d", cfg.PrefetchRowsCount, tt.wantPrefetch)
			}
			if tt.wantRowLimit != 0 && cfg.RowLimit != tt.wantRowLimit {
				t.Errorf("RowLimit = %d, want %d", cfg.RowLimit, tt.wantRowLimit)
			}
			if tt.wantTimezoneName != "" && cfg.TZName != tt.wantTimezoneName {
				t.Errorf("TZName = %q, want %q", cfg.TZName, tt.wantTimezoneName)
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
			name: "empty config gets UTC + backend pool/timeout/row-limit defaults",
			in:   Config{},
			want: Config{
				TZName:             DefaultTimezone,
				ConnectionPoolSize: DefaultConnectionPoolSize,
				DataProxyTimeout:   DefaultDataProxyTimeout,
				RowLimit:           DefaultRowLimit,
			},
		},
		{
			name: "existing values are preserved",
			in: Config{
				TZName:             "Europe/Berlin",
				ConnectionPoolSize: 10,
				DataProxyTimeout:   30,
				RowLimit:           5,
				PrefetchRowsCount:  7,
			},
			want: Config{
				TZName:             "Europe/Berlin",
				ConnectionPoolSize: 10,
				DataProxyTimeout:   30,
				RowLimit:           5,
				PrefetchRowsCount:  7,
			},
		},
		{
			name: "non-positive row limit is reset to default",
			in:   Config{TZName: "UTC", ConnectionPoolSize: 1, DataProxyTimeout: 1, RowLimit: -5},
			want: Config{TZName: "UTC", ConnectionPoolSize: 1, DataProxyTimeout: 1, RowLimit: DefaultRowLimit},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.TZName != tt.want.TZName {
				t.Errorf("TZName = %q, want %q", got.TZName, tt.want.TZName)
			}
			if got.ConnectionPoolSize != tt.want.ConnectionPoolSize {
				t.Errorf("ConnectionPoolSize = %d, want %d", got.ConnectionPoolSize, tt.want.ConnectionPoolSize)
			}
			if got.DataProxyTimeout != tt.want.DataProxyTimeout {
				t.Errorf("DataProxyTimeout = %d, want %d", got.DataProxyTimeout, tt.want.DataProxyTimeout)
			}
			if got.RowLimit != tt.want.RowLimit {
				t.Errorf("RowLimit = %d, want %d", got.RowLimit, tt.want.RowLimit)
			}
			if got.PrefetchRowsCount != tt.want.PrefetchRowsCount {
				t.Errorf("PrefetchRowsCount = %d, want %d", got.PrefetchRowsCount, tt.want.PrefetchRowsCount)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	pw := map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "p"}

	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "tcp basic happy path",
			cfg:  Config{URL: "localhost:1521", Database: "db", User: "u", DecryptedSecureJSONData: pw},
		},
		{
			name:    "tcp basic missing host",
			cfg:     Config{Database: "db", User: "u", DecryptedSecureJSONData: pw},
			wantErr: "host (root.url) is required",
		},
		{
			name:    "tcp basic missing database",
			cfg:     Config{URL: "localhost:1521", User: "u", DecryptedSecureJSONData: pw},
			wantErr: "database name (jsonData.database) is required",
		},
		{
			name:    "tcp basic missing user",
			cfg:     Config{URL: "localhost:1521", Database: "db", DecryptedSecureJSONData: pw},
			wantErr: "user (jsonData.user) is required",
		},
		{
			name:    "tcp basic missing password",
			cfg:     Config{URL: "localhost:1521", Database: "db", User: "u", DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
			wantErr: "password (secureJsonData.password) is required",
		},
		{
			name: "tns basic happy path",
			cfg:  Config{UseTNSNamesBasedConnection: true, TNSNamesEntry: "entry", User: "u", DecryptedSecureJSONData: pw},
		},
		{
			name:    "tns basic missing tnsNamesEntry",
			cfg:     Config{UseTNSNamesBasedConnection: true, User: "u", DecryptedSecureJSONData: pw},
			wantErr: "tnsNamesEntry (jsonData.tnsNamesEntry) is required",
		},
		{
			name: "tns basic legacy tns in url is accepted",
			cfg:  Config{UseTNSNamesBasedConnection: true, URL: "legacy-entry", User: "u", DecryptedSecureJSONData: pw},
		},
		{
			name: "tns kerberos happy path (no user/password)",
			cfg:  Config{UseTNSNamesBasedConnection: true, UseKerberosAuthentication: true, TNSNamesEntry: "entry", DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
		},
		{
			name: "tcp kerberos accepted by backend (no user/password, needs url+database)",
			cfg:  Config{UseKerberosAuthentication: true, URL: "localhost:1521", Database: "db", DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
		},
		{
			name:    "negative connection pool size",
			cfg:     Config{URL: "localhost:1521", Database: "db", User: "u", ConnectionPoolSize: -1, DecryptedSecureJSONData: pw},
			wantErr: "connectionPoolSize must be non-negative",
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

func TestEffectiveTNSNamesEntry(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "explicit entry wins",
			cfg:  Config{UseTNSNamesBasedConnection: true, TNSNamesEntry: "explicit", URL: "root"},
			want: "explicit",
		},
		{
			name: "falls back to root url when entry empty",
			cfg:  Config{UseTNSNamesBasedConnection: true, URL: "root"},
			want: "root",
		},
		{
			name: "no fallback when not a tnsnames connection",
			cfg:  Config{URL: "root"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.EffectiveTNSNamesEntry(); got != tt.want {
				t.Errorf("EffectiveTNSNamesEntry() = %q, want %q", got, tt.want)
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
