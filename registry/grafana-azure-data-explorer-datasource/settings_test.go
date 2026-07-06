package azuredataexplorerdatasource

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

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
		name         string
		settings     backend.DataSourceInstanceSettings
		wantErr      string
		wantAuthType AuthType
	}{
		{
			name: "clientsecret ok",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":   "clientsecret",
						"azureCloud": "AzureCloud",
						"tenantId":   "t",
						"clientId":   "c",
					},
					"clusterUrl":   "https://cluster.kusto.windows.net",
					"queryTimeout": "45s",
				}),
				DecryptedSecureJSONData: map[string]string{"azureClientSecret": "s"},
			},
			wantAuthType: AuthTypeClientSecret,
		},
		{
			name: "clientsecret missing secret fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "clientsecret"},
				}),
			},
			wantErr: "azureClientSecret",
		},
		{
			name: "clientsecret falls back to legacy clientSecret",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "clientsecret"},
				}),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "legacy"},
			},
			wantAuthType: AuthTypeClientSecret,
		},
		{
			name: "clientsecret-obo requires oauthPassThru",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "clientsecret-obo"},
				}),
				DecryptedSecureJSONData: map[string]string{"azureClientSecret": "s"},
			},
			wantErr: "oauthPassThru",
		},
		{
			name: "clientsecret-obo ok with oauthPassThru",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":   "clientsecret-obo",
						"azureCloud": "AzureCloud",
						"tenantId":   "t",
						"clientId":   "c",
					},
					"oauthPassThru": true,
				}),
				DecryptedSecureJSONData: map[string]string{"azureClientSecret": "s"},
			},
			wantAuthType: AuthTypeClientSecretOBO,
		},
		{
			name: "managed identity ok",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "msi"},
				}),
			},
			wantAuthType: AuthTypeManagedIdentity,
		},
		{
			name: "workload identity ok",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "workloadidentity"},
				}),
			},
			wantAuthType: AuthTypeWorkloadIdentity,
		},
		{
			name: "current user ok",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "currentuser"},
				}),
			},
			wantAuthType: AuthTypeCurrentUser,
		},
		{
			name: "empty authType with azureCredentials fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{"azureCredentials": map[string]any{}}),
			},
			wantErr: "authType",
		},
		{
			name: "unknown authType fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "bogus"},
				}),
			},
			wantErr: "unknown",
		},
		{
			name: "legacy clientsecret ok",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCloud": "azuremonitor",
					"tenantId":   "t",
					"clientId":   "c",
				}),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "s"},
			},
			wantAuthType: AuthTypeClientSecret,
		},
		{
			name: "legacy with onBehalfOf ok when oauthPassThru true",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCloud":    "azuremonitor",
					"tenantId":      "t",
					"clientId":      "c",
					"onBehalfOf":    true,
					"oauthPassThru": true,
				}),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "s"},
			},
			wantAuthType: AuthTypeClientSecretOBO,
		},
		{
			name: "legacy with onBehalfOf fails without oauthPassThru",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"tenantId":   "t",
					"clientId":   "c",
					"onBehalfOf": true,
				}),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "s"},
			},
			wantErr: "oauthPassThru",
		},
		{
			name: "legacy missing secret fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"tenantId": "t",
					"clientId": "c",
				}),
			},
			wantErr: "legacy credentials require",
		},
		{
			name: "invalid queryTimeout errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"queryTimeout": "not-a-duration",
				}),
			},
			wantErr: "queryTimeout",
		},
		{
			name: "queryTimeout too big errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"queryTimeout": "2h",
				}),
			},
			wantErr: "one-hour maximum",
		},
		{
			name: "unknown dataConsistency errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"dataConsistency": "eventual",
				}),
			},
			wantErr: "dataConsistency",
		},
		{
			name: "malformed jsonData",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name:     "empty settings validate cleanly (no auth chosen yet)",
			settings: backend.DataSourceInstanceSettings{},
		},
		{
			name: "OpenAI key parses into secure data",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "msi"},
				}),
				DecryptedSecureJSONData: map[string]string{"OpenAIAPIKey": "sk-abc"},
			},
			wantAuthType: AuthTypeManagedIdentity,
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
			if tc.wantAuthType != "" && cfg.EffectiveAuthType() != tc.wantAuthType {
				t.Errorf("EffectiveAuthType: want %q, got %q", tc.wantAuthType, cfg.EffectiveAuthType())
			}
		})
	}
}

