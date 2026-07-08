package postgresqldatasource

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
		wantURL  string
		wantUser string
		wantDB   string
		wantMode TLSMode
	}{
		{
			name: "basic ok",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "localhost:5432",
				User:                    "u",
				JSONData:                mustJSON(t, map[string]any{"database": "metrics"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantURL:  "localhost:5432",
			wantUser: "u",
			wantDB:   "metrics",
			wantMode: TLSModeRequire,
		},
		{
			name: "legacy root database",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:5432",
				User:     "u",
				Database: "legacy",
				JSONData: json.RawMessage(`{}`),
			},
			wantDB: "legacy",
		},
		{
			name: "missing url",
			settings: backend.DataSourceInstanceSettings{
				User:     "u",
				JSONData: mustJSON(t, map[string]any{"database": "db"}),
			},
			wantErr: "host URL",
		},
		{
			name: "missing user",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:5432",
				JSONData: mustJSON(t, map[string]any{"database": "db"}),
			},
			wantErr: "username",
		},
		{
			name: "missing database",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:5432",
				User:     "u",
				JSONData: json.RawMessage(`{}`),
			},
			wantErr: "database name",
		},
		{
			name: "unknown sslmode",
			settings: backend.DataSourceInstanceSettings{
				URL:      "u",
				User:     "u",
				JSONData: mustJSON(t, map[string]any{"database": "db", "sslmode": "prefer"}),
			},
			wantErr: "unknown sslmode",
		},
		{
			name: "verify-ca + file-content requires ca cert",
			settings: backend.DataSourceInstanceSettings{
				URL:      "u",
				User:     "u",
				JSONData: mustJSON(t, map[string]any{"database": "db", "sslmode": "verify-ca", "tlsConfigurationMethod": "file-content"}),
			},
			wantErr: "tlsCACert",
		},
		{
			name: "verify-full + file-content with ca cert ok",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "u",
				User:                    "u",
				JSONData:                mustJSON(t, map[string]any{"database": "db", "sslmode": "verify-full", "tlsConfigurationMethod": "file-content"}),
				DecryptedSecureJSONData: map[string]string{"tlsCACert": "CERT"},
			},
			wantMode: TLSModeVerifyFull,
		},
		{
			name: "invalid jsonData",
			settings: backend.DataSourceInstanceSettings{
				URL:      "u",
				User:     "u",
				JSONData: json.RawMessage(`{`),
			},
			wantErr: "parse jsonData",
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
			if tc.wantURL != "" && cfg.URL != tc.wantURL {
				t.Errorf("URL: want %q, got %q", tc.wantURL, cfg.URL)
			}
			if tc.wantUser != "" && cfg.User != tc.wantUser {
				t.Errorf("User: want %q, got %q", tc.wantUser, cfg.User)
			}
			if tc.wantDB != "" && cfg.EffectiveDatabase() != tc.wantDB {
				t.Errorf("EffectiveDatabase: want %q, got %q", tc.wantDB, cfg.EffectiveDatabase())
			}
			if tc.wantMode != "" && cfg.SSLMode != tc.wantMode {
				t.Errorf("SSLMode: want %q, got %q", tc.wantMode, cfg.SSLMode)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	c := Config{}
	c.ApplyDefaults()
	if c.SSLMode != TLSModeRequire {
		t.Errorf("SSLMode: want require, got %q", c.SSLMode)
	}
	if c.TLSConfigurationMethod != TLSMethodFilePath {
		t.Errorf("TLSConfigurationMethod: want file-path, got %q", c.TLSConfigurationMethod)
	}

	// Already-set values must be preserved.
	c2 := Config{SSLMode: TLSModeDisable, TLSConfigurationMethod: TLSMethodFileContent}
	c2.ApplyDefaults()
	if c2.SSLMode != TLSModeDisable {
		t.Errorf("SSLMode overwritten: got %q", c2.SSLMode)
	}
	if c2.TLSConfigurationMethod != TLSMethodFileContent {
		t.Errorf("TLSConfigurationMethod overwritten: got %q", c2.TLSConfigurationMethod)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "ok",
			cfg:  Config{URL: "u", User: "u", JSONDatabase: "db"},
		},
		{
			name:    "no db",
			cfg:     Config{URL: "u", User: "u"},
			wantErr: "database",
		},
		{
			name: "root db fallback ok",
			cfg:  Config{URL: "u", User: "u", Database: "db"},
		},
		{
			name:    "bad sslmode",
			cfg:     Config{URL: "u", User: "u", JSONDatabase: "db", SSLMode: "prefer"},
			wantErr: "unknown sslmode",
		},
		{
			name:    "bad tls method",
			cfg:     Config{URL: "u", User: "u", JSONDatabase: "db", TLSConfigurationMethod: "inline"},
			wantErr: "unknown tlsConfigurationMethod",
		},
		{
			name:    "verify-full inline needs ca",
			cfg:     Config{URL: "u", User: "u", JSONDatabase: "db", SSLMode: TLSModeVerifyFull, TLSConfigurationMethod: TLSMethodFileContent},
			wantErr: "tlsCACert",
		},
		{
			name: "verify-full inline with ca ok",
			cfg: Config{
				URL: "u", User: "u", JSONDatabase: "db",
				SSLMode: TLSModeVerifyFull, TLSConfigurationMethod: TLSMethodFileContent,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyTLSCACert: "CERT"},
			},
		},
		{
			name:    "negative maxOpenConns",
			cfg:     Config{URL: "u", User: "u", JSONDatabase: "db", MaxOpenConns: -1},
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
