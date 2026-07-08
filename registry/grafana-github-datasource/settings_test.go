package githubdatasource

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
		name            string
		example         string // schema.go SettingsExamples key ("" = use inline settings)
		settings        backend.DataSourceInstanceSettings
		wantErr         error
		wantAuthType    AuthType
		wantPlan        LicenseType
		wantGithubURL   string
		wantSecureKeys  SecureJsonDataConfig
		wantAccessToken string
		wantPrivateKey  string
		checkAppIDs     bool
		wantAppID       int64
		wantInstallID   int64
		wantAppIDStr    string
		wantInstallStr  string
	}{
		{
			// The default schema example intentionally has an empty accessToken
			// placeholder, so LoadConfig's Validate step is expected to reject it.
			name:    "default example fails validation (empty accessToken placeholder)",
			example: "",
			wantErr: errors.New("access token is required"),
		},
		{
			name:            "personal access token",
			example:         "personalAccessToken",
			wantAuthType:    AuthTypePAT,
			wantPlan:        LicenseTypeBasic,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "github_pat_XXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name:           "github app",
			example:        "githubApp",
			wantAuthType:   AuthTypeGithubApp,
			wantPlan:       LicenseTypeBasic,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPrivateKey},
			wantPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
			checkAppIDs:    true,
			wantAppID:      123456,
			wantInstallID:  12345678,
			wantAppIDStr:   "123456",
			wantInstallStr: "12345678",
		},
		{
			name:            "enterprise cloud",
			example:         "enterpriseCloud",
			wantAuthType:    AuthTypePAT,
			wantPlan:        LicenseTypeEnterpriseCloud,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "github_pat_XXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name:            "enterprise server",
			example:         "enterpriseServer",
			wantAuthType:    AuthTypePAT,
			wantPlan:        LicenseTypeEnterpriseServer,
			wantGithubURL:   "https://github.example.com",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "github_pat_XXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name:           "github app enterprise server",
			example:        "githubAppEnterpriseServer",
			wantAuthType:   AuthTypeGithubApp,
			wantPlan:       LicenseTypeEnterpriseServer,
			wantGithubURL:  "https://github.example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPrivateKey},
			wantPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
			checkAppIDs:    true,
			wantAppID:      123456,
			wantInstallID:  12345678,
		},
		{
			name:            "legacy access token defaults to PAT",
			example:         "legacyAccessTokenOnly",
			wantAuthType:    AuthTypePAT,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "github_pat_XXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "invalid appId errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"selectedAuthType":"github-app","appId":"not-a-number"}`),
			},
			wantErr: errors.New("error parsing app id"),
		},
		{
			name: "PAT skips numeric appId parsing",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"selectedAuthType":"personal-access-token","appId":"not-a-number"}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantAuthType: AuthTypePAT,
		},
		{
			name: "legacy github app with numeric ids normalizes to string",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"selectedAuthType":"github-app","appId":123456,"installationId":12345678}`),
				DecryptedSecureJSONData: map[string]string{"privateKey": "pem"},
			},
			wantAuthType:   AuthTypeGithubApp,
			checkAppIDs:    true,
			wantAppID:      123456,
			wantInstallID:  12345678,
			wantAppIDStr:   "123456",
			wantInstallStr: "12345678",
		},
		{
			name: "github app with empty appId errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"selectedAuthType":"github-app"}`),
			},
			wantErr: errors.New("error parsing app id"),
		},
		{
			name: "github app with empty installationId errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"selectedAuthType":"github-app","appId":"1"}`),
			},
			wantErr: errors.New("error parsing installation id"),
		},
		{
			// After ApplyDefaults, an empty config becomes PAT + basic, which
			// requires an accessToken — Validate rejects it.
			name:     "empty settings default to PAT and fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("access token is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" || (tt.settings.JSONData == nil && tt.settings.DecryptedSecureJSONData == nil && tt.wantErr == nil) {
				if _, ok := SettingsExamples().Examples[tt.example]; ok {
					settings = settingsFromExample(t, tt.example)
				}
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

			if tt.wantAuthType != "" && cfg.SelectedAuthType != tt.wantAuthType {
				t.Errorf("SelectedAuthType = %q, want %q", cfg.SelectedAuthType, tt.wantAuthType)
			}
			if tt.wantPlan != "" && cfg.GithubPlan != tt.wantPlan {
				t.Errorf("GithubPlan = %q, want %q", cfg.GithubPlan, tt.wantPlan)
			}
			if tt.wantGithubURL != "" && cfg.GitHubURL != tt.wantGithubURL {
				t.Errorf("GitHubURL = %q, want %q", cfg.GitHubURL, tt.wantGithubURL)
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
			if tt.wantAccessToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken] != tt.wantAccessToken {
				t.Errorf("Secrets[accessToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken], tt.wantAccessToken)
			}
			if tt.wantPrivateKey != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPrivateKey] != tt.wantPrivateKey {
				t.Errorf("Secrets[privateKey] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPrivateKey], tt.wantPrivateKey)
			}
			if tt.checkAppIDs {
				if cfg.AppIdInt64 != tt.wantAppID {
					t.Errorf("AppIdInt64 = %d, want %d", cfg.AppIdInt64, tt.wantAppID)
				}
				if cfg.InstallationIdInt64 != tt.wantInstallID {
					t.Errorf("InstallationIdInt64 = %d, want %d", cfg.InstallationIdInt64, tt.wantInstallID)
				}
			}
			if tt.wantAppIDStr != "" && cfg.AppId != tt.wantAppIDStr {
				t.Errorf("AppId = %q, want %q", cfg.AppId, tt.wantAppIDStr)
			}
			if tt.wantInstallStr != "" && cfg.InstallationId != tt.wantInstallStr {
				t.Errorf("InstallationId = %q, want %q", cfg.InstallationId, tt.wantInstallStr)
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
			name: "empty config gets PAT + basic",
			in:   Config{},
			want: Config{SelectedAuthType: AuthTypePAT, GithubPlan: LicenseTypeBasic},
		},
		{
			name: "existing auth type is preserved",
			in:   Config{SelectedAuthType: AuthTypeGithubApp},
			want: Config{SelectedAuthType: AuthTypeGithubApp, GithubPlan: LicenseTypeBasic},
		},
		{
			name: "existing plan is preserved",
			in:   Config{GithubPlan: LicenseTypeEnterpriseServer},
			want: Config{SelectedAuthType: AuthTypePAT, GithubPlan: LicenseTypeEnterpriseServer},
		},
		{
			name: "unrelated zero fields untouched",
			in:   Config{SelectedAuthType: AuthTypePAT, GithubPlan: LicenseTypeBasic},
			want: Config{SelectedAuthType: AuthTypePAT, GithubPlan: LicenseTypeBasic},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.SelectedAuthType != tt.want.SelectedAuthType {
				t.Errorf("SelectedAuthType = %q, want %q", got.SelectedAuthType, tt.want.SelectedAuthType)
			}
			if got.GithubPlan != tt.want.GithubPlan {
				t.Errorf("GithubPlan = %q, want %q", got.GithubPlan, tt.want.GithubPlan)
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
			name: "PAT with accessToken",
			cfg: Config{
				SelectedAuthType:        AuthTypePAT,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
		},
		{
			name:    "PAT without accessToken errors",
			cfg:     Config{SelectedAuthType: AuthTypePAT},
			wantErr: "access token is required",
		},
		{
			name: "github-app happy path",
			cfg: Config{
				SelectedAuthType:        AuthTypeGithubApp,
				AppIdInt64:              123,
				InstallationIdInt64:     456,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPrivateKey: "pem"},
			},
		},
		{
			name:    "github-app missing everything",
			cfg:     Config{SelectedAuthType: AuthTypeGithubApp},
			wantErr: "appId is required",
		},
		{
			name: "github-app missing privateKey",
			cfg: Config{
				SelectedAuthType:    AuthTypeGithubApp,
				AppIdInt64:          123,
				InstallationIdInt64: 456,
			},
			wantErr: "privateKey is required",
		},
		{
			name:    "empty auth type errors",
			cfg:     Config{},
			wantErr: "selectedAuthType is required",
		},
		{
			name:    "unknown auth type errors",
			cfg:     Config{SelectedAuthType: "bogus"},
			wantErr: `unknown selectedAuthType "bogus"`,
		},
		{
			name: "enterprise server without url errors",
			cfg: Config{
				SelectedAuthType:        AuthTypePAT,
				GithubPlan:              LicenseTypeEnterpriseServer,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
			wantErr: "githubUrl is required",
		},
		{
			name: "enterprise server with url",
			cfg: Config{
				SelectedAuthType:        AuthTypePAT,
				GithubPlan:              LicenseTypeEnterpriseServer,
				GitHubURL:               "https://github.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
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
