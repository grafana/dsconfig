package stackdriver

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
		wantProject   string
		wantOAuthPass bool
		wantUniverse  string
		wantImperson  bool
		wantWIFPool   string
		wantSecrets   map[SecureJsonDataKey]string
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
			wantAuth:    AuthTypeJWT,
			wantProject: "proj",
			wantSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyPrivateKey: "PEM"},
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
			wantProject: "proj",
		},
		{
			name: "JWT without any private key source fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
					"defaultProject":     "proj",
					"clientEmail":        "sa@proj.iam.gserviceaccount.com",
					"tokenUri":           "https://oauth2.googleapis.com/token",
				}),
			},
			wantErr: "secureJsonData.privateKey or jsonData.privateKeyPath is required",
		},
		{
			name: "GCE happy path (no defaultProject — metadata server resolves it)",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "gce",
				}),
			},
			wantAuth: AuthTypeGCE,
		},
		{
			name: "GCE with defaultProject override still succeeds",
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
			name: "GCE tolerates the frontend-only gceDefaultProject cache field",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "gce",
					"gceDefaultProject":  "cached-proj",
				}),
			},
			wantAuth: AuthTypeGCE,
		},
		{
			name: "forwardOAuthIdentity happy path sets oauthPassThru true",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "forwardOAuthIdentity",
					"defaultProject":     "proj",
				}),
			},
			wantAuth:      AuthTypeForwardOAuthIdentity,
			wantProject:   "proj",
			wantOAuthPass: true,
		},
		{
			name: "forwardOAuthIdentity without defaultProject fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "forwardOAuthIdentity",
				}),
			},
			wantErr: "defaultProject is required for 'forwardOAuthIdentity'",
		},
		{
			name: "workloadIdentityFederation happy path sets oauthPassThru true",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType":           "workloadIdentityFederation",
					"workloadIdentityPoolProvider": "projects/1/locations/global/workloadIdentityPools/p/providers/prov",
					"defaultProject":               "proj",
				}),
			},
			wantAuth:      AuthTypeWorkloadIdentityFederation,
			wantProject:   "proj",
			wantOAuthPass: true,
			wantWIFPool:   "projects/1/locations/global/workloadIdentityPools/p/providers/prov",
		},
		{
			name: "workloadIdentityFederation missing pool provider fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "workloadIdentityFederation",
					"defaultProject":     "proj",
				}),
			},
			wantErr: "workloadIdentityPoolProvider is required",
		},
		{
			name: "workloadIdentityFederation missing defaultProject fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType":           "workloadIdentityFederation",
					"workloadIdentityPoolProvider": "projects/1/locations/global/workloadIdentityPools/p/providers/prov",
				}),
			},
			wantErr: "defaultProject is required for 'workloadIdentityFederation'",
		},
		{
			name: "unknown auth type fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "basic",
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
			name: "impersonation on JWT",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType":          "jwt",
					"defaultProject":              "proj",
					"clientEmail":                 "caller@proj.iam.gserviceaccount.com",
					"tokenUri":                    "https://oauth2.googleapis.com/token",
					"usingImpersonation":          true,
					"serviceAccountToImpersonate": "target@proj.iam.gserviceaccount.com",
				}),
				DecryptedSecureJSONData: map[string]string{"privateKey": "PEM"},
			},
			wantAuth:     AuthTypeJWT,
			wantProject:  "proj",
			wantImperson: true,
			wantSecrets:  map[SecureJsonDataKey]string{SecureJsonDataKeyPrivateKey: "PEM"},
		},
		{
			name: "universeDomain populated",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
					"defaultProject":     "proj",
					"clientEmail":        "sa@proj.iam.gserviceaccount.com",
					"tokenUri":           "https://oauth2.googleapis.com/token",
					"universeDomain":     "googleapis.mtls.google.com",
				}),
				DecryptedSecureJSONData: map[string]string{"privateKey": "PEM"},
			},
			wantAuth:     AuthTypeJWT,
			wantProject:  "proj",
			wantUniverse: "googleapis.mtls.google.com",
			wantSecrets:  map[SecureJsonDataKey]string{SecureJsonDataKeyPrivateKey: "PEM"},
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
			if tc.wantProject != "" && cfg.DefaultProject != tc.wantProject {
				t.Errorf("DefaultProject: want %q, got %q", tc.wantProject, cfg.DefaultProject)
			}
			if tc.wantUniverse != "" && cfg.UniverseDomain != tc.wantUniverse {
				t.Errorf("UniverseDomain: want %q, got %q", tc.wantUniverse, cfg.UniverseDomain)
			}
			if tc.wantImperson && !cfg.UsingImpersonation {
				t.Errorf("UsingImpersonation: want true, got false")
			}
			if tc.wantWIFPool != "" && cfg.WorkloadIdentityPoolProvider != tc.wantWIFPool {
				t.Errorf("WorkloadIdentityPoolProvider: want %q, got %q", tc.wantWIFPool, cfg.WorkloadIdentityPoolProvider)
			}
			if cfg.OAuthPassthroughEnabled != tc.wantOAuthPass {
				t.Errorf("OAuthPassthroughEnabled: want %v, got %v", tc.wantOAuthPass, cfg.OAuthPassthroughEnabled)
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
		name          string
		cfg           Config
		wantAuth      AuthType
		wantOAuthPass bool
	}{
		{"empty auth defaults to jwt (no oauthPassThru)", Config{}, AuthTypeJWT, false},
		{"already-set jwt (no oauthPassThru)", Config{AuthenticationType: AuthTypeJWT}, AuthTypeJWT, false},
		{"already-set gce (no oauthPassThru)", Config{AuthenticationType: AuthTypeGCE}, AuthTypeGCE, false},
		{"forwardOAuthIdentity implies oauthPassThru", Config{AuthenticationType: AuthTypeForwardOAuthIdentity}, AuthTypeForwardOAuthIdentity, true},
		{"WIF implies oauthPassThru", Config{AuthenticationType: AuthTypeWorkloadIdentityFederation}, AuthTypeWorkloadIdentityFederation, true},
		{
			"switching to jwt clears a previously-set oauthPassThru",
			Config{AuthenticationType: AuthTypeJWT, OAuthPassthroughEnabled: true},
			AuthTypeJWT, false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.ApplyDefaults()
			if tc.cfg.AuthenticationType != tc.wantAuth {
				t.Errorf("AuthenticationType: want %q, got %q", tc.wantAuth, tc.cfg.AuthenticationType)
			}
			if tc.cfg.OAuthPassthroughEnabled != tc.wantOAuthPass {
				t.Errorf("OAuthPassthroughEnabled: want %v, got %v", tc.wantOAuthPass, tc.cfg.OAuthPassthroughEnabled)
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
			name: "jwt with privateKeyPath ok",
			cfg: Config{
				AuthenticationType: AuthTypeJWT,
				DefaultProject:     "p",
				ClientEmail:        "e",
				TokenURI:           "u",
				PrivateKeyPath:     "/etc/sa.json",
			},
		},
		{
			name: "jwt without any private key",
			cfg: Config{
				AuthenticationType: AuthTypeJWT,
				DefaultProject:     "p",
				ClientEmail:        "e",
				TokenURI:           "u",
			},
			wantErr: "secureJsonData.privateKey or jsonData.privateKeyPath is required",
		},
		{
			name:    "jwt missing every field",
			cfg:     Config{AuthenticationType: AuthTypeJWT},
			wantErr: "defaultProject is required",
		},
		{
			name: "gce ok without any field",
			cfg:  Config{AuthenticationType: AuthTypeGCE},
		},
		{
			name:    "forwardOAuthIdentity missing defaultProject",
			cfg:     Config{AuthenticationType: AuthTypeForwardOAuthIdentity},
			wantErr: "defaultProject is required for 'forwardOAuthIdentity'",
		},
		{
			name: "forwardOAuthIdentity ok with defaultProject",
			cfg:  Config{AuthenticationType: AuthTypeForwardOAuthIdentity, DefaultProject: "proj"},
		},
		{
			name: "workloadIdentityFederation ok with pool provider and project",
			cfg: Config{
				AuthenticationType:           AuthTypeWorkloadIdentityFederation,
				WorkloadIdentityPoolProvider: "projects/1/locations/global/workloadIdentityPools/p/providers/prov",
				DefaultProject:               "proj",
			},
		},
		{
			name:    "workloadIdentityFederation without pool provider",
			cfg:     Config{AuthenticationType: AuthTypeWorkloadIdentityFederation, DefaultProject: "proj"},
			wantErr: "workloadIdentityPoolProvider is required",
		},
		{
			name:    "workloadIdentityFederation without defaultProject",
			cfg:     Config{AuthenticationType: AuthTypeWorkloadIdentityFederation, WorkloadIdentityPoolProvider: "x"},
			wantErr: "defaultProject is required for 'workloadIdentityFederation'",
		},
		{
			name:    "empty auth",
			cfg:     Config{},
			wantErr: "authenticationType is required",
		},
		{
			name:    "unknown auth",
			cfg:     Config{AuthenticationType: "basic"},
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
