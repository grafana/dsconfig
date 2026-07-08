package mysqldatasource

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
		name        string
		settings    backend.DataSourceInstanceSettings
		wantErr     string
		wantURL     string
		wantUser    string
		wantDB      string
		wantSecrets map[SecureJsonDataKey]string
	}{
		{
			name: "basic auth ok",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "localhost:3306",
				User:                    "reader",
				JSONData:                mustJSON(t, map[string]any{"database": "metrics"}),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantURL:     "localhost:3306",
			wantUser:    "reader",
			wantDB:      "metrics",
			wantSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw"},
		},
		{
			name: "legacy root database falls back",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:3306",
				User:     "reader",
				Database: "legacy_db",
				JSONData: json.RawMessage(`{}`),
			},
			wantURL:  "localhost:3306",
			wantUser: "reader",
			wantDB:   "legacy_db",
		},
		{
			name: "missing URL fails",
			settings: backend.DataSourceInstanceSettings{
				User:     "reader",
				JSONData: json.RawMessage(`{}`),
			},
			wantErr: "host URL",
		},
		{
			name: "missing user fails",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:3306",
				JSONData: json.RawMessage(`{}`),
			},
			wantErr: "username",
		},
		{
			name: "tlsAuth requires cert+key",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:3306",
				User:     "reader",
				JSONData: mustJSON(t, map[string]any{"tlsAuth": true}),
			},
			wantErr: "tlsClientCert",
		},
		{
			name: "tlsAuthWithCACert requires ca",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:3306",
				User:     "reader",
				JSONData: mustJSON(t, map[string]any{"tlsAuthWithCACert": true}),
			},
			wantErr: "tlsCACert",
		},
		{
			name: "tlsAuth ok with cert+key",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:3306",
				User:     "reader",
				JSONData: mustJSON(t, map[string]any{"tlsAuth": true}),
				DecryptedSecureJSONData: map[string]string{
					"tlsClientCert": "CERT",
					"tlsClientKey":  "KEY",
				},
			},
			wantURL:  "localhost:3306",
			wantUser: "reader",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "x",
				User:     "y",
				JSONData: json.RawMessage(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "negative maxOpenConns fails",
			settings: backend.DataSourceInstanceSettings{
				URL:      "localhost:3306",
				User:     "reader",
				JSONData: mustJSON(t, map[string]any{"maxOpenConns": -1}),
			},
			wantErr: "maxOpenConns must be non-negative",
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
			for k, want := range tc.wantSecrets {
				if got := cfg.DecryptedSecureJSONData[k]; got != want {
					t.Errorf("DecryptedSecureJSONData[%s]: want %q, got %q", k, want, got)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		wantAuto bool
	}{
		{"empty pool defaults maxIdleConnsAuto=true", Config{}, true},
		{"maxOpenConns already set leaves autoIdle=false", Config{MaxOpenConns: 10}, false},
		{"maxIdleConns already set leaves autoIdle=false", Config{MaxIdleConns: 5}, false},
		{"explicit false stays false when pool already tuned", Config{MaxOpenConns: 10, MaxIdleConnsAuto: false}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.ApplyDefaults()
			if tc.cfg.MaxIdleConnsAuto != tc.wantAuto {
				t.Errorf("MaxIdleConnsAuto: want %v, got %v", tc.wantAuto, tc.cfg.MaxIdleConnsAuto)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "basic ok",
			cfg:  Config{URL: "localhost:3306", User: "u"},
		},
		{
			name:    "no url",
			cfg:     Config{User: "u"},
			wantErr: "host URL",
		},
		{
			name:    "no user",
			cfg:     Config{URL: "localhost:3306"},
			wantErr: "username",
		},
		{
			name: "tlsAuth needs cert+key",
			cfg: Config{
				URL: "u", User: "u", TLSAuth: true,
			},
			wantErr: "tlsClientCert",
		},
		{
			name: "tlsAuth ok",
			cfg: Config{
				URL: "u", User: "u", TLSAuth: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "CERT",
					SecureJsonDataKeyTLSClientKey:  "KEY",
				},
			},
		},
		{
			name:    "tlsAuthWithCACert needs ca",
			cfg:     Config{URL: "u", User: "u", TLSAuthWithCACert: true},
			wantErr: "tlsCACert",
		},
		{
			name:    "negative maxOpenConns",
			cfg:     Config{URL: "u", User: "u", MaxOpenConns: -1},
			wantErr: "maxOpenConns",
		},
		{
			name:    "negative maxIdleConns",
			cfg:     Config{URL: "u", User: "u", MaxIdleConns: -1},
			wantErr: "maxIdleConns",
		},
		{
			name:    "negative connMaxLifetime",
			cfg:     Config{URL: "u", User: "u", ConnMaxLifetime: -1},
			wantErr: "connMaxLifetime",
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
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestEffectiveDatabase(t *testing.T) {
	if got := (Config{Database: "root_db"}).EffectiveDatabase(); got != "root_db" {
		t.Errorf("root fallback: got %q", got)
	}
	if got := (Config{Database: "root_db", JSONDatabase: "json_db"}).EffectiveDatabase(); got != "json_db" {
		t.Errorf("jsonData precedence: got %q", got)
	}
	if got := (Config{}).EffectiveDatabase(); got != "" {
		t.Errorf("empty: got %q", got)
	}
}
