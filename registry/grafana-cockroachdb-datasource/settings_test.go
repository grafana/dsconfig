package cockroachdbdatasource

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with jsonData and secureJsonData) into the
// backend.DataSourceInstanceSettings shape LoadConfig expects. CockroachDB
// stores url/user/database in jsonData, so no root fields are set here.
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
		name         string
		example      string // schema.go SettingsExamples key ("" means use inline settings)
		settings     backend.DataSourceInstanceSettings
		wantErr      string
		wantURL      string
		wantUser     string
		wantDB       string
		wantAuthType AuthType
		wantPassword string
		wantMode     TLSMode
		wantMethod   TLSMethod
	}{
		{
			// The default example intentionally leaves url/user/database and the
			// password empty, so LoadConfig's Validate step rejects it.
			name:    "default example fails validation",
			example: "",
			wantErr: "host URL",
		},
		{
			name:         "sql auth example",
			example:      "sqlAuth",
			wantURL:      "localhost:26257",
			wantUser:     "grafana_reader",
			wantDB:       "defaultdb",
			wantAuthType: AuthTypeSQL,
			wantPassword: "changeme",
			wantMode:     TLSModeRequire,       // ApplyDefaults
			wantMethod:   TLSMethodFileContent, // ApplyDefaults
		},
		{
			name:         "kerberos auth example (no password)",
			example:      "kerberosAuth",
			wantAuthType: AuthTypeKerberos,
			wantURL:      "crdb.internal:26257",
			wantDB:       "defaultdb",
		},
		{
			name:         "tls verify-full file-path example",
			example:      "tlsVerifyFullFilePath",
			wantAuthType: AuthTypeTLS,
			wantMode:     TLSModeVerifyFull,
			wantMethod:   TLSMethodFilePath,
			wantPassword: "changeme",
		},
		{
			name:         "tls verify-ca file-content example",
			example:      "tlsVerifyCAFileContent",
			wantAuthType: AuthTypeTLS,
			wantMode:     TLSModeVerifyCA,
			wantMethod:   TLSMethodFileContent,
			wantPassword: "changeme",
		},
		{
			name: "sql auth inline ok + pool defaults",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "SQL Authentication", "url": "h:26257", "user": "u", "database": "db"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantURL:      "h:26257",
			wantUser:     "u",
			wantDB:       "db",
			wantAuthType: AuthTypeSQL,
			wantPassword: "pw",
		},
		{
			name: "missing url",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "SQL Authentication", "user": "u", "database": "db"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantErr: "host URL",
		},
		{
			name: "missing user",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "SQL Authentication", "url": "h:26257", "database": "db"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantErr: "username",
		},
		{
			name: "missing database",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "SQL Authentication", "url": "h:26257", "user": "u"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantErr: "database name",
		},
		{
			name: "sql auth missing password",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{"authType": "SQL Authentication", "url": "h:26257", "user": "u", "database": "db"}),
			},
			wantErr: "password",
		},
		{
			name: "kerberos missing credential cache",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{"authType": "Kerberos Authentication", "url": "h:26257", "user": "u", "database": "db"}),
			},
			wantErr: "credentialCache",
		},
		{
			name: "kerberos with credential cache ok (no password)",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{"authType": "Kerberos Authentication", "url": "h:26257", "user": "u", "database": "db", "credentialCache": "/tmp/krb5cc_1000"}),
			},
			wantAuthType: AuthTypeKerberos,
		},
		{
			name: "unknown authType",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "LDAP", "url": "h:26257", "user": "u", "database": "db"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantErr: "unknown authType",
		},
		{
			name: "unknown sslmode",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "TLS/SSL Authentication", "url": "h:26257", "user": "u", "database": "db", "sslmode": "prefer"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantErr: "unknown sslmode",
		},
		{
			name: "tls file-path missing certs",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "TLS/SSL Authentication", "url": "h:26257", "user": "u", "database": "db", "sslmode": "verify-full", "tlsConfigurationMethod": "file-path"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantErr: "file-path TLS requires",
		},
		{
			name: "tls file-content missing certs",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "TLS/SSL Authentication", "url": "h:26257", "user": "u", "database": "db", "sslmode": "verify-ca", "tlsConfigurationMethod": "file-content"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantErr: "file-content TLS requires",
		},
		{
			name: "tls file-content with certs ok",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{"authType": "TLS/SSL Authentication", "url": "h:26257", "user": "u", "database": "db", "sslmode": "verify-ca", "tlsConfigurationMethod": "file-content"}),
				DecryptedSecureJSONData: map[string]string{
					"password":      "pw",
					"tlsCACert":     "CA",
					"tlsClientCert": "CERT",
					"tlsClientKey":  "KEY",
				},
			},
			wantAuthType: AuthTypeTLS,
			wantMode:     TLSModeVerifyCA,
			wantMethod:   TLSMethodFileContent,
		},
		{
			name: "tls disable skips cert requirement",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                mustJSON(t, map[string]any{"authType": "TLS/SSL Authentication", "url": "h:26257", "user": "u", "database": "db", "sslmode": "disable"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantAuthType: AuthTypeTLS,
			wantMode:     TLSModeDisable,
		},
		{
			name: "invalid jsonData",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{`),
			},
			wantErr: "parse jsonData",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			settings := tc.settings
			if tc.example != "" || tc.name == "default example fails validation" {
				settings = settingsFromExample(t, tc.example)
			}

			cfg, err := LoadConfig(context.Background(), settings)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantURL != "" && cfg.URL != tc.wantURL {
				t.Errorf("URL: want %q, got %q", tc.wantURL, cfg.URL)
			}
			if tc.wantUser != "" && cfg.User != tc.wantUser {
				t.Errorf("User: want %q, got %q", tc.wantUser, cfg.User)
			}
			if tc.wantDB != "" && cfg.Database != tc.wantDB {
				t.Errorf("Database: want %q, got %q", tc.wantDB, cfg.Database)
			}
			if tc.wantAuthType != "" && cfg.AuthType != tc.wantAuthType {
				t.Errorf("AuthType: want %q, got %q", tc.wantAuthType, cfg.AuthType)
			}
			if tc.wantPassword != "" && cfg.Password() != tc.wantPassword {
				t.Errorf("Password: want %q, got %q", tc.wantPassword, cfg.Password())
			}
			if tc.wantMode != "" && cfg.SSLMode != tc.wantMode {
				t.Errorf("SSLMode: want %q, got %q", tc.wantMode, cfg.SSLMode)
			}
			if tc.wantMethod != "" && cfg.TLSConfigurationMethod != tc.wantMethod {
				t.Errorf("TLSConfigurationMethod: want %q, got %q", tc.wantMethod, cfg.TLSConfigurationMethod)
			}
			// Pool defaults are always applied.
			if cfg.MaxOpenConns != DefaultMaxOpenConns {
				t.Errorf("MaxOpenConns: want %d, got %d", DefaultMaxOpenConns, cfg.MaxOpenConns)
			}
			if cfg.MaxIdleConns != DefaultMaxIdleConns {
				t.Errorf("MaxIdleConns: want %d, got %d", DefaultMaxIdleConns, cfg.MaxIdleConns)
			}
			if cfg.ConnMaxLifetime != DefaultConnMaxLifetime {
				t.Errorf("ConnMaxLifetime: want %d, got %d", DefaultConnMaxLifetime, cfg.ConnMaxLifetime)
			}
			if cfg.QueryTimeout != DefaultQueryTimeout {
				t.Errorf("QueryTimeout: want %d, got %d", DefaultQueryTimeout, cfg.QueryTimeout)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	t.Run("empty config gets pool + discriminator defaults", func(t *testing.T) {
		c := Config{}
		c.ApplyDefaults()
		if c.MaxOpenConns != DefaultMaxOpenConns || c.MaxIdleConns != DefaultMaxIdleConns {
			t.Errorf("pool defaults not applied: %+v", c)
		}
		if c.ConnMaxLifetime != DefaultConnMaxLifetime || c.QueryTimeout != DefaultQueryTimeout {
			t.Errorf("lifetime/timeout defaults not applied: %+v", c)
		}
		if c.SSLMode != TLSModeRequire {
			t.Errorf("SSLMode: want require, got %q", c.SSLMode)
		}
		if c.TLSConfigurationMethod != TLSMethodFileContent {
			t.Errorf("TLSConfigurationMethod: want file-content, got %q", c.TLSConfigurationMethod)
		}
	})

	t.Run("query timeout clamps to minimum 5", func(t *testing.T) {
		c := Config{QueryTimeout: 2}
		c.ApplyDefaults()
		if c.QueryTimeout != MinQueryTimeout {
			t.Errorf("QueryTimeout: want %d, got %d", MinQueryTimeout, c.QueryTimeout)
		}
	})

	t.Run("query timeout clamps to maximum 600", func(t *testing.T) {
		c := Config{QueryTimeout: 1000}
		c.ApplyDefaults()
		if c.QueryTimeout != MaxQueryTimeout {
			t.Errorf("QueryTimeout: want %d, got %d", MaxQueryTimeout, c.QueryTimeout)
		}
	})

	t.Run("already-set discriminators preserved", func(t *testing.T) {
		c := Config{SSLMode: TLSModeDisable, TLSConfigurationMethod: TLSMethodFilePath}
		c.ApplyDefaults()
		if c.SSLMode != TLSModeDisable {
			t.Errorf("SSLMode overwritten: got %q", c.SSLMode)
		}
		if c.TLSConfigurationMethod != TLSMethodFilePath {
			t.Errorf("TLSConfigurationMethod overwritten: got %q", c.TLSConfigurationMethod)
		}
	})
}

func TestValidate(t *testing.T) {
	base := func() Config {
		return Config{
			URL: "h:26257", User: "u", Database: "db", AuthType: AuthTypeSQL,
			DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw"},
		}
	}
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{name: "sql ok", cfg: base()},
		{
			name:    "missing everything",
			cfg:     Config{},
			wantErr: "host URL",
		},
		{
			name: "kerberos without password ok",
			cfg: Config{
				URL: "h:26257", User: "u", Database: "db", AuthType: AuthTypeKerberos,
				CredentialCache: "/tmp/krb5cc_1000",
			},
		},
		{
			name: "kerberos without credential cache",
			cfg: Config{
				URL: "h:26257", User: "u", Database: "db", AuthType: AuthTypeKerberos,
			},
			wantErr: "credentialCache",
		},
		{
			name:    "unknown auth type",
			cfg:     Config{URL: "h:26257", User: "u", Database: "db", AuthType: "bogus", DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw"}},
			wantErr: "unknown authType",
		},
		{
			name: "tls file-content needs all three secrets",
			cfg: Config{
				URL: "h:26257", User: "u", Database: "db", AuthType: AuthTypeTLS,
				SSLMode: TLSModeVerifyCA, TLSConfigurationMethod: TLSMethodFileContent,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw", SecureJsonDataKeyTLSCACert: "CA"},
			},
			wantErr: "file-content TLS requires",
		},
		{
			name: "tls file-path needs all three paths",
			cfg: Config{
				URL: "h:26257", User: "u", Database: "db", AuthType: AuthTypeTLS,
				SSLMode: TLSModeVerifyFull, TLSConfigurationMethod: TLSMethodFilePath,
				SSLRootCertFile:         "/ca.crt",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw"},
			},
			wantErr: "file-path TLS requires",
		},
		{
			name: "negative maxOpenConns",
			cfg: func() Config {
				c := base()
				c.MaxOpenConns = -1
				return c
			}(),
			wantErr: "maxOpenConns",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestSecureJsonDataKeys(t *testing.T) {
	if len(SecureJsonDataKeys) != 4 {
		t.Fatalf("expected 4 secure keys, got %d", len(SecureJsonDataKeys))
	}
	want := map[SecureJsonDataKey]bool{
		SecureJsonDataKeyPassword:      true,
		SecureJsonDataKeyTLSCACert:     true,
		SecureJsonDataKeyTLSClientCert: true,
		SecureJsonDataKeyTLSClientKey:  true,
	}
	for _, k := range SecureJsonDataKeys {
		if !want[k] {
			t.Errorf("unexpected secure key %q", k)
		}
	}
}
