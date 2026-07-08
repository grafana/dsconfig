package azuremonitordatasource

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
			name: "clientcertificate PEM ok",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":          "clientcertificate",
						"certificateFormat": "pem",
					},
				}),
				DecryptedSecureJSONData: map[string]string{
					"clientCertificate": "pem-body",
					"privateKey":        "pem-key",
				},
			},
			wantAuthType: AuthTypeClientCertificate,
		},
		{
			name: "clientcertificate PEM missing privateKey fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":          "clientcertificate",
						"certificateFormat": "pem",
					},
				}),
				DecryptedSecureJSONData: map[string]string{"clientCertificate": "pem-body"},
			},
			wantErr: "privateKey",
		},
		{
			name: "clientcertificate PFX ok",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":          "clientcertificate",
						"certificateFormat": "pfx",
					},
				}),
				DecryptedSecureJSONData: map[string]string{
					"clientCertificate":   "pfx-bundle",
					"certificatePassword": "pw",
				},
			},
			wantAuthType: AuthTypeClientCertificate,
		},
		{
			name: "clientcertificate PFX missing password fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":          "clientcertificate",
						"certificateFormat": "pfx",
					},
				}),
				DecryptedSecureJSONData: map[string]string{"clientCertificate": "pfx-bundle"},
			},
			wantErr: "certificatePassword",
		},
		{
			name: "clientcertificate missing format fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "clientcertificate"},
				}),
				DecryptedSecureJSONData: map[string]string{"clientCertificate": "body"},
			},
			wantErr: "certificateFormat",
		},
		{
			name: "clientcertificate unknown format fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":          "clientcertificate",
						"certificateFormat": "der",
					},
				}),
				DecryptedSecureJSONData: map[string]string{"clientCertificate": "body"},
			},
			wantErr: "unknown certificateFormat",
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
			name: "current user ok (no fallback)",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{"authType": "currentuser"},
				}),
			},
			wantAuthType: AuthTypeCurrentUser,
		},
		{
			name: "ad-password missing secret fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType": "ad-password",
						"userId":   "u@contoso.onmicrosoft.com",
						"clientId": "c",
					},
				}),
			},
			wantErr: "password",
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
					"azureAuthType": "clientsecret",
					"cloudName":     "azuremonitor",
					"tenantId":      "t",
					"clientId":      "c",
				}),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "s"},
			},
			// EffectiveAuthType comes from azureAuthType when no azureCredentials.
			wantAuthType: AuthTypeClientSecret,
		},
		{
			name: "legacy clientsecret without secret fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{"azureAuthType": "clientsecret"}),
			},
			wantErr: "legacy clientsecret",
		},
		{
			name: "customized cloud ok with routes",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":   "clientsecret",
						"azureCloud": "AzureCustomizedCloud",
						"tenantId":   "t",
						"clientId":   "c",
					},
					"cloudName":        "customizedazuremonitor",
					"customizedRoutes": map[string]any{"Azure Monitor": map[string]any{"URL": "https://x"}},
				}),
				DecryptedSecureJSONData: map[string]string{"azureClientSecret": "s"},
			},
			wantAuthType: AuthTypeClientSecret,
		},
		{
			name: "customized cloud missing routes fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials": map[string]any{
						"authType":   "clientsecret",
						"azureCloud": "AzureCustomizedCloud",
					},
					"cloudName": "customizedazuremonitor",
				}),
				DecryptedSecureJSONData: map[string]string{"azureClientSecret": "s"},
			},
			wantErr: "customizedRoutes",
		},
		{
			name: "azureLogAnalyticsSameAs=false fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials":        map[string]any{"authType": "msi"},
					"azureLogAnalyticsSameAs": false,
				}),
			},
			wantErr: "azureLogAnalyticsSameAs=false",
		},
		{
			name: "azureLogAnalyticsSameAs string=\"false\" fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials":        map[string]any{"authType": "msi"},
					"azureLogAnalyticsSameAs": "false",
				}),
			},
			wantErr: "azureLogAnalyticsSameAs=false",
		},
		{
			name: "azureLogAnalyticsSameAs=true is fine",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials":        map[string]any{"authType": "msi"},
					"azureLogAnalyticsSameAs": true,
				}),
			},
			wantAuthType: AuthTypeManagedIdentity,
		},
		{
			name: "azureLogAnalyticsSameAs junk string fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: mustJSON(t, map[string]any{
					"azureCredentials":        map[string]any{"authType": "msi"},
					"azureLogAnalyticsSameAs": "definitely-not-a-bool",
				}),
			},
			wantErr: "azureLogAnalyticsSameAs",
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
				AzureAuthType:    AuthTypeClientSecret,
			},
			want: AuthTypeManagedIdentity,
		},
		{
			name: "legacy fallback when no modern",
			cfg:  Config{AzureAuthType: AuthTypeClientSecret},
			want: AuthTypeClientSecret,
		},
		{
			name: "malformed azureCredentials falls back to legacy",
			cfg: Config{
				AzureCredentials: json.RawMessage(`not-json`),
				AzureAuthType:    AuthTypeClientSecret,
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

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "modern msi",
			cfg: Config{
				AzureCredentials: json.RawMessage(`{"authType":"msi"}`),
			},
		},
		{
			name: "modern clientsecret with azureClientSecret",
			cfg: Config{
				AzureCredentials:        json.RawMessage(`{"authType":"clientsecret"}`),
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAzureClientSecret: "s"},
			},
		},
		{
			name: "modern clientsecret-obo with legacy clientSecret",
			cfg: Config{
				AzureCredentials:        json.RawMessage(`{"authType":"clientsecret-obo"}`),
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyClientSecret: "s"},
			},
		},
		{
			name: "modern clientsecret without any secret",
			cfg: Config{
				AzureCredentials: json.RawMessage(`{"authType":"clientsecret"}`),
			},
			wantErr: "azureClientSecret",
		},
		{
			name: "legacy clientsecret without secret",
			cfg: Config{
				AzureAuthType: AuthTypeClientSecret,
			},
			wantErr: "legacy clientsecret",
		},
		{
			name: "customized cloud requires routes",
			cfg: Config{
				CloudName: string(LegacyCloudNameCustomized),
				AzureCredentials: json.RawMessage(
					`{"authType":"clientsecret","azureCloud":"AzureCustomizedCloud"}`,
				),
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAzureClientSecret: "s"},
			},
			wantErr: "customizedRoutes",
		},
		{
			name: "azureLogAnalyticsSameAs false surfaces the deprecated-error",
			cfg: Config{
				AzureCredentials:        json.RawMessage(`{"authType":"msi"}`),
				AzureLogAnalyticsSameAs: json.RawMessage(`false`),
			},
			wantErr: "azureLogAnalyticsSameAs=false",
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

func TestApplyDefaults(t *testing.T) {
	// ApplyDefaults is a no-op today because Azure Monitor's editor picks
	// its default credential shape based on Grafana's runtime config
	// (managedIdentityEnabled / workloadIdentityEnabled / userIdentityEnabled).
	// The test locks in the no-op contract so accidental additions get caught.
	c := Config{}
	c.ApplyDefaults()
	if c.EffectiveAuthType() != "" {
		t.Errorf("ApplyDefaults must not synthesize a default auth type; got %q", c.EffectiveAuthType())
	}
	if len(c.AzureCredentials) != 0 {
		t.Errorf("ApplyDefaults must not populate AzureCredentials; got %s", string(c.AzureCredentials))
	}
}
