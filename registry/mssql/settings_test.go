package mssqldatasource

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

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		settings backend.DataSourceInstanceSettings
		wantErr  string
		wantAuth AuthType
	}{
		{
			name: "sql auth ok",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "localhost:1433",
				User:                    "u",
				JSONData:                mustJSON(t, map[string]any{"database": "db", "authenticationType": "SQL Server Authentication"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantAuth: AuthTypeSQL,
		},
		{
			name: "sql auth without password fails",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				User:     "u",
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "SQL Server Authentication"}),
			},
			wantErr: "password",
		},
		{
			name: "missing url fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "SQL Server Authentication"}),
			},
			wantErr: "host",
		},
		{
			name: "missing database fails",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				User:     "u",
				JSONData: mustJSON(t, map[string]any{"authenticationType": "SQL Server Authentication"}),
			},
			wantErr: "database",
		},
		{
			name: "windows authentication ok without user/password",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "Windows Authentication"}),
			},
			wantAuth: AuthTypeWindows,
		},
		{
			name: "kerberos keytab requires keytab path",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				User:     "u@EX.COM",
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "Windows AD: Keytab"}),
			},
			wantErr: "keytabFilePath",
		},
		{
			name: "kerberos keytab ok with path",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				User:     "u@EX.COM",
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "Windows AD: Keytab", "keytabFilePath": "/etc/kt"}),
			},
			wantAuth: AuthTypeKerberosKeytab,
		},
		{
			name: "kerberos cache requires cache path",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "Windows AD: Credential cache"}),
			},
			wantErr: "credentialCache",
		},
		{
			name: "kerberos cache file requires user + lookup file",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "Windows AD: Credential cache file"}),
			},
			wantErr: "username",
		},
		{
			name: "unknown auth type",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "LDAP"}),
			},
			wantErr: "unknown authenticationType",
		},
		{
			name: "unknown encrypt option",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "localhost:1433",
				User:                    "u",
				JSONData:                mustJSON(t, map[string]any{"database": "db", "authenticationType": "SQL Server Authentication", "encrypt": "yes"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantErr: "unknown encrypt",
		},
		{
			name: "invalid jsonData",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:1433",
				JSONData: json.RawMessage(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "azure AD ok with credentials",
			settings: backend.DataSourceInstanceSettings{
				URL:      "mssql.database.windows.net:1433",
				JSONData: mustJSON(t, map[string]any{"database": "db", "authenticationType": "Azure AD Authentication", "azureCredentials": map[string]any{"authType": "msi"}}),
			},
			wantAuth: AuthTypeAzureAD,
		},
		{
			name: "legacy root database falls back",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "localhost:1433",
				User:                    "u",
				Database:                "legacy",
				JSONData:                mustJSON(t, map[string]any{"authenticationType": "SQL Server Authentication"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantAuth: AuthTypeSQL,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := LoadConfig(context.Background(), tc.settings)
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
			if tc.wantAuth != "" && cfg.AuthenticationType != tc.wantAuth {
				t.Errorf("AuthenticationType: want %q, got %q", tc.wantAuth, cfg.AuthenticationType)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	c := Config{}
	c.ApplyDefaults()
	if c.AuthenticationType != AuthTypeSQL {
		t.Errorf("AuthenticationType: want %q, got %q", AuthTypeSQL, c.AuthenticationType)
	}
	if c.Encrypt != EncryptFalse {
		t.Errorf("Encrypt: want %q, got %q", EncryptFalse, c.Encrypt)
	}
	if c.UDPConnectionLimit != 1 {
		t.Errorf("UDPConnectionLimit: want 1, got %d", c.UDPConnectionLimit)
	}

	// Already-set values preserved.
	c2 := Config{AuthenticationType: AuthTypeAzureAD, Encrypt: EncryptTrue, UDPConnectionLimit: 5}
	c2.ApplyDefaults()
	if c2.AuthenticationType != AuthTypeAzureAD || c2.Encrypt != EncryptTrue || c2.UDPConnectionLimit != 5 {
		t.Errorf("defaults clobbered explicit values: %+v", c2)
	}
}

func TestValidate(t *testing.T) {
	base := Config{URL: "u", JSONDatabase: "db"}
	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr string
	}{
		{
			name: "sql auth ok",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeSQL
				c.User = "u"
				c.DecryptedSecureJSONData = map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw"}
			},
		},
		{
			name:    "no auth type",
			mutate:  func(c *Config) {},
			wantErr: "authenticationType",
		},
		{
			name: "sql without user",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeSQL
				c.DecryptedSecureJSONData = map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw"}
			},
			wantErr: "username",
		},
		{
			name: "windows sso ok",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeWindows
			},
		},
		{
			name: "azure ok",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeAzureAD
			},
		},
		{
			name: "kerberos keytab ok",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeKerberosKeytab
				c.User = "u@EX.COM"
				c.KeytabFilePath = "/etc/kt"
			},
		},
		{
			name: "kerberos cache lookup file needs both user and lookup file",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeKerberosCacheLookupFile
				c.User = "u@EX.COM"
			},
			wantErr: "credentialCacheLookupFile",
		},
		{
			name: "bad encrypt",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeSQL
				c.User = "u"
				c.DecryptedSecureJSONData = map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw"}
				c.Encrypt = "yes"
			},
			wantErr: "unknown encrypt",
		},
		{
			name: "negative timeout",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeWindows
				c.ConnectionTimeout = -1
			},
			wantErr: "connectionTimeout",
		},
		{
			name: "negative UDP limit",
			mutate: func(c *Config) {
				c.AuthenticationType = AuthTypeWindows
				c.UDPConnectionLimit = -1
			},
			wantErr: "UDPConnectionLimit",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := base
			tc.mutate(&c)
			err := c.Validate()
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

func TestEffectiveDatabase(t *testing.T) {
	if got := (Config{Database: "root"}).EffectiveDatabase(); got != "root" {
		t.Errorf("root fallback: got %q", got)
	}
	if got := (Config{Database: "root", JSONDatabase: "json"}).EffectiveDatabase(); got != "json" {
		t.Errorf("jsonData precedence: got %q", got)
	}
}
