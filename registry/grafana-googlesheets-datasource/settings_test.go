package googlesheetsdatasource

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
		name          string
		settings      backend.DataSourceInstanceSettings
		wantErr       string
		wantAuth      AuthType
		wantSheetID   string
		wantProject   string
		checkSecrets  map[SecureJsonDataKey]string
		wantClientEml string
		wantTokenURI  string
		wantKeyPath   string
	}{
		{
			name: "default example (jwt) — missing credentials fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
				}),
			},
			wantErr: "defaultProject is required",
		},
		{
			name: "API key happy path",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "key",
				}),
				DecryptedSecureJSONData: map[string]string{"apiKey": "AIzaKEY"},
			},
			wantAuth:     AuthTypeAPIKey,
			checkSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "AIzaKEY"},
		},
		{
			name: "API key without key fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "key",
				}),
			},
			wantErr: "apiKey is required",
		},
		{
			name: "JWT happy path with inline private key",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
					"defaultProject":     "proj",
					"clientEmail":        "sa@proj.iam.gserviceaccount.com",
					"tokenUri":           "https://oauth2.googleapis.com/token",
				}),
				DecryptedSecureJSONData: map[string]string{"privateKey": "PEM"},
			},
			wantAuth:      AuthTypeJWT,
			wantProject:   "proj",
			wantClientEml: "sa@proj.iam.gserviceaccount.com",
			wantTokenURI:  "https://oauth2.googleapis.com/token",
			checkSecrets:  map[SecureJsonDataKey]string{SecureJsonDataKeyPrivateKey: "PEM"},
		},
		{
			name: "JWT happy path with privateKeyPath (no inline secret)",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
					"defaultProject":     "proj",
					"clientEmail":        "sa@proj.iam.gserviceaccount.com",
					"tokenUri":           "https://oauth2.googleapis.com/token",
					"privateKeyPath":     "/etc/secrets/sa.json",
				}),
			},
			wantAuth:    AuthTypeJWT,
			wantKeyPath: "/etc/secrets/sa.json",
		},
		{
			name: "JWT missing defaultProject fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
					"clientEmail":        "sa@proj.iam.gserviceaccount.com",
					"tokenUri":           "https://oauth2.googleapis.com/token",
				}),
				DecryptedSecureJSONData: map[string]string{"privateKey": "PEM"},
			},
			wantErr: "defaultProject is required",
		},
		{
			name: "JWT missing all credentials fails validation with joined errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
				}),
			},
			wantErr: "defaultProject is required",
		},
		{
			name: "GCE happy path",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "gce",
					"defaultProject":     "proj",
				}),
			},
			wantAuth:    AuthTypeGCE,
			wantProject: "proj",
		},
		{
			name: "GCE without defaultProject still succeeds (defaultProject is optional for GCE)",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "gce",
				}),
			},
			wantAuth: AuthTypeGCE,
		},
		{
			name: "legacy authType migrated to authenticationType",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authType": "key",
				}),
				DecryptedSecureJSONData: map[string]string{"apiKey": "AIzaLEGACY"},
			},
			wantAuth:     AuthTypeAPIKey,
			checkSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "AIzaLEGACY"},
		},
		{
			name: "legacy authType and new authenticationType — authType wins (matches upstream)",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authType":           "key",
					"authenticationType": "gce",
				}),
				DecryptedSecureJSONData: map[string]string{"apiKey": "AIzaLEGACY"},
			},
			wantAuth:     AuthTypeAPIKey,
			checkSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "AIzaLEGACY"},
		},
		{
			name: "unknown authenticationType fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "oauth",
				}),
			},
			wantErr: "unknown authenticationType",
		},
		{
			name: "empty settings default to jwt and fail validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{}`),
			},
			wantErr: "defaultProject is required",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "defaultSheetID is loaded",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "key",
					"defaultSheetID":     "sheet-abc",
				}),
				DecryptedSecureJSONData: map[string]string{"apiKey": "AIzaKEY"},
			},
			wantAuth:    AuthTypeAPIKey,
			wantSheetID: "sheet-abc",
		},
		{
			name: "legacy jwt secret blob is decrypted and preserved",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
					"defaultProject":     "proj",
					"clientEmail":        "sa@proj.iam.gserviceaccount.com",
					"tokenUri":           "https://oauth2.googleapis.com/token",
				}),
				DecryptedSecureJSONData: map[string]string{
					"privateKey": "PEM",
					"jwt":        "{\"legacy\":true}",
				},
			},
			wantAuth: AuthTypeJWT,
			checkSecrets: map[SecureJsonDataKey]string{
				SecureJsonDataKeyPrivateKey: "PEM",
				SecureJsonDataKeyJWT:        "{\"legacy\":true}",
			},
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
			if tc.wantSheetID != "" && cfg.DefaultSheetID != tc.wantSheetID {
				t.Errorf("DefaultSheetID: want %q, got %q", tc.wantSheetID, cfg.DefaultSheetID)
			}
			if tc.wantProject != "" && cfg.DefaultProject != tc.wantProject {
				t.Errorf("DefaultProject: want %q, got %q", tc.wantProject, cfg.DefaultProject)
			}
			if tc.wantClientEml != "" && cfg.ClientEmail != tc.wantClientEml {
				t.Errorf("ClientEmail: want %q, got %q", tc.wantClientEml, cfg.ClientEmail)
			}
			if tc.wantTokenURI != "" && cfg.TokenURI != tc.wantTokenURI {
				t.Errorf("TokenURI: want %q, got %q", tc.wantTokenURI, cfg.TokenURI)
			}
			if tc.wantKeyPath != "" && cfg.PrivateKeyPath != tc.wantKeyPath {
				t.Errorf("PrivateKeyPath: want %q, got %q", tc.wantKeyPath, cfg.PrivateKeyPath)
			}
			for k, want := range tc.checkSecrets {
				if got := cfg.DecryptedSecureJSONData[k]; got != want {
					t.Errorf("DecryptedSecureJSONData[%s]: want %q, got %q", k, want, got)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		want    AuthType
		wantErr bool
	}{
		{"empty auth defaults to jwt", Config{}, AuthTypeJWT, false},
		{"already-set auth is untouched (key)", Config{AuthenticationType: AuthTypeAPIKey}, AuthTypeAPIKey, false},
		{"already-set auth is untouched (gce)", Config{AuthenticationType: AuthTypeGCE}, AuthTypeGCE, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.ApplyDefaults()
			if tc.cfg.AuthenticationType != tc.want {
				t.Errorf("AuthenticationType: want %q, got %q", tc.want, tc.cfg.AuthenticationType)
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
			name: "key + apiKey ok",
			cfg: Config{
				AuthenticationType:      AuthTypeAPIKey,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "AIza"},
			},
		},
		{
			name:    "key without apiKey",
			cfg:     Config{AuthenticationType: AuthTypeAPIKey},
			wantErr: "apiKey is required",
		},
		{
			name: "jwt with inline private key ok",
			cfg: Config{
				AuthenticationType:      AuthTypeJWT,
				DefaultProject:          "p",
				ClientEmail:             "e",
				TokenURI:                "u",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPrivateKey: "PEM"},
			},
		},
		{
			name: "jwt with privateKeyPath ok (no inline secret)",
			cfg: Config{
				AuthenticationType: AuthTypeJWT,
				DefaultProject:     "p",
				ClientEmail:        "e",
				TokenURI:           "u",
				PrivateKeyPath:     "/etc/sa.json",
			},
		},
		{
			name: "jwt without private key or path",
			cfg: Config{
				AuthenticationType: AuthTypeJWT,
				DefaultProject:     "p",
				ClientEmail:        "e",
				TokenURI:           "u",
			},
			wantErr: "secureJsonData.privateKey or jsonData.privateKeyPath is required",
		},
		{
			name: "jwt missing every field surfaces every error",
			cfg: Config{
				AuthenticationType: AuthTypeJWT,
			},
			wantErr: "defaultProject is required",
		},
		{
			name: "gce ok with defaultProject",
			cfg:  Config{AuthenticationType: AuthTypeGCE, DefaultProject: "p"},
		},
		{
			name: "gce ok without defaultProject",
			cfg:  Config{AuthenticationType: AuthTypeGCE},
		},
		{
			name:    "empty auth",
			cfg:     Config{},
			wantErr: "authenticationType is required",
		},
		{
			name:    "unknown auth",
			cfg:     Config{AuthenticationType: "oauth"},
			wantErr: "unknown authenticationType",
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
