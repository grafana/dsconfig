package databricksdatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with jsonData and secureJsonData) into the
// backend.DataSourceInstanceSettings shape LoadConfig expects.
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
		name           string
		example        string // SettingsExamples key ("" handled via inline settings otherwise)
		settings       backend.DataSourceInstanceSettings
		wantErr        error
		wantAuthType   AuthType
		wantHost       string
		wantClientID   string
		wantTenantID   string
		wantAzureCloud string
		wantSecureKeys SecureJsonDataConfig
	}{
		{
			// The default schema example has empty host/httpPath/token, so
			// LoadConfig's Validate step is expected to reject it.
			name:    "default example fails validation",
			example: "",
			wantErr: ErrMissingHost,
		},
		{
			name:           "personal access token",
			example:        "personalAccessToken",
			wantAuthType:   AuthTypePat,
			wantHost:       "dbc-a1b2c3d4-e5f6.cloud.databricks.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			name:         "oauth passthrough needs no secret",
			example:      "oauthPassthrough",
			wantAuthType: AuthTypeOauthPT,
			wantHost:     "dbc-a1b2c3d4-e5f6.cloud.databricks.com",
		},
		{
			name:           "oauth m2m",
			example:        "oauthM2M",
			wantAuthType:   AuthTypeOauthM2M,
			wantClientID:   "11111111-1111-1111-1111-111111111111",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyClientSecret},
		},
		{
			name:           "azure m2m",
			example:        "azureM2M",
			wantAuthType:   AuthTypeAzureM2M,
			wantClientID:   "11111111-1111-1111-1111-111111111111",
			wantTenantID:   "22222222-2222-2222-2222-222222222222",
			wantAzureCloud: "AzureCloud",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyClientSecret},
		},
		{
			name:           "azure on-behalf-of",
			example:        "azureOBO",
			wantAuthType:   AuthTypeOauthOBO,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAzureClientSecret},
		},
		{
			name: "legacy config without authType defaults to Pat",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"host":"h","httpPath":"p"}`),
				DecryptedSecureJSONData: map[string]string{"token": "tok"},
			},
			wantAuthType:   AuthTypePat,
			wantHost:       "h",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			name: "azure m2m accepts backend clientID/tenantID casing",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"AzureM2M","host":"h","httpPath":"p","tenantID":"t-upper","clientID":"c-upper","azureCloud":"AzureCloud"}`),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "s"},
			},
			wantAuthType: AuthTypeAzureM2M,
			wantClientID: "c-upper",
			wantTenantID: "t-upper",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "missing host errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"Pat","httpPath":"p"}`),
				DecryptedSecureJSONData: map[string]string{"token": "tok"},
			},
			wantErr: ErrMissingHost,
		},
		{
			name: "pat missing token errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"Pat","host":"h","httpPath":"p"}`),
			},
			wantErr: ErrMissingToken,
		},
		{
			name: "azure m2m missing tenantId errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"AzureM2M","host":"h","httpPath":"p","clientId":"c"}`),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "s"},
			},
			wantErr: ErrMissingTenantID,
		},
		{
			name: "on-behalf-of without oauthPassThru errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"OauthOBO","host":"h","httpPath":"p","azureCredentials":{"authType":"clientsecret-obo"}}`),
				DecryptedSecureJSONData: map[string]string{"azureClientSecret": "s"},
			},
			wantErr: ErrInvalidOAuth,
		},
		{
			name: "unknown auth type errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"bogus","host":"h","httpPath":"p"}`),
			},
			wantErr: errors.New(`unknown authType "bogus"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Tests with no inline settings pull their input from the named
			// SettingsExamples entry (tt.example, "" = the default example).
			settings := tt.settings
			if tt.settings.JSONData == nil && tt.settings.DecryptedSecureJSONData == nil {
				settings = settingsFromExample(t, tt.example)
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("LoadConfig: expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Fatalf("LoadConfig: error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}

			if tt.wantAuthType != "" && cfg.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", cfg.AuthType, tt.wantAuthType)
			}
			if tt.wantHost != "" && cfg.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", cfg.Host, tt.wantHost)
			}
			if tt.wantClientID != "" && cfg.ClientID != tt.wantClientID {
				t.Errorf("ClientID = %q, want %q", cfg.ClientID, tt.wantClientID)
			}
			if tt.wantTenantID != "" && cfg.TenantID != tt.wantTenantID {
				t.Errorf("TenantID = %q, want %q", cfg.TenantID, tt.wantTenantID)
			}
			if tt.wantAzureCloud != "" && cfg.AzureCloud != tt.wantAzureCloud {
				t.Errorf("AzureCloud = %q, want %q", cfg.AzureCloud, tt.wantAzureCloud)
			}
			// CloudFetch is force-defaulted to true on every successful load.
			if !cfg.CloudFetch {
				t.Errorf("CloudFetch = false, want true (backend force-enable)")
			}
			if tt.wantSecureKeys != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if _, ok := cfg.DecryptedSecureJSONData[key]; ok {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantSecureKeys) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantSecureKeys)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name           string
		in             Config
		wantAuthType   AuthType
		wantCloudFetch bool
		wantAzureCloud string
	}{
		{
			name:           "empty config gets Pat + cloudFetch",
			in:             Config{},
			wantAuthType:   AuthTypePat,
			wantCloudFetch: true,
		},
		{
			name:           "existing auth type is preserved",
			in:             Config{AuthType: AuthTypeOauthM2M},
			wantAuthType:   AuthTypeOauthM2M,
			wantCloudFetch: true,
		},
		{
			name:           "azure m2m defaults cloud to AzureCloud",
			in:             Config{AuthType: AuthTypeAzureM2M},
			wantAuthType:   AuthTypeAzureM2M,
			wantCloudFetch: true,
			wantAzureCloud: "AzureCloud",
		},
		{
			name:           "azure m2m preserves explicit cloud",
			in:             Config{AuthType: AuthTypeAzureM2M, AzureCloud: "AzureChinaCloud"},
			wantAuthType:   AuthTypeAzureM2M,
			wantCloudFetch: true,
			wantAzureCloud: "AzureChinaCloud",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuthType)
			}
			if got.CloudFetch != tt.wantCloudFetch {
				t.Errorf("CloudFetch = %v, want %v", got.CloudFetch, tt.wantCloudFetch)
			}
			if got.AzureCloud != tt.wantAzureCloud {
				t.Errorf("AzureCloud = %q, want %q", got.AzureCloud, tt.wantAzureCloud)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "pat happy path",
			cfg: Config{
				Host: "h", HTTPPath: "p", AuthType: AuthTypePat,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "tok"},
			},
		},
		{
			name: "unknown auth treated as pat requires token",
			cfg: Config{
				Host: "h", HTTPPath: "p", AuthType: AuthTypeUnknown,
			},
			wantErr: "missing token",
		},
		{
			name:    "missing host and httpPath",
			cfg:     Config{AuthType: AuthTypeOauthPT},
			wantErr: "missing host",
		},
		{
			name: "oauth passthrough happy path",
			cfg:  Config{Host: "h", HTTPPath: "p", AuthType: AuthTypeOauthPT},
		},
		{
			name: "oauth m2m missing client secret",
			cfg: Config{
				Host: "h", HTTPPath: "p", AuthType: AuthTypeOauthM2M, ClientID: "c",
			},
			wantErr: "missing clientSecret",
		},
		{
			name: "oauth m2m happy path",
			cfg: Config{
				Host: "h", HTTPPath: "p", AuthType: AuthTypeOauthM2M, ClientID: "c",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyClientSecret: "s"},
			},
		},
		{
			name: "azure m2m happy path",
			cfg: Config{
				Host: "h", HTTPPath: "p", AuthType: AuthTypeAzureM2M, TenantID: "t", ClientID: "c",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyClientSecret: "s"},
			},
		},
		{
			name: "on-behalf-of happy path",
			cfg: Config{
				Host: "h", HTTPPath: "p", AuthType: AuthTypeOauthOBO, OAuthPassThru: true,
				AzureCredentials:        json.RawMessage(`{"authType":"clientsecret-obo"}`),
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAzureClientSecret: "s"},
			},
		},
		{
			name: "on-behalf-of missing secret",
			cfg: Config{
				Host: "h", HTTPPath: "p", AuthType: AuthTypeOauthOBO, OAuthPassThru: true,
				AzureCredentials: json.RawMessage(`{"authType":"clientsecret-obo"}`),
			},
			wantErr: "azureClientSecret is required",
		},
		{
			name:    "unknown auth type",
			cfg:     Config{Host: "h", HTTPPath: "p", AuthType: "bogus"},
			wantErr: `unknown authType "bogus"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate: unexpected error %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate: expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate: error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
