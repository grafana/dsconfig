package yugabytedatasource

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
		wantHost string
		wantPort string
	}{
		{
			name: "basic ok",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "localhost:5433",
				User:                    "yugabyte",
				JSONData:                mustJSON(t, map[string]any{"database": "yb_demo"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantURL:  "localhost:5433",
			wantUser: "yugabyte",
			wantDB:   "yb_demo",
			wantHost: "localhost",
			wantPort: "5433",
		},
		{
			name: "ipv6 ok",
			settings: backend.DataSourceInstanceSettings{
				URL:      "[::1]:5433",
				User:     "yugabyte",
				JSONData: mustJSON(t, map[string]any{"database": "yb_demo"}),
			},
			wantHost: "::1",
			wantPort: "5433",
			wantDB:   "yb_demo",
		},
		{
			name: "url missing port",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost",
				User:     "yugabyte",
				JSONData: mustJSON(t, map[string]any{"database": "yb_demo"}),
			},
			wantErr: "parse url",
		},
		{
			name: "url has scheme fails split",
			settings: backend.DataSourceInstanceSettings{
				URL:      "postgres://localhost:5433",
				User:     "yugabyte",
				JSONData: mustJSON(t, map[string]any{"database": "yb_demo"}),
			},
			wantErr: "parse url",
		},
		{
			name: "missing user",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:5433",
				JSONData: mustJSON(t, map[string]any{"database": "yb_demo"}),
			},
			wantErr: "username",
		},
		{
			name: "missing database",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:5433",
				User:     "yugabyte",
				JSONData: json.RawMessage(`{}`),
			},
			wantErr: "database",
		},
		{
			name: "invalid jsonData",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:5433",
				User:     "yugabyte",
				JSONData: json.RawMessage(`{invalid}`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "empty jsonData accepted then defaulted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:5433",
				User:     "yugabyte",
				JSONData: nil,
			},
			wantErr: "database",
		},
		{
			name: "secure key propagated",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "localhost:5433",
				User:                    "yugabyte",
				JSONData:                mustJSON(t, map[string]any{"database": "yb_demo"}),
				DecryptedSecureJSONData: map[string]string{"password": "s3cret"},
			},
			wantDB: "yb_demo",
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
			if tc.wantDB != "" && cfg.Database != tc.wantDB {
				t.Errorf("Database: want %q, got %q", tc.wantDB, cfg.Database)
			}
			if tc.wantHost != "" && cfg.Connection.Host != tc.wantHost {
				t.Errorf("Connection.Host: want %q, got %q", tc.wantHost, cfg.Connection.Host)
			}
			if tc.wantPort != "" && cfg.Connection.Port != tc.wantPort {
				t.Errorf("Connection.Port: want %q, got %q", tc.wantPort, cfg.Connection.Port)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	// ApplyDefaults is a no-op for Yugabyte; assert it does not mutate Config
	// (the plugin has no discriminator or license field that needs a default).
	c := Config{
		URL:      "localhost:5433",
		User:     "yugabyte",
		Database: "yb_demo",
		Connection: Connection{
			URL: "localhost:5433", Host: "localhost", Port: "5433",
		},
	}
	c.ApplyDefaults()
	if c.URL != "localhost:5433" || c.User != "yugabyte" || c.Database != "yb_demo" {
		t.Errorf("ApplyDefaults must be a no-op, got %+v", c)
	}
	if c.Connection.Host != "localhost" || c.Connection.Port != "5433" {
		t.Errorf("ApplyDefaults must not touch Connection, got %+v", c.Connection)
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
			cfg: Config{
				URL:      "localhost:5433",
				User:     "yugabyte",
				Database: "yb_demo",
				Connection: Connection{
					URL: "localhost:5433", Host: "localhost", Port: "5433",
				},
			},
		},
		{
			name:    "missing url",
			cfg:     Config{User: "u", Database: "db"},
			wantErr: "host URL",
		},
		{
			name: "url has no port",
			cfg: Config{
				URL: "localhost", User: "u", Database: "db",
				Connection: Connection{URL: "localhost", Host: "localhost", Port: ""},
			},
			wantErr: "host:port",
		},
		{
			name:    "missing user",
			cfg:     Config{URL: "u:1", Database: "db", Connection: Connection{URL: "u:1", Host: "u", Port: "1"}},
			wantErr: "username",
		},
		{
			name:    "missing database",
			cfg:     Config{URL: "u:1", User: "u", Connection: Connection{URL: "u:1", Host: "u", Port: "1"}},
			wantErr: "database",
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
	// Guard against accidental additions — the plugin currently has exactly one secret.
	if len(SecureJsonDataKeys) != 1 {
		t.Fatalf("expected 1 secure key, got %d", len(SecureJsonDataKeys))
	}
	if SecureJsonDataKeys[0] != SecureJsonDataKeyPassword {
		t.Fatalf("expected password, got %q", SecureJsonDataKeys[0])
	}
}
