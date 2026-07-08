package bigquerydatasource

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
		wantMaxBytes  int64
		wantLocation  string
		wantEndpoint  string
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
			name: "JWT missing all credentials surfaces multiple errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
				}),
			},
			wantErr: "defaultProject is required",
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
			name: "GCE without defaultProject still succeeds",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "gce",
				}),
			},
			wantAuth: AuthTypeGCE,
		},
		{
			name: "forwardOAuthIdentity happy path sets oauthPassThru true",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "forwardOAuthIdentity",
				}),
			},
			wantAuth:      AuthTypeForwardOAuthIdentity,
			wantOAuthPass: true,
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
				}),
			},
			wantErr: "workloadIdentityPoolProvider is required",
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
			name: "MaxBytesBilled is loaded and processingLocation, serviceEndpoint",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "jwt",
					"defaultProject":     "proj",
					"clientEmail":        "sa@proj.iam.gserviceaccount.com",
					"tokenUri":           "https://oauth2.googleapis.com/token",
					"processingLocation": "EU",
					"serviceEndpoint":    "https://bigquery.googleapis.com/bigquery/v2/",
					"MaxBytesBilled":     int64(5242880),
				}),
				DecryptedSecureJSONData: map[string]string{"privateKey": "PEM"},
			},
			wantAuth:     AuthTypeJWT,
			wantProject:  "proj",
			wantLocation: "EU",
			wantEndpoint: "https://bigquery.googleapis.com/bigquery/v2/",
			wantMaxBytes: 5242880,
			wantSecrets:  map[SecureJsonDataKey]string{SecureJsonDataKeyPrivateKey: "PEM"},
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
			name: "negative MaxBytesBilled fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "gce",
					"MaxBytesBilled":     int64(-1),
				}),
			},
			wantErr: "MaxBytesBilled must be non-negative",
		},
		{
			name: "unknown queryPriority fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"authenticationType": "gce",
					"queryPriority":      "IMMEDIATE",
				}),
			},
			wantErr: "unknown queryPriority",
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
			if tc.wantLocation != "" && cfg.ProcessingLocation != tc.wantLocation {
				t.Errorf("ProcessingLocation: want %q, got %q", tc.wantLocation, cfg.ProcessingLocation)
			}
			if tc.wantEndpoint != "" && cfg.ServiceEndpoint != tc.wantEndpoint {
				t.Errorf("ServiceEndpoint: want %q, got %q", tc.wantEndpoint, cfg.ServiceEndpoint)
			}
			if tc.wantMaxBytes != 0 && cfg.MaxBytesBilled != tc.wantMaxBytes {
				t.Errorf("MaxBytesBilled: want %d, got %d", tc.wantMaxBytes, cfg.MaxBytesBilled)
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
			name: "gce ok",
			cfg:  Config{AuthenticationType: AuthTypeGCE},
		},
		{
			name: "forwardOAuthIdentity ok",
			cfg:  Config{AuthenticationType: AuthTypeForwardOAuthIdentity},
		},
		{
			name: "workloadIdentityFederation ok with pool provider",
			cfg: Config{
				AuthenticationType:           AuthTypeWorkloadIdentityFederation,
				WorkloadIdentityPoolProvider: "projects/1/locations/global/workloadIdentityPools/p/providers/prov",
			},
		},
		{
			name:    "workloadIdentityFederation without pool provider",
			cfg:     Config{AuthenticationType: AuthTypeWorkloadIdentityFederation},
			wantErr: "workloadIdentityPoolProvider is required",
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
		{
			name:    "negative MaxBytesBilled",
			cfg:     Config{AuthenticationType: AuthTypeGCE, MaxBytesBilled: -1},
			wantErr: "MaxBytesBilled must be non-negative",
		},
		{
			name:    "unknown queryPriority",
			cfg:     Config{AuthenticationType: AuthTypeGCE, QueryPriority: "IMMEDIATE"},
			wantErr: "unknown queryPriority",
		},
		{
			name: "known queryPriority values ok",
			cfg: Config{
				AuthenticationType: AuthTypeGCE,
				QueryPriority:      QueryPriorityBatch,
			},
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