func TestEffectiveAuthType(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want AuthType
	}{
		{
			name: "empty",
			cfg:  Config{},
			want: "",
		},
		{
			name: "modern beats legacy",
			cfg: Config{
				AzureCredentials: json.RawMessage(`{"authType":"msi"}`),
				TenantID:         "t",
				ClientID:         "c",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyClientSecret: "s",
				},
			},
			want: AuthTypeManagedIdentity,
		},
		{
			name: "legacy tuple resolves to clientsecret",
			cfg: Config{
				TenantID: "t",
				ClientID: "c",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyClientSecret: "s",
				},
			},
			want: AuthTypeClientSecret,
		},
		{
			name: "legacy tuple with onBehalfOf resolves to obo",
			cfg: Config{
				TenantID:   "t",
				ClientID:   "c",
				OnBehalfOf: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyClientSecret: "s",
				},
			},
			want: AuthTypeClientSecretOBO,
		},
		{
			name: "legacy tuple incomplete resolves to empty",
			cfg:  Config{TenantID: "t", ClientID: "c"},
			want: "",
		},
		{
			name: "malformed azureCredentials falls back",
			cfg: Config{
				AzureCredentials: json.RawMessage(`not-json`),
				TenantID:         "t",
				ClientID:         "c",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyClientSecret: "s",
				},
			},
			want: AuthTypeClientSecret,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.EffectiveAuthType(); got != tc.want {
				t.Errorf("EffectiveAuthType: want %q, got %q", tc.want, got)
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
			name: "empty gets strong + visual",
			in:   Config{},
			want: Config{DataConsistency: DataConsistencyStrong, DefaultEditorMode: EditorModeVisual},
		},
		{
			name: "existing consistency preserved",
			in:   Config{DataConsistency: DataConsistencyWeak},
			want: Config{DataConsistency: DataConsistencyWeak, DefaultEditorMode: EditorModeVisual},
		},
		{
			name: "existing editor mode preserved",
			in:   Config{DefaultEditorMode: EditorModeRaw},
			want: Config{DataConsistency: DataConsistencyStrong, DefaultEditorMode: EditorModeRaw},
		},
		{
			name: "unrelated fields untouched",
			in:   Config{ClusterURL: "https://x", DataConsistency: DataConsistencyStrong, DefaultEditorMode: EditorModeVisual},
			want: Config{ClusterURL: "https://x", DataConsistency: DataConsistencyStrong, DefaultEditorMode: EditorModeVisual},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.DataConsistency != tt.want.DataConsistency {
				t.Errorf("DataConsistency = %q, want %q", got.DataConsistency, tt.want.DataConsistency)
			}
			if got.DefaultEditorMode != tt.want.DefaultEditorMode {
				t.Errorf("DefaultEditorMode = %q, want %q", got.DefaultEditorMode, tt.want.DefaultEditorMode)
			}
			if got.ClusterURL != tt.want.ClusterURL {
				t.Errorf("ClusterURL = %q, want %q", got.ClusterURL, tt.want.ClusterURL)
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
			name: "modern msi",
			cfg:  Config{AzureCredentials: json.RawMessage(`{"authType":"msi"}`)},
		},
		{
			name: "modern clientsecret with azureClientSecret",
			cfg: Config{
				AzureCredentials: json.RawMessage(`{"authType":"clientsecret"}`),
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAzureClientSecret: "s",
				},
			},
		},
		{
			name: "modern clientsecret-obo without oauthPassThru fails",
			cfg: Config{
				AzureCredentials: json.RawMessage(`{"authType":"clientsecret-obo"}`),
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAzureClientSecret: "s",
				},
			},
			wantErr: "oauthPassThru",
		},
		{
			name: "modern clientsecret without any secret",
			cfg: Config{
				AzureCredentials: json.RawMessage(`{"authType":"clientsecret"}`),
			},
			wantErr: "azureClientSecret",
		},
		{
			name:    "queryTimeout > 1h",
			cfg:     Config{QueryTimeout: 2 * time.Hour},
			wantErr: "one-hour maximum",
		},
		{
			name:    "unknown data consistency",
			cfg:     Config{DataConsistency: "eventual"},
			wantErr: "dataConsistency",
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
